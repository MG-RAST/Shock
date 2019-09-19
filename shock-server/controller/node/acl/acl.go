//Package acl implements /node/:id/acl resource
package acl

import (
	"errors"
	"net/http"
	"strings"

	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/node/locker"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/MG-RAST/golib/go-uuid/uuid"
	"github.com/MG-RAST/golib/stretchr/goweb/context"
	mgo "gopkg.in/mgo.v2"
)

var (
	validAclTypes = map[string]bool{"all": true, "read": true, "write": true, "delete": true, "owner": true,
		"public_all": true, "public_read": true, "public_write": true, "public_delete": true}
)

// GET, POST, PUT, DELETE: /node/{nid}/acl/
// GET is the only action implemented here.
func AclRequest(ctx context.Context) {
	nid := ctx.PathValue("nid")
	rmeth := ctx.HttpRequest().Method

	u, err := request.Authenticate(ctx.HttpRequest())
	if err != nil && err.Error() != e.NoAuth {
		request.AuthError(err, ctx)
		return
	}

	// public user (no auth) can perform a GET operation with the proper node permissions
	if u == nil {
		if rmeth == "GET" && conf.ANON_READ {
			u = &user.User{Uuid: "public"}
		} else {
			responder.RespondWithError(ctx, http.StatusUnauthorized, e.NoAuth)
			return
		}
	}

	// Load node by id
	n, err := node.Load(nid)
	if err != nil {
		if err == mgo.ErrNotFound {
			logger.Error("err@node_Acl: (node.Load) id=" + nid + ": " + e.NodeNotFound)
			responder.RespondWithError(ctx, http.StatusNotFound, e.NodeNotFound)
			return
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			err_msg := "err@node_Acl: (node.Load) id=" + nid + ": " + err.Error()
			logger.Error(err_msg)
			responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
			return
		}
	}

	// Only the owner, an admin, or someone with read access can view acl's.
	//
	// NOTE: If the node is publicly owned, then anyone can view all acl's. The owner can only
	//       be "public" when anonymous node creation (ANON_WRITE) is enabled in Shock config.

	rights := n.Acl.Check(u.Uuid)
	if n.Acl.Owner != u.Uuid && u.Admin == false && n.Acl.Owner != "public" && rights["read"] == false {
		logger.Error("err@node_Acl: (Authenticate) id=" + nid + ": " + e.UnAuth)
		responder.RespondWithError(ctx, http.StatusUnauthorized, e.UnAuth)
		return
	}

	if rmeth == "GET" {
		query := ctx.HttpRequest().URL.Query()
		verbosity := ""
		if _, ok := query["verbosity"]; ok {
			verbosity = query.Get("verbosity")
		}
		responder.RespondWithData(ctx, n.Acl.FormatDisplayAcl(verbosity))
	} else {
		responder.RespondWithError(ctx, http.StatusNotImplemented, "This request type is not implemented.")
	}
	return
}

// GET, POST, PUT, DELETE: /node/{nid}/acl/{type}
func AclTypedRequest(ctx context.Context) {
	nid := ctx.PathValue("nid")
	rtype := ctx.PathValue("type")
	rmeth := ctx.HttpRequest().Method
	query := ctx.HttpRequest().URL.Query()
	verbosity := ""
	if _, ok := query["verbosity"]; ok {
		verbosity = query.Get("verbosity")
	}

	u, err := request.Authenticate(ctx.HttpRequest())
	if err != nil && err.Error() != e.NoAuth {
		request.AuthError(err, ctx)
		return
	}

	if !validAclTypes[rtype] {
		responder.RespondWithError(ctx, http.StatusBadRequest, "Invalid acl type")
		return
	}

	// Load node by id
	n, err := node.Load(nid)
	if err != nil {
		if err == mgo.ErrNotFound {
			logger.Error("err@node_Acl: (node.Load) id=" + nid + ": " + e.NodeNotFound)
			responder.RespondWithError(ctx, http.StatusNotFound, e.NodeNotFound)
			return
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			err_msg := "err@node_Acl: (node.Load) id=" + nid + ": " + err.Error()
			logger.Error(err_msg)
			responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
			return
		}
	}

	// public user (no auth) can perform a GET operation given the proper node permissions
	if u == nil {
		rights := n.Acl.Check("public")
		if rmeth == "GET" && conf.ANON_READ && (rights["read"] || n.Acl.Owner == "public") {
			responder.RespondWithData(ctx, n.Acl.FormatDisplayAcl(verbosity))
			return
		} else {
			responder.RespondWithError(ctx, http.StatusUnauthorized, e.NoAuth)
			return
		}
	}

	// lock node
	err = locker.NodeLockMgr.LockNode(nid)
	if err != nil {
		err_msg := "err@node_Acl: (LockNode) id=" + nid + ": " + err.Error()
		logger.Error(err_msg)
		responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
		return
	}
	defer locker.NodeLockMgr.UnlockNode(nid)

	// Users that are not an admin or the node owner can only delete themselves from an ACL.
	if n.Acl.Owner != u.Uuid && u.Admin == false {
		// Users that are not an admin or the node owner cannot remove public from ACL's.
		if rtype == "public_read" || rtype == "public_write" || rtype == "public_delete" || rtype == "public_all" {
			err_msg := "Users that are not node owners can only delete themselves from ACLs."
			logger.Error("err@node_Acl: id=" + nid + ": " + err_msg)
			responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
			return
		}

		// Parse user list
		ids, err := parseAclRequestTyped(ctx)
		if err != nil {
			err_msg := "err@node_Acl: (parseAclRequestTyped) id=" + nid + ": " + err.Error()
			logger.Error(err_msg)
			responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
			return
		}
		if rmeth == "DELETE" {
			if len(ids) != 1 || (len(ids) == 1 && ids[0] != u.Uuid) {
				err_msg := "Users that are not node owners can delete only themselves from ACLs."
				logger.Error("err@node_Acl: id=" + nid + ": " + err_msg)
				responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
				return
			}
			if rtype == "owner" {
				err_msg := "Deleting node ownership is not a supported request type."
				logger.Error("err@node_Acl: id=" + nid + ": " + err_msg)
				responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
				return
			}
			if rtype == "all" {
				n.Acl.UnSet(ids[0], map[string]bool{"read": true, "write": true, "delete": true})
			} else {
				n.Acl.UnSet(ids[0], map[string]bool{rtype: true})
			}
			n.Save()
			responder.RespondWithData(ctx, n.Acl.FormatDisplayAcl(verbosity))
			return
		}
		err_msg := "Users that are not node owners can only delete themselves from ACLs."
		logger.Error("err@node_Acl: id=" + nid + ": " + err_msg)
		responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
		return
	}

	// At this point we know we're dealing with an admin or the node owner.
	// Admins and node owners can view/edit/delete ACLs
	if rmeth == "GET" {
		responder.RespondWithData(ctx, n.Acl.FormatDisplayAcl(verbosity))
		return
	} else if rmeth == "POST" || rmeth == "PUT" {
		if rtype == "public_read" || rtype == "public_write" || rtype == "public_delete" || rtype == "public_all" {
			if rtype == "public_read" {
				n.Acl.Set("public", map[string]bool{"read": true})
			} else if rtype == "public_write" {
				n.Acl.Set("public", map[string]bool{"write": true})
			} else if rtype == "public_delete" {
				n.Acl.Set("public", map[string]bool{"delete": true})
			} else if rtype == "public_all" {
				n.Acl.Set("public", map[string]bool{"read": true, "write": true, "delete": true})
			}
			n.Save()
			responder.RespondWithData(ctx, n.Acl.FormatDisplayAcl(verbosity))
			return
		}

		// Parse user list
		ids, err := parseAclRequestTyped(ctx)
		if err != nil {
			err_msg := "err@node_Acl: (parseAclRequestTyped) id=" + nid + ": " + err.Error()
			logger.Error(err_msg)
			responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
			return
		}
		if rtype == "owner" {
			if len(ids) == 1 {
				n.Acl.SetOwner(ids[0])
			} else {
				err_msg := "Too many users. Nodes may have only one owner."
				logger.Error("err@node_Acl: id=" + nid + ": " + err_msg)
				responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
				return
			}
		} else if rtype == "all" {
			for _, i := range ids {
				n.Acl.Set(i, map[string]bool{"read": true, "write": true, "delete": true})
			}
		} else {
			for _, i := range ids {
				n.Acl.Set(i, map[string]bool{rtype: true})
			}
		}
		n.Save()
		responder.RespondWithData(ctx, n.Acl.FormatDisplayAcl(verbosity))
		return
	} else if rmeth == "DELETE" {
		if rtype == "public_read" || rtype == "public_write" || rtype == "public_delete" || rtype == "public_all" {
			if rtype == "public_read" {
				n.Acl.UnSet("public", map[string]bool{"read": true})
			} else if rtype == "public_write" {
				n.Acl.UnSet("public", map[string]bool{"write": true})
			} else if rtype == "public_delete" {
				n.Acl.UnSet("public", map[string]bool{"delete": true})
			} else if rtype == "public_all" {
				n.Acl.UnSet("public", map[string]bool{"read": true, "write": true, "delete": true})
			}
			n.Save()
			responder.RespondWithData(ctx, n.Acl.FormatDisplayAcl(verbosity))
			return
		}

		// Parse user list
		ids, err := parseAclRequestTyped(ctx)
		if err != nil {
			err_msg := "err@node_Acl: (parseAclRequestTyped) id=" + nid + ": " + err.Error()
			logger.Error(err_msg)
			responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
			return
		}
		if rtype == "owner" {
			err_msg := "Deleting ownership is not a supported request type."
			logger.Error("err@node_Acl: id=" + nid + ": " + err_msg)
			responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
			return
		} else if rtype == "all" {
			for _, i := range ids {
				n.Acl.UnSet(i, map[string]bool{"read": true, "write": true, "delete": true})
			}
		} else {
			for _, i := range ids {
				n.Acl.UnSet(i, map[string]bool{rtype: true})
			}
		}
		n.Save()
		responder.RespondWithData(ctx, n.Acl.FormatDisplayAcl(verbosity))
		return
	} else {
		responder.RespondWithError(ctx, http.StatusNotImplemented, "This request type is not implemented.")
		return
	}
}

func parseAclRequestTyped(ctx context.Context) (ids []string, err error) {
	var users []string
	query := ctx.HttpRequest().URL.Query()
	params, _, err := request.ParseMultipartForm(ctx.HttpRequest())
	if _, ok := query["users"]; ok && err != nil && strings.Contains(err.Error(), http.ErrNotMultipart.ErrorString) {
		users = strings.Split(query.Get("users"), ",")
	} else if params["users"] != "" {
		users = strings.Split(params["users"], ",")
	} else {
		return nil, errors.New("Action requires list of comma separated usernames in 'users' parameter")
	}
	for _, v := range users {
		if uuid.Parse(v) != nil {
			ids = append(ids, v)
		} else {
			u := user.User{Username: v}
			if err := u.SetMongoInfo(); err != nil {
				return nil, err
			}
			ids = append(ids, u.Uuid)
		}
	}
	return ids, nil
}

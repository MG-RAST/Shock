//Package acl implements /node/:id/acl resource
package acl

import (
	"errors"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/MG-RAST/golib/go-uuid/uuid"
	"github.com/MG-RAST/golib/mgo"
	"github.com/MG-RAST/golib/stretchr/goweb/context"
	"net/http"
	"strings"
)

var (
	validAclTypes = map[string]bool{"all": true, "read": true, "write": true, "delete": true, "owner": true}
)

// GET, POST, PUT, DELETE: /node/{nid}/acl/
// GET is the only action implemented here.
func AclRequest(ctx context.Context) {
	nid := ctx.PathValue("nid")

	u, err := request.Authenticate(ctx.HttpRequest())
	if err != nil && err.Error() != e.NoAuth {
		request.AuthError(err, ctx)
		return
	}

	// acl require auth even for public data
	if u == nil {
		responder.RespondWithError(ctx, http.StatusUnauthorized, e.NoAuth)
		return
	}

	// Load node and handle user unauthorized
	n, err := node.LoadUnauth(nid)
	if err != nil {
		if err == mgo.ErrNotFound {
			responder.RespondWithError(ctx, http.StatusNotFound, "Node not found")
			return
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			err_msg := "Err@node_Read:LoadNode: " + nid + ":" + err.Error()
			logger.Error(err_msg)
			responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
			return
		}
	}

	// only the owner or an admin can view/edit acl's unless owner="" or owner=nil
	// note: owner can only be empty when anonymous node creation is enabled in shock config.
	if n.Acl.Owner != u.Uuid && n.Acl.Owner != "" && !u.Admin {
		err_msg := "Only the node owner can edit/view node ACL's"
		logger.Error(err_msg)
		responder.RespondWithError(ctx, http.StatusUnauthorized, err_msg)
		return
	}

	if ctx.HttpRequest().Method == "GET" {
		responder.RespondWithData(ctx, n.Acl)
	} else {
		responder.RespondWithError(ctx, http.StatusNotImplemented, "This request type is not implemented.")
	}
	return
}

// GET, POST, PUT, DELETE: /node/{nid}/acl/{type}
func AclTypedRequest(ctx context.Context) {
	nid := ctx.PathValue("nid")
	rtype := ctx.PathValue("type")

	u, err := request.Authenticate(ctx.HttpRequest())
	if err != nil && err.Error() != e.NoAuth {
		request.AuthError(err, ctx)
		return
	}

	if !validAclTypes[rtype] {
		responder.RespondWithError(ctx, http.StatusBadRequest, "Invalid acl type")
		return
	}

	// acl require auth even for public data
	if u == nil {
		responder.RespondWithError(ctx, http.StatusUnauthorized, e.NoAuth)
		return
	}

	// Load node and handle user unauthorized
	n, err := node.LoadUnauth(nid)
	if err != nil {
		if err == mgo.ErrNotFound {
			responder.RespondWithError(ctx, http.StatusNotFound, "Node not found")
			return
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			err_msg := "Err@node_Read:LoadNode: " + nid + ":" + err.Error()
			logger.Error(err_msg)
			responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
			return
		}
	}

	// only the owner or an admin can view/edit acl's unless owner="" or owner=nil
	// note: owner can only be empty when anonymous node creation is enabled in shock config.
	if n.Acl.Owner != u.Uuid && n.Acl.Owner != "" && !u.Admin {
		err_msg := "Only the node owner can edit/view node ACL's"
		logger.Error(err_msg)
		responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
		return
	}

	requestMethod := ctx.HttpRequest().Method
	if requestMethod != "GET" {
		ids, err := parseAclRequestTyped(ctx)
		if err != nil {
			responder.RespondWithError(ctx, http.StatusBadRequest, err.Error())
			return
		}
		if requestMethod == "POST" || requestMethod == "PUT" {
			if rtype == "owner" {
				if len(ids) == 1 {
					n.Acl.SetOwner(ids[0])
				} else {
					responder.RespondWithError(ctx, http.StatusBadRequest, "Too many users. Nodes may have only one owner.")
					return
				}
			} else if rtype == "all" {
				for _, atype := range []string{"read", "write", "delete"} {
					for _, i := range ids {
						n.Acl.Set(i, map[string]bool{atype: true})
					}
				}
			} else {
				for _, i := range ids {
					n.Acl.Set(i, map[string]bool{rtype: true})
				}
			}
			n.Save()
		} else if requestMethod == "DELETE" {
			if rtype == "owner" {
				responder.RespondWithError(ctx, http.StatusBadRequest, "Deleting ownership is not a supported request type.")
				return
			} else if rtype == "all" {
				for _, atype := range []string{"read", "write", "delete"} {
					for _, i := range ids {
						n.Acl.UnSet(i, map[string]bool{atype: true})
					}
				}
			} else {
				for _, i := range ids {
					n.Acl.UnSet(i, map[string]bool{rtype: true})
				}
			}
			n.Save()
		} else {
			responder.RespondWithError(ctx, http.StatusNotImplemented, "This request type is not implemented.")
			return
		}
	}

	switch rtype {
	default:
		responder.RespondWithError(ctx, http.StatusNotImplemented, "This request type is not implemented.")
	case "read":
		responder.RespondWithData(ctx, map[string][]string{"read": n.Acl.Read})
	case "write":
		responder.RespondWithData(ctx, map[string][]string{"write": n.Acl.Write})
	case "delete":
		responder.RespondWithData(ctx, map[string][]string{"delete": n.Acl.Delete})
	case "owner":
		responder.RespondWithData(ctx, map[string]string{"owner": n.Acl.Owner})
	case "all":
		responder.RespondWithData(ctx, n.Acl)
	}

	return
}

func parseAclRequestTyped(ctx context.Context) (ids []string, err error) {
	var users []string
	query := ctx.HttpRequest().URL.Query()
	params, _, err := request.ParseMultipartForm(ctx.HttpRequest())
	if _, ok := query["users"]; ok && err != nil && err.Error() == "request Content-Type isn't multipart/form-data" {
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

//Package acl implements /node/:id/acl resource
package acl

import (
	"errors"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/MG-RAST/Shock/shock-server/util"
	"github.com/MG-RAST/golib/goweb"
	"net/http"
	"strings"
)

var (
	validAclTypes = map[string]bool{"all": true, "read": true, "write": true, "delete": true, "owner": true}
)

// GET, POST, PUT, DELETE: /node/{nid}/acl/
// GET is the only action implemented here.
var Controller goweb.ControllerFunc = func(cx *goweb.Context) {
	request.Log(cx.Request)
	u, err := request.Authenticate(cx.Request)
	if err != nil && err.Error() != e.NoAuth {
		request.AuthError(err, cx)
		return
	}

	// acl require auth even for public data
	if u == nil {
		cx.RespondWithErrorMessage(e.NoAuth, http.StatusUnauthorized)
		return
	}

	// Load node and handle user unauthorized
	id := cx.PathParams["nid"]
	n, err := node.Load(id, u.Uuid)
	if err != nil {
		if err.Error() == e.UnAuth {
			cx.RespondWithErrorMessage(err.Error(), http.StatusUnauthorized)
			return
		} else if err.Error() == e.MongoDocNotFound {
			cx.RespondWithNotFound()
			return
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			err_msg := "Err@node_Read:LoadNode: " + err.Error()
			logger.Error(err_msg)
			cx.RespondWithErrorMessage(err_msg, http.StatusInternalServerError)
			return
		}
	}

	rights := n.Acl.Check(u.Uuid)
	if cx.Request.Method == "GET" {
		if u.Uuid == n.Acl.Owner || rights["read"] {
			cx.RespondWithData(n.Acl)
		} else {
			cx.RespondWithErrorMessage(e.UnAuth, http.StatusUnauthorized)
			return
		}
	} else {
		cx.RespondWithErrorMessage("This request type is not implemented.", http.StatusNotImplemented)
	}
	return
}

// GET, POST, PUT, DELETE: /node/{nid}/acl/{type}
var ControllerTyped goweb.ControllerFunc = func(cx *goweb.Context) {
	request.Log(cx.Request)
	u, err := request.Authenticate(cx.Request)
	if err != nil && err.Error() != e.NoAuth {
		request.AuthError(err, cx)
		return
	}

	// acl require auth even for public data
	if u == nil {
		cx.RespondWithErrorMessage(e.NoAuth, http.StatusUnauthorized)
		return
	}

	rtype := cx.PathParams["type"]
	if !validAclTypes[rtype] {
		cx.RespondWithErrorMessage("Invalid acl type", http.StatusBadRequest)
		return
	}

	// Load node and handle user unauthorized
	id := cx.PathParams["nid"]
	n, err := node.Load(id, u.Uuid)
	if err != nil {
		if err.Error() == e.UnAuth {
			cx.RespondWithErrorMessage(err.Error(), http.StatusUnauthorized)
			return
		} else if err.Error() == e.MongoDocNotFound {
			cx.RespondWithNotFound()
			return
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			err_msg := "Err@node_Read:LoadNode: " + err.Error()
			logger.Error(err_msg)
			cx.RespondWithErrorMessage(err_msg, http.StatusInternalServerError)
			return
		}
	}

	rights := n.Acl.Check(u.Uuid)
	if cx.Request.Method != "GET" {
		ids, err := parseAclRequestTyped(cx)
		if err != nil {
			cx.RespondWithErrorMessage(err.Error(), http.StatusBadRequest)
			return
		}
		if (cx.Request.Method == "POST" || cx.Request.Method == "PUT") && (u.Uuid == n.Acl.Owner || rights["write"]) {
			if rtype == "owner" {
				if u.Uuid == n.Acl.Owner {
					if len(ids) == 1 {
						n.Acl.SetOwner(ids[0])
					} else {
						cx.RespondWithErrorMessage("Too many users. Nodes may have only one owner.", http.StatusBadRequest)
						return
					}
				} else {
					cx.RespondWithErrorMessage("Only owner can change ownership of Node.", http.StatusBadRequest)
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
		} else if cx.Request.Method == "DELETE" && (u.Uuid == n.Acl.Owner || rights["delete"]) {
			if rtype == "owner" {
				cx.RespondWithErrorMessage("Deleting ownership is not a supported request type.", http.StatusBadRequest)
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
			cx.RespondWithErrorMessage(e.UnAuth, http.StatusUnauthorized)
			return
		}
	}

	if u.Uuid == n.Acl.Owner || rights["read"] {
		switch rtype {
		case "read":
			cx.RespondWithData(map[string][]string{"read": n.Acl.Read})
		case "write":
			cx.RespondWithData(map[string][]string{"write": n.Acl.Write})
		case "delete":
			cx.RespondWithData(map[string][]string{"delete": n.Acl.Delete})
		case "owner":
			cx.RespondWithData(map[string]string{"owner": n.Acl.Owner})
		case "all":
			cx.RespondWithData(n.Acl)
		}
	} else {
		cx.RespondWithErrorMessage(e.UnAuth, http.StatusUnauthorized)
		return
	}
	return
}

func parseAclRequestTyped(cx *goweb.Context) (ids []string, err error) {
	var users []string
	query := util.Q(cx.Request.URL.Query())
	params, _, err := request.ParseMultipartForm(cx.Request)
	if err != nil && err.Error() == "request Content-Type isn't multipart/form-data" && query.Has("users") {
		users = strings.Split(query.Value("users"), ",")
	} else if params["users"] != "" {
		users = strings.Split(params["users"], ",")
	} else {
		return nil, errors.New("Action requires list of comma separated usernames in 'users' parameter")
	}
	for _, v := range users {
		if isUuid(v) {
			ids = append(ids, v)
		} else {
			u := user.User{Username: v}
			if err := u.SetUuid(); err != nil {
				return nil, err
			}
			ids = append(ids, u.Uuid)
		}
	}
	return ids, nil
}

func isUuid(s string) bool {
	if strings.Count(s, "-") == 4 {
		return true
	}
	return false
}

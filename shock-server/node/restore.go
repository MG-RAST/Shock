package node

import (
	"fmt"
	"net/http"

	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/go-chi/chi/v5"
	mgo "gopkg.in/mgo.v2"
)

// GET, PUT, DELETE: /node/{nid}/restore/  , for PUT send body: {loc}, specify -H "Content-Type: application/json"
func RestoreRequest(w http.ResponseWriter, r *http.Request) {
	nid := chi.URLParam(r, "nid")
	value := chi.URLParam(r, "val")

	rmeth := r.Method

	u, err := request.Authenticate(r)
	if err != nil && err.Error() != e.NoAuth {
		request.AuthError(err, w, r)
		return
	}

	// public user (no auth) can be used in some cases
	if u == nil && conf.USE_AUTH {
		if (rmeth == "GET" && conf.ANON_READ) || (rmeth == "POST" && conf.ANON_WRITE) || (rmeth == "DELETE" && conf.ANON_WRITE) {
			u = &user.User{Uuid: "public"}
		} else {
			responder.RespondWithError(w, r, http.StatusUnauthorized, e.NoAuth)
			return
		}
	}

	// Load node by id
	n, err := Load(nid)
	if err != nil {
		if err == mgo.ErrNotFound {
			logger.Error("(LocationsRequest) (node.Load) id=" + nid + ": " + e.NodeNotFound)
			responder.RespondWithError(w, r, http.StatusNotFound, e.NodeNotFound)
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			errMsg := "(LocationsRequest) (node.Load) id=" + nid + ":" + err.Error()
			logger.Error(errMsg)
			responder.RespondWithError(w, r, http.StatusInternalServerError, errMsg)
		}
		return
	}

	if conf.USE_AUTH {
		rights := n.Acl.Check(u.Uuid)
		if n.Acl.Owner != u.Uuid && u.Admin == false && n.Acl.Owner != "public" && rights["read"] == false {
			logger.Error("err@node_Acl: (Authenticate) id=" + nid + ": " + e.UnAuth)
			responder.RespondWithError(w, r, http.StatusUnauthorized, e.UnAuth)
			return
		}
	}
	switch rmeth {

	//case "PUT": //PUT is idempotent
	case "GET":
		if value == "" { // /node/{nid}/restore/
			restore := n.GetRestore()
			responder.RespondWithData(w, r, restore)
			return
		}
		// we might have to handle an error here

	case "POST": // append
		if conf.USE_AUTH && !u.Admin {
			errmsg := e.UnAuth
			if conf.DEBUG_AUTH {
				errmsg = "admin required"
			}

			responder.RespondWithError(w, r, http.StatusUnauthorized, errmsg) //
			return
		}

		if value == "true" {
			n.SetRestore()
		} else { // FALSE case
			n.UnSetRestore()
		}

		n.Save()

		restore := n.GetRestore()
		responder.RespondWithData(w, r, restore)

		// of case
	default:
		errMsg := fmt.Sprintf("(LocationsRequest) %s not supported", rmeth)
		responder.RespondWithError(w, r, http.StatusInternalServerError, errMsg)

	}

	return
}

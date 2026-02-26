package node

import (
	"net/http"

	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/go-chi/chi/v5"
	mgo "gopkg.in/mgo.v2"
)

// DELETE: /node/{id}
func (cr *NodeController) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	u, err := request.Authenticate(r)
	if err != nil && err.Error() != e.NoAuth {
		request.AuthError(err, w, r)
		return
	}

	// public user (no auth) can be used in some cases
	if u == nil {
		if conf.ANON_DELETE {
			u = &user.User{Uuid: "public"}
		} else {
			responder.RespondWithError(w, r, http.StatusUnauthorized, e.NoAuth)
			return
		}
	}

	// Load node by id
	n, err := node.Load(id)
	if err != nil {
		if err == mgo.ErrNotFound {
			logger.Error("err@node_Delete: (node.Load) id=" + id + ": " + e.NodeNotFound)
			responder.RespondWithError(w, r, http.StatusNotFound, e.NodeNotFound)
			return
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			err_msg := "err@node_Delete: (node.Load) id=" + id + ": " + err.Error()
			logger.Error(err_msg)
			responder.RespondWithError(w, r, http.StatusInternalServerError, err_msg)
			return
		}
	}

	rights := n.Acl.Check(u.Uuid)
	prights := n.Acl.Check("public")
	if rights["delete"] == false && u.Admin == false && n.Acl.Owner != u.Uuid && prights["delete"] == false {
		logger.Error("err@node_Delete: (Authenticate) id=" + id + ": " + e.UnAuth)
		responder.RespondWithError(w, r, http.StatusUnauthorized, e.UnAuth)
		return
	}

	if _, err := n.Delete(); err == nil {
		responder.RespondOK(w, r)
	} else {
		err_msg := "err@node_Delete: (node.Delete) " + err.Error()
		logger.Error(err_msg)
		responder.RespondWithError(w, r, http.StatusInternalServerError, err_msg)
	}
}

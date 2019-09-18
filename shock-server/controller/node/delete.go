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
	"github.com/MG-RAST/golib/stretchr/goweb/context"
	mgo "gopkg.in/mgo.v2"
)

// DELETE: /node/{id}
func (cr *NodeController) Delete(id string, ctx context.Context) error {
	u, err := request.Authenticate(ctx.HttpRequest())
	if err != nil && err.Error() != e.NoAuth {
		return request.AuthError(err, ctx)
	}

	// public user (no auth) can be used in some cases
	if u == nil {
		if conf.ANON_DELETE {
			u = &user.User{Uuid: "public"}
		} else {
			return responder.RespondWithError(ctx, http.StatusUnauthorized, e.NoAuth)
		}
	}

	// Load node by id
	n, err := node.Load(id)
	if err != nil {
		if err == mgo.ErrNotFound {
			logger.Error("err@node_Delete: (node.Load) id=" + id + ": " + e.NodeNotFound)
			return responder.RespondWithError(ctx, http.StatusNotFound, e.NodeNotFound)
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			err_msg := "err@node_Delete: (node.Load) id=" + id + ": " + err.Error()
			logger.Error(err_msg)
			return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
		}
	}

	rights := n.Acl.Check(u.Uuid)
	prights := n.Acl.Check("public")
	if rights["delete"] == false && u.Admin == false && n.Acl.Owner != u.Uuid && prights["delete"] == false {
		logger.Error("err@node_Delete: (Authenticate) id=" + id + ": " + e.UnAuth)
		return responder.RespondWithError(ctx, http.StatusUnauthorized, e.UnAuth)
	}

	if _, err := n.Delete(); err == nil {
		return responder.RespondOK(ctx)
	} else {
		err_msg := "err@node_Delete: (node.Delete) " + err.Error()
		logger.Error(err_msg)
		return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
	}
}

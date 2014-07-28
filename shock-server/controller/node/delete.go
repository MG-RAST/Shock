package node

import (
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/MG-RAST/golib/stretchr/goweb/context"
	"net/http"
)

// DELETE: /node/{id}
func (cr *NodeController) Delete(id string, ctx context.Context) error {
	u, err := request.Authenticate(ctx.HttpRequest())
	if err != nil && err.Error() != e.NoAuth {
		return request.AuthError(err, ctx)
	}

	if u == nil {
		return responder.RespondWithError(ctx, http.StatusUnauthorized, e.NoAuth)
	}

	// Load node and handle user unauthorized
	n, err := node.Load(id, u)
	if err != nil {
		if err.Error() == e.UnAuth {
			return responder.RespondWithError(ctx, http.StatusUnauthorized, e.UnAuth)
		} else if err.Error() == e.MongoDocNotFound {
			return responder.RespondWithError(ctx, http.StatusNotFound, "Node not found")
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			err_msg := "Err@node_Read:Delete: " + err.Error()
			logger.Error(err_msg)
			return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
		}
	}

	rights := n.Acl.Check(u.Uuid)
	if !rights["delete"] {
		return responder.RespondWithError(ctx, http.StatusUnauthorized, e.UnAuth)
	}

	if err := n.Delete(); err == nil {
		return responder.RespondOK(ctx)
	} else {
		err_msg := "Err@node_Delete:Delete: " + err.Error()
		logger.Error(err_msg)
		return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
	}
}

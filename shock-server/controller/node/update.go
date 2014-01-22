package node

import (
	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/stretchr/goweb/context"
	"net/http"
)

// PUT: /node/{id} -> multipart-form
func (cr *NodeController) Replace(id string, ctx context.Context) error {
	u, err := request.Authenticate(ctx.HttpRequest())
	if err != nil && err.Error() != e.NoAuth {
		return request.AuthError(err, ctx)
	}

	// Fake public user
	if u == nil {
		u = &user.User{Uuid: ""}
	}

	n, err := node.Load(id, u.Uuid)
	if err != nil {
		if err.Error() == e.UnAuth {
			return responder.RespondWithError(ctx, http.StatusUnauthorized, e.UnAuth)
		} else if err.Error() == e.MongoDocNotFound {
			return responder.RespondWithError(ctx, http.StatusNotFound, "Node not found")
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			err_msg := "Err@node_Update:LoadNode: " + err.Error()
			logger.Error(err_msg)
			return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
		}
	}

	if conf.Bool(conf.Conf["perf-log"]) {
		logger.Perf("START PUT data: " + id)
	}
	params, files, err := request.ParseMultipartForm(ctx.HttpRequest())
	if err != nil {
		err_msg := "err@node_ParseMultipartForm: " + err.Error()
		logger.Error(err_msg)
		return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
	}

	err = n.Update(params, files)
	if err != nil {
		err_msg := "err@node_Update: " + id + ": " + err.Error()
		logger.Error(err_msg)
		return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
	}
	responder.RespondWithData(ctx, n)
	if conf.Bool(conf.Conf["perf-log"]) {
		logger.Perf("END PUT data: " + id)
	}
	return nil
}

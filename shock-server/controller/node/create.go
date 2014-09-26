package node

import (
	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/MG-RAST/golib/stretchr/goweb/context"
	"net/http"
)

// POST: /node
func (cr *NodeController) Create(ctx context.Context) error {
	u, err := request.Authenticate(ctx.HttpRequest())
	if err != nil && err.Error() != e.NoAuth {
		return request.AuthError(err, ctx)
	}

	// public user
	if u == nil {
		if conf.Bool(conf.Conf["anon-write"]) {
			u = &user.User{Uuid: "public"}
		} else {
			return responder.RespondWithError(ctx, http.StatusUnauthorized, e.NoAuth)
		}
	}

	// Parse uploaded form
	params, files, err := request.ParseMultipartForm(ctx.HttpRequest())
	if err != nil {
		if err.Error() == "request Content-Type isn't multipart/form-data" {
			// If not multipart/form-data it will try to read the Body of the
			// request. If the Body is not empty it will create a file from
			// the Body contents. If the Body is empty it will create an empty
			// node.
			if ctx.HttpRequest().ContentLength != 0 {
				params, files, err = request.DataUpload(ctx.HttpRequest())
				if err != nil {
					err_msg := "Error uploading data from request body:" + err.Error()
					logger.Error(err_msg)
					return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
				}
			}

			n, cn_err := node.CreateNodeUpload(u, params, files)

			if cn_err != nil {
				err_msg := "Error at create empty node: " + cn_err.Error()
				logger.Error(err_msg)
				return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
			}
			if n == nil {
				// Not sure how you could get an empty node with no error
				// Assume it's the user's fault
				return responder.RespondWithError(ctx, http.StatusBadRequest, "Error, could not create node.")
			} else {
				return responder.RespondWithData(ctx, n)
			}
		} else {
			// Some error other than request encoding. Theoretically
			// could be a lost db connection between user lookup and parsing.
			// Blame the user, Its probaby their fault anyway.
			err_msg := "Error parsing form: " + err.Error()
			logger.Error(err_msg)
			return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
		}
	}

	// Create node
	n, err := node.CreateNodeUpload(u, params, files)
	if err != nil {
		err_msg := "err@node_CreateNodeUpload: " + err.Error()
		logger.Error(err_msg)
		return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
	}
	return responder.RespondWithData(ctx, n)
}

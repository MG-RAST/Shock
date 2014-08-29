package node

import (
	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/MG-RAST/golib/mgo"
	"github.com/MG-RAST/golib/stretchr/goweb/context"
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

	n, err := node.Load(id, u)
	if err != nil {
		if err.Error() == e.UnAuth {
			return responder.RespondWithError(ctx, http.StatusUnauthorized, e.UnAuth)
		} else if err == mgo.ErrNotFound {
			return responder.RespondWithError(ctx, http.StatusNotFound, "Node not found")
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			err_msg := "Err@node_Update:LoadNode: " + id + ":" + err.Error()
			logger.Error(err_msg)
			return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
		}
	}

	rights := n.Acl.Check(u.Uuid)
	if !rights["write"] {
		return responder.RespondWithError(ctx, http.StatusUnauthorized, e.UnAuth)
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

	if _, hasCopyData := params["copy_data"]; hasCopyData {
		_, err = node.Load(params["copy_data"], u)
		if err != nil {
			return request.AuthError(err, ctx)
		}
	}

	if _, hasParentNode := params["parent_node"]; hasParentNode {
		_, err = node.Load(params["parent_node"], u)
		if err != nil {
			return request.AuthError(err, ctx)
		}
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

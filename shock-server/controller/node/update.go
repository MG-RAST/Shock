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
	mgo "gopkg.in/mgo.v2"
	"net/http"
)

// PUT: /node/{id} -> multipart-form
func (cr *NodeController) Replace(id string, ctx context.Context) error {
	u, err := request.Authenticate(ctx.HttpRequest())
	if err != nil && err.Error() != e.NoAuth {
		return request.AuthError(err, ctx)
	}

	// public user (no auth) can be used in some cases
	if u == nil {
		if conf.ANON_WRITE {
			u = &user.User{Uuid: "public"}
		} else {
			return responder.RespondWithError(ctx, http.StatusUnauthorized, e.NoAuth)
		}
	}

	// Load node by id
	n, err := node.Load(id)
	if err != nil {
		if err == mgo.ErrNotFound {
			return responder.RespondWithError(ctx, http.StatusNotFound, e.NodeNotFound)
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			err_msg := "Err@node_Update:LoadNode: " + id + ":" + err.Error()
			logger.Error(err_msg)
			return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
		}
	}

	rights := n.Acl.Check(u.Uuid)
	prights := n.Acl.Check("public")
	if rights["write"] == false && u.Admin == false && n.Acl.Owner != u.Uuid && prights["write"] == false {
		return responder.RespondWithError(ctx, http.StatusUnauthorized, e.UnAuth)
	}

	if conf.LOG_PERF {
		logger.Perf("START PUT data: " + id)
	}
	params, files, err := request.ParseMultipartForm(ctx.HttpRequest())
	// clean up temp dir !!
	defer node.RemoveAllFormFiles(files)
	if err != nil {
		err_msg := "err@node_ParseMultipartForm: " + err.Error()
		logger.Error(err_msg)
		return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
	}

	// need delete rights to set expiration
	if _, hasExpiration := params["expiration"]; hasExpiration {
		if rights["delete"] == false && u.Admin == false && n.Acl.Owner != u.Uuid && prights["delete"] == false {
			return responder.RespondWithError(ctx, http.StatusUnauthorized, e.UnAuth)
		}
	}

	if _, hasCopyData := params["copy_data"]; hasCopyData {
		_, err = node.Load(params["copy_data"])
		if err != nil {
			return request.AuthError(err, ctx)
		}
	}

	if _, hasParentNode := params["parent_node"]; hasParentNode {
		_, err = node.Load(params["parent_node"])
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
	if conf.LOG_PERF {
		logger.Perf("END PUT data: " + id)
	}
	return nil
}

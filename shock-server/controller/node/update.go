package node

import (
	"net/http"

	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/MG-RAST/Shock/shock-server/node/locker"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/MG-RAST/golib/stretchr/goweb/context"
	mgo "gopkg.in/mgo.v2"
)

// PUT: /node/{id} -> multipart-form
func (cr *NodeController) Replace(id string, ctx context.Context) (err error) {
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

	// lock before loading
	err = locker.NodeLockMgr.LockNode(id)
	if err != nil {
		err_msg := "err@node_Update: (LockMgr.LockNode) id=" + id + ": " + err.Error()
		logger.Error(err_msg)
		return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
	}
	defer locker.NodeLockMgr.UnlockNode(id)

	// Load node by id
	n, err := node.Load(id)
	if err != nil {
		if err == mgo.ErrNotFound {
			logger.Error("err@node_Update: (node.Load) id=" + id + ": " + e.NodeNotFound)
			return responder.RespondWithError(ctx, http.StatusNotFound, e.NodeNotFound)
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			err_msg := "err@node_Update: (node.Load) " + id + ": " + err.Error()
			logger.Error(err_msg)
			return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
		}
	}

	rights := n.Acl.Check(u.Uuid)
	prights := n.Acl.Check("public")
	if rights["write"] == false && u.Admin == false && n.Acl.Owner != u.Uuid && prights["write"] == false {
		logger.Error("err@node_Update: (Authenticate) id=" + id + ": " + e.UnAuth)
		return responder.RespondWithError(ctx, http.StatusUnauthorized, e.UnAuth)
	}

	if conf.LOG_PERF {
		logger.Perf("START PUT data: " + id)
	}
	params, files, err := request.ParseMultipartForm(ctx.HttpRequest())
	// clean up temp dir !!
	defer file.RemoveAllFormFiles(files)
	if err != nil {
		err_msg := "err@node_Update: (ParseMultipartForm) id=" + id + ": " + err.Error()
		logger.Error(err_msg)
		return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
	}

	// need delete rights to set expiration
	if _, hasExpiration := params["expiration"]; hasExpiration {
		if rights["delete"] == false && u.Admin == false && n.Acl.Owner != u.Uuid && prights["delete"] == false {
			logger.Error("err@node_Update: (Authenticate) id=" + id + ": " + e.UnAuth)
			return responder.RespondWithError(ctx, http.StatusUnauthorized, e.UnAuth)
		}
	}

	if copy_data_id, hasCopyData := params["copy_data"]; hasCopyData {
		var copy_data_node *node.Node
		copy_data_node, err = node.Load(copy_data_id)
		if err != nil {
			return
		}

		rights := copy_data_node.Acl.Check(u.Uuid)
		if copy_data_node.Acl.Owner != u.Uuid && u.Admin == false && copy_data_node.Acl.Owner != "public" && rights["read"] == false {
			logger.Error("err@node_Update: (Authenticate) id=" + copy_data_id + ": " + e.UnAuth)
			responder.RespondWithError(ctx, http.StatusUnauthorized, e.UnAuth)
			err = request.AuthError(err, ctx)
			return
		}
	}

	if parentNode_id, hasParentNode := params["parent_node"]; hasParentNode {
		var parentNode *node.Node
		parentNode, err = node.Load(parentNode_id)
		if err != nil {
			return
		}

		rights := parentNode.Acl.Check(u.Uuid)
		if parentNode.Acl.Owner != u.Uuid && u.Admin == false && parentNode.Acl.Owner != "public" && rights["read"] == false {
			logger.Error("err@node_Update: (Authenticate) id=" + parentNode_id + ": " + e.UnAuth)
			responder.RespondWithError(ctx, http.StatusUnauthorized, e.UnAuth)
			err = request.AuthError(err, ctx)
			return
		}
	}

	err = n.Update(params, files, true)
	if err != nil {
		err_msg := "err@node_Update: (node.Update) id=" + id + ": " + err.Error()
		logger.Error(err_msg)
		return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
	}
	responder.RespondWithData(ctx, n)
	if conf.LOG_PERF {
		logger.Perf("END PUT data: " + id)
	}
	return nil
}

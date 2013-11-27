package node

import (
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/golib/goweb"
	"net/http"
)

// DELETE: /node/{id}
func (cr *Controller) Delete(id string, cx *goweb.Context) {
	request.Log(cx.Request)
	u, err := request.Authenticate(cx.Request)
	if err != nil && err.Error() != e.NoAuth {
		request.AuthError(err, cx)
		return
	}
	if u == nil {
		cx.RespondWithErrorMessage(e.NoAuth, http.StatusUnauthorized)
		return
	}

	// Load node and handle user unauthorized
	n, err := node.Load(id, u.Uuid)
	if err != nil {
		if err.Error() == e.UnAuth {
			cx.RespondWithErrorMessage(e.NoAuth, http.StatusUnauthorized)
			return
		} else if err.Error() == e.MongoDocNotFound {
			cx.RespondWithNotFound()
			return
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			err_msg := "Err@node_Read:Delete: " + err.Error()
			logger.Error(err_msg)
			cx.RespondWithErrorMessage(err_msg, http.StatusInternalServerError)
			return
		}
	}

	if err := n.Delete(); err == nil {
		cx.RespondWithOK()
		return
	} else {
		err_msg := "Err@node_Delet:Delete: " + err.Error()
		logger.Error(err_msg)
		cx.RespondWithErrorMessage(err_msg, http.StatusInternalServerError)
	}
	return
}

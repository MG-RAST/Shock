package node

import (
	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/MG-RAST/golib/goweb"
	"net/http"
)

// POST: /node
func (cr *Controller) Create(cx *goweb.Context) {
	// Log Request and check for Auth
	request.Log(cx.Request)
	u, err := request.Authenticate(cx.Request)
	if err != nil && err.Error() != e.NoAuth {
		request.AuthError(err, cx)
		return
	}

	// Fake public user
	if u == nil {
		if conf.Bool(conf.Conf["anon-write"]) {
			u = &user.User{Uuid: ""}
		} else {
			cx.RespondWithErrorMessage(e.NoAuth, http.StatusUnauthorized)
			return
		}
	}

	// Parse uploaded form
	params, files, err := request.ParseMultipartForm(cx.Request)
	if err != nil {
		if err.Error() == "request Content-Type isn't multipart/form-data" {
			// If not multipart/form-data it will try to read the Body of the
			// request. If the Body is not empty it will create a file from
			// the Body contents. If the Body is empty it will create an empty
			// node.
			if cx.Request.ContentLength != 0 {
				params, files, err = request.DataUpload(cx.Request)
				if err != nil {
					logger.Error("Error uploading data from request body.\n")
					cx.RespondWithError(http.StatusInternalServerError)
					return
				}
			}

			n, cn_err := node.CreateNodeUpload(u, params, files)
			if cn_err != nil {
				if cx.Request.ContentLength != 0 {
					logger.Error("Error creating node from data upload: " + cn_err.Error())
					cx.RespondWithError(http.StatusInternalServerError)
					return
				} else {
					logger.Error("Error at create empty node: " + cn_err.Error())
					cx.RespondWithError(http.StatusInternalServerError)
					return
				}

			}
			if n == nil {
				// Not sure how you could get an empty node with no error
				// Assume it's the user's fault
				cx.RespondWithError(http.StatusBadRequest)
				return
			} else {
				cx.RespondWithData(n)
				return
			}
		} else {
			// Some error other than request encoding. Theoretically
			// could be a lost db connection between user lookup and parsing.
			// Blame the user, Its probaby their fault anyway.
			logger.Error("Error parsing form: " + err.Error())
			cx.RespondWithError(http.StatusBadRequest)
			return
		}
	}

	// Create node
	n, err := node.CreateNodeUpload(u, params, files)
	if err != nil {
		logger.Error("err@node_CreateNodeUpload: " + err.Error())
		cx.RespondWithErrorMessage(err.Error(), http.StatusBadRequest)
		return
	}
	cx.RespondWithData(n)
	return
}

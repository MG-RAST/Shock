// Package node implements /node resource
package node

import (
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/golib/goweb"
	"net/http"
)

type Controller struct{}

// Options: /node
func (cr *Controller) Options(cx *goweb.Context) {
	request.Log(cx.Request)
	cx.RespondWithOK()
	return
}

// Will not implement
// PUT: /node
func (cr *Controller) UpdateMany(cx *goweb.Context) {
	request.Log(cx.Request)
	cx.RespondWithErrorMessage("This request type is not implemented.", http.StatusNotImplemented)
}

// DELETE: /node
func (cr *Controller) DeleteMany(cx *goweb.Context) {
	request.Log(cx.Request)
	cx.RespondWithErrorMessage("This request type is not implemented.", http.StatusNotImplemented)
}

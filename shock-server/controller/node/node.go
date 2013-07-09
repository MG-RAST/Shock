package node

import (
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/jaredwilkening/goweb"
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
	cx.RespondWithError(http.StatusNotImplemented)
}

// DELETE: /node
func (cr *Controller) DeleteMany(cx *goweb.Context) {
	request.Log(cx.Request)
	cx.RespondWithError(http.StatusNotImplemented)
}

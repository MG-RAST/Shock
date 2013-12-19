// Package node implements /node resource
package node

import (
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/stretchr/goweb/context"
	"net/http"
)

type NodeController struct{}

// Options: /node
func (cr *NodeController) Options(ctx context.Context) {
	responder.RespondOK(ctx)
	return
}

// Will not implement
// PUT: /node
func (cr *NodeController) UpdateMany(ctx context.Context) {
	responder.RespondWithError(ctx, http.StatusNotImplemented, "This request type is not implemented.")
	return
}

// DELETE: /node
func (cr *NodeController) DeleteMany(ctx context.Context) {
	responder.RespondWithError(ctx, http.StatusNotImplemented, "This request type is not implemented.")
	return
}

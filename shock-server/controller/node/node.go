// Package node implements /node resource
package node

import (
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/MG-RAST/golib/stretchr/goweb/context"
	"net/http"
)

type NodeController struct{}

// Options: /node
func (cr *NodeController) Options(ctx context.Context) error {
	return responder.RespondOK(ctx)
}

// Will not implement
// PUT: /node
func (cr *NodeController) UpdateMany(ctx context.Context) error {
	return responder.RespondWithError(ctx, http.StatusNotImplemented, "This request type is not implemented.")
}

// DELETE: /node
func (cr *NodeController) DeleteMany(ctx context.Context) error {
	return responder.RespondWithError(ctx, http.StatusNotImplemented, "This request type is not implemented.")
}

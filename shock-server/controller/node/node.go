// Package node implements /node resource
package node

import (
	"github.com/MG-RAST/Shock/shock-server/responder"
	"net/http"
)

type NodeController struct{}

// Options: /node
func (cr *NodeController) Options(w http.ResponseWriter, r *http.Request) {
	responder.RespondOK(w, r)
}

// Will not implement
// PUT: /node
func (cr *NodeController) UpdateMany(w http.ResponseWriter, r *http.Request) {
	responder.RespondWithError(w, r, http.StatusNotImplemented, "This request type is not implemented.")
}

// DELETE: /node
func (cr *NodeController) DeleteMany(w http.ResponseWriter, r *http.Request) {
	responder.RespondWithError(w, r, http.StatusNotImplemented, "This request type is not implemented.")
}

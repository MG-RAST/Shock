package controller

import (
	"github.com/MG-RAST/Shock/shock-server/controller/node"
	"github.com/MG-RAST/Shock/shock-server/controller/node/acl"
	"github.com/MG-RAST/Shock/shock-server/controller/node/index"
	"github.com/MG-RAST/Shock/shock-server/controller/preauth"
	"github.com/jaredwilkening/goweb"
)

type Controller struct {
	Node    *node.Controller
	Index   goweb.ControllerFunc
	Acl     map[string]goweb.ControllerFunc
	Preauth func(*goweb.Context)
}

func New() *Controller {
	return &Controller{
		Node:    new(node.Controller),
		Index:   index.Controller,
		Acl:     map[string]goweb.ControllerFunc{"base": acl.Controller, "typed": acl.ControllerTyped},
		Preauth: preauth.PreAuthRequest,
	}
}

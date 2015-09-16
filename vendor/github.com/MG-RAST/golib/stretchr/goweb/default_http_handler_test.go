package goweb

import (
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/goweb/handlers"
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/testify/assert"
	"testing"
)

func TestSetDefaultHttpHandler(t *testing.T) {

	handler := new(handlers.HttpHandler)

	if assert.NotEqual(t, handler, defaultHttpHandler) {

		SetDefaultHttpHandler(handler)

		assert.Equal(t, handler, defaultHttpHandler)

	}

}

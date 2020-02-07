package types

import (
	"fmt"
	"net/http"

	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/MG-RAST/golib/stretchr/goweb/context"
)

// when a node is uploaded and has a supported Type, set the priority automatically

// GET, /types/{type}/{function} specify -H "Content-Type: application/json"
func TypeRequest(ctx context.Context) {

	typeID := ctx.PathValue("type")
	function := ctx.PathValue("function")

	fmt.Printf("v received typeID: %s\n", typeID)
	logger.Debug(2, "(TypeRequest) received typeID: %s", typeID)

	rmeth := ctx.HttpRequest().Method

	u, err := request.Authenticate(ctx.HttpRequest())
	if err != nil && err.Error() != e.NoAuth {
		request.AuthError(err, ctx)
		return
	}

	// public user cannot use this
	if (u == nil) && conf.USE_AUTH {
		errMsg := "admin required"
		//errMsg := e.UnAuth
		responder.RespondWithError(ctx, http.StatusUnauthorized, errMsg)
		return
	}

	if (u != nil) && (!u.Admin) && conf.USE_AUTH {
		errMsg := e.UnAuth
		logger.Debug(2, "(TypeRequest) attempt to use as non admin (user: %s)", u.Username)
		responder.RespondWithError(ctx, http.StatusInternalServerError, errMsg)
		return
	}

	if rmeth != "GET" {
		errMsg := fmt.Sprintf("(TypeRequest) %s not supported", rmeth)
		responder.RespondWithError(ctx, http.StatusInternalServerError, errMsg)
		return
	}

	//fmt.Printf("TypeRequest passed auth bits and rmeth \n")

	// print details for one typeID or list all types

	typeEntry, ok := conf.TypesMap[typeID]
	if !ok {
		list := ""
		for x, _ := range conf.TypesMap {
			list += x + ","
		}
		err = fmt.Errorf("(TypeRequest) type %s not found (found: %s)", typeID, list)
		responder.RespondWithError(ctx, http.StatusInternalServerError, err.Error())
		//	fmt.Printf("TypeRequest LOAD error \n")
		return
	}

	//fmt.Printf("conf.TypesMap[typeID] worked \n")

	// ensure we only list nodes with Priority higher or equal to the one defined for the location

	switch function {

	case "info":

		// spew.Dump(locConf)
		responder.RespondWithData(ctx, typeEntry)
		return

	default:

		errMsg := fmt.Sprintf("(TypeRequest) %s not supported", function)
		responder.RespondWithError(ctx, http.StatusInternalServerError, errMsg)

	}

	return
}

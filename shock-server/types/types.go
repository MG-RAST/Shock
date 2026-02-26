package types

import (
	"fmt"
	"net/http"

	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/go-chi/chi/v5"
)

// when a node is uploaded and has a supported Type, set the priority automatically

// GET, /types/{type}/{function} specify -H "Content-Type: application/json"
func TypeRequest(w http.ResponseWriter, r *http.Request) {

	typeID := chi.URLParam(r, "type")
	function := chi.URLParam(r, "function")

	fmt.Printf("v received typeID: %s\n", typeID)
	logger.Debug(2, "(TypeRequest) received typeID: %s", typeID)

	rmeth := r.Method

	u, err := request.Authenticate(r)
	if err != nil && err.Error() != e.NoAuth {
		request.AuthError(err, w, r)
		return
	}

	// public user cannot use this
	if (u == nil) && conf.USE_AUTH {
		errMsg := "admin required"
		responder.RespondWithError(w, r, http.StatusUnauthorized, errMsg)
		return
	}

	if (u != nil) && (!u.Admin) && conf.USE_AUTH {
		errMsg := e.UnAuth
		logger.Debug(2, "(TypeRequest) attempt to use as non admin (user: %s)", u.Username)
		responder.RespondWithError(w, r, http.StatusInternalServerError, errMsg)
		return
	}

	if rmeth != "GET" {
		errMsg := fmt.Sprintf("(TypeRequest) %s not supported", rmeth)
		responder.RespondWithError(w, r, http.StatusInternalServerError, errMsg)
		return
	}

	// print details for one typeID or list all types

	typeEntry, ok := conf.TypesMap[typeID]
	if !ok {
		list := ""
		for x, _ := range conf.TypesMap {
			list += x + ","
		}
		err = fmt.Errorf("(TypeRequest) type %s not found (found: %s)", typeID, list)
		responder.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// ensure we only list nodes with Priority higher or equal to the one defined for the location

	switch function {

	case "info":

		responder.RespondWithData(w, r, typeEntry)
		return

	default:

		errMsg := fmt.Sprintf("(TypeRequest) %s not supported", function)
		responder.RespondWithError(w, r, http.StatusInternalServerError, errMsg)

	}

	return
}

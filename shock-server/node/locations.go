package node

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/go-chi/chi/v5"
	mgo "gopkg.in/mgo.v2"
)

// GET, PUT, DELETE: /node/{nid}/locations/{loc}  , for PUT send body: {loc}, specify -H "Content-Type: application/json"
func LocationsRequest(w http.ResponseWriter, r *http.Request) {
	nid := chi.URLParam(r, "nid")
	locationID := chi.URLParam(r, "loc")

	rmeth := r.Method

	u, err := request.Authenticate(r)
	if err != nil && err.Error() != e.NoAuth {
		request.AuthError(err, w, r)
		return
	}

	// public user (no auth) can be used in some cases
	if u == nil && conf.USE_AUTH {
		if (rmeth == "GET" && conf.ANON_READ) || (rmeth == "POST" && conf.ANON_WRITE) || (rmeth == "DELETE" && conf.ANON_WRITE) {
			u = &user.User{Uuid: "public"}
		} else {
			responder.RespondWithError(w, r, http.StatusUnauthorized, e.NoAuth)
			return
		}
	}

	// Load node by id
	n, err := Load(nid)
	if err != nil {
		if err == mgo.ErrNotFound {
			logger.Error("(LocationsRequest) (node.Load) id=" + nid + ": " + e.NodeNotFound)
			responder.RespondWithError(w, r, http.StatusNotFound, e.NodeNotFound)
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			errMsg := "(LocationsRequest) (node.Load) id=" + nid + ":" + err.Error()
			logger.Error(errMsg)
			responder.RespondWithError(w, r, http.StatusInternalServerError, errMsg)
		}
		return
	}

	if conf.USE_AUTH {
		rights := n.Acl.Check(u.Uuid)
		if n.Acl.Owner != u.Uuid && u.Admin == false && n.Acl.Owner != "public" && rights["read"] == false {
			logger.Error("err@node_Acl: (Authenticate) id=" + nid + ": " + e.UnAuth)
			responder.RespondWithError(w, r, http.StatusUnauthorized, e.UnAuth)
			return
		}
	}
	switch rmeth {

	//case "PUT": //PUT is idempotent
	case "GET":
		if locationID == "" { // /node/{nid}/locations
			locations := n.GetLocations()
			responder.RespondWithData(w, r, locations)
			return
		} else { // /node/{nid}/locations/{loc}
			var thisLocation Location
			thisLocation, err = n.GetLocation(locationID)
			if err != nil {
				responder.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
				return
			}
			responder.RespondWithData(w, r, thisLocation)
			return

		}
	case "POST": // append
		if conf.USE_AUTH && !u.Admin {
			errmsg := e.UnAuth
			if conf.DEBUG_AUTH {
				errmsg = "admin required"
			}

			responder.RespondWithError(w, r, http.StatusUnauthorized, errmsg) //
			return
		}

		var locationObjectMap map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&locationObjectMap); err != nil {
			responder.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
			return
		}

		newLocation := Location{}

		locIDIf, ok := locationObjectMap["id"]
		if !ok {
			responder.RespondWithError(w, r, http.StatusInternalServerError, "id not found")
			return
		}

		if locIDIf == "" {
			responder.RespondWithError(w, r, http.StatusInternalServerError, "id empty")
			return
		}

		locID, ok := locIDIf.(string)
		if !ok {
			responder.RespondWithError(w, r, http.StatusInternalServerError, "id not a string")
			return
		}

		newLocation.ID = locID
		storedIf, ok := locationObjectMap["stored"]
		if ok {

			stored, ok := storedIf.(bool)
			if ok {
				newLocation.Stored = stored
			}
		}

		requestedDateIf, ok := locationObjectMap["requestedDate"]
		if ok {
			requestedDateStr, ok := requestedDateIf.(string)
			if !ok {
				responder.RespondWithError(w, r, http.StatusInternalServerError, "timestamp not a string")
				return
			}
			var requestedDate time.Time
			requestedDate, err = time.Parse(shortDateForm, requestedDateStr)
			if err != nil {

				requestedDate, err = time.Parse(longDateForm, requestedDateStr)
				if err != nil {
					responder.RespondWithError(w, r, http.StatusInternalServerError, "could not parse timestamp")
					return
				}
			}

			newLocation.RequestedDate = &requestedDate
		}

		err = n.AddLocation(newLocation)
		if err != nil {
			responder.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
			return
		}
		n.Save()
		responder.RespondWithData(w, r, n.Locations)

		return
	case "DELETE":
		if !u.Admin {
			responder.RespondWithError(w, r, http.StatusUnauthorized, e.UnAuth)
			return
		}
		if locationID != "" { // /node/{nid}/locations/{loc}
			err = n.DeleteLocation(locationID)
			if err != nil {
				responder.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
				return
			}
			n.Save()
			responder.RespondWithData(w, r, n.Locations)
		} else { // /node/{nid}/locations

			err = n.DeleteLocations()
			if err != nil {
				responder.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
				return
			}
			n.Save()
			responder.RespondOK(w, r)

		}
	default:
		errMsg := fmt.Sprintf("(LocationsRequest) %s not supported", rmeth)
		responder.RespondWithError(w, r, http.StatusInternalServerError, errMsg)

	}

	return
}

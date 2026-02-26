package location

import (
	"fmt"
	"net/http"

	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/go-chi/chi/v5"
	"gopkg.in/mgo.v2/bson"
)

// Load location
func Load(locID string) (loc *conf.LocationConfig, err error) {
	loc, ok := conf.LocationsMap[locID]
	if !ok {
		err = fmt.Errorf("(Location->Load) Location %s not found", locID)
		return
	}
	return
}

// LocRequest support GET for info|present|missing|inflight
func LocRequest(w http.ResponseWriter, r *http.Request) {

	locationID := chi.URLParam(r, "loc")
	function := chi.URLParam(r, "function")

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
		responder.RespondWithError(w, r, http.StatusInternalServerError, errMsg)
		return
	}

	if rmeth != "GET" {
		errMsg := fmt.Sprintf("(LocRequest) %s not supported", rmeth)
		responder.RespondWithError(w, r, http.StatusInternalServerError, errMsg)
		return
	}

	locConf, err := Load(locationID)
	if err != nil {
		responder.RespondWithError(w, r, http.StatusInternalServerError, err.Error())

		return
	}

	// ensure we only list nodes with Priority higher or equal to the one defined for the location

	MinPrio := locConf.MinPriority

	// find Node Types with Priority > MinPrio
	nodes := node.Nodes{}
	matchesminprioquery := bson.M{"priority": bson.M{"$ge": MinPrio}} // the node has a priority higher than the Locations minimum threshold

	switch function {

	case "missing":

		// 1) find nodes that have specific locationID but are not stored
		locationStoredFalseQuery := bson.M{"locations": bson.M{"id": "anltsm", "stored": false}}

		// 2) array with locations but specific location
		noLocationQuery := bson.M{"locations.id": bson.M{"$ne": locationID}}

		// 3) array is missing completely
		locationArrayMissing := bson.M{"locations": bson.M{"$exists": "false"}}

		// combine 1-3) we ensure we only list nodes with Priority higher or equal to the one defined for the location
		allNodesLocationMissingQuery := bson.M{"$or": []bson.M{locationStoredFalseQuery, noLocationQuery, locationArrayMissing}}

		// add priority
		missingNodesWithHighPriorityQuery := bson.M{"$and": []bson.M{allNodesLocationMissingQuery, matchesminprioquery}}

		// nodes with no JSON priority but Attr.Type that has a priority
		nodes.GetAll(missingNodesWithHighPriorityQuery)

		// list all nodes without Location set or marked as Location.stored==false  MongoDB
		responder.RespondWithData(w, r, nodes)
		return

	// 	list all nodes marked as stored==true in Location in MongoDB
	case "present":
		query := bson.M{"locations.stored": bson.M{"$eq": "true"}}
		nodes.GetAll(query)
		responder.RespondWithData(w, r, nodes)
		return

		// // list all nodes marked as Location.stored==false and priority
	case "inflight":
		locationstoredfalsequery := bson.M{"locations.stored": bson.M{"$eq": "false"}}
		query := bson.M{"$and": []bson.M{locationstoredfalsequery, matchesminprioquery}}
		nodes.GetAll(query)
		responder.RespondWithData(w, r, nodes)
		return

		// list all nodes that need to be restored from tape
	case "restore":
		query := bson.M{"locations.restore": bson.M{"$eq": "true"}}
		nodes.GetAll(query)
		responder.RespondWithData(w, r, nodes)
		return

	//return config info for Node.
	case "info":
		responder.RespondWithData(w, r, locConf)
		return

	default:

		errMsg := fmt.Sprintf("(Location) %s not supported", function)
		responder.RespondWithError(w, r, http.StatusInternalServerError, errMsg)

	}

	return
}

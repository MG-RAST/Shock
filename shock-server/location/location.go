package location

import (
	"fmt"
	"net/http"

	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/MG-RAST/golib/stretchr/goweb/context"
	"github.com/davecgh/go-spew/spew"
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

// GET, /location/{loc}/{function}, specify -H "Content-Type: application/json"
func LocRequest(ctx context.Context) {

	locationID := ctx.PathValue("loc")
	function := ctx.PathValue("function")

	fmt.Printf("LocRequest received locationID: %s, function: %s\n", locationID, function)
	logger.Debug(2, "(LocRequest) received locationID: %s, function: %s", locationID, function)

	rmeth := ctx.HttpRequest().Method

	u, err := request.Authenticate(ctx.HttpRequest())
	if err != nil && err.Error() != e.NoAuth {
		request.AuthError(err, ctx)
		return
	}

	// public user cannot use this
	if (u == nil) && (!conf.DEBUG_AUTH) {
		errMsg := "admin required"
		//errMsg := e.UnAuth
		responder.RespondWithError(ctx, http.StatusUnauthorized, errMsg)
		return
	}

	if (u != nil) && (!u.Admin) && (!conf.DEBUG_AUTH) {
		errMsg := e.UnAuth
		logger.Debug(2, "(LocRequest) attempt to use as non admin (user: %s)", u.Username)
		responder.RespondWithError(ctx, http.StatusInternalServerError, errMsg)
		return
	}

	if rmeth != "GET" {
		errMsg := fmt.Sprintf("(LocRequest) %s not supported", rmeth)
		responder.RespondWithError(ctx, http.StatusInternalServerError, errMsg)
		return
	}

	fmt.Printf("LocRequest passed auth bits and rmeth \n")

	locConf, err := Load(locationID)
	if err != nil {
		responder.RespondWithError(ctx, http.StatusInternalServerError, err.Error())
		fmt.Printf("LocRequest LOAD error \n")

		return
	}

	fmt.Printf("LocRequest worked \n")

	// ensure we only list nodes with Priority higher or equal to the one defined for the location

	switch function {

	case "missing":
		nodes := node.Nodes{}

		// we ensure we only list nodes with Priority higher or equal to the one defined for the location

		//query := bson.M{"$and": []bson.M{"Location": bson.M{"$ne": locationID}}, bson.M{"Priority": bson.M{"$gt": locConf.MinPriority}
		query := bson.M{"Location": bson.M{"$ne": locationID}}

		// 	bson.M{"$or": []interface{}{
		// 		bson.M{"$and": []interface{}{
		// 			bson.M{"Priority": "$gt": locConf.MinPriority},
		// 			bson.M{"Location.Id": "$eq": locationID},
		// 			bson.M{"Location.stored": "$eq": locConf.MinPriority},
		// 		},
		// 	},
		// }

		nodes.GetAll(query)

		spew.Dump(nodes)

		// list all nodes without Location set or marked as Location.stored==false  MongoDB
		responder.RespondWithData(ctx, nodes)
		return

	// 	// list all nodes marked as stored==true in Location in MongoDB
	// case "present":
	// 	present :=
	// 		responder.RespondWithData(ctx, present)
	// 	return

	// // list all nodes marked as Location.stored==false
	// case "inflight":
	// 	inflight :=
	// 		responder.RespondWithData(ctx, inflight)
	// 	return

	//return config info for Node.
	case "info":

		// spew.Dump(locConf)
		responder.RespondWithData(ctx, locConf)
		return

	default:

		errMsg := fmt.Sprintf("(Location) %s not supported", function)
		responder.RespondWithError(ctx, http.StatusInternalServerError, errMsg)

	}

	// if locationID == "" { // /node/{nid}/locations
	// 	locations := n.GetLocations()
	// 	responder.RespondWithData(ctx, locations)
	// 	return
	// } else { // /node/{nid}/locations/{loc}
	// 	var thisLocation Location
	// 	thisLocation, err = n.GetLocation(locationID)
	// 	if err != nil {
	// 		responder.RespondWithError(ctx, http.StatusInternalServerError, err.Error())
	// 		return
	// 	}
	// 	responder.RespondWithData(ctx, thisLocation)
	// 	return

	// }

	// if locationID != "" { // /node/{nid}/locations/{loc}
	// 	err = n.DeleteLocation(locationID)
	// 	if err != nil {
	// 		responder.RespondWithError(ctx, http.StatusInternalServerError, err.Error())
	// 		return
	// 	}
	// 	n.Save()
	// 	responder.RespondWithData(ctx, n.Locations)
	// } else { // /node/{nid}/locations

	// 	err = n.DeleteLocations()
	// 	if err != nil {
	// 		responder.RespondWithError(ctx, http.StatusInternalServerError, err.Error())
	// 		return
	// 	}
	// 	n.Save()
	// 	responder.RespondOK(ctx)

	// }

	// }

	return
}

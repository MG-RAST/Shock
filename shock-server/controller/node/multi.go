package node

import (
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/util"
	"github.com/jaredwilkening/goweb"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strings"
)

// GET: /node
// To do:
// - Iterate node queries
func (cr *Controller) ReadMany(cx *goweb.Context) {
	// Log Request and check for Auth
	request.Log(cx.Request)
	u, err := request.Authenticate(cx.Request)
	if err != nil && err.Error() != e.NoAuth {
		request.AuthError(err, cx)
		return
	}

	// Gather query params
	query := request.Q(cx.Request.URL.Query())

	// Setup query and nodes objects
	q := bson.M{}
	nodes := node.Nodes{}

	if u != nil {
		// Admin sees all
		if !u.Admin {
			q["$or"] = []bson.M{bson.M{"acl.read": []string{}}, bson.M{"acl.read": u.Uuid}, bson.M{"acl.owner": u.Uuid}}
		}
	} else {
		if conf.Bool(conf.Conf["anon-read"]) {
			// select on only nodes with no read rights set
			q["acl.read"] = []string{}
		} else {
			cx.RespondWithErrorMessage(e.NoAuth, http.StatusUnauthorized)
			return
		}
	}

	// Gather params to make db query. Do not include the
	// following list.
	paramlist := map[string]int{"limit": 1, "offset": 1, "query": 1, "querynode": 1}
	if query.Has("query") {
		for key, val := range query.All() {
			_, s := paramlist[key]
			if !s {
				q[fmt.Sprintf("attributes.%s", key)] = val[0]
			}
		}
	} else if query.Has("querynode") {
		for key, val := range query.All() {
			if key == "type" {
				querytypes := strings.Split(query.Value("type"), ",")
				q["type"] = bson.M{"$all": querytypes}
			} else {
				_, s := paramlist[key]
				if !s {
					q[key] = val[0]
				}
			}
		}
	}

	// defaults
	limit := 25
	offset := 0
	if query.Has("limit") {
		limit = util.ToInt(query.Value("limit"))
	}
	if query.Has("offset") {
		offset = util.ToInt(query.Value("offset"))
	}

	// Get nodes from db
	count, err := nodes.GetPaginated(q, limit, offset)
	if err != nil {
		logger.Error("err " + err.Error())
		cx.RespondWithError(http.StatusBadRequest)
		return
	}
	cx.RespondWithPaginatedData(nodes, limit, offset, count)
	return
}

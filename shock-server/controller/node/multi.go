package node

import (
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/MG-RAST/Shock/shock-server/util"
	"github.com/MG-RAST/golib/mgo/bson"
	"github.com/MG-RAST/golib/stretchr/goweb/context"
	"net/http"
	"strconv"
)

type M map[string]interface{}
type ListOfMaps []M

// GET: /node
// To do:
// - Iterate node queries
func (cr *NodeController) ReadMany(ctx context.Context) error {
	u, err := request.Authenticate(ctx.HttpRequest())
	if err != nil && err.Error() != e.NoAuth {
		return request.AuthError(err, ctx)
	}

	// Gather query params
	query := ctx.HttpRequest().URL.Query()

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
			return responder.RespondWithError(ctx, http.StatusUnauthorized, e.NoAuth)
		}
	}

	// Gather params to make db query. Do not include the
	// following list.
	paramlist := map[string]int{"limit": 1, "offset": 1, "query": 1, "querynode": 1}
	if _, ok := query["query"]; ok {
		for key := range query {
			if _, found := paramlist[key]; !found {
				keyStr := fmt.Sprintf("attributes.%s", key)
				value := query.Get(key)
				if value != "" {
					if numValue, err := strconv.Atoi(value); err == nil {
						q["$or"] = ListOfMaps{{keyStr: value}, {keyStr: numValue}}
					} else if value == "null" {
						q["$or"] = ListOfMaps{{keyStr: value}, {keyStr: nil}}
					} else {
						q[keyStr] = value
					}
				} else {
					existsMap := map[string]bool{
						"$exists": true,
					}
					q[keyStr] = existsMap
				}
			}
		}
	} else if _, ok := query["querynode"]; ok {
		for key := range query {
			if _, found := paramlist[key]; !found {
				value := query.Get(key)
				if value != "" {
					if numValue, err := strconv.Atoi(value); err == nil {
						q["$or"] = ListOfMaps{{key: value}, {key: numValue}}
					} else if value == "null" {
						q["$or"] = ListOfMaps{{key: value}, {key: nil}}
					} else {
						q[key] = value
					}
				} else {
					existsMap := map[string]bool{
						"$exists": true,
					}
					q[key] = existsMap
				}
			}
		}
	}

	// defaults
	limit := 25
	offset := 0
	if _, ok := query["limit"]; ok {
		limit = util.ToInt(query.Get("limit"))
	}
	if _, ok := query["offset"]; ok {
		offset = util.ToInt(query.Get("offset"))
	}

	// Get nodes from db
	count, err := nodes.GetPaginated(q, limit, offset)
	if err != nil {
		err_msg := "err " + err.Error()
		logger.Error(err_msg)
		return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
	}
	return responder.RespondWithPaginatedData(ctx, nodes, limit, offset, count)
}

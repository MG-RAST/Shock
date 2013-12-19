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
	"github.com/stretchr/goweb/context"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strings"
)

// GET: /node
// To do:
// - Iterate node queries
func (cr *NodeController) ReadMany(ctx context.Context) error {
	u, err := request.Authenticate(ctx.HttpRequest())
	if err != nil && err.Error() != e.NoAuth {
		return request.AuthError(err, ctx)
	}

	// Gather query params
	query := ctx.QueryParams()

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
	if query.Has("query") {
		for key := range query {
			_, s := paramlist[key]
			if !s {
				q[fmt.Sprintf("attributes.%s", key)] = ctx.QueryValue(key)
			}
		}
	} else if query.Has("querynode") {
		for key := range query {
			if key == "type" {
				querytypes := strings.Split(ctx.QueryValue("type"), ",")
				q["type"] = bson.M{"$all": querytypes}
			} else {
				_, s := paramlist[key]
				if !s {
					q[key] = ctx.QueryValue(key)
				}
			}
		}
	}

	// defaults
	limit := 25
	offset := 0
	if query.Has("limit") {
		limit = util.ToInt(ctx.QueryValue("limit"))
	}
	if query.Has("offset") {
		offset = util.ToInt(ctx.QueryValue("offset"))
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

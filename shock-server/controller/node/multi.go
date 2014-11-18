package node

import (
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/MG-RAST/Shock/shock-server/util"
	"github.com/MG-RAST/golib/go-uuid/uuid"
	"github.com/MG-RAST/golib/mgo/bson"
	"github.com/MG-RAST/golib/stretchr/goweb/context"
	"net/http"
	"strconv"
	"strings"
	//	"time"
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
	query := ctx.HttpRequest().URL.Query()

	// Setup query and nodes objects
	q := bson.M{}
	qAcls := bson.M{}
	qOpts := bson.M{}
	qPerm := bson.M{}
	nodes := node.Nodes{}

	if u != nil {
		// Admin sees all
		if !u.Admin {
			qPerm["$or"] = []bson.M{bson.M{"acl.read": "public"}, bson.M{"acl.read": u.Uuid}, bson.M{"acl.owner": u.Uuid}}
		}
	} else {
		if conf.ANON_READ {
			// select on only nodes that are publicly readable
			qPerm["acl.read"] = "public"
		} else {
			return responder.RespondWithError(ctx, http.StatusUnauthorized, e.NoAuth)
		}
	}

	// bson.M is a convenient alias for a map[string]interface{} map, useful for dealing with BSON in a native way.
	var OptsMArray []bson.M

	// Gather params to make db query. Do not include the following list.
	if _, ok := query["query"]; ok {
		for key := range query {
			paramlist := map[string]int{"limit": 1, "offset": 1, "query": 1}
			if _, found := paramlist[key]; !found {
				keyStr := fmt.Sprintf("attributes.%s", key)
				for _, value := range query[key] {
					if value != "" {
						if numValue, err := strconv.Atoi(value); err == nil {
							OptsMArray = append(OptsMArray, bson.M{"$or": []bson.M{bson.M{keyStr: value}, bson.M{keyStr: numValue}}})
						} else if value == "null" {
							OptsMArray = append(OptsMArray, bson.M{"$or": []bson.M{bson.M{keyStr: value}, bson.M{keyStr: nil}}})
						} else {
							OptsMArray = append(OptsMArray, bson.M{keyStr: value})
						}
					} else {
						OptsMArray = append(OptsMArray, bson.M{keyStr: map[string]bool{"$exists": true}})
					}
				}
			}
		}
	} else if _, ok := query["querynode"]; ok {
		for key := range query {
			paramlist := map[string]int{"limit": 1, "offset": 1, "querynode": 1, "owner": 1, "read": 1, "write": 1, "delete": 1, "public_owner": 1, "public_read": 1, "public_write": 1, "public_delete": 1}
			if _, found := paramlist[key]; !found {
				for _, value := range query[key] {
					if value != "" {
						if numValue, err := strconv.Atoi(value); err == nil {
							OptsMArray = append(OptsMArray, bson.M{"$or": []bson.M{bson.M{key: value}, bson.M{key: numValue}}})
						} else if value == "null" {
							OptsMArray = append(OptsMArray, bson.M{"$or": []bson.M{bson.M{key: value}, bson.M{key: nil}}})
						} else {
							OptsMArray = append(OptsMArray, bson.M{key: value})
						}
					} else {
						OptsMArray = append(OptsMArray, bson.M{key: map[string]bool{"$exists": true}})
					}
				}
			}
		}
	}

	if len(OptsMArray) > 0 {
		qOpts["$and"] = OptsMArray
	}

	// bson.M is a convenient alias for a map[string]interface{} map, useful for dealing with BSON in a native way.
	var AclsMArray []bson.M

	// Allowing user to query based on ACL's with a comma-separated list of users.
	// Restricting ACL queries to just the querynode operation.
	// Users can be written as a username or a UUID.
	if _, qok := query["querynode"]; qok {
		for _, atype := range []string{"owner", "read", "write", "delete"} {
			if _, ok := query[atype]; ok {
				users := strings.Split(query.Get(atype), ",")
				for _, v := range users {
					if uuid.Parse(v) != nil {
						AclsMArray = append(AclsMArray, bson.M{"acl." + atype: v})
					} else {
						u := user.User{Username: v}
						if err := u.SetMongoInfo(); err != nil {
							err_msg := "err " + err.Error()
							logger.Error(err_msg)
							return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
						}
						AclsMArray = append(AclsMArray, bson.M{"acl." + atype: u.Uuid})
					}
				}
			}
		}
		// Allowing users to query based on whether ACL is public
		for _, atype := range []string{"owner", "read", "write", "delete"} {
			if _, ok := query["public_"+atype]; ok {
				AclsMArray = append(AclsMArray, bson.M{"acl." + atype: "public"})
			}
		}
	}

	if len(AclsMArray) > 0 {
		qAcls["$and"] = AclsMArray
	}

	// Combine permissions query with query parameters and ACL query into one AND clause
	q["$and"] = []bson.M{qPerm, qOpts, qAcls}

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

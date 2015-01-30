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
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	RangeRegex = regexp.MustCompile(`^([\[\]]{1})(.*);(.*)([\]\[]{1})$`)
)

const (
	longDateForm  = "2006-01-02T15:04:05-07:00"
	shortDateForm = "2006-01-02"
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
	// Note: query is composed of 3 sub-query objects:
	// 1) qPerm - user permissions (system-defined)
	// 2) qOpts - query options (user-defined)
	// 3) qAcls - ACL queries (user-defined)
	q := bson.M{}
	qPerm := bson.M{}
	qOpts := bson.M{}
	qAcls := bson.M{}
	nodes := node.Nodes{}

	if u != nil {
		// Skip this part if user is an admin
		if !u.Admin {
			qPerm["$or"] = []bson.M{bson.M{"acl.read": "public"}, bson.M{"acl.read": u.Uuid}, bson.M{"acl.owner": u.Uuid}}
		}
	} else {
		// User is anonymous
		if conf.ANON_READ {
			// select on only nodes that are publicly readable
			qPerm["acl.read"] = "public"
		} else {
			return responder.RespondWithError(ctx, http.StatusUnauthorized, e.NoAuth)
		}
	}

	// bson.M is a convenient alias for a map[string]interface{} map, useful for dealing with BSON in a native way.
	var OptsMArray []bson.M

	// default sort field and direction (can only be changed with querynode operator, not query operator)
	order := "created_on"
	direction := "-"

	// Gather params to make db query. Do not include the following list.
	if _, ok := query["query"]; ok {
		paramlist := map[string]int{"limit": 1, "offset": 1, "query": 1}
		for key := range query {
			if _, found := paramlist[key]; !found {
				keyStr := fmt.Sprintf("attributes.%s", key)
				for _, value := range query[key] {
					if value != "" {
						OptsMArray = append(OptsMArray, parseOption(keyStr, value))
					} else {
						OptsMArray = append(OptsMArray, bson.M{keyStr: map[string]bool{"$exists": true}})
					}
				}
			}
		}
	} else if _, ok := query["querynode"]; ok {
		paramlist := map[string]int{"limit": 1, "offset": 1, "querynode": 1, "order": 1, "direction": 1, "owner": 1, "read": 1, "write": 1, "delete": 1, "public_owner": 1, "public_read": 1, "public_write": 1, "public_delete": 1}
		for key := range query {
			if _, found := paramlist[key]; !found {
				for _, value := range query[key] {
					if value != "" {
						OptsMArray = append(OptsMArray, parseOption(key, value))
					} else {
						OptsMArray = append(OptsMArray, bson.M{key: map[string]bool{"$exists": true}})
					}
				}
			}
		}
		if _, ok := query["order"]; ok {
			order = query.Get("order")
		}
		if _, ok := query["direction"]; ok {
			if query.Get("direction") == "asc" {
				direction = ""
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
	order = direction + order
	count, err := nodes.GetPaginated(q, limit, offset, order)
	if err != nil {
		err_msg := "err " + err.Error()
		logger.Error(err_msg)
		return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
	}
	return responder.RespondWithPaginatedData(ctx, nodes, limit, offset, count)
}

func parseOption(key string, value string) bson.M {
	not := false
	// If value starts with ! then set flag to encapsulate query with $not operator and
	//  remove ! character from the beginning of the value string
	if value[0] == '!' {
		value = value[1:]
		not = true
	}

	// Parsing query option into bson.M query object
	opt := bson.M{}

	// Only one of the following conditions can be met at a time
	// mongodb doesn't allow for negating the entire query so the logic
	//   has to be written for each case below when the not flag is set.
	if numValue, err := strconv.Atoi(value); err == nil {
		// numeric values
		if not {
			opt = bson.M{"$and": []bson.M{bson.M{key: bson.M{"$ne": value}}, bson.M{key: bson.M{"$ne": numValue}}}}
		} else {
			opt = bson.M{"$or": []bson.M{bson.M{key: value}, bson.M{key: numValue}}}
		}
	} else if value == "null" {
		// value is "null" => nil
		if not {
			opt = bson.M{"$and": []bson.M{bson.M{key: bson.M{"$ne": value}}, bson.M{key: bson.M{"$ne": nil}}}}
		} else {
			opt = bson.M{"$or": []bson.M{bson.M{key: value}, bson.M{key: nil}}}
		}
	} else if matches := RangeRegex.FindStringSubmatch(value); len(matches) > 0 {
		// value matches the regex for a range
		lowerBound := bson.M{}
		upperBound := bson.M{}
		var val1 interface{} = matches[2]
		var val2 interface{} = matches[3]
		parseTypedValue(&val1)
		parseTypedValue(&val2)
		if not {
			if matches[1] == "[" {
				lowerBound = bson.M{key: bson.M{"$lt": val1}}
			} else {
				lowerBound = bson.M{key: bson.M{"$lte": val1}}
			}
			if matches[4] == "]" {
				upperBound = bson.M{key: bson.M{"$gt": val2}}
			} else {
				upperBound = bson.M{key: bson.M{"$gte": val2}}
			}
			opt = bson.M{"$or": []bson.M{lowerBound, upperBound}}
		} else {
			if matches[1] == "[" {
				lowerBound = bson.M{key: bson.M{"$gte": val1}}
			} else {
				lowerBound = bson.M{key: bson.M{"$gt": val1}}
			}
			if matches[4] == "]" {
				upperBound = bson.M{key: bson.M{"$lte": val2}}
			} else {
				upperBound = bson.M{key: bson.M{"$lt": val2}}
			}
			opt = bson.M{"$and": []bson.M{lowerBound, upperBound}}
		}
	} else if string(value[0]) == "*" || string(value[len(value)-1]) == "*" {
		// value starts or ends with wildcard, or both
		// Note: The $not operator could probably be used for some of these queries but
		//       the $not operator does not support operations with the $regex operator
		//       thus I have built the opposite regexes below for the "not" option.
		if not {
			if string(value[0]) != "*" {
				value = value[0 : len(value)-1]
				opt = bson.M{key: bson.M{"$regex": "^(?!" + value + ").*$"}}
			} else if string(value[len(value)-1]) != "*" {
				value = value[1:]
				opt = bson.M{key: bson.M{"$regex": "^.*(?<!" + value + ")$"}}
			} else {
				value = value[1 : len(value)-1]
				opt = bson.M{key: bson.M{"$regex": "^((?!" + value + ").)*$"}}
			}
		} else {
			if string(value[0]) != "*" {
				value = value[0 : len(value)-1]
				opt = bson.M{key: bson.M{"$regex": "^" + value}}
			} else if string(value[len(value)-1]) != "*" {
				value = value[1:]
				opt = bson.M{key: bson.M{"$regex": value + "$"}}
			} else {
				value = value[1 : len(value)-1]
				opt = bson.M{key: bson.M{"$regex": value}}
			}
		}
	} else {
		if not {
			opt = bson.M{key: bson.M{"$ne": value}}
		} else {
			opt = bson.M{key: value}
		}
	}
	return opt
}

func parseTypedValue(i *interface{}) {
	if val, err := strconv.Atoi((*i).(string)); err == nil {
		*i = val
	} else if t, err := time.Parse(longDateForm, (*i).(string)); err == nil {
		*i = t
	} else if t, err := time.Parse(shortDateForm, (*i).(string)); err == nil {
		*i = t
	}
	return
}

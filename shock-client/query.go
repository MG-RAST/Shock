package main

import (
	"net/url"
	"strconv"
	"strings"
)

type queryNode struct {
	values   url.Values
	prefix   string
	distinct bool
	full     bool
}

func newQueryNode() queryNode {
	return queryNode{
		values:   url.Values{},
		prefix:   "",
		distinct: false,
		full:     false,
	}
}

func (q *queryNode) processFlags(queries arrayFlags) {
	for _, val := range queries {
		parts := strings.Split(val, ":")
		if len(parts) == 2 {
			name := q.prefix + parts[0]
			q.values.Set(name, parts[1])
		}
	}
}

func (q *queryNode) addOptions() {
	if limit != 0 {
		q.values.Set("limit", strconv.Itoa(limit))
	}
	if offset != 0 {
		q.values.Set("offset", strconv.Itoa(offset))
	}
	if (direction != "") && validateCV("direction", direction) {
		q.values.Set("direction", direction)
	}
	if order != "" {
		q.values.Set("order", order)
	}
	if distinct != "" {
		q.distinct = true
		q.values.Set("distinct", distinct)
	}
}

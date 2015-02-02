package node

import (
	"github.com/MG-RAST/golib/mgo/bson"
)

// Node array type
type Nodes []*Node

func (n *Nodes) GetAll(q bson.M) (err error) {
	_, err = dbFind(q, n, "", nil)
	return
}

func (n *Nodes) GetPaginated(q bson.M, limit int, offset int, order string) (count int, err error) {
	count, err = dbFind(q, n, order, map[string]int{"limit": limit, "offset": offset})
	return
}

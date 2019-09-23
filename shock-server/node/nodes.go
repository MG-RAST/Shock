package node

import (
	"gopkg.in/mgo.v2/bson"
)

// Node array type
type Nodes []*Node

func (n *Nodes) DBInit() {
	for i := 0; i < len(*n); i++ {
		(*n)[i].DBInit()
	}
}

func (n *Nodes) GetAll(q bson.M) (err error) {
	_, err = dbFind(q, n, "", nil)
	return
}

func (n *Nodes) GetAllD(q bson.D) (err error) {
	_, err = dbFindD(q, n, "", nil)
	return
}

func (n *Nodes) GetPaginated(q bson.M, limit int, offset int, order string) (count int, err error) {
	count, err = dbFind(q, n, order, map[string]int{"limit": limit, "offset": offset})
	return
}

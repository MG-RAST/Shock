package store

import (
	"labix.org/v2/mgo/bson"
)

// Node array type
type Nodes []*Node

func (n *Nodes) GetAll(q bson.M) (err error) {
	db, err := DBConnect()
	if err != nil {
		return
	}
	defer db.Close()
	_, err = db.Find(q, n, nil)
	return
}

func (n *Nodes) GetPaginated(q bson.M, limit int, offset int) (count int, err error) {
	db, err := DBConnect()
	if err != nil {
		return
	}
	defer db.Close()
	count, err = db.Find(q, n, map[string]int{"limit": limit, "offset": offset})
	return
}

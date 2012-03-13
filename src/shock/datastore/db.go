package datastore

import (
	"fmt"
	mgo "launchpad.net/mgo"
	bson "launchpad.net/mgo/bson"
)

type db struct {
	Nodes *mgo.Collection
	Session *mgo.Session
}
	
func DBConnect() (d *db, err error) {
	MONGODBHOST := "localhost"
	session, err := mgo.Dial(MONGODBHOST); if err != nil { return }
	d = &db{Nodes: session.DB("ShockDB").C("Nodes"), Session : session}	
	return
}

func (d *db) Upsert(node *Node) (err error) {
	_, err = d.Nodes.Upsert(bson.M{"id": node.Id}, &node)
	return
}

func (d *db) FindById(id string, result *Node) (err error) {	
	err = d.Nodes.Find(bson.M{"id": id}).One(&result)
	return
}

func (d *db) GetAll(q bson.M, results *Nodes) (err error) {
	fmt.Println(q)
	err = d.Nodes.Find(q).All(results)
	return
} 

func (d *db) GetAllLimitOffset(q bson.M, results *Nodes, limit int, offset int) (err error) {
	err = d.Nodes.Find(q).Limit(limit).Skip(offset).All(results)
	return
} 

func (d *db) Close() () {
	d.Session.Close()
	return
}

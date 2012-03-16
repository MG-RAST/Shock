package datastore

import (
	"errors"
	"time"
	mgo "launchpad.net/mgo"
	bson "launchpad.net/mgo/bson"
	conf "shock/conf"
)

const (
	DbTimeout = time.Duration(time.Second*1)
)

func init() {
	d, err := DBConnect(); if err != nil { panic(errors.New("No reachable mongodb servers.")) }	
	idIdx := mgo.Index{Key: []string{"id"}, Unique: true}
	err = d.Nodes.EnsureIndex(idIdx); if err != nil { panic(err) }
}

type db struct {
	Nodes *mgo.Collection
	Session *mgo.Session
}
	
func DBConnect() (d *db, err error) {
	session, err := mgo.DialWithTimeout(*conf.MONGODB, DbTimeout); if err != nil { return }
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

func (d *db) FindByIdAuth(id string, uuid string, result *Node) (err error) {	
	err = d.Nodes.Find(bson.M{"id": id}).One(&result); if err != nil { return }
	rights := result.Acl.check(uuid)
	if !rights["read"] {
		err = errors.New("User Unauthorized")
	}
	return
}

func (d *db) GetAll(q bson.M, results *Nodes) (err error) {
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

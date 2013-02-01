package store

import (
	"fmt"
	"github.com/MG-RAST/Shock/conf"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"os"
	"time"
)

const (
	DbTimeout = time.Duration(time.Second * 1)
)

func init() {
	InitDB()
}

type db struct {
	Nodes   *mgo.Collection
	Session *mgo.Session
}

func InitDB() {
	d, err := DBConnect()
	defer d.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: no reachable mongodb servers")
		os.Exit(1)
	}
	idIdx := mgo.Index{Key: []string{"id"}, Unique: true}
	err = d.Nodes.EnsureIndex(idIdx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: fatal mongodb initialization error: %v", err)
		os.Exit(1)
	}
}

func DBConnect() (d *db, err error) {
	session, err := mgo.DialWithTimeout(conf.MONGODB, DbTimeout)
	if err != nil {
		return
	}
	d = &db{Nodes: session.DB("ShockDB").C("Nodes"), Session: session}
	return
}

func DropDB() (err error) {
	d, err := DBConnect()
	defer d.Close()
	if err != nil {
		return err
	}
	return d.Nodes.DropCollection()
}

func (d *db) Delete(q bson.M) (err error) {
	_, err = d.Nodes.RemoveAll(q)
	return
}

func (d *db) Upsert(node *Node) (err error) {
	_, err = d.Nodes.Upsert(bson.M{"id": node.Id}, &node)
	return
}

func (d *db) Find(q bson.M, results *Nodes, options map[string]int) (err error) {
	if limit, haslimit := options["limit"]; haslimit {
		if offset, hasoffset := options["offset"]; hasoffset {
			err = d.Nodes.Find(q).Limit(limit).Skip(offset).All(results)
			return
		}
	}
	err = d.Nodes.Find(q).All(results)
	return
}

func (d *db) FindOne(q bson.M, result *Node) (err error) {
	err = d.Nodes.Find(q).One(&result)
	return
}

func (d *db) Close() {
	d.Session.Close()
	return
}

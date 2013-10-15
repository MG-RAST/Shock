package node

import (
	"errors"
	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/db"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	"path/filepath"
)

/*
var DB *mgo.Collection

func Initialize() {
	DB = db.Connection.DB.C("Nodes")
	DB.EnsureIndex(mgo.Index{Key: []string{"id"}, Unique: true})
}
*/

func dbDelete(q bson.M) (err error) {
	session := db.Connection.Session.Copy()
	DB := session.DB(conf.Conf["mongodb-database"]).C("Nodes")
	_, err = DB.RemoveAll(q)
	session.Close()
	return
}

func dbUpsert(n *Node) (err error) {
	session := db.Connection.Session.Copy()
	DB := session.DB(conf.Conf["mongodb-database"]).C("Nodes")
	_, err = DB.Upsert(bson.M{"id": n.Id}, &n)
	session.Close()
	return
}

func dbFind(q bson.M, results *Nodes, options map[string]int) (count int, err error) {
	session := db.Connection.Session.Copy()
	DB := session.DB(conf.Conf["mongodb-database"]).C("Nodes")
	if limit, has := options["limit"]; has {
		if offset, has := options["offset"]; has {
			query := DB.Find(q)
			if count, err = query.Count(); err != nil {
				return 0, err
			}
			err = query.Limit(limit).Skip(offset).All(results)
			return
		} else {
			return 0, errors.New("store.db.Find options limit and offset must be used together")
		}
	}
	err = DB.Find(q).All(results)
	session.Close()
	return
}

func Load(id string, uuid string) (n *Node, err error) {
	session := db.Connection.Session.Copy()
	DB := session.DB(conf.Conf["mongodb-database"]).C("Nodes")
	n = new(Node)
	if err = DB.Find(bson.M{"id": id}).One(&n); err == nil {
		rights := n.Acl.Check(uuid)
		if !rights["read"] {
			return nil, errors.New("User Unauthorized")
		}
		return n, nil
	} else {
		return nil, err
	}
	session.Close()
	return
}

func LoadUnauth(id string) (n *Node, err error) {
	session := db.Connection.Session.Copy()
	DB := session.DB(conf.Conf["mongodb-database"]).C("Nodes")
	n = new(Node)
	if err = DB.Find(bson.M{"id": id}).One(&n); err == nil {
		return n, nil
	} else {
		return nil, err
	}
	session.Close()
	return
}

func LoadNodes(ids []string) (n Nodes, err error) {
	if _, err = dbFind(bson.M{"id": bson.M{"$in": ids}}, &n, nil); err == nil {
		return n, err
	}
	return nil, err
}

func ReloadFromDisk(path string) (err error) {
	id := filepath.Base(path)
	nbson, err := ioutil.ReadFile(path + "/" + id + ".bson")
	if err != nil {
		return
	}
	node := new(Node)
	if err = bson.Unmarshal(nbson, &node); err == nil {
		if err = dbUpsert(node); err != nil {
			return err
		}
	}
	return
}

package node

import (
	"errors"
	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/db"
	"github.com/MG-RAST/golib/mgo"
	"github.com/MG-RAST/golib/mgo/bson"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// Initialize creates a copy of the mongodb connection and then uses that connection to
// create the Nodes collection in mongodb. Then, it ensures that there is a unique index
// on the id key in this collection, creating the index if necessary.
func Initialize() {
	session := db.Connection.Session.Copy()
	defer session.Close()
	c := session.DB(conf.MONGODB_DATABASE).C("Nodes")
	c.EnsureIndex(mgo.Index{Key: []string{"acl.owner"}, Background: true})
	c.EnsureIndex(mgo.Index{Key: []string{"acl.read"}, Background: true})
	c.EnsureIndex(mgo.Index{Key: []string{"acl.write"}, Background: true})
	c.EnsureIndex(mgo.Index{Key: []string{"acl.delete"}, Background: true})
	c.EnsureIndex(mgo.Index{Key: []string{"created_on"}, Background: true})
	c.EnsureIndex(mgo.Index{Key: []string{"file.path"}, Background: true})
	c.EnsureIndex(mgo.Index{Key: []string{"file.virtual_parts"}, Background: true})
	c.EnsureIndex(mgo.Index{Key: []string{"id"}, Unique: true})
	if conf.MONGODB_ATTRIBUTE_INDEXES != "" {
		for _, v := range strings.Split(conf.MONGODB_ATTRIBUTE_INDEXES, ",") {
			v = "attributes." + strings.TrimSpace(v)
			c.EnsureIndex(mgo.Index{Key: []string{v}, Background: true})
		}
	}
}

func dbDelete(q bson.M) (err error) {
	session := db.Connection.Session.Copy()
	defer session.Close()
	c := session.DB(conf.MONGODB_DATABASE).C("Nodes")
	_, err = c.RemoveAll(q)
	return
}

func dbUpsert(n *Node) (err error) {
	session := db.Connection.Session.Copy()
	defer session.Close()
	c := session.DB(conf.MONGODB_DATABASE).C("Nodes")
	_, err = c.Upsert(bson.M{"id": n.Id}, &n)
	return
}

func dbFind(q bson.M, results *Nodes, order string, options map[string]int) (count int, err error) {
	session := db.Connection.Session.Copy()
	defer session.Close()
	c := session.DB(conf.MONGODB_DATABASE).C("Nodes")
	if order == "" {
		order = "created_on"
	}
	if limit, has := options["limit"]; has {
		if offset, has := options["offset"]; has {
			query := c.Find(q).Sort(order)
			if count, err = query.Count(); err != nil {
				return 0, err
			}
			err = query.Limit(limit).Skip(offset).All(results)
			return
		} else {
			return 0, errors.New("store.db.Find options limit and offset must be used together")
		}
	}
	err = c.Find(q).Sort(order).All(results)
	return
}

func Load(id string) (n *Node, err error) {
	session := db.Connection.Session.Copy()
	defer session.Close()
	c := session.DB(conf.MONGODB_DATABASE).C("Nodes")
	n = new(Node)
	if err = c.Find(bson.M{"id": id}).One(&n); err == nil {
		return n, nil
	} else {
		return nil, err
	}
	return
}

func LoadNodes(ids []string) (n Nodes, err error) {
	if _, err = dbFind(bson.M{"id": bson.M{"$in": ids}}, &n, "", nil); err == nil {
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

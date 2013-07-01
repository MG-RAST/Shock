package preauth

import (
	"github.com/MG-RAST/Shock/shock-server/db"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

var DB *mgo.Collection

type PreAuth struct {
	Id        string
	Type      string
	NodeId    string
	Options   map[string]string
	ValidTill time.Time
}

func Initialize() {
	DB = db.Connection.DB.C("PreAuth")
	DB.EnsureIndex(mgo.Index{Key: []string{"id"}, Unique: true})
}

func New(id, t, nid string, options map[string]string) (p *PreAuth, err error) {
	p = &PreAuth{Id: id, Type: t, NodeId: nid, Options: options, ValidTill: time.Now().AddDate(0, 0, 1)}
	if _, err = DB.Upsert(bson.M{"id": p.Id}, &p); err != nil {
		return nil, err
	}
	return p, nil
}

func Load(id string) (p *PreAuth, err error) {
	p = &PreAuth{}
	if err = DB.Find(bson.M{"id": id}).One(&p); err != nil {
		return nil, err
	}
	return p, nil
}

func Delete(id string) (err error) {
	_, err = DB.RemoveAll(bson.M{"id": id})
	return err
}

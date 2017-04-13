// Package preauth implements persistent storage and retrieval of preauth requests
package preauth

import (
	"github.com/MG-RAST/Shock/shock-server/db"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

// Database collection handle
var DB *mgo.Collection

type PreAuthResponse struct {
	Url       string `json:"url"`
	ValidTill string `json:"validtill"`
	Format    string `json:"format"`
	Filename  string `json:"filename"`
	Files     int    `json:"files"`
	Size      int64  `json:"size"`
}

type PreAuth struct {
	Id        string            `bson:"id" json:"id"`
	Type      string            `bson:"type" json:"type"`
	Nodes     []string          `bson:"nodes" json:"nodes"`
	Options   map[string]string `bson:"options" json:"options"`
	ValidTill time.Time         `bson:"validtill" json:"validtill"`
}

// Initialize is an explicit init. Requires db.Initialize
// Indexes are applied to the collection at this time.
func Initialize() {
	DB = db.Connection.DB.C("PreAuth")
	DB.EnsureIndex(mgo.Index{Key: []string{"id"}, Unique: true})
}

// New preauth takes the id, type, node ids, and a map of options
func New(id string, t string, nids []string, options map[string]string) (p *PreAuth, err error) {
	p = &PreAuth{Id: id, Type: t, Nodes: nids, Options: options, ValidTill: time.Now().AddDate(0, 1, 0)}
	if _, err = DB.Upsert(bson.M{"id": p.Id}, &p); err != nil {
		return nil, err
	}
	return p, nil
}

// Load preauth by id
func Load(id string) (p *PreAuth, err error) {
	p = &PreAuth{}
	if err = DB.Find(bson.M{"id": id}).One(&p); err != nil {
		return nil, err
	}
	return p, nil
}

// Delete preauth by id
func Delete(id string) (err error) {
	_, err = DB.RemoveAll(bson.M{"id": id})
	return err
}

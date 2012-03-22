package user

import (
	"errors"
	"github.com/MG-RAST/Shock/conf"
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
	"time"
)

const (
	DbTimeout = time.Duration(time.Second * 1)
)

func init() {
	d, err := DBConnect()
	if err != nil {
		panic(errors.New("No reachable mongodb servers."))
	}
	uuidIdx := mgo.Index{Key: []string{"uuid"}, Unique: true}
	nameIdx := mgo.Index{Key: []string{"name"}, Unique: true}
	err = d.User.EnsureIndex(uuidIdx)
	if err != nil {
		panic(err)
	}
	err = d.User.EnsureIndex(nameIdx)
	if err != nil {
		panic(err)
	}
}

func DBConnect() (d *db, err error) {
	session, err := mgo.DialWithTimeout(*conf.MONGODB, DbTimeout)
	if err != nil {
		return
	}
	d = &db{User: session.DB("ShockDB").C("Users"), Session: session}
	return
}

type db struct {
	User    *mgo.Collection
	Session *mgo.Session
}

func (d *db) AdminGet(u *Users) (err error) {
	err = d.User.Find(nil).All(u)
	return
}

func (d *db) GetUser(u *User) (err error) {
	if u.Uuid != "" {
		err = d.User.Find(bson.M{"uuid": u.Uuid}).One(&u)
	} else {
		err = d.User.Find(bson.M{"name": u.Name, "passwd": u.Passwd}).One(&u)
	}
	return
}

func (d *db) Insert(user *User) (err error) {
	err = d.User.Insert(&user)
	return
}

func (d *db) Close() {
	d.Session.Close()
	return
}

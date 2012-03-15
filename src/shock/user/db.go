package user

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
	uuidIdx := mgo.Index{Key: []string{"uuid"}, Unique: true}
	nameIdx := mgo.Index{Key: []string{"name"}, Unique: true}
	err = d.User.EnsureIndex(uuidIdx); if err != nil { panic(err) }
	err = d.User.EnsureIndex(nameIdx); if err != nil { panic(err) }
}

type db struct {
	User *mgo.Collection
	Session *mgo.Session
}
	
func DBConnect() (d *db, err error) {
	session, err := mgo.DialWithTimeout(*conf.MONGODB, DbTimeout); if err != nil { return }
	d = &db{User: session.DB("ShockDB").C("Users"), Session : session}	
	return
}

func (d *db) Upsert(user *User) (err error) {
	_, err = d.User.Upsert(bson.M{"uuid": user.Uuid}, &user)
	return
}

func (d *db) Close() () {
	d.Session.Close()
	return
}

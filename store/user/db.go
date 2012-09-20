package user

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
	d, err := DBConnect()
	if err != nil {
		fmt.Fprintln(os.Stderr, "user: no reachable mongodb server")
		os.Exit(1)
	}
	uuidIdx := mgo.Index{Key: []string{"uuid"}, Unique: true}
	nameIdx := mgo.Index{Key: []string{"username"}, Unique: true}
	err = d.User.EnsureIndex(uuidIdx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "user: fatal initialization error: %v", err)
		os.Exit(1)
	}
	err = d.User.EnsureIndex(nameIdx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "user: fatal initialization error: %v", err)
		os.Exit(1)
	}
}

func DBConnect() (d *db, err error) {
	session, err := mgo.DialWithTimeout(conf.MONGODB, DbTimeout)
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

func (d *db) GetUuid(username string) (uuid string, err error) {
	u := User{}
	err = d.User.Find(bson.M{"username": username}).One(&u)
	if err == nil {
		return u.Uuid, nil
	}
	return "", err
}

func (d *db) GetUser(u *User) (err error) {
	if u.Uuid != "" {
		err = d.User.Find(bson.M{"uuid": u.Uuid}).One(&u)
	} else {
		err = d.User.Find(bson.M{"username": u.Username, "password": u.Password}).One(&u)
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

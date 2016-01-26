// Package db to connect to mongodb
package db

import (
	"errors"
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	mgo "github.com/MG-RAST/Shock/vendor/gopkg.in/mgo.v2"
	"time"
)

const (
	DbTimeout    = time.Duration(time.Second * 1200)
	DialTimeout  = time.Duration(time.Second * 10)
	DialAttempts = 5
)

var (
	Connection connection
)

type connection struct {
	dbname   string
	username string
	password string
	Session  *mgo.Session
	DB       *mgo.Database
}

func Initialize() (err error) {
	c := connection{}

	// test connection
	canDial := false
	for i := 0; i < DialAttempts; i++ {
		s, err := mgo.DialWithTimeout(conf.MONGODB_HOSTS, DialTimeout)
		s.Close()
		if err == nil {
			canDial = true
			break
		}
	}
	if !canDial {
		return errors.New(fmt.Sprintf("no reachable mongodb server(s) at %s", conf.MONGODB_HOSTS))
	}

	// get handle
	s, err := mgo.DialWithTimeout(conf.MONGODB_HOSTS, DbTimeout)
	if err != nil {
		return errors.New(fmt.Sprintf("no reachable mongodb server(s) at %s", conf.MONGODB_HOSTS))
	}
	c.Session = s
	c.DB = c.Session.DB(conf.MONGODB_DATABASE)
	if conf.MONGODB_USER != "" && conf.MONGODB_PASSWORD != "" {
		c.DB.Login(conf.MONGODB_USER, conf.MONGODB_PASSWORD)
	}
	Connection = c
	return
}

func Drop() error {
	return Connection.DB.DropDatabase()
}

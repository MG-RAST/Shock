package user

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"time"
)

// Array of User
type Users []UserForListing

type UserForListing struct {
	Uuid  string `bson:"uuid" json:"uuid"`
	Name  string `bson:"name" json:"name"`
	Admin bool   `bson:"admin" json:"admin"`
}

// User struct
type User struct {
	Uuid   string `bson:"uuid" json:"uuid"`
	Name   string `bson:"name" json:"name"`
	Passwd string `bson:"passwd" json:"passwd"`
	Admin  bool   `bson:"admin" json:"admin"`
}

func New(name string, passwd string, isAdmin bool) (u *User, err error) {
	u = &User{Name: name, Passwd: passwd, Admin: isAdmin}
	u.SetUuid()
	err = u.Save()
	if err != nil {
		u = nil
	}
	return
}

func Authenticate(name string, passwd string) (u *User, err error) {
	d, err := DBConnect()
	if err != nil {
		return
	}
	defer d.Close()
	u = &User{Name: name, Passwd: passwd}
	err = d.GetUser(u)
	if err != nil {
		u = nil
	}
	return
}

func FindByUuid(uuid string) (u *User, err error) {
	d, err := DBConnect()
	if err != nil {
		return
	}
	defer d.Close()
	u = &User{Uuid: uuid}
	err = d.GetUser(u)
	if err != nil {
		u = nil
	}
	return
}

func AdminGet(u *Users) (err error) {
	d, err := DBConnect()
	if err != nil {
		return
	}
	defer d.Close()
	err = d.AdminGet(u)
	return
}

func (u *User) RemovePasswd() (nu *User) {
	nu = &User{Uuid: u.Uuid, Name: u.Name, Passwd: "**********", Admin: u.Admin}
	return
}

func (u *User) SetUuid() {
	var s []byte
	h := md5.New()
	h.Write([]byte(fmt.Sprint(u.Name, u.Passwd, time.Now().String(), rand.Float64())))
	s = h.Sum(s)
	u.Uuid = fmt.Sprintf("%x", s)
}

func (u *User) Save() (err error) {
	d, err := DBConnect()
	if err != nil {
		return
	}
	defer d.Close()
	err = d.Insert(u)
	return
}

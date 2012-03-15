package user

import (
	"fmt"
	"time"	
	"crypto/md5"
	"math/rand"
)

// User struct
type User struct {
	Uuid		string
	Name		string
	Passwd		string
	IsAdmin		bool
}
	
/*
u, err := user.New("jared", "brony", true); if err != nil {
	fmt.Println("User create error:", err.Error())
}
*/
	
func New(name string, passwd string, isAdmin bool) (u *User, err error){
	u = &User{Name : name, Passwd : passwd, IsAdmin : isAdmin}
	u.SetUuid()
	err = u.Save(); if err != nil { u = nil }
	return	
}

func Authenticate(name string, passwd string) (u *User, err error){
	d, err := DBConnect(); if err != nil { return }
	defer d.Close()	
	u = &User{Name : name, Passwd : passwd}	
	err = d.GetUser(u); if err != nil { u = nil }
	return
}

func (u *User) SetUuid() {
	var s []byte
	h := md5.New()
	h.Write([]byte(fmt.Sprint(u.Name, u.Passwd, time.Now().String(), rand.Float64())))	
	s = h.Sum(s)
	u.Uuid = fmt.Sprintf("%x",s)
}

func (u *User) Save() (err error){
	d, err := DBConnect(); if err != nil { return }
	defer d.Close()	
	err = d.Insert(u)
	return
}
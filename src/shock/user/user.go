package user

import (
	"fmt"
	"time"	
	"crypto/md5"
	"math/rand"
	"errors"
)

// User struct
type User struct {
	Uuid		string
	Name		string
	Passwd		string
	IsAdmin		bool
}

func New(name string, passwd string, isAdmin bool) (u *User, err error){
	u = &User{Name : name, Passwd : passwd, IsAdmin : isAdmin}
	u.SetUuid()
	//err = errors.New("User not found")
	return	
}

func Authenticate(user string, passwd string) (u *User, err error){
	u = &User{}
	err = errors.New("User not found")
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
	return
}
package user

import (
	"github.com/MG-RAST/Shock/shock-server/store/uuid"
)

// Array of User
type Users []User

// User struct
type User struct {
	Uuid         string      `bson:"uuid" json:"uuid"`
	Username     string      `bson:"username" json:"username"`
	FullName     string      `bson:"fullname" json:"fullname"`
	Email        string      `bson:"email" json:"email"`
	Password     string      `bson:"password" json:"-"`
	Admin        bool        `bson:"shock_admin" json:"shock_admin"`
	CustomFields interface{} `bson:"custom_fields" json:"custom_fields"`
}

func New(username string, password string, isAdmin bool) (u *User, err error) {
	u = &User{Uuid: uuid.New(), Username: username, Password: password, Admin: isAdmin}
	if err = u.Save(); err != nil {
		u = nil
	}
	return
}

func FindByUuid(uuid string) (u *User, err error) {
	if d, err := DBConnect(); err == nil {
		defer d.Close()
		u = &User{Uuid: uuid}
		if err = d.GetUser(u); err != nil {
			u = nil
		}
	}
	return
}

func AdminGet(u *Users) (err error) {
	if d, err := DBConnect(); err == nil {
		defer d.Close()
		err = d.AdminGet(u)
	}
	return
}

func (u *User) SetUuid() (err error) {
	if d, err := DBConnect(); err == nil {
		defer d.Close()
		if uu, err := d.GetUuid(u.Email); err == nil {
			u.Uuid = uu
			return nil
		} else {
			u.Uuid = uuid.New()
			if err := d.Insert(u); err != nil {
				return err
			}
		}
	}
	return
}

func (u *User) Save() (err error) {
	if d, err := DBConnect(); err == nil {
		defer d.Close()
		err = d.Insert(u)
	}
	return
}

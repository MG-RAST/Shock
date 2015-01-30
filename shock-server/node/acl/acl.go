package acl

import (
	"github.com/MG-RAST/Shock/shock-server/user"
)

// Node.Acl struct
type Acl struct {
	Owner  string   `bson:"owner" json:"owner"`
	Read   []string `bson:"read" json:"read"`
	Write  []string `bson:"write" json:"write"`
	Delete []string `bson:"delete" json:"delete"`
}

// struct for public status of ACL's
type publicAcl struct {
	Read   bool `bson:"read" json:"read"`
	Write  bool `bson:"write" json:"write"`
	Delete bool `bson:"delete" json:"delete"`
}

// ACL struct that is returned to user for verbosity = minimal (default)
type DisplayAcl struct {
	Owner  string    `bson:"owner" json:"owner"`
	Read   []string  `bson:"read" json:"read"`
	Write  []string  `bson:"write" json:"write"`
	Delete []string  `bson:"delete" json:"delete"`
	Public publicAcl `bson:"public" json:"public"`
}

// ACL struct that is returned to user for verbosity = full
type DisplayVerboseAcl struct {
	Owner  user.User   `bson:"owner" json:"owner"`
	Read   []user.User `bson:"read" json:"read"`
	Write  []user.User `bson:"write" json:"write"`
	Delete []user.User `bson:"delete" json:"delete"`
	Public publicAcl   `bson:"public" json:"public"`
}

type Rights map[string]bool

func (a *Acl) FormatDisplayAcl(verbosity string) (i interface{}) {
	if verbosity == "full" {
		dAcl := new(DisplayVerboseAcl)
		if u, err := user.FindByUuid(a.Owner); err == nil {
			dAcl.Owner = *u
		}
		dAcl.Public.Read = false
		dAcl.Public.Write = false
		dAcl.Public.Delete = false
		for _, v := range a.Read {
			if v == "public" {
				dAcl.Public.Read = true
			} else {
				if u, err := user.FindByUuid(v); err == nil {
					dAcl.Read = insertUser(dAcl.Read, *u)
				}
			}
		}
		for _, v := range a.Write {
			if v == "public" {
				dAcl.Public.Write = true
			} else {
				if u, err := user.FindByUuid(v); err == nil {
					dAcl.Write = insertUser(dAcl.Write, *u)
				}
			}
		}
		for _, v := range a.Delete {
			if v == "public" {
				dAcl.Public.Delete = true
			} else {
				if u, err := user.FindByUuid(v); err == nil {
					dAcl.Delete = insertUser(dAcl.Delete, *u)
				}
			}
		}
		i = dAcl
	} else {
		dAcl := new(DisplayAcl)
		dAcl.Owner = a.Owner
		dAcl.Read = []string{}
		dAcl.Write = []string{}
		dAcl.Delete = []string{}
		dAcl.Public.Read = false
		dAcl.Public.Write = false
		dAcl.Public.Delete = false
		for _, v := range a.Read {
			if v == "public" {
				dAcl.Public.Read = true
			} else {
				dAcl.Read = insert(dAcl.Read, v)
			}
		}
		for _, v := range a.Write {
			if v == "public" {
				dAcl.Public.Write = true
			} else {
				dAcl.Write = insert(dAcl.Write, v)
			}
		}
		for _, v := range a.Delete {
			if v == "public" {
				dAcl.Public.Delete = true
			} else {
				dAcl.Delete = insert(dAcl.Delete, v)
			}
		}
		i = dAcl
	}
	return
}

func (a *Acl) SetOwner(str string) {
	a.Owner = str
	return
}

func (a *Acl) UnSet(str string, r Rights) {
	if r["read"] {
		a.Read = del(a.Read, str)
	}
	if r["write"] {
		a.Write = del(a.Write, str)
	}
	if r["delete"] {
		a.Delete = del(a.Delete, str)
	}
	return
}

func (a *Acl) Set(str string, r Rights) {
	if r["read"] {
		a.Read = insert(a.Read, str)
	}
	if r["write"] {
		a.Write = insert(a.Write, str)
	}
	if r["delete"] {
		a.Delete = insert(a.Delete, str)
	}
	return
}

func (a *Acl) Check(str string) (r Rights) {
	r = Rights{"read": false, "write": false, "delete": false}
	acls := map[string][]string{"read": a.Read, "write": a.Write, "delete": a.Delete}
	for k, v := range acls {
		for _, id := range v {
			if str == id {
				r[k] = true
				break
			}
		}
	}
	return
}

func del(arr []string, s string) (narr []string) {
	narr = []string{}
	for i, item := range arr {
		if item != s {
			narr = append(narr, item)
		} else {
			narr = append(narr, arr[i+1:]...)
			break
		}
	}
	return
}

func insert(arr []string, s string) []string {
	for _, item := range arr {
		if item == s {
			return arr
		}
	}
	return append(arr, s)
}

func insertUser(arr []user.User, s user.User) []user.User {
	for _, item := range arr {
		if item == s {
			return arr
		}
	}
	return append(arr, s)
}

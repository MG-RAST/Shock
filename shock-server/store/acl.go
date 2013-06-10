package store

import ()

// Node.Acl struct
type Acl struct {
	Owner  string   `bson:"owner" json:"owner"`
	Read   []string `bson:"read" json:"read"`
	Write  []string `bson:"write" json:"write"`
	Delete []string `bson:"delete" json:"delete"`
}

type rights map[string]bool

func (a *Acl) SetOwner(uuid string) {
	a.Owner = uuid
	return
}

func (a *Acl) UnSet(uuid string, r rights) {
	if r["read"] {
		a.Read = del(a.Read, uuid)
	}
	if r["write"] {
		a.Write = del(a.Write, uuid)
	}
	if r["delete"] {
		a.Delete = del(a.Delete, uuid)
	}
	return
}

func (a *Acl) Set(uuid string, r rights) {
	if r["read"] {
		a.Read = insert(a.Read, uuid)
	}
	if r["write"] {
		a.Write = insert(a.Write, uuid)
	}
	if r["delete"] {
		a.Delete = insert(a.Delete, uuid)
	}
	return
}

func (a *Acl) Check(uuid string) (r rights) {
	r = rights{"read": false, "write": false, "delete": false}
	acls := map[string][]string{"read": a.Read, "write": a.Write, "delete": a.Delete}
	for k, v := range acls {
		if len(v) == 0 {
			r[k] = true
		} else {
			for _, id := range a.Read {
				if uuid == id {
					r[k] = true
					break
				}
			}
		}
	}
	return
}

func del(arr []string, s string) (narr []string) {
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

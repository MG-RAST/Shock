package datastore

import ()

// Node.acl struct
type acl struct {
	Read   []string `bson:"read" json:"read"`
	Write  []string `bson:"write" json:"write"`
	Delete []string `bson:"delete" json:"delete"`
}

type rights map[string]bool

func (a *acl) set(uuid string, r rights) {	
	if r["read"] { a.Read = append(a.Read, uuid) } 
	if r["write"] { a.Write = append(a.Write, uuid) } 
	if r["delete"] { a.Delete = append(a.Delete, uuid) }
	return
}

func (a *acl) check(uuid string) (r rights) {
	r = rights{"read":false, "write":false, "delete":false}
	acls := map[string][]string{"read":a.Read, "write":a.Write, "delete":a.Delete}
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
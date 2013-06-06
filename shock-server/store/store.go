package store

import (
	"errors"
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/store/user"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	"path/filepath"
	"time"
)

type PreAuth struct {
	Id        string
	Type      string
	NodeId    string
	Options   map[string]string
	ValidTill time.Time
}

func NewPreAuth(id, t, nid string, options map[string]string) (p *PreAuth, err error) {
	p = &PreAuth{Id: id, Type: t, NodeId: nid, Options: options, ValidTill: time.Now().AddDate(0, 0, 1)}
	if db, err := DBConnect(); err == nil {
		defer db.Close()
		if err = db.AddPreAuth(p); err != nil {
			return nil, err
		} else {
			return p, nil
		}
	} else {
		return nil, err
	}
	return
}

func LoadPreAuth(id string) (p *PreAuth, err error) {
	p = &PreAuth{}
	if db, err := DBConnect(); err == nil {
		defer db.Close()
		err = db.FindPreAuth(id, p)
		return p, err
	} else {
		return nil, err
	}
	return
}

func DeletePreAuth(id string) (err error) {
	db, err := DBConnect()
	defer db.Close()
	if err == nil {
		err = db.DeletePreAuth(id)
	}
	return err
}

func LoadNode(id string, uuid string) (node *Node, err error) {
	if db, err := DBConnect(); err == nil {
		defer db.Close()
		node = new(Node)
		if err = db.FindOne(bson.M{"id": id}, node); err == nil {
			rights := node.Acl.Check(uuid)
			if !rights["read"] {
				return nil, errors.New("User Unauthorized")
			}
			return node, nil
		} else {
			return nil, err
		}
	}
	return nil, err
}

func LoadNodeUnauth(id string) (node *Node, err error) {
	if db, err := DBConnect(); err == nil {
		defer db.Close()
		node = new(Node)
		if err = db.FindOne(bson.M{"id": id}, node); err == nil {
			return node, nil
		} else {
			return nil, err
		}
	}
	return nil, err
}

func LoadNodes(ids []string) (nodes Nodes, err error) {
	if db, err := DBConnect(); err == nil {
		defer db.Close()
		if _, err = db.Find(bson.M{"id": bson.M{"$in": ids}}, &nodes, nil); err == nil {
			return nodes, err
		} else {
			return nil, err
		}
	}
	return nil, err
}

func LoadNodeFromDisk(id string) (node *Node, err error) {
	path := getPath(id)
	nbson, err := ioutil.ReadFile(path + "/" + id + ".bson")
	if err != nil {
		return
	}
	node = new(Node)
	err = bson.Unmarshal(nbson, &node)
	if err != nil {
		node = nil
	}
	return
}

func ReloadFromDisk(path string) (err error) {
	id := filepath.Base(path)
	nbson, err := ioutil.ReadFile(path + "/" + id + ".bson")
	if err != nil {
		return
	}
	node := new(Node)
	err = bson.Unmarshal(nbson, &node)
	if err == nil {
		db, er := DBConnect()
		if er != nil {
			err = er
		}
		defer db.Close()
		err = db.Upsert(node)
		if err != nil {
			err = er
		}
	}
	return
}

func NewNode() (node *Node) {
	node = new(Node)
	node.Indexes = make(map[string]IdxInfo)
	node.File.Checksum = make(map[string]string)
	node.setId()
	return
}

func CreateNodeUpload(u *user.User, params map[string]string, files FormFiles) (node *Node, err error) {
	node = NewNode()
	if u.Uuid != "" {
		node.Acl.SetOwner(u.Uuid)
		node.Acl.Set(u.Uuid, rights{"read": true, "write": true, "delete": true})
	} else {
		node.Acl = acl{Owner: "", Read: make([]string, 0), Write: make([]string, 0), Delete: make([]string, 0)}
	}
	err = node.Mkdir()
	if err != nil {
		return
	}
	err = node.Update(params, files)
	if err != nil {
		return
	}
	err = node.Save()
	return
}

func getPath(id string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s", conf.DATA_PATH, id[0:2], id[2:4], id[4:6], id)
}

func getIndexPath(id string) string {
	return fmt.Sprintf("%s/idx", getPath(id))
}

/*
func CreateNode(filePath string, attrPath string) (node *Node, error) {
	var (
		attrStat *os.FileInfo
		in       *os.File
		out      *os.File
	)
	node = new(Node)
	node.Indexes = make(map[string]string)
	node.File.Checksum = make(map[string]string)
	node.setId()

	err = node.Mkdir(); if err != nil {	return }
	if filePath != "" {
		fileStat, err := os.Stat(filePath); if err != nil {	return }

		if fileStat.IsDirectory() {
			err = errors.New("directory found: wft?")
			return
		}
		var bytesRead int = 1
		md5h := md5.New()
		sha1h := sha1.New()
		in, err = os.Open(filePath)
		if err != nil {
			return
		}
		defer in.Close()
		out, err = os.Create(node.DataPath())
		if err != nil {
			return
		}
		defer out.Close()
		for bytesRead > 0 {
			buffer := make([]byte, 10240)
			bytesRead, err = in.Read(buffer)
			if err != nil && err.String() == "EOF" {
				err = nil
			} else if err != nil {
				return
			}
			md5h.Write(buffer[0:bytesRead])
			sha1h.Write(buffer[0:bytesRead])
			out.Write(buffer[0:bytesRead])
		}
		// set file struct
		node.File.Name = filepath.Base(filePath)
		node.File.Size = fileStat.Size
		node.File.Checksum["md5"] = fmt.Sprintf("%x", md5h.Sum())
		node.File.Checksum["sha1"] = fmt.Sprintf("%x", sha1h.Sum())
	}
	if attrPath != "" {
		attrStat, err = os.Stat(attrPath)
		if err != nil {
			return
		}
		if attrStat.IsDirectory() {
			err = errors.New("directory found: wft?")
			return
		} else {
			var attributes []byte
			attributes, err = ioutil.ReadFile(attrPath)
			if err != nil {
				return
			}
			err = json.Unmarshal(attributes, &node.Attributes)
			if err != nil {
				return
			}
		}
	}
	err = node.Save()
	return
}
*/
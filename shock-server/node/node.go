package node

import (
	"encoding/json"
	"errors"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/node/acl"
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/MG-RAST/Shock/shock-server/node/file/index"
	"github.com/MG-RAST/Shock/shock-server/user"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	"os"
)

type Node struct {
	Id           string            `bson:"id" json:"id"`
	Version      string            `bson:"version" json:"version"`
	File         file.File         `bson:"file" json:"file"`
	Attributes   interface{}       `bson:"attributes" json:"attributes"`
	Public       bool              `bson:"public" json:"public"`
	Indexes      Indexes           `bson:"indexes" json:"indexes"`
	Acl          acl.Acl           `bson:"acl" json:"-"`
	VersionParts map[string]string `bson:"version_parts" json:"-"`
	Tags         []string          `bson:"tags" json:"tags"`
	Revisions    []Node            `bson:"revisions" json:"-"`
	Linkages     []linkage         `bson:"linkage" json:"linkages"`
	CreatedOn    string            `bson:"created_on" json:"created_on"`
	LastModified string            `bson:"last_modified" json:"last_modified"`
}

type linkage struct {
	Type      string   `bson: "relation" json:"relation"`
	Ids       []string `bson:"ids" json:"ids"`
	Operation string   `bson:"operation" json:"operation"`
}

type Indexes map[string]IdxInfo

type IdxInfo struct {
	Type        string `bson:"index_type" json:"-"`
	TotalUnits  int64  `bson:"total_units" json:"total_units"`
	AvgUnitSize int64  `bson:"average_unit_size" json:"average_unit_size"`
}

type FormFiles map[string]FormFile

type FormFile struct {
	Name     string
	Path     string
	Checksum map[string]string
}

func New() (node *Node) {
	node = new(Node)
	node.Indexes = make(map[string]IdxInfo)
	node.File.Checksum = make(map[string]string)
	node.setId()
	node.LastModified = "-"
	return
}

func LoadFromDisk(id string) (n *Node, err error) {
	if len(id) < 6 {
		return nil, errors.New("Node ID must be at least 6 characters in length")
	}
	path := getPath(id)
	if nbson, err := ioutil.ReadFile(path + "/" + id + ".bson"); err != nil {
		return nil, errors.New("Node does not exist")
	} else {
		n = new(Node)
		if err = bson.Unmarshal(nbson, &n); err != nil {
			return nil, err
		}
	}
	return
}

func CreateNodeUpload(u *user.User, params map[string]string, files FormFiles) (node *Node, err error) {
	node = New()
	if u.Uuid != "" {
		node.Acl.SetOwner(u.Uuid)
		node.Acl.Set(u.Uuid, acl.Rights{"read": true, "write": true, "delete": true})
		node.Public = false
	} else {
		node.Acl = acl.Acl{Owner: "", Read: make([]string, 0), Write: make([]string, 0), Delete: make([]string, 0)}
		node.Public = true
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

func (node *Node) FileReader() (reader file.ReaderAt, err error) {
	if node.File.Virtual {
		readers := []file.ReaderAt{}
		nodes := Nodes{}
		if _, err := dbFind(bson.M{"id": bson.M{"$in": node.File.VirtualParts}}, &nodes, nil); err != nil {
			return nil, err
		}
		if len(nodes) > 0 {
			for _, n := range nodes {
				if r, err := n.FileReader(); err == nil {
					readers = append(readers, r)
				} else {
					return nil, err
				}
			}
		}
		return file.MultiReaderAt(readers...), nil
	}
	return os.Open(node.FilePath())
}

// Index functions
func (node *Node) Index(name string) (idx index.Index, err error) {
	if index.Has(name) {
		idx = index.NewVirtual(name, node.FilePath(), node.File.Size, 10240)
	} else {
		idx = index.New()
		err = idx.Load(node.IndexPath() + "/" + name + ".idx")
	}
	return
}

func (node *Node) Delete() (err error) {
	// check to make sure this node isn't referenced by a vnode
	nodes := Nodes{}
	if _, err = dbFind(bson.M{"virtual_parts": node.Id}, &nodes, nil); err != nil {
		return err
	}
	if len(nodes) != 0 {
		return errors.New(e.NodeReferenced)
	} else {
		if err = dbDelete(bson.M{"id": node.Id}); err != nil {
			return err
		}
	}
	return node.Rmdir()
}

func (node *Node) SetIndexInfo(indextype string, idxinfo IdxInfo) (err error) {
	node.Indexes[indextype] = idxinfo
	err = node.Save()
	return
}

func (node *Node) SetFileFormat(format string) (err error) {
	node.File.Format = format
	err = node.Save()
	return
}

func (node *Node) SetAttributes(attr FormFile) (err error) {
	attributes, err := ioutil.ReadFile(attr.Path)
	if err != nil {
		return
	}
	err = json.Unmarshal(attributes, &node.Attributes)
	if err != nil {
		return
	}
	err = node.Save()
	return
}

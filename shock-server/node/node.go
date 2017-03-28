package node

import (
	"encoding/json"
	"errors"
	"fmt"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/node/acl"
	"github.com/MG-RAST/Shock/shock-server/node/archive"
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/MG-RAST/Shock/shock-server/node/file/index"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/MG-RAST/Shock/shock-server/util"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

type Node struct {
	Id           string            `bson:"id" json:"id"`
	Version      string            `bson:"version" json:"version"`
	File         file.File         `bson:"file" json:"file"`
	Attributes   interface{}       `bson:"attributes" json:"attributes"`
	Indexes      Indexes           `bson:"indexes" json:"indexes"`
	Acl          acl.Acl           `bson:"acl" json:"-"`
	VersionParts map[string]string `bson:"version_parts" json:"version_parts"`
	Tags         []string          `bson:"tags" json:"tags"`
	Revisions    []Node            `bson:"revisions" json:"-"`
	Linkages     []linkage         `bson:"linkage" json:"linkage"`
	Priority     int               `bson:"priority" json:"priority"`
	CreatedOn    time.Time         `bson:"created_on" json:"created_on"`
	LastModified time.Time         `bson:"last_modified" json:"last_modified"`
	Expiration   time.Time         `bson:"expiration" json:"expiration"` // 0 means no expiration
	Type         string            `bson:"type" json:"type"`
	Subset       Subset            `bson:"subset" json:"-"`
	Parts        *PartsList        `bson:"parts" json:"parts"`
}

type linkage struct {
	Type      string   `bson:"relation" json:"relation"`
	Ids       []string `bson:"ids" json:"ids"`
	Operation string   `bson:"operation" json:"operation"`
}

type Indexes map[string]IdxInfo

type IdxInfo struct {
	Type        string    `bson:"index_type" json:"-"`
	TotalUnits  int64     `bson:"total_units" json:"total_units"`
	AvgUnitSize int64     `bson:"average_unit_size" json:"average_unit_size"`
	Format      string    `bson:"format" json:"-"`
	CreatedOn   time.Time `bson:"created_on" json:"created_on"`
}

type FormFiles map[string]FormFile

type FormFile struct {
	Name     string
	Path     string
	Checksum map[string]string
}

func (formfile *FormFile) Remove() {
	if _, err := os.Stat(formfile.Path); err == nil {
		os.Remove(formfile.Path)
	}
	return
}

func RemoveAllFormFiles(formfiles FormFiles) {
	for _, formfile := range formfiles {
		formfile.Remove()
	}
	return
}

// Subset is used to store information about a subset node's parent and its index.
// A subset node's index defines the subset of the data file that this node represents.
// A subset node's index is immutable after it is defined.
type Subset struct {
	Parent Parent            `bson:"parent" json:"-"`
	Index  SubsetNodeIdxInfo `bson:"index" json:"-"`
}

type Parent struct {
	Id        string `bson:"id" json:"-"`
	IndexName string `bson:"index_name" json:"-"`
}

type SubsetNodeIdxInfo struct {
	Path        string `bson:"path" json:"-"`
	TotalUnits  int64  `bson:"total_units" json:"-"`
	AvgUnitSize int64  `bson:"average_unit_size" json:"-"`
	Format      string `bson:"format" json:"-"`
}

func New() (node *Node) {
	node = new(Node)
	node.Indexes = make(map[string]IdxInfo)
	node.File.Checksum = make(map[string]string)
	node.setId()
	return
}

func LoadFromDisk(id string) (n *Node, err error) {
	if len(id) < 6 {
		return nil, errors.New("Node ID must be at least 6 characters in length")
	}
	path := getPath(id)
	if nbson, err := ioutil.ReadFile(path + "/" + id + ".bson"); err != nil {
		return nil, errors.New(e.NodeDoesNotExist)
	} else {
		n = new(Node)
		if err = bson.Unmarshal(nbson, &n); err != nil {
			return nil, err
		}
	}
	return
}

func CreateNodeUpload(u *user.User, params map[string]string, files FormFiles) (node *Node, err error) {
	// if copying node or creating subset node from parent, check if user has rights to the original node
	if _, hasCopyData := params["copy_data"]; hasCopyData {
		_, err = Load(params["copy_data"])
		if err != nil {
			return
		}
	}
	if _, hasParentNode := params["parent_node"]; hasParentNode {
		_, err = Load(params["parent_node"])
		if err != nil {
			return
		}
	}

	node = New()
	node.Type = "basic"

	node.Acl.SetOwner(u.Uuid)
	node.Acl.Set(u.Uuid, acl.Rights{"read": true, "write": true, "delete": true})

	err = node.Mkdir()
	if err != nil {
		return
	}

	err = node.Update(params, files)
	if err != nil {
		node.Rmdir()
		return
	}

	err = node.Save()
	return
}

func CreateNodesFromArchive(u *user.User, params map[string]string, files FormFiles, archiveId string) (nodes []*Node, err error) {
	// get parent node
	archiveNode, err := Load(archiveId)
	if err != nil {
		return nil, err
	}
	if archiveNode.File.Size == 0 {
		return nil, errors.New("parent archive node has no file")
	}

	// get format
	aFormat, hasFormat := params["archive_format"]
	if !hasFormat {
		return nil, errors.New("missing archive_format parameter. use one of: " + archive.ArchiveList)
	}
	if !archive.IsValidArchive(aFormat) {
		return nil, errors.New("invalid archive_format parameter. use one of: " + archive.ArchiveList)
	}

	// get attributes
	var atttributes interface{}
	if attrFile, ok := files["attributes"]; ok {
		defer attrFile.Remove()
		attr, err := ioutil.ReadFile(attrFile.Path)
		if err != nil {
			return nil, err
		}
		if err = json.Unmarshal(attr, &atttributes); err != nil {
			return nil, err
		}
	} else if attrStr, ok := params["attributes_str"]; ok {
		if err = json.Unmarshal([]byte(attrStr), &atttributes); err != nil {
			return nil, err
		}
	}

	// get files / delete unpack dir when done
	fileList, unpackDir, err := archive.FilesFromArchive(aFormat, archiveNode.FilePath())
	defer os.RemoveAll(unpackDir)
	if err != nil {
		return nil, err
	}

	// preserve acls
	_, preserveAcls := params["preserve_acls"]

	// build nodes
	var tempNodes []*Node
	for _, file := range fileList {
		// create link
		link := linkage{Type: "parent", Operation: aFormat, Ids: []string{archiveId}}
		// create and populate node
		node := New()
		node.Type = "basic"
		node.Linkages = append(node.Linkages, link)
		node.Attributes = atttributes

		if preserveAcls {
			// copy over acls from parent node
			node.Acl = archiveNode.Acl
		}
		// this user needs to be owner of new nodes
		node.Acl.SetOwner(u.Uuid)
		node.Acl.Set(u.Uuid, acl.Rights{"read": true, "write": true, "delete": true})

		if err = node.Mkdir(); err != nil {
			return nil, err
		}
		// set file
		f := FormFile{Name: file.Name, Path: file.Path, Checksum: file.Checksum}
		if err = node.SetFile(f); err != nil {
			node.Rmdir()
			return nil, err
		}
		tempNodes = append(tempNodes, node)
	}

	// save nodes, only return those that were created / saved
	for _, n := range tempNodes {
		if err = n.Save(); err != nil {
			n.Rmdir()
			return nil, err
		}
		nodes = append(nodes, n)
	}
	return
}

func (node *Node) FileReader() (reader file.ReaderAt, err error) {
	if node.File.Virtual {
		readers := []file.ReaderAt{}
		nodes := Nodes{}
		if _, err := dbFind(bson.M{"id": bson.M{"$in": node.File.VirtualParts}}, &nodes, "", nil); err != nil {
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

func (node *Node) DynamicIndex(name string) (idx index.Index, err error) {
	if index.Has(name) {
		idx = index.NewVirtual(name, node.FilePath(), node.File.Size, 10240)
	} else {
		if _, has := node.Indexes[name]; has {
			idx = index.New()
		} else {
			err_str := fmt.Sprintf("Node %s does not have index of type %s.", node.Id, name)
			err = errors.New(err_str)
		}
	}
	return
}

func (node *Node) Delete() (err error) {
	// check to make sure this node isn't referenced by a vnode
	virtualNodes := Nodes{}
	if _, err = dbFind(bson.M{"file.virtual_parts": node.Id}, &virtualNodes, "", nil); err != nil {
		return err
	}
	if len(virtualNodes) != 0 {
		return errors.New(e.NodeReferenced)
	}

	// Check to see if this node has a data file and if it's referenced by another node.
	// If it is, we will move the data file to the first node we find, and point all other nodes to that node's path
	dataFilePath := fmt.Sprintf("%s/%s.data", getPath(node.Id), node.Id)
	dataFileExists := true
	if _, ferr := os.Stat(dataFilePath); os.IsNotExist(ferr) {
		dataFileExists = false
	}
	newDataFilePath := ""
	copiedNodes := Nodes{}
	if _, err = dbFind(bson.M{"file.path": dataFilePath}, &copiedNodes, "", nil); err != nil {
		return err
	}
	if len(copiedNodes) != 0 && dataFileExists {
		for index, copiedNode := range copiedNodes {
			if index == 0 {
				newDataFilePath = fmt.Sprintf("%s/%s.data", getPath(copiedNode.Id), copiedNode.Id)
				if rerr := os.Rename(dataFilePath, newDataFilePath); rerr != nil {
					if _, cerr := util.CopyFile(dataFilePath, newDataFilePath); cerr != nil {
						return errors.New("This node has a data file linked to another node and the data file could not be copied elsewhere to allow for node deletion.")
					}
				}
				copiedNode.File.Path = ""
				copiedNode.Save()
			} else {
				copiedNode.File.Path = newDataFilePath
				copiedNode.Save()
			}
		}
	}

	if err = dbDelete(bson.M{"id": node.Id}); err != nil {
		return err
	}
	return node.Rmdir()
}

func (node *Node) DeleteIndex(indextype string) (err error) {
	delete(node.Indexes, indextype)
	IndexFilePath := fmt.Sprintf("%s/%s.idx", node.IndexPath(), indextype)
	if err = os.Remove(IndexFilePath); err != nil {
		return
	}
	err = node.Save()
	return
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

func (node *Node) SetPriority(priority int) (err error) {
	node.Priority = priority
	err = node.Save()
	return
}

func (node *Node) SetExpiration(expire string) (err error) {
	parts := ExpireRegex.FindStringSubmatch(expire)
	if len(parts) == 0 {
		return errors.New("expiration format is invalid")
	}
	var expireTime time.Duration
	expireNum, _ := strconv.Atoi(parts[1])
	currTime := time.Now()

	switch parts[2] {
	case "M":
		expireTime = time.Duration(expireNum) * time.Minute
	case "H":
		expireTime = time.Duration(expireNum) * time.Hour
	case "D":
		expireTime = time.Duration(expireNum*24) * time.Hour
	}

	node.Expiration = currTime.Add(expireTime)
	err = node.Save()
	return
}

func (node *Node) RemoveExpiration() (err error) {
	// reset to empty time
	node.Expiration = time.Time{}
	err = node.Save()
	return
}

func (node *Node) ClearRevisions() (err error) {
	// empty the revisions array
	node.Revisions = nil
	err = node.Save()
	return
}

func (node *Node) SetAttributes(attr FormFile) (err error) {
	defer attr.Remove()
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

func (node *Node) SetAttributesFromString(attributes string) (err error) {
	err = json.Unmarshal([]byte(attributes), &node.Attributes)
	if err != nil {
		return
	}
	err = node.Save()
	return
}

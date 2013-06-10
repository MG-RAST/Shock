package store

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/store/type/index"
	"github.com/MG-RAST/Shock/shock-server/store/type/index/virtual"
	"github.com/MG-RAST/Shock/shock-server/store/uuid"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	virtIdx = mappy{
		"size": true,
	}
)

type Node struct {
	Id           string            `bson:"id" json:"id"`
	Version      string            `bson:"version" json:"version"`
	File         file              `bson:"file" json:"file"`
	Attributes   interface{}       `bson:"attributes" json:"attributes"`
	Indexes      indexes           `bson:"indexes" json:"indexes"`
	Acl          Acl               `bson:"acl" json:"-"`
	VersionParts map[string]string `bson:"version_parts" json:"-"`
	Tags         []string          `bson:"tags" json:"tags"`
	Revisions    []Node            `bson:"revisions" json:"-"`
	Linkages     []linkage         `bson:"linkage" json:"linkages"`
}

type file struct {
	Name         string            `bson:"name" json:"name"`
	Size         int64             `bson:"size" json:"size"`
	Checksum     map[string]string `bson:"checksum" json:"checksum"`
	Format       string            `bson:"format" json:"format"`
	Path         string            `bson:"path" json:"-"`
	Virtual      bool              `bson:"virtual" json:"virtual"`
	VirtualParts []string          `bson:"virtual_parts" json:"virtual_parts"`
}

type partsList struct {
	Count  int         `json:"count"`
	Length int         `json:"length"`
	Parts  []partsFile `json:"parts"`
}

type linkage struct {
	Type      string   `bson: "relation" json:"relation"`
	Ids       []string `bson:"ids" json:"ids"`
	Operation string   `bson:"operation" json:"operation"`
}

type indexes map[string]IdxInfo

type IdxInfo struct {
	Type        string `bson:"index_type" json:"-"`
	TotalUnits  int64  `bson:"total_units" json:"total_units"`
	AvgUnitSize int64  `bson:"average_unit_size" json:"average_unit_size"`
}

type partsFile []string

type FormFiles map[string]FormFile

type FormFile struct {
	Name     string
	Path     string
	Checksum map[string]string
}

// HasFoo functions
func (node *Node) HasFile() bool {
	if node.File.Name == "" && node.File.Size == 0 && len(node.File.Checksum) == 0 && node.File.Path == "" {
		return false
	}
	return true
}

func (node *Node) HasIndex(index string) bool {
	if virtIdx[index] {
		return true
	} else {
		if node.HasFile() {
			if _, err := os.Stat(node.IndexPath() + "/" + index); err == nil {
				return true
			}
		}
	}
	return false
}

func (node *Node) HasParent() bool {
	for _, linkage := range node.Linkages {
		if linkage.Type == "parent" {
			return true
		}
	}
	return false
}

// Path functions
func (node *Node) Path() string {
	return getPath(node.Id)
}

func (node *Node) IndexPath() string {
	return getIndexPath(node.Id)
}

func (node *Node) FileReader() (reader ReaderAt, err error) {
	if node.File.Virtual {
		readers := []ReaderAt{}
		nodes := Nodes{}
		if db, err := DBConnect(); err == nil {
			defer db.Close()
			if _, err := db.Find(bson.M{"id": bson.M{"$in": node.File.VirtualParts}}, &nodes, nil); err != nil {
				return nil, err
			}
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
		return MultiReaderAt(readers...), nil
	}
	return os.Open(node.FilePath())
}

func (node *Node) FilePath() string {
	if node.File.Path != "" {
		return node.File.Path
	}
	return getPath(node.Id) + "/" + node.Id + ".data"
}

func (node *Node) FileExt() string {
	if node.File.Name != "" {
		return filepath.Ext(node.File.Name)
	}
	return ""
}

// Index functions
func (node *Node) Index(name string) (idx index.Index, err error) {
	if virtual.Has(name) {
		idx = virtual.New(name, node.FilePath(), node.File.Size, 10240)
	} else {
		idx = index.New()
		err = idx.Load(node.IndexPath() + "/" + name + ".idx")
	}
	return
}

// Parts functions
func (node *Node) loadParts() (p *partsList, err error) {
	pf, err := ioutil.ReadFile(node.partsListPath())
	if err != nil {
		return
	}
	p = &partsList{}
	err = json.Unmarshal(pf, &p)
	if err != nil {
		return
	}
	return
}

func (node *Node) writeParts(p *partsList) (err error) {
	pm, _ := json.Marshal(p)
	//os.Remove(node.partsListPath())
	err = ioutil.WriteFile(node.partsListPath(), []byte(pm), 0644)
	return
}

func (node *Node) partsCount() int {
	LockMgr.LockPartOp()
	p, err := node.loadParts()
	LockMgr.UnlockPartOp()
	if err != nil {
		return -1
	}
	return p.Count
}

func (node *Node) initParts(count int) (err error) {
	err = os.MkdirAll(fmt.Sprintf("%s/parts", node.Path()), 0777)
	p := &partsList{Count: count, Length: 0, Parts: make([]partsFile, count)}
	err = node.writeParts(p)
	return
}

func (node *Node) Delete() (err error) {
	// check to make sure this node isn't referenced by a vnode
	nodes := Nodes{}
	if db, err := DBConnect(); err == nil {
		defer db.Close()
		if _, err = db.Find(bson.M{"virtual_parts": node.Id}, &nodes, nil); err != nil {
			return err
		}
		if len(nodes) != 0 {
			return errors.New(e.NodeReferenced)
		} else {
			if err = db.Delete(bson.M{"id": node.Id}); err != nil {
				return err
			}
		}
	} else {
		return err
	}
	return node.Rmdir()
}

func (node *Node) addVirtualParts(ids []string) (err error) {
	nodes := Nodes{}
	if db, err := DBConnect(); err == nil {
		defer db.Close()
		if _, err := db.Find(bson.M{"id": bson.M{"$in": ids}}, &nodes, nil); err != nil {
			return err
		}
	} else {
		return err
	}
	if len(ids) != len(nodes) {
		return errors.New("unable to load all node ids.")
	}
	node.File.Virtual = true
	for _, n := range nodes {
		if n.HasFile() {
			node.File.VirtualParts = append(node.File.VirtualParts, n.Id)
		} else {
			return errors.New("node %s: has no file. All nodes in source must have files.")
		}
	}
	if reader, err := node.FileReader(); err == nil {
		md5h := md5.New()
		sha1h := sha1.New()
		buffer := make([]byte, 32*1024)
		size := 0
		for {
			n, err := reader.Read(buffer)
			if n == 0 || err != nil {
				break
			}
			md5h.Write(buffer[0:n])
			sha1h.Write(buffer[0:n])
			size = size + n
		}

		var md5s, sha1s []byte
		md5s = md5h.Sum(md5s)
		sha1s = sha1h.Sum(sha1s)
		node.File.Checksum["md5"] = fmt.Sprintf("%x", md5s)
		node.File.Checksum["sha1"] = fmt.Sprintf("%x", sha1s)
		node.File.Size = int64(size)
	} else {
		return err
	}
	err = node.Save()
	return
}

func (node *Node) addPart(n int, file *FormFile) (err error) {

	LockMgr.LockPartOp()
	defer LockMgr.UnlockPartOp()

	// load
	p, err := node.loadParts()
	if err != nil {
		return
	}

	// modify
	if len(p.Parts[n]) > 0 {
		err = errors.New(e.FileImut)
		return
	}
	part := partsFile{file.Name, file.Checksum["md5"]}
	p.Parts[n] = part
	p.Length = p.Length + 1

	err = os.Rename(file.Path, fmt.Sprintf("%s/parts/%d", node.Path(), n))
	if err != nil {
		return
	}

	// rewrite
	err = node.writeParts(p)
	if err != nil {
		return
	}

	// create file if done
	if p.Length == p.Count {
		err = node.SetFileFromParts(p)
		if err != nil {
			return
		}
	}
	return
}

func (node *Node) partsListPath() string {
	return node.Path() + "/parts/parts.json"
}

func (node *Node) SetFileFromPath(path string) (err error) {
	fileStat, err := os.Stat(path)
	if err != nil {
		return
	}
	node.File.Name = fileStat.Name()
	node.File.Size = fileStat.Size()
	node.File.Path = path

	md5h := md5.New()
	sha1h := sha1.New()
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	for {
		buffer := make([]byte, 10240)
		n, err := f.Read(buffer)
		if n == 0 || err != nil {
			break
		}
		md5h.Write(buffer[0:n])
		sha1h.Write(buffer[0:n])
	}
	node.File.Checksum["md5"] = fmt.Sprintf("%x", md5h.Sum(nil))
	node.File.Checksum["sha1"] = fmt.Sprintf("%x", sha1h.Sum(nil))
	err = node.Save()
	return
}

func (node *Node) SetFileFromParts(p *partsList) (err error) {
	out, err := os.Create(fmt.Sprintf("%s/%s.data", node.Path(), node.Id))
	if err != nil {
		return
	}
	defer out.Close()
	md5h := md5.New()
	sha1h := sha1.New()
	for i := 0; i < p.Count; i++ {
		part, err := os.Open(fmt.Sprintf("%s/parts/%d", node.Path(), i))
		if err != nil {
			return err
		}
		for {
			buffer := make([]byte, 10240)
			n, err := part.Read(buffer)
			if n == 0 || err != nil {
				break
			}
			out.Write(buffer[0:n])
			md5h.Write(buffer[0:n])
			sha1h.Write(buffer[0:n])
		}
		part.Close()
	}
	fileStat, err := os.Stat(fmt.Sprintf("%s/%s.data", node.Path(), node.Id))
	if err != nil {
		return
	}
	node.File.Name = node.Id
	node.File.Size = fileStat.Size()
	node.File.Checksum["md5"] = fmt.Sprintf("%x", md5h.Sum(nil))
	node.File.Checksum["sha1"] = fmt.Sprintf("%x", sha1h.Sum(nil))
	err = node.Save()
	return
}

//Modification functions
func (node *Node) Update(params map[string]string, files FormFiles) (err error) {
	// Exclusive conditions
	// 1. has files[upload] (regular upload)
	// 2. has params[parts] (partial upload support)
	// 3. has params[type] & params[source] (v_node)
	// 4. has params[path] (set from local path)
	//
	// All condition allow setting of attributes

	_, isRegularUpload := files["upload"]
	_, isPartialUpload := params["parts"]
	isVirtualNode := false
	if t, hasType := params["type"]; hasType && t == "virtual" {
		isVirtualNode = true
	}
	_, isPathUpload := params["path"]

	// Check exclusive conditions
	if (isRegularUpload && isPartialUpload) || (isRegularUpload && isVirtualNode) || (isRegularUpload && isPathUpload) {
		return errors.New("upload parameter incompatible with parts, path and/or type parmeter(s)")
	} else if (isPartialUpload && isVirtualNode) || (isPartialUpload && isPathUpload) {
		return errors.New("parts parameter incompatible with type and/or path parmeter(s)")
	} else if isVirtualNode && isPathUpload {
		return errors.New("type parameter incompatible with path parmeter")
	}

	// Check if immutiable
	if (isRegularUpload || isPartialUpload || isVirtualNode || isPathUpload) && node.HasFile() {
		return errors.New(e.FileImut)
	}

	if isRegularUpload {
		if err = node.SetFile(files["upload"]); err != nil {
			return err
		}
		delete(files, "upload")
	} else if isPartialUpload {
		if node.partsCount() > 0 {
			return errors.New("parts already set")
		}
		n, err := strconv.Atoi(params["parts"])
		if err != nil {
			return err
		}
		if n < 1 {
			return errors.New("parts cannot be less than 1")
		}
		if err = node.initParts(n); err != nil {
			return err
		}
	} else if isVirtualNode {
		if source, hasSource := params["source"]; hasSource {
			ids := strings.Split(source, ",")
			node.addVirtualParts(ids)
		} else {
			return errors.New("type virtual requires source parameter")
		}
	} else if isPathUpload {
		if len(conf.LOCAL_PATHS) > 0 {
			for _, p := range conf.LOCAL_PATHS {
				if strings.HasPrefix(params["path"], p) {
					if err = node.SetFileFromPath(params["path"]); err != nil {
						return err
					} else {
						return nil
					}
				}
			}
			return errors.New("file not in local files path. Please contact your Shock administrator.")
		} else {
			return errors.New("local files path uploads must be configured. Please contact your Shock administrator.")
		}
	}

	// set attributes from file
	if _, hasAttr := files["attributes"]; hasAttr {
		if err = node.SetAttributes(files["attributes"]); err != nil {
			return err
		}
		os.Remove(files["attributes"].Path)
		delete(files, "attributes")
	}

	// handle part file

	parts_count := node.partsCount()
	if parts_count > 1 {
		for key, file := range files {
			if node.HasFile() {
				return errors.New(e.FileImut)
			}
			keyn, errf := strconv.Atoi(key)
			if errf == nil && keyn <= parts_count {
				err = node.addPart(keyn-1, &file)
				if err != nil {
					return
				}
			}
		}
	}

	// update relatives
	if _, hasRelation := params["linkage"]; hasRelation {
		ltype := params["linkage"]

		if ltype == "parent" {
			if node.HasParent() {
				return errors.New(e.ProvenanceImut)
			}
		}
		var ids string
		if _, hasIds := params["ids"]; hasIds {
			ids = params["ids"]
		} else {
			return errors.New("missing ids for updating relativs")
		}
		var operation string
		if _, hasOp := params["operation"]; hasOp {
			operation = params["operation"]
		}
		if err = node.UpdateLinkages(ltype, ids, operation); err != nil {
			return err
		}
	}

	//update node tags
	if _, hasDataType := params["tags"]; hasDataType {
		if err = node.UpdateDataTags(params["tags"]); err != nil {
			return err
		}
	}

	//update file format
	if _, hasFormat := params["format"]; hasFormat {
		if node.File.Format != "" {
			return errors.New(fmt.Sprintf("file format already set:%s", node.File.Format))
		}
		if err = node.SetFileFormat(params["format"]); err != nil {
			return err
		}
	}

	return
}

func (node *Node) Save() (err error) {
	db, err := DBConnect()
	if err != nil {
		return
	}
	defer db.Close()
	node.UpdateVersion()
	if len(node.Revisions) == 0 || node.Revisions[len(node.Revisions)-1].Version != node.Version {
		n := Node{node.Id, node.Version, node.File, node.Attributes, node.Indexes, node.Acl, node.VersionParts, node.Tags, nil, node.Linkages}
		node.Revisions = append(node.Revisions, n)
	}
	bsonPath := fmt.Sprintf("%s/%s.bson", node.Path(), node.Id)
	os.Remove(bsonPath)
	nbson, err := bson.Marshal(node)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(bsonPath, nbson, 0644)
	if err != nil {
		return
	}
	err = db.Upsert(node)
	if err != nil {
		return
	}
	return
}

func (node *Node) Rmdir() (err error) {
	return os.RemoveAll(node.Path())
}

func (node *Node) Mkdir() (err error) {
	err = os.MkdirAll(node.Path(), 0777)
	if err != nil {
		return
	}
	err = os.MkdirAll(node.IndexPath(), 0777)
	if err != nil {
		return
	}
	return
}

func (node *Node) UpdateVersion() (err error) {
	parts := make(map[string]string)
	h := md5.New()
	version := node.Id
	for name, value := range map[string]interface{}{"file_ver": node.File, "attributes_ver": node.Attributes, "acl_ver": node.Acl} {
		m, er := json.Marshal(value)
		if er != nil {
			return
		}
		h.Write(m)
		sum := fmt.Sprintf("%x", h.Sum(nil))
		parts[name] = sum
		version = version + ":" + sum
		h.Reset()
	}
	h.Write([]byte(version))
	node.Version = fmt.Sprintf("%x", h.Sum(nil))
	node.VersionParts = parts
	return
}

func (node *Node) setId() {
	node.Id = uuid.New()
	return
}

func (node *Node) SetFile(file FormFile) (err error) {
	fileStat, err := os.Stat(file.Path)
	if err != nil {
		return
	}
	os.Rename(file.Path, node.FilePath())
	node.File.Name = file.Name
	node.File.Size = fileStat.Size()
	node.File.Checksum = file.Checksum

	//fill size index info
	totalunits := node.File.Size / conf.CHUNK_SIZE
	m := node.File.Size % conf.CHUNK_SIZE
	if m != 0 {
		totalunits += 1
	}
	node.Indexes["size"] = IdxInfo{
		Type:        "size",
		TotalUnits:  totalunits,
		AvgUnitSize: conf.CHUNK_SIZE,
	}

	err = node.Save()
	return
}

func (node *Node) SetIndexInfo(indextype string, idxinfo IdxInfo) (err error) {
	node.Indexes[indextype] = idxinfo
	err = node.Save()
	return
}

func (node *Node) UpdateLinkages(ltype string, ids string, operation string) (err error) {
	var link linkage
	link.Type = ltype
	idList := strings.Split(ids, ",")
	for _, id := range idList {
		link.Ids = append(link.Ids, id)
	}
	link.Operation = operation
	node.Linkages = append(node.Linkages, link)
	err = node.Save()
	return
}

func (node *Node) UpdateDataTags(types string) (err error) {
	tagslist := strings.Split(types, ",")
	for _, newtag := range tagslist {
		if contains(node.Tags, newtag) {
			continue
		}
		node.Tags = append(node.Tags, newtag)
	}
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

// Other
func (node *Node) ToJson() (s string, err error) {
	m, err := json.Marshal(node)
	s = string(m)
	return
}

func contains(list []string, elem string) bool {
	for _, t := range list {
		if t == elem {
			return true
		}
	}
	return false
}

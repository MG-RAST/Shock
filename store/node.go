package store

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MG-RAST/Shock/store/type/index"
	"github.com/MG-RAST/Shock/store/type/index/virtual"
	"io/ioutil"
	"launchpad.net/mgo/bson"
	"math/rand"
	"os"
	"strconv"
	"time"
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
	Indexes      map[string]string `bson:"indexes" json:"indexes"`
	Acl          acl               `bson:"acl" json:"acl"`
	VersionParts map[string]string `bson:"version_parts" json:"version_parts"`
}

type file struct {
	Name     string            `bson:"name" json:"name"`
	Size     int64             `bson:"size" json:"size"`
	Checksum map[string]string `bson:"checksum" json:"checksum"`
}

type partsList struct {
	Count  int         `json:"count"`
	Length int         `json:"length"`
	Parts  []partsFile `json:"parts"`
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
	if node.File.Name == "" && node.File.Size == 0 && len(node.File.Checksum) == 0 {
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

// Path functions
func (node *Node) Path() string {
	return getPath(node.Id)
}

func (node *Node) IndexPath() string {
	return getIndexPath(node.Id)
}

func (node *Node) DataPath() string {
	return getPath(node.Id) + "/" + node.Id + ".data"
}

// Index functions
func (node *Node) Index(name string) (idx index.Index, err error) {
	if virtual.Has(name) {
		idx = virtual.New(name, node.DataPath(), node.File.Size, 10240)
	} else {
		idx = index.New()
		err = idx.Load(node.IndexPath() + "/" + name)
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
	os.Remove(node.partsListPath())
	err = ioutil.WriteFile(node.partsListPath(), []byte(pm), 0644)
	return
}

func (node *Node) partsCount() int {
	p, err := node.loadParts()
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

func (node *Node) addPart(n int, file *FormFile) (err error) {
	// load
	p, err := node.loadParts()
	if err != nil {
		return
	}

	// modify
	if len(p.Parts[n]) > 0 {
		err = errors.New("node part already exists and is immutable")
		return
	}
	part := partsFile{file.Name, file.Checksum["md5"]}
	p.Parts[n] = part
	p.Length = p.Length + 1
	os.Rename(file.Path, fmt.Sprintf("%s/parts/%d", node.Path(), n))

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

	var md5s, sha1s []byte
	md5h.Sum(md5s)
	node.File.Checksum["md5"] = fmt.Sprintf("%x", md5s)
	sha1h.Sum(sha1s)
	node.File.Checksum["sha1"] = fmt.Sprintf("%x", sha1s)
	err = node.Save()
	return
}

//Modification functions
func (node *Node) Update(params map[string]string, files FormFiles) (err error) {
	_, hasParts := params["parts"]
	if hasParts && node.partsCount() < 0 {
		if node.HasFile() {
			return errors.New("file alreay set and is immutable")
		}
		n, err := strconv.Atoi(params["parts"])
		if err != nil {
			return err
		}
		if n < 1 {
			return errors.New("parts cannot be less than 1")
		}
		err = node.initParts(n)
		if err != nil {
			return err
		}
	}
	_, hasFile := files["file"]
	if hasFile && !node.HasFile() {
		err = node.SetFile(files["file"])
		if err != nil {
			return err
		}
		delete(files, "file")
	} else if hasFile {
		return errors.New("node file immutable")
	}
	_, hasAttr := files["attributes"]
	if hasAttr && node.Attributes == nil {
		err = node.SetAttributes(files["attributes"])
		if err != nil {
			return err
		}
		os.Remove(files["attributes"].Path)
		delete(files, "attributes")
	} else if hasAttr {
		return errors.New("node attributes immutable")
	}
	pc := node.partsCount()
	if pc > 1 {
		for key, file := range files {
			if node.HasFile() {
				return errors.New("node file already set and is immutable")
			}
			keyn, errf := strconv.Atoi(key)
			if errf == nil && keyn <= pc {
				err = node.addPart(keyn-1, &file)
				if err != nil {
					return
				}
			}
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
	var fsum, attrsum, aclsum, versum []byte
	parts := make(map[string]string)
	h := md5.New()

	// checksum file
	m, err := json.Marshal(node.File)
	if err != nil {
		return
	}
	h.Write(m)
	fsum = h.Sum(fsum)
	parts["file_ver"] = fmt.Sprintf("%x", fsum)
	h.Reset()

	// checksum attributes
	m, err = json.Marshal(node.Attributes)
	if err != nil {
		return
	}
	h.Write(m)
	attrsum = h.Sum(attrsum)
	parts["attributes_ver"] = fmt.Sprintf("%x", attrsum)
	h.Reset()

	// checksum acl
	m, err = json.Marshal(node.Acl)
	if err != nil {
		return
	}
	h.Write(m)
	aclsum = h.Sum(aclsum)
	parts["acl_ver"] = fmt.Sprintf("%x", aclsum)
	h.Reset()

	// node version
	h.Write([]byte(fmt.Sprintf("%s:%s:%s:%s", node.Id, parts["file_ver"], parts["attributes_ver"], parts["acl_ver"])))
	versum = h.Sum(versum)
	node.Version = fmt.Sprintf("%x", versum)
	node.VersionParts = parts
	return
}

func (node *Node) setId() {
	var s []byte
	h := md5.New()
	h.Write([]byte(fmt.Sprint(time.Now().String(), rand.Float64())))
	s = h.Sum(s)
	node.Id = fmt.Sprintf("%x", s)
	/*
		id, _ := uuid.NewV5(uuid.NamespaceURL, []byte("shock"))	
		node.Id = id.String()
	*/
	return
}

func (node *Node) SetFile(file FormFile) (err error) {
	fileStat, err := os.Stat(file.Path)
	if err != nil {
		return
	}
	os.Rename(file.Path, node.DataPath())
	node.File.Name = file.Name
	node.File.Size = fileStat.Size()
	node.File.Checksum = file.Checksum
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

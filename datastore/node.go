package datastore

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	e "github.com/MG-RAST/Shock/errors"
	"io/ioutil"
	"launchpad.net/mgo/bson"
	"math/rand"
	"os"
	"strconv"
	"time"
)

type Node struct {
	Id         string            `bson:"id" json:"id"`
	File       nodeFile          `bson:"file" json:"file"`
	Attributes interface{}       `bson:"attributes" json:"attributes"`
	Indexes    map[string]string `bson:"indexes" json:"indexes"`
	Acl        acl               `bson:"acl" json:"acl"`
}

type nodeFile struct {
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

func (n *nodeFile) Empty() bool {
	if n.Name == "" && n.Size == 0 && len(n.Checksum) == 0 {
		return true
	}
	return false
}

func (n *nodeFile) SizeIndex(chunkSize int64) (idx *Index) {
	var i int64
	idx = NewIndex()
	idx.Type = "size"
	idx.CType = "none"
	idx.Version = 1
	for i = 0; i < n.Size; i += chunkSize {
		idx.Idx = append(idx.Idx, Record{i, (i + chunkSize)})
	}
	return
}

func (n *nodeFile) SizeOffset(part int64, chunkSize int64) (position int64, length int64, err error) {
	if part < 1 {
		err = errors.New(e.InvalidIndex)
	} else if n.Size > ((part - 1) * chunkSize) {
		position = ((part - 1) * chunkSize)
		if n.Size < (part * chunkSize) {
			length = n.Size - position
		} else {
			length = chunkSize
		}
	} else {
		if part == int64(1) {
			position = 0
			length = n.Size
		} else {
			err = errors.New(e.InvalidIndex)
		}
	}
	return
}

func (node *Node) Update(params map[string]string, files FormFiles) (err error) {
	_, hasParts := params["parts"]
	if hasParts && node.partsCount() < 0 {
		if !node.File.Empty() {
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
	if hasFile && node.File.Empty() {
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
			if !node.File.Empty() {
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

func (node *Node) Path() string {
	return getPath(node.Id)
}

func (node *Node) DataPath() string {
	return fmt.Sprintf("%s/%s.data", getPath(node.Id), node.Id)
}

func (node *Node) Mkdir() (err error) {
	err = os.MkdirAll(node.Path(), 0777)
	return
}

func (node *Node) ToJson() (s string, err error) {
	m, err := json.Marshal(node)
	s = string(m)
	return
}

func (node *Node) Save() (err error) {
	//jsonPath := fmt.Sprintf("%s/%s.json", node.Path(), node.Id)
	//os.Remove(jsonPath)
	//n, err := node.ToJson(); if err != nil { return }	
	//err = ioutil.WriteFile(jsonPath, []byte(n), 0644); if err != nil { return }

	db, err := DBConnect()
	if err != nil {
		return
	}
	defer db.Close()

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

func (node *Node) partsListPath() string {
	return fmt.Sprintf("%s/parts/parts.json", node.Path())
}

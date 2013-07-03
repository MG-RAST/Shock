package node

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	"os"
)

type partsFile []string

type partsList struct {
	Count  int         `json:"count"`
	Length int         `json:"length"`
	Parts  []partsFile `json:"parts"`
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

func (node *Node) addVirtualParts(ids []string) (err error) {
	nodes := Nodes{}
	if _, err := dbFind(bson.M{"id": bson.M{"$in": ids}}, &nodes, nil); err != nil {
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
	// load
	p, err := node.loadParts()
	if err != nil {
		return err
	}

	// modify
	if len(p.Parts[n]) > 0 {
		return errors.New(e.FileImut)
	}
	part := partsFile{file.Name, file.Checksum["md5"]}
	p.Parts[n] = part
	p.Length = p.Length + 1

	if err = os.Rename(file.Path, fmt.Sprintf("%s/parts/%d", node.Path(), n)); err != nil {
		return err
	}

	// rewrite
	if err = node.writeParts(p); err != nil {
		return err
	}

	// create file if done
	if p.Length == p.Count {
		if err = node.SetFileFromParts(p); err != nil {
			return err
		}
	}
	return
}

func (node *Node) partsListPath() string {
	return node.Path() + "/parts/parts.json"
}

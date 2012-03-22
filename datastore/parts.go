package datastore

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

// PartsList struct
type partsList struct {
	Count  int         `json:"count"`
	Length int         `json:"length"`
	Parts  []partsFile `json:"parts"`
}

type partsFile []string

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

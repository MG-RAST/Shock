package node

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"io/ioutil"
	"github.com/MG-RAST/golib/mgo/bson"
	"os"
	"strconv"
)

type partsFile []string

type partsList struct {
	Count  int         `json:"count"`
	Length int         `json:"length"`
	VarLen bool        `json:"varlen"`
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

func (node *Node) isVarLen() bool {
	p, err := node.loadParts()
	if err != nil {
		return false
	}
	return p.VarLen
}

func (node *Node) initParts(partsCount string) (err error) {
	// Function should only be called with a postive integer or string 'unknown'
	count, cerr := strconv.Atoi(partsCount)
	if partsCount != "unknown" && cerr != nil {
		return cerr
	}
	p := &partsList{}
	err = os.MkdirAll(fmt.Sprintf("%s/parts", node.Path()), 0777)
	if partsCount == "unknown" {
		p = &partsList{Count: 0, Length: 0, VarLen: true, Parts: make([]partsFile, 0)}
	} else {
		p = &partsList{Count: count, Length: 0, VarLen: false, Parts: make([]partsFile, count)}
	}
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
		defer reader.Close()
		md5h := md5.New()
		buffer := make([]byte, 32*1024)
		size := 0
		for {
			n, err := reader.Read(buffer)
			if n == 0 || err != nil {
				break
			}
			md5h.Write(buffer[0:n])
			size = size + n
		}

		var md5s []byte
		md5s = md5h.Sum(md5s)
		node.File.Checksum["md5"] = fmt.Sprintf("%x", md5s)
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

	if n >= p.Count && !p.VarLen {
		return errors.New("part number is greater than node length: " + strconv.Itoa(p.Length))
	}

	if n < p.Count && len(p.Parts[n]) > 0 {
		return errors.New(e.FileImut)
	}

	// create part
	part := partsFile{file.Name, file.Checksum["md5"]}

	// add part to node
	if p.VarLen == true && n >= p.Count {
		for i := p.Count; i < n; i = i + 1 {
			p.Parts = append(p.Parts, partsFile{})
			p.Count = p.Count + 1
		}
		p.Parts = append(p.Parts, part)
		p.Count = p.Count + 1
		p.Length = p.Length + 1
	} else {
		p.Parts[n] = part
		p.Length = p.Length + 1
	}

	// put part into data directory
	if err = os.Rename(file.Path, fmt.Sprintf("%s/parts/%d", node.Path(), n+1)); err != nil {
		return err
	}

	// rewrite
	if err = node.writeParts(p); err != nil {
		return err
	}

	// create file if done with non-variable length node
	if !p.VarLen && p.Length == p.Count {
		if err = node.SetFileFromParts(p, false); err != nil {
			return err
		}
		if err = os.RemoveAll(node.Path() + "/parts/"); err != nil {
			return err
		}
	}
	return
}

func (node *Node) closeVarLenPartial() (err error) {
	p, err := node.loadParts()
	if err != nil {
		return err
	}

	// Second param says we will allow empty parts in merging of those parts
	if err = node.SetFileFromParts(p, true); err != nil {
		return err
	}
	if err = os.RemoveAll(node.Path() + "/parts/"); err != nil {
		return err
	}
	return
}

func (node *Node) partsListPath() string {
	return node.Path() + "/parts/parts.json"
}

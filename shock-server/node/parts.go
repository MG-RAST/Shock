package node

import (
	"crypto/md5"
	"errors"
	"fmt"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/vendor/gopkg.in/mgo.v2/bson"
	"io"
	"os"
	"strconv"
)

type partsFile []string

type PartsList struct {
	Count       int         `bson:"count" json:"count"`
	Length      int         `bson:"length" json:"length"`
	VarLen      bool        `bson:"varlen" json:"varlen"`
	Parts       []partsFile `bson:"parts" json:"parts"`
	Compression string      `bson:"compression" json:"compression"`
}

func (node *Node) initParts(partsCount string, compressionFormat string) (err error) {
	// Function should only be called with a postive integer or string 'unknown'
	count, cerr := strconv.Atoi(partsCount)
	if partsCount != "unknown" && cerr != nil {
		return cerr
	}
	if err = os.MkdirAll(fmt.Sprintf("%s/parts", node.Path()), 0777); err != nil {
		return err
	}

	varlen := false
	if partsCount == "unknown" {
		count = 0
		varlen = true
	}

	node.Type = "parts"
	node.Parts = &PartsList{
		Count:       count,
		Length:      0,
		VarLen:      varlen,
		Parts:       make([]partsFile, count),
		Compression: compressionFormat,
	}
	if err = node.Save(); err != nil {
		return err
	}

	// add node id to LockMgr
	LockMgr.AddNode(node.Id)
	return
}

func (node *Node) addVirtualParts(ids []string) (err error) {
	nodes := Nodes{}
	if _, err := dbFind(bson.M{"id": bson.M{"$in": ids}}, &nodes, "", nil); err != nil {
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
		n, err := io.Copy(md5h, reader)
		if err != nil {
			return err
		}
		node.File.Checksum["md5"] = fmt.Sprintf("%x", md5h.Sum(nil))
		node.File.Size = n
	} else {
		return err
	}
	err = node.Save()
	return
}

func (node *Node) addPart(n int, file *FormFile) (err error) {
	if n >= node.Parts.Count && !node.Parts.VarLen {
		return errors.New("part number is greater than node length: " + strconv.Itoa(node.Parts.Length))
	}

	if n < node.Parts.Count && len(node.Parts.Parts[n]) > 0 {
		return errors.New(e.FileImut)
	}

	// create part
	part := partsFile{file.Name, file.Checksum["md5"]}

	// add part to node
	if node.Parts.VarLen == true && n >= node.Parts.Count {
		for i := node.Parts.Count; i < n; i = i + 1 {
			node.Parts.Parts = append(node.Parts.Parts, partsFile{})
			node.Parts.Count = node.Parts.Count + 1
		}
		node.Parts.Parts = append(node.Parts.Parts, part)
		node.Parts.Count = node.Parts.Count + 1
		node.Parts.Length = node.Parts.Length + 1
	} else {
		node.Parts.Parts[n] = part
		node.Parts.Length = node.Parts.Length + 1
	}

	// put part into data directory
	if err = os.Rename(file.Path, fmt.Sprintf("%s/parts/%d", node.Path(), n+1)); err != nil {
		return err
	}
	err = node.Save()
	return
}

func (node *Node) closeParts(allowEmpty bool) (err error) {
	// Second param says we will allow empty parts in merging of those parts
	// true for variable length parts
	if err = node.SetFileFromParts(allowEmpty); err != nil {
		return err
	}
	if err = os.RemoveAll(node.Path() + "/parts/"); err != nil {
		return err
	}
	// remove node id from LockMgr
	LockMgr.RemoveNode(node.Id)
	return
}

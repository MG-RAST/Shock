package datastore

import (
	"fmt"
	"time"
	"crypto/md5"
	"crypto/sha256"
	"rand"
	"os"
	"json"
	"path/filepath"
	"io/ioutil"
)

type Node struct {
	Id         string            `json:"id"`
	File       Nodefile          `json:"file"`
	Attributes interface{}       `json:"attributes"`
	Indexes    map[string]string `json:"indexes"`
	Acl        ACL               `json:"acl"`
}

type Nodefile struct {
	Name     string            `json:"name"`
	Size     int64             `json:"size"`
	Checksum map[string]string `json:"checksum"`
}

type ACL struct {
	Read   []string `json:"read"`
	Write  []string `json:"write"`
	Delete []string `json:"delete"`
}

func LoadNode(id string) (node *Node, err os.Error) {
	path := getPath(id)
	var nJson []byte
	nJson, err = ioutil.ReadFile(fmt.Sprintf("%s/%s.json", path, id))
	if err != nil {
		return
	}
	node = new(Node)
	err = json.Unmarshal(nJson, &node)
	if err != nil {
		return
	}
	return
}

func CreateNode(filePath string, attrPath string) (node *Node, err os.Error) {
	var (
		fileStat *os.FileInfo
		attrStat *os.FileInfo
		in       *os.File
		out      *os.File
	)
	node = new(Node)
	node.Indexes = make(map[string]string)
	node.File.Checksum = make(map[string]string)
	node.setId()

	err = node.Mkdir()
	if err != nil {
		return
	}
	if filePath != "" {
		fileStat, err = os.Stat(filePath)
		if err != nil {
			return
		}
		if fileStat.IsDirectory() {
			err = os.NewError("directory found: wft?")
			return
		}
		var bytesRead int = 1
		md5h := md5.New()
		sha256h := sha256.New()
		in, err = os.Open(filePath)
		if err != nil {
			return
		}
		defer in.Close()
		out, err = os.Create(fmt.Sprintf("%s/%s.data", node.Path(), node.Id))
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
			sha256h.Write(buffer[0:bytesRead])
			out.Write(buffer[0:bytesRead])
		}
		// set file struct
		node.File.Name = filepath.Base(filePath)
		node.File.Size = fileStat.Size
		node.File.Checksum["md5"] = fmt.Sprintf("%x", md5h.Sum())
		node.File.Checksum["sha256"] = fmt.Sprintf("%x", sha256h.Sum())
	}
	if attrPath != "" {
		attrStat, err = os.Stat(attrPath)
		if err != nil {
			return
		}
		if attrStat.IsDirectory() {
			err = os.NewError("directory found: wft?")
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
	return
}

func (node *Node) setId() {
	h := md5.New()
	h.Write([]byte(fmt.Sprint(time.LocalTime().String(), rand.Float64())))
	node.Id = fmt.Sprintf("%x", h.Sum())
	return
}

func getPath(id string) (path string) {
	DATAROOT := "/Users/jared/projects/GoShockData"
	path = fmt.Sprintf("%s/%s/%s/%s/%s", DATAROOT, id[0:2], id[2:4], id[4:6], id)
	return
}

func (node *Node) Path() (path string) {
	path = getPath(node.Id)
	return
}

func (node *Node) Mkdir() (err os.Error) {
	err = os.MkdirAll(node.Path(), 0777)
	return
}

func (node *Node) ToJson() (s string, err os.Error) {
	m, err := json.Marshal(node)
	s = string(m)
	return
}

func (node *Node) Save() (err os.Error) {
	var n string
	jsonPath := fmt.Sprintf("%s/%s.json", node.Path(), node.Id)
	os.Remove(jsonPath)
	n, err = node.ToJson()
	if err != nil {
		return
	}
	err = ioutil.WriteFile(jsonPath, []byte(n), 0644)
	if err != nil {
		return
	}
	return
}

package datastore

import (
	"fmt"
	//"time"	
	//"crypto/md5"
	//"crypto/sha1"
	//"path/filepath"
	//"math/rand"
	"os"
	"errors"
	"encoding/json"	
	"io/ioutil"
	"strconv"
	uuid "github.com/nu7hatch/gouuid"
)

// Node struct
type Node struct {
	Id         string            `json:"id"`
	File       nodeFile          `json:"file"`	
	Attributes interface{}       `json:"attributes"`
	Indexes    map[string]string `json:"indexes"`
	Acl        acl               `json:"acl"`
}

// Node.nodefile struct
type nodeFile struct {
	Name     string            `json:"name"`
	Size     int64             `json:"size"`
	Checksum map[string]string `json:"checksum"`
}

// Node.acl struct
type acl struct {
	Read   []string `json:"read"`
	Write  []string `json:"write"`
	Delete []string `json:"delete"`
}

// FormFiles struct
type FormFiles map[string]FormFile

type FormFile struct {
	Name string
	Path string
	Checksum map[string]string	
}

func (n *nodeFile) Empty() (bool){
	if n.Name == "" && n.Size == 0 && len(n.Checksum) == 0 {
		return true
	}
	return false
}

func LoadNode(id string) (node *Node, err error) {
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

/*
func CreateNode(filePath string, attrPath string) (node *Node, error) {
	var (
		attrStat *os.FileInfo
		in       *os.File
		out      *os.File
	)
	node = new(Node)
	node.Indexes = make(map[string]string)
	node.File.Checksum = make(map[string]string)
	node.setId()

	err = node.Mkdir(); if err != nil {	return }
	if filePath != "" {
		fileStat, err := os.Stat(filePath); if err != nil {	return }

		if fileStat.IsDirectory() {
			err = errors.New("directory found: wft?")
			return
		}
		var bytesRead int = 1
		md5h := md5.New()
		sha1h := sha1.New()
		in, err = os.Open(filePath)
		if err != nil {
			return
		}
		defer in.Close()
		out, err = os.Create(node.DataPath())
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
			sha1h.Write(buffer[0:bytesRead])
			out.Write(buffer[0:bytesRead])
		}
		// set file struct
		node.File.Name = filepath.Base(filePath)
		node.File.Size = fileStat.Size
		node.File.Checksum["md5"] = fmt.Sprintf("%x", md5h.Sum())
		node.File.Checksum["sha1"] = fmt.Sprintf("%x", sha1h.Sum())
	}
	if attrPath != "" {
		attrStat, err = os.Stat(attrPath)
		if err != nil {
			return
		}
		if attrStat.IsDirectory() {
			err = errors.New("directory found: wft?")
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
	err = node.Save()
	return
}
*/

func CreateNodeUpload(params map[string]string, files FormFiles) (node *Node, err error) {
	node = new(Node)
	node.Indexes = make(map[string]string)
	node.File.Checksum = make(map[string]string)
	node.setId()

	err = node.Mkdir(); if err != nil { return }
	err = node.Update(params, files); if err != nil { return }
	err = node.Save()
	return
}

func (node *Node) Update(params map[string]string, files FormFiles) (err error){	
	_, hasParts := params["parts"]
	if hasParts && node.partsCount() < 0 {
		if !node.File.Empty() { return errors.New("file alreay set and is immutable") }		
		n, err := strconv.Atoi(params["parts"]); if err != nil {return err}
		if n < 1 {
			return errors.New("parts cannot be less than 1")
		}
		err = node.initParts(n); if err != nil {return err}
	}
	
	_, hasFile := files["file"]
	if hasFile && node.File.Empty() {
		err = node.SetFile(files["file"]); if err != nil {return err}
		delete(files, "file")
	} else if hasFile {
		return errors.New("node file immutable") 
	}
	_, hasAttr := files["attributes"]
	if hasAttr && node.Attributes == nil  {	
		err = node.SetAttributes(files["attributes"]); if err != nil {return err}
		os.Remove(files["attributes"].Path)
		delete(files, "attributes")		
	} else if hasAttr {
		return errors.New("node attributes immutable")
	}	
	
	pc := node.partsCount()
	if pc > 1 {
		for key, file := range files {
			if !node.File.Empty() { return errors.New("node file already set and is immutable") }
			keyn, errf := strconv.Atoi(key)
			if errf == nil && keyn <= pc {
				err = node.addPart(keyn-1, &file); if err != nil {return} 
			}
		}		
	}
	return
}

func (node *Node) SetFile(file FormFile) (err error){
	fileStat, err := os.Stat(file.Path); if err != nil { return }		
	os.Rename(file.Path, node.DataPath())
	node.File.Name = file.Name
	node.File.Size = fileStat.Size()
	node.File.Checksum = file.Checksum
	err = node.Save()
	return
}

func (node *Node) SetAttributes(attr FormFile) (err error){
	attributes, err := ioutil.ReadFile(attr.Path); if err != nil { return	}
	err = json.Unmarshal(attributes, &node.Attributes); if err != nil { return	}
	err = node.Save()
	return
}

func (node *Node) setId() {
	/*
	var s []byte
	h := md5.New()
	h.Write([]byte(fmt.Sprint(time.Now().String(), rand.Float64())))	
	h.Sum(s)
	node.Id = fmt.Sprintf("%x",s)
	*/	
	id, _ := uuid.NewV5(uuid.NamespaceURL, []byte("shock"))	
	node.Id = id.String()
	return
}

func getPath(id string) (string) {
	DATAROOT := "/Users/jared/projects/GoShockData"
	return fmt.Sprintf("%s/%s/%s/%s/%s", DATAROOT, id[0:2], id[2:4], id[4:6], id)
}

func (node *Node) Path() (string) {
	return getPath(node.Id)
}

func (node *Node) DataPath() (string) {
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
	jsonPath := fmt.Sprintf("%s/%s.json", node.Path(), node.Id)
	os.Remove(jsonPath)
	n, err := node.ToJson(); if err != nil { return }
	err = ioutil.WriteFile(jsonPath, []byte(n), 0644); if err != nil { return }
	return
}

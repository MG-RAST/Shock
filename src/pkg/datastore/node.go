package datastore

import (
	"fmt"
	"time"
	"crypto/md5"
	"crypto/sha1"
	"rand"
	"os"
	"json"
	"path/filepath"
	"io/ioutil"
	"strconv"
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

// PartsList struct
type partsList struct {
	Count int 			`json:"count"`
	Length int 			`json:"length"`
	Parts []partsFile   `json:"parts"`
}

type partsFile []string

func (n *nodeFile) Empty() (bool){
	if n.Name == "" && n.Size == 0 && len(n.Checksum) == 0 {
		return true
	}
	return false
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
		sha1h := sha1.New()
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
	err = node.Save()
	return
}

func CreateNodeUpload(params map[string]string, files FormFiles) (node *Node, err os.Error) {
	node = new(Node)
	node.Indexes = make(map[string]string)
	node.File.Checksum = make(map[string]string)
	node.setId()

	err = node.Mkdir(); if err != nil { return }
	err = node.Update(params, files); if err != nil { return }
	err = node.Save()
	return
}

func (node *Node) Update(params map[string]string, files FormFiles) (err os.Error){	
	_, hasParts := params["parts"]
	if hasParts && node.partsCount() < 0 {
		if !node.File.Empty() { return os.NewError("file alreay set and is immutable") }		
		n, err := strconv.Atoi(params["parts"]); if err != nil {return}
		if n < 1 {
			return os.NewError("parts cannot be less than 1")
		}
		err = node.initParts(n); if err != nil {return}
	}
	
	_, hasFile := files["file"]
	if hasFile && node.File.Empty() {
		err = node.SetFile(files["file"]); if err != nil {return}
		files["file"] = FormFile{}, false
	} else if hasFile {
		err = os.NewError("node file immutable")
		return 
	}
	_, hasAttr := files["attributes"]
	if hasAttr && node.Attributes == nil  {	
		err = node.SetAttributes(files["attributes"]); if err != nil {return}
		os.Remove(files["attributes"].Path)		
		files["attributes"] = FormFile{}, false
	} else if hasAttr {
		err = os.NewError("node attributes immutable")
		return 
	}	
	
	pc := node.partsCount()
	if pc > 1 {
		for key, file := range files {
			if !node.File.Empty() { return os.NewError("node file already set and is immutable") }
			keyn, errf := strconv.Atoi(key)
			if errf == nil && keyn <= pc {
				err = node.addPart(keyn-1, file); if err != nil {return} 
			}
		}		
	}
	return
}

func (node *Node) loadParts() (p *partsList, err os.Error){
	pf, err := ioutil.ReadFile(node.partsListPath()); if err != nil { return }
	p = partsList{}
	err = json.Unmarshal(pf, &p); if err != nil { return }
	return
}

func (node *Node) writeParts(p *partsList) (err os.Error){
	pm, _ := json.Marshal(p)
	os.Remove(node.partsListPath())
	err = ioutil.WriteFile(node.partsListPath(), []byte(pm), 0644)
	return
} 

func (node *Node) partsCount() (int){
	p, err := node.loadParts(); if err != nil { return -1 } 
	return p.Count
}

func (node *Node) initParts(count int) (err os.Error){
	err = os.MkdirAll(fmt.Sprintf("%s/parts", node.Path()), 0777)
	p := partsList{Count : count, Length : 0, Parts: make([]partsFile, count)}
	err = node.writeParts(p)
	return
}

func (node *Node) addPart(n int, file *FormFile) (err os.Error){
	// load
	p, err := node.loadParts(); if err != nil { return } 
	
	// modify
	if len(p.Parts[n]) > 0 {
		err = os.NewError("node part already exists and is immutable")
		return
	}
	part := partsFile{file.Name, file.Checksum["md5"]}
	p.Parts[n] = part
	p.Length = p.Length + 1 
	os.Rename(file.Path, fmt.Sprintf("%s/parts/%d", node.Path(), n))

	// rewrite	
	err = node.writeParts(p); if err != nil { return }		

	// create file if done
	if p.Length == p.Count {
		err = node.SetFileFromParts(p); if err != nil { return }
	}
	return
}

func (node *Node) SetFileFromParts(p *partsList) (err os.Error){
	out, err := os.Create(fmt.Sprintf("%s/%s.data", node.Path(), node.Id)); if err != nil { return }
	defer out.Close()
	md5h := md5.New()
	sha1h := sha1.New()	
	for i := 0; i < p.Count; i++ {
		part, err := os.Open(fmt.Sprintf("%s/parts/%d", node.Path(), i)); if err != nil { return }
		for {
			buffer := make([]byte, 10240)
			n, err := part.Read(buffer)
			if n == 0 || err != nil { break }
			out.Write(buffer[0:n])
			md5h.Write(buffer[0:n])
			sha1h.Write(buffer[0:n]) 
		}
		part.Close()		
	}
	fileStat, err := os.Stat(fmt.Sprintf("%s/%s.data", node.Path(), node.Id)); if err != nil { return }		
	node.File.Name = node.Id
	node.File.Size = fileStat.Size
	node.File.Checksum["md5"] = fmt.Sprintf("%x", md5h.Sum())		
	node.File.Checksum["sha1"] = fmt.Sprintf("%x", sha1h.Sum())	
	err = node.Save()
	return
}

func (node *Node) partsListPath() (string){
	return fmt.Sprintf("%s/parts/parts.json", node.Path())	
}

func (node *Node) SetFile(file FormFile) (err os.Error){
	fileStat, err := os.Stat(file.Path); if err != nil { return }		
	os.Rename(file.Path, fmt.Sprintf("%s/%s.data", node.Path(), node.Id))
	node.File.Name = file.Name
	node.File.Size = fileStat.Size
	node.File.Checksum = file.Checksum
	err = node.Save()
	return
}

func (node *Node) SetAttributes(attr FormFile) (err os.Error){
	attributes, err := ioutil.ReadFile(attr.Path); if err != nil { return	}
	err = json.Unmarshal(attributes, &node.Attributes); if err != nil { return	}
	err = node.Save()
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
	jsonPath := fmt.Sprintf("%s/%s.json", node.Path(), node.Id)
	os.Remove(jsonPath)
	n, err := node.ToJson(); if err != nil { return }
	err = ioutil.WriteFile(jsonPath, []byte(n), 0644); if err != nil { return }
	return
}

package node

import (
	"crypto/md5"
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	"os"
)

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

func (node *Node) SetFileFromPath(path string) (err error) {
	fileStat, err := os.Stat(path)
	if err != nil {
		return
	}
	node.File.Name = fileStat.Name()
	node.File.Size = fileStat.Size()
	node.File.Path = path

	md5h := md5.New()
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
	}
	node.File.Checksum["md5"] = fmt.Sprintf("%x", md5h.Sum(nil))
	err = node.Save()
	return
}

func (node *Node) SetFileFromParts(p *partsList, allowEmpty bool) (err error) {
	out, err := os.Create(fmt.Sprintf("%s/%s.data", node.Path(), node.Id))
	if err != nil {
		return
	}
	defer out.Close()
	md5h := md5.New()
	for i := 1; i <= p.Count; i++ {
		filename := fmt.Sprintf("%s/parts/%d", node.Path(), i)

		// skip this portion unless either
		// 1. file exists, or
		// 2. file does not exist and allowEmpty == false
		if _, errf := os.Stat(filename); errf == nil || (errf != nil && allowEmpty == false) {
			part, err := os.Open(filename)
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
			}
			part.Close()
		}
	}
	fileStat, err := os.Stat(fmt.Sprintf("%s/%s.data", node.Path(), node.Id))
	if err != nil {
		return
	}
	node.File.Name = node.Id
	node.File.Size = fileStat.Size()
	node.File.Checksum["md5"] = fmt.Sprintf("%x", md5h.Sum(nil))
	err = node.Save()
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

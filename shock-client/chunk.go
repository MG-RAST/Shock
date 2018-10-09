package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	sc "github.com/MG-RAST/go-shock-client"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path"
	"regexp"
	"strconv"
)

const MaxBuffer = 64 * 1024

var chunkRegex = regexp.MustCompile(`^(\d+)(K|M|G)$`)

type chunkUploader struct {
	md5   string
	file  string
	name  string
	parts int
	chunk int64
}

func newChunkUploader(f string, c string) (cu chunkUploader) {
	cu = chunkUploader{
		file: f,
		name: path.Base(f),
	}
	cu.setMd5()
	if c != "" {
		cu.setSizes(c)
	}
	return
}

func (cu *chunkUploader) validateChunkNode(node *sc.ShockNode) (msg string) {
	// basic check
	if (node.Type != "parts") || (node.Parts == nil) {
		return "node " + node.Id + " is not a valid parts node"
	}
	if node.Parts.Count == node.Parts.Length {
		return "node " + node.Id + " has already completed upload"
	}
	// attr check
	attr := node.Attributes.(map[string]interface{})
	_, mok := attr["md5sum"]
	_, pok := attr["parts_size"]
	_, cok := attr["chunk_size"]
	if !mok || !pok || !cok {
		return "node " + node.Id + " missing required attributes"
	}

	aMd5 := attr["md5sum"].(string)
	aParts := int(attr["parts_size"].(float64))
	aChunk := int64(attr["chunk_size"].(float64))

	if aMd5 != cu.md5 {
		return fmt.Sprintf("checksum of %s does not match origional file started on node %s", cu.name, node.Id)
	}
	if node.Parts.Count != aParts {
		return "invalid parts node: node.attributes.parts_size != node.parts.count"
	}
	// set cu values
	cu.parts = aParts
	cu.chunk = aChunk
	return
}

func (cu *chunkUploader) getAttr() (attr map[string]interface{}) {
	attr = make(map[string]interface{})
	attr["type"] = "temp"
	attr["md5sum"] = cu.md5
	attr["file_name"] = cu.file
	attr["parts_size"] = cu.parts
	attr["chunk_size"] = cu.chunk
	return
}

func (cu *chunkUploader) setSizes(c string) {
	fi, err := os.Stat(cu.file)
	if err != nil {
		exitError(err.Error())
	}
	matched := chunkRegex.FindStringSubmatch(c)
	if len(matched) == 0 {
		exitError("chunk format is invalid")
	}

	chunkBytes, _ := strconv.ParseInt(matched[1], 10, 64)
	switch matched[2] {
	case "K":
		chunkBytes = chunkBytes * 1024
	case "M":
		chunkBytes = chunkBytes * 1024 * 1024
	case "G":
		chunkBytes = chunkBytes * 1024 * 1024 * 1024
	}

	quotient := fi.Size() / chunkBytes
	remainder := fi.Size() % chunkBytes
	if quotient > 100 {
		exitError("too many part uploads created, please specify a larger chunk size")
	}

	var calcParts int
	if remainder == 0 {
		calcParts = int(quotient)
	} else {
		calcParts = int(quotient + 1)
	}
	cu.parts = calcParts
	cu.chunk = chunkBytes
}

func (cu *chunkUploader) setMd5() {
	f, err := os.Open(cu.file)
	if err != nil {
		exitError(err.Error())
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		exitError(err.Error())
	}
	cu.md5 = fmt.Sprintf("%x", h.Sum(nil))
}

func (cu *chunkUploader) uploadParts(nid string, start int, dir string) (err error) {
	if start >= cu.parts {
		err = errors.New("invalid start position")
		return
	}
	fmt.Printf("node: %s, %d parts to upload\n", nid, cu.parts-start+1)

	var currSize int64
	var fh *os.File
	fh, err = os.Open(cu.file)
	if err != nil {
		return
	}
	defer fh.Close()

	for i := start; i <= cu.parts; i++ {
		tempFile := path.Join(dir, randomStr(12))
		_, err = os.Create(tempFile)
		if err != nil {
			return
		}
		for {
			if currSize >= cu.chunk {
				break
			}
			bufferSize := int(math.Min(MaxBuffer, float64(cu.chunk-currSize)))
			byteBuffer := make([]byte, bufferSize)
			_, ferr := fh.Read(byteBuffer)
			if ferr == io.EOF {
				break
			}
			ioutil.WriteFile(tempFile, byteBuffer, os.ModeAppend)
			currSize += int64(bufferSize)
		}
		currSize = 0
		// upload file part
		_, err = client.PutOrPostFile(tempFile, nid, i, "", "", nil, nil)
		if err != nil {
			return
		}
		os.Remove(tempFile)
		fmt.Printf("part %d uploaded\n", i)
	}
	return
}

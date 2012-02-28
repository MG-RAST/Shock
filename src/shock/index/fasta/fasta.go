package fasta

import (
	"shock/indexer"
	"shock/index"
	"os"
)

func init(){
	index.RegisterIndexer(index.FASTA, New)
}

type rec struct {
	offset   int64
	length   int64
	checksum string
}

type idx struct {
	index        map[int64]rec
	pattern      string
	checksumType string
}

func New() indexer.Indexer {
	i := new(idx)
	return i
}

func (i *idx) Read(p []uint8) (n int, err os.Error) {
	return
}

func (i *idx) Write(p []byte) (nn int, err os.Error) {
	return
}

func (i *idx) Save(filename string) (err os.Error) {
	return
}

func (i *idx) Load(filename string) (err os.Error) {
	return
}
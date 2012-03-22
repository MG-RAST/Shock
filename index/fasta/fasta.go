package fasta

import (
	"github.com/MG-RAST/Shock/index"
	"github.com/MG-RAST/Shock/indexer"
)

func init() {
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

func (i *idx) Read(p []uint8) (n int, err error) {
	return
}

func (i *idx) Write(p []byte) (nn int, err error) {
	return
}

func (i *idx) Save(filename string) (err error) {
	return
}

func (i *idx) Load(filename string) (err error) {
	return
}

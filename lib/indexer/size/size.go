package size

import (
	"indexer"
)

func init(){
	indexer.RegisterIndexer(indexer.SIZE, New)
}

type record struct {
	offset   int64
	length   int64
	checksum string
}

type index struct {
	index        map[int64]record
	pattern      string
	checksumType string
}

func New() indexer.Indexer {
	idx := new(index)
	return idx
}

func (idx *index) Read(p []byte) (nn int, err os.Error) {
	return
}

func (idx *index) Write(p []byte) (nn int, err os.Error) {
	return
}

func (idx *index) Save(filename string) (err os.Error) {
	return
}

func (idx *index) Load(filename string) (err os.Error) {
	return
}
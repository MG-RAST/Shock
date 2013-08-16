package index

import (
	"os"
)

type size struct{}

func NewSizeIndexer(f *os.File) Indexer {
	return &size{}
}

func (i *size) Create() (count int64, err error) {
	return
}

func (i *size) Dump(f string) error {
	return nil
}

func (i *size) Close() (err error) {
	return
}

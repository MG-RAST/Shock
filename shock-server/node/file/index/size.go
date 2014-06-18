package index

import (
	"os"
)

type size struct{}

func NewSizeIndexer(f *os.File, nType string, snFormat string, snIndexPath string) Indexer {
	return &size{}
}

func (i *size) Create(file string) (count int64, format string, err error) {
	return
}

func (i *size) Close() (err error) {
	return
}

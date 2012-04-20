package size

import (
	"github.com/MG-RAST/Shock/store/type/index"
	"io"
)

type indexer struct{}

func NewIndexer(f io.ReadCloser) index.Indexer {
	return &indexer{}
}

func (i *indexer) Create() (err error) {
	return
}

func (i *indexer) Dump(f string) error {
	return nil
}

func (i *indexer) Close() (err error) {
	return
}

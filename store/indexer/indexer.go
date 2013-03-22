package indexer

import (
	"github.com/MG-RAST/Shock/store/indexer/chunkrecord"
	"github.com/MG-RAST/Shock/store/indexer/record"
	"github.com/MG-RAST/Shock/store/type/index"
	"io"
)

type indexerFunc func(io.ReadCloser) index.Indexer

var (
	Indexers = map[string]indexerFunc{
		"record":      record.NewIndexer,
		"size":        record.NewIndexer,
		"chunkrecord": chunkrecord.NewIndexer,
	}
)

func Indexer(i string) indexerFunc {
	return Indexers[i]
}

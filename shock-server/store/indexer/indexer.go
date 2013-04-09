package indexer

import (
	"github.com/MG-RAST/Shock/shock-server/store/indexer/chunkrecord"
	"github.com/MG-RAST/Shock/shock-server/store/indexer/record"
	"github.com/MG-RAST/Shock/shock-server/store/type/index"
	"os"
)

type indexerFunc func(*os.File) index.Indexer

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

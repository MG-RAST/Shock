package index

import (
	"shock/indexer"
)

type Indexer uint

const (
	SIZE  Indexer = 1 + iota // in package index/size
	FASTA                    // in package index/fasta
	FASTQ                    // in package index/fastq	
	maxIndexer
)

var indexers = make([]func() indexer.Indexer, maxIndexer)

func (idx Indexer) New() indexer.Indexer {
	if idx > 0 && idx < maxIndexer {
		f := indexers[idx]
		if f != nil {
			return f()
		}
	}
	return nil
}

func RegisterIndexer(idx Indexer, f func() indexer.Indexer) {
	if idx >= maxIndexer {
		panic("indexer: RegisterIndexer of unknown indexer function")
	}
	indexers[idx] = f
}

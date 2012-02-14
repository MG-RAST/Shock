package indexer

import (
	"io"
)

const (
	SIZE  uint = 1 + iota // in package indexer/size
	FASTA                 // in package indexer/fasta
	FASTQ                 // in package indexer/fastq	
)

type Indexer interface {
	// Read file based on index file
	io.Reader

	// Write to stream creation
	io.Writer

	// Save index to disk 
	Save(string) os.Error

	// Load index from disk
	Load(string) *Index
}

var indexers = make([]func() Indexer, maxIndexer)

func (idx Indexer) New() Indexer {
	if idx > 0 && idx < maxIndexer {
		f := indexers[idx]
		if f != nil {
			return f()
		}
	}
	return nil
}

func RegisterIndexer(idx Indexer, f func() Indexer) {
	if idx >= maxIndexer {
		panic("indexer: RegisterIndexer of unknown indexer function")
	}
	indexers[idx] = f
}

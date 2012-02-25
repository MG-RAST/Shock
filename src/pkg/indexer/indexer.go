package indexer

import (
	"io"
)

type record struct {
	offset   int64
	length   int64
	checksum string
}

type Index struct {
	idx []rec
}

type Indexer interface {
	// Write to stream creation	
	io.Writer
	
	FromFile() bool
	
	SetPath(string)
	
	Result() *Index
}

/*
type Indexer interface {
	// Read file based on index file
	io.Reader

	// Write to stream creation
	io.Writer

	// Save index to disk 
	Save(string) os.Error

	// Load index from disk
	Load(string) os.Error
}
*/
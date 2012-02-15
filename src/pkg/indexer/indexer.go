package indexer

import (
	"os"
	"io"
)

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

package chunkrecord_test

import (
	. "github.com/MG-RAST/Shock/shock-server/indexer/chunkrecord"
	"os"
	"testing"
)

var (
	//testFile = "/Users/jared/test/60MB.fna"
	testFile = "/Users/jared/test/1GB.fna"
)

func TestCreate(t *testing.T) {
	fh, _ := os.Open(testFile)
	i := NewIndexer(fh)
	i.Create()
}

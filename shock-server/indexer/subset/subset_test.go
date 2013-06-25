package subset_test

import (
	//"fmt"
	. "github.com/MG-RAST/Shock/shock-server/indexer/subset"
	"os"
	"testing"
)

var (
	fullFile   = "../../testdata/10kb.fna"
	subsetFile = "../../testdata/10kb_subset.fna"
)

func TestCreate(t *testing.T) {
	f1, _ := os.Open(fullFile)
	f2, _ := os.Open(subsetFile)
	s := NewSubset(f1, f2)
	if err := s.Create("/tmp/subset_test.binary"); err != nil {
		println(err.Error())
	}
}

func TestLoad(t *testing.T) {

}

func TestReader(t *testing.T) {

}

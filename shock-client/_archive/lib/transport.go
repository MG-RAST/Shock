package lib

import (
	"os"
)

const MB = int64(1048576)

var (
	partSize int64
)

func init() {
	partSize = 100 * MB
}

type Trans struct {
	Url   string
	Count int
	Path  string
	Name  string
	Size  int64
	Parts []OLP
}

type OLP struct {
	Offset int64
	Length int64
	Part   int
}

func PartionUpload(url string, path string) (t Trans, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	if fi, _ := f.Stat(); err == nil {
		t.Url = url
		t.Path = path
		t.Name = fi.Name()
		t.Size = fi.Size()
		var i int64 = 0
		count := 0
		for ; i+partSize < fi.Size(); i += partSize + 1 {
			count += 1
			t.Parts = append(t.Parts, OLP{Offset: i, Length: partSize, Part: count})
		}
		if i != fi.Size() {
			count += 1
			t.Parts = append(t.Parts, OLP{Offset: i, Length: fi.Size() - i, Part: count})
		}
		t.Count = count
	}
	return
}

package index

import (
	"github.com/MG-RAST/Shock/shock-server/node/file/format/multi"
	"github.com/MG-RAST/Shock/shock-server/node/file/format/seq"
	"io"
	"os"
)

type record struct {
	f     *os.File
	r     seq.Reader
	Index *Idx
}

func NewRecordIndexer(f *os.File) Indexer {
	return &record{
		f:     f,
		r:     multi.NewReader(f),
		Index: New(),
	}
}

func (i *record) Create() (count int64, err error) {
	curr := int64(0)
	count = 0
	for {
		buf := make([]byte, 32*1024)
		n, er := i.r.ReadRaw(buf)
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
		i.Index.Append([]int64{curr, int64(n)})
		curr += int64(n)
		count += 1
	}
	return
}

func (i *record) Dump(f string) error {
	return i.Index.Dump(f)
}

func (i *record) Close() (err error) {
	i.f.Close()
	return
}

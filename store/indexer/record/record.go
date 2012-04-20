package record

import (
	"github.com/MG-RAST/Shock/store/type/index"
	"github.com/MG-RAST/Shock/store/type/sequence/multi"
	"github.com/MG-RAST/Shock/store/type/sequence/seq"
	"io"
)

type indexer struct {
	f     io.ReadCloser
	r     seq.ReadCloser
	Index *index.Idx
}

func NewIndexer(f io.ReadCloser) index.Indexer {
	return &indexer{
		f:     f,
		r:     multi.NewReader(f),
		Index: index.New(),
	}
}

func (i *indexer) Create() (err error) {
	curr := int64(0)
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
	}
	return
}

func (i *indexer) Dump(f string) error {
	return i.Index.Dump(f)
}

func (i *indexer) Close() (err error) {
	i.f.Close()
	return
}

package chunkrecord

import (
	"github.com/MG-RAST/Shock/shock-server/store/type/index"
	"github.com/MG-RAST/Shock/shock-server/store/type/sequence/multi"
	"github.com/MG-RAST/Shock/shock-server/store/type/sequence/seq"
	"io"
	"os"
)

type indexer struct {
	f     *os.File
	r     seq.Reader
	Index *index.Idx
	size  int64
}

func NewIndexer(f *os.File) index.Indexer {
	fi, _ := f.Stat()
	return &indexer{
		f:     f,
		size:  fi.Size(),
		r:     multi.NewReader(f),
		Index: index.New(),
	}
}

func (i *indexer) Create() (count int64, err error) {
	curr := int64(0)
	count = 0
	for {
		n, er := i.r.SeekChunk(curr)
		if er != nil {
			if er == io.EOF {
				i.Index.Append([]int64{curr, i.size - curr})
				count += 1
			} else {
				err = er
			}
			break
		} else {
			i.Index.Append([]int64{curr, n})
			curr += int64(n)
			count += 1
		}
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

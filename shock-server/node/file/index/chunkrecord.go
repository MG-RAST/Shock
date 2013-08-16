package index

import (
	"github.com/MG-RAST/Shock/shock-server/node/file/format/multi"
	"github.com/MG-RAST/Shock/shock-server/node/file/format/seq"
	"io"
	"os"
)

type chunkRecord struct {
	f     *os.File
	r     seq.Reader
	Index *Idx
	size  int64
}

func NewChunkRecordIndexer(f *os.File) Indexer {
	fi, _ := f.Stat()
	return &chunkRecord{
		f:     f,
		size:  fi.Size(),
		r:     multi.NewReader(f),
		Index: New(),
	}
}

func (i *chunkRecord) Create() (count int64, err error) {
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

func (i *chunkRecord) Dump(f string) error {
	return i.Index.Dump(f)
}

func (i *chunkRecord) Close() (err error) {
	i.f.Close()
	return
}

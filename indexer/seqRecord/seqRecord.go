package seqRecord

import (
	"github.com/MG-RAST/Shock/datastore"
	"github.com/MG-RAST/Shock/types/sequence/multi"
	"github.com/MG-RAST/Shock/types/sequence/seq"
	"io"
)

type Indexer struct {
	f     io.ReadCloser
	r     seq.ReadCloser
	Index *datastore.BinaryIndex
}

func NewIndexer(f io.ReadCloser) *Indexer {
	return &Indexer{
		f:     f,
		r:     multi.NewReader(f),
		Index: datastore.NewBinaryIndex(),
	}
}

func (i *Indexer) Create() (err error) {
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

func (i *Indexer) Close() (err error) {
	i.f.Close()
	return
}

package index

import (
	"github.com/MG-RAST/Shock/shock-server/node/file/format/line"
	"io"
	"os"
)

type lineRecord struct {
	f     *os.File
	r     line.LineReader
	Index *Idx
}

func NewLineIndexer(f *os.File) Indexer {
	return &lineRecord{
		f:     f,
		r:     line.NewReader(f),
		Index: New(),
	}
}

func (l *lineRecord) Create() (count int64, err error) {
	curr := int64(0)
	count = 0
	for {
		n, er := l.r.GetReadOffset()
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
		l.Index.Append([]int64{curr, int64(n)})
		curr += int64(n)
		count += 1
	}
	return
}

func (l *lineRecord) Dump(f string) error {
	return l.Index.Dump(f)
}

func (l *lineRecord) Close() (err error) {
	l.f.Close()
	return
}

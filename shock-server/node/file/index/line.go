package index

import (
	"encoding/binary"
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/node/file/format/line"
	"io"
	"math/rand"
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

func (l *lineRecord) Create(file string) (count int64, err error) {
	tmpFilePath := fmt.Sprintf("%s/temp/%d%d.idx", conf.Conf["data-path"], rand.Int(), rand.Int())

	f, err := os.Create(tmpFilePath)
	if err != nil {
		return
	}
	defer f.Close()

	curr := int64(0)
	count = 0
	for {
		n, er := l.r.GetReadOffset()
		if er != nil {
			if er != io.EOF {
				err = er
				return
			}
			break
		}
		binary.Write(f, binary.LittleEndian, curr)
		binary.Write(f, binary.LittleEndian, int64(n))
		curr += int64(n)
		count += 1
	}
	err = os.Rename(tmpFilePath, file)

	return
}

func (l *lineRecord) Close() (err error) {
	l.f.Close()
	return
}

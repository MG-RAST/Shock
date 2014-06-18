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
	t     string
	snf   string
	snp   string
	Index *Idx
}

func NewLineIndexer(f *os.File, nType string, snFormat string, snIndexPath string) Indexer {
	return &lineRecord{
		f:     f,
		r:     line.NewReader(f),
		t:     nType,
		snf:   snFormat,
		snp:   snIndexPath,
		Index: New(),
	}
}

func (l *lineRecord) Create(file string) (count int64, format string, err error) {
	tmpFilePath := fmt.Sprintf("%s/temp/%d%d.idx", conf.Conf["data-path"], rand.Int(), rand.Int())

	f, err := os.Create(tmpFilePath)
	if err != nil {
		return
	}
	defer f.Close()

	format = "array"
	curr := int64(0)
	count = 0
	buffer_pos := 0 // used to track the location in our byte array

	// Writing index file in 16MB chunks
	var b [16777216]byte
	for {
		// io.EOF error does not get returned from GetReadOffset() until all lines have been read
		n, er := l.r.GetReadOffset()
		if er != nil {
			if er != io.EOF {
				err = er
				return
			}
			break
		}
		x := (buffer_pos * 16)
		if x == 16777216 {
			f.Write(b[:])
			buffer_pos = 0
			x = 0
		}
		y := x + 8
		z := x + 16

		binary.LittleEndian.PutUint64(b[x:y], uint64(curr))
		binary.LittleEndian.PutUint64(b[y:z], uint64(n))
		curr += int64(n)
		count += 1
		buffer_pos += 1
	}
	if buffer_pos != 0 {
		f.Write(b[:buffer_pos*16])
	}

	err = os.Rename(tmpFilePath, file)

	return
}

func (l *lineRecord) Close() (err error) {
	l.f.Close()
	return
}

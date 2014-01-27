package index

import (
	"encoding/binary"
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/node/file/format/multi"
	"github.com/MG-RAST/Shock/shock-server/node/file/format/seq"
	"io"
	"math/rand"
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

func (i *chunkRecord) Create(file string) (count int64, err error) {
	tmpFilePath := fmt.Sprintf("%s/temp/%d%d.idx", conf.Conf["data-path"], rand.Int(), rand.Int())

	f, err := os.Create(tmpFilePath)
	if err != nil {
		return
	}
	defer f.Close()

	curr := int64(0)
	count = 0
	for {
		n, er := i.r.SeekChunk(curr)
		if er != nil {
			if er != io.EOF {
				err = er
				return
			}
			binary.Write(f, binary.LittleEndian, curr)
			binary.Write(f, binary.LittleEndian, i.size-curr)
			count += 1
			break
		} else {
			binary.Write(f, binary.LittleEndian, curr)
			binary.Write(f, binary.LittleEndian, n)
			curr += int64(n)
			count += 1
		}
	}
	err = os.Rename(tmpFilePath, file)

	return
}

func (i *chunkRecord) Close() (err error) {
	i.f.Close()
	return
}

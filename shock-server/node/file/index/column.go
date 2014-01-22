package index

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/node/file/format/line"
	"io"
	"math/rand"
	"os"
)

type column struct {
	f     *os.File
	r     line.LineReader
	Index *Idx
}

func NewColumnIndexer(f *os.File) column {
	return column{
		f:     f,
		r:     line.NewReader(f),
		Index: New(),
	}
}

func (c *column) Create(string) (count int64, err error) {
	return
}

func CreateColumnIndex(c *column, column int, ofile string) (count int64, err error) {
	tmpFilePath := fmt.Sprintf("%s/temp/%d%d.idx", conf.Conf["data-path"], rand.Int(), rand.Int())

	f, err := os.Create(tmpFilePath)
	if err != nil {
		return
	}
	defer f.Close()

	curr := int64(0) // stores the offset position of the current index
	count = 0        // stores the number of indexed positions and get returned
	total_n := 0     // stores the number of bytes read for the current index record
	line_count := 0  // stores the number of lines that have been read from the data file
	prev_str := ""   // keeps track of the string of the specified column of the previous line
	for {
		buf := make([]byte, 32*1024)
		n, er := c.r.ReadRaw(buf)
		if er != nil {
			if er != io.EOF {
				err = er
				return
			}
			break
		}

		// split line by columns and test if column value has changed
		slices := bytes.Split(buf, []byte("\t"))
		if len(slices) < column-1 {
			return 0, errors.New("Specified column does not exist for all lines in file.")
		}

		str := string(slices[column-1])
		if prev_str != str && line_count != 0 {
			binary.Write(f, binary.LittleEndian, curr)
			binary.Write(f, binary.LittleEndian, int64(total_n))
			curr += int64(total_n)
			count += 1
			total_n = 0
			prev_str = str
		}
		if line_count == 0 {
			prev_str = str
		}
		total_n += n
		line_count += 1
	}

	binary.Write(f, binary.LittleEndian, curr)
	binary.Write(f, binary.LittleEndian, int64(total_n))

	err = os.Rename(tmpFilePath, ofile)

	return
}

func (c *column) Close() (err error) {
	c.f.Close()
	return
}

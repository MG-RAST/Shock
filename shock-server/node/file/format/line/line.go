// Package to read and write lines of a file
package line

import (
	"bufio"
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"io"
)

type Reader struct {
	f io.Reader
	r *bufio.Reader
}

type LineReader interface {
	ReadRaw(p []byte) (int, error)
	GetReadOffset() (int, error)
}

func NewReader(f file.SectionReader) LineReader {
	return &Reader{
		f: f,
		r: bufio.NewReader(f),
	}
}

// Read a single line and return it or an error.
func (self *Reader) ReadRaw(p []byte) (n int, err error) {
	if self.r == nil {
		self.r = bufio.NewReader(self.f)
	}
	for {
		read, er := self.r.ReadBytes('\n')
		n += len(read)
		if len(read) > 1 {
			copy(p[0:len(read)], read[0:len(read)])
			break
		} else if er != nil {
			err = er
			break
		}
	}
	return
}

// Read a single line and return the offset for indexing.
func (self *Reader) GetReadOffset() (n int, err error) {
	if self.r == nil {
		self.r = bufio.NewReader(self.f)
	}
	for {
		read, er := self.r.ReadBytes('\n')
		n += len(read)
		if len(read) > 1 {
			break
		} else if er != nil {
			err = er
			break
		}
	}
	return
}

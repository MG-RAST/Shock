// Package to read and index lines of a file
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
	ReadLine() ([]byte, error)
	GetReadOffset() (int, error)
}

func NewReader(f file.SectionReader) LineReader {
	return &Reader{
		f: f,
		r: bufio.NewReader(f),
	}
}

// Read a single line and return it or an error.
func (self *Reader) ReadLine() (p []byte, err error) {
	if self.r == nil {
		self.r = bufio.NewReader(self.f)
	}
	p, err = self.r.ReadBytes('\n')
	return
}

// Read a single line and return the offset for indexing.
func (self *Reader) GetReadOffset() (n int, err error) {
	if self.r == nil {
		self.r = bufio.NewReader(self.f)
	}
	p, err := self.r.ReadBytes('\n')
	return len(p), err
}

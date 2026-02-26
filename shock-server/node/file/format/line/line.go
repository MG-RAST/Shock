// Package to read and index lines of a file
package line

import (
	"bufio"
	"io"

	"github.com/MG-RAST/Shock/shock-server/node/file"
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
// If data is read but EOF is encountered without a newline, return the data without an error.
// Only return EOF when no data is read.
func (self *Reader) ReadLine() (p []byte, err error) {
	if self.r == nil {
		self.r = bufio.NewReader(self.f)
	}
	p, err = self.r.ReadBytes('\n')

	// If we read some data but got EOF without a newline, return the data without an error
	if err == io.EOF && len(p) > 0 {
		err = nil
	}
	return
}

// Read a single line and return the offset for indexing.
// If data is read but EOF is encountered without a newline, return the offset without an error.
// Only return EOF when no data is read.
func (self *Reader) GetReadOffset() (n int, err error) {
	if self.r == nil {
		self.r = bufio.NewReader(self.f)
	}
	var p []byte
	p, err = self.r.ReadBytes('\n')
	n = len(p)

	// If we read some data but got EOF without a newline, return the offset without an error
	if err == io.EOF && n > 0 {
		err = nil
	}
	return
}

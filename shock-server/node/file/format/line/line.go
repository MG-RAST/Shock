// Package to read and index lines of a file
package line

import (
	"bufio"
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"io"
)

type Reader struct {
	f io.Reader
	s *bufio.Scanner
}

type LineReader interface {
	ReadRaw(p []byte) (int, error)
	GetReadOffset() (int, error)
}

func NewReader(f file.SectionReader) LineReader {
	return &Reader{
		f: f,
		s: bufio.NewScanner(f),
	}
}

// Read a single line and return it or an error.
func (self *Reader) ReadRaw(p []byte) (n int, err error) {
	if self.s == nil {
		self.s = bufio.NewScanner(self.f)
	}
	for {
		if self.s.Scan() == false {
			err = io.EOF
			break
		}
		line := self.s.Text()
		length := len(line)
		n += length + 1
		if length > 0 {
			copy(p[0:length], line[0:length])
			break
		} else if er := self.s.Err(); er != nil {
			err = er
			break
		}
	}
	return
}

// Read a single line and return the offset for indexing.
func (self *Reader) GetReadOffset() (n int, err error) {
	if self.s == nil {
		self.s = bufio.NewScanner(self.f)
	}
	for {
		if self.s.Scan() == false {
			err = io.EOF
			break
		}
		line := self.s.Text()
		length := len(line)
		n += length + 1
		if length > 0 {
			break
		} else if er := self.s.Err(); er != nil {
			err = er
			break
		}
	}
	return
}

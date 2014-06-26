// Package to read and write FASTA format files
package fasta

// Modified under the terms of GPL3 from
// Dan Kortschak github.com/kortschak/BioGo

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/MG-RAST/Shock/shock-server/node/file/format/seq"
	"io"
	"os"
	"regexp"
)

var (
	Regex = regexp.MustCompile(`^[\n\r]*>\S+[\S\t ]*[\n\r]+[A-Za-z\- ]+[\n\r]+`)
)

// Fasta sequence format reader type.
type Reader struct {
	f file.SectionReader
	r *bufio.Reader
}

// Returns a new fasta format reader using f.
func NewReader(f file.SectionReader) seq.ReadRewinder {
	return &Reader{
		f: f,
		r: nil,
	}
}

// Read a single sequence and return it or an error.
func (self *Reader) Read() (sequence *seq.Seq, err error) {
	if self.r == nil {
		self.r = bufio.NewReader(self.f)
	}
	var label, body []byte
	for {
		read, err := self.r.ReadBytes('>')
		if len(read) > 1 {
			lines := bytes.Split(read, []byte{'\n'})
			if len(lines) > 1 {
				label = lines[0]
				body = bytes.Join(lines[1:len(lines)-1], []byte{})
			}
			break
		} else if err != nil {
			return nil, io.EOF
		}
	}
	if len(label) > 0 && len(body) > 0 {
		sequence = seq.New(label, body, nil)
	} else {
		return nil, errors.New("Invalid fasta entry")
	}
	return
}

// Read a single sequence and return it or an error.
func (self *Reader) ReadRaw(p []byte) (n int, err error) {
	if self.r == nil {
		self.r = bufio.NewReader(self.f)
	}
	p[n] = byte('>')
	n = 1
	for {
		read, er := self.r.ReadBytes('>')
		if len(read) > 1 {
			copy(p[n:n+len(read)-1], read[0:len(read)-1])
			n += len(read) - 1
			break
		} else if er != nil {
			err = er
			break
		}
	}
	return
}

// Read a single sequence and return read offset for indexing.
func (self *Reader) GetReadOffset() (n int, err error) {
	if self.r == nil {
		self.r = bufio.NewReader(self.f)
	}
	n = 1
	for {
		read, er := self.r.ReadBytes('>')
		if len(read) > 1 {
			n += len(read) - 1
			break
		} else if er != nil {
			err = er
			break
		}
	}
	return
}

// seek sequences which add up to a size close to the configured chunk size (conf.CHUNK_SIZE, e.g. 1M)
func (self *Reader) SeekChunk(offSet int64) (n int64, err error) {
	winSize := int64(32768)
	r := io.NewSectionReader(self.f, offSet+conf.CHUNK_SIZE-winSize, winSize)
	buf := make([]byte, winSize)
	if n, err := r.Read(buf); err != nil {
		return int64(n), err
	}
	if pos := bytes.LastIndex(buf, []byte(">")); pos == -1 {
		indexPos, err := self.SeekChunk(offSet + winSize)
		return (winSize + indexPos), err
	} else {
		return conf.CHUNK_SIZE - winSize + int64(pos), nil
	}
	return
}

// Rewind the reader.
func (self *Reader) Rewind() (err error) {
	if s, ok := self.f.(io.Seeker); ok {
		_, err = s.Seek(0, 0)
		self.r = bufio.NewReader(self.f)
	} else {
		err = errors.New("Not a Seeker")
	}
	return
}

// Fasta sequence format writer type.
type Writer struct {
	f io.WriteCloser
	w *bufio.Writer
}

// Returns a new fasta format writer using f.
func NewWriter(f io.WriteCloser, width int) *Writer {
	return &Writer{
		f: f,
		w: bufio.NewWriter(f),
	}
}

// Returns a new fasta format writer using a filename, truncating any existing file.
// If appending is required use NewWriter and os.OpenFile.
func NewWriterName(name string, width int) (w *Writer, err error) {
	var f *os.File
	if f, err = os.Create(name); err != nil {
		return
	}
	return NewWriter(f, width), nil
}

// Write a single sequence and return the number of bytes written and any error.
func (self *Writer) Write(s *seq.Seq) (n int, err error) {
	return Format(s, self.w)
}

// Format a single sequence into fasta string
func Format(s *seq.Seq, w io.Writer) (n int, err error) {
	return w.Write([]byte(">" + string(s.ID) + "\n" + string(s.Seq) + "\n"))
}

// Flush the writer.
func (self *Writer) Flush() error {
	return self.w.Flush()
}

// Close the writer, flushing any unwritten sequence.
func (self *Writer) Close() (err error) {
	if err = self.w.Flush(); err != nil {
		return
	}
	return self.f.Close()
}

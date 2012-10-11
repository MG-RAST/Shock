// Package to read and write FASTA format files
package fasta

// Modified under the terms of GPL3 from 
// Dan Kortschak github.com/kortschak/BioGo 

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/MG-RAST/Shock/store/type/sequence/seq"
	"io"
	"os"
)

// Fasta sequence format reader type.
type Reader struct {
	f io.Reader
	r *bufio.Reader
}

// Returns a new fasta format reader using f.
func NewReader(f io.Reader) *Reader {
	return &Reader{
		f: f,
		r: bufio.NewReader(f),
	}
}

// Read a single sequence and return it or an error.
func (self *Reader) Read() (sequence *seq.Seq, err error) {
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

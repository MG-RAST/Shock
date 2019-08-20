// Package to read and write SAM format files
package sam

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"regexp"

	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/MG-RAST/Shock/shock-server/node/file/format/seq"
)

var (
	Regex = regexp.MustCompile(`^[\n\r]*[@[A-Z][A-Z][ \t]+[\S \t]+[\n\r]]*`)
)

// Sam sequence format reader type.
type Reader struct {
	f io.Reader
	r *bufio.Reader
}

// Returns a new Sam format reader using r.
func NewReader(f file.SectionReader) seq.ReadRewinder {
	return &Reader{
		f: f,
		r: bufio.NewReader(f),
	}
}

// Returns a new Sam format reader using a filename.
func NewReaderName(name string) (r seq.ReadRewinder, err error) {
	var f *os.File
	if f, err = os.Open(name); err != nil {
		return
	}
	return NewReader(f), nil
}

// Read a single sequence and return it or an error.
func (self *Reader) Read() (sequence *seq.Seq, err error) {
	var line, label, seqBody []byte
	sequence = &seq.Seq{}

	for {
		if line, err = self.r.ReadBytes('\n'); err == nil {
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			line = bytes.TrimSpace(line)
			if len(line) == 0 {
				continue
			}

			//skip header
			if line[0] == '@' {
				continue
			}

			seqBody = line
			fields := bytes.Split(line, []byte{'\t'})
			if len(fields) < 11 {
				return nil, errors.New("sam alignment fields less than 11")
			}

			label = fields[0]

			break
		} else {
			return
		}
	}

	sequence = seq.New(label, seqBody, nil)

	return
}

// Read a single sequence and return read offset for indexing.
func (self *Reader) GetReadOffset() (n int, err error) {
	for {
		read, er := self.r.ReadBytes('\n')
		n += len(read)
		if len(read) > 1 {
			if read[0] == '@' {
				continue
			}
			break
		} else if er != nil {
			err = er
			break
		}
	}
	return
}

func (self *Reader) SeekChunk(offSet int64, lastIndex bool) (n int64, err error) {
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

// Returns a new sam format writer using f.
func NewWriter(f io.WriteCloser, width int) *Writer {
	return &Writer{
		f: f,
		w: bufio.NewWriter(f),
	}
}

// Returns a new sam format writer using a filename, truncating any existing file.
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

//To-do: not finished
//Format a single sequence into sam string
func Format(s *seq.Seq, w io.Writer) (n int, err error) {
	return w.Write([]byte(string(s.Seq) + "\n"))
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

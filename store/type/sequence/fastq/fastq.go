// Package to read and write FASTQ format files
package fastq

// Modified under the terms of GPL3 from 
// Dan Kortschak github.com/kortschak/BioGo

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/MG-RAST/Shock/conf"
	"github.com/MG-RAST/Shock/store"
	"github.com/MG-RAST/Shock/store/type/sequence/seq"
	"io"
	"os"
	"regexp"
)

var (
	Regex = regexp.MustCompile(`@[\S ]+[\n\r]+[A-Za-z]+[\n\r]+\+\S*[\n\r]+\S*[\n\r]*`)
)

// Fastq sequence format reader type.
type Reader struct {
	f store.SectionReader
	r *bufio.Reader
}

// Returns a new fastq format reader using r.
func NewReader(f store.SectionReader) seq.ReadRewinder {
	return &Reader{
		f: f,
		r: nil,
	}
}

// Returns a new fastq format reader using a filename.
func NewReaderName(name string) (r seq.ReadRewinder, err error) {
	var f *os.File
	if f, err = os.Open(name); err != nil {
		return
	}
	return NewReader(f), nil
}

// Read a single sequence and return it or an error.
// TODO: Does not read interleaved fastq.
func (self *Reader) Read() (sequence *seq.Seq, err error) {
	if self.r == nil {
		self.r = bufio.NewReader(self.f)
	}
	var line, label, seqBody, qualBody []byte
	sequence = &seq.Seq{}

	inQual := false
READ:
	for {
		if line, err = self.r.ReadBytes('\n'); err == nil {
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			line = bytes.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			switch {
			case !inQual && line[0] == '@':
				label = line[1:]
			case !inQual && line[0] == '+':
				if len(label) == 0 {
					return nil, errors.New("No ID line parsed at +line in fastq format")
				}
				if len(line) > 1 && bytes.Compare(label, line[1:]) != 0 {
					return nil, errors.New("Quality ID does not match sequence ID")
				}
				inQual = true
			case !inQual:
				line = bytes.Join(bytes.Fields(line), nil)
				seqBody = append(seqBody, line...)
			case inQual:
				line = bytes.Join(bytes.Fields(line), nil)
				qualBody = append(qualBody, line...)
				if len(qualBody) >= len(seqBody) {
					break READ
				}
			}
		} else {
			return
		}
	}

	if len(seqBody) != len(qualBody) {
		return nil, errors.New("Quality length does not match sequence length")
	}
	sequence = seq.New(label, seqBody, qualBody)

	return
}

func (self *Reader) ReadRaw(p []byte) (n int, err error) {
	if self.r == nil {
		self.r = bufio.NewReader(self.f)
	}
	curr := 0
	id, err := self.r.ReadBytes('\n')
	if err != nil {
		return 0, err
	} else if !bytes.HasPrefix(id, []byte{'@'}) {
		return 0, errors.New("Invalid format: id line does not start with @")
	}
	copy(p[curr:len(id)+curr], id)
	curr += len(id)

	seq, err := self.r.ReadBytes('\n')
	if err != nil {
		return 0, err
	}
	copy(p[curr:len(seq)+curr], seq)
	curr += len(seq)

	plus, err := self.r.ReadBytes('\n')
	if err != nil {
		return 0, err
	} else if !bytes.HasPrefix(plus, []byte{'+'}) {
		return 0, errors.New("Invalid format: plus line does not start with +")
	}
	copy(p[curr:len(plus)+curr], plus)
	curr += len(plus)

	qual, err := self.r.ReadBytes('\n')
	if err != nil {
		return 0, err
	} else if len(seq) != len(qual) {
		return 0, errors.New("Invalid format: length of sequence and quality lines do not match")
	}
	copy(p[curr:len(qual)+curr], qual)
	n = curr + len(qual)
	return
}

// seek sequences which add up to a size close to the configured chunk size (conf.CHUNK_SIZE, e.g. 1M)
func (self *Reader) SeekChunk(offSet int64) (n int64, err error) {
	r := io.NewSectionReader(self.f, offSet+conf.CHUNK_SIZE-32768, 32768)
	buf := make([]byte, 32768)
	if n, err := r.Read(buf); err != nil {
		return int64(n), err
	}
	if pos := bytes.LastIndex(buf, []byte("@")); pos == -1 {
		return self.SeekChunk(offSet + conf.CHUNK_SIZE)
	} else {
		return conf.CHUNK_SIZE - 32768 + int64(pos), nil
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

// Fastq sequence format writer type.
type Writer struct {
	f io.WriteCloser
	w *bufio.Writer
}

// Returns a new fastq format writer using w.
func NewWriter(f io.WriteCloser) *Writer {
	return &Writer{
		f: f,
		w: bufio.NewWriter(f),
	}
}

// Returns a new fastq format writer using a filename, truncating any existing file.
// If appending is required use NewWriter and os.OpenFile.
func NewWriterName(name string) (w *Writer, err error) {
	var f *os.File
	if f, err = os.Create(name); err != nil {
		return
	}
	return NewWriter(f), nil
}

// Write a single sequence and return the number of bytes written and any error.
func (self *Writer) Write(s *seq.Seq) (n int, err error) {
	if s.Qual == nil {
		return 0, errors.New("No quality associated with sequence")
	}
	if len(s.Seq) == len(s.Qual) {
		n, err = Format(s, self.w)
		return
	} else {
		return 0, errors.New("Sequence length and quality length do not match")
	}

	return
}

// Format a single sequence into fastq string
func Format(s *seq.Seq, w io.Writer) (n int, err error) {
	return w.Write([]byte("@" + string(s.ID) + "\n" + string(s.Seq) + "\n+\n" + string(s.Qual) + "\n"))
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

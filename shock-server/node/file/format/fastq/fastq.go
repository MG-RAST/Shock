// Package to read and write FASTQ format files
package fastq

// Modified under the terms of GPL3 from
// Dan Kortschak github.com/kortschak/BioGo

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"math"
	"os"
	"regexp"

	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/MG-RAST/Shock/shock-server/node/file/format/seq"
)

var (
	Regex  = regexp.MustCompile(`^[\n\r]*@\S+[\S\t ]*[\n\r]+[A-Za-z\-]+[\n\r]+\+[\S\t ]*[\n\r]+\S*[\n\r]+`)
	Record = regexp.MustCompile(`@\S(.*)?[\n\r]+[A-Za-z\-]+[\n\r]+\+(.*)?[\n\r]+(\S+)[\n\r]+`)
)

// Fastq sequence format reader type.
type Reader struct {
	f file.SectionReader
	r *bufio.Reader
}

// Returns a new fastq format reader using r.
func NewReader(f file.SectionReader) seq.ReadRewinder {
	return &Reader{
		f: f,
		r: nil,
	}
}

// Returns a new fastq format reader using a filename.
func NewReaderName(name string) (r seq.ReadRewinder, err error) {
	var f *os.File
	if f, err = os.Open(name); err != nil { // used to be os.Open(name);
		//if f, err = node.FMOpen(name); err != nil { // used to be os.Open(name);
		return
	}
	return NewReader(f), nil
}

// Read a single sequence and return it or an error.
func (self *Reader) Read() (sequence *seq.Seq, err error) {
	if self.r == nil {
		self.r = bufio.NewReader(self.f)
	}
	var seqId, seqBody, qualId, qualBody []byte

	// skip empty lines only at eof
	empty := false
	for {
		seqId, err = self.r.ReadBytes('\n')
		if err != nil {
			break
		}
		if len(seqId) > 1 {
			break
		}
		empty = true
	}

	if err == io.EOF {
		if len(seqId) > 0 {
			err = errors.New("Invalid format: truncated fastq record")
		}
		return
	} else if err != nil {
		return
	} else if empty {
		err = errors.New("Invalid format: empty line(s) between records")
		return
	} else if !bytes.HasPrefix(seqId, []byte{'@'}) {
		err = errors.New("Invalid format: id line does not start with @")
		return
	}
	seqId = bytes.TrimSpace(seqId[1:])
	if len(seqId) == 0 {
		err = errors.New("Invalid format: missing sequence ID")
		return
	}

	seqBody, err = self.r.ReadBytes('\n')
	if err == io.EOF {
		err = errors.New("Invalid format: truncated fastq record")
		return
	} else if err != nil {
		return
	}
	seqBody = bytes.TrimSpace(seqBody)
	if len(seqBody) == 0 {
		err = errors.New("Invalid format: empty sequence")
		return
	}

	qualId, err = self.r.ReadBytes('\n')
	if err == io.EOF {
		err = errors.New("Invalid format: truncated fastq record")
		return
	} else if err != nil {
		return
	} else if !bytes.HasPrefix(qualId, []byte{'+'}) {
		err = errors.New("Invalid format: plus line does not start with +")
		return
	}
	qualId = bytes.TrimSpace(qualId)
	if (len(qualId) > 1) && (bytes.Compare(seqId, qualId[1:]) != 0) {
		err = errors.New("Invalid format: quality ID does not match sequence ID")
		return
	}

	qualBody, err = self.r.ReadBytes('\n')
	if (err != nil) && (err != io.EOF) {
		return
	}
	qualBody = bytes.TrimSpace(qualBody)
	if len(seqBody) != len(qualBody) {
		err = errors.New("Invalid format: length of sequence and quality lines do not match")
		return
	}

	sequence = seq.New(seqId, seqBody, qualBody)
	return
}

// Read a single sequence and return read offset for indexing.
func (self *Reader) GetReadOffset() (n int, err error) {
	if self.r == nil {
		self.r = bufio.NewReader(self.f)
	}
	var seqId, seqBody, qualId, qualBody []byte
	curr := 0

	// skip empty lines only at eof
	empty := false
	for {
		seqId, err = self.r.ReadBytes('\n')
		if err != nil {
			break
		}
		if len(seqId) > 1 {
			break
		}
		empty = true
	}

	if err == io.EOF {
		if len(seqId) > 0 {
			err = errors.New("Invalid format: truncated fastq record")
		}
		return
	} else if err != nil {
		return
	} else if empty {
		err = errors.New("Invalid format: empty line(s) between records")
		return
	} else if !bytes.HasPrefix(seqId, []byte{'@'}) {
		err = errors.New("Invalid format: id line does not start with @")
		return
	} else if len(seqId) == 2 {
		err = errors.New("Invalid format: missing sequence ID")
		return
	}
	curr += len(seqId)

	seqBody, err = self.r.ReadBytes('\n')
	if err == io.EOF {
		err = errors.New("Invalid format: truncated fastq record")
		return
	} else if err != nil {
		return
	} else if len(seqBody) == 1 {
		err = errors.New("Invalid format: empty sequence")
		return
	}
	curr += len(seqBody)

	qualId, err = self.r.ReadBytes('\n')
	if err == io.EOF {
		err = errors.New("Invalid format: truncated fastq record")
		return
	} else if err != nil {
		return
	} else if !bytes.HasPrefix(qualId, []byte{'+'}) {
		err = errors.New("Invalid format: plus line does not start with +")
		return
	}
	qualIdTrim := bytes.TrimSpace(qualId)
	if (len(qualIdTrim) > 1) && (bytes.Compare(bytes.TrimSpace(seqId[1:]), qualIdTrim[1:]) != 0) {
		err = errors.New("Invalid format: quality ID does not match sequence ID")
		return
	}
	curr += len(qualId)

	qualBody, err = self.r.ReadBytes('\n')
	if (err != nil) && (err != io.EOF) {
		return
	}
	if len(bytes.TrimSpace(seqBody)) != len(bytes.TrimSpace(qualBody)) {
		err = errors.New("Invalid format: length of sequence and quality lines do not match")
		return
	}

	n = curr + len(qualBody)
	return
}

// seek sequences which add up to a size close to the configured chunk size (conf.CHUNK_SIZE, e.g. 1M)
func (self *Reader) SeekChunk(offSet int64, lastIndex bool) (n int64, err error) {
	winSize := int64(32768)
	r := io.NewSectionReader(self.f, offSet+conf.CHUNK_SIZE-winSize, winSize)
	buf := make([]byte, winSize)
	if n, err := r.Read(buf); err != nil {
		// EOF reached
		return int64(n), err
	}
	// recursivly extend by window size until new record found
	// first time get last record in window, succesive times get first record
	var loc []int

	if lastIndex {
		locs := Record.FindAllIndex(buf, -1)
		if locs != nil {
			loc = locs[len(locs)-1]
		}
	} else {
		loc = Record.FindIndex(buf)
	}
	if loc == nil {
		indexPos, err := self.SeekChunk(offSet+winSize, false)
		return (winSize + indexPos), err
	}
	// done, last record for this chunk found
	pos := int64(math.Min(float64(loc[1]), float64(len(buf)-1)))
	return conf.CHUNK_SIZE - winSize + pos, nil
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

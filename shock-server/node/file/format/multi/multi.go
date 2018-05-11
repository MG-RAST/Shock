// Package to read and auto-detect format of fasta & fastq files
package multi

import (
	"errors"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/MG-RAST/Shock/shock-server/node/file/format/fasta"
	"github.com/MG-RAST/Shock/shock-server/node/file/format/fastq"
	"github.com/MG-RAST/Shock/shock-server/node/file/format/sam"
	"github.com/MG-RAST/Shock/shock-server/node/file/format/seq"
	"io"
	"regexp"
)

//the order matters as it determines the order for checking format.
var validators = map[string]*regexp.Regexp{
	"fasta": fasta.Regex,
	"fastq": fastq.Regex,
	"sam":   sam.Regex,
}

var readers = map[string]func(f file.SectionReader) seq.ReadRewinder{
	"fasta": fasta.NewReader,
	"fastq": fastq.NewReader,
	"sam":   sam.NewReader,
}

type Reader struct {
	f      file.SectionReader
	r      seq.ReadRewinder
	format string
}

func NewReader(f file.SectionReader) *Reader {
	return &Reader{
		f:      f,
		r:      nil,
		format: "",
	}
}

func (r *Reader) DetermineFormat() error {
	if r.format != "" && r.r != nil {
		return nil
	}

	reader := io.NewSectionReader(r.f, 0, 32768)
	buf := make([]byte, 32768)
	if _, err := reader.Read(buf); err != nil && err != io.EOF {
		return err
	}

	for format, re := range validators {
		if re.Match(buf) {
			r.format = format
			r.r = readers[format](r.f)
			return nil
		}
	}
	return errors.New(e.InvalidFileTypeForFilter)
}

func (r *Reader) Read() (*seq.Seq, error) {
	if r.r == nil {
		err := r.DetermineFormat()
		if err != nil {
			return nil, err
		}
	}
	return r.r.Read()
}

func (r *Reader) GetReadOffset() (n int, err error) {
	if r.r == nil {
		err := r.DetermineFormat()
		if err != nil {
			return 0, err
		}
	}
	return r.r.GetReadOffset()
}

func (r *Reader) SeekChunk(carryOver int64) (n int64, err error) {
	if r.r == nil {
		err := r.DetermineFormat()
		if err != nil {
			return 0, err
		}
	}
	return r.r.SeekChunk(carryOver)
}

func (r *Reader) Format(s *seq.Seq, w io.Writer) (n int, err error) {
	switch {
	case r.format == "fastq":
		return fastq.Format(s, w)
	case r.format == "fasta":
		return fasta.Format(s, w)
	case r.format == "sam":
		return sam.Format(s, w)
	}
	return 0, errors.New("unknown sequence format")
}

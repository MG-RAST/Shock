package multi

import (
	"errors"
	e "github.com/MG-RAST/Shock/errors"
	"github.com/MG-RAST/Shock/store"
	"github.com/MG-RAST/Shock/store/type/sequence/fasta"
	"github.com/MG-RAST/Shock/store/type/sequence/fastq"
	"github.com/MG-RAST/Shock/store/type/sequence/sam"
	"github.com/MG-RAST/Shock/store/type/sequence/seq"
	"io"
	"regexp"
)

//the order matters as it determines the order for checking format.
var valid_format = [3]string{"sam", "fastq", "fasta"}

var readers = map[string]func(f store.SectionReader) seq.ReadRewinder{
	"fasta": fasta.NewReader,
	"fastq": fastq.NewReader,
	"sam":   sam.NewReader,
}

var validators = map[string]*regexp.Regexp{
	"fasta": fasta.Regex,
	"fastq": fastq.Regex,
	"sam":   sam.Regex,
}

type Reader struct {
	f      store.SectionReader
	r      seq.ReadRewinder
	format string
}

func NewReader(f store.SectionReader) *Reader {
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
	if _, err := reader.Read(buf); err != nil {
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

func (r *Reader) ReadRaw(p []byte) (n int, err error) {
	if r.r == nil {
		err := r.DetermineFormat()
		if err != nil {
			return 0, err
		}
	}
	return r.r.ReadRaw(p)
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
	case r.format == "fasta":
		return fasta.Format(s, w)
	case r.format == "fastq":
		return fastq.Format(s, w)
	case r.format == "sam":
		return sam.Format(s, w)
	}
	return 0, errors.New("unknown sequence format")
}

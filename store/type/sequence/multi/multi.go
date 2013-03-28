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
)

//the order matters as it determines the order for checking format.
var valid_format = [3]string{"sam", "fastq", "fasta"}

type Reader struct {
	f       store.SectionReader
	r       seq.ReadRewinder
	formats map[string]seq.ReadRewinder
	format  string
}

func NewReader(f store.SectionReader) *Reader {
	return &Reader{
		f: f,
		r: nil,
		formats: map[string]seq.ReadRewinder{
			"fasta": fasta.NewReader(f),
			"fastq": fastq.NewReader(f),
			"sam":   sam.NewReader(f),
		},
		format: "",
	}
}

func (r *Reader) determineFormat() error {
	if r.format != "" {
		return nil
	}
	for _, f := range valid_format {
		var er error
		reader := r.formats[f]
		for i := 0; i < 1; i++ {
			_, er = reader.Read()
			if er != nil {
				break
			}
		}
		reader.Rewind()
		if er == nil {
			r.r = reader
			r.format = f
			return nil
		}

	}
	return errors.New(e.InvalidFileTypeForFilter)
}

func (r *Reader) Read() (*seq.Seq, error) {
	if r.r == nil {
		err := r.determineFormat()
		if err != nil {
			return nil, err
		}
	}
	return r.r.Read()
}

func (r *Reader) ReadRaw(p []byte) (n int, err error) {
	if r.r == nil {
		err := r.determineFormat()
		if err != nil {
			return 0, err
		}
	}
	return r.r.ReadRaw(p)
}

func (r *Reader) SeekChunk(carryOver int64) (n int64, err error) {
	if r.r == nil {
		err := r.determineFormat()
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

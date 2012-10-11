package multi

import (
	"errors"
	e "github.com/MG-RAST/Shock/errors"
	"github.com/MG-RAST/Shock/store/type/sequence/fasta"
	"github.com/MG-RAST/Shock/store/type/sequence/fastq"
	"github.com/MG-RAST/Shock/store/type/sequence/seq"
	"io"
)

type Reader struct {
	f       io.Reader
	r       seq.ReadRewinder
	formats map[string]seq.ReadRewinder
	format  string
}

func NewReader(f io.Reader) *Reader {
	return &Reader{
		f: f,
		r: nil,
		formats: map[string]seq.ReadRewinder{
			"fasta": fasta.NewReader(f),
			"fastq": fastq.NewReader(f),
		},
		format: "",
	}
}

func (r *Reader) determineFormat() error {
	for f, reader := range r.formats {
		var er error
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

func (r *Reader) Format(s *seq.Seq, w io.Writer) (n int, err error) {
	if r.format == "fasta" {
		return fasta.Format(s, w)
	}
	return fastq.Format(s, w)
}

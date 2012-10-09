package multi

import (
	"errors"
	"fmt"
	e "github.com/MG-RAST/Shock/errors"
	"github.com/MG-RAST/Shock/store/type/sequence/fasta"
	"github.com/MG-RAST/Shock/store/type/sequence/fastq"
	"github.com/MG-RAST/Shock/store/type/sequence/sam"
	"github.com/MG-RAST/Shock/store/type/sequence/seq"
	"io"
)

//the order matters as it determines the order for checking format.
var valid_format = [3]string{"sam", "fastq", "fasta"}

type Reader struct {
	f       io.ReadCloser
	r       seq.ReadRewindCloser
	formats map[string]seq.ReadRewindCloser
	format  string
}

func NewReader(f io.ReadCloser) *Reader {
	return &Reader{
		f: f,
		r: nil,
		formats: map[string]seq.ReadRewindCloser{
			"fasta": fasta.NewReader(f),
			"fastq": fastq.NewReader(f),
			"sam":   sam.NewReader(f),
		},
		format: "",
	}
}

func (r *Reader) determineFormat() error {
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

// This cause server to hang without even printing debug
// statements. Not sure wtf is up with that.
func (r *Reader) Close() error {
	fmt.Println("Starting close")
	for format, reader := range r.formats {
		fmt.Println("Closing:", format)
		reader.Close()
	}
	fmt.Println("Done close")
	return nil
}

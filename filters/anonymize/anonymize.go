package anonymize

import (	
	"io"
	"errors"
	e "github.com/MG-RAST/Shock/errors"
	"github.com/kortschak/BioGo/io/seqio/fasta"
	"github.com/kortschak/BioGo/io/seqio/fastq"
	"github.com/kortschak/BioGo/io/seqio"
	"fmt"
)

type Reader struct {
	f		io.ReadCloser
	r		seqio.Reader
	rfasta	*fasta.Reader
	rfastq	*fastq.Reader
	counter	int
}

func NewReader(f io.ReadCloser) *Reader {
	return &Reader{ f: f, r: nil, rfasta: fasta.NewReader(f), rfastq: fastq.NewReader(f), counter: 1}
}

func (r *Reader) determineSeqType() (err error){	
	_, fastaE := r.rfasta.Read()
	_, fastqE := r.rfastq.Read()
	if fastaE != nil && fastqE != nil {
		err = errors.New(e.InvalidFileTypeForFilter)
	} else if fastaE != nil {
		err = r.rfastq.Rewind(); if err != nil {
			return
		}
		r.r = r.rfastq
	} else {
		err = r.rfastq.Rewind(); if err != nil {
			return
		}
		r.r = r.rfasta
	}
	return
}

func (r *Reader) Read(p []byte) (n int, err error) {
	if r.r == nil {
		err = r.determineSeqType(); if err != nil {
			return
		}
	} 
	seq, err := r.r.Read(); if err != nil {
		return
	}
	seq.ID = fmt.Sprintf("%d", r.counter)
	r.counter += 1
	record := []byte(fmt.Sprintf(">%s\n%s\n",seq.ID, seq.Seq))
	copy(p[0:len(record)],record)
	n = len(record)
	return
}

func (r *Reader) Close() {
	r.rfasta.Close()
	r.rfastq.Close()
	return
}



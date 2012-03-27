package anonymize

import (
	"fmt"
	"github.com/MG-RAST/Shock/types/sequence"
	"io"
)

type Reader struct {
	f        io.ReadCloser
	r        *sequence.Reader
	counter  int
	overflow []byte
}

func NewReader(f io.ReadCloser) io.ReadCloser {
	return &Reader{
		f:        f,
		r:        sequence.NewReader(f),
		counter:  1,
		overflow: nil,
	}
}

func (r *Reader) Read(p []byte) (n int, err error) {
	ln := 0
	if r.overflow != nil {
		copy(p[0:len(r.overflow)], r.overflow)
		ln = len(r.overflow)
	}
	for {
		seq, er := r.r.Read()
		if er != nil {
			err = er
			break
		}
		seq.ID = fmt.Sprintf("%d", r.counter)
		r.counter += 1
		record := []byte(r.r.Format(seq))
		if ln+len(record) < cap(p) {
			copy(p[ln:ln+len(record)], record)
			ln = ln + len(record)
		} else {
			r.overflow = record
			break
		}
	}
	n = ln
	return
}

func (r *Reader) Close() error {
	r.Close()
	return nil
}

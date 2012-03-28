package anonymize

import (
	"bytes"
	"fmt"
	"github.com/MG-RAST/Shock/types/sequence/seq"
	"github.com/MG-RAST/Shock/types/sequence/multi"
	"io"
)

type Reader struct {
	f        io.ReadCloser
	r        seq.ReadCloser
	counter  int
	overflow []byte
}

func NewReader(f io.ReadCloser) io.ReadCloser {
	return &Reader{
		f:        f,
		r:        multi.NewReader(f),
		counter:  1,
		overflow: nil,
	}
}

func (r *Reader) Read(p []byte) (n int, err error) {
	n = 0
	buf := bytes.NewBuffer(nil)
	if r.overflow != nil {
		ln, _ := buf.Write(r.overflow)
		n += ln
	}
	for {
		seq, er := r.r.Read()
		if er != nil {
			if er == io.EOF {
				copy(p[0:n],buf.Bytes()[0:n])
			} 
			err = er
			break
		}
		seq.ID = []byte(fmt.Sprint(r.counter))
		r.counter += 1
		ln, _ := r.r.Format(seq, buf)
		if n+ln > cap(p) {
			copy(p[0:n],buf.Bytes()[0:n])
			r.overflow = buf.Bytes()[n:]
			break
		} else {			
			n += ln
		}	
	}
	return
}

func (r *Reader) Close() error {
	r.Close()
	return nil
}

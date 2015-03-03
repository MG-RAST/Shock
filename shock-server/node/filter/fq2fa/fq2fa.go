package fq2fa

import (
	"bytes"
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/MG-RAST/Shock/shock-server/node/file/format/fasta"
	"github.com/MG-RAST/Shock/shock-server/node/file/format/fastq"
	"github.com/MG-RAST/Shock/shock-server/node/file/format/seq"
	"io"
)

type Reader struct {
	f        file.SectionReader
	r        seq.Reader
	overflow []byte
}

func NewReader(f file.SectionReader, n string) io.Reader {
	return &Reader{
		f:        f,
		r:        fastq.NewReader(f),
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
				copy(p[0:n], buf.Bytes()[0:n])
			}
			err = er
			break
		}
		ln, _ := fasta.Format(seq, buf)
		if n+ln > cap(p) {
			copy(p[0:n], buf.Bytes()[0:n])
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

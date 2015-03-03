package zip

import (
	"archive/zip"
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"io"
)

type Reader struct {
	f file.SectionReader
	r *io.PipeReader
}

func NewReader(f file.SectionReader, n string) io.Reader {
	pr, pw := io.Pipe()
	zw := zip.NewWriter(pw)
	zf, _ := zw.Create(n)
	go func() {
		io.Copy(zf, f)
		zw.Close()
		pw.Close()
	}()
	return &Reader{
		f: f,
		r: pr,
	}
}

func (r *Reader) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	return
}

func (r *Reader) Close() error {
	r.r.Close()
	return nil
}

package httpclient

import (
	"bytes"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

type Form struct {
	params      []param
	files       []file
	handles     []*os.File
	Length      int64
	Reader      io.Reader
	ContentType string
}

type param struct {
	k string
	v string
}

type file struct {
	n string
	p string
}

func NewForm() (f *Form) {
	return &Form{}
}

func (f *Form) AddFile(name, path string) {
	f.files = append(f.files, file{n: name, p: path})
	return
}

func (f *Form) AddParam(key, val string) {
	f.params = append(f.params, param{k: key, v: val})
	return
}

func (f *Form) Close() {
	for _, fh := range f.handles {
		fh.Close()
	}
}

func (f *Form) Create() (err error) {
	readers := []io.Reader{}
	buf := bytes.NewBufferString("")
	writer := multipart.NewWriter(buf)

	for _, p := range f.params {
		writer.WriteField(p.k, p.v)
	}
	readers = append(readers, bytes.NewBufferString(buf.String()))
	f.Length += int64(buf.Len())
	buf.Reset()

	for _, file := range f.files {
		writer.CreateFormFile(file.n, filepath.Base(file.p))
		readers = append(readers, bytes.NewBufferString(buf.String()))
		f.Length += int64(buf.Len())
		buf.Reset()
		if fh, err := os.Open(file.p); err == nil {
			if fi, err := fh.Stat(); err == nil {
				f.Length += fi.Size()
			} else {
				return err
			}
			readers = append(readers, fh)
			f.handles = append(f.handles, fh)
		} else {
			return err
		}
	}

	writer.Close()
	readers = append(readers, bytes.NewBufferString(buf.String()))
	f.Length += int64(buf.Len())

	f.Reader = io.MultiReader(readers...)
	f.ContentType = writer.FormDataContentType()
	return
}

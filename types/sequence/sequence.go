package sequence

import (
	"bytes"
	"errors"
	"fmt"
	e "github.com/MG-RAST/Shock/errors"
	"github.com/kortschak/BioGo/io/seqio"
	"github.com/kortschak/BioGo/io/seqio/fasta"
	"github.com/kortschak/BioGo/io/seqio/fastq"
	seq "github.com/kortschak/BioGo/seq"
	"io"
)

type Reader struct {
	f       io.ReadCloser
	r       seqio.Reader
	formats map[string]seqio.Reader
	format  string
}

type BufferWriteCloser struct {
	Buf *bytes.Buffer
}

func NewBufferWriteCloser(buf []byte) *BufferWriteCloser {
	return &BufferWriteCloser{Buf: bytes.NewBuffer(buf)}
}

func (b *BufferWriteCloser) Write(p []byte) (int, error) {
	fmt.Println("write - p:", p)
	return b.Buf.Write(p)
}

func (b *BufferWriteCloser) WriteString(s string) (int, error) {
	fmt.Println("writestring - s:", s)
	return b.Buf.WriteString(s)
}

func (b *BufferWriteCloser) Close() error {
	return nil
}

func NewReader(f io.ReadCloser) *Reader {
	//fastq := fastq.NewReader(f)
	//fastq.Encoding = seq.Illumina1_8
	return &Reader{
		f: f,
		r: nil,
		formats: map[string]seqio.Reader{
			"fasta": fasta.NewReader(f),
			//"fastq": fastq,
		},
		format: "",
	}
}

func (r *Reader) determineFormat() (err error) {
	for f, reader := range r.formats {
		_, error := reader.Read()
		if error == nil {
			reader.Rewind()
			r.r = reader
			r.format = f
			return
		} else {
			fmt.Println(error.Error())
		}
	}
	err = errors.New(e.InvalidFileTypeForFilter)
	return
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

func (r *Reader) Format(s *seq.Seq) (record string) {
	if r.format == "fasta" {
		record = fasta.Format(s, 0)
	} else if r.format == "fastq" {
		record = fastq.Format(s, false, seq.Illumina1_8)
	}
	return
}

// This cause server to hang without even printing debug
// statements. Not sure wtf is up with that.
func (r *Reader) Close() {
	fmt.Println("Starting close")
	for format, reader := range r.formats {
		fmt.Println("Closing:", format)
		reader.Close()
	}
	fmt.Println("Done close")
	return
}

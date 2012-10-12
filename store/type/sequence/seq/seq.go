package seq

import (
	"io"
)

type Seq struct {
	ID   []byte
	Seq  []byte
	Qual []byte
}

func New(id []byte, seq []byte, qual []byte) *Seq {
	return &Seq{
		ID:   id,
		Seq:  seq,
		Qual: qual,
	}
}

type ReadFormater interface {
	Read() (*Seq, error)
	Format(*Seq, io.Writer) (int, error)
}

type Reader interface {
	Read() (*Seq, error)
	ReadRaw(p []byte) (int, error)
}

type ReadRewinder interface {
	Read() (*Seq, error)
	ReadRaw(p []byte) (int, error)
	Rewind() error
}

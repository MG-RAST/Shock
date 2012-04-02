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

type ReadFormatCloser interface {
	Read() (*Seq, error)
	Format(*Seq, io.Writer) (int, error)
	Close() error
}

type ReadCloser interface {
	Read() (*Seq, error)
	Close() error
}

type ReadRewindCloser interface {
	Read() (*Seq, error)
	Rewind() error
	Close() error
}

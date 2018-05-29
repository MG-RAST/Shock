// Package contains interfaces for fasta & fastq & and multi packages
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
	GetReadOffset() (int, error)
	SeekChunk(int64, bool) (int64, error)
}

type ReadRewinder interface {
	Read() (*Seq, error)
	GetReadOffset() (int, error)
	SeekChunk(int64, bool) (int64, error)
	Rewind() error
}

package subset

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

const chunkSize int64 = 32768

type part struct {
	offset []int64
	length int64
	n      int
}

type subset struct {
	b       [][]byte
	read    []int
	offsets []int64
	fh      []*os.File
	fi      []os.FileInfo
	parts   []*part
	current *part
}

func NewSubset(a *os.File, b *os.File) *subset {
	return &subset{
		b:       [][]byte{make([]byte, chunkSize), make([]byte, chunkSize)},
		read:    []int{0, 0},
		offsets: []int64{0, 0},
		fh:      []*os.File{a, b},
		fi:      []os.FileInfo{},
		parts:   []*part{},
		current: nil,
	}
}

func (s *subset) bufA() []byte {
	return s.b[0][0:s.read[0]]
}

func (s *subset) bufB() []byte {
	return s.b[1][0:s.read[1]]
}

func (s *subset) Read() (err error) {
	a := io.NewSectionReader(s.fh[0], s.offsets[0], chunkSize)
	b := io.NewSectionReader(s.fh[1], s.offsets[1], chunkSize)
	s.read[0], err = a.Read(s.b[0])
	if err != nil && err != io.EOF {
		return err
	}
	s.read[1], err = b.Read(s.b[1])
	if err != nil && err != io.EOF {
		return err
	}
	return
}

func (s *subset) overlap() {
	a, b := s.bufA(), s.bufB()
	length := int64(0)
	for i := 0; i < s.read[0] && i < s.read[1]; i += 1 {
		if a[i] != b[i] {
			break
		} else {
			length += 1
		}
	}
	s.current.length += length
	if s.current.length > 0 {
		s.parts = append(s.parts, s.current)
	} else {
		//println("not found")
	}
	s.current = nil
}

func (s *subset) key() (key []byte, err error) {
	last := s.parts[len(s.parts)-1]
	key = make([]byte, 100)
	b := io.NewSectionReader(s.fh[1], last.offset[1]+last.length, 100)
	n, err := b.Read(key)
	return key[0:n], err
}

func (s *subset) seak() {
	last := s.parts[len(s.parts)-1]
	s.offsets[1] = last.offset[1] + last.length
	key, _ := s.key()
	if len(key) > 0 {
		if i := bytes.Index(s.bufA()[last.length:], key); i != -1 {
			s.offsets[0] = last.offset[0] + last.length + int64(i)
		} else {
			//fmt.Printf("%s\n", key)
		}
	}
}

// read, [match, seek] till empty. repeat. validate
func (s *subset) findParts() (err error) {
	if s.current == nil {
		s.current = &part{offset: []int64{s.offsets[0], s.offsets[1]}, n: len(s.parts)}
	}
	s.Read()
	if s.read[1] > 0 {
		if bytes.EqualFold(s.bufA(), s.bufB()) {
			s.current.length += int64(s.read[0])
			s.offsets[0] += int64(s.read[0])
			s.offsets[1] += int64(s.read[1])
		} else {
			s.overlap()
			s.seak()
		}
		return s.findParts()
	} else {
		if s.current.length > 0 {
			s.parts = append(s.parts, s.current)
		}
	}
	return
}

func (s *subset) Validate() (err error) {
	if len(s.parts) == 0 {
		return errors.New("Failed validation: index is empty")
	}

	size := int64(0)
	for _, p := range s.parts {
		size += p.length
	}

	if size != s.fi[1].Size() {
		return errors.New("Failed validation: index size does not equal target size")
	}
	return nil
}

func (s *subset) write(path string) (err error) {
	if f, err := os.Create(path); err != nil {
		return err
	} else {
		defer f.Close()
		for _, p := range s.parts {
			binary.Write(f, binary.LittleEndian, p.offset)
			binary.Write(f, binary.LittleEndian, p.length)
		}
	}
	return
}

func (s *subset) Create(path string) (err error) {
	for _, fh := range s.fh {
		if fi, err := fh.Stat(); err == nil {
			s.fi = append(s.fi, fi)
		} else {
			return err
		}
	}
	if err = s.findParts(); err != nil {
		return errors.New("Failed index creation: " + err.Error())
	}
	for _, p := range s.parts {
		fmt.Printf("%d %d %d\n", p.n, p.offset[0], p.length)
	}
	if err = s.Validate(); err != nil {
		return err
	}
	err = s.write(path)
	return
}

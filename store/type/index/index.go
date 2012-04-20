package index

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
)

type Indexer interface {
	Dump(string) error
	Create() error
	Close() error
}

type Index interface {
	Set(map[string]interface{})
	Type() string
	Append([]int64)
	Part(string) (int64, int64, error)
	Dump(string) error
	Load(string) error
}

type Idx struct {
	T      string
	Idx    [][]int64
	Length int
}

func New() *Idx {
	return &Idx{
		T:      "file",
		Idx:    [][]int64{},
		Length: 0,
	}
}

func (i *Idx) Append(rec []int64) {
	i.Idx = append(i.Idx, rec)
	i.Length += 1
}

func (i *Idx) Set(inter map[string]interface{}) {
	return
}

func (i *Idx) Type() string {
	return i.T
}

func (i *Idx) Part(part string) (pos int64, length int64, err error) {
	if strings.Contains(part, "-") {
		startend := strings.Split(part, "-")
		start, startEr := strconv.ParseInt(startend[0], 10, 64)
		end, endEr := strconv.ParseInt(startend[1], 10, 64)
		if startEr != nil || endEr != nil || start <= 0 || start > int64(i.Length) || end <= 0 || end > int64(i.Length) {
			err = errors.New("")
			return
		}
		pos = i.Idx[(start - 1)][0]
		length = (i.Idx[(end - 1)][0] - i.Idx[(start - 1)][0]) + i.Idx[(end - 1)][1]
	} else {
		p, er := strconv.ParseInt(part, 10, 64)
		if er != nil || p <= 0 || p > int64(i.Length) {
			err = errors.New("")
			return
		}
		pos = i.Idx[(p - 1)][0]
		length = i.Idx[(p - 1)][1]
	}
	return
}

func (i *Idx) Dump(file string) (err error) {
	f, err := os.Create(file)
	defer f.Close()
	if err != nil {
		return
	}
	for _, rec := range i.Idx {
		binary.Write(f, binary.LittleEndian, rec[0])
		binary.Write(f, binary.LittleEndian, rec[1])
	}
	return
}

func (i *Idx) Load(file string) (err error) {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		return
	}
	for {
		rec := make([]int64, 2)
		er := binary.Read(f, binary.LittleEndian, &rec[0])
		if er != nil {
			if er != io.EOF {
				err = er
			}
			return
		}
		er = binary.Read(f, binary.LittleEndian, &rec[1])
		if er != nil {
			if er != io.EOF {
				err = er
			}
			return
		}
		i.Append(rec)
	}
	return
}

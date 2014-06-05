package index

import (
	"errors"
	"strconv"
	"strings"
)

type partFunc func(string, *vIndex) (int64, int64, error)

var (
	virtual = map[string]partFunc{
		"size": SizePart,
	}
)

func Has(t string) bool {
	if _, has := virtual[t]; has {
		return true
	}
	return false
}

type vIndex struct {
	T         string
	path      string
	size      int64
	ChunkSize int64
	partF     partFunc
}

func NewVirtual(t string, p string, s int64, c int64) *vIndex {
	if partFunc, has := virtual[t]; has {
		return &vIndex{
			T:         "virtual",
			path:      p,
			size:      s,
			ChunkSize: c,
			partF:     partFunc,
		}
	}
	return &vIndex{}
}

func (v *vIndex) Part(p string) (pos int64, length int64, err error) {
	return v.partF(p, v)
}

func SizePart(part string, v *vIndex) (pos int64, length int64, err error) {
	if strings.Contains(part, "-") {
		startend := strings.Split(part, "-")
		start, startEr := strconv.ParseInt(startend[0], 10, 64)
		end, endEr := strconv.ParseInt(startend[1], 10, 64)
		if startEr != nil || endEr != nil || start <= 0 || (start-1)*v.ChunkSize > v.size || end <= 0 || (end-1)*v.ChunkSize > v.size {
			err = errors.New("")
			return
		}
		pos = (start - 1) * v.ChunkSize
		fullReads := (end-1)*v.ChunkSize - (start-1)*v.ChunkSize
		if fullReads+v.ChunkSize+pos > v.size {
			length = fullReads + (v.size - (pos + fullReads))
		} else {
			length = fullReads + v.ChunkSize
		}
	} else {
		p, er := strconv.ParseInt(part, 10, 64)
		if er != nil || p <= 0 || (p-1)*v.ChunkSize > v.size {
			err = errors.New("")
			return
		}
		pos = (p - 1) * v.ChunkSize
		if v.size-pos < v.ChunkSize {
			length = v.size - pos
		} else {
			length = v.ChunkSize
		}
	}
	return
}

func (v *vIndex) Set(i map[string]interface{}) {
	if cv, has := i["ChunkSize"]; has {
		if chunksize, ok := cv.(int64); ok {
			v.ChunkSize = chunksize
		}
	}
	return
}

func (v *vIndex) Type() string {
	return v.T
}

func (v *vIndex) GetLength() int64 {
	return v.size
}

// Empty functions to fulfil interface
func (v *vIndex) Append(a []int64) {
	return
}

func (v *vIndex) Load(string) error {
	return nil
}

func (v *vIndex) Range(string) ([][]int64, error) {
	return nil, nil
}

func (v *vIndex) DynamicPart(string, string, int64) (int64, int64, error) {
	return -1, -1, nil
}

func (v *vIndex) DynamicRange(string, string, int64) ([][]int64, error) {
	return nil, nil
}

package index

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/node/file/format/line"
	"io"
	"math/rand"
	"os"
	"strconv"
)

type subset struct {
	f     *os.File
	r     line.LineReader
	Index *Idx
}

func NewSubsetIndexer(f *os.File) subset {
	return subset{
		f:     f,
		r:     line.NewReader(f),
		Index: New(),
	}
}

func (s *subset) Create(string) (count int64, err error) {
	return
}

func CreateSubsetIndex(s *subset, ofile string, ifile string) (count int64, size int64, format string, err error) {
	tmpFilePath := fmt.Sprintf("%s/temp/%d%d.idx", conf.Conf["data-path"], rand.Int(), rand.Int())

	f, err := os.Create(tmpFilePath)
	if err != nil {
		return
	}
	defer f.Close()

	parent_idx := New()
	err = parent_idx.Load(ifile)
	if err != nil {
		return
	}

	count = 0
	size = 0
	format = "array"
	prev_int := int(0)
	for {
		buf, er := s.r.ReadLine()
		n := len(buf)
		if er != nil {
			if er != io.EOF {
				err = er
				return
			}
			break
		}
		// skip empty line
		if n <= 1 {
			continue
		}
		// int from line
		str := string(buf[:n-1])
		curr_int, er := strconv.Atoi(str)
		if er != nil {
			err = er
			return
		}

		if curr_int <= prev_int {
			err = errors.New(fmt.Sprintf("Subset indices must be numerically sorted and non-redundant, found value %d after value %d", curr_int, prev_int))
			return
		}

		if curr_int > parent_idx.Length {
			err = errors.New(fmt.Sprintf("Subset index: %d does not exist in parent index file.", curr_int))
			return
		}

		offset := parent_idx.Idx[curr_int-1][0]
		length := parent_idx.Idx[curr_int-1][1]
		binary.Write(f, binary.LittleEndian, offset)
		binary.Write(f, binary.LittleEndian, length)
		count += 1
		size += length
		prev_int = curr_int
	}
	err = os.Rename(tmpFilePath, ofile)

	return
}

func CreateSubsetNodeIndexes(s *subset, cofile string, ofile string, ifile string) (coCount int64, oCount int64, oSize int64, oFormat string, err error) {
	// create temporary output file (oTmpFilePath) for subset index and temporary output file (coTmpFilePath) for compressed subset index
	oTmpFilePath := fmt.Sprintf("%s/temp/%d%d.idx", conf.Conf["data-path"], rand.Int(), rand.Int())
	coTmpFilePath := fmt.Sprintf("%s/temp/%d%d.idx", conf.Conf["data-path"], rand.Int(), rand.Int())

	of, err := os.Create(oTmpFilePath)
	if err != nil {
		return
	}
	defer of.Close()

	cof, err := os.Create(coTmpFilePath)
	if err != nil {
		return
	}
	defer cof.Close()

	parent_idx := New()
	err = parent_idx.Load(ifile)
	if err != nil {
		return
	}

	// these are attributes of the subset index output, "matrix" format will be supported in the future
	oCount = 0
	oSize = 0
	oFormat = "array"

	// coCount is the number of indices created for the compressed output index
	coCount = 0

	//coOffset and coLength store the offset and length for a given compressed output index entry
	coOffset := int64(0)
	coLength := int64(0)

	// these store the previous offset and length for concatenating contiguous reads into one entry for compressed index
	prevOffset := int64(0)
	prevLength := int64(0)

	// stores previous integer from subset file line
	prev_int := int(0)

	for {
		buf, er := s.r.ReadLine()
		n := len(buf)
		if er != nil {
			if er != io.EOF {
				err = er
				return
			}
			break
		}
		// skip empty line
		if n <= 1 {
			continue
		}
		// int from line
		str := string(buf[:n-1])
		curr_int, er := strconv.Atoi(str)
		if er != nil {
			err = er
			return
		}

		if curr_int <= prev_int {
			err = errors.New(fmt.Sprintf("Subset indices must be numerically sorted and non-redundant, found value %d after value %d", curr_int, prev_int))
			return
		}

		if curr_int > parent_idx.Length {
			err = errors.New(fmt.Sprintf("Subset index: %d does not exist in parent index file.", curr_int))
			return
		}

		offset := parent_idx.Idx[curr_int-1][0]
		length := parent_idx.Idx[curr_int-1][1]
		binary.Write(of, binary.LittleEndian, offset)
		binary.Write(of, binary.LittleEndian, length)
		oCount += 1
		oSize += length

		// compressed index handling
		if prev_int != 0 && offset != prevOffset+prevLength {
			binary.Write(cof, binary.LittleEndian, coOffset)
			binary.Write(cof, binary.LittleEndian, coLength)
			coOffset = offset
			coLength = length
			coCount += 1
		} else if prev_int == 0 {
			coOffset = offset
			coLength += length
		} else {
			coLength += length
		}

		prev_int = curr_int
		prevOffset = offset
		prevLength = length
	}

	binary.Write(cof, binary.LittleEndian, coOffset)
	binary.Write(cof, binary.LittleEndian, coLength)

	err = os.Rename(coTmpFilePath, cofile)
	if err != nil {
		return
	}
	err = os.Rename(oTmpFilePath, ofile)

	return
}

func (s *subset) Close() (err error) {
	s.f.Close()
	return
}

package index

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"

	e "github.com/MG-RAST/Shock/shock-server/errors"
)

// indexerFunc Constructor function of Indexer objects
//   example call: (f, n.Type, n.Subset.Index.Format, n.IndexPath()+"/"+n.Subset.Parent.IndexName+".idx")
//   nType: node type
//   snFormat: "sn=subset node" index format, e.g. matrix, array
//   snIndexPath: subset node index file location
type indexerFunc func(*os.File, string, string, string) Indexer

var (
	Indexers = map[string]indexerFunc{
		"chunkrecord": NewChunkRecordIndexer,
		"line":        NewLineIndexer,
		"record":      NewRecordIndexer,
		"size":        NewSizeIndexer,
	}
)

type Indexer interface {
	Create(string) (int64, string, error) // actual file parsing
	Close() error
}

type Index interface {
	Set(map[string]interface{})
	Type() string
	GetLength() int64
	Part(string, string, int64) (int64, int64, error)
	Range(string, string, int64) ([][]int64, error)
}

type Idx struct {
	T      string
	Length int
}

func New() *Idx {
	return &Idx{
		T:      "file",
		Length: 0,
	}
}

func (i *Idx) Set(inter map[string]interface{}) {
	return
}

func (i *Idx) Type() string {
	return i.T
}

func (i *Idx) GetLength() int64 {
	return int64(i.Length)
}

func (i *Idx) Part(part string, idxFilePath string, idxLength int64) (pos int64, length int64, err error) {
	// this function is for returning a single pos and length for a given range
	// used for non-subset indices where the records are contiguous for the data file
	f, err := os.Open(idxFilePath)
	if err != nil {
		err = errors.New(e.IndexNoFile)
		return
	}
	defer f.Close()

	if strings.Contains(part, "-") {
		startend := strings.Split(part, "-")
		start, startEr := strconv.ParseInt(startend[0], 10, 64)
		end, endEr := strconv.ParseInt(startend[1], 10, 64)
		if startEr != nil || endEr != nil || start <= 0 || start > int64(idxLength) || end <= 0 || end > int64(idxLength) {
			err = errors.New(e.InvalidIndexRange)
			return
		}

		// read start offset and length from index file
		sr := io.NewSectionReader(f, (start-1)*16, 16)
		srec := make([]int64, 2)
		binary.Read(sr, binary.LittleEndian, &srec[0])
		binary.Read(sr, binary.LittleEndian, &srec[1])

		// read end offset and length from index file
		sr = io.NewSectionReader(f, (end-1)*16, 16)
		erec := make([]int64, 2)
		binary.Read(sr, binary.LittleEndian, &erec[0])
		binary.Read(sr, binary.LittleEndian, &erec[1])

		pos = srec[0]
		length = (erec[0] - srec[0]) + erec[1]
	} else {
		p, er := strconv.ParseInt(part, 10, 64)
		if er != nil || p <= 0 || p > int64(idxLength) {
			err = errors.New(e.IndexOutBounds)
			return
		}

		// read offset and length from index file
		sr := io.NewSectionReader(f, (p-1)*16, 16)
		rec := make([]int64, 2)
		binary.Read(sr, binary.LittleEndian, &rec[0])
		binary.Read(sr, binary.LittleEndian, &rec[1])

		pos = rec[0]
		length = rec[1]
	}
	return
}

func (i *Idx) Range(part string, idxFilePath string, idxLength int64) (recs [][]int64, err error) {
	// this function is for returning an array of [pos, length] for a given range
	// used for subset indices where the records are not contiguous for the data file
	f, err := os.Open(idxFilePath)
	if err != nil {
		err = errors.New(e.IndexNoFile)
		return
	}
	defer f.Close()

	if strings.Contains(part, "-") {
		startend := strings.Split(part, "-")
		start, startEr := strconv.ParseInt(startend[0], 10, 64)
		end, endEr := strconv.ParseInt(startend[1], 10, 64)
		if startEr != nil || endEr != nil || start <= 0 || start > int64(idxLength) || end <= 0 || end > int64(idxLength) {
			err = errors.New(e.InvalidIndexRange)
			return
		}

		// read beginning offset and length from index file
		sr := io.NewSectionReader(f, (start-1)*16, 16)
		rec := make([]int64, 2)
		binary.Read(sr, binary.LittleEndian, &rec[0])
		binary.Read(sr, binary.LittleEndian, &rec[1])

		curPos := rec[0]
		curLen := rec[1]
		// we only have one record
		if start == end {
			recs = append(recs, []int64{curPos, curLen})
			return
		}
		// this loop tries to only return seperate [pos, length] sets for non-contiguous records
		for x := start; x <= end-1; x++ {
			// reading next offset and length from index file
			sr = io.NewSectionReader(f, x*16, 16)
			binary.Read(sr, binary.LittleEndian, &rec[0])
			binary.Read(sr, binary.LittleEndian, &rec[1])

			nextPos := rec[0]
			nextLen := rec[1]
			// special case - last record
			if x == (end - 1) {
				if curLen == (nextPos - curPos) {
					recs = append(recs, []int64{curPos, curLen + nextLen})
				} else {
					recs = append(recs, []int64{curPos, curLen})
					recs = append(recs, []int64{nextPos, nextLen})
				}
				break
			}
			if curLen == (nextPos - curPos) {
				curLen = curLen + nextLen
				continue
			}
			recs = append(recs, []int64{curPos, curLen})
			curPos = nextPos
			curLen = nextLen
		}
	} else {
		p, er := strconv.ParseInt(part, 10, 64)
		if er != nil || p <= 0 || p > int64(idxLength) {
			err = errors.New(e.IndexOutBounds)
			return
		}

		// read offset and length from index file
		sr := io.NewSectionReader(f, (p-1)*16, 16)
		rec := make([]int64, 2)
		binary.Read(sr, binary.LittleEndian, &rec[0])
		binary.Read(sr, binary.LittleEndian, &rec[1])

		recs = append(recs, []int64{rec[0], rec[1]})
	}
	return
}

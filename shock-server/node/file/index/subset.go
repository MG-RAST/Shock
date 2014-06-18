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

// We anticipate that the input index format (iformat) will be either "array" or "matrix" and for
// subset indexes, the output format should be the same as the input format.  Only input index format
// type "array" is currently supported.
func CreateSubsetIndex(s *subset, oifile string, ifile string, iformat string, ilength int64) (count int64, size int64, err error) {
	if iformat == "array" {
		tmpFilePath := fmt.Sprintf("%s/temp/%d%d.idx", conf.Conf["data-path"], rand.Int(), rand.Int())

		ifh, err := os.Open(ifile)
		if err != nil {
			return -1, -1, err
		}
		defer ifh.Close()

		ofh, err := os.Create(tmpFilePath)
		if err != nil {
			return -1, -1, err
		}
		defer ofh.Close()

		count = 0
		size = 0
		prev_int := int(0)
		buffer_pos := 0 // used to track the location in our output byte array

		// Writing index file in 16MB chunks
		var b [16777216]byte
		for {
			buf, er := s.r.ReadLine()
			n := len(buf)
			if er != nil {
				if er != io.EOF {
					err = er
					return -1, -1, err
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
				return -1, -1, err
			}

			if curr_int <= prev_int {
				err = errors.New(fmt.Sprintf("Subset indices must be numerically sorted and non-redundant, found value %d after value %d", curr_int, prev_int))
				return -1, -1, err
			}

			if int64(curr_int) > ilength {
				err = errors.New(fmt.Sprintf("Subset index: %d does not exist in parent index file.", curr_int))
				return -1, -1, err
			}

			var ibuf [16]byte
			_, er = ifh.ReadAt(ibuf[0:16], int64((curr_int-1)*16))
			if er != nil {
				err = errors.New(fmt.Sprintf("Subset index could not read parent index file for part: %d", curr_int))
				return -1, -1, err
			}

			offset := int64(binary.LittleEndian.Uint64(ibuf[0:8]))
			length := int64(binary.LittleEndian.Uint64(ibuf[8:16]))

			x := (buffer_pos * 16)
			if x == 16777216 {
				ofh.Write(b[:])
				buffer_pos = 0
				x = 0
			}
			y := x + 8
			z := x + 16

			binary.LittleEndian.PutUint64(b[x:y], uint64(offset))
			binary.LittleEndian.PutUint64(b[y:z], uint64(length))

			count += 1
			size += length
			prev_int = curr_int
			buffer_pos += 1
		}
		if buffer_pos != 0 {
			ofh.Write(b[:buffer_pos*16])
		}

		err = os.Rename(tmpFilePath, oifile)

		return count, size, err
	} else {
		return -1, -1, errors.New("Subset index does not currently support the format of your parent index: " + iformat)
	}
}

// We anticipate that the input index format (iformat) will be either "array" or "matrix" and for
// subset nodes, the output index format should be the same as the input format.  Only input index
// format type "array" is currently supported.
func CreateSubsetNodeIndexes(s *subset, cofile string, ofile string, ifile string, iformat string, ilength int64) (coCount int64, oCount int64, oSize int64, err error) {
	if iformat == "array" {
		// create temporary output file (oTmpFilePath) for subset index and temporary output file (coTmpFilePath) for compressed subset index
		oTmpFilePath := fmt.Sprintf("%s/temp/%d%d.idx", conf.Conf["data-path"], rand.Int(), rand.Int())
		coTmpFilePath := fmt.Sprintf("%s/temp/%d%d.idx", conf.Conf["data-path"], rand.Int(), rand.Int())

		var ifh *os.File
		ifh, err = os.Open(ifile)
		if err != nil {
			return
		}
		defer ifh.Close()

		var ofh *os.File
		ofh, err = os.Create(oTmpFilePath)
		if err != nil {
			return
		}
		defer ofh.Close()

		var cofh *os.File
		cofh, err = os.Create(coTmpFilePath)
		if err != nil {
			return
		}
		defer cofh.Close()

		// these are attributes of the subset index output, "matrix" format will be supported in the future
		oCount = 0
		oSize = 0

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

		// used to track the location in our output byte array
		buffer_pos1 := 0
		buffer_pos2 := 0

		// Writing index files in 16MB chunks
		var b1 [16777216]byte
		var b2 [16777216]byte

		for {
			buf, er := s.r.ReadLine()
			n := len(buf)
			if er != nil {
				if er != io.EOF {
					err = er
					return -1, -1, -1, err
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

			if int64(curr_int) > ilength {
				err = errors.New(fmt.Sprintf("Subset index: %d does not exist in parent index file.", curr_int))
				return
			}

			var ibuf [16]byte
			_, er = ifh.ReadAt(ibuf[0:16], int64((curr_int-1)*16))
			if er != nil {
				err = errors.New(fmt.Sprintf("Subset index could not read parent index file for part: %d", curr_int))
				return
			}

			offset := int64(binary.LittleEndian.Uint64(ibuf[0:8]))
			length := int64(binary.LittleEndian.Uint64(ibuf[8:16]))

			x := (buffer_pos1 * 16)
			if x == 16777216 {
				ofh.Write(b1[:])
				buffer_pos1 = 0
				x = 0
			}
			y := x + 8
			z := x + 16

			binary.LittleEndian.PutUint64(b1[x:y], uint64(offset))
			binary.LittleEndian.PutUint64(b1[y:z], uint64(length))

			oCount += 1
			oSize += length
			buffer_pos1 += 1

			// compressed index handling
			if prev_int != 0 && offset != prevOffset+prevLength {
				x := (buffer_pos2 * 16)
				if x == 16777216 {
					cofh.Write(b2[:])
					buffer_pos2 = 0
					x = 0
				}
				y := x + 8
				z := x + 16

				binary.LittleEndian.PutUint64(b2[x:y], uint64(coOffset))
				binary.LittleEndian.PutUint64(b2[y:z], uint64(coLength))

				coOffset = offset
				coLength = length
				coCount += 1
				buffer_pos2 += 1
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

		if buffer_pos1 != 0 {
			ofh.Write(b1[:buffer_pos1*16])
		}

		x := (buffer_pos2 * 16)
		if x == 16777216 {
			cofh.Write(b2[:])
			buffer_pos2 = 0
			x = 0
		}

		if oSize != 0 {
			binary.LittleEndian.PutUint64(b2[x:x+8], uint64(coOffset))
			binary.LittleEndian.PutUint64(b2[x+8:x+16], uint64(coLength))
			coCount += 1
			buffer_pos2 += 1
			cofh.Write(b2[:buffer_pos2*16])
		}

		err = os.Rename(coTmpFilePath, cofile)
		if err != nil {
			return
		}
		err = os.Rename(oTmpFilePath, ofile)
	} else {
		err = errors.New("Subset node does not currently support the format of your parent index: " + iformat)
		return
	}
	return
}

func (s *subset) Close() (err error) {
	s.f.Close()
	return
}

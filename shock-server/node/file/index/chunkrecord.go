package index

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/node/file/format/multi"
	"github.com/MG-RAST/Shock/shock-server/node/file/format/seq"
	"io"
	"math/rand"
	"os"
)

type chunkRecord struct {
	f     *os.File
	r     seq.Reader
	t     string
	snf   string
	snrp  string
	Index *Idx
	size  int64
}

// We pass our index instantiation function the data file that we will be
// creating an index to, the node type, and the format of the subset node
// index (if the node is of type subset, otherwise snFomrat == "")
func NewChunkRecordIndexer(f *os.File, nType string, snFormat string, snRecordIndexPath string) Indexer {
	fi, _ := f.Stat()
	return &chunkRecord{
		f:     f,
		r:     multi.NewReader(f),
		t:     nType,
		snf:   snFormat,
		snrp:  snRecordIndexPath,
		Index: New(),
		size:  fi.Size(),
	}
}

func (i *chunkRecord) Create(file string) (count int64, format string, err error) {
	if i.t != "subset" {
		tmpFilePath := fmt.Sprintf("%s/temp/%d%d.idx", conf.PATH_DATA, rand.Int(), rand.Int())

		var f *os.File
		f, err = os.Create(tmpFilePath)
		if err != nil {
			return
		}
		defer f.Close()

		format = "array"
		eof := false
		curr := int64(0)
		count = 0
		buffer_pos := 0 // used to track the location in our byte array

		// Writing index file in 16MB chunks
		var b [16777216]byte
		for {
			n, er := i.r.SeekChunk(curr, true)
			if er != nil {
				if er != io.EOF {
					err = er
					return
				}
				eof = true
			}

			// Calculating position in byte array
			x := (buffer_pos * 16)
			if x == 16777216 {
				f.Write(b[:])
				buffer_pos = 0
				x = 0
			}
			y := x + 8
			z := x + 16

			binary.LittleEndian.PutUint64(b[x:y], uint64(curr))
			if eof {
				binary.LittleEndian.PutUint64(b[y:z], uint64(i.size-curr))
			} else {
				binary.LittleEndian.PutUint64(b[y:z], uint64(n))
			}
			curr += int64(n)
			count += 1
			buffer_pos += 1

			if eof {
				break
			}
		}
		if buffer_pos != 0 {
			f.Write(b[:buffer_pos*16])
		}
		err = os.Rename(tmpFilePath, file)

		return
	} else {
		// Handling different types of subset nodes.
		if i.snf == "matrix" {
			err = errors.New("Shock does not currently support the creation of chunkrecord indices for subset nodes derived from a matrix formatted index.")
			return
		} else {
			// We need to open a file handle (rfh) for the subset node's record index file.
			var rfh *os.File
			rfh, err = os.Open(i.snrp)
			if err != nil {
				return
			}
			defer rfh.Close()

			// For matrix formatted indexes, we are basically building a list of lists.  The parent index file will give the offset and length
			// for a list of indexes in the record index file.  In this way we can build a chunkrecord index that will point to a list of record indices.
			parentTmpFilePath := fmt.Sprintf("%s/temp/%d%d.idx", conf.PATH_DATA, rand.Int(), rand.Int())

			var pfh *os.File
			pfh, err = os.Create(parentTmpFilePath)
			if err != nil {
				return
			}
			defer pfh.Close()

			format = "matrix"

			// This variable stores the counter for the record index that we are reading.
			riCount := int64(0)

			// These variables store the number of chunkrecord indices and their current offset and length positions in the child index.
			count = 0
			parentOffset := int64(0)
			parentLength := int64(0)

			// This variable stores the current length of the chunkrecord index.
			chunkRecordLength := int64(0)

			parentBufferPos := int64(0) // used to track the location in our parent byte array

			// Writing parent index file in 16MB chunks.
			var pbuf [16777216]byte

			for {
				// Read first offset and length from node's record index.
				var ibuf [16]byte
				_, er := rfh.ReadAt(ibuf[0:16], int64((riCount)*16))
				if er != nil {
					if er != io.EOF {
						err = errors.New(fmt.Sprintf("Could not read record index file for part: %d", riCount))
						return
					}
					break
				}

				// Ignoring first 8 bytes because that's the offset of the record index and is not needed for creating the chunkrecord index.
				riLength := int64(binary.LittleEndian.Uint64(ibuf[8:16]))
				riCount += 1

				if chunkRecordLength == 0 {
					parentOffset = (riCount - 1) * 16
					parentLength = 16
				} else {
					parentLength += 16
				}

				// If length of current chunkrecord plus the current record is over 1MB, add information to index buffer, then reset variables.
				if chunkRecordLength+riLength >= 1048576 {

					// If chunkrecord is empty without current record, include current record in index, otherwise don't.
					if chunkRecordLength == 0 {
						// write current offset and length to index buffer
						x := (parentBufferPos * 16)
						binary.LittleEndian.PutUint64(pbuf[x:x+8], uint64(parentOffset))
						binary.LittleEndian.PutUint64(pbuf[x+8:x+16], uint64(parentLength))

						// If index buffer is full, print it to disk and reset the buffer position, otherwise increase buffer position.
						if x+16 == 16777216 {
							pfh.Write(pbuf[:])
							parentBufferPos = 0
						} else {
							parentBufferPos += 1
						}
						chunkRecordLength = 0
						parentLength = 0
					} else {
						// write current offset and length to index buffer
						x := (parentBufferPos * 16)
						binary.LittleEndian.PutUint64(pbuf[x:x+8], uint64(parentOffset))
						binary.LittleEndian.PutUint64(pbuf[x+8:x+16], uint64(parentLength-16))

						// If index buffer is full, print it to disk and reset the buffer position, otherwise increase buffer position.
						if x+16 == 16777216 {
							pfh.Write(pbuf[:])
							parentBufferPos = 0
						} else {
							parentBufferPos += 1
						}
						chunkRecordLength = riLength
						parentOffset = (riCount - 1) * 16
						parentLength = 16
					}

					count += 1
				} else {
					chunkRecordLength += riLength
				}
			}

			// If chunkRecordLength is not empty, need to record last chunk.
			if chunkRecordLength != 0 {
				x := (parentBufferPos * 16)
				binary.LittleEndian.PutUint64(pbuf[x:x+8], uint64(parentOffset))
				binary.LittleEndian.PutUint64(pbuf[x+8:x+16], uint64(parentLength))
				parentBufferPos += 1
				count += 1
			}

			// If index buffer is not empty, print it to disk.
			if parentBufferPos != 0 {
				pfh.Write(pbuf[:parentBufferPos*16])
			}

			// Move temporary index file to the desired path.
			err = os.Rename(parentTmpFilePath, file)

			return
		}
	}
}

func (i *chunkRecord) Close() (err error) {
	i.f.Close()
	return
}

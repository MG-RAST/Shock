package store

import (
	//"fmt"
	"io"
	"os"
)

type SectionReader interface {
	io.Reader
	io.ReaderAt
}

type ReaderAt interface {
	io.Reader
	io.ReaderAt
	Stat() (os.FileInfo, error)
}

type multiReaderAt struct {
	readers    []ReaderAt
	boundaries []multifd
	size       int64
}

type multifd struct {
	start int64
	end   int64
	size  int64
}

func (mr *multiReaderAt) Read(p []byte) (n int, err error) {
	for len(mr.readers) > 0 {
		n, err = mr.readers[0].Read(p)
		if n > 0 || err != io.EOF {
			if err == io.EOF {
				// Don't return EOF yet. There may be more bytes
				// in the remaining readers.
				err = nil
			}
			return
		}
		mr.readers = mr.readers[1:]
	}
	return 0, io.EOF
}

func (mr *multiReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	startF, endF := 0, 0
	startPos, endPos, length := int64(0), int64(0), int64(len(p))

	if off > mr.size {
		return 0, io.EOF
	}

	// find start
	for i, fd := range mr.boundaries {
		if off >= fd.start && off <= fd.end {
			startF = i
			startPos = off - fd.start
			break
		}
	}

	// find end    
	if off+length > mr.size {
		endF = len(mr.readers) - 1
		endPos = mr.size - mr.boundaries[endF].start
	} else {
		for i, fd := range mr.boundaries {
			if off+length >= fd.start && off+length <= fd.end {
				endF = i
				endPos = off + length - fd.start
				break
			}
		}
	}

	if startF == endF {
		// read startpos till endpos
		// println("--> readat: startpos till endpos")
		// fmt.Printf("file: %d, offset: %d, length: %d\n", startF, startPos, endPos-startPos)
		return mr.readers[startF].ReadAt(p[0:length], startPos)
	} else {
		buffPos := 0
		for i := startF; i <= endF; i++ {
			if i == startF {
				// read startpos till end of file
				// println("--> readat: startpos till end of file")
				// fmt.Printf("file: %d, offset: %d, length: %d, buffPos: %d\n", i, startPos, mr.boundaries[i].size-startPos, buffPos)
				if rn, err := mr.readers[i].ReadAt(p[buffPos:buffPos+int(mr.boundaries[i].size-startPos)], startPos); err != nil && err != io.EOF {
					return 0, err
				} else {
					buffPos = buffPos + int(mr.boundaries[i].size-startPos)
					n = n + rn
				}
			} else if i == endF {
				// read start of file till endpos
				// println("--> readat: start of file till endpos")
				// fmt.Printf("file: %d, offset: %d, length: %d, buffPos: %d\n", i, 0, endPos, buffPos)
				if rn, err := mr.readers[i].ReadAt(p[buffPos:buffPos+int(endPos)], 0); err != nil && err != io.EOF {
					println("--> error here: ", err.Error())
					return 0, err
				} else {
					buffPos = buffPos + int(endPos)
					n = n + rn
				}
			} else {
				// read entire file
				// println("--> readat: entire file")
				// fmt.Printf("file: %d, offset: %d, length: %d, buffPos: %d\n", i, 0, mr.boundaries[i].size, buffPos)
				if rn, err := mr.readers[i].ReadAt(p[buffPos:buffPos+int(mr.boundaries[i].size)], 0); err != nil && err != io.EOF {
					return 0, err
				} else {
					buffPos = buffPos + int(mr.boundaries[i].size)
					n = n + rn
				}
			}
		}
	}
	if n < int(length) {
		return n, io.EOF
	}
	return
}

// do not use
func (mr *multiReaderAt) Stat() (fi os.FileInfo, err error) {
	return
}

// MultiReader returns a Reader that's the logical concatenation of
// the provided input readers.  They're read sequentially.  Once all
// inputs are drained, Read will return EOF.
func MultiReaderAt(readers ...ReaderAt) ReaderAt {
	mr := &multiReaderAt{readers: readers}
	b := []multifd{}
	start := int64(0)
	for _, r := range mr.readers {
		fi, _ := r.Stat()
		b = append(b, multifd{start: start, end: start + fi.Size(), size: fi.Size()})
		start = start + fi.Size()
	}
	mr.boundaries = b
	mr.size = b[len(b)-1].end
	return mr
}

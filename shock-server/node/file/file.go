// Package contains Node File struct and MultiReaderAt implementation
package file

import (
	"io"
	"os"
	"time"

	"github.com/MG-RAST/Shock/shock-server/node/locker"
)

// File is the Node file structure. Contains the json/bson marshalling controls.
type File struct {
	Name         string            `bson:"name" json:"name"`
	Size         int64             `bson:"size" json:"size"`
	Checksum     map[string]string `bson:"checksum" json:"checksum"`
	Format       string            `bson:"format" json:"format"`
	Path         string            `bson:"path" json:"-"`
	Virtual      bool              `bson:"virtual" json:"virtual"`
	VirtualParts []string          `bson:"virtual_parts" json:"virtual_parts"`
	CreatedOn    time.Time         `bson:"created_on" json:"created_on"`
	Locked       *locker.LockInfo  `bson:"-" json:"locked"`
}

type FormFiles map[string]FormFile

type FormFile struct {
	Name     string
	Path     string
	Checksum map[string]string
}

func (formfile *FormFile) Remove() {
	if _, err := os.Stat(formfile.Path); err == nil {
		os.Remove(formfile.Path)
	}
	return
}

func RemoveAllFormFiles(formfiles FormFiles) {
	for _, formfile := range formfiles {
		formfile.Remove()
	}
	return
}

// FileInfo for streaming file content
type FileInfo struct {
	R        []SectionReader
	E        error
	ESection int
	Body     io.ReadCloser
	Name     string
	Size     int64
	ModTime  time.Time
	Checksum string
}

// SectionReader interface required for MultiReaderAt
type SectionReader interface {
	io.Reader
	io.ReaderAt
}

// ReaderAt interface that is compatiable with os.File types.
type ReaderAt interface {
	SectionReader
	Stat() (os.FileInfo, error)
	Close() error
}

// multifd contains file boundary information
type multifd struct {
	start int64
	end   int64
	size  int64
}

// multiReaderAt is private struct for the multi-file ReaderAt
// that provides the ablity to use indexes with vitrual files.
type multiReaderAt struct {
	readers    []ReaderAt
	boundaries []multifd
	size       int64
}

// MultiReaderAt returns a ReaderAt that's the logical concatenation of
// the provided input readers. BUG / KNOW-ISSUE: all file handles are opened
// initially. May not be suitiable for large numbers of files.
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

// Read same as io.MultiReader
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

// ReadAt is the magic sauce. Heavily commented to include all logic.
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

// Required for the ReaderAt interface but non-implemented
func (mr *multiReaderAt) Stat() (fi os.FileInfo, err error) {
	return
}

// Required for the ReaderAt interface but non-implemented
func (mr *multiReaderAt) Close() (err error) {
	return
}

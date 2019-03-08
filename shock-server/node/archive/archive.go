package archive

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node/file"
)

var validUncompress = []string{"gzip", "bzip2"}
var validCompress = []string{"gzip", "zip"}
var validToArchive = []string{"zip", "tar"}
var validArchive = []string{"zip", "tar", "tar.gz", "tar.bz2"}
var ArchiveList = strings.Join(validArchive, ", ")

func IsValidToArchive(a string) bool {
	for _, b := range validToArchive {
		if b == a {
			return true
		}
	}
	return false
}

func IsValidArchive(a string) bool {
	for _, b := range validArchive {
		if b == a {
			return true
		}
	}
	return false
}

func IsValidUncompress(a string) bool {
	for _, b := range validUncompress {
		if b == a {
			return true
		}
	}
	return false
}

func IsValidCompress(a string) bool {
	for _, b := range validCompress {
		if b == a {
			return true
		}
	}
	return false
}

func FilesFromArchive(format string, filePath string) (fileList []file.FormFile, unpackDir string, err error) {
	// set unpack dir
	unpackDir = fmt.Sprintf("%s/temp/%d%d", conf.PATH_DATA, rand.Int(), rand.Int())
	if merr := os.Mkdir(unpackDir, 0777); merr != nil {
		logger.Error("err:@node_unpack: " + err.Error())
	}

	// magic to unpack archive
	if format == "zip" {
		fileList, err = unZip(filePath, unpackDir)
	} else if format == "tar" {
		fileList, err = unTar(filePath, unpackDir, "")
	} else if format == "tar.gz" {
		fileList, err = unTar(filePath, unpackDir, "gzip")
	} else if format == "tar.bz2" {
		fileList, err = unTar(filePath, unpackDir, "bzip2")
	} else {
		return nil, unpackDir, errors.New("invalid archive format. must be one of: " + ArchiveList)
	}
	return
}

func unTar(filePath string, unpackDir string, compression string) (fileList []file.FormFile, err error) {
	// get file handle
	openFile, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer openFile.Close()

	// filter through compression if use
	ucReader, ucErr := UncompressReader(compression, openFile)
	if ucErr != nil {
		return nil, ucErr
	}

	// extract tarball
	tarBallReader := tar.NewReader(ucReader)
	for {
		header, nerr := tarBallReader.Next()
		if nerr != nil {
			if nerr == io.EOF {
				break
			}
			return nil, nerr
		}

		// set names
		path := filepath.Join(unpackDir, header.Name)
		baseName := filepath.Base(header.Name)
		// skip hidden
		if strings.HasPrefix(baseName, ".") {
			continue
		}

		// only handle real files and dirs, ignore links
		switch header.Typeflag {
		// handle regualar file
		case tar.TypeReg:
			// open output file
			var writer *os.File
			if writer, err = os.Create(path); err != nil {
				return
			}
			// get md5
			md5h := md5.New()
			dst := io.MultiWriter(writer, md5h)
			// write it
			_, err = io.Copy(dst, tarBallReader)
			writer.Close()
			if err != nil {
				return
			}
			// add to filelist
			ffile := file.FormFile{Name: baseName, Path: path, Checksum: make(map[string]string)}
			ffile.Checksum["md5"] = fmt.Sprintf("%x", md5h.Sum(nil))
			fileList = append(fileList, ffile)
		case tar.TypeDir:
			// handle directory
			if merr := os.MkdirAll(path, 0777); merr != nil {
				logger.Error("err:@node_untar: " + err.Error())
			}
		default:
		}
	}
	return
}

func unZip(filePath string, unpackDir string) (fileList []file.FormFile, err error) {
	// open file with unzip
	zipReader, err := zip.OpenReader(filePath)
	if err != nil {
		return
	}

	// extract archive
	for _, zf := range zipReader.File {
		// set names
		path := filepath.Join(unpackDir, zf.Name)
		baseName := filepath.Base(zf.Name)
		// skip hidden
		if strings.HasPrefix(baseName, ".") {
			continue
		}

		if zf.FileInfo().IsDir() {
			// handle directory
			if merr := os.MkdirAll(path, 0777); merr != nil {
				logger.Error("err:@node_untar: " + err.Error())
			}
		} else {
			// open output file
			writer, werr := os.Create(path)
			if werr != nil {
				return nil, werr
			}
			// open input stream
			zfh, zerr := zf.Open()
			if zerr != nil {
				writer.Close()
				return nil, zerr
			}
			// get md5
			md5h := md5.New()
			dst := io.MultiWriter(writer, md5h)
			// write it
			_, err = io.Copy(dst, zfh)
			writer.Close()
			zfh.Close()
			if err != nil {
				return
			}
			// add to filelist
			ffile := file.FormFile{Name: baseName, Path: path, Checksum: make(map[string]string)}
			ffile.Checksum["md5"] = fmt.Sprintf("%x", md5h.Sum(nil))
			fileList = append(fileList, ffile)
		}
	}
	return
}

func ArchiveReader(format string, files []*file.FileInfo) (outReader io.ReadCloser) {
	pReader, pWriter := io.Pipe()
	if format == "tar" {
		tWriter := tar.NewWriter(pWriter)
		go func() {
			fileNames := map[string]int{}
			for _, f := range files {
				fileName := f.Name
				if num, ok := fileNames[f.Name]; ok {
					fileName = fmt.Sprintf("%s.%d", fileName, num+1)
					fileNames[f.Name] = num + 1
				} else {
					fileNames[f.Name] = 1
				}
				fHdr := &tar.Header{Name: fileName, Mode: 0660, ModTime: f.ModTime, Size: f.Size}
				err := tWriter.WriteHeader(fHdr)
				if err != nil {
					logger.Error("(ArchiveReader) tWriter.WriteHeader returned: " + err.Error())
					tWriter.Close()
					pWriter.Close()
					return
				}
				io.Copy(tWriter, f.Body)
				if f.Checksum != "" {
					cHdr := &tar.Header{Name: fileName + ".md5", Mode: 0660, ModTime: f.ModTime, Size: int64(len(f.Checksum))}
					tWriter.WriteHeader(cHdr)
					cBuf := bytes.NewBufferString(f.Checksum)
					if cBuf != nil {
						io.Copy(tWriter, cBuf)
					}
				}
			}
			tWriter.Close()
			pWriter.Close()
		}()
	} else if format == "zip" {
		zWriter := zip.NewWriter(pWriter)
		go func() {
			fileNames := map[string]int{}
			for _, f := range files {
				fileName := f.Name
				if num, ok := fileNames[f.Name]; ok {
					fileName = fmt.Sprintf("%s.%d", fileName, num+1)
					fileNames[f.Name] = num + 1
				} else {
					fileNames[f.Name] = 1
				}
				zHdr := &zip.FileHeader{Name: fileName, UncompressedSize64: uint64(f.Size)}
				zHdr.SetModTime(f.ModTime)
				zFile, zferr := zWriter.CreateHeader(zHdr)
				if zferr != nil {
					logger.Error("(ArchiveReader) zWriter.CreateHeader returned: " + zferr.Error())
					zWriter.Close()
					pWriter.Close()
					return
				}
				io.Copy(zFile, f.Body)
				if f.Checksum != "" {
					cHdr := &zip.FileHeader{Name: fileName + ".md5", UncompressedSize64: uint64(len(f.Checksum))}
					cHdr.SetModTime(f.ModTime)
					zSum, zcerr := zWriter.CreateHeader(cHdr)
					cBuf := bytes.NewBufferString(f.Checksum)
					if (zcerr != nil) && (cBuf != nil) {
						io.Copy(zSum, cBuf)
					}
				}

			}
			zWriter.Close()
			pWriter.Close()
		}()
	} else {
		// no valid archive, pipe each inReader into one stream
		go func() {
			for _, f := range files {
				io.Copy(pWriter, f.Body)
			}
			pWriter.Close()
		}()
	}
	return pReader
}

func CompressReader(format string, filename string, inReader io.ReadCloser) (outReader io.ReadCloser) {
	if IsValidCompress(format) {
		pReader, pWriter := io.Pipe()
		if format == "gzip" {
			gWriter := gzip.NewWriter(pWriter)
			go func() {
				gWriter.Header.Name = filename
				gWriter.Header.ModTime = time.Now()
				io.Copy(gWriter, inReader)
				gWriter.Close()
				pWriter.Close()
			}()
		} else if format == "zip" {
			zWriter := zip.NewWriter(pWriter)
			go func() {
				zHdr := &zip.FileHeader{Name: filename}
				zHdr.SetModTime(time.Now())
				zFile, _ := zWriter.CreateHeader(zHdr)
				io.Copy(zFile, inReader)
				zWriter.Close()
				pWriter.Close()
			}()
		}
		return pReader
	}
	// just return input reader if no valid compression
	return inReader
}

func UncompressReader(format string, inReader io.Reader) (outReader io.Reader, err error) {
	if IsValidUncompress(format) {
		if format == "gzip" {
			gReader, gerr := gzip.NewReader(inReader)
			if gerr != nil {
				return nil, gerr
			}
			return gReader, nil
		} else if format == "bzip2" {
			bReader := bzip2.NewReader(inReader)
			return bReader, nil
		}
	}
	return inReader, nil
}

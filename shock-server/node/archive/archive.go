package archive

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

var validUncompress = []string{"gzip", "bzip2"}
var validCompress = []string{"gzip", "zip"}
var validArchive = []string{"zip", "tar", "tar.gz", "tar.bz2"}
var ArchiveList = strings.Join(validArchive, ", ")

type FormFile struct {
	Name     string
	Path     string
	Checksum map[string]string
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

func FilesFromArchive(format string, filePath string) (fileList []FormFile, unpackDir string, err error) {
	// set unpack dir
	unpackDir = fmt.Sprintf("%s/temp/%d%d", conf.PATH_DATA, rand.Int(), rand.Int())
	if err = os.Mkdir(unpackDir, 0777); err != nil {
		return
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

func unTar(filePath string, unpackDir string, compression string) (fileList []FormFile, err error) {
	// get file handle
	openFile, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer openFile.Close()

	// tarball handle
	var tarBallReader *tar.Reader

	// filter through compression if use
	if compression == "gzip" {
		gReader, gerr := gzip.NewReader(openFile)
		if gerr != nil {
			return nil, gerr
		}
		defer gReader.Close()
		tarBallReader = tar.NewReader(gReader)
	} else if compression == "bzip2" {
		bReader := bzip2.NewReader(openFile)
		tarBallReader = tar.NewReader(bReader)
	} else {
		tarBallReader = tar.NewReader(openFile)
	}

	// extract tarball
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
		// handle directory
		case tar.TypeDir:
			if err = os.MkdirAll(path, 0777); err != nil {
				return
			}
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
			ffile := FormFile{Name: baseName, Path: path, Checksum: make(map[string]string)}
			ffile.Checksum["md5"] = fmt.Sprintf("%x", md5h.Sum(nil))
			fileList = append(fileList, ffile)
		default:
		}
	}
	return
}

func unZip(filePath string, unpackDir string) (fileList []FormFile, err error) {
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
			if err = os.MkdirAll(path, 0777); err != nil {
				return
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
			ffile := FormFile{Name: baseName, Path: path, Checksum: make(map[string]string)}
			ffile.Checksum["md5"] = fmt.Sprintf("%x", md5h.Sum(nil))
			fileList = append(fileList, ffile)
		}
	}
	return
}

func CompressReader(format string, filename string, inReader io.ReadCloser) (outReader io.ReadCloser) {
	if IsValidCompress(format) {
		pReader, pWriter := io.Pipe()
		if format == "gzip" {
			gWriter := gzip.NewWriter(pWriter)
			gWriter.Header.Name = filename
			go func() {
				io.Copy(gWriter, inReader)
				gWriter.Close()
				pWriter.Close()
			}()
		} else if format == "zip" {
			zWriter := zip.NewWriter(pWriter)
			zFile, _ := zWriter.Create(filename)
			go func() {
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
			gWriter, gerr := gzip.NewReader(inReader)
			if gerr != nil {
				return nil, gerr
			}
			return gWriter, nil
		} else if format == "bzip2" {
			bWriter := bzip2.NewReader(inReader)
			return bWriter, nil
		}
	}
	return inReader, nil
}

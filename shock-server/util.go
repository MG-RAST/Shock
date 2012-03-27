package main

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/MG-RAST/Shock/conf"
	ds "github.com/MG-RAST/Shock/datastore"
	"github.com/MG-RAST/Shock/user"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type SectionReaderCloser struct {
	f  *os.File
	sr *io.SectionReader
}

// io.SectionReader doesn't implement close. Why? No one knows.
func NewSectionReaderCloser(f *os.File, off int64, n int64) *SectionReaderCloser {
	return &SectionReaderCloser{
		f:  f,
		sr: io.NewSectionReader(f, off, n),
	}
}

func (s *SectionReaderCloser) Read(p []byte) (n int, err error) {
	return s.sr.Read(p)
}

func (s *SectionReaderCloser) Seek(offset int64, whence int) (ret int64, err error) {
	return s.sr.Seek(offset, whence)
}

func (s *SectionReaderCloser) ReadAt(p []byte, off int64) (n int, err error) {
	return s.sr.ReadAt(p, off)
}

func (s *SectionReaderCloser) Close() error {
	return s.f.Close()
}

type streamer struct {
	rs          []io.ReadCloser
	ws          http.ResponseWriter
	contentType string
	filename    string
	size        int64
	filter      func(io.ReadCloser) io.ReadCloser
}

func (s *streamer) stream() (err error) {
	s.ws.Header().Set("Content-Type", s.contentType)
	s.ws.Header().Set("Content-Disposition", fmt.Sprintf(":attachment;filename=%s", s.filename))
	if s.size > 0 && s.filter == nil {
		s.ws.Header().Set("Content-Length", fmt.Sprint(s.size))
	}
	for _, sr := range s.rs {
		var rs io.ReadCloser
		if s.filter != nil {
			rs = s.filter(sr)
		} else {
			rs = sr
		}
		_, err = io.Copy(s.ws, rs)
		if err != nil {
			return
		}
	}
	return
}

// helper function for create & update
func ParseMultipartForm(r *http.Request) (params map[string]string, files ds.FormFiles, err error) {
	params = make(map[string]string)
	files = make(ds.FormFiles)
	md5h := md5.New()
	sha1h := sha1.New()
	reader, err := r.MultipartReader()
	if err != nil {
		return
	}
	for {
		part, err := reader.NextPart()
		if err != nil {
			break
		}
		if part.FileName() == "" {
			buffer := make([]byte, 32*1024)
			n, err := part.Read(buffer)
			if n == 0 || err != nil {
				break
			}
			params[part.FormName()] = fmt.Sprintf("%s", buffer[0:n])
		} else {
			tmpPath := fmt.Sprintf("%s/temp/%d%d", *conf.DATAROOT, rand.Int(), rand.Int())
			files[part.FormName()] = ds.FormFile{Name: part.FileName(), Path: tmpPath, Checksum: make(map[string]string)}
			tmpFile, err := os.Create(tmpPath)
			if err != nil {
				break
			}
			buffer := make([]byte, 32*1024)
			for {
				n, err := part.Read(buffer)
				if n == 0 || err != nil {
					break
				}
				tmpFile.Write(buffer[0:n])
				md5h.Write(buffer[0:n])
				sha1h.Write(buffer[0:n])
			}

			var md5s, sha1s []byte
			md5s = md5h.Sum(md5s)
			sha1s = sha1h.Sum(sha1s)
			files[part.FormName()].Checksum["md5"] = fmt.Sprintf("%x", md5s)
			files[part.FormName()].Checksum["sha1"] = fmt.Sprintf("%x", sha1s)

			tmpFile.Close()
			md5h.Reset()
			sha1h.Reset()
		}
	}
	if err != nil {
		return
	}
	return
}

func LogRequest(req *http.Request) {
	host, _, _ := net.SplitHostPort(req.RemoteAddr)
	// failed attempt to get the host in ipv4
	//addrs, _ := net.LookupIP(host)	
	//fmt.Println(addrs)
	prefix := fmt.Sprintf("%s [%s]", host, time.Now().Format(time.RFC1123))
	suffix := ""
	if _, auth := req.Header["Authorization"]; auth {
		suffix = "AUTH"
	}
	url := ""
	if req.URL.RawQuery != "" {
		url = fmt.Sprintf("%s %s?%s", req.Method, req.URL.Path, req.URL.RawQuery)
	} else {
		url = fmt.Sprintf("%s %s", req.Method, req.URL.Path)
	}
	fmt.Printf("%s %q %s\n", prefix, url, suffix)
}

func AuthenticateRequest(req *http.Request) (u *user.User, err error) {
	if _, ok := req.Header["Authorization"]; !ok {
		err = errors.New("No Authorization")
		return
	}
	header := req.Header.Get("Authorization")
	tmpAuthArray := strings.Split(header, " ")

	authValues, err := base64.URLEncoding.DecodeString(tmpAuthArray[1])
	if err != nil {
		err = errors.New("Failed to decode encoded auth settings in http request.")
		return
	}

	authValuesArray := strings.Split(string(authValues), ":")
	name := authValuesArray[0]
	passwd := authValuesArray[1]
	u, err = user.Authenticate(name, passwd)
	return
}

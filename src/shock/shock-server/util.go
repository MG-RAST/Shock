package main

import (
	"net/http"
	"fmt"
	"math/rand"
	"os"
	"io"
	"crypto/md5"
	"crypto/sha1"
	ds "shock/datastore"
	conf "shock/conf"
	"goweb"
)

type streamer struct {
	rs io.Reader
	ws http.ResponseWriter
	contentType string
	filename string
	size int64
}

func (s *streamer) stream() (err error){
	s.ws.Header().Set("Content-Type", s.contentType)
	s.ws.Header().Set("Content-Disposition", fmt.Sprintf(":attachment;filename=%s", s.filename))
	s.ws.Header().Set("Content-Length", fmt.Sprint(s.size))
	_, err = io.Copy(s.ws, s.rs)
	return	
}

// helper function for create & update
func ParseMultipartForm(r *http.Request) (params map[string]string, files ds.FormFiles, err error){
	params = make(map[string]string)
	files = make(ds.FormFiles)
	md5h := md5.New()
	sha1h := sha1.New()	
	reader, err := r.MultipartReader(); if err != nil { return }
	for {
		part, err := reader.NextPart(); if err != nil { break }
		if part.FileName() == "" {
			buffer := make([]byte, 32*1024)
			n, err := part.Read(buffer)
			if n == 0 || err != nil { break }
			params[part.FormName()] = fmt.Sprintf("%s", buffer[0:n])
		} else {
			tmpPath := fmt.Sprintf("%s/temp/%d%d", *conf.DATAROOT, rand.Int(), rand.Int())
			files[part.FormName()] = ds.FormFile{Name: part.FileName(), Path: tmpPath, Checksum: make(map[string]string)}
			tmpFile, err := os.Create(tmpPath); if err != nil { break }
			buffer := make([]byte, 32*1024)
			for {
				n, err := part.Read(buffer)
				if n == 0 || err != nil { break }
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
	if err != nil { return }
	return
}

func logReq(cx *goweb.Context) () {
	if cx.Request.URL.RawQuery != "" {
		fmt.Printf("%s: %s?%s\n", cx.Request.Method, cx.Request.URL.Path, cx.Request.URL.RawQuery)
	} else {
		fmt.Printf("%s: %s\n", cx.Request.Method, cx.Request.URL.Path)
	}
}

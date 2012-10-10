package main

import (
	//"compress/gzip"
	"crypto/md5"
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/MG-RAST/Shock/conf"
	e "github.com/MG-RAST/Shock/errors"
	"github.com/MG-RAST/Shock/store"
	"github.com/MG-RAST/Shock/store/filter"
	"github.com/MG-RAST/Shock/store/user"
	"github.com/MG-RAST/Shock/store/user/auth"
	"github.com/jaredwilkening/goweb"
	"io"
	"math/rand"
	//"mime/multipart"
	"net"
	"net/http"
	"os"
	"strings"
)

var (
	logo = "\n" +
		" +-------------+  +----+    +----+  +--------------+  +--------------+  +----+      +----+\n" +
		" |             |  |    |    |    |  |              |  |              |  |    |      |    |\n" +
		" |    +--------+  |    |    |    |  |    +----+    |  |    +---------+  |    |      |    |\n" +
		" |    |           |    +----+    |  |    |    |    |  |    |            |    |     |    |\n" +
		" |    +--------+  |              |  |    |    |    |  |    |            |    |    |    |\n" +
		" |             |  |    +----+    |  |    |    |    |  |    |            |    |   |    |\n" +
		" +--------+    |  |    |    |    |  |    |    |    |  |    |            |    +---+    +-+\n" +
		"          |    |  |    |    |    |  |    |    |    |  |    |            |               |\n" +
		" +--------+    |  |    |    |    |  |    +----+    |  |    +---------+  |    +-----+    |\n" +
		" |             |  |    |    |    |  |              |  |              |  |    |     |    |\n" +
		" +-------------+  +----+    +----+  +--------------+  +--------------+  +----+     +----+\n"
)

func printLogo() {
	fmt.Println(logo)
	return
}

type Query struct {
	list map[string][]string
}

func (q *Query) Has(key string) bool {
	if _, has := q.list[key]; has {
		return true
	}
	return false
}

func (q *Query) Value(key string) string {
	return q.list[key][0]
}

func (q *Query) List(key string) []string {
	return q.list[key]
}

func (q *Query) All() map[string][]string {
	return q.list
}

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
	filter      filter.FilterFunc
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
func ParseMultipartForm(r *http.Request) (params map[string]string, files store.FormFiles, err error) {
	params = make(map[string]string)
	files = make(store.FormFiles)
	md5h := md5.New()
	sha1h := sha1.New()
	reader, err := r.MultipartReader()
	if err != nil {
		return
	}
	for {
		if part, err := reader.NextPart(); err == nil {
			if part.FileName() == "" {
				buffer := make([]byte, 32*1024)
				n, err := part.Read(buffer)
				if n == 0 || err != nil {
					break
				}
				params[part.FormName()] = fmt.Sprintf("%s", buffer[0:n])
			} else {
				tmpPath := fmt.Sprintf("%s/temp/%d%d", conf.DATA_PATH, rand.Int(), rand.Int())
				/*
					if fname[len(fname)-3:] == ".gz" && params["decompress"] == "true" {
						fname = fname[:len(fname)-3]
						reader, err = gzip.NewReader(&part)
						if err != nil {
							break
						}
					} else {
						reader = &part
					}
				*/
				files[part.FormName()] = store.FormFile{Name: part.FileName(), Path: tmpPath, Checksum: make(map[string]string)}
				if tmpFile, err := os.Create(tmpPath); err == nil {
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

					md5h.Reset()
					sha1h.Reset()
					tmpFile.Close()
				} else {
					return nil, nil, err
				}
			}
		} else if err.Error() != "EOF" {
			fmt.Println("err here")
			return nil, nil, err
		} else {
			break
		}
	}
	return
}

type resource struct {
	R []string `json:"resources"`
	U string   `json:"url"`
	D string   `json:"documentation"`
	C string   `json:"contact"`
	I string   `json:"id"`
	T string   `json:"type"`
}

func ResourceDescription(cx *goweb.Context) {
	LogRequest(cx.Request)
	host := ""
	if strings.Contains(cx.Request.Host, ":") {
		split := strings.Split(cx.Request.Host, ":")
		host = split[0]
	} else {
		host = cx.Request.Host
	}
	r := resource{
		R: []string{"node", "user"},
		U: "http://" + host + ":" + fmt.Sprint(conf.API_PORT) + "/",
		D: "http://" + host + ":" + fmt.Sprint(conf.SITE_PORT) + "/",
		C: conf.ADMIN_EMAIL,
		I: "Shock",
		T: "Shock",
	}
	cx.WriteResponse(r, 200)
}

func Site(cx *goweb.Context) {
	LogRequest(cx.Request)
	http.ServeFile(cx.ResponseWriter, cx.Request, conf.SITE_PATH+"/pages/main.html")
}

func RawDir(cx *goweb.Context) {
	LogRequest(cx.Request)
	http.ServeFile(cx.ResponseWriter, cx.Request, fmt.Sprintf("%s%s", conf.DATA_PATH, cx.Request.URL.Path))
}

func AssetsDir(cx *goweb.Context) {
	LogRequest(cx.Request)
	http.ServeFile(cx.ResponseWriter, cx.Request, conf.SITE_PATH+cx.Request.URL.Path)
}

func LogRequest(req *http.Request) {
	host, _, _ := net.SplitHostPort(req.RemoteAddr)
	// failed attempt to get the host in ipv4
	//addrs, _ := net.LookupIP(host)	
	//fmt.Println(addrs)
	suffix := ""
	if _, auth := req.Header["Authorization"]; auth {
		suffix = " AUTH"
	}
	url := ""
	if req.URL.RawQuery != "" {
		url = fmt.Sprintf("%s %s?%s", req.Method, req.URL.Path, req.URL.RawQuery)
	} else {
		url = fmt.Sprintf("%s %s", req.Method, req.URL.Path)
	}
	log.Info("access", host+" \""+url+suffix+"\"")
}

func AuthenticateRequest(req *http.Request) (u *user.User, err error) {
	if _, ok := req.Header["Authorization"]; !ok {
		err = errors.New(e.NoAuth)
		return
	}
	header := req.Header.Get("Authorization")
	u, err = auth.Authenticate(header)
	return
}

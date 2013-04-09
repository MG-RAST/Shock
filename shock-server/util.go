package main

import (
	"crypto/md5"
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/store"
	"github.com/MG-RAST/Shock/shock-server/store/user"
	"github.com/MG-RAST/Shock/shock-server/store/user/auth"
	"github.com/jaredwilkening/goweb"
	"hash"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const logo = `
 +-------------+  +----+    +----+  +--------------+  +--------------+  +----+      +----+
 |             |  |    |    |    |  |              |  |              |  |    |      |    |
 |    +--------+  |    |    |    |  |    +----+    |  |    +---------+  |    |      |    |
 |    |           |    +----+    |  |    |    |    |  |    |            |    |     |    |
 |    +--------+  |              |  |    |    |    |  |    |            |    |    |    |
 |             |  |    +----+    |  |    |    |    |  |    |            |    |   |    |
 +--------+    |  |    |    |    |  |    |    |    |  |    |            |    +---+    +-+
          |    |  |    |    |    |  |    |    |    |  |    |            |               |
 +--------+    |  |    |    |    |  |    +----+    |  |    +---------+  |    +-----+    |
 |             |  |    |    |    |  |              |  |              |  |    |     |    |
 +-------------+  +----+    +----+  +--------------+  +--------------+  +----+     +----+`

const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890"

type checkSumCom struct {
	buf      []byte
	n        int
	checksum string
}

type urlResponse struct {
	Url       string `json:"url"`
	ValidTill string `json:"validtill"`
}

func RandString(l int) (s string) {
	rand.Seed(time.Now().UTC().UnixNano())
	c := make([]byte, l)
	for i := 0; i < l; i++ {
		c[i] = chars[rand.Intn(len(chars))]
	}
	return string(c)
}

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

// helper function for create & update
func ParseMultipartForm(r *http.Request) (params map[string]string, files store.FormFiles, err error) {
	params = make(map[string]string)
	files = make(store.FormFiles)
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
					md5c := make(chan checkSumCom)
					sha1c := make(chan checkSumCom)
					writeChecksum(md5.New, md5c)
					writeChecksum(sha1.New, sha1c)
					for {
						n, err := part.Read(buffer)
						if n == 0 || err != nil {
							md5c <- checkSumCom{n: 0}
							sha1c <- checkSumCom{n: 0}
							break
						}
						md5c <- checkSumCom{buf: buffer[0:n], n: n}
						sha1c <- checkSumCom{buf: buffer[0:n], n: n}
						tmpFile.Write(buffer[0:n])
					}
					md5r := <-md5c
					sha1r := <-sha1c
					files[part.FormName()].Checksum["md5"] = md5r.checksum
					files[part.FormName()].Checksum["sha1"] = sha1r.checksum
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

func writeChecksum(f func() hash.Hash, c chan checkSumCom) {
	go func() {
		h := f()
		for {
			select {
			case b := <-c:
				if b.n == 0 {
					c <- checkSumCom{checksum: fmt.Sprintf("%x", h.Sum(nil))}
					return
				} else {
					h.Write(b.buf[0:b.n])
				}
			}
		}
	}()
}

type resource struct {
	R []string `json:"resources"`
	U string   `json:"url"`
	D string   `json:"documentation"`
	C string   `json:"contact"`
	I string   `json:"id"`
	T string   `json:"type"`
}

func RespondOk(cx *goweb.Context) {
	LogRequest(cx.Request)
	cx.RespondWithOK()
	return
}

func apiUrl(cx *goweb.Context) string {
	if conf.API_URL != "" {
		return conf.API_URL
	}
	return "http://" + cx.Request.Host
}

func siteUrl(cx *goweb.Context) string {
	if conf.SITE_URL != "" {
		return conf.SITE_URL
	} else if strings.Contains(cx.Request.Host, ":") {
		return fmt.Sprintf("http://%s:%d", strings.Split(cx.Request.Host, ":")[0], conf.SITE_PORT)
	}
	return "http://" + cx.Request.Host
}

func ResourceDescription(cx *goweb.Context) {
	LogRequest(cx.Request)
	r := resource{
		R: []string{"node", "user"},
		U: apiUrl(cx) + "/",
		D: siteUrl(cx) + "/",
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

func handleAuthError(err error, cx *goweb.Context) {
	switch err.Error() {
	case e.MongoDocNotFound:
		cx.RespondWithErrorMessage("Invalid username or password", http.StatusBadRequest)
		return
	case e.InvalidAuth:
		cx.RespondWithErrorMessage("Invalid Authorization header", http.StatusBadRequest)
		return
	}
	log.Error("Error at Auth: " + err.Error())
	cx.RespondWithError(http.StatusInternalServerError)
	return
}

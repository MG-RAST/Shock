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
	"os/exec"
	"path/filepath"
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

type streamer struct {
	rs          []store.SectionReader
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
		var rs io.Reader
		if s.filter != nil {
			rs = s.filter(sr)
		} else {
			rs = sr
		}
		_, err := io.Copy(s.ws, rs)
		if err != nil {
			return err
		}
	}
	return
}

func (s *streamer) stream_samtools(filePath string, region string, args ...string) (err error) {
	//involking samtools in command line:
	//samtools view [-c] [-H] [-f INT] ... filname.bam [region]

	argv := []string{}
	argv = append(argv, "view")
	argv = append(argv, args...)
	argv = append(argv, filePath)

	if region != "" {
		argv = append(argv, region)
	}

	LoadBamIndex(filePath)

	cmd := exec.Command("samtools", argv...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	err = cmd.Start()
	if err != nil {
		return err
	}

	go io.Copy(s.ws, stdout)

	err = cmd.Wait()
	if err != nil {
		return err
	}

	UnLoadBamIndex(filePath)

	return
}

//helper function to translate args in URL query to samtools args
//manual: http://samtools.sourceforge.net/samtools.shtml
func ParseSamtoolsArgs(query *Query) (argv []string, err error) {

	var (
		filter_options = map[string]string{
			"head":     "-h",
			"headonly": "-H",
			"count":    "-c",
		}
		valued_options = map[string]string{
			"flag":      "-f",
			"lib":       "-l",
			"mapq":      "-q",
			"readgroup": "-r",
		}
	)

	for src, des := range filter_options {
		if query.Has(src) {
			argv = append(argv, des)
		}
	}

	for src, des := range valued_options {
		if query.Has(src) {
			if val := query.Value(src); val != "" {
				argv = append(argv, des)
				argv = append(argv, val)
			} else {
				return nil, errors.New(fmt.Sprintf("required value not found for query arg: %s ", src))
			}
		}
	}
	return argv, nil
}

func CreateBamIndex(bamFile string) (err error) {
	err = exec.Command("samtools", "index", bamFile).Run()
	if err != nil {
		return err
	}

	baiFile := fmt.Sprintf("%s.bai", bamFile)
	idxPath := fmt.Sprintf("%s/idx/", filepath.Dir(bamFile))

	err = exec.Command("mv", baiFile, idxPath).Run()
	if err != nil {
		return err
	}

	return
}

func LoadBamIndex(bamFile string) (err error) {
	bamFileDir := filepath.Dir(bamFile)
	bamFileName := filepath.Base(bamFile)
	targetBai := fmt.Sprintf("%s/%s.bai", bamFileDir, bamFileName)
	srcBai := fmt.Sprintf("%s/idx/%s.bai", bamFileDir, bamFileName)
	err = exec.Command("ln", "-s", srcBai, targetBai).Run()
	if err != nil {
		return err
	}
	return
}

func UnLoadBamIndex(bamFile string) (err error) {
	bamFileDir := filepath.Dir(bamFile)
	bamFileName := filepath.Base(bamFile)
	targetBai := fmt.Sprintf("%s/%s.bai", bamFileDir, bamFileName)
	err = exec.Command("rm", targetBai).Run()
	if err != nil {
		return err
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

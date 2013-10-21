package request

import (
	"crypto/md5"
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/auth"
	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/jaredwilkening/goweb"
	"hash"
	"math/rand"
	"net"
	"net/http"
	"os"
)

type checkSumCom struct {
	buf      []byte
	n        int
	checksum string
}

func Log(req *http.Request) {
	host, _, _ := net.SplitHostPort(req.RemoteAddr)
	// failed attempt to get the host in ipv4
	//addrs, _ := net.LookupIP(host)
	//fmt.Println(addrs)
	suffix := ""
	if _, auth := req.Header["Authorization"]; auth {
		suffix += " AUTH"
	}

	if l, has := req.Header["Content-Length"]; has {
		suffix += " Content-Length: " + l[0]
	}
	url := ""
	if req.URL.RawQuery != "" {
		url = fmt.Sprintf("%s %s?%s", req.Method, req.URL.Path, req.URL.RawQuery)
	} else {
		url = fmt.Sprintf("%s %s", req.Method, req.URL.Path)
	}
	logger.Info("access", host+" \""+url+suffix+"\"")
}

func Authenticate(req *http.Request) (u *user.User, err error) {
	if _, ok := req.Header["Authorization"]; !ok {
		err = errors.New(e.NoAuth)
		return
	}
	header := req.Header.Get("Authorization")
	u, err = auth.Authenticate(header)
	return
}

func AuthError(err error, cx *goweb.Context) {
	if err.Error() == e.InvalidAuth {
		cx.RespondWithErrorMessage("Invalid authorization header or content", http.StatusBadRequest)
		return
	}
	logger.Error("Error at Auth: " + err.Error())
	cx.RespondWithError(http.StatusInternalServerError)
	return
}

// helper function for create & update
func ParseMultipartForm(r *http.Request) (params map[string]string, files node.FormFiles, err error) {
	params = make(map[string]string)
	files = make(node.FormFiles)
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
				tmpPath := fmt.Sprintf("%s/temp/%d%d", conf.Conf["data-path"], rand.Int(), rand.Int())
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
				files[part.FormName()] = node.FormFile{Name: part.FileName(), Path: tmpPath, Checksum: make(map[string]string)}
				if tmpFile, err := os.Create(tmpPath); err == nil {
					defer tmpFile.Close()
					md5c := make(chan checkSumCom)
					sha1c := make(chan checkSumCom)
					writeChecksum(md5.New, md5c)
					writeChecksum(sha1.New, sha1c)
					for {
						buffer := make([]byte, 32*1024)
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

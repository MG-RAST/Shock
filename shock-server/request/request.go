package request

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/auth"
	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/MG-RAST/Shock/shock-server/util"
	"github.com/MG-RAST/golib/goweb"
	"hash"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
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
	err_msg := "Error at Auth: " + err.Error()
	logger.Error(err_msg)
	cx.RespondWithErrorMessage(err_msg, http.StatusInternalServerError)
	return
}

// helper function to create a node from an http data post
func DataUpload(r *http.Request) (params map[string]string, files node.FormFiles, err error) {
	params = make(map[string]string)
	files = make(node.FormFiles)
	tmpPath := fmt.Sprintf("%s/temp/%d%d", conf.Conf["data-path"], rand.Int(), rand.Int())

	files["upload"] = node.FormFile{Name: "filename", Path: tmpPath, Checksum: make(map[string]string)}
	if tmpFile, err := os.Create(tmpPath); err == nil {
		defer tmpFile.Close()
		md5c := make(chan checkSumCom)
		writeChecksum(md5.New, md5c)
		for {
			buffer := make([]byte, 32*1024)
			n, err := r.Body.Read(buffer)
			if n == 0 || err != nil {
				md5c <- checkSumCom{n: 0}
				break
			}
			md5c <- checkSumCom{buf: buffer[0:n], n: n}
			tmpFile.Write(buffer[0:n])
		}
		md5r := <-md5c
		files["upload"].Checksum["md5"] = md5r.checksum
	} else {
		return nil, nil, err
	}

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

	// arrays to check for valid param and file form names for node creation and updating, and also acl modification
	// Note: indexing and querying do not use this function and thus we don't have to accept those field names.
	validParams := []string{"action", "all", "delete", "format", "ids", "linkage", "operation", "owner", "parts", "path", "read", "source", "tags", "type", "users", "write"}
	validFiles := []string{"attributes", "upload"}

	for {
		if part, err := reader.NextPart(); err == nil {
			// params don't have a FileName() and files must have FormName() of either "upload", "attributes", or an integer
			if part.FileName() == "" {
				if !util.StringInSlice(part.FormName(), validParams) {
					return nil, nil, errors.New("invalid param: " + part.FormName())
				}
				buffer := make([]byte, 32*1024)
				n, err := part.Read(buffer)
				if n == 0 || err != nil {
					break
				}
				params[part.FormName()] = fmt.Sprintf("%s", buffer[0:n])
			} else {
				if _, er := strconv.Atoi(part.FormName()); er != nil && !util.StringInSlice(part.FormName(), validFiles) {
					return nil, nil, errors.New("invalid file param: " + part.FormName())
				}
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
					writeChecksum(md5.New, md5c)
					for {
						buffer := make([]byte, 32*1024)
						n, err := part.Read(buffer)
						if n == 0 || err != nil {
							md5c <- checkSumCom{n: 0}
							break
						}
						md5c <- checkSumCom{buf: buffer[0:n], n: n}
						tmpFile.Write(buffer[0:n])
					}
					md5r := <-md5c
					files[part.FormName()].Checksum["md5"] = md5r.checksum
				} else {
					return nil, nil, err
				}
			}
		} else if err.Error() != "EOF" {
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

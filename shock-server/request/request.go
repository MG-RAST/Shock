package request

import (
	"crypto/md5"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/MG-RAST/Shock/shock-server/auth"
	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node/archive"
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/MG-RAST/Shock/shock-server/util"
	"github.com/MG-RAST/golib/httpclient"
	"github.com/MG-RAST/golib/stretchr/goweb/context"
	"github.com/jum/tinyftp"
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
	logger.Infof("%s \"%s%s\"", host, url, suffix)
}

// Authenticate _
func Authenticate(req *http.Request) (u *user.User, err error) {
	if _, ok := req.Header["Authorization"]; !ok {
		err = errors.New(e.NoAuth)
		return
	}
	header := req.Header.Get("Authorization")
	u, err = auth.Authenticate(header)
	return
}

// AuthError _
func AuthError(err error, ctx context.Context) error {

	if conf.DEBUG_AUTH {
		return responder.RespondWithError(ctx, http.StatusBadRequest, err.Error())
	}

	if err.Error() == e.InvalidAuth {
		return responder.RespondWithError(ctx, http.StatusBadRequest, "Invalid authorization header or content")
	}
	err_msg := "Error at Auth: " + err.Error()
	logger.Error(err_msg)
	return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
}

// DataUpload helper function to create a node from an http data post (not multi-part)
func DataUpload(r *http.Request) (params map[string]string, files file.FormFiles, err error) {
	params = make(map[string]string)
	files = make(file.FormFiles)
	tmpPath := fmt.Sprintf("%s/temp/%d%d", conf.PATH_DATA, rand.Int(), rand.Int())

	// tmpPath, err := ioutil.TempFile(conf.PATH_CACHE+"/temp/", "")
	// if err != nil {
	// 	log.Fatalf("(DataUpload)  cannot create temporary file: %s [Err: %s]", tmpPath, err.Error())
	// 	return
	// }

	files["upload"] = file.FormFile{Name: "filename", Path: tmpPath, Checksum: make(map[string]string)}
	if tmpFile, err := os.Create(tmpPath); err == nil {
		defer tmpFile.Close()
		md5h := md5.New()
		dst := io.MultiWriter(tmpFile, md5h)
		if _, err = io.Copy(dst, r.Body); err != nil {
			return nil, nil, err
		}
		files["upload"].Checksum["md5"] = fmt.Sprintf("%x", md5h.Sum(nil))
	} else {
		return nil, nil, err
	}

	return
}

// ParseMultipartForm helper function for create & update
func ParseMultipartForm(r *http.Request) (params map[string]string, files file.FormFiles, err error) {

	params = make(map[string]string)
	files = make(file.FormFiles)
	reader, err := r.MultipartReader()
	if err != nil {
		err = fmt.Errorf("(ParseMultipartForm) MultipartReader returned: %s", err.Error())
		return
	}

	tmpPath := ""
	for { // read until EOF
		//logger.Error("ParseMultipartForm forloop")
		var part *multipart.Part
		part, err = reader.NextPart()
		if err != nil {
			//fmt.Printf("(ParseMultipartForm) NextPart returned: %s\n", err.Error())
			//logger.Error("NextPart returned " + err.Error())
			if err == io.EOF {
				err = nil
			} else {
				params = nil
				err = fmt.Errorf("(ParseMultipartForm) reader.NextPart returned: %s", err.Error())
			}

			return
		}
		//fmt.Printf("(ParseMultipartForm) part.FormName(): %s\n", part.FormName())
		//fmt.Printf("(ParseMultipartForm) part.FileName(): %s\n", part.FileName())
		// params don't have a FileName()
		// files must have FormName() of either "upload", "gzip", "bzip2", "attributes", "subset_indices", or an integer
		if part.FileName() == "" {
			if !util.IsValidParamName(part.FormName()) {
				return nil, files, errors.New("invalid param: " + part.FormName())
			}
			buffer := make([]byte, 32*1024)
			var n int
			n, err = part.Read(buffer)
			if n == 0 { // first handle n !
				return
			}
			if err != nil {
				//fmt.Printf("(ParseMultipartForm) part.Read n: %d\n", n)
				//fmt.Printf("(ParseMultipartForm) part.Read: %s\n", err.Error())
				if err != io.EOF {
					// return errors other than EOF
					return
				}
				// ignore EOF
			}

			formValue := fmt.Sprintf("%s", buffer[0:n])
			//fmt.Printf("(ParseMultipartForm) part.FormName(): %s\n", part.FormName())
			if part.FormName() == "upload_url" {
				tempDir := path.Join(conf.PATH_DATA, "temp")
				var tmpFile *os.File
				tmpFile, err = ioutil.TempFile(tempDir, "")
				if err != nil {
					err = fmt.Errorf("(ParseMultipartForm) ioutil.TempFile returned: %s", err.Error())
					return
				}
				defer tmpFile.Close()
				tmpPath = tmpFile.Name()
				//tmpPath = fmt.Sprintf("%s/temp/%d%d", conf.PATH_DATA, rand.Int(), rand.Int())
				files[part.FormName()] = file.FormFile{Name: "", Path: tmpPath, Checksum: make(map[string]string)}
				// download from url
				//var tmpFile *os.File
				// tmpFile, err = os.Create(tmpPath)
				// if err != nil {
				// 	err = fmt.Errorf("(ParseMultipartForm) os.Create returned: %s", err.Error())
				// 	return
				// }

				var tmpform = files[part.FormName()]
				md5h := md5.New()
				dst := io.MultiWriter(tmpFile, md5h)
				var fileName string
				var body io.ReadCloser
				fileName, body, err = fetchFileStream(formValue)
				if err != nil {
					return nil, files, errors.New("unable to stream url: " + err.Error())
				}
				defer body.Close()
				if _, err = io.Copy(dst, body); err != nil {
					err = fmt.Errorf("(ParseMultipartForm) io.Copy returned: %s", err.Error())
					return nil, files, err
				}
				tmpform.Name = fileName
				tmpform.Checksum["md5"] = fmt.Sprintf("%x", md5h.Sum(nil))
				files[part.FormName()] = tmpform

			} else {
				// regular form field
				params[part.FormName()] = formValue
			}
		} else {
			// determine file type
			isSubsetFile := false
			if part.FormName() == "subset_indices" {
				isSubsetFile = true
			}
			isPartsFile := false
			if _, er := strconv.Atoi(part.FormName()); er == nil {
				isPartsFile = true
			}

			//if util.IsValidParamName(part.FormName()) {

			//}

			if !isPartsFile && !util.IsValidFileName(part.FormName()) {
				return nil, files, fmt.Errorf("invalid file param (part.FileName()=%s): %s ", part.FileName(), part.FormName())
			}
			// download it
			tempDir := path.Join(conf.PATH_DATA, "temp")
			var tmpFile *os.File
			tmpFile, err = ioutil.TempFile(tempDir, "")
			if err != nil {
				err = fmt.Errorf("(ParseMultipartForm) ioutil.TempFile returned: %s", err.Error())
				return
			}
			defer tmpFile.Close()

			tmpPath = tmpFile.Name()
			//tmpPath = fmt.Sprintf("%s/temp/%d%d", conf.PATH_DATA, rand.Int(), rand.Int())
			files[part.FormName()] = file.FormFile{Name: part.FileName(), Path: tmpPath, Checksum: make(map[string]string)}
			//var tmpFile *os.File
			// tmpFile, err = os.Create(tmpPath)
			// if err != nil {
			// 	err = fmt.Errorf("(ParseMultipartForm) os.Create returned: %s", err.Error())
			// 	return
			// }

			if util.IsValidUploadFile(part.FormName()) || isPartsFile || isSubsetFile {
				// handle upload or parts files
				var tmpform = files[part.FormName()]
				md5h := md5.New()
				dst := io.MultiWriter(tmpFile, md5h)
				ucReader, ucErr := archive.UncompressReader(part.FormName(), part)
				if ucErr != nil {
					ucErr = fmt.Errorf("(ParseMultipartForm) archive.UncompressReader returned: %s", ucErr.Error())
					return nil, files, ucErr
				}
				if _, err = io.Copy(dst, ucReader); err != nil {
					err = fmt.Errorf("(ParseMultipartForm) io.Copy returned: %s", err.Error())
					return nil, files, err
				}
				if archive.IsValidUncompress(part.FormName()) {
					tmpform.Name = util.StripSuffix(part.FileName())
				}
				tmpform.Checksum["md5"] = fmt.Sprintf("%x", md5h.Sum(nil))
				files[part.FormName()] = tmpform
			} else {
				// handle file where md5 not needed
				if _, err = io.Copy(tmpFile, part); err != nil {
					err = fmt.Errorf("(ParseMultipartForm) io.Copy returned: %s", err.Error())
					return nil, files, err
				}
			}

		}

	}
	return
}

func fetchFileStream(urlStr string) (f string, r io.ReadCloser, err error) {
	u, _ := url.Parse(urlStr)
	if (u.Scheme == "") || (u.Host == "") || (u.Path == "") {
		return "", nil, errors.New("Not a valid url: " + urlStr)
	}
	pathParts := strings.Split(strings.TrimRight(u.Path, "/"), "/")
	fileName := pathParts[len(pathParts)-1]
	cleanPath := strings.Join(pathParts, "/")

	if (u.Scheme == "http") || (u.Scheme == "https") {
		res, err := httpclient.Get(u.String(), httpclient.Header{}, nil, nil)
		if err != nil {
			return "", nil, errors.New("httpclient returned: " + err.Error())
		}
		if res.StatusCode != 200 { //err in fetching data
			resbody, _ := ioutil.ReadAll(res.Body)
			return "", nil, errors.New(fmt.Sprintf("url=%s, res=%s", u.String(), resbody))
		}
		return fileName, res.Body, err
	} else if u.Scheme == "ftp" {
		// set port if missing
		ftpHost := u.Host
		hostParts := strings.Split(u.Host, ":")
		if len(hostParts) == 1 {
			ftpHost = u.Host + ":21"
		}
		c, _, _, err := tinyftp.Dial("tcp", ftpHost)
		if err != nil {
			return "", nil, errors.New("ftpclient returned: " + err.Error())
		}
		defer c.Close()
		if _, _, err = c.Login("", ""); err != nil {
			return "", nil, errors.New("ftpclient returned: " + err.Error())
		}
		dconn, _, _, err := c.RetrieveFrom(cleanPath)
		if err != nil {
			return "", nil, errors.New("ftpclient returned: " + err.Error())
		}
		return fileName, dconn, err
	}
	return "", nil, errors.New("unsupported protocol scheme: " + u.Scheme)
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

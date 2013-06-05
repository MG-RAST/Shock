package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MG-RAST/Shock/shock-client/conf"
	client "github.com/MG-RAST/Shock/shock-client/lib/httpclient"
	"io"
	"io/ioutil"
	"net/http"
	"os/user"
	"strconv"
	"syscall"
	"time"
)

type Opts map[string]string
type Acls struct {
	R bool
	W bool
	D bool
}

func SetTokenAuth(t *Token) {
	client.SetTokenAuth(t.AccessToken)
}

func SetBasicAuth(username, password string) {
	client.SetBasicAuth(username, password)
}

func OAuthToken(username, password string) (t *Token, err error) {
	client.SetBasicAuth(username, password)
	res, err := client.Get(conf.Auth.TokenUrl, client.Header{}, nil)
	if err == nil {
		if res.StatusCode == http.StatusCreated {
			if body, err := ioutil.ReadAll(res.Body); err == nil {
				if err = json.Unmarshal(body, &t); err != nil {
					return nil, err
				}
			}
		} else {
			return nil, errors.New("Authentication failed: Unexpected response status: " + res.Status)
		}
	} else {
		return nil, err
	}
	return
}

func (t *Token) ExpiresInDays() int {
	d := time.Duration(t.ExpiresIn) * time.Second
	return int(d.Hours()) / 24
}

func (t *Token) Path() (path string) {
	u, _ := user.Current()
	return u.HomeDir + "/.shock-client.token"
}

func (t *Token) Delete() (err error) {
	if err := syscall.Unlink(t.Path()); err != nil {
		if err.Error() == "no such file or directory" {
			return nil
		} else {
			return err
		}
	}
	return nil
}

func (t *Token) Load() (err error) {
	if f, err := ioutil.ReadFile(t.Path()); err != nil {
		return err
	} else {
		if err = json.Unmarshal(f, &t); err != nil {
			return err
		}
	}
	return
}

func (t *Token) Store() (err error) {
	m, _ := json.Marshal(t)
	return ioutil.WriteFile(t.Path(), m, 0600)
}

func (n *Node) Create(opts Opts) (err error) {
	return n.createOrUpdate(opts)
}

func (n *Node) Update(opts Opts) (err error) {
	return n.createOrUpdate(opts)
}

func (n *Node) createOrUpdate(opts Opts) (err error) {
	url := conf.Server.Url + "/node"
	method := "POST"
	if n.Id != "" {
		url += "/" + n.Id
		method = "PUT"
	}
	form := client.NewForm()
	if opts.HasKey("attributes") {
		form.AddFile("attributes", opts.Value("attributes"))
	}
	if opts.HasKey("upload_type") {
		switch opts.Value("upload_type") {
		case "full":
			if opts.HasKey("full") {
				form.AddFile("upload", opts.Value("full"))
			} else {
				return errors.New("missing file parameter: upload")
			}
		case "parts":
			if opts.HasKey("parts") {
				form.AddParam("parts", opts.Value("parts"))
			} else {
				return errors.New("missing partial upload parameter: parts")
			}
		case "part":
			if opts.HasKey("part") && opts.HasKey("file") {
				println(opts.Value("part"), opts.Value("file"))
				form.AddParam(opts.Value("part"), opts.Value("file"))
			} else {
				return errors.New("missing partial upload parameter: part or file")
			}
		case "remote_path":
			if opts.HasKey("remote_path") {
				form.AddParam("path", opts.Value("remote_path"))
			} else {
				return errors.New("missing remote path parameter: path")
			}
		case "virtual_file":
			if opts.HasKey("virtual_file") {
				form.AddParam("type", "virtual")
				form.AddParam("source", opts.Value("virtual_file"))
			} else {
				return errors.New("missing virtual node parameter: source")
			}
		}
	}
	form.Create()

	// "Authorization": "OAuth "+token,
	headers := client.Header{
		"Content-Type":   form.ContentType,
		"Content-Length": strconv.FormatInt(form.Length, 10),
	}

	if res, err := client.Do(method, url, headers, form.Reader); err == nil {
		if res.StatusCode == 200 {
			r := WNode{Data: n}
			body, _ := ioutil.ReadAll(res.Body)
			if err = json.Unmarshal(body, &r); err == nil {
				return err
			}
		} else {
			r := Wrapper{}
			body, _ := ioutil.ReadAll(res.Body)
			if err = json.Unmarshal(body, &r); err == nil {
				return errors.New(res.Status + ": " + (*r.Error)[0])
			} else {
				return errors.New("request error: " + res.Status)
			}
		}
	} else {
		return err
	}
	return nil
}

func (n *Node) Download(opts Opts) (download io.Reader, err error) {
	if n.Id == "" {
		return nil, errors.New("missing node Id")
	}
	url := conf.Server.Url + "/node/" + n.Id + "?download"
	if opts.HasKey("index") {
		url += "&index=" + opts.Value("index")
		if opts.HasKey("parts") {
			url += "&part=" + opts.Value("parts")
		} else {
			return nil, errors.New("missing index parameter: part")
		}
		if opts.HasKey("index_options") {
			// index options should be in form key=value
			url += "&" + opts.Value("index_options")
		}
	}

	res, err := client.Get(url, client.Header{}, nil)
	if err == nil {
		if res.StatusCode == 200 {
			return res.Body, nil
		} else {
			r := Wrapper{}
			body, _ := ioutil.ReadAll(res.Body)
			if err = json.Unmarshal(body, &r); err == nil {
				return nil, errors.New(res.Status + ": " + (*r.Error)[0])
			} else {
				return nil, errors.New("request error: " + res.Status)
			}
		}
	}
	return nil, err
}

func (n *Node) String() string {
	m, _ := json.Marshal(n)
	return string(m)
}

func (n *Node) PP() {
	n.PrettyPrint()
}

func (n *Node) PrettyPrint() {
	m, _ := json.MarshalIndent(n, "", "    ")
	fmt.Printf("%s\n", m)
}

func (n *Node) Get() (err error) {
	if n.Id == "" {
		return errors.New("missing node Id")
	}
	url := conf.Server.Url + "/node/" + n.Id
	res, err := client.Get(url, client.Header{}, nil)
	if err == nil {
		if res.StatusCode == 200 {
			r := WNode{Data: n}
			body, _ := ioutil.ReadAll(res.Body)
			if err = json.Unmarshal(body, &r); err == nil {
				return err
			}
		} else {
			r := Wrapper{}
			body, _ := ioutil.ReadAll(res.Body)
			if err = json.Unmarshal(body, &r); err == nil {
				return errors.New(res.Status + ": " + (*r.Error)[0])
			} else {
				return errors.New("request error: " + res.Status)
			}
		}
	}
	return
}

func (n *Node) aclMod(method, acl, users string) (err error) {
	if n.Id == "" {
		return errors.New("missing node Id")
	}
	url := ""
	if acl != "owner" {
		url = conf.Server.Url + "/node/" + n.Id + "/acl?" + acl + "=" + users
	} else {
		url = conf.Server.Url + "/node/" + n.Id + "/acl/owner?users=" + users
	}
	println(method + " : " + url)

	res, err := client.Do(method, url, client.Header{}, nil)
	if err == nil {
		if res.StatusCode == 200 {
			r := WAcl{Data: &n.Acl}
			body, _ := ioutil.ReadAll(res.Body)
			if err = json.Unmarshal(body, &r); err == nil {
				return err
			}
		} else {
			r := Wrapper{}
			body, _ := ioutil.ReadAll(res.Body)
			if err = json.Unmarshal(body, &r); err == nil {
				return errors.New(res.Status + ": " + (*r.Error)[0])
			} else {
				println(err.Error())
				return errors.New("request error: " + res.Status)
			}
		}
	} else {
		return err
	}
	return
}

func (n *Node) AclAdd(acl, users string) (err error) {
	return n.aclMod("PUT", acl, users)
}

func (n *Node) AclRemove(acl, users string) (err error) {
	return n.aclMod("DELETE", acl, users)
}

func (n *Node) AclChown(user string) (err error) {
	return n.aclMod("PUT", "owner", user)
}

func (o *Opts) HasKey(key string) bool {
	if _, has := (*o)[key]; has {
		return true
	}
	return false
}

func (o *Opts) Value(key string) string {
	val, _ := (*o)[key]
	return val
}

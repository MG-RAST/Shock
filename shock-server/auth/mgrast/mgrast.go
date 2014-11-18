// Package globus implements MG-RAST OAuth authentication
package mgrast

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/user"
	"io/ioutil"
	"net/http"
	"strings"
)

type resErr struct {
	error string `json:"error"`
}

type credentials struct {
	Uname string `json:"login"`
	Fname string `json:"firstname"`
	Lname string `json:"lastname"`
	Email string `json:"email"`
}

func authHeaderType(header string) string {
	tmp := strings.Split(header, " ")
	if len(tmp) > 1 {
		return strings.ToLower(tmp[0])
	}
	return ""
}

// Auth takes the request authorization header and returns
// user
func Auth(header string) (*user.User, error) {
	switch authHeaderType(header) {
	case "mgrast", "oauth":
		return authToken(strings.Split(header, " ")[1])
	case "basic":
		return nil, errors.New("This authentication method does not support username/password authentication. Please use MG-RAST your token.")
	default:
		return nil, errors.New("Invalid authentication header.")
	}
}

// authToken validiates token by fetching user information.
func authToken(t string) (u *user.User, err error) {
	client := &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	req, err := http.NewRequest("GET", conf.Conf["mgrast_oauth_url"], nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Auth", t)
	if resp, err := client.Do(req); err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			if body, err := ioutil.ReadAll(resp.Body); err == nil {
				u = &user.User{}
				c := &credentials{}
				if err = json.Unmarshal(body, &c); err != nil {
					return nil, err
				} else {
					if c.Uname == "" {
						return nil, errors.New(e.InvalidAuth)
					}
					u.Username = c.Uname
					u.Fullname = c.Fname + " " + c.Lname
					u.Email = c.Email
					if err = u.SetMongoInfo(); err != nil {
						return nil, err
					}
				}
			}
		} else if resp.StatusCode == http.StatusForbidden {
			return nil, errors.New(e.InvalidAuth)
		} else {
			err_str := "Authentication failed: Unexpected response status: " + resp.Status
			logger.Error(err_str)
			return nil, errors.New(err_str)
		}
	} else {
		return nil, err
	}
	return
}

package globus

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"github.com/MG-RAST/Shock/conf"
	"github.com/MG-RAST/Shock/store/user"
	"io/ioutil"
	"net/http"
	"strings"
)

type Token struct {
	AccessToken     string      `json:"access_token"`
	AccessTokenHash string      `json:"access_token_hash"`
	ClientId        string      `json:"client_id"`
	ExpiresIn       int         `json:"expires_in"`
	Expiry          int         `json:"expiry"`
	IssuedOn        int         `json:"issued_on"`
	Lifetime        int         `json:"lifetime"`
	Scopes          interface{} `json:"scopes"`
	TokenId         string      `json:"token_id"`
	TokeType        string      `json:"token_type"`
	UserName        string      `json:"user_name"`
}

func AuthUsernamePassword(u string, p string) (usr *user.User, err error) {
	if t, err := FetchToken(u, p); err == nil {
		return FetchProfile(t.AccessToken)
	} else {
		return nil, err
	}
	return
}

func AuthToken(t string) (*user.User, error) {
	return FetchProfile(t)
}

func FetchToken(u string, p string) (t *Token, err error) {
	client := &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	req, err := http.NewRequest("GET", conf.GLOBUS_TOKEN_URL, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(u, p)
	if resp, err := client.Do(req); err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusCreated {
			if body, err := ioutil.ReadAll(resp.Body); err == nil {
				if err = json.Unmarshal(body, &t); err != nil {
					return nil, err
				}
			}
		} else {
			return nil, errors.New("Authentication failed: Unexpected response status: " + resp.Status)
		}
	} else {
		return nil, err
	}
	return
}

func FetchProfile(t string) (u *user.User, err error) {
	client := &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	req, err := http.NewRequest("GET", conf.GLOBUS_PROFILE_URL+"/"+clientId(t), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Globus-Goauthtoken "+t)
	if resp, err := client.Do(req); err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			if body, err := ioutil.ReadAll(resp.Body); err == nil {
				u = &user.User{}
				if err = json.Unmarshal(body, &u); err != nil {
					return nil, err
				} else {
					if err = u.SetUuid(); err != nil {
						return nil, err
					}
				}
			}
		} else {
			return nil, errors.New("Authentication failed: Unexpected response status: " + resp.Status)
		}
	} else {
		return nil, err
	}
	return
}

func clientId(t string) string {
	for _, part := range strings.Split(t, "|") {
		if kv := strings.Split(part, "="); kv[0] == "client_id" {
			return kv[1]
		}
	}
	return ""
}

func ValidToken(header string) bool {
	return strings.Contains(header, "Globus-Goauthtoken ") || strings.Contains(header, "Oauth ")
}

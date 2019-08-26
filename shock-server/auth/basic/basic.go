// Package basic implements basic auth decoding and
// self contained user authentication
package basic

import (
	"encoding/base64"
	"errors"
	"strings"

	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/user"
)

// DecodeHeader takes the request authorization header and returns
// username and password if it is correctly encoded.
func DecodeHeader(header string) (username string, password string, err error) {
	headerArray := strings.Split(header, " ")
	if len(headerArray) != 2 {
		err = errors.New(e.InvalidAuth)
		return
	}
	bearer := headerArray[0]
	token := headerArray[1]
	if strings.ToLower(bearer) == "basic" {

		if val, err := base64.URLEncoding.DecodeString(token); err == nil {
			tmp := strings.Split(string(val), ":")
			if len(tmp) >= 2 {
				return tmp[0], tmp[1], nil
			} else {
				return "", "", errors.New(e.InvalidAuth)
			}
		} else {
			return "", "", errors.New(e.InvalidAuth)
		}

	}
	return "", "", errors.New(e.InvalidAuth)
}

// Auth takes the request authorization header and returns
// user
func Auth(header string) (u *user.User, err error) {
	//fmt.Printf("auth: %s\n", header)
	username, password, err := DecodeHeader(header)
	//fmt.Printf("auth: %s %s\n", username, password)
	if err == nil {
		return user.FindByUsernamePassword(username, password)
	}
	return nil, err
}

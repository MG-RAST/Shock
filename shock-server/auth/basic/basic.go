// Package basic implements basic auth decoding and
// self contained user authentication
package basic

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/user"
)

// DecodeHeader takes the request authorization header and returns
// username and password if it is correctly encoded.
func DecodeHeader(header string) (username string, password string, err error) {
	headerArray := strings.Split(header, " ")
	if len(headerArray) != 2 {
		if conf.DEBUG_AUTH {
			err = errors.New("(basic/DecodeHeader) len(headerArray) != 2")
			return
		}
		err = errors.New(e.InvalidAuth)
		return
	}
	bearer := headerArray[0]
	token := headerArray[1]
	if strings.ToLower(bearer) == "basic" {

		var val []byte
		val, err = base64.URLEncoding.DecodeString(token)
		if err != nil {
			if conf.DEBUG_AUTH {
				err = errors.New("(basic/DecodeHeader) " + err.Error())
				return
			}
			return "", "", errors.New(e.InvalidAuth)
		}

		tmp := strings.Split(string(val), ":")
		if len(tmp) <= 1 {
			if conf.DEBUG_AUTH {
				err = errors.New("(basic/DecodeHeader) len(tmp) <=1")
				return
			}
			err = errors.New(e.InvalidAuth)
			return
		}
		username = tmp[0]
		password = tmp[1]
		return
	}

	if conf.DEBUG_AUTH {
		err = errors.New("(basic/DecodeHeader) bearer \"basic\" is missing")
		return
	}
	err = errors.New(e.InvalidAuth)
	return

}

// Auth takes the request authorization header and returns
// user
func Auth(header string) (u *user.User, err error) {
	//fmt.Printf("auth: %s\n", header)
	username, password, err := DecodeHeader(header)
	if err != nil {
		if conf.DEBUG_AUTH {
			err = fmt.Errorf("(Basic/Auth) DecodeHeader returned: %s", err.Error())
		}
		return
	}
	//fmt.Printf("auth: %s %s\n", username, password)

	u, err = user.FindByUsernamePassword(username, password)
	if err != nil {
		if conf.DEBUG_AUTH {
			err = fmt.Errorf("(Basic/Auth) user.FindByUsernamePassword returned: %s", err.Error())
		}
		return
	}
	return
}

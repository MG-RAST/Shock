package basic

import (
	"encoding/base64"
	"errors"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/user"
	"strings"
)

func DecodeHeader(header string) (string, string, error) {
	if strings.Split(header, " ")[0] == "Basic" {
		if val, err := base64.URLEncoding.DecodeString(strings.Split(header, " ")[1]); err == nil {
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

func Auth(username string, password string) (u *user.User, err error) {
	return user.FindByUsernamePassword(username, password)
}

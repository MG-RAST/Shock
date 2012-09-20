package auth

import (
	"github.com/MG-RAST/Shock/conf"
	"github.com/MG-RAST/Shock/store/user"
	"github.com/MG-RAST/Shock/store/user/auth/basic"
	"github.com/MG-RAST/Shock/store/user/auth/globus"
)

func Authenticate(header string) (u *user.User, err error) {
	switch conf.AUTH_TYPE {
	case "globus":
		if globus.ValidToken(header) {
			// check cache
			// auth from server			
			if u, err = globus.AuthToken(header); err == nil {
				return
			} else {
				return nil, err
			}
			// cache results
		} else {
			if username, password, err := basic.DecodeHeader(header); err == nil {
				if u, err := globus.AuthUsernamePassword(username, password); err == nil {
					return u, nil
				} else {
					return nil, err
				}
			} else {
				return nil, err
			}
		}
	case "oauth":
		// stub
	case "basic":
		if username, password, err := basic.DecodeHeader(header); err == nil {
			return AuthByUsernamePassword(username, password)
		} else {
			return nil, err
		}
	}
	return
}

func AuthByUsernamePassword(username string, password string) (u *user.User, err error) {
	if d, err := user.DBConnect(); err == nil {
		defer d.Close()
		u = &user.User{Username: username, Password: password}
		if err = d.GetUser(u); err != nil {
			u = nil
		}
	}
	return
}

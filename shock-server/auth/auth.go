package auth

import (
	"errors"
	"github.com/MG-RAST/Shock/shock-server/auth/basic"
	"github.com/MG-RAST/Shock/shock-server/auth/globus"
	"github.com/MG-RAST/Shock/shock-server/auth/mgrast"
	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/user"
)

var authCache cache
var authMethods []func(string) (*user.User, error)

func Initialize() {
	authCache = cache{m: make(map[string]cacheValue)}
	authMethods = []func(string) (*user.User, error){}
	if conf.Conf["basic_auth"] != "" {
		authMethods = append(authMethods, basic.Auth)
	}
	if conf.Conf["globus_token_url"] != "" && conf.Conf["globus_profile_url"] != "" {
		authMethods = append(authMethods, globus.Auth)
	}
	if conf.Conf["mgrast_oauth_url"] != "" {
		authMethods = append(authMethods, mgrast.Auth)
	}
}

func Authenticate(header string) (u *user.User, err error) {
	if u = authCache.lookup(header); u != nil {
		return u, nil
	} else {
		for _, auth := range authMethods {
			if u, _ := auth(header); u != nil {
				authCache.add(header, u)
				return u, nil
			}
		}
	}
	return nil, errors.New(e.InvalidAuth)
}

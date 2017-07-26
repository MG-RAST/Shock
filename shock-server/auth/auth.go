// Package auth implements http request authentication
package auth

import (
	"errors"
	"github.com/MG-RAST/Shock/shock-server/auth/globus"
	"github.com/MG-RAST/Shock/shock-server/auth/oauth"
	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/user"
)

// authCache is a
var authCache cache
var authMethods []func(string) (*user.User, error)

func Initialize() {
	authCache = cache{m: make(map[string]cacheValue)}
	authMethods = []func(string) (*user.User, error){}
	if len(conf.AUTH_OAUTH) > 0 {
		authMethods = append(authMethods, oauth.Auth)
	}
	if conf.AUTH_GLOBUS_TOKEN_URL != "" && conf.AUTH_GLOBUS_PROFILE_URL != "" {
		authMethods = append(authMethods, globus.Auth)
	}
}

func Authenticate(header string) (u *user.User, err error) {
	var lastErr error
	if u = authCache.lookup(header); u != nil {
		return u, nil
	} else {
		for _, auth := range authMethods {
			u, err := auth(header)
			if u != nil && err == nil {
				authCache.add(header, u)
				return u, nil
			}
			if err != nil {
				lastErr = err
			}
		}
	}
	// log actual error, return consistant invalid auth to user
	if lastErr != nil {
		logger.Error("Err@auth.Authenticate: " + lastErr.Error())
	}
	return nil, errors.New(e.InvalidAuth)
}

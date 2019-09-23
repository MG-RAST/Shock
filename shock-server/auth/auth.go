// Package auth implements http request authentication
package auth

import (
	"errors"
	"fmt"

	"github.com/MG-RAST/Shock/shock-server/auth/basic"
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

// Initialize _
func Initialize() {
	authCache = cache{m: make(map[string]cacheValue)}
	authMethods = []func(string) (*user.User, error){}
	if len(conf.AUTH_OAUTH) > 0 {
		authMethods = append(authMethods, oauth.Auth)
	}
	if conf.AUTH_GLOBUS_TOKEN_URL != "" && conf.AUTH_GLOBUS_PROFILE_URL != "" {
		authMethods = append(authMethods, globus.Auth)
	}

	if conf.AUTH_BASIC {
		authMethods = append(authMethods, basic.Auth)
	}
}

// Authenticate _
func Authenticate(header string) (u *user.User, err error) {
	u = authCache.lookup(header)
	if u != nil {
		return
	}

	for _, auth := range authMethods {
		u, err = auth(header)
		if err != nil {

			if conf.DEBUG_AUTH && len(authMethods) == 1 {
				err = fmt.Errorf("(Authenticate) authMethod returned: %s", err.Error())
				return
			}

			// log actual error, return consistant invalid auth to user
			logger.Error("Err@auth.Authenticate: " + err.Error())
			continue
		}

		if u != nil {
			authCache.add(header, u)
		}
		return

	}

	if conf.DEBUG_AUTH {
		err = fmt.Errorf("(Authenticate) No authMethod matched (count: %d)", len(authMethods))
		return
	}
	return nil, errors.New(e.InvalidAuth)
}

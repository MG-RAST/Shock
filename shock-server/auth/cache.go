package auth

import (
	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/user"
	"sync"
	"time"
)

type cache struct {
	sync.Mutex
	m map[string]cacheValue
}

type cacheValue struct {
	expires time.Time
	user    *user.User
}

func (c *cache) lookup(header string) *user.User {
	if v, ok := c.m[header]; ok {
		if time.Now().Before(v.expires) {
			return v.user
		} else {
			c.Lock()
			defer c.Unlock()
			delete(c.m, header)
		}
	}
	return nil
}

func (c *cache) add(header string, u *user.User) {
	c.Lock()
	defer c.Unlock()
	c.m[header] = cacheValue{
		expires: time.Now().Add(time.Duration(conf.AUTH_CACHE_TIMEOUT) * time.Minute),
		//expires: time.Now().Add(1 * time.Minute),
		user: u,
	}
	return
}

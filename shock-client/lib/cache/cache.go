package cache

/*
   shock client package
       - async (cache)
       - get 
       - put (slow bleed)
       - invalidate
*/

import (
	"fmt"
	"github.com/MG-RAST/Shock/shock-client/conf"
	"math/rand"
	"os"
	"time"
)

type Cache struct {
	Dir         string
	Server      string
	getChan     chan message
	putChan     chan message
	connectPool chan bool
}

type message struct {
	str  string
	wait chan bool
}

func New() (c *Cache, err error) {
	c = &Cache{
		Dir:         conf.Cache.Dir,
		Server:      conf.Server.Url,
		getChan:     make(chan message),
		putChan:     make(chan message),
		connectPool: make(chan bool, conf.Cache.MaxConnections),
	}
	go func() {
		for {
			select {
			case m := <-c.getChan:
				//c.connectPool

				time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)
				if m.wait != nil {
					m.wait <- true
				}
			case m := <-c.putChan:
				fmt.Println("Running: put:" + m.str)
			}
		}
	}()
	return
}

func (c *Cache) Get(uuid string) {
	m := message{str: uuid, wait: make(chan bool)}
	c.getChan <- m
	<-m.wait
	return
}

func (c *Cache) GetAsync(uuid string) {
	m := message{str: uuid, wait: nil}
	c.getChan <- m
	return
}

func (c *Cache) Put(uuid, remote string) {
	m := message{str: uuid, wait: make(chan bool)}
	c.putChan <- m
	<-m.wait
}

func (c *Cache) PutAsync(uuid, remote string) {
	m := message{str: uuid, wait: nil}
	c.putChan <- m
}

func (c *Cache) Delete(uuid string) {
	return
}

func exists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func initCache() {
	d := conf.Cache.Dir
	if !exists(d + "/temp") {
		os.Mkdir(d+"/temp", 0777)
	}
}

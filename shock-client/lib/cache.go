package lib

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
	"os"
)

type Cache struct {
	Dir     string
	Server  string
	getChan chan message
	putChan chan message
}

type message struct {
	str  string
	wait chan bool
}

func New() (c *Cache, err error) {
	c = &Cache{
		Dir:     conf.Cache.Dir,
		Server:  conf.Server.Url,
		getChan: make(chan message),
		putChan: make(chan message),
	}

	for i := 0; i < conf.Cache.MaxConnections; i++ {
		go func() {
			for {
				select {
				case m := <-c.getChan:
					fmt.Println("get: Grabbing Object: " + m.str)
					fmt.Println("get: Splitting into part: " + m.str)
				case m := <-c.putChan:
					fmt.Println("put: Spliting: " + m.str)
					fmt.Println("put: Dumping to connectionPool: " + m.str)
				}
			}
		}()
	}
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

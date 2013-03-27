package cache_test

import (
	"fmt"
	. "github.com/MG-RAST/Shock/shock-client/cache"
	"testing"
)

func TestNew(t *testing.T) {
	c, _ := New()
	c.GetAsync("foo1")
	c.Get("foo2")
	c.Get("bar3")
}

func TestPartionFile(t *testing.T) {
	if res, err := PartionFile("localhost:8080/node/1", "/Users/jared/ANL/Calhoun_2.fna"); err == nil {
		fmt.Printf("%#v\n", res)
		for _, p := range res.Parts {
			fmt.Printf("%d\t%d\t%d\n", p.Part, p.Offset, p.Length)
		}
	} else {
		println("printing error")
		println(err.Error())
	}
	return
}

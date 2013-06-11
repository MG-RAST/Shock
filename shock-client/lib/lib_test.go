package lib_test

import (
	//"io/ioutil"
	//"io"
	"fmt"
	. "github.com/MG-RAST/Shock/shock-client/lib"
	"testing"
)

func TestCreate(t *testing.T) {
	n := Node{}
	if err := n.Create(Opts{"attributes": ""}); err != nil {
		fmt.Printf("%s\n", err.Error())
	}
	fmt.Printf("%s\n", n.String())
	n.PP()
}

/*
func TestDownload(t *testing.T) {
	n := Node{}
	if err := n.Create(Opts{"upload_type": "full", "upload": ""}); err != nil {
		fmt.Printf("%s\n", err.Error())
	} else {
		fmt.Printf("%#v\n", n)
	}
	if fh, err := n.Download(Opts{}); err != nil {
		fmt.Printf("%s\n", err.Error())
	} else {
	    f, _ := ioutil.TempFile("/tmp", "test-download")
        if w, err := io.Copy(f, fh); err == nil {
            fmt.Printf("Full file: %d bytes\n", w)
        }
	}

    if fh, err := n.Download(Opts{"index": "size", "index_options": "chunksize=1024", "part": "1"}); err != nil {
    	fmt.Printf("%s\n", err.Error())
    } else {
        f, _ := ioutil.TempFile("/tmp", "test-download")
        if w, err := io.Copy(f, fh); err == nil {
            fmt.Printf("Size index 1024 btyes: %d bytes\n", w)
        }
    }
}

func TestAcls(t *testing.T) {
	SetBasicAuth("", "")
	n := Node{}
	if err := n.Create(Opts{"attributes": ""}); err != nil {
		fmt.Printf("%s\n", err.Error())
	}
	if err := n.AclAdd("write", "frank@gmail.com,bob@gmail.com"); err != nil {
		fmt.Printf("%s\n", err.Error())
	} else {
		fmt.Printf("After add: %#v\n", n.Acl)
	}
	if err := n.AclRemove("write", "frank@gmail.com,bob@gmail.com"); err != nil {
		fmt.Printf("%s\n", err.Error())
	} else {
		fmt.Printf("After rm: %#v\n", n.Acl)
	}
	if err := n.AclChown("frank@gmail.com"); err != nil {
		fmt.Printf("%s\n", err.Error())
	} else {
		fmt.Printf("After chown: %#v\n", n.Acl)
	}
}
*/

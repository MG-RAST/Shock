package httpclient

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestForm(t *testing.T) {
	f := NewForm()
	f.AddParam("upload_type", "true")
	f.AddParam("parts", "25")

	gopath := os.Getenv("GOPATH")
	f.AddFile("upload", gopath+"/src/github.com/MG-RAST/Shock/shock-server/testdata/nr_subset1.fa")
	f.AddFile("upload", gopath+"/src/github.com/MG-RAST/Shock/shock-server/testdata/nr_subset2.fa")
	if err := f.Create(); err != nil {
		println(err.Error())
	}
	println("Content-Type: " + f.ContentType)
	fmt.Printf("Content-Length: %d\n", f.Length)
	if form, err := ioutil.ReadAll(f.Reader); err != nil {
		println(err.Error())
	} else {
		fmt.Printf("%s", form)
		fmt.Printf("len: %d\n", len(form))
	}
}

package store_test

import (
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	. "github.com/MG-RAST/Shock/shock-server/store"
	"github.com/MG-RAST/Shock/shock-server/store/user"
	"math/rand"
	"os"
	"testing"
)

var test_uuid = ""
var test_uuids = []string{"703ce2a0-54cf-4af1-ab14-2de05305f823", "e4ca342a-2338-448b-bdea-d890370eddc2", "883c5c9d-3881-4647-9c9e-4d283cc9c253"}
var temp_node_id string

//var params map[string]string

func TestNewNode(t *testing.T) {
	fmt.Println("in testing NewNode()...")
	newnode := NewNode()

	err := newnode.Mkdir()
	if err != nil {
		t.Errorf("failed in NewNode: %v", err.Error())
	}

	err = newnode.Save()
	if err != nil {
		t.Errorf("failed in NewNode: %v", err.Error())
	}

	fmt.Println(newnode)

	temp_node_id = newnode.Id
	fmt.Println("new node id=", temp_node_id)
	if newnode == nil {
		t.Errorf("failed in NewNode")
	}
}

func TestLoadNode(t *testing.T) {
	fmt.Println("in testing LoadNode()")
	node, err := LoadNode(temp_node_id, test_uuid)
	fmt.Printf("node=%#v\n", node)
	if err != nil {
		fmt.Errorf("Load Node error: %v ", err)
	}
}

func TestReLoadNodeFromDisk(t *testing.T) {
	fmt.Println("in TestReLoadNodeFromDisk()...")
	node, err := LoadNodeFromDisk(temp_node_id)
	fmt.Printf("node=%#v\n", node)
	if err != nil {
		fmt.Errorf("Load Node error: %v ", err)
	}
}

func TestCreateNodeUpload(t *testing.T) {
	fmt.Println("in TestCreateNodeUpload()")

	u := &user.User{Uuid: ""}

	params := make(map[string]string)
	params["key1"] = "value1"
	params["key2"] = "value2"
	fmt.Println("params=", params)

	files := make(FormFiles)
	tmpPath := fmt.Sprintf("%s/temp/%d%d", conf.DATA_PATH, rand.Int(), rand.Int())
	fmt.Println("tmpPath=", tmpPath)
	formfile1 := FormFile{Name: "./testdata/10kb.fna", Path: tmpPath, Checksum: make(map[string]string)}
	files["file1"] = formfile1
	fmt.Println("files=", formfile1)

	node, err := CreateNodeUpload(u, params, files)
	fmt.Printf("node=%#v\n", node)

	if err != nil {
		fmt.Errorf("CreateNodeUpload error: %v ", err)
	}

}

func TestMultiReaderAt(t *testing.T) {
	fmt.Println("in TestMultiReaderAt()")
	readers := []ReaderAt{}
	for _, fn := range []string{"./testdata/10kb.fna", "./testdata/40kb.fna"} {
		if r, err := os.Open(fn); err == nil {
			readers = append(readers, r)
		} else {
			fmt.Errorf("MultiReaderAt error: %v ", err)
		}
	}
	mr := MultiReaderAt(readers...)
	fmt.Printf("mutlireaderat: %#v\n", mr)
	buffer := make([]byte, 2000)
	fmt.Println("---------------")
	for _, offset := range []int64{0, 100, 1000, 11000, 25000, 56000} {
		n, _ := mr.ReadAt(buffer, offset)
		fmt.Printf("read: %d\n", n)
		fmt.Printf("buffer: \n%s\n", buffer[0:n])
		fmt.Println("---------------")
	}
}

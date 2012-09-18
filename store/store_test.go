package store_test

import (
	"fmt"
	. "github.com/MG-RAST/Shock/store"
	"github.com/MG-RAST/Shock/store/user"
	"github.com/MG-RAST/Shock/conf"
	"math/rand"
	"testing"
)

var test_uuid = ""
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
	fmt.Printf("node=%v", node)
	if err != nil {
		fmt.Errorf("Load Node error: %v ", err)
	}
}

func TestReLoadNodeFromDisk(t *testing.T) {
    fmt.Println("in TestReLoadNodeFromDisk()...")
    node, err := LoadNodeFromDisk(temp_node_id)
	fmt.Printf("node=%v", node)
	if err != nil {
		fmt.Errorf("Load Node error: %v ", err)
	}
}

func TestCreateNodeUpload(t *testing.T) {
    fmt.Println("in TestCreateNodeUpload()");
    
    u := &user.User{Uuid: ""}
    
    params := make(map[string]string)
    params["key1"] = "value1"
    params["key2"] = "value2"
    fmt.Println("params=", params)  
    
    files := make(FormFiles)
    tmpPath := fmt.Sprintf("%s/temp/%d%d", conf.DATAPATH, rand.Int(), rand.Int())
    fmt.Println("tmpPath=", tmpPath)
    formfile1 := FormFile{Name: "tshock", Path: tmpPath, Checksum: make(map[string]string)}
    files["file1"] = formfile1  
    fmt.Println("files=",formfile1)
    
    node, err := CreateNodeUpload(u, params, files)
    fmt.Printf("node=%v", node)
    
    if err != nil {
    	fmt.Errorf("CreateNodeUpload error: %v ", err)
	}
    
    
}

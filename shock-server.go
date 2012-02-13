package main

import (
	"flag"
	"fmt"
	//"goweb"
)

// Command line options
var (
	PORT = flag.Int("port", 8000, "the port to listen on")
	DATAROOT = "/Users/jared/projects/GoShockData"
)

func init() {}

func main() {
	flag.Parse()

	n, err := CreateNode("notes","test.json"); if err != nil {
		fmt.Println("hells bells: "+err.String())
	}
	err = n.Save(); if err != nil {
		fmt.Println("hells bells: "+err.String())
	}
	fmt.Println(n.Path())	 
	
	//goweb.MapRest("/node", new(NodeController))
	//goweb.ListenAndServe(":"+fmt.Sprintf("%d", *PORT))  
}


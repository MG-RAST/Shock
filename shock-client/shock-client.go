package main

import (
	"flag"
	"fmt"
	sc "github.com/MG-RAST/go-shock-client"
	"net/url"
	"os"
	"strings"
)

var (
	flags      *flag.FlagSet
	archive    string
	attributes string
	bearer     string
	chunk      int
	column     int
	compressed string
	copy       string
	direction  string
	expiration string
	filename   string
	filepath   string
	force      bool
	format     string
	index      string
	limit      int
	md5        bool
	offset     int
	order      string
	part       string
	remote     string
	shock_url  string
	token      string
	unexpire   bool
	virtual    string
)

func stub(x string) {
	fmt.Println("hello world: " + x)
}

func main() {
	if len(os.Args) < 2 {
		exitError("missing command")
	}

	flags = setFlags()
	flags.Parse(os.Args[1:])
	args := flags.Args()

	url, auth := getUserInfo()
	client := sc.NewShockClient(url, auth)

	command := args[0]
	switch args[0] {
	case "help":
		exitHelp()
	case "info":
		stub(command)
	case "create", "update":
		stub(command)
	case "index":
		stub(command)
	case "delete":
		stub(command)
	case "get":
		stub(command)
	case "query":
		stub(command)
	case "download":
		stub(command)
	case "acl", "public", "chown":
		stub(command)
	}

}

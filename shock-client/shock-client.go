package main

import (
	"flag"
	"fmt"
	sc "github.com/MG-RAST/go-shock-client"
	"os"
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
	index      string
	limit      int
	md5        bool
	offset     int
	order      string
	output     string
	part       string
	pretty     bool
	remote     string
	shock_url  string
	token      string
	unexpire   bool
	virtual    string
)

func stub(x string) {
	fmt.Println("not implamented: " + x)
}

func main() {
	if len(os.Args) < 2 {
		exitError("missing command")
	}
	command := os.Args[1]
	if command == "help" {
		exitHelp()
	}

	flags = setFlags()
	flags.Parse(os.Args[2:])
	//args := flags.Args()

	url, auth := getUserInfo()
	client := sc.NewShockClient(url, auth, true)

	switch command {
	case "help":
		exitHelp()
	case "info":
		info, err := client.ServerInfo()
		if err != nil {
			exitError(err.Error())
		}
		exitOutput(&info)
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
	default:
		exitError("invalid command: " + command)
	}

}

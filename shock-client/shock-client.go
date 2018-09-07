package main

import (
	"flag"
	"fmt"
	sc "github.com/MG-RAST/go-shock-client"
	"io"
	"net/url"
	"os"
	"strconv"
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
	length     int
	limit      int
	md5        bool
	offset     int
	order      string
	output     string
	part       string
	pretty     bool
	remote     string
	seek       int
	shock_url  string
	token      string
	unexpire   bool
	virtual    string
)

func stub(x string) {
	fmt.Println("not implamented: " + x)
}

func buildDownloadUrl(host string, id string) string {
	var query url.Values
	query.Add("download", "")

	if (index != "") && (part != "") {
		if !validateCV("index", index) {
			exitError("invalid index type")
		}
		query.Add("index", index)
		query.Add("part", part)
	} else if (seek > -1) && (length > 0) {
		query.Add("seek", strconv.Itoa(seek))
		query.Add("length", strconv.Itoa(length))
	}

	var myurl *url.URL
	myurl, err := url.ParseRequestURI(host)
	if err != nil {
		exitError("error parsing shock url")
	}
	(*myurl).Path = "/node/" + id
	(*myurl).RawQuery = query.Encode()

	return myurl.String()
}

func main() {
	if len(os.Args) < 2 {
		exitError("missing command")
	}
	command := os.Args[1]
	if (command == "help") || (command == "-h") || (command == "--help") {
		exitHelp()
	}

	flags = setFlags()
	flags.Parse(os.Args[2:])
	args := flags.Args()

	host, auth := getUserInfo()
	client := sc.NewShockClient(host, auth, false)

	switch command {
	case "info":
		info, err := client.ServerInfo()
		if err != nil {
			exitError(err.Error())
		}
		exitOutput(&info)
	case "create", "update":
		stub(command)
	case "index":
		if len(args) < 2 {
			exitError("missing required arguments")
		}
		if !validateCV("index", args[1]) {
			exitError("invalid index type")
		}
		if (args[1] == "column") && (column < 1) {
			exitError("invalid column position")
		}
		err := client.PutIndexQuery(args[0], args[1], force, column)
		if err != nil {
			exitError(err.Error())
		}
	case "delete":
		if len(args) < 1 {
			exitError("missing required ID")
		}
		err := client.DeleteNode(args[0])
		if err != nil {
			exitError(err.Error())
		}
	case "get":
		if len(args) < 1 {
			exitError("missing required ID")
		}
		node, err := client.GetNode(args[0])
		if err != nil {
			exitError(err.Error())
		}
		exitOutput(&node)
	case "query":
		stub(command)
	case "download":
		if len(args) < 1 {
			exitError("missing required ID")
		}
		downUrl := buildDownloadUrl(host, args[0])
		if (output == "") || (output == "-") || (output == "stdout") {
			body, err := sc.FetchShockStream(downUrl, auth)
			if err != nil {
				exitError(err.Error())
			}
			defer body.Close()
			_, err = io.Copy(os.Stdout, body)
			if err != nil {
				exitError(err.Error())
			}
		} else {
			size, md5sum, err := sc.FetchFile(output, downUrl, auth, "", md5)
			if err != nil {
				exitError(err.Error())
			}
			fmt.Printf("download complete\nfile\t%s\nsize\t%d\nmd5\t%s\n", output, size, md5sum)
		}
	case "acl":
		if (len(args) > 1) && (args[1] == "get") {
			acl, err := client.GetAcl(args[0])
			if err != nil {
				exitError(err.Error())
			}
			exitOutput(&acl.Data)
		} else if len(args) > 3 {
			if !validateCV("acl", args[2]) {
				exitError("invalid acl type")
			}
			if args[1] == "add" {
				err := client.PutAcl(args[0], args[2], args[3])
				if err != nil {
					exitError(err.Error())
				}
			}
			if args[1] == "delete" {
				err := client.DeleteAcl(args[0], args[2], args[3])
				if err != nil {
					exitError(err.Error())
				}
			}
		} else {
			exitError("missing required arguments")
		}
	case "public":
		if len(args) < 1 {
			exitError("missing required ID")
		}
		err := client.MakePublic(args[0])
		if err != nil {
			exitError(err.Error())
		}
	case "chown":
		if len(args) < 2 {
			exitError("missing required arguments")
		}
		err := client.ChownNode(args[0], args[1])
		if err != nil {
			exitError(err.Error())
		}
	default:
		exitError("invalid command: " + command)
	}
	os.Exit(0)
}

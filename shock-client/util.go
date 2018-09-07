package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

const USAGE = `Usage: shock-client <command> [options...] [args..]

Global Options:
    --shock_url=<url>           URL of shock-server, or env SHOCK_URL
    --token=<s>                 User token for authorization, or env TOKEN
    --bearer=<s>                Bearer token, default is 'mgrast', or env BEARER
    --output=<p>                File to write ouput too, default is stdout
    --pretty                    JSON output is pretty-print format, default is condensed

Commands:

help / --help / -h              This help message

info
    Note: returns shock-server info from base url

create [options...]
    --attributes=<p>            JSON formated attribute file
                                Note: Attributes will replace all current attributes
    --chunk=<i>                 Upload in chunks, able to resume if disconnected
                                Note: for --filepath option only
    --filename=<s>              Filename to use, default is upload-file basename
    --expiration=<s>            Epiration time: [0-9]+[MHD], default is no expiration
    --compressed=<gzip|bzip2>   Uncompress after upload, default is off
    --archive=<tar|zip>         Unpack after upload, default is off
    
    Mutualy exclusive options:
    --filepath=<p>              Path to file
    --part=<s>                  Number of parts to be uploaded
    --virtual=<s>               Comma seperated list of node ids
    --remote=<p>                URI of remote file
    --copy=<id>                 ID of node to copy

update [options...] <id>
    Note: options are the same as create with additional below
    --unexpire                  Remove expiration
    --part=<s> --filepath=<p>   The part number to be uploaded and path to file
                                Note: parts must be set

index [options...] <id> <line|column|record|chunkrecord|bai>
    --force                     Forces index to be rebuilt if exists
    --column=<i>                Column number to index by if using 'column' type

delete [options...] <id>

get [options...] <id>

query [options...] 
    --limit=<i>                 Max number of output nodes, default is 25
    --offset=<i>                Position to start rerieving nodes at, default is 0
    --order=<s>                 Field to sort nodes by, default is 'created_on'
    --direction=<asc|desc>      Direction to sort by, default is 'desc'
    --attributes=<s>            Attribute field and value to search with
                                Note: use "FIELD_NAME:VALUE"
    --other=<s>                 Any non-attributes node field and value to search with
    --distinct=<s>              An attribute field name, returns all unique values in that field

download [options...] <id>
    --md5                       Create md5 checksum of downloaded file
                                Note: requires output file
    Mutualy exclusive options:
    --index=<s> --part={s}      Name of index and part(s) to retrieve, may be a range eg. 1-10
    --seek=<i> --length=<i>     Download bytes from file for given seek and length

acl [options...] <id> <get|add|delete> <all|read|write|delete> <users>
    Note: users are in the form of comma delimited list of user-names or uuids,
          not required for get action

public [options...] <id>

chown [options...] <id> <user>
    Note: user is user-name or uuid 

`

var CV = map[string]map[string]bool{
	"acl":        map[string]bool{"all": true, "delete": true, "read": true, "write": true},
	"archive":    map[string]bool{"tar": true, "zip": true},
	"compressed": map[string]bool{"bzip2": true, "gzip": true},
	"direction":  map[string]bool{"asc": true, "desc": true},
	"index":      map[string]bool{"bai": true, "chunkrecord": true, "column": true, "line": true, "record": true, "size": true},
}

func validateCV(name string, value string) bool {
	if _, ok := CV[name]; ok {
		if _, ok := CV[name][value]; ok {
			return true
		}
	}
	return false
}

func exitHelp() {
	fmt.Fprintln(os.Stdout, USAGE)
	os.Exit(0)
}

func exitError(msg string) {
	fmt.Fprintln(os.Stderr, USAGE)
	if msg != "" {
		fmt.Fprintln(os.Stderr, "Error: "+msg)
	}
	os.Exit(1)
}

func exitOutput(v interface{}) {
	var b []byte
	var e error
	if pretty {
		b, e = json.MarshalIndent(v, "", "   ")
	} else {
		b, e = json.Marshal(v)
	}
	if e != nil {
		exitError(e.Error())
	}
	if (output == "") || (output == "-") || (output == "stdout") {
		fmt.Println(string(b))
	} else {
		b = append(b, '\n')
		e = ioutil.WriteFile(output, b, 0644)
		if e != nil {
			exitError(e.Error())
		}
	}
	os.Exit(0)
}

func setFlags() (flags *flag.FlagSet) {
	flags = flag.NewFlagSet("shockclient", flag.ContinueOnError)
	flags.StringVar(&archive, "archive", "", "")
	flags.StringVar(&attributes, "attributes", "", "")
	flags.StringVar(&bearer, "bearer", "mgrast", "")
	flags.IntVar(&chunk, "chunk", 0, "")
	flags.IntVar(&column, "column", 0, "")
	flags.StringVar(&compressed, "compressed", "", "")
	flags.StringVar(&copy, "copy", "", "")
	flags.StringVar(&direction, "direction", "", "")
	flags.StringVar(&expiration, "expiration", "", "")
	flags.StringVar(&filename, "filename", "", "")
	flags.StringVar(&filepath, "filepath", "", "")
	flags.BoolVar(&force, "force", false, "")
	flags.StringVar(&index, "index", "", "")
	flags.IntVar(&length, "length", 0, "")
	flags.IntVar(&limit, "limit", 0, "")
	flags.BoolVar(&md5, "md5", false, "")
	flags.IntVar(&offset, "offset", 0, "")
	flags.StringVar(&order, "order", "", "")
	flags.StringVar(&output, "output", "stdout", "")
	flags.StringVar(&part, "part", "", "")
	flags.BoolVar(&pretty, "pretty", false, "")
	flags.StringVar(&remote, "remote", "", "")
	flags.IntVar(&seek, "seek", -1, "")
	flags.StringVar(&shock_url, "shock_url", "", "")
	flags.StringVar(&token, "token", "", "")
	flags.BoolVar(&unexpire, "unexpire", false, "")
	flags.StringVar(&virtual, "virtual", "", "")
	return
}

func getUserInfo() (host string, auth string) {
	// set from env if exists
	if shock_url == "" {
		shock_url = os.Getenv("SHOCK_URL")
	}
	if token == "" {
		token = os.Getenv("TOKEN")
	}
	if os.Getenv("BEARER") != "" {
		bearer = os.Getenv("BEARER")
	}
	// test and return
	if token != "" {
		auth = bearer + " " + token
	}
	if shock_url == "" {
		exitError("missing required --shock_url or SHOCK_URL")
	}
	host = shock_url
	return
}

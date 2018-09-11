package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"
)

const USAGE = `Usage: shock-client <command> [options...] [args..]

Global Options:
    --shock_url=<url>           URL of shock-server, or env SHOCK_URL
    --token=<s>                 User token for authorization, or env TOKEN
    --bearer=<s>                Bearer token, default is 'mgrast', or env BEARER
    --output=<p>                File to write ouput too, default is stdout
    --pretty                    JSON output is pretty-print format, default is condensed
    --debug                     run shock client in debug mode

Commands:

help / --help / -h              This help message

info
    Note: returns shock-server info from base url

create [options...]
    --expiration=<s>            Epiration time: use intiger followed by unit [0-9]+[MHD], default is no expiration
    --filename=<s>              Filename to use, default is upload-file basename
    --attributes=<p>            JSON formated attribute file, with --chunk option, applied at end of upload
                                Note: Attributes will replace all current attributes
    
    --dir=<p                    Directory for temp files, used with --chunk
    --chunk=<s>                 Upload size of each chunk, able to resume if disconnected
                                Use integer followed by unit: [0-9]+[KMG]
                                Note: this uses parts node, runs mutliple shock uploads,
                                      only used with --filepath option
    
    --compression=<gzip|bzip2>  Uncompress after upload, default is off
                                Note: used with --filepath and --part options
    
    Mutualy exclusive options:
    --filepath=<p>              Path to file
    --part=<s>                  Number of parts to be uploaded
    --virtual=<s>               Comma seperated list of node ids
    --remote=<url>              URI of remote file
    --copy=<id>                 ID of node to copy

update [options...] <id>
    Note: options are the same as create with additional below
    --unexpire                  Remove expiration
    --part=<s> --filepath=<p>   The part number to be uploaded and path to file
                                Note: parts must be set

resume [options...] <id>
    Note: this is for resuming an incomplete upload using: creatut/update --chunk
    --attributes=<p>            JSON formated attribute file, applied at end up upload
    --filepath=<p>              Path to file
    --dir=<p                    Directory for temp files

unpack [options] <id>
    Note: creates mutliple nodes from a parent archive format node
    --attributes=<p>            JSON formated attribute file
                                Note: same attributes applied to all created nodes
    --archive=<tar|tar.gz|tar.bz2|zip>

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
    
    Query field and values, may be used more than once each
    Note: use "FIELD_NAME:VALUE" with no whitespace
    --attribute=<field:value>   Attribute field and value to search with, may be used more than once
    --other=<field:value>       Any non-attributes node field and value to search with, may be used more than once
    
    Note: this option returns a list of fields names and not nodes
    --distinct=<s>              An attribute field name, returns all unique values in that field

download [options...] <id>
    --md5                       Create md5 checksum of downloaded file
                                Note: requires output file
    Mutualy exclusive options:
    --index=<s> --parts={s}      Name of index and part(s) to retrieve, may be a range eg. 1-10
    --seek=<i> --length=<i>     Download bytes from file for given seek and length

acl [options...] <id> <get|add|delete> <all|read|write|delete> <users>
    Note: users are in the form of comma delimited list of user-names or uuids,
          not required for get action

public [options...] <id> <add|delete>

chown [options...] <id> <user>
    Note: user is user-name or uuid 

`

var CV = map[string]map[string]bool{
	"acl":         map[string]bool{"all": true, "delete": true, "read": true, "write": true},
	"archive":     map[string]bool{"tar": true, "tar.gz": true, "tar.bz2": true, "zip": true},
	"compression": map[string]bool{"bzip2": true, "gzip": true},
	"direction":   map[string]bool{"asc": true, "desc": true},
	"index":       map[string]bool{"bai": true, "chunkrecord": true, "column": true, "line": true, "record": true, "size": true},
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
	flags.Var(&attrQuery, "attribute", "")
	flags.StringVar(&bearer, "bearer", "mgrast", "")
	flags.StringVar(&chunk, "chunk", "", "")
	flags.IntVar(&column, "column", 0, "")
	flags.StringVar(&compression, "compression", "", "")
	flags.StringVar(&copy, "copy", "", "")
	flags.BoolVar(&debug, "debug", false, "")
	flags.StringVar(&dir, "dir", ".", "")
	flags.StringVar(&direction, "direction", "", "")
	flags.StringVar(&expiration, "expiration", "", "")
	flags.StringVar(&filename, "filename", "", "")
	flags.StringVar(&filepath, "filepath", "", "")
	flags.BoolVar(&force, "force", false, "")
	flags.StringVar(&index, "index", "", "")
	flags.IntVar(&length, "length", 0, "")
	flags.IntVar(&limit, "limit", 0, "")
	flags.BoolVar(&md5sum, "md5", false, "")
	flags.IntVar(&offset, "offset", 0, "")
	flags.StringVar(&order, "order", "", "")
	flags.Var(&otherQuery, "other", "")
	flags.StringVar(&output, "output", "stdout", "")
	flags.IntVar(&part, "part", 0, "")
	flags.StringVar(&parts, "parts", "", "")
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

func chunkUploadAttr() (attr map[string]interface{}) {
	p, c := partsFromChunk()
	attr["type"] = "temp"
	attr["md5sum"] = md5ForFilePath()
	attr["file_name"] = path.Base(filepath)
	attr["parts_size"] = p
	attr["chunk_size"] = c
	return
}

func partsFromChunk() (calcParts int, chunkBytes int64) {
	fi, err := os.Stat(filepath)
	if err != nil {
		exitError(err.Error())
	}
	matched := chunkRegex.FindStringSubmatch(chunk)
	if len(matched) == 0 {
		exitError("chunk format is invalid")
	}

	chunkBytes, _ = strconv.ParseInt(matched[0], 10, 64)
	switch matched[2] {
	case "K":
		chunkBytes = chunkBytes * 1024
	case "M":
		chunkBytes = chunkBytes * 1024 * 1024
	case "G":
		chunkBytes = chunkBytes * 1024 * 1024 * 1024
	}

	quotient := fi.Size() / chunkBytes
	remainder := fi.Size() % chunkBytes
	if quotient > 100 {
		exitError("too many part uploads created, please specify a larger chunk size")
	}

	if remainder == 0 {
		calcParts = int(quotient)
	} else {
		calcParts = int(quotient + 1)
	}
	return
}

func md5ForFilePath() string {
	f, err := os.Open(filepath)
	if err != nil {
		exitError(err.Error())
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		exitError(err.Error())
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func uploadPartsFiles(nid string, start int, attr map[string]interface{}) (err error) {
	partSize, pok := attr["parts_size"].(int)
	chunkSize, cok := attr["chunk_size"].(int64)
	if !(pok && cok) {
		err = errors.New("invalid node attributes for chunk / resume upload")
		return
	}
	if start >= partSize {
		err = errors.New("invalid parts start and/or total size")
		return
	}
	fmt.Printf("node: %s, %d parts to upload\n", nid, partSize-start)

	var currSize int64
	var fh *os.File
	fh, err = os.Open(filepath)
	if err != nil {
		return
	}
	defer fh.Close()

	for i := start; i <= partSize; i++ {
		tempFile := path.Join(dir, randomStr(12))
		_, err = os.Create(tempFile)
		if err != nil {
			return
		}
		for {
			if currSize >= chunkSize {
				break
			}
			bufferSize := int(math.Min(MaxBuffer, float64(chunkSize-currSize)))
			byteBuffer := make([]byte, bufferSize)
			_, ferr := fh.Read(byteBuffer)
			if ferr == io.EOF {
				break
			}
			ioutil.WriteFile(tempFile, byteBuffer, os.ModeAppend)
			currSize += int64(bufferSize)
		}
		currSize = 0
		// upload file part
		_, err = client.PutOrPostFile(tempFile, nid, i, "", "", nil, nil)
		if err != nil {
			return
		}
		os.Remove(tempFile)
		fmt.Printf("part %d uploaded\n", i)
	}
	return
}

func buildDownloadUrl(host string, id string) string {
	var query url.Values
	query.Add("download", "")

	if (index != "") && (parts != "") {
		if !validateCV("index", index) {
			exitError("invalid index type")
		}
		query.Add("index", index)
		query.Add("part", parts)
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

func isDir(d string) bool {
	fi, err := os.Stat(d)
	if err != nil {
		return false
	}
	return fi.IsDir()
}

func randomStr(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	rand.Read(b)
	s := base64.StdEncoding.EncodeToString(b)
	return s[:n]
}

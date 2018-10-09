package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

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
	//fmt.Fprintln(os.Stderr, USAGE)
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

func getUserInfo() (host string, auth string) {
	// set from env if exists
	if os.Getenv("SHOCK_URL") != "" {
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
	host = shock_url
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
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var CV = map[string]map[string]bool{
	"acl":         {"all": true, "delete": true, "read": true, "write": true},
	"archive":     {"tar": true, "tar.gz": true, "tar.bz2": true, "zip": true},
	"compression": {"bzip2": true, "gzip": true},
	"direction":   {"asc": true, "desc": true},
	"index":       {"bai": true, "chunkrecord": true, "column": true, "line": true, "record": true, "size": true},
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
	fmt.Fprint(os.Stdout, USAGE)
	os.Exit(0)
}

func exitError(msg string) {
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
		e = os.WriteFile(output, b, 0644)
		if e != nil {
			exitError(e.Error())
		}
	}
	os.Exit(0)
}

func getUserInfo() (host string, tkn string, br string) {
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
	host = shock_url
	tkn = token
	br = bearer
	return
}

func buildDownloadUrl(host string, id string) string {
	return host + "/node/" + id + "?download"
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

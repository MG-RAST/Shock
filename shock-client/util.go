package main

import (
	"errors"
	"fmt"
	"github.com/MG-RAST/Shock/shock-client/lib"
	"io"
	"os"
)

type Set map[string]bool

var (
	fileOptionKeys = []string{"full", "parts", "part", "virtual_file", "remote_path"}
)

func (s *Set) Has(key string) bool {
	_, has := (*s)[key]
	return has
}

func ne(s *string) bool {
	if (*s) != "" {
		return true
	}
	return false
}

func fileOptions(options map[string]*string) (t string, err error) {
	c := 0
	for _, i := range fileOptionKeys {
		if options[i] != nil && ne(options[i]) {
			t = i
			if c += 1; c > 1 {
				return "", errors.New("file options are multially exclusive")
			}
		}
	}
	return
}

func downloadChunk(n lib.Node, opts lib.Opts, filename string, offset int64, c chan int) {
	oh, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		fmt.Printf("Error open output file %s: %s\n", filename, err.Error())
	}
	defer oh.Close()

	_, err = oh.Seek(offset, os.SEEK_SET)
	if err != nil {
		fmt.Printf("Error seek output file %s: %s\n", filename, err.Error())
	}

	if ih, err := n.Download(opts); err != nil {
		fmt.Printf("Error downloading %s: %s\n", n.Id, err.Error())
	} else {
		if s, err := io.Copy(oh, ih); err != nil {
			fmt.Printf("Error writing output: %s\n", err.Error())
		} else {
			fmt.Printf("Success. Wrote %d bytes\n", s)
		}
	}
	c <- 1
}

package main

import (
	"errors"
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

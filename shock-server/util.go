package main

import (
	"fmt"
	"path"
	"strings"
)

const logo = `
 +-------------+  +----+    +----+  +--------------+  +--------------+  +----+      +----+
 |             |  |    |    |    |  |              |  |              |  |    |      |    |
 |    +--------+  |    |    |    |  |    +----+    |  |    +---------+  |    |      |    |
 |    |           |    +----+    |  |    |    |    |  |    |            |    |     |    |
 |    +--------+  |              |  |    |    |    |  |    |            |    |    |    |
 |             |  |    +----+    |  |    |    |    |  |    |            |    |   |    |
 +--------+    |  |    |    |    |  |    |    |    |  |    |            |    +---+    +-+
          |    |  |    |    |    |  |    |    |    |  |    |            |               |
 +--------+    |  |    |    |    |  |    +----+    |  |    +---------+  |    +-----+    |
 |             |  |    |    |    |  |              |  |              |  |    |     |    |
 +-------------+  +----+    +----+  +--------------+  +--------------+  +----+     +----+`

func printLogo() {
	fmt.Println(logo)
	return
}

// Path2uuid extract uuid from path
func Path2uuid(filepath string) string {

	ext := path.Ext(filepath)                     // identify extension
	filename := strings.TrimSuffix(filepath, ext) // find filename
	uuid := path.Base(filename)                   // implement basename cmd

	return uuid
}

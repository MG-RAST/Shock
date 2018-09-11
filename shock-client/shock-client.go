package main

import (
	"flag"
	"fmt"
	sc "github.com/MG-RAST/go-shock-client"
	"io"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, " ")
}
func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type queryNode struct {
	values url.Values
	prefix string
	full   bool
}

func newQueryNode() queryNode {
	return queryNode{
		values: url.Values{},
		prefix: "",
		full:   false,
	}
}
func (q queryNode) processFlags(queries arrayFlags) {
	for _, val := range queries {
		parts := strings.Split(val, ":")
		if len(parts) == 2 {
			name := q.prefix + parts[0]
			q.values.Set(name, parts[1])
		}
	}
}
func (q queryNode) addOptions() {
	if limit != 0 {
		q.values.Set("limit", strconv.Itoa(limit))
	}
	if offset != 0 {
		q.values.Set("offset", strconv.Itoa(offset))
	}
	if (direction != "") && validateCV("direction", direction) {
		q.values.Set("direction", direction)
	}
	if order != "" {
		q.values.Set("order", order)
	}
}

const MaxBuffer = 64 * 1024

var chunkRegex = regexp.MustCompile(`^(\d+)(K|M|G)$`)
var client *sc.ShockClient

var (
	flags       *flag.FlagSet
	archive     string
	attributes  string
	attrQuery   arrayFlags
	bearer      string
	chunk       string
	column      int
	compression string
	copy        string
	debug       bool
	dir         string
	direction   string
	expiration  string
	filename    string
	filepath    string
	force       bool
	index       string
	length      int
	limit       int
	md5sum      bool
	offset      int
	order       string
	otherQuery  arrayFlags
	output      string
	part        int
	parts       string
	pretty      bool
	remote      string
	seek        int
	shock_url   string
	token       string
	unexpire    bool
	virtual     string
)

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
	client = &sc.ShockClient{
		Host:  host,
		Token: auth,
		Debug: debug,
	}

	var err error
	switch command {
	case "info":
		var info *sc.ShockResponseMap
		info, err = client.ServerInfo()
		if err == nil {
			exitOutput(&info)
		}
	case "create", "update":
		var nid string
		if len(args) > 0 {
			nid = args[0]
		}
		var opts map[string]string
		if filename != "" {
			opts["file_name"] = filename
		}
		if expiration != "" {
			opts["expiration"] = expiration
		}
		if (command == "update") && unexpire {
			opts["remove_expiration"] = "true"
		}
		if compression != "" {
			if !validateCV("compression", compression) {
				exitError("invalid compress type")
			}
		}
		// do one of these
		if (part > 0) && (filepath == "") {
			// set parts node
			if compression != "" {
				opts["compression"] = compression
			}
			opts["parts"] = strconv.Itoa(part)
			nid, err = client.PutOrPostFile("", nid, 0, attributes, "parts", opts, nil)
		} else if (part > 0) && (filepath != "") && (command == "update") {
			// put part file, update only
			nid, err = client.PutOrPostFile(filepath, nid, part, attributes, "", opts, nil)
		} else if virtual != "" {
			// set virtual file
			opts["virtual_file"] = virtual
			nid, err = client.PutOrPostFile("", nid, 0, attributes, "virtual", opts, nil)
		} else if remote != "" {
			// add file by remote url
			opts["remote_url"] = remote
			nid, err = client.PutOrPostFile("", nid, 0, attributes, "remote", opts, nil)
		} else if copy != "" {
			// copy file from another node
			opts["parent_node"] = copy
			opts["copy_indexes"] = "true"
			if attributes == "" {
				opts["copy_attributes"] = "true"
			}
			nid, err = client.PutOrPostFile("", nid, 0, attributes, "copy", opts, nil)
		} else if (chunk != "") && (filepath != "") {
			if dir != "." && (!isDir(dir)) {
				exitError("invalid --dir path")
			}
			// special auto-parts upload, able to resume
			if compression != "" {
				opts["compression"] = compression
			}
			// first set parts and temp attributes, calculate # of parts
			tempAttr := chunkUploadAttr()
			partSize := tempAttr["parts_size"].(int)
			opts["parts"] = strconv.Itoa(partSize)
			nid, err = client.PutOrPostFile("", nid, 0, "", "parts", opts, tempAttr)
			if err != nil {
				exitError(err.Error())
			}
			// upload parts in series
			err = uploadPartsFiles(nid, 1, tempAttr)
			if err != nil {
				exitError("error uploading " + filepath + " in chunks")
			}
			if attributes == "" {
				var attrMap map[string]interface{}
				client.UpdateAttributes(nid, "", attrMap)
			} else {
				client.UpdateAttributes(nid, attributes, nil)
			}
		} else if filepath != "" {
			// basic file upload
			if compression != "" {
				opts["compression"] = compression
			}
			nid, err = client.PutOrPostFile(filepath, nid, 0, attributes, "", opts, nil)
		} else {
			exitError("invalid option combination")
		}
		if err == nil {
			fmt.Printf("%sd node: %s\n", command, nid)
		}
	case "resume":
		if len(args) < 1 {
			exitError("missing required ID")
		}
		if filepath == "" {
			exitError("missing required --filepath")
		}
		var node *sc.ShockNode
		var attr map[string]interface{}
		node, err = client.GetNode(args[0])
		if err != nil {
			exitError(err.Error())
		}
		if (node.Type != "parts") || (node.Parts == nil) {
			exitError("node " + args[0] + " is not a valid parts node")
		}
		if node.Parts.Count == node.Parts.Length {
			exitError("node " + args[0] + " has already completed upload")
		}
		if node.Parts.Count != attr["parts_size"].(int) {
			exitError("invalid parts node: node.attributes.parts_size != node.parts.count")
		}
		attr = node.Attributes.(map[string]interface{})
		nodeMD5, ok := attr["md5sum"]
		fileMD5 := md5ForFilePath()
		if !ok || (nodeMD5 != fileMD5) {
			exitError(fmt.Sprintf("checksum of %s does not match origional file started on node %s", filepath, args[0]))
		}
		err = uploadPartsFiles(args[0], node.Parts.Length+1, attr)
		if err != nil {
			exitError("error uploading " + filepath + " in chunks")
		}
		if attributes == "" {
			var attrMap map[string]interface{}
			client.UpdateAttributes(args[0], "", attrMap)
		} else {
			client.UpdateAttributes(args[0], attributes, nil)
		}
	case "unpack":
		if len(args) < 1 {
			exitError("missing required ID")
		}
		if !validateCV("archive", archive) {
			exitError("invalid archive type")
		}
		var nodes interface{}
		nodes, err = client.UnpackArchiveNode(args[0], archive, attributes)
		if err == nil {
			exitOutput(&nodes)
		}
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
		err = client.PutIndexQuery(args[0], args[1], force, column)
	case "delete":
		if len(args) < 1 {
			exitError("missing required ID")
		}
		err = client.DeleteNode(args[0])
	case "get":
		if len(args) < 1 {
			exitError("missing required ID")
		}
		var node *sc.ShockNode
		node, err = client.GetNode(args[0])
		if err == nil {
			exitOutput(&node)
		}
	case "query":
		query := newQueryNode()
		if len(otherQuery) > 0 {
			query.processFlags(otherQuery)
			query.prefix = "attributes."
			query.full = true
		}
		if len(attrQuery) > 0 {
			query.processFlags(attrQuery)
		}
		query.addOptions()
		var sqr *sc.ShockQueryResponse
		if query.full {
			sqr, err = client.QueryFull(query.values)
		} else {
			sqr, err = client.Query(query.values)
		}
		if err == nil {
			exitOutput(&sqr)
		}
	case "download":
		if len(args) < 1 {
			exitError("missing required ID")
		}
		downUrl := buildDownloadUrl(host, args[0])
		if (output == "") || (output == "-") || (output == "stdout") {
			body, berr := sc.FetchShockStream(downUrl, auth)
			if berr != nil {
				exitError(berr.Error())
			}
			defer body.Close()
			_, err = io.Copy(os.Stdout, body)
		} else {
			var size int64
			var checksum string
			size, checksum, err = sc.FetchFile(output, downUrl, auth, "", md5sum)
			if err == nil {
				fmt.Printf("download complete\nfile\t%s\nsize\t%d\nmd5\t%s\n", output, size, checksum)
			}
		}
	case "acl":
		if (len(args) > 1) && (args[1] == "get") {
			var acl *sc.ShockResponseGeneric
			acl, err = client.GetAcl(args[0])
			if err == nil {
				exitOutput(&acl.Data)
			}
		} else if len(args) > 3 {
			if !validateCV("acl", args[2]) {
				exitError("invalid acl type")
			}
			if args[1] == "add" {
				err = client.PutAcl(args[0], args[2], args[3])
			}
			if args[1] == "delete" {
				err = client.DeleteAcl(args[0], args[2], args[3])
			}
		} else {
			exitError("missing required arguments")
		}
	case "public":
		if len(args) < 2 {
			exitError("missing required arguments")
		}
		if args[1] == "add" {
			err = client.MakePublic(args[0])
		}
		if args[1] == "delete" {
			err = client.DeleteAcl(args[0], "public_read", "")
		}
	case "chown":
		if len(args) < 2 {
			exitError("missing required arguments")
		}
		err = client.ChownNode(args[0], args[1])
	default:
		exitError("invalid command: " + command)
	}

	if err != nil {
		exitError(err.Error())
	}
	os.Exit(0)
}

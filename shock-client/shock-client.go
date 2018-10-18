package main

import (
	"fmt"
	sc "github.com/MG-RAST/go-shock-client"
	"io"
	"os"
	"strconv"
)

var client *sc.ShockClient

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
	client = sc.NewShockClient(host, auth, debug)

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
		opts := make(map[string]string)
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
			// special auto-parts upload, able to resume
			if dir != "." && (!isDir(dir)) {
				exitError("invalid --dir path")
			}
			// first set parts and temp attributes, calculate # of parts
			cu := newChunkUploader(filepath, chunk)
			opts["parts"] = strconv.Itoa(cu.parts)
			if compression != "" {
				opts["compression"] = compression
			}
			tempAttr := cu.getAttr()
			nid, err = client.PutOrPostFile("", nid, 0, "", "parts", opts, tempAttr)
			if err != nil {
				exitError(err.Error())
			}
			// upload parts in series
			err = cu.uploadParts(nid, 1, dir)
			if err != nil {
				exitError("error uploading " + filepath + " in chunks: " + err.Error())
			}
			// final attributes
			if attributes == "" {
				attrMap := make(map[string]interface{})
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
		if dir != "." && (!isDir(dir)) {
			exitError("invalid --dir path")
		}
		if filepath == "" {
			exitError("missing required --filepath")
		}
		var node *sc.ShockNode
		node, err = client.GetNode(args[0])
		if err != nil {
			exitError(err.Error())
		}
		// validate node and get info
		cu := newChunkUploader(filepath, "")
		errMsg := cu.validateChunkNode(node)
		if errMsg != "" {
			exitError(errMsg)
		}
		// upload remaining parts in series
		err = cu.uploadParts(args[0], node.Parts.Length+1, dir)
		if err != nil {
			exitError("error uploading " + filepath + " in chunks: " + err.Error())
		}
		// final attributes
		if attributes == "" {
			attrMap := make(map[string]interface{})
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

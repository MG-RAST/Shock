package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"

	sc "github.com/MG-RAST/Shock/clients/shock-go"
)

var client *sc.Client

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

	host, tkn, br := getUserInfo()
	client = sc.New(host, sc.WithToken(tkn), sc.WithAuthType(br), sc.WithDebug(debug))
	ctx := context.Background()

	var err error
	switch command {
	case "info":
		var info *sc.ServerInfo
		info, err = client.ServerInfo(ctx)
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
			nid, err = client.PutOrPostFile(ctx, "", nid, 0, attributes, "parts", opts, nil)
		} else if (part > 0) && (filepath != "") && (command == "update") {
			// put part file, update only
			nid, err = client.PutOrPostFile(ctx, filepath, nid, part, attributes, "", opts, nil)
		} else if virtual != "" {
			// set virtual file
			opts["virtual_file"] = virtual
			nid, err = client.PutOrPostFile(ctx, "", nid, 0, attributes, "virtual", opts, nil)
		} else if remote != "" {
			// add file by remote url
			opts["remote_url"] = remote
			nid, err = client.PutOrPostFile(ctx, "", nid, 0, attributes, "remote", opts, nil)
		} else if copy != "" {
			// copy file from another node
			opts["parent_node"] = copy
			opts["copy_indexes"] = "true"
			if attributes == "" {
				opts["copy_attributes"] = "true"
			}
			nid, err = client.PutOrPostFile(ctx, "", nid, 0, attributes, "copy", opts, nil)
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
			nid, err = client.PutOrPostFile(ctx, "", nid, 0, "", "parts", opts, tempAttr)
			if err != nil {
				exitError(err.Error())
			}
			// upload parts in series
			err = cu.uploadParts(ctx, nid, 1, dir)
			if err != nil {
				exitError("error uploading " + filepath + " in chunks: " + err.Error())
			}
			// final attributes
			if attributes == "" {
				attrMap := make(map[string]interface{})
				client.UpdateAttributes(ctx, nid, "", attrMap)
			} else {
				client.UpdateAttributes(ctx, nid, attributes, nil)
			}
		} else if filepath != "" {
			// basic file upload
			if compression != "" {
				opts["compression"] = compression
			}
			nid, err = client.PutOrPostFile(ctx, filepath, nid, 0, attributes, "", opts, nil)
		} else {
			// just make empty node with provided options
			nid, err = client.PutOrPostFile(ctx, "", nid, 0, attributes, "", opts, nil)
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
		var node *sc.Node
		node, err = client.GetNode(ctx, args[0])
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
		err = cu.uploadParts(ctx, args[0], node.Parts.Length+1, dir)
		if err != nil {
			exitError("error uploading " + filepath + " in chunks: " + err.Error())
		}
		// final attributes
		if attributes == "" {
			attrMap := make(map[string]interface{})
			client.UpdateAttributes(ctx, args[0], "", attrMap)
		} else {
			client.UpdateAttributes(ctx, args[0], attributes, nil)
		}
	case "unpack":
		if len(args) < 1 {
			exitError("missing required ID")
		}
		if !validateCV("archive", archive) {
			exitError("invalid archive type")
		}
		var nodes []sc.Node
		nodes, err = client.UnpackArchiveNode(ctx, args[0], archive, attributes)
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
		err = client.PutIndexQuery(ctx, args[0], args[1], force, column)
	case "delete":
		if len(args) < 1 {
			exitError("missing required ID")
		}
		err = client.DeleteNode(ctx, args[0])
	case "get":
		if len(args) < 1 {
			exitError("missing required ID")
		}
		var node *sc.Node
		node, err = client.GetNode(ctx, args[0])
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
		if query.distinct {
			if query.full {
				query.values.Set("querynode", "")
			} else {
				query.values.Set("query", "")
			}
			var result interface{}
			result, err = client.QueryDistinctRaw(ctx, query.values)
			if err == nil {
				exitOutput(&result)
			}
		} else {
			if query.full {
				query.values.Set("querynode", "")
			} else {
				query.values.Set("query", "")
			}
			var sqr *sc.QueryResult
			sqr, err = client.QueryRaw(ctx, query.values)
			if err == nil {
				exitOutput(&sqr)
			}
		}
	case "download":
		if len(args) < 1 {
			exitError("missing required ID")
		}
		if (output == "") || (output == "-") || (output == "stdout") {
			body, berr := client.Download(ctx, args[0])
			if berr != nil {
				exitError(berr.Error())
			}
			defer body.Close()
			_, err = io.Copy(os.Stdout, body)
		} else {
			var dlOpts []sc.DownloadOption
			if md5sum {
				dlOpts = append(dlOpts, sc.WithComputeMD5())
			}
			var size int64
			var checksum string
			size, checksum, err = client.DownloadToFile(ctx, args[0], output, dlOpts...)
			if err == nil {
				fmt.Printf("download complete\nfile\t%s\nsize\t%d\nmd5\t%s\n", output, size, checksum)
			}
		}
	case "acl":
		if (len(args) > 1) && (args[1] == "get") {
			var acl *sc.DisplayAcl
			acl, err = client.GetAcl(ctx, args[0])
			if err == nil {
				exitOutput(&acl)
			}
		} else if len(args) > 3 {
			if !validateCV("acl", args[2]) {
				exitError("invalid acl type")
			}
			if args[1] == "add" {
				_, err = client.PutAcl(ctx, args[0], args[2], args[3])
			}
			if args[1] == "delete" {
				_, err = client.DeleteAcl(ctx, args[0], args[2], args[3])
			}
		} else {
			exitError("missing required arguments")
		}
	case "public":
		if len(args) < 2 {
			exitError("missing required arguments")
		}
		if args[1] == "add" {
			_, err = client.MakePublic(ctx, args[0])
		}
		if args[1] == "delete" {
			_, err = client.DeleteAcl(ctx, args[0], "public_read", "")
		}
	case "chown":
		if len(args) < 2 {
			exitError("missing required arguments")
		}
		_, err = client.ChownNode(ctx, args[0], args[1])
	default:
		exitError("invalid command: " + command)
	}

	if err != nil {
		exitError(err.Error())
	}
	os.Exit(0)
}

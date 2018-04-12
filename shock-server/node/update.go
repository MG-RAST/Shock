package node

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/node/archive"
	"github.com/MG-RAST/Shock/shock-server/util"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

//Modification functions
func (node *Node) Update(params map[string]string, files FormFiles) (err error) {
	// Exclusive conditions
	// 1.1. has files[upload] (regular upload)
	// 1.2. has files[gzip] (compressed upload)
	// 1.3. has files[bzip2] (compressed upload)
	// 2. has params[parts] (partial upload support)
	// 3. has params[type] & params[source] (v_node)
	// 4. has params[path] (set from local path)
	// 5. has params[copy_data] (create node by copying data from another node)
	// 6. has params[parent_node] (create node by specifying subset of records in a parent node)
	//
	// All condition allow setting of attributes
	//
	// state is saved to mongodb at end of update function

	// global lock on a node that is being updated
	err = LockMgr.LockNode(node.Id)
	if err != nil {
		return
	}
	defer LockMgr.UnlockNode(node.Id)

	// refresh node state
	var n *Node
	n, err = Load(node.Id)
	if err != nil {
		return
	}
	node = n

	for _, u := range util.ValidUpload {
		if _, uploadMisplaced := params[u]; uploadMisplaced {
			return errors.New(fmt.Sprintf("%s form field must be file encoded", u))
		}
	}

	isRegularUpload := false
	uploadFile := ""
	uploadCount := 0
	for _, u := range util.ValidUpload {
		if _, hasRegularUpload := files[u]; hasRegularUpload {
			isRegularUpload = true
			uploadFile = u
			uploadCount += 1
		}
	}
	if uploadCount > 1 {
		return errors.New("only one upload file allowed")
	}

	isUrlUpload := false
	if _, hasUrlUpload := files["upload_url"]; hasUrlUpload {
		isUrlUpload = true
	}

	_, isPartialUpload := params["parts"]
	hasPartsFile := false
	for key, _ := range files {
		if _, errf := strconv.Atoi(key); errf == nil {
			hasPartsFile = true
		}
	}

	isVirtualNode := false
	if t, hasType := params["type"]; hasType && t == "virtual" {
		isVirtualNode = true
	}
	_, isPathUpload := params["path"]
	_, isCopyUpload := params["copy_data"]
	_, isSubsetUpload := params["parent_node"]

	// Check exclusive conditions
	if isRegularUpload && (isUrlUpload || isPartialUpload || isPathUpload || isVirtualNode || isCopyUpload || isSubsetUpload) {
		return errors.New("upload parameter incompatible with upload_url, parts, path, type, copy_data and/or parent_node parameter(s)")
	} else if isUrlUpload && (isRegularUpload || isPartialUpload || isPathUpload || isVirtualNode || isCopyUpload || isSubsetUpload) {
		return errors.New("upload_url parameter incompatible with upload, parts, path, type, copy_data and/or parent_node parameter(s)")
	} else if isPartialUpload && (isVirtualNode || isPathUpload || isCopyUpload || isSubsetUpload) {
		return errors.New("parts parameter incompatible with type, path, copy_data and/or parent_node parameter(s)")
	} else if isVirtualNode && (isPathUpload || isCopyUpload || isSubsetUpload) {
		return errors.New("type parameter incompatible with path, copy_data and/or parent_node parameter")
	} else if isPathUpload && (isCopyUpload || isSubsetUpload) {
		return errors.New("path parameter incompatible with copy_data and/or parent_node parameter")
	} else if isCopyUpload && isSubsetUpload {
		return errors.New("copy_data parameter incompatible with parent_node parameter")
	} else if hasPartsFile && (isRegularUpload || isUrlUpload) {
		return errors.New("parts file and upload or upload_url parameters are incompatible")
	} else if (node.Type == "parts") && (isRegularUpload || isUrlUpload) {
		return errors.New("parts node and upload or upload_url parameters are incompatible")
	} else if isPartialUpload && hasPartsFile {
		return errors.New("can not upload parts file when creating parts node")
	}

	// Check if immutable
	if node.HasFile() && (isRegularUpload || isUrlUpload || isPartialUpload || hasPartsFile || isVirtualNode || isPathUpload || isCopyUpload || isSubsetUpload) {
		return errors.New(e.FileImut)
	}

	if isRegularUpload {
		if err = node.SetFile(files[uploadFile]); err != nil {
			return err
		}
		delete(files, uploadFile)
	} else if isUrlUpload {
		if err = node.SetFile(files["upload_url"]); err != nil {
			return err
		}
		delete(files, "upload_url")
	} else if isPartialUpload {
		// close variable length parts
		if params["parts"] == "close" {
			if (node.Type != "parts") || (node.Parts == nil) || !node.Parts.VarLen {
				return errors.New("can only call 'close' on unknown parts node")
			}
			if err = node.closeParts(true); err != nil {
				return err
			}
		} else if (node.Parts != nil) && (node.Parts.VarLen || node.Parts.Count > 0) {
			return errors.New("parts already set")
		} else {
			// set parts struct
			var compressionFormat string = ""
			if compress, ok := params["compression"]; ok {
				if archive.IsValidUncompress(compress) {
					compressionFormat = compress
				}
			}
			if params["parts"] == "unknown" {
				if err = node.initParts("unknown", compressionFormat); err != nil {
					return err
				}
			} else {
				n, err := strconv.Atoi(params["parts"])
				if err != nil {
					return errors.New("parts must be an integer or 'unknown'")
				}
				if n < 1 {
					return errors.New("parts cannot be less than 1")
				}
				if err = node.initParts(params["parts"], compressionFormat); err != nil {
					return err
				}
			}
		}
	} else if isVirtualNode {
		node.Type = "virtual"
		if source, hasSource := params["source"]; hasSource {
			ids := strings.Split(source, ",")
			node.addVirtualParts(ids)
		} else {
			return errors.New("type virtual requires source parameter")
		}
	} else if isPathUpload {
		if action, hasAction := params["action"]; !hasAction || (action != "copy_file" && action != "move_file" && action != "keep_file") {
			return errors.New("path upload requires action field equal to copy_file, move_file or keep_file")
		}
		localpaths := strings.Split(conf.PATH_LOCAL, ",")
		if len(localpaths) <= 0 {
			return errors.New("local files path uploads must be configured. Please contact your Shock administrator.")
		}
		var success = false
		for _, p := range localpaths {
			if strings.HasPrefix(params["path"], p) {
				if err = node.SetFileFromPath(params["path"], params["action"]); err != nil {
					return err
				} else {
					success = true
				}
			}
		}
		if !success {
			return errors.New("file not in local files path. Please contact your Shock administrator.")
		}
	} else if isCopyUpload {
		var n *Node
		n, err = Load(params["copy_data"])
		if err != nil {
			return err
		}

		if n.File.Virtual {
			return errors.New("copy_data parameter points to a virtual node, invalid operation.")
		}

		// Copy node file information
		node.File.Name = n.File.Name
		node.File.Size = n.File.Size
		node.File.Checksum = n.File.Checksum
		node.File.Format = n.File.Format
		node.File.CreatedOn = time.Now()

		if n.Type == "subset" {
			node.Subset = n.Subset
			subsetIndexFile := n.Path() + "/" + n.Id + ".subset.idx"
			// The subset index file is required for creating a copy of a subset node.
			if _, err := os.Stat(subsetIndexFile); err == nil {
				if _, cerr := util.CopyFile(subsetIndexFile, node.Path()+"/"+node.Id+".subset.idx"); cerr != nil {
					return cerr
				}
			} else {
				return err
			}
			node.Type = "subset"
		} else {
			node.Type = "copy"
		}

		// Copy node attributes
		if _, copyAttributes := params["copy_attributes"]; copyAttributes {
			node.Attributes = n.Attributes
		}

		// Copy node indexes
		if _, copyIndex := params["copy_indexes"]; copyIndex && (len(n.Indexes) > 0) {
			// loop through parent indexes
			for idxType, idxInfo := range n.Indexes {
				parentFile := n.IndexPath() + "/" + idxType + ".idx"
				if _, err := os.Stat(parentFile); err == nil {
					// copy file if exists
					if _, cerr := util.CopyFile(parentFile, node.IndexPath()+"/"+idxType+".idx"); cerr != nil {
						return cerr
					}
				}
				// copy index struct
				node.SetIndexInfo(idxType, idxInfo)
			}
		} else if sizeIndex, exists := n.Indexes["size"]; exists {
			// just copy size index
			node.SetIndexInfo("size", sizeIndex)
		}

		if n.File.Path == "" {
			node.File.Path = fmt.Sprintf("%s/%s.data", getPath(params["copy_data"]), params["copy_data"])
		} else {
			node.File.Path = n.File.Path
		}
	} else if isSubsetUpload {
		fInfo, statErr := os.Stat(files["subset_indices"].Path)
		if statErr != nil {
			return errors.New("Could not stat uploaded subset_indices file.")
		}
		node.Type = "subset"

		if fInfo.Size() == 0 {
			// if upload file is empty, make a basic node with empty file
			if err = node.SetFile(files["subset_indices"]); err != nil {
				return err
			}
			delete(files, "subset_indices")
		} else {
			// process subset upload
			_, hasParentIndex := params["parent_index"]
			if !hasParentIndex {
				return errors.New("parent_index is a required parameter for creating a subset node.")
			}

			var n *Node
			n, err = Load(params["parent_node"])
			if err != nil {
				return err
			}

			if n.File.Virtual {
				return errors.New("parent_node parameter points to a virtual node, invalid operation.")
			}

			if _, indexExists := n.Indexes[params["parent_index"]]; !indexExists {
				return errors.New("Index '" + params["parent_index"] + "' does not exist for parent node.")
			}

			parentIndexFile := n.IndexPath() + "/" + params["parent_index"] + ".idx"
			if _, statErr := os.Stat(parentIndexFile); statErr != nil {
				return errors.New("Could not stat index file for parent node where parent node = '" + params["parent_node"] + "' and index = '" + params["parent_index"] + "'.")
			}

			// Copy node file information
			node.File.Name = n.File.Name
			node.File.Format = n.File.Format
			node.Subset.Parent.Id = params["parent_node"]
			node.Subset.Parent.IndexName = params["parent_index"]

			if n.File.Path == "" {
				node.File.Path = fmt.Sprintf("%s/%s.data", getPath(params["parent_node"]), params["parent_node"])
			} else {
				node.File.Path = n.File.Path
			}

			if _, hasSubsetList := files["subset_indices"]; hasSubsetList {
				if err = node.SetFileFromSubset(files["subset_indices"]); err != nil {
					return err
				}
				delete(files, "subset_indices")
			}
		}
	}

	// set attributes from file
	if _, hasAttr := files["attributes"]; hasAttr {
		if _, hasAttrStr := params["attributes_str"]; hasAttrStr {
			return errors.New("Cannot define an attributes file and an attributes_str parameter in the same request.")
		}
		if err = node.SetAttributes(files["attributes"]); err != nil {
			return err
		}
		delete(files, "attributes")
	}

	// set attributes from json string
	if _, hasAttrStr := params["attributes_str"]; hasAttrStr {
		if _, hasAttr := files["attributes"]; hasAttr {
			return errors.New("Cannot define an attributes file and an attributes_str parameter in the same request.")
		}
		if err = node.SetAttributesFromString(params["attributes_str"]); err != nil {
			return err
		}
	}

	// set filename string
	if _, hasFileNameStr := params["file_name"]; hasFileNameStr {
		node.File.Name = params["file_name"]
	}

	// update relatives
	if _, hasRelation := params["linkage"]; hasRelation {
		ltype := params["linkage"]

		if ltype == "parent" {
			if node.HasParent() {
				return errors.New(e.ProvenanceImut)
			}
		}
		var ids string
		if _, hasIds := params["ids"]; hasIds {
			ids = params["ids"]
		} else {
			return errors.New("missing ids for updating relatives")
		}
		var operation string
		if _, hasOp := params["operation"]; hasOp {
			operation = params["operation"]
		}
		node.UpdateLinkages(ltype, ids, operation)
	}

	// update node tags
	if _, hasDataType := params["tags"]; hasDataType {
		node.UpdateDataTags(params["tags"])
	}

	// update file format
	if _, hasFormat := params["format"]; hasFormat {
		if node.File.Format != "" {
			return errors.New(fmt.Sprintf("file format already set:%s", node.File.Format))
		}
		node.File.Format = params["format"]
	}

	// update priority
	if _, hasPriority := params["priority"]; hasPriority {
		priority, err := strconv.Atoi(params["priority"])
		if err != nil {
			return errors.New("priority must be an integer")
		}
		node.Priority = priority
	}

	// update node expiration
	if _, hasExpiration := params["expiration"]; hasExpiration {
		if err = node.SetExpiration(params["expiration"]); err != nil {
			return err
		}
	}
	if _, hasRemove := params["remove_expiration"]; hasRemove {
		// reset to empty time
		node.Expiration = time.Time{}
	}

	// clear node revisions
	if _, hasClearRevisions := params["clear_revisions"]; hasClearRevisions {
		// empty the revisions array
		node.Revisions = []Node{}
	}

	// handle part file
	if hasPartsFile {
		if node.HasFile() {
			return errors.New(e.FileImut)
		}
		if (node.Type != "parts") || (node.Parts == nil) {
			return errors.New("This is not a parts node and thus does not support uploading in parts.")
		}

		if node.Parts.Count > 0 || node.Parts.VarLen {
			for key, file := range files {
				keyn, errf := strconv.Atoi(key)
				if errf == nil && (keyn <= node.Parts.Count || node.Parts.VarLen) {
					if err = node.addPart(keyn-1, &file); err != nil {
						return err
					}
				}
			}
		} else {
			return errors.New("Unable to retrieve parts info for node.")
		}
		// all parts are in, close it
		if !node.Parts.VarLen && node.Parts.Length == node.Parts.Count {
			if err = node.closeParts(false); err != nil {
				return err
			}
		}
	}

	// save only after all updates applied
	err = node.Save()
	return
}

func (node *Node) Save() (err error) {
	// update versions
	previousVersion := node.Version
	node.UpdateVersion()
	// only add to revisions if not new and has changed and allow revisions
	if (previousVersion != "") && (previousVersion != node.Version) && (conf.MAX_REVISIONS != 0) {
		n := Node{node.Id, node.Version, node.File, node.Attributes, node.Indexes, node.Acl, node.VersionParts, node.Tags, nil, node.Linkages, node.Priority, node.CreatedOn, node.LastModified, node.Expiration, node.Type, node.Subset, node.Parts}
		newRevisions := []Node{n}
		if len(node.Revisions) > 0 {
			newRevisions = append(newRevisions, node.Revisions...) // prepend, latest revisions in front
		}
		// adjust revisions based on config
		// <0 keep all ; >0 keep max
		if (conf.MAX_REVISIONS < 0) || (len(newRevisions) <= conf.MAX_REVISIONS) {
			node.Revisions = newRevisions
		} else {
			node.Revisions = newRevisions[:conf.MAX_REVISIONS] // keep most recent MAX_REVISIONS
		}
	}
	if node.CreatedOn.String() == "0001-01-01 00:00:00 +0000 UTC" {
		node.CreatedOn = time.Now()
	} else {
		node.LastModified = time.Now()
	}
	// get bson, test size and print
	nbson, err := bson.Marshal(node)
	if err != nil {
		return err
	}
	if len(nbson) >= DocumentMaxByte {
		return errors.New(fmt.Sprintf("bson document size is greater than limit of %d bytes", DocumentMaxByte))
	}
	bsonPath := fmt.Sprintf("%s/%s.bson", node.Path(), node.Id)
	os.Remove(bsonPath)
	if err := ioutil.WriteFile(bsonPath, nbson, 0644); err != nil {
		// dir path may be missing, recreate and try again
		if err := node.Mkdir(); err != nil {
			return err
		}
		if err := ioutil.WriteFile(bsonPath, nbson, 0644); err != nil {
			return err
		}
	}
	// save node to mongodb
	if err := dbUpsert(node); err != nil {
		return err
	}
	return
}

func (node *Node) UpdateVersion() (err error) {
	h := md5.New()
	version := node.Id
	versionParts := make(map[string]string)
	partMap := map[string]interface{}{"file_ver": node.File, "indexes_ver": node.Indexes, "attributes_ver": node.Attributes, "acl_ver": node.Acl}

	// need to keep map ordered
	partKeys := []string{}
	for k, _ := range partMap {
		partKeys = append(partKeys, k)
	}
	sort.Strings(partKeys)

	for _, k := range partKeys {
		j, er := json.Marshal(partMap[k])
		if er != nil {
			return
		}
		// need to sort bytes to deal with unordered json
		sj := SortByteArray(j)
		h.Write(sj)
		sum := fmt.Sprintf("%x", h.Sum(nil))
		versionParts[k] = sum
		version = version + sum
		h.Reset()
	}
	h.Write([]byte(version))
	node.Version = fmt.Sprintf("%x", h.Sum(nil))
	node.VersionParts = versionParts
	return
}

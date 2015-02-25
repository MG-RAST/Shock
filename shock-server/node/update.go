package node

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/util"
	"github.com/MG-RAST/golib/mgo/bson"
	"io/ioutil"
	"os"
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
	// Note that all paths for node operations in this function must end with "err = node.Save()" to save node state.

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
	if (isRegularUpload && isPartialUpload) || (isRegularUpload && isVirtualNode) || (isRegularUpload && isPathUpload) || (isRegularUpload && isCopyUpload) || (isRegularUpload && isSubsetUpload) {
		return errors.New("upload parameter incompatible with parts, path, type, copy_data and/or parent_node parameter(s)")
	} else if (isPartialUpload && isVirtualNode) || (isPartialUpload && isPathUpload) || (isPartialUpload && isCopyUpload) || (isPartialUpload && isSubsetUpload) {
		return errors.New("parts parameter incompatible with type, path, copy_data and/or parent_node parameter(s)")
	} else if (isVirtualNode && isPathUpload) || (isVirtualNode && isCopyUpload) || (isVirtualNode && isSubsetUpload) {
		return errors.New("type parameter incompatible with path, copy_data and/or parent_node parameter")
	} else if (isPathUpload && isCopyUpload) || (isPathUpload && isSubsetUpload) {
		return errors.New("path parameter incompatible with copy_data and/or parent_node parameter")
	} else if isCopyUpload && isSubsetUpload {
		return errors.New("copy_data parameter incompatible with parent_node parameter")
	} else if isRegularUpload && hasPartsFile {
		return errors.New("upload file and parts file are incompatible")
	}

	// Check if immutable
	if (isRegularUpload || isPartialUpload || hasPartsFile || isVirtualNode || isPathUpload || isCopyUpload || isSubsetUpload) && node.HasFile() {
		return errors.New(e.FileImut)
	}

	if isRegularUpload {
		if err = node.SetFile(files[uploadFile]); err != nil {
			return err
		}
		delete(files, uploadFile)
	} else if isPartialUpload {
		node.Type = "parts"
		if params["parts"] == "unknown" {
			if err = node.initParts("unknown"); err != nil {
				return err
			}
		} else if params["parts"] == "close" {
			if err = node.closeVarLenPartial(); err != nil {
				return err
			}
		} else if node.isVarLen() || node.partsCount() > 0 {
			return errors.New("parts already set")
		} else {
			n, err := strconv.Atoi(params["parts"])
			if err != nil {
				return errors.New("parts must be an integer or 'unknown'")
			}
			if n < 1 {
				return errors.New("parts cannot be less than 1")
			}
			if err = node.initParts(params["parts"]); err != nil {
				return err
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
				if err := node.SetIndexInfo(idxType, idxInfo); err != nil {
					return err
				}
			}
		} else if sizeIndex, exists := n.Indexes["size"]; exists {
			// just copy size index
			if err := node.SetIndexInfo("size", sizeIndex); err != nil {
				return err
			}
		}

		if n.File.Path == "" {
			node.File.Path = fmt.Sprintf("%s/%s.data", getPath(params["copy_data"]), params["copy_data"])
		} else {
			node.File.Path = n.File.Path
		}

		err = node.Save()
		if err != nil {
			return
		}
	} else if isSubsetUpload {
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
		node.Type = "subset"
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
		} else {
			err = node.Save()
			if err != nil {
				return
			}
		}
	}

	if _, hasSubsetList := files["subset_indices"]; hasSubsetList {
		if err = node.SetFileFromSubset(files["subset_indices"]); err != nil {
			return err
		}
		delete(files, "subset_indices")
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
		delete(params, "attributes_str")
	}

	// set filename string
	if _, hasFileNameStr := params["file_name"]; hasFileNameStr {
		node.File.Name = params["file_name"]
		if err = node.Save(); err != nil {
			return err
		}
		delete(params, "file_name")
	}

	// handle part file
	if hasPartsFile {
		if node.HasFile() {
			return errors.New(e.FileImut)
		}
		if node.Type != "parts" {
			return errors.New("This is not a parts node and thus does not support uploading in parts.")
		}
		LockMgr.LockPartOp()
		parts_count := node.partsCount()
		if parts_count > 0 || node.isVarLen() {
			for key, file := range files {
				keyn, errf := strconv.Atoi(key)
				if errf == nil && (keyn <= parts_count || node.isVarLen()) {
					err = node.addPart(keyn-1, &file)
					if err != nil {
						LockMgr.UnlockPartOp()
						return err
					}
				}
			}
		} else {
			LockMgr.UnlockPartOp()
			return errors.New("Unable to retrieve parts info for node.")
		}
		LockMgr.UnlockPartOp()
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
		if err = node.UpdateLinkages(ltype, ids, operation); err != nil {
			return err
		}
	}

	//update node tags
	if _, hasDataType := params["tags"]; hasDataType {
		if err = node.UpdateDataTags(params["tags"]); err != nil {
			return err
		}
	}

	//update file format
	if _, hasFormat := params["format"]; hasFormat {
		if node.File.Format != "" {
			return errors.New(fmt.Sprintf("file format already set:%s", node.File.Format))
		}
		if err = node.SetFileFormat(params["format"]); err != nil {
			return err
		}
	}
	return
}

func (node *Node) Save() (err error) {
	node.UpdateVersion()
	if len(node.Revisions) == 0 || node.Revisions[len(node.Revisions)-1].Version != node.Version {
		n := Node{node.Id, node.Version, node.File, node.Attributes, node.Indexes, node.Acl, node.VersionParts, node.Tags, nil, node.Linkages, node.CreatedOn, node.LastModified, node.Type, node.Subset}
		node.Revisions = append(node.Revisions, n)
	}
	if node.CreatedOn.String() == "0001-01-01 00:00:00 +0000 UTC" {
		node.CreatedOn = time.Now()
	} else {
		node.LastModified = time.Now()
	}

	bsonPath := fmt.Sprintf("%s/%s.bson", node.Path(), node.Id)
	os.Remove(bsonPath)
	nbson, err := bson.Marshal(node)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(bsonPath, nbson, 0644)
	if err != nil {
		return
	}
	err = dbUpsert(node)
	if err != nil {
		return
	}
	return
}

func (node *Node) UpdateVersion() (err error) {
	parts := make(map[string]string)
	h := md5.New()
	version := node.Id
	for name, value := range map[string]interface{}{"file_ver": node.File, "attributes_ver": node.Attributes, "acl_ver": node.Acl} {
		m, er := json.Marshal(value)
		if er != nil {
			return
		}
		h.Write(m)
		sum := fmt.Sprintf("%x", h.Sum(nil))
		parts[name] = sum
		version = version + ":" + sum
		h.Reset()
	}
	h.Write([]byte(version))
	node.Version = fmt.Sprintf("%x", h.Sum(nil))
	node.VersionParts = parts
	return
}

func (node *Node) UpdateLinkages(ltype string, ids string, operation string) (err error) {
	var link linkage
	link.Type = ltype
	idList := strings.Split(ids, ",")
	for _, id := range idList {
		link.Ids = append(link.Ids, id)
	}
	link.Operation = operation
	node.Linkages = append(node.Linkages, link)
	err = node.Save()
	return
}

func (node *Node) UpdateDataTags(types string) (err error) {
	tagslist := strings.Split(types, ",")
	for _, newtag := range tagslist {
		if contains(node.Tags, newtag) {
			continue
		}
		node.Tags = append(node.Tags, newtag)
	}
	err = node.Save()
	return
}

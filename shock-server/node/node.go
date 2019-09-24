package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node/acl"
	"github.com/MG-RAST/Shock/shock-server/node/archive"
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/MG-RAST/Shock/shock-server/node/file/index"
	"github.com/MG-RAST/Shock/shock-server/node/locker"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/MG-RAST/Shock/shock-server/util"
	"gopkg.in/mgo.v2/bson"
)

type Node struct {
	Id           string            `bson:"id" json:"id"`
	Version      string            `bson:"version" json:"version"`
	File         file.File         `bson:"file" json:"file"`
	Attributes   interface{}       `bson:"attributes" json:"attributes"`
	Indexes      Indexes           `bson:"indexes" json:"indexes"`
	Acl          acl.Acl           `bson:"acl" json:"-"`
	VersionParts map[string]string `bson:"version_parts" json:"version_parts"`
	Tags         []string          `bson:"tags" json:"tags"`
	Revisions    []Node            `bson:"revisions" json:"-"`
	Linkages     []linkage         `bson:"linkage" json:"linkage"`
	Priority     int               `bson:"priority" json:"priority"`
	CreatedOn    time.Time         `bson:"created_on" json:"created_on"`
	LastModified time.Time         `bson:"last_modified" json:"last_modified"`
	Expiration   time.Time         `bson:"expiration" json:"expiration"` // 0 means no expiration of Node
	Type         string            `bson:"type" json:"type"`
	Subset       Subset            `bson:"subset" json:"-"`
	Parts        *PartsList        `bson:"parts" json:"parts"`
	Locations    []Location        `bson:"locations" json:"locations"` // see below
	Restore      bool              `bson:"restore" json:"restore"`     // has a restore request been observed
}

// Location a data type to represent storage locations (defined in LocationConfig) and status of data in flight
type Location struct {
	ID            string     `bson:"id" json:"id"`                                           // name of the location, if present data is verified to exist in said location
	Stored        bool       `bson:"stored,omitempty" json:"stored,omitempty"`               //
	RequestedDate *time.Time `bson:"requestedDate,omitempty" json:"requestedDate,omitempty"` // what is the date the data item was send on its way
}

type linkage struct {
	Type      string   `bson:"relation" json:"relation"`
	Ids       []string `bson:"ids" json:"ids"`
	Operation string   `bson:"operation" json:"operation"`
}

type Indexes map[string]*IdxInfo

type IdxInfo struct {
	Type        string           `bson:"index_type" json:"-"`
	TotalUnits  int64            `bson:"total_units" json:"total_units"`
	AvgUnitSize int64            `bson:"average_unit_size" json:"average_unit_size"`
	Format      string           `bson:"format" json:"-"`
	CreatedOn   time.Time        `bson:"created_on" json:"created_on"`
	Locked      *locker.LockInfo `bson:"-" json:"locked"`
}

// Subset is used to store information about a subset node's parent and its index.
// A subset node's index defines the subset of the data file that this node represents.
// A subset node's index is immutable after it is defined.
type Subset struct {
	Parent Parent            `bson:"parent" json:"-"`
	Index  SubsetNodeIdxInfo `bson:"index" json:"-"`
}

type Parent struct {
	Id        string `bson:"id" json:"-"`
	IndexName string `bson:"index_name" json:"-"`
}

type SubsetNodeIdxInfo struct {
	Path        string `bson:"path" json:"-"`
	TotalUnits  int64  `bson:"total_units" json:"-"`
	AvgUnitSize int64  `bson:"average_unit_size" json:"-"`
	Format      string `bson:"format" json:"-"`
}

const (
	longDateForm  = "2006-01-02T15:04:05-07:00"
	shortDateForm = "2006-01-02"
)

func New(uuid string) (node *Node) {
	node = new(Node)
	node.Indexes = make(map[string]*IdxInfo)
	node.File.Checksum = make(map[string]string)

	if uuid == "" {
		node.setId()
	} else {

		logger.Infof("(Node-->New) we need to check with the upstream node (UUID-Master) if UUID is available ")

	}

	return
}

func (node *Node) DBInit() {
	node.File.Locked = locker.FileLockMgr.Get(node.Id)
	for name, info := range node.Indexes {
		info.Locked = locker.IndexLockMgr.Get(node.Id, name)
	}
}

func CreateNodeUpload(u *user.User, params map[string]string, files file.FormFiles) (node *Node, err error) {
	// if copying node or creating subset node from parent, check if user has rights to the original node

	checkSumMD5, hasCheckSumMD5 := params["checksum-md5"] //TODO make checksum generic using strings.Split("-") ?
	if hasCheckSumMD5 {
		matchingNodes := Nodes{}

		_, err = dbFind(bson.M{"file.checksum.md5": checkSumMD5, "type": "basic"}, &matchingNodes, "", nil) // TODO search in public and owner nodes only
		if err != nil {
			return nil, err
		}

		if len(matchingNodes) > 0 {
			var matchingNode *Node
			matchingNode = matchingNodes[0]

			params["copy_data"] = matchingNode.Id
			delete(params, "path")

		} else {
			// node not found, continue as usual
		}
	}

	if copy_data_id, hasCopyData := params["copy_data"]; hasCopyData {
		var copy_data_node *Node
		copy_data_node, err = Load(copy_data_id)
		if err != nil {
			return
		}

		rights := copy_data_node.Acl.Check(u.Uuid)
		if copy_data_node.Acl.Owner != u.Uuid && u.Admin == false && copy_data_node.Acl.Owner != "public" && rights["read"] == false {
			logger.Error("err@CreateNodeUpload: (Authenticate) id=" + copy_data_id + ": " + e.UnAuth)
			//responder.RespondWithError(ctx, http.StatusUnauthorized, e.UnAuth)
			//err = request.AuthError(err, ctx)
			err = fmt.Errorf("copy_data_node auth error")
			return
		}
	}

	if parentNode_id, hasParentNode := params["parent_node"]; hasParentNode {
		var parentNode *Node
		parentNode, err = Load(parentNode_id)
		if err != nil {
			return
		}

		rights := parentNode.Acl.Check(u.Uuid)
		if parentNode.Acl.Owner != u.Uuid && u.Admin == false && parentNode.Acl.Owner != "public" && rights["read"] == false {
			logger.Error("err@CreateNodeUpload: (Authenticate) id=" + parentNode_id + ": " + e.UnAuth)
			//responder.RespondWithError(ctx, http.StatusUnauthorized, e.UnAuth)
			//err = request.AuthError(err, ctx)
			err = fmt.Errorf("parentNode auth error")
			return
		}
	}

	node = New("")
	node.Type = "basic"

	node.Acl.SetOwner(u.Uuid)
	node.Acl.Set(u.Uuid, acl.Rights{"read": true, "write": true, "delete": true})

	err = node.Mkdir()
	if err != nil {
		return
	}

	// update saves node
	err = node.Update(params, files, false)
	if err != nil {
		err = fmt.Errorf("(node.Update) %s", err.Error())
		node.Rmdir()
	}

	return
}

func CreateNodesFromArchive(u *user.User, params map[string]string, files file.FormFiles, archiveId string) (nodes []*Node, err error) {
	// get parent node
	archiveNode, err := Load(archiveId)
	if err != nil {
		return nil, err
	}
	if archiveNode.File.Size == 0 {
		return nil, errors.New("parent archive node has no file")
	}

	// get format
	aFormat, hasFormat := params["archive_format"]
	if !hasFormat {
		return nil, errors.New("missing archive_format parameter. use one of: " + archive.ArchiveList)
	}
	if !archive.IsValidArchive(aFormat) {
		return nil, errors.New("invalid archive_format parameter. use one of: " + archive.ArchiveList)
	}

	// get attributes
	var attributes interface{}
	if attrFile, ok := files["attributes"]; ok {
		defer attrFile.Remove()
		attr, err := ioutil.ReadFile(attrFile.Path)
		if err != nil {
			return nil, err
		}
		if err = json.Unmarshal(attr, &attributes); err != nil {
			return nil, err
		}
	} else if attrStr, ok := params["attributes_str"]; ok {
		if err = json.Unmarshal([]byte(attrStr), &attributes); err != nil {
			return nil, err
		}
	}

	// get files / delete unpack dir when done
	fileList, unpackDir, err := archive.FilesFromArchive(aFormat, archiveNode.FilePath())
	defer os.RemoveAll(unpackDir)
	if err != nil {
		return nil, err
	}

	// preserve acls
	_, preserveAcls := params["preserve_acls"]

	// build nodes
	var tempNodes []*Node
	for _, f := range fileList {
		// create link
		link := linkage{Type: "parent", Operation: aFormat, Ids: []string{archiveId}}
		// create and populate node
		node := New("")
		node.Type = "basic"
		node.Linkages = append(node.Linkages, link)
		node.Attributes = attributes

		if preserveAcls {
			// copy over acls from parent node
			node.Acl = archiveNode.Acl
		}
		// this user needs to be owner of new nodes
		node.Acl.SetOwner(u.Uuid)
		node.Acl.Set(u.Uuid, acl.Rights{"read": true, "write": true, "delete": true})

		if err = node.Mkdir(); err != nil {
			return nil, err
		}
		// set file
		ffile := file.FormFile{Name: f.Name, Path: f.Path, Checksum: f.Checksum}
		if err = node.SetFile(ffile); err != nil {
			node.Rmdir()
			return nil, err
		}
		tempNodes = append(tempNodes, node)
	}

	// save nodes, only return those that were created / saved
	for _, n := range tempNodes {
		if serr := n.Save(); serr != nil {
			n.Rmdir()
			continue
		}
		nodes = append(nodes, n)
	}
	return
}

func (node *Node) FileReader() (reader file.ReaderAt, err error) {
	if node.File.Virtual {
		readers := []file.ReaderAt{}
		nodes := Nodes{}
		if _, err := dbFind(bson.M{"id": bson.M{"$in": node.File.VirtualParts}}, &nodes, "", nil); err != nil {
			return nil, err
		}
		if len(nodes) > 0 {
			for _, n := range nodes {
				if r, err := n.FileReader(); err == nil {
					readers = append(readers, r)
				} else {
					return nil, err
				}
			}
		}
		return file.MultiReaderAt(readers...), nil
	}
	return FMOpen(node.FilePath())
}

func (node *Node) DynamicIndex(name string) (idx index.Index, err error) {
	if index.Has(name) {
		idx = index.NewVirtual(name, node.FilePath(), node.File.Size, 10240)
	} else {
		if _, has := node.Indexes[name]; has {
			idx = index.New()
		} else {
			err_str := fmt.Sprintf("Node %s does not have index of type %s.", node.Id, name)
			err = errors.New(err_str)
		}
	}
	return
}

//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************

// DeleteFiles delete the files on disk while keeping information in Mongo
// FMopen will stage back the files from external Locations if requested
// data for nodes will subsequently be cached in PATH_CACHE not stored in PATH_DATA
func (node *Node) deleteFiles() (err error) {
	// lock node
	err = locker.NodeLockMgr.LockNode(node.Id)
	if err != nil {
		return
	}
	defer locker.NodeLockMgr.Remove(node.Id)

	// Check to see if this node has a data file and if it's referenced by another node.
	// If it is, we will move the data file to the first node we find, and point all other nodes to that node's path
	dataFilePath := fmt.Sprintf("%s/%s.data", getPath(node.Id), node.Id)
	dataFileExists := true
	if _, ferr := os.Stat(dataFilePath); os.IsNotExist(ferr) {
		dataFileExists = false
	}
	newDataFilePath := ""
	copiedNodes := Nodes{}
	if _, err = dbFind(bson.M{"file.path": dataFilePath}, &copiedNodes, "", nil); err != nil {
		return err
	}
	if len(copiedNodes) != 0 && dataFileExists {
		for index, copiedNode := range copiedNodes {
			// lock copynode for save
			err = locker.NodeLockMgr.LockNode(copiedNode.Id)
			if err != nil {
				err = errors.New("This node has a data file linked to another node which could not be locked during data file copy: " + err.Error())
				return
			}
			defer locker.NodeLockMgr.UnlockNode(copiedNode.Id)

			if index == 0 {
				newDataFilePath = fmt.Sprintf("%s/%s.data", getPath(copiedNode.Id), copiedNode.Id)
				if rerr := os.Rename(dataFilePath, newDataFilePath); rerr != nil {
					if _, cerr := util.CopyFile(dataFilePath, newDataFilePath); cerr != nil {
						err = errors.New("This node has a data file linked to another node and the data file could not be copied elsewhere to allow for node deletion.")
						return
					}
				}
				copiedNode.File.Path = ""
				copiedNode.Save()
			} else {
				copiedNode.File.Path = newDataFilePath
				copiedNode.Save()
			}
		}
	}

	// we shoudl really delete the index files as well
	logger.Debug(1, "(Node->DeleteFiles) we still need to delete the index files!")
	// IndexFilePath := fmt.Sprintf("%s/%s.idx", node.IndexPath(), indextype)
	// if err = os.Remove(IndexFilePath); err != nil {
	// 	return
	// }

	return
}

//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************

// ExpireNode -- delete the node from Mongo and Disk
func (node *Node) Delete() (deleted bool, err error) {
	// lock node
	err = locker.NodeLockMgr.LockNode(node.Id)
	if err != nil {
		return
	}
	defer locker.NodeLockMgr.Remove(node.Id)

	// check to make sure this node isn't referenced by a vnode
	virtualNodes := Nodes{}
	if _, err = dbFind(bson.M{"file.virtual_parts": node.Id}, &virtualNodes, "", nil); err != nil {
		return
	}
	if len(virtualNodes) != 0 {
		err = errors.New(e.NodeReferenced)
		return
	}

	// Check to see if this node has a data file and if it's referenced by another node.
	// If it is, we will move the data file to the first node we find, and point all other nodes to that node's path
	dataFilePath := fmt.Sprintf("%s/%s.data", getPath(node.Id), node.Id)
	dataFileExists := true
	if _, ferr := os.Stat(dataFilePath); os.IsNotExist(ferr) {
		dataFileExists = false
	}
	newDataFilePath := ""
	copiedNodes := Nodes{}
	if _, err = dbFind(bson.M{"file.path": dataFilePath}, &copiedNodes, "", nil); err != nil {
		return
	}
	if len(copiedNodes) != 0 && dataFileExists {
		for index, copiedNode := range copiedNodes {
			// lock copynode for save
			err = locker.NodeLockMgr.LockNode(copiedNode.Id)
			if err != nil {
				err = errors.New("This node has a data file linked to another node which could not be locked during data file copy: " + err.Error())
				return
			}
			defer locker.NodeLockMgr.UnlockNode(copiedNode.Id)

			if index == 0 {
				newDataFilePath = fmt.Sprintf("%s/%s.data", getPath(copiedNode.Id), copiedNode.Id)
				if rerr := os.Rename(dataFilePath, newDataFilePath); rerr != nil {
					if _, cerr := util.CopyFile(dataFilePath, newDataFilePath); cerr != nil {
						err = errors.New("This node has a data file linked to another node and the data file could not be copied elsewhere to allow for node deletion.")
						return
					}
				}
				copiedNode.File.Path = ""
				copiedNode.Save()
			} else {
				copiedNode.File.Path = newDataFilePath
				copiedNode.Save()
			}
		}
	}

	err = dbDelete(bson.M{"id": node.Id})
	if err != nil {
		logger.Debug(2, "(Node->Delete) we failed to delete %s from Mongo database", node.Id, err.Error())
		return
	}
	err = node.Rmdir()
	if err != nil {
		logger.Debug(2, "(Node->Delete) we failed to delete %s from disk", node.Id, err.Error())
		return
	}
	deleted = true
	return
}

//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
// ExpireNodeFiles -- remove files in DATA_PATH if present on >= conf.MIN_REPLICA_COUNT external Locations
func (n *Node) ExpireNodeFiles() (deleted bool, err error) {

	// lock node
	err = locker.NodeLockMgr.LockNode(n.Id)
	if err != nil {
		return
	}
	defer locker.NodeLockMgr.Remove(n.Id)

	// leave if we are not set up to remove NodeFiles
	if conf.NODE_DATA_REMOVAL == false {
		deleted = false
		return
	}

LocationsLoop:
	for _, loc := range n.Locations {
		counter := 0 // we need at least N locations before we can erase data files on local disk
		// delete only if other locations exist
		locObj, ok := conf.LocationsMap[loc.ID]
		if !ok {
			logger.Errorf("(Reaper-->FileReaper) location %s is not defined in this server instance \n ", loc)
			continue LocationsLoop
		}
		//fmt.Printf("(Reaper-->FileReaper) locObj.Persistent =  %b  \n ", locObj.Persistent)
		if locObj.Persistent == true {
			logger.Debug(2, "(Reaper-->FileReaper) has remote Location (%s) removing from Data: %s", loc.ID, n.Id)
			counter++ // increment counter
		}
		if counter >= conf.MIN_REPLICA_COUNT {
			err = n.deleteFiles() // delete all data files for node in PATH_DATA NOTE: this is different from PATH_CACHE
			if err != nil {
				logger.Errorf("(Reaper-->FileReaper) files for node %s could not be deleted (Err: %s) ", n.Id, err.Error())
				continue
			}
			deleted = true
			return
		}
	}
	return
}

//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
// DeleteIndex for node
func (node *Node) DeleteIndex(indextype string) (err error) {
	// lock node
	err = locker.NodeLockMgr.LockNode(node.Id)
	if err != nil {
		return
	}
	defer locker.NodeLockMgr.UnlockNode(node.Id)

	delete(node.Indexes, indextype)
	IndexFilePath := fmt.Sprintf("%s/%s.idx", node.IndexPath(), indextype)
	if err = os.Remove(IndexFilePath); err != nil {
		return
	}
	err = node.Save()
	return
}

func (node *Node) SetIndexInfo(indextype string, idxinfo *IdxInfo) {
	node.Indexes[indextype] = idxinfo
}

func (node *Node) SetExpiration(expire string) (err error) {
	parts := ExpireRegex.FindStringSubmatch(expire)
	if len(parts) == 0 {
		return errors.New("expiration format is invalid")
	}
	var expireTime time.Duration
	expireNum, _ := strconv.Atoi(parts[1])
	currTime := time.Now()

	switch parts[2] {
	case "M":
		expireTime = time.Duration(expireNum) * time.Minute
	case "H":
		expireTime = time.Duration(expireNum) * time.Hour
	case "D":
		expireTime = time.Duration(expireNum*24) * time.Hour
	}

	node.Expiration = currTime.Add(expireTime)
	return
}

func (node *Node) SetAttributes(attr file.FormFile) (err error) {
	defer attr.Remove()
	attributes, err := ioutil.ReadFile(attr.Path)
	if err != nil {
		return
	}
	err = json.Unmarshal(attributes, &node.Attributes)
	if err != nil {
		return
	}
	return
}

func (node *Node) SetAttributesFromString(attributes string) (err error) {
	err = json.Unmarshal([]byte(attributes), &node.Attributes)
	if err != nil {
		return
	}
	return
}

func (node *Node) UpdateDataTags(types string) {
	tagslist := strings.Split(types, ",")
	for _, newtag := range tagslist {
		if contains(node.Tags, newtag) {
			continue
		}
		node.Tags = append(node.Tags, newtag)
	}
}

func (node *Node) UpdateLinkages(ltype string, ids string, operation string) {
	var link linkage
	link.Type = ltype
	idList := strings.Split(ids, ",")
	for _, id := range idList {
		link.Ids = append(link.Ids, id)
	}
	link.Operation = operation
	node.Linkages = append(node.Linkages, link)
}

// AddLocation _
func (node *Node) AddLocation(loc Location) (err error) {
	if node.Locations == nil {
		node.Locations = []Location{loc}
		return
	}

	for _, location := range node.Locations {
		if location.ID == loc.ID {
			err = fmt.Errorf("%s already exists", loc.ID)
			return
		}
	}

	node.Locations = append(node.Locations, loc)
	return
}

// GetLocations _
func (node *Node) GetLocations() (locations []Location) {

	locations = node.Locations
	if locations == nil {
		locations = []Location{}
	}

	return
}

// GetLocation _
func (node *Node) GetLocation(locID string) (myLocation Location, err error) {
	if node.Locations == nil {
		err = fmt.Errorf("location %s not found", locID)
		return
	}

	for _, location := range node.Locations {
		if location.ID == locID {
			myLocation = location
			return
		}

	}

	err = fmt.Errorf("location %s not found", locID)

	return
}

// DeleteLocation _
func (node *Node) DeleteLocation(locID string) (err error) {
	if node.Locations == nil {
		err = fmt.Errorf("location %s not found", locID)
		return
	}

	newLocations := []Location{}
	found := false
	for _, location := range node.Locations {
		if location.ID == locID {
			found = true
			continue
		}
		newLocations = append(newLocations, location)
	}

	if !found {
		err = fmt.Errorf("location %s not found", locID)
		return
	}
	node.Locations = newLocations
	return
}

// DeleteLocations _
func (node *Node) DeleteLocations() (err error) {
	node.Locations = []Location{}
	return
}

// GetRestore return true if node has been marked for restoring from external Location
func (node *Node) GetRestore() (stat bool) {

	stat = node.Restore

	return
}

// SetRestore set Restore value to true to mark node for restoring from external Location
func (node *Node) SetRestore() {
	node.Restore = true
	return
}

// UnsetRestore set Restore value to false, node restore has been requested e.g. via TSM client
func (node *Node) UnSetRestore() {

	node.Restore = false
	return
}

package shock

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/MG-RAST/golib/httpclient"
)

// TODO use Token

var WAIT_SLEEP = 30 * time.Second
var WAIT_TIMEOUT = time.Duration(-1) * time.Hour
var SHOCK_TIMEOUT = 60 * time.Second
var DATA_SUFFIX = "?download"

type ShockClient struct {
	Host  string
	Token string
	Debug bool
}

type ShockResponse struct {
	Code int        `bson:"status" json:"status"`
	Data *ShockNode `bson:"data" json:"data"`
	Errs []string   `bson:"error" json:"error"`
}

type ShockResponseGeneric struct {
	Code int         `bson:"status" json:"status"`
	Data interface{} `bson:"data" json:"data"`
	Errs []string    `bson:"error" json:"error"`
}

type ShockResponseMap map[string]interface{}

type ShockQueryResponse struct {
	Code       int         `bson:"status" json:"status"`
	Data       []ShockNode `bson:"data" json:"data"`
	Errs       []string    `bson:"error" json:"error"`
	Limit      int         `bson:"limit" json:"limit"`
	Offset     int         `bson:"offset" json:"offset"`
	TotalCount int         `bson:"total_count" json:"total_count"`
}

type ShockNode struct {
	Id           string             `bson:"id" json:"id"`
	Version      string             `bson:"version" json:"version"`
	File         shockFile          `bson:"file" json:"file"`
	Attributes   interface{}        `bson:"attributes" json:"attributes"`
	Indexes      map[string]IdxInfo `bson:"indexes" json:"indexes"`
	VersionParts map[string]string  `bson:"version_parts" json:"-"`
	Tags         []string           `bson:"tags" json:"tags"`
	Linkages     []linkage          `bson:"linkage" json:"linkages"`
	Priority     int                `bson:"priority" json:"priority"`
	CreatedOn    time.Time          `bson:"created_on" json:"created_on"`
	LastModified time.Time          `bson:"last_modified" json:"last_modified"`
	Expiration   time.Time          `bson:"expiration" json:"expiration"` // 0 means no expiration
	Type         string             `bson:"type" json:"type"`
	Parts        *partsList         `bson:"parts" json:"parts"`
}

type shockFile struct {
	Name         string            `bson:"name" json:"name"`
	Size         int64             `bson:"size" json:"size"`
	Checksum     map[string]string `bson:"checksum" json:"checksum"`
	Format       string            `bson:"format" json:"format"`
	Virtual      bool              `bson:"virtual" json:"virtual"`
	VirtualParts []string          `bson:"virtual_parts" json:"virtual_parts"`
	CreatedOn    time.Time         `bson:"created_on" json:"created_on"`
	Locked       *lockInfo         `bson:"-" json:"locked"`
}

type IdxInfo struct {
	TotalUnits  int64     `bson:"total_units" json:"total_units"`
	AvgUnitSize int64     `bson:"average_unit_size" json:"average_unit_size"`
	CreatedOn   time.Time `bson:"created_on" json:"created_on"`
	Locked      *lockInfo `bson:"-" json:"locked"`
}

type linkage struct {
	Type      string   `bson: "relation" json:"relation"`
	Ids       []string `bson:"ids" json:"ids"`
	Operation string   `bson:"operation" json:"operation"`
}

type lockInfo struct {
	CreatedOn time.Time `bson:"-" json:"created_on"`
	Error     string    `bson:"-" json:"error"`
}

type partsFile []string

type partsList struct {
	Count       int         `bson:"count" json:"count"`
	Length      int         `bson:"length" json:"length"`
	VarLen      bool        `bson:"varlen" json:"varlen"`
	Parts       []partsFile `bson:"parts" json:"parts"`
	Compression string      `bson:"compression" json:"compression"`
}

type Opts map[string]string

func (o *Opts) HasKey(key string) bool {
	if _, has := (*o)[key]; has {
		return true
	}
	return false
}

func (o *Opts) Value(key string) string {
	val, _ := (*o)[key]
	return val
}

func NewShockClient(shock_url string, shock_auth string, debug bool) (sc *ShockClient) {
	sc = &ShockClient{Host: shock_url, Token: shock_auth, Debug: debug}
	return
}

// *** low-level functions ***

func (sc *ShockClient) getRequest(resource string, query url.Values, response interface{}) (err error) {
	return sc.doRequest("GET", resource, query, response)
}

func (sc *ShockClient) putRequest(resource string, query url.Values, response interface{}) (err error) {
	return sc.doRequest("PUT", resource, query, response)
}

func (sc *ShockClient) postRequest(resource string, query url.Values, response interface{}) (err error) {
	return sc.doRequest("POST", resource, query, response)
}

func (sc *ShockClient) deleteRequest(resource string, query url.Values, response interface{}) (err error) {
	return sc.doRequest("DELETE", resource, query, response)
}

// for GET / PUT / POST / DELETE with no multipart from
func (sc *ShockClient) doRequestString(method string, resource string, query url.Values) (jsonstream []byte, err error) {
	var myurl *url.URL
	myurl, err = url.ParseRequestURI(sc.Host)
	if err != nil {
		return
	}

	(*myurl).Path = resource
	(*myurl).RawQuery = query.Encode()
	shockurl := myurl.String()

	if sc.Debug {
		fmt.Fprintf(os.Stdout, "doRequest url: %s\n", shockurl)
	}
	if len(shockurl) < 5 {
		err = errors.New("could not parse shockurl: " + shockurl)
		return
	}

	var user *httpclient.Auth
	if sc.Token != "" {
		user = httpclient.GetUserByTokenAuth(sc.Token)
	}

	var res *http.Response
	res, err = httpclient.Do(method, shockurl, httpclient.Header{}, nil, user)
	if err != nil {
		return
	}

	defer res.Body.Close()

	jsonstream, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	return
}

// for GET / PUT / POST / DELETE with no multipart from
func (sc *ShockClient) doRequest(method string, resource string, query url.Values, response interface{}) (err error) {
	var jsonstream []byte
	jsonstream, err = sc.doRequestString(method, resource, query)
	if err != nil {
		return
	}
	err = json.Unmarshal(jsonstream, response)
	return
}

// for PUT / POST with multipart form
func (sc *ShockClient) createOrUpdate(opts Opts, nodeid string, nodeattr map[string]interface{}) (node *ShockNode, err error) {
	host := sc.Host
	token := sc.Token

	if host == "" {
		err = errors.New("(createOrUpdate) host is not defined in Shock node")
		return
	}

	url := host + "/node"
	method := "POST"
	if nodeid != "" {
		url += "/" + nodeid
		method = "PUT"
	}
	form := httpclient.NewForm()

	// attributes
	if opts.HasKey("attributes") {
		form.AddFile("attributes", opts.Value("attributes"))
	}
	if len(nodeattr) != 0 {
		var nodeattr_json []byte
		nodeattr_json, err = json.Marshal(nodeattr)
		if err != nil {
			err = errors.New("(createOrUpdate) error marshalling NodeAttr")
			return
		}
		form.AddParam("attributes_str", string(nodeattr_json[:]))
	}

	// expiration
	if opts.HasKey("expiration") {
		form.AddParam("expiration", opts.Value("expiration"))
	}
	if opts.HasKey("remove_expiration") {
		form.AddParam("remove_expiration", "1")
	}

	// filename
	if opts.HasKey("file_name") {
		form.AddParam("file_name", opts.Value("file_name"))
	}

	var uploadType string
	if opts.HasKey("upload_type") {
		uploadType = opts.Value("upload_type")
	}

	if uploadType != "" {
		switch uploadType {
		case "basic":

			// upload string as file
			if opts.HasKey("file_content") {
				fileContent := opts.Value("file_content")
				fileContentLen := int64(len(fileContent))
				reader := bytes.NewBufferString(fileContent)
				form.AddFileReader("upload", reader, fileContentLen)
			}

			if opts.HasKey("file") {
				if opts.HasKey("compression") {
					form.AddFile(opts.Value("compression"), opts.Value("file"))
				} else {
					form.AddFile("upload", opts.Value("file"))
				}
			}
			if opts.HasKey("checksum-md5") {
				form.AddParam("checksum-md5", opts.Value("checksum-md5"))
			}

		case "parts":
			if opts.HasKey("parts") {
				form.AddParam("parts", opts.Value("parts"))
			} else {
				err = errors.New("(createOrUpdate) (case:parts) missing partial upload parameter: parts")
				return
			}
			if opts.HasKey("compression") {
				form.AddParam("compression", opts.Value("compression"))
			}
		case "part":
			if opts.HasKey("part") && opts.HasKey("file") {
				form.AddFile(opts.Value("part"), opts.Value("file"))
			} else {
				err = errors.New("(createOrUpdate) (case:part) missing partial upload parameter: part or file")
				return
			}
		case "remote":
			if opts.HasKey("remote_url") {
				form.AddParam("upload_url", opts.Value("remote_url"))
			} else {
				err = errors.New("(createOrUpdate) (case:remote_url) missing remote path parameter: path")
				return
			}
		case "virtual":
			if opts.HasKey("virtual_file") {
				form.AddParam("type", "virtual")
				form.AddParam("source", opts.Value("virtual_file"))
			} else {
				err = errors.New("(createOrUpdate) (case:virtual_file) missing virtual node parameter: source")
				return
			}
		case "copy":
			if opts.HasKey("parent_node") {
				form.AddParam("copy_data", opts.Value("parent_node"))
			} else {
				err = errors.New("(createOrUpdate) (case:copy) missing copy node parameter: parent_node")
				return
			}
			if opts.HasKey("copy_indexes") {
				form.AddParam("copy_indexes", "1")
			}
			if opts.HasKey("copy_attributes") {
				form.AddParam("copy_attributes", "1")
			}
		case "subset":
			if opts.HasKey("parent_node") && opts.HasKey("parent_index") && opts.HasKey("file") {
				form.AddParam("parent_node", opts.Value("parent_node"))
				form.AddParam("parent_index", opts.Value("parent_index"))
				form.AddFile("subset_indices", opts.Value("file"))
			} else {
				err = errors.New("(createOrUpdate) (case:subset) missing subset node parameter: parent_node or parent_index or file")
				return
			}
		}
	}

	err = form.Create()
	if err != nil {
		err = fmt.Errorf("(createOrUpdate) form.Create returned: %s", err.Error())
		return
	}

	headers := httpclient.Header{
		"Content-Type":   []string{form.ContentType},
		"Content-Length": []string{strconv.FormatInt(form.Length, 10)},
	}
	var user *httpclient.Auth
	if token != "" {
		user = httpclient.GetUserByTokenAuth(token)
	}
	if sc.Debug {
		fmt.Printf("(createOrUpdate) url: %s %s\n", method, url)
		fmt.Printf("multipart form: ContentType=%s Length=%d\n", form.ContentType, form.Length)
	}
	var res *http.Response
	res, err = httpclient.Do(method, url, headers, form.Reader, user)
	if err != nil {
		err = fmt.Errorf("(createOrUpdate) httpclient.Do returned: %s", err.Error())
		return
	}

	defer res.Body.Close()
	jsonstream, err := ioutil.ReadAll(res.Body)
	if err != nil {
		err = fmt.Errorf("(createOrUpdate) ioutil.ReadAll returned: %s", err.Error())
		return
	}
	response := new(ShockResponse)
	err = json.Unmarshal(jsonstream, response)
	if err != nil {
		err = fmt.Errorf("(createOrUpdate) (httpclient.Do) failed to marshal response:\"%s\" (err: %s)", jsonstream, err.Error())
		return
	}
	if len(response.Errs) > 0 {
		err = fmt.Errorf("(createOrUpdate) type=%s, method=%s, url=%s, error=%s", uploadType, method, url, strings.Join(response.Errs, ","))
		return
	}
	node = response.Data
	return
}

// *** high-level functions ***

// unpack archive node, creates multiple nodes
func (sc *ShockClient) UnpackArchiveNode(nodeid string, format string, attrfile string) (nodes interface{}, err error) {
	form := httpclient.NewForm()
	if attrfile != "" {
		form.AddFile("attributes", attrfile)
	}
	form.AddParam("unpack_node", nodeid)
	form.AddParam("archive_format", format)

	err = form.Create()
	if err != nil {
		err = fmt.Errorf("(UnpackArchiveNode) form.Create returned: %s", err.Error())
		return
	}

	headers := httpclient.Header{
		"Content-Type":   []string{form.ContentType},
		"Content-Length": []string{strconv.FormatInt(form.Length, 10)},
	}

	var user *httpclient.Auth
	if sc.Token != "" {
		user = httpclient.GetUserByTokenAuth(sc.Token)
	}

	var res *http.Response
	res, err = httpclient.Do("POST", sc.Host+"/node", headers, form.Reader, user)
	if err != nil {
		err = fmt.Errorf("(UnpackArchiveNode) httpclient.Do returned: %s", err.Error())
		return
	}

	defer res.Body.Close()
	jsonstream, err := ioutil.ReadAll(res.Body)
	if err != nil {
		err = fmt.Errorf("(UnpackArchiveNode) ioutil.ReadAll returned: %s", err.Error())
		return
	}
	response := new(ShockResponseGeneric)
	err = json.Unmarshal(jsonstream, response)
	if err != nil {
		err = fmt.Errorf("(UnpackArchiveNode) failed to marshal response:\"%s\" (err: %s)", jsonstream, err.Error())
		return
	}

	if len(response.Errs) > 0 {
		err = fmt.Errorf("(UnpackArchiveNode) error=%s", strings.Join(response.Errs, ","))
		return
	}

	return response.Data, nil
}

func (sc *ShockClient) ServerInfo() (srm *ShockResponseMap, err error) {
	srm = new(ShockResponseMap)
	err = sc.getRequest("", nil, &srm)
	return
}

func (sc *ShockClient) WaitIndex(nodeid string, indexname string) (index IdxInfo, err error) {
	if indexname == "" {
		return
	}
	var node *ShockNode
	var has bool
	startTime := time.Now()
	for {
		node, err = sc.GetNode(nodeid)
		if err != nil {
			return
		}
		index, has = node.Indexes[indexname]
		if !has {
			err = fmt.Errorf("(shock.WaitIndex) index does not exist: node=%s, index=%s", nodeid, indexname)
			return
		}
		if (index.Locked != nil) && (index.Locked.Error != "") {
			// we have an error, return it
			err = fmt.Errorf("(shock.WaitIndex) error creating index: node=%s, index=%s, error=%s", nodeid, indexname, index.Locked.Error)
			return
		}
		if index.Locked == nil {
			// no longer locked
			return
		}
		// need a resonable timeout
		currTime := time.Now()
		expireTime := currTime.Add(WAIT_TIMEOUT) // adding negative time to current time, because no subtraction function
		if startTime.Before(expireTime) {
			err = fmt.Errorf("(shock.WaitIndex) timeout waiting on index lock: node=%s, index=%s", nodeid, indexname)
			return
		}
		time.Sleep(WAIT_SLEEP)
	}
	return
}

func (sc *ShockClient) WaitFile(nodeid string) (node *ShockNode, err error) {
	startTime := time.Now()
	for {
		node, err = sc.GetNode(nodeid)
		if err != nil {
			return
		}
		if (node.File.Locked != nil) && (node.File.Locked.Error != "") {
			// we have an error, return it
			err = fmt.Errorf("(shock.WaitFile) error waiting on file lock: node=%s, error=%s", nodeid, node.File.Locked.Error)
			return
		}
		if node.File.Locked == nil {
			// no longer locked
			return
		}
		// need a resonable timeout
		currTime := time.Now()
		expireTime := currTime.Add(WAIT_TIMEOUT) // adding negative time to current time, because no subtraction function
		if startTime.Before(expireTime) {
			err = fmt.Errorf("(shock.WaitFile) timeout waiting on file lock: node=%s", nodeid)
			return
		}
		time.Sleep(WAIT_SLEEP)
	}
	return
}

// upload file to node, creates new node if no nodeid given, handle different node types
func (sc *ShockClient) PutOrPostFile(filename string, nodeid string, rank int, attrfile string, ntype string, formopts map[string]string, nodeattr map[string]interface{}) (newNodeid string, err error) {
	opts := Opts{}
	fi, _ := os.Stat(filename)

	if (attrfile != "") && (rank < 2) {
		opts["attributes"] = attrfile
	}
	if filename != "" {
		opts["file"] = filename
	}
	if rank == 0 {
		opts["upload_type"] = "basic"
	} else {
		opts["upload_type"] = "part"
		opts["part"] = strconv.Itoa(rank)
	}

	// apply form options for different node types
	if (ntype == "subset") && (rank == 0) && (fi.Size() == 0) {
		opts["upload_type"] = "basic"
	} else if ((ntype == "copy") || (ntype == "subset") || (ntype == "remote") || (ntype == "virtual") || (ntype == "parts")) && (len(formopts) > 0) {
		opts["upload_type"] = ntype
		for k, v := range formopts {
			opts[k] = v
		}
	}
	if v, ok := formopts["compression"]; ok {
		opts["compression"] = v
	}

	var node *ShockNode
	node, err = sc.createOrUpdate(opts, nodeid, nodeattr)
	if err != nil {
		err = fmt.Errorf("(PutFileToShock) failed (%s): %v", sc.Host, err)
		return
	}
	if node != nil {
		newNodeid = node.Id
	}
	return
}

// PostFile create basic node with file POST
// use filepath OR fileContent for upload
func (sc *ShockClient) PostFile(filepath string, fileContent string, filename string) (nodeid string, err error) {
	if sc.Debug {
		fmt.Fprintf(os.Stdout, "(PostFile) filepath: %s\n", filepath)
	}
	opts := Opts{
		"upload_type": "basic",
	}

	if filepath != "" && fileContent != "" {
		err = fmt.Errorf("(PostFile) do use filepath and fileContent arguments at the same time", sc.Host, err)
		return
	}

	if filepath != "" {
		opts["file"] = filepath
	}

	if fileContent != "" {
		opts["file_content"] = fileContent
	}

	if filename != "" {
		opts["file_name"] = filename
	}

	var node *ShockNode
	node, err = sc.createOrUpdate(opts, "", nil)
	if err != nil {
		return
	}
	if node != nil {
		nodeid = node.Id
	}
	return
}

// PostFileLazy create basic node if file is not in shock already, otherwise returns a existing node
// createCopyNode boolean, if true, should return a copynode if node exists
func (sc *ShockClient) PostFileLazy(filepath string, filename string, fileContent string, createCopyNode bool) (nodeid string, err error) {

	if createCopyNode {
		err = fmt.Errorf("(PostFileLazy) createCopyNode not implemented yet")
		return
	}

	var md5sum string
	if filepath != "" {
		md5Filename := filepath + ".md5"

		if sc.Debug {
			fmt.Fprintf(os.Stdout, "(PostFileLazy) filepath=%s  (md5Filename=%s)\n", filepath, md5Filename)
		}

		//var md5FileInfo os.FileInfo
		_, err = os.Stat(md5Filename)
		if err != nil {
			// did not find .md5 file, calculate md5

			if sc.Debug {
				fmt.Fprintf(os.Stdout, "(PostFileLazy) calculating md5sum... (%s, %s)\n", md5Filename, err.Error())
			}

			md5sum, err = GetMD5FromFile(filepath)
			if err != nil {
				err = fmt.Errorf("(PostFileLazy) GetMD5FromFile returned: %s", err.Error())
				return
			}

			baseName := path.Base(filepath)
			md5sumFileContent := md5sum + " " + baseName

			if sc.Debug {
				fmt.Fprintf(os.Stdout, "(PostFileLazy) calculated md5sum: %s\n", md5sum)
			}

			err = ioutil.WriteFile(md5Filename, []byte(md5sumFileContent), 0644)
			if err != nil {
				fmt.Fprintf(os.Stdout, "(PostFileLazy) could not write md5sum, ioutil.WriteFile returned: %s, continue anyway", err.Error())
				err = nil
			}

		} else {
			// .md5 file exits
			var md5sumByteArray []byte
			md5sumByteArray, err = ioutil.ReadFile(md5Filename) // just pass the file name
			if err != nil {
				err = fmt.Errorf("(PostFileLazy) could not read md5sum, ioutil.ReadFile returned: %s", err.Error())
				return
			}

			md5sum = string(md5sumByteArray[0:32])

			if sc.Debug {
				fmt.Fprintf(os.Stdout, "(PostFileLazy) got cached md5sum: %s\n", md5sum)
			}
		}

		var ok bool
		nodeid, ok, err = sc.GetNodeByMD5(md5sum)
		if err != nil {
			fmt.Fprintf(os.Stdout, "(PostFileLazy) GetNodeByMD5 returned: %s", err.Error())
			err = nil
		}

		if ok {
			return
		}
	}
	nodeid, err = sc.PostFile(filepath, fileContent, filename)
	if err != nil {
		err = fmt.Errorf("(PostFileLazy) sc.PostFile returned: %s", err.Error())
		return
	}

	return
}

// GetNodeByMD5 _
func (sc *ShockClient) GetNodeByMD5(md5sum string) (nodeid string, ok bool, err error) {
	var searchQuery url.Values
	searchQuery = make(url.Values)
	searchQuery["file.checksum.md5"] = []string{md5sum}
	searchQuery["type"] = []string{"basic"}

	var sr *ShockQueryResponse
	sr, err = sc.QueryFull(searchQuery)
	if err != nil {
		err = fmt.Errorf("(GetNodeByMD5) sc.QueryFull returned: %s", err.Error())
		return
	}
	if sr.Code != 200 {
		err = fmt.Errorf("(GetNodeByMD5) sc.QueryFull returned error code %d (%s)", sr.Code, strings.Join(sr.Errs, ","))
		return
	}

	if len(sr.Data) > 0 {
		nodeid = sr.Data[0].Id
		if sc.Debug {
			fmt.Fprintf(os.Stdout, "(GetNodeByMD5) found existing node: %s\n", nodeid)
		}
		ok = true
		return
	}
	if sc.Debug {
		fmt.Fprintf(os.Stdout, "(GetNodeByMD5) file not found in Shock\n")
	}

	ok = false

	return
}

// GetMD5FromFile _
func GetMD5FromFile(filePath string) (md5sum string, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := md5.New()
	_, err = io.Copy(h, f)
	if err != nil {
		log.Fatal(err)
		return
	}

	md5sum = fmt.Sprintf("%x", h.Sum(nil))
	return
}

// PostFileWithAttributes create basic node with file POST
func (sc *ShockClient) PostFileWithAttributes(filepath string, filename string, nodeattr map[string]interface{}) (node *ShockNode, err error) {
	opts := Opts{
		"upload_type": "basic",
		"file":        filepath,
	}
	if filename != "" {
		opts["file_name"] = filename
	}
	node, err = sc.createOrUpdate(opts, "", nodeattr)
	return
}

// CreateNode create empty node, basic or parts
func (sc *ShockClient) CreateNode(filename string, numParts int) (nodeid string, err error) {
	sr := new(ShockResponse)
	err = sc.postRequest("/node", nil, &sr)
	if err != nil {
		err = fmt.Errorf("(CreateNode) %s", err.Error())
		return
	}

	if len(sr.Errs) > 0 {
		err = fmt.Errorf("(CreateNode) %s", strings.Join(sr.Errs, ","))
		return
	}
	if sr.Data != nil {
		nodeid = sr.Data.Id
	}

	// create "parts" for output splits
	if numParts > 1 {
		opts := Opts{"upload_type": "parts", "file_name": path.Base(filename), "parts": strconv.Itoa(numParts)}
		_, err = sc.createOrUpdate(opts, nodeid, nil)
		if err != nil {
			err = fmt.Errorf("(CreateNode) node=%s: %s", nodeid, err.Error())
		}
	}
	return
}

// UpdateAttributes update node attributes
func (sc *ShockClient) UpdateAttributes(nodeid string, attrfile string, nodeattr map[string]interface{}) (err error) {
	opts := Opts{}
	if attrfile != "" {
		opts["attributes"] = attrfile
	}
	_, err = sc.createOrUpdate(opts, nodeid, nodeattr)
	return
}

// UpdateFilename change node filename
func (sc *ShockClient) UpdateFilename(nodeid string, filename string) (err error) {
	opts := Opts{
		"upload_type": "basic",
		"file_name":   filename,
	}
	_, err = sc.createOrUpdate(opts, nodeid, nil)
	return
}

// Expiration add / modify / delete expiration
func (sc *ShockClient) Expiration(nodeid string, value int, format byte) (err error) {
	opts := Opts{}
	if value == 0 {
		opts["remove_expiration"] = ""
	} else if value > 0 {
		opts["expiration"] = strconv.Itoa(value) + string(format)
	} else {
		return
	}
	_, err = sc.createOrUpdate(opts, nodeid, nil)
	return
}

// PutIndex _
func (sc *ShockClient) PutIndex(nodeid string, indexname string) (err error) {
	if indexname == "" {
		return
	}
	sr := new(ShockResponseGeneric)
	err = sc.putRequest("/node/"+nodeid+"/index/"+indexname, nil, &sr)
	if err != nil {
		err = fmt.Errorf("(PutIndex) node=%s, index=%s: %s", nodeid, indexname, err.Error())
		return
	}
	if len(sr.Errs) > 0 {
		err = fmt.Errorf("(PutIndex) node=%s, index=%s: %s", nodeid, indexname, strings.Join(sr.Errs, ","))
	}
	return
}

// PutIndexQuery _
func (sc *ShockClient) PutIndexQuery(nodeid string, indexname string, force bool, column int) (err error) {
	if indexname == "" {
		return
	}
	query := url.Values{}
	if force {
		query.Set("force_rebuild", "1")
	}
	if column > 0 {
		query.Set("number", strconv.Itoa(column))
	}
	sr := new(ShockResponseGeneric)
	err = sc.putRequest("/node/"+nodeid+"/index/"+indexname, query, &sr)
	if err != nil {
		err = fmt.Errorf("(PutIndexQuery) node=%s, index=%s: %s", nodeid, indexname, err.Error())
		return
	}
	if len(sr.Errs) > 0 {
		err = fmt.Errorf("(PutIndexQuery) node=%s, index=%s: %s", nodeid, indexname, strings.Join(sr.Errs, ","))
	}
	return
}

// PutAcl _
func (sc *ShockClient) PutAcl(nodeid string, acltype string, username string) (err error) {
	if (acltype == "") || (username == "") {
		return
	}
	query := url.Values{}
	query.Set("users", username)

	sr := new(ShockResponseGeneric)
	err = sc.putRequest("/node/"+nodeid+"/acl/"+acltype, query, &sr)
	if err != nil {
		err = fmt.Errorf("(PutAcl) node=%s, acl=%s, user=%s: %s", nodeid, acltype, username, err.Error())
		return
	}

	if len(sr.Errs) > 0 {
		err = fmt.Errorf("(PutAcl) node=%s, acl=%s, user=%s: %s", nodeid, acltype, username, strings.Join(sr.Errs, ","))
	}
	return
}

// DeleteAcl _
func (sc *ShockClient) DeleteAcl(nodeid string, acltype string, username string) (err error) {
	if acltype == "" {
		return
	}
	query := url.Values{}
	if !strings.HasPrefix(acltype, "public") {
		query.Set("users", username)
	}

	sr := new(ShockResponseGeneric)
	err = sc.deleteRequest("/node/"+nodeid+"/acl/"+acltype, query, &sr)
	if err != nil {
		err = fmt.Errorf("(DeleteAcl) node=%s, acl=%s, user=%s: %s", nodeid, acltype, username, err.Error())
		return
	}

	if len(sr.Errs) > 0 {
		err = fmt.Errorf("(DeleteAcl) node=%s, acl=%s, user=%s: %s", nodeid, acltype, username, strings.Join(sr.Errs, ","))
	}
	return
}

// GetAcl _
func (sc *ShockClient) GetAcl(nodeid string) (sr *ShockResponseGeneric, err error) {
	sr = new(ShockResponseGeneric)
	err = sc.getRequest("/node/"+nodeid+"/acl", nil, &sr)
	if err != nil {
		err = fmt.Errorf("(GetAcl) node=%s: %s", nodeid, err.Error())
		return
	}

	if len(sr.Errs) > 0 {
		err = fmt.Errorf("(GetAcl) node=%s: %s", nodeid, strings.Join(sr.Errs, ","))
	}
	return
}

// MakePublic _
func (sc *ShockClient) MakePublic(nodeid string) (err error) {
	sr := new(ShockResponseGeneric)
	err = sc.putRequest("/node/"+nodeid+"/acl/public_read", nil, &sr)
	if err != nil {
		err = fmt.Errorf("(MakePublic) node=%s: %s", nodeid, err.Error())
		return
	}

	if len(sr.Errs) > 0 {
		err = fmt.Errorf("(MakePublic) node=%s: %s", nodeid, strings.Join(sr.Errs, ","))
	}
	return
}

// ChownNode _
func (sc *ShockClient) ChownNode(nodeid string, username string) (err error) {
	query := url.Values{}
	query.Set("users", username)

	sr := new(ShockResponseGeneric)
	err = sc.putRequest("/node/"+nodeid+"/acl/owner", query, &sr)
	if err != nil {
		err = fmt.Errorf("(ChownNode) node=%s: %s", nodeid, err.Error())
		return
	}

	if len(sr.Errs) > 0 {
		err = fmt.Errorf("(ChownNode) node=%s: %s", nodeid, strings.Join(sr.Errs, ","))
	}
	return
}

// GetNodeDownloadUrl _
func (sc *ShockClient) GetNodeDownloadUrl(node ShockNode) (downloadUrl string, err error) {
	var myurl *url.URL
	myurl, err = url.ParseRequestURI(sc.Host)
	if err != nil {
		return
	}
	(*myurl).Path = fmt.Sprint("node/", node.Id)
	(*myurl).RawQuery = "download"
	downloadUrl = myurl.String()
	return
}

// Query _
func (sc *ShockClient) Query(query url.Values) (sr *ShockQueryResponse, err error) {
	query.Set("query", "")
	sr, err = sc.nodeQuery(query)
	return
}

// QueryFull _
func (sc *ShockClient) QueryFull(query url.Values) (sr *ShockQueryResponse, err error) {
	query.Set("querynode", "")
	sr, err = sc.nodeQuery(query)
	return
}

// QueryDistinct _
func (sc *ShockClient) QueryDistinct(query url.Values) (sr *ShockResponseGeneric, err error) {
	sr = new(ShockResponseGeneric)
	err = sc.getRequest("/node", query, &sr)
	if err != nil {
		err = fmt.Errorf("(QueryDistinct) %s", err.Error())
		return
	}
	if len(sr.Errs) > 0 {
		err = fmt.Errorf("(QueryDistinct) %s", strings.Join(sr.Errs, ","))
	}
	return
}

// nodeQuery
func (sc *ShockClient) nodeQuery(query url.Values) (sr *ShockQueryResponse, err error) {
	sr = new(ShockQueryResponse)
	err = sc.getRequest("/node", query, &sr)
	if err != nil {
		err = fmt.Errorf("(Query) %s", err.Error())
		return
	}
	if len(sr.Errs) > 0 {
		err = fmt.Errorf("(Query) %s", strings.Join(sr.Errs, ","))
	}
	return
}

// QueryPaginated is not supported with the current vendored httpclient.
// The httpclient.RestClient type is not available in this version.
func (sc *ShockClient) QueryPaginated(resource string, query url.Values, limit int, offset int) (err error) {
	err = fmt.Errorf("(QueryPaginated) not supported: httpclient.RestClient not available")
	return
}

// GetNode _
func (sc *ShockClient) GetNode(nodeid string) (node *ShockNode, err error) {
	sr := new(ShockResponse)
	err = sc.getRequest("/node/"+nodeid, nil, &sr)
	if err != nil {
		err = fmt.Errorf("(GetNode) node=%s: %s", nodeid, err.Error())
		return
	}

	if len(sr.Errs) > 0 {
		err = fmt.Errorf("(GetNode) node=%s: %s", nodeid, strings.Join(sr.Errs, ","))
		return
	}

	node = sr.Data
	if node == nil {
		err = fmt.Errorf("(GetNode) node=%s: empty node returned from Shock", nodeid)
	}
	return
}

// DeleteNode _
func (sc *ShockClient) DeleteNode(nodeid string) (err error) {
	sr := new(ShockResponseGeneric)
	err = sc.deleteRequest("/node/"+nodeid, nil, &sr)
	if err != nil {
		err = fmt.Errorf("(DeleteNode) node=%s: %s", nodeid, err.Error())
		return
	}

	if len(sr.Errs) > 0 {
		err = fmt.Errorf("(DeleteNode) node=%s: %s", nodeid, strings.Join(sr.Errs, ","))
	}
	return
}

// old-style functions that probably should be refactored
// ShockGet _
func ShockGet(host string, nodeid string, token string) (node *ShockNode, err error) {
	if host == "" || nodeid == "" {
		err = errors.New("empty shock host or node id")
		return
	}
	sc := ShockClient{Host: host, Token: token}

	node, err = sc.GetNode(nodeid)
	return
}

// ShockDelete _
func ShockDelete(host string, nodeid string, token string) (err error) {
	if host == "" || nodeid == "" {
		return errors.New("empty shock host or node id")
	}
	sc := ShockClient{Host: host, Token: token}

	err = sc.DeleteNode(nodeid)
	return
}

// FetchFile fetch file by shock url
func FetchFile(filename string, url string, token string, uncompress string, computeMD5 bool) (size int64, md5sum string, err error) {
	var localfile *os.File
	localfile, err = os.Create(filename)
	if err != nil {
		return
	}
	defer localfile.Close()

	var body io.ReadCloser
	body, err = FetchShockStream(url, token)
	if err != nil {
		err = errors.New("(FetchFile) " + err.Error())
		return
	}
	defer body.Close()

	// set md5 compute
	md5h := md5.New()

	if uncompress == "" {
		// split stream to file and md5
		var dst io.Writer
		if computeMD5 {
			dst = io.MultiWriter(localfile, md5h)
		} else {
			dst = localfile
		}
		size, err = io.Copy(dst, body)
		if err != nil {
			return
		}
	} else if uncompress == "gzip" {
		// split stream to gzip and md5
		var input io.ReadCloser
		if computeMD5 {
			pReader, pWriter := io.Pipe()
			defer pReader.Close()
			dst := io.MultiWriter(pWriter, md5h)
			go func() {
				io.Copy(dst, body)
				pWriter.Close()
			}()
			input = pReader
		} else {
			input = body
		}

		gr, gerr := gzip.NewReader(input)
		if gerr != nil {
			err = gerr
			return
		}
		defer gr.Close()
		size, err = io.Copy(localfile, gr)
		if err != nil {
			return
		}
	} else {
		err = errors.New("(FetchFile) uncompress method unknown: " + uncompress)
		return
	}

	if computeMD5 {
		md5sum = fmt.Sprintf("%x", md5h.Sum(nil))
	}
	return
}

// FetchShockStream _
func FetchShockStream(url string, token string) (r io.ReadCloser, err error) {
	var user *httpclient.Auth
	if token != "" {
		user = httpclient.GetUserByTokenAuth(token)
	}

	//download file from Shock
	var res *http.Response
	res, err = httpclient.Get(url, httpclient.Header{}, nil, user)
	if err != nil {
		err = errors.New("(FetchShockStream) httpclient.Get returned: " + err.Error())
		return
	}

	if res.StatusCode != 200 { //err in fetching data
		resbody, _ := ioutil.ReadAll(res.Body)
		err = fmt.Errorf("(FetchShockStream) url=%s, res=%s", url, resbody)
		return
	}

	r = res.Body
	return
}

// CopyFile _
// source:  http://stackoverflow.com/a/22259280
// TODO this is not shock related, need another package
func CopyFile(src, dst string) (size int64, err error) {
	var srcFile *os.File
	srcFile, err = os.Open(src)
	if err != nil {
		return
	}
	defer srcFile.Close()

	var srcFileStat os.FileInfo
	srcFileStat, err = srcFile.Stat()
	if err != nil {
		return
	}

	if !srcFileStat.Mode().IsRegular() {
		err = fmt.Errorf("%s is not a regular file", src)
		return
	}

	var dstFile *os.File
	dstFile, err = os.Create(dst)
	if err != nil {
		return
	}
	defer dstFile.Close()
	size, err = io.Copy(dstFile, srcFile)
	return
}

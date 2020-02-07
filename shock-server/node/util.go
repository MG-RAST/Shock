package node

import (
	"archive/zip"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"google.golang.org/api/option"

	"cloud.google.com/go/storage"

	"github.com/Azure/azure-storage-blob-go/azblob"

	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/logger"
)

type mappy map[string]bool

// TransitMap store UUID and bool value if object is in transit
//var TransitMap map[string]bool

// TransitMapLock lock write access to the CacheMap
var TransitMapMutex = sync.RWMutex{}

func IsInMappy(item string, mp mappy) bool {
	if _, ok := mp[item]; ok {
		return true
	}
	return false
}

var virtIdx = mappy{"size": true}

type sortBytes []byte

func (b sortBytes) Less(i, j int) bool {
	return b[i] < b[j]
}

func (b sortBytes) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b sortBytes) Len() int {
	return len(b)
}

func SortByteArray(b []byte) []byte {
	sb := make([]byte, len(b))
	copy(sb, b)
	sort.Sort(sortBytes(sb))
	return sb
}

//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************

// func Open(name string) (*File, error)
//  Open opens the named file for reading from local disk
//  Open file on list of remote stores

// FMOpen drop in replacement for os.Open that attempts to download files from locations stored in MongoDB
func FMOpen(filepath string) (f *os.File, err error) {

	// try to read file from disk
	f, err = os.Open(filepath) // this will also open a sym link from the cache location

	if err == nil {
		return
	}

	// extract UUID from path(should be path2UUID function really)
	ext := path.Ext(filepath)                     // identify extension
	filename := strings.TrimSuffix(filepath, ext) // find filename
	uuid := path.Base(filename)                   // implement basename cmd

	var nodeInstance, _ = Load(uuid)

	// lock access to map
	for true {
		TransitMapMutex.Lock()
		if !nodeInstance.CheckTransit() {
			break
		}
		TransitMapMutex.Unlock()
		time.Sleep(1 * time.Second)
	}
	nodeInstance.SetInTransit()
	TransitMapMutex.Unlock()

	defer nodeInstance.UnSetInTransitLocked() // this is locked internally

	// create the directory infrastructure for node and index
	err = nodeInstance.Mkdir()
	if err != nil {
		return
	}

	success := false
LocationLoop:
	for _, location := range nodeInstance.Locations {

		locationConfig, ok := conf.LocationsMap[location.ID]
		if !ok {
			err = fmt.Errorf("(FMOpen) location unknown %s", location.ID)
			return
		}

		var md5sum string
		begin := time.Now()

		// below are the location types that will allow streaming download, there are also batch
		// download location types that are handled by external scripts
		switch locationConfig.Type {
		case "S3":
			err, md5sum = S3Download(uuid, nodeInstance, locationConfig)

		case "Azure":
			err, md5sum = AzureDownload(uuid, nodeInstance, locationConfig)

		case "Shock":
			// this should be expanded to handle Shock servers sharing the same Mongo instance
			err, md5sum = ShockDownload(uuid, nodeInstance, locationConfig)

		// case "IRods"
		// 	err, md5sum = IRodsDownload(uuid, nodeInstance, locationConfig)

		case "Daos":
			// this should call a DAOS downloader
			err, md5sum = DaosDownload(uuid, nodeInstance)

		default:
			err = fmt.Errorf("(FMOpen) Location type %s not supported", locationConfig.Type)
			logger.Errorf("(FMOpen) Location type %s not supported", locationConfig.Type)
			err = nil
			continue
		}

		// catch broken download
		if err != nil {
			err = fmt.Errorf("(FMOpen) %s download returned: %s", locationConfig.Type, err.Error())
			logger.Errorf("(FMOpen) %s download returned: %s", locationConfig.Type, err.Error())
			err = nil
			continue
		}

		nodeMd5, ok := nodeInstance.File.Checksum["md5"]
		if !ok { // if the node has no MD5 we cannot compare and no download will work, needs to be fixed
			err = fmt.Errorf("(FMOpen) node %s has no MD5", nodeInstance.Id)
			logger.Errorf("%s", err.Error())
			return
		}

		if md5sum != nodeMd5 {
			logger.Errorf("(FMOpen) md5sum=%s AND nodeMd5= %s", md5sum, nodeMd5)
			logger.Errorf("(FMOpen) %s download returned: %s", locationConfig.Type, err.Error())
			continue LocationLoop
		}

		success = true
		duration := time.Now().Sub(begin)
		logger.Infof("(FMOpen) %s downloaded, UUID: %s, duration: %d, size:%d", locationConfig.Type, uuid, int(duration.Seconds()), int(nodeInstance.File.Size))
		// exit the loop
		break LocationLoop

	} // of for location loop

	// error report in case of failure
	if !success {

		loclist := ""
		for _, loc := range nodeInstance.Locations {
			loclist += loc.ID + ","
		}

		err = fmt.Errorf("(FMOpen) Object (%s) not found in any location [%s]", uuid, loclist)

		if err != nil {
			// debug output
			err = fmt.Errorf("(FMOpen) returned: %s", err.Error())
			return
		}
	}

	// create file handle for newly downloaded file on local disk
	// we use the symlink we have created here
	f, err = os.Open(filepath)

	// check if we encounter an error
	if err != nil {
		err = fmt.Errorf("(FMOpen) error opening file %s, after download d: %s", filepath, err.Error())
		return
	}
	return
}

//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************

// S3Download download a file and its indices from an S3 source using an external boto3 python script
func S3Download(uuid string, nodeInstance *Node, location *conf.LocationConfig) (err error, md5sum string) {
	functionName := "S3Download"

	itemkey := fmt.Sprintf("%s.data", uuid)
	indexfile := fmt.Sprintf("%s.idx.zip", uuid) // the zipped contents of the idx directory in S3

	//fmt.Printf("(%s) downloading node: %s \n", functionName, uuid)

	tmpfile, err := ioutil.TempFile(conf.PATH_CACHE, "")
	if err != nil {
		log.Fatalf("(%s) cannot create temporary file: %s [Err: %s]", functionName, uuid, err.Error())
		return
	}
	tmpfile.Close()

	baseArgString := fmt.Sprintf("boto-s3-download.py --bucket=%s --region=%s --tmpfile=%s --s3endpoint=%s",
		location.Bucket, location.Region, tmpfile.Name(), location.URL)
	argString := fmt.Sprintf("%s --object=%s",
		baseArgString, itemkey)
	args := strings.Fields(argString)
	cmd := exec.Command(args[0], args[1:]...)

	// passing parameters to external program via ENV variables, see https://boto3.amazonaws.com/v1/documentation/api/latest/guide/configuration.html
	// this is more secure than cmd-line
	envAuth := fmt.Sprintf("AWS_ACCESS_KEY_ID = %s", location.AuthKey)
	envSecret := fmt.Sprintf("AWS_SECRET_ACCESS_KEY = %s", location.SecretKey)
	newEnv := append(os.Environ(), envAuth, envSecret)
	cmd.Env = newEnv

	// run and capture the output
	out, err := cmd.Output()
	if err != nil {
		logger.Debug(1, "(%s) cmd.Run(%s) failed with %s\n", functionName, cmd, err)
		//fmt.Printf("(%s) cmd.Run(%s) failed with %s\n", functionName, cmd, err.Error())
		return
	}

	//md5sum = fmt.Sprintf("%s", string(out))

	md5sum = string(out)
	md5sum = strings.TrimRight(md5sum, "\r\n")
	//fmt.Printf("node: %s DONE \n", itemkey)

	// move the bits into place
	err = handleDataFile(tmpfile.Name(), uuid, functionName)
	if err != nil {
		logger.Debug(3, "(%s) error moving directory structure and symkink into place for : %s [Err: %s]", functionName, uuid, err.Error())
		return
	}

	// ##############################################################################
	// ##############################################################################
	// ##############################################################################
	//index bits now
	tmpfile, err = ioutil.TempFile(conf.PATH_CACHE, "")
	if err != nil {
		log.Fatalf("(%s) cannot create temporary file: %s [Err: %s]", functionName, uuid, err.Error())
		return
	}
	tmpfile.Close()

	argString = fmt.Sprintf("%s --object=%s", baseArgString, indexfile)
	args = strings.Fields(argString)
	cmd = exec.Command(args[0], args[1:]...)

	// passing parameters to external program via ENV variables, see https://boto3.amazonaws.com/v1/documentation/api/latest/guide/configuration.html
	// this is more secure than cmd-line
	cmd.Env = newEnv

	// run and capture the output
	out, err = cmd.Output()
	if err != nil {
		logger.Debug(1, "(%s) cmd.Run(%s) failed with %s\n", functionName, cmd, err)
		fmt.Printf("(%s) cmd.Run(%s) failed with %s\n", functionName, cmd, err.Error())
		err = nil
		return
	}

	//logger.Infof("Downloaded: %s (%d Bytes) \n", file.Name(), numBytes)
	err = handleIdxZipFile(tmpfile, uuid, "S3Download")
	if err != nil {
		logger.Debug(3, "(S3Download) error moving index directory structure and symkink into place for : %s [Err: %s]", uuid, err.Error())
		return
	}

	return
}

//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************

//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************

// AzureDownload support for downloading off https://github.com/daos-stack
func AzureDownload(uuid string, nodeInstance *Node, location *conf.LocationConfig) (err error, md5sum string) {

	itemkey := fmt.Sprintf("%s.data", uuid)
	indexfile := fmt.Sprintf("%s.idx.zip", uuid) // the zipped contents of the idx directory in S3

	tmpfile, err := ioutil.TempFile(conf.PATH_CACHE, "")
	if err != nil {
		log.Fatalf("(GCloudStoreDownload)  cannot create temporary file: %s [Err: %s]", uuid, err.Error())
		return
	}
	defer tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	// Create a default request pipeline using your storage account name and account key.
	credential, err := azblob.NewSharedKeyCredential(location.Account, location.SecretKey)
	if err != nil {
		logger.Debug(3, "(AzureDownload) error authenticating account: %s [Err: %s]", location.Account, err.Error())
		return
	}

	// Azure specific bits
	pipeline := azblob.NewPipeline(credential, azblob.PipelineOptions{})
	// create context
	ctx := context.Background() // This example uses a never-expiring context

	// From the Azure portal, get your storage account blob service URL endpoint.
	myURL, err := url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", location.Account, location.Container, itemkey))
	if err != nil {
		logger.Debug(3, "(AzureDownload) URL malformed: %s [Err: %s]", myURL, err.Error())
		return
	}

	// Create a ServiceURL for our node
	blobURL := azblob.NewBlobURL(*myURL, pipeline)

	// download the file contents
	err = azblob.DownloadBlobToFile(ctx, blobURL, 0, azblob.CountToEnd, tmpfile, azblob.DownloadFromBlobOptions{
		BlockSize: 4 * 1024 * 1024, Parallelism: 16})
	if err != nil {
		logger.Debug(3, "(AzureDownload) error downloading blob: %s [Err: %s]", uuid, err.Error())
		return
	}

	var dst io.Writer
	md5h := md5.New()
	dst = md5h

	_, err = io.Copy(dst, tmpfile)
	if err != nil {
		// md5 checksum creation did not work
		return
	}

	md5sum = fmt.Sprintf("%x", md5h.Sum(nil))

	err = handleDataFile(tmpfile.Name(), uuid, "AzureDownload")
	if err != nil {
		logger.Debug(3, "(AzureDownload) error moving directory structure and symkink into place for : %s [Err: %s]", uuid, err.Error())
		return
	}
	tmpfile.Close()

	// index bits now
	tmpfile, err = ioutil.TempFile(conf.PATH_CACHE, "")
	if err != nil {
		log.Fatalf("(GCloudStoreDownload) cannot create temporary file: %s [Err: %s]", uuid, err.Error())
		return
	}
	defer tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	// From the Azure portal, get your storage account blob service URL endpoint.
	myURL, _ = url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", location.Account, location.Container, indexfile))

	// Create a ServiceURL for our node
	blobURL = azblob.NewBlobURL(*myURL, pipeline)

	// download the file contents
	err = azblob.DownloadBlobToFile(ctx, blobURL, 0, azblob.CountToEnd, tmpfile, azblob.DownloadFromBlobOptions{
		BlockSize: 4 * 1024 * 1024, Parallelism: 16})

	if err != nil {
		logger.Debug(3, "(AzureDownload) error downloading blob: %s [Err: %s]", uuid, err.Error())
		return
	}

	err = handleDataFile(tmpfile.Name(), uuid, "AzureDownload")
	if err != nil {
		logger.Debug(3, "(AzureDownload) error moving index directory structure and symkink into place for : %s [Err: %s]", uuid, err.Error())
		return
	}
	tmpfile.Close()
	return
}

//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************

// GCloudStoreDownload support for downloading off https://github.com/daos-stack
func GCloudStoreDownload(uuid string, nodeInstance *Node, location *conf.LocationConfig) (err error, md5sum string) {

	itemkey := fmt.Sprintf("%s.data", uuid)
	indexfile := fmt.Sprintf("%s.idx.zip", uuid) // the zipped contents of the idx directory in S3

	tmpfile, err := ioutil.TempFile(conf.PATH_CACHE, "")
	if err != nil {
		log.Fatalf("(GCloudStoreDownload)  cannot create temporary file: %s [Err: %s]", uuid, err.Error())
		return
	}
	defer tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	// GCS specific part
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithAPIKey(location.SecretKey))
	if err != nil {
		log.Fatalf("(GCloudStoreDownload)Failed to create GCloud client: %s", err)
		return
	}

	//  a Bucket handle
	bucket := client.Bucket(location.Bucket)

	//	a handle for our file
	obj := bucket.Object(itemkey)

	// read an object
	reader, err := obj.NewReader(ctx)
	if err != nil {
		log.Fatalf("(GCloudStoreDownload)  blob not found: %s [Err: %s]", uuid, err.Error())
		return
	}
	defer reader.Close()

	var dst io.Writer
	md5h := md5.New()
	dst = io.MultiWriter(tmpfile, md5h)

	_, err = io.Copy(dst, reader)
	if err != nil {
		logger.Debug(3, "(GCloudStoreDownload) error downloading blob: %s [Err: %s]", uuid, err.Error())
		return
	}
	// end GCS specific
	md5sum = fmt.Sprintf("%x", md5h.Sum(nil))

	err = handleDataFile(tmpfile.Name(), uuid, "GCloudStoreDownload")
	if err != nil {
		logger.Debug(3, "(GCloudStoreDownload) error moving directory structure and symkink into place for : %s [Err: %s]", uuid, err.Error())
		return
	}

	tmpfile.Close()

	// download index files as well

	tmpfile, err = ioutil.TempFile(conf.PATH_CACHE, "")
	if err != nil {
		log.Fatalf("(GCloudStoreDownload) cannot create temporary file: %s [Err: %s]", uuid, err.Error())
		return
	}
	defer tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	//	a handle for our file
	obj = bucket.Object(indexfile)

	// read an object
	reader, err = obj.NewReader(ctx)
	if err != nil {
		log.Fatalf("(GCloudStoreDownload) blob not found: %s [Err: %s]", indexfile, err.Error())
	}
	defer reader.Close()
	_, err = io.Copy(tmpfile, reader)
	if err != nil {
		logger.Debug(3, "(GCloudStoreDownload) error downloading blob: %s [Err: %s]", indexfile, err.Error())
		return
	}

	err = handleIdxZipFile(tmpfile, uuid, "GCloudStoreDownload")
	if err != nil {
		logger.Debug(3, "(GCloudStoreDownload) error moving index structures and symkink into place for : %s [Err: %s]", uuid, err.Error())
		return
	}
	return
}

//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************

//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************

// // IRodsDownload download from IRods
//  // https://github.com/jjacquay712/GoRODS/blob/master/HOWTO.md
// func IRodsDownload(uuid string, nodeInstance *Node, location *conf.LocationConfig) (err error, md5sum string) {

// 	itemkey := fmt.Sprintf("%s.data", uuid)
// 	indexfile := fmt.Sprintf("%s.idx.zip", uuid) // the zipped contents of the idx directory in S3

// 	tmpfile, err := ioutil.TempFile(conf.PATH_CACHE, "")
// 	if err != nil {
// 		log.Fatalf("(IRodsDownload)  cannot create temporary file: %s [Err: %s]", uuid, err.Error())
// 		return
// 	}
// 	defer tmpfile.Close()
// 	defer os.Remove(tmpfile.Name())

// 	// irods connection init
// 	client, err := gorods.New(gorods.ConnectionOptions{
// 		Type: gorods.UserDefined,

// 		Host: location.Hostname,
// 		Port: location.Port,
// 		Zone: location.Zone,

// 		Username: location.User,
// 		Password: location.Password,
// 	})

// 	// Ensure the client initialized successfully and connected to the iCAT server
// 	if err != nil {
// 		log.Fatalf("(IRodsDownload) cannot init connection: %s [Err: %s]", uuid, err.Error())
// 		return
// 	}

// 	// the paths for the objects in iRods include the Zone
// 	itempath := filepath.Join(location.Zone, itemkey)
// 	indexpath := filepath.Join(location.Zone, indexfile)

// 	//irods file retrieval
// 	// Open a data object reference
// 	err = client.OpenDataObject(itempath, func(myFile *gorods.DataObj, con *gorods.Connection)

// 	if err != nil {
// 		log.Fatalf("(IRodsDownload) cannot find iRODs object: %s [Err: %s]", uuid, err.Error())
// 		return
// 	}

// 	var dst io.Writer
// 	md5h := md5.New()
// 	dst = io.MultiWriter(tmpfile, md5h)

// 	//
// 	outBuff := make(chan *ByteArr, 100)

// 	go func() {
// 		err := obj.ReadChunkFree(10240000, func(chunk *ByteArr) {outBuff <- chunk } )

// 		if err != nil {
// 			log.Fatalf("(IRodsDownload) cannot find iRODs object: %s [Err: %s]", uuid, err.Error())
// 			return
// 		}

// 		close(outBuff)
// 	}()

// 	// write the contents of the buffer
// 	for b := range outBuff {
// 		dst.Write(b.Contents)
// 		b.Free()
// 	}

// 	// end iRODS specific
// 	md5sum = fmt.Sprintf("%x", md5h.Sum(nil))

// 	err = handleDataFile(tmpfile, uuid, "IRodsDownload")
// 	if err != nil {
// 		logger.Debug(3, "(IRodsDownload) error moving directory structure and symkink into place for : %s [Err: %s]", uuid, err.Error())
// 		return
// 	}

// 	tmpfile.Close()

// 	// download index files as well

// 	tmpfile, err = ioutil.TempFile(conf.PATH_CACHE, "")
// 	if err != nil {
// 		log.Fatalf("(IRodsDownload) cannot create temporary file: %s [Err: %s]", uuid, err.Error())
// 		return
// 	}
// 	defer tmpfile.Close()
// 	defer os.Remove(tmpfile.Name())

// 	//irods file retrieval
// 	// Open a data object reference
// 	err = client.OpenDataObject(indexpath, func(myFile *gorods.DataObj, con *gorods.Connection)

// 	if err != nil {
// 		log.Fatalf("(IRodsDownload) cannot find iRODs object: %s [Err: %s]", uuid, err.Error())
// 		return
// 	}

// 	var dst io.Writer
// 	md5h := md5.New()
// 	dst = io.MultiWriter(tmpfile, md5h)

// 	//
// 	outBuff := make(chan *ByteArr, 100)

// 	go func() {
// 		err := obj.ReadChunkFree(10240000, func(chunk *ByteArr) {outBuff <- chunk } )

// 		if err != nil {
// 			log.Fatalf("(IRodsDownload) cannot find iRODs object: %s [Err: %s]", uuid, err.Error())
// 			return
// 		}

// 		close(outBuff)
// 	}()

// 	// write the contents of the buffer
// 	for b := range outBuff {
// 		dst.Write(b.Contents)
// 		b.Free()
// 	}

// 	}

// 	err = handleIdxZipFile(tmpfile, uuid, "IRodsDownload")
// 	if err != nil {
// 		logger.Debug(3, "(IRodsDownload) error moving index structures and symkink into place for : %s [Err: %s]", uuid, err.Error())
// 		return
// 	}
// 	return
// }

//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************

// DaosDownload support for downloading off https://github.com/daos-stack
func DaosDownload(uuid string, nodeInstance *Node) (err error, md5sum string) {
	logger.Infof("(S3Download--> DAOS ) needs to be implemented !! \n")

	return
}

//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************

// ShockDownload download a file from a Shock server
func ShockDownload(uuid string, nodeInstance *Node, location *conf.LocationConfig) (err error, md5sum string) {

	itemkey := fmt.Sprintf("%s.data", uuid)
	indexfile := fmt.Sprintf("%s.idx.zip", uuid) // the zipped contents of the idx directory in S3

	tmpfile, err := ioutil.TempFile(conf.PATH_CACHE, "")
	if err != nil {
		log.Fatalf("(GCloudStoreDownload)  cannot create temporary file: %s [Err: %s]", uuid, err.Error())
		return
	}
	defer tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	var dst io.Writer
	md5h := md5.New()
	dst = io.MultiWriter(tmpfile, md5h)

	// a transport helps with proxies, TLS configuration, keep-alives, compression, and other settings
	transport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}
	client := &http.Client{Transport: transport}

	// we expect the form "authorization: mgrast 12345678A123456789012345" the auth has to make sense for the remote Shock instance
	authkey := fmt.Sprintf("%s", location.AuthKey)

	url := fmt.Sprintf("%s/%s?download", location.URL, uuid)

	// create a request to enable adding auth to header

	request, err := http.NewRequest("GET", url, nil)
	// ...
	request.Header.Add("authorization", authkey)
	resp, err := client.Do(request)

	// For control over proxies, TLS configuration, keep-alives, compression, and other settings, create a Transport:

	// tr := &http.Transport{
	// 	MaxIdleConns:       10,
	// 	IdleConnTimeout:    30 * time.Second,
	// 	DisableCompression: true,
	// }
	// client := &http.Client{Transport: tr}
	// resp, err := client.Get("https://example.com")

	// Get the data

	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", resp.Status)
		return
	}

	// Writer the body to file
	_, err = io.Copy(dst, resp.Body)
	if err != nil {
		log.Fatalf("(ShockDownload) Unable to download item %q for %s, %v", itemkey, uuid, err)
		return
	}

	md5sum = fmt.Sprintf("%x", md5h.Sum(nil))

	err = handleDataFile(tmpfile.Name(), uuid, "ShockDownload")
	if err != nil {
		logger.Debug(3, "(ShockDownload) error moving directory structure and symkink into place for : %s [Err: %s]", uuid, err.Error())
		return
	}
	// end of SHOCK specific part

	// now download the indices
	tmpfile.Close()

	tmpfile, err = ioutil.TempFile(conf.PATH_CACHE, "")
	if err != nil {
		log.Fatalf("(GCloudStoreDownload) cannot create temporary file: %s [Err: %s]", uuid, err.Error())
		return
	}
	defer tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	url = fmt.Sprintf("%s/%s.?download", location.URL, indexfile)

	// Get the data
	resp, err = http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", resp.Status)
		return
	}

	// Writer the body to file
	_, err = io.Copy(tmpfile, resp.Body)
	if err != nil {
		logger.Debug(1, "(ShockDownload) Unable to download item %q for %s, %v", indexfile, indexfile, err)
		return
	}
	// unzip the index file
	err = handleIdxZipFile(tmpfile, uuid, "ShockDownload")
	if err != nil {
		logger.Debug(3, "(ShockDownload) error moving index directory structures and symkink into place for : %s [Err: %s]", uuid, err.Error())
		return
	}

	return
}

// Unzip will decompress a zip archive, moving all files and folders q
// within the zip file (parameter 1) to an output directory (parameter 2).
func Unzip(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

// handleIdxZip handle <uuid>.idx.zip files
// accept a file handle for the data file,
// create the directory infrastructure in Cache and Data and create symlinks
func handleIdxZipFile(fp *os.File, uuid string, funcName string) (err error) {

	// download the zipped archive with all the indexed
	indexfile := fmt.Sprintf("%s.idx.zip", uuid) // the zipped contents of the idx directory in S3
	cacheindexpath := uuid2CachePath(uuid)
	indexpath := uuid2Path(uuid)

	indextemppath := fmt.Sprintf("%s/%s", conf.PATH_LOCAL, indexfile) // use the PATH_LOCAL to configure tmp dir

	file, err := os.Create(indextemppath)
	if err != nil {
		logger.Infof("(%s) attempting create index temp dir: %s FAILED", funcName, indextemppath)
		return
	}
	defer file.Close()

	if _, err := io.Copy(file, fp); err != nil {
		logger.Debug(3, "(%s) error coipying file: %s [Err: %s]", funcName, indexfile, err.Error())
	}
	file.Close()

	_, err = Unzip(indexfile, cacheindexpath) // unzip into idx folder
	if err != nil {
		// debug output
		err = fmt.Errorf("(%s) error decompressing d: %s", err.Error())
		return
	}
	// remove the zip file
	err = os.Remove(indextemppath)
	if err != nil {
		// debug output
		err = fmt.Errorf("(%s) error removing temp file d: %s", err.Error())
		return
	}

	// add sym link from cacheItemPath to itemPath
	err = os.Symlink(cacheindexpath+"/idx", indexpath+"/idx")
	if err != nil {
		log.Fatalf("(%s) Unable to create symlink from %s to %s, %s", funcName, cacheindexpath, indexpath, err.Error())
		return
	}

	return
}

// handleDataFile handle <uuid>.data files
// accept a file handle for the data file,
// unpack it into the right cache location for uuid
// create the directory infrastructure in Cache and Data and create symlinks
func handleDataFile(filename string, uuid string, funcName string) (err error) {

	//logger.Infof("(FMOpen-> handleDataFile) in ")

	// 4) Download the item from the bucket. If an error occurs, log it and exit. Otherwise, notify the user that the download succeeded.
	// needs to create a full path
	cacheitempath := uuid2CachePath(uuid)
	cacheitemfile := path.Join(cacheitempath, uuid+".data")

	itempath := uuid2Path(uuid)
	itemfile := path.Join(itempath, uuid+".data")

	in, err := os.Open(filename)
	if err != nil {
		log.Fatalf("(%s) Cannot open file %s for reading [%s]", filename, err.Error())
		return
	}
	in.Close()

	//logger.Infof("(FMOpen-> handleDataFile) cacheFile: %s, itemfile: %s", cacheitemfile, itemfile)

	// create cache dir path
	err = os.MkdirAll(cacheitempath, 0777)
	if err != nil {
		log.Fatalf("(%s) Unable to create cache path for item %s [%s], %s", funcName, cacheitemfile, cacheitempath, err.Error())
		return
	}
	//logger.Infof("(FMOpen-> handleDataFile) created cache path ")

	err = os.MkdirAll(itempath, 0777)
	if err != nil {
		log.Fatalf("(%s) Unable to create path for item %s [%s], %s", funcName, itemfile, itempath, err.Error())
		return
	}

	//logger.Infof("(FMOpen-> handleDataFile) created item path ")

	// create a handle for the cache item here
	err = os.Rename(filename, cacheitemfile) // move the tmpfile into the correct cache path
	if err != nil {
		logger.Infof("(%s) moving tmpfile (%s) to new path (%) failed: %s FAILED", funcName, filename, cacheitemfile)
		return
	}
	//logger.Infof("(FMOpen-> handleDataFile) past create local Cache file for uuid: %s at Path: %s [Err: %s]", funcName, uuid, cacheitempath, err.Error())

	// add sym link from cacheItemPath to itemPath
	err = os.Symlink(cacheitemfile, itemfile)
	if err != nil {
		log.Fatalf("(%s) Unable to create symlink from %s to %s, %s", funcName, cacheitemfile, itemfile, err.Error())
		return
	}
	//logger.Infof("(FMOpen-> handleDataFile) created symlink")

	return
}

// // Transitlock - lock the mutex controlling access to the Transitlock
// func (node *Node) TransitMapLock() {
// 	TransitMapMutex.Lock()
// 	defer TransitMapMutex.Unlock()

// 	TransitMap[node.Id] = true
// 	return
// }

// // TransitUnlock - unlock the mutex controlling access to the Transitlock
// func (node *Node) TransitMapUnlock() {
// 	TransitMapMutex.Lock()
// 	defer TransitMapMutex.Unlock()

// 	delete(TransitMap, node.Id)
// 	return
// }

// CheckTransit - return true if Node is currently being uploaded to an external Location
func (node *Node) CheckTransit() (locked bool) {
	_, locked = conf.TransitMap[node.Id]
	return
}

// SetInTransit - return true if Node is currently being uploaded to an external Location
func (node *Node) SetInTransit() {

	conf.TransitMap[node.Id] = struct{}{}

	return
}

// UnSetInTransitLocked - return true if Node is currently being uploaded to an external Location
func (node *Node) UnSetInTransitLocked() {
	TransitMapMutex.Lock()
	defer TransitMapMutex.Unlock()

	delete(conf.TransitMap, node.Id)
	return
}

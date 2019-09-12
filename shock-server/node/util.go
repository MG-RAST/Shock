package node

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"google.golang.org/api/option"

	"cloud.google.com/go/storage"

	"github.com/Azure/azure-storage-blob-go/azblob"

	"github.com/MG-RAST/Shock/shock-server/cache"
	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/logger"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type mappy map[string]bool

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

	if err != nil {
		return
	}

	// extract UUID from path(should be path2UUID function really)
	ext := path.Ext(filepath)                     // identify extension
	filename := strings.TrimSuffix(filepath, ext) // find filename
	uuid := path.Base(filename)                   // implement basename cmd

	// check if we encounter an error AND if the file is cached (e.g. UUID.cache exists and UUID.data is a symlink)
	if err == nil {
		// update cache LRU info
		cache.Touch(uuid)
		return
	}

	// WE NEED TO LOCK THIS NODE at this point

	var nodeInstance, _ = Load(uuid)

	// create the directory infrastructure for node and index
	err = nodeInstance.Mkdir()
	if err != nil {
		return
	}

	// ideally we should loop over all instances of remotes
	//for _, locationStr := range nodeInstance.Locations {
Loop:
	//for _, locationStr := range []string{"anls3_anlseq"} {
	for _, location := range nodeInstance.Locations {

		locationConfig, ok := conf.LocationsMap[location.ID]
		if !ok {
			err = fmt.Errorf("(FMOpen) location unknown %s", location.ID)
			return
		}
		// debug
		//	spew.Dump(location)

		switch locationConfig.Type {
		// we implement only S3 for now
		case "S3":
			err = S3Download(uuid, nodeInstance, locationConfig)
			if err != nil {
				// debug output
				err = fmt.Errorf("(FMOpen) S3download returned: %s", err.Error())
				return
			}
			break Loop
		case "Azure":
			err = AzureDownload(uuid, nodeInstance, locationConfig)
			if err != nil {
				// debug output
				err = fmt.Errorf("(FMOpen) Azure returned: %s", err.Error())
				return
			}
			break Loop

		case "Shock":
			// this should be expanded to handle Shock servers sharing the same Mongo instance
			err = ShockDownload(uuid, nodeInstance, locationConfig)
			if err != nil {
				// debug output
				err = fmt.Errorf("(FMOpen) ShockDownload returned: %s", err.Error())
				return
			}
			break Loop

		case "Daos":
			// this should call a DAOS downloader
			err = DaosDownload(uuid, nodeInstance)
			if err != nil {
				// debug output
				err = fmt.Errorf("(FMOpen) DaosDownload returned: %s", err.Error())
				return
			}
			break Loop

		default:
			err = fmt.Errorf("(FMOpen) Location type %s not supported", locationConfig.Type)
			return
		}
		// if we are here we did not find what we needed
		err = fmt.Errorf("(FMOpen) Object (%s) not found in any location", uuid)

		if err != nil {
			// debug output
			err = fmt.Errorf("(FMOpen) returned: %s", err.Error())
			return
		}

	}
	// notify the Cache of the new local file
	cache.Add(uuid, nodeInstance.File.Size)
	// WE NEED TO REMOE THE LOCK ON THE NODE...

	// create file handle for newly downloaded file on local disk
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

// S3Download download a file and its indices from an S3 source
func S3Download(uuid string, nodeInstance *Node, location *conf.LocationConfig) (err error) {

	// return error if file not found in S3bucket
	//fmt.Printf("(S3Download) attempting download, UUID: %s, nodeID: %s from: %s\n", uuid, nodeInstance.Id, location.URL)

	Bucket := location.Bucket
	logger.Infof("(S3Download) attempting download, UUID: %s, nodeID: %s, Bucket:%s", uuid, nodeInstance.Id, Bucket)

	// 2) Create an AWS session
	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(location.AuthKey, location.SecretKey, ""),
		Endpoint:    aws.String(location.URL),
		Region:      aws.String(location.Region),

		//DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	}
	sess, err := session.NewSession(s3Config)
	if err != nil {
		logger.Errorf("(S3Download) creating S3 session failed with Endpoint: %s, Region: %s, Bucket: %s, Authkey: %s, SessionKey: %s (err: %s)",
			aws.String(location.URL),
			aws.String("us-east-1"),
			Bucket,
			location.AuthKey,
			location.SecretKey,
			err.Error())
		return
	}

	// 3) Create a new AWS S3 downloader
	downloader := s3manager.NewDownloader(sess)

	// 4) Download the item from the bucket. If an error occurs, log it and exit. Otherwise, notify the user that the download succeeded.
	// needs to create a full path
	cacheitempath := uuid2CachePath(uuid)
	cacheitemfile := fmt.Sprintf("%s/%s.data", cacheitempath, uuid)

	itempath := uuid2Path(uuid)
	itemS3key := fmt.Sprintf("%s.data", uuid)
	itemfile := fmt.Sprintf("%s/%s.data", itempath, uuid)

	// create cache dir path
	err = os.MkdirAll(cacheitempath, 0777)
	if err != nil {
		log.Fatalf("(S3Download) Unable to create cache path for item %s [%s], %s", cacheitemfile, cacheitempath, err.Error())
		return
	}
	err = os.MkdirAll(itempath, 0777)
	if err != nil {
		log.Fatalf("(S3Download) Unable to create path for item %s [%s], %s", itemfile, itempath, err.Error())
		return
	}

	//logger.Infof("(S3Download) attempting download, UUID: %s, itemS3key: %s", uuid, itemS3key)
	// create a cache item here
	file, err := os.Create(cacheitemfile)
	defer file.Close()

	if err != nil {
		return
	}

	_, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(Bucket),
			Key:    aws.String(itemS3key),
		})

	if err != nil {
		log.Fatalf("(S3Download) Unable to download item %q for %s, %s", itemS3key, itemfile, err.Error())
		return
	}

	file.Close()

	// add sym link from cacheItemPath to itemPath
	err = os.Symlink(cacheitemfile, itemfile)
	if err != nil {
		log.Fatalf("(S3Download) Unable to download to create symlink from %s to %s, %s", cacheitemfile, itemfile, err.Error())
		return
	}

	// download the zipped archive with all the indexed
	indexfile := fmt.Sprintf("%s.idx.zip", uuid) // the zipped contents of the idx directory in S3
	cacheindexpath := uuid2CachePath(uuid)
	indexpath := uuid2Path(uuid)

	indextemppath := fmt.Sprintf("%s/%s", conf.PATH_LOCAL, indexfile) // use the PATH_LOCAL to configure tmp dir

	//	logger.Infof("(S3Download) attempting index download (indexfile: %s), cacheindexpath: %s\n", indexfile, cacheindexpath)

	file, err = os.Create(indextemppath)
	if err != nil {
		logger.Infof("(S3Download) attempting create index temp dir: %s FAILED", indextemppath)

		return
	}
	defer file.Close()

	_, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(Bucket),
			Key:    aws.String(indexfile),
		})

	file.Close()

	if err != nil {

		// check if S3 tells us there is no file
		if strings.HasPrefix(err.Error(), "NoSuchKey") {
			// we did not find an index
			//	logger.Infof("no index for %s", uuid)
			return nil // do not report an error
		}
		log.Fatalf("(S3Download) Unable to download item %q for %s, %s", itemS3key, itemfile, err.Error())

		return err
	}
	//logger.Infof("Downloaded: %s (%d Bytes) \n", file.Name(), numBytes)

	// unzip the index file
	_, err = Unzip(indexfile, cacheindexpath) // unzip into idx folder
	if err != nil {
		// debug output
		err = fmt.Errorf("(S3Download) error decompressing d: %s", err.Error())
		return
	}
	// remove the zip file
	err = os.Remove(indextemppath)
	if err != nil {
		// debug output
		err = fmt.Errorf("(S3Download) error removing temp file d: %s", err.Error())
		return
	}

	// add sym link from cacheItemPath to itemPath
	err = os.Symlink(cacheindexpath+"/idx", indexpath+"/idx")
	if err != nil {
		log.Fatalf("(S3Download) Unable to download to create symlink from %s to %s, %s", cacheindexpath, indexpath, err.Error())
		return
	}

	return
}

//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************

// TSMDownload support for downloading files of an existing IBM Tivoli service
func TSMDownload(uuid string, nodeInstance *Node) (err error) {
	logger.Infof("(S3Download--> TSMDownload ) needs to be implemented !! \n")

	// the turn around time here is ~12-24 hours
	// check a dedicated TSMrestore directory in the temp area
	// move file sform there

	// add .data and .idx.zip files to the list of files to be downloaded from TSM

	//

	return
}

//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************

// AzureDownload support for downloading off https://github.com/daos-stack
func AzureDownload(uuid string, nodeInstance *Node, location *conf.LocationConfig) (err error) {

	// 4) Download the item from the bucket. If an error occurs, log it and exit. Otherwise, notify the user that the download succeeded.
	// needs to create a full path
	cacheitempath := uuid2CachePath(uuid)
	cacheitemfile := fmt.Sprintf("%s/%s.data", cacheitempath, uuid)

	itempath := uuid2Path(uuid)
	itemkey := fmt.Sprintf("%s.data", uuid)
	itemfile := fmt.Sprintf("%s/%s.data", itempath, uuid)

	// create cache dir path
	err = os.MkdirAll(cacheitempath, 0777)
	if err != nil {
		log.Fatalf("(AzureDownload) Unable to create cache path for item %s [%s], %s", cacheitemfile, cacheitempath, err.Error())
		return
	}
	err = os.MkdirAll(itempath, 0777)
	if err != nil {
		log.Fatalf("(AzureDownload) Unable to create path for item %s [%s], %s", itemfile, itempath, err.Error())
		return
	}

	// create a cache item here
	file, err := os.Create(cacheitemfile)
	defer file.Close()

	if err != nil {
		logger.Debug(3, "(AzureDownload) cannot create local file handle for: %s, itemkey: %s", cacheitemfile, uuid)
		return
	}

	// Create a default request pipeline using your storage account name and account key.
	credential, err := azblob.NewSharedKeyCredential(location.Account, location.SecretKey)
	if err != nil {
		logger.Debug(3, "(AzureDownload) error authenticating account: %s [Err: %s]", location.Account, err.Error())
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
	err = azblob.DownloadBlobToFile(ctx, blobURL, 0, azblob.CountToEnd, file, azblob.DownloadFromBlobOptions{
		BlockSize: 4 * 1024 * 1024, Parallelism: 16})
	if err != nil {
		logger.Debug(3, "(AzureDownload) error downloading blob: %s [Err: %s]", uuid, err.Error())
		return
	}

	// end Azure specific bits

	file.Close()

	// add sym link from cacheItemPath to itemPath
	err = os.Symlink(cacheitemfile, itemfile)
	if err != nil {
		log.Fatalf("(AzureDownload) Unable to download to create symlink from %s to %s, %s", cacheitemfile, itemfile, err.Error())
		return
	}

	// download index files as well

	// download the zipped archive with all the indexed
	indexfile := fmt.Sprintf("%s.idx.zip", uuid) // the zipped contents of the idx directory in S3
	cacheindexpath := uuid2CachePath(uuid)
	indexpath := uuid2Path(uuid)

	indextemppath := fmt.Sprintf("%s/%s", conf.PATH_LOCAL, indexfile) // use the PATH_LOCAL to configure tmp dir

	//	logger.Infof("(S3Download) attempting index download (indexfile: %s), cacheindexpath: %s\n", indexfile, cacheindexpath)

	file, err = os.Create(indextemppath)
	if err != nil {
		logger.Infof("(AzureDownload) attempting create index temp dir: %s FAILED", indextemppath)

		return
	}
	defer file.Close()

	// From the Azure portal, get your storage account blob service URL endpoint.
	myURL, _ = url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", location.Account, location.Container, indexfile))

	// Create a ServiceURL for our node
	blobURL = azblob.NewBlobURL(*myURL, pipeline)

	// download the file contents
	err = azblob.DownloadBlobToFile(ctx, blobURL, 0, azblob.CountToEnd, file, azblob.DownloadFromBlobOptions{
		BlockSize: 4 * 1024 * 1024, Parallelism: 16})

	if err != nil {
		logger.Debug(3, "(AzureDownload) error downloading blob: %s [Err: %s]", uuid, err.Error())
	}

	file.Close()

	// unzip the index file
	_, err = Unzip(indexfile, cacheindexpath) // unzip into idx folder
	if err != nil {
		// debug output
		err = fmt.Errorf("(AzureDownload) error decompressing d: %s", err.Error())
		return
	}
	// remove the zip file
	err = os.Remove(indextemppath)
	if err != nil {
		// debug output
		err = fmt.Errorf("(AzureDownload) error removing temp file d: %s", err.Error())
		return
	}

	// add sym link from cacheItemPath to itemPath
	err = os.Symlink(cacheindexpath+"/idx", indexpath+"/idx")
	if err != nil {
		log.Fatalf("(AzureDownload) Unable to download to create symlink from %s to %s, %s", cacheindexpath, indexpath, err.Error())
		return
	}

	return
}

//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************

// GCloudStoreDownload support for downloading off https://github.com/daos-stack
func GCloudStoreDownload(uuid string, nodeInstance *Node, location *conf.LocationConfig) (err error) {

	// 4) Download the item from the bucket. If an error occurs, log it and exit. Otherwise, notify the user that the download succeeded.
	// needs to create a full path
	cacheitempath := uuid2CachePath(uuid)
	cacheitemfile := fmt.Sprintf("%s/%s.data", cacheitempath, uuid)

	itempath := uuid2Path(uuid)
	itemkey := fmt.Sprintf("%s.data", uuid)
	itemfile := fmt.Sprintf("%s/%s.data", itempath, uuid)

	// create cache dir path
	err = os.MkdirAll(cacheitempath, 0777)
	if err != nil {
		log.Fatalf("(GCloudStoreDownload) Unable to create cache path for item %s [%s], %s", cacheitemfile, cacheitempath, err.Error())
		return
	}
	err = os.MkdirAll(itempath, 0777)
	if err != nil {
		log.Fatalf("(GCloudStoreDownload) Unable to create path for item %s [%s], %s", itemfile, itempath, err.Error())
		return
	}

	// create a cache item here
	file, err := os.Create(cacheitemfile)
	defer file.Close()

	if err != nil {
		logger.Debug(3, "(GCloudStoreDownload) cannot create local file handle for: %s, itemkey: %s", cacheitemfile, uuid)
		return
	}

	// GCS specific part
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithAPIKey(location.SecretKey))
	if err != nil {
		log.Fatalf("(GCloudStoreDownload)Failed to create GCloud client: %s", err)
	}

	//  a Bucket handle
	bucket := client.Bucket(location.Bucket)

	//	a handle for our file
	obj := bucket.Object(itemkey)

	// read an object
	r, err := obj.NewReader(ctx)
	if err != nil {
		log.Fatalf("(GCloudStoreDownload)  blob not found: %s [Err: %s]", uuid, err.Error())
	}
	defer r.Close()
	if _, err := io.Copy(file, r); err != nil {
		logger.Debug(3, "(GCloudStoreDownload) error downloading blob: %s [Err: %s]", uuid, err.Error())
	}
	// end GCS specific
	file.Close()

	// add sym link from cacheItemPath to itemPath
	err = os.Symlink(cacheitemfile, itemfile)
	if err != nil {
		log.Fatalf("(GCloudStoreDownload) Unable to download to create symlink from %s to %s, %s", cacheitemfile, itemfile, err.Error())
		return
	}

	// download index files as well

	// download the zipped archive with all the indexed
	indexfile := fmt.Sprintf("%s.idx.zip", uuid) // the zipped contents of the idx directory in S3
	cacheindexpath := uuid2CachePath(uuid)
	indexpath := uuid2Path(uuid)

	indextemppath := fmt.Sprintf("%s/%s", conf.PATH_LOCAL, indexfile) // use the PATH_LOCAL to configure tmp dir

	//	logger.Infof("(GCloudStoreDownload) attempting index download (indexfile: %s), cacheindexpath: %s\n", indexfile, cacheindexpath)

	file, err = os.Create(indextemppath)
	if err != nil {
		logger.Infof("(GCloudDownload) attempting create index temp dir: %s FAILED", indextemppath)

		return
	}
	defer file.Close()

	/////
	//	a handle for our file
	obj = bucket.Object(indexfile)

	// read an object
	r, err = obj.NewReader(ctx)
	if err != nil {
		log.Fatalf("(GCloudStoreDownload)  blob not found: %s [Err: %s]", indexfile, err.Error())
	}
	defer r.Close()
	if _, err := io.Copy(file, r); err != nil {
		logger.Debug(3, "(GCloudStoreDownload) error downloading blob: %s [Err: %s]", indexfile, err.Error())
	}
	// end GCS specific

	file.Close()

	// unzip the index file
	_, err = Unzip(indexfile, cacheindexpath) // unzip into idx folder
	if err != nil {
		// debug output
		err = fmt.Errorf("(GCloudStoreDownload) error decompressing d: %s", err.Error())
		return
	}
	// remove the zip file
	err = os.Remove(indextemppath)
	if err != nil {
		// debug output
		err = fmt.Errorf("(GCloudStoreDownload) error removing temp file d: %s", err.Error())
		return
	}

	// add sym link from cacheItemPath to itemPath
	err = os.Symlink(cacheindexpath+"/idx", indexpath+"/idx")
	if err != nil {
		log.Fatalf("(GCloudStoreDownload) Unable to download to create symlink from %s to %s, %s", cacheindexpath, indexpath, err.Error())
		return
	}

	return
}

//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************

// DaosDownload support for downloading off https://github.com/daos-stack
func DaosDownload(uuid string, nodeInstance *Node) (err error) {
	logger.Infof("(S3Download--> DAOS ) needs to be implemented !! \n")

	return
}

//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************

// ShockDownload download a file from a Shock server
func ShockDownload(uuid string, nodeInstance *Node, location *conf.LocationConfig) (err error) {

	logger.Debug(1, "(ShockDownload) needs to be redone for cache item path")

	// return error if file not found in S3bucket
	logger.Debug(1, "(ShockDownload) attempting download, UUID: %s, nodeID: %s", uuid, nodeInstance.Id)

	cacheitempath := uuid2CachePath(uuid)
	cacheitemfile := fmt.Sprintf("%s/%s.data", cacheitempath, uuid)

	itempath := uuid2Path(uuid)
	itemfile := fmt.Sprintf("%s/%s.data", itempath, uuid)

	// create cache dir path
	err = os.MkdirAll(cacheitempath, 0777)
	if err != nil {
		log.Fatalf("(GCloudStoreDownload) Unable to create cache path for item %s [%s], %s", cacheitemfile, cacheitempath, err.Error())
		return
	}
	err = os.MkdirAll(itempath, 0777)
	if err != nil {
		log.Fatalf("(GCloudStoreDownload) Unable to create path for item %s [%s], %s", itemfile, itempath, err.Error())
		return
	}

	// create a cache item here
	file, err := os.Create(cacheitemfile)
	defer file.Close()

	// SHOCK specific part

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
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Fatalf("(ShockDownload) Unable to download item %q for %s, %v", itemfile, itemfile, err)
		return
	}
	// end of SHOCK specific part

	// now download the indices
	file.Close()

	//	time.Sleep(time.Second * 3)

	fmt.Printf("(ShockDownload) downloaded, UUID: %s, itemS3key: %s", uuid, itemfile)
	logger.Infof("(ShockDownload)  downloaded, UUID: %s, itemS3key: %s", uuid, itemfile)

	//this will download the indices via a new API feature in SHOCK
	// download the zipped archive with all the indexed
	// indexfile := fmt.Sprintf("%s.idx", uuid) // the zipped contents of the idx directory in S3
	// indexpath := uuid2CachePath(uuid)
	// indextemppath := fmt.Sprintf("%s/idx.zip", indexpath)
	// indexdir := fmt.Sprintf("%s/idx", indexpath)

	indexfile := fmt.Sprintf("%s.idx.zip", uuid) // the zipped contents of the idx directory in S3
	cacheindexpath := uuid2CachePath(uuid)
	indexpath := uuid2Path(uuid)

	indextemppath := fmt.Sprintf("%s/%s", conf.PATH_LOCAL, indexfile) // use the PATH_LOCAL to configure tmp dir

	url = fmt.Sprintf("%s/%s.?download", location.URL, indexfile)

	// Create the file
	file, err = os.Create(indextemppath)
	if err != nil {
		return
	}
	defer file.Close()

	// Get the data
	resp, err = http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		logger.Debug(1, "(ShockDownload) Unable to download item %q for %s, %v", itemfile, itemfile, err)
		return
	}
	// unzip the index file
	_, err = Unzip(indexfile, cacheindexpath) // unzip into idx folder
	if err != nil {
		logger.Debug(1, "(ShockDownload) error decompressing %q, %v", itemfile, err)
		// debug output
		err = fmt.Errorf("(ShockDownload) error decompressing d: %s", err.Error())
		return
	}
	// remove the zip file
	err = os.Remove(indextemppath)
	if err != nil {
		// debug output
		logger.Debug(1, "(ShockDownload) error removing temp file d: %s", err.Error())
		err = fmt.Errorf("(ShockDownload) error removing temp file d: %s", err.Error())
		return
	}

	// add sym link from cacheItemPath to itemPath
	err = os.Symlink(cacheindexpath+"/idx", indexpath+"/idx")
	if err != nil {
		log.Fatalf("(ShockDownload) Unable to download to create symlink from %s to %s, %s", cacheindexpath, indexpath, err.Error())
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

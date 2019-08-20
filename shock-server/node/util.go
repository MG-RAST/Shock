package node

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

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

// func Open(name string) (*File, error)
//  Open opens the named file for reading from local disk
//  Open file on list of remote stores

// FMOpen drop in replacement for os.Open that attempts to download files from locations stored in MongoDB
func FMOpen(filepath string) (f *os.File, err error) {

	// try to read file from disk
	f, err = os.Open(filepath)

	// check if we encounter an error
	if err == nil {
		return
	}

	// WE NEED TO LOCK THIS NODE at this point

	// extract UUID from path(should be path2UUID function really)
	ext := path.Ext(filepath)                     // identify extension
	filename := strings.TrimSuffix(filepath, ext) // find filename
	uuid := path.Base(filename)                   // implement basename cmd

	var nodeInstance, _ = Load(uuid)

	// create the directory infrastructure for node and index
	err = nodeInstance.Mkdir()
	if err != nil {
		return
	}

	// ideally we should loop over all instances of remotes
	//for _, locationStr := range nodeInstance.Locations {
Loop:
	for _, locationStr := range []string{"anls3_anlseq"} {

		location, ok := conf.LocationsMap[locationStr]
		if !ok {
			err = fmt.Errorf("(FM) location unknown %s", locationStr)
			return
		}
		// debug
		//	spew.Dump(location)

		switch location.Type {
		// we implement only S3 for now
		case "S3":
			err = S3Download(uuid, nodeInstance, location)
			if err != nil {
				// debug output
				err = fmt.Errorf("(FMOpen) S3download returned: %s", err.Error())
				return
			}
			break Loop
		case "Shock":
			// this should be expanded to handle Shock servers sharing the same Mongo instance
			err = ShockDownload(uuid, nodeInstance)
			if err != nil {
				// debug output
				err = fmt.Errorf("(FMOpen) ShockDownload returned: %s", err.Error())
				return
			}
			break Loop
		default:
			err = fmt.Errorf("(FMOpen) Location type %s not supported", location.Type)
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

// S3Download download a file and its indices from an S3 source
func S3Download(uuid string, nodeInstance *Node, location *conf.Location) (err error) {

	// return error if file not found in S3bucket
	fmt.Printf("(S3Download) attempting download, UUID: %s, nodeID: %s from: %s\n", uuid, nodeInstance.Id, location.URL)
	//logger.Infof("(S3Download) attempting download, UUID: %s, nodeID: %s", uuid, nodeInstance.Id)

	Bucket := location.Bucket

	// 2) Create an AWS session
	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(location.AuthKey, location.SecretKey, ""),
		Endpoint:    aws.String(location.URL),
		Region:      aws.String("us-east-1"),
		//DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	}
	sess := session.New(s3Config)

	// 3) Create a new AWS S3 downloader
	downloader := s3manager.NewDownloader(sess)

	// 4) Download the item from the bucket. If an error occurs, log it and exit. Otherwise, notify the user that the download succeeded.
	// needs to create a full path
	itempath := uuid2Path(uuid)
	itemS3key := fmt.Sprintf("%s.data", uuid)
	itemfile := fmt.Sprintf("%s/%s.data", itempath, uuid)

	//	fmt.Printf("(S3Download) attempting download, UUID: %s, itemS3key: %s", uuid, itemS3key)
	//logger.Infof("(S3Download) attempting download, UUID: %s, itemS3key: %s", uuid, itemS3key)

	file, err := os.Create(itemfile)
	if err != nil {
		return
	}

	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(Bucket),
			Key:    aws.String(itemS3key),
		})

	if err != nil {
		log.Fatalf("(S3Download) Unable to download item %q for %s, %v", itemS3key, itemfile, err)
		return
	}

	file.Close()

	//	time.Sleep(time.Second * 3)

	//fmt.Printf("(S3Download) downloaded, UUID: %s, itemS3key: %s", uuid, itemS3key)
	//logger.Infof("(S3Download)  downloaded, UUID: %s, itemS3key: %s", uuid, itemS3key)

	if false {
		// download the zipped archive with all the indexed
		indexfile := fmt.Sprintf("%s.idx", uuid) // the zipped contents of the idx directory in S3
		indexpath := uuid2Path(uuid)
		indextemppath := fmt.Sprintf("%s/idx.zip", indexpath)
		indexdir := fmt.Sprintf("%s/idx", indexpath)

		file, err = os.Create(indextemppath)
		if err != nil {
			return
		}
		defer file.Close()

		numBytes, err = downloader.Download(file,
			&s3.GetObjectInput{
				Bucket: aws.String(Bucket),
				Key:    aws.String(indexfile),
			})

		// we should distinguish between file not found and transfer error

		// if there is no index archive in S3
		if err == nil {
			// log.Fatalf("Unable to download index %q, %v", indexfile, err)

			fmt.Println("Downloaded", file.Name(), numBytes, "bytes")

			// unzip the index file
			_, err = Unzip(indexfile, indexdir) // unzip into idx folder
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
		}
	}
	return
}

// ShockDownload download a file from a Shock server
func ShockDownload(uuid string, nodeInstance *Node) (err error) {

	// return error if file not found in S3bucket
	fmt.Printf("(ShockDownload) attempting download, UUID: %s, nodeID: %s", uuid, nodeInstance.Id)
	logger.Infof("(ShockDownload) attempting download, UUID: %s, nodeID: %s", uuid, nodeInstance.Id)

	// authkey
	var authkey = "blah" // this needs to be read from the auth config file (YAML file)
	var baseurl = "blah" // this needs to be read from the locations object, e.g. "http://shock.mg-rast.org/node/<UUID>?download"

	// dfds
	itempath := uuid2Path(uuid)
	itemS3key := fmt.Sprintf("%s.data", uuid)
	_ = itemS3key
	itemfile := fmt.Sprintf("%s/%s.data", itempath, uuid)

	url := fmt.Sprintf("%s/%s?download", baseurl, uuid)
	options := fmt.Sprint("%s", authkey)
	_ = options

	// Create the file
	out, err := os.Create(itemfile)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatalf("(ShockDownload) Unable to download item %q for %s, %v", itemfile, itemfile, err)
		return
	}

	// now download the indices
	out.Close()

	//	time.Sleep(time.Second * 3)

	fmt.Printf("(ShockDownload) downloaded, UUID: %s, itemS3key: %s", uuid, itemfile)
	logger.Infof("(ShockDownload)  downloaded, UUID: %s, itemS3key: %s", uuid, itemfile)

	//this will download the indices via a new API feature in SHOCK
	if false {
		// download the zipped archive with all the indexed
		indexfile := fmt.Sprintf("%s.idx", uuid) // the zipped contents of the idx directory in S3
		indexpath := uuid2Path(uuid)
		indextemppath := fmt.Sprintf("%s/idx.zip", indexpath)
		indexdir := fmt.Sprintf("%s/idx", indexpath)

		url = fmt.Sprintf("%s/%s.?download", baseurl, indexfile)
		options = fmt.Sprint("%s", authkey)

		// Create the file
		out, err = os.Create(itemfile)
		if err != nil {
			return err
		}
		defer out.Close()

		// Get the data
		resp, err = http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Check server response
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("bad status: %s", resp.Status)
		}

		// Writer the body to file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			log.Fatalf("(ShockDownload) Unable to download item %q for %s, %v", itemfile, itemfile, err)
			return
		}
		// unzip the index file
		_, err = Unzip(indexfile, indexdir) // unzip into idx folder
		if err != nil {
			// debug output
			err = fmt.Errorf("(ShockDownload) error decompressing d: %s", err.Error())
			return
		}
		// remove the zip file
		err = os.Remove(indextemppath)
		if err != nil {
			// debug output
			err = fmt.Errorf("(ShockDownload) error removing temp file d: %s", err.Error())
			return
		}

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

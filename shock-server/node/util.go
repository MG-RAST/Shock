package node

import (
	"fmt"
	"log"
	"os"
	"path"
	"sort"
	"strings"

	. "github.com/MG-RAST/Shock/shock-server/Location"

	"github.com/aws/aws-sdk-go/aws"
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

	// extract UUID from path
	ext := path.Ext(filepath)                     // identify extension
	filename := strings.TrimSuffix(filepath, ext) // find filename
	uuid := path.Base(filename)                   // implement basename cmd

	var nodeInstance, _ = Load(uuid)

	var id2loc map[string]*Location

	var locations []Location

	locations = append(locations, Location{ID: "ANLS3", Type: "S3"})

	id2loc = make(map[string]*Location)

	for _, loc := range locations {
		id2loc[loc.ID] = &loc
	}

	// ideally we should loop over all instances of remotes
	for _, locationStr := range nodeInstance.Locations {

		location, ok := id2loc[locationStr]
		if !ok {
			err = fmt.Errorf("location unknown %s", locationStr)
			return
		}

		switch location.Type {

		// we implement only S3 for now
		case "S3":
			f, err = S3Open(uuid, nodeInstance)
			if err == nil {
				return
			}
		case "Shock":
			// this should be expanded to handle Shock servers sharing the same Mongo instance
			return
		default:
			err = fmt.Errorf("(FMOpen) Location type %s not supported", location.Type)
			return
		}
		// if we are here we did not find what we needed
		err = fmt.Errorf("(FMOpen) Object (%s) not found in any location", uuid)

		f, err = S3Open(uuid, nodeInstance)

		if err != nil {
			// debug output
			err = fmt.Errorf("(FMOpen) S3Open returned: %s", err.Error())
			return
		}
	}

	return
}

// atttempt to download the missing node from an S3 implementation and return a file handle

// S3Open download and open a file from an S3 source return filehandle on local storage
func S3Open(uuid string, nodeInstance *Node) (f *os.File, err error) {

	// return error if file not found in S3bucket
	fmt.Printf("(S3Open) download, UUID: %s, nodeID: %s", uuid, nodeInstance.Id)

	// NOTE: you need to store your AWS credentials in ~/.aws/credentials

	// 1) Define your bucket and item names
	bucket := "Shock" //unclear
	item := uuid

	// 2) Create an AWS session
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")}, // NEEDs changes
	)

	// 3) Create a new AWS S3 downloader
	downloader := s3manager.NewDownloader(sess)

	// 4) Download the item from the bucket. If an error occurs, log it and exit. Otherwise, notify the user that the download succeeded.
	file, err := os.Create(item)
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(item),
		})

	if err != nil {
		log.Fatalf("Unable to download item %q, %v", item, err)
	}

	fmt.Println("Downloaded", file.Name(), numBytes, "bytes")

	// s3 client lib...
	if err != nil {
		// debug output
		err = fmt.Errorf("(FMOpen) S3 download returned: %s", err.Error())
		return
	}

	var path = getPath(uuid) // set this to the correct shock filename use method from helper package

	// try to read file from disk
	f, err = os.Open(path)

	// check if we encounter an error
	if err == nil {
		return
	}
	return
}

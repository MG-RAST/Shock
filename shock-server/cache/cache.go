package cache

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/logger"
	// remove for production :-)
)

// Item information to manage file expiration on cache servers
type Item struct {
	UUID      string    `bson:"uuid" json:"uuid" `                   // e.g. node UUID
	Access    time.Time `bson:"last_accessed" json:"last_accessed" ` // e.g. access time
	Type      string    `bson:"type" json:"type"`
	Size      int64     `bson:"size" json:"size"  `        // e.g. size in bytes
	CreatedOn time.Time `bson:"url" json:"url" yaml:"URL"` // e.g. creation of local copy
}

// CacheMap store UUID, size, type, atime and ctime in separate (sorted) table for access by fileReaper
var CacheMap map[string]*Item

// CacheMapLock lock write access to the CacheMap
var CacheMapLock = sync.RWMutex{}

// Path2uuid extract uuid from path
func path2uuid(fpath string) string {

	ext := path.Ext(fpath)                     // identify extension
	filename := strings.TrimSuffix(fpath, ext) // find filename
	uuid := path.Base(filename)                // implement basename cmd

	return uuid
}

// Initialize find all *.data files for nodes and populate cache
// only call if global --is-cache=true
func Initialize() (err error) {

	if conf.PATH_CACHE == "" { // no PATH_CACHE set will stop
		logger.Infof("(cache) not initializing; not configured)")
		return
	}

	//logger.Info(fmt.Sprintf("(cache) initializing: %s)\n ", DataRoot))
	//
	Pattern := fmt.Sprintf("%s/*/*/*/*/*.data", conf.PATH_CACHE)

	//debug
	logger.Infof("(cache->Initialize) listing files for Pattern: %s", Pattern)

	nodefiles, err := filepath.Glob(Pattern)

	if err != nil {
		logger.Errorf("(cache->Initialize) error reading %s (Error:%s)", Pattern, err.Error())
		return
	}
	CacheMap = make(map[string]*Item)

	for _, file := range nodefiles {

		var fileinfo os.FileInfo
		fileinfo, err = os.Stat(file)

		if err != nil {
			logger.Errorf("(cache->Initialize) error reading %s (Error:%s)", file, err.Error())
			continue
		}
		filename := path2uuid(file)

		var entry Item

		entry.UUID = filename
		entry.Size = fileinfo.Size()
		entry.CreatedOn = fileinfo.ModTime()
		//Item.Access = ""
		now := time.Now()
		age := entry.CreatedOn
		diff := now.Sub(age)
		hours := diff.Hours()

		logger.Infof("(cache->Initialize) added UUID %s, Size: %d, age(h): %f", entry.UUID, entry.Size, hours)

		// add the map bits
		CacheMap[entry.UUID] = &entry

	}
	return
}

// Add an entry to the Cache for ID
func Add(ID string, size int64) {

	var entry Item

	entry.UUID = ID
	entry.Size = size
	entry.CreatedOn = time.Now()

	CacheMapLock.Lock()
	defer CacheMapLock.Unlock()

	CacheMap[entry.UUID] = &entry

	logger.Infof("(Cache-->Add) added file: %s with size: %d", ID, size)

	return
}

// Remove an entry to the CacheMap and the file on disk
func Remove(ID string) (err error) {
	// return immediately if system is not setup to be cache
	if conf.PATH_CACHE == "" {
		return
	}

	cachepath := fmt.Sprintf("%s/*/*/*/%s", conf.PATH_CACHE, ID) // the data file in cache
	itempath := fmt.Sprintf("%s/*/*/*/%s", conf.PATH_DATA, ID)   // the symlink

	// remove cacheitem
	err = os.RemoveAll(cachepath)
	if err != nil {
		logger.Infof("(Cache-->Remove) cannot remove %s from cache (%s)", cachepath, err.Error())
	}

	// remove link
	err = os.RemoveAll(itempath)
	if err != nil {
		logger.Infof("(Cache-->Remove) cannot remove symlink %s (%s)", cachepath, err.Error())
	}

	// remove index files and index sym link as well
	cacheindexdir := fmt.Sprintf("%s/*/*/*/%s/idx", conf.PATH_CACHE, ID) // the data file in cache
	indexdir := fmt.Sprintf("%s/*/*/*/%s/idx", conf.PATH_DATA, ID)       // the symlink

	err = os.RemoveAll(cacheindexdir)
	if err != nil {
		logger.Infof("(Cache-->Remove) cannot remove %s from cache (%s)", cacheindexdir, err.Error())
	}

	err = os.RemoveAll(indexdir)
	if err != nil {
		logger.Infof("(Cache-->Remove) cannot symlink %s from cache (%s)", indexdir, err.Error())
	}

	CacheMapLock.Lock()
	defer CacheMapLock.Unlock()
	_, ok := CacheMap[ID]
	if !ok {
		logger.Infof("(Cache-->Remove) cannot remove ID: [%s] from CacheMap (%s) ", ID, err.Error())
		return
	}
	// remove object from Map and remove Cache Entry
	delete(CacheMap, ID)
	return

}

// Touch update cache LRU info
func Touch(ID string) {

	//spew.Dump(CacheMap)

	_, ok := CacheMap[ID]
	if !ok {
		return
	}
	CacheMapLock.Lock()
	defer CacheMapLock.Unlock()

	CacheMap[ID].Access = time.Now()
	logger.Infof("(Cache-->Touch) lru for  %s updated to %s", ID, time.Now())

}

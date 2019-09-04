package node

import (
	"regexp"
	"time"

	"github.com/MG-RAST/Shock/shock-server/cache"
	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node/locker"
	"gopkg.in/mgo.v2/bson"
)

var (
	Ttl         *NodeReaper
	ExpireRegex = regexp.MustCompile(`^(\d+)(M|H|D)$`)
)

func InitReaper() {
	Ttl = NewNodeReaper()
}

type NodeReaper struct{}

func NewNodeReaper() *NodeReaper {
	return &NodeReaper{}
}

func (nr *NodeReaper) Handle() {
	waitDuration := time.Duration(conf.EXPIRE_WAIT) * time.Minute
MainLoop:
	for {

		// sleep
		time.Sleep(waitDuration)
		// query to get expired nodes
		nodes := Nodes{}
		query := nr.getQuery()
		nodes.GetAll(query)

	NodeLoop:
		// loop thru all nodes
		for _, n := range nodes {
			logger.Infof("Deleting expired node: %s", n.Id)
			// delete expired nodes
			if err := n.Delete(); err != nil {
				err_msg := "err:@node_delete: " + err.Error()
				logger.Error(err_msg)
			}

			// write info for node migration
			logger.Debug(1, "writeing info for node migration: %s", n.Id)
			// CODE MISSING HERE FOR WRITING FILES WITH INFO FOR MIGRATION

			//  node data deletion
			logger.Debug(1, "node data deletion: %s", n.Id)

			//var loc location
			for _, loc := range n.Locations {
				var counter int = 0 // we need at least N locations before we can erase data files on local disk

				// delete only if other locations exist
				locObj, ok := conf.LocationsMap[loc.ID]

				if !ok {
					logger.Errorf("(Reaper-->FileReaper) location %s is not defined in this server instance \n ", loc)
					continue
				}
				//fmt.Printf("(Reaper-->FileReaper) locObj.Persistent =  %b  \n ", locObj.Persistent)
				if locObj.Persistent == true {
					logger.Debug(2, "(Reaper-->FileReaper) has remote Location (%s) removing from Data: %s", loc.ID, n.Id)
					counter++ // increment counter
				}
				if counter >= conf.MIN_REPLICA_COUNT {
					err := n.DeleteFiles() // delete all data files for node in PATH_DATA NOTE: this is different from PATH_CACHE
					if err != nil {
						logger.Errorf("(Reaper-->FileReaper) files for node %s could not be deleted (Err: %s) ", n.Id, err.Error())
						continue
					}
					continue NodeLoop
					// the innermost loop
				}
				///

			}
			// garbage collection: remove old nodes from Lockers, value is hours old
			locker.NodeLockMgr.RemoveOld(1)
			locker.FileLockMgr.RemoveOld(6)
			locker.IndexLockMgr.RemoveOld(6)

			// we do not start deletings files if we are not in cache mode
			// we might want to change this, if we are in shock-migrate or we do not have a cache_path we skip this
			if conf.PATH_CACHE == "" {
				continue MainLoop
			}
		CacheMapLoop:
			// start a FILE REAPER that loops thru CacheMap[*]
			for ID := range cache.CacheMap {

				logger.Debug(3, "(Reaper-->FileReaper) checking %s in cache\n", ID)

				now := time.Now()
				lru := cache.CacheMap[ID].Access
				diff := now.Sub(lru)

				// we use a very simple scheme for caching initially (file not used for 1 day)
				if diff.Hours() < float64(conf.CACHE_TTL) {
					logger.Debug(3, "Reaper-->FileReaper) not deleting %s from cache it was last accessed %s hours ago\n", ID, diff.Hours())
					continue CacheMapLoop
				}

				n, err := Load(ID)
				if err != nil {
					logger.Debug(1, "(Reaper-->FileReaper) Cannot access CacheMapItem[%s] (%s)", ID, err.Error())
					continue CacheMapLoop
				}
				cache.Remove(ID)

				logger.Errorf("(Reaper-->FileReaper) cannot delete %s from cache [This should not happen!!]", ID)
			}
		}
	}
	return
}

func (nr *NodeReaper) getQuery() (query bson.M) {
	hasExpire := bson.M{"expiration": bson.M{"$exists": true}}   // has the field
	toExpire := bson.M{"expiration": bson.M{"$ne": time.Time{}}} // value has been set, not default
	isExpired := bson.M{"expiration": bson.M{"$lt": time.Now()}} // value is too old
	query = bson.M{"$and": []bson.M{hasExpire, toExpire, isExpired}}
	return
}

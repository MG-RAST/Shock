package node

import (
	"fmt"
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
	for {

		// sleep
		time.Sleep(waitDuration)
		// query to get expired nodes
		nodes := Nodes{}
		query := nr.getQuery()
		nodes.GetAll(query)
		// delete expired nodes
		for _, n := range nodes {
			logger.Infof("Deleting expired node: %s", n.Id)
			if err := n.Delete(); err != nil {
				err_msg := "err:@node_delete: " + err.Error()
				logger.Error(err_msg)
			}
		}
		// garbage collection: remove old nodes from Lockers, value is hours old
		locker.NodeLockMgr.RemoveOld(1)
		locker.FileLockMgr.RemoveOld(6)
		locker.IndexLockMgr.RemoveOld(6)

		// we do not start deletings files if we are not in cache mode
		if conf.PATH_CACHE == "" {
			return
		}
	Loop2:
		// start a FILE REAPER that loops thru CacheMap[*]
		for ID := range cache.CacheMap {

			fmt.Printf("(Reaper-->FileReaper) checking %s in cache\n", ID)

			now := time.Now()
			lru := cache.CacheMap[ID].Access

			diff := now.Sub(lru)

			// we use a very simple scheme for caching initially (file not used for 1 day)
			if diff.Hours() < 24 {
				//	fmt.Printf("(Reaper-->FileReaper) not deleting %s from cache it was last accessed %s hours ago\n", ID, diff.Hours())
				continue
			}

			// START HERE ON MONDAY
			// ideally we would only have things in the cache that have a remote location
			// maybe init cache from Mongo instead of local filesystem? it'd be faster...

			//fmt.Printf("(Reaper-->FileReaper) trying to delete %s from cache\n ", ID)

			n, _ := Load(ID)
			for _, loc := range n.Locations {
				// delete only if other locations exist
				locObj, ok := conf.LocationsMap[loc]
				if !ok {
					logger.Info(fmt.Sprintf("(Reaper-->FileReaper) location %s is not OK \n ", loc))
					continue
				}
				//fmt.Printf("(Reaper-->FileReaper) locObj.Persistent =  %b  \n ", locObj.Persistent)

				if locObj.Persistent == true {
					logger.Info(fmt.Sprintf("(Reaper-->FileReaper) has remote Location (%s) removing from Cache: %s \n", loc, ID))

					cache.Remove(ID)
					break Loop2 // the innermost loop
				}
			}

			logger.Info(fmt.Sprintf("(Reaper-->FileReaper) cannot delete %s from cache\n ", ID))

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

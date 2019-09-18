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

		// loop thru all nodes
		for _, n := range nodes {
			logger.Debug(6, "Reaper->Checking for expired node: %s", n.Id)

			// delete expired nodes
			deleted, err := n.ExpireNode()
			if err != nil {
				err_msg := "err:@node_ExpireNode: " + err.Error()
				logger.Error(err_msg)
			}
			if deleted == true {
				// node is gone
				continue
			}

			// global conf NODE_DATA_REMOVAL has to be true to delete files in DATA PATH
			if conf.NODE_DATA_REMOVAL == true {
				expired, err := n.ExpireNodeFiles()
				if err != nil {
					err_msg := "err:@nExpireNodeFiles: " + err.Error()
					logger.Error(err_msg)
				}
				if expired == true {
					// node is gone
					continue
				}
			}

			// garbage collection: remove old nodes from Lockers, value is hours old
			locker.NodeLockMgr.RemoveOld(1)
			locker.FileLockMgr.RemoveOld(6)
			locker.IndexLockMgr.RemoveOld(6)

			//  ************************ ************************ ************************ ************************ ************************ ************************ ************************ ************************
			// we do not start deletings files if we are not in cache mode
			// we might want to change this, if we are in shock-migrate or we do not have a cache_path we skip this
			if conf.PATH_CACHE == "" {
				continue MainLoop
			}
		}
	CacheMapLoop:
		// start a FILE REAPER that loops thru CacheMap[*]
		for ID := range cache.CacheMap {
			logger.Debug(3, "(Reaper-->CacheMapLoop) checking %s in cache\n", ID)

			now := time.Now()
			lru := cache.CacheMap[ID].Access
			diff := now.Sub(lru)

			// we use a very simple scheme for caching initially (file not used for 1 day)
			if diff.Hours() < float64(conf.CACHE_TTL) {
				logger.Debug(3, "Reaper-->CacheMapLoop) not deleting %s from cache it was last accessed %s hours ago\n", ID, diff.Hours())
				continue CacheMapLoop
			}

			cache.Remove(ID)
			logger.Errorf("(Reaper-->CacheMapLoop) cannot delete %s from cache [This should not happen!!]", ID)
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

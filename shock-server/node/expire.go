package node

import (
	"regexp"
	"time"

<<<<<<< HEAD
=======
	"github.com/MG-RAST/Shock/shock-server/cache"
>>>>>>> dc34e8103804a3797c83c529391486b4e1d66fd0
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

		// loop thru all nodes
		for _, n := range nodes {
			logger.Debug(6, "Reaper->Checking for expired node: %s", n.Id)

			// delete expired nodes
			deleted, err := n.Delete()
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

package node

import (
	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node/locker"
	"gopkg.in/mgo.v2/bson"
	"regexp"
	"time"
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
	}
}

func (nr *NodeReaper) getQuery() (query bson.M) {
	hasExpire := bson.M{"expiration": bson.M{"$exists": true}}   // has the field
	toExpire := bson.M{"expiration": bson.M{"$ne": time.Time{}}} // value has been set, not default
	isExpired := bson.M{"expiration": bson.M{"$lt": time.Now()}} // value is too old
	query = bson.M{"$and": []bson.M{hasExpire, toExpire, isExpired}}
	return
}

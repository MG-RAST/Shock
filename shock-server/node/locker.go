package node

import (
	"fmt"
	"sync"
	"time"
)

var (
	LockMgr = NewLocker()
)

func NewLocker() *Locker {
	return &Locker{
		nodes: map[string]*NodeLock{},
	}
}

type Locker struct {
	nodes map[string]*NodeLock
	sync.Mutex
}

type NodeLock struct {
	isLocked  bool
	updated   time.Time
	writeLock chan int
}

func (n *NodeLock) init() {
	n.isLocked = false
	n.updated = time.Now()
	n.writeLock <- 1 // Put the initial value into the channel
}

func (n *NodeLock) lock(id string) (err error) {
	select {
	case <-n.writeLock: // Grab the ticket - here is where we wait
	case <-time.After(time.Minute * 30):
		err = fmt.Errorf("Timeout!! Waited 30 mins on lock for node %s", id)
		return
	}
	n.isLocked = true
	n.updated = time.Now()
	return
}

func (n *NodeLock) unlock() {
	n.isLocked = false
	n.updated = time.Now()
	n.writeLock <- 1 // Release the ticket
}

func (l *Locker) LockNode(id string) (err error) {
	// add if missing, may happen if shock restarted
	if _, ok := l.nodes[id]; !ok {
		l.AddNode(id)
	}
	err = l.nodes[id].lock(id)
	return
}

func (l *Locker) UnlockNode(id string) {
	// skip missing id
	if _, ok := l.nodes[id]; ok {
		l.nodes[id].unlock()
	}
}

func (l *Locker) GetLocked() (ids []string) {
	l.Lock()
	for id, n := range l.nodes {
		if n.isLocked {
			ids = append(ids, id)
		}
	}
	l.Unlock()
	return
}

func (l *Locker) AddNode(id string) {
	l.Lock()
	l.nodes[id] = new(NodeLock)
	l.nodes[id].init()
	l.Unlock()
}

func (l *Locker) RemoveNode(id string) {
	l.Lock()
	delete(l.nodes, id)
	l.Unlock()
}

func (l *Locker) RemoveOldNodes(hours int) {
	currTime := time.Now()
	expireTime := currTime.Add(time.Duration(hours*-1) * time.Hour)
	l.Lock()
	for id, n := range l.nodes {
		if (!n.isLocked) && n.updated.Before(expireTime) {
			delete(l.nodes, id)
		}
	}
	l.Unlock()
}

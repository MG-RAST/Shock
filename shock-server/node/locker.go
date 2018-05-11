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

func NewNodeLock(id string) (n *NodeLock) {
	n = &NodeLock{
		id:        id,
		isLocked:  false,
		updated:   time.Now(),
		writeLock: make(chan int, 1),
	}
	n.writeLock <- 1 // Put the initial value into the channel
	return
}

type NodeLock struct {
	id        string
	isLocked  bool
	updated   time.Time
	writeLock chan int
}

func (n *NodeLock) lock() (err error) {
	select {
	case <-n.writeLock: // Grab the ticket - here is where we wait
	case <-time.After(time.Minute * 30):
		err = fmt.Errorf("Timeout!! Waited 30 mins on lock for node %s", n.id)
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
	l.Lock()
	defer l.Unlock()
	if _, ok := l.nodes[id]; !ok {
		l.nodes[id] = NewNodeLock(id)
	}
	err = l.nodes[id].lock()
	return
}

func (l *Locker) UnlockNode(id string) {
	l.Lock()
	defer l.Unlock()
	if _, ok := l.nodes[id]; ok {
		l.nodes[id].unlock()
	}
}

func (l *Locker) GetLocked() (ids []string) {
	l.Lock()
	defer l.Unlock()
	for id, n := range l.nodes {
		if n.isLocked {
			ids = append(ids, id)
		}
	}
	return
}

func (l *Locker) GetAll() (ids []string) {
	l.Lock()
	defer l.Unlock()
	for id, _ := range l.nodes {
		ids = append(ids, id)
	}
	return
}

func (l *Locker) RemoveNode(id string) {
	l.Lock()
	defer l.Unlock()
	delete(l.nodes, id)
}

func (l *Locker) RemoveOldNodes(hours int) {
	l.Lock()
	defer l.Unlock()
	currTime := time.Now()
	expireTime := currTime.Add(time.Duration(hours*-1) * time.Hour)
	l.Lock()
	for id, n := range l.nodes {
		if (!n.isLocked) && n.updated.Before(expireTime) {
			delete(l.nodes, id)
		}
	}
}

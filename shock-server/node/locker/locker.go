package locker

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/MG-RAST/Shock/shock-server/logger"
)

var (
	NodeLockMgr  = NewNodeLocker()
	FileLockMgr  = NewFileLocker()
	IndexLockMgr = NewIndexLocker()
	lockDebug    = false
)

// NodeLocker prevents a process from loading node from mongo while another process is update the same node
func NewNodeLocker() *NodeLocker {
	return &NodeLocker{
		nodes: make(map[string]*NodeLock),
	}
}

type NodeLocker struct {
	nodes map[string]*NodeLock
	sync.Mutex
}

// FileLocker prevents a node file from being downloaded while locked
func NewFileLocker() *FileLocker {
	return &FileLocker{
		nodes: make(map[string]*LockInfo),
	}
}

type FileLocker struct {
	nodes map[string]*LockInfo
	sync.Mutex
}

// IndexLocker prevents a node file from being downloaded using the locked index
func NewIndexLocker() *IndexLocker {
	return &IndexLocker{
		nodes: make(map[string]map[string]*LockInfo),
	}
}

type IndexLocker struct {
	nodes map[string]map[string]*LockInfo
	sync.Mutex
}

func NewLockInfo() *LockInfo {
	return &LockInfo{CreatedOn: time.Now()}
}

type LockInfo struct {
	CreatedOn time.Time `bson:"-" json:"created_on"`
	Error     string    `bson:"-" json:"error"`
}

// ############## NodeLocker ##############

func NewNodeLock(id string) (n *NodeLock) {
	if lockDebug {
		fmt.Printf("locker.NewNodeLock gid=%d node=%s\n", getGID(), id)
	}
	n = &NodeLock{
		Id:        id,
		IsLocked:  false,
		Updated:   time.Now(),
		writeLock: make(chan int, 1),
	}
	n.writeLock <- 1 // Put the initial value into the channel
	return
}

type NodeLock struct {
	Id        string    `bson:"-" json:"id"`
	IsLocked  bool      `bson:"-" json:"locked"`
	Updated   time.Time `bson:"-" json:"updated_on"`
	writeLock chan int
}

func (n *NodeLock) lock() (err error) {
	if lockDebug {
		fmt.Printf("start NodeLock.lock gid=%d node=%s\n", getGID(), n.Id)
	}
	select {
	case <-n.writeLock: // Grab the ticket - here is where we wait
	case <-time.After(time.Minute * 30):
		err = fmt.Errorf("Timeout!! Waited 30 mins on lock for node %s", n.Id)
		return
	}
	n.IsLocked = true
	n.Updated = time.Now()
	if lockDebug {
		fmt.Printf("end NodeLock.lock gid=%d node=%s\n", getGID(), n.Id)
	}
	return
}

func (n *NodeLock) unlock() {
	if lockDebug {
		fmt.Printf("start NodeLock.unlock gid=%d node=%s\n", getGID(), n.Id)
	}
	n.IsLocked = false
	n.Updated = time.Now()
	n.writeLock <- 1 // Release the ticket
	if lockDebug {
		fmt.Printf("end NodeLock.unlock gid=%d node=%s\n", getGID(), n.Id)
	}
}

func (l *NodeLocker) LockNode(id string) (err error) {
	l.Add(id)
	err = l.nodes[id].lock() // here is where we wait
	return
}

func (l *NodeLocker) UnlockNode(id string) {
	l.Lock()
	defer l.Unlock()
	if _, ok := l.nodes[id]; ok {
		l.nodes[id].unlock()
	}
}

func (l *NodeLocker) GetLocked() (nodes []*NodeLock) {
	l.Lock()
	defer l.Unlock()
	for _, n := range l.nodes {
		if n.IsLocked {
			nodes = append(nodes, n)
		}
	}
	return
}

func (l *NodeLocker) GetAll() (nodes []*NodeLock) {
	l.Lock()
	defer l.Unlock()
	for _, n := range l.nodes {
		nodes = append(nodes, n)
	}
	return
}

func (l *NodeLocker) Add(id string) {
	// add if missing, may happen if shock restarted
	l.Lock()
	defer l.Unlock()
	if _, ok := l.nodes[id]; !ok {
		l.nodes[id] = NewNodeLock(id)
	}
}

func (l *NodeLocker) Remove(id string) {
	l.Lock()
	defer l.Unlock()
	delete(l.nodes, id)
}

func (l *NodeLocker) RemoveOld(hours int) {
	l.Lock()
	defer l.Unlock()
	currTime := time.Now()
	expireTime := currTime.Add(time.Duration(hours*-1) * time.Hour)
	for id, n := range l.nodes {
		if (!n.IsLocked) && n.Updated.Before(expireTime) {
			delete(l.nodes, id)
		}
	}
}

// ############## FileLocker ##############

func (f *FileLocker) GetAll() map[string]*LockInfo {
	f.Lock()
	defer f.Unlock()
	return f.nodes
}

func (f *FileLocker) Get(id string) *LockInfo {
	f.Lock()
	defer f.Unlock()
	if info, ok := f.nodes[id]; ok {
		return info
	}
	return nil
}

func (f *FileLocker) Add(id string) *LockInfo {
	f.Lock()
	defer f.Unlock()
	f.nodes[id] = NewLockInfo()
	return f.nodes[id]
}

func (f *FileLocker) Error(id string, err error) {
	if err == nil {
		return
	}
	f.Lock()
	defer f.Unlock()
	if info, ok := f.nodes[id]; ok {
		info.Error = err.Error()
		logger.Errorf("error during asynchronous file processing node=%s: %s", id, err.Error())
	}
}

func (f *FileLocker) Remove(id string) {
	f.Lock()
	defer f.Unlock()
	delete(f.nodes, id)
}

func (f *FileLocker) RemoveOld(hours int) {
	f.Lock()
	defer f.Unlock()
	currTime := time.Now()
	expireTime := currTime.Add(time.Duration(hours*-1) * time.Hour)
	for id, info := range f.nodes {
		if info.CreatedOn.Before(expireTime) {
			logger.Errorf("Removing stale file lock: node=%s", id)
			delete(f.nodes, id)
		}
	}
}

// ############## IndexLocker ##############

func (i *IndexLocker) GetAll() map[string]map[string]*LockInfo {
	i.Lock()
	defer i.Unlock()
	return i.nodes
}

func (i *IndexLocker) Get(id string, name string) *LockInfo {
	i.Lock()
	defer i.Unlock()
	if names, nok := i.nodes[id]; nok {
		if info, iok := names[name]; iok {
			return info
		}
	}
	return nil
}

func (i *IndexLocker) Add(id string, name string) *LockInfo {
	i.Lock()
	defer i.Unlock()
	if _, ok := i.nodes[id]; !ok {
		i.nodes[id] = make(map[string]*LockInfo)
	}
	i.nodes[id][name] = NewLockInfo()
	return i.nodes[id][name]
}

func (i *IndexLocker) Error(id string, name string, err error) {
	if err == nil {
		return
	}
	i.Lock()
	defer i.Unlock()
	if names, nok := i.nodes[id]; nok {
		if info, iok := names[name]; iok {
			info.Error = err.Error()
			logger.Error(fmt.Sprintf("error during asynchronous indexing node=%s, index=%s: %s", id, name, err.Error()))
		}
	}
}

func (i *IndexLocker) Remove(id string, name string) {
	i.Lock()
	defer i.Unlock()
	if names, ok := i.nodes[id]; ok {
		delete(names, name)
		if len(names) == 0 {
			delete(i.nodes, id)
		}
	}
}

func (i *IndexLocker) RemoveOld(hours int) {
	i.Lock()
	defer i.Unlock()
	currTime := time.Now()
	expireTime := currTime.Add(time.Duration(hours*-1) * time.Hour)
	for id, names := range i.nodes {
		for n, info := range names {
			if info.CreatedOn.Before(expireTime) {
				logger.Error(fmt.Sprintf("Removing stale index lock: node=%s, index=%s", id, n))
				delete(names, n)
			}
		}
		if len(names) == 0 {
			delete(i.nodes, id)
		}
	}
}

// hacky function for debugging
func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

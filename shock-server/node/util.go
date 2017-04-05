package node

import (
    "sort"
    "sync"
)

type mappy map[string]bool

func IsInMappy(item string, mp mappy) bool {
	if _, ok := mp[item]; ok {
		return true
	}
	return false
}

var virtIdx = mappy{"size": true}

var (
	LockMgr = NewLocker()
)

type Locker struct {
	nLock map[string]*NodeLock
}

type NodeLock struct {
	sync.Mutex
}

func NewLocker() *Locker {
	return &Locker{
		nLock: map[string]*NodeLock{},
	}
}

func (l *Locker) LockNode(id string) {
	// add if missing, may happen if shock restarted
	if _, ok := l.nLock[id]; !ok {
		l.nLock[id] = new(NodeLock)
	}
	l.nLock[id].Lock()
}

func (l *Locker) UnlockNode(id string) {
	// skip missing id
	if _, ok := l.nLock[id]; ok {
		l.nLock[id].Unlock()
	}
}

func (l *Locker) AddNode(id string) {
	if _, ok := l.nLock[id]; !ok {
		l.nLock[id] = new(NodeLock)
	}
}

func (l *Locker) RemoveNode(id string) {
	delete(l.nLock, id)
}

func (l *Locker) GetNodes() (ids []string) {
	for id, _ := range l.nLock {
		ids = append(ids, id)
	}
	return
}

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

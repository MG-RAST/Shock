package node

import ()

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
	partLock chan bool //semaphore for checkout (mutual exclusion between different clients)
}

func NewLocker() *Locker {
	return &Locker{
		partLock: make(chan bool, 1), //non-blocking buffered channel
	}
}

func (l *Locker) LockPartOp() {
	l.partLock <- true
}

func (l *Locker) UnlockPartOp() {
	<-l.partLock
}

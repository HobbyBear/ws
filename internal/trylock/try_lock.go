package trylock

import (
	"sync/atomic"
)

const (
	LockedFlag   int32 = 1
	UnlockedFlag int32 = 0
)

type Mutex struct {
	status *int32
}

func NewMutex() *Mutex {
	status := UnlockedFlag
	return &Mutex{
		status: &status,
	}
}

func (m *Mutex) Unlock() {
	atomic.StoreInt32(m.status, UnlockedFlag)
}

func (m *Mutex) TryLock() bool {
	if atomic.CompareAndSwapInt32(m.status, UnlockedFlag, LockedFlag) {
		return true
	}
	return false
}

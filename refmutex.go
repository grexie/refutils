package refutils

import "sync"

type RefMutex struct {
	mutex     sync.Mutex
	refMutex  sync.Mutex
	refCount  uint64
	refLocker *refLocker
}

func (rm *RefMutex) Lock() {
	rm.mutex.Lock()
}

func (rm *RefMutex) Unlock() {
	rm.mutex.Unlock()
}

func (rm *RefMutex) RefLock() {
	rm.refMutex.Lock()
	defer rm.refMutex.Unlock()

	if rm.refCount == 0 {
		rm.mutex.Lock()
	}
	rm.refCount++
}

func (rm *RefMutex) RefUnlock() {
	rm.refMutex.Lock()
	defer rm.refMutex.Unlock()

	rm.refCount--
	if rm.refCount == 0 {
		rm.mutex.Unlock()
	}
}

func (rm *RefMutex) RefLocker() sync.Locker {
	if rm.refLocker == nil {
		rm.refLocker = &refLocker{rm}
	}
	return rm.refLocker
}

type refLocker struct {
	refMutex *RefMutex
}

func (rl *refLocker) Lock() {
	rl.refMutex.RefLock()
}

func (rl *refLocker) Unlock() {
	rl.refMutex.RefUnlock()
}

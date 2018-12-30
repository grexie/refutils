package refutils

import "sync"

// RefMutex provides a mutual exclusion lock with two sync.Locker interfaces. The master lock takes a global lock on
// the RefMutex, whereas RefLocker (and corresponding RefLock and RefUnlock methods) allow for multiple locks even when
// the global lock has been requested.
//
// This differs from sync.RWLock in that once a ref-lock has been obtained, further ref-locks are guaranteed until all
// ref-locks have been released. Consequently, once a ref-lock has been obtained, no master-locks can be obtained until
// all ref-locks have been released. Finally, only one master-lock may exist at any one time. If a master-lock is
// obtained then ref-locks will block until the master-lock is released.
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

	if rm.refCount == 0 {
		panic("ref-unlocked of ref-unlocked mutex")
	}

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

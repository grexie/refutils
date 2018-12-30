package refutils

import (
	"testing"
	"time"
)

func waitLockFail(t *testing.T, ch chan bool, timeout int, message string) {
	select {
	case <-time.After(time.Duration(timeout) * time.Millisecond):
	case <-ch:
		t.Error(message)
	}
}

func waitLockPass(t *testing.T, ch chan bool, timeout int, message string) {
	select {
	case <-time.After(time.Duration(timeout) * time.Millisecond):
		t.Error(message)
	case <-ch:
	}
}

func TestRefLocking(t *testing.T) {
	m := RefMutex{}
	l := m.RefLocker()

	ch := make(chan bool, 1)

	l.Lock()
	l.Lock()
	l.Lock()

	go func() {
		m.Lock()
		ch <- true
	}()

	waitLockFail(t, ch, 200, "obtained master-lock while ref-locked")

	l.Unlock()
	l.Unlock()

	waitLockFail(t, ch, 200, "obtained master-lock while ref-locked")

	l.Unlock()

	waitLockPass(t, ch, 200, "failed to obtain master-lock while not ref-locked")
}

func TestMasterLocking(t *testing.T) {
	m := RefMutex{}

	ch := make(chan bool, 1)

	m.Lock()

	for i := 0; i < 3; i++ {
		go func() {
			m.RefLock()
			ch <- true
		}()
	}

	waitLockFail(t, ch, 200, "obtained ref-lock while master-locked")

	m.Unlock()

	waitLockPass(t, ch, 200, "failed to obtain ref-lock while not master-locked")
	waitLockPass(t, ch, 200, "failed to obtain ref-lock while not master-locked")
	waitLockPass(t, ch, 200, "failed to obtain ref-lock while not master-locked")
}

func TestOverUnlock(t *testing.T) {
	m := RefMutex{}

	defer func() {
		if r := recover(); r == nil {
			t.Error("able to ref-unlock an unlocked mutex")
		}
	}()

	m.RefUnlock()
}

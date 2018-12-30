package refutils

import (
	"runtime"
	"testing"
	"time"
)

type MyStruct struct {
	RefHolder
}

func TestRefMapConsistency(t *testing.T) {
	r := NewRefMap("test")

	s := &MyStruct{}
	id := r.Ref(s)
	if r.Length() != 1 {
		t.Error("length of map should be 1")
	}
	r.Unref(s)
	if r.Length() != 0 {
		t.Error("length of map should be 0")
	}
	if r.Ref(s) != id {
		t.Errorf("inconsistent id: %d", id)
	}
	if r.Length() != 1 {
		t.Error("length of map should be 1")
	}
}

func TestRefMapConsistencyAfterRelease(t *testing.T) {
	r := NewRefMap("test")

	s := &MyStruct{}
	id := r.Ref(s)
	r.Ref(s)
	if r.Length() != 1 {
		t.Error("length of map should be 1")
	}
	r.Release(s)
	if r.Length() != 0 {
		t.Error("length of map should be 0")
	}
	if r.Ref(s) != id {
		t.Errorf("inconsistent id: %d", id)
	}
	if r.Length() != 1 {
		t.Error("length of map should be 1")
	}
}

func TestMultipleReferences(t *testing.T) {
	r1 := NewRefMap("test1")
	r2 := NewRefMap("test2")

	s1 := &MyStruct{}
	s2 := &MyStruct{}
	id1 := r1.Ref(s1)
	id2 := r2.Ref(s1)
	id3 := r1.Ref(s2)
	id4 := r2.Ref(s2)

	if s1.GetID("test1") != id1 {
		t.Error("ids don't match")
	}

	if s1.GetID("test2") != id2 {
		t.Error("ids don't match")
	}

	if s2.GetID("test1") != id3 {
		t.Error("ids don't match")
	}

	if s2.GetID("test2") != id4 {
		t.Error("ids don't match")
	}
}

func TestRefMapMultipleUnref(t *testing.T) {
	r := NewRefMap("test")

	s := &MyStruct{}
	r.Ref(s)
	r.Ref(s)
	if r.Length() != 1 {
		t.Error("length of map should be 1")
	}
	r.Unref(s)
	if r.Length() != 1 {
		t.Error("length of map should be 1")
	}
	r.Unref(s)
	if r.Length() != 0 {
		t.Error("length of map should be 0")
	}
}

func TestRefMapOverUnref(t *testing.T) {
	r := NewRefMap("test")

	s := &MyStruct{}
	r.Ref(s)
	r.Ref(s)
	if r.Length() != 1 {
		t.Error("length of map should be 1")
	}
	r.Unref(s)
	if r.Length() != 1 {
		t.Error("length of map should be 1")
	}
	r.Unref(s)
	if r.Length() != 0 {
		t.Error("length of map should be 0")
	}
}

func TestWeakRefMapIsWeak(t *testing.T) {
	r := NewWeakRefMap("test")

	ch := make(chan bool, 1)

	s := &MyStruct{}
	runtime.SetFinalizer(s, func(s *MyStruct) {
		ch <- true
	})
	r.Ref(s)
	s = nil
	runtime.GC()

	select {
	case <-time.After(time.Duration(200) * time.Millisecond):
		t.Error("timeout waiting for finalizer")
	case <-ch:
	}

	r.ReleaseAll()
}

func TestWeakRefMapIsStrong(t *testing.T) {
	r := NewRefMap("test")

	ch := make(chan bool, 1)

	s := &MyStruct{}
	runtime.SetFinalizer(s, func(s *MyStruct) {
		ch <- true
	})
	r.Ref(s)
	s = nil
	runtime.GC()

	select {
	case <-time.After(time.Duration(200) * time.Millisecond):
	case <-ch:
		t.Error("reference was garbage collected")
	}

	r.ReleaseAll()
}

func TestRefMapGet(t *testing.T) {
	r := NewWeakRefMap("test")

	s := &MyStruct{}
	id := r.Ref(s)
	s2 := r.Get(id).(*MyStruct)
	if s != s2 {
		t.Error("returned interface differs to the one added")
	}
}

func TestRefMapReleaseAll(t *testing.T) {
	r := NewRefMap("test")

	s1 := &MyStruct{}
	s2 := &MyStruct{}
	s3 := &MyStruct{}

	r.Ref(s1)
	r.Ref(s2)
	r.Ref(s3)

	if r.Length() != 3 {
		t.Error("length of map should be 3")
	}

	r.ReleaseAll()

	if r.Length() != 0 {
		t.Error("length of map should be 0")
	}
}

func TestRefMapRefs(t *testing.T) {
	r := NewRefMap("test")

	s1 := &MyStruct{}
	s2 := &MyStruct{}
	s3 := &MyStruct{}

	id1 := r.Ref(s1)
	id2 := r.Ref(s2)
	id3 := r.Ref(s3)

	refs := r.Refs()

	if len(refs) != 3 {
		t.Error("length of map should be 3")
	}

	r.ReleaseAll()

	if len(refs) != 3 {
		t.Error("length of map should be 3")
	}

	if refs[id1] != s1 {
		t.Error("reference s1 doesn't match refs[id1]")
	}

	if refs[id2] != s2 {
		t.Error("reference s2 doesn't match refs[id2]")
	}

	if refs[id3] != s3 {
		t.Error("reference s3 doesn't match refs[id3]")
	}
}

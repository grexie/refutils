package refutils

import (
	"reflect"
	"sync"
	"unsafe"
)

type id uint32

type Ref interface {
	GetID(name string) id
	SetID(name string, id id)
}

type RefHolder struct {
	ids      map[string]id
	idsMutex sync.Mutex
}

func (o *RefHolder) GetID(name string) id {
	o.idsMutex.Lock()
	defer o.idsMutex.Unlock()

	if o.ids == nil {
		return 0
	}

	if id, ok := o.ids[name]; !ok {
		return 0
	} else {
		return id
	}
}

func (o *RefHolder) SetID(name string, i id) {
	o.idsMutex.Lock()
	defer o.idsMutex.Unlock()

	if o.ids == nil {
		o.ids = map[string]id{}
	}

	o.ids[name] = i
}

type refMapEntry struct {
	pointer uintptr
	ref     unsafe.Pointer
	count   uint32
}

type RefMap struct {
	name    string
	entries map[id]*refMapEntry
	refType reflect.Type
	nextId  id
	mutex   sync.RWMutex
	strong  bool
}

func NewRefMap(name string, RefType reflect.Type) *RefMap {
	return &RefMap{
		name:    name,
		entries: map[id]*refMapEntry{},
		refType: RefType,
		strong:  true,
	}
}

func NewWeakRefMap(name string, RefType reflect.Type) *RefMap {
	return &RefMap{
		name:    name,
		entries: map[id]*refMapEntry{},
		refType: RefType,
		strong:  false,
	}
}

func (rm *RefMap) Refs() map[id]Ref {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	refs := map[id]Ref{}
	for i, entry := range rm.entries {
		// nolint: vet
		refs[i] = reflect.NewAt(rm.refType.Elem(), unsafe.Pointer(entry.pointer)).Interface().(Ref)
	}
	return refs
}

func (rm *RefMap) Length() int {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	return len(rm.entries)
}

func (rm *RefMap) Get(id id) Ref {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	// nolint: vet
	return reflect.NewAt(rm.refType.Elem(), unsafe.Pointer(rm.entries[id].pointer)).Interface().(Ref)
}

func (rm *RefMap) Ref(r Ref) id {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	var e *refMapEntry
	id := r.GetID(rm.name)
	if id == 0 {
		rm.nextId++
		id = rm.nextId
		r.SetID(rm.name, id)
	}
	if e = rm.entries[id]; e == nil {
		if rm.strong {
			e = &refMapEntry{reflect.ValueOf(r).Pointer(), unsafe.Pointer(reflect.ValueOf(r).Pointer()), 0}
			rm.entries[id] = e
		} else {
			e = &refMapEntry{reflect.ValueOf(r).Pointer(), nil, 0}
			rm.entries[id] = e
		}
	}
	e.count++
	return id
}

func (rm *RefMap) Unref(r Ref) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	id := r.GetID(rm.name)
	if e, ok := rm.entries[id]; ok && e.count <= 1 {
		delete(rm.entries, id)
	} else if ok {
		e.count--
	}
}

func (rm *RefMap) Release(r Ref) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	id := r.GetID(rm.name)
	if _, ok := rm.entries[id]; ok {
		delete(rm.entries, id)
	}
}

func (rm *RefMap) Clear() {
	rm.entries = nil
}

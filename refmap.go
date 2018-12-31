package refutils

import (
	"reflect"
	"sync"
	"unsafe"
)

type ID uint64

type Ref interface {
	getID(name string) ID
	setID(name string, id ID)
}

type RefHolder struct {
	ids      map[string]ID
	idsMutex sync.Mutex
}

func (o *RefHolder) getID(name string) ID {
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

func (o *RefHolder) setID(name string, i ID) {
	o.idsMutex.Lock()
	defer o.idsMutex.Unlock()

	if o.ids == nil {
		o.ids = map[string]ID{}
	}

	o.ids[name] = i
}

type refMapEntry struct {
	pointer uintptr
	ref     Ref
	refType reflect.Type
	count   uint64
}

// RefMap provides a thread-safe strong and weak reference map for objects implementing Ref (GetID and SetID for uint64
// IDs). The RefObject struct can be inherited by your structs to implement the required maps and mechanisms for RefMap
// to work. RefObject implemented the Ref interface.
//
// RefMap will set a monotonically increasing identifier in the Ref that it is provided by Ref and Unref. Once an object
// is referenced using Ref, it will remain referenced within the map until it is dereferenced using Unref.
type RefMap struct {
	name    string
	entries map[ID]*refMapEntry
	nextId  ID
	mutex   sync.RWMutex
	strong  bool
}

func NewRefMap(name string) *RefMap {
	return &RefMap{
		name:    name,
		entries: map[ID]*refMapEntry{},
		strong:  true,
	}
}

func NewWeakRefMap(name string) *RefMap {
	return &RefMap{
		name:    name,
		entries: map[ID]*refMapEntry{},
		strong:  false,
	}
}

func (rm *RefMap) Refs() map[ID]Ref {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	refs := map[ID]Ref{}
	for i, entry := range rm.entries {
		// nolint: vet
		refs[i] = reflect.NewAt(entry.refType, unsafe.Pointer(entry.pointer)).Interface().(Ref)
	}
	return refs
}

func (rm *RefMap) Length() int {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	return len(rm.entries)
}

func (rm *RefMap) Get(id ID) Ref {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	entry := rm.entries[id]

	// nolint: vet
	return reflect.NewAt(entry.refType, unsafe.Pointer(entry.pointer)).Interface().(Ref)
}

func (rm *RefMap) GetID(r Ref) ID {
	return r.getID(rm.name)
}

func (rm *RefMap) Ref(r Ref) ID {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	var e *refMapEntry
	id := r.getID(rm.name)
	if id == 0 {
		rm.nextId++
		id = rm.nextId
		r.setID(rm.name, id)
	}
	if e = rm.entries[id]; e == nil {
		refType := reflect.TypeOf(r).Elem()
		if rm.strong {
			e = &refMapEntry{reflect.ValueOf(r).Pointer(), r, refType, 0}
			rm.entries[id] = e
		} else {
			e = &refMapEntry{reflect.ValueOf(r).Pointer(), nil, refType, 0}
			rm.entries[id] = e
		}
	}
	e.count++
	return id
}

func (rm *RefMap) Unref(r Ref) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	id := r.getID(rm.name)
	if e, ok := rm.entries[id]; ok && e.count <= 1 {
		delete(rm.entries, id)
	} else if ok {
		e.count--
	}
}

func (rm *RefMap) Release(r Ref) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	id := r.getID(rm.name)
	if _, ok := rm.entries[id]; ok {
		delete(rm.entries, id)
	}
}

func (rm *RefMap) ReleaseAll() {
	rm.entries = map[ID]*refMapEntry{}
}

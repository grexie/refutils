# RefUtils Go Library

Collection of utility interfaces for use with referencing of objects. Used heavily in [go-v8](https://github.com/behrsin/go-v8).

## Usage

```
import (
  refutils "github.com/behrsin/go-refutils"
)
```

## RefMutex

Offers a mutual exclusion lock with two sync.Locker interfaces. The master lock takes a global lock on the RefMutex, whereas `RefLocker` (and corresponding `RefLock` and `RefUnlock` methods) allow for multiple locks even when the global lock has been requested. This differs from `sync.RWLock` in that once a ref-lock has been obtained, further ref-locks are guaranteed until all ref-locks have been released. Consequently, once a ref-lock has been obtained, no master-locks can be obtained until all ref-locks have been released. Finally, only one master-lock may exist at any one time. If a master-lock is obtained then ref-locks will block until the master-lock is released.

## RefMap

Provides a thread-safe strong and weak reference map for objects implementing `Ref` (`GetID` and `SetID` for `uint32` IDs). The `RefObject` struct can be inherited by your structs to implement the required maps and mechanisms for RefMap to work. `RefObject` implemented the Ref interface.

`RefMap` will set a monotonically increasing identifier in the `Ref` that it is provided by `Ref` and `Unref`. Once an object is referenced using `Ref`, it will remain referenced within the map until it is dereferenced using `Unref`.

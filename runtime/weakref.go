// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package grumpy

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

var (
	// WeakRefType is the object representing the Python 'weakref' type.
	WeakRefType = newBasisType("weakref", reflect.TypeOf(WeakRef{}), toWeakRefUnsafe, ObjectType)
)

type weakRefState int

const (
	weakRefStateNew weakRefState = iota
	weakRefStateUsed
	weakRefStateDead
)

// WeakRef represents Python 'weakref' objects.
type WeakRef struct {
	Object
	ptr       uintptr
	mutex     sync.Mutex
	state     weakRefState
	callbacks []*Object
	hash      *Object
}

func toWeakRefUnsafe(o *Object) *WeakRef {
	return (*WeakRef)(o.toPointer())
}

// get returns r's referent, or nil if r is "dead".
func (r *WeakRef) get() *Object {
	if r.state == weakRefStateDead {
		return nil
	}
	r.state = weakRefStateUsed
	return (*Object)(unsafe.Pointer(r.ptr))
}

// ToObject upcasts r to an Object.
func (r *WeakRef) ToObject() *Object {
	return &r.Object
}

func weakRefCall(f *Frame, callable *Object, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkFunctionArgs(f, "__call__", args); raised != nil {
		return nil, raised
	}
	r := toWeakRefUnsafe(callable)
	r.mutex.Lock()
	o := r.get()
	r.mutex.Unlock()
	if o == nil {
		o = None
	}
	return o, nil
}

func weakRefHash(f *Frame, o *Object) (result *Object, raised *BaseException) {
	r := toWeakRefUnsafe(o)
	var referent *Object
	r.mutex.Lock()
	if r.hash != nil {
		result = r.hash
	} else {
		referent = r.get()
	}
	r.mutex.Unlock()
	if referent != nil {
		var hash *Int
		hash, raised = Hash(f, referent)
		if raised == nil {
			result = hash.ToObject()
			r.mutex.Lock()
			r.hash = result
			r.mutex.Unlock()
		}
	} else if result == nil {
		raised = f.RaiseType(TypeErrorType, "weak object has gone away")
	}
	return result, raised
}

func weakRefNew(f *Frame, t *Type, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkFunctionVarArgs(f, "__new__", args, ObjectType); raised != nil {
		return nil, raised
	}
	argc := len(args)
	if argc > 2 {
		format := "__new__ expected at most 2 arguments, got %d"
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, argc))
	}
	o := args[0]
	nilPtr := unsafe.Pointer(nil)
	addr := (*unsafe.Pointer)(unsafe.Pointer(&o.ref))
	var r *WeakRef
	// Atomically fetch or initialize o.ref.
	for {
		p := atomic.LoadPointer(addr)
		if p != nilPtr {
			r = (*WeakRef)(p)
			break
		} else {
			r = &WeakRef{Object: Object{typ: WeakRefType}, ptr: uintptr(o.toPointer())}
			if atomic.CompareAndSwapPointer(addr, nilPtr, r.toPointer()) {
				runtime.SetFinalizer(o, weakRefFinalizeReferent)
				break
			}
		}
	}
	if argc > 1 {
		r.mutex.Lock()
		r.callbacks = append(r.callbacks, args[1])
		r.mutex.Unlock()
	}
	return r.ToObject(), nil
}

func weakRefRepr(f *Frame, o *Object) (*Object, *BaseException) {
	r := toWeakRefUnsafe(o)
	r.mutex.Lock()
	p := r.get()
	r.mutex.Unlock()
	s := "dead"
	if p != nil {
		s = fmt.Sprintf("to '%s' at %p", p.Type().Name(), p)
	}
	return NewStr(fmt.Sprintf("<weakref at %p; %s>", r, s)).ToObject(), nil
}

func initWeakRefType(map[string]*Object) {
	WeakRefType.slots.Call = &callSlot{weakRefCall}
	WeakRefType.slots.Hash = &unaryOpSlot{weakRefHash}
	WeakRefType.slots.New = &newSlot{weakRefNew}
	WeakRefType.slots.Repr = &unaryOpSlot{weakRefRepr}
}

func weakRefFinalizeReferent(o *Object) {
	// Note that although o should be the last reference to that object
	// (since this is its finalizer), in the time between the runtime
	// scheduling this finalizer and the Lock() call below, r may have
	// handed out another reference to o. So we can't simply mark r "dead".
	addr := (*unsafe.Pointer)(unsafe.Pointer(&o.ref))
	r := (*WeakRef)(atomic.LoadPointer(addr))
	numCallbacks := 0
	var callbacks []*Object
	r.mutex.Lock()
	switch r.state {
	case weakRefStateNew:
		// State "new" means that no references have been handed out by
		// r and therefore o is the only live reference.
		r.state = weakRefStateDead
		numCallbacks = len(r.callbacks)
		callbacks = make([]*Object, numCallbacks)
		copy(callbacks, r.callbacks)
	case weakRefStateUsed:
		// Most likely it's safe to mark r "dead" at this point, but
		// because a reference was handed out at some point, play it
		// safe and reset the finalizer. If no more references are
		// handed out before the next finalize then it will be "dead".
		r.state = weakRefStateNew
		runtime.SetFinalizer(o, weakRefFinalizeReferent)
	}
	r.mutex.Unlock()
	// Don't hold r.mutex while invoking callbacks in case they access r
	// and attempt to acquire the mutex.
	for i := numCallbacks - 1; i >= 0; i-- {
		f := NewRootFrame()
		if _, raised := callbacks[i].Call(f, Args{r.ToObject()}, nil); raised != nil {
			Stderr.writeString(FormatExc(f))
		}
	}
}

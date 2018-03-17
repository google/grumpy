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
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	argsCacheSize = 16
	argsCacheArgc = 6
)

type threadState struct {
	reprState    map[*Object]bool
	excValue     *BaseException
	excTraceback *Traceback
	// argsCache is a small, per-thread LIFO cache for arg lists. Entries
	// have a fixed capacity so calls to functions with larger parameter
	// lists will be allocated afresh each time. Args freed when the cache
	// is full are dropped. If the cache is empty then a new args slice
	// will be allocated.
	argsCache []Args

	// frameCache is a local cache of allocated frames almost ready for
	// reuse. The cache is maintained through the Frame `back` pointer as a
	// singly linked list.
	frameCache *Frame
}

func newThreadState() *threadState {
	return &threadState{argsCache: make([]Args, 0, argsCacheSize)}
}

// recursiveMutex implements a typical reentrant lock, similar to Python's
// RLock. Lock can be called multiple times for the same frame stack.
type recursiveMutex struct {
	mutex       sync.Mutex
	threadState *threadState
	count       int
}

func (m *recursiveMutex) Lock(f *Frame) {
	p := (*unsafe.Pointer)(unsafe.Pointer(&m.threadState))
	if (*threadState)(atomic.LoadPointer(p)) != f.threadState {
		// m.threadState != f.threadState implies m is not held by this
		// thread and therefore we won't deadlock acquiring the mutex.
		m.mutex.Lock()
		// m.threadState is now guaranteed to be empty (otherwise we
		// couldn't have acquired m.mutex) so store our own thread ID.
		atomic.StorePointer(p, unsafe.Pointer(f.threadState))
		m.count++
	} else {
		m.count++
	}
}

func (m *recursiveMutex) Unlock(f *Frame) {
	p := (*unsafe.Pointer)(unsafe.Pointer(&m.threadState))
	if (*threadState)(atomic.LoadPointer(p)) != f.threadState {
		logFatal("recursiveMutex.Unlock: frame did not match that passed to Lock")
	}
	// Since we're unlocking, we must hold m.mutex, so this is safe.
	if m.count <= 0 {
		logFatal("recursiveMutex.Unlock: Unlock called too many times")
	}
	m.count--
	if m.count == 0 {
		atomic.StorePointer(p, unsafe.Pointer(nil))
		m.mutex.Unlock()
	}
}

// TryableMutex is a mutex-like object that also supports TryLock().
type TryableMutex struct {
	c chan bool
}

// NewTryableMutex returns a new TryableMutex.
func NewTryableMutex() *TryableMutex {
	m := &TryableMutex{make(chan bool, 1)}
	m.Unlock()
	return m
}

// Lock blocks until the mutex is available and then acquires a lock.
func (m *TryableMutex) Lock() {
	<-m.c
}

// TryLock returns true and acquires a lock if the mutex is available, otherwise
// it returns false.
func (m *TryableMutex) TryLock() bool {
	select {
	case <-m.c:
		return true
	default:
		return false
	}
}

// Unlock releases the mutex's lock.
func (m *TryableMutex) Unlock() {
	m.c <- true
}

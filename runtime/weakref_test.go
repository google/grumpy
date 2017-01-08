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
	"runtime"
	"testing"
	"time"
)

func TestWeakRefCall(t *testing.T) {
	aliveRef, alive, deadRef := makeWeakRefsForTest()
	dupRef := newTestWeakRef(alive, nil)
	cases := []invokeTestCase{
		{args: wrapArgs(aliveRef), want: alive},
		{args: wrapArgs(dupRef), want: alive},
		{args: wrapArgs(deadRef), want: None},
		{args: wrapArgs(aliveRef, 123), wantExc: mustCreateException(TypeErrorType, "'__call__' requires 0 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(WeakRefType, "__call__", &cas); err != "" {
			t.Error(err)
		}
	}
	runtime.KeepAlive(alive)
}

func TestWeakRefHash(t *testing.T) {
	aliveRef, alive, deadRef := makeWeakRefsForTest()
	hashedRef, hashed, _ := makeWeakRefsForTest()
	if _, raised := Hash(NewRootFrame(), hashedRef.ToObject()); raised != nil {
		t.Fatal(raised)
	}
	runtime.KeepAlive(hashed)
	hashed = nil
	weakRefMustDie(hashedRef)
	unhashable := NewList().ToObject()
	unhashableRef := newTestWeakRef(unhashable, nil)
	cases := []invokeTestCase{
		{args: wrapArgs(aliveRef), want: NewInt(hashString("foo")).ToObject()},
		{args: wrapArgs(deadRef), wantExc: mustCreateException(TypeErrorType, "weak object has gone away")},
		{args: wrapArgs(hashedRef), want: NewInt(hashString("foo")).ToObject()},
		{args: wrapArgs(unhashableRef), wantExc: mustCreateException(TypeErrorType, "unhashable type: 'list'")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(WeakRefType, "__hash__", &cas); err != "" {
			t.Error(err)
		}
	}
	runtime.KeepAlive(alive)
	runtime.KeepAlive(unhashable)
}

func TestWeakRefNew(t *testing.T) {
	alive := NewStr("foo").ToObject()
	aliveRef := newTestWeakRef(alive, nil)
	cases := []invokeTestCase{
		{args: wrapArgs(alive), want: aliveRef.ToObject()},
		{wantExc: mustCreateException(TypeErrorType, "'__new__' requires 1 arguments")},
		{args: wrapArgs("foo", "bar", "baz"), wantExc: mustCreateException(TypeErrorType, "__new__ expected at most 2 arguments, got 3")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(WeakRefType.ToObject(), &cas); err != "" {
			t.Error(err)
		}
	}
	runtime.KeepAlive(alive)
}

func TestWeakRefNewCallback(t *testing.T) {
	callbackChannel := make(chan *WeakRef)
	callback := wrapFuncForTest(func(f *Frame, r *WeakRef) {
		callbackChannel <- r
	})
	r := newTestWeakRef(newObject(ObjectType), callback)
	weakRefMustDie(r)
	if r.get() != nil {
		t.Fatalf("expected weakref %v to be dead", r)
	}
	if callbackGot := <-callbackChannel; callbackGot != r {
		t.Fatalf("callback got %v, want %v", callbackGot, r)
	}
}

func TestWeakRefNewCallbackRaises(t *testing.T) {
	// It's not easy to verify that the exception is output properly, but
	// we can at least make sure the program doesn't blow up if the
	// callback raises.
	callback := wrapFuncForTest(func(f *Frame, r *WeakRef) *BaseException {
		return f.RaiseType(RuntimeErrorType, "foo")
	})
	r := newTestWeakRef(newObject(ObjectType), callback)
	weakRefMustDie(r)
	if r.get() != nil {
		t.Fatalf("expected weakref %v to be dead", r)
	}
}

func TestWeakRefStrRepr(t *testing.T) {
	aliveRef, alive, deadRef := makeWeakRefsForTest()
	cases := []invokeTestCase{
		{args: wrapArgs(aliveRef), want: NewStr(fmt.Sprintf("<weakref at %p; to 'str' at %p>", aliveRef, alive)).ToObject()},
		{args: wrapArgs(deadRef), want: NewStr(fmt.Sprintf("<weakref at %p; dead>", deadRef)).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(ToStr), &cas); err != "" {
			t.Error(err)
		}
		if err := runInvokeTestCase(wrapFuncForTest(Repr), &cas); err != "" {
			t.Error(err)
		}
	}
	runtime.KeepAlive(alive)
}

func newTestWeakRef(o, callback *Object) *WeakRef {
	args := Args{o}
	if callback != nil {
		args = Args{o, callback}
	}
	return toWeakRefUnsafe(mustNotRaise(WeakRefType.Call(NewRootFrame(), args, nil)))
}

func makeWeakRefsForTest() (*WeakRef, *Object, *WeakRef) {
	alive := NewStr("foo").ToObject()
	aliveRef := newTestWeakRef(alive, nil)
	dead := NewFloat(3.14).ToObject()
	deadRef := newTestWeakRef(dead, nil)
	dead = nil
	weakRefMustDie(deadRef)
	return aliveRef, alive, deadRef
}

func weakRefMustDie(r *WeakRef) {
	r.mutex.Lock()
	o := r.get()
	r.mutex.Unlock()
	if o == nil {
		return
	}
	doneChannel := make(chan bool)
	callback := wrapFuncForTest(func(f *Frame, r *WeakRef) {
		close(doneChannel)
	})
	mustNotRaise(WeakRefType.Call(NewRootFrame(), Args{o, callback}, nil))
	o = nil
	timeoutChannel := make(chan bool)
	go func() {
		// Finalizers run some time after GC, thus the Sleep call. In
		// our case o's finalizer will have to run twice, so loop and
		// GC repeatedly. In theory, twice should be enough, but in
		// practice there are race conditions and things to contend
		// with so just loop a bunch of times.
		wait := 10 * time.Millisecond
		for t := time.Duration(0); t < time.Second; t += wait {
			runtime.GC()
			time.Sleep(wait)
		}
		close(timeoutChannel)
	}()
	select {
	case <-doneChannel:
		return
	case <-timeoutChannel:
		panic(fmt.Sprintf("weakref %v did not die", r))
	}
}

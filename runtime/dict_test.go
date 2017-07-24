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
	"reflect"
	"regexp"
	"runtime"
	"sync"
	"testing"
	"time"
)

// hashFoo is the hash of the string 'foo'. We use this to validate some corner
// cases around hash collision below.
// NOTE: Inline func helps support 32bit systems.
var hashFoo = NewInt(func(i int64) int { return int(i) }(-4177197833195190597)).ToObject()

func TestNewStringDict(t *testing.T) {
	cases := []struct {
		m    map[string]*Object
		want *Dict
	}{
		{nil, NewDict()},
		{map[string]*Object{"baz": NewFloat(3.14).ToObject()}, newTestDict("baz", 3.14)},
		{map[string]*Object{"foo": NewInt(2).ToObject(), "bar": NewInt(4).ToObject()}, newTestDict("bar", 4, "foo", 2)},
	}
	for _, cas := range cases {
		fun := newBuiltinFunction("newStringDict", func(*Frame, Args, KWArgs) (*Object, *BaseException) {
			return newStringDict(cas.m).ToObject(), nil
		}).ToObject()
		if err := runInvokeTestCase(fun, &invokeTestCase{want: cas.want.ToObject()}); err != "" {
			t.Error(err)
		}
	}
}

func TestDictClear(t *testing.T) {
	clear := mustNotRaise(GetAttr(NewRootFrame(), DictType.ToObject(), NewStr("clear"), nil))
	fun := newBuiltinFunction("TestDictClear", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		if _, raised := clear.Call(f, args, nil); raised != nil {
			return nil, raised
		}
		return args[0], nil
	}).ToObject()
	cases := []invokeTestCase{
		{args: wrapArgs(NewDict()), want: NewDict().ToObject()},
		{args: wrapArgs(newStringDict(map[string]*Object{"foo": NewInt(1).ToObject()})), want: NewDict().ToObject()},
		{args: wrapArgs(newTestDict(2, None, "baz", 3.14)), want: NewDict().ToObject()},
		{args: wrapArgs(NewDict(), NewList()), wantExc: mustCreateException(TypeErrorType, "'clear' of 'dict' requires 1 arguments")},
		{args: wrapArgs(NewDict(), None), wantExc: mustCreateException(TypeErrorType, "'clear' of 'dict' requires 1 arguments")},
		{args: wrapArgs(None), wantExc: mustCreateException(TypeErrorType, "unbound method clear() must be called with dict instance as first argument (got NoneType instance instead)")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestDictContains(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(NewDict(), "foo"), want: False.ToObject()},
		{args: wrapArgs(newTestDict("foo", 1, "bar", 2), "foo"), want: True.ToObject()},
		{args: wrapArgs(newTestDict(3, "foo", "bar", 42), 42), want: False.ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(DictType, "__contains__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestDictDelItem(t *testing.T) {
	fun := newBuiltinFunction("TestDictDelItem", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		if raised := checkMethodArgs(f, "TestDictDelItem", args, DictType, ObjectType); raised != nil {
			return nil, raised
		}
		if raised := DelItem(f, args[0], args[1]); raised != nil {
			return nil, raised
		}
		return args[0], nil
	}).ToObject()
	testDict := newTestDict("a", 1, "b", 2, "c", 3)
	cases := []invokeTestCase{
		{args: wrapArgs(newTestDict("foo", 1), "foo"), want: NewDict().ToObject()},
		{args: wrapArgs(NewDict(), 10), wantExc: mustCreateException(KeyErrorType, "10")},
		{args: wrapArgs(testDict, "a"), want: newTestDict("b", 2, "c", 3).ToObject()},
		{args: wrapArgs(testDict, "c"), want: newTestDict("b", 2).ToObject()},
		{args: wrapArgs(testDict, "a"), wantExc: mustCreateException(KeyErrorType, "a")},
		{args: wrapArgs(NewDict(), NewList()), wantExc: mustCreateException(TypeErrorType, "unhashable type: 'list'")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestDictDelItemString(t *testing.T) {
	fun := newBuiltinFunction("TestDictDelItemString", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		if raised := checkMethodArgs(f, "TestDictDelItemString", args, DictType, StrType); raised != nil {
			return nil, raised
		}
		deleted, raised := toDictUnsafe(args[0]).DelItemString(f, toStrUnsafe(args[1]).Value())
		if raised != nil {
			return nil, raised
		}
		return newTestTuple(deleted, args[0]).ToObject(), nil
	}).ToObject()
	cases := []invokeTestCase{
		{args: wrapArgs(newTestDict("foo", 1), "foo"), want: newTestTuple(true, NewDict()).ToObject()},
		{args: wrapArgs(NewDict(), "qux"), want: newTestTuple(false, NewDict()).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestDictEqNE(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, v, w *Object) (*Object, *BaseException) {
		eq, raised := Eq(f, v, w)
		if raised != nil {
			return nil, raised
		}
		ne, raised := NE(f, v, w)
		if raised != nil {
			return nil, raised
		}
		valid := GetBool(eq == True.ToObject() && ne == False.ToObject() || eq == False.ToObject() && ne == True.ToObject()).ToObject()
		if raised := Assert(f, valid, NewStr("invalid values for __eq__ or __ne__").ToObject()); raised != nil {
			return nil, raised
		}
		return eq, nil
	})
	f := NewRootFrame()
	large1, large2 := NewDict(), NewDict()
	largeSize := 100
	for i := 0; i < largeSize; i++ {
		s, raised := ToStr(f, NewInt(i).ToObject())
		if raised != nil {
			t.Fatal(raised)
		}
		large1.SetItem(f, NewInt(i).ToObject(), s.ToObject())
		s, raised = ToStr(f, NewInt(largeSize-i-1).ToObject())
		if raised != nil {
			t.Fatal(raised)
		}
		large2.SetItem(f, NewInt(largeSize-i-1).ToObject(), s.ToObject())
	}
	o := newObject(ObjectType)
	cases := []invokeTestCase{
		{args: wrapArgs(NewDict(), NewDict()), want: True.ToObject()},
		{args: wrapArgs(NewDict(), newTestDict("foo", true)), want: False.ToObject()},
		{args: wrapArgs(newTestDict("foo", "foo"), newTestDict("foo", "foo")), want: True.ToObject()},
		{args: wrapArgs(newTestDict("foo", true), newTestDict("bar", true)), want: False.ToObject()},
		{args: wrapArgs(newTestDict("foo", true), newTestDict("foo", newObject(ObjectType))), want: False.ToObject()},
		{args: wrapArgs(newTestDict("foo", true, "bar", false), newTestDict("bar", true)), want: False.ToObject()},
		{args: wrapArgs(newTestDict("foo", o, "bar", o), newTestDict("foo", o, "bar", o)), want: True.ToObject()},
		{args: wrapArgs(newTestDict(2, None, "foo", o), newTestDict("foo", o, 2, None)), want: True.ToObject()},
		{args: wrapArgs(large1, large2), want: True.ToObject()},
		{args: wrapArgs(NewDict(), 123), want: False.ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestDictGet(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(NewDict(), "foo"), want: None},
		{args: wrapArgs(newTestDict("foo", 1, "bar", 2), "foo"), want: NewInt(1).ToObject()},
		{args: wrapArgs(newTestDict(3, "foo", "bar", 42), 42, "nope"), want: NewStr("nope").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(DictType, "get", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestDictGetItem(t *testing.T) {
	getItem := newBuiltinFunction("TestDictGetItem", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		if raised := checkFunctionArgs(f, "TestDictGetItem", args, DictType, ObjectType); raised != nil {
			return nil, raised
		}
		result, raised := toDictUnsafe(args[0]).GetItem(f, args[1])
		if raised == nil && result == nil {
			result = None
		}
		return result, raised
	}).ToObject()
	f := NewRootFrame()
	h, raised := Hash(f, NewStr("foo").ToObject())
	if raised != nil {
		t.Fatal(raised)
	}
	if b, raised := IsTrue(f, mustNotRaise(NE(f, h.ToObject(), hashFoo))); raised != nil {
		t.Fatal(raised)
	} else if b {
		t.Fatalf("hash('foo') = %v, want %v", h, hashFoo)
	}
	deletedItemDict := newTestDict(hashFoo, true, "foo", true)
	deletedItemDict.DelItem(f, hashFoo)
	cases := []invokeTestCase{
		{args: wrapArgs(NewDict(), "foo"), want: None},
		{args: wrapArgs(newStringDict(map[string]*Object{"foo": True.ToObject()}), "foo"), want: True.ToObject()},
		{args: wrapArgs(newTestDict(2, "bar", "baz", 3.14), 2), want: NewStr("bar").ToObject()},
		{args: wrapArgs(newTestDict(2, "bar", "baz", 3.14), 3), want: None},
		{args: wrapArgs(deletedItemDict, hashFoo), want: None},
		{args: wrapArgs(NewDict(), NewList()), wantExc: mustCreateException(TypeErrorType, "unhashable type: 'list'")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(getItem, &cas); err != "" {
			t.Error(err)
		}
	}
}

// BenchmarkDictGetItem is to keep an eye on the speed of contended dict access
// in a fast read loop.
func BenchmarkDictGetItem(b *testing.B) {
	d := newTestDict(
		"foo", 1,
		"bar", 2,
		None, 3,
		4, 5)
	k := NewInt(4).ToObject()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		f := NewRootFrame()
		var ret *Object
		var raised *BaseException
		for pb.Next() {
			ret, raised = d.GetItem(f, k)
		}
		runtime.KeepAlive(ret)
		runtime.KeepAlive(raised)
	})
}

func BenchmarkDictIterItems(b *testing.B) {
	bench := func(d *Dict) func(*testing.B) {
		return func(b *testing.B) {
			f := NewRootFrame()
			args := f.MakeArgs(1)
			args[0] = d.ToObject()
			b.ResetTimer()

			var ret *Object
			var raised *BaseException
			for i := 0; i < b.N; i++ {
				iter, _ := dictIterItems(f, args, nil)
				for {
					ret, raised = Next(f, iter)
					if raised != nil {
						if !raised.isInstance(StopIterationType) {
							b.Fatalf("iteration failed with: %v", raised)
						}
						f.RestoreExc(nil, nil)
						break
					}
				}
			}
			runtime.KeepAlive(ret)
			runtime.KeepAlive(raised)
		}
	}

	b.Run("0-elements", bench(newTestDict()))
	b.Run("1-elements", bench(newTestDict(1, 2)))
	b.Run("2-elements", bench(newTestDict(1, 2, 3, 4)))
	b.Run("3-elements", bench(newTestDict(1, 2, 3, 4, 5, 6)))
	b.Run("4-elements", bench(newTestDict(1, 2, 3, 4, 5, 6, 7, 8)))
	b.Run("5-elements", bench(newTestDict(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)))
	b.Run("6-elements", bench(newTestDict(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12)))
	b.Run("7-elements", bench(newTestDict(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14)))
	b.Run("8-elements", bench(newTestDict(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16)))
}

func BenchmarkDictIterKeys(b *testing.B) {
	bench := func(d *Dict) func(*testing.B) {
		return func(b *testing.B) {
			f := NewRootFrame()
			args := f.MakeArgs(1)
			args[0] = d.ToObject()
			b.ResetTimer()

			var ret *Object
			var raised *BaseException
			for i := 0; i < b.N; i++ {
				iter, _ := dictIterKeys(f, args, nil)
				for {
					ret, raised = Next(f, iter)
					if raised != nil {
						if !raised.isInstance(StopIterationType) {
							b.Fatalf("iteration failed with: %v", raised)
						}
						f.RestoreExc(nil, nil)
						break
					}
				}
			}
			runtime.KeepAlive(ret)
			runtime.KeepAlive(raised)
		}
	}

	b.Run("0-elements", bench(newTestDict()))
	b.Run("1-elements", bench(newTestDict(1, 2)))
	b.Run("2-elements", bench(newTestDict(1, 2, 3, 4)))
	b.Run("3-elements", bench(newTestDict(1, 2, 3, 4, 5, 6)))
	b.Run("4-elements", bench(newTestDict(1, 2, 3, 4, 5, 6, 7, 8)))
	b.Run("5-elements", bench(newTestDict(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)))
	b.Run("6-elements", bench(newTestDict(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12)))
	b.Run("7-elements", bench(newTestDict(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14)))
	b.Run("8-elements", bench(newTestDict(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16)))
}

func BenchmarkDictIterValues(b *testing.B) {
	bench := func(d *Dict) func(*testing.B) {
		return func(b *testing.B) {
			f := NewRootFrame()
			args := f.MakeArgs(1)
			args[0] = d.ToObject()
			b.ResetTimer()

			var ret *Object
			var raised *BaseException
			for i := 0; i < b.N; i++ {
				iter, _ := dictIterValues(f, args, nil)
				for {
					ret, raised = Next(f, iter)
					if raised != nil {
						if !raised.isInstance(StopIterationType) {
							b.Fatalf("iteration failed with: %v", raised)
						}
						f.RestoreExc(nil, nil)
						break
					}
				}
			}
			runtime.KeepAlive(ret)
			runtime.KeepAlive(raised)
		}
	}

	b.Run("0-elements", bench(newTestDict()))
	b.Run("1-elements", bench(newTestDict(1, 2)))
	b.Run("2-elements", bench(newTestDict(1, 2, 3, 4)))
	b.Run("3-elements", bench(newTestDict(1, 2, 3, 4, 5, 6)))
	b.Run("4-elements", bench(newTestDict(1, 2, 3, 4, 5, 6, 7, 8)))
	b.Run("5-elements", bench(newTestDict(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)))
	b.Run("6-elements", bench(newTestDict(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12)))
	b.Run("7-elements", bench(newTestDict(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14)))
	b.Run("8-elements", bench(newTestDict(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16)))
}

func TestDictGetItemString(t *testing.T) {
	getItemString := newBuiltinFunction("TestDictGetItemString", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		if raised := checkFunctionArgs(f, "TestDictGetItem", args, DictType, StrType); raised != nil {
			return nil, raised
		}
		result, raised := toDictUnsafe(args[0]).GetItemString(f, toStrUnsafe(args[1]).Value())
		if raised == nil && result == nil {
			result = None
		}
		return result, raised
	}).ToObject()
	cases := []invokeTestCase{
		{args: wrapArgs(NewDict(), "foo"), want: None},
		{args: wrapArgs(newTestDict("foo", true), "foo"), want: True.ToObject()},
		{args: wrapArgs(newTestDict(2, "bar", "baz", 3.14), "baz"), want: NewFloat(3.14).ToObject()},
		{args: wrapArgs(newTestDict(2, "bar", "baz", 3.14), "qux"), want: None},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(getItemString, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestDictHasKey(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(NewDict(), "foo"), want: False.ToObject()},
		{args: wrapArgs(newTestDict("foo", 1, "bar", 2), "foo"), want: True.ToObject()},
		{args: wrapArgs(newTestDict(3, "foo", "bar", 42), 42), want: False.ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(DictType, "has_key", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestDictItemIteratorIter(t *testing.T) {
	iter := &newDictItemIterator(NewDict()).Object
	cas := &invokeTestCase{args: wrapArgs(iter), want: iter}
	if err := runInvokeMethodTestCase(dictItemIteratorType, "__iter__", cas); err != "" {
		t.Error(err)
	}
}

func TestDictItemIterModified(t *testing.T) {
	f := NewRootFrame()
	iterItems := mustNotRaise(GetAttr(f, DictType.ToObject(), NewStr("iteritems"), nil))
	d := NewDict()
	iter := mustNotRaise(iterItems.Call(f, wrapArgs(d), nil))
	if raised := d.SetItemString(f, "foo", None); raised != nil {
		t.Fatal(raised)
	}
	cas := invokeTestCase{
		args:    wrapArgs(iter),
		wantExc: mustCreateException(RuntimeErrorType, "dictionary changed during iteration"),
	}
	if err := runInvokeMethodTestCase(dictItemIteratorType, "next", &cas); err != "" {
		t.Error(err)
	}
}

func TestDictIter(t *testing.T) {
	iter := newBuiltinFunction("TestDictIter", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		if raised := checkFunctionArgs(f, "TestDictIter", args, DictType); raised != nil {
			return nil, raised
		}
		iter, raised := Iter(f, args[0])
		if raised != nil {
			return nil, raised
		}
		return TupleType.Call(f, []*Object{iter}, nil)
	}).ToObject()
	f := NewRootFrame()
	deletedItemDict := newTestDict(hashFoo, None, "foo", None)
	deletedItemDict.DelItem(f, hashFoo)
	cases := []invokeTestCase{
		{args: wrapArgs(NewDict()), want: NewTuple().ToObject()},
		{args: wrapArgs(newStringDict(map[string]*Object{"foo": NewInt(1).ToObject(), "bar": NewInt(2).ToObject()})), want: newTestTuple("foo", "bar").ToObject()},
		{args: wrapArgs(newTestDict(123, True, "foo", False)), want: newTestTuple(123, "foo").ToObject()},
		{args: wrapArgs(deletedItemDict), want: newTestTuple("foo").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(iter, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestDictIterKeys(t *testing.T) {
	iterkeys := mustNotRaise(GetAttr(NewRootFrame(), DictType.ToObject(), NewStr("iterkeys"), nil))
	fun := wrapFuncForTest(func(f *Frame, args ...*Object) (*Object, *BaseException) {
		iter, raised := iterkeys.Call(f, args, nil)
		if raised != nil {
			return nil, raised
		}
		return TupleType.Call(f, Args{iter}, nil)
	})
	cases := []invokeTestCase{
		{args: wrapArgs(NewDict()), want: NewTuple().ToObject()},
		{args: wrapArgs(newTestDict("foo", 1, "bar", 2)), want: newTestTuple("foo", "bar").ToObject()},
		{args: wrapArgs(NewDict(), "bad"), wantExc: mustCreateException(TypeErrorType, "'iterkeys' of 'dict' requires 1 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestDictIterValues(t *testing.T) {
	itervalues := mustNotRaise(GetAttr(NewRootFrame(), DictType.ToObject(), NewStr("itervalues"), nil))
	fun := wrapFuncForTest(func(f *Frame, args ...*Object) (*Object, *BaseException) {
		iter, raised := itervalues.Call(f, args, nil)
		if raised != nil {
			return nil, raised
		}
		return TupleType.Call(f, Args{iter}, nil)
	})
	cases := []invokeTestCase{
		{args: wrapArgs(NewDict()), want: NewTuple().ToObject()},
		{args: wrapArgs(newTestDict("foo", 1, "bar", 2)), want: newTestTuple(1, 2).ToObject()},
		{args: wrapArgs(NewDict(), "bad"), wantExc: mustCreateException(TypeErrorType, "'itervalues' of 'dict' requires 1 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

// Tests dict.items and dict.iteritems.
func TestDictItems(t *testing.T) {
	f := NewRootFrame()
	iterItems := mustNotRaise(GetAttr(f, DictType.ToObject(), NewStr("iteritems"), nil))
	items := newBuiltinFunction("TestDictIterItems", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		if raised := checkFunctionArgs(f, "TestDictIterItems", args, DictType); raised != nil {
			return nil, raised
		}
		iter, raised := iterItems.Call(f, []*Object{args[0]}, nil)
		if raised != nil {
			return nil, raised
		}
		return ListType.Call(f, []*Object{iter}, nil)
	}).ToObject()
	deletedItemDict := newTestDict(hashFoo, None, "foo", None)
	deletedItemDict.DelItem(f, hashFoo)
	cases := []invokeTestCase{
		{args: wrapArgs(NewDict()), want: NewList().ToObject()},
		{args: wrapArgs(newStringDict(map[string]*Object{"foo": NewInt(1).ToObject(), "bar": NewInt(2).ToObject()})), want: newTestList(newTestTuple("foo", 1), newTestTuple("bar", 2)).ToObject()},
		{args: wrapArgs(newTestDict(123, True, "foo", False)), want: newTestList(newTestTuple(123, true), newTestTuple("foo", false)).ToObject()},
		{args: wrapArgs(deletedItemDict), want: newTestList(newTestTuple("foo", None)).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(items, &cas); err != "" {
			t.Error(err)
		}
		if err := runInvokeMethodTestCase(DictType, "items", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestDictKeyIteratorIter(t *testing.T) {
	iter := &newDictKeyIterator(NewDict()).Object
	cas := &invokeTestCase{args: wrapArgs(iter), want: iter}
	if err := runInvokeMethodTestCase(dictKeyIteratorType, "__iter__", cas); err != "" {
		t.Error(err)
	}
}

func TestDictKeyIterModified(t *testing.T) {
	f := NewRootFrame()
	d := NewDict()
	iter := mustNotRaise(Iter(f, d.ToObject()))
	if raised := d.SetItemString(f, "foo", None); raised != nil {
		t.Fatal(raised)
	}
	cas := invokeTestCase{
		args:    wrapArgs(iter),
		wantExc: mustCreateException(RuntimeErrorType, "dictionary changed during iteration"),
	}
	if err := runInvokeMethodTestCase(dictKeyIteratorType, "next", &cas); err != "" {
		t.Error(err)
	}
}

func TestDictKeys(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(NewDict()), want: NewList().ToObject()},
		{args: wrapArgs(newTestDict("foo", None, 42, None)), want: newTestList(42, "foo").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(DictType, "keys", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestDictPop(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(newTestDict("foo", 42), "foo"), want: NewInt(42).ToObject()},
		{args: wrapArgs(NewDict(), "foo", 42), want: NewInt(42).ToObject()},
		{args: wrapArgs(NewDict(), "foo"), wantExc: mustCreateException(KeyErrorType, "foo")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(DictType, "pop", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestDictPopItem(t *testing.T) {
	popItem := mustNotRaise(GetAttr(NewRootFrame(), DictType.ToObject(), NewStr("popitem"), nil))
	fun := wrapFuncForTest(func(f *Frame, d *Dict) (*Object, *BaseException) {
		result := NewDict()
		item, raised := popItem.Call(f, wrapArgs(d), nil)
		for ; raised == nil; item, raised = popItem.Call(f, wrapArgs(d), nil) {
			t := toTupleUnsafe(item)
			result.SetItem(f, t.GetItem(0), t.GetItem(1))
		}
		if raised != nil {
			if !raised.isInstance(KeyErrorType) {
				return nil, raised
			}
			f.RestoreExc(nil, nil)
		}
		if raised = Assert(f, GetBool(d.Len() == 0).ToObject(), nil); raised != nil {
			return nil, raised
		}
		return result.ToObject(), nil
	})
	cases := []invokeTestCase{
		{args: wrapArgs(newTestDict("foo", 42)), want: newTestDict("foo", 42).ToObject()},
		{args: wrapArgs(newTestDict("foo", 42, 123, "bar")), want: newTestDict("foo", 42, 123, "bar").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestDictNewInit(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(), want: NewDict().ToObject()},
		{args: wrapArgs(newTestDict("foo", 42)), want: newTestDict("foo", 42).ToObject()},
		{args: wrapArgs(), kwargs: wrapKWArgs("foo", 42), want: newTestDict("foo", 42).ToObject()},
		{args: wrapArgs(newTestDict("foo", 42)), kwargs: wrapKWArgs("foo", "bar"), want: newTestDict("foo", "bar").ToObject()},
		{args: wrapArgs(newTestList(newTestTuple("baz", 42))), kwargs: wrapKWArgs("foo", "bar"), want: newTestDict("baz", 42, "foo", "bar").ToObject()},
		{args: wrapArgs(True), wantExc: mustCreateException(TypeErrorType, "'bool' object is not iterable")},
		{args: wrapArgs(NewList(), "foo"), wantExc: mustCreateException(TypeErrorType, "'__init__' requires 1 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(DictType.ToObject(), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestDictNewRaises(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, "'__new__' requires 1 arguments")},
		{args: wrapArgs(123), wantExc: mustCreateException(TypeErrorType, `'__new__' requires a 'type' object but received a "int"`)},
		{args: wrapArgs(NoneType), wantExc: mustCreateException(TypeErrorType, "dict.__new__(NoneType): NoneType is not a subtype of dict")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(DictType, "__new__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestDictSetDefault(t *testing.T) {
	setDefaultMethod := mustNotRaise(GetAttr(NewRootFrame(), DictType.ToObject(), NewStr("setdefault"), nil))
	setDefault := newBuiltinFunction("TestDictSetDefault", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		i, raised := setDefaultMethod.Call(f, args, kwargs)
		if raised != nil {
			return nil, raised
		}
		return NewTuple(i, args[0]).ToObject(), nil
	}).ToObject()
	cases := []invokeTestCase{
		{args: wrapArgs(NewDict(), "foo"), want: newTestTuple(None, newTestDict("foo", None)).ToObject()},
		{args: wrapArgs(NewDict(), "foo", 42), want: newTestTuple(42, newTestDict("foo", 42)).ToObject()},
		{args: wrapArgs(newTestDict("foo", 42), "foo"), want: newTestTuple(42, newTestDict("foo", 42)).ToObject()},
		{args: wrapArgs(newTestDict("foo", 42), "foo", 43), want: newTestTuple(42, newTestDict("foo", 42)).ToObject()},
		{args: wrapArgs(NewDict()), wantExc: mustCreateException(TypeErrorType, "setdefault expected at least 1 arguments, got 0")},
		{args: wrapArgs(NewDict(), "foo", "bar", "baz"), wantExc: mustCreateException(TypeErrorType, "setdefault expected at most 2 arguments, got 3")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(setDefault, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestDictSetItem(t *testing.T) {
	setItem := newBuiltinFunction("TestDictSetItem", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		if raised := checkFunctionArgs(f, "TestDictSetItem", args, DictType, ObjectType, ObjectType); raised != nil {
			return nil, raised
		}
		d := toDictUnsafe(args[0])
		if raised := d.SetItem(f, args[1], args[2]); raised != nil {
			return nil, raised
		}
		return d.ToObject(), nil
	}).ToObject()
	f := NewRootFrame()
	o := newObject(ObjectType)
	deletedItemDict := newStringDict(map[string]*Object{"foo": None})
	if _, raised := deletedItemDict.DelItemString(f, "foo"); raised != nil {
		t.Fatal(raised)
	}
	modifiedDict := newTestDict(0, None)
	modifiedType := newTestClass("Foo", []*Type{IntType}, newStringDict(map[string]*Object{
		"__eq__": newBuiltinFunction("__eq__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			for i := 1000; i < 1100; i++ {
				if raised := modifiedDict.SetItem(f, NewInt(i).ToObject(), None); raised != nil {
					return nil, raised
				}
			}
			return False.ToObject(), nil
		}).ToObject(),
	}))
	cases := []invokeTestCase{
		{args: wrapArgs(NewDict(), "foo", o), want: newStringDict(map[string]*Object{"foo": o}).ToObject()},
		{args: wrapArgs(newStringDict(map[string]*Object{"foo": NewInt(1).ToObject()}), "foo", 2), want: newStringDict(map[string]*Object{"foo": NewInt(2).ToObject()}).ToObject()},
		{args: wrapArgs(newTestDict(2, None, "baz", 3.14), 2, o), want: newTestDict(2, o, "baz", 3.14).ToObject()},
		{args: wrapArgs(deletedItemDict, "foo", o), want: newStringDict(map[string]*Object{"foo": o}).ToObject()},
		{args: wrapArgs(NewDict(), NewList(), None), wantExc: mustCreateException(TypeErrorType, "unhashable type: 'list'")},
		{args: wrapArgs(modifiedDict, newObject(modifiedType), None), wantExc: mustCreateException(RuntimeErrorType, "dictionary changed during write")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(setItem, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestDictSetItemString(t *testing.T) {
	setItemString := newBuiltinFunction("TestDictSetItemString", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		if raised := checkFunctionArgs(f, "TestDictSetItemString", args, DictType, StrType, ObjectType); raised != nil {
			return nil, raised
		}
		d := toDictUnsafe(args[0])
		if raised := d.SetItemString(f, toStrUnsafe(args[1]).Value(), args[2]); raised != nil {
			return nil, raised
		}
		return d.ToObject(), nil
	}).ToObject()
	o := newObject(ObjectType)
	cases := []invokeTestCase{
		{args: wrapArgs(NewDict(), "foo", o), want: newStringDict(map[string]*Object{"foo": o}).ToObject()},
		{args: wrapArgs(newStringDict(map[string]*Object{"foo": NewInt(1).ToObject()}), "foo", 2), want: newStringDict(map[string]*Object{"foo": NewInt(2).ToObject()}).ToObject()},
		{args: wrapArgs(newTestDict(2, None, "baz", 3.14), "baz", o), want: newTestDict(2, None, "baz", o).ToObject()},
		{args: wrapArgs(newTestDict(hashFoo, o, "foo", None), "foo", 3.14), want: newTestDict(hashFoo, o, "foo", 3.14).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(setItemString, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestDictStrRepr(t *testing.T) {
	recursiveDict := NewDict()
	if raised := recursiveDict.SetItemString(NewRootFrame(), "key", recursiveDict.ToObject()); raised != nil {
		t.Fatal(raised)
	}
	cases := []struct {
		o            *Object
		wantPatterns []string
	}{
		{NewDict().ToObject(), []string{"^{}$"}},
		{newStringDict(map[string]*Object{"foo": NewStr("foo value").ToObject()}).ToObject(), []string{`^\{'foo': 'foo value'\}$`}},
		{newStringDict(map[string]*Object{"foo": NewStr("foo value").ToObject(), "bar": NewStr("bar value").ToObject()}).ToObject(), []string{`^{.*, .*}$`, `'foo': 'foo value'`, `'bar': 'bar value'`}},
		{recursiveDict.ToObject(), []string{`^{'key': {\.\.\.}}$`}},
	}
	for _, cas := range cases {
		fun := wrapFuncForTest(func(f *Frame) *BaseException {
			for _, pattern := range cas.wantPatterns {
				re := regexp.MustCompile(pattern)
				s, raised := ToStr(f, cas.o)
				if raised != nil {
					return raised
				}
				if !re.MatchString(s.Value()) {
					t.Errorf("str(%v) = %v, want %q", cas.o, s, re)
				}
				s, raised = Repr(f, cas.o)
				if raised != nil {
					return raised
				}
				if !re.MatchString(s.Value()) {
					t.Errorf("repr(%v) = %v, want %q", cas.o, s, re)
				}
			}
			return nil
		})
		if err := runInvokeTestCase(fun, &invokeTestCase{want: None}); err != "" {
			t.Error(err)
		}
	}
}

func TestDictUpdate(t *testing.T) {
	updateMethod := mustNotRaise(GetAttr(NewRootFrame(), DictType.ToObject(), NewStr("update"), nil))
	update := newBuiltinFunction("TestDictUpdate", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		if raised := checkFunctionVarArgs(f, "TestDictUpdate", args, DictType); raised != nil {
			return nil, raised
		}
		if _, raised := updateMethod.Call(f, args, kwargs); raised != nil {
			return nil, raised
		}
		return args[0], nil
	}).ToObject()
	cases := []invokeTestCase{
		{args: wrapArgs(newTestDict(42, "foo")), want: newTestDict(42, "foo").ToObject()},
		{args: wrapArgs(NewDict(), NewDict()), want: NewDict().ToObject()},
		{args: wrapArgs(NewDict(), newTestDict("foo", 42, "bar", 43)), want: newTestDict("foo", 42, "bar", 43).ToObject()},
		{args: wrapArgs(newTestDict(123, None), newTestDict(124, True)), want: newTestDict(123, None, 124, True).ToObject()},
		{args: wrapArgs(newTestDict("foo", 3.14), newTestDict("foo", "bar")), want: newTestDict("foo", "bar").ToObject()},
		{args: wrapArgs(NewDict(), NewTuple()), want: NewDict().ToObject()},
		{args: wrapArgs(NewDict(), newTestList(newTestTuple("foo", 42), newTestTuple("bar", 43))), want: newTestDict("foo", 42, "bar", 43).ToObject()},
		{args: wrapArgs(newTestDict(123, None), newTestTuple(newTestTuple(124, True))), want: newTestDict(123, None, 124, True).ToObject()},
		{args: wrapArgs(newTestDict("foo", 3.14), newTestList(newTestList("foo", "bar"))), want: newTestDict("foo", "bar").ToObject()},
		{args: wrapArgs(NewDict(), None), wantExc: mustCreateException(TypeErrorType, "'NoneType' object is not iterable")},
		{args: wrapArgs(NewDict(), newTestTuple(newTestList(None, 42, "foo"))), wantExc: mustCreateException(ValueErrorType, "dictionary update sequence element has length 3; 2 is required")},
		{args: wrapArgs(NewDict()), want: NewDict().ToObject()},
		{args: wrapArgs(NewDict()), kwargs: wrapKWArgs("foo", "bar"), want: newTestDict("foo", "bar").ToObject()},
		{args: wrapArgs(newTestDict("foo", 1, "bar", 3.14), newTestDict("foo", 2)), kwargs: wrapKWArgs("foo", 3), want: newTestDict("foo", 3, "bar", 3.14).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(update, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestDictValues(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(NewDict()), want: NewList().ToObject()},
		{args: wrapArgs(newTestDict("foo", 1, "bar", 2)), want: newTestList(1, 2).ToObject()},
		{args: wrapArgs(NewDict(), "bad"), wantExc: mustCreateException(TypeErrorType, "'values' of 'dict' requires 1 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(DictType, "values", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestParallelDictUpdates(t *testing.T) {
	keys := []*Object{
		NewStr("abc").ToObject(),
		NewStr("def").ToObject(),
		NewStr("ghi").ToObject(),
		NewStr("jkl").ToObject(),
		NewStr("mno").ToObject(),
		NewStr("pqr").ToObject(),
		NewStr("stu").ToObject(),
		NewStr("vwx").ToObject(),
		NewStr("yz0").ToObject(),
		NewStr("123").ToObject(),
		NewStr("456").ToObject(),
		NewStr("789").ToObject(),
		NewStr("ABC").ToObject(),
		NewStr("DEF").ToObject(),
		NewStr("GHI").ToObject(),
		NewStr("JKL").ToObject(),
		NewStr("MNO").ToObject(),
		NewStr("PQR").ToObject(),
		NewStr("STU").ToObject(),
		NewStr("VWX").ToObject(),
		NewStr("YZ)").ToObject(),
		NewStr("!@#").ToObject(),
		NewStr("$%^").ToObject(),
		NewStr("&*(").ToObject(),
	}

	var started, finished sync.WaitGroup
	stop := make(chan struct{})
	runner := func(f func(*Frame, *Object, int)) {
		for i := 0; i < 8; i++ {
			started.Add(1)
			finished.Add(1)
			go func() {
				defer finished.Done()
				frame := NewRootFrame()
				i := 0
				for _, k := range keys {
					f(frame, k, i)
					frame.RestoreExc(nil, nil)
					i++
				}
				started.Done()
				for {
					if _, ok := <-stop; !ok {
						break
					}
					for _, k := range keys {
						f(frame, k, i)
						frame.RestoreExc(nil, nil)
						i++
					}
				}
			}()
		}
	}

	d := NewDict().ToObject()
	runner(func(f *Frame, k *Object, _ int) {
		GetItem(f, d, k)
	})

	runner(func(f *Frame, k *Object, i int) {
		mustNotRaise(nil, SetItem(f, d, k, NewInt(i).ToObject()))
	})

	runner(func(f *Frame, k *Object, _ int) {
		DelItem(f, d, k)
	})

	started.Wait()
	time.AfterFunc(time.Second, func() { close(stop) })
	finished.Wait()
}

func newTestDict(elems ...interface{}) *Dict {
	if len(elems)%2 != 0 {
		panic("invalid test dict spec")
	}
	numItems := len(elems) / 2
	d := NewDict()
	f := NewRootFrame()
	for i := 0; i < numItems; i++ {
		k := mustNotRaise(WrapNative(f, reflect.ValueOf(elems[i*2])))
		v := mustNotRaise(WrapNative(f, reflect.ValueOf(elems[i*2+1])))
		d.SetItem(f, k, v)
	}
	return d
}

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
	"testing"
)

func TestSetAdd(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, s *Set, args ...*Object) (*Object, *BaseException) {
		add, raised := GetAttr(f, s.ToObject(), NewStr("add"), nil)
		if raised != nil {
			return nil, raised
		}
		if _, raised := add.Call(f, args, nil); raised != nil {
			return nil, raised
		}
		return s.ToObject(), nil
	})
	cases := []invokeTestCase{
		{args: wrapArgs(NewSet(), "foo"), want: newTestSet("foo").ToObject()},
		{args: wrapArgs(newTestSet(1, 2, 3), 2), want: newTestSet(1, 2, 3).ToObject()},
		{args: wrapArgs(NewSet(), NewList()), wantExc: mustCreateException(TypeErrorType, "unhashable type: 'list'")},
		{args: wrapArgs(NewSet(), "foo", "bar"), wantExc: mustCreateException(TypeErrorType, "'add' of 'set' requires 2 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestSetContains(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(NewSet(), "foo"), want: False.ToObject()},
		{args: wrapArgs(newTestSet(1, 2), 2), want: True.ToObject()},
		{args: wrapArgs(newTestSet(3, "foo"), 42), want: False.ToObject()},
		{args: wrapArgs(NewSet(), NewList()), wantExc: mustCreateException(TypeErrorType, "unhashable type: 'list'")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(SetType, "__contains__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestSetDiscard(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, s *Set, args ...*Object) (*Object, *BaseException) {
		discard, raised := GetAttr(f, s.ToObject(), NewStr("discard"), nil)
		if raised != nil {
			return nil, raised
		}
		if _, raised := discard.Call(f, args, nil); raised != nil {
			return nil, raised
		}
		return s.ToObject(), nil
	})
	cases := []invokeTestCase{
		{args: wrapArgs(newTestSet(1, 2, 3), 2), want: newTestSet(1, 3).ToObject()},
		{args: wrapArgs(newTestSet("foo", 3), "foo"), want: newTestSet(3).ToObject()},
		{args: wrapArgs(NewSet(), NewList()), wantExc: mustCreateException(TypeErrorType, "unhashable type: 'list'")},
		{args: wrapArgs(NewSet(), "foo", "bar"), wantExc: mustCreateException(TypeErrorType, "'discard' of 'set' requires 2 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestSetCompare(t *testing.T) {
	modifiedSet := newTestSet(0)
	modifiedType := newTestClass("Foo", []*Type{IntType}, newStringDict(map[string]*Object{
		"__eq__": newBuiltinFunction("__eq__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			modifiedSet.Add(f, NewStr("baz").ToObject())
			return False.ToObject(), nil
		}).ToObject(),
	}))
	cases := []invokeTestCase{
		{args: wrapArgs(NewSet(), newTestSet("foo")), want: compareAllResultLT},
		{args: wrapArgs(newTestSet(1, 2, 3), newTestSet(3, 2, 1)), want: compareAllResultEq},
		{args: wrapArgs(newTestSet("foo", 3.14), newObject(ObjectType)), want: newTestTuple(false, false, false, true, true, true).ToObject()},
		{args: wrapArgs(123, newTestSet("baz")), want: newTestTuple(true, true, false, true, false, false).ToObject()},
		{args: wrapArgs(mustNotRaise(SetType.Call(newFrame(nil), wrapArgs(newTestRange(100)), nil)), mustNotRaise(SetType.Call(newFrame(nil), wrapArgs(newTestRange(100)), nil))), want: compareAllResultEq},
		{args: wrapArgs(modifiedSet, newTestSet(newObject(modifiedType))), wantExc: mustCreateException(RuntimeErrorType, "set changed during iteration")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(compareAll, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestSetIsSubset(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(NewSet(), newTestSet("foo")), want: True.ToObject()},
		{args: wrapArgs(newTestSet(1, 2), newTestSet(2, 3)), want: False.ToObject()},
		{args: wrapArgs(newTestSet("foo"), newTestTuple("bar")), want: False.ToObject()},
		{args: wrapArgs(mustNotRaise(SetType.Call(newFrame(nil), wrapArgs(newTestRange(42)), nil)), newTestRange(42)), want: True.ToObject()},
		{args: wrapArgs(NewSet(), 123), wantExc: mustCreateException(TypeErrorType, "'int' object is not iterable")},
		{args: wrapArgs(NewSet(), "foo", "bar"), wantExc: mustCreateException(TypeErrorType, "'issubset' of 'set' requires 2 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(SetType, "issubset", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestSetIsSuperset(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(NewSet(), newTestSet("foo")), want: False.ToObject()},
		{args: wrapArgs(newTestSet(1, 2), newTestSet(2, 3)), want: False.ToObject()},
		{args: wrapArgs(newTestSet("foo"), newTestTuple("bar")), want: False.ToObject()},
		{args: wrapArgs(mustNotRaise(SetType.Call(newFrame(nil), wrapArgs(newTestRange(42)), nil)), newTestRange(42)), want: True.ToObject()},
		{args: wrapArgs(NewSet(), 123), wantExc: mustCreateException(TypeErrorType, "'int' object is not iterable")},
		{args: wrapArgs(NewSet(), "foo", "bar"), wantExc: mustCreateException(TypeErrorType, "'issuperset' of 'set' requires 2 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(SetType, "issuperset", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestSetIsTrue(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(NewSet()), want: False.ToObject()},
		{args: wrapArgs(newTestSet("foo", None)), want: True.ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(IsTrue), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestSetIter(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, s *Set) (*Tuple, *BaseException) {
		var result *Tuple
		raised := seqApply(f, s.ToObject(), func(elems []*Object, _ bool) *BaseException {
			result = NewTuple(elems...)
			return nil
		})
		if raised != nil {
			return nil, raised
		}
		return result, nil
	})
	cases := []invokeTestCase{
		{args: wrapArgs(NewSet()), want: NewTuple().ToObject()},
		{args: wrapArgs(newTestSet(1, 2, 3)), want: newTestTuple(1, 2, 3).ToObject()},
		{args: wrapArgs(newTestSet("foo", 3.14)), want: newTestTuple(3.14, "foo").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestSetLen(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(NewSet()), want: NewInt(0).ToObject()},
		{args: wrapArgs(newTestSet(1, 2, 3)), want: NewInt(3).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(SetType, "__len__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestSetNewInit(t *testing.T) {
	cases := []invokeTestCase{
		{want: NewSet().ToObject()},
		{args: wrapArgs(newTestTuple("foo", "bar")), want: newTestSet("foo", "bar").ToObject()},
		{args: wrapArgs("abba"), want: newTestSet("a", "b").ToObject()},
		{args: wrapArgs(3.14), wantExc: mustCreateException(TypeErrorType, "'float' object is not iterable")},
		{args: wrapArgs(NewTuple(), 123), wantExc: mustCreateException(TypeErrorType, "set expected at most 1 arguments, got 2")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(SetType.ToObject(), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestSetRemove(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, s *Set, args ...*Object) (*Object, *BaseException) {
		remove, raised := GetAttr(f, s.ToObject(), NewStr("remove"), nil)
		if raised != nil {
			return nil, raised
		}
		if _, raised := remove.Call(f, args, nil); raised != nil {
			return nil, raised
		}
		return s.ToObject(), nil
	})
	cases := []invokeTestCase{
		{args: wrapArgs(newTestSet(1, 2, 3), 2), want: newTestSet(1, 3).ToObject()},
		{args: wrapArgs(newTestSet("foo", 3), "foo"), want: newTestSet(3).ToObject()},
		{args: wrapArgs(NewSet(), "foo"), wantExc: mustCreateException(KeyErrorType, "foo")},
		{args: wrapArgs(NewSet(), NewList()), wantExc: mustCreateException(TypeErrorType, "unhashable type: 'list'")},
		{args: wrapArgs(NewSet(), "foo", "bar"), wantExc: mustCreateException(TypeErrorType, "'remove' of 'set' requires 2 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestSetStrRepr(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(NewSet()), want: NewStr("set([])").ToObject()},
		{args: wrapArgs(newTestSet("foo")), want: NewStr("set(['foo'])").ToObject()},
		{args: wrapArgs(newTestSet(TupleType, ExceptionType)), want: NewStr("set([<type 'tuple'>, <type 'Exception'>])").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(ToStr), &cas); err != "" {
			t.Error(err)
		}
		if err := runInvokeTestCase(wrapFuncForTest(Repr), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestSetUpdate(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, s *Set, args ...*Object) (*Object, *BaseException) {
		update, raised := GetAttr(f, s.ToObject(), NewStr("update"), nil)
		if raised != nil {
			return nil, raised
		}
		if _, raised := update.Call(f, args, nil); raised != nil {
			return nil, raised
		}
		return s.ToObject(), nil
	})
	cases := []invokeTestCase{
		{args: wrapArgs(NewSet(), "foo"), want: newTestSet("f", "o").ToObject()},
		{args: wrapArgs(NewSet(), newTestDict(1, "1", 2, "2")), want: newTestSet(1, 2).ToObject()},
		{args: wrapArgs(NewSet(), newTestTuple("foo", "bar", "bar")), want: newTestSet("foo", "bar").ToObject()},
		{args: wrapArgs(NewSet(), newTestTuple(NewDict())), wantExc: mustCreateException(TypeErrorType, "unhashable type: 'dict'")},
		{args: wrapArgs(NewSet(), 123), wantExc: mustCreateException(TypeErrorType, "'int' object is not iterable")},
		{args: wrapArgs(NewSet(), "foo", "bar"), wantExc: mustCreateException(TypeErrorType, "'update' of 'set' requires 2 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func newTestSet(elems ...interface{}) *Set {
	f := newFrame(nil)
	wrappedElems, raised := seqWrapEach(f, elems...)
	if raised != nil {
		panic(raised)
	}
	s := NewSet()
	for _, elem := range wrappedElems {
		if _, raised := s.Add(f, elem); raised != nil {
			panic(raised)
		}
	}
	return s
}

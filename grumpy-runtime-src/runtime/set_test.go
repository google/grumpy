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
	f := NewRootFrame()
	for _, typ := range []*Type{SetType, FrozenSetType} {
		cases := []invokeTestCase{
			{args: wrapArgs(mustNotRaise(typ.Call(f, nil, nil)), "foo"), want: False.ToObject()},
			{args: wrapArgs(mustNotRaise(typ.Call(f, wrapArgs(newTestTuple(1, 2)), nil)), 2), want: True.ToObject()},
			{args: wrapArgs(mustNotRaise(typ.Call(f, wrapArgs(newTestTuple(3, "foo")), nil)), 42), want: False.ToObject()},
			{args: wrapArgs(mustNotRaise(typ.Call(f, nil, nil)), NewList()), wantExc: mustCreateException(TypeErrorType, "unhashable type: 'list'")},
		}
		for _, cas := range cases {
			if err := runInvokeMethodTestCase(typ, "__contains__", &cas); err != "" {
				t.Error(err)
			}
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
		{args: wrapArgs(mustNotRaise(SetType.Call(NewRootFrame(), wrapArgs(newTestRange(100)), nil)), mustNotRaise(SetType.Call(NewRootFrame(), wrapArgs(newTestRange(100)), nil))), want: compareAllResultEq},
		{args: wrapArgs(modifiedSet, newTestSet(newObject(modifiedType))), wantExc: mustCreateException(RuntimeErrorType, "set changed during iteration")},
		{args: wrapArgs(newTestFrozenSet(), newTestFrozenSet("foo")), want: compareAllResultLT},
		{args: wrapArgs(newTestFrozenSet(1, 2, 3), newTestFrozenSet(3, 2, 1)), want: compareAllResultEq},
		{args: wrapArgs(newTestFrozenSet("foo", 3.14), newObject(ObjectType)), want: newTestTuple(true, true, false, true, false, false).ToObject()},
		{args: wrapArgs(123, newTestFrozenSet("baz")), want: newTestTuple(false, false, false, true, true, true).ToObject()},
		{args: wrapArgs(mustNotRaise(FrozenSetType.Call(NewRootFrame(), wrapArgs(newTestRange(100)), nil)), mustNotRaise(FrozenSetType.Call(NewRootFrame(), wrapArgs(newTestRange(100)), nil))), want: compareAllResultEq},
		{args: wrapArgs(newTestFrozenSet(), NewSet()), want: newTestTuple(false, true, true, false, true, false).ToObject()},
		{args: wrapArgs(newTestSet("foo", "bar"), newTestFrozenSet("foo", "bar")), want: newTestTuple(false, true, true, false, true, false).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(compareAll, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestSetIsSubset(t *testing.T) {
	f := NewRootFrame()
	for _, typ := range []*Type{SetType, FrozenSetType} {
		cases := []invokeTestCase{
			{args: wrapArgs(mustNotRaise(typ.Call(f, nil, nil)), newTestSet("foo")), want: True.ToObject()},
			{args: wrapArgs(mustNotRaise(typ.Call(f, nil, nil)), newTestFrozenSet("foo")), want: True.ToObject()},
			{args: wrapArgs(mustNotRaise(typ.Call(f, wrapArgs(newTestTuple(1, 2)), nil)), newTestSet(2, 3)), want: False.ToObject()},
			{args: wrapArgs(mustNotRaise(typ.Call(f, wrapArgs(newTestTuple(1, 2)), nil)), newTestFrozenSet(2, 3)), want: False.ToObject()},
			{args: wrapArgs(mustNotRaise(typ.Call(f, wrapArgs(newTestTuple("foo")), nil)), newTestTuple("bar")), want: False.ToObject()},
			{args: wrapArgs(mustNotRaise(typ.Call(f, wrapArgs(newTestRange(42)), nil)), newTestRange(42)), want: True.ToObject()},
			{args: wrapArgs(mustNotRaise(typ.Call(f, nil, nil)), 123), wantExc: mustCreateException(TypeErrorType, "'int' object is not iterable")},
			{args: wrapArgs(mustNotRaise(typ.Call(f, nil, nil)), "foo", "bar"), wantExc: mustCreateException(TypeErrorType, fmt.Sprintf("'issubset' of '%s' requires 2 arguments", typ.Name()))},
		}
		for _, cas := range cases {
			if err := runInvokeMethodTestCase(typ, "issubset", &cas); err != "" {
				t.Error(err)
			}
		}
	}
}

func TestSetIsSuperset(t *testing.T) {
	f := NewRootFrame()
	for _, typ := range []*Type{SetType, FrozenSetType} {
		cases := []invokeTestCase{
			{args: wrapArgs(mustNotRaise(typ.Call(f, nil, nil)), newTestSet("foo")), want: False.ToObject()},
			{args: wrapArgs(mustNotRaise(typ.Call(f, wrapArgs(newTestTuple(1, 2)), nil)), newTestSet(2, 3)), want: False.ToObject()},
			{args: wrapArgs(mustNotRaise(typ.Call(f, wrapArgs(newTestTuple("foo")), nil)), newTestTuple("bar")), want: False.ToObject()},
			{args: wrapArgs(mustNotRaise(typ.Call(f, wrapArgs(newTestRange(42)), nil)), newTestRange(42)), want: True.ToObject()},
			{args: wrapArgs(mustNotRaise(typ.Call(f, nil, nil)), 123), wantExc: mustCreateException(TypeErrorType, "'int' object is not iterable")},
			{args: wrapArgs(mustNotRaise(typ.Call(f, nil, nil)), "foo", "bar"), wantExc: mustCreateException(TypeErrorType, fmt.Sprintf("'issuperset' of '%s' requires 2 arguments", typ.Name()))},
		}
		for _, cas := range cases {
			if err := runInvokeMethodTestCase(typ, "issuperset", &cas); err != "" {
				t.Error(err)
			}
		}
	}
}

func TestSetIsTrue(t *testing.T) {
	f := NewRootFrame()
	for _, typ := range []*Type{SetType, FrozenSetType} {
		cases := []invokeTestCase{
			{args: wrapArgs(mustNotRaise(typ.Call(f, nil, nil))), want: False.ToObject()},
			{args: wrapArgs(mustNotRaise(typ.Call(f, wrapArgs(newTestTuple("foo", None)), nil))), want: True.ToObject()},
		}
		for _, cas := range cases {
			if err := runInvokeTestCase(wrapFuncForTest(IsTrue), &cas); err != "" {
				t.Error(err)
			}
		}
	}
}

func TestSetIter(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, s *Object) (*Tuple, *BaseException) {
		var result *Tuple
		raised := seqApply(f, s, func(elems []*Object, _ bool) *BaseException {
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
		{args: wrapArgs(newTestSet("foo", 3.14)), want: newTestTuple("foo", 3.14).ToObject()},
		{args: wrapArgs(newTestFrozenSet()), want: NewTuple().ToObject()},
		{args: wrapArgs(newTestFrozenSet(1, 2, 3)), want: newTestTuple(1, 2, 3).ToObject()},
		{args: wrapArgs(newTestFrozenSet("foo", 3.14)), want: newTestTuple("foo", 3.14).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestSetLen(t *testing.T) {
	f := NewRootFrame()
	for _, typ := range []*Type{SetType, FrozenSetType} {
		cases := []invokeTestCase{
			{args: wrapArgs(mustNotRaise(typ.Call(f, nil, nil))), want: NewInt(0).ToObject()},
			{args: wrapArgs(mustNotRaise(typ.Call(f, wrapArgs(newTestTuple(1, 2, 3)), nil))), want: NewInt(3).ToObject()},
		}
		for _, cas := range cases {
			if err := runInvokeMethodTestCase(typ, "__len__", &cas); err != "" {
				t.Error(err)
			}
		}
	}
}

func TestSetNewInit(t *testing.T) {
	f := NewRootFrame()
	for _, typ := range []*Type{SetType, FrozenSetType} {
		cases := []invokeTestCase{
			{want: NewSet().ToObject()},
			{args: wrapArgs(newTestTuple("foo", "bar")), want: mustNotRaise(typ.Call(f, wrapArgs(newTestTuple("foo", "bar")), nil))},
			{args: wrapArgs("abba"), want: mustNotRaise(typ.Call(f, wrapArgs(newTestTuple("a", "b")), nil))},
			{args: wrapArgs(3.14), wantExc: mustCreateException(TypeErrorType, "'float' object is not iterable")},
			{args: wrapArgs(NewTuple(), 1, 2, 3), wantExc: mustCreateException(TypeErrorType, fmt.Sprintf("%s expected at most 1 arguments, got 4", typ.Name()))},
		}
		for _, cas := range cases {
			if err := runInvokeTestCase(typ.ToObject(), &cas); err != "" {
				t.Error(err)
			}
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
		{args: wrapArgs(newTestFrozenSet()), want: NewStr("frozenset([])").ToObject()},
		{args: wrapArgs(newTestFrozenSet("foo")), want: NewStr("frozenset(['foo'])").ToObject()},
		{args: wrapArgs(newTestFrozenSet(TupleType, ExceptionType)), want: NewStr("frozenset([<type 'tuple'>, <type 'Exception'>])").ToObject()},
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
	f := NewRootFrame()
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

func newTestFrozenSet(elems ...interface{}) *FrozenSet {
	f := NewRootFrame()
	wrappedElems, raised := seqWrapEach(f, elems...)
	if raised != nil {
		panic(raised)
	}
	d := NewDict()
	for _, elem := range wrappedElems {
		if raised := d.SetItem(f, elem, None); raised != nil {
			panic(raised)
		}
	}
	return &FrozenSet{Object{typ: FrozenSetType}, d}
}

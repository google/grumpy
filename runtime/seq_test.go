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

func TestSeqApply(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, seq *Object) (*Object, *BaseException) {
		var got *Object
		raised := seqApply(f, seq, func(elems []*Object, borrowed bool) *BaseException {
			got = newTestTuple(NewTuple(elems...), GetBool(borrowed)).ToObject()
			return nil
		})
		if raised != nil {
			return nil, raised
		}
		return got, nil
	})
	cases := []invokeTestCase{
		{args: wrapArgs(NewTuple()), want: newTestTuple(NewTuple(), true).ToObject()},
		{args: wrapArgs(newTestList("foo", "bar")), want: newTestTuple(newTestTuple("foo", "bar"), true).ToObject()},
		{args: wrapArgs(newTestDict("foo", None)), want: newTestTuple(newTestTuple("foo"), false).ToObject()},
		{args: wrapArgs(42), wantExc: mustCreateException(TypeErrorType, "'int' object is not iterable")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestSeqCount(t *testing.T) {
	badEqType := newTestClass("Eq", []*Type{IntType}, newStringDict(map[string]*Object{
		"__eq__": newBuiltinFunction("__eq__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return nil, f.RaiseType(TypeErrorType, "uh oh")
		}).ToObject(),
	}))
	badNonZeroType := newTestClass("BadNonZeroType", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__nonzero__": newBuiltinFunction("__nonzero__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return nil, f.RaiseType(TypeErrorType, "uh oh")
		}).ToObject(),
	}))
	worseEqType := newTestClass("WorseEqCmp", []*Type{IntType}, newStringDict(map[string]*Object{
		"__eq__": newBuiltinFunction("__eq__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return newObject(badNonZeroType), nil
		}).ToObject(),
	}))
	cases := []invokeTestCase{
		{args: wrapArgs(newTestList(), NewInt(1)), want: NewInt(0).ToObject()},
		{args: wrapArgs(newTestList(1, 2, 3, 4, 5, 6), NewInt(7)), want: NewInt(0).ToObject()},
		{args: wrapArgs(newTestList(1, 2, 3, 2, 2, 4), NewInt(2)), want: NewInt(3).ToObject()},
		{args: wrapArgs(newTestList(1, None, None, 3), None), want: NewInt(2).ToObject()},
		{args: wrapArgs(newTestList("a", "b", "c", "d", "e"), NewStr("c")), want: NewInt(1).ToObject()},
		{args: wrapArgs(newTestList(newObject(badEqType)), newObject(badEqType)), wantExc: mustCreateException(TypeErrorType, "uh oh")},
		{args: wrapArgs(newTestList(newObject(worseEqType)), newObject(worseEqType)), wantExc: mustCreateException(TypeErrorType, "uh oh")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(seqCount), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestSeqForEach(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, seq *Object) (*Object, *BaseException) {
		elems := []*Object{}
		raised := seqForEach(f, seq, func(elem *Object) *BaseException {
			elems = append(elems, elem)
			return nil
		})
		if raised != nil {
			return nil, raised
		}
		return NewTuple(elems...).ToObject(), nil
	})
	cases := []invokeTestCase{
		{args: wrapArgs(NewList()), want: NewTuple().ToObject()},
		{args: wrapArgs(newTestDict("foo", 1, "bar", 2)), want: newTestTuple("foo", "bar").ToObject()},
		{args: wrapArgs(123), wantExc: mustCreateException(TypeErrorType, "'int' object is not iterable")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestSeqIterator(t *testing.T) {
	fun := newBuiltinFunction("TestSeqIterator", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		return TupleType.Call(f, args, nil)
	}).ToObject()
	exhaustedIter := newSeqIterator(NewStr("foo").ToObject())
	TupleType.Call(NewRootFrame(), []*Object{exhaustedIter}, nil)
	cases := []invokeTestCase{
		{args: wrapArgs(newSeqIterator(NewStr("bar").ToObject())), want: newTestTuple("b", "a", "r").ToObject()},
		{args: wrapArgs(newSeqIterator(newTestTuple(123, 456).ToObject())), want: newTestTuple(123, 456).ToObject()},
		{args: wrapArgs(exhaustedIter), want: NewTuple().ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestSeqNew(t *testing.T) {
	fun := newBuiltinFunction("TestSeqNew", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		elems, raised := seqNew(f, args)
		if raised != nil {
			return nil, raised
		}
		return NewTuple(elems...).ToObject(), nil
	}).ToObject()
	cases := []invokeTestCase{
		{want: NewTuple().ToObject()},
		{args: wrapArgs(newTestTuple("foo", "bar")), want: newTestTuple("foo", "bar").ToObject()},
		{args: wrapArgs(newTestDict("foo", 1)), want: newTestTuple("foo").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

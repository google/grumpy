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

func TestEnumerate(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, args ...*Object) (*Object, *BaseException) {
		e, raised := enumerateType.Call(f, args, nil)
		if raised != nil {
			return nil, raised
		}
		return ListType.Call(f, Args{e}, nil)
	})
	cases := []invokeTestCase{
		{args: wrapArgs(NewTuple()), want: NewList().ToObject()},
		{args: wrapArgs(newTestList("foo", "bar")), want: newTestList(newTestTuple(0, "foo"), newTestTuple(1, "bar")).ToObject()},
		{args: wrapArgs(newTestTuple("foo", "bar"), 1), want: newTestList(newTestTuple(1, "bar")).ToObject()},
		{args: wrapArgs(newTestList("foo", "bar"), 128), want: NewList().ToObject()},
		{args: wrapArgs(newTestTuple(42), -300), want: newTestList(newTestTuple(0, 42)).ToObject()},
		{args: wrapArgs(NewTuple(), 3.14), wantExc: mustCreateException(TypeErrorType, "float object cannot be interpreted as an index")},
		{args: wrapArgs(123), wantExc: mustCreateException(TypeErrorType, "'int' object is not iterable")},
		{args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, "'__new__' requires 2 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestRangeIteratorIter(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame) *BaseException {
		xrange, raised := xrangeType.Call(f, wrapArgs(5), nil)
		if raised != nil {
			return raised
		}
		iter, raised := Iter(f, xrange)
		if raised != nil {
			return raised
		}
		if !iter.isInstance(rangeIteratorType) {
			t.Errorf("iter(xrange(5)) = %v, want rangeiterator", iter)
		}
		got, raised := Iter(f, iter)
		if raised != nil {
			return raised
		}
		if got != iter {
			t.Errorf("iter(%[1]v) = %[2]v, want %[1]v", iter, got)
		}
		return nil
	})
	if err := runInvokeTestCase(fun, &invokeTestCase{want: None}); err != "" {
		t.Error(err)
	}
}

func TestXRangeGetItem(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(newTestXRange(10), 3), want: NewInt(3).ToObject()},
		{args: wrapArgs(newTestXRange(10, 12), 1), want: NewInt(11).ToObject()},
		{args: wrapArgs(newTestXRange(5, -2, -3), 2), want: NewInt(-1).ToObject()},
		{args: wrapArgs(newTestXRange(3), 100), wantExc: mustCreateException(IndexErrorType, "index out of range")},
		{args: wrapArgs(newTestXRange(5), newTestSlice(1, 3)), wantExc: mustCreateException(TypeErrorType, "sequence index must be integer, not 'slice'")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(xrangeType, "__getitem__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestXRangeLen(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(newTestXRange(10)), want: NewInt(10).ToObject()},
		{args: wrapArgs(newTestXRange(10, 12)), want: NewInt(2).ToObject()},
		{args: wrapArgs(newTestXRange(5, 16, 5)), want: NewInt(3).ToObject()},
		{args: wrapArgs(newTestXRange(5, -2, -3)), want: NewInt(3).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(xrangeType, "__len__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestXRangeNew(t *testing.T) {
	fun := newBuiltinFunction("TestXRangeNew", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		xrange, raised := xrangeType.Call(f, args, nil)
		if raised != nil {
			return nil, raised
		}
		return ListType.Call(f, []*Object{xrange}, nil)
	}).ToObject()
	cases := []invokeTestCase{
		{args: wrapArgs(5), want: newTestList(0, 1, 2, 3, 4).ToObject()},
		{args: wrapArgs(-3, 1), want: newTestList(-3, -2, -1, 0).ToObject()},
		{args: wrapArgs(10, 0), want: NewList().ToObject()},
		{args: wrapArgs(4, 7, 3), want: newTestList(4).ToObject()},
		{args: wrapArgs(4, 8, 3), want: newTestList(4, 7).ToObject()},
		{args: wrapArgs(-12, -21, -5), want: newTestList(-12, -17).ToObject()},
		{args: wrapArgs(-12, -22, -5), want: newTestList(-12, -17).ToObject()},
		{args: wrapArgs(-12, -23, -5), want: newTestList(-12, -17, -22).ToObject()},
		{args: wrapArgs(4, -4), want: NewList().ToObject()},
		{args: wrapArgs(-26, MinInt), want: NewList().ToObject()},
		{args: wrapArgs(1, 2, 0), wantExc: mustCreateException(ValueErrorType, "xrange() arg 3 must not be zero")},
		{args: wrapArgs(0, MinInt, -1), wantExc: mustCreateException(OverflowErrorType, "result too large")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestXRangeRepr(t *testing.T) {
	fun := newBuiltinFunction("TestXRangeRepr", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		xrange, raised := xrangeType.Call(f, args, nil)
		if raised != nil {
			return nil, raised
		}
		s, raised := Repr(f, xrange)
		if raised != nil {
			return nil, raised
		}
		return s.ToObject(), nil
	}).ToObject()
	cases := []invokeTestCase{
		{args: wrapArgs(42), want: NewStr("xrange(42)").ToObject()},
		{args: wrapArgs(42, 48), want: NewStr("xrange(42, 48)").ToObject()},
		{args: wrapArgs(42, 10), want: NewStr("xrange(42, 42)").ToObject()},
		{args: wrapArgs(-10, 10), want: NewStr("xrange(-10, 10)").ToObject()},
		{args: wrapArgs(-10, 10, 10), want: NewStr("xrange(-10, 10, 10)").ToObject()},
		{args: wrapArgs(-10, 10, 3), want: NewStr("xrange(-10, 11, 3)").ToObject()},
		{args: wrapArgs(4, 7, 3), want: NewStr("xrange(4, 7, 3)").ToObject()},
		{args: wrapArgs(4, 8, 3), want: NewStr("xrange(4, 10, 3)").ToObject()},
		{args: wrapArgs(-10, 10, -3), want: NewStr("xrange(-10, -10, -3)").ToObject()},
		{args: wrapArgs(3, 3, -5), want: NewStr("xrange(3, 3, -5)").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func newTestXRange(args ...interface{}) *Object {
	return mustNotRaise(xrangeType.Call(NewRootFrame(), wrapArgs(args...), nil))
}

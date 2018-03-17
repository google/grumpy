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

func TestSliceCalcSize(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, s *Slice, numElems int) (*Object, *BaseException) {
		start, stop, step, sliceLen, raised := s.calcSlice(f, numElems)
		if raised != nil {
			return nil, raised
		}
		return newTestTuple(start, stop, step, sliceLen).ToObject(), nil
	})
	cases := []invokeTestCase{
		{args: wrapArgs(newTestSlice(4), 6), want: newTestTuple(0, 4, 1, 4).ToObject()},
		{args: wrapArgs(newTestSlice(-8), 3), want: newTestTuple(0, 0, 1, 0).ToObject()},
		{args: wrapArgs(newTestSlice(0, 10), 3), want: newTestTuple(0, 3, 1, 3).ToObject()},
		{args: wrapArgs(newTestSlice(1, 2, newObject(ObjectType)), 0), wantExc: mustCreateException(TypeErrorType, errBadSliceIndex)},
		{args: wrapArgs(newTestSlice(newObject(ObjectType)), 10), wantExc: mustCreateException(TypeErrorType, errBadSliceIndex)},
		{args: wrapArgs(newTestSlice(newObject(ObjectType), 4), 10), wantExc: mustCreateException(TypeErrorType, errBadSliceIndex)},
		{args: wrapArgs(newTestSlice(1.0, 4), 10), wantExc: mustCreateException(TypeErrorType, errBadSliceIndex)},
		{args: wrapArgs(newTestSlice(1, 2, 0), 3), wantExc: mustCreateException(ValueErrorType, "slice step cannot be zero")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestSliceCompare(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(newTestSlice(None), newTestSlice(None)), want: compareAllResultEq},
		{args: wrapArgs(newTestSlice(2), newTestSlice(1)), want: compareAllResultGT},
		{args: wrapArgs(newTestSlice(1, 2), newTestSlice(1, 3)), want: compareAllResultLT},
		{args: wrapArgs(newTestSlice(1, 2, 3), newTestSlice(1, 2)), want: compareAllResultGT},
		{args: wrapArgs(None, newTestSlice(1, 2)), want: compareAllResultLT},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(compareAll, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestSliceNew(t *testing.T) {
	cases := []invokeTestCase{
		{args: nil, wantExc: mustCreateException(TypeErrorType, "'__new__' of 'object' requires 3 arguments")},
		{args: wrapArgs(10), want: (&Slice{Object{typ: SliceType}, nil, NewInt(10).ToObject(), nil}).ToObject()},
		{args: wrapArgs(1.2, "foo"), want: (&Slice{Object{typ: SliceType}, NewFloat(1.2).ToObject(), NewStr("foo").ToObject(), nil}).ToObject()},
		{args: wrapArgs(None, None, true), want: (&Slice{Object{typ: SliceType}, None, None, True.ToObject()}).ToObject()},
		{args: wrapArgs(1, 2, 3, 4), wantExc: mustCreateException(TypeErrorType, "'__new__' of 'object' requires 3 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(SliceType.ToObject(), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestSliceStrRepr(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(newTestSlice(3.14)), want: NewStr("slice(None, 3.14, None)").ToObject()},
		{args: wrapArgs(newTestSlice("foo", "bar")), want: NewStr("slice('foo', 'bar', None)").ToObject()},
		{args: wrapArgs(newTestSlice(1, 2, 3)), want: NewStr("slice(1, 2, 3)").ToObject()},
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

func newTestSlice(args ...interface{}) *Object {
	return mustNotRaise(SliceType.Call(NewRootFrame(), wrapArgs(args...), nil))
}

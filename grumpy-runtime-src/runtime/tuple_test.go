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
	"testing"
)

func TestNewTuple(t *testing.T) {
	cases := [][]*Object{
		nil,
		{newObject(ObjectType)},
		{newObject(ObjectType), newObject(ObjectType)},
	}
	for _, args := range cases {
		tuple := NewTuple(args...)
		if !reflect.DeepEqual(tuple.elems, args) {
			t.Errorf("NewTuple(%v) = %v, want %v", args, tuple.elems, args)
		}
	}
}

func TestTupleBinaryOps(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, fn binaryOpFunc, v, w *Object) (*Object, *BaseException) {
		return fn(f, v, w)
	})
	cases := []invokeTestCase{
		{args: wrapArgs(Add, newTestTuple(3), newTestTuple("foo")), want: newTestTuple(3, "foo").ToObject()},
		{args: wrapArgs(Add, NewTuple(None), NewTuple()), want: NewTuple(None).ToObject()},
		{args: wrapArgs(Add, NewTuple(), newObject(ObjectType)), wantExc: mustCreateException(TypeErrorType, "unsupported operand type(s) for +: 'tuple' and 'object'")},
		{args: wrapArgs(Add, None, NewTuple()), wantExc: mustCreateException(TypeErrorType, "unsupported operand type(s) for +: 'NoneType' and 'tuple'")},
		{args: wrapArgs(Mul, NewTuple(), 10), want: NewTuple().ToObject()},
		{args: wrapArgs(Mul, newTestTuple("baz"), -2), want: NewTuple().ToObject()},
		{args: wrapArgs(Mul, newTestTuple(None, None), 0), want: NewTuple().ToObject()},
		{args: wrapArgs(Mul, newTestTuple(1, "bar"), 2), want: newTestTuple(1, "bar", 1, "bar").ToObject()},
		{args: wrapArgs(Mul, 1, newTestTuple(1, "bar")), want: newTestTuple(1, "bar").ToObject()},
		{args: wrapArgs(Mul, newObject(ObjectType), newTestTuple(newObject(ObjectType))), wantExc: mustCreateException(TypeErrorType, "unsupported operand type(s) for *: 'object' and 'tuple'")},
		{args: wrapArgs(Mul, NewTuple(newObject(ObjectType)), NewTuple()), wantExc: mustCreateException(TypeErrorType, "unsupported operand type(s) for *: 'tuple' and 'tuple'")},
		{args: wrapArgs(Mul, NewTuple(None, None), MaxInt), wantExc: mustCreateException(OverflowErrorType, "result too large")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestTupleCompare(t *testing.T) {
	o := newObject(ObjectType)
	cases := []invokeTestCase{
		{args: wrapArgs(NewTuple(), NewTuple()), want: compareAllResultEq},
		{args: wrapArgs(newTestTuple("foo", o), newTestTuple("foo", o)), want: compareAllResultEq},
		{args: wrapArgs(newTestTuple(4), newTestTuple(3, 0)), want: compareAllResultGT},
		{args: wrapArgs(newTestTuple(4), newTestTuple(4, 3, 0)), want: compareAllResultLT},
		{args: wrapArgs(NewTuple(o), NewTuple()), want: compareAllResultGT},
		{args: wrapArgs(NewTuple(o), newTestTuple("foo")), want: compareAllResultLT},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(compareAll, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestTupleCompareNotImplemented(t *testing.T) {
	cas := invokeTestCase{args: wrapArgs(NewTuple(), 3), want: NotImplemented}
	if err := runInvokeMethodTestCase(TupleType, "__eq__", &cas); err != "" {
		t.Error(err)
	}
}

func TestTupleContains(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(newTestTuple("foo", 42, "bar"), 1), want: False.ToObject()},
		{args: wrapArgs(newTestTuple("foo", 42, "bar"), "foo"), want: True.ToObject()},
		{args: wrapArgs(newTestTuple("foo", 42, "bar"), 42), want: True.ToObject()},
		{args: wrapArgs(newTestTuple("foo", 42, "bar"), "bar"), want: True.ToObject()},
		{args: wrapArgs(NewTuple(), newTestSlice(50, 100)), want: False.ToObject()},
		{args: wrapArgs(newTestTuple(1, 2, 3, 4, 5), newTestSlice(1, None, 2)), want: False.ToObject()},
		{args: wrapArgs(NewTuple(), 1), want: False.ToObject()},
		{args: wrapArgs(newTestTuple(32), -100), want: False.ToObject()},
		{args: wrapArgs(newTestTuple(1, 2, 3), newTestSlice(1, None, 0)), want: False.ToObject()},
		{args: wrapArgs(newTestTuple(true), None), want: False.ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(TupleType, "__contains__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestTupleCount(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(NewTuple(), NewInt(1)), want: NewInt(0).ToObject()},
		{args: wrapArgs(NewTuple(None, None, None), None), want: NewInt(3).ToObject()},
		{args: wrapArgs(NewTuple()), wantExc: mustCreateException(TypeErrorType, "'count' of 'tuple' requires 2 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(TupleType, "count", &cas); err != "" {
			t.Error(err)
		}
	}
}

func BenchmarkTupleContains(b *testing.B) {
	b.Run("false-3", func(b *testing.B) {
		t := newTestTuple("foo", 42, "bar").ToObject()
		a := wrapArgs(1)[0]
		f := NewRootFrame()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Contains(f, t, a)
		}
	})

	b.Run("false-10", func(b *testing.B) {
		t := newTestTuple("foo", 42, "bar", "foo", 42, "bar", "foo", 42, "bar", "baz").ToObject()
		a := wrapArgs(1)[0]
		f := NewRootFrame()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Contains(f, t, a)
		}
	})

	b.Run("true-3.1", func(b *testing.B) {
		t := newTestTuple("foo", 42, "bar").ToObject()
		a := wrapArgs("foo")[0]
		f := NewRootFrame()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Contains(f, t, a)
		}
	})

	b.Run("true-3.3", func(b *testing.B) {
		t := newTestTuple("foo", 42, "bar").ToObject()
		a := wrapArgs("bar")[0]
		f := NewRootFrame()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Contains(f, t, a)
		}
	})

	b.Run("true-10.10", func(b *testing.B) {
		t := newTestTuple("foo", 42, "bar", "foo", 42, "bar", "foo", 42, "bar", "baz").ToObject()
		a := wrapArgs("baz")[0]
		f := NewRootFrame()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Contains(f, t, a)
		}
	})
}

func TestTupleGetItem(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(newTestTuple("foo", 42, "bar"), 1), want: NewInt(42).ToObject()},
		{args: wrapArgs(newTestTuple("foo", 42, "bar"), -3), want: NewStr("foo").ToObject()},
		{args: wrapArgs(NewTuple(), newTestSlice(50, 100)), want: NewTuple().ToObject()},
		{args: wrapArgs(newTestTuple(1, 2, 3, 4, 5), newTestSlice(1, None, 2)), want: newTestTuple(2, 4).ToObject()},
		{args: wrapArgs(NewTuple(), 1), wantExc: mustCreateException(IndexErrorType, "index out of range")},
		{args: wrapArgs(newTestTuple(32), -100), wantExc: mustCreateException(IndexErrorType, "index out of range")},
		{args: wrapArgs(newTestTuple(1, 2, 3), newTestSlice(1, None, 0)), wantExc: mustCreateException(ValueErrorType, "slice step cannot be zero")},
		{args: wrapArgs(newTestTuple(true), None), wantExc: mustCreateException(TypeErrorType, "sequence indices must be integers, not NoneType")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(TupleType, "__getitem__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestTupleLen(t *testing.T) {
	tuple := newTestTuple("foo", 42, "bar")
	if got := tuple.Len(); got != 3 {
		t.Errorf("%v.Len() = %v, want 3", tuple, got)
	}
}

func TestTupleNew(t *testing.T) {
	cases := []invokeTestCase{
		{want: NewTuple().ToObject()},
		{args: wrapArgs(newTestTuple(1, 2, 3)), want: newTestTuple(1, 2, 3).ToObject()},
		{args: wrapArgs(newTestDict(1, "foo", "bar", None)), want: newTestTuple(1, "bar").ToObject()},
		{args: wrapArgs(42), wantExc: mustCreateException(TypeErrorType, "'int' object is not iterable")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(TupleType.ToObject(), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestTupleStrRepr(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, o *Object) (*Tuple, *BaseException) {
		str, raised := ToStr(f, o)
		if raised != nil {
			return nil, raised
		}
		repr, raised := Repr(f, o)
		if raised != nil {
			return nil, raised
		}
		return newTestTuple(str, repr), nil
	})
	cases := []invokeTestCase{
		{args: wrapArgs(NewTuple()), want: newTestTuple("()", "()").ToObject()},
		{args: wrapArgs(newTestTuple("foo")), want: newTestTuple("('foo',)", "('foo',)").ToObject()},
		{args: wrapArgs(newTestTuple(TupleType, ExceptionType)), want: newTestTuple("(<type 'tuple'>, <type 'Exception'>)", "(<type 'tuple'>, <type 'Exception'>)").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestTupleIter(t *testing.T) {
	o := newObject(ObjectType)
	cases := []invokeTestCase{
		{args: wrapArgs(NewTuple()), want: NewList().ToObject()},
		{args: wrapArgs(newTestTuple(1, o, "foo")), want: newTestList(1, o, "foo").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(ListType.ToObject(), &cas); err != "" {
			t.Error(err)
		}
	}
}

func newTestTuple(elems ...interface{}) *Tuple {
	return NewTuple(wrapArgs(elems...)...)
}

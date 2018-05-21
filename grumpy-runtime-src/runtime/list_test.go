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
	"math/big"
	"reflect"
	"testing"
)

func TestNewList(t *testing.T) {
	cases := [][]*Object{
		nil,
		[]*Object{newObject(ObjectType)},
		[]*Object{newObject(ObjectType), newObject(ObjectType)},
	}
	for _, args := range cases {
		l := NewList(args...)
		if !reflect.DeepEqual(l.elems, args) {
			t.Errorf("NewList(%v) = %v, want %v", args, l.elems, args)
		}
	}
}

func TestListBinaryOps(t *testing.T) {
	cases := []struct {
		fun     func(f *Frame, v, w *Object) (*Object, *BaseException)
		v, w    *Object
		want    *Object
		wantExc *BaseException
	}{
		{Add, newTestList(3).ToObject(), newTestList("foo").ToObject(), newTestList(3, "foo").ToObject(), nil},
		{Add, NewList(None).ToObject(), NewList().ToObject(), NewList(None).ToObject(), nil},
		{Add, NewList().ToObject(), newObject(ObjectType), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for +: 'list' and 'object'")},
		{Add, None, NewList().ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for +: 'NoneType' and 'list'")},
		{Mul, NewList().ToObject(), NewInt(10).ToObject(), NewList().ToObject(), nil},
		{Mul, newTestList("baz").ToObject(), NewInt(-2).ToObject(), NewList().ToObject(), nil},
		{Mul, NewList(None, None).ToObject(), NewInt(0).ToObject(), NewList().ToObject(), nil},
		{Mul, newTestList(1, "bar").ToObject(), NewInt(2).ToObject(), newTestList(1, "bar", 1, "bar").ToObject(), nil},
		{Mul, NewInt(1).ToObject(), newTestList(1, "bar").ToObject(), newTestList(1, "bar").ToObject(), nil},
		{Mul, newObject(ObjectType), NewList(newObject(ObjectType)).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for *: 'object' and 'list'")},
		{Mul, NewList(newObject(ObjectType)).ToObject(), NewList().ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for *: 'list' and 'list'")},
		{Mul, NewList(None, None).ToObject(), NewInt(MaxInt).ToObject(), nil, mustCreateException(OverflowErrorType, "result too large")},
	}
	for _, cas := range cases {
		testCase := invokeTestCase{args: wrapArgs(cas.v, cas.w), want: cas.want, wantExc: cas.wantExc}
		if err := runInvokeTestCase(wrapFuncForTest(cas.fun), &testCase); err != "" {
			t.Error(err)
		}
	}
}

func TestListCompare(t *testing.T) {
	o := newObject(ObjectType)
	cases := []invokeTestCase{
		{args: wrapArgs(NewList(), NewList()), want: compareAllResultEq},
		{args: wrapArgs(newTestList("foo", o), newTestList("foo", o)), want: compareAllResultEq},
		{args: wrapArgs(newTestList(4), newTestList(3, 0)), want: compareAllResultGT},
		{args: wrapArgs(newTestList(4), newTestList(4, 3, 0)), want: compareAllResultLT},
		{args: wrapArgs(NewList(o), NewList()), want: compareAllResultGT},
		{args: wrapArgs(NewList(o), newTestList("foo")), want: compareAllResultLT},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(compareAll, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestListCount(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(NewList(), NewInt(1)), want: NewInt(0).ToObject()},
		{args: wrapArgs(NewList(None, None, None), None), want: NewInt(3).ToObject()},
		{args: wrapArgs(newTestList()), wantExc: mustCreateException(TypeErrorType, "'count' of 'list' requires 2 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(ListType, "count", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestListDelItem(t *testing.T) {
	badIndexType := newTestClass("badIndex", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__index__": newBuiltinFunction("__index__", func(f *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return nil, f.RaiseType(ValueErrorType, "wut")
		}).ToObject(),
	}))
	delItem := mustNotRaise(GetAttr(NewRootFrame(), ListType.ToObject(), NewStr("__delitem__"), nil))
	fun := wrapFuncForTest(func(f *Frame, l *List, key *Object) (*Object, *BaseException) {
		_, raised := delItem.Call(f, wrapArgs(l, key), nil)
		if raised != nil {
			return nil, raised
		}
		return l.ToObject(), nil
	})
	cases := []invokeTestCase{
		{args: wrapArgs(newTestRange(3), 0), want: newTestList(1, 2).ToObject()},
		{args: wrapArgs(newTestRange(3), 2), want: newTestList(0, 1).ToObject()},
		{args: wrapArgs(NewList(), 101), wantExc: mustCreateException(IndexErrorType, "index out of range")},
		{args: wrapArgs(NewList(), newTestSlice(50, 100)), want: NewList().ToObject()},
		{args: wrapArgs(newTestList(1, 2, 3, 4, 5), newTestSlice(1, 3, None)), want: newTestList(1, 4, 5).ToObject()},
		{args: wrapArgs(newTestList(1, 2, 3, 4, 5), newTestSlice(1, None, 2)), want: newTestList(1, 3, 5).ToObject()},
		{args: wrapArgs(newTestList(1, 2, 3, 4, 5), newTestSlice(big.NewInt(1), None, 2)), want: newTestList(1, 3, 5).ToObject()},
		{args: wrapArgs(newTestList(1, 2, 3, 4, 5), newTestSlice(1, big.NewInt(5), 2)), want: newTestList(1, 3, 5).ToObject()},
		{args: wrapArgs(newTestList(1, 2, 3, 4, 5), newTestSlice(1, None, big.NewInt(2))), want: newTestList(1, 3, 5).ToObject()},
		{args: wrapArgs(newTestList(1, 2, 3, 4, 5), newTestSlice(1.0, 3, None)), wantExc: mustCreateException(TypeErrorType, errBadSliceIndex)},
		{args: wrapArgs(newTestList(1, 2, 3, 4, 5), newTestSlice(None, None, 4)), want: newTestList(2, 3, 4).ToObject()},
		{args: wrapArgs(newTestRange(10), newTestSlice(1, 8, 3)), want: newTestList(0, 2, 3, 5, 6, 8, 9).ToObject()},
		{args: wrapArgs(newTestList(1, 2, 3), newTestSlice(1, None, 0)), wantExc: mustCreateException(ValueErrorType, "slice step cannot be zero")},
		{args: wrapArgs(newTestList(true), None), wantExc: mustCreateException(TypeErrorType, "list indices must be integers, not NoneType")},
		{args: wrapArgs(newTestList(true), newObject(badIndexType)), wantExc: mustCreateException(ValueErrorType, "wut")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestListIndex(t *testing.T) {
	intIndexType := newTestClass("IntIndex", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__index__": newBuiltinFunction("__index__", func(f *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return NewInt(0).ToObject(), nil
		}).ToObject(),
	}))
	cases := []invokeTestCase{
		// {args: wrapArgs(newTestList(), 1, "foo"), wantExc: mustCreateException(TypeErrorType, "slice indices must be integers or None or have an __index__ method")},
		{args: wrapArgs(newTestList(10, 20, 30), 20), want: NewInt(1).ToObject()},
		{args: wrapArgs(newTestList(10, 20, 30), 20, newObject(intIndexType)), want: NewInt(1).ToObject()},
		{args: wrapArgs(newTestList(0, "foo", "bar"), "foo"), want: NewInt(1).ToObject()},
		{args: wrapArgs(newTestList(0, 1, 2, 3, 4), 3, 3), want: NewInt(3).ToObject()},
		{args: wrapArgs(newTestList(0, 2.0, 2, 3, 4, 2, 1, "foo"), 3, 3), want: NewInt(3).ToObject()},
		{args: wrapArgs(newTestList(0, 1, 2, 3, 4), 3, 4), wantExc: mustCreateException(ValueErrorType, "3 is not in list")},
		{args: wrapArgs(newTestList(0, 1, 2, 3, 4), 3, 0, 4), want: NewInt(3).ToObject()},
		{args: wrapArgs(newTestList(0, 1, 2, 3, 4), 3, 0, 3), wantExc: mustCreateException(ValueErrorType, "3 is not in list")},
		{args: wrapArgs(newTestList(0, 1, 2, 3, 4), 3, -2), want: NewInt(3).ToObject()},
		{args: wrapArgs(newTestList(0, 1, 2, 3, 4), 3, -1), wantExc: mustCreateException(ValueErrorType, "3 is not in list")},
		{args: wrapArgs(newTestList(0, 1, 2, 3, 4), 3, 0, -1), want: NewInt(3).ToObject()},
		{args: wrapArgs(newTestList(0, 1, 2, 3, 4), 3, 0, -2), wantExc: mustCreateException(ValueErrorType, "3 is not in list")},
		{args: wrapArgs(newTestList(0, 1, 2, 3, 4), 3, 0, 999), want: NewInt(3).ToObject()},
		{args: wrapArgs(newTestList(0, 1, 2, 3, 4), "foo", 0, 999), wantExc: mustCreateException(ValueErrorType, "'foo' is not in list")},
		{args: wrapArgs(newTestList(0, 1, 2, 3, 4), 3, 999), wantExc: mustCreateException(ValueErrorType, "3 is not in list")},
		{args: wrapArgs(newTestList(0, 1, 2, 3, 4), 3, 5, 0), wantExc: mustCreateException(ValueErrorType, "3 is not in list")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(ListType, "index", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestListRemove(t *testing.T) {
	fun := newBuiltinFunction("TestListRemove", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		rem, raised := GetAttr(f, ListType.ToObject(), NewStr("remove"), nil)
		if raised != nil {
			return nil, raised
		}
		if _, raised := rem.Call(f, args, nil); raised != nil {
			return nil, raised
		}
		return args[0], nil
	}).ToObject()
	cases := []invokeTestCase{
		{args: wrapArgs(newTestList(1, 2, 3), 2), want: newTestList(1, 3).ToObject()},
		{args: wrapArgs(newTestList(1, 2, 3, 2, 1), 2), want: newTestList(1, 3, 2, 1).ToObject()},
		{args: wrapArgs(NewList()), wantExc: mustCreateException(TypeErrorType, "'remove' of 'list' requires 2 arguments")},
		{args: wrapArgs(NewList(), 1), wantExc: mustCreateException(ValueErrorType, "list.remove(x): x not in list")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func BenchmarkListContains(b *testing.B) {
	b.Run("false-3", func(b *testing.B) {
		t := newTestList("foo", 42, "bar").ToObject()
		a := wrapArgs(1)[0]
		f := NewRootFrame()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Contains(f, t, a)
		}
	})

	b.Run("false-10", func(b *testing.B) {
		t := newTestList("foo", 42, "bar", "foo", 42, "bar", "foo", 42, "bar", "baz").ToObject()
		a := wrapArgs(1)[0]
		f := NewRootFrame()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Contains(f, t, a)
		}
	})

	b.Run("true-3.1", func(b *testing.B) {
		t := newTestList("foo", 42, "bar").ToObject()
		a := wrapArgs("foo")[0]
		f := NewRootFrame()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Contains(f, t, a)
		}
	})

	b.Run("true-3.3", func(b *testing.B) {
		t := newTestList("foo", 42, "bar").ToObject()
		a := wrapArgs("bar")[0]
		f := NewRootFrame()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Contains(f, t, a)
		}
	})

	b.Run("true-10.10", func(b *testing.B) {
		t := newTestList("foo", 42, "bar", "foo", 42, "bar", "foo", 42, "bar", "baz").ToObject()
		a := wrapArgs("baz")[0]
		f := NewRootFrame()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Contains(f, t, a)
		}
	})
}

func TestListGetItem(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(newTestRange(20), 0), want: NewInt(0).ToObject()},
		{args: wrapArgs(newTestRange(20), 19), want: NewInt(19).ToObject()},
		{args: wrapArgs(NewList(), 101), wantExc: mustCreateException(IndexErrorType, "index out of range")},
		{args: wrapArgs(NewList(), newTestSlice(50, 100)), want: NewList().ToObject()},
		{args: wrapArgs(newTestList(1, 2, 3, 4, 5), newTestSlice(1, 3, None)), want: newTestList(2, 3).ToObject()},
		{args: wrapArgs(newTestList(1, 2, 3, 4, 5), newTestSlice(1, None, 2)), want: newTestList(2, 4).ToObject()},
		{args: wrapArgs(newTestList(1, 2, 3, 4, 5), newTestSlice(big.NewInt(1), None, 2)), want: newTestList(2, 4).ToObject()},
		{args: wrapArgs(newTestList(1, 2, 3, 4, 5), newTestSlice(1, big.NewInt(5), 2)), want: newTestList(2, 4).ToObject()},
		{args: wrapArgs(newTestList(1, 2, 3, 4, 5), newTestSlice(1, None, big.NewInt(2))), want: newTestList(2, 4).ToObject()},
		{args: wrapArgs(newTestList(1, 2, 3, 4, 5), newTestSlice(1.0, 3, None)), wantExc: mustCreateException(TypeErrorType, errBadSliceIndex)},
		{args: wrapArgs(newTestList(1, 2, 3), newTestSlice(1, None, 0)), wantExc: mustCreateException(ValueErrorType, "slice step cannot be zero")},
		{args: wrapArgs(newTestList(true), None), wantExc: mustCreateException(TypeErrorType, "list indices must be integers, not NoneType")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(ListType, "__getitem__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestListInplaceOps(t *testing.T) {
	cases := []struct {
		fun     func(f *Frame, v, w *Object) (*Object, *BaseException)
		v, w    *Object
		want    *Object
		wantExc *BaseException
	}{
		{IAdd, newTestList(3).ToObject(), newTestList("foo").ToObject(), newTestList(3, "foo").ToObject(), nil},
		{IAdd, NewList(None).ToObject(), NewList().ToObject(), NewList(None).ToObject(), nil},
		{IAdd, NewList().ToObject(), newObject(ObjectType), nil, mustCreateException(TypeErrorType, "'object' object is not iterable")},
		{IMul, NewList().ToObject(), NewInt(10).ToObject(), NewList().ToObject(), nil},
		{IMul, newTestList("baz").ToObject(), NewInt(-2).ToObject(), NewList().ToObject(), nil},
		{IMul, NewList().ToObject(), None, nil, mustCreateException(TypeErrorType, "can't multiply sequence by non-int of type 'NoneType'")},
	}
	for _, cas := range cases {
		switch got, result := checkInvokeResult(wrapFuncForTest(cas.fun), []*Object{cas.v, cas.w}, cas.want, cas.wantExc); result {
		case checkInvokeResultExceptionMismatch:
			t.Errorf("%s(%v, %v) raised %v, want %v", getFuncName(cas.fun), cas.v, cas.w, got, cas.wantExc)
		case checkInvokeResultReturnValueMismatch:
			t.Errorf("%s(%v, %v) = %v, want %v", getFuncName(cas.fun), cas.v, cas.w, got, cas.want)
		default:
			if got != nil && got != cas.v {
				t.Errorf("%s(%v, %v) did not return identity", getFuncName(cas.fun), cas.v, cas.w)
			}
		}
	}
}

func TestListIsTrue(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(NewList()), want: False.ToObject()},
		{args: wrapArgs(newTestList("foo", None)), want: True.ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(IsTrue), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestListAppend(t *testing.T) {
	fun := newBuiltinFunction("TestListAppend", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		if raised := checkFunctionArgs(f, "TestListAppend", args, ListType, ObjectType); raised != nil {
			return nil, raised
		}
		app, raised := GetAttr(f, ListType.ToObject(), NewStr("append"), nil)
		if raised != nil {
			return nil, raised
		}
		if _, raised := app.Call(f, args, nil); raised != nil {
			return nil, raised
		}
		return args[0], nil
	}).ToObject()
	cases := []invokeTestCase{
		{args: wrapArgs(NewList(), None), want: NewList(None).ToObject()},
		{args: wrapArgs(NewList(None), 42), want: newTestList(None, 42).ToObject()},
		{args: wrapArgs(newTestList(None, 42), "foo"), want: newTestList(None, 42, "foo").ToObject()},
		{args: wrapArgs(newTestRange(100), 100), want: newTestRange(101).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestListExtend(t *testing.T) {
	extend := mustNotRaise(GetAttr(NewRootFrame(), ListType.ToObject(), NewStr("extend"), nil))
	fun := newBuiltinFunction("TestListExtend", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		if _, raised := extend.Call(f, args, nil); raised != nil {
			return nil, raised
		}
		return args[0], nil
	}).ToObject()
	cases := []invokeTestCase{
		{args: wrapArgs(newTestList(), newTestTuple()), want: newTestList().ToObject()},
		{args: wrapArgs(newTestList(), newTestList()), want: newTestList().ToObject()},
		{args: wrapArgs(newTestList(3), newTestList("foo")), want: newTestList(3, "foo").ToObject()},
		{args: wrapArgs(newTestList(), newTestList("foo")), want: newTestList("foo").ToObject()},
		{args: wrapArgs(newTestList(3), newTestList()), want: newTestList(3).ToObject()},
		{args: wrapArgs(NewStr(""), newTestList()), wantExc: mustCreateException(TypeErrorType, "unbound method extend() must be called with list instance as first argument (got str instance instead)")},
		{args: wrapArgs(None, None), wantExc: mustCreateException(TypeErrorType, "unbound method extend() must be called with list instance as first argument (got NoneType instance instead)")},
		{args: wrapArgs(newTestList(3), None), wantExc: mustCreateException(TypeErrorType, "'NoneType' object is not iterable")},
		{args: wrapArgs(newTestRange(5), newTestList(3)), want: newTestList(0, 1, 2, 3, 4, 3).ToObject()},
		{args: wrapArgs(newTestRange(5), newTestList(3)), want: newTestList(0, 1, 2, 3, 4, 3).ToObject()},
		{args: wrapArgs(newTestTuple(1, 2, 3), newTestList(3)), wantExc: mustCreateException(TypeErrorType, "unbound method extend() must be called with list instance as first argument (got tuple instance instead)")},
		{args: wrapArgs(newTestList(4), newTestTuple(1, 2, 3)), want: newTestList(4, 1, 2, 3).ToObject()},
		{args: wrapArgs(newTestList()), wantExc: mustCreateException(TypeErrorType, "extend() takes exactly one argument (1 given)")},
		{args: wrapArgs(newTestList(), newTestTuple(), newTestTuple()), wantExc: mustCreateException(TypeErrorType, "extend() takes exactly one argument (3 given)")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestListLen(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(NewList()), want: NewInt(0).ToObject()},
		{args: wrapArgs(NewList(None, None, None)), want: NewInt(3).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(ListType, "__len__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestListNew(t *testing.T) {
	cases := []invokeTestCase{
		{want: NewList().ToObject()},
		{args: wrapArgs(newTestTuple(1, 2, 3)), want: newTestList(1, 2, 3).ToObject()},
		{args: wrapArgs(newTestDict(1, "foo", "bar", None)), want: newTestList(1, "bar").ToObject()},
		{args: wrapArgs(42), wantExc: mustCreateException(TypeErrorType, "'int' object is not iterable")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(ListType.ToObject(), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestListReverse(t *testing.T) {
	reverse := mustNotRaise(GetAttr(NewRootFrame(), ListType.ToObject(), NewStr("reverse"), nil))
	fun := wrapFuncForTest(func(f *Frame, o *Object, args ...*Object) (*Object, *BaseException) {
		_, raised := reverse.Call(f, append(Args{o}, args...), nil)
		if raised != nil {
			return nil, raised
		}
		return o, nil
	})
	cases := []invokeTestCase{
		{args: wrapArgs(NewList()), want: NewList().ToObject()},
		{args: wrapArgs(newTestList(1, 2, 3)), want: newTestList(3, 2, 1).ToObject()},
		{args: wrapArgs(NewList(), 123), wantExc: mustCreateException(TypeErrorType, "'reverse' of 'list' requires 1 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestListStrRepr(t *testing.T) {
	recursiveList := newTestList("foo").ToObject()
	listAppend(NewRootFrame(), []*Object{recursiveList, recursiveList}, nil)
	cases := []invokeTestCase{
		{args: wrapArgs(NewList()), want: NewStr("[]").ToObject()},
		{args: wrapArgs(newTestList("foo")), want: NewStr("['foo']").ToObject()},
		{args: wrapArgs(newTestList(TupleType, ExceptionType)), want: NewStr("[<type 'tuple'>, <type 'Exception'>]").ToObject()},
		{args: wrapArgs(recursiveList), want: NewStr("['foo', [...]]").ToObject()},
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

func TestListInsert(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, l *List, args ...*Object) (*Object, *BaseException) {
		insert, raised := GetAttr(NewRootFrame(), l.ToObject(), NewStr("insert"), nil)
		if raised != nil {
			return nil, raised
		}
		if _, raised := insert.Call(f, args, nil); raised != nil {
			return nil, raised
		}
		return l.ToObject(), nil
	})
	cases := []invokeTestCase{
		{args: wrapArgs(NewList(), 0, None), want: NewList(None).ToObject()},
		{args: wrapArgs(newTestList("bar"), -100, "foo"), want: newTestList("foo", "bar").ToObject()},
		{args: wrapArgs(newTestList("foo", "bar"), 101, "baz"), want: newTestList("foo", "bar", "baz").ToObject()},
		{args: wrapArgs(newTestList("a", "c"), 1, "b"), want: newTestList("a", "b", "c").ToObject()},
		{args: wrapArgs(newTestList(1, 2), 0, 0), want: newTestList(0, 1, 2).ToObject()},
		{args: wrapArgs(NewList()), wantExc: mustCreateException(TypeErrorType, "'insert' of 'list' requires 3 arguments")},
		{args: wrapArgs(NewList(), "foo", 123), wantExc: mustCreateException(TypeErrorType, "'insert' requires a 'int' object but received a 'str'")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestListIter(t *testing.T) {
	fun := newBuiltinFunction("TestListIter", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		if raised := checkFunctionArgs(f, "TestListIter", args, ListType); raised != nil {
			return nil, raised
		}
		var got []*Object
		iter, raised := Iter(f, args[0])
		if raised != nil {
			return nil, raised
		}
		raised = seqApply(f, iter, func(elems []*Object, _ bool) *BaseException {
			got = make([]*Object, len(elems))
			copy(got, elems)
			return nil
		})
		if raised != nil {
			return nil, raised
		}
		return NewList(got...).ToObject(), nil
	}).ToObject()
	o := newObject(ObjectType)
	cases := []invokeTestCase{
		{args: wrapArgs(NewList()), want: NewList().ToObject()},
		{args: wrapArgs(newTestList(1, o, "foo")), want: newTestList(1, o, "foo").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestListIteratorIter(t *testing.T) {
	iter := newListIterator(NewList())
	cas := &invokeTestCase{args: wrapArgs(iter), want: iter}
	if err := runInvokeMethodTestCase(listIteratorType, "__iter__", cas); err != "" {
		t.Error(err)
	}
}

func TestListPop(t *testing.T) {
	pop := mustNotRaise(GetAttr(NewRootFrame(), ListType.ToObject(), NewStr("pop"), nil))
	fun := wrapFuncForTest(func(f *Frame, l *List, args ...*Object) (*Tuple, *BaseException) {
		result, raised := pop.Call(f, append(Args{l.ToObject()}, args...), nil)
		if raised != nil {
			return nil, raised
		}
		return newTestTuple(result, l), nil
	})
	cases := []invokeTestCase{
		{args: wrapArgs(newTestList(1)), want: newTestTuple(1, newTestList().ToObject()).ToObject()},
		{args: wrapArgs(newTestList(1), 0), want: newTestTuple(1, newTestList().ToObject()).ToObject()},
		{args: wrapArgs(newTestList(-1, 0, 1)), want: newTestTuple(1, newTestList(-1, 0).ToObject()).ToObject()},
		{args: wrapArgs(newTestList(-1, 0, 1), 0), want: newTestTuple(-1, newTestList(0, 1).ToObject()).ToObject()},
		{args: wrapArgs(newTestList(-1, 0, 1), NewLong(big.NewInt(1))), want: newTestTuple(0, newTestList(-1, 1).ToObject()).ToObject()},
		{args: wrapArgs(newTestList(-1, 0, 1), None), wantExc: mustCreateException(TypeErrorType, "an integer is required")},
		{args: wrapArgs(newTestList(-1, 0, 1), None), wantExc: mustCreateException(TypeErrorType, "an integer is required")},
		{args: wrapArgs(newTestList(-1, 0, 1), 3), wantExc: mustCreateException(IndexErrorType, "list index out of range")},
		{args: wrapArgs(newTestList()), wantExc: mustCreateException(IndexErrorType, "list index out of range")},
		{args: wrapArgs(newTestList(), 0), wantExc: mustCreateException(IndexErrorType, "list index out of range")},
		{args: wrapArgs(newTestList(), 1), wantExc: mustCreateException(IndexErrorType, "list index out of range")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestListSetItem(t *testing.T) {
	fun := newBuiltinFunction("TestListSetItem", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		// Check that there is at least one arg, but otherwise leave
		// the validation to __setitem__.
		if raised := checkFunctionVarArgs(f, "TestListSetItem", args, ObjectType); raised != nil {
			return nil, raised
		}
		setitem, raised := GetAttr(f, args[0], NewStr("__setitem__"), nil)
		if raised != nil {
			return nil, raised
		}
		_, raised = setitem.Call(f, args[1:], nil)
		if raised != nil {
			return nil, raised
		}
		return args[0], nil
	}).ToObject()
	cases := []invokeTestCase{
		{args: wrapArgs(newTestList("foo", "bar"), 1, None), want: newTestList("foo", None).ToObject()},
		{args: wrapArgs(newTestList(1, 2, 3), newTestSlice(0), newTestList(0)), want: newTestList(0, 1, 2, 3).ToObject()},
		{args: wrapArgs(newTestList(1, 2, 3), newTestSlice(1), newTestList(4)), want: newTestList(4, 2, 3).ToObject()},
		{args: wrapArgs(newTestList(1, 2, 3), newTestSlice(2, None), newTestList("foo")), want: newTestList(1, 2, "foo").ToObject()},
		{args: wrapArgs(newTestList(1, 2, 3), newTestSlice(100, None), newTestList("foo")), want: newTestList(1, 2, 3, "foo").ToObject()},
		{args: wrapArgs(newTestList(1, 2, 4, 5), newTestSlice(1, None, 2), newTestTuple("foo", "bar")), want: newTestList(1, "foo", 4, "bar").ToObject()},
		{args: wrapArgs(newTestList(1, 2, 3), newTestSlice(None, None, 2), newTestList("foo")), wantExc: mustCreateException(ValueErrorType, "attempt to assign sequence of size 1 to extended slice of size 2")},
		{args: wrapArgs(newTestRange(100), newTestSlice(None, None), NewList()), want: NewList().ToObject()},
		{args: wrapArgs(NewList(), newTestSlice(4, 8, 0), NewList()), wantExc: mustCreateException(ValueErrorType, "slice step cannot be zero")},
		{args: wrapArgs(newTestList("foo", "bar"), -100, None), wantExc: mustCreateException(IndexErrorType, "index out of range")},
		{args: wrapArgs(NewList(), 101, None), wantExc: mustCreateException(IndexErrorType, "index out of range")},
		{args: wrapArgs(newTestList(true), None, false), wantExc: mustCreateException(TypeErrorType, "list indices must be integers, not NoneType")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestListSort(t *testing.T) {
	sort := mustNotRaise(GetAttr(NewRootFrame(), ListType.ToObject(), NewStr("sort"), nil))
	fun := newBuiltinFunction("TestListSort", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		if _, raised := sort.Call(f, args, nil); raised != nil {
			return nil, raised
		}
		return args[0], nil
	}).ToObject()
	cases := []invokeTestCase{
		{args: wrapArgs(NewList()), want: NewList().ToObject()},
		{args: wrapArgs(newTestList("foo", "bar")), want: newTestList("bar", "foo").ToObject()},
		{args: wrapArgs(newTestList(true, false)), want: newTestList(false, true).ToObject()},
		{args: wrapArgs(newTestList(1, 2, 0, 3)), want: newTestRange(4).ToObject()},
		{args: wrapArgs(newTestRange(100)), want: newTestRange(100).ToObject()},
		{args: wrapArgs(1), wantExc: mustCreateException(TypeErrorType, "unbound method sort() must be called with list instance as first argument (got int instance instead)")},
		{args: wrapArgs(NewList(), 1), wantExc: mustCreateException(TypeErrorType, "'sort' of 'list' requires 1 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func newTestRange(n int) *List {
	elems := make([]*Object, n)
	for i := 0; i < n; i++ {
		elems[i] = NewInt(i).ToObject()
	}
	return NewList(elems...)
}

func newTestList(elems ...interface{}) *List {
	listElems, raised := seqWrapEach(NewRootFrame(), elems...)
	if raised != nil {
		panic(raised)
	}
	return NewList(listElems...)
}

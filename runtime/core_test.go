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
	"bytes"
	"fmt"
	"io"
	"math/big"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"testing"
)

func TestAssert(t *testing.T) {
	assert := newBuiltinFunction("TestAssert", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		switch argc := len(args); argc {
		case 1:
			if raised := Assert(f, args[0], nil); raised != nil {
				return nil, raised
			}
		case 2:
			if raised := Assert(f, args[0], args[1]); raised != nil {
				return nil, raised
			}
		default:
			return nil, f.RaiseType(SystemErrorType, fmt.Sprintf("Assert expected 1 or 2 args, got %d", argc))
		}
		return None, nil
	}).ToObject()
	emptyAssert := toBaseExceptionUnsafe(mustNotRaise(AssertionErrorType.Call(newFrame(nil), nil, nil)))
	cases := []invokeTestCase{
		{args: wrapArgs(true), want: None},
		{args: wrapArgs(NewTuple(None)), want: None},
		{args: wrapArgs(None), wantExc: emptyAssert},
		{args: wrapArgs(NewDict()), wantExc: emptyAssert},
		{args: wrapArgs(false, "foo"), wantExc: mustCreateException(AssertionErrorType, "foo")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(assert, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestBinaryOps(t *testing.T) {
	fooType := newTestClass("Foo", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__add__": newBuiltinFunction("__add__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return NewStr("foo add").ToObject(), nil
		}).ToObject(),
		"__radd__": newBuiltinFunction("__add__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return NewStr("foo radd").ToObject(), nil
		}).ToObject(),
	}))
	barType := newTestClass("Bar", []*Type{fooType}, NewDict())
	bazType := newTestClass("Baz", []*Type{IntType}, newStringDict(map[string]*Object{
		"__rdiv__": newBuiltinFunction("__rdiv__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			s, raised := ToStr(f, args[1])
			if raised != nil {
				return nil, raised
			}
			return s.ToObject(), nil
		}).ToObject(),
	}))
	inplaceType := newTestClass("Inplace", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__iadd__": newBuiltinFunction("__iadd__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return args[1], nil
		}).ToObject(),
		"__iand__": newBuiltinFunction("__iand__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return args[1], nil
		}).ToObject(),
		"__idiv__": newBuiltinFunction("__idiv__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return args[1], nil
		}).ToObject(),
		"__imod__": newBuiltinFunction("__imod__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return args[1], nil
		}).ToObject(),
		"__imul__": newBuiltinFunction("__imul__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return args[1], nil
		}).ToObject(),
		"__ior__": newBuiltinFunction("__ior__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return args[1], nil
		}).ToObject(),
		"__isub__": newBuiltinFunction("__isub__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return args[1], nil
		}).ToObject(),
		"__ixor__": newBuiltinFunction("__ixor__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return args[1], nil
		}).ToObject(),
	}))
	cases := []struct {
		fun     func(f *Frame, v, w *Object) (*Object, *BaseException)
		v, w    *Object
		want    *Object
		wantExc *BaseException
	}{
		{Add, NewStr("foo").ToObject(), NewStr("bar").ToObject(), NewStr("foobar").ToObject(), nil},
		{Add, NewStr("foo").ToObject(), NewStr("bar").ToObject(), NewStr("foobar").ToObject(), nil},
		{Add, newObject(fooType), newObject(ObjectType), NewStr("foo add").ToObject(), nil},
		{And, NewInt(-42).ToObject(), NewInt(244).ToObject(), NewInt(212).ToObject(), nil},
		{And, NewInt(42).ToObject(), NewStr("foo").ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for &: 'int' and 'str'")},
		{Add, newObject(fooType), newObject(barType), NewStr("foo add").ToObject(), nil},
		{Div, NewInt(123).ToObject(), newObject(bazType), NewStr("123").ToObject(), nil},
		{IAdd, NewStr("foo").ToObject(), NewStr("bar").ToObject(), NewStr("foobar").ToObject(), nil},
		{IAdd, NewStr("foo").ToObject(), NewStr("bar").ToObject(), NewStr("foobar").ToObject(), nil},
		{IAdd, newObject(fooType), newObject(ObjectType), NewStr("foo add").ToObject(), nil},
		{IAdd, newObject(inplaceType), NewStr("foo").ToObject(), NewStr("foo").ToObject(), nil},
		{IAnd, NewInt(9).ToObject(), NewInt(12).ToObject(), NewInt(8).ToObject(), nil},
		{IAnd, newObject(inplaceType), NewStr("foo").ToObject(), NewStr("foo").ToObject(), nil},
		{IAnd, newObject(ObjectType), newObject(fooType), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for &: 'object' and 'Foo'")},
		{IDiv, NewInt(123).ToObject(), newObject(bazType), NewStr("123").ToObject(), nil},
		{IDiv, newObject(inplaceType), NewInt(42).ToObject(), NewInt(42).ToObject(), nil},
		{IMod, NewInt(24).ToObject(), NewInt(6).ToObject(), NewInt(0).ToObject(), nil},
		{IMod, newObject(inplaceType), NewFloat(3.14).ToObject(), NewFloat(3.14).ToObject(), nil},
		{IMul, NewStr("foo").ToObject(), NewInt(3).ToObject(), NewStr("foofoofoo").ToObject(), nil},
		{IMul, newObject(inplaceType), True.ToObject(), True.ToObject(), nil},
		{IMul, newObject(ObjectType), newObject(fooType), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for *: 'object' and 'Foo'")},
		{IOr, newObject(inplaceType), NewInt(42).ToObject(), NewInt(42).ToObject(), nil},
		{IOr, NewInt(9).ToObject(), NewInt(12).ToObject(), NewInt(13).ToObject(), nil},
		{IOr, newObject(ObjectType), newObject(fooType), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for |: 'object' and 'Foo'")},
		{ISub, NewInt(3).ToObject(), NewInt(-3).ToObject(), NewInt(6).ToObject(), nil},
		{ISub, newObject(inplaceType), None, None, nil},
		{IXor, newObject(inplaceType), None, None, nil},
		{IXor, NewInt(9).ToObject(), NewInt(12).ToObject(), NewInt(5).ToObject(), nil},
		{IXor, newObject(ObjectType), newObject(fooType), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for ^: 'object' and 'Foo'")},
		{Mod, NewInt(24).ToObject(), NewInt(6).ToObject(), NewInt(0).ToObject(), nil},
		{Mul, NewStr("foo").ToObject(), NewInt(3).ToObject(), NewStr("foofoofoo").ToObject(), nil},
		{Mul, newObject(ObjectType), newObject(fooType), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for *: 'object' and 'Foo'")},
		{Or, NewInt(-42).ToObject(), NewInt(244).ToObject(), NewInt(-10).ToObject(), nil},
		{Or, NewInt(42).ToObject(), NewStr("foo").ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for |: 'int' and 'str'")},
		{Sub, NewInt(3).ToObject(), NewInt(-3).ToObject(), NewInt(6).ToObject(), nil},
		{Xor, NewInt(-42).ToObject(), NewInt(244).ToObject(), NewInt(-222).ToObject(), nil},
		{Xor, NewInt(42).ToObject(), NewStr("foo").ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for ^: 'int' and 'str'")},
	}
	for _, cas := range cases {
		testCase := invokeTestCase{wrapArgs(cas.v, cas.w), nil, cas.want, cas.wantExc}
		if err := runInvokeTestCase(wrapFuncForTest(cas.fun), &testCase); err != "" {
			t.Error(err)
		}
	}
}

func TestCompareDefault(t *testing.T) {
	o1, o2 := newObject(ObjectType), newObject(ObjectType)
	// Make sure uintptr(o1) < uintptr(o2).
	if uintptr(o1.toPointer()) > uintptr(o2.toPointer()) {
		o1, o2 = o2, o1
	}
	// When type names are equal, comparison should fall back to comparing
	// the pointer values of the types of the objects.
	fakeObjectType := newTestClass("object", []*Type{ObjectType}, NewDict())
	o3, o4 := newObject(fakeObjectType), newObject(ObjectType)
	if uintptr(o3.typ.toPointer()) > uintptr(o4.typ.toPointer()) {
		o3, o4 = o4, o3
	}
	// An int subtype that equals anything, but doesn't override other
	// comparison methods.
	eqType := newTestClass("Eq", []*Type{IntType}, newStringDict(map[string]*Object{
		"__eq__": newBuiltinFunction("__eq__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return True.ToObject(), nil
		}).ToObject(),
		"__repr__": newBuiltinFunction("__repr__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return NewStr("<Foo>").ToObject(), nil
		}).ToObject(),
	}))
	cases := []invokeTestCase{
		{args: wrapArgs(true, o1), want: compareAllResultLT},
		{args: wrapArgs(o1, -306), want: compareAllResultGT},
		{args: wrapArgs(-306, o1), want: compareAllResultLT},
		{args: wrapArgs(NewList(), None), want: compareAllResultGT},
		{args: wrapArgs(None, "foo"), want: compareAllResultLT},
		{args: wrapArgs(o1, o1), want: compareAllResultEq},
		{args: wrapArgs(o1, o2), want: compareAllResultLT},
		{args: wrapArgs(o2, o1), want: compareAllResultGT},
		{args: wrapArgs(o3, o4), want: compareAllResultLT},
		{args: wrapArgs(o4, o3), want: compareAllResultGT},
		// The equality test should dispatch to the eqType instance and
		// return true.
		{args: wrapArgs(42, newObject(eqType)), want: newTestTuple(false, false, true, true, true, true).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(compareAll, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestContains(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(NewTuple(), 42), want: False.ToObject()},
		{args: wrapArgs(newTestList("foo", "bar"), "bar"), want: True.ToObject()},
		{args: wrapArgs(newTestDict(1, "foo", 2, "bar", 3, "baz"), 2), want: True.ToObject()},
		{args: wrapArgs("foobar", "ooba"), want: True.ToObject()},
		{args: wrapArgs("qux", "ooba"), want: False.ToObject()},
		{args: wrapArgs(3.14, None), wantExc: mustCreateException(TypeErrorType, "'float' object is not iterable")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(Contains), &cas); err != "" {
			t.Error(err)
		}
	}
}

// DelAttr is tested in TestObjectDelAttr.

func TestDelItem(t *testing.T) {
	delItem := newBuiltinFunction("TestDelItem", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		if raised := checkFunctionArgs(f, "TestDelItem", args, ObjectType, ObjectType); raised != nil {
			return nil, raised
		}
		o := args[0]
		if raised := DelItem(f, o, args[1]); raised != nil {
			return nil, raised
		}
		return args[0], nil
	}).ToObject()
	cases := []invokeTestCase{
		{args: wrapArgs(newTestDict("foo", None), "foo"), want: NewDict().ToObject()},
		{args: wrapArgs(NewDict(), "foo"), wantExc: mustCreateException(KeyErrorType, "foo")},
		{args: wrapArgs(123, "bar"), wantExc: mustCreateException(TypeErrorType, "'int' object does not support item deletion")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(delItem, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestFormatException(t *testing.T) {
	f := newFrame(nil)
	cases := []struct {
		o    *Object
		want string
	}{
		{mustNotRaise(ExceptionType.Call(f, nil, nil)), "Exception\n"},
		{mustNotRaise(AttributeErrorType.Call(f, wrapArgs(""), nil)), "AttributeError\n"},
		{mustNotRaise(TypeErrorType.Call(f, wrapArgs(123), nil)), "TypeError: 123\n"},
		{mustNotRaise(AttributeErrorType.Call(f, wrapArgs("hello", "there"), nil)), "AttributeError: ('hello', 'there')\n"},
	}
	for _, cas := range cases {
		if !cas.o.isInstance(BaseExceptionType) {
			t.Errorf("expected FormatException() input to be BaseException, got %s", cas.o.typ.Name())
		} else if got, raised := FormatException(f, toBaseExceptionUnsafe(cas.o)); raised != nil {
			t.Errorf("FormatException(%v) raised %v, want nil", cas.o, raised)
		} else if got != cas.want {
			t.Errorf("FormatException(%v) = %q, want %q", cas.o, got, cas.want)
		}
	}
}

func TestGetAttr(t *testing.T) {
	getAttr := newBuiltinFunction("TestGetAttr", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		expectedTypes := []*Type{ObjectType, StrType, ObjectType}
		argc := len(args)
		if argc == 2 {
			expectedTypes = expectedTypes[:2]
		}
		if raised := checkFunctionArgs(f, "TestGetAttr", args, expectedTypes...); raised != nil {
			return nil, raised
		}
		var def *Object
		if argc > 2 {
			def = args[2]
		}
		s, raised := ToStr(f, args[1])
		if raised != nil {
			return nil, raised
		}
		return GetAttr(f, args[0], s, def)
	}).ToObject()
	fooResult := newObject(ObjectType)
	fooType := newTestClass("Foo", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__getattribute__": newBuiltinFunction("__getattribute__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return fooResult, nil
		}).ToObject(),
	}))
	barType := newTestClass("Bar", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__getattribute__": newBuiltinFunction("__getattribute__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return nil, f.RaiseType(TypeErrorType, "uh oh")
		}).ToObject(),
	}))
	cases := []invokeTestCase{
		{args: wrapArgs(newObject(fooType), "bar"), want: fooResult},
		{args: wrapArgs(newObject(fooType), "baz", None), want: fooResult},
		{args: wrapArgs(newObject(ObjectType), "qux", None), want: None},
		{args: wrapArgs(NewTuple(), "noexist"), wantExc: mustCreateException(AttributeErrorType, "'tuple' object has no attribute 'noexist'")},
		{args: wrapArgs(DictType, "noexist"), wantExc: mustCreateException(AttributeErrorType, "type object 'dict' has no attribute 'noexist'")},
		{args: wrapArgs(newObject(barType), "noexist"), wantExc: mustCreateException(TypeErrorType, "uh oh")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(getAttr, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestGetItem(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(newStringDict(map[string]*Object{"foo": None}), "foo"), want: None},
		{args: wrapArgs(NewDict(), "bar"), wantExc: mustCreateException(KeyErrorType, "bar")},
		{args: wrapArgs(true, "baz"), wantExc: mustCreateException(TypeErrorType, "'bool' object has no attribute '__getitem__'")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(GetItem), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestHash(t *testing.T) {
	badHash := newTestClass("badHash", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__hash__": newBuiltinFunction("__hash__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return args[0], nil
		}).ToObject(),
	}))
	o := newObject(ObjectType)
	cases := []invokeTestCase{
		{args: wrapArgs("foo"), want: hashFoo},
		{args: wrapArgs(123), want: NewInt(123).ToObject()},
		{args: wrapArgs(o), want: NewInt(int(uintptr(o.toPointer()))).ToObject()},
		{args: wrapArgs(NewList()), wantExc: mustCreateException(TypeErrorType, "unhashable type: 'list'")},
		{args: wrapArgs(NewDict()), wantExc: mustCreateException(TypeErrorType, "unhashable type: 'dict'")},
		{args: wrapArgs(newObject(badHash)), wantExc: mustCreateException(TypeErrorType, "an integer is required")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(Hash), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestIndex(t *testing.T) {
	goodType := newTestClass("GoodIndex", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__index__": newBuiltinFunction("__index__", func(f *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return NewInt(123).ToObject(), nil
		}).ToObject(),
	}))
	longType := newTestClass("LongIndex", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__index__": newBuiltinFunction("__index__", func(f *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return NewLong(big.NewInt(123)).ToObject(), nil
		}).ToObject(),
	}))
	raiseType := newTestClass("RaiseIndex", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__index__": newBuiltinFunction("__index__", func(f *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return nil, f.RaiseType(RuntimeErrorType, "uh oh")
		}).ToObject(),
	}))
	badType := newTestClass("BadIndex", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__index__": newBuiltinFunction("__index__", func(f *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return NewFloat(3.14).ToObject(), nil
		}).ToObject(),
	}))
	cases := []invokeTestCase{
		{args: wrapArgs(42), want: NewInt(42).ToObject()},
		{args: wrapArgs(newObject(goodType)), want: NewInt(123).ToObject()},
		{args: wrapArgs(newObject(longType)), want: NewLong(big.NewInt(123)).ToObject()},
		{args: wrapArgs(newObject(raiseType)), wantExc: mustCreateException(RuntimeErrorType, "uh oh")},
		{args: wrapArgs(newObject(badType)), wantExc: mustCreateException(TypeErrorType, "__index__ returned non-(int,long) (type float)")},
		{args: wrapArgs("abc"), want: None},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(Index), &cas); err != "" {
			t.Error(err)
		}
	}
	cases = []invokeTestCase{
		{args: wrapArgs(42), want: NewInt(42).ToObject()},
		{args: wrapArgs(newObject(goodType)), want: NewInt(123).ToObject()},
		{args: wrapArgs(newObject(raiseType)), wantExc: mustCreateException(RuntimeErrorType, "uh oh")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(cas.args[0].typ, "__index__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestInvert(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(42), want: NewInt(-43).ToObject()},
		{args: wrapArgs(0), want: NewInt(-1).ToObject()},
		{args: wrapArgs(-35935), want: NewInt(35934).ToObject()},
		{args: wrapArgs("foo"), wantExc: mustCreateException(TypeErrorType, "bad operand type for unary ~: 'str'")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(Invert), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestIsInstanceIsSubclass(t *testing.T) {
	fooType := newTestClass("Foo", []*Type{ObjectType}, NewDict())
	barType := newTestClass("Bar", []*Type{fooType, IntType}, NewDict())
	cases := []struct {
		o         *Object
		classinfo *Object
		want      *Object
		wantExc   *BaseException
	}{
		{newObject(ObjectType), ObjectType.ToObject(), True.ToObject(), nil},
		{NewInt(42).ToObject(), StrType.ToObject(), False.ToObject(), nil},
		{None, NewTuple(NoneType.ToObject(), IntType.ToObject()).ToObject(), True.ToObject(), nil},
		{NewStr("foo").ToObject(), NewTuple(NoneType.ToObject(), IntType.ToObject()).ToObject(), False.ToObject(), nil},
		{NewStr("foo").ToObject(), NewTuple(IntType.ToObject(), NoneType.ToObject()).ToObject(), False.ToObject(), nil},
		{None, NewTuple().ToObject(), False.ToObject(), nil},
		{newObject(barType), fooType.ToObject(), True.ToObject(), nil},
		{newObject(barType), IntType.ToObject(), True.ToObject(), nil},
		{newObject(fooType), IntType.ToObject(), False.ToObject(), nil},
		{newObject(ObjectType), None, nil, mustCreateException(TypeErrorType, "classinfo must be a type or tuple of types")},
		{newObject(ObjectType), NewTuple(None).ToObject(), nil, mustCreateException(TypeErrorType, "classinfo must be a type or tuple of types")},
	}
	for _, cas := range cases {
		// IsInstance
		testCase := invokeTestCase{args: wrapArgs(cas.o, cas.classinfo), want: cas.want, wantExc: cas.wantExc}
		if err := runInvokeTestCase(wrapFuncForTest(IsInstance), &testCase); err != "" {
			t.Error(err)
		}
		// IsSubclass
		testCase.args = wrapArgs(cas.o.Type(), cas.classinfo)
		if err := runInvokeTestCase(wrapFuncForTest(IsSubclass), &testCase); err != "" {
			t.Error(err)
		}
	}
	// Test that IsSubclass raises when first arg is not a type.
	testCase := invokeTestCase{args: wrapArgs(None, NoneType), wantExc: mustCreateException(TypeErrorType, "issubclass() arg 1 must be a class")}
	if err := runInvokeTestCase(wrapFuncForTest(IsSubclass), &testCase); err != "" {
		t.Error(err)
	}
}

func TestIsTrue(t *testing.T) {
	badNonZeroType := newTestClass("BadNonZeroType", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__nonzero__": newBuiltinFunction("__nonzero__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return None, nil
		}).ToObject(),
	}))
	badLenType := newTestClass("BadLen", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__len__": newBuiltinFunction("__len__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return None, nil
		}).ToObject(),
	}))
	cases := []invokeTestCase{
		// Bool
		{args: wrapArgs(true), want: True.ToObject()},
		{args: wrapArgs(false), want: False.ToObject()},
		// Dict
		{args: wrapArgs(NewDict()), want: False.ToObject()},
		{args: wrapArgs(newStringDict(map[string]*Object{"foo": True.ToObject()})), want: True.ToObject()},
		// Int
		{args: wrapArgs(0), want: False.ToObject()},
		{args: wrapArgs(-1020), want: True.ToObject()},
		{args: wrapArgs(1698391283), want: True.ToObject()},
		// None
		{args: wrapArgs(None), want: False.ToObject()},
		// Object
		{args: wrapArgs(newObject(ObjectType)), want: True.ToObject()},
		// Str
		{args: wrapArgs(""), want: False.ToObject()},
		{args: wrapArgs("\x00"), want: True.ToObject()},
		{args: wrapArgs("foo"), want: True.ToObject()},
		// Tuple
		{args: wrapArgs(NewTuple()), want: False.ToObject()},
		{args: wrapArgs(newTestTuple("foo", None)), want: True.ToObject()},
		// Funky types
		{args: wrapArgs(newObject(badNonZeroType)), wantExc: mustCreateException(TypeErrorType, "__nonzero__ should return bool, returned NoneType")},
		{args: wrapArgs(newObject(badLenType)), wantExc: mustCreateException(TypeErrorType, "an integer is required")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(IsTrue), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestIter(t *testing.T) {
	fun := newBuiltinFunction("TestIter", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		if argc := len(args); argc != 1 {
			return nil, f.RaiseType(SystemErrorType, fmt.Sprintf("Iter expected 1 arg, got %d", argc))
		}
		i, raised := Iter(f, args[0])
		if raised != nil {
			return nil, raised
		}
		return Next(f, i)
	}).ToObject()
	cases := []invokeTestCase{
		{args: wrapArgs(NewTuple()), wantExc: mustCreateException(StopIterationType, "")},
		{args: wrapArgs(newTestTuple(42, "foo")), want: NewInt(42).ToObject()},
		{args: wrapArgs(newTestList("foo")), want: NewStr("foo").ToObject()},
		{args: wrapArgs("foo"), want: NewStr("f").ToObject()},
		{args: wrapArgs(123), wantExc: mustCreateException(TypeErrorType, "'int' object is not iterable")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestNext(t *testing.T) {
	fun := newBuiltinFunction("TestNext", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		if argc := len(args); argc != 1 {
			return nil, f.RaiseType(SystemErrorType, fmt.Sprintf("Next expected 1 arg, got %d", argc))
		}
		iter := args[0]
		var elems []*Object
		elem, raised := Next(f, iter)
		for ; raised == nil; elem, raised = Next(f, iter) {
			elems = append(elems, elem)
		}
		if !raised.isInstance(StopIterationType) {
			return nil, raised
		}
		f.RestoreExc(nil, nil)
		return NewTuple(elems...).ToObject(), nil
	}).ToObject()
	testElems := []*Object{NewInt(42).ToObject(), NewStr("foo").ToObject(), newObject(ObjectType)}
	cases := []invokeTestCase{
		{args: wrapArgs(mustNotRaise(Iter(newFrame(nil), NewTuple().ToObject()))), want: NewTuple().ToObject()},
		{args: wrapArgs(mustNotRaise(Iter(newFrame(nil), NewTuple(testElems...).ToObject()))), want: NewTuple(testElems...).ToObject()},
		{args: wrapArgs(mustNotRaise(Iter(newFrame(nil), NewList(testElems...).ToObject()))), want: NewTuple(testElems...).ToObject()},
		{args: wrapArgs(123), wantExc: mustCreateException(TypeErrorType, "int object is not an iterator")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestLen(t *testing.T) {
	badLenType := newTestClass("BadLen", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__len__": newBuiltinFunction("__len__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return None, nil
		}).ToObject(),
	}))
	cases := []invokeTestCase{
		{args: wrapArgs(NewDict()), want: NewInt(0).ToObject()},
		{args: wrapArgs(newStringDict(map[string]*Object{"foo": NewStr("foo value").ToObject(), "bar": NewStr("bar value").ToObject()})), want: NewInt(2).ToObject()},
		{args: wrapArgs(NewTuple()), want: NewInt(0).ToObject()},
		{args: wrapArgs(NewTuple(None, None, None)), want: NewInt(3).ToObject()},
		{args: wrapArgs(10), wantExc: mustCreateException(TypeErrorType, "object of type 'int' has no len()")},
		{args: wrapArgs(newObject(badLenType)), wantExc: mustCreateException(TypeErrorType, "an integer is required")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(Len), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestLenRaise(t *testing.T) {
	testTypes := []*Type{
		DictType,
		TupleType,
	}
	for _, typ := range testTypes {
		cases := []invokeTestCase{
			{args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, fmt.Sprintf("unbound method __len__() must be called with %s instance as first argument (got nothing instead)", typ.Name()))},
			{args: wrapArgs(newObject(ObjectType)), wantExc: mustCreateException(TypeErrorType, fmt.Sprintf("unbound method __len__() must be called with %s instance as first argument (got object instance instead)", typ.Name()))},
			{args: wrapArgs(newObject(ObjectType), newObject(ObjectType)), wantExc: mustCreateException(TypeErrorType, fmt.Sprintf("unbound method __len__() must be called with %s instance as first argument (got object instance instead)", typ.Name()))},
		}
		for _, cas := range cases {
			if err := runInvokeMethodTestCase(typ, "__len__", &cas); err != "" {
				t.Error(err)
			}
		}
	}
}

func TestInvokePositionalArgs(t *testing.T) {
	fun := newBuiltinFunction("TestInvokePositionalArgs", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		return NewTuple(args.makeCopy()...).ToObject(), nil
	}).ToObject()
	cases := []struct {
		varargs *Object
		args    Args
		want    *Object
	}{
		{nil, nil, NewTuple().ToObject()},
		{NewTuple(NewInt(2).ToObject()).ToObject(), nil, NewTuple(NewInt(2).ToObject()).ToObject()},
		{nil, []*Object{NewStr("foo").ToObject()}, NewTuple(NewStr("foo").ToObject()).ToObject()},
		{NewTuple(NewFloat(3.14).ToObject()).ToObject(), []*Object{NewStr("foo").ToObject()}, NewTuple(NewStr("foo").ToObject(), NewFloat(3.14).ToObject()).ToObject()},
		{NewList(NewFloat(3.14).ToObject()).ToObject(), []*Object{NewStr("foo").ToObject()}, NewTuple(NewStr("foo").ToObject(), NewFloat(3.14).ToObject()).ToObject()},
	}
	for _, cas := range cases {
		got, raised := Invoke(newFrame(nil), fun, cas.args, cas.varargs, nil, nil)
		switch checkResult(got, cas.want, raised, nil) {
		case checkInvokeResultExceptionMismatch:
			t.Errorf("PackArgs(%v, %v) raised %v, want nil", cas.args, cas.varargs, raised)
		case checkInvokeResultReturnValueMismatch:
			t.Errorf("PackArgs(%v, %v) = %v, want %v", cas.args, cas.varargs, got, cas.want)
		}
	}
}

func TestInvokeKeywordArgs(t *testing.T) {
	fun := newBuiltinFunction("TestInvokeKeywordArgs", func(f *Frame, _ Args, kwargs KWArgs) (*Object, *BaseException) {
		got := map[string]*Object{}
		for _, kw := range kwargs {
			got[kw.Name] = kw.Value
		}
		return newStringDict(got).ToObject(), nil
	}).ToObject()
	d := NewDict()
	d.SetItem(newFrame(nil), NewInt(123).ToObject(), None)
	cases := []struct {
		keywords KWArgs
		kwargs   *Object
		want     *Object
		wantExc  *BaseException
	}{
		{nil, nil, NewDict().ToObject(), nil},
		{wrapKWArgs("foo", 42), nil, newTestDict("foo", 42).ToObject(), nil},
		{nil, newTestDict("foo", None).ToObject(), newTestDict("foo", None).ToObject(), nil},
		{wrapKWArgs("foo", 42), newTestDict("bar", None).ToObject(), newTestDict("foo", 42, "bar", None).ToObject(), nil},
		{nil, NewList().ToObject(), nil, mustCreateException(TypeErrorType, "argument after ** must be a dict, not list")},
		{nil, d.ToObject(), nil, mustCreateException(TypeErrorType, "keywords must be strings")},
	}
	for _, cas := range cases {
		got, raised := Invoke(newFrame(nil), fun, nil, nil, cas.keywords, cas.kwargs)
		switch checkResult(got, cas.want, raised, cas.wantExc) {
		case checkInvokeResultExceptionMismatch:
			t.Errorf("PackKwargs(%v, %v) raised %v, want %v", cas.keywords, cas.kwargs, raised, cas.wantExc)
		case checkInvokeResultReturnValueMismatch:
			t.Errorf("PackKwargs(%v, %v) = %v, want %v", cas.keywords, cas.kwargs, got, cas.want)
		}
	}
}

func TestPrint(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, args *Tuple, nl bool) (string, *BaseException) {
		r, w, err := os.Pipe()
		if err != nil {
			return "", f.RaiseType(RuntimeErrorType, fmt.Sprintf("failed to open pipe: %v", err))
		}
		oldStdout := os.Stdout
		os.Stdout = w
		defer func() {
			os.Stdout = oldStdout
		}()
		done := make(chan struct{})
		var raised *BaseException
		go func() {
			defer close(done)
			defer w.Close()
			raised = Print(newFrame(nil), args.elems, nl)
		}()
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, r); err != nil {
			return "", f.RaiseType(RuntimeErrorType, fmt.Sprintf("failed to copy buffer: %v", err))
		}
		<-done
		if raised != nil {
			return "", raised
		}
		return buf.String(), nil
	})
	cases := []invokeTestCase{
		{args: wrapArgs(NewTuple(), true), want: NewStr("\n").ToObject()},
		{args: wrapArgs(NewTuple(), false), want: NewStr("").ToObject()},
		{args: wrapArgs(newTestTuple("abc", 123), true), want: NewStr("abc 123\n").ToObject()},
		{args: wrapArgs(newTestTuple("foo"), false), want: NewStr("foo ").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestReprRaise(t *testing.T) {
	testTypes := []*Type{
		BaseExceptionType,
		BoolType,
		DictType,
		IntType,
		FunctionType,
		StrType,
		TupleType,
		TypeType,
	}
	for _, typ := range testTypes {
		cases := []invokeTestCase{
			{args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, fmt.Sprintf("unbound method __repr__() must be called with %s instance as first argument (got nothing instead)", typ.Name()))},
			{args: wrapArgs(newObject(ObjectType)), wantExc: mustCreateException(TypeErrorType, fmt.Sprintf("unbound method __repr__() must be called with %s instance as first argument (got object instance instead)", typ.Name()))},
		}
		for _, cas := range cases {
			if err := runInvokeMethodTestCase(typ, "__repr__", &cas); err != "" {
				t.Error(err)
			}
		}
	}
}

func TestReprMethodReturnsNonStr(t *testing.T) {
	// Don't use runInvokeTestCase since it takes repr(args) and in this
	// case repr will raise.
	typ := newTestClass("Foo", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__repr__": newBuiltinFunction("__repr__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return None, nil
		}).ToObject(),
	}))
	_, raised := Repr(newFrame(nil), newObject(typ))
	wantExc := mustCreateException(TypeErrorType, "__repr__ returned non-string (type NoneType)")
	if !exceptionsAreEquivalent(raised, wantExc) {
		t.Errorf(`Repr() raised %v, want %v`, raised, wantExc)
	}
}

func TestResolveClass(t *testing.T) {
	f := newFrame(nil)
	cases := []struct {
		class   *Dict
		local   *Object
		globals *Dict
		name    string
		want    *Object
		wantExc *BaseException
	}{
		{newStringDict(map[string]*Object{"foo": NewStr("bar").ToObject()}), NewStr("baz").ToObject(), NewDict(), "foo", NewStr("bar").ToObject(), nil},
		{newStringDict(map[string]*Object{"str": NewInt(42).ToObject()}), nil, NewDict(), "str", NewInt(42).ToObject(), nil},
		{NewDict(), nil, newStringDict(map[string]*Object{"foo": NewStr("bar").ToObject()}), "foo", NewStr("bar").ToObject(), nil},
		{NewDict(), nil, NewDict(), "str", StrType.ToObject(), nil},
		{NewDict(), nil, NewDict(), "foo", nil, mustCreateException(NameErrorType, "name 'foo' is not defined")},
	}
	for _, cas := range cases {
		f.globals = cas.globals
		got, raised := ResolveClass(f, cas.class, cas.local, NewStr(cas.name))
		switch checkResult(got, cas.want, raised, cas.wantExc) {
		case checkInvokeResultExceptionMismatch:
			t.Errorf("ResolveClass(%v, %q) raised %v, want %v", cas.globals, cas.name, raised, cas.wantExc)
		case checkInvokeResultReturnValueMismatch:
			t.Errorf("ResolveClass(%v, %q) = %v, want %v", cas.globals, cas.name, got, cas.want)
		}
	}
}

func TestResolveGlobal(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, globals *Dict, name *Str) (*Object, *BaseException) {
		f.globals = globals
		return ResolveGlobal(f, name)
	})
	cases := []invokeTestCase{
		{args: wrapArgs(newStringDict(map[string]*Object{"foo": NewStr("bar").ToObject()}), "foo"), want: NewStr("bar").ToObject()},
		{args: wrapArgs(NewDict(), "str"), want: StrType.ToObject()},
		{args: wrapArgs(NewDict(), "foo"), wantExc: mustCreateException(NameErrorType, "name 'foo' is not defined")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestCheckLocal(t *testing.T) {
	o := newObject(ObjectType)
	cases := []invokeTestCase{
		{args: wrapArgs(o, "foo"), want: None},
		{args: wrapArgs(UnboundLocal, "bar"), wantExc: mustCreateException(UnboundLocalErrorType, "local variable 'bar' referenced before assignment")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(CheckLocal), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestSetItem(t *testing.T) {
	setItem := newBuiltinFunction("TestSetItem", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		if raised := checkFunctionArgs(f, "TestSetItem", args, ObjectType, ObjectType, ObjectType); raised != nil {
			return nil, raised
		}
		o := args[0]
		if raised := SetItem(f, o, args[1], args[2]); raised != nil {
			return nil, raised
		}
		return o, nil
	}).ToObject()
	cases := []invokeTestCase{
		{args: wrapArgs(NewDict(), "bar", None), want: newTestDict("bar", None).ToObject()},
		{args: wrapArgs(123, "bar", None), wantExc: mustCreateException(TypeErrorType, "'int' object has no attribute '__setitem__'")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(setItem, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestStartThread(t *testing.T) {
	c := make(chan bool)
	callable := newBuiltinFunction("TestStartThread", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		close(c)
		return None, nil
	}).ToObject()
	StartThread(callable)
	// Deadlock indicates the thread didn't start.
	<-c
}

func TestStartThreadRaises(t *testing.T) {
	// Since there's no way to notify that the goroutine has returned we
	// can't actually test the exception output but we can at least make
	// sure the callable ran and didn't blow up the rest of the program.
	c := make(chan bool)
	callable := newBuiltinFunction("TestStartThreadRaises", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		defer close(c)
		return nil, f.RaiseType(ExceptionType, "foo")
	}).ToObject()
	StartThread(callable)
	<-c
}

func TestTie(t *testing.T) {
	targets := make([]*Object, 3)
	cases := []struct {
		t       TieTarget
		o       *Object
		want    *Object
		wantExc *BaseException
	}{
		{TieTarget{Target: &targets[0]}, NewInt(42).ToObject(), NewTuple(NewInt(42).ToObject()).ToObject(), nil},
		{TieTarget{Target: &targets[0]}, NewTuple().ToObject(), NewTuple(NewTuple().ToObject()).ToObject(), nil},
		{
			TieTarget{
				Children: []TieTarget{{Target: &targets[0]}, {Target: &targets[1]}},
			},
			NewList(NewStr("foo").ToObject(), NewStr("bar").ToObject()).ToObject(),
			NewTuple(NewStr("foo").ToObject(), NewStr("bar").ToObject()).ToObject(),
			nil,
		},
		{
			TieTarget{
				Children: []TieTarget{
					{Target: &targets[0]},
					{Children: []TieTarget{{Target: &targets[1]}, {Target: &targets[2]}}},
				},
			},
			NewTuple(NewStr("foo").ToObject(), NewTuple(NewStr("bar").ToObject(), NewStr("baz").ToObject()).ToObject()).ToObject(),
			NewTuple(NewStr("foo").ToObject(), NewStr("bar").ToObject(), NewStr("baz").ToObject()).ToObject(),
			nil,
		},
		{
			TieTarget{
				Children: []TieTarget{
					{Target: &targets[0]},
					{Target: &targets[1]},
				},
			},
			NewList(NewStr("foo").ToObject()).ToObject(),
			nil,
			mustCreateException(TypeErrorType, "need more than 1 values to unpack"),
		},
		{
			TieTarget{Children: []TieTarget{{Target: &targets[0]}}},
			NewTuple(NewInt(1).ToObject(), NewInt(2).ToObject()).ToObject(),
			nil,
			mustCreateException(ValueErrorType, "too many values to unpack"),
		},
	}
	for _, cas := range cases {
		for i := range targets {
			targets[i] = nil
		}
		var got *Object
		raised := Tie(newFrame(nil), cas.t, cas.o)
		if raised == nil {
			var elems []*Object
			for _, t := range targets {
				if t == nil {
					break
				}
				elems = append(elems, t)
			}
			got = NewTuple(elems...).ToObject()
		}
		switch checkResult(got, cas.want, raised, cas.wantExc) {
		case checkInvokeResultExceptionMismatch:
			t.Errorf("Tie(%+v, %v) raised %v, want %v", cas.t, cas.o, raised, cas.wantExc)
		case checkInvokeResultReturnValueMismatch:
			t.Errorf("Tie(%+v, %v) = %v, want %v", cas.t, cas.o, got, cas.want)
		}
	}
}

func TestToNative(t *testing.T) {
	foo := newObject(ObjectType)
	cases := []struct {
		o       *Object
		want    interface{}
		wantExc *BaseException
	}{
		{True.ToObject(), true, nil},
		{NewInt(42).ToObject(), 42, nil},
		{NewStr("bar").ToObject(), "bar", nil},
		{foo, foo, nil},
	}
	for _, cas := range cases {
		got, raised := ToNative(newFrame(nil), cas.o)
		if !exceptionsAreEquivalent(raised, cas.wantExc) {
			t.Errorf("ToNative(%v) raised %v, want %v", cas.o, raised, cas.wantExc)
		} else if raised == nil && (!got.IsValid() || !reflect.DeepEqual(got.Interface(), cas.want)) {
			t.Errorf("ToNative(%v) = %v, want %v", cas.o, got, cas.want)
		}
	}
}

// SetAttr is tested in TestObjectSetAttr.

func exceptionsAreEquivalent(e1 *BaseException, e2 *BaseException) bool {
	if e1 == nil && e2 == nil {
		return true
	}
	if e1 == nil || e2 == nil {
		return false
	}
	if e1.typ != e2.typ {
		return false
	}
	if e1.args == nil && e2.args == nil {
		return true
	}
	if e1.args == nil || e2.args == nil {
		return false
	}
	f := newFrame(nil)
	b, raised := IsTrue(f, mustNotRaise(Eq(f, e1.args.ToObject(), e2.args.ToObject())))
	if raised != nil {
		panic(raised)
	}
	return b
}

func getFuncName(f interface{}) string {
	s := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	return regexp.MustCompile(`\w+$`).FindString(s)
}

// wrapFuncForTest creates a callable object that invokes fun, passing the
// current frame as its first argument followed by caller provided args.
func wrapFuncForTest(fun interface{}) *Object {
	return newBuiltinFunction(getFuncName(fun), func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		callable, raised := WrapNative(f, reflect.ValueOf(fun))
		if raised != nil {
			return nil, raised
		}
		argc := len(args)
		nativeArgs := make(Args, argc+1, argc+1)
		nativeArgs[0] = f.ToObject()
		copy(nativeArgs[1:], args)
		return callable.Call(f, nativeArgs, nil)
	}).ToObject()
}

func mustCreateException(t *Type, msg string) *BaseException {
	if !t.isSubclass(BaseExceptionType) {
		panic(fmt.Sprintf("type does not inherit from BaseException: %s", t.Name()))
	}
	e := toBaseExceptionUnsafe(newObject(t))
	if msg == "" {
		e.args = NewTuple()
	} else {
		e.args = NewTuple(NewStr(msg).ToObject())
	}
	return e
}

func mustNotRaise(o *Object, raised *BaseException) *Object {
	if raised != nil {
		panic(raised)
	}
	return o
}

var (
	compareAll = wrapFuncForTest(func(f *Frame, v, w *Object) (*Object, *BaseException) {
		lt, raised := LT(f, v, w)
		if raised != nil {
			return nil, raised
		}
		le, raised := LE(f, v, w)
		if raised != nil {
			return nil, raised
		}
		eq, raised := Eq(f, v, w)
		if raised != nil {
			return nil, raised
		}
		ne, raised := NE(f, v, w)
		if raised != nil {
			return nil, raised
		}
		ge, raised := GE(f, v, w)
		if raised != nil {
			return nil, raised
		}
		gt, raised := GT(f, v, w)
		if raised != nil {
			return nil, raised
		}
		return NewTuple(lt, le, eq, ne, ge, gt).ToObject(), nil
	})
	compareAllResultLT = newTestTuple(true, true, false, true, false, false).ToObject()
	compareAllResultEq = newTestTuple(false, true, true, false, true, false).ToObject()
	compareAllResultGT = newTestTuple(false, false, false, true, true, true).ToObject()
)

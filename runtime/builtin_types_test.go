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
	"math/big"
	"testing"
)

func TestBuiltinFuncs(t *testing.T) {
	f := newFrame(nil)
	objectDir := ObjectType.dict.Keys(f)
	objectDir.Sort(f)
	fooType := newTestClass("Foo", []*Type{ObjectType}, newStringDict(map[string]*Object{"bar": None}))
	fooTypeDir := NewList(objectDir.elems...)
	fooTypeDir.Append(NewStr("bar").ToObject())
	fooTypeDir.Sort(f)
	foo := newObject(fooType)
	SetAttr(f, foo, NewStr("baz"), None)
	fooDir := NewList(fooTypeDir.elems...)
	fooDir.Append(NewStr("baz").ToObject())
	fooDir.Sort(f)
	iter := mustNotRaise(Iter(f, mustNotRaise(xrangeType.Call(f, wrapArgs(5), nil))))
	neg := wrapFuncForTest(func(f *Frame, i int) int { return -i })
	raiseKey := wrapFuncForTest(func(f *Frame, o *Object) *BaseException { return f.RaiseType(RuntimeErrorType, "foo") })
	hexOctType := newTestClass("HexOct", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__hex__": newBuiltinFunction("__hex__", func(f *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return NewStr("0xhexadecimal").ToObject(), nil
		}).ToObject(),
		"__oct__": newBuiltinFunction("__hex__", func(f *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return NewStr("0octal").ToObject(), nil
		}).ToObject(),
	}))
	indexType := newTestClass("Index", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__index__": newBuiltinFunction("__index__", func(f *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return NewInt(123).ToObject(), nil
		}).ToObject(),
	}))
	cases := []struct {
		f       string
		args    Args
		kwargs  KWArgs
		want    *Object
		wantExc *BaseException
	}{
		{f: "bin", args: wrapArgs(64 + 8 + 1), want: NewStr("0b1001001").ToObject()},
		{f: "bin", args: wrapArgs(MinInt), want: NewStr(fmt.Sprintf("-0b%b0", -(MinInt >> 1))).ToObject()},
		{f: "bin", args: wrapArgs(0), want: NewStr("0b0").ToObject()},
		{f: "bin", args: wrapArgs(1), want: NewStr("0b1").ToObject()},
		{f: "bin", args: wrapArgs(-1), want: NewStr("-0b1").ToObject()},
		{f: "bin", args: wrapArgs(big.NewInt(-1)), want: NewStr("-0b1").ToObject()},
		{f: "bin", args: wrapArgs("foo"), wantExc: mustCreateException(TypeErrorType, "str object cannot be interpreted as an index")},
		{f: "bin", args: wrapArgs(0.1), wantExc: mustCreateException(TypeErrorType, "float object cannot be interpreted as an index")},
		{f: "bin", args: wrapArgs(1, 2, 3), wantExc: mustCreateException(TypeErrorType, "'bin' requires 1 arguments")},
		{f: "bin", args: wrapArgs(newObject(indexType)), want: NewStr("0b1111011").ToObject()},
		{f: "chr", args: wrapArgs(0), want: NewStr("\x00").ToObject()},
		{f: "chr", args: wrapArgs(65), want: NewStr("A").ToObject()},
		{f: "chr", args: wrapArgs(300), wantExc: mustCreateException(ValueErrorType, "chr() arg not in range(256)")},
		{f: "chr", args: wrapArgs(-1), wantExc: mustCreateException(ValueErrorType, "chr() arg not in range(256)")},
		{f: "chr", args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, "'chr' requires 1 arguments")},
		{f: "dir", args: wrapArgs(newObject(ObjectType)), want: objectDir.ToObject()},
		{f: "dir", args: wrapArgs(newObject(fooType)), want: fooTypeDir.ToObject()},
		{f: "dir", args: wrapArgs(foo), want: fooDir.ToObject()},
		{f: "dir", args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, "'dir' requires 1 arguments")},
		{f: "getattr", args: wrapArgs(None, NewStr("foo").ToObject(), NewStr("bar").ToObject()), want: NewStr("bar").ToObject()},
		{f: "getattr", args: wrapArgs(None, NewStr("foo").ToObject()), wantExc: mustCreateException(AttributeErrorType, "'NoneType' object has no attribute 'foo'")},
		{f: "hasattr", args: wrapArgs(newObject(ObjectType), NewStr("foo").ToObject()), want: False.ToObject()},
		{f: "hasattr", args: wrapArgs(foo, NewStr("bar").ToObject()), want: True.ToObject()},
		{f: "hasattr", args: wrapArgs(foo, NewStr("baz").ToObject()), want: True.ToObject()},
		{f: "hasattr", args: wrapArgs(foo, NewStr("qux").ToObject()), want: False.ToObject()},
		{f: "hash", args: wrapArgs(123), want: NewInt(123).ToObject()},
		{f: "hash", args: wrapArgs("foo"), want: hashFoo},
		{f: "hash", args: wrapArgs(NewList()), wantExc: mustCreateException(TypeErrorType, "unhashable type: 'list'")},
		{f: "hex", args: wrapArgs(0x63adbeef), want: NewStr("0x63adbeef").ToObject()},
		{f: "hex", args: wrapArgs(0), want: NewStr("0x0").ToObject()},
		{f: "hex", args: wrapArgs(1), want: NewStr("0x1").ToObject()},
		{f: "hex", args: wrapArgs(-1), want: NewStr("-0x1").ToObject()},
		{f: "hex", args: wrapArgs(big.NewInt(-1)), want: NewStr("-0x1L").ToObject()},
		{f: "hex", args: wrapArgs("foo"), wantExc: mustCreateException(TypeErrorType, "hex() argument can't be converted to hex")},
		{f: "hex", args: wrapArgs(0.1), wantExc: mustCreateException(TypeErrorType, "hex() argument can't be converted to hex")},
		{f: "hex", args: wrapArgs(1, 2, 3), wantExc: mustCreateException(TypeErrorType, "'hex' requires 1 arguments")},
		{f: "hex", args: wrapArgs(newObject(hexOctType)), want: NewStr("0xhexadecimal").ToObject()},
		{f: "id", args: wrapArgs(foo), want: NewInt(int(uintptr(foo.toPointer()))).ToObject()},
		{f: "id", args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, "'id' requires 1 arguments")},
		{f: "isinstance", args: wrapArgs(NewInt(42).ToObject(), IntType.ToObject()), want: True.ToObject()},
		{f: "isinstance", args: wrapArgs(NewStr("foo").ToObject(), TupleType.ToObject()), want: False.ToObject()},
		{f: "isinstance", args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, "'isinstance' requires 2 arguments")},
		{f: "issubclass", args: wrapArgs(IntType, IntType), want: True.ToObject()},
		{f: "issubclass", args: wrapArgs(fooType, IntType), want: False.ToObject()},
		{f: "issubclass", args: wrapArgs(fooType, ObjectType), want: True.ToObject()},
		{f: "issubclass", args: wrapArgs(FloatType, newTestTuple(IntType, StrType)), want: False.ToObject()},
		{f: "issubclass", args: wrapArgs(FloatType, newTestTuple(IntType, FloatType)), want: True.ToObject()},
		{f: "issubclass", args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, "'issubclass' requires 2 arguments")},
		{f: "iter", args: wrapArgs(iter), want: iter},
		{f: "iter", args: wrapArgs(42), wantExc: mustCreateException(TypeErrorType, "'int' object is not iterable")},
		{f: "len", args: wrapArgs(newTestList(1, 2, 3)), want: NewInt(3).ToObject()},
		{f: "len", args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, "'len' requires 1 arguments")},
		{f: "max", args: wrapArgs(2, 3, 1), want: NewInt(3).ToObject()},
		{f: "max", args: wrapArgs("bar", "foo"), want: NewStr("foo").ToObject()},
		{f: "max", args: wrapArgs(newTestList(2, 3, 1)), want: NewInt(3).ToObject()},
		{f: "max", args: wrapArgs(newTestList("bar", "foo")), want: NewStr("foo").ToObject()},
		{f: "max", args: wrapArgs(2, 3, 1), want: NewInt(3).ToObject()},
		{f: "max", args: wrapArgs("bar", "foo"), want: NewStr("foo").ToObject()},
		{f: "max", args: wrapArgs(newTestList(2, 3, 1)), want: NewInt(3).ToObject()},
		{f: "max", args: wrapArgs(newTestList("bar", "foo")), want: NewStr("foo").ToObject()},
		{f: "max", args: wrapArgs(2, 3, 1), kwargs: wrapKWArgs("key", neg), want: NewInt(1).ToObject()},
		{f: "max", args: wrapArgs(1, 2, 3), kwargs: wrapKWArgs("key", neg), want: NewInt(1).ToObject()},
		{f: "max", args: wrapArgs(newTestList(2, 3, 1)), kwargs: wrapKWArgs("key", neg), want: NewInt(1).ToObject()},
		{f: "max", args: wrapArgs(newTestList(1, 2, 3)), kwargs: wrapKWArgs("key", neg), want: NewInt(1).ToObject()},
		{f: "max", args: wrapArgs(2, 3, 1), kwargs: wrapKWArgs("key", neg), want: NewInt(1).ToObject()},
		{f: "max", args: wrapArgs(1, 2, 3), kwargs: wrapKWArgs("key", neg), want: NewInt(1).ToObject()},
		{f: "max", args: wrapArgs(newTestList(2, 3, 1)), kwargs: wrapKWArgs("key", neg), want: NewInt(1).ToObject()},
		{f: "max", args: wrapArgs(newTestList(1, 2, 3)), kwargs: wrapKWArgs("key", neg), want: NewInt(1).ToObject()},
		{f: "max", args: wrapArgs(newTestList("foo")), want: NewStr("foo").ToObject()},
		{f: "max", args: wrapArgs(1), wantExc: mustCreateException(TypeErrorType, "'int' object is not iterable")},
		{f: "max", args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, "'max' requires 1 arguments")},
		{f: "max", args: wrapArgs(newTestList()), wantExc: mustCreateException(ValueErrorType, "max() arg is an empty sequence")},
		{f: "max", args: wrapArgs(1, 2), kwargs: wrapKWArgs("key", raiseKey), wantExc: mustCreateException(RuntimeErrorType, "foo")},
		{f: "min", args: wrapArgs(2, 3, 1), want: NewInt(1).ToObject()},
		{f: "min", args: wrapArgs("bar", "foo"), want: NewStr("bar").ToObject()},
		{f: "min", args: wrapArgs(newTestList(2, 3, 1)), want: NewInt(1).ToObject()},
		{f: "min", args: wrapArgs(newTestList("bar", "foo")), want: NewStr("bar").ToObject()},
		{f: "min", args: wrapArgs(2, 3, 1), want: NewInt(1).ToObject()},
		{f: "min", args: wrapArgs("bar", "foo"), want: NewStr("bar").ToObject()},
		{f: "min", args: wrapArgs(newTestList(2, 3, 1)), want: NewInt(1).ToObject()},
		{f: "min", args: wrapArgs(newTestList("bar", "foo")), want: NewStr("bar").ToObject()},
		{f: "min", args: wrapArgs(2, 3, 1), kwargs: wrapKWArgs("key", neg), want: NewInt(3).ToObject()},
		{f: "min", args: wrapArgs(1, 2, 3), kwargs: wrapKWArgs("key", neg), want: NewInt(3).ToObject()},
		{f: "min", args: wrapArgs(newTestList(2, 3, 1)), kwargs: wrapKWArgs("key", neg), want: NewInt(3).ToObject()},
		{f: "min", args: wrapArgs(newTestList(1, 2, 3)), kwargs: wrapKWArgs("key", neg), want: NewInt(3).ToObject()},
		{f: "min", args: wrapArgs(2, 3, 1), kwargs: wrapKWArgs("key", neg), want: NewInt(3).ToObject()},
		{f: "min", args: wrapArgs(1, 2, 3), kwargs: wrapKWArgs("key", neg), want: NewInt(3).ToObject()},
		{f: "min", args: wrapArgs(newTestList(2, 3, 1)), kwargs: wrapKWArgs("key", neg), want: NewInt(3).ToObject()},
		{f: "min", args: wrapArgs(newTestList(1, 2, 3)), kwargs: wrapKWArgs("key", neg), want: NewInt(3).ToObject()},
		{f: "min", args: wrapArgs(newTestList("foo")), want: NewStr("foo").ToObject()},
		{f: "min", args: wrapArgs(1), wantExc: mustCreateException(TypeErrorType, "'int' object is not iterable")},
		{f: "min", args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, "'min' requires 1 arguments")},
		{f: "min", args: wrapArgs(newTestList()), wantExc: mustCreateException(ValueErrorType, "min() arg is an empty sequence")},
		{f: "min", args: wrapArgs(1, 2), kwargs: wrapKWArgs("key", raiseKey), wantExc: mustCreateException(RuntimeErrorType, "foo")},
		{f: "oct", args: wrapArgs(077), want: NewStr("077").ToObject()},
		{f: "oct", args: wrapArgs(0), want: NewStr("0").ToObject()},
		{f: "oct", args: wrapArgs(1), want: NewStr("01").ToObject()},
		{f: "oct", args: wrapArgs(-1), want: NewStr("-01").ToObject()},
		{f: "oct", args: wrapArgs(big.NewInt(-1)), want: NewStr("-01L").ToObject()},
		{f: "oct", args: wrapArgs("foo"), wantExc: mustCreateException(TypeErrorType, "oct() argument can't be converted to oct")},
		{f: "oct", args: wrapArgs(0.1), wantExc: mustCreateException(TypeErrorType, "oct() argument can't be converted to oct")},
		{f: "oct", args: wrapArgs(1, 2, 3), wantExc: mustCreateException(TypeErrorType, "'oct' requires 1 arguments")},
		{f: "oct", args: wrapArgs(newObject(hexOctType)), want: NewStr("0octal").ToObject()},
		{f: "ord", args: wrapArgs("a"), want: NewInt(97).ToObject()},
		{f: "ord", args: wrapArgs(NewUnicode("樂")), want: NewInt(63764).ToObject()},
		{f: "ord", args: wrapArgs("foo"), wantExc: mustCreateException(ValueErrorType, "ord() expected a character, but string of length 3 found")},
		{f: "ord", args: wrapArgs(NewUnicode("волн")), wantExc: mustCreateException(ValueErrorType, "ord() expected a character, but string of length 4 found")},
		{f: "ord", args: wrapArgs(1, 2, 3), wantExc: mustCreateException(TypeErrorType, "'ord' requires 1 arguments")},
		{f: "range", args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, "'__new__' of 'int' requires 3 arguments")},
		{f: "range", args: wrapArgs(3), want: newTestList(0, 1, 2).ToObject()},
		{f: "range", args: wrapArgs(10, 0), want: NewList().ToObject()},
		{f: "range", args: wrapArgs(-12, -23, -5), want: newTestList(-12, -17, -22).ToObject()},
		{f: "repr", args: wrapArgs(123), want: NewStr("123").ToObject()},
		{f: "repr", args: wrapArgs(NewUnicode("abc")), want: NewStr("u'abc'").ToObject()},
		{f: "repr", args: wrapArgs(newTestTuple("foo", "bar")), want: NewStr("('foo', 'bar')").ToObject()},
		{f: "repr", args: wrapArgs("a", "b", "c"), wantExc: mustCreateException(TypeErrorType, "'repr' requires 1 arguments")},
		{f: "unichr", args: wrapArgs(0), want: NewUnicode("\x00").ToObject()},
		{f: "unichr", args: wrapArgs(65), want: NewStr("A").ToObject()},
		{f: "unichr", args: wrapArgs(0x120000), wantExc: mustCreateException(ValueErrorType, "unichr() arg not in range(0x10ffff)")},
		{f: "unichr", args: wrapArgs(-1), wantExc: mustCreateException(ValueErrorType, "unichr() arg not in range(0x10ffff)")},
		{f: "unichr", args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, "'unichr' requires 1 arguments")},
	}
	for _, cas := range cases {
		fun := mustNotRaise(Builtins.GetItemString(newFrame(nil), cas.f))
		if fun == nil {
			t.Fatalf("%s not found in builtins: %v", cas.f, Builtins)
		}
		testCase := invokeTestCase{args: cas.args, kwargs: cas.kwargs, want: cas.want, wantExc: cas.wantExc}
		if err := runInvokeTestCase(fun, &testCase); err != "" {
			t.Error(err)
		}
	}
}

func TestBuiltinGlobals(t *testing.T) {
	f := newFrame(nil)
	f.globals = newTestDict("foo", 1, "bar", 2, 42, None)
	globals := mustNotRaise(Builtins.GetItemString(f, "globals"))
	got, raised := globals.Call(f, nil, nil)
	want := newTestDict("foo", 1, "bar", 2, 42, None).ToObject()
	switch checkResult(got, want, raised, nil) {
	case checkInvokeResultExceptionMismatch:
		t.Errorf("globals() = %v, want %v", got, want)
	case checkInvokeResultReturnValueMismatch:
		t.Errorf("globals() raised %v, want nil", raised)
	}
}

func TestNoneRepr(t *testing.T) {
	cas := invokeTestCase{args: wrapArgs(None), want: NewStr("None").ToObject()}
	if err := runInvokeMethodTestCase(NoneType, "__repr__", &cas); err != "" {
		t.Error(err)
	}
}

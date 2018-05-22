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
	"testing"
)

func TestBuiltinDelAttr(t *testing.T) {
	f := NewRootFrame()
	delattr := mustNotRaise(Builtins.GetItemString(f, "delattr"))
	fooType := newTestClass("Foo", []*Type{ObjectType}, NewDict())
	fooForDelAttr := newObject(fooType)
	fooValue := newObject(ObjectType)
	mustNotRaise(nil, SetAttr(f, fooForDelAttr, NewStr("bar"), fooValue))
	fun := wrapFuncForTest(func(f *Frame, args ...*Object) (*Object, *BaseException) {
		result, raised := delattr.Call(f, args, nil)
		if raised != nil {
			return nil, raised
		}
		val, raised := GetAttr(f, args[0], toStrUnsafe(args[1]), None)
		return newTestTuple(result, val == fooValue).ToObject(), nil
	})
	cases := []invokeTestCase{
		{args: wrapArgs(fooForDelAttr, "bar"), want: newTestTuple(None, False.ToObject()).ToObject()},
		{args: wrapArgs(fooForDelAttr, "baz"), wantExc: mustCreateException(AttributeErrorType, "'Foo' object has no attribute 'baz'")},
		{args: wrapArgs(fooForDelAttr), wantExc: mustCreateException(TypeErrorType, "'delattr' requires 2 arguments")},
		{args: wrapArgs(fooForDelAttr, "foo", "bar"), wantExc: mustCreateException(TypeErrorType, "'delattr' requires 2 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestBuiltinFuncs(t *testing.T) {
	f := NewRootFrame()
	objectDir := ObjectType.Dict().Keys(f)
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
	dirModule := newTestModule("foo", "foo.py")
	if raised := dirModule.Dict().SetItemString(NewRootFrame(), "bar", newObject(ObjectType)); raised != nil {
		panic(raised)
	}
	dirModuleDir := dirModule.Dict().Keys(NewRootFrame())
	if raised := dirModuleDir.Sort(NewRootFrame()); raised != nil {
		panic(raised)
	}
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
	badNonZeroType := newTestClass("BadNonZeroType", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__nonzero__": newBuiltinFunction("__nonzero__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return nil, f.RaiseType(RuntimeErrorType, "foo")
		}).ToObject(),
	}))
	badNextType := newTestClass("BadNextType", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"next": newBuiltinFunction("next", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return nil, f.RaiseType(RuntimeErrorType, "foo")
		}).ToObject(),
	}))
	badIterType := newTestClass("BadIterType", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__iter__": newBuiltinFunction("__iter__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return newObject(badNextType), nil
		}).ToObject(),
	}))
	addType := newTestClass("Add", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__add__": newBuiltinFunction("__add__", func(f *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return NewInt(1).ToObject(), nil
		}).ToObject(),
	}))
	fooBuiltinFunc := newBuiltinFunction("foo", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		return newTestTuple(NewTuple(args.makeCopy()...), kwargs.makeDict()).ToObject(), nil
	}).ToObject()
	fooFunc := NewFunction(NewCode("foo", "foo.py", nil, CodeFlagVarArg, func(f *Frame, args []*Object) (*Object, *BaseException) {
		return args[0], nil
	}), nil)
	cases := []struct {
		f       string
		args    Args
		kwargs  KWArgs
		want    *Object
		wantExc *BaseException
	}{
		{f: "abs", args: wrapArgs(1, 2, 3), wantExc: mustCreateException(TypeErrorType, "'abs' requires 1 arguments")},
		{f: "abs", args: wrapArgs(1), want: NewInt(1).ToObject()},
		{f: "abs", args: wrapArgs(-1), want: NewInt(1).ToObject()},
		{f: "abs", args: wrapArgs(big.NewInt(2)), want: NewLong(big.NewInt(2)).ToObject()},
		{f: "abs", args: wrapArgs(big.NewInt(-2)), want: NewLong(big.NewInt(2)).ToObject()},
		{f: "abs", args: wrapArgs(NewFloat(3.4)), want: NewFloat(3.4).ToObject()},
		{f: "abs", args: wrapArgs(NewFloat(-3.4)), want: NewFloat(3.4).ToObject()},
		{f: "abs", args: wrapArgs(MinInt), want: NewLong(big.NewInt(MinInt).Neg(minIntBig)).ToObject()},
		{f: "abs", args: wrapArgs(NewStr("a")), wantExc: mustCreateException(TypeErrorType, "bad operand type for abs(): 'str'")},
		{f: "all", args: wrapArgs(newTestList()), want: True.ToObject()},
		{f: "all", args: wrapArgs(newTestList(1, 2, 3)), want: True.ToObject()},
		{f: "all", args: wrapArgs(newTestList(1, 0, 1)), want: False.ToObject()},
		{f: "all", args: wrapArgs(13), wantExc: mustCreateException(TypeErrorType, "'int' object is not iterable")},
		{f: "all", args: wrapArgs(newTestList(newObject(badNonZeroType))), wantExc: mustCreateException(RuntimeErrorType, "foo")},
		{f: "all", args: wrapArgs(newObject(badIterType)), wantExc: mustCreateException(RuntimeErrorType, "foo")},
		{f: "any", args: wrapArgs(newTestList()), want: False.ToObject()},
		{f: "any", args: wrapArgs(newTestList(1, 2, 3)), want: True.ToObject()},
		{f: "any", args: wrapArgs(newTestList(1, 0, 1)), want: True.ToObject()},
		{f: "any", args: wrapArgs(newTestList(0, 0, 0)), want: False.ToObject()},
		{f: "any", args: wrapArgs(newTestList(False.ToObject(), False.ToObject())), want: False.ToObject()},
		{f: "any", args: wrapArgs(13), wantExc: mustCreateException(TypeErrorType, "'int' object is not iterable")},
		{f: "any", args: wrapArgs(newTestList(newObject(badNonZeroType))), wantExc: mustCreateException(RuntimeErrorType, "foo")},
		{f: "any", args: wrapArgs(newObject(badIterType)), wantExc: mustCreateException(RuntimeErrorType, "foo")},
		{f: "bin", args: wrapArgs(64 + 8 + 1), want: NewStr("0b1001001").ToObject()},
		{f: "bin", args: wrapArgs(MinInt), want: NewStr(fmt.Sprintf("-0b%b0", -(MinInt >> 1))).ToObject()},
		{f: "bin", args: wrapArgs(0), want: NewStr("0b0").ToObject()},
		{f: "bin", args: wrapArgs(1), want: NewStr("0b1").ToObject()},
		{f: "bin", args: wrapArgs(-1), want: NewStr("-0b1").ToObject()},
		{f: "bin", args: wrapArgs(big.NewInt(-1)), want: NewStr("-0b1").ToObject()},
		{f: "bin", args: wrapArgs("foo"), wantExc: mustCreateException(TypeErrorType, "str object cannot be interpreted as an index")},
		{f: "bin", args: wrapArgs(0.1), wantExc: mustCreateException(TypeErrorType, "float object cannot be interpreted as an index")},
		{f: "bin", args: wrapArgs(1, 2, 3), wantExc: mustCreateException(TypeErrorType, "'bin' requires 1 arguments")},
		{f: "bin", args: wrapArgs(newTestIndexObject(123)), want: NewStr("0b1111011").ToObject()},
		{f: "callable", args: wrapArgs(fooBuiltinFunc), want: True.ToObject()},
		{f: "callable", args: wrapArgs(fooFunc), want: True.ToObject()},
		{f: "callable", args: wrapArgs(0), want: False.ToObject()},
		{f: "callable", args: wrapArgs(0.1), want: False.ToObject()},
		{f: "callable", args: wrapArgs("foo"), want: False.ToObject()},
		{f: "callable", args: wrapArgs(newTestDict("foo", 1, "bar", 2)), want: False.ToObject()},
		{f: "callable", args: wrapArgs(newTestList(1, 2, 3)), want: False.ToObject()},
		{f: "callable", args: wrapArgs(iter), want: False.ToObject()},
		{f: "callable", args: wrapArgs(1, 2), wantExc: mustCreateException(TypeErrorType, "'callable' requires 1 arguments")},
		{f: "chr", args: wrapArgs(0), want: NewStr("\x00").ToObject()},
		{f: "chr", args: wrapArgs(65), want: NewStr("A").ToObject()},
		{f: "chr", args: wrapArgs(300), wantExc: mustCreateException(ValueErrorType, "chr() arg not in range(256)")},
		{f: "chr", args: wrapArgs(-1), wantExc: mustCreateException(ValueErrorType, "chr() arg not in range(256)")},
		{f: "chr", args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, "'chr' requires 1 arguments")},
		{f: "dir", args: wrapArgs(newObject(ObjectType)), want: objectDir.ToObject()},
		{f: "dir", args: wrapArgs(newObject(fooType)), want: fooTypeDir.ToObject()},
		{f: "dir", args: wrapArgs(fooType), want: fooTypeDir.ToObject()},
		{f: "dir", args: wrapArgs(foo), want: fooDir.ToObject()},
		{f: "dir", args: wrapArgs(dirModule), want: dirModuleDir.ToObject()},
		{f: "dir", args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, "'dir' requires 1 arguments")},
		{f: "divmod", args: wrapArgs(12, 7), want: NewTuple2(NewInt(1).ToObject(), NewInt(5).ToObject()).ToObject()},
		{f: "divmod", args: wrapArgs(-12, 7), want: NewTuple2(NewInt(-2).ToObject(), NewInt(2).ToObject()).ToObject()},
		{f: "divmod", args: wrapArgs(12, -7), want: NewTuple2(NewInt(-2).ToObject(), NewInt(-2).ToObject()).ToObject()},
		{f: "divmod", args: wrapArgs(-12, -7), want: NewTuple2(NewInt(1).ToObject(), NewInt(-5).ToObject()).ToObject()},
		{f: "divmod", args: wrapArgs(MaxInt, MinInt), want: NewTuple2(NewInt(-1).ToObject(), NewInt(-1).ToObject()).ToObject()},
		{f: "divmod", args: wrapArgs(MinInt, MaxInt), want: NewTuple2(NewInt(-2).ToObject(), NewInt(MaxInt-1).ToObject()).ToObject()},
		{f: "divmod", args: wrapArgs(MinInt, -1), want: NewTuple2(NewLong(new(big.Int).Neg(minIntBig)).ToObject(), NewLong(big.NewInt(0)).ToObject()).ToObject()},
		{f: "divmod", args: wrapArgs(big.NewInt(12), big.NewInt(7)), want: NewTuple2(NewLong(big.NewInt(1)).ToObject(), NewLong(big.NewInt(5)).ToObject()).ToObject()},
		{f: "divmod", args: wrapArgs(big.NewInt(-12), big.NewInt(7)), want: NewTuple2(NewLong(big.NewInt(-2)).ToObject(), NewLong(big.NewInt(2)).ToObject()).ToObject()},
		{f: "divmod", args: wrapArgs(big.NewInt(12), big.NewInt(-7)), want: NewTuple2(NewLong(big.NewInt(-2)).ToObject(), NewLong(big.NewInt(-2)).ToObject()).ToObject()},
		{f: "divmod", args: wrapArgs(big.NewInt(-12), big.NewInt(-7)), want: NewTuple2(NewLong(big.NewInt(1)).ToObject(), NewLong(big.NewInt(-5)).ToObject()).ToObject()},
		{f: "divmod", args: wrapArgs(3.25, 1.0), want: NewTuple2(NewFloat(3.0).ToObject(), NewFloat(0.25).ToObject()).ToObject()},
		{f: "divmod", args: wrapArgs(-3.25, 1.0), want: NewTuple2(NewFloat(-4.0).ToObject(), NewFloat(0.75).ToObject()).ToObject()},
		{f: "divmod", args: wrapArgs(3.25, -1.0), want: NewTuple2(NewFloat(-4.0).ToObject(), NewFloat(-0.75).ToObject()).ToObject()},
		{f: "divmod", args: wrapArgs(-3.25, -1.0), want: NewTuple2(NewFloat(3.0).ToObject(), NewFloat(-0.25).ToObject()).ToObject()},
		{f: "divmod", args: wrapArgs(NewStr("a"), NewStr("b")), wantExc: mustCreateException(TypeErrorType, "unsupported operand type(s) for divmod(): 'str' and 'str'")},
		{f: "divmod", args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, "'divmod' requires 2 arguments")},
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
		{f: "map", args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, "map() requires at least two args")},
		{f: "map", args: wrapArgs(StrType), wantExc: mustCreateException(TypeErrorType, "map() requires at least two args")},
		{f: "map", args: wrapArgs(None, newTestList()), want: newTestList().ToObject()},
		{f: "map", args: wrapArgs(None, newTestList(1, 2, 3)), want: newTestList(1, 2, 3).ToObject()},
		{f: "map", args: wrapArgs(None, newTestDict("foo", 1, "bar", 3)), want: newTestList("foo", "bar").ToObject()},
		{f: "map", args: wrapArgs(None, None), wantExc: mustCreateException(TypeErrorType, "'NoneType' object is not iterable")},
		{f: "map", args: wrapArgs(StrType, None), wantExc: mustCreateException(TypeErrorType, "'NoneType' object is not iterable")},
		{f: "map", args: wrapArgs(StrType, newTestList(), None), wantExc: mustCreateException(TypeErrorType, "'NoneType' object is not iterable")},
		{f: "map", args: wrapArgs(newTestList(), newTestList(1, 2, 3)), wantExc: mustCreateException(TypeErrorType, "'list' object is not callable")},
		{f: "map", args: wrapArgs(StrType, newTestList()), want: newTestList().ToObject()},
		{f: "map", args: wrapArgs(StrType, newTestList(1, 2, 3)), want: newTestList("1", "2", "3").ToObject()},
		{f: "map", args: wrapArgs(StrType, newTestList(-1, -2, -3)), want: newTestList("-1", "-2", "-3").ToObject()},
		{f: "map", args: wrapArgs(IntType, newTestList("1", "2", "3")), want: newTestList(1, 2, 3).ToObject()},
		{f: "map", args: wrapArgs(IntType, newTestList("-1", "-2", "-3")), want: newTestList(-1, -2, -3).ToObject()},
		{f: "map", args: wrapArgs(IntType, "123"), want: newTestList(1, 2, 3).ToObject()},
		{f: "map", args: wrapArgs(IntType, newTestDict("1", "11", "2", "22")), want: newTestList(1, 2).ToObject()},
		{f: "map", args: wrapArgs(IntType, 1), wantExc: mustCreateException(TypeErrorType, "'int' object is not iterable")},
		{f: "map", args: wrapArgs(1, newTestList(1, 2, 3)), wantExc: mustCreateException(TypeErrorType, "'int' object is not callable")},
		{f: "map", args: wrapArgs(StrType, newTestList(), 1), wantExc: mustCreateException(TypeErrorType, "'int' object is not iterable")},
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
		{f: "round", args: wrapArgs(1234.567), want: NewFloat(1235).ToObject()},
		{f: "round", args: wrapArgs(1234.111), want: NewFloat(1234).ToObject()},
		{f: "round", args: wrapArgs(-1234.567), want: NewFloat(-1235).ToObject()},
		{f: "round", args: wrapArgs(-1234.111), want: NewFloat(-1234).ToObject()},
		{f: "round", args: wrapArgs(1234.567, newTestIndexObject(0)), want: NewFloat(1235).ToObject()},
		{f: "round", args: wrapArgs("foo"), wantExc: mustCreateException(TypeErrorType, "a float is required")},
		{f: "round", args: wrapArgs(12.5, 0), want: NewFloat(13.0).ToObject()},
		{f: "round", args: wrapArgs(-12.5, 0), want: NewFloat(-13.0).ToObject()},
		{f: "round", args: wrapArgs(12.5, 3), want: NewFloat(12.5).ToObject()},
		{f: "round", args: wrapArgs(1234.5, 1), want: NewFloat(1234.5).ToObject()},
		{f: "round", args: wrapArgs(1234.5, 1), want: NewFloat(1234.5).ToObject()},
		{f: "round", args: wrapArgs(1234.56, 1), want: NewFloat(1234.6).ToObject()},
		{f: "round", args: wrapArgs(-1234.56, 1), want: NewFloat(-1234.6).ToObject()},
		{f: "round", args: wrapArgs(-1234.56, -2), want: NewFloat(-1200.0).ToObject()},
		{f: "round", args: wrapArgs(-1234.56, -8), want: NewFloat(0.0).ToObject()},
		{f: "round", args: wrapArgs(63.4, -3), want: NewFloat(0.0).ToObject()},
		{f: "round", args: wrapArgs(63.4, -2), want: NewFloat(100.0).ToObject()},
		{f: "sorted", args: wrapArgs(NewList()), want: NewList().ToObject()},
		{f: "sorted", args: wrapArgs(newTestList("foo", "bar")), want: newTestList("bar", "foo").ToObject()},
		{f: "sorted", args: wrapArgs(newTestList(true, false)), want: newTestList(false, true).ToObject()},
		{f: "sorted", args: wrapArgs(newTestList(1, 2, 0, 3)), want: newTestRange(4).ToObject()},
		{f: "sorted", args: wrapArgs(newTestRange(100)), want: newTestRange(100).ToObject()},
		{f: "sorted", args: wrapArgs(newTestTuple(1, 2, 0, 3)), want: newTestRange(4).ToObject()},
		{f: "sorted", args: wrapArgs(newTestDict("foo", 1, "bar", 2)), want: newTestList("bar", "foo").ToObject()},
		{f: "sorted", args: wrapArgs(1), wantExc: mustCreateException(TypeErrorType, "'int' object is not iterable")},
		{f: "sorted", args: wrapArgs(newTestList("foo", "bar"), 2), wantExc: mustCreateException(TypeErrorType, "'sorted' requires 1 arguments")},
		{f: "sum", args: wrapArgs(newTestList(1, 2, 3, 4)), want: NewInt(10).ToObject()},
		{f: "sum", args: wrapArgs(newTestList(1, 2), 3), want: NewFloat(6).ToObject()},
		{f: "sum", args: wrapArgs(newTestList(2, 1.1)), want: NewFloat(3.1).ToObject()},
		{f: "sum", args: wrapArgs(newTestList(2, 1.1, 2)), want: NewFloat(5.1).ToObject()},
		{f: "sum", args: wrapArgs(newTestList(2, 1.1, 2.0)), want: NewFloat(5.1).ToObject()},
		{f: "sum", args: wrapArgs(newTestList(1), newObject(addType)), want: NewInt(1).ToObject()},
		{f: "sum", args: wrapArgs(newTestList(newObject(addType)), newObject(addType)), want: NewInt(1).ToObject()},
		{f: "unichr", args: wrapArgs(0), want: NewUnicode("\x00").ToObject()},
		{f: "unichr", args: wrapArgs(65), want: NewStr("A").ToObject()},
		{f: "unichr", args: wrapArgs(0x120000), wantExc: mustCreateException(ValueErrorType, "unichr() arg not in range(0x10ffff)")},
		{f: "unichr", args: wrapArgs(-1), wantExc: mustCreateException(ValueErrorType, "unichr() arg not in range(0x10ffff)")},
		{f: "unichr", args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, "'unichr' requires 1 arguments")},
		{f: "zip", args: wrapArgs(), want: newTestList().ToObject()},
		{f: "zip", args: wrapArgs(newTestTuple()), want: newTestList().ToObject()},
		{f: "zip", args: wrapArgs(newTestList()), want: newTestList().ToObject()},
		{f: "zip", args: wrapArgs(newTestList(1)), want: newTestList(newTestTuple(1).ToObject()).ToObject()},
		{f: "zip", args: wrapArgs(newTestList(1, 2, 3)), want: newTestList(newTestTuple(1).ToObject(), newTestTuple(2).ToObject(), newTestTuple(3).ToObject()).ToObject()},
		{f: "zip", args: wrapArgs(newTestRange(3)), want: newTestList(newTestTuple(0).ToObject(), newTestTuple(1).ToObject(), newTestTuple(2).ToObject()).ToObject()},
		{f: "zip", args: wrapArgs(newTestTuple(1, 2, 3), newTestTuple(4, 5, 6)), want: NewList(newTestTuple(1, 4).ToObject(), newTestTuple(2, 5).ToObject(), newTestTuple(3, 6).ToObject()).ToObject()},
		{f: "zip", args: wrapArgs(newTestTuple(1, 2, 3), newTestTuple(4, 5)), want: NewList(newTestTuple(1, 4).ToObject(), newTestTuple(2, 5).ToObject()).ToObject()},
		{f: "zip", args: wrapArgs(newTestTuple(1, 2), newTestTuple(4, 5, 5)), want: NewList(newTestTuple(1, 4).ToObject(), newTestTuple(2, 5).ToObject()).ToObject()},
		{f: "zip", args: wrapArgs(1), wantExc: mustCreateException(TypeErrorType, "'int' object is not iterable")},
		{f: "zip", args: wrapArgs(newTestDict("foo", 1, "bar", 2)), want: newTestList(newTestTuple("foo").ToObject(), newTestTuple("bar").ToObject()).ToObject()},
	}
	for _, cas := range cases {
		fun := mustNotRaise(Builtins.GetItemString(NewRootFrame(), cas.f))
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
	f := NewRootFrame()
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

func TestEllipsisRepr(t *testing.T) {
	cas := invokeTestCase{args: wrapArgs(Ellipsis), want: NewStr("Ellipsis").ToObject()}
	if err := runInvokeMethodTestCase(EllipsisType, "__repr__", &cas); err != "" {
		t.Error(err)
	}
}

func TestNoneRepr(t *testing.T) {
	cas := invokeTestCase{args: wrapArgs(None), want: NewStr("None").ToObject()}
	if err := runInvokeMethodTestCase(NoneType, "__repr__", &cas); err != "" {
		t.Error(err)
	}
}

func TestNotImplementedRepr(t *testing.T) {
	cas := invokeTestCase{args: wrapArgs(NotImplemented), want: NewStr("NotImplemented").ToObject()}
	if err := runInvokeMethodTestCase(NotImplementedType, "__repr__", &cas); err != "" {
		t.Error(err)
	}
}

// captureStdout invokes a function closure which writes to stdout and captures
// its output as string.
func captureStdout(f *Frame, fn func() *BaseException) (string, *BaseException) {
	r, w, err := os.Pipe()
	if err != nil {
		return "", f.RaiseType(RuntimeErrorType, fmt.Sprintf("failed to open pipe: %v", err))
	}
	oldStdout := Stdout
	Stdout = NewFileFromFD(w.Fd(), nil)
	defer func() {
		Stdout = oldStdout
	}()
	done := make(chan struct{})
	var raised *BaseException
	go func() {
		defer close(done)
		defer w.Close()
		raised = fn()
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
}

// TODO(corona10): Re-enable once #282 is addressed.
/*func TestBuiltinPrint(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, args *Tuple, kwargs KWArgs) (string, *BaseException) {
		return captureStdout(f, func() *BaseException {
			_, raised := builtinPrint(NewRootFrame(), args.elems, kwargs)
			return raised
		})
	})
	cases := []invokeTestCase{
		{args: wrapArgs(NewTuple(), wrapKWArgs()), want: NewStr("\n").ToObject()},
		{args: wrapArgs(newTestTuple("abc"), wrapKWArgs()), want: NewStr("abc\n").ToObject()},
		{args: wrapArgs(newTestTuple("abc", 123), wrapKWArgs()), want: NewStr("abc 123\n").ToObject()},
		{args: wrapArgs(newTestTuple("abc", 123), wrapKWArgs("sep", "")), want: NewStr("abc123\n").ToObject()},
		{args: wrapArgs(newTestTuple("abc", 123), wrapKWArgs("end", "")), want: NewStr("abc 123").ToObject()},
		{args: wrapArgs(newTestTuple("abc", 123), wrapKWArgs("sep", "XX", "end", "--")), want: NewStr("abcXX123--").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}*/

func TestBuiltinSetAttr(t *testing.T) {
	setattr := mustNotRaise(Builtins.GetItemString(NewRootFrame(), "setattr"))
	fooType := newTestClass("Foo", []*Type{ObjectType}, newStringDict(map[string]*Object{}))
	foo := newObject(fooType)
	fun := wrapFuncForTest(func(f *Frame, args ...*Object) (*Object, *BaseException) {
		result, raised := setattr.Call(f, args, nil)
		if raised != nil {
			return nil, raised
		}
		val, raised := GetAttr(f, args[0], toStrUnsafe(args[1]), nil)
		if raised != nil {
			return nil, raised
		}
		return newTestTuple(result, val).ToObject(), nil
	})
	cases := []invokeTestCase{
		{args: wrapArgs(foo), wantExc: mustCreateException(TypeErrorType, "'setattr' requires 3 arguments")},
		{args: wrapArgs(newObject(fooType), "foo", "bar"), want: newTestTuple(None, "bar").ToObject()},
		{args: wrapArgs(newObject(fooType), "foo", 123), want: newTestTuple(None, 123).ToObject()},
		{args: wrapArgs(foo, "foo"), wantExc: mustCreateException(TypeErrorType, "'setattr' requires 3 arguments")},
		{args: wrapArgs(foo, "foo", 123, None), wantExc: mustCreateException(TypeErrorType, "'setattr' requires 3 arguments")},
		{args: wrapArgs(foo, 123, 123), wantExc: mustCreateException(TypeErrorType, "'setattr' requires a 'str' object but received a \"int\"")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

// TODO(corona10): Re-enable once #282 is addressed.
/*func TestRawInput(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, s string, args ...*Object) (*Object, *BaseException) {
		// Create a fake Stdin for input test.
		stdinFile, w, err := os.Pipe()
		if err != nil {
			return nil, f.RaiseType(RuntimeErrorType, fmt.Sprintf("failed to open pipe: %v", err))
		}

		go func() {
			w.Write([]byte(s))
			w.Close()
		}()

		oldStdin := Stdin
		Stdin = NewFileFromFD(stdinFile.Fd(), nil)
		defer func() {
			Stdin = oldStdin
			stdinFile.Close()
		}()

		var input *Object
		output, raised := captureStdout(f, func() *BaseException {
			in, raised := builtinRawInput(f, args, nil)
			input = in
			return raised
		})

		if raised != nil {
			return nil, raised
		}

		return newTestTuple(input, output).ToObject(), nil
	})

	cases := []invokeTestCase{
		{args: wrapArgs("HelloGrumpy\n", ""), want: newTestTuple("HelloGrumpy", "").ToObject()},
		{args: wrapArgs("HelloGrumpy\n", "ShouldBeShown\nShouldBeShown\t"), want: newTestTuple("HelloGrumpy", "ShouldBeShown\nShouldBeShown\t").ToObject()},
		{args: wrapArgs("HelloGrumpy\n", 5, 4), wantExc: mustCreateException(TypeErrorType, "[raw_]input expcted at most 1 arguments, got 2")},
		{args: wrapArgs("HelloGrumpy\nHelloGrumpy\n", ""), want: newTestTuple("HelloGrumpy", "").ToObject()},
		{args: wrapArgs("HelloGrumpy\nHelloGrumpy\n", "ShouldBeShown\nShouldBeShown\t"), want: newTestTuple("HelloGrumpy", "ShouldBeShown\nShouldBeShown\t").ToObject()},
		{args: wrapArgs("HelloGrumpy\nHelloGrumpy\n", 5, 4), wantExc: mustCreateException(TypeErrorType, "[raw_]input expcted at most 1 arguments, got 2")},
		{args: wrapArgs("", ""), wantExc: mustCreateException(EOFErrorType, "EOF when reading a line")},
		{args: wrapArgs("", "ShouldBeShown\nShouldBeShown\t"), wantExc: mustCreateException(EOFErrorType, "EOF when reading a line")},
		{args: wrapArgs("", 5, 4), wantExc: mustCreateException(TypeErrorType, "[raw_]input expcted at most 1 arguments, got 2")},
	}

	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}

}*/

func newTestIndexObject(index int) *Object {
	indexType := newTestClass("Index", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__index__": newBuiltinFunction("__index__", func(f *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return NewInt(index).ToObject(), nil
		}).ToObject(),
	}))
	return newObject(indexType)
}

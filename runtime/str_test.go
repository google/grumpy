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
	"reflect"
	"testing"
)

func TestNewStr(t *testing.T) {
	expected := &Str{Object: Object{typ: StrType}, value: "foo"}
	s := NewStr("foo")
	if !reflect.DeepEqual(s, expected) {
		t.Errorf(`NewStr("foo") = %+v, expected %+v`, *s, *expected)
	}
}

// >>> hash("foo")
// -4177197833195190597
// >>> hash("bar")
// 327024216814240868
// >>> hash("baz")
// 327024216814240876
func TestHashString(t *testing.T) {
	cases := []struct {
		value string
		hash  int
	}{
		{"foo", -4177197833195190597},
		{"bar", 327024216814240868},
		{"baz", 327024216814240876},
	}
	for _, cas := range cases {
		if h := hashString(cas.value); h != cas.hash {
			t.Errorf("hashString(%q) = %d, expected %d", cas.value, h, cas.hash)
		}
	}
}

func TestStrBinaryOps(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, fn binaryOpFunc, v *Object, w *Object) (*Object, *BaseException) {
		return fn(f, v, w)
	})
	cases := []invokeTestCase{
		{args: wrapArgs(Add, "foo", "bar"), want: NewStr("foobar").ToObject()},
		{args: wrapArgs(Add, "foo", NewUnicode("bar")), want: NewUnicode("foobar").ToObject()},
		{args: wrapArgs(Add, "baz", ""), want: NewStr("baz").ToObject()},
		{args: wrapArgs(Add, "", newObject(ObjectType)), wantExc: mustCreateException(TypeErrorType, "unsupported operand type(s) for +: 'str' and 'object'")},
		{args: wrapArgs(Add, None, ""), wantExc: mustCreateException(TypeErrorType, "unsupported operand type(s) for +: 'NoneType' and 'str'")},
		{args: wrapArgs(Mod, "%s", 42), want: NewStr("42").ToObject()},
		{args: wrapArgs(Mod, "%f", 3.14), want: NewStr("3.140000").ToObject()},
		{args: wrapArgs(Mod, "%%", NewTuple()), want: NewStr("%").ToObject()},
		{args: wrapArgs(Mod, "%r", "abc"), want: NewStr("'abc'").ToObject()},
		{args: wrapArgs(Mod, "%s %s", true), wantExc: mustCreateException(TypeErrorType, "not enough arguments for format string")},
		{args: wrapArgs(Mod, "%Z", None), wantExc: mustCreateException(ValueErrorType, "invalid format spec")},
		{args: wrapArgs(Mod, "%s", NewDict()), wantExc: mustCreateException(NotImplementedErrorType, "mappings not yet supported")},
		{args: wrapArgs(Mod, "% d", 23), wantExc: mustCreateException(NotImplementedErrorType, "conversion flags not yet supported")},
		{args: wrapArgs(Mod, "%.3f", 102.1), wantExc: mustCreateException(NotImplementedErrorType, "field width not yet supported")},
		{args: wrapArgs(Mod, "%x", 24), wantExc: mustCreateException(NotImplementedErrorType, "conversion type not yet supported: x")},
		{args: wrapArgs(Mod, "%f", None), wantExc: mustCreateException(TypeErrorType, "float argument required, not NoneType")},
		{args: wrapArgs(Mod, "%s", newTestTuple(123, None)), wantExc: mustCreateException(TypeErrorType, "not all arguments converted during string formatting")},
		{args: wrapArgs(Mul, "", 10), want: NewStr("").ToObject()},
		{args: wrapArgs(Mul, "foo", -2), want: NewStr("").ToObject()},
		{args: wrapArgs(Mul, "foobar", 0), want: NewStr("").ToObject()},
		{args: wrapArgs(Mul, "aloha", 2), want: NewStr("alohaaloha").ToObject()},
		{args: wrapArgs(Mul, 1, "baz"), want: NewStr("baz").ToObject()},
		{args: wrapArgs(Mul, newObject(ObjectType), "qux"), wantExc: mustCreateException(TypeErrorType, "unsupported operand type(s) for *: 'object' and 'str'")},
		{args: wrapArgs(Mul, "foo", ""), wantExc: mustCreateException(TypeErrorType, "unsupported operand type(s) for *: 'str' and 'str'")},
		{args: wrapArgs(Mul, "bar", MaxInt), wantExc: mustCreateException(OverflowErrorType, "result too large")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestStrCompare(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs("", ""), want: compareAllResultEq},
		{args: wrapArgs("foo", "foo"), want: compareAllResultEq},
		{args: wrapArgs("", "foo"), want: compareAllResultLT},
		{args: wrapArgs("foo", ""), want: compareAllResultGT},
		{args: wrapArgs("bar", "baz"), want: compareAllResultLT},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(compareAll, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestStrContains(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs("foobar", "foo"), want: True.ToObject()},
		{args: wrapArgs("abcdef", "bar"), want: False.ToObject()},
		{args: wrapArgs("", ""), want: True.ToObject()},
		{args: wrapArgs("foobar", NewUnicode("bar")), want: True.ToObject()},
		{args: wrapArgs("", 102.1), wantExc: mustCreateException(TypeErrorType, "'in <string>' requires string as left operand, not float")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(StrType, "__contains__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestStrDecode(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs("foo"), want: NewUnicode("foo").ToObject()},
		{args: wrapArgs("foo\xffbar", "utf8", "replace"), want: NewUnicode("foo\ufffdbar").ToObject()},
		{args: wrapArgs("foo\xffbar", "utf8", "ignore"), want: NewUnicode("foobar").ToObject()},
		// Bad error handler name only triggers LookupError when an
		// error is encountered.
		{args: wrapArgs("foobar", "utf8", "noexist"), want: NewUnicode("foobar").ToObject()},
		{args: wrapArgs("foo\xffbar", "utf8", "noexist"), wantExc: mustCreateException(LookupErrorType, "unknown error handler name 'noexist'")},
		{args: wrapArgs("foobar", "noexist"), wantExc: mustCreateException(LookupErrorType, "unknown encoding: noexist")},
		{args: wrapArgs("foo\xffbar"), wantExc: mustCreateException(UnicodeDecodeErrorType, "'utf8' codec can't decode byte 0xff in position 3")},
		// Surrogates are not valid UTF-8 and should raise, unlike
		// CPython 2.x.
		{args: wrapArgs("foo\xef\xbf\xbdbar", "utf8", "strict"), wantExc: mustCreateException(UnicodeDecodeErrorType, "'utf8' codec can't decode byte 0xef in position 3")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(StrType, "decode", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestStrGetItem(t *testing.T) {
	intIndexType := newTestClass("IntIndex", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__index__": newBuiltinFunction("__index__", func(f *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return NewInt(2).ToObject(), nil
		}).ToObject(),
	}))
	longIndexType := newTestClass("LongIndex", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__index__": newBuiltinFunction("__index__", func(f *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return NewLong(big.NewInt(2)).ToObject(), nil
		}).ToObject(),
	}))
	cases := []invokeTestCase{
		{args: wrapArgs("bar", 1), want: NewStr("a").ToObject()},
		{args: wrapArgs("foo", 3.14), wantExc: mustCreateException(TypeErrorType, "string indices must be integers, not float")},
		{args: wrapArgs("bar", big.NewInt(1)), want: NewStr("a").ToObject()},
		{args: wrapArgs("baz", -1), want: NewStr("z").ToObject()},
		{args: wrapArgs("baz", newObject(intIndexType)), want: NewStr("z").ToObject()},
		{args: wrapArgs("baz", newObject(longIndexType)), want: NewStr("z").ToObject()},
		{args: wrapArgs("baz", -4), wantExc: mustCreateException(IndexErrorType, "index out of range")},
		{args: wrapArgs("", 0), wantExc: mustCreateException(IndexErrorType, "index out of range")},
		{args: wrapArgs("foo", 3), wantExc: mustCreateException(IndexErrorType, "index out of range")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(StrType, "__getitem__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestStrNew(t *testing.T) {
	dummy := newObject(ObjectType)
	dummyStr := NewStr(fmt.Sprintf("<object object at %p>", dummy))
	fooType := newTestClass("Foo", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__str__": newBuiltinFunction("__str__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return NewInt(123).ToObject(), nil
		}).ToObject(),
	}))
	foo := newObject(fooType)
	strictEqType := newTestClassStrictEq("StrictEq", StrType)
	subType := newTestClass("SubType", []*Type{StrType}, newStringDict(map[string]*Object{}))
	subTypeObject := (&Str{Object: Object{typ: subType}, value: "abc"}).ToObject()
	goodSlotType := newTestClass("GoodSlot", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__str__": newBuiltinFunction("__str__", func(_ *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return NewStr("abc").ToObject(), nil
		}).ToObject(),
	}))
	badSlotType := newTestClass("BadSlot", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__str__": newBuiltinFunction("__str__", func(_ *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return newObject(ObjectType), nil
		}).ToObject(),
	}))
	slotSubTypeType := newTestClass("SlotSubType", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__str__": newBuiltinFunction("__str__", func(_ *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return subTypeObject, nil
		}).ToObject(),
	}))
	cases := []invokeTestCase{
		{wantExc: mustCreateException(TypeErrorType, "'__new__' requires 1 arguments")},
		{args: wrapArgs(IntType.ToObject()), wantExc: mustCreateException(TypeErrorType, "str.__new__(int): int is not a subtype of str")},
		{args: wrapArgs(StrType.ToObject(), NewInt(1).ToObject(), NewInt(2).ToObject()), wantExc: mustCreateException(TypeErrorType, "str() takes at most 1 argument (2 given)")},
		{args: wrapArgs(StrType.ToObject(), foo), wantExc: mustCreateException(TypeErrorType, "__str__ returned non-string (type int)")},
		{args: wrapArgs(StrType.ToObject()), want: NewStr("").ToObject()},
		{args: wrapArgs(StrType.ToObject(), NewDict().ToObject()), want: NewStr("{}").ToObject()},
		{args: wrapArgs(StrType.ToObject(), dummy), want: dummyStr.ToObject()},
		{args: wrapArgs(strictEqType, "foo"), want: (&Str{Object: Object{typ: strictEqType}, value: "foo"}).ToObject()},
		{args: wrapArgs(StrType, newObject(goodSlotType)), want: NewStr("abc").ToObject()},
		{args: wrapArgs(StrType, newObject(badSlotType)), wantExc: mustCreateException(TypeErrorType, "__str__ returned non-string (type object)")},
		{args: wrapArgs(StrType, newObject(slotSubTypeType)), want: subTypeObject},
		{args: wrapArgs(strictEqType, newObject(goodSlotType)), want: (&Str{Object: Object{typ: strictEqType}, value: "abc"}).ToObject()},
		{args: wrapArgs(strictEqType, newObject(badSlotType)), wantExc: mustCreateException(TypeErrorType, "__str__ returned non-string (type object)")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(StrType, "__new__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestStrRepr(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs("foo"), want: NewStr(`'foo'`).ToObject()},
		{args: wrapArgs("on\nmultiple\nlines"), want: NewStr(`'on\nmultiple\nlines'`).ToObject()},
		{args: wrapArgs("\x00\x00"), want: NewStr(`'\x00\x00'`).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(StrType, "__repr__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestStrMethods(t *testing.T) {
	cases := []struct {
		methodName string
		args       Args
		want       *Object
		wantExc    *BaseException
	}{
		{"join", wrapArgs(",", newTestList("foo", "bar")), NewStr("foo,bar").ToObject(), nil},
		{"join", wrapArgs(":", newTestList("foo", "bar", NewUnicode("baz"))), NewUnicode("foo:bar:baz").ToObject(), nil},
		{"join", wrapArgs("nope", NewTuple()), NewStr("").ToObject(), nil},
		{"join", wrapArgs("nope", newTestTuple("foo")), NewStr("foo").ToObject(), nil},
		{"join", wrapArgs(",", newTestList("foo", "bar", 3.14)), nil, mustCreateException(TypeErrorType, "sequence item 2: expected string, float found")},
		{"join", wrapArgs("\xff", newTestList(NewUnicode("foo"), NewUnicode("bar"))), nil, mustCreateException(UnicodeDecodeErrorType, "'utf8' codec can't decode byte 0xff in position 0")},
		{"split", wrapArgs("foo,bar", ","), newTestList("foo", "bar").ToObject(), nil},
		{"split", wrapArgs("1,2,3", ",", 1), newTestList("1", "2,3").ToObject(), nil},
		{"split", wrapArgs("a \tb\nc"), newTestList("a", "b", "c").ToObject(), nil},
		{"split", wrapArgs("a \tb\nc", None), newTestList("a", "b", "c").ToObject(), nil},
		{"split", wrapArgs("a \tb\nc", None, -1), newTestList("a", "b", "c").ToObject(), nil},
		{"split", wrapArgs("a \tb\nc", None, 1), newTestList("a", "b\nc").ToObject(), nil},
		{"split", wrapArgs("foo", 1), nil, mustCreateException(TypeErrorType, "expected a str separator")},
		{"split", wrapArgs("foo", ""), nil, mustCreateException(ValueErrorType, "empty separator")},
		{"startswith", wrapArgs("", ""), True.ToObject(), nil},
		{"startswith", wrapArgs("", "", 1), True.ToObject(), nil},
		{"startswith", wrapArgs("foobar", "foo"), True.ToObject(), nil},
		{"startswith", wrapArgs("foobar", "foo", 2), False.ToObject(), nil},
		{"startswith", wrapArgs("foobar", "bar", 3), True.ToObject(), nil},
		{"startswith", wrapArgs("foobar", "bar", 3, 5), False.ToObject(), nil},
		{"startswith", wrapArgs("foobar", "bar", 5, 3), False.ToObject(), nil},
		{"startswith", wrapArgs("foo", "foobar"), False.ToObject(), nil},
		{"startswith", wrapArgs("foo", newTestTuple("foobar", "fo").ToObject()), True.ToObject(), nil},
		{"startswith", wrapArgs("foo", 123), nil, mustCreateException(TypeErrorType, "startswith first arg must be str or tuple, not int")},
		{"startswith", wrapArgs("foo", newTestTuple(123).ToObject()), nil, mustCreateException(TypeErrorType, "expected a str")},
	}
	for _, cas := range cases {
		testCase := invokeTestCase{args: cas.args, want: cas.want, wantExc: cas.wantExc}
		if err := runInvokeMethodTestCase(StrType, cas.methodName, &testCase); err != "" {
			t.Error(err)
		}
	}
}

func TestStrStr(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs("foo"), want: NewStr("foo").ToObject()},
		{args: wrapArgs("on\nmultiple\nlines"), want: NewStr("on\nmultiple\nlines").ToObject()},
		{args: wrapArgs("\x00\x00"), want: NewStr("\x00\x00").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(StrType, "__str__", &cas); err != "" {
			t.Error(err)
		}
	}
}

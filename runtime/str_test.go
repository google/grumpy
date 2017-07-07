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
	"runtime"
	"testing"
)

func TestNewStr(t *testing.T) {
	expected := &Str{Object: Object{typ: StrType}, value: "foo"}
	s := NewStr("foo")
	if !reflect.DeepEqual(s, expected) {
		t.Errorf(`NewStr("foo") = %+v, expected %+v`, *s, *expected)
	}
}

func BenchmarkNewStr(b *testing.B) {
	var ret *Str
	for i := 0; i < b.N; i++ {
		ret = NewStr("foo")
	}
	runtime.KeepAlive(ret)
}

// # On a 64bit system:
// >>> hash("foo")
// -4177197833195190597
// >>> hash("bar")
// 327024216814240868
// >>> hash("baz")
// 327024216814240876
func TestHashString(t *testing.T) {
	truncateInt := func(i int64) int { return int(i) } // Support for 32bit platforms
	cases := []struct {
		value string
		hash  int
	}{
		{"foo", truncateInt(-4177197833195190597)},
		{"bar", truncateInt(327024216814240868)},
		{"baz", truncateInt(327024216814240876)},
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
		{args: wrapArgs(Mod, "%3s", 42), want: NewStr(" 42").ToObject()},
		{args: wrapArgs(Mod, "%03s", 42), want: NewStr(" 42").ToObject()},
		{args: wrapArgs(Mod, "%f", 3.14), want: NewStr("3.140000").ToObject()},
		{args: wrapArgs(Mod, "%10f", 3.14), want: NewStr("  3.140000").ToObject()},
		{args: wrapArgs(Mod, "%010f", 3.14), want: NewStr("003.140000").ToObject()},
		{args: wrapArgs(Mod, "abc %d", NewLong(big.NewInt(123))), want: NewStr("abc 123").ToObject()},
		{args: wrapArgs(Mod, "%d", 3.14), want: NewStr("3").ToObject()},
		{args: wrapArgs(Mod, "%%", NewTuple()), want: NewStr("%").ToObject()},
		{args: wrapArgs(Mod, "%3%", NewTuple()), want: NewStr("  %").ToObject()},
		{args: wrapArgs(Mod, "%03%", NewTuple()), want: NewStr("  %").ToObject()},
		{args: wrapArgs(Mod, "%r", "abc"), want: NewStr("'abc'").ToObject()},
		{args: wrapArgs(Mod, "%6r", "abc"), want: NewStr(" 'abc'").ToObject()},
		{args: wrapArgs(Mod, "%06r", "abc"), want: NewStr(" 'abc'").ToObject()},
		{args: wrapArgs(Mod, "%s %s", true), wantExc: mustCreateException(TypeErrorType, "not enough arguments for format string")},
		{args: wrapArgs(Mod, "%Z", None), wantExc: mustCreateException(ValueErrorType, "invalid format spec")},
		{args: wrapArgs(Mod, "%s", NewDict()), wantExc: mustCreateException(NotImplementedErrorType, "mappings not yet supported")},
		{args: wrapArgs(Mod, "% d", 23), wantExc: mustCreateException(NotImplementedErrorType, "conversion flags not yet supported")},
		{args: wrapArgs(Mod, "%.3f", 102.1), wantExc: mustCreateException(NotImplementedErrorType, "field width not yet supported")},
		{args: wrapArgs(Mod, "%x", 0x1f), want: NewStr("1f").ToObject()},
		{args: wrapArgs(Mod, "%X", 0xffff), want: NewStr("FFFF").ToObject()},
		{args: wrapArgs(Mod, "%x", 1.2), want: NewStr("1").ToObject()},
		{args: wrapArgs(Mod, "abc %x", NewLong(big.NewInt(123))), want: NewStr("abc 7b").ToObject()},
		{args: wrapArgs(Mod, "%x", None), wantExc: mustCreateException(TypeErrorType, "an integer is required")},
		{args: wrapArgs(Mod, "%f", None), wantExc: mustCreateException(TypeErrorType, "float argument required, not NoneType")},
		{args: wrapArgs(Mod, "%s", newTestTuple(123, None)), wantExc: mustCreateException(TypeErrorType, "not all arguments converted during string formatting")},
		{args: wrapArgs(Mod, "%d", newTestTuple("123")), wantExc: mustCreateException(TypeErrorType, "an integer is required")},
		{args: wrapArgs(Mod, "%o", newTestTuple(123)), want: NewStr("173").ToObject()},
		{args: wrapArgs(Mod, "%o", 8), want: NewStr("10").ToObject()},
		{args: wrapArgs(Mod, "%o", -8), want: NewStr("-10").ToObject()},
		{args: wrapArgs(Mod, "%03o", newTestTuple(123)), want: NewStr("173").ToObject()},
		{args: wrapArgs(Mod, "%04o", newTestTuple(123)), want: NewStr("0173").ToObject()},
		{args: wrapArgs(Mod, "%o", newTestTuple("123")), wantExc: mustCreateException(TypeErrorType, "an integer is required")},
		{args: wrapArgs(Mod, "%o", None), wantExc: mustCreateException(TypeErrorType, "an integer is required")},
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
		{args: wrapArgs("foo", 3.14), wantExc: mustCreateException(TypeErrorType, "string indices must be integers or slice, not float")},
		{args: wrapArgs("bar", big.NewInt(1)), want: NewStr("a").ToObject()},
		{args: wrapArgs("baz", -1), want: NewStr("z").ToObject()},
		{args: wrapArgs("baz", newObject(intIndexType)), want: NewStr("z").ToObject()},
		{args: wrapArgs("baz", newObject(longIndexType)), want: NewStr("z").ToObject()},
		{args: wrapArgs("baz", -4), wantExc: mustCreateException(IndexErrorType, "index out of range")},
		{args: wrapArgs("", 0), wantExc: mustCreateException(IndexErrorType, "index out of range")},
		{args: wrapArgs("foo", 3), wantExc: mustCreateException(IndexErrorType, "index out of range")},
		{args: wrapArgs("bar", newTestSlice(None, 2)), want: NewStr("ba").ToObject()},
		{args: wrapArgs("bar", newTestSlice(1, 3)), want: NewStr("ar").ToObject()},
		{args: wrapArgs("bar", newTestSlice(1, None)), want: NewStr("ar").ToObject()},
		{args: wrapArgs("foobarbaz", newTestSlice(1, 8, 2)), want: NewStr("obra").ToObject()},
		{args: wrapArgs("abc", newTestSlice(None, None, -1)), want: NewStr("cba").ToObject()},
		{args: wrapArgs("bar", newTestSlice(1, 2, 0)), wantExc: mustCreateException(ValueErrorType, "slice step cannot be zero")},
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
	fooType := newTestClass("Foo", []*Type{ObjectType}, newStringDict(map[string]*Object{"bar": None}))
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
	intIntType := newTestClass("IntInt", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__int__": newBuiltinFunction("__int__", func(f *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return NewInt(2).ToObject(), nil
		}).ToObject(),
	}))
	longIntType := newTestClass("LongInt", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__int__": newBuiltinFunction("__int__", func(f *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return NewLong(big.NewInt(2)).ToObject(), nil
		}).ToObject(),
	}))
	cases := []struct {
		methodName string
		args       Args
		want       *Object
		wantExc    *BaseException
	}{
		{"capitalize", wrapArgs(""), NewStr("").ToObject(), nil},
		{"capitalize", wrapArgs("foobar"), NewStr("Foobar").ToObject(), nil},
		{"capitalize", wrapArgs("FOOBAR"), NewStr("Foobar").ToObject(), nil},
		{"capitalize", wrapArgs("ùBAR"), NewStr("ùbar").ToObject(), nil},
		{"capitalize", wrapArgs("вол"), NewStr("вол").ToObject(), nil},
		{"capitalize", wrapArgs("foobar", 123), nil, mustCreateException(TypeErrorType, "'capitalize' of 'str' requires 1 arguments")},
		{"capitalize", wrapArgs("ВОЛ"), NewStr("ВОЛ").ToObject(), nil},
		{"center", wrapArgs("foobar", 9, "#"), NewStr("##foobar#").ToObject(), nil},
		{"center", wrapArgs("foobar", 10, "#"), NewStr("##foobar##").ToObject(), nil},
		{"center", wrapArgs("foobar", 3, "#"), NewStr("foobar").ToObject(), nil},
		{"center", wrapArgs("foobar", -1, "#"), NewStr("foobar").ToObject(), nil},
		{"center", wrapArgs("foobar", 10, "##"), nil, mustCreateException(TypeErrorType, "center() argument 2 must be char, not str")},
		{"center", wrapArgs("foobar", 10, ""), nil, mustCreateException(TypeErrorType, "center() argument 2 must be char, not str")},
		{"count", wrapArgs("", "a"), NewInt(0).ToObject(), nil},
		{"count", wrapArgs("five", ""), NewInt(5).ToObject(), nil},
		{"count", wrapArgs("abba", "bb"), NewInt(1).ToObject(), nil},
		{"count", wrapArgs("abbba", "bb"), NewInt(1).ToObject(), nil},
		{"count", wrapArgs("abbbba", "bb"), NewInt(2).ToObject(), nil},
		{"count", wrapArgs("abcdeffdeabcb", "b"), NewInt(3).ToObject(), nil},
		{"count", wrapArgs(""), nil, mustCreateException(TypeErrorType, "'count' of 'str' requires 2 arguments")},
		{"endswith", wrapArgs("", ""), True.ToObject(), nil},
		{"endswith", wrapArgs("", "", 1), False.ToObject(), nil},
		{"endswith", wrapArgs("foobar", "bar"), True.ToObject(), nil},
		{"endswith", wrapArgs("foobar", "bar", 0, -2), False.ToObject(), nil},
		{"endswith", wrapArgs("foobar", "foo", 0, 3), True.ToObject(), nil},
		{"endswith", wrapArgs("foobar", "bar", 3, 5), False.ToObject(), nil},
		{"endswith", wrapArgs("foobar", "bar", 5, 3), False.ToObject(), nil},
		{"endswith", wrapArgs("bar", "foobar"), False.ToObject(), nil},
		{"endswith", wrapArgs("foo", newTestTuple("barfoo", "oo").ToObject()), True.ToObject(), nil},
		{"endswith", wrapArgs("foo", 123), nil, mustCreateException(TypeErrorType, "endswith first arg must be str, unicode, or tuple, not int")},
		{"endswith", wrapArgs("foo", newTestTuple(123).ToObject()), nil, mustCreateException(TypeErrorType, "expected a str")},
		{"find", wrapArgs("", ""), NewInt(0).ToObject(), nil},
		{"find", wrapArgs("", "", 1), NewInt(-1).ToObject(), nil},
		{"find", wrapArgs("", "", -1), NewInt(0).ToObject(), nil},
		{"find", wrapArgs("", "", None, -1), NewInt(0).ToObject(), nil},
		{"find", wrapArgs("foobar", "bar"), NewInt(3).ToObject(), nil},
		{"find", wrapArgs("foobar", "bar", fooType), nil, mustCreateException(TypeErrorType, "slice indices must be integers or None or have an __index__ method")},
		{"find", wrapArgs("foobar", "bar", NewInt(MaxInt)), NewInt(-1).ToObject(), nil},
		{"find", wrapArgs("foobar", "bar", None, NewInt(MaxInt)), NewInt(3).ToObject(), nil},
		{"find", wrapArgs("foobar", "bar", newObject(intIndexType)), NewInt(3).ToObject(), nil},
		{"find", wrapArgs("foobar", "bar", None, newObject(intIndexType)), NewInt(-1).ToObject(), nil},
		{"find", wrapArgs("foobar", "bar", newObject(longIndexType)), NewInt(3).ToObject(), nil},
		{"find", wrapArgs("foobar", "bar", None, newObject(longIndexType)), NewInt(-1).ToObject(), nil},
		// TODO: Support unicode substring.
		{"find", wrapArgs("foobar", NewUnicode("bar")), nil, mustCreateException(TypeErrorType, "'find/index' requires a 'str' object but received a 'unicode'")},
		{"find", wrapArgs("foobar", "bar", "baz"), nil, mustCreateException(TypeErrorType, "slice indices must be integers or None or have an __index__ method")},
		{"find", wrapArgs("foobar", "bar", 0, "baz"), nil, mustCreateException(TypeErrorType, "slice indices must be integers or None or have an __index__ method")},
		{"find", wrapArgs("foobar", "bar", None), NewInt(3).ToObject(), nil},
		{"find", wrapArgs("foobar", "bar", 0, None), NewInt(3).ToObject(), nil},
		{"find", wrapArgs("foobar", "bar", 0, -2), NewInt(-1).ToObject(), nil},
		{"find", wrapArgs("foobar", "foo", 0, 3), NewInt(0).ToObject(), nil},
		{"find", wrapArgs("foobar", "foo", 10), NewInt(-1).ToObject(), nil},
		{"find", wrapArgs("foobar", "foo", 3, 3), NewInt(-1).ToObject(), nil},
		{"find", wrapArgs("foobar", "bar", 3, 5), NewInt(-1).ToObject(), nil},
		{"find", wrapArgs("foobar", "bar", 5, 3), NewInt(-1).ToObject(), nil},
		{"find", wrapArgs("bar", "foobar"), NewInt(-1).ToObject(), nil},
		{"find", wrapArgs("bar", "a", 1, 10), NewInt(1).ToObject(), nil},
		{"find", wrapArgs("bar", "a", NewLong(big.NewInt(1)), 10), NewInt(1).ToObject(), nil},
		{"find", wrapArgs("bar", "a", 0, NewLong(big.NewInt(2))), NewInt(1).ToObject(), nil},
		{"find", wrapArgs("bar", "a", 1, 3), NewInt(1).ToObject(), nil},
		{"find", wrapArgs("bar", "a", 0, -1), NewInt(1).ToObject(), nil},
		{"find", wrapArgs("foo", newTestTuple("barfoo", "oo").ToObject()), nil, mustCreateException(TypeErrorType, "'find/index' requires a 'str' object but received a 'tuple'")},
		{"find", wrapArgs("foo", 123), nil, mustCreateException(TypeErrorType, "'find/index' requires a 'str' object but received a 'int'")},
		{"index", wrapArgs("", ""), NewInt(0).ToObject(), nil},
		{"index", wrapArgs("", "", 1), nil, mustCreateException(ValueErrorType, "substring not found")},
		{"index", wrapArgs("", "", -1), NewInt(0).ToObject(), nil},
		{"index", wrapArgs("", "", None, -1), NewInt(0).ToObject(), nil},
		{"index", wrapArgs("foobar", "bar"), NewInt(3).ToObject(), nil},
		{"index", wrapArgs("foobar", "bar", fooType), nil, mustCreateException(TypeErrorType, "slice indices must be integers or None or have an __index__ method")},
		{"index", wrapArgs("foobar", "bar", NewInt(MaxInt)), nil, mustCreateException(ValueErrorType, "substring not found")},
		{"index", wrapArgs("foobar", "bar", None, NewInt(MaxInt)), NewInt(3).ToObject(), nil},
		{"index", wrapArgs("foobar", "bar", newObject(intIndexType)), NewInt(3).ToObject(), nil},
		{"index", wrapArgs("foobar", "bar", None, newObject(intIndexType)), nil, mustCreateException(ValueErrorType, "substring not found")},
		{"index", wrapArgs("foobar", "bar", newObject(longIndexType)), NewInt(3).ToObject(), nil},
		{"index", wrapArgs("foobar", "bar", None, newObject(longIndexType)), nil, mustCreateException(ValueErrorType, "substring not found")},
		//TODO: Support unicode substring.
		{"index", wrapArgs("foobar", NewUnicode("bar")), nil, mustCreateException(TypeErrorType, "'find/index' requires a 'str' object but received a 'unicode'")},
		{"index", wrapArgs("foobar", "bar", "baz"), nil, mustCreateException(TypeErrorType, "slice indices must be integers or None or have an __index__ method")},
		{"index", wrapArgs("foobar", "bar", 0, "baz"), nil, mustCreateException(TypeErrorType, "slice indices must be integers or None or have an __index__ method")},
		{"index", wrapArgs("foobar", "bar", None), NewInt(3).ToObject(), nil},
		{"index", wrapArgs("foobar", "bar", 0, None), NewInt(3).ToObject(), nil},
		{"index", wrapArgs("foobar", "bar", 0, -2), nil, mustCreateException(ValueErrorType, "substring not found")},
		{"index", wrapArgs("foobar", "foo", 0, 3), NewInt(0).ToObject(), nil},
		{"index", wrapArgs("foobar", "foo", 10), nil, mustCreateException(ValueErrorType, "substring not found")},
		{"index", wrapArgs("foobar", "foo", 3, 3), nil, mustCreateException(ValueErrorType, "substring not found")},
		{"index", wrapArgs("foobar", "bar", 3, 5), nil, mustCreateException(ValueErrorType, "substring not found")},
		{"index", wrapArgs("foobar", "bar", 5, 3), nil, mustCreateException(ValueErrorType, "substring not found")},
		{"index", wrapArgs("bar", "foobar"), nil, mustCreateException(ValueErrorType, "substring not found")},
		{"index", wrapArgs("bar", "a", 1, 10), NewInt(1).ToObject(), nil},
		{"index", wrapArgs("bar", "a", NewLong(big.NewInt(1)), 10), NewInt(1).ToObject(), nil},
		{"index", wrapArgs("bar", "a", 0, NewLong(big.NewInt(2))), NewInt(1).ToObject(), nil},
		{"index", wrapArgs("bar", "a", 1, 3), NewInt(1).ToObject(), nil},
		{"index", wrapArgs("bar", "a", 0, -1), NewInt(1).ToObject(), nil},
		{"index", wrapArgs("foo", newTestTuple("barfoo", "oo").ToObject()), nil, mustCreateException(TypeErrorType, "'find/index' requires a 'str' object but received a 'tuple'")},
		{"index", wrapArgs("foo", 123), nil, mustCreateException(TypeErrorType, "'find/index' requires a 'str' object but received a 'int'")},
		{"index", wrapArgs("barbaz", "ba"), NewInt(0).ToObject(), nil},
		{"index", wrapArgs("barbaz", "ba", 1), NewInt(3).ToObject(), nil},
		{"isalnum", wrapArgs("123abc"), True.ToObject(), nil},
		{"isalnum", wrapArgs(""), False.ToObject(), nil},
		{"isalnum", wrapArgs("#$%"), False.ToObject(), nil},
		{"isalnum", wrapArgs("abc#123"), False.ToObject(), nil},
		{"isalnum", wrapArgs("123abc", "efg"), nil, mustCreateException(TypeErrorType, "'isalnum' of 'str' requires 1 arguments")},
		{"isalpha", wrapArgs("xyz"), True.ToObject(), nil},
		{"isalpha", wrapArgs(""), False.ToObject(), nil},
		{"isalpha", wrapArgs("#$%"), False.ToObject(), nil},
		{"isalpha", wrapArgs("abc#123"), False.ToObject(), nil},
		{"isalpha", wrapArgs("absd", "efg"), nil, mustCreateException(TypeErrorType, "'isalpha' of 'str' requires 1 arguments")},
		{"isdigit", wrapArgs("abc"), False.ToObject(), nil},
		{"isdigit", wrapArgs("123"), True.ToObject(), nil},
		{"isdigit", wrapArgs(""), False.ToObject(), nil},
		{"isdigit", wrapArgs("abc#123"), False.ToObject(), nil},
		{"isdigit", wrapArgs("123", "456"), nil, mustCreateException(TypeErrorType, "'isdigit' of 'str' requires 1 arguments")},
		{"islower", wrapArgs("abc"), True.ToObject(), nil},
		{"islower", wrapArgs("ABC"), False.ToObject(), nil},
		{"islower", wrapArgs(""), False.ToObject(), nil},
		{"islower", wrapArgs("abc#123"), False.ToObject(), nil},
		{"islower", wrapArgs("123", "456"), nil, mustCreateException(TypeErrorType, "'islower' of 'str' requires 1 arguments")},
		{"isupper", wrapArgs("abc"), False.ToObject(), nil},
		{"isupper", wrapArgs("ABC"), True.ToObject(), nil},
		{"isupper", wrapArgs(""), False.ToObject(), nil},
		{"isupper", wrapArgs("abc#123"), False.ToObject(), nil},
		{"isupper", wrapArgs("123", "456"), nil, mustCreateException(TypeErrorType, "'isupper' of 'str' requires 1 arguments")},
		{"isspace", wrapArgs(""), False.ToObject(), nil},
		{"isspace", wrapArgs(" "), True.ToObject(), nil},
		{"isspace", wrapArgs("\n\t\v\f\r      "), True.ToObject(), nil},
		{"isspace", wrapArgs(""), False.ToObject(), nil},
		{"isspace", wrapArgs("asdad"), False.ToObject(), nil},
		{"isspace", wrapArgs("       "), True.ToObject(), nil},
		{"isspace", wrapArgs("    ", "456"), nil, mustCreateException(TypeErrorType, "'isspace' of 'str' requires 1 arguments")},
		{"istitle", wrapArgs("abc"), False.ToObject(), nil},
		{"istitle", wrapArgs("Abc&D"), True.ToObject(), nil},
		{"istitle", wrapArgs("ABc&D"), False.ToObject(), nil},
		{"istitle", wrapArgs(""), False.ToObject(), nil},
		{"istitle", wrapArgs("abc#123"), False.ToObject(), nil},
		{"istitle", wrapArgs("ABc&D", "456"), nil, mustCreateException(TypeErrorType, "'istitle' of 'str' requires 1 arguments")},
		{"join", wrapArgs(",", newTestList("foo", "bar")), NewStr("foo,bar").ToObject(), nil},
		{"join", wrapArgs(":", newTestList("foo", "bar", NewUnicode("baz"))), NewUnicode("foo:bar:baz").ToObject(), nil},
		{"join", wrapArgs("nope", NewTuple()), NewStr("").ToObject(), nil},
		{"join", wrapArgs("nope", newTestTuple("foo")), NewStr("foo").ToObject(), nil},
		{"join", wrapArgs(",", newTestList("foo", "bar", 3.14)), nil, mustCreateException(TypeErrorType, "sequence item 2: expected string, float found")},
		{"join", wrapArgs("\xff", newTestList(NewUnicode("foo"), NewUnicode("bar"))), nil, mustCreateException(UnicodeDecodeErrorType, "'utf8' codec can't decode byte 0xff in position 0")},
		{"ljust", wrapArgs("foobar", 10, "#"), NewStr("foobar####").ToObject(), nil},
		{"ljust", wrapArgs("foobar", 3, "#"), NewStr("foobar").ToObject(), nil},
		{"ljust", wrapArgs("foobar", -1, "#"), NewStr("foobar").ToObject(), nil},
		{"ljust", wrapArgs("foobar", 10, "##"), nil, mustCreateException(TypeErrorType, "ljust() argument 2 must be char, not str")},
		{"ljust", wrapArgs("foobar", 10, ""), nil, mustCreateException(TypeErrorType, "ljust() argument 2 must be char, not str")},
		{"lower", wrapArgs(""), NewStr("").ToObject(), nil},
		{"lower", wrapArgs("a"), NewStr("a").ToObject(), nil},
		{"lower", wrapArgs("A"), NewStr("a").ToObject(), nil},
		{"lower", wrapArgs(" A"), NewStr(" a").ToObject(), nil},
		{"lower", wrapArgs("abc"), NewStr("abc").ToObject(), nil},
		{"lower", wrapArgs("ABC"), NewStr("abc").ToObject(), nil},
		{"lower", wrapArgs("aBC"), NewStr("abc").ToObject(), nil},
		{"lower", wrapArgs("abc def", 123), nil, mustCreateException(TypeErrorType, "'lower' of 'str' requires 1 arguments")},
		{"lower", wrapArgs(123), nil, mustCreateException(TypeErrorType, "unbound method lower() must be called with str instance as first argument (got int instance instead)")},
		{"lower", wrapArgs("вол"), NewStr("вол").ToObject(), nil},
		{"lower", wrapArgs("ВОЛ"), NewStr("ВОЛ").ToObject(), nil},
		{"lstrip", wrapArgs("foo "), NewStr("foo ").ToObject(), nil},
		{"lstrip", wrapArgs(" foo bar "), NewStr("foo bar ").ToObject(), nil},
		{"lstrip", wrapArgs("foo foo", "o"), NewStr("foo foo").ToObject(), nil},
		{"lstrip", wrapArgs("foo foo", "f"), NewStr("oo foo").ToObject(), nil},
		{"lstrip", wrapArgs("foo bar", "abr"), NewStr("foo bar").ToObject(), nil},
		{"lstrip", wrapArgs("foo bar", "fo"), NewStr(" bar").ToObject(), nil},
		{"lstrip", wrapArgs("foo", NewUnicode("f")), NewUnicode("oo").ToObject(), nil},
		{"lstrip", wrapArgs("123", 3), nil, mustCreateException(TypeErrorType, "strip arg must be None, str or unicode")},
		{"lstrip", wrapArgs("foo", "bar", "baz"), nil, mustCreateException(TypeErrorType, "'strip' of 'str' requires 2 arguments")},
		{"lstrip", wrapArgs("\xfboo", NewUnicode("o")), nil, mustCreateException(UnicodeDecodeErrorType, "'utf8' codec can't decode byte 0xfb in position 0")},
		{"lstrip", wrapArgs("foo", NewUnicode("o")), NewUnicode("f").ToObject(), nil},
		{"rfind", wrapArgs("", ""), NewInt(0).ToObject(), nil},
		{"rfind", wrapArgs("", "", 1), NewInt(-1).ToObject(), nil},
		{"rfind", wrapArgs("", "", -1), NewInt(0).ToObject(), nil},
		{"rfind", wrapArgs("", "", None, -1), NewInt(0).ToObject(), nil},
		{"rfind", wrapArgs("foobar", "bar"), NewInt(3).ToObject(), nil},
		{"rfind", wrapArgs("foobar", "bar", fooType), nil, mustCreateException(TypeErrorType, "slice indices must be integers or None or have an __index__ method")},
		{"rfind", wrapArgs("foobar", "bar", NewInt(MaxInt)), NewInt(-1).ToObject(), nil},
		{"rfind", wrapArgs("foobar", "bar", None, NewInt(MaxInt)), NewInt(3).ToObject(), nil},
		{"rfind", wrapArgs("foobar", "bar", newObject(intIndexType)), NewInt(3).ToObject(), nil},
		{"rfind", wrapArgs("foobar", "bar", None, newObject(intIndexType)), NewInt(-1).ToObject(), nil},
		{"rfind", wrapArgs("foobar", "bar", newObject(longIndexType)), NewInt(3).ToObject(), nil},
		{"rfind", wrapArgs("foobar", "bar", None, newObject(longIndexType)), NewInt(-1).ToObject(), nil},
		//r TODO: Support unicode substring.
		{"rfind", wrapArgs("foobar", NewUnicode("bar")), nil, mustCreateException(TypeErrorType, "'find/index' requires a 'str' object but received a 'unicode'")},
		{"rfind", wrapArgs("foobar", "bar", "baz"), nil, mustCreateException(TypeErrorType, "slice indices must be integers or None or have an __index__ method")},
		{"rfind", wrapArgs("foobar", "bar", 0, "baz"), nil, mustCreateException(TypeErrorType, "slice indices must be integers or None or have an __index__ method")},
		{"rfind", wrapArgs("foobar", "bar", None), NewInt(3).ToObject(), nil},
		{"rfind", wrapArgs("foobar", "bar", 0, None), NewInt(3).ToObject(), nil},
		{"rfind", wrapArgs("foobar", "bar", 0, -2), NewInt(-1).ToObject(), nil},
		{"rfind", wrapArgs("foobar", "foo", 0, 3), NewInt(0).ToObject(), nil},
		{"rfind", wrapArgs("foobar", "foo", 10), NewInt(-1).ToObject(), nil},
		{"rfind", wrapArgs("foobar", "foo", 3, 3), NewInt(-1).ToObject(), nil},
		{"rfind", wrapArgs("foobar", "bar", 3, 5), NewInt(-1).ToObject(), nil},
		{"rfind", wrapArgs("foobar", "bar", 5, 3), NewInt(-1).ToObject(), nil},
		{"rfind", wrapArgs("bar", "foobar"), NewInt(-1).ToObject(), nil},
		{"rfind", wrapArgs("bar", "a", 1, 10), NewInt(1).ToObject(), nil},
		{"rfind", wrapArgs("bar", "a", NewLong(big.NewInt(1)), 10), NewInt(1).ToObject(), nil},
		{"rfind", wrapArgs("bar", "a", 0, NewLong(big.NewInt(2))), NewInt(1).ToObject(), nil},
		{"rfind", wrapArgs("bar", "a", 1, 3), NewInt(1).ToObject(), nil},
		{"rfind", wrapArgs("bar", "a", 0, -1), NewInt(1).ToObject(), nil},
		{"rfind", wrapArgs("foo", newTestTuple("barfoo", "oo").ToObject()), nil, mustCreateException(TypeErrorType, "'find/index' requires a 'str' object but received a 'tuple'")},
		{"rfind", wrapArgs("foo", 123), nil, mustCreateException(TypeErrorType, "'find/index' requires a 'str' object but received a 'int'")},
		{"rfind", wrapArgs("barbaz", "ba"), NewInt(3).ToObject(), nil},
		{"rfind", wrapArgs("barbaz", "ba", None, 4), NewInt(0).ToObject(), nil},
		{"rindex", wrapArgs("", ""), NewInt(0).ToObject(), nil},
		{"rindex", wrapArgs("", "", 1), nil, mustCreateException(ValueErrorType, "substring not found")},
		{"rindex", wrapArgs("", "", -1), NewInt(0).ToObject(), nil},
		{"rindex", wrapArgs("", "", None, -1), NewInt(0).ToObject(), nil},
		{"rindex", wrapArgs("foobar", "bar"), NewInt(3).ToObject(), nil},
		{"rindex", wrapArgs("foobar", "bar", fooType), nil, mustCreateException(TypeErrorType, "slice indices must be integers or None or have an __index__ method")},
		{"rindex", wrapArgs("foobar", "bar", NewInt(MaxInt)), nil, mustCreateException(ValueErrorType, "substring not found")},
		{"rindex", wrapArgs("foobar", "bar", None, NewInt(MaxInt)), NewInt(3).ToObject(), nil},
		{"rindex", wrapArgs("foobar", "bar", newObject(intIndexType)), NewInt(3).ToObject(), nil},
		{"rindex", wrapArgs("foobar", "bar", None, newObject(intIndexType)), nil, mustCreateException(ValueErrorType, "substring not found")},
		{"rindex", wrapArgs("foobar", "bar", newObject(longIndexType)), NewInt(3).ToObject(), nil},
		{"rindex", wrapArgs("foobar", "bar", None, newObject(longIndexType)), nil, mustCreateException(ValueErrorType, "substring not found")},
		// TODO: Support unicode substring.
		{"rindex", wrapArgs("foobar", NewUnicode("bar")), nil, mustCreateException(TypeErrorType, "'find/index' requires a 'str' object but received a 'unicode'")},
		{"rindex", wrapArgs("foobar", "bar", "baz"), nil, mustCreateException(TypeErrorType, "slice indices must be integers or None or have an __index__ method")},
		{"rindex", wrapArgs("foobar", "bar", 0, "baz"), nil, mustCreateException(TypeErrorType, "slice indices must be integers or None or have an __index__ method")},
		{"rindex", wrapArgs("foobar", "bar", None), NewInt(3).ToObject(), nil},
		{"rindex", wrapArgs("foobar", "bar", 0, None), NewInt(3).ToObject(), nil},
		{"rindex", wrapArgs("foobar", "bar", 0, -2), nil, mustCreateException(ValueErrorType, "substring not found")},
		{"rindex", wrapArgs("foobar", "foo", 0, 3), NewInt(0).ToObject(), nil},
		{"rindex", wrapArgs("foobar", "foo", 10), nil, mustCreateException(ValueErrorType, "substring not found")},
		{"rindex", wrapArgs("foobar", "foo", 3, 3), nil, mustCreateException(ValueErrorType, "substring not found")},
		{"rindex", wrapArgs("foobar", "bar", 3, 5), nil, mustCreateException(ValueErrorType, "substring not found")},
		{"rindex", wrapArgs("foobar", "bar", 5, 3), nil, mustCreateException(ValueErrorType, "substring not found")},
		{"rindex", wrapArgs("bar", "foobar"), nil, mustCreateException(ValueErrorType, "substring not found")},
		{"rindex", wrapArgs("bar", "a", 1, 10), NewInt(1).ToObject(), nil},
		{"rindex", wrapArgs("bar", "a", NewLong(big.NewInt(1)), 10), NewInt(1).ToObject(), nil},
		{"rindex", wrapArgs("bar", "a", 0, NewLong(big.NewInt(2))), NewInt(1).ToObject(), nil},
		{"rindex", wrapArgs("bar", "a", 1, 3), NewInt(1).ToObject(), nil},
		{"rindex", wrapArgs("bar", "a", 0, -1), NewInt(1).ToObject(), nil},
		{"rindex", wrapArgs("foo", newTestTuple("barfoo", "oo").ToObject()), nil, mustCreateException(TypeErrorType, "'find/index' requires a 'str' object but received a 'tuple'")},
		{"rindex", wrapArgs("foo", 123), nil, mustCreateException(TypeErrorType, "'find/index' requires a 'str' object but received a 'int'")},
		{"rindex", wrapArgs("barbaz", "ba"), NewInt(3).ToObject(), nil},
		{"rindex", wrapArgs("barbaz", "ba", None, 4), NewInt(0).ToObject(), nil},
		{"rjust", wrapArgs("foobar", 10, "#"), NewStr("####foobar").ToObject(), nil},
		{"rjust", wrapArgs("foobar", 3, "#"), NewStr("foobar").ToObject(), nil},
		{"rjust", wrapArgs("foobar", -1, "#"), NewStr("foobar").ToObject(), nil},
		{"rjust", wrapArgs("foobar", 10, "##"), nil, mustCreateException(TypeErrorType, "rjust() argument 2 must be char, not str")},
		{"rjust", wrapArgs("foobar", 10, ""), nil, mustCreateException(TypeErrorType, "rjust() argument 2 must be char, not str")},
		{"split", wrapArgs("foo,bar", ","), newTestList("foo", "bar").ToObject(), nil},
		{"split", wrapArgs("1,2,3", ",", 1), newTestList("1", "2,3").ToObject(), nil},
		{"split", wrapArgs("a \tb\nc"), newTestList("a", "b", "c").ToObject(), nil},
		{"split", wrapArgs("a \tb\nc", None), newTestList("a", "b", "c").ToObject(), nil},
		{"split", wrapArgs("a \tb\nc", None, -1), newTestList("a", "b", "c").ToObject(), nil},
		{"split", wrapArgs("a \tb\nc", None, 1), newTestList("a", "b\nc").ToObject(), nil},
		{"split", wrapArgs("foo", 1), nil, mustCreateException(TypeErrorType, "expected a str separator")},
		{"split", wrapArgs("foo", ""), nil, mustCreateException(ValueErrorType, "empty separator")},
		{"split", wrapArgs(""), newTestList().ToObject(), nil},
		{"split", wrapArgs(" "), newTestList().ToObject(), nil},
		{"split", wrapArgs("", "x"), newTestList("").ToObject(), nil},
		{"split", wrapArgs(" ", " ", 1), newTestList("", "").ToObject(), nil},
		{"split", wrapArgs("aa", "a", 2), newTestList("", "", "").ToObject(), nil},
		{"split", wrapArgs(" a ", "a"), newTestList(" ", " ").ToObject(), nil},
		{"split", wrapArgs("a b c d", None, 1), newTestList("a", "b c d").ToObject(), nil},
		{"split", wrapArgs("a b c d "), newTestList("a", "b", "c", "d").ToObject(), nil},
		{"split", wrapArgs(" a b c d ", None, 1), newTestList("a", "b c d ").ToObject(), nil},
		{"split", wrapArgs("   a b c d ", None, 0), newTestList("a b c d ").ToObject(), nil},
		{"splitlines", wrapArgs(""), NewList().ToObject(), nil},
		{"splitlines", wrapArgs("\n"), newTestList("").ToObject(), nil},
		{"splitlines", wrapArgs("foo"), newTestList("foo").ToObject(), nil},
		{"splitlines", wrapArgs("foo\r"), newTestList("foo").ToObject(), nil},
		{"splitlines", wrapArgs("foo\r", true), newTestList("foo\r").ToObject(), nil},
		{"splitlines", wrapArgs("foo\r\nbar\n", big.NewInt(12)), newTestList("foo\r\n", "bar\n").ToObject(), nil},
		{"splitlines", wrapArgs("foo\n\r\nbar\n\n"), newTestList("foo", "", "bar", "").ToObject(), nil},
		{"splitlines", wrapArgs("foo", newObject(ObjectType)), nil, mustCreateException(TypeErrorType, "an integer is required")},
		{"splitlines", wrapArgs("foo", "bar", "baz"), nil, mustCreateException(TypeErrorType, "'splitlines' of 'str' requires 2 arguments")},
		{"splitlines", wrapArgs("foo", overflowLong), nil, mustCreateException(OverflowErrorType, "Python int too large to convert to a Go int")},
		{"startswith", wrapArgs("", ""), True.ToObject(), nil},
		{"startswith", wrapArgs("", "", 1), False.ToObject(), nil},
		{"startswith", wrapArgs("foobar", "foo"), True.ToObject(), nil},
		{"startswith", wrapArgs("foobar", "foo", 2), False.ToObject(), nil},
		{"startswith", wrapArgs("foobar", "bar", 3), True.ToObject(), nil},
		{"startswith", wrapArgs("foobar", "bar", 3, 5), False.ToObject(), nil},
		{"startswith", wrapArgs("foobar", "bar", 5, 3), False.ToObject(), nil},
		{"startswith", wrapArgs("foo", "foobar"), False.ToObject(), nil},
		{"startswith", wrapArgs("foo", newTestTuple("foobar", "fo").ToObject()), True.ToObject(), nil},
		{"startswith", wrapArgs("foo", 123), nil, mustCreateException(TypeErrorType, "startswith first arg must be str, unicode, or tuple, not int")},
		{"startswith", wrapArgs("foo", "f", "123"), nil, mustCreateException(TypeErrorType, "'startswith' requires a 'int' object but received a 'str'")},
		{"startswith", wrapArgs("foo", newTestTuple(123).ToObject()), nil, mustCreateException(TypeErrorType, "expected a str")},
		{"strip", wrapArgs("foo "), NewStr("foo").ToObject(), nil},
		{"strip", wrapArgs(" foo bar "), NewStr("foo bar").ToObject(), nil},
		{"strip", wrapArgs("foo foo", "o"), NewStr("foo f").ToObject(), nil},
		{"strip", wrapArgs("foo bar", "abr"), NewStr("foo ").ToObject(), nil},
		{"strip", wrapArgs("foo", NewUnicode("o")), NewUnicode("f").ToObject(), nil},
		{"strip", wrapArgs("123", 3), nil, mustCreateException(TypeErrorType, "strip arg must be None, str or unicode")},
		{"strip", wrapArgs("foo", "bar", "baz"), nil, mustCreateException(TypeErrorType, "'strip' of 'str' requires 2 arguments")},
		{"strip", wrapArgs("\xfboo", NewUnicode("o")), nil, mustCreateException(UnicodeDecodeErrorType, "'utf8' codec can't decode byte 0xfb in position 0")},
		{"strip", wrapArgs("foo", NewUnicode("o")), NewUnicode("f").ToObject(), nil},
		{"replace", wrapArgs("one!two!three!", "!", "@", 1), NewStr("one@two!three!").ToObject(), nil},
		{"replace", wrapArgs("one!two!three!", "!", ""), NewStr("onetwothree").ToObject(), nil},
		{"replace", wrapArgs("one!two!three!", "!", "@", 2), NewStr("one@two@three!").ToObject(), nil},
		{"replace", wrapArgs("one!two!three!", "!", "@", 3), NewStr("one@two@three@").ToObject(), nil},
		{"replace", wrapArgs("one!two!three!", "!", "@", 4), NewStr("one@two@three@").ToObject(), nil},
		{"replace", wrapArgs("one!two!three!", "!", "@", 0), NewStr("one!two!three!").ToObject(), nil},
		{"replace", wrapArgs("one!two!three!", "!", "@"), NewStr("one@two@three@").ToObject(), nil},
		{"replace", wrapArgs("one!two!three!", "x", "@"), NewStr("one!two!three!").ToObject(), nil},
		{"replace", wrapArgs("one!two!three!", "x", "@", 2), NewStr("one!two!three!").ToObject(), nil},
		{"replace", wrapArgs("\xd0\xb2\xd0\xbe\xd0\xbb", "", "\x00", -1), NewStr("\x00\xd0\x00\xb2\x00\xd0\x00\xbe\x00\xd0\x00\xbb\x00").ToObject(), nil},
		{"replace", wrapArgs("\xd0\xb2\xd0\xbe\xd0\xbb", "", "\x01\x02", -1), NewStr("\x01\x02\xd0\x01\x02\xb2\x01\x02\xd0\x01\x02\xbe\x01\x02\xd0\x01\x02\xbb\x01\x02").ToObject(), nil},
		{"replace", wrapArgs("abc", "", "-"), NewStr("-a-b-c-").ToObject(), nil},
		{"replace", wrapArgs("abc", "", "-", 3), NewStr("-a-b-c").ToObject(), nil},
		{"replace", wrapArgs("abc", "", "-", 0), NewStr("abc").ToObject(), nil},
		{"replace", wrapArgs("", "", ""), NewStr("").ToObject(), nil},
		{"replace", wrapArgs("", "", "a"), NewStr("a").ToObject(), nil},
		{"replace", wrapArgs("abc", "a", "--", 0), NewStr("abc").ToObject(), nil},
		{"replace", wrapArgs("abc", "xy", "--"), NewStr("abc").ToObject(), nil},
		{"replace", wrapArgs("123", "123", ""), NewStr("").ToObject(), nil},
		{"replace", wrapArgs("123123", "123", ""), NewStr("").ToObject(), nil},
		{"replace", wrapArgs("123x123", "123", ""), NewStr("x").ToObject(), nil},
		{"replace", wrapArgs("one!two!three!", "!", "@", NewLong(big.NewInt(1))), NewStr("one@two!three!").ToObject(), nil},
		{"replace", wrapArgs("foobar", "bar", "baz", newObject(intIntType)), NewStr("foobaz").ToObject(), nil},
		{"replace", wrapArgs("foobar", "bar", "baz", newObject(longIntType)), NewStr("foobaz").ToObject(), nil},
		{"replace", wrapArgs("", "", "x"), NewStr("x").ToObject(), nil},
		{"replace", wrapArgs("", "", "x", -1), NewStr("x").ToObject(), nil},
		{"replace", wrapArgs("", "", "x", 0), NewStr("").ToObject(), nil},
		{"replace", wrapArgs("", "", "x", 1), NewStr("").ToObject(), nil},
		{"replace", wrapArgs("", "", "x", 1000), NewStr("").ToObject(), nil},
		// TODO: Support unicode substring.
		{"replace", wrapArgs("foobar", "", NewUnicode("bar")), nil, mustCreateException(TypeErrorType, "'replace' requires a 'str' object but received a 'unicode'")},
		{"replace", wrapArgs("foobar", NewUnicode("bar"), ""), nil, mustCreateException(TypeErrorType, "'replace' requires a 'str' object but received a 'unicode'")},
		{"replace", wrapArgs("foobar", "bar", "baz", None), nil, mustCreateException(TypeErrorType, "an integer is required")},
		{"replace", wrapArgs("foobar", "bar", "baz", newObject(intIndexType)), nil, mustCreateException(TypeErrorType, "an integer is required")},
		{"replace", wrapArgs("foobar", "bar", "baz", newObject(longIndexType)), nil, mustCreateException(TypeErrorType, "an integer is required")},
		{"rstrip", wrapArgs("foo "), NewStr("foo").ToObject(), nil},
		{"rstrip", wrapArgs(" foo bar "), NewStr(" foo bar").ToObject(), nil},
		{"rstrip", wrapArgs("foo foo", "o"), NewStr("foo f").ToObject(), nil},
		{"rstrip", wrapArgs("foo bar", "abr"), NewStr("foo ").ToObject(), nil},
		{"rstrip", wrapArgs("foo", NewUnicode("o")), NewUnicode("f").ToObject(), nil},
		{"rstrip", wrapArgs("123", 3), nil, mustCreateException(TypeErrorType, "strip arg must be None, str or unicode")},
		{"rstrip", wrapArgs("foo", "bar", "baz"), nil, mustCreateException(TypeErrorType, "'strip' of 'str' requires 2 arguments")},
		{"rstrip", wrapArgs("\xfboo", NewUnicode("o")), nil, mustCreateException(UnicodeDecodeErrorType, "'utf8' codec can't decode byte 0xfb in position 0")},
		{"rstrip", wrapArgs("foo", NewUnicode("o")), NewUnicode("f").ToObject(), nil},
		{"title", wrapArgs(""), NewStr("").ToObject(), nil},
		{"title", wrapArgs("a"), NewStr("A").ToObject(), nil},
		{"title", wrapArgs("A"), NewStr("A").ToObject(), nil},
		{"title", wrapArgs(" a"), NewStr(" A").ToObject(), nil},
		{"title", wrapArgs("abc def"), NewStr("Abc Def").ToObject(), nil},
		{"title", wrapArgs("ABC DEF"), NewStr("Abc Def").ToObject(), nil},
		{"title", wrapArgs("aBC dEF"), NewStr("Abc Def").ToObject(), nil},
		{"title", wrapArgs("abc def", 123), nil, mustCreateException(TypeErrorType, "'title' of 'str' requires 1 arguments")},
		{"title", wrapArgs(123), nil, mustCreateException(TypeErrorType, "unbound method title() must be called with str instance as first argument (got int instance instead)")},
		{"title", wrapArgs("вол"), NewStr("вол").ToObject(), nil},
		{"title", wrapArgs("ВОЛ"), NewStr("ВОЛ").ToObject(), nil},
		{"upper", wrapArgs(""), NewStr("").ToObject(), nil},
		{"upper", wrapArgs("a"), NewStr("A").ToObject(), nil},
		{"upper", wrapArgs("A"), NewStr("A").ToObject(), nil},
		{"upper", wrapArgs(" a"), NewStr(" A").ToObject(), nil},
		{"upper", wrapArgs("abc"), NewStr("ABC").ToObject(), nil},
		{"upper", wrapArgs("ABC"), NewStr("ABC").ToObject(), nil},
		{"upper", wrapArgs("aBC"), NewStr("ABC").ToObject(), nil},
		{"upper", wrapArgs("abc def", 123), nil, mustCreateException(TypeErrorType, "'upper' of 'str' requires 1 arguments")},
		{"upper", wrapArgs(123), nil, mustCreateException(TypeErrorType, "unbound method upper() must be called with str instance as first argument (got int instance instead)")},
		{"upper", wrapArgs("вол"), NewStr("вол").ToObject(), nil},
		{"upper", wrapArgs("ВОЛ"), NewStr("ВОЛ").ToObject(), nil},
		{"zfill", wrapArgs("123", 2), NewStr("123").ToObject(), nil},
		{"zfill", wrapArgs("123", 3), NewStr("123").ToObject(), nil},
		{"zfill", wrapArgs("123", 4), NewStr("0123").ToObject(), nil},
		{"zfill", wrapArgs("+123", 3), NewStr("+123").ToObject(), nil},
		{"zfill", wrapArgs("+123", 4), NewStr("+123").ToObject(), nil},
		{"zfill", wrapArgs("+123", 5), NewStr("+0123").ToObject(), nil},
		{"zfill", wrapArgs("-123", 3), NewStr("-123").ToObject(), nil},
		{"zfill", wrapArgs("-123", 4), NewStr("-123").ToObject(), nil},
		{"zfill", wrapArgs("-123", 5), NewStr("-0123").ToObject(), nil},
		{"zfill", wrapArgs("123", NewLong(big.NewInt(3))), NewStr("123").ToObject(), nil},
		{"zfill", wrapArgs("123", NewLong(big.NewInt(5))), NewStr("00123").ToObject(), nil},
		{"zfill", wrapArgs("", 0), NewStr("").ToObject(), nil},
		{"zfill", wrapArgs("", 1), NewStr("0").ToObject(), nil},
		{"zfill", wrapArgs("", 3), NewStr("000").ToObject(), nil},
		{"zfill", wrapArgs("", -1), NewStr("").ToObject(), nil},
		{"zfill", wrapArgs("34", 1), NewStr("34").ToObject(), nil},
		{"zfill", wrapArgs("34", 4), NewStr("0034").ToObject(), nil},
		{"zfill", wrapArgs("34", None), nil, mustCreateException(TypeErrorType, "an integer is required")},
		{"zfill", wrapArgs("", True), NewStr("0").ToObject(), nil},
		{"zfill", wrapArgs("", False), NewStr("").ToObject(), nil},
		{"zfill", wrapArgs("34", NewStr("test")), nil, mustCreateException(TypeErrorType, "an integer is required")},
		{"zfill", wrapArgs("34"), nil, mustCreateException(TypeErrorType, "'zfill' of 'str' requires 2 arguments")},
		{"swapcase", wrapArgs(""), NewStr("").ToObject(), nil},
		{"swapcase", wrapArgs("a"), NewStr("A").ToObject(), nil},
		{"swapcase", wrapArgs("A"), NewStr("a").ToObject(), nil},
		{"swapcase", wrapArgs(" A"), NewStr(" a").ToObject(), nil},
		{"swapcase", wrapArgs("abc"), NewStr("ABC").ToObject(), nil},
		{"swapcase", wrapArgs("ABC"), NewStr("abc").ToObject(), nil},
		{"swapcase", wrapArgs("aBC"), NewStr("Abc").ToObject(), nil},
		{"swapcase", wrapArgs("abc def", 123), nil, mustCreateException(TypeErrorType, "'swapcase' of 'str' requires 1 arguments")},
		{"swapcase", wrapArgs(123), nil, mustCreateException(TypeErrorType, "unbound method swapcase() must be called with str instance as first argument (got int instance instead)")},
		{"swapcase", wrapArgs("вол"), NewStr("вол").ToObject(), nil},
		{"swapcase", wrapArgs("ВОЛ"), NewStr("ВОЛ").ToObject(), nil},
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

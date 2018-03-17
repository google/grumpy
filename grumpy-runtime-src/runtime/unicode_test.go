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
	"reflect"
	"testing"
	"unicode"
)

func TestUnicodeNewUnicode(t *testing.T) {
	cases := []struct {
		s    string
		want []rune
	}{
		// Invalid utf-8 characters should not be present in unicode
		// objects, but if that happens they're substituted with the
		// replacement character U+FFFD.
		{"foo\xffbar", []rune{'f', 'o', 'o', '\uFFFD', 'b', 'a', 'r'}},
		// U+D800 is a surrogate that Python 2.x encodes to UTF-8 as
		// \xed\xa0\x80 but Go treats each code unit as a bad rune.
		{"\xed\xa0\x80", []rune{'\uFFFD', '\uFFFD', '\uFFFD'}},
	}
	for _, cas := range cases {
		got := NewUnicode(cas.s).Value()
		if !reflect.DeepEqual(got, cas.want) {
			t.Errorf("NewUnicode(%q) = %v, want %v", cas.s, got, cas.want)
		}
	}
}

func TestUnicodeBinaryOps(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, fn func(f *Frame, v, w *Object) (*Object, *BaseException), v, w *Object) (*Object, *BaseException) {
		return fn(f, v, w)
	})
	cases := []invokeTestCase{
		{args: wrapArgs(Add, NewUnicode("foo"), NewUnicode("bar")), want: NewUnicode("foobar").ToObject()},
		{args: wrapArgs(Add, NewUnicode("foo"), "bar"), want: NewUnicode("foobar").ToObject()},
		{args: wrapArgs(Add, "foo", NewUnicode("bar")), want: NewUnicode("foobar").ToObject()},
		{args: wrapArgs(Add, NewUnicode("baz"), NewUnicode("")), want: NewUnicode("baz").ToObject()},
		{args: wrapArgs(Add, NewUnicode(""), newObject(ObjectType)), wantExc: mustCreateException(TypeErrorType, "coercing to Unicode: need string, object found")},
		{args: wrapArgs(Add, None, NewUnicode("")), wantExc: mustCreateException(TypeErrorType, "unsupported operand type(s) for +: 'NoneType' and 'unicode'")},
		{args: wrapArgs(Mul, NewUnicode(""), 10), want: NewUnicode("").ToObject()},
		{args: wrapArgs(Mul, NewUnicode("foo"), -2), want: NewUnicode("").ToObject()},
		{args: wrapArgs(Mul, NewUnicode("foobar"), 0), want: NewUnicode("").ToObject()},
		{args: wrapArgs(Mul, NewUnicode("aloha"), 2), want: NewUnicode("alohaaloha").ToObject()},
		{args: wrapArgs(Mul, 1, NewUnicode("baz")), want: NewUnicode("baz").ToObject()},
		{args: wrapArgs(Mul, newObject(ObjectType), NewUnicode("qux")), wantExc: mustCreateException(TypeErrorType, "unsupported operand type(s) for *: 'object' and 'unicode'")},
		{args: wrapArgs(Mul, NewUnicode("foo"), NewUnicode("")), wantExc: mustCreateException(TypeErrorType, "unsupported operand type(s) for *: 'unicode' and 'unicode'")},
		{args: wrapArgs(Mul, NewUnicode("bar"), MaxInt), wantExc: mustCreateException(OverflowErrorType, "result too large")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestUnicodeCompare(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(NewUnicode(""), NewUnicode("")), want: compareAllResultEq},
		{args: wrapArgs(NewUnicode(""), ""), want: compareAllResultEq},
		{args: wrapArgs(NewStr(""), NewUnicode("")), want: compareAllResultEq},
		{args: wrapArgs(NewUnicode("樂"), NewUnicode("樂")), want: compareAllResultEq},
		{args: wrapArgs(NewUnicode("樂"), "樂"), want: compareAllResultEq},
		{args: wrapArgs(NewStr("樂"), NewUnicode("樂")), want: compareAllResultEq},
		{args: wrapArgs(NewUnicode("вол"), NewUnicode("волн")), want: compareAllResultLT},
		{args: wrapArgs(NewUnicode("вол"), "волн"), want: compareAllResultLT},
		{args: wrapArgs(NewStr("вол"), NewUnicode("волн")), want: compareAllResultLT},
		{args: wrapArgs(NewUnicode("bar"), NewUnicode("baz")), want: compareAllResultLT},
		{args: wrapArgs(NewUnicode("bar"), "baz"), want: compareAllResultLT},
		{args: wrapArgs(NewStr("bar"), NewUnicode("baz")), want: compareAllResultLT},
		{args: wrapArgs(NewUnicode("abc"), None), want: compareAllResultGT},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(compareAll, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestUnicodeContains(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(NewUnicode("foobar"), NewUnicode("foo")), want: True.ToObject()},
		{args: wrapArgs(NewUnicode("abcdef"), NewUnicode("bar")), want: False.ToObject()},
		{args: wrapArgs(NewUnicode(""), NewUnicode("")), want: True.ToObject()},
		{args: wrapArgs(NewUnicode(""), 102.1), wantExc: mustCreateException(TypeErrorType, "coercing to Unicode: need string, float found")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(UnicodeType, "__contains__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestUnicodeEncode(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(NewUnicode("foo")), want: NewStr("foo").ToObject()},
		{args: wrapArgs(NewUnicode("foob\u0300ar"), "utf8"), want: NewStr("foob\u0300ar").ToObject()},
		{args: wrapArgs(NewUnicode("foo"), "noexist", "strict"), wantExc: mustCreateException(LookupErrorType, "unknown encoding: noexist")},
		{args: wrapArgs(NewUnicodeFromRunes([]rune{'в', 'о', 'л', 'н'}), "utf8", "strict"), want: NewStr("\xd0\xb2\xd0\xbe\xd0\xbb\xd0\xbd").ToObject()},
		{args: wrapArgs(NewUnicodeFromRunes([]rune{'\xff'}), "utf8"), want: NewStr("\xc3\xbf").ToObject()},
		{args: wrapArgs(NewUnicodeFromRunes([]rune{0xD800})), wantExc: mustCreateException(UnicodeEncodeErrorType, `'utf8' codec can't encode character \ud800 in position 0`)},
		{args: wrapArgs(NewUnicodeFromRunes([]rune{unicode.MaxRune + 1}), "utf8", "replace"), want: NewStr("\xef\xbf\xbd").ToObject()},
		{args: wrapArgs(NewUnicodeFromRunes([]rune{0xFFFFFF}), "utf8", "ignore"), want: NewStr("").ToObject()},
		{args: wrapArgs(NewUnicodeFromRunes([]rune{0xFFFFFF}), "utf8", "noexist"), wantExc: mustCreateException(LookupErrorType, "unknown error handler name 'noexist'")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(UnicodeType, "encode", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestUnicodeGetItem(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(NewUnicode("bar"), 1), want: NewUnicode("a").ToObject()},
		{args: wrapArgs(NewUnicode("foo"), 3.14), wantExc: mustCreateException(TypeErrorType, "unicode indices must be integers or slice, not float")},
		{args: wrapArgs(NewUnicode("baz"), -1), want: NewUnicode("z").ToObject()},
		{args: wrapArgs(NewUnicode("baz"), -4), wantExc: mustCreateException(IndexErrorType, "index out of range")},
		{args: wrapArgs(NewUnicode(""), 0), wantExc: mustCreateException(IndexErrorType, "index out of range")},
		{args: wrapArgs(NewUnicode("foo"), 3), wantExc: mustCreateException(IndexErrorType, "index out of range")},
		{args: wrapArgs(NewUnicode("bar"), newTestSlice(None, 2)), want: NewStr("ba").ToObject()},
		{args: wrapArgs(NewUnicode("bar"), newTestSlice(1, 3)), want: NewStr("ar").ToObject()},
		{args: wrapArgs(NewUnicode("bar"), newTestSlice(1, None)), want: NewStr("ar").ToObject()},
		{args: wrapArgs(NewUnicode("foobarbaz"), newTestSlice(1, 8, 2)), want: NewStr("obra").ToObject()},
		{args: wrapArgs(NewUnicode("bar"), newTestSlice(1, 2, 0)), wantExc: mustCreateException(ValueErrorType, "slice step cannot be zero")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(UnicodeType, "__getitem__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestUnicodeHash(t *testing.T) {
	truncateInt := func(i int64) int { return int(i) } // Support for 32bit systems.
	cases := []invokeTestCase{
		{args: wrapArgs(NewUnicode("foo")), want: NewInt(truncateInt(-4177197833195190597)).ToObject()},
		{args: wrapArgs(NewUnicode("bar")), want: NewInt(truncateInt(327024216814240868)).ToObject()},
		{args: wrapArgs(NewUnicode("baz")), want: NewInt(truncateInt(327024216814240876)).ToObject()},
		{args: wrapArgs(NewUnicode("")), want: NewInt(0).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(UnicodeType, "__hash__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestUnicodeLen(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(NewUnicode("foo")), want: NewInt(3).ToObject()},
		{args: wrapArgs(NewUnicode("")), want: NewInt(0).ToObject()},
		{args: wrapArgs(NewUnicode("волн")), want: NewInt(4).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(UnicodeType, "__len__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestUnicodeMethods(t *testing.T) {
	cases := []struct {
		methodName string
		args       Args
		want       *Object
		wantExc    *BaseException
	}{
		{"join", wrapArgs(NewUnicode(","), newTestList("foo", "bar")), NewUnicode("foo,bar").ToObject(), nil},
		{"join", wrapArgs(NewUnicode(":"), newTestList(NewUnicode("foo"), "bar", NewUnicode("baz"))), NewUnicode("foo:bar:baz").ToObject(), nil},
		{"join", wrapArgs(NewUnicode("nope"), NewTuple()), NewUnicode("").ToObject(), nil},
		{"join", wrapArgs(NewUnicode("nope"), newTestTuple(NewUnicode("foo"))), NewUnicode("foo").ToObject(), nil},
		{"join", wrapArgs(NewUnicode(","), newTestList("foo", "bar", 3.14)), nil, mustCreateException(TypeErrorType, "coercing to Unicode: need string, float found")},
		{"strip", wrapArgs(NewUnicode("foo ")), NewStr("foo").ToObject(), nil},
		{"strip", wrapArgs(NewUnicode(" foo bar ")), NewStr("foo bar").ToObject(), nil},
		{"strip", wrapArgs(NewUnicode("foo foo"), "o"), NewStr("foo f").ToObject(), nil},
		{"strip", wrapArgs(NewUnicode("foo bar"), "abr"), NewStr("foo ").ToObject(), nil},
		{"strip", wrapArgs(NewUnicode("foo"), NewUnicode("o")), NewUnicode("f").ToObject(), nil},
		{"strip", wrapArgs(NewUnicode("123"), 3), nil, mustCreateException(TypeErrorType, "coercing to Unicode: need string, int found")},
		{"strip", wrapArgs(NewUnicode("foo"), "bar", "baz"), nil, mustCreateException(TypeErrorType, "'strip' of 'unicode' requires 2 arguments")},
		{"strip", wrapArgs(NewUnicode("foo"), NewUnicode("o")), NewUnicode("f").ToObject(), nil},
	}
	for _, cas := range cases {
		testCase := invokeTestCase{args: cas.args, want: cas.want, wantExc: cas.wantExc}
		if err := runInvokeMethodTestCase(UnicodeType, cas.methodName, &testCase); err != "" {
			t.Error(err)
		}
	}
}
func TestUnicodeNative(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, s *Unicode) (string, *BaseException) {
		native, raised := ToNative(f, s.ToObject())
		if raised != nil {
			return "", raised
		}
		got, ok := native.Interface().(string)
		if raised := Assert(f, GetBool(ok).ToObject(), nil); raised != nil {
			return "", raised
		}
		return got, nil
	})
	cases := []invokeTestCase{
		{args: wrapArgs(NewUnicode("волн")), want: NewStr("волн").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestUnicodeRepr(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(NewUnicode("foo")), want: NewStr("u'foo'").ToObject()},
		{args: wrapArgs(NewUnicode("on\nmultiple\nlines")), want: NewStr(`u'on\nmultiple\nlines'`).ToObject()},
		{args: wrapArgs(NewUnicode("a\u0300")), want: NewStr(`u'a\u0300'`).ToObject()},
		{args: wrapArgs(NewUnicodeFromRunes([]rune{'h', 'o', 'l', 0xFF})), want: NewStr(`u'hol\xff'`).ToObject()},
		{args: wrapArgs(NewUnicodeFromRunes([]rune{0x10163})), want: NewStr(`u'\U00010163'`).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(UnicodeType, "__repr__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestUnicodeNew(t *testing.T) {
	fooType := newTestClass("Foo", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__unicode__": newBuiltinFunction("__unicode__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return NewStr("foo").ToObject(), nil
		}).ToObject(),
	}))
	strictEqType := newTestClassStrictEq("StrictEq", UnicodeType)
	cases := []invokeTestCase{
		{args: wrapArgs(UnicodeType), want: NewUnicode("").ToObject()},
		{args: wrapArgs(UnicodeType, NewUnicode("foo")), want: NewUnicode("foo").ToObject()},
		{args: wrapArgs(UnicodeType, newObject(fooType)), want: NewUnicode("foo").ToObject()},
		{args: wrapArgs(UnicodeType, "foobar"), want: NewUnicode("foobar").ToObject()},
		{args: wrapArgs(UnicodeType, "foo\xffbar"), wantExc: mustCreateException(UnicodeDecodeErrorType, "'utf8' codec can't decode byte 0xff in position 3")},
		{args: wrapArgs(UnicodeType, 123), want: NewUnicode("123").ToObject()},
		{args: wrapArgs(UnicodeType, 3.14, "utf8"), wantExc: mustCreateException(TypeErrorType, "coercing to Unicode: need str, float found")},
		{args: wrapArgs(UnicodeType, "baz", "utf8"), want: NewUnicode("baz").ToObject()},
		{args: wrapArgs(UnicodeType, "baz", "utf-8"), want: NewUnicode("baz").ToObject()},
		{args: wrapArgs(UnicodeType, "foo\xffbar", "utf_8"), wantExc: mustCreateException(UnicodeDecodeErrorType, "'utf_8' codec can't decode byte 0xff in position 3")},
		{args: wrapArgs(UnicodeType, "foo\xffbar", "UTF8", "ignore"), want: NewUnicode("foobar").ToObject()},
		{args: wrapArgs(UnicodeType, "foo\xffbar", "utf8", "replace"), want: NewUnicode("foo\ufffdbar").ToObject()},
		{args: wrapArgs(UnicodeType, "\xff", "utf-8", "noexist"), wantExc: mustCreateException(LookupErrorType, "unknown error handler name 'noexist'")},
		{args: wrapArgs(UnicodeType, "\xff", "utf16"), wantExc: mustCreateException(LookupErrorType, "unknown encoding: utf16")},
		{args: wrapArgs(strictEqType, NewUnicode("foo")), want: (&Unicode{Object{typ: strictEqType}, bytes.Runes([]byte("foo"))}).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(UnicodeType, "__new__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestUnicodeNewNotSubtype(t *testing.T) {
	cas := invokeTestCase{args: wrapArgs(IntType), wantExc: mustCreateException(TypeErrorType, "unicode.__new__(int): int is not a subtype of unicode")}
	if err := runInvokeMethodTestCase(UnicodeType, "__new__", &cas); err != "" {
		t.Error(err)
	}
}

func TestUnicodeNewSubclass(t *testing.T) {
	fooType := newTestClass("Foo", []*Type{UnicodeType}, NewDict())
	bar := (&Unicode{Object{typ: fooType}, bytes.Runes([]byte("bar"))}).ToObject()
	fun := wrapFuncForTest(func(f *Frame) *BaseException {
		got, raised := UnicodeType.Call(f, []*Object{bar}, nil)
		if raised != nil {
			return raised
		}
		if got.typ != UnicodeType {
			t.Errorf(`unicode(Foo("bar")) = %v, want u"bar"`, got)
			return nil
		}
		ne, raised := NE(f, got, NewUnicode("bar").ToObject())
		if raised != nil {
			return raised
		}
		isTrue, raised := IsTrue(f, ne)
		if raised != nil {
			return raised
		}
		if isTrue {
			t.Errorf(`unicode(Foo("bar")) = %v, want u"bar"`, got)
		}
		return nil
	})
	if err := runInvokeTestCase(fun, &invokeTestCase{want: None}); err != "" {
		t.Error(err)
	}
}

func TestUnicodeStr(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(NewUnicode("foo")), want: NewStr("foo").ToObject()},
		{args: wrapArgs(NewUnicode("on\nmultiple\nlines")), want: NewStr("on\nmultiple\nlines").ToObject()},
		{args: wrapArgs(NewUnicode("a\u0300")), want: NewStr("a\u0300").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(UnicodeType, "__str__", &cas); err != "" {
			t.Error(err)
		}
	}
}

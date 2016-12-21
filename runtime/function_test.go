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
	"regexp"
	"testing"
)

func TestNewFunctionKeywordsCheck(t *testing.T) {
	oldLogFatal := logFatal
	defer func() { logFatal = oldLogFatal }()
	var got string
	logFatal = func(msg string) {
		got = msg
	}
	NewFunction("foo", []FunctionArg{{"bar", None}, {"baz", nil}}, "", "", nil)
	if want := "foo() non-keyword arg baz after keyword arg"; got != want {
		t.Errorf("NewFunction logged %q, want %q", got, want)
	}
}

func TestNewFunction(t *testing.T) {
	fn := func(f *Frame, args []*Object) (*Object, *BaseException) {
		return NewTuple(Args(args).makeCopy()...).ToObject(), nil
	}
	cases := []struct {
		fun *Function
		invokeTestCase
	}{
		{
			NewFunction("f1", nil, "", "", fn),
			invokeTestCase{want: NewTuple().ToObject()},
		},
		{
			NewFunction("f2", []FunctionArg{{"a", nil}}, "", "", fn),
			invokeTestCase{args: wrapArgs(123), want: newTestTuple(123).ToObject()},
		},
		{
			NewFunction("f2", []FunctionArg{{"a", nil}}, "", "", fn),
			invokeTestCase{kwargs: wrapKWArgs("a", "apple"), want: newTestTuple("apple").ToObject()},
		},
		{
			NewFunction("f2", []FunctionArg{{"a", nil}}, "", "", fn),
			invokeTestCase{kwargs: wrapKWArgs("b", "bear"), wantExc: mustCreateException(TypeErrorType, "f2() got an unexpected keyword argument 'b'")},
		},
		{
			NewFunction("f2", []FunctionArg{{"a", nil}}, "", "", fn),
			invokeTestCase{wantExc: mustCreateException(TypeErrorType, "f2() takes at least 1 arguments (0 given)")},
		},
		{
			NewFunction("f2", []FunctionArg{{"a", nil}}, "", "", fn),
			invokeTestCase{args: wrapArgs(1, 2, 3), wantExc: mustCreateException(TypeErrorType, "f2() takes 1 arguments (3 given)")},
		},
		{
			NewFunction("f3", []FunctionArg{{"a", nil}, {"b", nil}}, "", "", fn),
			invokeTestCase{args: wrapArgs(1, 2), want: newTestTuple(1, 2).ToObject()},
		},
		{
			NewFunction("f3", []FunctionArg{{"a", nil}, {"b", nil}}, "", "", fn),
			invokeTestCase{args: wrapArgs(1), kwargs: wrapKWArgs("b", "bear"), want: newTestTuple(1, "bear").ToObject()},
		},
		{
			NewFunction("f3", []FunctionArg{{"a", nil}, {"b", nil}}, "", "", fn),
			invokeTestCase{kwargs: wrapKWArgs("b", "bear", "a", "apple"), want: newTestTuple("apple", "bear").ToObject()},
		},
		{
			NewFunction("f3", []FunctionArg{{"a", nil}, {"b", nil}}, "", "", fn),
			invokeTestCase{args: wrapArgs(1), kwargs: wrapKWArgs("a", "alpha"), wantExc: mustCreateException(TypeErrorType, "f3() got multiple values for keyword argument 'a'")},
		},
		{
			NewFunction("f4", []FunctionArg{{"a", nil}, {"b", None}}, "", "", fn),
			invokeTestCase{args: wrapArgs(123), want: newTestTuple(123, None).ToObject()},
		},
		{
			NewFunction("f4", []FunctionArg{{"a", nil}, {"b", None}}, "", "", fn),
			invokeTestCase{args: wrapArgs(123, "bar"), want: newTestTuple(123, "bar").ToObject()},
		},
		{
			NewFunction("f4", []FunctionArg{{"a", nil}, {"b", None}}, "", "", fn),
			invokeTestCase{kwargs: wrapKWArgs("a", 123, "b", "bar"), want: newTestTuple(123, "bar").ToObject()},
		},
		{
			NewFunction("f5", []FunctionArg{{"a", nil}}, "args", "", fn),
			invokeTestCase{args: wrapArgs(1), want: newTestTuple(1, NewTuple()).ToObject()},
		},
		{
			NewFunction("f5", []FunctionArg{{"a", nil}}, "args", "", fn),
			invokeTestCase{args: wrapArgs(1, 2, 3), want: newTestTuple(1, newTestTuple(2, 3)).ToObject()},
		},
		{
			NewFunction("f6", []FunctionArg{{"a", nil}}, "", "kwargs", fn),
			invokeTestCase{args: wrapArgs("bar"), want: newTestTuple("bar", NewDict()).ToObject()},
		},
		{
			NewFunction("f6", []FunctionArg{{"a", nil}}, "", "kwargs", fn),
			invokeTestCase{kwargs: wrapKWArgs("a", "apple", "b", "bear"), want: newTestTuple("apple", newTestDict("b", "bear")).ToObject()},
		},
		{
			NewFunction("f6", []FunctionArg{{"a", nil}}, "", "kwargs", fn),
			invokeTestCase{args: wrapArgs("bar"), kwargs: wrapKWArgs("b", "baz", "c", "qux"), want: newTestTuple("bar", newTestDict("b", "baz", "c", "qux")).ToObject()},
		},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(cas.fun.ToObject(), &cas.invokeTestCase); err != "" {
			t.Error(err)
		}
	}
}

func TestFunctionCall(t *testing.T) {
	fun := newBuiltinFunction("fun", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		return newTestTuple(NewTuple(args.makeCopy()...), kwargs.makeDict()).ToObject(), nil
	}).ToObject()
	cases := []invokeTestCase{
		{args: wrapArgs(fun, 123, "abc"), kwargs: wrapKWArgs("b", "bear"), want: newTestTuple(newTestTuple(123, "abc"), newTestDict("b", "bear")).ToObject()},
		{wantExc: mustCreateException(TypeErrorType, "unbound method __call__() must be called with function instance as first argument (got nothing instead)")},
		{args: wrapArgs(newObject(ObjectType)), wantExc: mustCreateException(TypeErrorType, "unbound method __call__() must be called with function instance as first argument (got object instance instead)")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(FunctionType, "__call__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestFunctionGet(t *testing.T) {
	appendMethod := mustNotRaise(GetAttr(newFrame(nil), NewList().ToObject(), NewStr("append"), nil))
	if !appendMethod.isInstance(MethodType) {
		t.Errorf("list.append = %v, want instancemethod", appendMethod)
	}
}

func TestFunctionName(t *testing.T) {
	fun := newBuiltinFunction("TestFunctionName", func(f *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
		foo := newBuiltinFunction("foo", func(*Frame, Args, KWArgs) (*Object, *BaseException) { return None, nil })
		return GetAttr(f, foo.ToObject(), NewStr("__name__"), nil)
	}).ToObject()
	if err := runInvokeTestCase(fun, &invokeTestCase{want: NewStr("foo").ToObject()}); err != "" {
		t.Error(err)
	}
}

func TestFunctionStrRepr(t *testing.T) {
	fn := func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) { return nil, nil }
	cases := []struct {
		o           *Object
		wantPattern string
	}{
		{newBuiltinFunction("foo", fn).ToObject(), `^<function foo at \w+>$`},
		{newBuiltinFunction("some big function name", fn).ToObject(), `^<function some big function name at \w+>$`},
	}
	for _, cas := range cases {
		fun := wrapFuncForTest(func(f *Frame) *BaseException {
			re := regexp.MustCompile(cas.wantPattern)
			s, raised := ToStr(f, cas.o)
			if raised != nil {
				return raised
			}
			if !re.MatchString(s.Value()) {
				t.Errorf("str(%v) = %v, want %q", cas.o, s, re)
			}
			s, raised = Repr(f, cas.o)
			if raised != nil {
				return raised
			}
			if !re.MatchString(s.Value()) {
				t.Errorf("repr(%v) = %v, want %q", cas.o, s, re)
			}
			return nil
		})
		if err := runInvokeTestCase(fun, &invokeTestCase{want: None}); err != "" {
			t.Error(err)
		}
	}
}

func TestStaticMethodGet(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(newStaticMethod(NewStr("abc").ToObject()), 123, IntType), want: NewStr("abc").ToObject()},
		{args: wrapArgs(newStaticMethod(nil), 123, IntType), wantExc: mustCreateException(RuntimeErrorType, "uninitialized staticmethod object")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(StaticMethodType, "__get__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestStaticMethodInit(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, args ...*Object) (*Object, *BaseException) {
		m, raised := StaticMethodType.Call(f, args, nil)
		if raised != nil {
			return nil, raised
		}
		get, raised := GetAttr(f, m, NewStr("__get__"), nil)
		if raised != nil {
			return nil, raised
		}
		return get.Call(f, wrapArgs(123, IntType), nil)
	})
	cases := []invokeTestCase{
		{args: wrapArgs(3.14), want: NewFloat(3.14).ToObject()},
		{wantExc: mustCreateException(TypeErrorType, "'__init__' requires 1 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

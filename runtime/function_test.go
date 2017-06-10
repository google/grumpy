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

func TestFunctionCall(t *testing.T) {
	foo := newBuiltinFunction("foo", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		return newTestTuple(NewTuple(args.makeCopy()...), kwargs.makeDict()).ToObject(), nil
	}).ToObject()
	bar := NewFunction(NewCode("bar", "bar.py", nil, CodeFlagVarArg, func(f *Frame, args []*Object) (*Object, *BaseException) {
		return args[0], nil
	}), nil)
	cases := []invokeTestCase{
		{args: wrapArgs(foo, 123, "abc"), kwargs: wrapKWArgs("b", "bear"), want: newTestTuple(newTestTuple(123, "abc"), newTestDict("b", "bear")).ToObject()},
		{args: wrapArgs(bar, "bar", "baz"), want: newTestTuple("bar", "baz").ToObject()},
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
	appendMethod := mustNotRaise(GetAttr(NewRootFrame(), NewList().ToObject(), NewStr("append"), nil))
	if !appendMethod.isInstance(MethodType) {
		t.Errorf("list.append = %v, want instancemethod", appendMethod)
	}
}

func TestFunctionName(t *testing.T) {
	fun := newBuiltinFunction("TestFunctionName", func(f *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
		foo := newBuiltinFunction("foo", func(*Frame, Args, KWArgs) (*Object, *BaseException) { return None, nil })
		return GetAttr(f, foo.ToObject(), internedName, nil)
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

func TestClassMethodGet(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, meth *classMethod, args ...*Object) (*Object, *BaseException) {
		get, raised := GetAttr(f, meth.ToObject(), NewStr("__get__"), nil)
		if raised != nil {
			return nil, raised
		}
		callable, raised := get.Call(f, args, nil)
		if raised != nil {
			return nil, raised
		}
		return callable.Call(f, nil, nil)
	})
	echoFunc := wrapFuncForTest(func(f *Frame, args ...*Object) *Tuple {
		return NewTuple(args...)
	})
	cases := []invokeTestCase{
		{args: wrapArgs(newClassMethod(echoFunc), ObjectType, ObjectType), want: NewTuple(ObjectType.ToObject()).ToObject()},
		{args: wrapArgs(newClassMethod(NewStr("abc").ToObject()), 123, IntType), wantExc: mustCreateException(TypeErrorType, "first argument must be callable")},
		{args: wrapArgs(newClassMethod(nil), 123, IntType), wantExc: mustCreateException(RuntimeErrorType, "uninitialized classmethod object")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestClassMethodInit(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, args ...*Object) (*Object, *BaseException) {
		m, raised := ClassMethodType.Call(f, args, nil)
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
		// {args: wrapArgs(3.14), want: NewFloat(3.14).ToObject()},
		{wantExc: mustCreateException(TypeErrorType, "'__init__' requires 1 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

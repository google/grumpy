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
	"testing"
)

func TestMethodCall(t *testing.T) {
	foo := newBuiltinFunction("foo", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		return NewTuple(args.makeCopy()...).ToObject(), nil
	}).ToObject()
	self := newObject(ObjectType)
	arg0 := NewInt(123).ToObject()
	arg1 := NewStr("abc").ToObject()
	cases := []invokeTestCase{
		{args: wrapArgs(newTestMethod(foo, self, ObjectType.ToObject())), want: NewTuple(self).ToObject()},
		{args: wrapArgs(newTestMethod(foo, None, ObjectType.ToObject()), self), want: NewTuple(self).ToObject()},
		{args: wrapArgs(newTestMethod(foo, self, ObjectType.ToObject()), arg0, arg1), want: NewTuple(self, arg0, arg1).ToObject()},
		{args: wrapArgs(newTestMethod(foo, None, ObjectType.ToObject()), self, arg0, arg1), want: NewTuple(self, arg0, arg1).ToObject()},
		{args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, "unbound method __call__() must be called with instancemethod instance as first argument (got nothing instead)")},
		{args: wrapArgs(newObject(ObjectType)), wantExc: mustCreateException(TypeErrorType, "unbound method __call__() must be called with instancemethod instance as first argument (got object instance instead)")},
		{args: wrapArgs(newTestMethod(foo, None, IntType.ToObject()), newObject(ObjectType)), wantExc: mustCreateException(TypeErrorType, "unbound method foo() must be called with int instance as first argument (got object instance instead)")},
		{args: wrapArgs(newTestMethod(foo, None, IntType.ToObject())), wantExc: mustCreateException(TypeErrorType, "unbound method foo() must be called with int instance as first argument (got nothing instead)")},
		{args: wrapArgs(newTestMethod(foo, None, None), None), wantExc: mustCreateException(TypeErrorType, "classinfo must be a type or tuple of types")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(MethodType, "__call__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestMethodNew(t *testing.T) {
	cases := []invokeTestCase{
		{wantExc: mustCreateException(TypeErrorType, "'__new__' requires 3 arguments")},
		{args: Args{None, None, None}, wantExc: mustCreateException(TypeErrorType, "first argument must be callable")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(MethodType.ToObject(), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestMethodStrRepr(t *testing.T) {
	foo := newBuiltinFunction("foo", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) { return None, nil }).ToObject()
	cases := []invokeTestCase{
		{args: wrapArgs(newTestMethod(foo, None, StrType.ToObject())), want: NewStr("<unbound method str.foo>").ToObject()},
		{args: wrapArgs(newTestMethod(foo, NewStr("wut").ToObject(), StrType.ToObject())), want: NewStr("<bound method str.foo of 'wut'>").ToObject()},
		{args: wrapArgs(newTestMethod(foo, NewInt(123).ToObject(), TupleType.ToObject())), want: NewStr("<bound method tuple.foo of 123>").ToObject()},
		{args: wrapArgs(newTestMethod(foo, None, None)), want: NewStr("<unbound method ?.foo>").ToObject()},
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

func newTestMethod(function, self, class *Object) *Method {
	return toMethodUnsafe(mustNotRaise(MethodType.Call(NewRootFrame(), Args{function, self, class}, nil)))
}

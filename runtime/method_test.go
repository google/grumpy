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
	})
	self := newObject(ObjectType)
	arg0 := NewInt(123).ToObject()
	arg1 := NewStr("abc").ToObject()
	cases := []invokeTestCase{
		{args: wrapArgs(NewMethod(foo, self, ObjectType)), want: NewTuple(self).ToObject()},
		{args: wrapArgs(NewMethod(foo, None, ObjectType), self), want: NewTuple(self).ToObject()},
		{args: wrapArgs(NewMethod(foo, self, ObjectType), arg0, arg1), want: NewTuple(self, arg0, arg1).ToObject()},
		{args: wrapArgs(NewMethod(foo, None, ObjectType), self, arg0, arg1), want: NewTuple(self, arg0, arg1).ToObject()},
		{args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, "unbound method __call__() must be called with instancemethod instance as first argument (got nothing instead)")},
		{args: wrapArgs(newObject(ObjectType)), wantExc: mustCreateException(TypeErrorType, "unbound method __call__() must be called with instancemethod instance as first argument (got object instance instead)")},
		{args: wrapArgs(NewMethod(foo, None, IntType), newObject(ObjectType)), wantExc: mustCreateException(TypeErrorType, "unbound method foo() must be called with int instance as first argument (got object instance instead)")},
		{args: wrapArgs(NewMethod(foo, None, IntType)), wantExc: mustCreateException(TypeErrorType, "unbound method foo() must be called with int instance as first argument (got nothing instead)")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(MethodType, "__call__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestMethodStrRepr(t *testing.T) {
	foo := newBuiltinFunction("foo", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) { return None, nil })
	cases := []invokeTestCase{
		{args: wrapArgs(NewMethod(foo, None, StrType)), want: NewStr("<unbound method str.foo>").ToObject()},
		{args: wrapArgs(NewMethod(foo, NewStr("wut").ToObject(), StrType)), want: NewStr("<bound method str.foo of 'wut'>").ToObject()},
		{args: wrapArgs(NewMethod(foo, NewInt(123).ToObject(), TupleType)), want: NewStr("<bound method tuple.foo of 123>").ToObject()},
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

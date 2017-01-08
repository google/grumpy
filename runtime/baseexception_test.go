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

func TestBaseExceptionCreate(t *testing.T) {
	emptyExc := toBaseExceptionUnsafe(newObject(ExceptionType))
	emptyExc.args = NewTuple()
	cases := []struct {
		t       *Type
		args    *Tuple
		wantRet *Object
	}{
		{ExceptionType, NewTuple(), emptyExc.ToObject()},
		{TypeErrorType, NewTuple(NewStr("abc").ToObject()), mustCreateException(TypeErrorType, "abc").ToObject()},
	}
	for _, cas := range cases {
		got, match := checkInvokeResult(cas.t.ToObject(), cas.args.elems, cas.wantRet, nil)
		if match == checkInvokeResultExceptionMismatch {
			t.Errorf("%s%v raised %v, want none", cas.t.Name(), cas.args, got)
		} else if match == checkInvokeResultReturnValueMismatch {
			t.Errorf("%s%v = %v, want %v", cas.t.Name(), cas.args, got, cas.wantRet)
		}
	}
}

func TestBaseExceptionInitRaise(t *testing.T) {
	cas := invokeTestCase{
		args:    nil,
		wantExc: mustCreateException(TypeErrorType, "unbound method __init__() must be called with BaseException instance as first argument (got nothing instead)"),
	}
	if err := runInvokeMethodTestCase(BaseExceptionType, "__init__", &cas); err != "" {
		t.Error(err)
	}
}

func TestBaseExceptionRepr(t *testing.T) {
	fooExc := toBaseExceptionUnsafe(newObject(ExceptionType))
	fooExc.args = NewTuple(NewStr("foo").ToObject())
	recursiveExc := toBaseExceptionUnsafe(newObject(ExceptionType))
	recursiveExc.args = NewTuple(recursiveExc.ToObject())
	cases := []invokeTestCase{
		{args: wrapArgs(newObject(TypeErrorType)), want: NewStr("TypeError()").ToObject()},
		{args: wrapArgs(fooExc), want: NewStr("Exception('foo',)").ToObject()},
		{args: wrapArgs(recursiveExc), want: NewStr("Exception(Exception(...),)").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(Repr), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestBaseExceptionStr(t *testing.T) {
	f := NewRootFrame()
	cases := []invokeTestCase{
		{args: wrapArgs(newObject(TypeErrorType)), want: NewStr("").ToObject()},
		{args: wrapArgs(mustNotRaise(ExceptionType.Call(f, wrapArgs(""), nil))), want: NewStr("").ToObject()},
		{args: wrapArgs(mustNotRaise(ExceptionType.Call(f, wrapArgs("foo"), nil))), want: NewStr("foo").ToObject()},
		{args: wrapArgs(mustNotRaise(TypeErrorType.Call(f, wrapArgs(NewTuple(), 3), nil))), want: NewStr("((), 3)").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(ToStr), &cas); err != "" {
			t.Error(err)
		}
	}
}

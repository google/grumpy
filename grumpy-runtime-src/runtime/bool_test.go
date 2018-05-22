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

func TestBoolCompare(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(true, true), want: compareAllResultEq},
		{args: wrapArgs(true, false), want: compareAllResultGT},
		{args: wrapArgs(true, 1), want: compareAllResultEq},
		{args: wrapArgs(true, -1), want: compareAllResultGT},
		{args: wrapArgs(false, 0), want: compareAllResultEq},
		{args: wrapArgs(false, 1000), want: compareAllResultLT},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(compareAll, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestBoolCreate(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(None), wantExc: mustCreateException(TypeErrorType, `'__new__' requires a 'type' object but received a "NoneType"`)},
		{args: wrapArgs(BoolType), want: False.ToObject()},
		{args: wrapArgs(BoolType, None), want: False.ToObject()},
		{args: wrapArgs(BoolType, ""), want: False.ToObject()},
		{args: wrapArgs(BoolType, true), want: True.ToObject()},
		{args: wrapArgs(BoolType, newObject(ObjectType)), want: True.ToObject()},
		{args: wrapArgs(ObjectType), wantExc: mustCreateException(TypeErrorType, "bool.__new__(object): object is not a subtype of bool")},
		{args: wrapArgs(BoolType, "foo", "bar"), wantExc: mustCreateException(TypeErrorType, "bool() takes at most 1 argument (2 given)")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(BoolType, "__new__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestBoolStrRepr(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(true), want: NewStr("True").ToObject()},
		{args: wrapArgs(false), want: NewStr("False").ToObject()},
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

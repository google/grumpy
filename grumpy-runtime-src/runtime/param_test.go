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

func TestParamSpecValidate(t *testing.T) {
	testFunc := newBuiltinFunction("TestParamSpecValidate", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		if len(args) < 1 {
			return nil, f.RaiseType(TypeErrorType, "not enough args")
		}
		val, raised := ToNative(f, args[0])
		if raised != nil {
			return nil, raised
		}
		s, ok := val.Interface().(*ParamSpec)
		if !ok {
			return nil, f.RaiseType(TypeErrorType, "expected ParamSpec arg")
		}
		validated := make([]*Object, s.Count)
		if raised := s.Validate(f, validated, args[1:], kwargs); raised != nil {
			return nil, raised
		}
		return NewTuple(validated...).ToObject(), nil
	})
	cases := []invokeTestCase{
		invokeTestCase{args: wrapArgs(NewParamSpec("f1", nil, false, false)), want: NewTuple().ToObject()},
		invokeTestCase{args: wrapArgs(NewParamSpec("f2", []Param{{"a", nil}}, false, false), 123), want: newTestTuple(123).ToObject()},
		invokeTestCase{args: wrapArgs(NewParamSpec("f2", []Param{{"a", nil}}, false, false)), kwargs: wrapKWArgs("a", "apple"), want: newTestTuple("apple").ToObject()},
		invokeTestCase{args: wrapArgs(NewParamSpec("f2", []Param{{"a", nil}}, false, false)), kwargs: wrapKWArgs("b", "bear"), wantExc: mustCreateException(TypeErrorType, "f2() got an unexpected keyword argument 'b'")},
		invokeTestCase{args: wrapArgs(NewParamSpec("f2", []Param{{"a", nil}}, false, false)), wantExc: mustCreateException(TypeErrorType, "f2() takes at least 1 arguments (0 given)")},
		invokeTestCase{args: wrapArgs(NewParamSpec("f2", []Param{{"a", nil}}, false, false), 1, 2, 3), wantExc: mustCreateException(TypeErrorType, "f2() takes 1 arguments (3 given)")},
		invokeTestCase{args: wrapArgs(NewParamSpec("f3", []Param{{"a", nil}, {"b", nil}}, false, false), 1, 2), want: newTestTuple(1, 2).ToObject()},
		invokeTestCase{args: wrapArgs(NewParamSpec("f3", []Param{{"a", nil}, {"b", nil}}, false, false), 1), kwargs: wrapKWArgs("b", "bear"), want: newTestTuple(1, "bear").ToObject()},
		invokeTestCase{args: wrapArgs(NewParamSpec("f3", []Param{{"a", nil}, {"b", nil}}, false, false)), kwargs: wrapKWArgs("b", "bear", "a", "apple"), want: newTestTuple("apple", "bear").ToObject()},
		invokeTestCase{args: wrapArgs(NewParamSpec("f3", []Param{{"a", nil}, {"b", nil}}, false, false), 1), kwargs: wrapKWArgs("a", "alpha"), wantExc: mustCreateException(TypeErrorType, "f3() got multiple values for keyword argument 'a'")},
		invokeTestCase{args: wrapArgs(NewParamSpec("f4", []Param{{"a", nil}, {"b", None}}, false, false), 123), want: newTestTuple(123, None).ToObject()},
		invokeTestCase{args: wrapArgs(NewParamSpec("f4", []Param{{"a", nil}, {"b", None}}, false, false), 123, "bar"), want: newTestTuple(123, "bar").ToObject()},
		invokeTestCase{args: wrapArgs(NewParamSpec("f4", []Param{{"a", nil}, {"b", None}}, false, false)), kwargs: wrapKWArgs("a", 123, "b", "bar"), want: newTestTuple(123, "bar").ToObject()},
		invokeTestCase{args: wrapArgs(NewParamSpec("f5", []Param{{"a", nil}}, true, false), 1), want: newTestTuple(1, NewTuple()).ToObject()},
		invokeTestCase{args: wrapArgs(NewParamSpec("f5", []Param{{"a", nil}}, true, false), 1, 2, 3), want: newTestTuple(1, newTestTuple(2, 3)).ToObject()},
		invokeTestCase{args: wrapArgs(NewParamSpec("f6", []Param{{"a", nil}}, false, true), "bar"), want: newTestTuple("bar", NewDict()).ToObject()},
		invokeTestCase{args: wrapArgs(NewParamSpec("f6", []Param{{"a", nil}}, false, true)), kwargs: wrapKWArgs("a", "apple", "b", "bear"), want: newTestTuple("apple", newTestDict("b", "bear")).ToObject()},
		invokeTestCase{args: wrapArgs(NewParamSpec("f6", []Param{{"a", nil}}, false, true), "bar"), kwargs: wrapKWArgs("b", "baz", "c", "qux"), want: newTestTuple("bar", newTestDict("b", "baz", "c", "qux")).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(testFunc.ToObject(), &cas); err != "" {
			t.Error(err)
		}
	}
}

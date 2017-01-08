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

func TestGeneratorNext(t *testing.T) {
	f := NewRootFrame()
	var recursive *Object
	recursiveFn := func(*Object) (*Object, *BaseException) {
		next, raised := GetAttr(f, recursive, NewStr("next"), nil)
		if raised != nil {
			return nil, raised
		}
		return next.Call(f, nil, nil)
	}
	recursive = NewGenerator(f, recursiveFn).ToObject()
	emptyFn := func(*Object) (*Object, *BaseException) {
		return nil, nil
	}
	exhausted := NewGenerator(NewRootFrame(), emptyFn).ToObject()
	mustNotRaise(ListType.Call(NewRootFrame(), Args{exhausted}, nil))
	cases := []invokeTestCase{
		invokeTestCase{args: wrapArgs(recursive), wantExc: mustCreateException(ValueErrorType, "generator already executing")},
		invokeTestCase{args: wrapArgs(exhausted), wantExc: toBaseExceptionUnsafe(mustNotRaise(StopIterationType.Call(NewRootFrame(), nil, nil)))},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(GeneratorType, "next", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestGeneratorSend(t *testing.T) {
	emptyFn := func(*Object) (*Object, *BaseException) {
		return nil, nil
	}
	cases := []invokeTestCase{
		invokeTestCase{args: wrapArgs(NewGenerator(NewRootFrame(), emptyFn), 123), wantExc: mustCreateException(TypeErrorType, "can't send non-None value to a just-started generator")},
		invokeTestCase{args: wrapArgs(NewGenerator(NewRootFrame(), emptyFn), "foo", "bar"), wantExc: mustCreateException(TypeErrorType, "'send' of 'generator' requires 2 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(GeneratorType, "send", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestGeneratorSimple(t *testing.T) {
	f := NewRootFrame()
	fn := func(*Object) (*Object, *BaseException) {
		switch f.State() {
		case 0:
			goto Start
		case 1:
			goto Yield1
		case 2:
			goto Yield2
		default:
			t.Fatalf("got invalid state %d", f.State())
		}
	Start:
		f.PushCheckpoint(1)
		return NewStr("foo").ToObject(), nil
	Yield1:
		f.PushCheckpoint(2)
		return NewStr("bar").ToObject(), nil
	Yield2:
		return nil, nil
	}
	cas := &invokeTestCase{
		args: wrapArgs(NewGenerator(f, fn)),
		want: newTestList("foo", "bar").ToObject(),
	}
	if err := runInvokeTestCase(ListType.ToObject(), cas); err != "" {
		t.Error(err)
	}
}

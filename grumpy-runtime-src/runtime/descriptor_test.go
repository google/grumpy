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

func TestPropertyDelete(t *testing.T) {
	dummy := newObject(ObjectType)
	cases := []invokeTestCase{
		{args: wrapArgs(newProperty(nil, nil, wrapFuncForTest(func(f *Frame, o *Object) (*Object, *BaseException) { return None, nil })), dummy), want: None},
		{args: wrapArgs(newProperty(nil, nil, wrapFuncForTest(func(f *Frame, o *Object) (*Object, *BaseException) { return nil, f.RaiseType(ValueErrorType, "bar") })), dummy), wantExc: mustCreateException(ValueErrorType, "bar")},
		{args: wrapArgs(newProperty(nil, nil, nil), dummy), wantExc: mustCreateException(AttributeErrorType, "can't delete attribute")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(PropertyType, "__delete__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestPropertyGet(t *testing.T) {
	dummy := newObject(ObjectType)
	cases := []invokeTestCase{
		{args: wrapArgs(newProperty(wrapFuncForTest(func(f *Frame, o *Object) (*Object, *BaseException) { return o, nil }), nil, nil), dummy, ObjectType), want: dummy},
		{args: wrapArgs(newProperty(wrapFuncForTest(func(f *Frame, o *Object) (*Object, *BaseException) { return nil, f.RaiseType(ValueErrorType, "bar") }), nil, nil), dummy, ObjectType), wantExc: mustCreateException(ValueErrorType, "bar")},
		{args: wrapArgs(newProperty(nil, nil, nil), dummy, ObjectType), wantExc: mustCreateException(AttributeErrorType, "unreadable attribute")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(PropertyType, "__get__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestPropertyInit(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, args ...*Object) (*Object, *BaseException) {
		o, raised := PropertyType.Call(f, args, nil)
		if raised != nil {
			return nil, raised
		}
		p := toPropertyUnsafe(o)
		return newTestTuple(p.get, p.set, p.del).ToObject(), nil
	})
	cases := []invokeTestCase{
		{want: NewTuple(None, None, None).ToObject()},
		{args: wrapArgs("foo"), want: newTestTuple("foo", None, None).ToObject()},
		{args: wrapArgs("foo", None), want: newTestTuple("foo", None, None).ToObject()},
		{args: wrapArgs("foo", None, "bar"), want: newTestTuple("foo", None, "bar").ToObject()},
		{args: wrapArgs(1, 2, 3, 4), wantExc: mustCreateException(TypeErrorType, "'__init__' requires 3 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestPropertySet(t *testing.T) {
	dummy := newObject(ObjectType)
	cases := []invokeTestCase{
		{args: wrapArgs(newProperty(nil, wrapFuncForTest(func(_ *Frame, _, _ *Object) (*Object, *BaseException) { return None, nil }), nil), dummy, 123), want: None},
		{args: wrapArgs(newProperty(nil, wrapFuncForTest(func(f *Frame, _, _ *Object) (*Object, *BaseException) { return nil, f.RaiseType(ValueErrorType, "bar") }), nil), dummy, 123), wantExc: mustCreateException(ValueErrorType, "bar")},
		{args: wrapArgs(newProperty(nil, nil, nil), dummy, 123), wantExc: mustCreateException(AttributeErrorType, "can't set attribute")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(PropertyType, "__set__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestMakeStructFieldDescriptor(t *testing.T) {
	e := mustNotRaise(RuntimeErrorType.Call(NewRootFrame(), wrapArgs("foo"), nil))
	fun := newBuiltinFunction("TestMakeStructFieldDescriptor", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		if raised := checkMethodArgs(f, "TestMakeStructFieldDescriptor", args, TypeType, StrType, StrType, ObjectType); raised != nil {
			return nil, raised
		}
		t := toTypeUnsafe(args[0])
		desc := makeStructFieldDescriptor(t, toStrUnsafe(args[1]).Value(), toStrUnsafe(args[2]).Value(), fieldDescriptorRO)
		get, raised := GetAttr(f, desc, NewStr("__get__"), nil)
		if raised != nil {
			return nil, raised
		}
		return get.Call(f, wrapArgs(args[3], t), nil)
	}).ToObject()
	cases := []invokeTestCase{
		{args: wrapArgs(ObjectType, "dict", "__dict__", newObject(ObjectType)), want: None},
		{args: wrapArgs(ObjectType, "dict", "__dict__", newBuiltinFunction("foo", func(*Frame, Args, KWArgs) (*Object, *BaseException) { return nil, nil })), want: NewDict().ToObject()},
		{args: wrapArgs(IntType, "value", "value", 42), want: NewInt(42).ToObject()},
		{args: wrapArgs(StrType, "value", "value", 42), wantExc: mustCreateException(TypeErrorType, "descriptor 'value' for 'str' objects doesn't apply to 'int' objects")},
		{args: wrapArgs(BaseExceptionType, "args", "args", e), want: NewTuple(NewStr("foo").ToObject()).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestMakeStructFieldDescriptorRWGet(t *testing.T) {
	fun := newBuiltinFunction("TestMakeStructFieldDescriptorRW_get", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		if raised := checkMethodArgs(f, "TestMakeStructFieldDescriptorRW_get", args, TypeType, StrType, StrType, ObjectType); raised != nil {
			return nil, raised
		}
		t := toTypeUnsafe(args[0])
		desc := makeStructFieldDescriptor(t, toStrUnsafe(args[1]).Value(), toStrUnsafe(args[2]).Value(), fieldDescriptorRW)
		get, raised := GetAttr(f, desc, NewStr("__get__"), nil)
		if raised != nil {
			return nil, raised
		}
		return get.Call(f, wrapArgs(args[3], t), nil)
	}).ToObject()
	cases := []invokeTestCase{
		{args: wrapArgs(FileType, "Softspace", "softspace", newObject(FileType)), want: NewInt(0).ToObject()},
		{args: wrapArgs(FileType, "Softspace", "softspace", 42), wantExc: mustCreateException(TypeErrorType, "descriptor 'softspace' for 'file' objects doesn't apply to 'int' objects")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestMakeStructFieldDescriptorRWSet(t *testing.T) {
	fun := newBuiltinFunction("TestMakeStructFieldDescriptorRW_set", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		if raised := checkMethodArgs(f, "TestMakeStructFieldDescriptorRW_set", args, TypeType, StrType, StrType, ObjectType, ObjectType); raised != nil {
			return nil, raised
		}
		t := toTypeUnsafe(args[0])
		desc := makeStructFieldDescriptor(t, toStrUnsafe(args[1]).Value(), toStrUnsafe(args[2]).Value(), fieldDescriptorRW)
		set, raised := GetAttr(f, desc, NewStr("__set__"), nil)
		if raised != nil {
			return nil, raised
		}
		return set.Call(f, wrapArgs(args[3], args[4]), nil)
	}).ToObject()
	cases := []invokeTestCase{
		{args: wrapArgs(FileType, "Softspace", "softspace", newObject(FileType), NewInt(0).ToObject()), want: None},
		{args: wrapArgs(FileType, "Softspace", "softspace", newObject(FileType), NewInt(0)), want: None},
		{args: wrapArgs(FileType, "Softspace", "softspace", newObject(FileType), "wrong"), wantExc: mustCreateException(TypeErrorType, "an int is required")},
		{args: wrapArgs(FileType, "Softspace", "softspace", 42, NewInt(0)), wantExc: mustCreateException(TypeErrorType, "descriptor 'softspace' for 'file' objects doesn't apply to 'int' objects")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

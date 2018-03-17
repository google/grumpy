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

func TestSuperInitErrors(t *testing.T) {
	// Only tests __init__ error cases. Non-error cases are tested
	// implicitly by TestSuperGetAttribute.
	cases := []invokeTestCase{
		{wantExc: mustCreateException(TypeErrorType, "'__init__' requires 2 arguments")},
		{args: wrapArgs(FloatType, 123), wantExc: mustCreateException(TypeErrorType, "super(type, obj): obj must be an instance or subtype of type")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(superType.ToObject(), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestSuperGetAttribute(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, t *Type, o *Object) (*Object, *BaseException) {
		sup, raised := superType.Call(f, wrapArgs(t, o), nil)
		if raised != nil {
			return nil, raised
		}
		return GetAttr(f, sup, NewStr("attr"), nil)
	})
	// top, left, bottom, right refer to parts of a diamond hierarchy.
	topType := newTestClass("Top", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"attr": NewStr("top").ToObject(),
	}))
	top := newObject(topType)
	leftType := newTestClass("Left", []*Type{topType}, newStringDict(map[string]*Object{
		"attr": NewStr("left").ToObject(),
	}))
	left := newObject(leftType)
	rightType := newTestClass("Right", []*Type{topType}, newStringDict(map[string]*Object{
		"attr": newProperty(newBuiltinFunction("attr", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
			s := "right"
			if args[0] == nil {
				// When the "instance" argument is nil, the
				// descriptor is unbound.
				s = "rightType"
			}
			return NewStr(s).ToObject(), nil
		}).ToObject(), nil, nil).ToObject(),
	}))
	right := newObject(rightType)
	bottomType := newTestClass("Bottom", []*Type{leftType, rightType}, newStringDict(map[string]*Object{
		"attr": NewStr("bottom").ToObject(),
	}))
	bottom := newObject(bottomType)
	cases := []invokeTestCase{
		{args: wrapArgs(bottomType, bottom), want: NewStr("left").ToObject()},
		{args: wrapArgs(bottomType, bottomType), want: NewStr("left").ToObject()},
		{args: wrapArgs(leftType, bottom), want: NewStr("right").ToObject()},
		{args: wrapArgs(leftType, bottomType), want: NewStr("rightType").ToObject()},
		{args: wrapArgs(rightType, bottom), want: NewStr("top").ToObject()},
		{args: wrapArgs(rightType, bottomType), want: NewStr("top").ToObject()},
		{args: wrapArgs(topType, bottom), wantExc: mustCreateException(AttributeErrorType, "'super' object has no attribute 'attr'")},
		{args: wrapArgs(topType, bottomType), wantExc: mustCreateException(AttributeErrorType, "'super' object has no attribute 'attr'")},
		{args: wrapArgs(leftType, left), want: NewStr("top").ToObject()},
		{args: wrapArgs(leftType, leftType), want: NewStr("top").ToObject()},
		{args: wrapArgs(topType, left), wantExc: mustCreateException(AttributeErrorType, "'super' object has no attribute 'attr'")},
		{args: wrapArgs(topType, leftType), wantExc: mustCreateException(AttributeErrorType, "'super' object has no attribute 'attr'")},
		{args: wrapArgs(rightType, right), want: NewStr("top").ToObject()},
		{args: wrapArgs(rightType, rightType), want: NewStr("top").ToObject()},
		{args: wrapArgs(topType, right), wantExc: mustCreateException(AttributeErrorType, "'super' object has no attribute 'attr'")},
		{args: wrapArgs(topType, rightType), wantExc: mustCreateException(AttributeErrorType, "'super' object has no attribute 'attr'")},
		{args: wrapArgs(topType, top), wantExc: mustCreateException(AttributeErrorType, "'super' object has no attribute 'attr'")},
		{args: wrapArgs(topType, topType), wantExc: mustCreateException(AttributeErrorType, "'super' object has no attribute 'attr'")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

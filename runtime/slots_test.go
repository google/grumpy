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
	"reflect"
	"testing"
)

func TestSlotMakeCallable(t *testing.T) {
	fooType := newTestClass("Foo", []*Type{ObjectType}, NewDict())
	foo := newObject(fooType)
	// fun returns a tuple: (ret, args) where ret is the return value of
	// the callable produced by makeCallable and args are the arguments
	// that were passed into the callable.
	fun := wrapFuncForTest(func(f *Frame, s slot, ret *Object, args ...*Object) (*Object, *BaseException) {
		gotArgs := None
		prepareTestSlot(s, &gotArgs, ret)
		callable := s.makeCallable(fooType, "__slot__")
		if callable == nil {
			// This slot does not produce a callable, so just
			// return None.
			return None, nil
		}
		result, raised := callable.Call(f, args, nil)
		if raised != nil {
			return nil, raised
		}
		return NewTuple(result, gotArgs).ToObject(), nil
	})
	cases := []invokeTestCase{
		{args: wrapArgs(&basisSlot{}, None), want: None},
		{args: wrapArgs(&binaryOpSlot{}, "foo", foo, 123), want: newTestTuple("foo", newTestTuple(foo, 123)).ToObject()},
		{args: wrapArgs(&binaryOpSlot{}, None, "abc", 123), wantExc: mustCreateException(TypeErrorType, "'__slot__' requires a 'Foo' object but received a 'str'")},
		{args: wrapArgs(&delAttrSlot{}, None, foo, "bar"), want: newTestTuple(None, newTestTuple(foo, "bar")).ToObject()},
		{args: wrapArgs(&delAttrSlot{}, None, foo, 3.14), wantExc: mustCreateException(TypeErrorType, "'__slot__' requires a 'str' object but received a 'float'")},
		{args: wrapArgs(&delAttrSlot{}, RuntimeErrorType, foo, "bar"), wantExc: mustCreateException(RuntimeErrorType, "")},
		{args: wrapArgs(&deleteSlot{}, None, foo, "bar"), want: newTestTuple(None, newTestTuple(foo, "bar")).ToObject()},
		{args: wrapArgs(&deleteSlot{}, None, foo, 1, 2, 3), wantExc: mustCreateException(TypeErrorType, "'__slot__' of 'Foo' requires 2 arguments")},
		{args: wrapArgs(&deleteSlot{}, RuntimeErrorType, foo, "bar"), wantExc: mustCreateException(RuntimeErrorType, "")},
		{args: wrapArgs(&delItemSlot{}, None, foo, "bar"), want: newTestTuple(None, newTestTuple(foo, "bar")).ToObject()},
		{args: wrapArgs(&delItemSlot{}, None, foo, 1, 2, 3), wantExc: mustCreateException(TypeErrorType, "'__slot__' of 'Foo' requires 2 arguments")},
		{args: wrapArgs(&delItemSlot{}, RuntimeErrorType, foo, "bar"), wantExc: mustCreateException(RuntimeErrorType, "")},
		{args: wrapArgs(&getAttributeSlot{}, None, foo, "bar"), want: newTestTuple(None, newTestTuple(foo, "bar")).ToObject()},
		{args: wrapArgs(&getAttributeSlot{}, None, foo, 3.14), wantExc: mustCreateException(TypeErrorType, "'__slot__' requires a 'str' object but received a 'float'")},
		{args: wrapArgs(&getAttributeSlot{}, RuntimeErrorType, foo, "bar"), wantExc: mustCreateException(RuntimeErrorType, "")},
		{args: wrapArgs(&getSlot{}, 3.14, foo, 123, IntType), want: newTestTuple(3.14, newTestTuple(foo, 123, IntType)).ToObject()},
		{args: wrapArgs(&getSlot{}, None, foo, "bar", "baz"), wantExc: mustCreateException(TypeErrorType, "'__slot__' requires a 'type' object but received a 'str'")},
		{args: wrapArgs(&nativeSlot{}, None), want: None},
		{args: wrapArgs(&setAttrSlot{}, None, foo, "bar", 123), want: newTestTuple(None, newTestTuple(foo, "bar", 123)).ToObject()},
		{args: wrapArgs(&setAttrSlot{}, None, foo, true, None), wantExc: mustCreateException(TypeErrorType, "'__slot__' requires a 'str' object but received a 'bool'")},
		{args: wrapArgs(&setAttrSlot{}, RuntimeErrorType, foo, "bar", "baz"), wantExc: mustCreateException(RuntimeErrorType, "")},
		{args: wrapArgs(&setItemSlot{}, None, foo, "bar", true), want: newTestTuple(None, newTestTuple(foo, "bar", true)).ToObject()},
		{args: wrapArgs(&setItemSlot{}, None, foo, 1, 2, 3), wantExc: mustCreateException(TypeErrorType, "'__slot__' of 'Foo' requires 3 arguments")},
		{args: wrapArgs(&setItemSlot{}, RuntimeErrorType, foo, "bar", "baz"), wantExc: mustCreateException(RuntimeErrorType, "")},
		{args: wrapArgs(&setSlot{}, None, foo, 3.14, false), want: newTestTuple(None, newTestTuple(foo, 3.14, false)).ToObject()},
		{args: wrapArgs(&setSlot{}, RuntimeErrorType, foo, "bar", "baz"), wantExc: mustCreateException(RuntimeErrorType, "")},
		{args: wrapArgs(&setSlot{}, None, foo, 1, 2, 3), wantExc: mustCreateException(TypeErrorType, "'__slot__' of 'Foo' requires 3 arguments")},
		{args: wrapArgs(&unaryOpSlot{}, 42, foo), want: newTestTuple(42, NewTuple(foo)).ToObject()},
		{args: wrapArgs(&unaryOpSlot{}, None, foo, "bar"), wantExc: mustCreateException(TypeErrorType, "'__slot__' of 'Foo' requires 1 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestSlotWrapCallable(t *testing.T) {
	// fun returns a tuple: (ret, args, kwargs) where ret is the return
	// value of the slot, args and kwargs are the positional and keyword
	// parameters passed to it.
	fun := wrapFuncForTest(func(f *Frame, s slot, ret *Object, args ...*Object) (*Object, *BaseException) {
		gotArgs := None
		gotKWArgs := None
		wrapped := newBuiltinFunction("wrapped", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			if ret.isInstance(TypeType) && toTypeUnsafe(ret).isSubclass(BaseExceptionType) {
				return nil, f.Raise(ret, nil, nil)
			}
			gotArgs = NewTuple(args.makeCopy()...).ToObject()
			gotKWArgs = kwargs.makeDict().ToObject()
			return ret, nil
		}).ToObject()
		s.wrapCallable(wrapped)
		fnField := reflect.ValueOf(s).Elem().Field(0)
		if fnField.IsNil() {
			// Return None to denote the slot was empty.
			return None, nil
		}
		// Wrap the underlying slot function (s.Fn) and call it.  This
		// is more convenient than using reflection to call it.
		fn, raised := WrapNative(f, fnField)
		if raised != nil {
			return nil, raised
		}
		result, raised := fn.Call(f, append(Args{f.ToObject()}, args...), nil)
		if raised != nil {
			return nil, raised
		}
		return NewTuple(result, gotArgs, gotKWArgs).ToObject(), nil
	})
	o := newObject(ObjectType)
	cases := []invokeTestCase{
		{args: wrapArgs(&basisSlot{}, "no"), want: None},
		{args: wrapArgs(&binaryOpSlot{}, "ret", "foo", "bar"), want: newTestTuple("ret", newTestTuple("foo", "bar"), NewDict()).ToObject()},
		{args: wrapArgs(&binaryOpSlot{}, RuntimeErrorType, "foo", "bar"), wantExc: mustCreateException(RuntimeErrorType, "")},
		{args: wrapArgs(&callSlot{}, "ret", true, wrapArgs(1, 2), None), want: newTestTuple("ret", newTestTuple(true, 1, 2), NewDict()).ToObject()},
		{args: wrapArgs(&callSlot{}, "ret", "foo", None, wrapKWArgs("a", "b")), want: newTestTuple("ret", newTestTuple("foo"), newTestDict("a", "b")).ToObject()},
		{args: wrapArgs(&callSlot{}, "ret", 3.14, wrapArgs(false), wrapKWArgs("foo", 42)), want: newTestTuple("ret", newTestTuple(3.14, false), newTestDict("foo", 42)).ToObject()},
		{args: wrapArgs(&callSlot{}, RuntimeErrorType, true, wrapArgs(1, 2), None), wantExc: mustCreateException(RuntimeErrorType, "")},
		{args: wrapArgs(&delAttrSlot{}, "ret", o, "foo"), want: newTestTuple(None, newTestTuple(o, "foo"), NewDict()).ToObject()},
		{args: wrapArgs(&delAttrSlot{}, RuntimeErrorType, o, "foo"), wantExc: mustCreateException(RuntimeErrorType, "")},
		{args: wrapArgs(&deleteSlot{}, "ret", o, 3.14), want: newTestTuple(None, newTestTuple(o, 3.14), NewDict()).ToObject()},
		{args: wrapArgs(&deleteSlot{}, RuntimeErrorType, o, 3.14), wantExc: mustCreateException(RuntimeErrorType, "")},
		{args: wrapArgs(&delItemSlot{}, "ret", o, false), want: newTestTuple(None, newTestTuple(o, false), NewDict()).ToObject()},
		{args: wrapArgs(&delItemSlot{}, RuntimeErrorType, o, false), wantExc: mustCreateException(RuntimeErrorType, "")},
		{args: wrapArgs(&getAttributeSlot{}, "ret", o, "foo"), want: newTestTuple("ret", newTestTuple(o, "foo"), NewDict()).ToObject()},
		{args: wrapArgs(&getAttributeSlot{}, RuntimeErrorType, o, "foo"), wantExc: mustCreateException(RuntimeErrorType, "")},
		{args: wrapArgs(&getSlot{}, "ret", o, "foo", SetType), want: newTestTuple("ret", newTestTuple(o, "foo", SetType), NewDict()).ToObject()},
		{args: wrapArgs(&getSlot{}, RuntimeErrorType, o, "foo", SetType), wantExc: mustCreateException(RuntimeErrorType, "")},
		{args: wrapArgs(&initSlot{}, "ret", true, wrapArgs(1, 2), None), want: newTestTuple("ret", newTestTuple(true, 1, 2), NewDict()).ToObject()},
		{args: wrapArgs(&initSlot{}, "ret", "foo", None, wrapKWArgs("a", "b")), want: newTestTuple("ret", newTestTuple("foo"), newTestDict("a", "b")).ToObject()},
		{args: wrapArgs(&initSlot{}, "ret", 3.14, wrapArgs(false), wrapKWArgs("foo", 42)), want: newTestTuple("ret", newTestTuple(3.14, false), newTestDict("foo", 42)).ToObject()},
		{args: wrapArgs(&initSlot{}, RuntimeErrorType, true, wrapArgs(1, 2), None), wantExc: mustCreateException(RuntimeErrorType, "")},
		{args: wrapArgs(&nativeSlot{}, "no"), want: None},
		{args: wrapArgs(&newSlot{}, "ret", StrType, wrapArgs(1, 2), None), want: newTestTuple("ret", newTestTuple(StrType, 1, 2), NewDict()).ToObject()},
		{args: wrapArgs(&newSlot{}, "ret", ObjectType, None, wrapKWArgs("a", "b")), want: newTestTuple("ret", newTestTuple(ObjectType), newTestDict("a", "b")).ToObject()},
		{args: wrapArgs(&newSlot{}, "ret", ListType, wrapArgs(false), wrapKWArgs("foo", 42)), want: newTestTuple("ret", newTestTuple(ListType, false), newTestDict("foo", 42)).ToObject()},
		{args: wrapArgs(&newSlot{}, RuntimeErrorType, IntType, wrapArgs(1, 2), None), wantExc: mustCreateException(RuntimeErrorType, "")},
		{args: wrapArgs(&setAttrSlot{}, "ret", o, "foo", 42), want: newTestTuple(None, newTestTuple(o, "foo", 42), NewDict()).ToObject()},
		{args: wrapArgs(&setAttrSlot{}, RuntimeErrorType, o, "foo", 42), wantExc: mustCreateException(RuntimeErrorType, "")},
		{args: wrapArgs(&setItemSlot{}, "ret", o, "foo", 42), want: newTestTuple(None, newTestTuple(o, "foo", 42), NewDict()).ToObject()},
		{args: wrapArgs(&setItemSlot{}, RuntimeErrorType, o, "foo", 42), wantExc: mustCreateException(RuntimeErrorType, "")},
		{args: wrapArgs(&setSlot{}, "ret", o, "foo", 42), want: newTestTuple(None, newTestTuple(o, "foo", 42), NewDict()).ToObject()},
		{args: wrapArgs(&setSlot{}, RuntimeErrorType, o, "foo", 42), wantExc: mustCreateException(RuntimeErrorType, "")},
		{args: wrapArgs(&unaryOpSlot{}, "ret", "foo"), want: newTestTuple("ret", newTestTuple("foo"), NewDict()).ToObject()},
		{args: wrapArgs(&unaryOpSlot{}, RuntimeErrorType, "foo"), wantExc: mustCreateException(RuntimeErrorType, "")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

// prepareTestSlot sets the Fn field of s to a function that assigns its
// parameters to gotArgs and returns ret. If ret is type inheriting from
// BaseException, an exception of that type will be raised instead.
func prepareTestSlot(s slot, gotArgs **Object, ret *Object) {
	fnField := reflect.ValueOf(s).Elem().Field(0)
	slotFuncType := fnField.Type()
	numIn := slotFuncType.NumIn()
	numOut := slotFuncType.NumOut()
	fnField.Set(reflect.MakeFunc(slotFuncType, func(argValues []reflect.Value) []reflect.Value {
		f := argValues[0].Interface().(*Frame)
		var raised *BaseException
		if ret.isInstance(TypeType) && toTypeUnsafe(ret).isSubclass(BaseExceptionType) {
			raised = f.Raise(ret, nil, nil)
		} else {
			// Copy the input args into *gotArgs.
			elems := make([]*Object, numIn-1)
			for i := 1; i < numIn; i++ {
				if elems[i-1], raised = WrapNative(f, argValues[i]); raised != nil {
					break
				}
			}
			if raised == nil {
				*gotArgs = NewTuple(elems...).ToObject()
			}
		}
		raisedValue := reflect.ValueOf(raised)
		if numOut == 1 {
			// This slot does only returns an exception so return
			// raised (which may be nil).
			return []reflect.Value{raisedValue}
		}
		// Slot returns a single value and an exception.
		retValue := reflect.ValueOf((*Object)(nil))
		if raised == nil {
			retValue = reflect.ValueOf(ret)
		}
		return []reflect.Value{retValue, raisedValue}
	}))
}

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
	"fmt"
	"reflect"
	"regexp"
	"testing"
)

func TestObjectCall(t *testing.T) {
	arg0 := newObject(ObjectType)
	arg1 := newObject(ObjectType)
	args := wrapArgs(arg0, arg1)
	kwargs := wrapKWArgs("kwarg", newObject(ObjectType))
	kwargsDict := kwargs.makeDict()
	fn := func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		kwargsOrNone := None
		if len(kwargs) > 0 {
			kwargsOrNone = kwargs.makeDict().ToObject()
		}
		return newTestTuple(NewTuple(args.makeCopy()...), kwargsOrNone).ToObject(), nil
	}
	foo := newBuiltinFunction("foo", fn).ToObject()
	typ := newTestClass("Foo", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__call__": newBuiltinFunction("__call__", fn).ToObject(),
	}))
	callable := newObject(typ)
	raisesFunc := newBuiltinFunction("bar", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		return nil, f.RaiseType(RuntimeErrorType, "bar")
	}).ToObject()
	cases := []struct {
		callable *Object
		invokeTestCase
	}{
		{foo, invokeTestCase{args: args, kwargs: kwargs, want: newTestTuple(NewTuple(args...), kwargsDict).ToObject()}},
		{foo, invokeTestCase{args: args, want: newTestTuple(NewTuple(args...).ToObject(), None).ToObject()}},
		{foo, invokeTestCase{kwargs: kwargs, want: newTestTuple(NewTuple(), kwargsDict).ToObject()}},
		{foo, invokeTestCase{want: newTestTuple(NewTuple(), None).ToObject()}},
		{foo, invokeTestCase{args: wrapArgs(arg0), want: newTestTuple(NewTuple(arg0), None).ToObject()}},
		{callable, invokeTestCase{args: args, kwargs: kwargs, want: newTestTuple(NewTuple(callable, arg0, arg1), kwargsDict).ToObject()}},
		{newObject(ObjectType), invokeTestCase{wantExc: mustCreateException(TypeErrorType, "'object' object is not callable")}},
		{raisesFunc, invokeTestCase{wantExc: mustCreateException(RuntimeErrorType, "bar")}},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(cas.callable, &cas.invokeTestCase); err != "" {
			t.Error(err)
		}
	}
}

func TestNewObject(t *testing.T) {
	cases := []*Type{DictType, ObjectType, StrType, TypeType}
	for _, c := range cases {
		if o := newObject(c); o.Type() != c {
			t.Errorf("new object has type %q, want %q", o.Type().Name(), c.Name())
		}
	}
}

func TestObjectString(t *testing.T) {
	typ := newTestClass("Foo", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__repr__": newBuiltinFunction("__repr__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return nil, f.RaiseType(ExceptionType, "ruh roh")
		}).ToObject(),
	}))
	cases := []struct {
		o           *Object
		wantPattern string
	}{
		{newObject(ObjectType), `^<object object at \w+>$`},
		{NewTuple(NewStr("foo").ToObject(), NewStr("bar").ToObject()).ToObject(), `^\('foo', 'bar'\)$`},
		{ExceptionType.ToObject(), "^<type 'Exception'>$"},
		{NewStr("foo\nbar").ToObject(), `^'foo\\nbar'$`},
		{newObject(typ), `^<Foo object \(repr raised Exception\)>$`},
	}
	for _, cas := range cases {
		re := regexp.MustCompile(cas.wantPattern)
		s := cas.o.String()
		if matched := re.MatchString(s); !matched {
			t.Errorf("%v.String() = %q, doesn't match pattern %q", cas.o, s, re)
		}
	}
}

func TestObjectDelAttr(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, o *Object, name *Str) (*Object, *BaseException) {
		if raised := DelAttr(f, o, name); raised != nil {
			return nil, raised
		}
		return GetAttr(f, o, name, None)
	})
	dellerType := newTestClass("Deller", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__get__": newBuiltinFunction("__get__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			attr, raised := args[1].Dict().GetItemString(f, "attr")
			if raised != nil {
				return nil, raised
			}
			if attr == nil {
				return nil, f.RaiseType(AttributeErrorType, "attr")
			}
			return attr, nil
		}).ToObject(),
		"__delete__": newBuiltinFunction("__delete__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			deleted, raised := args[1].Dict().DelItemString(f, "attr")
			if raised != nil {
				return nil, raised
			}
			if !deleted {
				return nil, f.RaiseType(AttributeErrorType, "attr")
			}
			return None, nil
		}).ToObject(),
	}))
	fooType := newTestClass("Foo", []*Type{ObjectType}, newStringDict(map[string]*Object{"deller": newObject(dellerType)}))
	foo := newObject(fooType)
	if raised := foo.Dict().SetItemString(NewRootFrame(), "attr", NewInt(123).ToObject()); raised != nil {
		t.Fatal(raised)
	}
	cases := []invokeTestCase{
		{args: wrapArgs(foo, "deller"), want: None},
		{args: wrapArgs(newObject(fooType), "foo"), wantExc: mustCreateException(AttributeErrorType, "'Foo' object has no attribute 'foo'")},
		{args: wrapArgs(newObject(fooType), "deller"), wantExc: mustCreateException(AttributeErrorType, "attr")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestObjectGetAttribute(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, o *Object, name *Str) (*Object, *BaseException) {
		return GetAttr(f, o, name, nil)
	})
	// class Getter(object):
	//   def __get__(self, *args):
	//     return "got getter"
	getterType := newTestClass("Getter", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__get__": newBuiltinFunction("__get__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return NewStr("got getter").ToObject(), nil
		}).ToObject(),
	}))
	getter := newObject(getterType)
	// class Setter(object):
	//   def __get__(self, *args):
	//     return "got setter"
	//   def __set__(self, *args):
	//     pass
	setterType := newTestClass("Setter", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__get__": newBuiltinFunction("__get__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return NewStr("got setter").ToObject(), nil
		}).ToObject(),
		"__set__": newBuiltinFunction("__set__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return None, nil
		}).ToObject(),
	}))
	setter := newObject(setterType)
	// class Foo(object):
	//   pass
	// foo = Foo()
	fooType := newTestClass("Foo", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"bar":       NewInt(42).ToObject(),
		"baz":       NewStr("Foo's baz").ToObject(),
		"foogetter": getter,
		"foo":       NewInt(101).ToObject(),
		"barsetter": setter,
	}))
	foo := newObject(fooType)
	if raised := foo.Dict().SetItemString(NewRootFrame(), "fooattr", True.ToObject()); raised != nil {
		t.Fatal(raised)
	}
	if raised := foo.Dict().SetItemString(NewRootFrame(), "barattr", NewInt(-1).ToObject()); raised != nil {
		t.Fatal(raised)
	}
	if raised := foo.Dict().SetItemString(NewRootFrame(), "barsetter", NewStr("NOT setter").ToObject()); raised != nil {
		t.Fatal(raised)
	}
	cases := []invokeTestCase{
		{args: wrapArgs(foo, "bar"), want: NewInt(42).ToObject()},
		{args: wrapArgs(foo, "fooattr"), want: True.ToObject()},
		{args: wrapArgs(foo, "foogetter"), want: NewStr("got getter").ToObject()},
		{args: wrapArgs(foo, "bar"), want: NewInt(42).ToObject()},
		{args: wrapArgs(foo, "foo"), want: NewInt(101).ToObject()},
		{args: wrapArgs(foo, "barattr"), want: NewInt(-1).ToObject()},
		{args: wrapArgs(foo, "barsetter"), want: NewStr("got setter").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestObjectGetDict(t *testing.T) {
	fooType := newTestClass("Foo", []*Type{ObjectType}, NewDict())
	foo := newObject(fooType)
	if raised := SetAttr(NewRootFrame(), foo, NewStr("bar"), NewInt(123).ToObject()); raised != nil {
		panic(raised)
	}
	fun := wrapFuncForTest(func(f *Frame, o *Object) (*Object, *BaseException) {
		return GetAttr(f, o, NewStr("__dict__"), nil)
	})
	cases := []invokeTestCase{
		{args: wrapArgs(newObject(ObjectType)), wantExc: mustCreateException(AttributeErrorType, "'object' object has no attribute '__dict__'")},
		{args: wrapArgs(newObject(fooType)), want: NewDict().ToObject()},
		{args: wrapArgs(foo), want: newStringDict(map[string]*Object{"bar": NewInt(123).ToObject()}).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestObjectSetDict(t *testing.T) {
	fooType := newTestClass("Foo", []*Type{ObjectType}, NewDict())
	testDict := newStringDict(map[string]*Object{"bar": NewInt(123).ToObject()})
	fun := wrapFuncForTest(func(f *Frame, o, val *Object) (*Object, *BaseException) {
		if raised := SetAttr(f, o, NewStr("__dict__"), val); raised != nil {
			return nil, raised
		}
		d := o.Dict()
		if d == nil {
			return None, nil
		}
		return d.ToObject(), nil
	})
	cases := []invokeTestCase{
		{args: wrapArgs(newObject(ObjectType), NewDict()), wantExc: mustCreateException(AttributeErrorType, "'object' object has no attribute '__dict__'")},
		{args: wrapArgs(newObject(fooType), testDict), want: testDict.ToObject()},
		{args: wrapArgs(newObject(fooType), 123), wantExc: mustCreateException(TypeErrorType, "'_set_dict' requires a 'dict' object but received a 'int'")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestObjectNew(t *testing.T) {
	foo := makeTestType("Foo", ObjectType)
	foo.flags &= ^typeFlagInstantiable
	prepareType(foo)
	cases := []invokeTestCase{
		{args: wrapArgs(ExceptionType), want: newObject(ExceptionType)},
		{args: wrapArgs(IntType), want: NewInt(0).ToObject()},
		{wantExc: mustCreateException(TypeErrorType, "'__new__' requires 1 arguments")},
		{args: wrapArgs(None), wantExc: mustCreateException(TypeErrorType, `'__new__' requires a 'type' object but received a "NoneType"`)},
		{args: wrapArgs(foo), wantExc: mustCreateException(TypeErrorType, "object.__new__(Foo) is not safe, use Foo.__new__()")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(ObjectType, "__new__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestObjectReduce(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, method *Str, o *Object, args Args) (*Object, *BaseException) {
		// Call __reduce/reduce_ex__.
		reduce, raised := GetAttr(f, o, method, nil)
		if raised != nil {
			return nil, raised
		}
		result, raised := reduce.Call(f, args, nil)
		if raised != nil {
			return nil, raised
		}
		msg := NewStr(fmt.Sprintf("reduce must return a tuple, got %s", result.Type().Name())).ToObject()
		if raised := Assert(f, GetBool(result.isInstance(TupleType)).ToObject(), msg); raised != nil {
			return nil, raised
		}
		elems := toTupleUnsafe(result).elems
		numElems := len(elems)
		msg = NewStr(fmt.Sprintf("reduce must return a tuple with 2 <= len <= 5, got %d", numElems)).ToObject()
		if raised := Assert(f, GetBool(numElems >= 2 && numElems <= 5).ToObject(), msg); raised != nil {
			return nil, raised
		}
		newArgs := elems[1]
		msg = NewStr(fmt.Sprintf("reduce second return value must be tuple, got %s", newArgs.Type().Name())).ToObject()
		if raised := Assert(f, GetBool(newArgs.isInstance(TupleType)).ToObject(), msg); raised != nil {
			return nil, raised
		}
		// Call the reconstructor function with the args returned.
		reduced, raised := elems[0].Call(f, toTupleUnsafe(newArgs).elems, nil)
		if raised != nil {
			return nil, raised
		}
		// Return the reconstructed object, object state, list items
		// and dict items.
		state, list, dict := None, None, None
		if numElems > 2 && elems[2] != None {
			state = elems[2]
		}
		if numElems > 3 && elems[3] != None {
			if list, raised = ListType.Call(f, Args{elems[3]}, nil); raised != nil {
				return nil, raised
			}
		}
		if numElems > 4 && elems[4] != None {
			if dict, raised = DictType.Call(f, Args{elems[4]}, nil); raised != nil {
				return nil, raised
			}
		}
		return NewTuple(reduced, state, list, dict).ToObject(), nil
	})
	fooType := newTestClass("Foo", []*Type{StrType}, NewDict())
	fooNoDict := &Str{Object: Object{typ: fooType}, value: "fooNoDict"}
	// Calling __reduce_ex__ on a type that overrides __reduce__ should
	// forward to the call to __reduce__.
	reduceOverrideType := newTestClass("ReduceOverride", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__reduce__": newBuiltinFunction("__reduce__", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
			strNew, raised := GetAttr(f, StrType.ToObject(), NewStr("__new__"), nil)
			if raised != nil {
				return nil, raised
			}
			return newTestTuple(strNew, newTestTuple(StrType, "ReduceOverride")).ToObject(), nil
		}).ToObject(),
	}))
	getNewArgsRaisesType := newTestClass("GetNewArgsRaises", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__getnewargs__": newBuiltinFunction("__getnewargs__", func(f *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return nil, f.RaiseType(RuntimeErrorType, "uh oh")
		}).ToObject(),
	}))
	getNewArgsReturnsNonTupleType := newTestClass("GetNewArgsReturnsNonTuple", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__getnewargs__": newBuiltinFunction("__getnewargs__", func(f *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return NewInt(123).ToObject(), nil
		}).ToObject(),
	}))
	// Attempting to reduce an int will fail with "can't pickle" but
	// subclasses can be reduced.
	intSubclass := newTestClass("IntSubclass", []*Type{IntType}, NewDict())
	intSubclassInst := &Int{Object{typ: intSubclass}, 123}
	cases := []invokeTestCase{
		{args: wrapArgs("__reduce__", 42, Args{}), wantExc: mustCreateException(TypeErrorType, "can't pickle int objects")},
		{args: wrapArgs("__reduce__", 42, wrapArgs(2)), want: newTestTuple(42, None, None, None).ToObject()},
		{args: wrapArgs("__reduce_ex__", 42, Args{}), wantExc: mustCreateException(TypeErrorType, "can't pickle int objects")},
		{args: wrapArgs("__reduce__", 3.14, wrapArgs("bad proto")), wantExc: mustCreateException(TypeErrorType, "'__reduce__' requires a 'int' object but received a 'str'")},
		{args: wrapArgs("__reduce_ex__", 3.14, wrapArgs("bad proto")), wantExc: mustCreateException(TypeErrorType, "'__reduce_ex__' requires a 'int' object but received a 'str'")},
		{args: wrapArgs("__reduce__", newObject(fooType), Args{}), want: newTestTuple("", NewDict(), None, None).ToObject()},
		{args: wrapArgs("__reduce__", newObject(fooType), wrapArgs(2)), want: newTestTuple("", NewDict(), None, None).ToObject()},
		{args: wrapArgs("__reduce_ex__", newObject(fooType), Args{}), want: newTestTuple("", NewDict(), None, None).ToObject()},
		{args: wrapArgs("__reduce_ex__", newObject(reduceOverrideType), Args{}), want: newTestTuple("ReduceOverride", None, None, None).ToObject()},
		{args: wrapArgs("__reduce__", fooNoDict, Args{}), want: newTestTuple("fooNoDict", None, None, None).ToObject()},
		{args: wrapArgs("__reduce__", newTestList(1, 2, 3), wrapArgs(2)), want: newTestTuple(NewList(), None, newTestList(1, 2, 3), None).ToObject()},
		{args: wrapArgs("__reduce__", newTestDict("a", 1, "b", 2), wrapArgs(2)), want: newTestTuple(NewDict(), None, None, newTestDict("a", 1, "b", 2)).ToObject()},
		{args: wrapArgs("__reduce__", newObject(getNewArgsRaisesType), wrapArgs(2)), wantExc: mustCreateException(RuntimeErrorType, "uh oh")},
		{args: wrapArgs("__reduce__", newObject(getNewArgsReturnsNonTupleType), wrapArgs(2)), wantExc: mustCreateException(TypeErrorType, "__getnewargs__ should return a tuple, not 'int'")},
		{args: wrapArgs("__reduce__", newTestTuple("foo", "bar"), wrapArgs(2)), want: newTestTuple(newTestTuple("foo", "bar"), None, None, None).ToObject()},
		{args: wrapArgs("__reduce__", 3.14, wrapArgs(2)), want: newTestTuple(3.14, None, None, None).ToObject()},
		{args: wrapArgs("__reduce__", NewUnicode("abc"), wrapArgs(2)), want: newTestTuple(NewUnicode("abc"), None, None, None).ToObject()},
		{args: wrapArgs("__reduce__", intSubclassInst, Args{}), want: newTestTuple(intSubclassInst, None, None, None).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestObjectSetAttr(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, o *Object, name *Str, value *Object) (*Object, *BaseException) {
		if raised := SetAttr(f, o, name, value); raised != nil {
			return nil, raised
		}
		return GetAttr(f, o, name, None)
	})
	setterType := newTestClass("Setter", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__get__": newBuiltinFunction("__get__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			item, raised := args[1].Dict().GetItemString(f, "attr")
			if raised != nil {
				return nil, raised
			}
			if item == nil {
				return nil, raiseKeyError(f, NewStr("attr").ToObject())
			}
			return item, nil
		}).ToObject(),
		"__set__": newBuiltinFunction("__set__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			if raised := args[1].Dict().SetItemString(f, "attr", NewTuple(args.makeCopy()...).ToObject()); raised != nil {
				return nil, raised
			}
			return None, nil
		}).ToObject(),
	}))
	setter := newObject(setterType)
	fooType := newTestClass("Foo", []*Type{ObjectType}, newStringDict(map[string]*Object{"setter": setter}))
	foo := newObject(fooType)
	cases := []invokeTestCase{
		{args: wrapArgs(newObject(fooType), "foo", "abc"), want: NewStr("abc").ToObject()},
		{args: wrapArgs(foo, "setter", "baz"), want: NewTuple(setter, foo, NewStr("baz").ToObject()).ToObject()},
		{args: wrapArgs(newObject(ObjectType), "foo", 10), wantExc: mustCreateException(AttributeErrorType, "'object' has no attribute 'foo'")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestObjectStrRepr(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, o *Object, wantPattern string) *BaseException {
		re := regexp.MustCompile(wantPattern)
		s, raised := ToStr(f, o)
		if raised != nil {
			return raised
		}
		if !re.MatchString(s.Value()) {
			t.Errorf("str(%v) = %v, want %q", o, s, re)
		}
		s, raised = Repr(f, o)
		if raised != nil {
			return raised
		}
		if !re.MatchString(s.Value()) {
			t.Errorf("repr(%v) = %v, want %q", o, s, re)
		}
		return nil
	})
	type noReprMethodBasis struct{ Object }
	noReprMethodType := newType(TypeType, "noReprMethod", reflect.TypeOf(noReprMethodBasis{}), []*Type{}, NewDict())
	noReprMethodType.mro = []*Type{noReprMethodType}
	fooType := newTestClass("Foo", []*Type{ObjectType}, newTestDict("__module__", "foo.bar"))
	cases := []invokeTestCase{
		{args: wrapArgs(newObject(ObjectType), `^<object object at \w+>$`), want: None},
		{args: wrapArgs(newObject(noReprMethodType), `^<noReprMethod object at \w+>$`), want: None},
		{args: wrapArgs(newObject(fooType), `<foo\.bar\.Foo object at \w+>`), want: None},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

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
)

var (
	// FunctionType is the object representing the Python 'function' type.
	FunctionType = newBasisType("function", reflect.TypeOf(Function{}), toFunctionUnsafe, ObjectType)
	// StaticMethodType is the object representing the Python
	// 'staticmethod' type.
	StaticMethodType = newBasisType("staticmethod", reflect.TypeOf(staticMethod{}), toStaticMethodUnsafe, ObjectType)
	// ClassMethodType is the object representing the Python
	// 'classmethod' type.
	ClassMethodType = newBasisType("classmethod", reflect.TypeOf(classMethod{}), toClassMethodUnsafe, ObjectType)
)

// Args represent positional parameters in a call to a Python function.
type Args []*Object

func (a Args) makeCopy() Args {
	result := make(Args, len(a))
	copy(result, a)
	return result
}

// KWArg represents a keyword argument in a call to a Python function.
type KWArg struct {
	Name  string
	Value *Object
}

// KWArgs represents a list of keyword parameters in a call to a Python
// function.
type KWArgs []KWArg

// String returns a string representation of k, e.g. for debugging.
func (k KWArgs) String() string {
	return k.makeDict().String()
}

func (k KWArgs) get(name string, def *Object) *Object {
	for _, kwarg := range k {
		if kwarg.Name == name {
			return kwarg.Value
		}
	}
	return def
}

func (k KWArgs) makeDict() *Dict {
	m := map[string]*Object{}
	for _, kw := range k {
		m[kw.Name] = kw.Value
	}
	return newStringDict(m)
}

// Func is a Go function underlying a Python Function object.
type Func func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException)

// Function represents Python 'function' objects.
type Function struct {
	Object
	fn      Func
	name    string `attr:"__name__"`
	code    *Code  `attr:"func_code"`
	globals *Dict  `attr:"func_globals"`
}

// NewFunction creates a function object corresponding to a Python function
// taking the given args, vararg and kwarg. When called, the arguments are
// validated before calling fn. This includes checking that an appropriate
// number of arguments are provided, populating *args and **kwargs if
// necessary, etc.
func NewFunction(c *Code, globals *Dict) *Function {
	return &Function{Object{typ: FunctionType, dict: NewDict()}, nil, c.name, c, globals}
}

// newBuiltinFunction returns a function object with the given name that
// invokes fn when called.
func newBuiltinFunction(name string, fn Func) *Function {
	return &Function{Object: Object{typ: FunctionType, dict: NewDict()}, fn: fn, name: name}
}

func toFunctionUnsafe(o *Object) *Function {
	return (*Function)(o.toPointer())
}

// ToObject upcasts f to an Object.
func (f *Function) ToObject() *Object {
	return &f.Object
}

// Name returns f's name field.
func (f *Function) Name() string {
	return f.name
}

func functionCall(f *Frame, callable *Object, args Args, kwargs KWArgs) (*Object, *BaseException) {
	fun := toFunctionUnsafe(callable)
	code := fun.code
	if code == nil {
		return fun.fn(f, args, kwargs)
	}
	return code.Eval(f, fun.globals, args, kwargs)
}

func functionGet(f *Frame, desc, instance *Object, owner *Type) (*Object, *BaseException) {
	args := f.MakeArgs(3)
	args[0] = desc
	args[1] = instance
	args[2] = owner.ToObject()
	ret, raised := MethodType.Call(f, args, nil)
	f.FreeArgs(args)
	return ret, raised
}

func functionRepr(_ *Frame, o *Object) (*Object, *BaseException) {
	fun := toFunctionUnsafe(o)
	return NewStr(fmt.Sprintf("<%s %s at %p>", fun.typ.Name(), fun.Name(), fun)).ToObject(), nil
}

func initFunctionType(map[string]*Object) {
	FunctionType.flags &= ^(typeFlagInstantiable | typeFlagBasetype)
	FunctionType.slots.Call = &callSlot{functionCall}
	FunctionType.slots.Get = &getSlot{functionGet}
	FunctionType.slots.Repr = &unaryOpSlot{functionRepr}
}

// staticMethod represents Python 'staticmethod' objects.
type staticMethod struct {
	Object
	callable *Object
}

func newStaticMethod(callable *Object) *staticMethod {
	return &staticMethod{Object{typ: StaticMethodType}, callable}
}

func toStaticMethodUnsafe(o *Object) *staticMethod {
	return (*staticMethod)(o.toPointer())
}

// ToObject upcasts f to an Object.
func (m *staticMethod) ToObject() *Object {
	return &m.Object
}

func staticMethodGet(f *Frame, desc, _ *Object, _ *Type) (*Object, *BaseException) {
	m := toStaticMethodUnsafe(desc)
	if m.callable == nil {
		return nil, f.RaiseType(RuntimeErrorType, "uninitialized staticmethod object")
	}
	return m.callable, nil
}

func staticMethodInit(f *Frame, o *Object, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkFunctionArgs(f, "__init__", args, ObjectType); raised != nil {
		return nil, raised
	}
	toStaticMethodUnsafe(o).callable = args[0]
	return None, nil
}

func initStaticMethodType(map[string]*Object) {
	StaticMethodType.slots.Get = &getSlot{staticMethodGet}
	StaticMethodType.slots.Init = &initSlot{staticMethodInit}
}

// classMethod represents Python 'classmethod' objects.
type classMethod struct {
	Object
	callable *Object
}

func newClassMethod(callable *Object) *classMethod {
	return &classMethod{Object{typ: ClassMethodType}, callable}
}

func toClassMethodUnsafe(o *Object) *classMethod {
	return (*classMethod)(o.toPointer())
}

// ToObject upcasts f to an Object.
func (m *classMethod) ToObject() *Object {
	return &m.Object
}

func classMethodGet(f *Frame, desc, _ *Object, owner *Type) (*Object, *BaseException) {
	m := toClassMethodUnsafe(desc)
	if m.callable == nil {
		return nil, f.RaiseType(RuntimeErrorType, "uninitialized classmethod object")
	}
	args := f.MakeArgs(3)
	args[0] = m.callable
	args[1] = owner.ToObject()
	args[2] = args[1]
	ret, raised := MethodType.Call(f, args, nil)
	f.FreeArgs(args)
	return ret, raised
}

func classMethodInit(f *Frame, o *Object, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkFunctionArgs(f, "__init__", args, ObjectType); raised != nil {
		return nil, raised
	}
	toClassMethodUnsafe(o).callable = args[0]
	return None, nil
}

func initClassMethodType(map[string]*Object) {
	ClassMethodType.slots.Get = &getSlot{classMethodGet}
	ClassMethodType.slots.Init = &initSlot{classMethodInit}
}

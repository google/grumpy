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
	fn   Func
	name string `attr:"__name__"`
	spec *functionSpec
}

// FunctionArg describes a parameter to a Python function.
type FunctionArg struct {
	// Name is the argument name.
	Name string
	// Def is the default value to use if the argument is not provided. If
	// no default is specified then Def is nil.
	Def *Object
}

// functionSpec describes a Python user function.
type functionSpec struct {
	args        []FunctionArg
	fn          func(*Frame, []*Object) (*Object, *BaseException)
	specArgc    int
	numBodyArgs int
	varArgIndex int
	kwArgIndex  int
	minArgc     int
}

// NewFunction creates a function object corresponding to a Python function
// described by spec. When called, the arguments are validated before calling
// spec.Fn. This includes checking that an appropriate number of arguments are
// provided, populating *args and **kwargs if necessary, etc.
func NewFunction(name string, args []FunctionArg, vararg string, kwarg string, fn func(*Frame, []*Object) (*Object, *BaseException)) *Function {
	specArgc := len(args)
	numBodyArgs := specArgc
	varArgIndex, kwArgIndex := -1, -1
	if vararg != "" {
		varArgIndex = numBodyArgs
		numBodyArgs++
	}
	if kwarg != "" {
		kwArgIndex = numBodyArgs
		numBodyArgs++
	}
	minArgc := 0
	for ; minArgc < specArgc; minArgc++ {
		if args[minArgc].Def != nil {
			break
		}
	}
	for i := minArgc; i < specArgc; i++ {
		if args[i].Def == nil {
			format := "%s() non-keyword arg %s after keyword arg"
			logFatal(fmt.Sprintf(format, name, args[i].Name))
		}
	}
	spec := &functionSpec{
		args:        args,
		fn:          fn,
		specArgc:    specArgc,
		numBodyArgs: numBodyArgs,
		varArgIndex: varArgIndex,
		kwArgIndex:  kwArgIndex,
		minArgc:     minArgc,
	}
	return &Function{Object{typ: FunctionType, dict: NewDict()}, nil, name, spec}
}

// newBuiltinFunction returns a function object with the given name that
// invokes fn when called.
func newBuiltinFunction(name string, fn func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException)) *Function {
	return &Function{Object{typ: FunctionType, dict: NewDict()}, fn, name, nil}
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
	spec := fun.spec
	if spec == nil {
		return fun.fn(f, args, kwargs)
	}
	argc := len(args)
	if argc > spec.specArgc && spec.varArgIndex == -1 {
		format := "%s() takes %d arguments (%d given)"
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, fun.name, spec.specArgc, argc))
	}
	bodyArgs := f.MakeArgs(spec.numBodyArgs)
	i := 0
	for ; i < argc && i < spec.specArgc; i++ {
		bodyArgs[i] = args[i]
	}
	if spec.varArgIndex != -1 {
		bodyArgs[spec.varArgIndex] = NewTuple(args[i:].makeCopy()...).ToObject()
	}
	var kwargDict *Dict
	if spec.kwArgIndex != -1 {
		kwargDict = NewDict()
		bodyArgs[spec.kwArgIndex] = kwargDict.ToObject()
	}
	for _, kw := range kwargs {
		name := kw.Name
		j := 0
		for ; j < spec.specArgc; j++ {
			if spec.args[j].Name == name {
				if bodyArgs[j] != nil {
					format := "%s() got multiple values for keyword argument '%s'"
					return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, fun.name, name))
				}
				bodyArgs[j] = kw.Value
				break
			}
		}
		if j == spec.specArgc {
			if kwargDict == nil {
				format := "%s() got an unexpected keyword argument '%s'"
				return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, fun.name, name))
			}
			if raised := kwargDict.SetItemString(f, name, kw.Value); raised != nil {
				return nil, raised
			}
		}
	}
	for ; i < spec.specArgc; i++ {
		arg := spec.args[i]
		if bodyArgs[i] == nil {
			if arg.Def == nil {
				format := "%s() takes at least %d arguments (%d given)"
				return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, fun.name, spec.minArgc, argc))
			}
			bodyArgs[i] = arg.Def
		}
	}
	ret, raised := spec.fn(f, bodyArgs)
	f.FreeArgs(bodyArgs)
	return ret, raised
}

func functionGet(_ *Frame, desc, instance *Object, owner *Type) (*Object, *BaseException) {
	return NewMethod(toFunctionUnsafe(desc), instance, owner).ToObject(), nil
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

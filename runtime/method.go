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

// Method represents Python 'instancemethod' objects.
type Method struct {
	Object
	function *Function
	self     *Object
	class    *Type
	name     string `attr:"__name__"`
}

// NewMethod returns a method wrapping the given function belonging to class.
// When self is None the method is unbound, otherwise it is bound to self.
func NewMethod(function *Function, self *Object, class *Type) *Method {
	return &Method{Object{typ: MethodType}, function, self, class, function.Name()}
}

func toMethodUnsafe(o *Object) *Method {
	return (*Method)(o.toPointer())
}

// ToObject upcasts m to an Object.
func (m *Method) ToObject() *Object {
	return &m.Object
}

// MethodType is the object representing the Python 'instancemethod' type.
var MethodType = newBasisType("instancemethod", reflect.TypeOf(Method{}), toMethodUnsafe, ObjectType)

func methodCall(f *Frame, callable *Object, args Args, kwargs KWArgs) (*Object, *BaseException) {
	m := toMethodUnsafe(callable)
	var methodArgs []*Object
	argc := len(args)
	if m.self == None {
		if argc < 1 {
			format := "unbound method %s() must be called with %s instance as first argument (got nothing instead)"
			return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, m.name, m.class.Name()))
		}
		if !args[0].isInstance(m.class) {
			format := "unbound method %s() must be called with %s instance as first argument (got %s instance instead)"
			return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, m.name, m.class.Name(), args[0].typ.Name()))
		}
		methodArgs = args
	} else {
		methodArgs = make([]*Object, argc+1, argc+1)
		methodArgs[0] = m.self
		copy(methodArgs[1:], args)
	}
	return m.function.Call(f, methodArgs, kwargs)
}

func methodRepr(f *Frame, o *Object) (*Object, *BaseException) {
	m := toMethodUnsafe(o)
	s := ""
	if m.self == None {
		s = fmt.Sprintf("<unbound method %s.%s>", m.class.Name(), m.function.Name())
	} else {
		repr, raised := Repr(f, m.self)
		if raised != nil {
			return nil, raised
		}
		s = fmt.Sprintf("<bound method %s.%s of %s>", m.class.Name(), m.function.Name(), repr.Value())
	}
	return NewStr(s).ToObject(), nil
}

func initMethodType(map[string]*Object) {
	// TODO: Should be instantiable.
	MethodType.flags &= ^(typeFlagBasetype | typeFlagInstantiable)
	MethodType.slots.Call = &callSlot{methodCall}
	MethodType.slots.Repr = &unaryOpSlot{methodRepr}
}

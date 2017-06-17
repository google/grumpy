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
	function *Object `attr:"im_func"`
	self     *Object `attr:"im_self"`
	class    *Object `attr:"im_class"`
	name     string  `attr:"__name__"`
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
	argc := len(args)
	if m.self != nil {
		methodArgs := f.MakeArgs(argc + 1)
		methodArgs[0] = m.self
		copy(methodArgs[1:], args)
		result, raised := m.function.Call(f, methodArgs, kwargs)
		f.FreeArgs(methodArgs)
		return result, raised
	}
	if argc < 1 {
		className, raised := methodGetMemberName(f, m.class)
		if raised != nil {
			return nil, raised
		}
		format := "unbound method %s() must be called with %s " +
			"instance as first argument (got nothing instead)"
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, m.name, className))
	}
	// instancemethod.__new__ ensures that m.self and m.class are not both
	// nil. Since m.self is nil, we know that m.class is not.
	isInst, raised := IsInstance(f, args[0], m.class)
	if raised != nil {
		return nil, raised
	}
	if !isInst {
		className, raised := methodGetMemberName(f, m.class)
		if raised != nil {
			return nil, raised
		}
		format := "unbound method %s() must be called with %s " +
			"instance as first argument (got %s instance instead)"
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, m.name, className, args[0].typ.Name()))
	}
	return m.function.Call(f, args, kwargs)
}

func methodGet(f *Frame, desc, instance *Object, owner *Type) (*Object, *BaseException) {
	m := toMethodUnsafe(desc)
	if m.self != nil {
		// Don't bind a method that's already bound.
		return desc, nil
	}
	if m.class != nil {
		subcls, raised := IsSubclass(f, owner.ToObject(), m.class)
		if raised != nil {
			return nil, raised
		}
		if !subcls {
			// Don't bind if owner is not a subclass of m.class.
			return desc, nil
		}
	}
	return (&Method{Object{typ: MethodType}, m.function, instance, owner.ToObject(), m.name}).ToObject(), nil
}

func methodNew(f *Frame, t *Type, args Args, _ KWArgs) (*Object, *BaseException) {
	expectedTypes := []*Type{ObjectType, ObjectType, ObjectType}
	argc := len(args)
	if argc == 2 {
		expectedTypes = expectedTypes[:2]
	}
	if raised := checkFunctionArgs(f, "__new__", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	function, self := args[0], args[1]
	if self == None {
		self = nil
	}
	var class *Object
	if argc > 2 {
		class = args[2]
	} else if self == nil {
		return nil, f.RaiseType(TypeErrorType, "unbound methods must have non-NULL im_class")
	}
	if function.Type().slots.Call == nil {
		return nil, f.RaiseType(TypeErrorType, "first argument must be callable")
	}
	functionName, raised := methodGetMemberName(f, function)
	if raised != nil {
		return nil, raised
	}
	method := &Method{Object{typ: MethodType}, function, self, class, functionName}
	return method.ToObject(), nil
}

func methodRepr(f *Frame, o *Object) (*Object, *BaseException) {
	m := toMethodUnsafe(o)
	s := ""
	className, raised := methodGetMemberName(f, m.class)
	if raised != nil {
		return nil, raised
	}
	functionName, raised := methodGetMemberName(f, m.function)
	if raised != nil {
		return nil, raised
	}
	if m.self == nil {
		s = fmt.Sprintf("<unbound method %s.%s>", className, functionName)
	} else {
		repr, raised := Repr(f, m.self)
		if raised != nil {
			return nil, raised
		}
		s = fmt.Sprintf("<bound method %s.%s of %s>", className, functionName, repr.Value())
	}
	return NewStr(s).ToObject(), nil
}

func initMethodType(map[string]*Object) {
	MethodType.flags &= ^typeFlagBasetype
	MethodType.slots.Call = &callSlot{methodCall}
	MethodType.slots.Get = &getSlot{methodGet}
	MethodType.slots.Repr = &unaryOpSlot{methodRepr}
	MethodType.slots.New = &newSlot{methodNew}
}

func methodGetMemberName(f *Frame, o *Object) (string, *BaseException) {
	if o == nil {
		return "?", nil
	}
	name, raised := GetAttr(f, o, internedName, None)
	if raised != nil {
		return "", raised
	}
	if !name.isInstance(BaseStringType) {
		return "?", nil
	}
	nameStr, raised := ToStr(f, name)
	if raised != nil {
		return "", raised
	}
	return nameStr.Value(), nil
}

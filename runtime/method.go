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
	var methodArgs []*Object
	argc := len(args)
	if m.self == None {
		if argc < 1 {
			className, raised := methodGetMemberName(f, m.class)
			if raised != nil {
				return nil, raised
			}
			format := "unbound method %s() must be called with %s " +
				"instance as first argument (got nothing instead)"
			return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, m.name, className))
		}
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
		methodArgs = args
	} else {
		methodArgs = make([]*Object, argc+1, argc+1)
		methodArgs[0] = m.self
		copy(methodArgs[1:], args)
	}
	return m.function.Call(f, methodArgs, kwargs)
}

func methodNew(f *Frame, t *Type, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkFunctionArgs(f, "__new__", args, ObjectType, ObjectType, ObjectType); raised != nil {
		return nil, raised
	}
	if args[0].Type().slots.Call == nil {
		return nil, f.RaiseType(TypeErrorType, "first argument must be callable")
	}
	functionName, raised := methodGetMemberName(f, args[0])
	if raised != nil {
		return nil, raised
	}
	method := &Method{Object{typ: MethodType}, args[0], args[1], args[2], functionName}
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
	if m.self == None {
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
	MethodType.slots.Repr = &unaryOpSlot{methodRepr}
	MethodType.slots.New = &newSlot{methodNew}
}

func methodGetMemberName(f *Frame, o *Object) (string, *BaseException) {
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

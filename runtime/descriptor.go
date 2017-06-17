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

type fieldDescriptorType int

const (
	fieldDescriptorRO fieldDescriptorType = iota
	fieldDescriptorRW
)

// Property represents Python 'property' objects.
type Property struct {
	Object
	get, set, del *Object
}

func newProperty(get, set, del *Object) *Property {
	return &Property{Object{typ: PropertyType}, get, set, del}
}

func toPropertyUnsafe(o *Object) *Property {
	return (*Property)(o.toPointer())
}

// ToObject upcasts p to an Object.
func (p *Property) ToObject() *Object {
	return &p.Object
}

// PropertyType is the object representing the Python 'property' type.
var PropertyType = newBasisType("property", reflect.TypeOf(Property{}), toPropertyUnsafe, ObjectType)

func initPropertyType(map[string]*Object) {
	PropertyType.slots.Delete = &deleteSlot{propertyDelete}
	PropertyType.slots.Get = &getSlot{propertyGet}
	PropertyType.slots.Init = &initSlot{propertyInit}
	PropertyType.slots.Set = &setSlot{propertySet}
}

func propertyDelete(f *Frame, desc, inst *Object) *BaseException {
	p := toPropertyUnsafe(desc)
	if p.del == nil || p.del == None {
		return f.RaiseType(AttributeErrorType, "can't delete attribute")
	}
	_, raised := p.del.Call(f, Args{inst}, nil)
	return raised
}

func propertyGet(f *Frame, desc, instance *Object, _ *Type) (*Object, *BaseException) {
	p := toPropertyUnsafe(desc)
	if p.get == nil || p.get == None {
		return nil, f.RaiseType(AttributeErrorType, "unreadable attribute")
	}
	return p.get.Call(f, Args{instance}, nil)
}

func propertyInit(f *Frame, o *Object, args Args, _ KWArgs) (*Object, *BaseException) {
	expectedTypes := []*Type{ObjectType, ObjectType, ObjectType}
	argc := len(args)
	if argc < 3 {
		expectedTypes = expectedTypes[:argc]
	}
	if raised := checkFunctionArgs(f, "__init__", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	p := toPropertyUnsafe(o)
	if argc > 0 {
		p.get = args[0]
	}
	if argc > 1 {
		p.set = args[1]
	}
	if argc > 2 {
		p.del = args[2]
	}
	return None, nil
}

func propertySet(f *Frame, desc, inst, value *Object) *BaseException {
	p := toPropertyUnsafe(desc)
	if p.set == nil || p.set == None {
		return f.RaiseType(AttributeErrorType, "can't set attribute")
	}
	_, raised := p.set.Call(f, Args{inst, value}, nil)
	return raised
}

// makeStructFieldDescriptor creates a descriptor with a getter that returns
// the field given by fieldName from t's basis structure.
func makeStructFieldDescriptor(t *Type, fieldName, propertyName string, fieldMode fieldDescriptorType) *Object {
	field, ok := t.basis.FieldByName(fieldName)
	if !ok {
		logFatal(fmt.Sprintf("no such field %q for basis %s", fieldName, nativeTypeName(t.basis)))
	}

	getterFunc := func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		if raised := checkFunctionArgs(f, fieldName, args, ObjectType); raised != nil {
			return nil, raised
		}

		self := args[0]
		if !self.isInstance(t) {
			format := "descriptor '%s' for '%s' objects doesn't apply to '%s' objects"
			return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, propertyName, t.Name(), self.typ.Name()))
		}

		return WrapNative(f, t.slots.Basis.Fn(self).FieldByIndex(field.Index))
	}
	getter := newBuiltinFunction("_get"+fieldName, getterFunc).ToObject()

	setter := None
	if fieldMode == fieldDescriptorRW {
		if field.PkgPath != "" {
			logFatal(fmt.Sprintf("field '%q' is not public on Golang code. Please fix it.", fieldName))
		}

		setterFunc := func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
			if raised := checkFunctionArgs(f, fieldName, args, ObjectType, ObjectType); raised != nil {
				return nil, raised
			}

			self := args[0]
			newValue := args[1]

			if !self.isInstance(t) {
				format := "descriptor '%s' for '%s' objects doesn't apply to '%s' objects"
				return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, propertyName, t.Name(), self.typ.Name()))
			}

			val := t.slots.Basis.Fn(self).FieldByIndex(field.Index)
			converted, raised := maybeConvertValue(f, newValue, field.Type)
			if raised != nil {
				return nil, raised
			}

			val.Set(converted)
			return None, nil
		}

		setter = newBuiltinFunction("_set"+fieldName, setterFunc).ToObject()
	}
	return newProperty(getter, setter, None).ToObject()
}

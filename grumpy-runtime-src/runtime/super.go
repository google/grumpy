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
)

var (
	// superType is the object representing the Python 'super' type.
	superType = newBasisType("super", reflect.TypeOf(super{}), toSuperUnsafe, ObjectType)
)

type super struct {
	Object
	sub     *Type
	obj     *Object
	objType *Type
}

func toSuperUnsafe(o *Object) *super {
	return (*super)(o.toPointer())
}

func superInit(f *Frame, o *Object, args Args, _ KWArgs) (*Object, *BaseException) {
	// TODO: Support the unbound form of super.
	if raised := checkFunctionArgs(f, "__init__", args, TypeType, ObjectType); raised != nil {
		return nil, raised
	}
	sup := toSuperUnsafe(o)
	sub := toTypeUnsafe(args[0])
	obj := args[1]
	var objType *Type
	if obj.isInstance(TypeType) && toTypeUnsafe(obj).isSubclass(sub) {
		objType = toTypeUnsafe(obj)
	} else if obj.isInstance(sub) {
		objType = obj.typ
	} else {
		return nil, f.RaiseType(TypeErrorType, "super(type, obj): obj must be an instance or subtype of type")
	}
	sup.sub = sub
	sup.obj = obj
	sup.objType = objType
	return None, nil
}

func superGetAttribute(f *Frame, o *Object, name *Str) (*Object, *BaseException) {
	sup := toSuperUnsafe(o)
	// Tell the truth about the __class__ attribute.
	if sup.objType != nil && name.Value() != "__class__" {
		mro := sup.objType.mro
		n := len(mro)
		// Start from the immediate mro successor to the specified type.
		i := 0
		for i < n && mro[i] != sup.sub {
			i++
		}
		i++
		var inst *Object
		if sup.obj != sup.objType.ToObject() {
			inst = sup.obj
		}
		// Now do normal mro lookup from the successor type.
		for ; i < n; i++ {
			dict := mro[i].Dict()
			res, raised := dict.GetItem(f, name.ToObject())
			if raised != nil {
				return nil, raised
			}
			if res != nil {
				if get := res.typ.slots.Get; get != nil {
					// Found a descriptor so invoke it.
					return get.Fn(f, res, inst, sup.objType)
				}
				return res, nil
			}
		}
	}
	// Attribute not found on base classes so lookup the attr on the super
	// object itself. Most likely will AttributeError.
	return objectGetAttribute(f, o, name)
}

func initSuperType(map[string]*Object) {
	superType.slots.GetAttribute = &getAttributeSlot{superGetAttribute}
	superType.slots.Init = &initSlot{superInit}
}

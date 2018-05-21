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

// GetBool returns True if v is true, False otherwise.
func GetBool(v bool) *Int {
	if v {
		return True
	}
	return False
}

// BoolType is the object representing the Python 'bool' type.
var BoolType = newSimpleType("bool", IntType)

func boolNative(_ *Frame, o *Object) (reflect.Value, *BaseException) {
	return reflect.ValueOf(toIntUnsafe(o).Value() != 0), nil
}

func boolNew(f *Frame, _ *Type, args Args, _ KWArgs) (*Object, *BaseException) {
	argc := len(args)
	if argc == 0 {
		return False.ToObject(), nil
	}
	if argc != 1 {
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf("bool() takes at most 1 argument (%d given)", argc))
	}
	ret, raised := IsTrue(f, args[0])
	if raised != nil {
		return nil, raised
	}
	return GetBool(ret).ToObject(), nil
}

func boolRepr(_ *Frame, o *Object) (*Object, *BaseException) {
	i := toIntUnsafe(o)
	if i.Value() != 0 {
		return trueStr, nil
	}
	return falseStr, nil
}

func initBoolType(map[string]*Object) {
	BoolType.flags &= ^(typeFlagInstantiable | typeFlagBasetype)
	BoolType.slots.Native = &nativeSlot{boolNative}
	BoolType.slots.New = &newSlot{boolNew}
	BoolType.slots.Repr = &unaryOpSlot{boolRepr}
}

var (
	// True is the singleton bool object representing the Python 'True'
	// object.
	True = &Int{Object{typ: BoolType}, 1}
	// False is the singleton bool object representing the Python 'False'
	// object.
	False    = &Int{Object{typ: BoolType}, 0}
	trueStr  = NewStr("True").ToObject()
	falseStr = NewStr("False").ToObject()
)

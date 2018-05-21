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

// BaseException represents Python 'BaseException' objects.
type BaseException struct {
	Object
	args *Tuple
}

func toBaseExceptionUnsafe(o *Object) *BaseException {
	return (*BaseException)(o.toPointer())
}

// ToObject upcasts e to an Object.
func (e *BaseException) ToObject() *Object {
	return &e.Object
}

// BaseExceptionType corresponds to the Python type 'BaseException'.
var BaseExceptionType = newBasisType("BaseException", reflect.TypeOf(BaseException{}), toBaseExceptionUnsafe, ObjectType)

func baseExceptionInit(f *Frame, o *Object, args Args, kwargs KWArgs) (*Object, *BaseException) {
	e := toBaseExceptionUnsafe(o)
	e.args = NewTuple(args.makeCopy()...)
	return None, nil
}

func baseExceptionRepr(f *Frame, o *Object) (*Object, *BaseException) {
	e := toBaseExceptionUnsafe(o)
	argsString := "()"
	if e.args != nil {
		s, raised := Repr(f, e.args.ToObject())
		if raised != nil {
			return nil, raised
		}
		argsString = s.Value()
	}
	return NewStr(e.typ.Name() + argsString).ToObject(), nil
}

func baseExceptionStr(f *Frame, o *Object) (*Object, *BaseException) {
	e := toBaseExceptionUnsafe(o)
	if e.args == nil || len(e.args.elems) == 0 {
		return NewStr("").ToObject(), nil
	}
	if len(e.args.elems) == 1 {
		s, raised := ToStr(f, e.args.elems[0])
		return s.ToObject(), raised
	}
	s, raised := ToStr(f, e.args.ToObject())
	return s.ToObject(), raised
}

func initBaseExceptionType(map[string]*Object) {
	BaseExceptionType.slots.Init = &initSlot{baseExceptionInit}
	BaseExceptionType.slots.Repr = &unaryOpSlot{baseExceptionRepr}
	BaseExceptionType.slots.Str = &unaryOpSlot{baseExceptionStr}
}

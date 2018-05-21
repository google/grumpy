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
	"strings"
)

var (
	slotsType = reflect.TypeOf(typeSlots{})
	numSlots  = slotsType.NumField()
	slotNames = calcSlotNames()
)

type slot interface {
	// makeCallable returns a new callable object that forwards calls to
	// the receiving slot with the given slotName. It is used to populate
	// t's type dictionary so that slots are accessible from Python.
	makeCallable(t *Type, slotName string) *Object
	// wrapCallable updates the receiver slot to forward its calls to the
	// given callable. This method is called when a user defined type
	// defines a slot method in Python to override the slot.
	wrapCallable(callable *Object) bool
}

type basisSlot struct {
	Fn func(*Object) reflect.Value
}

func (s *basisSlot) makeCallable(t *Type, slotName string) *Object {
	return nil
}

func (s *basisSlot) wrapCallable(callable *Object) bool {
	return false
}

type binaryOpFunc func(*Frame, *Object, *Object) (*Object, *BaseException)

type binaryOpSlot struct {
	Fn binaryOpFunc
}

func (s *binaryOpSlot) makeCallable(t *Type, slotName string) *Object {
	return newBuiltinFunction(slotName, func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		if raised := checkMethodArgs(f, slotName, args, t, ObjectType); raised != nil {
			return nil, raised
		}
		return s.Fn(f, args[0], args[1])
	}).ToObject()
}

func (s *binaryOpSlot) wrapCallable(callable *Object) bool {
	s.Fn = func(f *Frame, v, w *Object) (*Object, *BaseException) {
		return callable.Call(f, Args{v, w}, nil)
	}
	return true
}

type callSlot struct {
	Fn func(*Frame, *Object, Args, KWArgs) (*Object, *BaseException)
}

func (s *callSlot) makeCallable(t *Type, _ string) *Object {
	return newBuiltinFunction("__call__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		if raised := checkMethodVarArgs(f, "__call__", args, t); raised != nil {
			return nil, raised
		}
		return t.slots.Call.Fn(f, args[0], args[1:], kwargs)
	}).ToObject()
}

func (s *callSlot) wrapCallable(callable *Object) bool {
	s.Fn = func(f *Frame, o *Object, args Args, kwargs KWArgs) (*Object, *BaseException) {
		callArgs := make(Args, len(args)+1)
		callArgs[0] = o
		copy(callArgs[1:], args)
		return callable.Call(f, callArgs, kwargs)
	}
	return true
}

type delAttrSlot struct {
	Fn func(*Frame, *Object, *Str) *BaseException
}

func (s *delAttrSlot) makeCallable(t *Type, slotName string) *Object {
	return newBuiltinFunction(slotName, func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		if raised := checkMethodArgs(f, slotName, args, t, StrType); raised != nil {
			return nil, raised
		}
		if raised := s.Fn(f, args[0], toStrUnsafe(args[1])); raised != nil {
			return nil, raised
		}
		return None, nil
	}).ToObject()
}

func (s *delAttrSlot) wrapCallable(callable *Object) bool {
	s.Fn = func(f *Frame, o *Object, name *Str) *BaseException {
		_, raised := callable.Call(f, Args{o, name.ToObject()}, nil)
		return raised
	}
	return true
}

type deleteSlot struct {
	Fn func(*Frame, *Object, *Object) *BaseException
}

func (s *deleteSlot) makeCallable(t *Type, slotName string) *Object {
	return newBuiltinFunction(slotName, func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		if raised := checkMethodArgs(f, slotName, args, t, ObjectType); raised != nil {
			return nil, raised
		}
		if raised := s.Fn(f, args[0], args[1]); raised != nil {
			return nil, raised
		}
		return None, nil
	}).ToObject()
}

func (s *deleteSlot) wrapCallable(callable *Object) bool {
	s.Fn = func(f *Frame, desc *Object, inst *Object) *BaseException {
		_, raised := callable.Call(f, Args{desc, inst}, nil)
		return raised
	}
	return true
}

type delItemSlot struct {
	Fn func(*Frame, *Object, *Object) *BaseException
}

func (s *delItemSlot) makeCallable(t *Type, slotName string) *Object {
	return newBuiltinFunction(slotName, func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		if raised := checkMethodArgs(f, slotName, args, t, ObjectType); raised != nil {
			return nil, raised
		}
		if raised := s.Fn(f, args[0], args[1]); raised != nil {
			return nil, raised
		}
		return None, nil
	}).ToObject()
}

func (s *delItemSlot) wrapCallable(callable *Object) bool {
	s.Fn = func(f *Frame, o *Object, key *Object) *BaseException {
		_, raised := callable.Call(f, Args{o, key}, nil)
		return raised
	}
	return true
}

type getAttributeSlot struct {
	Fn func(*Frame, *Object, *Str) (*Object, *BaseException)
}

func (s *getAttributeSlot) makeCallable(t *Type, slotName string) *Object {
	return newBuiltinFunction(slotName, func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		if raised := checkMethodArgs(f, slotName, args, t, StrType); raised != nil {
			return nil, raised
		}
		return s.Fn(f, args[0], toStrUnsafe(args[1]))
	}).ToObject()
}

func (s *getAttributeSlot) wrapCallable(callable *Object) bool {
	s.Fn = func(f *Frame, o *Object, name *Str) (*Object, *BaseException) {
		return callable.Call(f, Args{o, name.ToObject()}, nil)
	}
	return true
}

type getSlot struct {
	Fn func(*Frame, *Object, *Object, *Type) (*Object, *BaseException)
}

func (s *getSlot) makeCallable(t *Type, slotName string) *Object {
	return newBuiltinFunction(slotName, func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		if raised := checkMethodArgs(f, slotName, args, t, ObjectType, TypeType); raised != nil {
			return nil, raised
		}
		return s.Fn(f, args[0], args[1], toTypeUnsafe(args[2]))
	}).ToObject()
}

func (s *getSlot) wrapCallable(callable *Object) bool {
	s.Fn = func(f *Frame, desc, inst *Object, owner *Type) (*Object, *BaseException) {
		return callable.Call(f, Args{desc, inst, owner.ToObject()}, nil)
	}
	return true
}

type initSlot struct {
	Fn func(*Frame, *Object, Args, KWArgs) (*Object, *BaseException)
}

func (s *initSlot) makeCallable(t *Type, _ string) *Object {
	return newBuiltinFunction("__init__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		if raised := checkMethodVarArgs(f, "__init__", args, t); raised != nil {
			return nil, raised
		}
		return t.slots.Init.Fn(f, args[0], args[1:], kwargs)
	}).ToObject()
}

func (s *initSlot) wrapCallable(callable *Object) bool {
	s.Fn = func(f *Frame, o *Object, args Args, kwargs KWArgs) (*Object, *BaseException) {
		callArgs := make(Args, len(args)+1)
		callArgs[0] = o
		copy(callArgs[1:], args)
		return callable.Call(f, callArgs, kwargs)
	}
	return true
}

type nativeSlot struct {
	Fn func(*Frame, *Object) (reflect.Value, *BaseException)
}

func (s *nativeSlot) makeCallable(t *Type, slotName string) *Object {
	return nil
}

func (s *nativeSlot) wrapCallable(callable *Object) bool {
	return false
}

type newSlot struct {
	Fn func(*Frame, *Type, Args, KWArgs) (*Object, *BaseException)
}

func (s *newSlot) makeCallable(t *Type, _ string) *Object {
	return newStaticMethod(newBuiltinFunction("__new__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		if raised := checkFunctionVarArgs(f, "__new__", args, TypeType); raised != nil {
			return nil, raised
		}
		typeArg := toTypeUnsafe(args[0])
		if !typeArg.isSubclass(t) {
			format := "%[1]s.__new__(%[2]s): %[2]s is not a subtype of %[1]s"
			return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, t.Name(), typeArg.Name()))
		}
		return t.slots.New.Fn(f, typeArg, args[1:], kwargs)
	}).ToObject()).ToObject()
}

func (s *newSlot) wrapCallable(callable *Object) bool {
	s.Fn = func(f *Frame, t *Type, args Args, kwargs KWArgs) (*Object, *BaseException) {
		callArgs := make(Args, len(args)+1)
		callArgs[0] = t.ToObject()
		copy(callArgs[1:], args)
		return callable.Call(f, callArgs, kwargs)
	}
	return true
}

type setAttrSlot struct {
	Fn func(*Frame, *Object, *Str, *Object) *BaseException
}

func (s *setAttrSlot) makeCallable(t *Type, slotName string) *Object {
	return newBuiltinFunction(slotName, func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		if raised := checkMethodArgs(f, slotName, args, t, StrType, ObjectType); raised != nil {
			return nil, raised
		}
		if raised := s.Fn(f, args[0], toStrUnsafe(args[1]), args[2]); raised != nil {
			return nil, raised
		}
		return None, nil
	}).ToObject()
}

func (s *setAttrSlot) wrapCallable(callable *Object) bool {
	s.Fn = func(f *Frame, o *Object, name *Str, value *Object) *BaseException {
		_, raised := callable.Call(f, Args{o, name.ToObject(), value}, nil)
		return raised
	}
	return true
}

type setItemSlot struct {
	Fn func(*Frame, *Object, *Object, *Object) *BaseException
}

func (s *setItemSlot) makeCallable(t *Type, slotName string) *Object {
	return newBuiltinFunction(slotName, func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		if raised := checkMethodArgs(f, slotName, args, t, ObjectType, ObjectType); raised != nil {
			return nil, raised
		}
		if raised := s.Fn(f, args[0], args[1], args[2]); raised != nil {
			return nil, raised
		}
		return None, nil
	}).ToObject()
}

func (s *setItemSlot) wrapCallable(callable *Object) bool {
	s.Fn = func(f *Frame, o *Object, key *Object, value *Object) *BaseException {
		_, raised := callable.Call(f, Args{o, key, value}, nil)
		return raised
	}
	return true
}

type setSlot struct {
	Fn func(*Frame, *Object, *Object, *Object) *BaseException
}

func (s *setSlot) makeCallable(t *Type, slotName string) *Object {
	return newBuiltinFunction(slotName, func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		if raised := checkMethodArgs(f, slotName, args, t, ObjectType, ObjectType); raised != nil {
			return nil, raised
		}
		if raised := s.Fn(f, args[0], args[1], args[2]); raised != nil {
			return nil, raised
		}
		return None, nil
	}).ToObject()
}

func (s *setSlot) wrapCallable(callable *Object) bool {
	s.Fn = func(f *Frame, desc, inst, value *Object) *BaseException {
		_, raised := callable.Call(f, Args{desc, inst, value}, nil)
		return raised
	}
	return true
}

type unaryOpSlot struct {
	Fn func(*Frame, *Object) (*Object, *BaseException)
}

func (s *unaryOpSlot) makeCallable(t *Type, slotName string) *Object {
	return newBuiltinFunction(slotName, func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		if raised := checkMethodArgs(f, slotName, args, t); raised != nil {
			return nil, raised
		}
		return s.Fn(f, args[0])
	}).ToObject()
}

func (s *unaryOpSlot) wrapCallable(callable *Object) bool {
	s.Fn = func(f *Frame, o *Object) (*Object, *BaseException) {
		return callable.Call(f, Args{o}, nil)
	}
	return true
}

// typeSlots hold a type's special methods such as __eq__. During type
// initialization, any field that is not set for that type will be inherited
// according to the type's MRO. Therefore, any given field will be nil only if
// that method is not defined for the type nor any of its super classes.  Each
// slot is expected to be a pointer to a struct with a single function field.
// The wrapper structs permit comparison of like slots which is occasionally
// necessary to determine whether a function has been overridden by a subclass.
type typeSlots struct {
	Abs          *unaryOpSlot
	Add          *binaryOpSlot
	And          *binaryOpSlot
	Basis        *basisSlot
	Call         *callSlot
	Cmp          *binaryOpSlot
	Complex      *unaryOpSlot
	Contains     *binaryOpSlot
	DelAttr      *delAttrSlot
	Delete       *deleteSlot
	DelItem      *delItemSlot
	Div          *binaryOpSlot
	DivMod       *binaryOpSlot
	Eq           *binaryOpSlot
	Float        *unaryOpSlot
	FloorDiv     *binaryOpSlot
	GE           *binaryOpSlot
	Get          *getSlot
	GetAttribute *getAttributeSlot
	GetItem      *binaryOpSlot
	GT           *binaryOpSlot
	Hash         *unaryOpSlot
	Hex          *unaryOpSlot
	IAdd         *binaryOpSlot
	IAnd         *binaryOpSlot
	IDiv         *binaryOpSlot
	IDivMod      *binaryOpSlot
	IFloorDiv    *binaryOpSlot
	ILShift      *binaryOpSlot
	IMod         *binaryOpSlot
	IMul         *binaryOpSlot
	Index        *unaryOpSlot
	Init         *initSlot
	Int          *unaryOpSlot
	Invert       *unaryOpSlot
	IOr          *binaryOpSlot
	IPow         *binaryOpSlot
	IRShift      *binaryOpSlot
	ISub         *binaryOpSlot
	Iter         *unaryOpSlot
	IXor         *binaryOpSlot
	LE           *binaryOpSlot
	Len          *unaryOpSlot
	Long         *unaryOpSlot
	LShift       *binaryOpSlot
	LT           *binaryOpSlot
	Mod          *binaryOpSlot
	Mul          *binaryOpSlot
	Native       *nativeSlot
	NE           *binaryOpSlot
	Neg          *unaryOpSlot
	New          *newSlot
	Next         *unaryOpSlot
	NonZero      *unaryOpSlot
	Oct          *unaryOpSlot
	Or           *binaryOpSlot
	Pos          *unaryOpSlot
	Pow          *binaryOpSlot
	RAdd         *binaryOpSlot
	RAnd         *binaryOpSlot
	RDiv         *binaryOpSlot
	RDivMod      *binaryOpSlot
	Repr         *unaryOpSlot
	RFloorDiv    *binaryOpSlot
	RLShift      *binaryOpSlot
	RMod         *binaryOpSlot
	RMul         *binaryOpSlot
	ROr          *binaryOpSlot
	RPow         *binaryOpSlot
	RRShift      *binaryOpSlot
	RShift       *binaryOpSlot
	RSub         *binaryOpSlot
	RXor         *binaryOpSlot
	Set          *setSlot
	SetAttr      *setAttrSlot
	SetItem      *setItemSlot
	Str          *unaryOpSlot
	Sub          *binaryOpSlot
	Unicode      *unaryOpSlot
	Xor          *binaryOpSlot
}

func calcSlotNames() []string {
	names := make([]string, numSlots, numSlots)
	for i := 0; i < numSlots; i++ {
		field := slotsType.Field(i)
		name := ""
		if field.Name == "Next" {
			name = "next"
		} else {
			name = fmt.Sprintf("__%s__", strings.ToLower(field.Name))
		}
		names[i] = name
	}
	return names
}

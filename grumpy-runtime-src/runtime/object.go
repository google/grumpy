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
	"sync/atomic"
	"unsafe"
)

var (
	objectBasis             = reflect.TypeOf(Object{})
	objectReconstructorFunc = newBuiltinFunction("_reconstructor", objectReconstructor).ToObject()
	objectReduceFunc        = newBuiltinFunction("__reduce__", objectReduce).ToObject()
	// ObjectType is the object representing the Python 'object' type.
	//
	// We don't use newBasisType() here since that introduces an initialization
	// cycle between TypeType and ObjectType.
	ObjectType = &Type{
		name:  "object",
		basis: objectBasis,
		flags: typeFlagDefault,
		slots: typeSlots{Basis: &basisSlot{objectBasisFunc}},
	}
)

// Object represents Python 'object' objects.
type Object struct {
	typ  *Type `attr:"__class__"`
	dict *Dict
	ref  *WeakRef
}

func newObject(t *Type) *Object {
	var dict *Dict
	if t != ObjectType {
		dict = NewDict()
	}
	o := (*Object)(unsafe.Pointer(reflect.New(t.basis).Pointer()))
	o.typ = t
	o.setDict(dict)
	return o
}

// Call invokes the callable Python object o with the given positional and
// keyword args. args must be non-nil (but can be empty). kwargs can be nil.
func (o *Object) Call(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	call := o.Type().slots.Call
	if call == nil {
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf("'%s' object is not callable", o.Type().Name()))
	}
	return call.Fn(f, o, args, kwargs)
}

// Dict returns o's object dict, aka __dict__.
func (o *Object) Dict() *Dict {
	p := (*unsafe.Pointer)(unsafe.Pointer(&o.dict))
	return (*Dict)(atomic.LoadPointer(p))
}

func (o *Object) setDict(d *Dict) {
	p := (*unsafe.Pointer)(unsafe.Pointer(&o.dict))
	atomic.StorePointer(p, unsafe.Pointer(d))
}

// String returns a string representation of o, e.g. for debugging.
func (o *Object) String() string {
	if o == nil {
		return "nil"
	}
	s, raised := Repr(NewRootFrame(), o)
	if raised != nil {
		return fmt.Sprintf("<%s object (repr raised %s)>", o.typ.Name(), raised.typ.Name())
	}
	return s.Value()
}

// Type returns the Python type of o.
func (o *Object) Type() *Type {
	return o.typ
}

func (o *Object) toPointer() unsafe.Pointer {
	return unsafe.Pointer(o)
}

func (o *Object) isInstance(t *Type) bool {
	return o.typ.isSubclass(t)
}

func objectBasisFunc(o *Object) reflect.Value {
	return reflect.ValueOf(o).Elem()
}

func objectDelAttr(f *Frame, o *Object, name *Str) *BaseException {
	desc, raised := o.typ.mroLookup(f, name)
	if raised != nil {
		return raised
	}
	if desc != nil {
		if del := desc.Type().slots.Delete; del != nil {
			return del.Fn(f, desc, o)
		}
	}
	deleted := false
	d := o.Dict()
	if d != nil {
		deleted, raised = d.DelItem(f, name.ToObject())
		if raised != nil {
			return raised
		}
	}
	if !deleted {
		format := "'%s' object has no attribute '%s'"
		return f.RaiseType(AttributeErrorType, fmt.Sprintf(format, o.typ.Name(), name.Value()))
	}
	return nil
}

// objectGetAttribute implements the spec here:
// https://docs.python.org/2/reference/datamodel.html#invoking-descriptors
func objectGetAttribute(f *Frame, o *Object, name *Str) (*Object, *BaseException) {
	// Look for a data descriptor in the type.
	var typeGet *getSlot
	typeAttr, raised := o.typ.mroLookup(f, name)
	if raised != nil {
		return nil, raised
	}
	if typeAttr != nil {
		typeGet = typeAttr.typ.slots.Get
		if typeGet != nil && (typeAttr.typ.slots.Set != nil || typeAttr.typ.slots.Delete != nil) {
			return typeGet.Fn(f, typeAttr, o, o.Type())
		}
	}
	// Look in the object's dict.
	if d := o.Dict(); d != nil {
		value, raised := d.GetItem(f, name.ToObject())
		if value != nil || raised != nil {
			return value, raised
		}
	}
	// Use the (non-data) descriptor from the type.
	if typeGet != nil {
		return typeGet.Fn(f, typeAttr, o, o.Type())
	}
	// Return the ordinary type attribute.
	if typeAttr != nil {
		return typeAttr, nil
	}
	format := "'%s' object has no attribute '%s'"
	return nil, f.RaiseType(AttributeErrorType, fmt.Sprintf(format, o.typ.Name(), name.Value()))
}

func objectHash(f *Frame, o *Object) (*Object, *BaseException) {
	return NewInt(int(uintptr(o.toPointer()))).ToObject(), nil
}

func objectNew(f *Frame, t *Type, _ Args, _ KWArgs) (*Object, *BaseException) {
	if t.flags&typeFlagInstantiable == 0 {
		format := "object.__new__(%s) is not safe, use %s.__new__()"
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, t.Name(), t.Name()))
	}
	return newObject(t), nil
}

func objectReduce(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	expectedTypes := []*Type{ObjectType, IntType}
	argc := len(args)
	if argc == 1 {
		expectedTypes = expectedTypes[:1]
	}
	if raised := checkMethodArgs(f, "__reduce__", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	return objectReduceCommon(f, args)
}

func objectReduceEx(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	expectedTypes := []*Type{ObjectType, IntType}
	argc := len(args)
	if argc == 1 {
		expectedTypes = expectedTypes[:1]
	}
	if raised := checkMethodArgs(f, "__reduce_ex__", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	reduce, raised := args[0].typ.mroLookup(f, NewStr("__reduce__"))
	if raised != nil {
		return nil, raised
	}
	if reduce != nil && reduce != objectReduceFunc {
		// __reduce__ is overridden so prefer using it.
		return reduce.Call(f, args, nil)
	}
	return objectReduceCommon(f, args)
}

func objectSetAttr(f *Frame, o *Object, name *Str, value *Object) *BaseException {
	if typeAttr, raised := o.typ.mroLookup(f, name); raised != nil {
		return raised
	} else if typeAttr != nil {
		if typeSet := typeAttr.typ.slots.Set; typeSet != nil {
			return typeSet.Fn(f, typeAttr, o, value)
		}
	}
	if d := o.Dict(); d != nil {
		if raised := d.SetItem(f, name.ToObject(), value); raised == nil || !raised.isInstance(KeyErrorType) {
			return nil
		}
	}
	return f.RaiseType(AttributeErrorType, fmt.Sprintf("'%s' has no attribute '%s'", o.typ.Name(), name.Value()))
}

func initObjectType(dict map[string]*Object) {
	ObjectType.typ = TypeType
	dict["__reduce__"] = objectReduceFunc
	dict["__reduce_ex__"] = newBuiltinFunction("__reduce_ex__", objectReduceEx).ToObject()
	dict["__dict__"] = newProperty(newBuiltinFunction("_get_dict", objectGetDict).ToObject(), newBuiltinFunction("_set_dict", objectSetDict).ToObject(), nil).ToObject()
	ObjectType.slots.DelAttr = &delAttrSlot{objectDelAttr}
	ObjectType.slots.GetAttribute = &getAttributeSlot{objectGetAttribute}
	ObjectType.slots.Hash = &unaryOpSlot{objectHash}
	ObjectType.slots.New = &newSlot{objectNew}
	ObjectType.slots.SetAttr = &setAttrSlot{objectSetAttr}
}

// objectReconstructor builds an object from a class, its basis type and its
// state (e.g. its string or integer value). It is similar to the
// copy_reg._reconstructor function in CPython.
func objectReconstructor(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkFunctionArgs(f, "_reconstructor", args, TypeType, TypeType, ObjectType); raised != nil {
		return nil, raised
	}
	t, basisType, state := toTypeUnsafe(args[0]), toTypeUnsafe(args[1]), args[2]
	newMethod, raised := GetAttr(f, basisType.ToObject(), NewStr("__new__"), nil)
	if raised != nil {
		return nil, raised
	}
	o, raised := newMethod.Call(f, Args{t.ToObject(), state}, nil)
	if raised != nil {
		return nil, raised
	}
	if basisType != ObjectType {
		initMethod, raised := GetAttr(f, basisType.ToObject(), NewStr("__init__"), None)
		if raised != nil {
			return nil, raised
		}
		if initMethod != None {
			if _, raised := initMethod.Call(f, Args{o, state}, nil); raised != nil {
				return nil, raised
			}
		}
	}
	return o, nil
}

func objectReduceCommon(f *Frame, args Args) (*Object, *BaseException) {
	// TODO: Support __getstate__ and __getnewargs__.
	o := args[0]
	t := o.Type()
	proto := 0
	if len(args) > 1 {
		proto = toIntUnsafe(args[1]).Value()
	}
	var raised *BaseException
	if proto < 2 {
		basisType := basisTypes[t.basis]
		if basisType == t {
			// Basis types are handled elsewhere by the pickle and
			// copy frameworks. This matches behavior in
			// copy_reg._reduce_ex in CPython.
			return nil, f.RaiseType(TypeErrorType, fmt.Sprintf("can't pickle %s objects", t.Name()))
		}
		state := None
		if basisType != ObjectType {
			// For subclasses of basis types having state (e.g.
			// integer values), the state is captured by creating
			// an instance of that basis type.
			if state, raised = basisType.Call(f, Args{o}, nil); raised != nil {
				return nil, raised
			}
		}
		newArgs := NewTuple3(t.ToObject(), basisType.ToObject(), state).ToObject()
		if d := o.Dict(); d != nil {
			return NewTuple3(objectReconstructorFunc, newArgs, d.ToObject()).ToObject(), nil
		}
		return NewTuple2(objectReconstructorFunc, newArgs).ToObject(), nil
	}
	newArgs := []*Object{t.ToObject()}
	getNewArgsMethod, raised := GetAttr(f, o, NewStr("__getnewargs__"), None)
	if raised != nil {
		return nil, raised
	}
	if getNewArgsMethod != None {
		extraNewArgs, raised := getNewArgsMethod.Call(f, nil, nil)
		if raised != nil {
			return nil, raised
		}
		if !extraNewArgs.isInstance(TupleType) {
			format := "__getnewargs__ should return a tuple, not '%s'"
			return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, extraNewArgs.Type().Name()))
		}
		newArgs = append(newArgs, toTupleUnsafe(extraNewArgs).elems...)
	}
	dict := None
	if d := o.Dict(); d != nil {
		dict = d.ToObject()
	}
	// For proto >= 2 include list and dict items.
	listItems := None
	if o.isInstance(ListType) {
		if listItems, raised = Iter(f, o); raised != nil {
			return nil, raised
		}
	}
	dictItems := None
	if o.isInstance(DictType) {
		iterItems, raised := o.typ.mroLookup(f, NewStr("iteritems"))
		if raised != nil {
			return nil, raised
		}
		if iterItems != nil {
			if dictItems, raised = iterItems.Call(f, Args{o}, nil); raised != nil {
				return nil, raised
			}
		}
	}
	newFunc, raised := GetAttr(f, t.ToObject(), NewStr("__new__"), nil)
	if raised != nil {
		return nil, raised
	}
	return NewTuple5(newFunc, NewTuple(newArgs...).ToObject(), dict, listItems, dictItems).ToObject(), nil
}

func objectGetDict(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "_get_dict", args, ObjectType); raised != nil {
		return nil, raised
	}
	o := args[0]
	d := o.Dict()
	if d == nil {
		format := "'%s' object has no attribute '__dict__'"
		return nil, f.RaiseType(AttributeErrorType, fmt.Sprintf(format, o.typ.Name()))
	}
	return args[0].Dict().ToObject(), nil
}

func objectSetDict(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "_set_dict", args, ObjectType, DictType); raised != nil {
		return nil, raised
	}
	o := args[0]
	if o.Type() == ObjectType {
		format := "'%s' object has no attribute '__dict__'"
		return nil, f.RaiseType(AttributeErrorType, fmt.Sprintf(format, o.typ.Name()))
	}
	o.setDict(toDictUnsafe(args[1]))
	return None, nil
}

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

type typeFlag int

const (
	// Set when instances can be created via __new__. This is the default.
	// We need to be able to prohibit instantiation of certain internal
	// types like NoneType. CPython accomplishes this via tp_new == NULL but
	// we don't have a tp_new.
	typeFlagInstantiable typeFlag = 1 << iota
	// Set when the type can be used as a base class. This is the default.
	// Corresponds to the Py_TPFLAGS_BASETYPE flag in CPython.
	typeFlagBasetype typeFlag = 1 << iota
	typeFlagDefault           = typeFlagInstantiable | typeFlagBasetype
)

// Type represents Python 'type' objects.
type Type struct {
	Object
	name  string `attr:"__name__"`
	basis reflect.Type
	bases []*Type
	mro   []*Type
	flags typeFlag
	slots typeSlots
}

var basisTypes = map[reflect.Type]*Type{
	objectBasis: ObjectType,
	typeBasis:   TypeType,
}

// newClass creates a Python type with the given name, base classes and type
// dict. It is similar to the Python expression 'type(name, bases, dict)'.
func newClass(f *Frame, meta *Type, name string, bases []*Type, dict *Dict) (*Type, *BaseException) {
	numBases := len(bases)
	if numBases == 0 {
		return nil, f.RaiseType(TypeErrorType, "class must have base classes")
	}
	var basis reflect.Type
	for _, base := range bases {
		if base.flags&typeFlagBasetype == 0 {
			format := "type '%s' is not an acceptable base type"
			return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, base.Name()))
		}
		basis = basisSelect(basis, base.basis)
	}
	if basis == nil {
		return nil, f.RaiseType(TypeErrorType, "class layout error")
	}
	t := newType(meta, name, basis, bases, dict)
	// Populate slots for any special methods overridden in dict.
	slotsValue := reflect.ValueOf(&t.slots).Elem()
	for i := 0; i < numSlots; i++ {
		dictFunc, raised := dict.GetItemString(f, slotNames[i])
		if raised != nil {
			return nil, raised
		}
		if dictFunc != nil {
			slotField := slotsValue.Field(i)
			slotValue := reflect.New(slotField.Type().Elem())
			if slotValue.Interface().(slot).wrapCallable(dictFunc) {
				slotField.Set(slotValue)
			}
		}
	}
	if err := prepareType(t); err != "" {
		return nil, f.RaiseType(TypeErrorType, err)
	}
	// Set the __module__ attr if it's not already specified.
	mod, raised := dict.GetItemString(f, "__module__")
	if raised != nil {
		return nil, raised
	}
	if mod == nil {
		if raised := dict.SetItemString(f, "__module__", builtinStr.ToObject()); raised != nil {
			return nil, raised
		}
	}
	return t, nil
}

func newType(meta *Type, name string, basis reflect.Type, bases []*Type, dict *Dict) *Type {
	return &Type{
		Object: Object{typ: meta, dict: dict},
		name:   name,
		basis:  basis,
		bases:  bases,
		flags:  typeFlagDefault,
	}
}

func newBasisType(name string, basis reflect.Type, basisFunc interface{}, base *Type) *Type {
	if _, ok := basisTypes[basis]; ok {
		logFatal(fmt.Sprintf("type for basis already exists: %s", basis))
	}
	if basis.Kind() != reflect.Struct {
		logFatal(fmt.Sprintf("basis must be a struct not: %s", basis.Kind()))
	}
	if basis.NumField() == 0 {
		logFatal(fmt.Sprintf("1st field of basis must be base type's basis"))
	}
	if basis.Field(0).Type != base.basis {
		logFatal(fmt.Sprintf("1st field of basis must be base type's basis not: %s", basis.Field(0).Type))
	}
	basisFuncValue := reflect.ValueOf(basisFunc)
	basisFuncType := basisFuncValue.Type()
	if basisFuncValue.Kind() != reflect.Func || basisFuncType.NumIn() != 1 || basisFuncType.NumOut() != 1 ||
		basisFuncType.In(0) != reflect.PtrTo(objectBasis) || basisFuncType.Out(0) != reflect.PtrTo(basis) {
		logFatal(fmt.Sprintf("expected basis func of type func(*Object) *%s", nativeTypeName(basis)))
	}
	t := newType(TypeType, name, basis, []*Type{base}, nil)
	t.slots.Basis = &basisSlot{func(o *Object) reflect.Value {
		return basisFuncValue.Call([]reflect.Value{reflect.ValueOf(o)})[0].Elem()
	}}
	basisTypes[basis] = t
	return t
}

func newSimpleType(name string, base *Type) *Type {
	return newType(TypeType, name, base.basis, []*Type{base}, nil)
}

// prepareBuiltinType initializes the builtin typ by populating dict with
// struct field descriptors and slot wrappers, and then calling prepareType.
func prepareBuiltinType(typ *Type, init builtinTypeInit) {
	dict := map[string]*Object{"__module__": builtinStr.ToObject()}
	if init != nil {
		init(dict)
	}
	// For basis types, export field descriptors.
	if basis := typ.basis; basisTypes[basis] == typ {
		numFields := basis.NumField()
		for i := 0; i < numFields; i++ {
			field := basis.Field(i)
			if attr := field.Tag.Get("attr"); attr != "" {
				fieldMode := fieldDescriptorRO
				if mode := field.Tag.Get("attr_mode"); mode == "rw" {
					fieldMode = fieldDescriptorRW
				}
				dict[attr] = makeStructFieldDescriptor(typ, field.Name, attr, fieldMode)
			}
		}
	}
	// Create dict entries for slot methods.
	slotsValue := reflect.ValueOf(&typ.slots).Elem()
	for i := 0; i < numSlots; i++ {
		slotField := slotsValue.Field(i)
		if !slotField.IsNil() {
			slot := slotField.Interface().(slot)
			if fun := slot.makeCallable(typ, slotNames[i]); fun != nil {
				dict[slotNames[i]] = fun
			}
		}
	}
	typ.setDict(newStringDict(dict))
	if err := prepareType(typ); err != "" {
		logFatal(err)
	}
}

// prepareType calculates typ's mro and inherits its flags and slots from its
// base classes.
func prepareType(typ *Type) string {
	typ.mro = mroCalc(typ)
	if typ.mro == nil {
		return fmt.Sprintf("mro error for: %s", typ.name)
	}
	for _, base := range typ.mro {
		if base.flags&typeFlagInstantiable == 0 {
			typ.flags &^= typeFlagInstantiable
		}
		if base.flags&typeFlagBasetype == 0 {
			typ.flags &^= typeFlagBasetype
		}
	}
	// Inherit slots from typ's mro.
	slotsValue := reflect.ValueOf(&typ.slots).Elem()
	for i := 0; i < numSlots; i++ {
		slotField := slotsValue.Field(i)
		if slotField.IsNil() {
			for _, base := range typ.mro {
				baseSlotFunc := reflect.ValueOf(base.slots).Field(i)
				if !baseSlotFunc.IsNil() {
					slotField.Set(baseSlotFunc)
					break
				}
			}
		}
	}
	return ""
}

// Precondition: At least one of seqs is non-empty.
func mroMerge(seqs [][]*Type) []*Type {
	var res []*Type
	numSeqs := len(seqs)
	hasNonEmptySeqs := true
	for hasNonEmptySeqs {
		var cand *Type
		for i := 0; i < numSeqs && cand == nil; i++ {
			// The next candidate will be absent from or at the head
			// of all lists. If we try a candidate and we find it's
			// somewhere past the head of one of the lists, reject.
			seq := seqs[i]
			if len(seq) == 0 {
				continue
			}
			cand = seq[0]
		RejectCandidate:
			for _, seq := range seqs {
				numElems := len(seq)
				for j := 1; j < numElems; j++ {
					if seq[j] == cand {
						cand = nil
						break RejectCandidate
					}
				}
			}
		}
		if cand == nil {
			// We could not find a candidate meaning that the
			// hierarchy is inconsistent.
			return nil
		}
		res = append(res, cand)
		hasNonEmptySeqs = false
		for i, seq := range seqs {
			if len(seq) > 0 {
				if seq[0] == cand {
					// Remove the candidate from all lists
					// (it will only be found at the head of
					// any list because otherwise it would
					// have been rejected above.)
					seqs[i] = seq[1:]
				}
				if len(seqs[i]) > 0 {
					hasNonEmptySeqs = true
				}
			}
		}
	}
	return res
}

func mroCalc(t *Type) []*Type {
	seqs := [][]*Type{{t}}
	for _, b := range t.bases {
		seqs = append(seqs, b.mro)
	}
	seqs = append(seqs, t.bases)
	return mroMerge(seqs)
}

func toTypeUnsafe(o *Object) *Type {
	return (*Type)(o.toPointer())
}

// ToObject upcasts t to an Object.
func (t *Type) ToObject() *Object {
	return &t.Object
}

// Name returns t's name field.
func (t *Type) Name() string {
	return t.name
}

// FullName returns t's fully qualified name including the module.
func (t *Type) FullName(f *Frame) (string, *BaseException) {
	moduleAttr, raised := t.Dict().GetItemString(f, "__module__")
	if raised != nil {
		return "", raised
	}
	if moduleAttr == nil {
		return t.Name(), nil
	}
	if moduleAttr.isInstance(StrType) {
		if s := toStrUnsafe(moduleAttr).Value(); s != "__builtin__" {
			return fmt.Sprintf("%s.%s", s, t.Name()), nil
		}
	}
	return t.Name(), nil
}

func (t *Type) isSubclass(super *Type) bool {
	for _, b := range t.mro {
		if b == super {
			return true
		}
	}
	return false
}

func (t *Type) mroLookup(f *Frame, name *Str) (*Object, *BaseException) {
	for _, t := range t.mro {
		v, raised := t.Dict().GetItem(f, name.ToObject())
		if v != nil || raised != nil {
			return v, raised
		}
	}
	return nil, nil
}

var typeBasis = reflect.TypeOf(Type{})

func typeBasisFunc(o *Object) reflect.Value {
	return reflect.ValueOf(toTypeUnsafe(o)).Elem()
}

// TypeType is the object representing the Python 'type' type.
//
// Don't use newType() since that depends on the initialization of
// TypeType.
var TypeType = &Type{
	name:  "type",
	basis: typeBasis,
	bases: []*Type{ObjectType},
	flags: typeFlagDefault,
	slots: typeSlots{Basis: &basisSlot{typeBasisFunc}},
}

func typeCall(f *Frame, callable *Object, args Args, kwargs KWArgs) (*Object, *BaseException) {
	t := toTypeUnsafe(callable)
	newFunc := t.slots.New
	if newFunc == nil {
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf("type %s has no __new__", t.Name()))
	}
	o, raised := newFunc.Fn(f, t, args, kwargs)
	if raised != nil {
		return nil, raised
	}
	if init := o.Type().slots.Init; init != nil {
		if _, raised := init.Fn(f, o, args, kwargs); raised != nil {
			return nil, raised
		}
	}
	return o, nil
}

// typeGetAttribute is very similar to objectGetAttribute except that it uses
// MRO to resolve dict attributes rather than just the type's own dict and the
// exception message is slightly different.
func typeGetAttribute(f *Frame, o *Object, name *Str) (*Object, *BaseException) {
	t := toTypeUnsafe(o)
	// Look for a data descriptor in the metaclass.
	var metaGet *getSlot
	metaType := t.typ
	metaAttr, raised := metaType.mroLookup(f, name)
	if raised != nil {
		return nil, raised
	}
	if metaAttr != nil {
		metaGet = metaAttr.typ.slots.Get
		if metaGet != nil && (metaAttr.typ.slots.Set != nil || metaAttr.typ.slots.Delete != nil) {
			return metaGet.Fn(f, metaAttr, t.ToObject(), metaType)
		}
	}
	// Look in dict of this type and its bases.
	attr, raised := t.mroLookup(f, name)
	if raised != nil {
		return nil, raised
	}
	if attr != nil {
		if get := attr.typ.slots.Get; get != nil {
			return get.Fn(f, attr, None, t)
		}
		return attr, nil
	}
	// Use the (non-data) descriptor from the metaclass.
	if metaGet != nil {
		return metaGet.Fn(f, metaAttr, t.ToObject(), metaType)
	}
	// Return the ordinary attr from metaclass.
	if metaAttr != nil {
		return metaAttr, nil
	}
	msg := fmt.Sprintf("type object '%s' has no attribute '%s'", t.Name(), name.Value())
	return nil, f.RaiseType(AttributeErrorType, msg)
}

func typeNew(f *Frame, t *Type, args Args, kwargs KWArgs) (*Object, *BaseException) {
	switch len(args) {
	case 0:
		return nil, f.RaiseType(TypeErrorType, "type() takes 1 or 3 arguments")
	case 1:
		return args[0].typ.ToObject(), nil
	}
	// case 3+
	if raised := checkMethodArgs(f, "__new__", args, StrType, TupleType, DictType); raised != nil {
		return nil, raised
	}
	name := toStrUnsafe(args[0]).Value()
	bases := toTupleUnsafe(args[1]).elems
	dict := toDictUnsafe(args[2])
	baseTypes := make([]*Type, len(bases))
	meta := t
	for i, o := range bases {
		if !o.isInstance(TypeType) {
			s, raised := Repr(f, o)
			if raised != nil {
				return nil, raised
			}
			return nil, f.RaiseType(TypeErrorType, fmt.Sprintf("not a valid base class: %s", s.Value()))
		}
		// Choose the most derived metaclass among all the bases to be
		// the metaclass for the new type.
		if o.typ.isSubclass(meta) {
			meta = o.typ
		} else if !meta.isSubclass(o.typ) {
			msg := "metaclass conflict: the metaclass of a derived class must " +
				"a be a (non-strict) subclass of the metaclasses of all its bases"
			return nil, f.RaiseType(TypeErrorType, msg)
		}
		baseTypes[i] = toTypeUnsafe(o)
	}
	ret, raised := newClass(f, meta, name, baseTypes, dict)
	if raised != nil {
		return nil, raised
	}
	return ret.ToObject(), nil
}

func typeRepr(f *Frame, o *Object) (*Object, *BaseException) {
	s, raised := toTypeUnsafe(o).FullName(f)
	if raised != nil {
		return nil, raised
	}
	return NewStr(fmt.Sprintf("<type '%s'>", s)).ToObject(), nil
}

func initTypeType(map[string]*Object) {
	TypeType.typ = TypeType
	TypeType.slots.Call = &callSlot{typeCall}
	TypeType.slots.GetAttribute = &getAttributeSlot{typeGetAttribute}
	TypeType.slots.New = &newSlot{typeNew}
	TypeType.slots.Repr = &unaryOpSlot{typeRepr}
}

// basisParent returns the immediate ancestor of basis, which is its first
// field. Returns nil when basis is objectBasis (the root of basis hierarchy.)
func basisParent(basis reflect.Type) reflect.Type {
	if basis == objectBasis {
		return nil
	}
	return basis.Field(0).Type
}

// basisSelect returns b1 if b2 inherits from it, b2 if b1 inherits from b2,
// otherwise nil. b1 can be nil in which case b2 is always returned.
func basisSelect(b1, b2 reflect.Type) reflect.Type {
	if b1 == nil {
		return b2
	}
	// Search up b1's inheritance chain to see if b2 is present.
	basis := b1
	for basis != nil && basis != b2 {
		basis = basisParent(basis)
	}
	if basis != nil {
		return b1
	}
	// Search up b2's inheritance chain to see if b1 is present.
	basis = b2
	for basis != nil && basis != b1 {
		basis = basisParent(basis)
	}
	if basis != nil {
		return b2
	}
	return nil
}

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

// Tuple represents Python 'tuple' objects.
//
// Tuples are thread safe by virtue of being immutable.
type Tuple struct {
	Object
	elems []*Object
}

// NewTuple returns a tuple containing the given elements.
func NewTuple(elems ...*Object) *Tuple {
	if len(elems) == 0 {
		return emptyTuple
	}
	return &Tuple{Object: Object{typ: TupleType}, elems: elems}
}

func toTupleUnsafe(o *Object) *Tuple {
	return (*Tuple)(o.toPointer())
}

// GetItem returns the i'th element of t. Bounds are unchecked and therefore
// this method will panic unless 0 <= i < t.Len().
func (t *Tuple) GetItem(i int) *Object {
	return t.elems[i]
}

// Len returns the number of elements in t.
func (t *Tuple) Len() int {
	return len(t.elems)
}

// ToObject upcasts t to an Object.
func (t *Tuple) ToObject() *Object {
	return &t.Object
}

// TupleType is the object representing the Python 'tuple' type.
var TupleType = newBasisType("tuple", reflect.TypeOf(Tuple{}), toTupleUnsafe, ObjectType)

var emptyTuple = &Tuple{Object: Object{typ: TupleType}}

func tupleAdd(f *Frame, v, w *Object) (*Object, *BaseException) {
	if !w.isInstance(TupleType) {
		return NotImplemented, nil
	}
	elems, raised := seqAdd(f, toTupleUnsafe(v).elems, toTupleUnsafe(w).elems)
	if raised != nil {
		return nil, raised
	}
	return NewTuple(elems...).ToObject(), nil
}

func tupleContains(f *Frame, t, v *Object) (*Object, *BaseException) {
	return seqContains(f, t, v)
}

func tupleEq(f *Frame, v, w *Object) (*Object, *BaseException) {
	return tupleCompare(f, toTupleUnsafe(v), w, Eq)
}

func tupleGE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return tupleCompare(f, toTupleUnsafe(v), w, GE)
}

func tupleGetItem(f *Frame, o, key *Object) (*Object, *BaseException) {
	t := toTupleUnsafe(o)
	item, elems, raised := seqGetItem(f, t.elems, key)
	if raised != nil {
		return nil, raised
	}
	if item != nil {
		return item, nil
	}
	return NewTuple(elems...).ToObject(), nil
}

func tupleGetNewArgs(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "__getnewargs__", args, TupleType); raised != nil {
		return nil, raised
	}
	return NewTuple(args[0]).ToObject(), nil
}

func tupleGT(f *Frame, v, w *Object) (*Object, *BaseException) {
	return tupleCompare(f, toTupleUnsafe(v), w, GT)
}

func tupleIter(f *Frame, o *Object) (*Object, *BaseException) {
	return newSliceIterator(reflect.ValueOf(toTupleUnsafe(o).elems)), nil
}

func tupleLE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return tupleCompare(f, toTupleUnsafe(v), w, LE)
}

func tupleLen(f *Frame, o *Object) (*Object, *BaseException) {
	return NewInt(len(toTupleUnsafe(o).elems)).ToObject(), nil
}

func tupleLT(f *Frame, v, w *Object) (*Object, *BaseException) {
	return tupleCompare(f, toTupleUnsafe(v), w, LT)
}

func tupleMul(f *Frame, v, w *Object) (*Object, *BaseException) {
	if !w.isInstance(IntType) {
		return NotImplemented, nil
	}
	elems, raised := seqMul(f, toTupleUnsafe(v).elems, toIntUnsafe(w).Value())
	if raised != nil {
		return nil, raised
	}
	return NewTuple(elems...).ToObject(), nil
}

func tupleNE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return tupleCompare(f, toTupleUnsafe(v), w, NE)
}

func tupleNew(f *Frame, t *Type, args Args, _ KWArgs) (*Object, *BaseException) {
	elems, raised := seqNew(f, args)
	if raised != nil {
		return nil, raised
	}
	tup := toTupleUnsafe(newObject(t))
	tup.elems = elems
	return tup.ToObject(), nil
}

func tupleRepr(f *Frame, o *Object) (*Object, *BaseException) {
	t := toTupleUnsafe(o)
	if f.reprEnter(t.ToObject()) {
		return NewStr("(...)").ToObject(), nil
	}
	s, raised := seqRepr(f, t.elems)
	f.reprLeave(t.ToObject())
	if raised != nil {
		return nil, raised
	}
	if len(t.elems) == 1 {
		s = fmt.Sprintf("(%s,)", s)
	} else {
		s = fmt.Sprintf("(%s)", s)
	}
	return NewStr(s).ToObject(), nil
}

func tupleRMul(f *Frame, v, w *Object) (*Object, *BaseException) {
	if !w.isInstance(IntType) {
		return NotImplemented, nil
	}
	elems, raised := seqMul(f, toTupleUnsafe(v).elems, toIntUnsafe(w).Value())
	if raised != nil {
		return nil, raised
	}
	return NewTuple(elems...).ToObject(), nil
}

func initTupleType(dict map[string]*Object) {
	dict["__getnewargs__"] = newBuiltinFunction("__getnewargs__", tupleGetNewArgs).ToObject()
	TupleType.slots.Add = &binaryOpSlot{tupleAdd}
	TupleType.slots.Contains = &binaryOpSlot{tupleContains}
	TupleType.slots.Eq = &binaryOpSlot{tupleEq}
	TupleType.slots.GE = &binaryOpSlot{tupleGE}
	TupleType.slots.GetItem = &binaryOpSlot{tupleGetItem}
	TupleType.slots.GT = &binaryOpSlot{tupleGT}
	TupleType.slots.Iter = &unaryOpSlot{tupleIter}
	TupleType.slots.LE = &binaryOpSlot{tupleLE}
	TupleType.slots.Len = &unaryOpSlot{tupleLen}
	TupleType.slots.LT = &binaryOpSlot{tupleLT}
	TupleType.slots.Mul = &binaryOpSlot{tupleMul}
	TupleType.slots.NE = &binaryOpSlot{tupleNE}
	TupleType.slots.New = &newSlot{tupleNew}
	TupleType.slots.Repr = &unaryOpSlot{tupleRepr}
	TupleType.slots.RMul = &binaryOpSlot{tupleRMul}
}

func tupleCompare(f *Frame, v *Tuple, w *Object, cmp binaryOpFunc) (*Object, *BaseException) {
	if !w.isInstance(TupleType) {
		return NotImplemented, nil
	}
	return seqCompare(f, v.elems, toTupleUnsafe(w).elems, cmp)
}

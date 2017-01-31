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

// Below are direct allocation versions of small Tuples. Rather than performing
// two allocations, one for the tuple object and one for the slice holding the
// elements, we allocate both objects at the same time in one block of memory.
// This both decreases the number of allocations overall as well as increases
// memory locality for tuple data. Both of which *should* improve time to
// allocate as well as read performance. The methods below are used by the
// compiler to create fixed size tuples when the size is known ahead of time.
//
// The number of specializations below were chosen first to cover all the fixed
// size tuple allocations in the runtime (currently 5), then filled out to
// cover the whole memory size class (see golang/src/runtime/sizeclasses.go for
// the table). On a 64bit system, a tuple of length 6 occupies 96 bytes - 48
// bytes for the tuple object and 6*8 (48) bytes of pointers.
//
// If methods are added or removed, then the constant MAX_DIRECT_TUPLE in
// compiler/util.py needs to be updated as well.

// NewTuple0 returns the empty tuple. This is mostly provided for the
// convenience of the compiler.
func NewTuple0() *Tuple { return emptyTuple }

// NewTuple1 returns a tuple of length 1 containing just elem0.
func NewTuple1(elem0 *Object) *Tuple {
	t := struct {
		tuple Tuple
		elems [1]*Object
	}{
		tuple: Tuple{Object: Object{typ: TupleType}},
		elems: [1]*Object{elem0},
	}
	t.tuple.elems = t.elems[:]
	return &t.tuple
}

// NewTuple2 returns a tuple of length 2 containing just elem0 and elem1.
func NewTuple2(elem0, elem1 *Object) *Tuple {
	t := struct {
		tuple Tuple
		elems [2]*Object
	}{
		tuple: Tuple{Object: Object{typ: TupleType}},
		elems: [2]*Object{elem0, elem1},
	}
	t.tuple.elems = t.elems[:]
	return &t.tuple
}

// NewTuple3 returns a tuple of length 3 containing elem0 to elem2.
func NewTuple3(elem0, elem1, elem2 *Object) *Tuple {
	t := struct {
		tuple Tuple
		elems [3]*Object
	}{
		tuple: Tuple{Object: Object{typ: TupleType}},
		elems: [3]*Object{elem0, elem1, elem2},
	}
	t.tuple.elems = t.elems[:]
	return &t.tuple
}

// NewTuple4 returns a tuple of length 4 containing elem0 to elem3.
func NewTuple4(elem0, elem1, elem2, elem3 *Object) *Tuple {
	t := struct {
		tuple Tuple
		elems [4]*Object
	}{
		tuple: Tuple{Object: Object{typ: TupleType}},
		elems: [4]*Object{elem0, elem1, elem2, elem3},
	}
	t.tuple.elems = t.elems[:]
	return &t.tuple
}

// NewTuple5 returns a tuple of length 5 containing elem0 to elem4.
func NewTuple5(elem0, elem1, elem2, elem3, elem4 *Object) *Tuple {
	t := struct {
		tuple Tuple
		elems [5]*Object
	}{
		tuple: Tuple{Object: Object{typ: TupleType}},
		elems: [5]*Object{elem0, elem1, elem2, elem3, elem4},
	}
	t.tuple.elems = t.elems[:]
	return &t.tuple
}

// NewTuple6 returns a tuple of length 6 containing elem0 to elem5.
func NewTuple6(elem0, elem1, elem2, elem3, elem4, elem5 *Object) *Tuple {
	t := struct {
		tuple Tuple
		elems [6]*Object
	}{
		tuple: Tuple{Object: Object{typ: TupleType}},
		elems: [6]*Object{elem0, elem1, elem2, elem3, elem4, elem5},
	}
	t.tuple.elems = t.elems[:]
	return &t.tuple
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

func tupleCount(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "count", args, TupleType, ObjectType); raised != nil {
		return nil, raised
	}
	return seqCount(f, args[0], args[1])
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
	return NewTuple1(args[0]).ToObject(), nil
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
	if t == TupleType && len(args) == 1 && args[0].typ == TupleType {
		// Tuples are immutable so just return the tuple provided.
		return args[0], nil
	}
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
	dict["count"] = newBuiltinFunction("count", tupleCount).ToObject()
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

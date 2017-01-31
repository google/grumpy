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

import "reflect"

const errBadSliceIndex = "slice indices must be integers or None or have an __index__ method"

var (
	// SliceType is the object representing the Python 'slice' type.
	SliceType = newBasisType("slice", reflect.TypeOf(Slice{}), toSliceUnsafe, ObjectType)
)

// Slice represents Python 'slice' objects.
type Slice struct {
	Object
	start *Object `attr:"start"`
	stop  *Object `attr:"stop"`
	step  *Object `attr:"step"`
}

func toSliceUnsafe(o *Object) *Slice {
	return (*Slice)(o.toPointer())
}

// calcSlice returns the three range indices (start, stop, step) and the length
// of the slice produced by slicing a sequence of length numElems by s. As with
// seqRange, the resulting indices can be used to iterate over the slice like:
//
// for i := start; i != stop; i += step { ... }
func (s *Slice) calcSlice(f *Frame, numElems int) (int, int, int, int, *BaseException) {
	step := 1
	if s.step != nil && s.step != None {
		if s.step.typ.slots.Index == nil {
			return 0, 0, 0, 0, f.RaiseType(TypeErrorType, errBadSliceIndex)
		}
		i, raised := IndexInt(f, s.step)
		if raised != nil {
			return 0, 0, 0, 0, raised
		}
		step = i
	}
	var startDef, stopDef int
	if step > 0 {
		startDef, stopDef = 0, numElems
	} else {
		startDef, stopDef = numElems-1, -1
	}
	start, raised := sliceClampIndex(f, s.start, startDef, numElems)
	if raised != nil {
		return 0, 0, 0, 0, raised
	}
	stop, raised := sliceClampIndex(f, s.stop, stopDef, numElems)
	if raised != nil {
		return 0, 0, 0, 0, raised
	}
	stop, sliceLen, result := seqRange(start, stop, step)
	switch result {
	case seqRangeZeroStep:
		return 0, 0, 0, 0, f.RaiseType(ValueErrorType, "slice step cannot be zero")
	case seqRangeOverflow:
		return 0, 0, 0, 0, f.RaiseType(OverflowErrorType, errResultTooLarge)
	}
	return start, stop, step, sliceLen, nil
}

// ToObject upcasts s to an Object.
func (s *Slice) ToObject() *Object {
	return &s.Object
}

func sliceEq(f *Frame, v, w *Object) (*Object, *BaseException) {
	return sliceCompare(f, toSliceUnsafe(v), w, Eq)
}

func sliceGE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return sliceCompare(f, toSliceUnsafe(v), w, GE)
}

func sliceGT(f *Frame, v, w *Object) (*Object, *BaseException) {
	return sliceCompare(f, toSliceUnsafe(v), w, GT)
}

func sliceLE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return sliceCompare(f, toSliceUnsafe(v), w, LE)
}

func sliceLT(f *Frame, v, w *Object) (*Object, *BaseException) {
	return sliceCompare(f, toSliceUnsafe(v), w, LT)
}

func sliceNE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return sliceCompare(f, toSliceUnsafe(v), w, NE)
}

func sliceNew(f *Frame, t *Type, args Args, _ KWArgs) (*Object, *BaseException) {
	expectedTypes := []*Type{ObjectType, ObjectType, ObjectType}
	argc := len(args)
	if argc >= 1 && argc <= 3 {
		expectedTypes = expectedTypes[:argc]
	}
	if raised := checkMethodArgs(f, "__new__", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	s := toSliceUnsafe(newObject(t))
	if argc == 1 {
		s.stop = args[0]
	} else {
		s.start = args[0]
		s.stop = args[1]
		if argc > 2 {
			s.step = args[2]
		}
	}
	return s.ToObject(), nil
}

func sliceRepr(f *Frame, o *Object) (*Object, *BaseException) {
	s := toSliceUnsafe(o)
	elem0, elem1, elem2 := None, s.stop, None
	if s.start != nil {
		elem0 = s.start
	}
	if s.step != nil {
		elem2 = s.step
	}
	r, raised := Repr(f, NewTuple3(elem0, elem1, elem2).ToObject())
	if raised != nil {
		return nil, raised
	}
	return NewStr("slice" + r.Value()).ToObject(), nil
}

func initSliceType(map[string]*Object) {
	SliceType.flags &^= typeFlagBasetype
	SliceType.slots.Eq = &binaryOpSlot{sliceEq}
	SliceType.slots.GE = &binaryOpSlot{sliceGE}
	SliceType.slots.GT = &binaryOpSlot{sliceGT}
	SliceType.slots.Hash = &unaryOpSlot{hashNotImplemented}
	SliceType.slots.LE = &binaryOpSlot{sliceLE}
	SliceType.slots.LT = &binaryOpSlot{sliceLT}
	SliceType.slots.NE = &binaryOpSlot{sliceNE}
	SliceType.slots.New = &newSlot{sliceNew}
	SliceType.slots.Repr = &unaryOpSlot{sliceRepr}
}

func sliceClampIndex(f *Frame, index *Object, def, seqLen int) (int, *BaseException) {
	if index == nil || index == None {
		return def, nil
	}
	if index.typ.slots.Index == nil {
		return 0, f.RaiseType(TypeErrorType, errBadSliceIndex)
	}
	i, raised := IndexInt(f, index)
	if raised != nil {
		return 0, raised
	}
	return seqClampIndex(i, seqLen), nil
}

func sliceCompare(f *Frame, v *Slice, w *Object, cmp binaryOpFunc) (*Object, *BaseException) {
	if !w.isInstance(SliceType) {
		return NotImplemented, nil
	}
	rhs := toSliceUnsafe(w)
	elems1, elems2 := []*Object{None, v.stop, None}, []*Object{None, rhs.stop, None}
	if v.start != nil {
		elems1[0] = v.start
	}
	if v.step != nil {
		elems1[2] = v.step
	}
	if rhs.start != nil {
		elems2[0] = rhs.start
	}
	if rhs.step != nil {
		elems2[2] = rhs.step
	}
	return seqCompare(f, elems1, elems2, cmp)
}

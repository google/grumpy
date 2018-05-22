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

// This file contains common code and helpers for sequence types.

import (
	"bytes"
	"fmt"
	"reflect"
	"sync"
)

var (
	seqIteratorType = newBasisType("iterator", reflect.TypeOf(seqIterator{}), toSeqIteratorUnsafe, ObjectType)
)

func seqAdd(f *Frame, elems1, elems2 []*Object) ([]*Object, *BaseException) {
	if len(elems1)+len(elems2) < 0 {
		// This indicates an int overflow.
		return nil, f.RaiseType(OverflowErrorType, errResultTooLarge)
	}
	return append(elems1, elems2...), nil
}

func seqCompare(f *Frame, elems1, elems2 []*Object, cmp binaryOpFunc) (*Object, *BaseException) {
	n1 := len(elems1)
	n2 := len(elems2)
	for i := 0; i < n1 && i < n2; i++ {
		eq, raised := Eq(f, elems1[i], elems2[i])
		if raised != nil {
			return nil, raised
		}
		if ret, raised := IsTrue(f, eq); raised != nil {
			return nil, raised
		} else if !ret {
			// We encountered an unequal element before the end of
			// either sequence so perform the comparison on the two
			// elements.
			return cmp(f, elems1[i], elems2[i])
		}
	}
	// One sequence is longer than the other, so do the comparison on the
	// lengths of the two sequences.
	return cmp(f, NewInt(n1).ToObject(), NewInt(n2).ToObject())
}

// seqApply calls fun with a slice of objects contained in the sequence object
// seq. If the second callback parameter is true, the slice is borrowed and the
// function must not modify the provided slice. Otherwise the slice is scratch
// and it may freely be used in any way. It will raise if seq is not a sequence
// object.
func seqApply(f *Frame, seq *Object, fun func([]*Object, bool) *BaseException) *BaseException {
	switch {
	// Don't use fast path referencing the underlying slice directly for
	// list and tuple subtypes. See comment in listextend in listobject.c.
	case seq.typ == ListType:
		l := toListUnsafe(seq)
		l.mutex.RLock()
		raised := fun(l.elems, true)
		l.mutex.RUnlock()
		return raised
	case seq.typ == TupleType:
		return fun(toTupleUnsafe(seq).elems, true)
	default:
		elems := []*Object{}
		raised := seqForEach(f, seq, func(elem *Object) *BaseException {
			elems = append(elems, elem)
			return nil
		})
		if raised != nil {
			return raised
		}
		return fun(elems, false)
	}
}

func seqCheckedIndex(f *Frame, seqLen, index int) (int, *BaseException) {
	if index < 0 {
		index = seqLen + index
	}
	if index < 0 || index >= seqLen {
		return 0, f.RaiseType(IndexErrorType, "index out of range")
	}
	return index, nil
}

func seqClampIndex(i, seqLen int) int {
	if i < 0 {
		i += seqLen
		if i < 0 {
			i = 0
		}
	}
	if i > seqLen {
		i = seqLen
	}
	return i
}

func seqContains(f *Frame, iterable *Object, v *Object) (*Object, *BaseException) {
	pred := func(o *Object) (bool, *BaseException) {
		eq, raised := Eq(f, v, o)
		if raised != nil {
			return false, raised
		}
		ret, raised := IsTrue(f, eq)
		if raised != nil {
			return false, raised
		}
		return ret, nil
	}
	foundEqItem, raised := seqFindFirst(f, iterable, pred)
	if raised != nil {
		return nil, raised
	}
	return GetBool(foundEqItem).ToObject(), raised
}

func seqCount(f *Frame, iterable *Object, v *Object) (*Object, *BaseException) {
	count := 0
	raised := seqForEach(f, iterable, func(o *Object) *BaseException {
		eq, raised := Eq(f, o, v)
		if raised != nil {
			return raised
		}
		t, raised := IsTrue(f, eq)
		if raised != nil {
			return raised
		}
		if t {
			count++
		}
		return nil
	})
	if raised != nil {
		return nil, raised
	}
	return NewInt(count).ToObject(), nil
}

func seqFindFirst(f *Frame, iterable *Object, pred func(*Object) (bool, *BaseException)) (bool, *BaseException) {
	iter, raised := Iter(f, iterable)
	if raised != nil {
		return false, raised
	}
	item, raised := Next(f, iter)
	for ; raised == nil; item, raised = Next(f, iter) {
		ret, raised := pred(item)
		if raised != nil {
			return false, raised
		}
		if ret {
			return true, nil
		}
	}
	if !raised.isInstance(StopIterationType) {
		return false, raised
	}
	f.RestoreExc(nil, nil)
	return false, nil
}

func seqFindElem(f *Frame, elems []*Object, o *Object) (int, *BaseException) {
	for i, elem := range elems {
		eq, raised := Eq(f, elem, o)
		if raised != nil {
			return -1, raised
		}
		found, raised := IsTrue(f, eq)
		if raised != nil {
			return -1, raised
		}
		if found {
			return i, nil
		}
	}
	return -1, nil
}

func seqForEach(f *Frame, iterable *Object, callback func(*Object) *BaseException) *BaseException {
	iter, raised := Iter(f, iterable)
	if raised != nil {
		return raised
	}
	item, raised := Next(f, iter)
	for ; raised == nil; item, raised = Next(f, iter) {
		if raised := callback(item); raised != nil {
			return raised
		}
	}
	if !raised.isInstance(StopIterationType) {
		return raised
	}
	f.RestoreExc(nil, nil)
	return nil
}

// seqGetItem returns a single element or a slice of elements of elems
// depending on whether index is an integer or a slice. If index is neither of
// those types then a TypeError is returned.
func seqGetItem(f *Frame, elems []*Object, index *Object) (*Object, []*Object, *BaseException) {
	switch {
	case index.typ.slots.Index != nil:
		i, raised := IndexInt(f, index)
		if raised != nil {
			return nil, nil, raised
		}
		i, raised = seqCheckedIndex(f, len(elems), i)
		if raised != nil {
			return nil, nil, raised
		}
		return elems[i], nil, nil
	case index.isInstance(SliceType):
		s := toSliceUnsafe(index)
		start, stop, step, sliceLen, raised := s.calcSlice(f, len(elems))
		if raised != nil {
			return nil, nil, raised
		}
		result := make([]*Object, sliceLen)
		i := 0
		for j := start; j != stop; j += step {
			result[i] = elems[j]
			i++
		}
		return nil, result, nil
	}
	return nil, nil, f.RaiseType(TypeErrorType, fmt.Sprintf("sequence indices must be integers, not %s", index.typ.Name()))
}

func seqMul(f *Frame, elems []*Object, n int) ([]*Object, *BaseException) {
	if n <= 0 {
		return nil, nil
	}
	numElems := len(elems)
	if numElems > MaxInt/n {
		return nil, f.RaiseType(OverflowErrorType, errResultTooLarge)
	}
	newNumElems := numElems * n
	resultElems := make([]*Object, newNumElems)
	for i := 0; i < newNumElems; i++ {
		resultElems[i] = elems[i%numElems]
	}
	return resultElems, nil
}

func seqNew(f *Frame, args Args) ([]*Object, *BaseException) {
	if len(args) == 0 {
		return nil, nil
	}
	if raised := checkMethodArgs(f, "__new__", args, ObjectType); raised != nil {
		return nil, raised
	}
	var result []*Object
	raised := seqApply(f, args[0], func(elems []*Object, borrowed bool) *BaseException {
		if borrowed {
			result = make([]*Object, len(elems))
			copy(result, elems)
		} else {
			result = elems
		}
		return nil
	})
	if raised != nil {
		return nil, raised
	}
	return result, nil
}

type seqRangeResult int

const (
	seqRangeOK seqRangeResult = iota
	seqRangeOverflow
	seqRangeZeroStep
)

// seqRange takes the bounds and stride defining a Python range (e.g.
// xrange(start, stop, step)) and returns three things:
//
// 1. The terminal value for the range when iterating
// 2. The length of the range (i.e. the number of iterations)
// 3. A status indicating whether the range is valid
//
// The terminal value can be used to iterate over the range as follows:
//
// for i := start; i != term; i += step { ... }
func seqRange(start, stop, step int) (int, int, seqRangeResult) {
	if step == 0 {
		return 0, 0, seqRangeZeroStep
	}
	if stop == start || (stop > start) != (step > 0) {
		// The step doesn't make progress towards the goal,
		// so return an empty range.
		return start, 0, seqRangeOK
	}
	if step > 0 {
		stop--
	} else {
		stop++
	}
	n := (stop-start)/step + 1
	if n < 0 {
		return 0, 0, seqRangeOverflow
	}
	return start + n*step, n, seqRangeOK
}

func seqRepr(f *Frame, elems []*Object) (string, *BaseException) {
	var buf bytes.Buffer
	for i, o := range elems {
		if i > 0 {
			buf.WriteString(", ")
		}
		s, raised := Repr(f, o)
		if raised != nil {
			return "", raised
		}
		buf.WriteString(s.Value())
		i++
	}
	return buf.String(), nil
}

func seqWrapEach(f *Frame, elems ...interface{}) ([]*Object, *BaseException) {
	result := make([]*Object, len(elems))
	for i, elem := range elems {
		var raised *BaseException
		if result[i], raised = WrapNative(f, reflect.ValueOf(elem)); raised != nil {
			return nil, raised
		}
	}
	return result, nil
}

type seqIterator struct {
	Object
	seq   *Object
	mutex sync.Mutex
	index int
}

func newSeqIterator(seq *Object) *Object {
	iter := &seqIterator{Object: Object{typ: seqIteratorType}, seq: seq}
	return &iter.Object
}

func toSeqIteratorUnsafe(o *Object) *seqIterator {
	return (*seqIterator)(o.toPointer())
}

func seqIteratorIter(f *Frame, o *Object) (*Object, *BaseException) {
	return o, nil
}

func seqIteratorNext(f *Frame, o *Object) (item *Object, raised *BaseException) {
	i := toSeqIteratorUnsafe(o)
	i.mutex.Lock()
	if i.seq == nil {
		raised = f.Raise(StopIterationType.ToObject(), nil, nil)
	} else if item, raised = GetItem(f, i.seq, NewInt(i.index).ToObject()); raised == nil {
		i.index++
	} else if raised.isInstance(IndexErrorType) {
		i.seq = nil
		raised = f.Raise(StopIterationType.ToObject(), nil, nil)
	}
	i.mutex.Unlock()
	return item, raised
}

func initSeqIteratorType(map[string]*Object) {
	seqIteratorType.flags &= ^(typeFlagBasetype | typeFlagInstantiable)
	seqIteratorType.slots.Iter = &unaryOpSlot{seqIteratorIter}
	seqIteratorType.slots.Next = &unaryOpSlot{seqIteratorNext}
}

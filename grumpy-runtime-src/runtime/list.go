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
	"sort"
	"sync"
)

// List represents Python 'list' objects.
//
// Lists are thread safe, however read operations are not necessarily atomic.
// E.g.  given the list l = [1, 2, 3] executing del l[1] in one thread may give
// repr(l) == [1, 2] in another which is never correct.
type List struct {
	Object
	mutex sync.RWMutex
	elems []*Object
}

// NewList returns a list containing the given elements.
func NewList(elems ...*Object) *List {
	l := &List{Object: Object{typ: ListType}}
	numElems := len(elems)
	l.resize(numElems)
	for i := 0; i < numElems; i++ {
		l.elems[i] = elems[i]
	}
	return l
}

func toListUnsafe(o *Object) *List {
	return (*List)(o.toPointer())
}

// ToObject upcasts l to an Object.
func (l *List) ToObject() *Object {
	return &l.Object
}

// Append adds o to the end of l.
func (l *List) Append(o *Object) {
	l.mutex.Lock()
	newLen := len(l.elems) + 1
	l.resize(newLen)
	l.elems[newLen-1] = o
	l.mutex.Unlock()
}

// DelItem removes the index'th element of l.
func (l *List) DelItem(f *Frame, index int) *BaseException {
	l.mutex.Lock()
	numElems := len(l.elems)
	i, raised := seqCheckedIndex(f, numElems, index)
	if raised == nil {
		copy(l.elems[i:numElems-1], l.elems[i+1:numElems])
		l.elems = l.elems[:numElems-1]
	}
	l.mutex.Unlock()
	return raised
}

// DelSlice removes the slice of l specified by s.
func (l *List) DelSlice(f *Frame, s *Slice) *BaseException {
	l.mutex.Lock()
	numListElems := len(l.elems)
	start, stop, step, numSliceElems, raised := s.calcSlice(f, numListElems)
	if raised == nil {
		if step == 1 {
			copy(l.elems[start:numListElems-numSliceElems], l.elems[stop:numListElems])
		} else {
			j := 0
			for i := start; i != stop; i += step {
				next := i + step
				if next > numListElems {
					next = numListElems
				}
				dest := l.elems[i-j : next-j-1]
				src := l.elems[i+1 : next]
				copy(dest, src)
				j++
			}
		}
		l.elems = l.elems[:numListElems-numSliceElems]
	}
	l.mutex.Unlock()
	return raised
}

// SetItem sets the index'th element of l to value.
func (l *List) SetItem(f *Frame, index int, value *Object) *BaseException {
	l.mutex.Lock()
	i, raised := seqCheckedIndex(f, len(l.elems), index)
	if raised == nil {
		l.elems[i] = value
	}
	l.mutex.Unlock()
	return raised
}

// SetSlice replaces the slice of l specified by s with the contents of value
// (an iterable).
func (l *List) SetSlice(f *Frame, s *Slice, value *Object) *BaseException {
	l.mutex.Lock()
	numListElems := len(l.elems)
	start, stop, step, numSliceElems, raised := s.calcSlice(f, numListElems)
	if raised == nil {
		raised = seqApply(f, value, func(elems []*Object, _ bool) *BaseException {
			numElems := len(elems)
			if step == 1 {
				tailElems := l.elems[stop:numListElems]
				l.resize(numListElems - numSliceElems + numElems)
				copy(l.elems[start+numElems:], tailElems)
				copy(l.elems[start:start+numElems], elems)
			} else if numSliceElems == numElems {
				i := 0
				for j := start; j != stop; j += step {
					l.elems[j] = elems[i]
					i++
				}
			} else {
				format := "attempt to assign sequence of size %d to extended slice of size %d"
				return f.RaiseType(ValueErrorType, fmt.Sprintf(format, numElems, numSliceElems))
			}
			return nil
		})
	}
	l.mutex.Unlock()
	return raised
}

// Sort reorders l so that its elements are in sorted order.
func (l *List) Sort(f *Frame) (raised *BaseException) {
	l.mutex.RLock()
	sorter := &listSorter{f, l, nil}
	defer func() {
		l.mutex.RUnlock()
		if val := recover(); val == nil {
			return
		} else if s, ok := val.(*listSorter); !ok || s != sorter {
			panic(val)
		}
		raised = sorter.raised
	}()
	// Python guarantees stability.  See note (9) in:
	// https://docs.python.org/2/library/stdtypes.html#mutable-sequence-types
	sort.Stable(sorter)
	return nil
}

// resize ensures that len(l.elems) == newLen, reallocating if necessary.
// NOTE: l.mutex must be locked when calling resize.
func (l *List) resize(newLen int) {
	if cap(l.elems) < newLen {
		// Borrowed from CPython's list_resize() in listobject.c.
		newCap := (newLen >> 3) + 3 + newLen
		if newLen >= 9 {
			newCap += 3
		}
		newElems := make([]*Object, len(l.elems), newCap)
		copy(newElems, l.elems)
		l.elems = newElems
	}
	l.elems = l.elems[:newLen]
}

// ListType is the object representing the Python 'list' type.
var ListType = newBasisType("list", reflect.TypeOf(List{}), toListUnsafe, ObjectType)

func listAdd(f *Frame, v, w *Object) (ret *Object, raised *BaseException) {
	if !w.isInstance(ListType) {
		return NotImplemented, nil
	}
	listV, listW := toListUnsafe(v), toListUnsafe(w)
	listV.mutex.RLock()
	listW.mutex.RLock()
	elems, raised := seqAdd(f, listV.elems, listW.elems)
	if raised == nil {
		ret = NewList(elems...).ToObject()
	}
	listW.mutex.RUnlock()
	listV.mutex.RUnlock()
	return ret, raised
}

func listAppend(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "append", args, ListType, ObjectType); raised != nil {
		return nil, raised
	}
	toListUnsafe(args[0]).Append(args[1])
	return None, nil
}

func listCount(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "count", args, ListType, ObjectType); raised != nil {
		return nil, raised
	}
	return seqCount(f, args[0], args[1])
}

func listDelItem(f *Frame, o *Object, key *Object) *BaseException {
	l := toListUnsafe(o)
	if key.isInstance(SliceType) {
		return l.DelSlice(f, toSliceUnsafe(key))
	}
	if key.typ.slots.Index == nil {
		format := "list indices must be integers, not %s"
		return f.RaiseType(TypeErrorType, fmt.Sprintf(format, key.Type().Name()))
	}
	index, raised := IndexInt(f, key)
	if raised != nil {
		return raised
	}
	return l.DelItem(f, index)
}

func listRemove(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "remove", args, ListType, ObjectType); raised != nil {
		return nil, raised
	}
	value := args[1]
	l := toListUnsafe(args[0])
	l.mutex.Lock()
	index, raised := seqFindElem(f, l.elems, value)
	if raised == nil {
		if index != -1 {
			l.elems = append(l.elems[:index], l.elems[index+1:]...)
		} else {
			raised = f.RaiseType(ValueErrorType, "list.remove(x): x not in list")
		}
	}
	l.mutex.Unlock()
	if raised != nil {
		return nil, raised
	}
	return None, nil
}

func listExtend(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	argc := len(args)
	if argc != 2 {
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf("extend() takes exactly one argument (%d given)", argc))
	}
	return listIAdd(f, args[0], args[1])
}

func listContains(f *Frame, l, v *Object) (*Object, *BaseException) {
	return seqContains(f, l, v)
}

func listEq(f *Frame, v, w *Object) (*Object, *BaseException) {
	return listCompare(f, toListUnsafe(v), w, Eq)
}

func listGE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return listCompare(f, toListUnsafe(v), w, GE)
}

func listGetItem(f *Frame, o, key *Object) (*Object, *BaseException) {
	l := toListUnsafe(o)
	if key.typ.slots.Index == nil && !key.isInstance(SliceType) {
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf("list indices must be integers, not %s", key.typ.Name()))
	}
	l.mutex.RLock()
	item, elems, raised := seqGetItem(f, l.elems, key)
	l.mutex.RUnlock()
	if raised != nil {
		return nil, raised
	}
	if item != nil {
		return item, nil
	}
	return NewList(elems...).ToObject(), nil
}

func listGT(f *Frame, v, w *Object) (*Object, *BaseException) {
	return listCompare(f, toListUnsafe(v), w, GT)
}

func listIAdd(f *Frame, v, w *Object) (*Object, *BaseException) {
	l := toListUnsafe(v)
	raised := seqForEach(f, w, func(o *Object) *BaseException {
		l.Append(o)
		return nil
	})
	if raised != nil {
		return nil, raised
	}
	return v, nil
}

func listIMul(f *Frame, v, w *Object) (*Object, *BaseException) {
	if !w.isInstance(IntType) {
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf("can't multiply sequence by non-int of type '%s'", w.typ.Name()))
	}
	l, n := toListUnsafe(v), toIntUnsafe(w).Value()
	l.mutex.Lock()
	elems, raised := seqMul(f, l.elems, n)
	if raised == nil {
		l.elems = elems
	}
	l.mutex.Unlock()
	if raised != nil {
		return nil, raised
	}
	return v, nil
}

func listInsert(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "insert", args, ListType, IntType, ObjectType); raised != nil {
		return nil, raised
	}
	l := toListUnsafe(args[0])
	l.mutex.Lock()
	elems := l.elems
	numElems := len(elems)
	i := seqClampIndex(toIntUnsafe(args[1]).Value(), numElems)
	l.resize(numElems + 1)
	// TODO: The resize() above may have done a copy so we're doing a lot
	// of extra work here. Optimize this.
	copy(l.elems[i+1:], elems[i:])
	l.elems[i] = args[2]
	l.mutex.Unlock()
	return None, nil
}

func listIter(f *Frame, o *Object) (*Object, *BaseException) {
	return newListIterator(toListUnsafe(o)), nil
}

func listLE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return listCompare(f, toListUnsafe(v), w, LE)
}

func listLen(f *Frame, o *Object) (*Object, *BaseException) {
	l := toListUnsafe(o)
	l.mutex.RLock()
	ret := NewInt(len(l.elems)).ToObject()
	l.mutex.RUnlock()
	return ret, nil
}

func listNew(f *Frame, t *Type, args Args, _ KWArgs) (*Object, *BaseException) {
	elems, raised := seqNew(f, args)
	if raised != nil {
		return nil, raised
	}
	l := toListUnsafe(newObject(t))
	l.elems = elems
	return l.ToObject(), nil
}

func listLT(f *Frame, v, w *Object) (*Object, *BaseException) {
	return listCompare(f, toListUnsafe(v), w, LT)
}

func listMul(f *Frame, v, w *Object) (*Object, *BaseException) {
	if !w.isInstance(IntType) {
		return NotImplemented, nil
	}
	l, n := toListUnsafe(v), toIntUnsafe(w).Value()
	l.mutex.RLock()
	elems, raised := seqMul(f, l.elems, n)
	l.mutex.RUnlock()
	if raised != nil {
		return nil, raised
	}
	return NewList(elems...).ToObject(), nil
}

func listNE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return listCompare(f, toListUnsafe(v), w, NE)
}

func listIndex(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	expectedTypes := []*Type{ListType, ObjectType, ObjectType, ObjectType}
	argc := len(args)
	var raised *BaseException
	if argc == 2 || argc == 3 {
		expectedTypes = expectedTypes[:argc]
	}
	if raised = checkMethodArgs(f, "index", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	l := toListUnsafe(args[0])
	l.mutex.RLock()
	numElems := len(l.elems)
	start, stop := 0, numElems
	if argc > 2 {
		start, raised = IndexInt(f, args[2])
		if raised != nil {
			l.mutex.RUnlock()
			return nil, raised
		}
	}
	if argc > 3 {
		stop, raised = IndexInt(f, args[3])
		if raised != nil {
			l.mutex.RUnlock()
			return nil, raised
		}
	}
	start, stop = adjustIndex(start, stop, numElems)
	value := args[1]
	index := -1
	if start < numElems && start < stop {
		index, raised = seqFindElem(f, l.elems[start:stop], value)
	}
	l.mutex.RUnlock()
	if raised != nil {
		return nil, raised
	}
	if index == -1 {
		return nil, f.RaiseType(ValueErrorType, fmt.Sprintf("%v is not in list", value))
	}
	return NewInt(index + start).ToObject(), nil
}

func listPop(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	argc := len(args)
	expectedTypes := []*Type{ListType, ObjectType}
	if argc == 1 {
		expectedTypes = expectedTypes[:1]
	}
	if raised := checkMethodArgs(f, "pop", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	i := -1
	if argc == 2 {
		var raised *BaseException
		i, raised = ToIntValue(f, args[1])
		if raised != nil {
			return nil, raised
		}
	}
	l := toListUnsafe(args[0])
	l.mutex.Lock()
	numElems := len(l.elems)
	if i < 0 {
		i += numElems
	}
	var item *Object
	var raised *BaseException
	if i >= numElems || i < 0 {
		raised = f.RaiseType(IndexErrorType, "list index out of range")
	} else {
		item = l.elems[i]
		l.elems = append(l.elems[:i], l.elems[i+1:]...)
	}
	l.mutex.Unlock()
	return item, raised
}

func listRepr(f *Frame, o *Object) (*Object, *BaseException) {
	l := toListUnsafe(o)
	if f.reprEnter(l.ToObject()) {
		return NewStr("[...]").ToObject(), nil
	}
	l.mutex.RLock()
	repr, raised := seqRepr(f, l.elems)
	l.mutex.RUnlock()
	f.reprLeave(l.ToObject())
	if raised != nil {
		return nil, raised
	}
	return NewStr(fmt.Sprintf("[%s]", repr)).ToObject(), nil
}

func listReverse(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "reverse", args, ListType); raised != nil {
		return nil, raised
	}
	l := toListUnsafe(args[0])
	l.mutex.Lock()
	halfLen := len(l.elems) / 2
	for i := 0; i < halfLen; i++ {
		j := len(l.elems) - i - 1
		l.elems[i], l.elems[j] = l.elems[j], l.elems[i]
	}
	l.mutex.Unlock()
	return None, nil
}

func listSetItem(f *Frame, o, key, value *Object) *BaseException {
	l := toListUnsafe(o)
	if key.typ.slots.Index != nil {
		i, raised := IndexInt(f, key)
		if raised != nil {
			return raised
		}
		return l.SetItem(f, i, value)
	}
	if key.isInstance(SliceType) {
		return l.SetSlice(f, toSliceUnsafe(key), value)
	}
	return f.RaiseType(TypeErrorType, fmt.Sprintf("list indices must be integers, not %s", key.Type().Name()))
}

func listSort(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	// TODO: Support (cmp=None, key=None, reverse=False)
	if raised := checkMethodArgs(f, "sort", args, ListType); raised != nil {
		return nil, raised
	}
	l := toListUnsafe(args[0])
	l.Sort(f)
	return None, nil
}

func initListType(dict map[string]*Object) {
	dict["append"] = newBuiltinFunction("append", listAppend).ToObject()
	dict["count"] = newBuiltinFunction("count", listCount).ToObject()
	dict["extend"] = newBuiltinFunction("extend", listExtend).ToObject()
	dict["index"] = newBuiltinFunction("index", listIndex).ToObject()
	dict["insert"] = newBuiltinFunction("insert", listInsert).ToObject()
	dict["pop"] = newBuiltinFunction("pop", listPop).ToObject()
	dict["remove"] = newBuiltinFunction("remove", listRemove).ToObject()
	dict["reverse"] = newBuiltinFunction("reverse", listReverse).ToObject()
	dict["sort"] = newBuiltinFunction("sort", listSort).ToObject()
	ListType.slots.Add = &binaryOpSlot{listAdd}
	ListType.slots.Contains = &binaryOpSlot{listContains}
	ListType.slots.DelItem = &delItemSlot{listDelItem}
	ListType.slots.Eq = &binaryOpSlot{listEq}
	ListType.slots.GE = &binaryOpSlot{listGE}
	ListType.slots.GetItem = &binaryOpSlot{listGetItem}
	ListType.slots.GT = &binaryOpSlot{listGT}
	ListType.slots.Hash = &unaryOpSlot{hashNotImplemented}
	ListType.slots.IAdd = &binaryOpSlot{listIAdd}
	ListType.slots.IMul = &binaryOpSlot{listIMul}
	ListType.slots.Iter = &unaryOpSlot{listIter}
	ListType.slots.LE = &binaryOpSlot{listLE}
	ListType.slots.Len = &unaryOpSlot{listLen}
	ListType.slots.LT = &binaryOpSlot{listLT}
	ListType.slots.Mul = &binaryOpSlot{listMul}
	ListType.slots.NE = &binaryOpSlot{listNE}
	ListType.slots.New = &newSlot{listNew}
	ListType.slots.Repr = &unaryOpSlot{listRepr}
	ListType.slots.RMul = &binaryOpSlot{listMul}
	ListType.slots.SetItem = &setItemSlot{listSetItem}
}

type listIterator struct {
	Object
	list  *List
	mutex sync.Mutex
	index int
}

func newListIterator(l *List) *Object {
	iter := &listIterator{Object: Object{typ: listIteratorType}, list: l}
	return &iter.Object
}

func toListIteratorUnsafe(o *Object) *listIterator {
	return (*listIterator)(o.toPointer())
}

var listIteratorType = newBasisType("listiterator", reflect.TypeOf(listIterator{}), toListIteratorUnsafe, ObjectType)

func listIteratorIter(f *Frame, o *Object) (*Object, *BaseException) {
	return o, nil
}

func listIteratorNext(f *Frame, o *Object) (ret *Object, raised *BaseException) {
	i := toListIteratorUnsafe(o)
	// Ensure that no mutations happen to the list.
	i.list.mutex.RLock()
	i.mutex.Lock()
	if i.index < len(i.list.elems) {
		ret = i.list.elems[i.index]
		i.index++
	} else {
		// Ensure that we raise StopIteration henceforth even if the
		// sequence grows subsequently.
		i.index = MaxInt
		raised = f.Raise(StopIterationType.ToObject(), nil, nil)
	}
	i.mutex.Unlock()
	i.list.mutex.RUnlock()
	return ret, raised
}

func initListIteratorType(map[string]*Object) {
	listIteratorType.flags &= ^(typeFlagBasetype | typeFlagInstantiable)
	listIteratorType.slots.Iter = &unaryOpSlot{listIteratorIter}
	listIteratorType.slots.Next = &unaryOpSlot{listIteratorNext}
}

func listCompare(f *Frame, v *List, w *Object, cmp binaryOpFunc) (*Object, *BaseException) {
	if !w.isInstance(ListType) {
		return NotImplemented, nil
	}
	listw := toListUnsafe(w)
	// Order of locking doesn't matter since we're doing a read lock.
	v.mutex.RLock()
	listw.mutex.RLock()
	ret, raised := seqCompare(f, v.elems, listw.elems, cmp)
	listw.mutex.RUnlock()
	v.mutex.RUnlock()
	return ret, raised
}

type listSorter struct {
	f      *Frame
	l      *List
	raised *BaseException
}

func (s *listSorter) Len() int {
	return len(s.l.elems)
}

func (s *listSorter) Less(i, j int) bool {
	lt, raised := LT(s.f, s.l.elems[i], s.l.elems[j])
	if raised != nil {
		s.raised = raised
		panic(s)
	}
	ret, raised := IsTrue(s.f, lt)
	if raised != nil {
		s.raised = raised
		panic(s)
	}
	return ret
}

func (s *listSorter) Swap(i, j int) {
	s.l.elems[i], s.l.elems[j] = s.l.elems[j], s.l.elems[i]
}

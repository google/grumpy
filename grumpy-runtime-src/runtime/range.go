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
	"sync"
)

var (
	// enumerateType is the object representing the Python 'enumerate' type.
	enumerateType = newBasisType("enumerate", reflect.TypeOf(enumerate{}), toEnumerateUnsafe, ObjectType)
	// rangeIteratorType is the object representing the Python 'rangeiterator' type.
	rangeIteratorType = newBasisType("rangeiterator", reflect.TypeOf(rangeIterator{}), toRangeIteratorUnsafe, ObjectType)
	// xrangeType is the object representing the Python 'xrange' type.
	xrangeType = newBasisType("xrange", reflect.TypeOf(xrange{}), toXRangeUnsafe, ObjectType)
)

type enumerate struct {
	Object
	mutex sync.Mutex
	index int
	iter  *Object
}

func toEnumerateUnsafe(o *Object) *enumerate {
	return (*enumerate)(o.toPointer())
}

func enumerateIter(f *Frame, o *Object) (*Object, *BaseException) {
	return o, nil
}

func enumerateNew(f *Frame, t *Type, args Args, _ KWArgs) (*Object, *BaseException) {
	expectedTypes := []*Type{ObjectType, ObjectType}
	argc := len(args)
	if argc == 1 {
		expectedTypes = expectedTypes[:1]
	}
	if raised := checkFunctionArgs(f, "__new__", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	index := 0
	if argc > 1 {
		if args[1].typ.slots.Index == nil {
			format := "%s object cannot be interpreted as an index"
			return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, args[1].typ.Name()))
		}
		// TODO: support long?
		i, raised := IndexInt(f, args[1])
		if raised != nil {
			return nil, raised
		}
		if i > 0 {
			index = i
		}
	}
	iter, raised := Iter(f, args[0])
	if raised != nil {
		return nil, raised
	}
	for i := 0; i < index; i++ {
		_, raised := Next(f, iter)
		if raised != nil {
			if !raised.isInstance(StopIterationType) {
				return nil, raised
			}
			index = -1
			f.RestoreExc(nil, nil)
			break
		}
	}
	var d *Dict
	if t != enumerateType {
		d = NewDict()
	}
	e := &enumerate{Object: Object{typ: t, dict: d}, index: index, iter: iter}
	return &e.Object, nil
}

func enumerateNext(f *Frame, o *Object) (ret *Object, raised *BaseException) {
	e := toEnumerateUnsafe(o)
	e.mutex.Lock()
	var item *Object
	if e.index != -1 {
		item, raised = Next(f, e.iter)
	}
	if raised == nil {
		if item == nil {
			raised = f.Raise(StopIterationType.ToObject(), nil, nil)
			e.index = -1
		} else {
			ret = NewTuple2(NewInt(e.index).ToObject(), item).ToObject()
			e.index++
		}
	}
	e.mutex.Unlock()
	return ret, raised
}

func initEnumerateType(map[string]*Object) {
	enumerateType.slots.Iter = &unaryOpSlot{enumerateIter}
	enumerateType.slots.Next = &unaryOpSlot{enumerateNext}
	enumerateType.slots.New = &newSlot{enumerateNew}
}

// TODO: Synchronize access to this structure.
type rangeIterator struct {
	Object
	i    int
	stop int
	step int
}

func toRangeIteratorUnsafe(o *Object) *rangeIterator {
	return (*rangeIterator)(o.toPointer())
}

func rangeIteratorIter(f *Frame, o *Object) (*Object, *BaseException) {
	return o, nil
}

func rangeIteratorNext(f *Frame, o *Object) (*Object, *BaseException) {
	iter := toRangeIteratorUnsafe(o)
	if iter.i == iter.stop {
		return nil, f.Raise(StopIterationType.ToObject(), nil, nil)
	}
	ret := NewInt(iter.i)
	iter.i += iter.step
	return ret.ToObject(), nil
}

func initRangeIteratorType(map[string]*Object) {
	rangeIteratorType.flags &^= typeFlagInstantiable | typeFlagBasetype
	rangeIteratorType.slots.Iter = &unaryOpSlot{rangeIteratorIter}
	rangeIteratorType.slots.Next = &unaryOpSlot{rangeIteratorNext}
}

type xrange struct {
	Object
	start int
	stop  int
	step  int
}

func toXRangeUnsafe(o *Object) *xrange {
	return (*xrange)(o.toPointer())
}

func xrangeGetItem(f *Frame, o, key *Object) (*Object, *BaseException) {
	if key.typ.slots.Index == nil {
		format := "sequence index must be integer, not '%s'"
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, key.typ.Name()))
	}
	i, raised := IndexInt(f, key)
	if raised != nil {
		return nil, raised
	}
	r := toXRangeUnsafe(o)
	i, raised = seqCheckedIndex(f, (r.stop-r.start)/r.step, i)
	if raised != nil {
		return nil, raised
	}
	return NewInt(r.start + i*r.step).ToObject(), nil
}

func xrangeIter(f *Frame, o *Object) (*Object, *BaseException) {
	r := toXRangeUnsafe(o)
	return &(&rangeIterator{Object{typ: rangeIteratorType}, r.start, r.stop, r.step}).Object, nil
}

func xrangeLen(f *Frame, o *Object) (*Object, *BaseException) {
	r := toXRangeUnsafe(o)
	return NewInt((r.stop - r.start) / r.step).ToObject(), nil
}

func xrangeNew(f *Frame, _ *Type, args Args, _ KWArgs) (*Object, *BaseException) {
	expectedTypes := []*Type{IntType, IntType, IntType}
	argc := len(args)
	if argc > 0 && argc < 3 {
		expectedTypes = expectedTypes[:argc]
	}
	if raised := checkMethodArgs(f, "__new__", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	start, stop, step := 0, 0, 1
	if argc == 1 {
		stop = toIntUnsafe(args[0]).Value()
	} else {
		start = toIntUnsafe(args[0]).Value()
		stop = toIntUnsafe(args[1]).Value()
		if argc > 2 {
			step = toIntUnsafe(args[2]).Value()
		}
	}
	stop, _, result := seqRange(start, stop, step)
	switch result {
	case seqRangeZeroStep:
		return nil, f.RaiseType(ValueErrorType, "xrange() arg 3 must not be zero")
	case seqRangeOverflow:
		return nil, f.RaiseType(OverflowErrorType, errResultTooLarge)
	}
	r := &xrange{Object: Object{typ: xrangeType}, start: start, stop: stop, step: step}
	return &r.Object, nil
}

func xrangeRepr(_ *Frame, o *Object) (*Object, *BaseException) {
	r := toXRangeUnsafe(o)
	s := ""
	if r.step != 1 {
		s = fmt.Sprintf("xrange(%d, %d, %d)", r.start, r.stop, r.step)
	} else if r.start != 0 {
		s = fmt.Sprintf("xrange(%d, %d)", r.start, r.stop)
	} else {
		s = fmt.Sprintf("xrange(%d)", r.stop)
	}
	return NewStr(s).ToObject(), nil
}

func initXRangeType(map[string]*Object) {
	xrangeType.flags &^= typeFlagBasetype
	xrangeType.slots.GetItem = &binaryOpSlot{xrangeGetItem}
	xrangeType.slots.Iter = &unaryOpSlot{xrangeIter}
	xrangeType.slots.Len = &unaryOpSlot{xrangeLen}
	xrangeType.slots.New = &newSlot{xrangeNew}
	xrangeType.slots.Repr = &unaryOpSlot{xrangeRepr}
}

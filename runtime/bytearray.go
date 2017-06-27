// Copyright 2017 Google Inc. All Rights Reserved.
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
	"bytes"
	"fmt"
	"reflect"
	"sync"
)

var (
	// ByteArrayType is the object representing the Python 'bytearray' type.
	ByteArrayType = newBasisType("bytearray", reflect.TypeOf(ByteArray{}), toByteArrayUnsafe, ObjectType)
)

// ByteArray represents Python 'bytearray' objects.
type ByteArray struct {
	Object
	mutex sync.RWMutex
	value []byte
}

func toByteArrayUnsafe(o *Object) *ByteArray {
	return (*ByteArray)(o.toPointer())
}

// ToObject upcasts a to an Object.
func (a *ByteArray) ToObject() *Object {
	return &a.Object
}

// Value returns the underlying bytes held by a.
func (a *ByteArray) Value() []byte {
	return a.value
}

func byteArrayEq(f *Frame, v, w *Object) (*Object, *BaseException) {
	return byteArrayCompare(v, w, False, True, False), nil
}

func byteArrayGE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return byteArrayCompare(v, w, False, True, True), nil
}

func byteArrayGetItem(f *Frame, o, key *Object) (result *Object, raised *BaseException) {
	a := toByteArrayUnsafe(o)
	if key.typ.slots.Index != nil {
		index, raised := IndexInt(f, key)
		if raised != nil {
			return nil, raised
		}
		a.mutex.RLock()
		elems := a.Value()
		if index, raised = seqCheckedIndex(f, len(elems), index); raised == nil {
			result = NewInt(int(elems[index])).ToObject()
		}
		a.mutex.RUnlock()
		return result, raised
	}
	if key.isInstance(SliceType) {
		a.mutex.RLock()
		elems := a.Value()
		start, stop, step, sliceLen, raised := toSliceUnsafe(key).calcSlice(f, len(elems))
		if raised == nil {
			value := make([]byte, sliceLen)
			if step == 1 {
				copy(value, elems[start:stop])
			} else {
				i := 0
				for j := start; j != stop; j += step {
					value[i] = elems[j]
					i++
				}
			}
			result = (&ByteArray{Object: Object{typ: ByteArrayType}, value: value}).ToObject()
		}
		a.mutex.RUnlock()
		return result, raised
	}
	return nil, f.RaiseType(TypeErrorType, fmt.Sprintf("bytearray indices must be integers or slice, not %s", key.typ.Name()))
}

func byteArrayGT(f *Frame, v, w *Object) (*Object, *BaseException) {
	return byteArrayCompare(v, w, False, False, True), nil
}

func byteArrayInit(f *Frame, o *Object, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkFunctionArgs(f, "__init__", args, IntType); raised != nil {
		return nil, raised
	}
	a := toByteArrayUnsafe(o)
	a.mutex.Lock()
	a.value = make([]byte, toIntUnsafe(args[0]).Value())
	a.mutex.Unlock()
	return None, nil
}

func byteArrayLE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return byteArrayCompare(v, w, True, True, False), nil
}

func byteArrayLT(f *Frame, v, w *Object) (*Object, *BaseException) {
	return byteArrayCompare(v, w, True, False, False), nil
}

func byteArrayNative(f *Frame, o *Object) (reflect.Value, *BaseException) {
	a := toByteArrayUnsafe(o)
	a.mutex.RLock()
	result := reflect.ValueOf(a.Value())
	a.mutex.RUnlock()
	return result, nil
}

func byteArrayNE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return byteArrayCompare(v, w, True, False, True), nil
}

func byteArrayRepr(f *Frame, o *Object) (*Object, *BaseException) {
	a := toByteArrayUnsafe(o)
	a.mutex.RLock()
	s, raised := Repr(f, NewStr(string(a.Value())).ToObject())
	a.mutex.RUnlock()
	if raised != nil {
		return nil, raised
	}
	return NewStr(fmt.Sprintf("bytearray(b%s)", s.Value())).ToObject(), nil
}

func byteArrayStr(f *Frame, o *Object) (*Object, *BaseException) {
	a := toByteArrayUnsafe(o)
	a.mutex.RLock()
	s := string(a.Value())
	a.mutex.RUnlock()
	return NewStr(s).ToObject(), nil
}

func initByteArrayType(dict map[string]*Object) {
	ByteArrayType.slots.Eq = &binaryOpSlot{byteArrayEq}
	ByteArrayType.slots.GE = &binaryOpSlot{byteArrayGE}
	ByteArrayType.slots.GetItem = &binaryOpSlot{byteArrayGetItem}
	ByteArrayType.slots.GT = &binaryOpSlot{byteArrayGT}
	ByteArrayType.slots.Init = &initSlot{byteArrayInit}
	ByteArrayType.slots.LE = &binaryOpSlot{byteArrayLE}
	ByteArrayType.slots.LT = &binaryOpSlot{byteArrayLT}
	ByteArrayType.slots.Native = &nativeSlot{byteArrayNative}
	ByteArrayType.slots.NE = &binaryOpSlot{byteArrayNE}
	ByteArrayType.slots.Repr = &unaryOpSlot{byteArrayRepr}
	ByteArrayType.slots.Str = &unaryOpSlot{byteArrayStr}
}

func byteArrayCompare(v, w *Object, ltResult, eqResult, gtResult *Int) *Object {
	if v == w {
		return eqResult.ToObject()
	}
	// For simplicity we make a copy of w if it's a str or bytearray. This
	// is inefficient and it may be useful to optimize.
	var data []byte
	switch {
	case w.isInstance(StrType):
		data = []byte(toStrUnsafe(w).Value())
	case w.isInstance(ByteArrayType):
		a := toByteArrayUnsafe(w)
		a.mutex.RLock()
		data = make([]byte, len(a.value))
		copy(data, a.value)
		a.mutex.RUnlock()
	default:
		return NotImplemented
	}
	a := toByteArrayUnsafe(v)
	a.mutex.RLock()
	cmp := bytes.Compare(a.value, data)
	a.mutex.RUnlock()
	switch cmp {
	case -1:
		return ltResult.ToObject()
	case 0:
		return eqResult.ToObject()
	default:
		return gtResult.ToObject()
	}
}

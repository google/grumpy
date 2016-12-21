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

var (
	// SetType is the object representing the Python 'set' type.
	SetType = newBasisType("set", reflect.TypeOf(Set{}), toSetUnsafe, ObjectType)
)

// Set represents Python 'set' objects.
type Set struct {
	Object
	dict *Dict
}

// NewSet returns an empty Set.
func NewSet() *Set {
	return &Set{Object{typ: SetType}, NewDict()}
}

func toSetUnsafe(o *Object) *Set {
	return (*Set)(o.toPointer())
}

// Add inserts key into s. If key already exists then does nothing.
func (s *Set) Add(f *Frame, key *Object) (bool, *BaseException) {
	return s.dict.putItem(f, key, None)
}

// Contains returns true if key exists in s.
func (s *Set) Contains(f *Frame, key *Object) (bool, *BaseException) {
	item, raised := s.dict.GetItem(f, key)
	if raised != nil {
		return false, raised
	}
	return item != nil, nil
}

// Remove erases key from s. If key is not in s then raises KeyError.
func (s *Set) Remove(f *Frame, key *Object) (bool, *BaseException) {
	return s.dict.DelItem(f, key)
}

// ToObject upcasts s to an Object.
func (s *Set) ToObject() *Object {
	return &s.Object
}

// Update inserts all elements in the iterable o into s.
func (s *Set) Update(f *Frame, o *Object) *BaseException {
	raised := seqForEach(f, o, func(key *Object) *BaseException {
		if raised := s.dict.SetItem(f, key, None); raised != nil {
			return raised
		}
		return nil
	})
	return raised
}

func setAdd(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "add", args, SetType, ObjectType); raised != nil {
		return nil, raised
	}
	if _, raised := toSetUnsafe(args[0]).Add(f, args[1]); raised != nil {
		return nil, raised
	}
	return None, nil
}

func setContains(f *Frame, seq, value *Object) (*Object, *BaseException) {
	contains, raised := toSetUnsafe(seq).Contains(f, value)
	if raised != nil {
		return nil, raised
	}
	return GetBool(contains).ToObject(), nil
}

func setDiscard(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "discard", args, SetType, ObjectType); raised != nil {
		return nil, raised
	}
	if _, raised := toSetUnsafe(args[0]).Remove(f, args[1]); raised != nil {
		return nil, raised
	}
	return None, nil
}

func setEq(f *Frame, v, w *Object) (*Object, *BaseException) {
	return setCompare(f, compareOpEq, toSetUnsafe(v), w)
}

func setGE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return setCompare(f, compareOpGE, toSetUnsafe(v), w)
}

func setGT(f *Frame, v, w *Object) (*Object, *BaseException) {
	return setCompare(f, compareOpGT, toSetUnsafe(v), w)
}

func setInit(f *Frame, o *Object, args Args, _ KWArgs) (*Object, *BaseException) {
	argc := len(args)
	if argc > 1 {
		return nil, f.RaiseType(TypeErrorType, "set expected at most 1 arguments, got 2")
	}
	s := toSetUnsafe(o)
	if argc == 1 {
		if raised := s.Update(f, args[0]); raised != nil {
			return nil, raised
		}
	}
	return None, nil
}

func setIsSubset(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "issubset", args, SetType, ObjectType); raised != nil {
		return nil, raised
	}
	s, raised := setFromSeq(f, args[1])
	if raised != nil {
		return nil, raised
	}
	return setCompare(f, compareOpLE, toSetUnsafe(args[0]), s.ToObject())
}

func setIsSuperset(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "issuperset", args, SetType, ObjectType); raised != nil {
		return nil, raised
	}
	s, raised := setFromSeq(f, args[1])
	if raised != nil {
		return nil, raised
	}
	return setCompare(f, compareOpGE, toSetUnsafe(args[0]), s.ToObject())
}

func setIter(f *Frame, o *Object) (*Object, *BaseException) {
	s := toSetUnsafe(o)
	s.dict.mutex.Lock(f)
	iter := &newDictKeyIterator(s.dict).Object
	s.dict.mutex.Unlock(f)
	return iter, nil
}

func setLE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return setCompare(f, compareOpLE, toSetUnsafe(v), w)
}

func setLen(f *Frame, o *Object) (*Object, *BaseException) {
	return NewInt(toSetUnsafe(o).dict.Len()).ToObject(), nil
}

func setLT(f *Frame, v, w *Object) (*Object, *BaseException) {
	return setCompare(f, compareOpLT, toSetUnsafe(v), w)
}

func setNE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return setCompare(f, compareOpNE, toSetUnsafe(v), w)
}

func setNew(f *Frame, t *Type, _ Args, _ KWArgs) (*Object, *BaseException) {
	s := toSetUnsafe(newObject(t))
	s.dict = NewDict()
	return s.ToObject(), nil
}

func setRemove(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "remove", args, SetType, ObjectType); raised != nil {
		return nil, raised
	}
	key := args[1]
	if removed, raised := toSetUnsafe(args[0]).Remove(f, key); raised != nil {
		return nil, raised
	} else if !removed {
		return nil, raiseKeyError(f, key)
	}
	return None, nil
}

func setRepr(f *Frame, o *Object) (*Object, *BaseException) {
	s := toSetUnsafe(o)
	if f.reprEnter(s.ToObject()) {
		return NewStr(fmt.Sprintf("%s(...)", s.typ.Name())).ToObject(), nil
	}
	repr, raised := Repr(f, s.dict.Keys(f).ToObject())
	f.reprLeave(s.ToObject())
	if raised != nil {
		return nil, raised
	}
	return NewStr(fmt.Sprintf("%s(%s)", s.typ.Name(), repr.Value())).ToObject(), nil
}

func setUpdate(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "update", args, SetType, ObjectType); raised != nil {
		return nil, raised
	}
	if raised := toSetUnsafe(args[0]).Update(f, args[1]); raised != nil {
		return nil, raised
	}
	return None, nil
}

func initSetType(dict map[string]*Object) {
	dict["add"] = newBuiltinFunction("add", setAdd).ToObject()
	dict["discard"] = newBuiltinFunction("discard", setDiscard).ToObject()
	dict["issubset"] = newBuiltinFunction("issubset", setIsSubset).ToObject()
	dict["issuperset"] = newBuiltinFunction("issuperset", setIsSuperset).ToObject()
	dict["remove"] = newBuiltinFunction("remove", setRemove).ToObject()
	dict["update"] = newBuiltinFunction("update", setUpdate).ToObject()
	SetType.slots.Contains = &binaryOpSlot{setContains}
	SetType.slots.Eq = &binaryOpSlot{setEq}
	SetType.slots.GE = &binaryOpSlot{setGE}
	SetType.slots.GT = &binaryOpSlot{setGT}
	SetType.slots.Hash = &unaryOpSlot{hashNotImplemented}
	SetType.slots.Init = &initSlot{setInit}
	SetType.slots.Iter = &unaryOpSlot{setIter}
	SetType.slots.LE = &binaryOpSlot{setLE}
	SetType.slots.Len = &unaryOpSlot{setLen}
	SetType.slots.LT = &binaryOpSlot{setLT}
	SetType.slots.NE = &binaryOpSlot{setNE}
	SetType.slots.New = &newSlot{setNew}
	SetType.slots.Repr = &unaryOpSlot{setRepr}
}

func setCompare(f *Frame, op compareOp, v *Set, w *Object) (*Object, *BaseException) {
	if !w.isInstance(SetType) {
		return NotImplemented, nil
	}
	s2 := toSetUnsafe(w)
	if op == compareOpGE || op == compareOpGT {
		op = op.swapped()
		v, s2 = s2, v
	}
	v.dict.mutex.Lock(f)
	iter := newDictEntryIterator(v.dict)
	g1 := newDictVersionGuard(v.dict)
	len1 := v.dict.Len()
	v.dict.mutex.Unlock(f)
	s2.dict.mutex.Lock(f)
	g2 := newDictVersionGuard(s2.dict)
	len2 := s2.dict.Len()
	s2.dict.mutex.Unlock(f)
	result := (op != compareOpNE)
	switch op {
	case compareOpLT:
		if len1 >= len2 {
			return False.ToObject(), nil
		}
	case compareOpLE:
		if len1 > len2 {
			return False.ToObject(), nil
		}
	case compareOpEq, compareOpNE:
		if len1 != len2 {
			return GetBool(!result).ToObject(), nil
		}
	}
	for entry := iter.next(); entry != nil; entry = iter.next() {
		contains, raised := s2.Contains(f, entry.key)
		if raised != nil {
			return nil, raised
		}
		if !contains {
			result = !result
			break
		}
	}
	if !g1.check() || !g2.check() {
		return nil, f.RaiseType(RuntimeErrorType, "set changed during iteration")
	}
	return GetBool(result).ToObject(), nil
}

func setFromSeq(f *Frame, seq *Object) (*Set, *BaseException) {
	if seq.isInstance(SetType) {
		return toSetUnsafe(seq), nil
	}
	o, raised := SetType.Call(f, Args{seq}, nil)
	if raised != nil {
		return nil, raised
	}
	return toSetUnsafe(o), nil
}

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
	"math"
	"math/big"
	"reflect"
	"strings"
	"sync"
)

// By convention in this file, we always use the variable
// name "z" or "m" (in the case of DivMod) to refer to
// a *big.Int that we intend to modify. For other big.Int
// values, we use "x", "y" or other names. Variables z and m
// must be allocated within the same function using a big.Int
// constructor.
// We must never modify the value field of a Long that has
// already been made available to the rest of the program,
// as this would violate the immutability of Python longs.

// Long represents Python 'long' objects.
type Long struct {
	Object
	value    big.Int
	hashOnce sync.Once
	hash     int
}

// NewLong returns a new Long holding the given value.
func NewLong(x *big.Int) *Long {
	result := Long{Object: Object{typ: LongType}}
	result.value.Set(x)
	return &result
}

// NewLongFromBytes returns a new Long holding the given bytes,
// interpreted as a big endian unsigned integer.
func NewLongFromBytes(b []byte) *Long {
	result := Long{Object: Object{typ: LongType}}
	result.value.SetBytes(b)
	return &result
}

func toLongUnsafe(o *Object) *Long {
	return (*Long)(o.toPointer())
}

// IntValue returns l's value as a plain int if it will not overflow.
// Otherwise raises OverflowErrorType.
func (l *Long) IntValue(f *Frame) (int, *BaseException) {
	if !numInIntRange(&l.value) {
		return 0, f.RaiseType(OverflowErrorType, "Python int too large to convert to a Go int")
	}
	return int(l.value.Int64()), nil
}

// ToObject upcasts l to an Object.
func (l *Long) ToObject() *Object {
	return &l.Object
}

// Value returns the underlying integer value held by l.
func (l *Long) Value() *big.Int {
	return new(big.Int).Set(&l.value)
}

// IsTrue returns false if l is zero, true otherwise.
func (l *Long) IsTrue() bool {
	return l.value.Sign() != 0
}

// Neg returns a new Long that is the negative of l.
func (l *Long) Neg() *Long {
	result := Long{Object: Object{typ: LongType}}
	result.value.Set(&l.value)
	result.value.Neg(&result.value)
	return &result
}

// LongType is the object representing the Python 'long' type.
var LongType = newBasisType("long", reflect.TypeOf(Long{}), toLongUnsafe, ObjectType)

func longAbs(z, x *big.Int) {
	z.Abs(x)
}

func longAdd(z, x, y *big.Int) {
	z.Add(x, y)
}

func longAnd(z, x, y *big.Int) {
	z.And(x, y)
}

func longDiv(z, x, y *big.Int) {
	m := big.Int{}
	longDivMod(x, y, z, &m)
}

func longDivAndMod(z, m, x, y *big.Int) {
	longDivMod(x, y, z, m)
}

func longEq(x, y *big.Int) bool {
	return x.Cmp(y) == 0
}

func longGE(x, y *big.Int) bool {
	return x.Cmp(y) >= 0
}

func longGetNewArgs(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "__getnewargs__", args, LongType); raised != nil {
		return nil, raised
	}
	return NewTuple1(args[0]).ToObject(), nil
}

func longGT(x, y *big.Int) bool {
	return x.Cmp(y) > 0
}

func longFloat(f *Frame, o *Object) (*Object, *BaseException) {
	flt, _ := new(big.Float).SetInt(&toLongUnsafe(o).value).Float64()
	if math.IsInf(flt, 0) {
		return nil, f.RaiseType(OverflowErrorType, "long int too large to convert to float")
	}
	return NewFloat(flt).ToObject(), nil
}

func hashBigInt(x *big.Int) int {
	// TODO: Make this hash match that of cpython.
	return hashString(x.Text(36))
}

func longHex(f *Frame, o *Object) (*Object, *BaseException) {
	val := numberToBase("0x", 16, o) + "L"
	return NewStr(val).ToObject(), nil
}

func longHash(f *Frame, o *Object) (*Object, *BaseException) {
	l := toLongUnsafe(o)
	l.hashOnce.Do(func() {
		// Be compatible with int hashes.
		if numInIntRange(&l.value) {
			l.hash = int(l.value.Int64())
		}
		l.hash = hashBigInt(&l.value)
	})
	return NewInt(l.hash).ToObject(), nil
}

func longIndex(_ *Frame, o *Object) (*Object, *BaseException) {
	return o, nil
}

func longInt(f *Frame, o *Object) (*Object, *BaseException) {
	if l := &toLongUnsafe(o).value; numInIntRange(l) {
		return NewInt(int(l.Int64())).ToObject(), nil
	}
	return o, nil
}

func longInvert(z, x *big.Int) {
	z.Not(x)
}

func longLE(x, y *big.Int) bool {
	return x.Cmp(y) <= 0
}

func longLShift(z, x *big.Int, n uint) {
	z.Lsh(x, n)
}

func longLong(f *Frame, o *Object) (*Object, *BaseException) {
	if o.typ == LongType {
		return o, nil
	}
	l := Long{Object: Object{typ: LongType}}
	l.value.Set(&toLongUnsafe(o).value)
	return l.ToObject(), nil
}

func longLT(x, y *big.Int) bool {
	return x.Cmp(y) < 0
}

func longMul(z, x, y *big.Int) {
	z.Mul(x, y)
}

func longMod(m, x, y *big.Int) {
	z := &big.Int{}
	longDivMod(x, y, z, m)
}

func longNative(f *Frame, o *Object) (reflect.Value, *BaseException) {
	return reflect.ValueOf(toLongUnsafe(o).Value()), nil
}

func longNE(x, y *big.Int) bool {
	return x.Cmp(y) != 0
}

func longNeg(z, x *big.Int) {
	z.Neg(x)
}

func longNew(f *Frame, t *Type, args Args, _ KWArgs) (*Object, *BaseException) {
	if t != LongType {
		// Allocate a plain long and then copy its value into an
		// object of the long subtype.
		i, raised := longNew(f, LongType, args, nil)
		if raised != nil {
			return nil, raised
		}
		result := toLongUnsafe(newObject(t))
		result.value = toLongUnsafe(i).value
		return result.ToObject(), nil
	}
	argc := len(args)
	if argc == 0 {
		return NewLong(big.NewInt(0)).ToObject(), nil
	}
	o := args[0]
	baseArg := 10
	if argc == 1 {
		if slot := o.typ.slots.Long; slot != nil {
			result, raised := slot.Fn(f, o)
			if raised != nil {
				return nil, raised
			}
			if !result.isInstance(LongType) {
				format := "__long__ returned non-long (type %s)"
				return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, result.typ.Name()))
			}
			return result, nil
		}
		if raised := checkMethodArgs(f, "__new__", args, StrType); raised != nil {
			return nil, raised
		}
	} else {
		if raised := checkMethodArgs(f, "__new__", args, StrType, IntType); raised != nil {
			return nil, raised
		}
		baseArg = toIntUnsafe(args[1]).Value()
		if baseArg != 0 && (baseArg < 2 || baseArg > 36) {
			return nil, f.RaiseType(ValueErrorType, "long() base must be >= 2 and <= 36")
		}
	}
	s := strings.TrimSpace(toStrUnsafe(o).Value())
	if len(s) > 0 && (s[len(s)-1] == 'L' || s[len(s)-1] == 'l') {
		s = s[:len(s)-1]
	}
	base := baseArg
	if len(s) > 2 {
		detectedBase := 0
		switch s[:2] {
		case "0b", "0B":
			detectedBase = 2
		case "0o", "0O":
			detectedBase = 8
		case "0x", "0X":
			detectedBase = 16
		}
		if detectedBase != 0 && (baseArg == 0 || baseArg == detectedBase) {
			s = s[2:]
			base = detectedBase
		}
	}
	if base == 0 {
		base = 10
	}
	i := big.Int{}
	if _, ok := i.SetString(s, base); !ok {
		format := "invalid literal for long() with base %d: %s"
		return nil, f.RaiseType(ValueErrorType, fmt.Sprintf(format, baseArg, toStrUnsafe(o).Value()))
	}
	return NewLong(&i).ToObject(), nil
}

func longNonZero(x *big.Int) bool {
	return x.Sign() != 0
}

func longOct(f *Frame, o *Object) (*Object, *BaseException) {
	val := numberToBase("0", 8, o) + "L"
	if val == "00L" {
		val = "0L"
	}
	return NewStr(val).ToObject(), nil
}

func longOr(z, x, y *big.Int) {
	z.Or(x, y)
}

func longPos(z, x *big.Int) {
	z.Set(x)
}

func longRepr(f *Frame, o *Object) (*Object, *BaseException) {
	return NewStr(toLongUnsafe(o).value.Text(10) + "L").ToObject(), nil
}

func longRShift(z, x *big.Int, n uint) {
	z.Rsh(x, n)
}

func longStr(f *Frame, o *Object) (*Object, *BaseException) {
	return NewStr(toLongUnsafe(o).value.Text(10)).ToObject(), nil
}

func longSub(z, x, y *big.Int) {
	z.Sub(x, y)
}

func longXor(z, x, y *big.Int) {
	z.Xor(x, y)
}

func initLongType(dict map[string]*Object) {
	dict["__getnewargs__"] = newBuiltinFunction("__getnewargs__", longGetNewArgs).ToObject()
	LongType.slots.Abs = longUnaryOpSlot(longAbs)
	LongType.slots.Add = longBinaryOpSlot(longAdd)
	LongType.slots.And = longBinaryOpSlot(longAnd)
	LongType.slots.Div = longDivModOpSlot(longDiv)
	LongType.slots.DivMod = longDivAndModOpSlot(longDivAndMod)
	LongType.slots.Eq = longBinaryBoolOpSlot(longEq)
	LongType.slots.Float = &unaryOpSlot{longFloat}
	LongType.slots.FloorDiv = longDivModOpSlot(longDiv)
	LongType.slots.GE = longBinaryBoolOpSlot(longGE)
	LongType.slots.GT = longBinaryBoolOpSlot(longGT)
	LongType.slots.Hash = &unaryOpSlot{longHash}
	LongType.slots.Hex = &unaryOpSlot{longHex}
	LongType.slots.Index = &unaryOpSlot{longIndex}
	LongType.slots.Int = &unaryOpSlot{longInt}
	LongType.slots.Invert = longUnaryOpSlot(longInvert)
	LongType.slots.LE = longBinaryBoolOpSlot(longLE)
	LongType.slots.LShift = longShiftOpSlot(longLShift)
	LongType.slots.LT = longBinaryBoolOpSlot(longLT)
	LongType.slots.Long = &unaryOpSlot{longLong}
	LongType.slots.Mod = longDivModOpSlot(longMod)
	LongType.slots.Mul = longBinaryOpSlot(longMul)
	LongType.slots.Native = &nativeSlot{longNative}
	LongType.slots.NE = longBinaryBoolOpSlot(longNE)
	LongType.slots.Neg = longUnaryOpSlot(longNeg)
	LongType.slots.New = &newSlot{longNew}
	LongType.slots.NonZero = longUnaryBoolOpSlot(longNonZero)
	LongType.slots.Oct = &unaryOpSlot{longOct}
	LongType.slots.Or = longBinaryOpSlot(longOr)
	LongType.slots.Pos = longUnaryOpSlot(longPos)
	// This operation can return a float, it must use binaryOpSlot directly.
	LongType.slots.Pow = &binaryOpSlot{longPow}
	LongType.slots.RAdd = longRBinaryOpSlot(longAdd)
	LongType.slots.RAnd = longRBinaryOpSlot(longAnd)
	LongType.slots.RDiv = longRDivModOpSlot(longDiv)
	LongType.slots.RDivMod = longRDivAndModOpSlot(longDivAndMod)
	LongType.slots.Repr = &unaryOpSlot{longRepr}
	LongType.slots.RFloorDiv = longRDivModOpSlot(longDiv)
	LongType.slots.RMod = longRDivModOpSlot(longMod)
	LongType.slots.RMul = longRBinaryOpSlot(longMul)
	LongType.slots.ROr = longRBinaryOpSlot(longOr)
	LongType.slots.RLShift = longRShiftOpSlot(longLShift)
	// This operation can return a float, it must use binaryOpSlot directly.
	LongType.slots.RPow = &binaryOpSlot{longRPow}
	LongType.slots.RRShift = longRShiftOpSlot(longRShift)
	LongType.slots.RShift = longShiftOpSlot(longRShift)
	LongType.slots.RSub = longRBinaryOpSlot(longSub)
	LongType.slots.RXor = longRBinaryOpSlot(longXor)
	LongType.slots.Str = &unaryOpSlot{longStr}
	LongType.slots.Sub = longBinaryOpSlot(longSub)
	LongType.slots.Xor = longBinaryOpSlot(longXor)
}

func longCallUnary(fun func(z, x *big.Int), v *Long) *Object {
	l := Long{Object: Object{typ: LongType}}
	fun(&l.value, &v.value)
	return l.ToObject()
}

func longCallUnaryBool(fun func(x *big.Int) bool, v *Long) *Object {
	return GetBool(fun(&v.value)).ToObject()
}

func longCallBinary(fun func(z, x, y *big.Int), v, w *Long) *Object {
	l := Long{Object: Object{typ: LongType}}
	fun(&l.value, &v.value, &w.value)
	return l.ToObject()
}

func longCallBinaryTuple(fun func(z, m, x, y *big.Int), v, w *Long) *Object {
	l := Long{Object: Object{typ: LongType}}
	ll := Long{Object: Object{typ: LongType}}
	fun(&l.value, &ll.value, &v.value, &w.value)
	return NewTuple2(l.ToObject(), ll.ToObject()).ToObject()
}

func longCallBinaryBool(fun func(x, y *big.Int) bool, v, w *Long) *Object {
	return GetBool(fun(&v.value, &w.value)).ToObject()
}

func longCallShift(fun func(z, x *big.Int, n uint), f *Frame, v, w *Long) (*Object, *BaseException) {
	if !numInIntRange(&w.value) {
		return nil, f.RaiseType(OverflowErrorType, "long int too large to convert to int")
	}
	if w.value.Sign() < 0 {
		return nil, f.RaiseType(ValueErrorType, "negative shift count")
	}
	l := Long{Object: Object{typ: LongType}}
	fun(&l.value, &v.value, uint(w.value.Int64()))
	return l.ToObject(), nil
}

func longCallDivMod(fun func(z, x, y *big.Int), f *Frame, v, w *Long) (*Object, *BaseException) {
	if w.value.Sign() == 0 {
		return nil, f.RaiseType(ZeroDivisionErrorType, "integer division or modulo by zero")
	}
	return longCallBinary(fun, v, w), nil
}

func longCallDivAndMod(fun func(z, m, x, y *big.Int), f *Frame, v, w *Long) (*Object, *BaseException) {
	if w.value.Sign() == 0 {
		return nil, f.RaiseType(ZeroDivisionErrorType, "integer division or modulo by zero")
	}
	return longCallBinaryTuple(fun, v, w), nil
}

func longUnaryOpSlot(fun func(z, x *big.Int)) *unaryOpSlot {
	f := func(_ *Frame, v *Object) (*Object, *BaseException) {
		return longCallUnary(fun, toLongUnsafe(v)), nil
	}
	return &unaryOpSlot{f}
}

func longUnaryBoolOpSlot(fun func(x *big.Int) bool) *unaryOpSlot {
	f := func(_ *Frame, v *Object) (*Object, *BaseException) {
		return longCallUnaryBool(fun, toLongUnsafe(v)), nil
	}
	return &unaryOpSlot{f}
}

func longBinaryOpSlot(fun func(z, x, y *big.Int)) *binaryOpSlot {
	f := func(_ *Frame, v, w *Object) (*Object, *BaseException) {
		if w.isInstance(IntType) {
			w = intToLong(toIntUnsafe(w)).ToObject()
		} else if !w.isInstance(LongType) {
			return NotImplemented, nil
		}
		return longCallBinary(fun, toLongUnsafe(v), toLongUnsafe(w)), nil
	}
	return &binaryOpSlot{f}
}

func longRBinaryOpSlot(fun func(z, x, y *big.Int)) *binaryOpSlot {
	f := func(_ *Frame, v, w *Object) (*Object, *BaseException) {
		if w.isInstance(IntType) {
			w = intToLong(toIntUnsafe(w)).ToObject()
		} else if !w.isInstance(LongType) {
			return NotImplemented, nil
		}
		return longCallBinary(fun, toLongUnsafe(w), toLongUnsafe(v)), nil
	}
	return &binaryOpSlot{f}
}

func longDivModOpSlot(fun func(z, x, y *big.Int)) *binaryOpSlot {
	f := func(f *Frame, v, w *Object) (*Object, *BaseException) {
		if w.isInstance(IntType) {
			w = intToLong(toIntUnsafe(w)).ToObject()
		} else if !w.isInstance(LongType) {
			return NotImplemented, nil
		}
		return longCallDivMod(fun, f, toLongUnsafe(v), toLongUnsafe(w))
	}
	return &binaryOpSlot{f}
}

func longRDivModOpSlot(fun func(z, x, y *big.Int)) *binaryOpSlot {
	f := func(f *Frame, v, w *Object) (*Object, *BaseException) {
		if w.isInstance(IntType) {
			w = intToLong(toIntUnsafe(w)).ToObject()
		} else if !w.isInstance(LongType) {
			return NotImplemented, nil
		}
		return longCallDivMod(fun, f, toLongUnsafe(w), toLongUnsafe(v))
	}
	return &binaryOpSlot{f}
}

func longDivAndModOpSlot(fun func(z, m, x, y *big.Int)) *binaryOpSlot {
	f := func(f *Frame, v, w *Object) (*Object, *BaseException) {
		if w.isInstance(IntType) {
			w = intToLong(toIntUnsafe(w)).ToObject()
		} else if !w.isInstance(LongType) {
			return NotImplemented, nil
		}
		return longCallDivAndMod(fun, f, toLongUnsafe(v), toLongUnsafe(w))
	}
	return &binaryOpSlot{f}
}

func longRDivAndModOpSlot(fun func(z, m, x, y *big.Int)) *binaryOpSlot {
	f := func(f *Frame, v, w *Object) (*Object, *BaseException) {
		if w.isInstance(IntType) {
			w = intToLong(toIntUnsafe(w)).ToObject()
		} else if !w.isInstance(LongType) {
			return NotImplemented, nil
		}
		return longCallDivAndMod(fun, f, toLongUnsafe(w), toLongUnsafe(v))
	}
	return &binaryOpSlot{f}
}

func longShiftOpSlot(fun func(z, x *big.Int, n uint)) *binaryOpSlot {
	f := func(f *Frame, v, w *Object) (*Object, *BaseException) {
		if w.isInstance(IntType) {
			w = intToLong(toIntUnsafe(w)).ToObject()
		} else if !w.isInstance(LongType) {
			return NotImplemented, nil
		}
		return longCallShift(fun, f, toLongUnsafe(v), toLongUnsafe(w))
	}
	return &binaryOpSlot{f}
}

func longRShiftOpSlot(fun func(z, x *big.Int, n uint)) *binaryOpSlot {
	f := func(f *Frame, v, w *Object) (*Object, *BaseException) {
		if w.isInstance(IntType) {
			w = intToLong(toIntUnsafe(w)).ToObject()
		} else if !w.isInstance(LongType) {
			return NotImplemented, nil
		}
		return longCallShift(fun, f, toLongUnsafe(w), toLongUnsafe(v))
	}
	return &binaryOpSlot{f}
}

func longBinaryBoolOpSlot(fun func(x, y *big.Int) bool) *binaryOpSlot {
	f := func(f *Frame, v, w *Object) (*Object, *BaseException) {
		if w.isInstance(IntType) {
			w = intToLong(toIntUnsafe(w)).ToObject()
		} else if !w.isInstance(LongType) {
			return NotImplemented, nil
		}
		return longCallBinaryBool(fun, toLongUnsafe(v), toLongUnsafe(w)), nil
	}
	return &binaryOpSlot{f}
}

func longRBinaryBoolOpSlot(fun func(x, y *big.Int) bool) *binaryOpSlot {
	f := func(f *Frame, v, w *Object) (*Object, *BaseException) {
		if w.isInstance(IntType) {
			w = intToLong(toIntUnsafe(w)).ToObject()
		} else if !w.isInstance(LongType) {
			return NotImplemented, nil
		}
		return longCallBinaryBool(fun, toLongUnsafe(w), toLongUnsafe(v)), nil
	}
	return &binaryOpSlot{f}
}

func longPow(f *Frame, v, w *Object) (*Object, *BaseException) {
	var wLong *big.Int

	vLong := toLongUnsafe(v).Value()
	if w.isInstance(LongType) {
		wLong = toLongUnsafe(w).Value()
	} else if w.isInstance(IntType) {
		wLong = big.NewInt(int64(toIntUnsafe(w).Value()))
	} else {
		return NotImplemented, nil
	}

	if wLong.Sign() < 0 {
		// The result will be a float, so we call the floating point function.
		var vFloat, wFloat *Object
		var raised *BaseException

		vFloat, raised = longFloat(f, v)
		if raised != nil {
			return nil, raised
		}
		// w might be an int or a long
		if w.isInstance(LongType) {
			wFloat, raised = longFloat(f, w)
			if raised != nil {
				return nil, raised
			}
		} else if w.isInstance(IntType) {
			wFloat = NewFloat(float64(toIntUnsafe(w).Value())).ToObject()
		} else {
			// This point should not be reachable
			return nil, f.RaiseType(SystemErrorType, "internal error in longPow")
		}
		return floatPow(f, vFloat, wFloat)
	}

	return NewLong(big.NewInt(0).Exp(vLong, wLong, nil)).ToObject(), nil
}

func longRPow(f *Frame, v, w *Object) (*Object, *BaseException) {
	if w.isInstance(LongType) {
		return longPow(f, w, v)
	}
	if w.isInstance(IntType) {
		wLong := NewLong(big.NewInt(int64(toIntUnsafe(w).Value()))).ToObject()
		return longPow(f, wLong, v)
	}
	return NotImplemented, nil
}

func longDivMod(x, y, z, m *big.Int) {
	z.QuoRem(x, y, m)
	if m.Sign() == -y.Sign() {
		// In Python the result of the modulo operator is always the
		// same sign as the divisor, whereas in Go, the result is
		// always the same sign as the dividend. Therefore we need to
		// do an adjustment when the sign of the modulo result differs
		// from that of the divisor.
		m.Add(m, y)
		// Relatedly, in Python the result of division truncates toward
		// negative infinity whereas it truncates toward zero in Go.
		// The fact that the signs of the divisor and the modulo result
		// differ implies that the quotient is also negative so we also
		// adjust the quotient here.
		z.Sub(z, big.NewInt(1))
	}
}

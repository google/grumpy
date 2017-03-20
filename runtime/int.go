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
	"strconv"
)

const (
	internedIntMin = -2
	internedIntMax = 300
)

var (
	internedInts = makeInternedInts()
)

// Int represents Python 'int' objects.
type Int struct {
	Object
	value int
}

// NewInt returns a new Int holding the given integer value.
func NewInt(value int) *Int {
	if value >= internedIntMin && value <= internedIntMax {
		return &internedInts[value-internedIntMin]
	}
	return &Int{Object{typ: IntType}, value}
}

func toIntUnsafe(o *Object) *Int {
	return (*Int)(o.toPointer())
}

// ToObject upcasts i to an Object.
func (i *Int) ToObject() *Object {
	return &i.Object
}

// Value returns the underlying integer value held by i.
func (i *Int) Value() int {
	return i.value
}

// IsTrue returns false if i is zero, true otherwise.
func (i *Int) IsTrue() bool {
	return i.Value() != 0
}

// IntType is the object representing the Python 'int' type.
var IntType = newBasisType("int", reflect.TypeOf(Int{}), toIntUnsafe, ObjectType)

func intAbs(f *Frame, o *Object) (*Object, *BaseException) {
	z := toIntUnsafe(o)
	if z.Value() > 0 {
		return z.ToObject(), nil
	}
	return intNeg(f, o)
}

func intAdd(f *Frame, v, w *Object) (*Object, *BaseException) {
	return intAddMulOp(f, "__add__", v, w, intCheckedAdd, longAdd)
}

func intAnd(f *Frame, v, w *Object) (*Object, *BaseException) {
	if !w.isInstance(IntType) {
		return NotImplemented, nil
	}
	return NewInt(toIntUnsafe(v).Value() & toIntUnsafe(w).Value()).ToObject(), nil
}

func intDiv(f *Frame, v, w *Object) (*Object, *BaseException) {
	return intDivModOp(f, "__div__", v, w, intCheckedDiv, longDiv)
}

func intDivMod(f *Frame, v, w *Object) (*Object, *BaseException) {
	return intDivAndModOp(f, "__divmod__", v, w, intCheckedDivMod, longDivAndMod)
}

func intEq(f *Frame, v, w *Object) (*Object, *BaseException) {
	return intCompare(compareOpEq, toIntUnsafe(v), w), nil
}

func intGE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return intCompare(compareOpGE, toIntUnsafe(v), w), nil
}

func intGetNewArgs(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "__getnewargs__", args, IntType); raised != nil {
		return nil, raised
	}
	return NewTuple1(args[0]).ToObject(), nil
}

func intGT(f *Frame, v, w *Object) (*Object, *BaseException) {
	return intCompare(compareOpGT, toIntUnsafe(v), w), nil
}

func intFloat(f *Frame, o *Object) (*Object, *BaseException) {
	i := toIntUnsafe(o).Value()
	return NewFloat(float64(i)).ToObject(), nil
}

func intHash(f *Frame, o *Object) (*Object, *BaseException) {
	return o, nil
}

func intHex(f *Frame, o *Object) (*Object, *BaseException) {
	val := numberToBase("0x", 16, o)
	return NewStr(val).ToObject(), nil
}

func intIndex(f *Frame, o *Object) (*Object, *BaseException) {
	return o, nil
}

func intInt(f *Frame, o *Object) (*Object, *BaseException) {
	if o.typ == IntType {
		return o, nil
	}
	return NewInt(toIntUnsafe(o).Value()).ToObject(), nil
}

func intInvert(f *Frame, o *Object) (*Object, *BaseException) {
	return NewInt(^toIntUnsafe(o).Value()).ToObject(), nil
}

func intLE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return intCompare(compareOpLE, toIntUnsafe(v), w), nil
}

func intLong(f *Frame, o *Object) (*Object, *BaseException) {
	return NewLong(big.NewInt(int64(toIntUnsafe(o).Value()))).ToObject(), nil
}

func intLShift(f *Frame, v, w *Object) (*Object, *BaseException) {
	return intShiftOp(f, v, w, func(v, w int) (int, int, bool) { return v, w, false })
}

func intLT(f *Frame, v, w *Object) (*Object, *BaseException) {
	return intCompare(compareOpLT, toIntUnsafe(v), w), nil
}

func intMod(f *Frame, v, w *Object) (*Object, *BaseException) {
	return intDivModOp(f, "__mod__", v, w, intCheckedMod, longMod)
}

func intMul(f *Frame, v, w *Object) (*Object, *BaseException) {
	return intAddMulOp(f, "__mul__", v, w, intCheckedMul, longMul)
}

func intNative(f *Frame, o *Object) (reflect.Value, *BaseException) {
	return reflect.ValueOf(toIntUnsafe(o).Value()), nil
}

func intNE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return intCompare(compareOpNE, toIntUnsafe(v), w), nil
}

func intNeg(f *Frame, o *Object) (*Object, *BaseException) {
	z := toIntUnsafe(o)
	if z.Value() == MinInt {
		nz := big.NewInt(int64(z.Value()))
		return NewLong(nz.Neg(nz)).ToObject(), nil
	}
	return NewInt(-z.Value()).ToObject(), nil
}

func intNew(f *Frame, t *Type, args Args, _ KWArgs) (*Object, *BaseException) {
	if len(args) == 0 {
		return newObject(t), nil
	}
	o := args[0]
	if len(args) == 1 && o.typ.slots.Int != nil {
		i, raised := ToInt(f, o)
		if raised != nil {
			return nil, raised
		}
		if t == IntType {
			return i, nil
		}
		n := 0
		if i.isInstance(LongType) {
			n, raised = toLongUnsafe(i).IntValue(f)
			if raised != nil {
				return nil, raised
			}
		} else {
			n = toIntUnsafe(i).Value()
		}
		ret := newObject(t)
		toIntUnsafe(ret).value = n
		return ret, nil
	}
	if len(args) > 2 {
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf("int() takes at most 2 arguments (%d given)", len(args)))
	}
	if !o.isInstance(StrType) {
		if len(args) == 2 {
			return nil, f.RaiseType(TypeErrorType, "int() can't convert non-string with explicit base")
		}
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf("int() argument must be a string or a number, not '%s'", o.typ.Name()))
	}
	s := toStrUnsafe(o).Value()
	base := 10
	if len(args) == 2 {
		var raised *BaseException
		base, raised = ToIntValue(f, args[1])
		if raised != nil {
			return nil, raised
		}
		if base < 0 || base == 1 || base > 36 {
			return nil, f.RaiseType(ValueErrorType, "int() base must be >= 2 and <= 36")
		}
	}
	i, ok := numParseInteger(new(big.Int), s, base)
	if !ok {
		format := "invalid literal for int() with base %d: %s"
		return nil, f.RaiseType(ValueErrorType, fmt.Sprintf(format, base, s))
	}
	if !numInIntRange(i) {
		if t == IntType {
			return NewLong(i).ToObject(), nil
		}
		return nil, f.RaiseType(OverflowErrorType, "Python int too large to convert to a Go int")
	}
	if t != IntType {
		o := newObject(t)
		toIntUnsafe(o).value = int(i.Int64())
		return o, nil
	}
	return NewInt(int(i.Int64())).ToObject(), nil
}

func intNonZero(f *Frame, o *Object) (*Object, *BaseException) {
	return GetBool(toIntUnsafe(o).Value() != 0).ToObject(), nil
}

func intOct(f *Frame, o *Object) (*Object, *BaseException) {
	val := numberToBase("0", 8, o)
	if val == "00" {
		val = "0"
	}
	return NewStr(val).ToObject(), nil
}

func intOr(f *Frame, v, w *Object) (*Object, *BaseException) {
	if !w.isInstance(IntType) {
		return NotImplemented, nil
	}
	return NewInt(toIntUnsafe(v).Value() | toIntUnsafe(w).Value()).ToObject(), nil
}

func intPos(f *Frame, o *Object) (*Object, *BaseException) {
	return o, nil
}

func intPow(f *Frame, v, w *Object) (*Object, *BaseException) {
	if w.isInstance(IntType) {
		// First try to use the faster floating point arithmetic
		// on the CPU, then falls back to slower methods.
		// IEEE float64 has 52bit of precision, so the result should be
		// less than MaxInt32 to be representable as an exact integer.
		// This assumes that int is at least 32bit.
		vInt := toIntUnsafe(v).Value()
		wInt := toIntUnsafe(w).Value()
		if 0 < vInt && vInt <= math.MaxInt32 && 0 < wInt && wInt <= math.MaxInt32 {
			res := math.Pow(float64(vInt), float64(wInt))
			// Can the result be interpreted as an int?
			if !math.IsNaN(res) && !math.IsInf(res, 0) && res <= math.MaxInt32 {
				return NewInt(int(res)).ToObject(), nil
			}
		}
		// Special cases.
		if vInt == 0 {
			if wInt < 0 {
				return nil, f.RaiseType(ZeroDivisionErrorType, "0.0 cannot be raised to a negative power")
			}
			if wInt == 0 {
				return NewInt(1).ToObject(), nil
			}
			return NewInt(0).ToObject(), nil
		}
		// If w < 0, the result must be a floating point number.
		// We convert both arguments to float and continue.
		if wInt < 0 {
			return floatPow(f, NewFloat(float64(vInt)).ToObject(), NewFloat(float64(wInt)).ToObject())
		}
		// Else we convert to Long and continue there.
		return longPow(f, NewLong(big.NewInt(int64(vInt))).ToObject(), NewLong(big.NewInt(int64(wInt))).ToObject())
	}
	return NotImplemented, nil
}

func intRAdd(f *Frame, v, w *Object) (*Object, *BaseException) {
	return intAddMulOp(f, "__radd__", v, w, intCheckedAdd, longAdd)
}

func intRDiv(f *Frame, v, w *Object) (*Object, *BaseException) {
	return intDivModOp(f, "__rdiv__", v, w, func(v, w int) (int, divModResult) {
		return intCheckedDiv(w, v)
	}, func(z, x, y *big.Int) {
		longDiv(z, y, x)
	})
}

func intRDivMod(f *Frame, v, w *Object) (*Object, *BaseException) {
	return intDivAndModOp(f, "__rdivmod__", v, w, func(v, w int) (int, int, divModResult) {
		return intCheckedDivMod(w, v)
	}, func(z, m, x, y *big.Int) {
		longDivAndMod(z, m, y, x)
	})
}

func intRepr(f *Frame, o *Object) (*Object, *BaseException) {
	return NewStr(strconv.FormatInt(int64(toIntUnsafe(o).Value()), 10)).ToObject(), nil
}

func intRMod(f *Frame, v, w *Object) (*Object, *BaseException) {
	return intDivModOp(f, "__rmod__", v, w, func(v, w int) (int, divModResult) {
		return intCheckedMod(w, v)
	}, func(z, x, y *big.Int) {
		longMod(z, y, x)
	})
}

func intRMul(f *Frame, v, w *Object) (*Object, *BaseException) {
	return intAddMulOp(f, "__rmul__", v, w, intCheckedMul, longMul)
}

func intRLShift(f *Frame, v, w *Object) (*Object, *BaseException) {
	return intShiftOp(f, v, w, func(v, w int) (int, int, bool) { return w, v, false })
}

func intRRShift(f *Frame, v, w *Object) (*Object, *BaseException) {
	return intShiftOp(f, v, w, func(v, w int) (int, int, bool) { return w, v, true })
}

func intRShift(f *Frame, v, w *Object) (*Object, *BaseException) {
	return intShiftOp(f, v, w, func(v, w int) (int, int, bool) { return v, w, true })
}

func intRSub(f *Frame, v, w *Object) (*Object, *BaseException) {
	return intAddMulOp(f, "__rsub__", v, w, func(v, w int) (int, bool) {
		return intCheckedSub(w, v)
	}, func(z, x, y *big.Int) {
		longSub(z, y, x)
	})
}

func intSub(f *Frame, v, w *Object) (*Object, *BaseException) {
	return intAddMulOp(f, "__sub__", v, w, intCheckedSub, longSub)
}

func intXor(f *Frame, v, w *Object) (*Object, *BaseException) {
	if !w.isInstance(IntType) {
		return NotImplemented, nil
	}
	return NewInt(toIntUnsafe(v).Value() ^ toIntUnsafe(w).Value()).ToObject(), nil
}

func initIntType(dict map[string]*Object) {
	dict["__getnewargs__"] = newBuiltinFunction("__getnewargs__", intGetNewArgs).ToObject()
	IntType.slots.Abs = &unaryOpSlot{intAbs}
	IntType.slots.Add = &binaryOpSlot{intAdd}
	IntType.slots.And = &binaryOpSlot{intAnd}
	IntType.slots.Div = &binaryOpSlot{intDiv}
	IntType.slots.DivMod = &binaryOpSlot{intDivMod}
	IntType.slots.Eq = &binaryOpSlot{intEq}
	IntType.slots.FloorDiv = &binaryOpSlot{intDiv}
	IntType.slots.GE = &binaryOpSlot{intGE}
	IntType.slots.GT = &binaryOpSlot{intGT}
	IntType.slots.Float = &unaryOpSlot{intFloat}
	IntType.slots.Hash = &unaryOpSlot{intHash}
	IntType.slots.Hex = &unaryOpSlot{intHex}
	IntType.slots.Index = &unaryOpSlot{intIndex}
	IntType.slots.Int = &unaryOpSlot{intInt}
	IntType.slots.Invert = &unaryOpSlot{intInvert}
	IntType.slots.LE = &binaryOpSlot{intLE}
	IntType.slots.LShift = &binaryOpSlot{intLShift}
	IntType.slots.LT = &binaryOpSlot{intLT}
	IntType.slots.Long = &unaryOpSlot{intLong}
	IntType.slots.Mod = &binaryOpSlot{intMod}
	IntType.slots.Mul = &binaryOpSlot{intMul}
	IntType.slots.Native = &nativeSlot{intNative}
	IntType.slots.NE = &binaryOpSlot{intNE}
	IntType.slots.Neg = &unaryOpSlot{intNeg}
	IntType.slots.New = &newSlot{intNew}
	IntType.slots.NonZero = &unaryOpSlot{intNonZero}
	IntType.slots.Oct = &unaryOpSlot{intOct}
	IntType.slots.Or = &binaryOpSlot{intOr}
	IntType.slots.Pos = &unaryOpSlot{intPos}
	IntType.slots.Pow = &binaryOpSlot{intPow}
	IntType.slots.RAdd = &binaryOpSlot{intRAdd}
	IntType.slots.RAnd = &binaryOpSlot{intAnd}
	IntType.slots.RDiv = &binaryOpSlot{intRDiv}
	IntType.slots.RDivMod = &binaryOpSlot{intRDivMod}
	IntType.slots.Repr = &unaryOpSlot{intRepr}
	IntType.slots.RFloorDiv = &binaryOpSlot{intRDiv}
	IntType.slots.RMod = &binaryOpSlot{intRMod}
	IntType.slots.RMul = &binaryOpSlot{intRMul}
	IntType.slots.ROr = &binaryOpSlot{intOr}
	IntType.slots.RLShift = &binaryOpSlot{intRLShift}
	IntType.slots.RRShift = &binaryOpSlot{intRRShift}
	IntType.slots.RShift = &binaryOpSlot{intRShift}
	IntType.slots.RSub = &binaryOpSlot{intRSub}
	IntType.slots.RXor = &binaryOpSlot{intXor}
	IntType.slots.Sub = &binaryOpSlot{intSub}
	IntType.slots.Xor = &binaryOpSlot{intXor}
}

type divModResult int

const (
	divModOK           divModResult = iota
	divModOverflow                  = iota
	divModZeroDivision              = iota
)

func intCompare(op compareOp, v *Int, w *Object) *Object {
	if !w.isInstance(IntType) {
		return NotImplemented
	}
	lhs, rhs := v.Value(), toIntUnsafe(w).Value()
	result := false
	switch op {
	case compareOpLT:
		result = lhs < rhs
	case compareOpLE:
		result = lhs <= rhs
	case compareOpEq:
		result = lhs == rhs
	case compareOpNE:
		result = lhs != rhs
	case compareOpGE:
		result = lhs >= rhs
	case compareOpGT:
		result = lhs > rhs
	}
	return GetBool(result).ToObject()
}

func intAddMulOp(f *Frame, method string, v, w *Object, fun func(v, w int) (int, bool), bigFun func(z, x, y *big.Int)) (*Object, *BaseException) {
	if !w.isInstance(IntType) {
		return NotImplemented, nil
	}
	r, ok := fun(toIntUnsafe(v).Value(), toIntUnsafe(w).Value())
	if !ok {
		return longCallBinary(bigFun, intToLong(toIntUnsafe(v)), intToLong(toIntUnsafe(w))), nil
	}
	return NewInt(r).ToObject(), nil
}

func intCheckedDiv(v, w int) (int, divModResult) {
	q, _, r := intCheckedDivMod(v, w)
	return q, r
}

func intCheckedDivMod(v, w int) (int, int, divModResult) {
	if w == 0 {
		return 0, 0, divModZeroDivision
	}
	if v == MinInt && w == -1 {
		return 0, 0, divModOverflow
	}
	q := v / w
	m := v % w
	if m != 0 && (w^m) < 0 {
		// In Python the result of the modulo operator is always the
		// same sign as the divisor, whereas in Go, the result is
		// always the same sign as the dividend. Therefore we need to
		// do an adjustment when the sign of the modulo result differs
		// from that of the divisor.
		m += w
		// Relatedly, in Python the result of division truncates toward
		// negative infinity whereas it truncates toward zero in Go.
		// The fact that the signs of the divisor and the modulo result
		// differ implies that the quotient is also negative so we also
		// adjust the quotient here.
		q--
	}
	return q, m, divModOK
}

func intCheckedAdd(v, w int) (int, bool) {
	if (v > 0 && w > MaxInt-v) || (v < 0 && w < MinInt-v) {
		return 0, false
	}
	return v + w, true
}

func intCheckedMod(v, w int) (int, divModResult) {
	_, m, r := intCheckedDivMod(v, w)
	return m, r
}

func intCheckedMul(v, w int) (int, bool) {
	if v == 0 || w == 0 || v == 1 || w == 1 {
		return v * w, true
	}
	// Since MinInt can only be multiplied by zero and one safely and we've
	// already handled that case above, we know this multiplication will
	// overflow. Unfortunately the division check below will fail to catch
	// this by coincidence: MinInt * -1 overflows to MinInt, causing the
	// expression x/w to overflow, coincidentally producing MinInt which
	// makes it seem as though the multiplication was correct.
	if v == MinInt || w == MinInt {
		return 0, false
	}
	x := v * w
	if x/w != v {
		return 0, false
	}
	return x, true
}

func intCheckedSub(v, w int) (int, bool) {
	if (w > 0 && v < MinInt+w) || (w < 0 && v > MaxInt+w) {
		return 0, false
	}
	return v - w, true
}

func intDivModOp(f *Frame, method string, v, w *Object, fun func(v, w int) (int, divModResult), bigFun func(z, x, y *big.Int)) (*Object, *BaseException) {
	if !w.isInstance(IntType) {
		return NotImplemented, nil
	}
	x, r := fun(toIntUnsafe(v).Value(), toIntUnsafe(w).Value())
	switch r {
	case divModOverflow:
		return longCallBinary(bigFun, intToLong(toIntUnsafe(v)), intToLong(toIntUnsafe(w))), nil
	case divModZeroDivision:
		return nil, f.RaiseType(ZeroDivisionErrorType, "integer division or modulo by zero")
	}
	return NewInt(x).ToObject(), nil
}

func intDivAndModOp(f *Frame, method string, v, w *Object, fun func(v, w int) (int, int, divModResult), bigFun func(z, m, x, y *big.Int)) (*Object, *BaseException) {
	if !w.isInstance(IntType) {
		return NotImplemented, nil
	}
	q, m, r := fun(toIntUnsafe(v).Value(), toIntUnsafe(w).Value())
	switch r {
	case divModOverflow:
		return longCallBinaryTuple(bigFun, intToLong(toIntUnsafe(v)), intToLong(toIntUnsafe(w))), nil
	case divModZeroDivision:
		return nil, f.RaiseType(ZeroDivisionErrorType, "integer division or modulo by zero")
	}
	return NewTuple2(NewInt(q).ToObject(), NewInt(m).ToObject()).ToObject(), nil
}

func intShiftOp(f *Frame, v, w *Object, fun func(int, int) (int, int, bool)) (*Object, *BaseException) {
	if !w.isInstance(IntType) {
		return NotImplemented, nil
	}
	lhs, rhs, rshift := fun(toIntUnsafe(v).Value(), toIntUnsafe(w).Value())
	if rhs < 0 {
		return nil, f.RaiseType(ValueErrorType, "negative shift count")
	}
	var result int
	n := uint(rhs)
	if rshift {
		result = lhs >> n
	} else {
		result = lhs << n
		if result>>n != lhs {
			return NewLong(new(big.Int).Lsh(big.NewInt(int64(lhs)), n)).ToObject(), nil
		}
	}
	return NewInt(result).ToObject(), nil
}

func intToLong(o *Int) *Long {
	return NewLong(big.NewInt(int64(o.Value())))
}

func makeInternedInts() [internedIntMax - internedIntMin + 1]Int {
	var ints [internedIntMax - internedIntMin + 1]Int
	for i := internedIntMin; i <= internedIntMax; i++ {
		ints[i-internedIntMin] = Int{Object{typ: IntType}, i}
	}
	return ints
}

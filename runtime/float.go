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
	"strings"
	"sync/atomic"
	"unicode"
	"unsafe"
)

// FloatType is the object representing the Python 'float' type.
var FloatType = newBasisType("float", reflect.TypeOf(Float{}), toFloatUnsafe, ObjectType)

// Float represents Python 'float' objects.
type Float struct {
	Object
	value float64
	hash  int
}

// NewFloat returns a new Float holding the given floating point value.
func NewFloat(value float64) *Float {
	return &Float{Object: Object{typ: FloatType}, value: value}
}

func toFloatUnsafe(o *Object) *Float {
	return (*Float)(o.toPointer())
}

// ToObject upcasts f to an Object.
func (f *Float) ToObject() *Object {
	return &f.Object
}

// Value returns the underlying floating point value held by f.
func (f *Float) Value() float64 {
	return f.value
}

func floatAbs(f *Frame, o *Object) (*Object, *BaseException) {
	z := toFloatUnsafe(o).Value()
	return NewFloat(math.Abs(z)).ToObject(), nil
}

func floatAdd(f *Frame, v, w *Object) (*Object, *BaseException) {
	return floatArithmeticOp(f, "__add__", v, w, func(v, w float64) float64 { return v + w })
}

func floatDiv(f *Frame, v, w *Object) (*Object, *BaseException) {
	return floatDivModOp(f, "__div__", v, w, func(v, w float64) (float64, bool) {
		if w == 0.0 {
			return 0, false
		}
		return v / w, true
	})
}

func floatDivMod(f *Frame, v, w *Object) (*Object, *BaseException) {
	return floatDivAndModOp(f, "__divmod__", v, w, func(v, w float64) (float64, float64, bool) {
		m, r := floatModFunc(v, w)
		if !r {
			return 0, 0, false
		}
		return math.Floor(v / w), m, true
	})
}

func floatEq(f *Frame, v, w *Object) (*Object, *BaseException) {
	return floatCompare(toFloatUnsafe(v), w, False, True, False), nil
}

func floatFloat(f *Frame, o *Object) (*Object, *BaseException) {
	return o, nil
}

func floatFloorDiv(f *Frame, v, w *Object) (*Object, *BaseException) {
	return floatDivModOp(f, "__floordiv__", v, w, func(v, w float64) (float64, bool) {
		if w == 0.0 {
			return 0, false
		}
		return math.Floor(v / w), true
	})
}

func floatGE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return floatCompare(toFloatUnsafe(v), w, False, True, True), nil
}

func floatGetNewArgs(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "__getnewargs__", args, FloatType); raised != nil {
		return nil, raised
	}
	return NewTuple1(args[0]).ToObject(), nil
}

func floatGT(f *Frame, v, w *Object) (*Object, *BaseException) {
	return floatCompare(toFloatUnsafe(v), w, False, False, True), nil
}

func floatHash(f *Frame, o *Object) (*Object, *BaseException) {
	v := toFloatUnsafe(o)
	p := (*unsafe.Pointer)(unsafe.Pointer(&v.hash))
	if lp := atomic.LoadPointer(p); lp != unsafe.Pointer(nil) {
		return (*Int)(lp).ToObject(), nil
	}
	hash := hashFloat(v.Value())
	if hash == -1 {
		hash--
	}
	h := NewInt(hash)
	atomic.StorePointer(p, unsafe.Pointer(h))

	return h.ToObject(), nil
}

func floatInt(f *Frame, o *Object) (*Object, *BaseException) {
	val := toFloatUnsafe(o).Value()
	if math.IsInf(val, 0) {
		return nil, f.RaiseType(OverflowErrorType, "cannot convert float infinity to integer")
	}
	if math.IsNaN(val) {
		return nil, f.RaiseType(OverflowErrorType, "cannot convert float NaN to integer")
	}
	i := big.Int{}
	big.NewFloat(val).Int(&i)
	if !numInIntRange(&i) {
		return NewLong(&i).ToObject(), nil
	}
	return NewInt(int(i.Int64())).ToObject(), nil
}

func floatLong(f *Frame, o *Object) (*Object, *BaseException) {
	val := toFloatUnsafe(o).Value()
	if math.IsInf(val, 0) {
		return nil, f.RaiseType(OverflowErrorType, "cannot convert float infinity to integer")
	}
	if math.IsNaN(val) {
		return nil, f.RaiseType(OverflowErrorType, "cannot convert float NaN to integer")
	}
	i, _ := big.NewFloat(val).Int(nil)
	return NewLong(i).ToObject(), nil
}

func floatLE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return floatCompare(toFloatUnsafe(v), w, True, True, False), nil
}

func floatLT(f *Frame, v, w *Object) (*Object, *BaseException) {
	return floatCompare(toFloatUnsafe(v), w, True, False, False), nil
}

func floatMod(f *Frame, v, w *Object) (*Object, *BaseException) {
	return floatDivModOp(f, "__mod__", v, w, floatModFunc)
}

func floatMul(f *Frame, v, w *Object) (*Object, *BaseException) {
	return floatArithmeticOp(f, "__mul__", v, w, func(v, w float64) float64 { return v * w })
}

func floatNative(f *Frame, o *Object) (reflect.Value, *BaseException) {
	return reflect.ValueOf(toFloatUnsafe(o).Value()), nil
}

func floatNE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return floatCompare(toFloatUnsafe(v), w, True, False, True), nil
}

func floatNeg(f *Frame, o *Object) (*Object, *BaseException) {
	z := toFloatUnsafe(o).Value()
	return NewFloat(-z).ToObject(), nil
}

func floatNew(f *Frame, t *Type, args Args, _ KWArgs) (*Object, *BaseException) {
	argc := len(args)
	if argc == 0 {
		return newObject(t), nil
	}
	if argc != 1 {
		return nil, f.RaiseType(TypeErrorType, "'__new__' of 'float' requires 0 or 1 arguments")
	}
	if t != FloatType {
		// Allocate a plain float then copy it's value into an object
		// of the float subtype.
		x, raised := floatNew(f, FloatType, args, nil)
		if raised != nil {
			return nil, raised
		}
		result := toFloatUnsafe(newObject(t))
		result.value = toFloatUnsafe(x).Value()
		return result.ToObject(), nil
	}
	o := args[0]
	if floatSlot := o.typ.slots.Float; floatSlot != nil {
		fl, raised := floatConvert(floatSlot, f, o)
		if raised != nil {
			return nil, raised
		}
		return fl.ToObject(), nil
	}
	if !o.isInstance(StrType) {
		return nil, f.RaiseType(TypeErrorType, "float() argument must be a string or a number")
	}
	s := toStrUnsafe(o).Value()
	result, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, f.RaiseType(ValueErrorType, fmt.Sprintf("could not convert string to float: %s", s))
	}
	return NewFloat(result).ToObject(), nil
}

func floatNonZero(f *Frame, o *Object) (*Object, *BaseException) {
	return GetBool(toFloatUnsafe(o).Value() != 0).ToObject(), nil
}

func floatPos(f *Frame, o *Object) (*Object, *BaseException) {
	return o, nil
}

func floatPow(f *Frame, v, w *Object) (*Object, *BaseException) {
	return floatArithmeticOp(f, "__pow__", v, w, func(v, w float64) float64 { return math.Pow(v, w) })
}

func floatRAdd(f *Frame, v, w *Object) (*Object, *BaseException) {
	return floatArithmeticOp(f, "__radd__", v, w, func(v, w float64) float64 { return w + v })
}

func floatRDiv(f *Frame, v, w *Object) (*Object, *BaseException) {
	return floatDivModOp(f, "__rdiv__", v, w, func(v, w float64) (float64, bool) {
		if v == 0.0 {
			return 0, false
		}
		return w / v, true
	})
}

func floatRDivMod(f *Frame, v, w *Object) (*Object, *BaseException) {
	return floatDivAndModOp(f, "__rdivmod__", v, w, func(v, w float64) (float64, float64, bool) {
		m, r := floatModFunc(w, v)
		if !r {
			return 0, 0, false
		}
		return w / v, m, true
	})
}

const (
	floatReprPrecision = 16
	floatStrPrecision  = 12
)

func floatRepr(f *Frame, o *Object) (*Object, *BaseException) {
	return NewStr(floatToString(toFloatUnsafe(o).Value(), floatReprPrecision)).ToObject(), nil
}

func floatRFloorDiv(f *Frame, v, w *Object) (*Object, *BaseException) {
	return floatDivModOp(f, "__rfloordiv__", v, w, func(v, w float64) (float64, bool) {
		if v == 0.0 {
			return 0, false
		}
		return math.Floor(w / v), true
	})
}

func floatRMod(f *Frame, v, w *Object) (*Object, *BaseException) {
	return floatDivModOp(f, "__rmod__", v, w, func(v, w float64) (float64, bool) {
		return floatModFunc(w, v)
	})
}

func floatRMul(f *Frame, v, w *Object) (*Object, *BaseException) {
	return floatArithmeticOp(f, "__rmul__", v, w, func(v, w float64) float64 { return w * v })
}

func floatRPow(f *Frame, v, w *Object) (*Object, *BaseException) {
	return floatArithmeticOp(f, "__rpow", v, w, func(v, w float64) float64 { return math.Pow(w, v) })
}

func floatRSub(f *Frame, v, w *Object) (*Object, *BaseException) {
	return floatArithmeticOp(f, "__rsub__", v, w, func(v, w float64) float64 { return w - v })
}

func floatStr(f *Frame, o *Object) (*Object, *BaseException) {
	return NewStr(floatToString(toFloatUnsafe(o).Value(), floatStrPrecision)).ToObject(), nil
}

func floatSub(f *Frame, v, w *Object) (*Object, *BaseException) {
	return floatArithmeticOp(f, "__sub__", v, w, func(v, w float64) float64 { return v - w })
}

func initFloatType(dict map[string]*Object) {
	dict["__getnewargs__"] = newBuiltinFunction("__getnewargs__", floatGetNewArgs).ToObject()
	FloatType.slots.Abs = &unaryOpSlot{floatAbs}
	FloatType.slots.Add = &binaryOpSlot{floatAdd}
	FloatType.slots.Div = &binaryOpSlot{floatDiv}
	FloatType.slots.DivMod = &binaryOpSlot{floatDivMod}
	FloatType.slots.Eq = &binaryOpSlot{floatEq}
	FloatType.slots.Float = &unaryOpSlot{floatFloat}
	FloatType.slots.FloorDiv = &binaryOpSlot{floatFloorDiv}
	FloatType.slots.GE = &binaryOpSlot{floatGE}
	FloatType.slots.GT = &binaryOpSlot{floatGT}
	FloatType.slots.Hash = &unaryOpSlot{floatHash}
	FloatType.slots.Int = &unaryOpSlot{floatInt}
	FloatType.slots.Long = &unaryOpSlot{floatLong}
	FloatType.slots.LE = &binaryOpSlot{floatLE}
	FloatType.slots.LT = &binaryOpSlot{floatLT}
	FloatType.slots.Mod = &binaryOpSlot{floatMod}
	FloatType.slots.Mul = &binaryOpSlot{floatMul}
	FloatType.slots.Native = &nativeSlot{floatNative}
	FloatType.slots.NE = &binaryOpSlot{floatNE}
	FloatType.slots.Neg = &unaryOpSlot{floatNeg}
	FloatType.slots.New = &newSlot{floatNew}
	FloatType.slots.NonZero = &unaryOpSlot{floatNonZero}
	FloatType.slots.Pos = &unaryOpSlot{floatPos}
	FloatType.slots.Pow = &binaryOpSlot{floatPow}
	FloatType.slots.RAdd = &binaryOpSlot{floatRAdd}
	FloatType.slots.RDiv = &binaryOpSlot{floatRDiv}
	FloatType.slots.RDivMod = &binaryOpSlot{floatRDivMod}
	FloatType.slots.Repr = &unaryOpSlot{floatRepr}
	FloatType.slots.RFloorDiv = &binaryOpSlot{floatRFloorDiv}
	FloatType.slots.RMod = &binaryOpSlot{floatRMod}
	FloatType.slots.RMul = &binaryOpSlot{floatRMul}
	FloatType.slots.RPow = &binaryOpSlot{floatRPow}
	FloatType.slots.RSub = &binaryOpSlot{floatRSub}
	FloatType.slots.Str = &unaryOpSlot{floatStr}
	FloatType.slots.Sub = &binaryOpSlot{floatSub}
}

func floatArithmeticOp(f *Frame, method string, v, w *Object, fun func(v, w float64) float64) (*Object, *BaseException) {
	floatW, ok := floatCoerce(w)
	if !ok {
		if math.IsInf(floatW, 0) {
			return nil, f.RaiseType(OverflowErrorType, "long int too large to convert to float")
		}
		return NotImplemented, nil
	}
	return NewFloat(fun(toFloatUnsafe(v).Value(), floatW)).ToObject(), nil
}

func floatCompare(v *Float, w *Object, ltResult, eqResult, gtResult *Int) *Object {
	lhs := v.Value()
	rhs, ok := floatCoerce(w)
	if !ok {
		if !math.IsInf(rhs, 0) {
			return NotImplemented
		}
		// When floatCoerce returns (Inf, false) it indicates an
		// overflow - abs(rhs) is between MaxFloat64 and Inf.
		// When comparing with infinite floats, rhs might as well be 0.
		// Otherwise, let the compare proceed normally as |rhs| might
		// as well be infinite, since it's outside the range of finite
		// floats.
		if math.IsInf(lhs, 0) {
			rhs = 0
		}
	}
	if lhs < rhs {
		return ltResult.ToObject()
	}
	if lhs == rhs {
		return eqResult.ToObject()
	}
	if lhs > rhs {
		return gtResult.ToObject()
	}
	// There must be a NaN involved, which always compares false, even to other NaNs.
	// This is true both in Go and in Python.
	return False.ToObject()
}

// floatCoerce will coerce any numeric type to a float. If all is
// well, it will return the float64 value, and true (OK). If an overflow
// occurs, it will return either (+Inf, false) or (-Inf, false) depending
// on whether the source value was too large or too small. Note that if the
// source number is an infinite float, the result will be infinite without
// overflow, (+-Inf, true).
// If the input is not a number, it will return (0, false).
func floatCoerce(o *Object) (float64, bool) {
	switch {
	case o.isInstance(IntType):
		return float64(toIntUnsafe(o).Value()), true
	case o.isInstance(LongType):
		f, _ := new(big.Float).SetInt(toLongUnsafe(o).Value()).Float64()
		// If f is infinite, that indicates the big.Int was too large
		// or too small to be represented as a float64. In that case,
		// indicate the overflow by returning (f, false).
		overflow := math.IsInf(f, 0)
		return f, !overflow
	case o.isInstance(FloatType):
		return toFloatUnsafe(o).Value(), true
	default:
		return 0, false
	}
}

func floatConvert(floatSlot *unaryOpSlot, f *Frame, o *Object) (*Float, *BaseException) {
	result, raised := floatSlot.Fn(f, o)
	if raised != nil {
		return nil, raised
	}
	if !result.isInstance(FloatType) {
		exc := fmt.Sprintf("__float__ returned non-float (type %s)", result.typ.Name())
		return nil, f.RaiseType(TypeErrorType, exc)
	}
	return toFloatUnsafe(result), nil
}

func floatDivModOp(f *Frame, method string, v, w *Object, fun func(v, w float64) (float64, bool)) (*Object, *BaseException) {
	floatW, ok := floatCoerce(w)
	if !ok {
		if math.IsInf(floatW, 0) {
			return nil, f.RaiseType(OverflowErrorType, "long int too large to convert to float")
		}
		return NotImplemented, nil
	}
	x, ok := fun(toFloatUnsafe(v).Value(), floatW)
	if !ok {
		return nil, f.RaiseType(ZeroDivisionErrorType, "float division or modulo by zero")
	}
	return NewFloat(x).ToObject(), nil
}

func floatDivAndModOp(f *Frame, method string, v, w *Object, fun func(v, w float64) (float64, float64, bool)) (*Object, *BaseException) {
	floatW, ok := floatCoerce(w)
	if !ok {
		if math.IsInf(floatW, 0) {
			return nil, f.RaiseType(OverflowErrorType, "long int too large to convert to float")
		}
		return NotImplemented, nil
	}
	q, m, ok := fun(toFloatUnsafe(v).Value(), floatW)
	if !ok {
		return nil, f.RaiseType(ZeroDivisionErrorType, "float division or modulo by zero")
	}
	return NewTuple2(NewFloat(q).ToObject(), NewFloat(m).ToObject()).ToObject(), nil
}

func hashFloat(v float64) int {
	if math.IsNaN(v) {
		return 0
	}

	if math.IsInf(v, 0) {
		if math.IsInf(v, 1) {
			return 314159
		}
		if math.IsInf(v, -1) {
			return -271828
		}
		return 0
	}

	_, fracPart := math.Modf(v)
	if fracPart == 0.0 {
		i := big.Int{}
		big.NewFloat(v).Int(&i)
		if numInIntRange(&i) {
			return int(i.Int64())
		}
		// TODO: hashBigInt() is not yet matched that of cpython or pypy.
		return hashBigInt(&i)
	}

	v, expo := math.Frexp(v)
	v *= 2147483648.0
	hiPart := int(v)
	v = (v - float64(hiPart)) * 2147483648.0
	x := int(hiPart + int(v) + (expo << 15))
	return x
}

func floatModFunc(v, w float64) (float64, bool) {
	if w == 0.0 {
		return 0, false
	}
	x := math.Mod(v, w)
	if x != 0 && math.Signbit(x) != math.Signbit(w) {
		// In Python the result of the modulo operator is
		// always the same sign as the divisor, whereas in Go,
		// the result is always the same sign as the dividend.
		// Therefore we need to do an adjustment when the sign
		// of the modulo result differs from that of the
		// divisor.
		x += w
	}
	return x, true
}

func floatToString(f float64, p int) string {
	s := unsignPositiveInf(strings.ToLower(strconv.FormatFloat(f, 'g', p, 64)))
	fun := func(r rune) bool {
		return !unicode.IsDigit(r)
	}
	if i := strings.IndexFunc(s, fun); i == -1 {
		s += ".0"
	}
	return s
}

func unsignPositiveInf(s string) string {
	if s == "+inf" {
		return "inf"
	}
	return s
}

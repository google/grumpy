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
	"errors"
	"fmt"
	"math"
	"math/cmplx"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// ComplexType is the object representing the Python 'complex' type.
var ComplexType = newBasisType("complex", reflect.TypeOf(Complex{}), toComplexUnsafe, ObjectType)

// Complex represents Python 'complex' objects.
type Complex struct {
	Object
	value complex128
}

// NewComplex returns a new Complex holding the given complex value.
func NewComplex(value complex128) *Complex {
	return &Complex{Object{typ: ComplexType}, value}
}

func toComplexUnsafe(o *Object) *Complex {
	return (*Complex)(o.toPointer())
}

// ToObject upcasts c to an Object.
func (c *Complex) ToObject() *Object {
	return &c.Object
}

// Value returns the underlying complex value held by c.
func (c *Complex) Value() complex128 {
	return c.value
}

func complexAbs(f *Frame, o *Object) (*Object, *BaseException) {
	c := toComplexUnsafe(o).Value()
	return NewFloat(cmplx.Abs(c)).ToObject(), nil
}

func complexAdd(f *Frame, v, w *Object) (*Object, *BaseException) {
	return complexArithmeticOp(f, "__add__", v, w, func(lhs, rhs complex128) complex128 {
		return lhs + rhs
	})
}

func complexCompareNotSupported(f *Frame, v, w *Object) (*Object, *BaseException) {
	if w.isInstance(IntType) || w.isInstance(LongType) || w.isInstance(FloatType) || w.isInstance(ComplexType) {
		return nil, f.RaiseType(TypeErrorType, "no ordering relation is defined for complex numbers")
	}
	return NotImplemented, nil
}

func complexComplex(f *Frame, o *Object) (*Object, *BaseException) {
	return o, nil
}

func complexDiv(f *Frame, v, w *Object) (*Object, *BaseException) {
	return complexDivModOp(f, "__div__", v, w, func(v, w complex128) (complex128, bool) {
		if w == 0 {
			return 0, false
		}
		return v / w, true
	})
}

func complexDivMod(f *Frame, v, w *Object) (*Object, *BaseException) {
	return complexDivAndModOp(f, "__divmod__", v, w, func(v, w complex128) (complex128, complex128, bool) {
		if w == 0 {
			return 0, 0, false
		}
		return complexFloorDivOp(v, w), complexModOp(v, w), true
	})
}

func complexEq(f *Frame, v, w *Object) (*Object, *BaseException) {
	e, ok := complexCompare(toComplexUnsafe(v), w)
	if !ok {
		return NotImplemented, nil
	}
	return GetBool(e).ToObject(), nil
}

func complexFloorDiv(f *Frame, v, w *Object) (*Object, *BaseException) {
	return complexDivModOp(f, "__floordiv__", v, w, func(v, w complex128) (complex128, bool) {
		if w == 0 {
			return 0, false
		}
		return complexFloorDivOp(v, w), true
	})
}

func complexHash(f *Frame, o *Object) (*Object, *BaseException) {
	v := toComplexUnsafe(o).Value()
	hashCombined := hashFloat(real(v)) + 1000003*hashFloat(imag(v))
	if hashCombined == -1 {
		hashCombined = -2
	}
	return NewInt(hashCombined).ToObject(), nil
}

func complexMod(f *Frame, v, w *Object) (*Object, *BaseException) {
	return complexDivModOp(f, "__mod__", v, w, func(v, w complex128) (complex128, bool) {
		if w == 0 {
			return 0, false
		}
		return complexModOp(v, w), true
	})
}

func complexMul(f *Frame, v, w *Object) (*Object, *BaseException) {
	return complexArithmeticOp(f, "__mul__", v, w, func(lhs, rhs complex128) complex128 {
		return lhs * rhs
	})
}

func complexNE(f *Frame, v, w *Object) (*Object, *BaseException) {
	e, ok := complexCompare(toComplexUnsafe(v), w)
	if !ok {
		return NotImplemented, nil
	}
	return GetBool(!e).ToObject(), nil
}

func complexNeg(f *Frame, o *Object) (*Object, *BaseException) {
	c := toComplexUnsafe(o).Value()
	return NewComplex(-c).ToObject(), nil
}

func complexNew(f *Frame, t *Type, args Args, _ KWArgs) (*Object, *BaseException) {
	argc := len(args)
	if argc == 0 {
		return newObject(t), nil
	}
	if argc > 2 {
		return nil, f.RaiseType(TypeErrorType, "'__new__' of 'complex' requires at most 2 arguments")
	}
	if t != ComplexType {
		// Allocate a plain complex then copy it's value into an object
		// of the complex subtype.
		x, raised := complexNew(f, ComplexType, args, nil)
		if raised != nil {
			return nil, raised
		}
		result := toComplexUnsafe(newObject(t))
		result.value = toComplexUnsafe(x).Value()
		return result.ToObject(), nil
	}
	if complexSlot := args[0].typ.slots.Complex; complexSlot != nil && argc == 1 {
		c, raised := complexConvert(complexSlot, f, args[0])
		if raised != nil {
			return nil, raised
		}
		return c.ToObject(), nil
	}
	if args[0].isInstance(StrType) {
		if argc > 1 {
			return nil, f.RaiseType(TypeErrorType, "complex() can't take second arg if first is a string")
		}
		s := toStrUnsafe(args[0]).Value()
		result, err := parseComplex(s)
		if err != nil {
			return nil, f.RaiseType(ValueErrorType, "complex() arg is a malformed string")
		}
		return NewComplex(result).ToObject(), nil
	}
	if argc > 1 && args[1].isInstance(StrType) {
		return nil, f.RaiseType(TypeErrorType, "complex() second arg can't be a string")
	}
	cr, raised := complex128Convert(f, args[0])
	if raised != nil {
		return nil, raised
	}
	var ci complex128
	if argc > 1 {
		ci, raised = complex128Convert(f, args[1])
		if raised != nil {
			return nil, raised
		}
	}

	// Logically it should be enough to return this:
	//  NewComplex(cr + ci*1i).ToObject()
	// But Go complex arithmatic is not satisfying all conditions, for instance:
	//  cr := complex(math.Inf(1), 0)
	//  ci := complex(math.Inf(-1), 0)
	//  fmt.Println(cr + ci*1i)
	// Output is (NaN-Infi), instead of (+Inf-Infi).
	return NewComplex(complex(real(cr)-imag(ci), imag(cr)+real(ci))).ToObject(), nil
}

func complexNonZero(f *Frame, o *Object) (*Object, *BaseException) {
	return GetBool(toComplexUnsafe(o).Value() != 0).ToObject(), nil
}

func complexPos(f *Frame, o *Object) (*Object, *BaseException) {
	return o, nil
}

func complexPow(f *Frame, v, w *Object) (*Object, *BaseException) {
	return complexArithmeticOp(f, "__pow__", v, w, func(lhs, rhs complex128) complex128 {
		return cmplx.Pow(lhs, rhs)
	})
}

func complexRAdd(f *Frame, v, w *Object) (*Object, *BaseException) {
	return complexArithmeticOp(f, "__radd__", v, w, func(lhs, rhs complex128) complex128 {
		return lhs + rhs
	})
}

func complexRDiv(f *Frame, v, w *Object) (*Object, *BaseException) {
	return complexDivModOp(f, "__rdiv__", v, w, func(v, w complex128) (complex128, bool) {
		if v == 0 {
			return 0, false
		}
		return w / v, true
	})
}

func complexRDivMod(f *Frame, v, w *Object) (*Object, *BaseException) {
	return complexDivAndModOp(f, "__rdivmod__", v, w, func(v, w complex128) (complex128, complex128, bool) {
		if v == 0 {
			return 0, 0, false
		}
		return complexFloorDivOp(w, v), complexModOp(w, v), true
	})
}

func complexRepr(f *Frame, o *Object) (*Object, *BaseException) {
	c := toComplexUnsafe(o).Value()
	rs, is := "", ""
	pre, post := "", ""
	sign := ""
	if real(c) == 0.0 {
		is = strconv.FormatFloat(imag(c), 'g', -1, 64)
	} else {
		pre = "("
		rs = strconv.FormatFloat(real(c), 'g', -1, 64)
		is = strconv.FormatFloat(imag(c), 'g', -1, 64)
		if imag(c) >= 0.0 || math.IsNaN(imag(c)) {
			sign = "+"
		}
		post = ")"
	}
	rs = unsignPositiveInf(strings.ToLower(rs))
	is = unsignPositiveInf(strings.ToLower(is))
	return NewStr(fmt.Sprintf("%s%s%s%sj%s", pre, rs, sign, is, post)).ToObject(), nil
}

func complexRFloorDiv(f *Frame, v, w *Object) (*Object, *BaseException) {
	return complexDivModOp(f, "__rfloordiv__", v, w, func(v, w complex128) (complex128, bool) {
		if v == 0 {
			return 0, false
		}
		return complexFloorDivOp(w, v), true
	})
}

func complexRMod(f *Frame, v, w *Object) (*Object, *BaseException) {
	return complexDivModOp(f, "__rmod__", v, w, func(v, w complex128) (complex128, bool) {
		if v == 0 {
			return 0, false
		}
		return complexModOp(w, v), true
	})
}

func complexRMul(f *Frame, v, w *Object) (*Object, *BaseException) {
	return complexArithmeticOp(f, "__rmul__", v, w, func(lhs, rhs complex128) complex128 {
		return rhs * lhs
	})
}

func complexRPow(f *Frame, v, w *Object) (*Object, *BaseException) {
	return complexArithmeticOp(f, "__rpow__", v, w, func(lhs, rhs complex128) complex128 {
		return cmplx.Pow(rhs, lhs)
	})
}

func complexRSub(f *Frame, v, w *Object) (*Object, *BaseException) {
	return complexArithmeticOp(f, "__rsub__", v, w, func(lhs, rhs complex128) complex128 {
		return rhs - lhs
	})
}

func complexSub(f *Frame, v, w *Object) (*Object, *BaseException) {
	return complexArithmeticOp(f, "__sub__", v, w, func(lhs, rhs complex128) complex128 {
		return lhs - rhs
	})
}

func initComplexType(dict map[string]*Object) {
	ComplexType.slots.Abs = &unaryOpSlot{complexAbs}
	ComplexType.slots.Add = &binaryOpSlot{complexAdd}
	ComplexType.slots.Complex = &unaryOpSlot{complexComplex}
	ComplexType.slots.Div = &binaryOpSlot{complexDiv}
	ComplexType.slots.DivMod = &binaryOpSlot{complexDivMod}
	ComplexType.slots.Eq = &binaryOpSlot{complexEq}
	ComplexType.slots.FloorDiv = &binaryOpSlot{complexFloorDiv}
	ComplexType.slots.GE = &binaryOpSlot{complexCompareNotSupported}
	ComplexType.slots.GT = &binaryOpSlot{complexCompareNotSupported}
	ComplexType.slots.Hash = &unaryOpSlot{complexHash}
	ComplexType.slots.LE = &binaryOpSlot{complexCompareNotSupported}
	ComplexType.slots.LT = &binaryOpSlot{complexCompareNotSupported}
	ComplexType.slots.Mod = &binaryOpSlot{complexMod}
	ComplexType.slots.Mul = &binaryOpSlot{complexMul}
	ComplexType.slots.NE = &binaryOpSlot{complexNE}
	ComplexType.slots.Neg = &unaryOpSlot{complexNeg}
	ComplexType.slots.New = &newSlot{complexNew}
	ComplexType.slots.NonZero = &unaryOpSlot{complexNonZero}
	ComplexType.slots.Pos = &unaryOpSlot{complexPos}
	ComplexType.slots.Pow = &binaryOpSlot{complexPow}
	ComplexType.slots.RAdd = &binaryOpSlot{complexRAdd}
	ComplexType.slots.RDiv = &binaryOpSlot{complexRDiv}
	ComplexType.slots.RDivMod = &binaryOpSlot{complexRDivMod}
	ComplexType.slots.RFloorDiv = &binaryOpSlot{complexRFloorDiv}
	ComplexType.slots.Repr = &unaryOpSlot{complexRepr}
	ComplexType.slots.RMod = &binaryOpSlot{complexRMod}
	ComplexType.slots.RMul = &binaryOpSlot{complexRMul}
	ComplexType.slots.RPow = &binaryOpSlot{complexRPow}
	ComplexType.slots.RSub = &binaryOpSlot{complexRSub}
	ComplexType.slots.Sub = &binaryOpSlot{complexSub}
}

func complex128Convert(f *Frame, o *Object) (complex128, *BaseException) {
	if complexSlot := o.typ.slots.Complex; complexSlot != nil {
		c, raised := complexConvert(complexSlot, f, o)
		if raised != nil {
			return complex(0, 0), raised
		}
		return c.Value(), nil
	} else if floatSlot := o.typ.slots.Float; floatSlot != nil {
		result, raised := floatConvert(floatSlot, f, o)
		if raised != nil {
			return complex(0, 0), raised
		}
		return complex(result.Value(), 0), nil
	} else {
		return complex(0, 0), f.RaiseType(TypeErrorType, "complex() argument must be a string or a number")
	}
}

func complexArithmeticOp(f *Frame, method string, v, w *Object, fun func(v, w complex128) complex128) (*Object, *BaseException) {
	if w.isInstance(ComplexType) {
		return NewComplex(fun(toComplexUnsafe(v).Value(), toComplexUnsafe(w).Value())).ToObject(), nil
	}

	floatW, ok := floatCoerce(w)
	if !ok {
		if math.IsInf(floatW, 0) {
			return nil, f.RaiseType(OverflowErrorType, "long int too large to convert to float")
		}
		return NotImplemented, nil
	}
	return NewComplex(fun(toComplexUnsafe(v).Value(), complex(floatW, 0))).ToObject(), nil
}

// complexCoerce will coerce any numeric type to a complex. If all is
// well, it will return the complex128 value, and true (OK). If an overflow
// occurs, it will return either (+Inf, false) or (-Inf, false) depending
// on whether the source value was too large or too small. Note that if the
// source number is an infinite float, the result will be infinite without
// overflow, (+-Inf, true).
// If the input is not a number, it will return (0, false).
func complexCoerce(o *Object) (complex128, bool) {
	if o.isInstance(ComplexType) {
		return toComplexUnsafe(o).Value(), true
	}
	floatO, ok := floatCoerce(o)
	if !ok {
		if math.IsInf(floatO, 0) {
			return complex(floatO, 0.0), false
		}
		return 0, false
	}
	return complex(floatO, 0.0), true
}

func complexCompare(v *Complex, w *Object) (bool, bool) {
	lhsr := real(v.Value())
	rhs, ok := complexCoerce(w)
	if !ok {
		return false, false
	}
	return lhsr == real(rhs) && imag(v.Value()) == imag(rhs), true
}

func complexConvert(complexSlot *unaryOpSlot, f *Frame, o *Object) (*Complex, *BaseException) {
	result, raised := complexSlot.Fn(f, o)
	if raised != nil {
		return nil, raised
	}
	if !result.isInstance(ComplexType) {
		exc := fmt.Sprintf("__complex__ returned non-complex (type %s)", result.typ.Name())
		return nil, f.RaiseType(TypeErrorType, exc)
	}
	return toComplexUnsafe(result), nil
}

func complexDivModOp(f *Frame, method string, v, w *Object, fun func(v, w complex128) (complex128, bool)) (*Object, *BaseException) {
	complexW, ok := complexCoerce(w)
	if !ok {
		if cmplx.IsInf(complexW) {
			return nil, f.RaiseType(OverflowErrorType, "long int too large to convert to complex")
		}
		return NotImplemented, nil
	}
	x, ok := fun(toComplexUnsafe(v).Value(), complexW)
	if !ok {
		return nil, f.RaiseType(ZeroDivisionErrorType, "complex division or modulo by zero")
	}
	return NewComplex(x).ToObject(), nil
}

func complexDivAndModOp(f *Frame, method string, v, w *Object, fun func(v, w complex128) (complex128, complex128, bool)) (*Object, *BaseException) {
	complexW, ok := complexCoerce(w)
	if !ok {
		if cmplx.IsInf(complexW) {
			return nil, f.RaiseType(OverflowErrorType, "long int too large to convert to complex")
		}
		return NotImplemented, nil
	}
	q, m, ok := fun(toComplexUnsafe(v).Value(), complexW)
	if !ok {
		return nil, f.RaiseType(ZeroDivisionErrorType, "complex division or modulo by zero")
	}
	return NewTuple2(NewComplex(q).ToObject(), NewComplex(m).ToObject()).ToObject(), nil
}

func complexFloorDivOp(v, w complex128) complex128 {
	return complex(math.Floor(real(v/w)), 0)
}

func complexModOp(v, w complex128) complex128 {
	return v - complexFloorDivOp(v, w)*w
}

const (
	blank = iota
	real1
	imag1
	real2
	sign2
	imag3
	real4
	sign5
	onlyJ
)

// ParseComplex converts the string s to a complex number.
// If string is well-formed (one of these forms: <float>, <float>j,
// <float><signed-float>j, <float><sign>j, <sign>j or j, where <float> is
// any numeric string that's acceptable by strconv.ParseFloat(s, 64)),
// ParseComplex returns the respective complex128 number.
func parseComplex(s string) (complex128, error) {
	c := strings.Count(s, "(")
	if (c > 1) || (c == 1 && strings.Count(s, ")") != 1) {
		return complex(0, 0), errors.New("Malformed complex string, more than one matching parantheses")
	}
	ts := strings.TrimSpace(s)
	ts = strings.Trim(ts, "()")
	ts = strings.TrimSpace(ts)
	re := `(?i)(?:(?:(?:(?:\d*\.\d+)|(?:\d+\.?))(?:[Ee][+-]?\d+)?)|(?:infinity)|(?:nan)|(?:inf))`
	fre := `[-+]?` + re
	sre := `[-+]` + re
	fsfj := `(?:(?P<real1>` + fre + `)(?P<imag1>` + sre + `)j)`
	fsj := `(?:(?P<real2>` + fre + `)(?P<sign2>[-+])j)`
	fj := `(?P<imag3>` + fre + `)j`
	f := `(?P<real4>` + fre + `)`
	sj := `(?P<sign5>[-+])j`
	j := `(?P<onlyJ>j)`
	r := regexp.MustCompile(`^(?:` + fsfj + `|` + fsj + `|` + fj + `|` + f + `|` + sj + `|` + j + `)$`)
	subs := r.FindStringSubmatch(ts)
	if subs == nil {
		return complex(0, 0), errors.New("Malformed complex string, no mathing pattern found")
	}
	if subs[real1] != "" && subs[imag1] != "" {
		r, _ := strconv.ParseFloat(unsignNaN(subs[real1]), 64)
		i, err := strconv.ParseFloat(unsignNaN(subs[imag1]), 64)
		return complex(r, i), err
	}
	if subs[real2] != "" && subs[sign2] != "" {
		r, err := strconv.ParseFloat(unsignNaN(subs[real2]), 64)
		if subs[sign2] == "-" {
			return complex(r, -1), err
		}
		return complex(r, 1), err
	}
	if subs[imag3] != "" {
		i, err := strconv.ParseFloat(unsignNaN(subs[imag3]), 64)
		return complex(0, i), err
	}
	if subs[real4] != "" {
		r, err := strconv.ParseFloat(unsignNaN(subs[real4]), 64)
		return complex(r, 0), err
	}
	if subs[sign5] != "" {
		if subs[sign5] == "-" {
			return complex(0, -1), nil
		}
		return complex(0, 1), nil
	}
	if subs[onlyJ] != "" {
		return complex(0, 1), nil
	}
	return complex(0, 0), errors.New("Malformed complex string")
}

func unsignNaN(s string) string {
	ls := strings.ToLower(s)
	if ls == "-nan" || ls == "+nan" {
		return "nan"
	}
	return s
}

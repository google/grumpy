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
	"reflect"
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
		if imag(c) >= 0.0 {
			sign = "+"
		}
		post = ")"
	}
	return NewStr(fmt.Sprintf("%s%s%s%sj%s", pre, rs, sign, is, post)).ToObject(), nil
}

func initComplexType(dict map[string]*Object) {
	ComplexType.slots.Eq = &binaryOpSlot{complexEq}
	ComplexType.slots.GE = &binaryOpSlot{complexGE}
	ComplexType.slots.GT = &binaryOpSlot{complexGT}
	ComplexType.slots.LE = &binaryOpSlot{complexLE}
	ComplexType.slots.LT = &binaryOpSlot{complexLT}
	ComplexType.slots.NE = &binaryOpSlot{complexNE}
	ComplexType.slots.New = &newSlot{complexNew}
	ComplexType.slots.Repr = &unaryOpSlot{complexRepr}
}

func complexEq(f *Frame, v, w *Object) (*Object, *BaseException) {
	if !v.isInstance(ComplexType) {
		format := "__eq__ received non-complex (type %s)"
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, v.typ.Name()))
	}
	return complexCompare(toComplexUnsafe(v), w, True, False), nil
}

func complexGE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return nil, f.RaiseType(TypeErrorType, "'__ge__' of 'complex' is not defined")
}

func complexGT(f *Frame, v, w *Object) (*Object, *BaseException) {
	return nil, f.RaiseType(TypeErrorType, "'__gt__' of 'complex' is not defined")
}

func complexLE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return nil, f.RaiseType(TypeErrorType, "'__le__' of 'complex' is not defined")
}

func complexLT(f *Frame, v, w *Object) (*Object, *BaseException) {
	return nil, f.RaiseType(TypeErrorType, "'__lt__' of 'complex' is not defined")
}

func complexNE(f *Frame, v, w *Object) (*Object, *BaseException) {
	if !v.isInstance(ComplexType) {
		format := "__ne__ received non-complex (type %s)"
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, v.typ.Name()))
	}
	return complexCompare(toComplexUnsafe(v), w, False, True), nil
}

func complexCompare(v *Complex, w *Object, eqResult *Int, neResult *Int) *Object {
	lhsr := real(v.Value())
	rhs, ok := complexCoerce(w)
	if !ok {
		return NotImplemented
	}
	if lhsr == real(rhs) && imag(v.Value()) == imag(rhs) {
		return eqResult.ToObject()
	}
	return neResult.ToObject()
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
	o := args[0]
	if argc == 1 {
		if complexSlot := o.typ.slots.Complex; complexSlot != nil {
			result, raised := complexSlot.Fn(f, o)
			if raised != nil {
				return nil, raised
			}
			if !result.isInstance(ComplexType) {
				format := "__complex__ returned non-complex (type %s)"
				return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, result.typ.Name()))
			}
			return result, nil
		}
		if floatSlot := o.typ.slots.Float; floatSlot != nil {
			result, raised := floatSlot.Fn(f, o)
			if raised != nil {
				return nil, raised
			}
			if !result.isInstance(FloatType) {
				exc := fmt.Sprintf("__float__ returned non-float (type %s)", result.typ.Name())
				return nil, f.RaiseType(TypeErrorType, exc)
			}
			f := toFloatUnsafe(result).Value()
			return NewComplex(complex(f, 0)).ToObject(), nil
		}
		if !o.isInstance(StrType) {
			return nil, f.RaiseType(TypeErrorType, "complex() argument must be a string or a number")
		}
		s := toStrUnsafe(o).Value()
		result, err := parseComplex(s)
		if err != nil {
			return nil, f.RaiseType(ValueErrorType, fmt.Sprintf("could not convert string to complex: %s", s))
		}
		return NewComplex(result).ToObject(), nil
	}

	// TODO: fix me
	return nil, nil
}

// ParseComplex converts the string s to a complex number.
// If string is well-formed (one of these forms: <float>, <float>j,
// <float><signed-float>j, <float><sign>j, <sign>j or j, where <float> is
// any numeric string that's acceptable by strconv.ParseFloat(s, 64)),
// ParseComplex returns the respective complex128 number.
func parseComplex(s string) (complex128, error) {
	ts := strings.Trim(s, "() ")
	if !strings.Contains(ts, "j") {
		result, err := strconv.ParseFloat(ts, 64)
		return complex(result, 0), err
	}
	ts = strings.Replace(ts, "j", "", -1)
	if len(ts) == 0 {
		return complex(0, 1), nil
	}
	a := splitBeforeSign(ts)
	l := len(a)
	if (l == 3 && a[0] == "") || (l == 2 && a[0] != "") {
		r, _ := strconv.ParseFloat(a[l-2], 64)
		if a[l-1] == "+" || a[l-1] == "-" {
			a[l-1] += "1"
		}
		i, err := strconv.ParseFloat(a[l-1], 64)
		return complex(r, i), err
	}
	if (l == 2 && a[0] == "") || (l == 1 && a[0] != "") {
		if a[l-1] == "+" || a[l-1] == "-" {
			a[l-1] += "1"
		}
		i, err := strconv.ParseFloat(a[l-1], 64)
		return complex(0, i), err
	}
	return complex(0, 0), nil
}

// SplitBeforeSign splits the string s before sign (+ or -) delimiter,
// and keeps the delimiter too.
func splitBeforeSign(s string) []string {
	a := []string{}
	if len(s) == 0 {
		return a
	}
	c := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '+' || s[i] == '-' {
			a = append(a, s[:i])
			s = s[i:]
			c++
		}
	}
	a = append(a, s)
	return a
}

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
	ComplexType.slots.GE = &binaryOpSlot{complexCompareNotSupported}
	ComplexType.slots.GT = &binaryOpSlot{complexCompareNotSupported}
	ComplexType.slots.LE = &binaryOpSlot{complexCompareNotSupported}
	ComplexType.slots.LT = &binaryOpSlot{complexCompareNotSupported}
	ComplexType.slots.NE = &binaryOpSlot{complexNE}
	ComplexType.slots.New = &newSlot{complexNew}
	ComplexType.slots.Repr = &unaryOpSlot{complexRepr}
}

func complexEq(f *Frame, v, w *Object) (*Object, *BaseException) {
	e, ok := complexCompare(toComplexUnsafe(v), w)
	if !ok {
		return NotImplemented, nil
	}
	return GetBool(e).ToObject(), nil
}

func complexNE(f *Frame, v, w *Object) (*Object, *BaseException) {
	e, ok := complexCompare(toComplexUnsafe(v), w)
	if !ok {
		return NotImplemented, nil
	}
	return GetBool(!e).ToObject(), nil
}

func complexCompare(v *Complex, w *Object) (bool, bool) {
	lhsr := real(v.Value())
	rhs, ok := complexCoerce(w)
	if !ok {
		return false, false
	}
	return lhsr == real(rhs) && imag(v.Value()) == imag(rhs), true
}

func complexCompareNotSupported(f *Frame, v, w *Object) (*Object, *BaseException) {
	if w.isInstance(IntType) || w.isInstance(LongType) || w.isInstance(FloatType) || w.isInstance(ComplexType) {
		return nil, f.RaiseType(TypeErrorType, "no ordering relation is defined for complex numbers")
	}
	return NotImplemented, nil
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
			result, raised := floatConvert(floatSlot, f, o)
			if raised != nil {
				return nil, raised
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
			return nil, f.RaiseType(ValueErrorType, "complex() arg is a malformed string")
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
	c := strings.Count(s, "(")
	if (c > 1) || (c == 1 && strings.Count(s, ")") != 1) {
		return complex(0, 0), errors.New("Malformed complex string, more than one matching parantheses")
	}
	ts := strings.Trim(s, "() ")
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
	res := make(map[string]string)
	for i, name := range r.SubexpNames() {
		// First one is the complete string
		if i != 0 {
			res[name] = subs[i]
		}
	}
	if res["real1"] != "" && res["imag1"] != "" {
		r, _ := strconv.ParseFloat(unsignNaN(res["real1"]), 64)
		i, err := strconv.ParseFloat(unsignNaN(res["imag1"]), 64)
		return complex(r, i), err
	}
	if res["real2"] != "" && res["sign2"] != "" {
		r, err := strconv.ParseFloat(unsignNaN(res["real2"]), 64)
		if res["sign2"] == "-" {
			return complex(r, -1), err
		}
		return complex(r, 1), err
	}
	if res["imag3"] != "" {
		i, err := strconv.ParseFloat(unsignNaN(res["imag3"]), 64)
		return complex(0, i), err
	}
	if res["real4"] != "" {
		r, err := strconv.ParseFloat(unsignNaN(res["real4"]), 64)
		return complex(r, 0), err
	}
	if res["sign5"] != "" {
		if res["sign5"] == "-" {
			return complex(0, -1), nil
		}
		return complex(0, 1), nil
	}
	if res["onlyJ"] != "" {
		return complex(0, 1), nil
	}
	return complex(0, 0), errors.New("Malformed complex string")
}

func unsignNaN(s string) string {
	us := strings.ToUpper(s)
	if us == "-NAN" || us == "+NAN" {
		return "NAN"
	}
	return s
}

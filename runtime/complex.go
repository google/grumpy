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

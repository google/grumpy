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
	"math"
	"math/big"
	"math/cmplx"
	"testing"
)

func TestComplexAbs(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(complex(0, 0)), want: NewFloat(0).ToObject()},
		{args: wrapArgs(complex(1, 1)), want: NewFloat(1.4142135623730951).ToObject()},
		{args: wrapArgs(complex(1, 2)), want: NewFloat(2.23606797749979).ToObject()},
		{args: wrapArgs(complex(3, 4)), want: NewFloat(5).ToObject()},
		{args: wrapArgs(complex(-3, 4)), want: NewFloat(5).ToObject()},
		{args: wrapArgs(complex(3, -4)), want: NewFloat(5).ToObject()},
		{args: wrapArgs(-complex(3, 4)), want: NewFloat(5).ToObject()},
		{args: wrapArgs(complex(0.123456e-3, 0)), want: NewFloat(0.000123456).ToObject()},
		{args: wrapArgs(complex(0.123456e-3, 3.14151692e+7)), want: NewFloat(31415169.2).ToObject()},
		{args: wrapArgs(complex(math.Inf(-1), 1.2)), want: NewFloat(math.Inf(1)).ToObject()},
		{args: wrapArgs(complex(3.4, math.Inf(1))), want: NewFloat(math.Inf(1)).ToObject()},
		{args: wrapArgs(complex(math.Inf(1), math.Inf(-1))), want: NewFloat(math.Inf(1)).ToObject()},
		{args: wrapArgs(complex(math.Inf(1), math.NaN())), want: NewFloat(math.Inf(1)).ToObject()},
		{args: wrapArgs(complex(math.NaN(), math.Inf(1))), want: NewFloat(math.Inf(1)).ToObject()},
		{args: wrapArgs(complex(math.NaN(), 5.6)), want: NewFloat(math.NaN()).ToObject()},
		{args: wrapArgs(complex(7.8, math.NaN())), want: NewFloat(math.NaN()).ToObject()},
	}
	for _, cas := range cases {
		switch got, match := checkInvokeResult(wrapFuncForTest(complexAbs), cas.args, cas.want, cas.wantExc); match {
		case checkInvokeResultReturnValueMismatch:
			if got == nil || cas.want == nil || !got.isInstance(FloatType) || !cas.want.isInstance(FloatType) ||
				!floatsAreSame(toFloatUnsafe(got).Value(), toFloatUnsafe(cas.want).Value()) {
				t.Errorf("complex.__abs__%v = %v, want %v", cas.args, got, cas.want)
			}
		}
	}
}

func TestComplexEq(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(complex(0, 0), 0), want: True.ToObject()},
		{args: wrapArgs(complex(1, 0), 0), want: False.ToObject()},
		{args: wrapArgs(complex(-12, 0), -12), want: True.ToObject()},
		{args: wrapArgs(complex(-12, 0), 1), want: False.ToObject()},
		{args: wrapArgs(complex(17.20, 0), 17.20), want: True.ToObject()},
		{args: wrapArgs(complex(1.2, 0), 17.20), want: False.ToObject()},
		{args: wrapArgs(complex(-4, 15), complex(-4, 15)), want: True.ToObject()},
		{args: wrapArgs(complex(-4, 15), complex(1, 2)), want: False.ToObject()},
		{args: wrapArgs(complex(math.Inf(1), 0), complex(math.Inf(1), 0)), want: True.ToObject()},
		{args: wrapArgs(complex(math.Inf(1), 0), complex(0, math.Inf(1))), want: False.ToObject()},
		{args: wrapArgs(complex(math.Inf(-1), 0), complex(math.Inf(-1), 0)), want: True.ToObject()},
		{args: wrapArgs(complex(math.Inf(-1), 0), complex(0, math.Inf(-1))), want: False.ToObject()},
		{args: wrapArgs(complex(math.Inf(1), math.Inf(1)), complex(math.Inf(1), math.Inf(1))), want: True.ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(complexEq), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestComplexBinaryOps(t *testing.T) {
	cases := []struct {
		fun     func(f *Frame, v, w *Object) (*Object, *BaseException)
		v, w    *Object
		want    *Object
		wantExc *BaseException
	}{
		{Add, NewComplex(1 + 3i).ToObject(), NewInt(1).ToObject(), NewComplex(2 + 3i).ToObject(), nil},
		{Add, NewComplex(1 + 3i).ToObject(), NewFloat(-1).ToObject(), NewComplex(3i).ToObject(), nil},
		{Add, NewComplex(1 + 3i).ToObject(), NewInt(1).ToObject(), NewComplex(2 + 3i).ToObject(), nil},
		{Add, NewComplex(1 + 3i).ToObject(), NewComplex(-1 - 3i).ToObject(), NewComplex(0i).ToObject(), nil},
		{Add, NewFloat(math.Inf(1)).ToObject(), NewComplex(3i).ToObject(), NewComplex(complex(math.Inf(1), 3)).ToObject(), nil},
		{Add, NewFloat(math.Inf(-1)).ToObject(), NewComplex(3i).ToObject(), NewComplex(complex(math.Inf(-1), 3)).ToObject(), nil},
		{Add, NewFloat(math.NaN()).ToObject(), NewComplex(3i).ToObject(), NewComplex(complex(math.NaN(), 3)).ToObject(), nil},
		{Add, NewComplex(cmplx.NaN()).ToObject(), NewComplex(3i).ToObject(), NewComplex(cmplx.NaN()).ToObject(), nil},
		{Add, NewFloat(math.Inf(-1)).ToObject(), NewComplex(complex(math.Inf(+1), 3)).ToObject(), NewComplex(complex(math.NaN(), 3)).ToObject(), nil},
		{Add, NewComplex(1 + 3i).ToObject(), None, nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for +: 'complex' and 'NoneType'")},
		{Add, None, NewComplex(1 + 3i).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for +: 'NoneType' and 'complex'")},
		{Add, NewInt(3).ToObject(), NewComplex(3i).ToObject(), NewComplex(3 + 3i).ToObject(), nil},
		{Add, NewLong(big.NewInt(9999999)).ToObject(), NewComplex(3i).ToObject(), NewComplex(9999999 + 3i).ToObject(), nil},
		{Add, NewFloat(3.5).ToObject(), NewComplex(3i).ToObject(), NewComplex(3.5 + 3i).ToObject(), nil},
		{Sub, NewComplex(1 + 3i).ToObject(), NewComplex(1 + 3i).ToObject(), NewComplex(0i).ToObject(), nil},
		{Sub, NewComplex(1 + 3i).ToObject(), NewComplex(3i).ToObject(), NewComplex(1).ToObject(), nil},
		{Sub, NewComplex(1 + 3i).ToObject(), NewFloat(1).ToObject(), NewComplex(3i).ToObject(), nil},
		{Sub, NewComplex(3i).ToObject(), NewFloat(1.2).ToObject(), NewComplex(-1.2 + 3i).ToObject(), nil},
		{Sub, NewComplex(1 + 3i).ToObject(), NewComplex(1 + 3i).ToObject(), NewComplex(0i).ToObject(), nil},
		{Sub, NewComplex(4 + 3i).ToObject(), NewInt(1).ToObject(), NewComplex(3 + 3i).ToObject(), nil},
		{Sub, NewComplex(4 + 3i).ToObject(), NewLong(big.NewInt(99994)).ToObject(), NewComplex(-99990 + 3i).ToObject(), nil},
		{Sub, NewFloat(math.Inf(1)).ToObject(), NewComplex(3i).ToObject(), NewComplex(complex(math.Inf(1), -3)).ToObject(), nil},
		{Sub, NewFloat(math.Inf(-1)).ToObject(), NewComplex(3i).ToObject(), NewComplex(complex(math.Inf(-1), -3)).ToObject(), nil},
		{Sub, NewComplex(1 + 3i).ToObject(), None, nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for -: 'complex' and 'NoneType'")},
		{Sub, None, NewComplex(1 + 3i).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for -: 'NoneType' and 'complex'")},
		{Sub, NewFloat(math.NaN()).ToObject(), NewComplex(3i).ToObject(), NewComplex(complex(math.NaN(), -3)).ToObject(), nil},
		{Sub, NewComplex(cmplx.NaN()).ToObject(), NewComplex(3i).ToObject(), NewComplex(cmplx.NaN()).ToObject(), nil},
		{Sub, NewFloat(math.Inf(-1)).ToObject(), NewComplex(complex(math.Inf(-1), 3)).ToObject(), NewComplex(complex(math.NaN(), -3)).ToObject(), nil},
		{Mul, NewComplex(1 + 3i).ToObject(), NewComplex(1 + 3i).ToObject(), NewComplex(-8 + 6i).ToObject(), nil},
		{Mul, NewComplex(1 + 3i).ToObject(), NewComplex(3i).ToObject(), NewComplex(-9 + 3i).ToObject(), nil},
		{Mul, NewComplex(1 + 3i).ToObject(), NewFloat(1).ToObject(), NewComplex(1 + 3i).ToObject(), nil},
		{Mul, NewComplex(3i).ToObject(), NewFloat(1.2).ToObject(), NewComplex(3.5999999999999996i).ToObject(), nil},
		{Mul, NewComplex(1 + 3i).ToObject(), NewComplex(1 + 3i).ToObject(), NewComplex(-8 + 6i).ToObject(), nil},
		{Mul, NewComplex(4 + 3i).ToObject(), NewInt(1).ToObject(), NewComplex(4 + 3i).ToObject(), nil},
		{Mul, NewComplex(4 + 3i).ToObject(), NewLong(big.NewInt(99994)).ToObject(), NewComplex(399976 + 299982i).ToObject(), nil},
		{Mul, NewFloat(math.Inf(1)).ToObject(), NewComplex(3i).ToObject(), NewComplex(complex(math.NaN(), math.Inf(1))).ToObject(), nil},
		{Mul, NewFloat(math.Inf(-1)).ToObject(), NewComplex(3i).ToObject(), NewComplex(complex(math.NaN(), math.Inf(-1))).ToObject(), nil},
		{Mul, NewComplex(1 + 3i).ToObject(), None, nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for *: 'complex' and 'NoneType'")},
		{Mul, None, NewComplex(1 + 3i).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for *: 'NoneType' and 'complex'")},
		{Mul, NewFloat(math.NaN()).ToObject(), NewComplex(3i).ToObject(), NewComplex(complex(math.NaN(), math.NaN())).ToObject(), nil},
		{Mul, NewComplex(cmplx.NaN()).ToObject(), NewComplex(3i).ToObject(), NewComplex(cmplx.NaN()).ToObject(), nil},
		{Mul, NewFloat(math.Inf(-1)).ToObject(), NewComplex(complex(math.Inf(-1), 3)).ToObject(), NewComplex(complex(math.Inf(1), math.NaN())).ToObject(), nil},
		{Pow, NewComplex(0i).ToObject(), NewComplex(0i).ToObject(), NewComplex(1 + 0i).ToObject(), nil},
		{Pow, NewComplex(-1 + 0i).ToObject(), NewComplex(1i).ToObject(), NewComplex(0.04321391826377226 + 0i).ToObject(), nil},
		{Pow, NewComplex(1 + 2i).ToObject(), NewComplex(1 + 2i).ToObject(), NewComplex(-0.22251715680177264 + 0.10070913113607538i).ToObject(), nil},
		{Pow, NewComplex(0i).ToObject(), NewComplex(-1 + 0i).ToObject(), NewComplex(complex(math.Inf(1), 0)).ToObject(), nil},
		{Pow, NewComplex(0i).ToObject(), NewComplex(-1 + 1i).ToObject(), NewComplex(complex(math.Inf(1), math.Inf(1))).ToObject(), nil},
		{Pow, NewComplex(complex(math.Inf(-1), 2)).ToObject(), NewComplex(1 + 2i).ToObject(), NewComplex(complex(math.NaN(), math.NaN())).ToObject(), nil},
		{Pow, NewComplex(1 + 2i).ToObject(), NewComplex(complex(1, math.Inf(1))).ToObject(), NewComplex(complex(math.NaN(), math.NaN())).ToObject(), nil},
		{Pow, NewComplex(complex(math.NaN(), 1)).ToObject(), NewComplex(3 + 4i).ToObject(), NewComplex(complex(math.NaN(), math.NaN())).ToObject(), nil},
		{Pow, NewComplex(3 + 4i).ToObject(), NewInt(3).ToObject(), NewComplex(-117 + 44.000000000000036i).ToObject(), nil},
		{Pow, NewComplex(3 + 4i).ToObject(), NewFloat(3.1415).ToObject(), NewComplex(-152.8892667678244 + 35.55533513049651i).ToObject(), nil},
		{Pow, NewComplex(3 + 4i).ToObject(), NewLong(big.NewInt(123)).ToObject(), NewComplex(5.393538720276193e+85 + 7.703512580443326e+85i).ToObject(), nil},
		{Pow, NewComplex(1 + 2i).ToObject(), NewStr("foo").ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for **: 'complex' and 'str'")},
		{Pow, NewStr("foo").ToObject(), NewComplex(1 + 2i).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for **: 'str' and 'complex'")},
	}

	for _, cas := range cases {
		switch got, result := checkInvokeResult(wrapFuncForTest(cas.fun), []*Object{cas.v, cas.w}, cas.want, cas.wantExc); result {
		case checkInvokeResultExceptionMismatch:
			t.Errorf("%s(%v, %v) raised %v, want %v", getFuncName(cas.fun), cas.v, cas.w, got, cas.wantExc)
		case checkInvokeResultReturnValueMismatch:
			if got == nil || cas.want == nil || !got.isInstance(ComplexType) || !cas.want.isInstance(ComplexType) ||
				!complexesAreSame(toComplexUnsafe(got).Value(), toComplexUnsafe(cas.want).Value()) {
				t.Errorf("%s(%v, %v) = %v, want %v", getFuncName(cas.fun), cas.v, cas.w, got, cas.want)
			}
		}
	}
}

func TestComplexCompareNotSupported(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(complex(1, 2), 1), wantExc: mustCreateException(TypeErrorType, "no ordering relation is defined for complex numbers")},
		{args: wrapArgs(complex(1, 2), 1.2), wantExc: mustCreateException(TypeErrorType, "no ordering relation is defined for complex numbers")},
		{args: wrapArgs(complex(1, 2), math.NaN()), wantExc: mustCreateException(TypeErrorType, "no ordering relation is defined for complex numbers")},
		{args: wrapArgs(complex(1, 2), math.Inf(-1)), wantExc: mustCreateException(TypeErrorType, "no ordering relation is defined for complex numbers")},
		{args: wrapArgs(complex(1, 2), "abc"), want: NotImplemented},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(complexCompareNotSupported), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestComplexNE(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(complex(0, 0), 0), want: False.ToObject()},
		{args: wrapArgs(complex(1, 0), 0), want: True.ToObject()},
		{args: wrapArgs(complex(-12, 0), -12), want: False.ToObject()},
		{args: wrapArgs(complex(-12, 0), 1), want: True.ToObject()},
		{args: wrapArgs(complex(17.20, 0), 17.20), want: False.ToObject()},
		{args: wrapArgs(complex(1.2, 0), 17.20), want: True.ToObject()},
		{args: wrapArgs(complex(-4, 15), complex(-4, 15)), want: False.ToObject()},
		{args: wrapArgs(complex(-4, 15), complex(1, 2)), want: True.ToObject()},
		{args: wrapArgs(complex(math.Inf(1), 0), complex(math.Inf(1), 0)), want: False.ToObject()},
		{args: wrapArgs(complex(math.Inf(1), 0), complex(0, math.Inf(1))), want: True.ToObject()},
		{args: wrapArgs(complex(math.Inf(-1), 0), complex(math.Inf(-1), 0)), want: False.ToObject()},
		{args: wrapArgs(complex(math.Inf(-1), 0), complex(0, math.Inf(-1))), want: True.ToObject()},
		{args: wrapArgs(complex(math.Inf(1), math.Inf(1)), complex(math.Inf(1), math.Inf(1))), want: False.ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(complexNE), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestComplexNew(t *testing.T) {
	complexNew := mustNotRaise(GetAttr(NewRootFrame(), ComplexType.ToObject(), NewStr("__new__"), nil))
	goodSlotType := newTestClass("GoodSlot", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__complex__": newBuiltinFunction("__complex__", func(_ *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return NewComplex(complex(1, 2)).ToObject(), nil
		}).ToObject(),
	}))
	badSlotType := newTestClass("BadSlot", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__complex__": newBuiltinFunction("__complex__", func(_ *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return newObject(ObjectType), nil
		}).ToObject(),
	}))
	strictEqType := newTestClassStrictEq("StrictEq", ComplexType)
	newStrictEq := func(v complex128) *Object {
		f := Complex{Object: Object{typ: strictEqType}, value: v}
		return f.ToObject()
	}
	subType := newTestClass("SubType", []*Type{ComplexType}, newStringDict(map[string]*Object{}))
	subTypeObject := (&Complex{Object: Object{typ: subType}, value: 3.14}).ToObject()
	slotSubTypeType := newTestClass("SlotSubType", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__complex__": newBuiltinFunction("__complex__", func(_ *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return subTypeObject, nil
		}).ToObject(),
	}))
	cases := []invokeTestCase{
		{args: wrapArgs(ComplexType), want: NewComplex(0).ToObject()},
		{args: wrapArgs(ComplexType, 56), want: NewComplex(complex(56, 0)).ToObject()},
		{args: wrapArgs(ComplexType, -12), want: NewComplex(complex(-12, 0)).ToObject()},
		{args: wrapArgs(ComplexType, 3.14), want: NewComplex(complex(3.14, 0)).ToObject()},
		{args: wrapArgs(ComplexType, -703.4), want: NewComplex(complex(-703.4, 0)).ToObject()},
		{args: wrapArgs(ComplexType, math.NaN()), want: NewComplex(complex(math.NaN(), 0)).ToObject()},
		{args: wrapArgs(ComplexType, math.Inf(1)), want: NewComplex(complex(math.Inf(1), 0)).ToObject()},
		{args: wrapArgs(ComplexType, math.Inf(-1)), want: NewComplex(complex(math.Inf(-1), 0)).ToObject()},
		{args: wrapArgs(ComplexType, biggestFloat), want: NewComplex(complex(math.MaxFloat64, 0)).ToObject()},
		{args: wrapArgs(ComplexType, new(big.Int).Neg(biggestFloat)), want: NewComplex(complex(-math.MaxFloat64, 0)).ToObject()},
		{args: wrapArgs(ComplexType, new(big.Int).Sub(big.NewInt(-1), biggestFloat)), wantExc: mustCreateException(OverflowErrorType, "long int too large to convert to float")},
		{args: wrapArgs(ComplexType, new(big.Int).Add(biggestFloat, big.NewInt(1))), wantExc: mustCreateException(OverflowErrorType, "long int too large to convert to float")},
		{args: wrapArgs(ComplexType, bigLongNumber), wantExc: mustCreateException(OverflowErrorType, "long int too large to convert to float")},
		{args: wrapArgs(ComplexType, complex(1, 2)), want: NewComplex(complex(1, 2)).ToObject()},
		{args: wrapArgs(ComplexType, complex(-0.0001e-1, 3.14151692)), want: NewComplex(complex(-0.00001, 3.14151692)).ToObject()},
		{args: wrapArgs(ComplexType, "23"), want: NewComplex(complex(23, 0)).ToObject()},
		{args: wrapArgs(ComplexType, "-516"), want: NewComplex(complex(-516, 0)).ToObject()},
		{args: wrapArgs(ComplexType, "1.003e4"), want: NewComplex(complex(10030, 0)).ToObject()},
		{args: wrapArgs(ComplexType, "151.7"), want: NewComplex(complex(151.7, 0)).ToObject()},
		{args: wrapArgs(ComplexType, "-74.02"), want: NewComplex(complex(-74.02, 0)).ToObject()},
		{args: wrapArgs(ComplexType, "+38.29"), want: NewComplex(complex(38.29, 0)).ToObject()},
		{args: wrapArgs(ComplexType, "8j"), want: NewComplex(complex(0, 8)).ToObject()},
		{args: wrapArgs(ComplexType, "-17j"), want: NewComplex(complex(0, -17)).ToObject()},
		{args: wrapArgs(ComplexType, "7.3j"), want: NewComplex(complex(0, 7.3)).ToObject()},
		{args: wrapArgs(ComplexType, "-4.786j"), want: NewComplex(complex(0, -4.786)).ToObject()},
		{args: wrapArgs(ComplexType, "+17.59123j"), want: NewComplex(complex(0, 17.59123)).ToObject()},
		{args: wrapArgs(ComplexType, "-3.0007e3j"), want: NewComplex(complex(0, -3000.7)).ToObject()},
		{args: wrapArgs(ComplexType, "1+2j"), want: NewComplex(complex(1, 2)).ToObject()},
		{args: wrapArgs(ComplexType, "3.1415-23j"), want: NewComplex(complex(3.1415, -23)).ToObject()},
		{args: wrapArgs(ComplexType, "-23+3.1415j"), want: NewComplex(complex(-23, 3.1415)).ToObject()},
		{args: wrapArgs(ComplexType, "+451.2192+384.27j"), want: NewComplex(complex(451.2192, 384.27)).ToObject()},
		{args: wrapArgs(ComplexType, "-38.378-283.28j"), want: NewComplex(complex(-38.378, -283.28)).ToObject()},
		{args: wrapArgs(ComplexType, "1.76123e2+0.000007e6j"), want: NewComplex(complex(176.123, 7)).ToObject()},
		{args: wrapArgs(ComplexType, "-nan+nanj"), want: NewComplex(complex(math.NaN(), math.NaN())).ToObject()},
		{args: wrapArgs(ComplexType, "inf-infj"), want: NewComplex(complex(math.Inf(1), math.Inf(-1))).ToObject()},
		{args: wrapArgs(ComplexType, 1, 2), want: NewComplex(complex(1, 2)).ToObject()},
		{args: wrapArgs(ComplexType, 7, 23.45), want: NewComplex(complex(7, 23.45)).ToObject()},
		{args: wrapArgs(ComplexType, 28.2537, -19), want: NewComplex(complex(28.2537, -19)).ToObject()},
		{args: wrapArgs(ComplexType, -3.14, -0.685), want: NewComplex(complex(-3.14, -0.685)).ToObject()},
		{args: wrapArgs(ComplexType, -47.234e+2, 2.374e+3), want: NewComplex(complex(-4723.4, 2374)).ToObject()},
		{args: wrapArgs(ComplexType, -4.5, new(big.Int).Neg(biggestFloat)), want: NewComplex(complex(-4.5, -math.MaxFloat64)).ToObject()},
		{args: wrapArgs(ComplexType, biggestFloat, biggestFloat), want: NewComplex(complex(math.MaxFloat64, math.MaxFloat64)).ToObject()},
		{args: wrapArgs(ComplexType, 5, math.NaN()), want: NewComplex(complex(5, math.NaN())).ToObject()},
		{args: wrapArgs(ComplexType, math.Inf(-1), -95), want: NewComplex(complex(math.Inf(-1), -95)).ToObject()},
		{args: wrapArgs(ComplexType, math.NaN(), math.NaN()), want: NewComplex(complex(math.NaN(), math.NaN())).ToObject()},
		{args: wrapArgs(ComplexType, math.Inf(1), math.Inf(-1)), want: NewComplex(complex(math.Inf(1), math.Inf(-1))).ToObject()},
		{args: wrapArgs(ComplexType, complex(-48.8, 0.7395), 5.448), want: NewComplex(complex(-48.8, 6.1875)).ToObject()},
		{args: wrapArgs(ComplexType, -3.14, complex(-4.5, -0.618)), want: NewComplex(complex(-2.5220000000000002, -4.5)).ToObject()},
		{args: wrapArgs(ComplexType, complex(1, 2), complex(3, 4)), want: NewComplex(complex(-3, 5)).ToObject()},
		{args: wrapArgs(ComplexType, complex(-2.47, 0.205e+2), complex(3.1, -0.4)), want: NewComplex(complex(-2.0700000000000003, 23.6)).ToObject()},
		{args: wrapArgs(ComplexType, "bar", 1.2), wantExc: mustCreateException(TypeErrorType, "complex() can't take second arg if first is a string")},
		{args: wrapArgs(ComplexType, "bar", None), wantExc: mustCreateException(TypeErrorType, "complex() can't take second arg if first is a string")},
		{args: wrapArgs(ComplexType, 1.2, "baz"), wantExc: mustCreateException(TypeErrorType, "complex() second arg can't be a string")},
		{args: wrapArgs(ComplexType, None, "baz"), wantExc: mustCreateException(TypeErrorType, "complex() second arg can't be a string")},
		{args: wrapArgs(ComplexType, newObject(goodSlotType)), want: NewComplex(complex(1, 2)).ToObject()},
		{args: wrapArgs(ComplexType, newObject(badSlotType)), wantExc: mustCreateException(TypeErrorType, "__complex__ returned non-complex (type object)")},
		{args: wrapArgs(ComplexType, newObject(slotSubTypeType)), want: subTypeObject},
		{args: wrapArgs(strictEqType, 3.14), want: newStrictEq(3.14)},
		{args: wrapArgs(strictEqType, newObject(goodSlotType)), want: newStrictEq(complex(1, 2))},
		{args: wrapArgs(strictEqType, newObject(badSlotType)), wantExc: mustCreateException(TypeErrorType, "__complex__ returned non-complex (type object)")},
		{args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, "'__new__' requires 1 arguments")},
		{args: wrapArgs(FloatType), wantExc: mustCreateException(TypeErrorType, "complex.__new__(float): float is not a subtype of complex")},
		{args: wrapArgs(ComplexType, None), wantExc: mustCreateException(TypeErrorType, "complex() argument must be a string or a number")},
		{args: wrapArgs(ComplexType, "foo"), wantExc: mustCreateException(ValueErrorType, "complex() arg is a malformed string")},
		{args: wrapArgs(ComplexType, 123, None, None), wantExc: mustCreateException(TypeErrorType, "'__new__' of 'complex' requires at most 2 arguments")},
	}
	for _, cas := range cases {
		switch got, match := checkInvokeResult(complexNew, cas.args, cas.want, cas.wantExc); match {
		case checkInvokeResultExceptionMismatch:
			t.Errorf("complex.__new__%v raised %v, want %v", cas.args, got, cas.wantExc)
		case checkInvokeResultReturnValueMismatch:
			// Handle NaN specially, since NaN != NaN.
			if got == nil || cas.want == nil || !got.isInstance(ComplexType) || !cas.want.isInstance(ComplexType) ||
				!cmplx.IsNaN(toComplexUnsafe(got).Value()) || !cmplx.IsNaN(toComplexUnsafe(cas.want).Value()) {
				t.Errorf("complex.__new__%v = %v, want %v", cas.args, got, cas.want)
			}
		}
	}
}

func TestComplexNonZero(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(complex(0, 0)), want: False.ToObject()},
		{args: wrapArgs(complex(.0, .0)), want: False.ToObject()},
		{args: wrapArgs(complex(0.0, 0.1)), want: True.ToObject()},
		{args: wrapArgs(complex(1, 0)), want: True.ToObject()},
		{args: wrapArgs(complex(3.14, -0.001e+5)), want: True.ToObject()},
		{args: wrapArgs(complex(math.NaN(), math.NaN())), want: True.ToObject()},
		{args: wrapArgs(complex(math.Inf(-1), math.Inf(1))), want: True.ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(complexNonZero), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestComplexPos(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(complex(0, 0)), want: NewComplex(complex(0, 0)).ToObject()},
		{args: wrapArgs(complex(42, -0.1)), want: NewComplex(complex(42, -0.1)).ToObject()},
		{args: wrapArgs(complex(-1.2, 375E+2)), want: NewComplex(complex(-1.2, 37500)).ToObject()},
		{args: wrapArgs(complex(5, math.NaN())), want: NewComplex(complex(5, math.NaN())).ToObject()},
		{args: wrapArgs(complex(math.Inf(1), 0.618)), want: NewComplex(complex(math.Inf(1), 0.618)).ToObject()},
	}
	for _, cas := range cases {
		switch got, match := checkInvokeResult(wrapFuncForTest(complexPos), cas.args, cas.want, cas.wantExc); match {
		case checkInvokeResultReturnValueMismatch:
			if got == nil || cas.want == nil || !got.isInstance(ComplexType) || !cas.want.isInstance(ComplexType) ||
				!complexesAreSame(toComplexUnsafe(got).Value(), toComplexUnsafe(cas.want).Value()) {
				t.Errorf("complex.__pos__%v = %v, want %v", cas.args, got, cas.want)
			}
		}
	}
}

func TestComplexRepr(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(complex(0.0, 0.0)), want: NewStr("0j").ToObject()},
		{args: wrapArgs(complex(0.0, 1.0)), want: NewStr("1j").ToObject()},
		{args: wrapArgs(complex(1.0, 2.0)), want: NewStr("(1+2j)").ToObject()},
		{args: wrapArgs(complex(3.1, -4.2)), want: NewStr("(3.1-4.2j)").ToObject()},
		{args: wrapArgs(complex(math.NaN(), math.NaN())), want: NewStr("(nan+nanj)").ToObject()},
		{args: wrapArgs(complex(math.Inf(-1), math.Inf(1))), want: NewStr("(-inf+infj)").ToObject()},
		{args: wrapArgs(complex(math.Inf(1), math.Inf(-1))), want: NewStr("(inf-infj)").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(Repr), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestParseComplex(t *testing.T) {
	var ErrSyntax = errors.New("invalid syntax")
	cases := []struct {
		s    string
		want complex128
		err  error
	}{
		{"5", complex(5, 0), nil},
		{"-3.14", complex(-3.14, 0), nil},
		{"1.8456e3", complex(1845.6, 0), nil},
		{"23j", complex(0, 23), nil},
		{"7j", complex(0, 7), nil},
		{"-365.12j", complex(0, -365.12), nil},
		{"1+2j", complex(1, 2), nil},
		{"-.3+.7j", complex(-0.3, 0.7), nil},
		{"-1.3+2.7j", complex(-1.3, 2.7), nil},
		{"48.39-20.3j", complex(48.39, -20.3), nil},
		{"-1.23e2-30.303j", complex(-123, -30.303), nil},
		{"-1.23e2-45.678e1j", complex(-123, -456.78), nil},
		{"nan+nanj", complex(math.NaN(), math.NaN()), nil},
		{"nan-nanj", complex(math.NaN(), math.NaN()), nil},
		{"-nan-nanj", complex(math.NaN(), math.NaN()), nil},
		{"inf+infj", complex(math.Inf(1), math.Inf(1)), nil},
		{"inf-infj", complex(math.Inf(1), math.Inf(-1)), nil},
		{"-inf-infj", complex(math.Inf(-1), math.Inf(-1)), nil},
		{"infINIty+infinityj", complex(math.Inf(1), math.Inf(1)), nil},
		{"3.4+j", complex(3.4, 1), nil},
		{"21.98-j", complex(21.98, -1), nil},
		{"+j", complex(0, 1), nil},
		{"-j", complex(0, -1), nil},
		{"j", complex(0, 1), nil},
		{"(2.1-3.4j)", complex(2.1, -3.4), nil},
		{"   (2.1-3.4j)    ", complex(2.1, -3.4), nil},
		{"   (   2.1-3.4j    )     ", complex(2.1, -3.4), nil},
		{" \t \n \r ( \t \n \r 2.1-3.4j \t \n \r ) \t \n \r ", complex(2.1, -3.4), nil},
		{"     3.14-15.16j   ", complex(3.14, -15.16), nil},
		{"(2.1-3.4j", complex(0, 0), ErrSyntax},
		{"((2.1-3.4j))", complex(0, 0), ErrSyntax},
		{"3.14 -15.16j", complex(0, 0), ErrSyntax},
		{"3.14- 15.16j", complex(0, 0), ErrSyntax},
		{"3.14-15.16 j", complex(0, 0), ErrSyntax},
		{"3.14 - 15.16 j", complex(0, 0), ErrSyntax},
		{"foo", complex(0, 0), ErrSyntax},
		{"foo+bar", complex(0, 0), ErrSyntax},
	}
	for _, cas := range cases {
		if got, _ := parseComplex(cas.s); !complexesAreSame(got, cas.want) {
			t.Errorf("parseComplex(%q) = %g, want %g", cas.s, got, cas.want)
		}
	}
}

func TestComplexHash(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(complex(0.0, 0.0)), want: NewInt(0).ToObject()},
		{args: wrapArgs(complex(0.0, 1.0)), want: NewInt(1000003).ToObject()},
		{args: wrapArgs(complex(1.0, 0.0)), want: NewInt(1).ToObject()},
		{args: wrapArgs(complex(3.1, -4.2)), want: NewInt(-1556830019620134).ToObject()},
		{args: wrapArgs(complex(3.1, 4.2)), want: NewInt(1557030815934348).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(complexHash), &cas); err != "" {
			t.Error(err)
		}
	}
}

func floatsAreSame(a, b float64) bool {
	return a == b || (math.IsNaN(a) && math.IsNaN(b))
}

func complexesAreSame(a, b complex128) bool {
	return floatsAreSame(real(a), real(b)) && floatsAreSame(imag(a), imag(b))
}

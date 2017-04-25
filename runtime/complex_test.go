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
		// {args: wrapArgs(ComplexType, complex(1, 2)), want: NewComplex(complex(1, 2)).ToObject()},
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
		// {args: wrapArgs(ComplexType, "-nan+nanj"), want: NewComplex(complex(math.NaN(), math.NaN())).ToObject()},
		{args: wrapArgs(ComplexType, "inf-infj"), want: NewComplex(complex(math.Inf(1), math.Inf(-1))).ToObject()},
		// {args: wrapArgs(ComplexType, 1, 2), want: NewComplex(complex(1, 2)).ToObject()},
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

func TestComplexRepr(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(complex(0.0, 0.0)), want: NewStr("0j").ToObject()},
		{args: wrapArgs(complex(0.0, 1.0)), want: NewStr("1j").ToObject()},
		{args: wrapArgs(complex(1.0, 2.0)), want: NewStr("(1+2j)").ToObject()},
		{args: wrapArgs(complex(3.1, -4.2)), want: NewStr("(3.1-4.2j)").ToObject()},
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
		if got, _ := parseComplex(cas.s); got != cas.want {
			// Handle NaN specially, since NaN != NaN.
			if !cmplx.IsNaN(got) || !cmplx.IsNaN(cas.want) {
				t.Errorf("parseComplex(%q) = %g, want %g", cas.s, got, cas.want)
			}
		}
	}
}

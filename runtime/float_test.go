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
	"math"
	"math/big"
	"testing"
)

var (
	// bigLongNumber is a number that is too large to be converted to
	// a float64.
	bigLongNumber = new(big.Int).Exp(big.NewInt(10), big.NewInt(1000), nil)
	// biggestFloat is the largest integer that can be converted to a
	// Python long (in CPython) without overflow.
	// Its value is 2**1024 - 2**(1024-54) - 1.
	biggestFloat = func(z *big.Int) *big.Int {
		z.SetBit(z, 1024, 1)
		z.Sub(z, big.NewInt(1))
		z.SetBit(z, 1024-54, 0)
		return z
	}(new(big.Int))
)

func TestFloatArithmeticOps(t *testing.T) {
	cases := []struct {
		fun     func(f *Frame, v, w *Object) (*Object, *BaseException)
		v, w    *Object
		want    *Object
		wantExc *BaseException
	}{
		{Add, NewFloat(1).ToObject(), NewFloat(1).ToObject(), NewFloat(2).ToObject(), nil},
		{Add, NewFloat(1.5).ToObject(), NewInt(1).ToObject(), NewFloat(2.5).ToObject(), nil},
		{Add, NewInt(1).ToObject(), NewFloat(1.5).ToObject(), NewFloat(2.5).ToObject(), nil},
		{Add, NewFloat(1.7976931348623157e+308).ToObject(), NewFloat(1.7976931348623157e+308).ToObject(), NewFloat(math.Inf(1)).ToObject(), nil},
		{Add, NewFloat(1.7976931348623157e+308).ToObject(), NewFloat(-1.7976931348623157e+308).ToObject(), NewFloat(0).ToObject(), nil},
		{Add, NewFloat(math.Inf(1)).ToObject(), NewFloat(math.Inf(1)).ToObject(), NewFloat(math.Inf(1)).ToObject(), nil},
		{Add, NewFloat(math.Inf(1)).ToObject(), NewFloat(math.Inf(-1)).ToObject(), NewFloat(math.NaN()).ToObject(), nil},
		{Add, newObject(ObjectType), NewFloat(-1).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for +: 'object' and 'float'")},
		{Div, NewFloat(12.5).ToObject(), NewFloat(4).ToObject(), NewFloat(3.125).ToObject(), nil},
		{Div, NewFloat(-12.5).ToObject(), NewInt(4).ToObject(), NewFloat(-3.125).ToObject(), nil},
		{Div, NewInt(25).ToObject(), NewFloat(5).ToObject(), NewFloat(5.0).ToObject(), nil},
		{Div, NewFloat(math.Inf(1)).ToObject(), NewFloat(math.Inf(1)).ToObject(), NewFloat(math.NaN()).ToObject(), nil},
		{Div, NewFloat(math.Inf(-1)).ToObject(), NewInt(-20).ToObject(), NewFloat(math.Inf(1)).ToObject(), nil},
		{Div, NewInt(1).ToObject(), NewFloat(math.Inf(1)).ToObject(), NewFloat(0).ToObject(), nil},
		{Div, newObject(ObjectType), NewFloat(1.1).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for /: 'object' and 'float'")},
		{Div, True.ToObject(), NewFloat(0).ToObject(), nil, mustCreateException(ZeroDivisionErrorType, "float division or modulo by zero")},
		{Div, NewFloat(math.Inf(1)).ToObject(), NewFloat(0).ToObject(), nil, mustCreateException(ZeroDivisionErrorType, "float division or modulo by zero")},
		{Div, NewFloat(1.0).ToObject(), NewLong(bigLongNumber).ToObject(), nil, mustCreateException(OverflowErrorType, "long int too large to convert to float")},
		{FloorDiv, NewFloat(12.5).ToObject(), NewFloat(4).ToObject(), NewFloat(3).ToObject(), nil},
		{FloorDiv, NewFloat(-12.5).ToObject(), NewInt(4).ToObject(), NewFloat(-4).ToObject(), nil},
		{FloorDiv, NewInt(25).ToObject(), NewFloat(5).ToObject(), NewFloat(5.0).ToObject(), nil},
		{FloorDiv, NewFloat(math.Inf(1)).ToObject(), NewFloat(math.Inf(1)).ToObject(), NewFloat(math.NaN()).ToObject(), nil},
		{FloorDiv, NewFloat(math.Inf(-1)).ToObject(), NewInt(-20).ToObject(), NewFloat(math.Inf(1)).ToObject(), nil},
		{FloorDiv, NewInt(1).ToObject(), NewFloat(math.Inf(1)).ToObject(), NewFloat(0).ToObject(), nil},
		{FloorDiv, newObject(ObjectType), NewFloat(1.1).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for //: 'object' and 'float'")},
		{FloorDiv, NewFloat(1.0).ToObject(), NewLong(bigLongNumber).ToObject(), nil, mustCreateException(OverflowErrorType, "long int too large to convert to float")},
		{FloorDiv, True.ToObject(), NewFloat(0).ToObject(), nil, mustCreateException(ZeroDivisionErrorType, "float division or modulo by zero")},
		{FloorDiv, NewFloat(math.Inf(1)).ToObject(), NewFloat(0).ToObject(), nil, mustCreateException(ZeroDivisionErrorType, "float division or modulo by zero")},
		{Mod, NewFloat(50.5).ToObject(), NewInt(10).ToObject(), NewFloat(0.5).ToObject(), nil},
		{Mod, NewFloat(50.5).ToObject(), NewFloat(-10).ToObject(), NewFloat(-9.5).ToObject(), nil},
		{Mod, NewFloat(-20.2).ToObject(), NewFloat(40).ToObject(), NewFloat(19.8).ToObject(), nil},
		{Mod, NewFloat(math.Inf(1)).ToObject(), NewInt(10).ToObject(), NewFloat(math.NaN()).ToObject(), nil},
		{Mod, NewInt(17).ToObject(), NewFloat(-4.25).ToObject(), NewFloat(0).ToObject(), nil},
		{Mod, NewInt(10).ToObject(), NewFloat(-8).ToObject(), NewFloat(-6).ToObject(), nil},
		{Mod, NewFloat(4.5).ToObject(), NewFloat(math.Inf(1)).ToObject(), NewFloat(4.5).ToObject(), nil},
		{Mod, NewFloat(4.5).ToObject(), NewFloat(math.Inf(-1)).ToObject(), NewFloat(math.Inf(-1)).ToObject(), nil},
		{Mod, NewFloat(math.Inf(1)).ToObject(), NewFloat(math.Inf(-1)).ToObject(), NewFloat(math.NaN()).ToObject(), nil},
		{Mod, None, NewFloat(42).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for %: 'NoneType' and 'float'")},
		{Mod, NewFloat(-32.25).ToObject(), NewInt(0).ToObject(), nil, mustCreateException(ZeroDivisionErrorType, "float division or modulo by zero")},
		{Mod, NewFloat(math.Inf(-1)).ToObject(), NewFloat(0).ToObject(), nil, mustCreateException(ZeroDivisionErrorType, "float division or modulo by zero")},
		{Mod, NewInt(2).ToObject(), NewFloat(0).ToObject(), nil, mustCreateException(ZeroDivisionErrorType, "float division or modulo by zero")},
		{Mul, NewFloat(1.2).ToObject(), True.ToObject(), NewFloat(1.2).ToObject(), nil},
		{Mul, NewInt(-4).ToObject(), NewFloat(1.2).ToObject(), NewFloat(-4.8).ToObject(), nil},
		{Mul, NewFloat(math.Inf(1)).ToObject(), NewInt(-5).ToObject(), NewFloat(math.Inf(-1)).ToObject(), nil},
		{Mul, False.ToObject(), NewFloat(math.Inf(1)).ToObject(), NewFloat(math.NaN()).ToObject(), nil},
		{Mul, None, NewFloat(1.5).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for *: 'NoneType' and 'float'")},
		{Pow, NewFloat(2.0).ToObject(), NewInt(10).ToObject(), NewFloat(1024.0).ToObject(), nil},
		{Pow, NewFloat(2.0).ToObject(), NewFloat(-2.0).ToObject(), NewFloat(0.25).ToObject(), nil},
		{Pow, newObject(ObjectType), NewFloat(2.0).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for **: 'object' and 'float'")},
		{Pow, NewFloat(2.0).ToObject(), newObject(ObjectType), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for **: 'float' and 'object'")},
		{Sub, NewFloat(21.3).ToObject(), NewFloat(35.6).ToObject(), NewFloat(-14.3).ToObject(), nil},
		{Sub, True.ToObject(), NewFloat(1.5).ToObject(), NewFloat(-0.5).ToObject(), nil},
		{Sub, NewFloat(1.0).ToObject(), NewList().ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for -: 'float' and 'list'")},
		{Sub, NewFloat(math.Inf(1)).ToObject(), NewFloat(math.Inf(1)).ToObject(), NewFloat(math.NaN()).ToObject(), nil},
	}
	for _, cas := range cases {
		switch got, result := checkInvokeResult(wrapFuncForTest(cas.fun), []*Object{cas.v, cas.w}, cas.want, cas.wantExc); result {
		case checkInvokeResultExceptionMismatch:
			t.Errorf("%s(%v, %v) raised %v, want %v", getFuncName(cas.fun), cas.v, cas.w, got, cas.wantExc)
		case checkInvokeResultReturnValueMismatch:
			// Handle NaN specially, since NaN != NaN.
			if got == nil || cas.want == nil || !got.isInstance(FloatType) || !cas.want.isInstance(FloatType) ||
				!math.IsNaN(toFloatUnsafe(got).Value()) || !math.IsNaN(toFloatUnsafe(cas.want).Value()) {
				t.Errorf("%s(%v, %v) = %v, want %v", getFuncName(cas.fun), cas.v, cas.w, got, cas.want)
			}
		}
	}
}

func TestFloatDivMod(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(12.5, 4.0), want: NewTuple2(NewFloat(3).ToObject(), NewFloat(0.5).ToObject()).ToObject()},
		{args: wrapArgs(-12.5, 4.0), want: NewTuple2(NewFloat(-4).ToObject(), NewFloat(3.5).ToObject()).ToObject()},
		{args: wrapArgs(25.0, 5.0), want: NewTuple2(NewFloat(5).ToObject(), NewFloat(0).ToObject()).ToObject()},
		{args: wrapArgs(-20.2, 40.0), want: NewTuple2(NewFloat(-1).ToObject(), NewFloat(19.8).ToObject()).ToObject()},
		{args: wrapArgs(math.Inf(1), math.Inf(1)), want: NewTuple2(NewFloat(math.NaN()).ToObject(), NewFloat(math.NaN()).ToObject()).ToObject()},
		{args: wrapArgs(math.Inf(1), math.Inf(-1)), want: NewTuple2(NewFloat(math.NaN()).ToObject(), NewFloat(math.NaN()).ToObject()).ToObject()},
		{args: wrapArgs(math.Inf(-1), -20.0), want: NewTuple2(NewFloat(math.Inf(1)).ToObject(), NewFloat(math.NaN()).ToObject()).ToObject()},
		{args: wrapArgs(1, math.Inf(1)), want: NewTuple2(NewFloat(0).ToObject(), NewFloat(1).ToObject()).ToObject()},
		{args: wrapArgs(newObject(ObjectType), 1.1), wantExc: mustCreateException(TypeErrorType, "unsupported operand type(s) for divmod(): 'object' and 'float'")},
		{args: wrapArgs(True.ToObject(), 0.0), wantExc: mustCreateException(ZeroDivisionErrorType, "float division or modulo by zero")},
		{args: wrapArgs(math.Inf(1), 0.0), wantExc: mustCreateException(ZeroDivisionErrorType, "float division or modulo by zero")},
		{args: wrapArgs(1.0, bigLongNumber), wantExc: mustCreateException(OverflowErrorType, "long int too large to convert to float")},
	}
	for _, cas := range cases {
		switch got, result := checkInvokeResult(wrapFuncForTest(DivMod), cas.args, cas.want, cas.wantExc); result {
		case checkInvokeResultExceptionMismatch:
			t.Errorf("float.__divmod__%v raised %v, want %v", cas.args, got, cas.wantExc)
		case checkInvokeResultReturnValueMismatch:
			// Handle NaN specially, since NaN != NaN.
			if got == nil || cas.want == nil || !got.isInstance(TupleType) || !cas.want.isInstance(TupleType) ||
				!isNaNTupleFloat(got, cas.want) {
				t.Errorf("float.__divmod__%v = %v, want %v", cas.args, got, cas.want)
			}
		}
	}
}

func isNaNTupleFloat(got, want *Object) bool {
	if toTupleUnsafe(got).Len() != toTupleUnsafe(want).Len() {
		return false
	}
	for i := 0; i < toTupleUnsafe(got).Len(); i++ {
		if math.IsNaN(toFloatUnsafe(toTupleUnsafe(got).GetItem(i)).Value()) &&
			math.IsNaN(toFloatUnsafe(toTupleUnsafe(want).GetItem(i)).Value()) {
			return true
		}
	}
	return false
}

func TestFloatCompare(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(1.0, 1.0), want: compareAllResultEq},
		{args: wrapArgs(30968.3958, 30968.3958), want: compareAllResultEq},
		{args: wrapArgs(-306.5, 101.0), want: compareAllResultLT},
		{args: wrapArgs(309683.958, 102.1), want: compareAllResultGT},
		{args: wrapArgs(0.9, 1), want: compareAllResultLT},
		{args: wrapArgs(0.0, 0), want: compareAllResultEq},
		{args: wrapArgs(1, 0.9), want: compareAllResultGT},
		{args: wrapArgs(0, 0.0), want: compareAllResultEq},
		{args: wrapArgs(0.0, None), want: compareAllResultGT},
		{args: wrapArgs(math.Inf(+1), bigLongNumber), want: compareAllResultGT},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(compareAll, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestFloatInt(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(IntType, -1209539058.2), want: NewInt(-1209539058).ToObject()},
		{args: wrapArgs(IntType, 2.994514758031654e+186), want: NewLong(func() *big.Int { i, _ := big.NewFloat(2.994514758031654e+186).Int(nil); return i }()).ToObject()},
		{args: wrapArgs(IntType, math.Inf(1)), wantExc: mustCreateException(OverflowErrorType, "cannot convert float infinity to integer")},
		{args: wrapArgs(IntType, math.Inf(-1)), wantExc: mustCreateException(OverflowErrorType, "cannot convert float infinity to integer")},
		{args: wrapArgs(IntType, math.NaN()), wantExc: mustCreateException(OverflowErrorType, "cannot convert float NaN to integer")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(IntType, "__new__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestFloatLong(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(LongType, -3209539058.2), want: NewLong(big.NewInt(-3209539058)).ToObject()},
		{args: wrapArgs(LongType, 2.994514758031654e+186), want: NewLong(func() *big.Int { i, _ := big.NewFloat(2.994514758031654e+186).Int(nil); return i }()).ToObject()},
		{args: wrapArgs(LongType, math.Inf(1)), wantExc: mustCreateException(OverflowErrorType, "cannot convert float infinity to integer")},
		{args: wrapArgs(LongType, math.Inf(-1)), wantExc: mustCreateException(OverflowErrorType, "cannot convert float infinity to integer")},
		{args: wrapArgs(LongType, math.NaN()), wantExc: mustCreateException(OverflowErrorType, "cannot convert float NaN to integer")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(LongType, "__new__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestFloatHash(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(NewFloat(0.0)), want: NewInt(0).ToObject()},
		{args: wrapArgs(NewFloat(3.14)), want: NewInt(3146129223).ToObject()},
		{args: wrapArgs(NewFloat(42.0)), want: NewInt(42).ToObject()},
		{args: wrapArgs(NewFloat(42.125)), want: NewInt(1413677056).ToObject()},
		{args: wrapArgs(NewFloat(math.Inf(1))), want: NewInt(314159).ToObject()},
		{args: wrapArgs(NewFloat(math.Inf(-1))), want: NewInt(-271828).ToObject()},
		{args: wrapArgs(NewFloat(math.NaN())), want: NewInt(0).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(floatHash), &cas); err != "" {
			t.Error(err)
		}
	}
}
func TestFloatIsTrue(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(0.0), want: False.ToObject()},
		{args: wrapArgs(0.0001), want: True.ToObject()},
		{args: wrapArgs(36983.91283), want: True.ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(IsTrue), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestFloatNew(t *testing.T) {
	floatNew := mustNotRaise(GetAttr(NewRootFrame(), FloatType.ToObject(), NewStr("__new__"), nil))
	strictEqType := newTestClassStrictEq("StrictEq", FloatType)
	newStrictEq := func(v float64) *Object {
		f := Float{Object: Object{typ: strictEqType}, value: v}
		return f.ToObject()
	}
	subType := newTestClass("SubType", []*Type{FloatType}, newStringDict(map[string]*Object{}))
	subTypeObject := (&Float{Object: Object{typ: subType}, value: 3.14}).ToObject()
	goodSlotType := newTestClass("GoodSlot", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__float__": newBuiltinFunction("__float__", func(_ *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return NewFloat(3.14).ToObject(), nil
		}).ToObject(),
	}))
	badSlotType := newTestClass("BadSlot", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__float__": newBuiltinFunction("__float__", func(_ *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return newObject(ObjectType), nil
		}).ToObject(),
	}))
	slotSubTypeType := newTestClass("SlotSubType", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__float__": newBuiltinFunction("__float__", func(_ *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return subTypeObject, nil
		}).ToObject(),
	}))

	cases := []invokeTestCase{
		{args: wrapArgs(FloatType), want: NewFloat(0).ToObject()},
		{args: wrapArgs(FloatType, 10.5), want: NewFloat(10.5).ToObject()},
		{args: wrapArgs(FloatType, -102.1), want: NewFloat(-102.1).ToObject()},
		{args: wrapArgs(FloatType, 42), want: NewFloat(42).ToObject()},
		{args: wrapArgs(FloatType, "1.024e3"), want: NewFloat(1024).ToObject()},
		{args: wrapArgs(FloatType, "-42"), want: NewFloat(-42).ToObject()},
		{args: wrapArgs(FloatType, math.Inf(1)), want: NewFloat(math.Inf(1)).ToObject()},
		{args: wrapArgs(FloatType, math.Inf(-1)), want: NewFloat(math.Inf(-1)).ToObject()},
		{args: wrapArgs(FloatType, math.NaN()), want: NewFloat(math.NaN()).ToObject()},
		{args: wrapArgs(FloatType, biggestFloat), want: NewFloat(math.MaxFloat64).ToObject()},
		{args: wrapArgs(FloatType, new(big.Int).Neg(biggestFloat)), want: NewFloat(-math.MaxFloat64).ToObject()},
		{args: wrapArgs(FloatType, new(big.Int).Sub(big.NewInt(-1), biggestFloat)), wantExc: mustCreateException(OverflowErrorType, "long int too large to convert to float")},
		{args: wrapArgs(FloatType, new(big.Int).Add(biggestFloat, big.NewInt(1))), wantExc: mustCreateException(OverflowErrorType, "long int too large to convert to float")},
		{args: wrapArgs(FloatType, bigLongNumber), wantExc: mustCreateException(OverflowErrorType, "long int too large to convert to float")},
		{args: wrapArgs(FloatType, newObject(goodSlotType)), want: NewFloat(3.14).ToObject()},
		{args: wrapArgs(FloatType, newObject(badSlotType)), wantExc: mustCreateException(TypeErrorType, "__float__ returned non-float (type object)")},
		{args: wrapArgs(FloatType, newObject(slotSubTypeType)), want: subTypeObject},
		{args: wrapArgs(strictEqType, 3.14), want: newStrictEq(3.14)},
		{args: wrapArgs(strictEqType, newObject(goodSlotType)), want: newStrictEq(3.14)},
		{args: wrapArgs(strictEqType, newObject(badSlotType)), wantExc: mustCreateException(TypeErrorType, "__float__ returned non-float (type object)")},
		{args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, "'__new__' requires 1 arguments")},
		{args: wrapArgs(IntType), wantExc: mustCreateException(TypeErrorType, "float.__new__(int): int is not a subtype of float")},
		{args: wrapArgs(FloatType, 123, None), wantExc: mustCreateException(TypeErrorType, "'__new__' of 'float' requires 0 or 1 arguments")},
		{args: wrapArgs(FloatType, "foo"), wantExc: mustCreateException(ValueErrorType, "could not convert string to float: foo")},
		{args: wrapArgs(FloatType, None), wantExc: mustCreateException(TypeErrorType, "float() argument must be a string or a number")},
	}
	for _, cas := range cases {
		switch got, match := checkInvokeResult(floatNew, cas.args, cas.want, cas.wantExc); match {
		case checkInvokeResultExceptionMismatch:
			t.Errorf("float.__new__%v raised %v, want %v", cas.args, got, cas.wantExc)
		case checkInvokeResultReturnValueMismatch:
			// Handle NaN specially, since NaN != NaN.
			if got == nil || cas.want == nil || !got.isInstance(FloatType) || !cas.want.isInstance(FloatType) ||
				!math.IsNaN(toFloatUnsafe(got).Value()) || !math.IsNaN(toFloatUnsafe(cas.want).Value()) {
				t.Errorf("float.__new__%v = %v, want %v", cas.args, got, cas.want)
			}
		}
	}
}

func TestFloatRepr(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(0.0), want: NewStr("0.0").ToObject()},
		{args: wrapArgs(0.1), want: NewStr("0.1").ToObject()},
		{args: wrapArgs(-303.5), want: NewStr("-303.5").ToObject()},
		{args: wrapArgs(231095835.0), want: NewStr("231095835.0").ToObject()},
		{args: wrapArgs(1e+6), want: NewStr("1000000.0").ToObject()},
		{args: wrapArgs(1e+15), want: NewStr("1000000000000000.0").ToObject()},
		{args: wrapArgs(1e+16), want: NewStr("1e+16").ToObject()},
		{args: wrapArgs(1E16), want: NewStr("1e+16").ToObject()},
		{args: wrapArgs(1e-6), want: NewStr("1e-06").ToObject()},
		{args: wrapArgs(math.Inf(1)), want: NewStr("inf").ToObject()},
		{args: wrapArgs(math.Inf(-1)), want: NewStr("-inf").ToObject()},
		{args: wrapArgs(math.NaN()), want: NewStr("nan").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(Repr), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestFloatStr(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(1.0), want: NewStr("1.0").ToObject()},
		{args: wrapArgs(-847.373), want: NewStr("-847.373").ToObject()},
		{args: wrapArgs(0.123456789123456789), want: NewStr("0.123456789123").ToObject()},
		{args: wrapArgs(1e+11), want: NewStr("100000000000.0").ToObject()},
		{args: wrapArgs(1e+12), want: NewStr("1e+12").ToObject()},
		{args: wrapArgs(1e-4), want: NewStr("0.0001").ToObject()},
		{args: wrapArgs(1e-5), want: NewStr("1e-05").ToObject()},
		{args: wrapArgs(math.Inf(1)), want: NewStr("inf").ToObject()},
		{args: wrapArgs(math.Inf(-1)), want: NewStr("-inf").ToObject()},
		{args: wrapArgs(math.NaN()), want: NewStr("nan").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(floatStr), &cas); err != "" {
			t.Error(err)
		}
	}
}

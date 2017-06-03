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
	"math/big"
	"runtime"
	"testing"
)

func TestIntBinaryOps(t *testing.T) {
	cases := []struct {
		fun     binaryOpFunc
		v, w    *Object
		want    *Object
		wantExc *BaseException
	}{
		{Add, NewInt(-100).ToObject(), NewInt(50).ToObject(), NewInt(-50).ToObject(), nil},
		{Add, newObject(ObjectType), NewInt(-100).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for +: 'object' and 'int'")},
		{Add, NewInt(MaxInt).ToObject(), NewInt(1).ToObject(), NewLong(new(big.Int).Add(maxIntBig, big.NewInt(1))).ToObject(), nil},
		{And, NewInt(-100).ToObject(), NewInt(50).ToObject(), NewInt(16).ToObject(), nil},
		{And, NewInt(MaxInt).ToObject(), NewInt(MinInt).ToObject(), NewInt(0).ToObject(), nil},
		{And, newObject(ObjectType), NewInt(-100).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for &: 'object' and 'int'")},
		{Div, NewInt(7).ToObject(), NewInt(3).ToObject(), NewInt(2).ToObject(), nil},
		{Div, NewInt(MaxInt).ToObject(), NewInt(MinInt).ToObject(), NewInt(-1).ToObject(), nil},
		{Div, NewInt(MinInt).ToObject(), NewInt(MaxInt).ToObject(), NewInt(-2).ToObject(), nil},
		{Div, NewList().ToObject(), NewInt(21).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for /: 'list' and 'int'")},
		{Div, NewInt(1).ToObject(), NewInt(0).ToObject(), nil, mustCreateException(ZeroDivisionErrorType, "integer division or modulo by zero")},
		{Div, NewInt(MinInt).ToObject(), NewInt(-1).ToObject(), NewLong(new(big.Int).Neg(minIntBig)).ToObject(), nil},
		{DivMod, NewInt(7).ToObject(), NewInt(3).ToObject(), NewTuple2(NewInt(2).ToObject(), NewInt(1).ToObject()).ToObject(), nil},
		{DivMod, NewInt(3).ToObject(), NewInt(-7).ToObject(), NewTuple2(NewInt(-1).ToObject(), NewInt(-4).ToObject()).ToObject(), nil},
		{DivMod, NewInt(MaxInt).ToObject(), NewInt(MinInt).ToObject(), NewTuple2(NewInt(-1).ToObject(), NewInt(-1).ToObject()).ToObject(), nil},
		{DivMod, NewInt(MinInt).ToObject(), NewInt(MaxInt).ToObject(), NewTuple2(NewInt(-2).ToObject(), NewInt(MaxInt-1).ToObject()).ToObject(), nil},
		{DivMod, NewList().ToObject(), NewInt(21).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for divmod(): 'list' and 'int'")},
		{DivMod, NewInt(1).ToObject(), NewInt(0).ToObject(), nil, mustCreateException(ZeroDivisionErrorType, "integer division or modulo by zero")},
		{DivMod, NewInt(MinInt).ToObject(), NewInt(-1).ToObject(), NewTuple2(NewLong(new(big.Int).Neg(minIntBig)).ToObject(), NewLong(big.NewInt(0)).ToObject()).ToObject(), nil},
		{FloorDiv, NewInt(7).ToObject(), NewInt(3).ToObject(), NewInt(2).ToObject(), nil},
		{FloorDiv, NewInt(MaxInt).ToObject(), NewInt(MinInt).ToObject(), NewInt(-1).ToObject(), nil},
		{FloorDiv, NewInt(MinInt).ToObject(), NewInt(MaxInt).ToObject(), NewInt(-2).ToObject(), nil},
		{FloorDiv, NewList().ToObject(), NewInt(21).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for //: 'list' and 'int'")},
		{FloorDiv, NewInt(1).ToObject(), NewInt(0).ToObject(), nil, mustCreateException(ZeroDivisionErrorType, "integer division or modulo by zero")},
		{FloorDiv, NewInt(MinInt).ToObject(), NewInt(-1).ToObject(), NewLong(new(big.Int).Neg(minIntBig)).ToObject(), nil},
		{LShift, NewInt(2).ToObject(), NewInt(4).ToObject(), NewInt(32).ToObject(), nil},
		{LShift, NewInt(-12).ToObject(), NewInt(10).ToObject(), NewInt(-12288).ToObject(), nil},
		{LShift, NewInt(10).ToObject(), NewInt(100).ToObject(), NewLong(new(big.Int).Lsh(big.NewInt(10), 100)).ToObject(), nil},
		{LShift, NewInt(2).ToObject(), NewInt(-5).ToObject(), nil, mustCreateException(ValueErrorType, "negative shift count")},
		{LShift, NewInt(4).ToObject(), NewFloat(3.14).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for <<: 'int' and 'float'")},
		{LShift, newObject(ObjectType), NewInt(4).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for <<: 'object' and 'int'")},
		{RShift, NewInt(87).ToObject(), NewInt(3).ToObject(), NewInt(10).ToObject(), nil},
		{RShift, NewInt(-101).ToObject(), NewInt(5).ToObject(), NewInt(-4).ToObject(), nil},
		{RShift, NewInt(12).ToObject(), NewInt(10).ToObject(), NewInt(0).ToObject(), nil},
		{RShift, NewInt(12).ToObject(), NewInt(-10).ToObject(), nil, mustCreateException(ValueErrorType, "negative shift count")},
		{RShift, NewInt(4).ToObject(), NewFloat(3.14).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for >>: 'int' and 'float'")},
		{RShift, newObject(ObjectType), NewInt(4).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for >>: 'object' and 'int'")},
		{RShift, NewInt(4).ToObject(), newObject(ObjectType), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for >>: 'int' and 'object'")},
		{Mod, NewInt(3).ToObject(), NewInt(-7).ToObject(), NewInt(-4).ToObject(), nil},
		{Mod, NewInt(MaxInt).ToObject(), NewInt(MinInt).ToObject(), NewInt(-1).ToObject(), nil},
		{Mod, NewInt(MinInt).ToObject(), NewInt(MaxInt).ToObject(), NewInt(MaxInt - 1).ToObject(), nil},
		{Mod, None, NewInt(-4).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for %: 'NoneType' and 'int'")},
		{Mod, NewInt(10).ToObject(), NewInt(0).ToObject(), nil, mustCreateException(ZeroDivisionErrorType, "integer division or modulo by zero")},
		{Mod, NewInt(MinInt).ToObject(), NewInt(-1).ToObject(), NewLong(big.NewInt(0)).ToObject(), nil},
		{Mul, NewInt(-1).ToObject(), NewInt(-3).ToObject(), NewInt(3).ToObject(), nil},
		{Mul, newObject(ObjectType), NewInt(101).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for *: 'object' and 'int'")},
		{Mul, NewInt(MaxInt).ToObject(), NewInt(MaxInt - 1).ToObject(), NewLong(new(big.Int).Mul(big.NewInt(MaxInt), big.NewInt(MaxInt-1))).ToObject(), nil},
		{Or, NewInt(-100).ToObject(), NewInt(50).ToObject(), NewInt(-66).ToObject(), nil},
		{Or, NewInt(MaxInt).ToObject(), NewInt(MinInt).ToObject(), NewInt(-1).ToObject(), nil},
		{Or, newObject(ObjectType), NewInt(-100).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for |: 'object' and 'int'")},
		{Pow, NewInt(2).ToObject(), NewInt(128).ToObject(), NewLong(big.NewInt(0).Exp(big.NewInt(2), big.NewInt(128), nil)).ToObject(), nil},
		{Pow, NewInt(2).ToObject(), newObject(ObjectType), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for **: 'int' and 'object'")},
		{Pow, NewInt(2).ToObject(), NewInt(-2).ToObject(), NewFloat(0.25).ToObject(), nil},
		{Pow, newObject(ObjectType), NewInt(2).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for **: 'object' and 'int'")},
		{Sub, NewInt(22).ToObject(), NewInt(18).ToObject(), NewInt(4).ToObject(), nil},
		{Sub, IntType.ToObject(), NewInt(42).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for -: 'type' and 'int'")},
		{Sub, NewInt(MinInt).ToObject(), NewInt(1).ToObject(), NewLong(new(big.Int).Sub(minIntBig, big.NewInt(1))).ToObject(), nil},
		{Xor, NewInt(-100).ToObject(), NewInt(50).ToObject(), NewInt(-82).ToObject(), nil},
		{Xor, NewInt(MaxInt).ToObject(), NewInt(MinInt).ToObject(), NewInt(-1).ToObject(), nil},
		{Xor, newObject(ObjectType), NewInt(-100).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for ^: 'object' and 'int'")},
	}
	for _, cas := range cases {
		testCase := invokeTestCase{args: wrapArgs(cas.v, cas.w), want: cas.want, wantExc: cas.wantExc}
		if err := runInvokeTestCase(wrapFuncForTest(cas.fun), &testCase); err != "" {
			t.Error(err)
		}
	}
}

func TestIntCompare(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(1, 1), want: compareAllResultEq},
		{args: wrapArgs(309683958, 309683958), want: compareAllResultEq},
		{args: wrapArgs(-306, 101), want: compareAllResultLT},
		{args: wrapArgs(309683958, 101), want: compareAllResultGT},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(compareAll, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestIntInvert(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(2592), want: NewInt(-2593).ToObject()},
		{args: wrapArgs(0), want: NewInt(-1).ToObject()},
		{args: wrapArgs(-43), want: NewInt(42).ToObject()},
		{args: wrapArgs(MaxInt), want: NewInt(MinInt).ToObject()},
		{args: wrapArgs(MinInt), want: NewInt(MaxInt).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(IntType, "__invert__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestIntNew(t *testing.T) {
	fooType := newTestClass("Foo", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__int__": newBuiltinFunction("__int__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return args[0], nil
		}).ToObject(),
	}))
	strictEqType := newTestClassStrictEq("StrictEq", IntType)
	subType := newTestClass("SubType", []*Type{IntType}, newStringDict(map[string]*Object{}))
	subTypeObject := (&Int{Object: Object{typ: subType}, value: 3}).ToObject()
	goodSlotType := newTestClass("GoodSlot", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__int__": newBuiltinFunction("__int__", func(_ *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return NewInt(3).ToObject(), nil
		}).ToObject(),
	}))
	badSlotType := newTestClass("BadSlot", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__int__": newBuiltinFunction("__int__", func(_ *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return newObject(ObjectType), nil
		}).ToObject(),
	}))
	slotSubTypeType := newTestClass("SlotSubType", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__int__": newBuiltinFunction("__int__", func(_ *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return subTypeObject, nil
		}).ToObject(),
	}))
	cases := []invokeTestCase{
		{args: wrapArgs(IntType), want: NewInt(0).ToObject()},
		{args: wrapArgs(IntType, "123"), want: NewInt(123).ToObject()},
		{args: wrapArgs(IntType, " \t123"), want: NewInt(123).ToObject()},
		{args: wrapArgs(IntType, "123 \t"), want: NewInt(123).ToObject()},
		{args: wrapArgs(IntType, "FF", 16), want: NewInt(255).ToObject()},
		{args: wrapArgs(IntType, "0xFF", 16), want: NewInt(255).ToObject()},
		{args: wrapArgs(IntType, "0xE", 0), want: NewInt(14).ToObject()},
		{args: wrapArgs(IntType, "0b101", 0), want: NewInt(5).ToObject()},
		{args: wrapArgs(IntType, "0o726", 0), want: NewInt(470).ToObject()},
		{args: wrapArgs(IntType, "0726", 0), want: NewInt(470).ToObject()},
		{args: wrapArgs(IntType, "102", 0), want: NewInt(102).ToObject()},
		{args: wrapArgs(IntType, 42), want: NewInt(42).ToObject()},
		{args: wrapArgs(IntType, -3.14), want: NewInt(-3).ToObject()},
		{args: wrapArgs(subType, overflowLong), wantExc: mustCreateException(OverflowErrorType, "Python int too large to convert to a Go int")},
		{args: wrapArgs(strictEqType, 42), want: (&Int{Object{typ: strictEqType}, 42}).ToObject()},
		{args: wrapArgs(IntType, newObject(goodSlotType)), want: NewInt(3).ToObject()},
		{args: wrapArgs(IntType, newObject(badSlotType)), wantExc: mustCreateException(TypeErrorType, "__int__ returned non-int (type object)")},
		{args: wrapArgs(IntType, newObject(slotSubTypeType)), want: subTypeObject},
		{args: wrapArgs(strictEqType, newObject(goodSlotType)), want: (&Int{Object{typ: strictEqType}, 3}).ToObject()},
		{args: wrapArgs(strictEqType, newObject(badSlotType)), wantExc: mustCreateException(TypeErrorType, "__int__ returned non-int (type object)")},
		{args: wrapArgs(IntType, "0xff"), wantExc: mustCreateException(ValueErrorType, "invalid literal for int() with base 10: 0xff")},
		{args: wrapArgs(IntType, ""), wantExc: mustCreateException(ValueErrorType, "invalid literal for int() with base 10: ")},
		{args: wrapArgs(IntType, " "), wantExc: mustCreateException(ValueErrorType, "invalid literal for int() with base 10:  ")},
		{args: wrapArgs(FloatType), wantExc: mustCreateException(TypeErrorType, "int.__new__(float): float is not a subtype of int")},
		{args: wrapArgs(IntType, "asldkfj", 1), wantExc: mustCreateException(ValueErrorType, "int() base must be >= 2 and <= 36")},
		{args: wrapArgs(IntType, "asldkfj", 37), wantExc: mustCreateException(ValueErrorType, "int() base must be >= 2 and <= 36")},
		{args: wrapArgs(IntType, "@#%*(#", 36), wantExc: mustCreateException(ValueErrorType, "invalid literal for int() with base 36: @#%*(#")},
		{args: wrapArgs(IntType, "123", overflowLong), wantExc: mustCreateException(OverflowErrorType, "Python int too large to convert to a Go int")},
		{args: wrapArgs(IntType, "32059823095809238509238590835"), want: NewLong(func() *big.Int { i, _ := new(big.Int).SetString("32059823095809238509238590835", 0); return i }()).ToObject()},
		{args: wrapArgs(IntType, newObject(ObjectType)), wantExc: mustCreateException(TypeErrorType, "int() argument must be a string or a number, not 'object'")},
		{args: wrapArgs(IntType, newObject(fooType)), wantExc: mustCreateException(TypeErrorType, "__int__ returned non-int (type Foo)")},
		{args: wrapArgs(IntType, 1, 2), wantExc: mustCreateException(TypeErrorType, "int() can't convert non-string with explicit base")},
		{args: wrapArgs(IntType, 1, 2, 3), wantExc: mustCreateException(TypeErrorType, "int() takes at most 2 arguments (3 given)")},
		{args: wrapArgs(IntType, "1", None), wantExc: mustCreateException(TypeErrorType, "an integer is required")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(IntType, "__new__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestIntNewInterned(t *testing.T) {
	// Make sure small integers are interned.
	fun := wrapFuncForTest(func(f *Frame, i *Int) (bool, *BaseException) {
		o, raised := IntType.Call(f, wrapArgs(i.Value()), nil)
		if raised != nil {
			return false, raised
		}
		return o == i.ToObject(), nil
	})
	cases := []invokeTestCase{
		{args: wrapArgs(-1001), want: False.ToObject()},
		{args: wrapArgs(0), want: True.ToObject()},
		{args: wrapArgs(100), want: True.ToObject()},
		{args: wrapArgs(120948298), want: False.ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func BenchmarkIntNew(b *testing.B) {
	b.Run("interned", func(b *testing.B) {
		var ret *Object
		for i := 0; i < b.N; i++ {
			ret = NewInt(1).ToObject()
		}
		runtime.KeepAlive(ret)
	})

	b.Run("not interned", func(b *testing.B) {
		var ret *Object
		for i := 0; i < b.N; i++ {
			ret = NewInt(internedIntMax + 5).ToObject()
		}
		runtime.KeepAlive(ret)
	})
}

func TestIntStrRepr(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(0), want: NewStr("0").ToObject()},
		{args: wrapArgs(-303), want: NewStr("-303").ToObject()},
		{args: wrapArgs(231095835), want: NewStr("231095835").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(ToStr), &cas); err != "" {
			t.Error(err)
		}
		if err := runInvokeTestCase(wrapFuncForTest(Repr), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestIntCheckedAddMul(t *testing.T) {
	cases := []struct {
		f      func(a, b int) (int, bool)
		a, b   int
		want   int
		wantOK bool
	}{
		{intCheckedAdd, 1, 2, 3, true},
		{intCheckedAdd, MaxInt, -1, MaxInt - 1, true},
		{intCheckedAdd, MaxInt, 0, MaxInt, true},
		{intCheckedAdd, MaxInt, 1, 0, false},
		{intCheckedAdd, MinInt, -1, 0, false},
		{intCheckedAdd, MinInt, 0, MinInt, true},
		{intCheckedAdd, MinInt, 1, MinInt + 1, true},
		{intCheckedMul, MaxInt, 1, MaxInt, true},
		{intCheckedMul, MaxInt, -1, MinInt + 1, true},
		{intCheckedMul, MinInt, -1, 0, false},
	}
	for _, cas := range cases {
		if got, gotOK := cas.f(cas.a, cas.b); got != cas.want || gotOK != cas.wantOK {
			t.Errorf("%s(%v, %v) = (%v, %v), want (%v, %v)", getFuncName(cas.f), cas.a, cas.b, got, gotOK, cas.want, cas.wantOK)
		}
		if got, gotOK := cas.f(cas.b, cas.a); got != cas.want || gotOK != cas.wantOK {
			t.Errorf("%s(%v, %v) = (%v, %v), want (%v, %v)", getFuncName(cas.f), cas.b, cas.a, got, gotOK, cas.want, cas.wantOK)
		}
	}
}

func TestIntCheckedDivMod(t *testing.T) {
	cases := []struct {
		f          func(a, b int) (int, divModResult)
		a, b       int
		want       int
		wantResult divModResult
	}{
		{intCheckedDiv, 872, 736, 1, divModOK},
		{intCheckedDiv, -320, 3, -107, divModOK},
		{intCheckedDiv, 7, 3, 2, divModOK},
		{intCheckedDiv, 7, -3, -3, divModOK},
		{intCheckedDiv, -7, 3, -3, divModOK},
		{intCheckedDiv, -7, -3, 2, divModOK},
		{intCheckedDiv, 3, 7, 0, divModOK},
		{intCheckedDiv, 3, -7, -1, divModOK},
		{intCheckedDiv, -3, 7, -1, divModOK},
		{intCheckedDiv, -3, -7, 0, divModOK},
		{intCheckedDiv, MaxInt, MaxInt, 1, divModOK},
		{intCheckedDiv, MaxInt, MinInt, -1, divModOK},
		{intCheckedDiv, MinInt, MaxInt, -2, divModOK},
		{intCheckedDiv, MinInt, MinInt, 1, divModOK},
		{intCheckedDiv, 22, 0, 0, divModZeroDivision},
		{intCheckedDiv, MinInt, -1, 0, divModOverflow},
		{intCheckedMod, -142, -118, -24, divModOK},
		{intCheckedMod, -225, 454, 229, divModOK},
		{intCheckedMod, 7, 3, 1, divModOK},
		{intCheckedMod, 7, -3, -2, divModOK},
		{intCheckedMod, -7, 3, 2, divModOK},
		{intCheckedMod, -7, -3, -1, divModOK},
		{intCheckedMod, 3, 7, 3, divModOK},
		{intCheckedMod, 3, -7, -4, divModOK},
		{intCheckedMod, -3, 7, 4, divModOK},
		{intCheckedMod, -3, -7, -3, divModOK},
		{intCheckedMod, MaxInt, MaxInt, 0, divModOK},
		{intCheckedMod, MaxInt, MinInt, -1, divModOK},
		{intCheckedMod, MinInt, MaxInt, MaxInt - 1, divModOK},
		{intCheckedMod, MinInt, MinInt, 0, divModOK},
		{intCheckedMod, -50, 0, 0, divModZeroDivision},
		{intCheckedMod, MinInt, -1, 0, divModOverflow},
	}
	for _, cas := range cases {
		if got, gotResult := cas.f(cas.a, cas.b); got != cas.want || gotResult != cas.wantResult {
			t.Errorf("%s(%v, %v) = (%v, %v), want (%v, %v)", getFuncName(cas.f), cas.a, cas.b, got, gotResult, cas.want, cas.wantResult)
		}
	}
}

func TestIntCheckedSub(t *testing.T) {
	cases := []struct {
		f      func(a, b int) (int, bool)
		a, b   int
		want   int
		wantOK bool
	}{
		{intCheckedSub, MaxInt, MaxInt, 0, true},
		{intCheckedSub, MaxInt, -1, 0, false},
		{intCheckedSub, MaxInt, 0, MaxInt, true},
		{intCheckedSub, MaxInt, 1, MaxInt - 1, true},
		{intCheckedSub, MinInt, -1, MinInt + 1, true},
		{intCheckedSub, MinInt, 0, MinInt, true},
		{intCheckedSub, MinInt, 1, 0, false},
		{intCheckedSub, MinInt, MinInt, 0, true},
		{intCheckedSub, -2, MaxInt, 0, false},
		{intCheckedSub, -1, MaxInt, MinInt, true},
		{intCheckedSub, 0, MaxInt, MinInt + 1, true},
		{intCheckedSub, -1, MinInt, MaxInt, true},
		{intCheckedSub, 0, MinInt, 0, false},
	}
	for _, cas := range cases {
		if got, gotOK := cas.f(cas.a, cas.b); got != cas.want || gotOK != cas.wantOK {
			t.Errorf("%s(%v, %v) = (%v, %v), want (%v, %v)", getFuncName(cas.f), cas.a, cas.b, got, gotOK, cas.want, cas.wantOK)
		}
	}
}

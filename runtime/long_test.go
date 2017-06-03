package grumpy

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"testing"
)

var overflowLong = big.NewInt(0).Add(maxIntBig, big.NewInt(101))

func TestLongBasis(t *testing.T) {
	got := LongType.slots.Basis.Fn(NewLong(big.NewInt(42)).ToObject()).Type()
	want := reflect.TypeOf(Long{})
	if got != want {
		t.Fatalf("LongType.slots.Basis.Fn(NewLong(big.NewInt(42).ToObject()).Type() = %v, want %v", got, want)
	}
}

func TestNewLongFromBytes(t *testing.T) {
	cases := []struct {
		bytes []byte
		want  string
	}{
		{bytes: []byte{0x01, 0x00}, want: "100"},
		{bytes: []byte{0x01, 0x02, 0x03}, want: "10203"},
		{bytes: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09},
			want: "10203040506070809"},
	}
	for _, cas := range cases {
		got := NewLongFromBytes(cas.bytes).value.Text(16)
		if got != cas.want {
			t.Errorf("NewLongFromBytes(%v).value.Text(16) = %v, want %v", cas.bytes, got, cas.want)
		}
	}
}

func TestLongReprStr(t *testing.T) {
	cases := []string{
		"0",
		"123",
		"-1",
		"3000",
		"42",
		fmt.Sprint(MaxInt),
		fmt.Sprint(MinInt),
		"10000000000000000",
	}
	for _, cas := range cases {
		i, _ := new(big.Int).SetString(cas, 0)
		o := NewLong(i).ToObject()
		repr, raised := o.typ.slots.Repr.Fn(nil, o)
		if raised != nil || toStrUnsafe(repr).Value() != cas+"L" {
			t.Errorf("(%sL).__repr__() = (%v, %v), want (%v, %v)", cas, toStrUnsafe(repr).Value(), raised, cas, nil)
		}
		str, raised := o.typ.slots.Str.Fn(nil, o)
		if raised != nil || toStrUnsafe(str).Value() != cas {
			t.Errorf("(%sL).__str__() = (%v, %v), want (%v, %v)", cas, toStrUnsafe(str).Value(), raised, cas, nil)
		}
	}
}

func TestLongNew(t *testing.T) {
	fooType := newTestClass("Foo", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__long__": newBuiltinFunction("__long__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return args[0], nil
		}).ToObject(),
	}))
	strictEqType := newTestClassStrictEq("StrictEq", LongType)
	newStrictEq := func(i *big.Int) *Object {
		l := Long{Object: Object{typ: strictEqType}}
		l.value.Set(i)
		return l.ToObject()
	}
	longSubType := newTestClass("LongSubType", []*Type{LongType}, newStringDict(map[string]*Object{}))
	cases := []invokeTestCase{
		{args: wrapArgs(LongType), want: NewLong(big.NewInt(0)).ToObject()},
		{args: wrapArgs(LongType, "123"), want: NewLong(big.NewInt(123)).ToObject()},
		{args: wrapArgs(LongType, "123L"), want: NewLong(big.NewInt(123)).ToObject()},
		{args: wrapArgs(LongType, "123l"), want: NewLong(big.NewInt(123)).ToObject()},
		{args: wrapArgs(LongType, " \t123L"), want: NewLong(big.NewInt(123)).ToObject()},
		{args: wrapArgs(LongType, "123L \t"), want: NewLong(big.NewInt(123)).ToObject()},
		{args: wrapArgs(LongType, "FF", 16), want: NewLong(big.NewInt(255)).ToObject()},
		{args: wrapArgs(LongType, "0xFFL", 16), want: NewLong(big.NewInt(255)).ToObject()},
		{args: wrapArgs(LongType, "0xE", 0), want: NewLong(big.NewInt(14)).ToObject()},
		{args: wrapArgs(LongType, "0b101L", 0), want: NewLong(big.NewInt(5)).ToObject()},
		{args: wrapArgs(LongType, "0o726", 0), want: NewLong(big.NewInt(470)).ToObject()},
		{args: wrapArgs(LongType, "102", 0), want: NewLong(big.NewInt(102)).ToObject()},
		{args: wrapArgs(LongType, 42), want: NewLong(big.NewInt(42)).ToObject()},
		{args: wrapArgs(LongType, -3.14), want: NewLong(big.NewInt(-3)).ToObject()},
		{args: wrapArgs(LongType, newObject(longSubType)), want: NewLong(big.NewInt(0)).ToObject()},
		{args: wrapArgs(strictEqType, big.NewInt(42)), want: newStrictEq(big.NewInt(42))},
		{args: wrapArgs(LongType, "0xff"), wantExc: mustCreateException(ValueErrorType, "invalid literal for long() with base 10: 0xff")},
		{args: wrapArgs(LongType, ""), wantExc: mustCreateException(ValueErrorType, "invalid literal for long() with base 10: ")},
		{args: wrapArgs(LongType, " "), wantExc: mustCreateException(ValueErrorType, "invalid literal for long() with base 10:  ")},
		{args: wrapArgs(FloatType), wantExc: mustCreateException(TypeErrorType, "long.__new__(float): float is not a subtype of long")},
		{args: wrapArgs(LongType, "asldkfj", 1), wantExc: mustCreateException(ValueErrorType, "long() base must be >= 2 and <= 36")},
		{args: wrapArgs(LongType, "asldkfj", 37), wantExc: mustCreateException(ValueErrorType, "long() base must be >= 2 and <= 36")},
		{args: wrapArgs(LongType, "@#%*(#", 36), wantExc: mustCreateException(ValueErrorType, "invalid literal for long() with base 36: @#%*(#")},
		{args: wrapArgs(LongType, "32059823095809238509238590835"), want: NewLong(func() *big.Int { i, _ := new(big.Int).SetString("32059823095809238509238590835", 0); return i }()).ToObject()},
		{args: wrapArgs(LongType, big.NewInt(3)), want: NewLong(big.NewInt(3)).ToObject()},
		{args: wrapArgs(LongType, NewInt(3)), want: NewLong(big.NewInt(3)).ToObject()},
		{args: wrapArgs(LongType, NewInt(3).ToObject()), want: NewLong(big.NewInt(3)).ToObject()},
		{args: wrapArgs(LongType, NewLong(big.NewInt(3))), want: NewLong(big.NewInt(3)).ToObject()},
		{args: wrapArgs(LongType, NewLong(big.NewInt(3)).ToObject()), want: NewLong(big.NewInt(3)).ToObject()},
		{args: wrapArgs(LongType, newObject(ObjectType)), wantExc: mustCreateException(TypeErrorType, "'__new__' requires a 'str' object but received a 'object'")},
		{args: wrapArgs(LongType, newObject(fooType)), wantExc: mustCreateException(TypeErrorType, "__long__ returned non-long (type Foo)")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(LongType, "__new__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestLongBinaryOps(t *testing.T) {
	cases := []struct {
		fun     binaryOpFunc
		v, w    interface{}
		want    *Object
		wantExc *BaseException
	}{
		{Add, -100, 50, NewLong(big.NewInt(-50)).ToObject(), nil},
		{Add, newObject(ObjectType), -100, nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for +: 'object' and 'long'")},
		{Add, MaxInt, 1, NewLong(new(big.Int).Add(maxIntBig, big.NewInt(1))).ToObject(), nil},
		{And, -100, 50, NewLong(big.NewInt(16)).ToObject(), nil},
		{And, MaxInt, MinInt, NewLong(big.NewInt(0)).ToObject(), nil},
		{And, newObject(ObjectType), -100, nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for &: 'object' and 'long'")},
		{Div, 7, 3, NewLong(big.NewInt(2)).ToObject(), nil},
		{Div, MaxInt, MinInt, NewLong(big.NewInt(-1)).ToObject(), nil},
		{Div, MinInt, MaxInt, NewLong(big.NewInt(-2)).ToObject(), nil},
		{Div, NewList().ToObject(), NewLong(big.NewInt(21)).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for /: 'list' and 'long'")},
		{Div, 1, 0, nil, mustCreateException(ZeroDivisionErrorType, "integer division or modulo by zero")},
		{Div, MinInt, -1, NewLong(new(big.Int).Neg(minIntBig)).ToObject(), nil},
		{DivMod, 7, 3, NewTuple2(NewLong(big.NewInt(2)).ToObject(), NewLong(big.NewInt(1)).ToObject()).ToObject(), nil},
		{DivMod, 3, -7, NewTuple2(NewLong(big.NewInt(-1)).ToObject(), NewLong(big.NewInt(-4)).ToObject()).ToObject(), nil},
		{DivMod, MaxInt, MinInt, NewTuple2(NewLong(big.NewInt(-1)).ToObject(), NewLong(big.NewInt(-1)).ToObject()).ToObject(), nil},
		{DivMod, MinInt, MaxInt, NewTuple2(NewLong(big.NewInt(-2)).ToObject(), NewLong(big.NewInt(MaxInt-1)).ToObject()).ToObject(), nil},
		{DivMod, MinInt, 1, NewTuple2(NewLong(big.NewInt(MinInt)).ToObject(), NewLong(big.NewInt(0)).ToObject()).ToObject(), nil},
		{DivMod, MinInt, -1, NewTuple2(NewLong(new(big.Int).Neg(minIntBig)).ToObject(), NewLong(big.NewInt(0)).ToObject()).ToObject(), nil},
		{DivMod, NewList().ToObject(), NewLong(big.NewInt(21)).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for divmod(): 'list' and 'long'")},
		{DivMod, 1, 0, nil, mustCreateException(ZeroDivisionErrorType, "integer division or modulo by zero")},
		{FloorDiv, 7, 3, NewLong(big.NewInt(2)).ToObject(), nil},
		{FloorDiv, MaxInt, MinInt, NewLong(big.NewInt(-1)).ToObject(), nil},
		{FloorDiv, MinInt, MaxInt, NewLong(big.NewInt(-2)).ToObject(), nil},
		{FloorDiv, NewList().ToObject(), NewLong(big.NewInt(21)).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for //: 'list' and 'long'")},
		{FloorDiv, 1, 0, nil, mustCreateException(ZeroDivisionErrorType, "integer division or modulo by zero")},
		{FloorDiv, MinInt, -1, NewLong(new(big.Int).Neg(minIntBig)).ToObject(), nil},
		{LShift, 2, 4, NewLong(big.NewInt(32)).ToObject(), nil},
		{LShift, 12, 10, NewLong(big.NewInt(12288)).ToObject(), nil},
		{LShift, 10, 100, NewLong(new(big.Int).Lsh(big.NewInt(10), 100)).ToObject(), nil},
		{LShift, 2, -5, nil, mustCreateException(ValueErrorType, "negative shift count")},
		{LShift, 4, NewFloat(3.14).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for <<: 'long' and 'float'")},
		{LShift, newObject(ObjectType), 4, nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for <<: 'object' and 'long'")},
		{RShift, 87, 3, NewLong(big.NewInt(10)).ToObject(), nil},
		{RShift, -101, 5, NewLong(big.NewInt(-4)).ToObject(), nil},
		{RShift, 12, NewInt(10).ToObject(), NewLong(big.NewInt(0)).ToObject(), nil},
		{RShift, 12, -10, nil, mustCreateException(ValueErrorType, "negative shift count")},
		{RShift, 4, NewFloat(3.14).ToObject(), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for >>: 'long' and 'float'")},
		{RShift, newObject(ObjectType), 4, nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for >>: 'object' and 'long'")},
		{RShift, 4, newObject(ObjectType), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for >>: 'long' and 'object'")},
		{Mod, 3, -7, NewLong(big.NewInt(-4)).ToObject(), nil},
		{Mod, MaxInt, MinInt, NewLong(big.NewInt(-1)).ToObject(), nil},
		{Mod, MinInt, MaxInt, NewLong(big.NewInt(int64(MaxInt) - 1)).ToObject(), nil},
		{Mod, None, 4, nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for %: 'NoneType' and 'long'")},
		{Mod, 10, 0, nil, mustCreateException(ZeroDivisionErrorType, "integer division or modulo by zero")},
		{Mod, MinInt, 1, NewLong(big.NewInt(0)).ToObject(), nil},
		{Mul, 1, 3, NewLong(big.NewInt(3)).ToObject(), nil},
		{Mul, newObject(ObjectType), 101, nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for *: 'object' and 'long'")},
		{Mul, int64(4294967295), int64(2147483649), NewLong(new(big.Int).Mul(big.NewInt(4294967295), big.NewInt(2147483649))).ToObject(), nil},
		{Or, -100, 50, NewLong(big.NewInt(-66)).ToObject(), nil},
		{Or, MaxInt, MinInt, NewLong(big.NewInt(-1)).ToObject(), nil},
		{Or, newObject(ObjectType), 100, nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for |: 'object' and 'long'")},
		{Pow, 2, 128, NewLong(big.NewInt(0).Exp(big.NewInt(2), big.NewInt(128), nil)).ToObject(), nil},
		{Pow, 2, -2, NewFloat(0.25).ToObject(), nil},
		{Pow, 2, newObject(ObjectType), nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for **: 'long' and 'object'")},
		{Sub, 22, 18, NewLong(big.NewInt(4)).ToObject(), nil},
		{Sub, IntType.ToObject(), 42, nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for -: 'type' and 'long'")},
		{Sub, MinInt, 1, NewLong(new(big.Int).Sub(minIntBig, big.NewInt(1))).ToObject(), nil},
		{Xor, -100, 50, NewLong(big.NewInt(-82)).ToObject(), nil},
		{Xor, MaxInt, MinInt, NewLong(big.NewInt(-1)).ToObject(), nil},
		{Xor, newObject(ObjectType), 100, nil, mustCreateException(TypeErrorType, "unsupported operand type(s) for ^: 'object' and 'long'")},
	}
	for _, cas := range cases {
		v := (*Object)(nil)
		switch casv := cas.v.(type) {
		case int:
			v = NewLong(big.NewInt(int64(casv))).ToObject()
		case int64:
			v = NewLong(big.NewInt(casv)).ToObject()
		case *big.Int:
			v = NewLong(casv).ToObject()
		case *Object:
			v = casv
		default:
			t.Errorf("invalid test case: %T", casv)
			continue
		}
		w := (*Object)(nil)
		switch casw := cas.w.(type) {
		case int:
			w = NewLong(big.NewInt(int64(casw))).ToObject()
		case int64:
			w = NewLong(big.NewInt(casw)).ToObject()
		case *big.Int:
			w = NewLong(casw).ToObject()
		case *Object:
			w = casw
		default:
			t.Errorf("invalid test case: %T", casw)
			continue
		}
		testCase := invokeTestCase{args: wrapArgs(v, w), want: cas.want, wantExc: cas.wantExc}
		if err := runInvokeTestCase(wrapFuncForTest(cas.fun), &testCase); err != "" {
			t.Error(err)
		}
	}
}

func TestLongCompare(t *testing.T) {
	// Equivalence classes of sample numbers, sorted from least to greatest, nil-separated
	googol, _ := big.NewFloat(1e100).Int(nil)
	numbers := []interface{}{
		math.Inf(-1), nil,
		-1e100, new(big.Int).Neg(googol), nil,
		new(big.Int).Lsh(big.NewInt(-1), 100), nil, // -2^100
		MinInt, nil,
		-306, -306.0, nil,
		1, big.NewInt(1), nil,
		309683958, big.NewInt(309683958), nil,
		MaxInt, nil,
		1e100, googol, nil,
		math.Inf(1), nil,
	}
	for i, v := range numbers {
		if v == nil {
			continue
		}
		want := compareAllResultEq
		for _, w := range numbers[i:] {
			if w == nil {
				// switching to a new equivalency class
				want = compareAllResultLT
				continue
			}
			cas := invokeTestCase{args: wrapArgs(v, w), want: want}
			if err := runInvokeTestCase(compareAll, &cas); err != "" {
				t.Error(err)
			}
		}
	}
}

func TestLongInvert(t *testing.T) {
	googol, _ := big.NewFloat(1e100).Int(nil)
	cases := []invokeTestCase{
		{args: wrapArgs(big.NewInt(2592)), want: NewLong(big.NewInt(-2593)).ToObject()},
		{args: wrapArgs(big.NewInt(0)), want: NewLong(big.NewInt(-1)).ToObject()},
		{args: wrapArgs(big.NewInt(-43)), want: NewLong(big.NewInt(42)).ToObject()},
		{args: wrapArgs(maxIntBig), want: NewLong(minIntBig).ToObject()},
		{args: wrapArgs(minIntBig), want: NewLong(maxIntBig).ToObject()},
		{args: wrapArgs(googol),
			want: NewLong(new(big.Int).Not(googol)).ToObject()},
		{args: wrapArgs(new(big.Int).Lsh(big.NewInt(-1), 100)),
			want: NewLong(new(big.Int).Not(new(big.Int).Lsh(big.NewInt(-1), 100))).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(LongType, "__invert__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestLongInt(t *testing.T) {
	googol, _ := big.NewFloat(1e100).Int(nil)
	cases := []invokeTestCase{
		{args: wrapArgs(big.NewInt(2592)), want: NewInt(2592).ToObject()},
		{args: wrapArgs(big.NewInt(0)), want: NewInt(0).ToObject()},
		{args: wrapArgs(big.NewInt(-43)), want: NewInt(-43).ToObject()},
		{args: wrapArgs(maxIntBig), want: NewInt(MaxInt).ToObject()},
		{args: wrapArgs(minIntBig), want: NewInt(MinInt).ToObject()},
		{args: wrapArgs(googol), want: NewLong(googol).ToObject()},
		{args: wrapArgs(new(big.Int).Lsh(big.NewInt(-1), 100)),
			want: NewLong(new(big.Int).Lsh(big.NewInt(-1), 100)).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(LongType, "__int__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestLongFloat(t *testing.T) {
	googol, _ := big.NewFloat(1e100).Int(nil)
	cases := []invokeTestCase{
		{args: wrapArgs(big.NewInt(2592)), want: NewFloat(2592).ToObject()},
		{args: wrapArgs(big.NewInt(0)), want: NewFloat(0).ToObject()},
		{args: wrapArgs(big.NewInt(-43)), want: NewFloat(-43).ToObject()},
		{args: wrapArgs(maxIntBig), want: NewFloat(float64(MaxInt)).ToObject()},
		{args: wrapArgs(minIntBig), want: NewFloat(float64(MinInt)).ToObject()},
		{args: wrapArgs(googol), want: NewFloat(1e100).ToObject()},
		{args: wrapArgs(new(big.Int).Lsh(big.NewInt(-1), 100)),
			want: NewFloat(-math.Pow(2, 100) + 1).ToObject()},
		{args: wrapArgs(new(big.Int).Lsh(big.NewInt(1), 10000)),
			wantExc: mustCreateException(OverflowErrorType, "long int too large to convert to float")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(LongType, "__float__", &cas); err != "" {
			t.Error(err)
		}
	}
}

// tests needed:
// ✓ arithmetic (long, long) -> long
// ✓   add
// ✓   sub
// ✓   mul
// ✓   div
// ✓   mod
// ✓ boolean logic (long, long) -> long
// ✓   and
// ✓   or
// ✓   xor
// ✓ shifts (long, int) -> long
// ✓   lsh
// ✓   rsh
// ✓ comparison (long, long) -> bool
// unary ops
//   hash    long -> int
//   nonzero long -> bool
// ✓   invert  long -> long
//   negate  long -> long   (this slot doesn't exist yet)
// ✓ int compatibility
// ✓   conversion
// ✓   comparison
// ✓ float compatibility
// ✓   conversion
// ✓   comparison
// ✓ parsing
// ✓   new
// ✓ formatting
// ✓   repr
// ✓   str
// native
//   istrue
//   native

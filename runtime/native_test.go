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
	"math/big"
	"reflect"
	"regexp"
	"testing"
)

func TestNativeMetaclassNew(t *testing.T) {
	var i int16
	intType := &newNativeType(reflect.TypeOf(i), IntType, NewDict()).Type
	fun := wrapFuncForTest(func(f *Frame, args ...*Object) *BaseException {
		newFunc, raised := GetAttr(f, intType.ToObject(), NewStr("new"), nil)
		if raised != nil {
			return raised
		}
		ret, raised := newFunc.Call(f, args, nil)
		if raised != nil {
			return raised
		}
		got, raised := ToNative(f, ret)
		if raised != nil {
			return raised
		}
		if got.Type() != reflect.TypeOf(&i) {
			t.Errorf("%v.new() returned a %s, want a *int16", intType, nativeTypeName(got.Type()))
		} else if p, ok := got.Interface().(*int16); !ok || p == nil || *p != 0 {
			t.Errorf("%v.new() returned %v, want &int16(0)", intType, got)
		}
		return nil
	})
	cases := []invokeTestCase{
		{want: None},
		{args: wrapArgs("abc"), wantExc: mustCreateException(TypeErrorType, "'new' of 'nativetype' requires 1 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestNativeFuncCall(t *testing.T) {
	cases := []struct {
		fun interface{}
		invokeTestCase
	}{
		{func() {}, invokeTestCase{want: None}},
		{func() float32 { return 2.0 }, invokeTestCase{want: NewFloat(2.0).ToObject()}},
		{func(s string) string { return s }, invokeTestCase{args: wrapArgs("foo"), want: NewStr("foo").ToObject()}},
		{func() (int, string) { return 42, "bar" }, invokeTestCase{want: newTestTuple(42, "bar").ToObject()}},
		{func(s ...string) int { return len(s) }, invokeTestCase{args: wrapArgs("foo", "bar"), want: NewInt(2).ToObject()}},
		{func() {}, invokeTestCase{args: wrapArgs(3.14), wantExc: mustCreateException(TypeErrorType, "native function takes 0 arguments, (1 given)")}},
		{func(int, ...string) {}, invokeTestCase{wantExc: mustCreateException(TypeErrorType, "native function takes at least 1 arguments, (0 given)")}},
	}
	for _, cas := range cases {
		n := &native{Object{typ: nativeFuncType}, reflect.ValueOf(cas.fun)}
		if err := runInvokeTestCase(n.ToObject(), &cas.invokeTestCase); err != "" {
			t.Error(err)
		}
	}
}

func TestNativeFuncName(t *testing.T) {
	re := regexp.MustCompile(`(\w+\.)*\w+$`)
	fun := wrapFuncForTest(func(f *Frame, o *Object) (string, *BaseException) {
		desc, raised := GetItem(f, nativeFuncType.Dict().ToObject(), NewStr("__name__").ToObject())
		if raised != nil {
			return "", raised
		}
		get, raised := GetAttr(f, desc, NewStr("__get__"), nil)
		if raised != nil {
			return "", raised
		}
		name, raised := get.Call(f, wrapArgs(o, nativeFuncType), nil)
		if raised != nil {
			return "", raised
		}
		if raised := Assert(f, GetBool(name.isInstance(StrType)).ToObject(), nil); raised != nil {
			return "", raised
		}
		return re.FindString(toStrUnsafe(name).Value()), nil
	})
	cases := []invokeTestCase{
		{args: wrapArgs(TestNativeFuncName), want: NewStr("grumpy.TestNativeFuncName").ToObject()},
		{args: wrapArgs(None), wantExc: mustCreateException(TypeErrorType, "'_get_name' requires a 'func' object but received a 'NoneType'")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestNativeFuncStrRepr(t *testing.T) {
	cases := []struct {
		args        Args
		wantPattern string
	}{
		{wrapArgs(TestNativeFuncStrRepr), `<func\(\*T\) .*grumpy\.TestNativeFuncStrRepr at 0x[a-f0-9]+>`},
		{wrapArgs(func() {}), `<func\(\) .*grumpy\.TestNativeFuncStrRepr\.\w+ at 0x[a-f0-9]+>`},
		{wrapArgs(Repr), `<func\(\*Frame, \*Object\) .*grumpy\.Repr at 0x[a-f0-9]+>`},
	}
	for _, cas := range cases {
		re := regexp.MustCompile(cas.wantPattern)
		fun := wrapFuncForTest(func(f *Frame, o *Object) *BaseException {
			s, raised := ToStr(f, o)
			if raised != nil {
				return raised
			}
			if !re.MatchString(s.Value()) {
				t.Errorf("str(%v) = %v, want %v", o, s, re)
			}
			s, raised = Repr(f, o)
			if raised != nil {
				return raised
			}
			if !re.MatchString(s.Value()) {
				t.Errorf("repr(%v) = %v, want %v", o, s, re)
			}
			return nil
		})
		if err := runInvokeTestCase(fun, &invokeTestCase{args: cas.args, want: None}); err != "" {
			t.Error(err)
		}
	}
}

func TestNativeSliceIter(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, slice interface{}) (*Object, *BaseException) {
		o, raised := WrapNative(f, reflect.ValueOf(slice))
		if raised != nil {
			return nil, raised
		}
		return TupleType.Call(f, []*Object{o}, nil)
	})
	o := newObject(ObjectType)
	cases := []invokeTestCase{
		{args: wrapArgs([]int{}), want: NewTuple().ToObject()},
		{args: wrapArgs([]string{"foo", "bar"}), want: newTestTuple("foo", "bar").ToObject()},
		{args: wrapArgs([]*Object{True.ToObject(), o}), want: newTestTuple(true, o).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestSliceIteratorIter(t *testing.T) {
	iter := newSliceIterator(reflect.ValueOf([]*Object{}))
	cas := &invokeTestCase{args: wrapArgs(iter), want: iter}
	if err := runInvokeMethodTestCase(sliceIteratorType, "__iter__", cas); err != "" {
		t.Error(err)
	}
}

func TestWrapNative(t *testing.T) {
	o := newObject(ObjectType)
	d := NewDict()
	i := 0
	n := &native{Object{typ: nativeType}, reflect.ValueOf(&i)}
	cases := []struct {
		value   interface{}
		want    *Object
		wantExc *BaseException
	}{
		{true, True.ToObject(), nil},
		{True, True.ToObject(), nil},
		{123, NewInt(123).ToObject(), nil},
		{int8(10), NewInt(10).ToObject(), nil},
		{float32(0.5), NewFloat(0.5).ToObject(), nil},
		{NewFloat(3.14), NewFloat(3.14).ToObject(), nil},
		{uint(MaxInt), NewInt(MaxInt).ToObject(), nil},
		{"foobar", NewStr("foobar").ToObject(), nil},
		{NewStr("foo"), NewStr("foo").ToObject(), nil},
		{uint64(MaxInt) + 100, NewLong(new(big.Int).SetUint64(uint64(MaxInt) + 100)).ToObject(), nil},
		{o, o, nil},
		{d, d.ToObject(), nil},
		{(*Object)(nil), None, nil},
		{uintptr(123), NewInt(123).ToObject(), nil},
		{n, n.ToObject(), nil},
		{(chan int)(nil), None, nil},
		{[]rune("hola"), NewUnicode("hola").ToObject(), nil},
		{big.NewInt(12345), NewLong(big.NewInt(12345)).ToObject(), nil},
		{*big.NewInt(12345), NewLong(big.NewInt(12345)).ToObject(), nil},
	}
	for _, cas := range cases {
		fun := wrapFuncForTest(func(f *Frame) (*Object, *BaseException) {
			return WrapNative(f, reflect.ValueOf(cas.value))
		})
		testCase := invokeTestCase{want: cas.want, wantExc: cas.wantExc}
		if err := runInvokeTestCase(fun, &testCase); err != "" {
			t.Error(err)
		}
	}
}

func TestWrapNativeFunc(t *testing.T) {
	foo := func() int { return 42 }
	wrappedFoo := mustNotRaise(WrapNative(NewRootFrame(), reflect.ValueOf(foo)))
	if err := runInvokeTestCase(wrappedFoo, &invokeTestCase{want: NewInt(42).ToObject()}); err != "" {
		t.Error(err)
	}
}

func TestWrapNativeInterface(t *testing.T) {
	// This seems to be the simplest way to get a reflect.Value that has
	// Interface kind.
	iVal := reflect.ValueOf(func() error { return errors.New("foo") }).Call(nil)[0]
	if iVal.Kind() != reflect.Interface {
		t.Fatalf("iVal.Kind() = %v, want interface", iVal.Kind())
	}
	o := mustNotRaise(WrapNative(NewRootFrame(), iVal))
	cas := &invokeTestCase{args: wrapArgs(o), want: NewStr("foo").ToObject()}
	if err := runInvokeMethodTestCase(o.typ, "Error", cas); err != "" {
		t.Error(err)
	}
	// Also test the nil interface case.
	nilVal := reflect.ValueOf(func() error { return nil }).Call(nil)[0]
	if nilVal.Kind() != reflect.Interface {
		t.Fatalf("nilVal.Kind() = %v, want interface", nilVal.Kind())
	}
	if o := mustNotRaise(WrapNative(NewRootFrame(), nilVal)); o != None {
		t.Errorf("WrapNative(%v) = %v, want None", nilVal, o)
	}
}

func TestWrapNativeOpaque(t *testing.T) {
	type fooStruct struct{}
	foo := &fooStruct{}
	fooVal := reflect.ValueOf(foo)
	fun := wrapFuncForTest(func(f *Frame) *BaseException {
		o, raised := WrapNative(f, fooVal)
		if raised != nil {
			return raised
		}
		if !o.isInstance(nativeType) {
			t.Errorf("WrapNative(%v) = %v, want %v", fooVal, o, foo)
		} else if v := toNativeUnsafe(o).value; v.Type() != reflect.TypeOf(foo) {
			t.Errorf("WrapNative(%v) = %v, want %v", fooVal, v, foo)
		} else if got := v.Interface().(*fooStruct); got != foo {
			t.Errorf("WrapNative(%v) = %v, want %v", fooVal, got, foo)
		}
		return nil
	})
	if err := runInvokeTestCase(fun, &invokeTestCase{want: None}); err != "" {
		t.Error(err)
	}
}

func TestGetNativeTypeCaches(t *testing.T) {
	foo := []struct{}{}
	typ := getNativeType(reflect.TypeOf(foo))
	if got := getNativeType(reflect.TypeOf(foo)); got != typ {
		t.Errorf("getNativeType(foo) = %v, want %v", got, typ)
	}
}

func TestGetNativeTypeFunc(t *testing.T) {
	if typ := getNativeType(reflect.TypeOf(func() {})); !typ.isSubclass(nativeFuncType) {
		t.Errorf("getNativeType(func() {}) = %v, want a subclass of func", typ)
	} else if name := typ.Name(); name != "func()" {
		t.Errorf(`%v.__name__ == %q, want "func()"`, typ, name)
	}
}

type testNativeType struct {
	data int64
}

func (n *testNativeType) Int64() int64 {
	return n.data
}

func TestGetNativeTypeMethods(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, o *Object) (*Object, *BaseException) {
		if raised := Assert(f, GetBool(o.isInstance(nativeType)).ToObject(), nil); raised != nil {
			return nil, raised
		}
		int64Method, raised := GetAttr(f, o.Type().ToObject(), NewStr("Int64"), nil)
		if raised != nil {
			return nil, raised
		}
		return int64Method.Call(f, []*Object{o}, nil)
	})
	cas := invokeTestCase{args: wrapArgs(&testNativeType{12}), want: NewInt(12).ToObject()}
	if err := runInvokeTestCase(fun, &cas); err != "" {
		t.Error(err)
	}
}

func TestGetNativeTypeSlice(t *testing.T) {
	if typ := getNativeType(reflect.TypeOf([]int{})); !typ.isSubclass(nativeSliceType) {
		t.Errorf("getNativeType([]int) = %v, want a subclass of slice", typ)
	} else if name := typ.Name(); name != "[]int" {
		t.Errorf(`%v.__name__ == %q, want "func()"`, typ, name)
	}
}

func TestGetNativeTypeTypedefs(t *testing.T) {
	type testBool bool
	type testInt int
	type testFloat float32
	type testString string
	cases := []struct {
		rtype reflect.Type
		super *Type
	}{
		{reflect.TypeOf(testBool(true)), BoolType},
		{reflect.TypeOf(testFloat(3.14)), FloatType},
		{reflect.TypeOf(testInt(42)), IntType},
		{reflect.TypeOf(testString("foo")), StrType},
	}
	for _, cas := range cases {
		if typ := getNativeType(cas.rtype); typ == cas.super || !typ.isSubclass(cas.super) {
			t.Errorf("getNativeType(%v) = %v, want a subclass of %v", cas.rtype, typ, cas.super)
		}
	}
}

func TestGetNativeTypeBigInts(t *testing.T) {
	cases := []struct {
		rtype reflect.Type
		typ   *Type
	}{
		{reflect.TypeOf(big.Int{}), LongType},
		{reflect.TypeOf((*big.Int)(nil)), LongType},
	}
	for _, cas := range cases {
		if typ := getNativeType(cas.rtype); typ != cas.typ {
			t.Errorf("getNativeType(%v) = %v, want %v", cas.rtype, typ, cas.typ)
		}
	}
}

func TestMaybeConvertValue(t *testing.T) {
	type fooStruct struct{}
	foo := &fooStruct{}
	fooNative := &native{Object{typ: nativeType}, reflect.ValueOf(&foo)}
	cases := []struct {
		o             *Object
		expectedRType reflect.Type
		want          interface{}
		wantExc       *BaseException
	}{
		{NewInt(42).ToObject(), reflect.TypeOf(int(0)), 42, nil},
		{NewFloat(0.5).ToObject(), reflect.TypeOf(float32(0)), float32(0.5), nil},
		{fooNative.ToObject(), reflect.TypeOf(&fooStruct{}), foo, nil},
		{None, reflect.TypeOf((*int)(nil)), (*int)(nil), nil},
		{None, reflect.TypeOf(""), nil, mustCreateException(TypeErrorType, "cannot convert None to string")},
	}
	for _, cas := range cases {
		fun := wrapFuncForTest(func(f *Frame) *BaseException {
			got, raised := maybeConvertValue(f, cas.o, cas.expectedRType)
			if raised != nil {
				return raised
			}
			if !got.IsValid() || !reflect.DeepEqual(got.Interface(), cas.want) {
				t.Errorf("maybeConvertValue(%v, %v) = %v, want %v", cas.o, nativeTypeName(cas.expectedRType), got, cas.want)
			}
			return nil
		})
		testCase := invokeTestCase{}
		if cas.wantExc != nil {
			testCase.wantExc = cas.wantExc
		} else {
			testCase.want = None
		}
		if err := runInvokeTestCase(fun, &testCase); err != "" {
			t.Error(err)
		}
	}
}

func TestNativeTypeName(t *testing.T) {
	type fooStruct struct{}
	cases := []struct {
		rtype reflect.Type
		want  string
	}{
		{reflect.TypeOf([4]int{}), "[4]int"},
		{reflect.TypeOf(make(chan *string)), "chan *string"},
		{reflect.TypeOf(func() {}), "func()"},
		{reflect.TypeOf(func(int, string) {}), "func(int, string)"},
		{reflect.TypeOf(func() int { return 0 }), "func() int"},
		{reflect.TypeOf(func() (int, float32) { return 0, 0.0 }), "func() (int, float32)"},
		{reflect.TypeOf(map[int]fooStruct{}), "map[int]fooStruct"},
		{reflect.TypeOf(&fooStruct{}), "*fooStruct"},
		{reflect.TypeOf([]byte{}), "[]uint8"},
		{reflect.TypeOf(struct{}{}), "anonymous struct"},
	}
	for _, cas := range cases {
		if got := nativeTypeName(cas.rtype); got != cas.want {
			t.Errorf("nativeTypeName(%v) = %q, want %q", cas.rtype, got, cas.want)
		}
	}
}

func wrapArgs(elems ...interface{}) Args {
	f := NewRootFrame()
	argc := len(elems)
	result := make(Args, argc, argc)
	var raised *BaseException
	for i, elem := range elems {
		if result[i], raised = WrapNative(f, reflect.ValueOf(elem)); raised != nil {
			panic(raised)
		}
	}
	return result
}

func wrapKWArgs(elems ...interface{}) KWArgs {
	if len(elems)%2 != 0 {
		panic("invalid kwargs")
	}
	numItems := len(elems) / 2
	kwargs := make(KWArgs, numItems, numItems)
	f := NewRootFrame()
	for i := 0; i < numItems; i++ {
		kwargs[i].Name = elems[i*2].(string)
		kwargs[i].Value = mustNotRaise(WrapNative(f, reflect.ValueOf(elems[i*2+1])))
	}
	return kwargs
}

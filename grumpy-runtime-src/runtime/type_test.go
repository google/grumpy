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
	"reflect"
	"strings"
	"testing"
)

func TestNewClass(t *testing.T) {
	type strBasisStruct struct{ Str }
	strBasisStructFunc := func(o *Object) *strBasisStruct { return (*strBasisStruct)(o.toPointer()) }
	fooType := newBasisType("Foo", reflect.TypeOf(strBasisStruct{}), strBasisStructFunc, StrType)
	defer delete(basisTypes, fooType.basis)
	fooType.setDict(NewDict())
	prepareType(fooType)
	cases := []struct {
		wantBasis reflect.Type
		invokeTestCase
	}{
		{objectBasis, invokeTestCase{args: wrapArgs([]*Type{ObjectType}), want: None}},
		{fooType.basis, invokeTestCase{args: wrapArgs([]*Type{fooType, StrType}), want: None}},
		{fooType.basis, invokeTestCase{args: wrapArgs([]*Type{fooType, StrType, ObjectType}), want: None}},
		{nil, invokeTestCase{args: wrapArgs([]*Type{}), wantExc: mustCreateException(TypeErrorType, "class must have base classes")}},
		{nil, invokeTestCase{args: wrapArgs([]*Type{BoolType, ObjectType}), wantExc: mustCreateException(TypeErrorType, "type 'bool' is not an acceptable base type")}},
		{nil, invokeTestCase{args: wrapArgs([]*Type{IntType, StrType}), wantExc: mustCreateException(TypeErrorType, "class layout error")}},
		{nil, invokeTestCase{args: wrapArgs([]*Type{StrType, fooType}), wantExc: mustCreateException(TypeErrorType, "mro error for: Foo")}},
	}
	for _, cas := range cases {
		fun := wrapFuncForTest(func(f *Frame, bases []*Type) *BaseException {
			cls, raised := newClass(f, TypeType, "Foo", bases, NewDict())
			if raised != nil {
				return raised
			}
			if cls.basis != cas.wantBasis {
				t.Errorf("type('Foo', %v, {}) had basis %v, want %v", bases, cls.basis, cas.wantBasis)
			}
			return nil
		})
		if err := runInvokeTestCase(fun, &cas.invokeTestCase); err != "" {
			t.Error(err)
		}
	}
}

func TestNewBasisType(t *testing.T) {
	type basisStruct struct{ Object }
	basisStructFunc := func(o *Object) *basisStruct { return (*basisStruct)(o.toPointer()) }
	basis := reflect.TypeOf(basisStruct{})
	typ := newBasisType("Foo", basis, basisStructFunc, ObjectType)
	defer delete(basisTypes, basis)
	if typ.Type() != TypeType {
		t.Errorf("got %q, want a type", typ.Type().Name())
	}
	if typ.Dict() != nil {
		t.Error("type's dict was expected to be nil")
	}
	wantBases := []*Type{ObjectType}
	if !reflect.DeepEqual(wantBases, typ.bases) {
		t.Errorf("typ.bases = %v, want %v, ", typ.bases, wantBases)
	}
	if typ.mro != nil {
		t.Errorf("type's mro expected to be nil, got %v", typ.mro)
	}
	if name := typ.Name(); name != "Foo" {
		t.Errorf(`Foo.Name() = %q, want "Foo"`, name)
	}
	foo := (*basisStruct)(newObject(typ).toPointer())
	if typ.slots.Basis == nil {
		t.Error("type's Basis slot is nil")
	} else if got := typ.slots.Basis.Fn(&foo.Object); got.Type() != basis || got.Addr().Interface().(*basisStruct) != foo {
		t.Errorf("Foo.__basis__(%v) = %v, want %v", &foo.Object, got, foo)
	}
}

func TestNewSimpleType(t *testing.T) {
	got := newSimpleType("Foo", ObjectType)
	if got.Object.typ != TypeType {
		t.Errorf(`newSimpleType got %q, want "type"`, got.Type().Name())
	}
	if got.basis != objectBasis {
		t.Errorf("newSimpleType result got basis %v, want %v", got.basis, objectBasis)
	}
	wantBases := []*Type{ObjectType}
	if !reflect.DeepEqual(got.bases, wantBases) {
		t.Errorf("newSimpleType got bases %v, want %v", got.bases, wantBases)
	}
	if name := got.Name(); name != "Foo" {
		t.Errorf(`Foo.Name() = %q, want "Foo"`, name)
	}
}

func TestInvalidBasisType(t *testing.T) {
	type intFieldStruct struct{ int }
	type emptyStruct struct{}
	type objectBasisStruct struct{ Object }
	oldLogFatal := logFatal
	defer func() { logFatal = oldLogFatal }()
	logFatal = func(msg string) { panic(msg) }
	cases := []struct {
		basis     reflect.Type
		basisFunc interface{}
		wantMsg   string
	}{
		{objectBasis, objectBasisFunc, "basis already exists"},
		{reflect.TypeOf(int(0)), objectBasisFunc, "basis must be a struct"},
		{reflect.TypeOf(emptyStruct{}), objectBasisFunc, "1st field of basis must be base type's basis"},
		{reflect.TypeOf(intFieldStruct{}), objectBasisFunc, "1st field of basis must be base type's basis not: int"},
		{reflect.TypeOf(objectBasisStruct{}), objectBasisFunc, "expected basis func of type func(*Object) *objectBasisStruct"},
	}
	for _, cas := range cases {
		func() {
			defer func() {
				if msg, ok := recover().(string); !ok || !strings.Contains(msg, cas.wantMsg) {
					t.Errorf("logFatal() called with %q, want error like %q", msg, cas.wantMsg)
				}
			}()
			newBasisType("Foo", cas.basis, cas.basisFunc, ObjectType)
		}()
	}
}

func TestPrepareType(t *testing.T) {
	type objectBasisStruct struct{ Object }
	objectBasisStructFunc := func(o *Object) *objectBasisStruct { return (*objectBasisStruct)(o.toPointer()) }
	type strBasisStruct struct{ Str }
	strBasisStructFunc := func(o *Object) *strBasisStruct { return (*strBasisStruct)(o.toPointer()) }
	cases := []struct {
		basis     reflect.Type
		basisFunc interface{}
		base      *Type
		wantMro   []*Type
	}{
		{reflect.TypeOf(objectBasisStruct{}), objectBasisStructFunc, ObjectType, []*Type{nil, ObjectType}},
		{reflect.TypeOf(strBasisStruct{}), strBasisStructFunc, StrType, []*Type{nil, StrType, BaseStringType, ObjectType}},
	}
	for _, cas := range cases {
		typ := newBasisType("Foo", cas.basis, cas.basisFunc, cas.base)
		defer delete(basisTypes, cas.basis)
		typ.setDict(NewDict())
		prepareType(typ)
		cas.wantMro[0] = typ
		if !reflect.DeepEqual(typ.mro, cas.wantMro) {
			t.Errorf("typ.mro = %v, want %v", typ.mro, cas.wantMro)
		}
	}
}

func makeTestType(name string, bases ...*Type) *Type {
	return newType(TypeType, name, nil, bases, NewDict())
}

func TestMroCalc(t *testing.T) {
	fooType := makeTestType("Foo", ObjectType)
	barType := makeTestType("Bar", StrType, fooType)
	bazType := makeTestType("Baz", fooType, StrType)
	// Boo has an inconsistent hierarchy since it's not possible to order
	// mro such that StrType is before fooType and fooType is also before
	// StrType.
	booType := makeTestType("Boo", barType, bazType)
	cases := []struct {
		typ     *Type
		wantMro []*Type
	}{
		{fooType, []*Type{fooType, ObjectType}},
		{barType, []*Type{barType, StrType, BaseStringType, fooType, ObjectType}},
		{bazType, []*Type{bazType, fooType, StrType, BaseStringType, ObjectType}},
		{booType, nil},
	}
	for _, cas := range cases {
		cas.typ.mro = mroCalc(cas.typ)
		if !reflect.DeepEqual(cas.wantMro, cas.typ.mro) {
			t.Errorf("%s.mro = %v, want %v", cas.typ.Name(), cas.typ.mro, cas.wantMro)
		}
	}
}

func TestTypeIsSubclass(t *testing.T) {
	fooType := makeTestType("Foo", ObjectType)
	prepareType(fooType)
	barType := makeTestType("Bar", StrType, fooType)
	prepareType(barType)
	cases := []struct {
		typ   *Type
		super *Type
		want  bool
	}{
		{fooType, ObjectType, true},
		{fooType, StrType, false},
		{barType, ObjectType, true},
		{barType, fooType, true},
		{barType, StrType, true},
		{barType, TypeType, false},
	}
	for _, cas := range cases {
		got := cas.typ.isSubclass(cas.super)
		if got != cas.want {
			t.Errorf("%s.isSubclass(%s) = %v, want %v", cas.typ.Name(), cas.super.Name(), got, cas.want)
		}
	}
}

func TestTypeCall(t *testing.T) {
	fooType := makeTestType("Foo")
	prepareType(fooType)
	emptyExc := toBaseExceptionUnsafe(newObject(ExceptionType))
	emptyExc.args = NewTuple()
	cases := []invokeTestCase{
		{wantExc: mustCreateException(TypeErrorType, "unbound method __call__() must be called with type instance as first argument (got nothing instead)")},
		{args: wrapArgs(42), wantExc: mustCreateException(TypeErrorType, "unbound method __call__() must be called with type instance as first argument (got int instance instead)")},
		{args: wrapArgs(fooType), wantExc: mustCreateException(TypeErrorType, "type Foo has no __new__")},
		{args: wrapArgs(IntType), want: NewInt(0).ToObject()},
		{args: wrapArgs(ExceptionType, "blah"), want: mustCreateException(ExceptionType, "blah").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(TypeType, "__call__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestNewWithSubclass(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(StrType, "abc"), want: None},
		{args: wrapArgs(IntType, 3), want: None},
		{args: wrapArgs(UnicodeType, "abc"), want: None},
	}
	simpleRepr := newBuiltinFunction("__repr__", func(_ *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		return NewStr(fmt.Sprintf("%s object", args[0].typ.Name())).ToObject(), nil
	}).ToObject()
	constantFunc := func(name string, value *Object) *Object {
		return newBuiltinFunction(name, func(_ *Frame, _ Args, _ KWArgs) (*Object, *BaseException) {
			return value, nil
		}).ToObject()
	}
	for _, cas := range cases {
		fun := wrapFuncForTest(func(f *Frame, basisType *Type, o *Object) *BaseException {
			subclassTypeName := "SubclassOf" + basisType.Name()
			// Create a subclass of the basis type.
			subclassType := newTestClass(subclassTypeName, []*Type{basisType}, newStringDict(map[string]*Object{
				"__repr__": simpleRepr,
			}))
			subclassObject, raised := subclassType.Call(f, Args{o}, nil)
			if raised != nil {
				return raised
			}
			slotName := "__" + basisType.Name() + "__"
			fooType := newTestClass("FooFor"+basisType.Name(), []*Type{ObjectType}, newStringDict(map[string]*Object{
				slotName:   constantFunc(slotName, subclassObject),
				"__repr__": simpleRepr,
			}))
			foo := newObject(fooType)
			// Test that <basistype>(subclassObject) returns an object of the basis type, not the subclass.
			got, raised := basisType.Call(f, Args{subclassObject}, nil)
			if raised != nil {
				return raised
			}
			if got.typ != basisType {
				t.Errorf("type(%s(%s)) = %s, want %s", basisType.Name(), subclassObject, got.typ.Name(), basisType.Name())
			}
			// Test that subclass objects returned from __<typename>__ slots are left intact.
			got, raised = basisType.Call(f, Args{foo}, nil)
			if raised != nil {
				return raised
			}
			if got.typ != subclassType {
				t.Errorf("type(%s(%s)) = %s, want %s", basisType.Name(), foo, got.typ.Name(), basisType.Name())
			}
			return nil
		})
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestTypeGetAttribute(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, o *Object, name *Str) (*Object, *BaseException) {
		return GetAttr(f, o, name, nil)
	})
	// class Getter(object):
	//   def __get__(self, *args):
	//     return "got getter"
	getterType := newTestClass("Getter", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__get__": newBuiltinFunction("__get__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return NewStr("got getter").ToObject(), nil
		}).ToObject(),
	}))
	getter := newObject(getterType)
	// class Setter(object):
	//   def __get__(self, *args):
	//     return "got setter"
	//   def __set__(self, *args):
	//     pass
	setterType := newTestClass("Setter", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"__get__": newBuiltinFunction("__get__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return NewStr("got setter").ToObject(), nil
		}).ToObject(),
		"__set__": newBuiltinFunction("__set__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return None, nil
		}).ToObject(),
	}))
	setter := newObject(setterType)
	// class Foo(object):
	//   pass
	fooType := newTestClass("Foo", []*Type{ObjectType}, newStringDict(map[string]*Object{
		"bar":       NewInt(42).ToObject(),
		"baz":       NewStr("Foo's baz").ToObject(),
		"foogetter": getter,
	}))
	// class BarMeta(type):
	//   pass
	barMetaType := newTestClass("BarMeta", []*Type{TypeType}, newStringDict(map[string]*Object{
		"bar":           NewStr("BarMeta's bar").ToObject(),
		"boo":           NewInt(123).ToObject(),
		"barmetagetter": getter,
		"barmetasetter": setter,
	}))
	// class Bar(Foo):
	//   __metaclass__ = BarMeta
	// bar = Bar()
	barType := &Type{Object: Object{typ: barMetaType}, name: "Bar", basis: fooType.basis, bases: []*Type{fooType}}
	barType.setDict(newTestDict("bar", "Bar's bar", "foo", 101, "barsetter", setter, "barmetasetter", "NOT setter"))
	bar := newObject(barType)
	prepareType(barType)
	cases := []invokeTestCase{
		{args: wrapArgs(fooType, "bar"), want: NewInt(42).ToObject()},
		{args: wrapArgs(fooType, "baz"), want: NewStr("Foo's baz").ToObject()},
		{args: wrapArgs(barMetaType, "barmetagetter"), want: NewStr("got getter").ToObject()},
		{args: wrapArgs(barType, "bar"), want: NewStr("Bar's bar").ToObject()},
		{args: wrapArgs(barType, "baz"), want: NewStr("Foo's baz").ToObject()},
		{args: wrapArgs(barType, "foo"), want: NewInt(101).ToObject()},
		{args: wrapArgs(barType, "barmetagetter"), want: NewStr("got getter").ToObject()},
		{args: wrapArgs(barType, "barmetasetter"), want: NewStr("got setter").ToObject()},
		{args: wrapArgs(barType, "boo"), want: NewInt(123).ToObject()},
		{args: wrapArgs(bar, "boo"), wantExc: mustCreateException(AttributeErrorType, "'Bar' object has no attribute 'boo'")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestTypeName(t *testing.T) {
	fooType := newTestClass("Foo", []*Type{ObjectType}, NewDict())
	fun := wrapFuncForTest(func(f *Frame, t *Type) (*Object, *BaseException) {
		return GetAttr(f, t.ToObject(), internedName, nil)
	})
	cas := invokeTestCase{args: wrapArgs(fooType), want: NewStr("Foo").ToObject()}
	if err := runInvokeTestCase(fun, &cas); err != "" {
		t.Error(err)
	}
}

func TestTypeNew(t *testing.T) {
	fooMetaType := newTestClass("FooMeta", []*Type{TypeType}, NewDict())
	fooType, raised := newClass(NewRootFrame(), fooMetaType, "Foo", []*Type{ObjectType}, NewDict())
	if raised != nil {
		panic(raised)
	}
	barMetaType := newTestClass("BarMeta", []*Type{TypeType}, NewDict())
	barType, raised := newClass(NewRootFrame(), barMetaType, "Bar", []*Type{ObjectType}, NewDict())
	if raised != nil {
		panic(raised)
	}
	var bazMetaType *Type
	bazMetaType = newTestClass("BazMeta", []*Type{barMetaType}, newStringDict(map[string]*Object{
		// Returns true if type(lhs) == rhs.
		"__eq__": newBuiltinFunction("__eq__", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
			if raised := checkMethodArgs(f, "__eq__", args, TypeType, TypeType); raised != nil {
				return nil, raised
			}
			return GetBool(args[0].typ == toTypeUnsafe(args[1])).ToObject(), nil
		}).ToObject(),
	}))
	bazType, raised := newClass(NewRootFrame(), bazMetaType, "Baz", []*Type{ObjectType}, NewDict())
	if raised != nil {
		panic(raised)
	}
	cases := []invokeTestCase{
		{wantExc: mustCreateException(TypeErrorType, "'__new__' requires 1 arguments")},
		{args: wrapArgs(TypeType), wantExc: mustCreateException(TypeErrorType, "type() takes 1 or 3 arguments")},
		{args: wrapArgs(TypeType, "foo", newTestTuple(false), NewDict()), wantExc: mustCreateException(TypeErrorType, "not a valid base class: False")},
		{args: wrapArgs(TypeType, None), want: NoneType.ToObject()},
		{args: wrapArgs(fooMetaType, "Qux", newTestTuple(fooType, barType), NewDict()), wantExc: mustCreateException(TypeErrorType, "metaclass conflict: the metaclass of a derived class must a be a (non-strict) subclass of the metaclasses of all its bases")},
		// Test that the metaclass of the result is the most derived
		// metaclass of the bases. In this case that should be
		// bazMetaType so pass bazMetaType to be compared by the __eq__
		// operator defined above.
		{args: wrapArgs(barMetaType, "Qux", newTestTuple(barType, bazType), NewDict()), want: bazMetaType.ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(TypeType, "__new__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestTypeNewResult(t *testing.T) {
	fooType := makeTestType("Foo", ObjectType)
	prepareType(fooType)
	fun := wrapFuncForTest(func(f *Frame) *BaseException {
		newFunc, raised := GetAttr(f, TypeType.ToObject(), NewStr("__new__"), nil)
		if raised != nil {
			return raised
		}
		ret, raised := newFunc.Call(f, wrapArgs(TypeType, "Bar", newTestTuple(fooType, StrType), NewDict()), nil)
		if raised != nil {
			return raised
		}
		if !ret.isInstance(TypeType) {
			t.Errorf("type('Bar', (Foo, str), {}) = %v, want type instance", ret)
		} else if typ := toTypeUnsafe(ret); typ.basis != StrType.basis {
			t.Errorf("type('Bar', (Foo, str), {}) basis is %v, want %v", typ.basis, StrType.basis)
		} else if wantMro := []*Type{typ, fooType, StrType, BaseStringType, ObjectType}; !reflect.DeepEqual(typ.mro, wantMro) {
			t.Errorf("type('Bar', (Foo, str), {}).__mro__ = %v, want %v", typ.mro, wantMro)
		}
		return nil
	})
	if err := runInvokeTestCase(fun, &invokeTestCase{want: None}); err != "" {
		t.Error(err)
	}
}

func TestTypeStrRepr(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, o *Object) (*Tuple, *BaseException) {
		str, raised := ToStr(f, o)
		if raised != nil {
			return nil, raised
		}
		repr, raised := Repr(f, o)
		if raised != nil {
			return nil, raised
		}
		return newTestTuple(str, repr), nil
	})
	fooType := newTestClass("Foo", []*Type{ObjectType}, newTestDict("__module__", "foo.bar"))
	cases := []invokeTestCase{
		{args: wrapArgs(TypeErrorType), want: newTestTuple("<type 'TypeError'>", "<type 'TypeError'>").ToObject()},
		{args: wrapArgs(TupleType), want: newTestTuple("<type 'tuple'>", "<type 'tuple'>").ToObject()},
		{args: wrapArgs(TypeType), want: newTestTuple("<type 'type'>", "<type 'type'>").ToObject()},
		{args: wrapArgs(fooType), want: newTestTuple("<type 'foo.bar.Foo'>", "<type 'foo.bar.Foo'>").ToObject()},
		{args: wrapArgs(mustNotRaise(WrapNative(NewRootFrame(), reflect.ValueOf(t))).Type()), want: newTestTuple("<type '*T'>", "<type '*T'>").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestTypeModule(t *testing.T) {
	fn := newBuiltinFunction("__module__", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		if raised := checkFunctionArgs(f, "__module__", args, TypeType); raised != nil {
			return nil, raised
		}
		mod, raised := toTypeUnsafe(args[0]).Dict().GetItemString(f, "__module__")
		if raised != nil || mod != nil {
			return mod, raised
		}
		return None, nil
	}).ToObject()
	fooType := newTestClass("Foo", []*Type{ObjectType}, newTestDict("__module__", "foo.bar"))
	barType := newTestClass("Bar", []*Type{ObjectType}, NewDict())
	cases := []invokeTestCase{
		{args: wrapArgs(IntType), want: NewStr("__builtin__").ToObject()},
		{args: wrapArgs(mustNotRaise(WrapNative(NewRootFrame(), reflect.ValueOf(t))).Type()), want: NewStr("__builtin__").ToObject()},
		{args: wrapArgs(fooType), want: NewStr("foo.bar").ToObject()},
		{args: wrapArgs(barType), want: NewStr("__builtin__").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fn, &cas); err != "" {
			t.Error(err)
		}
	}
}

func newTestClass(name string, bases []*Type, dict *Dict) *Type {
	t, raised := newClass(NewRootFrame(), TypeType, name, bases, dict)
	if raised != nil {
		panic(raised)
	}
	return t
}

// newTestClassStrictEq returns a new class that defines eq and ne operators
// that check whether the lhs and rhs have the same type and that the value
// fields are also equal. This is useful for testing that the builtin types
// return objects of the correct type for their __new__ method.
func newTestClassStrictEq(name string, base *Type) *Type {
	var t *Type
	t = newTestClass(name, []*Type{base}, newStringDict(map[string]*Object{
		"__repr__": newBuiltinFunction("__repr__", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
			if raised := checkMethodArgs(f, "__repr__", args, t); raised != nil {
				return nil, raised
			}
			repr, raised := GetAttr(f, base.ToObject(), NewStr("__repr__"), nil)
			if raised != nil {
				return nil, raised
			}
			s, raised := repr.Call(f, Args{args[0]}, nil)
			if raised != nil {
				return nil, raised
			}
			if !s.isInstance(StrType) {
				return nil, f.RaiseType(TypeErrorType, "__repr__ returned non-str")
			}
			return NewStr(fmt.Sprintf("%s(%s)", t.Name(), toStrUnsafe(s).Value())).ToObject(), nil
		}).ToObject(),
		"__eq__": newBuiltinFunction("__eq__", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
			if raised := checkMethodArgs(f, "__eq__", args, t, ObjectType); raised != nil {
				return nil, raised
			}
			if args[1].typ != t {
				return False.ToObject(), nil
			}
			return base.slots.Eq.Fn(f, args[0], args[1])
		}).ToObject(),
		"__ne__": newBuiltinFunction("__ne__", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
			if raised := checkMethodArgs(f, "__ne__", args, t, ObjectType); raised != nil {
				return nil, raised
			}
			o, raised := Eq(f, args[0], args[1])
			if raised != nil {
				return nil, raised
			}
			eq, raised := IsTrue(f, o)
			if raised != nil {
				return nil, raised
			}
			return GetBool(eq).ToObject(), nil
		}).ToObject(),
	}))
	return t
}

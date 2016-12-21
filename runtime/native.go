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
	"bytes"
	"fmt"
	"math/big"
	"reflect"
	"runtime"
	"sync"
	"unsafe"
)

var (
	nativeFuncType      = newSimpleType("func", nativeType)
	nativeMetaclassType = newBasisType("nativetype", reflect.TypeOf(nativeMetaclass{}), toNativeMetaclassUnsafe, TypeType)
	nativeSliceType     = newSimpleType("slice", nativeType)
	nativeType          = newBasisType("native", reflect.TypeOf(native{}), toNativeUnsafe, ObjectType)
	// Prepopulate the builtin primitive types so that WrapNative calls on
	// these kinds of values resolve directly to primitive Python types.
	nativeTypes = map[reflect.Type]*Type{
		reflect.TypeOf(bool(false)):     BoolType,
		reflect.TypeOf(float32(0)):      FloatType,
		reflect.TypeOf(float64(0)):      FloatType,
		reflect.TypeOf(int(0)):          IntType,
		reflect.TypeOf(int16(0)):        IntType,
		reflect.TypeOf(int32(0)):        IntType,
		reflect.TypeOf(int64(0)):        IntType,
		reflect.TypeOf(int8(0)):         IntType,
		reflect.TypeOf(string("")):      StrType,
		reflect.TypeOf(uint(0)):         IntType,
		reflect.TypeOf(uint16(0)):       IntType,
		reflect.TypeOf(uint32(0)):       IntType,
		reflect.TypeOf(uint64(0)):       IntType,
		reflect.TypeOf(uint8(0)):        IntType,
		reflect.TypeOf(uintptr(0)):      IntType,
		reflect.TypeOf([]rune(nil)):     UnicodeType,
		reflect.TypeOf(big.Int{}):       LongType,
		reflect.TypeOf((*big.Int)(nil)): LongType,
	}
	nativeTypesMutex  = sync.Mutex{}
	sliceIteratorType = newBasisType("sliceiterator", reflect.TypeOf(sliceIterator{}), toSliceIteratorUnsafe, ObjectType)
)

type nativeMetaclass struct {
	Type
	rtype reflect.Type
}

func toNativeMetaclassUnsafe(o *Object) *nativeMetaclass {
	return (*nativeMetaclass)(o.toPointer())
}

func newNativeType(rtype reflect.Type, base *Type, d *Dict) *nativeMetaclass {
	return &nativeMetaclass{
		Type{
			Object: Object{typ: nativeMetaclassType, dict: d},
			name:   nativeTypeName(rtype),
			basis:  base.basis,
			bases:  []*Type{base},
			flags:  typeFlagDefault,
		},
		rtype,
	}
}

func nativeMetaclassNew(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "new", args, nativeMetaclassType); raised != nil {
		return nil, raised
	}
	return WrapNative(f, reflect.New(toNativeMetaclassUnsafe(args[0]).rtype))
}

func initNativeMetaclassType(dict map[string]*Object) {
	nativeMetaclassType.flags &^= typeFlagInstantiable | typeFlagBasetype
	dict["new"] = newBuiltinFunction("new", nativeMetaclassNew).ToObject()
}

type native struct {
	Object
	value reflect.Value
}

func toNativeUnsafe(o *Object) *native {
	return (*native)(o.toPointer())
}

// ToObject upcasts n to an Object.
func (n *native) ToObject() *Object {
	return &n.Object
}

func nativeNative(f *Frame, o *Object) (reflect.Value, *BaseException) {
	return toNativeUnsafe(o).value, nil
}

func initNativeType(map[string]*Object) {
	nativeType.flags = typeFlagDefault &^ typeFlagInstantiable
	nativeType.slots.Native = &nativeSlot{nativeNative}
}

func nativeFuncCall(f *Frame, callable *Object, args Args, kwargs KWArgs) (*Object, *BaseException) {
	return nativeInvoke(f, toNativeUnsafe(callable).value, args)
}

func nativeFuncGetName(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "_get_name", args, nativeFuncType); raised != nil {
		return nil, raised
	}
	fun := runtime.FuncForPC(toNativeUnsafe(args[0]).value.Pointer())
	return NewStr(fun.Name()).ToObject(), nil
}

func nativeFuncRepr(f *Frame, o *Object) (*Object, *BaseException) {
	name, raised := GetAttr(f, o, NewStr("__name__"), NewStr("<unknown>").ToObject())
	if raised != nil {
		return nil, raised
	}
	nameStr, raised := ToStr(f, name)
	if raised != nil {
		return nil, raised
	}
	typeName := nativeTypeName(toNativeUnsafe(o).value.Type())
	return NewStr(fmt.Sprintf("<%s %s at %p>", typeName, nameStr.Value(), o)).ToObject(), nil
}

func initNativeFuncType(dict map[string]*Object) {
	dict["__name__"] = newProperty(newBuiltinFunction("_get_name", nativeFuncGetName).ToObject(), None, None).ToObject()
	nativeFuncType.slots.Call = &callSlot{nativeFuncCall}
	nativeFuncType.slots.Repr = &unaryOpSlot{nativeFuncRepr}
}

func nativeSliceIter(f *Frame, o *Object) (*Object, *BaseException) {
	return newSliceIterator(toNativeUnsafe(o).value), nil
}

func initNativeSliceType(map[string]*Object) {
	nativeSliceType.slots.Iter = &unaryOpSlot{nativeSliceIter}
}

type sliceIterator struct {
	Object
	slice    reflect.Value
	mutex    sync.Mutex
	numElems int
	index    int
}

func newSliceIterator(slice reflect.Value) *Object {
	iter := &sliceIterator{Object: Object{typ: sliceIteratorType}, slice: slice, numElems: slice.Len()}
	return &iter.Object
}

func toSliceIteratorUnsafe(o *Object) *sliceIterator {
	return (*sliceIterator)(o.toPointer())
}

func sliceIteratorIter(f *Frame, o *Object) (*Object, *BaseException) {
	return o, nil
}

func sliceIteratorNext(f *Frame, o *Object) (ret *Object, raised *BaseException) {
	i := toSliceIteratorUnsafe(o)
	i.mutex.Lock()
	if i.index < i.numElems {
		ret, raised = WrapNative(f, i.slice.Index(i.index))
		i.index++
	} else {
		raised = f.Raise(StopIterationType.ToObject(), nil, nil)
	}
	i.mutex.Unlock()
	return ret, raised
}

func initSliceIteratorType(map[string]*Object) {
	sliceIteratorType.flags &= ^(typeFlagBasetype | typeFlagInstantiable)
	sliceIteratorType.slots.Iter = &unaryOpSlot{sliceIteratorIter}
	sliceIteratorType.slots.Next = &unaryOpSlot{sliceIteratorNext}
}

// WrapNative takes a reflect.Value object and converts the underlying Go
// object to a Python object in the following way:
//
// - Primitive types are converted in the way you'd expect: Go int types map to
//   Python int, Go booleans to Python bool, etc. User-defined primitive Go types
//   are subclasses of the Python primitives.
// - *big.Int is represented by Python long.
// - Functions are represented by Python type that supports calling into native
//   functions.
// - Interfaces are converted to their concrete held type, or None if IsNil.
// - Other native types are wrapped in an opaque native type that does not
//   support directly accessing the underlying object from Python. When these
//   opaque objects are passed back into Go by native function calls, however,
//   they will be unwrapped back to their Go representation.
func WrapNative(f *Frame, v reflect.Value) (*Object, *BaseException) {
	switch v.Kind() {
	case reflect.Interface:
		if v.IsNil() {
			return None, nil
		}
		// Interfaces have undefined methods (Method() will return an
		// invalid func value). What we really want to wrap is the
		// underlying, concrete object.
		v = v.Elem()
	case reflect.Invalid:
		panic("zero reflect.Value passed to WrapNative")
	}

	t := getNativeType(v.Type())

	switch v.Kind() {
	// ===============
	// Primitive types
	// ===============
	// Primitive Go types are translated into primitive Python types or
	// subclasses of primitive Python types.
	case reflect.Bool:
		i := 0
		if v.Bool() {
			i = 1
		}
		// TODO: Make native bool subtypes singletons and add support
		// for __new__ so we can use t.Call() here.
		return (&Int{Object{typ: t}, i}).ToObject(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint8, reflect.Uint16:
		return t.Call(f, Args{NewInt(int(v.Int())).ToObject()}, nil)
	// Handle potentially large ints separately in case of overflow.
	case reflect.Int64:
		i := v.Int()
		if i < int64(MinInt) || i > int64(MaxInt) {
			return NewLong(big.NewInt(i)).ToObject(), nil
		}
		return t.Call(f, Args{NewInt(int(i)).ToObject()}, nil)
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		i := v.Uint()
		if i > uint64(MaxInt) {
			return t.Call(f, Args{NewLong((new(big.Int).SetUint64(i))).ToObject()}, nil)
		}
		return t.Call(f, Args{NewInt(int(i)).ToObject()}, nil)
	case reflect.Uintptr:
		// Treat uintptr as a opaque data encoded as a signed integer.
		i := int64(v.Uint())
		if i < int64(MinInt) || i > int64(MaxInt) {
			return NewLong(big.NewInt(i)).ToObject(), nil
		}
		return t.Call(f, Args{NewInt(int(i)).ToObject()}, nil)
	case reflect.Float32, reflect.Float64:
		x := v.Float()
		return t.Call(f, Args{NewFloat(x).ToObject()}, nil)
	case reflect.String:
		return t.Call(f, Args{NewStr(v.String()).ToObject()}, nil)
	case reflect.Slice:
		if v.Type().Elem() == reflect.TypeOf(rune(0)) {
			// Avoid reflect.Copy() and Interface()+copy() in case
			// this is an unexported field.
			// TODO: Implement a fast path that uses copy() when
			// v.CanInterface() is true.
			numRunes := v.Len()
			runes := make([]rune, numRunes)
			for i := 0; i < numRunes; i++ {
				runes[i] = rune(v.Index(i).Int())
			}
			return t.Call(f, Args{NewUnicodeFromRunes(runes).ToObject()}, nil)
		}

	// =============
	// Complex types
	// =============
	// Non-primitive types are always nativeType subclasses except in a few
	// specific cases which we handle below.
	case reflect.Ptr:
		if v.IsNil() {
			return None, nil
		}
		if v.Type() == reflect.TypeOf((*big.Int)(nil)) {
			i := v.Interface().(*big.Int)
			return t.Call(f, Args{NewLong(i).ToObject()}, nil)
		}
		if basis := v.Elem(); basisTypes[basis.Type()] != nil {
			// We have a basis type that is binary compatible with
			// Object.
			return (*Object)(unsafe.Pointer(basis.UnsafeAddr())), nil
		}
	case reflect.Struct:
		if i, ok := v.Interface().(big.Int); ok {
			return t.Call(f, Args{NewLong(&i).ToObject()}, nil)
		}
	case reflect.Chan, reflect.Func, reflect.Map:
		if v.IsNil() {
			return None, nil
		}
	}
	return (&native{Object{typ: t}, v}).ToObject(), nil
}

func getNativeType(rtype reflect.Type) *Type {
	nativeTypesMutex.Lock()
	t, ok := nativeTypes[rtype]
	if !ok {
		// Choose an appropriate base class for this kind of native
		// object.
		base := nativeType
		switch rtype.Kind() {
		case reflect.Bool:
			base = BoolType
		case reflect.Float32, reflect.Float64:
			base = FloatType
		case reflect.Func:
			base = nativeFuncType
		case reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8, reflect.Int, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8, reflect.Uint, reflect.Uintptr:
			base = IntType
		case reflect.Slice:
			base = nativeSliceType
		case reflect.String:
			base = StrType
		}
		d := map[string]*Object{"__module__": builtinStr.ToObject()}
		numMethod := rtype.NumMethod()
		for i := 0; i < numMethod; i++ {
			meth := rtype.Method(i)
			// A non-empty PkgPath indicates a private method that shouldn't
			// be registered.
			if meth.PkgPath == "" {
				d[meth.Name] = newNativeMethod(meth.Name, meth.Func)
			}
		}
		t = &newNativeType(rtype, base, newStringDict(d)).Type
		// This cannot fail since we're defining simple classes.
		if err := prepareType(t); err != "" {
			logFatal(err)
		}
	}
	nativeTypes[rtype] = t
	nativeTypesMutex.Unlock()
	return t
}

func newNativeMethod(name string, fun reflect.Value) *Object {
	return newBuiltinFunction(name, func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		return nativeInvoke(f, fun, args)
	}).ToObject()
}

func maybeConvertValue(f *Frame, o *Object, expectedRType reflect.Type) (reflect.Value, *BaseException) {
	if expectedRType.Kind() == reflect.Ptr {
		// When the expected type is some basis pointer, check if o is
		// an instance of that basis and use it if so.
		if t, ok := basisTypes[expectedRType.Elem()]; ok && o.isInstance(t) {
			return t.slots.Basis.Fn(o).Addr(), nil
		}
	}
	if o == None {
		switch expectedRType.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
			return reflect.Zero(expectedRType), nil
		default:
			return reflect.Value{}, f.RaiseType(TypeErrorType, fmt.Sprintf("cannot convert None to %s", expectedRType))
		}
	}
	val, raised := ToNative(f, o)
	if raised != nil {
		return reflect.Value{}, raised
	}
	rtype := val.Type()
	for {
		if rtype == expectedRType {
			return val, nil
		}
		if rtype.ConvertibleTo(expectedRType) {
			return val.Convert(expectedRType), nil
		}
		if rtype.Kind() == reflect.Ptr {
			val = val.Elem()
			rtype = val.Type()
			continue
		}
		break
	}
	return reflect.Value{}, f.RaiseType(TypeErrorType, fmt.Sprintf("cannot convert %s to %s", rtype, expectedRType))
}

func nativeFuncTypeName(rtype reflect.Type) string {
	var buf bytes.Buffer
	buf.WriteString("func(")
	numIn := rtype.NumIn()
	for i := 0; i < numIn; i++ {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(nativeTypeName(rtype.In(i)))
	}
	buf.WriteString(")")
	numOut := rtype.NumOut()
	if numOut == 1 {
		buf.WriteString(" ")
		buf.WriteString(nativeTypeName(rtype.Out(0)))
	} else if numOut > 1 {
		buf.WriteString(" (")
		for i := 0; i < numOut; i++ {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(nativeTypeName(rtype.Out(i)))
		}
		buf.WriteString(")")
	}
	return buf.String()
}

func nativeInvoke(f *Frame, fun reflect.Value, args Args) (ret *Object, raised *BaseException) {
	rtype := fun.Type()
	argc := len(args)
	expectedArgc := rtype.NumIn()
	fixedArgc := expectedArgc
	if rtype.IsVariadic() {
		fixedArgc--
	}
	if rtype.IsVariadic() && argc < fixedArgc {
		msg := fmt.Sprintf("native function takes at least %d arguments, (%d given)", fixedArgc, argc)
		return nil, f.RaiseType(TypeErrorType, msg)
	}
	if !rtype.IsVariadic() && argc != fixedArgc {
		msg := fmt.Sprintf("native function takes %d arguments, (%d given)", fixedArgc, argc)
		return nil, f.RaiseType(TypeErrorType, msg)
	}
	// Convert all the fixed args to their native types.
	nativeArgs := make([]reflect.Value, argc)
	for i := 0; i < fixedArgc; i++ {
		if nativeArgs[i], raised = maybeConvertValue(f, args[i], rtype.In(i)); raised != nil {
			return nil, raised
		}
	}
	if rtype.IsVariadic() {
		// The last input in a variadic function is a slice with elem type of the
		// var args.
		elementT := rtype.In(fixedArgc).Elem()
		for i := fixedArgc; i < argc; i++ {
			if nativeArgs[i], raised = maybeConvertValue(f, args[i], elementT); raised != nil {
				return nil, raised
			}
		}
	}
	result := fun.Call(nativeArgs)
	if e, _ := f.ExcInfo(); e != nil {
		return nil, e
	}
	numResults := len(result)
	if numResults > 0 && result[numResults-1].Type() == reflect.TypeOf((*BaseException)(nil)) {
		numResults--
		result = result[:numResults]
	}
	// Convert the return value slice to a single value when only one value is
	// returned, or to a Tuple, when many are returned.
	switch numResults {
	case 0:
		ret = None
	case 1:
		ret, raised = WrapNative(f, result[0])
	default:
		elems := make([]*Object, numResults)
		for i := 0; i < numResults; i++ {
			if elems[i], raised = WrapNative(f, result[i]); raised != nil {
				return nil, raised
			}
		}
		ret = NewTuple(elems...).ToObject()
	}
	return ret, raised
}

func nativeTypeName(rtype reflect.Type) string {
	if rtype.Name() != "" {
		return rtype.Name()
	}
	switch rtype.Kind() {
	case reflect.Array:
		return fmt.Sprintf("[%d]%s", rtype.Len(), nativeTypeName(rtype.Elem()))
	case reflect.Chan:
		return fmt.Sprintf("chan %s", nativeTypeName(rtype.Elem()))
	case reflect.Func:
		return nativeFuncTypeName(rtype)
	case reflect.Map:
		return fmt.Sprintf("map[%s]%s", nativeTypeName(rtype.Key()), nativeTypeName(rtype.Elem()))
	case reflect.Ptr:
		return fmt.Sprintf("*%s", nativeTypeName(rtype.Elem()))
	case reflect.Slice:
		return fmt.Sprintf("[]%s", nativeTypeName(rtype.Elem()))
	case reflect.Struct:
		return "anonymous struct"
	default:
		return "unknown"
	}
}

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
	"math/big"
	"os"
	"unicode"
)

var (
	// Builtins contains all of the Python built-in identifiers.
	Builtins   = NewDict()
	builtinStr = NewStr("__builtin__")
	// ExceptionTypes contains all builtin exception types.
	ExceptionTypes []*Type
	// NoneType is the object representing the Python 'NoneType' type.
	NoneType = newSimpleType("NoneType", ObjectType)
	// None is the singleton NoneType object representing the Python 'None'
	// object.
	None = &Object{typ: NoneType}
	// NotImplementedType is the object representing the Python
	// 'NotImplementedType' object.
	NotImplementedType = newSimpleType("NotImplementedType", ObjectType)
	// NotImplemented is the singleton NotImplementedType object
	// representing the Python 'NotImplemented' object.
	NotImplemented   = newObject(NotImplementedType)
	unboundLocalType = newSimpleType("UnboundLocalType", ObjectType)
	// UnboundLocal is a singleton held by local variables in generated
	// code before they are bound.
	UnboundLocal = newObject(unboundLocalType)
)

func noneRepr(*Frame, *Object) (*Object, *BaseException) {
	return NewStr("None").ToObject(), nil
}

func initNoneType(map[string]*Object) {
	NoneType.flags &= ^(typeFlagInstantiable | typeFlagBasetype)
	NoneType.slots.Repr = &unaryOpSlot{noneRepr}
}

func initNotImplementedType(map[string]*Object) {
	NotImplementedType.flags &= ^(typeFlagInstantiable | typeFlagBasetype)
}

func initUnboundLocalType(map[string]*Object) {
	unboundLocalType.flags &= ^(typeFlagInstantiable | typeFlagBasetype)
}

type typeState int

const (
	typeStateNotReady typeState = iota
	typeStateInitializing
	typeStateReady
)

type builtinTypeInit func(map[string]*Object)

type builtinTypeInfo struct {
	state  typeState
	init   builtinTypeInit
	global bool
}

var builtinTypes = map[*Type]*builtinTypeInfo{
	ArithmeticErrorType:           {global: true},
	AssertionErrorType:            {global: true},
	AttributeErrorType:            {global: true},
	BaseExceptionType:             {init: initBaseExceptionType, global: true},
	BaseStringType:                {init: initBaseStringType, global: true},
	BoolType:                      {init: initBoolType, global: true},
	BytesWarningType:              {global: true},
	CodeType:                      {},
	ClassMethodType:               {init: initClassMethodType, global: true},
	DeprecationWarningType:        {global: true},
	dictItemIteratorType:          {init: initDictItemIteratorType},
	dictKeyIteratorType:           {init: initDictKeyIteratorType},
	DictType:                      {init: initDictType, global: true},
	enumerateType:                 {init: initEnumerateType, global: true},
	EnvironmentErrorType:          {global: true},
	ExceptionType:                 {global: true},
	FileType:                      {init: initFileType, global: true},
	FloatType:                     {init: initFloatType, global: true},
	FrameType:                     {init: initFrameType},
	FrozenSetType:                 {init: initFrozenSetType, global: true},
	FunctionType:                  {init: initFunctionType},
	FutureWarningType:             {global: true},
	GeneratorType:                 {init: initGeneratorType},
	ImportErrorType:               {global: true},
	ImportWarningType:             {global: true},
	IndexErrorType:                {global: true},
	IntType:                       {init: initIntType, global: true},
	IOErrorType:                   {global: true},
	KeyErrorType:                  {global: true},
	listIteratorType:              {init: initListIteratorType},
	ListType:                      {init: initListType, global: true},
	LongType:                      {init: initLongType, global: true},
	LookupErrorType:               {global: true},
	MemoryErrorType:               {global: true},
	MethodType:                    {init: initMethodType},
	ModuleType:                    {init: initModuleType},
	NameErrorType:                 {global: true},
	nativeBoolMetaclassType:       {init: initNativeBoolMetaclassType},
	nativeFuncType:                {init: initNativeFuncType},
	nativeMetaclassType:           {init: initNativeMetaclassType},
	nativeSliceType:               {init: initNativeSliceType},
	nativeType:                    {init: initNativeType},
	NoneType:                      {init: initNoneType, global: true},
	NotImplementedErrorType:       {global: true},
	NotImplementedType:            {init: initNotImplementedType, global: true},
	ObjectType:                    {init: initObjectType, global: true},
	OSErrorType:                   {global: true},
	OverflowErrorType:             {global: true},
	PendingDeprecationWarningType: {global: true},
	PropertyType:                  {init: initPropertyType, global: true},
	rangeIteratorType:             {init: initRangeIteratorType, global: true},
	ReferenceErrorType:            {global: true},
	RuntimeErrorType:              {global: true},
	RuntimeWarningType:            {global: true},
	seqIteratorType:               {init: initSeqIteratorType},
	SetType:                       {init: initSetType, global: true},
	sliceIteratorType:             {init: initSliceIteratorType},
	SliceType:                     {init: initSliceType, global: true},
	StandardErrorType:             {global: true},
	StaticMethodType:              {init: initStaticMethodType, global: true},
	StopIterationType:             {global: true},
	StrType:                       {init: initStrType, global: true},
	superType:                     {init: initSuperType, global: true},
	SyntaxErrorType:               {global: true},
	SyntaxWarningType:             {global: true},
	SystemErrorType:               {global: true},
	SystemExitType:                {global: true, init: initSystemExitType},
	TracebackType:                 {init: initTracebackType},
	TupleType:                     {init: initTupleType, global: true},
	TypeErrorType:                 {global: true},
	TypeType:                      {init: initTypeType, global: true},
	UnboundLocalErrorType:         {global: true},
	unboundLocalType:              {init: initUnboundLocalType},
	UnicodeDecodeErrorType:        {global: true},
	UnicodeEncodeErrorType:        {global: true},
	UnicodeErrorType:              {global: true},
	UnicodeType:                   {init: initUnicodeType, global: true},
	UnicodeWarningType:            {global: true},
	UserWarningType:               {global: true},
	ValueErrorType:                {global: true},
	WarningType:                   {global: true},
	WeakRefType:                   {init: initWeakRefType},
	xrangeType:                    {init: initXRangeType, global: true},
	ZeroDivisionErrorType:         {global: true},
}

func initBuiltinType(typ *Type, info *builtinTypeInfo) {
	if info.state == typeStateReady {
		return
	}
	if info.state == typeStateInitializing {
		logFatal(fmt.Sprintf("cycle in type initialization for: %s", typ.name))
	}
	info.state = typeStateInitializing
	for _, base := range typ.bases {
		baseInfo, ok := builtinTypes[base]
		if !ok {
			logFatal(fmt.Sprintf("base type not registered for: %s", typ.name))
		}
		initBuiltinType(base, baseInfo)
	}
	prepareBuiltinType(typ, info.init)
	info.state = typeStateReady
	if typ.isSubclass(BaseExceptionType) {
		ExceptionTypes = append(ExceptionTypes, typ)
	}
}

func builtinAbs(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkFunctionArgs(f, "abs", args, ObjectType); raised != nil {
		return nil, raised
	}
	return Abs(f, args[0])
}

func builtinMapFn(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	argc := len(args)
	if argc < 2 {
		return nil, f.RaiseType(TypeErrorType, "map() requires at least two args")
	}
	result := make([]*Object, 0, 2)
	z, raised := zipLongest(f, args[1:])
	if raised != nil {
		return nil, raised
	}
	for _, tuple := range z {
		if args[0] == None {
			if argc == 2 {
				result = append(result, tuple[0])
			} else {
				result = append(result, NewTuple(tuple...).ToObject())
			}
		} else {
			ret, raised := args[0].Call(f, tuple, nil)
			if raised != nil {
				return nil, raised
			}
			result = append(result, ret)
		}
	}

	return NewList(result...).ToObject(), nil
}

func builtinAll(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkFunctionArgs(f, "all", args, ObjectType); raised != nil {
		return nil, raised
	}
	pred := func(o *Object) (bool, *BaseException) {
		ret, raised := IsTrue(f, o)
		if raised != nil {
			return false, raised
		}
		return !ret, nil
	}
	foundFalseItem, raised := seqFindFirst(f, args[0], pred)
	if raised != nil {
		return nil, raised
	}
	return GetBool(!foundFalseItem).ToObject(), raised
}

func builtinAny(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkFunctionArgs(f, "any", args, ObjectType); raised != nil {
		return nil, raised
	}
	pred := func(o *Object) (bool, *BaseException) {
		ret, raised := IsTrue(f, o)
		if raised != nil {
			return false, raised
		}
		return ret, nil
	}
	foundTrueItem, raised := seqFindFirst(f, args[0], pred)
	if raised != nil {
		return nil, raised
	}
	return GetBool(foundTrueItem).ToObject(), raised
}

func builtinBin(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkFunctionArgs(f, "bin", args, ObjectType); raised != nil {
		return nil, raised
	}
	index, raised := Index(f, args[0])
	if raised != nil {
		return nil, raised
	}
	if index == nil {
		format := "%s object cannot be interpreted as an index"
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, args[0].typ.Name()))
	}
	return NewStr(numberToBase("0b", 2, index)).ToObject(), nil
}

func builtinCallable(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkFunctionArgs(f, "callable", args, ObjectType); raised != nil {
		return nil, raised
	}
	o := args[0]
	if call := o.Type().slots.Call; call == nil {
		return False.ToObject(), nil
	}
	return True.ToObject(), nil
}

func builtinChr(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkFunctionArgs(f, "chr", args, IntType); raised != nil {
		return nil, raised
	}
	i := toIntUnsafe(args[0]).Value()
	if i < 0 || i > 255 {
		return nil, f.RaiseType(ValueErrorType, "chr() arg not in range(256)")
	}
	return NewStr(string([]byte{byte(i)})).ToObject(), nil
}

func builtinCmp(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkFunctionArgs(f, "cmp", args, ObjectType, ObjectType); raised != nil {
		return nil, raised
	}
	return Compare(f, args[0], args[1])
}

func builtinDir(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	// TODO: Support __dir__.
	if raised := checkFunctionArgs(f, "dir", args, ObjectType); raised != nil {
		return nil, raised
	}
	d := NewDict()
	o := args[0]
	if o.dict != nil {
		raised := seqForEach(f, o.dict.ToObject(), func(k *Object) *BaseException {
			return d.SetItem(f, k, None)
		})
		if raised != nil {
			return nil, raised
		}
	}
	for _, t := range o.typ.mro {
		raised := seqForEach(f, t.dict.ToObject(), func(k *Object) *BaseException {
			return d.SetItem(f, k, None)
		})
		if raised != nil {
			return nil, raised
		}
	}
	l := d.Keys(f)
	if raised := l.Sort(f); raised != nil {
		return nil, raised
	}
	return l.ToObject(), nil
}

func builtinFrame(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkFunctionArgs(f, "__frame__", args); raised != nil {
		return nil, raised
	}
	return f.ToObject(), nil
}

func builtinGetAttr(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	expectedTypes := []*Type{ObjectType, StrType, ObjectType}
	argc := len(args)
	if argc == 2 {
		expectedTypes = expectedTypes[:2]
	}
	if raised := checkFunctionArgs(f, "getattr", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	var def *Object
	if argc == 3 {
		def = args[2]
	}
	return GetAttr(f, args[0], toStrUnsafe(args[1]), def)
}

func builtinGlobals(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkFunctionArgs(f, "globals", args); raised != nil {
		return nil, raised
	}
	return f.globals.ToObject(), nil
}

func builtinHasAttr(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkFunctionArgs(f, "hasattr", args, ObjectType, StrType); raised != nil {
		return nil, raised
	}
	if _, raised := GetAttr(f, args[0], toStrUnsafe(args[1]), nil); raised != nil {
		if raised.isInstance(AttributeErrorType) {
			f.RestoreExc(nil, nil)
			return False.ToObject(), nil
		}
		return nil, raised
	}
	return True.ToObject(), nil
}

func builtinHash(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkFunctionArgs(f, "hash", args, ObjectType); raised != nil {
		return nil, raised
	}
	h, raised := Hash(f, args[0])
	if raised != nil {
		return nil, raised
	}
	return h.ToObject(), nil
}

func builtinHex(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	// In Python3 we would call __index__ similarly to builtinBin().
	if raised := checkFunctionArgs(f, "hex", args, ObjectType); raised != nil {
		return nil, raised
	}
	if method, raised := args[0].typ.mroLookup(f, NewStr("__hex__")); raised != nil {
		return nil, raised
	} else if method != nil {
		return method.Call(f, args, nil)
	}
	if !args[0].isInstance(IntType) && !args[0].isInstance(LongType) {
		return nil, f.RaiseType(TypeErrorType, "hex() argument can't be converted to hex")
	}
	s := numberToBase("0x", 16, args[0])
	if args[0].isInstance(LongType) {
		s += "L"
	}
	return NewStr(s).ToObject(), nil
}

func builtinID(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkFunctionArgs(f, "id", args, ObjectType); raised != nil {
		return nil, raised
	}
	return NewInt(int(uintptr(args[0].toPointer()))).ToObject(), nil
}

func builtinIsInstance(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkFunctionArgs(f, "isinstance", args, ObjectType, ObjectType); raised != nil {
		return nil, raised
	}
	ret, raised := IsInstance(f, args[0], args[1])
	if raised != nil {
		return nil, raised
	}
	return GetBool(ret).ToObject(), nil
}

func builtinIsSubclass(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkFunctionArgs(f, "issubclass", args, ObjectType, ObjectType); raised != nil {
		return nil, raised
	}
	ret, raised := IsSubclass(f, args[0], args[1])
	if raised != nil {
		return nil, raised
	}
	return GetBool(ret).ToObject(), nil
}

func builtinIter(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkFunctionArgs(f, "iter", args, ObjectType); raised != nil {
		return nil, raised
	}
	return Iter(f, args[0])
}

func builtinLen(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkFunctionArgs(f, "len", args, ObjectType); raised != nil {
		return nil, raised
	}
	ret, raised := Len(f, args[0])
	return ret.ToObject(), raised
}

func builtinMax(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	return builtinMinMax(f, true, args, kwargs)
}

func builtinMin(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	return builtinMinMax(f, false, args, kwargs)
}

func builtinNext(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkFunctionArgs(f, "next", args, ObjectType); raised != nil {
		return nil, raised
	}
	ret, raised := Next(f, args[0])
	if raised != nil {
		return nil, raised
	}
	if ret != nil {
		return ret, nil
	}
	return nil, f.Raise(StopIterationType.ToObject(), nil, nil)
}

func builtinOct(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	// In Python3 we would call __index__ similarly to builtinBin().
	if raised := checkFunctionArgs(f, "oct", args, ObjectType); raised != nil {
		return nil, raised
	}
	if method, raised := args[0].typ.mroLookup(f, NewStr("__oct__")); raised != nil {
		return nil, raised
	} else if method != nil {
		return method.Call(f, args, nil)
	}
	if !args[0].isInstance(IntType) && !args[0].isInstance(LongType) {
		return nil, f.RaiseType(TypeErrorType, "oct() argument can't be converted to oct")
	}
	s := numberToBase("0", 8, args[0])
	if args[0].isInstance(LongType) {
		s += "L"
	}
	// For oct(0), return "0", not "00".
	if s == "00" {
		s = "0"
	}
	return NewStr(s).ToObject(), nil
}

func builtinOpen(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	return FileType.Call(f, args, kwargs)
}

func builtinOrd(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	const lenMsg = "ord() expected a character, but string of length %d found"
	if raised := checkFunctionArgs(f, "ord", args, BaseStringType); raised != nil {
		return nil, raised
	}
	o := args[0]
	var result int
	if o.isInstance(StrType) {
		s := toStrUnsafe(o).Value()
		if numChars := len(s); numChars != 1 {
			return nil, f.RaiseType(ValueErrorType, fmt.Sprintf(lenMsg, numChars))
		}
		result = int(([]byte(s))[0])
	} else {
		s := toUnicodeUnsafe(o).Value()
		if numChars := len(s); numChars != 1 {
			return nil, f.RaiseType(ValueErrorType, fmt.Sprintf(lenMsg, numChars))
		}
		result = int(s[0])
	}
	return NewInt(result).ToObject(), nil
}

func builtinPrint(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	sep := " "
	end := "\n"
	file := os.Stdout
	for _, kwarg := range kwargs {
		switch kwarg.Name {
		case "sep":
			kwsep, raised := ToStr(f, kwarg.Value)
			if raised != nil {
				return nil, raised
			}
			sep = kwsep.Value()
		case "end":
			kwend, raised := ToStr(f, kwarg.Value)
			if raised != nil {
				return nil, raised
			}
			end = kwend.Value()
		case "file":
			// TODO: need to map Python sys.stdout, sys.stderr etc. to os.Stdout,
			// os.Stderr, but for other file-like objects would need to recover
			// to the file descriptor probably
		}
	}
	return nil, pyPrint(f, args, sep, end, file)
}

func builtinRange(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	r, raised := xrangeType.Call(f, args, nil)
	if raised != nil {
		return nil, raised
	}
	return ListType.Call(f, []*Object{r}, nil)
}

func builtinRepr(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkFunctionArgs(f, "repr", args, ObjectType); raised != nil {
		return nil, raised
	}
	s, raised := Repr(f, args[0])
	if raised != nil {
		return nil, raised
	}
	return s.ToObject(), nil
}

func builtinSorted(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	// TODO: Support (cmp=None, key=None, reverse=False)
	if raised := checkFunctionArgs(f, "sorted", args, ObjectType); raised != nil {
		return nil, raised
	}
	result, raised := ListType.Call(f, Args{args[0]}, nil)
	if raised != nil {
		return nil, raised
	}
	toListUnsafe(result).Sort(f)
	return result, nil
}

func builtinUniChr(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkFunctionArgs(f, "unichr", args, IntType); raised != nil {
		return nil, raised
	}
	i := toIntUnsafe(args[0]).Value()
	if i < 0 || i > unicode.MaxRune {
		return nil, f.RaiseType(ValueErrorType, fmt.Sprintf("unichr() arg not in range(0x%x)", unicode.MaxRune))
	}
	return NewUnicodeFromRunes([]rune{rune(i)}).ToObject(), nil
}

func builtinZip(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	argc := len(args)
	if argc == 0 {
		return NewList().ToObject(), nil
	}
	result := make([]*Object, 0, 2)
	iters, raised := initIters(f, args)
	if raised != nil {
		return nil, raised
	}

Outer:
	for {
		elems := make([]*Object, argc)
		for i, iter := range iters {
			elem, raised := Next(f, iter)
			if raised != nil {
				if raised.isInstance(StopIterationType) {
					break Outer
				}
				f.RestoreExc(nil, nil)
				return nil, raised
			}
			elems[i] = elem
		}
		result = append(result, NewTuple(elems...).ToObject())
	}
	return NewList(result...).ToObject(), nil
}

func init() {
	builtinMap := map[string]*Object{
		"__frame__":      newBuiltinFunction("__frame__", builtinFrame).ToObject(),
		"abs":            newBuiltinFunction("abs", builtinAbs).ToObject(),
		"all":            newBuiltinFunction("all", builtinAll).ToObject(),
		"any":            newBuiltinFunction("any", builtinAny).ToObject(),
		"bin":            newBuiltinFunction("bin", builtinBin).ToObject(),
		"callable":       newBuiltinFunction("callable", builtinCallable).ToObject(),
		"chr":            newBuiltinFunction("chr", builtinChr).ToObject(),
		"cmp":            newBuiltinFunction("cmp", builtinCmp).ToObject(),
		"dir":            newBuiltinFunction("dir", builtinDir).ToObject(),
		"False":          False.ToObject(),
		"getattr":        newBuiltinFunction("getattr", builtinGetAttr).ToObject(),
		"globals":        newBuiltinFunction("globals", builtinGlobals).ToObject(),
		"hasattr":        newBuiltinFunction("hasattr", builtinHasAttr).ToObject(),
		"hash":           newBuiltinFunction("hash", builtinHash).ToObject(),
		"hex":            newBuiltinFunction("hex", builtinHex).ToObject(),
		"id":             newBuiltinFunction("id", builtinID).ToObject(),
		"isinstance":     newBuiltinFunction("isinstance", builtinIsInstance).ToObject(),
		"issubclass":     newBuiltinFunction("issubclass", builtinIsSubclass).ToObject(),
		"iter":           newBuiltinFunction("iter", builtinIter).ToObject(),
		"len":            newBuiltinFunction("len", builtinLen).ToObject(),
		"map":            newBuiltinFunction("map", builtinMapFn).ToObject(),
		"max":            newBuiltinFunction("max", builtinMax).ToObject(),
		"min":            newBuiltinFunction("min", builtinMin).ToObject(),
		"next":           newBuiltinFunction("next", builtinNext).ToObject(),
		"None":           None,
		"NotImplemented": NotImplemented,
		"oct":            newBuiltinFunction("oct", builtinOct).ToObject(),
		"open":           newBuiltinFunction("open", builtinOpen).ToObject(),
		"ord":            newBuiltinFunction("ord", builtinOrd).ToObject(),
		"print":          newBuiltinFunction("print", builtinPrint).ToObject(),
		"range":          newBuiltinFunction("range", builtinRange).ToObject(),
		"repr":           newBuiltinFunction("repr", builtinRepr).ToObject(),
		"sorted":         newBuiltinFunction("sorted", builtinSorted).ToObject(),
		"True":           True.ToObject(),
		"unichr":         newBuiltinFunction("unichr", builtinUniChr).ToObject(),
		"zip":            newBuiltinFunction("zip", builtinZip).ToObject(),
	}
	// Do type initialization in two phases so that we don't have to think
	// about hard-to-understand cycles.
	for typ, info := range builtinTypes {
		initBuiltinType(typ, info)
		if info.global {
			builtinMap[typ.name] = typ.ToObject()
		}
	}
	for name := range builtinMap {
		InternStr(name)
	}
	Builtins = newStringDict(builtinMap)
}

// builtinMinMax implements the builtin min/max() functions. When doMax is
// true, the max is found, otherwise the min is found. There are two forms of
// the builtins. The first takes a single iterable argument and the result is
// the min/max of the elements of that sequence. The second form takes two or
// more args and returns the min/max of those. For more details see:
// https://docs.python.org/2/library/functions.html#min
func builtinMinMax(f *Frame, doMax bool, args Args, kwargs KWArgs) (*Object, *BaseException) {
	name := "min"
	if doMax {
		name = "max"
	}
	if raised := checkFunctionVarArgs(f, name, args, ObjectType); raised != nil {
		return nil, raised
	}
	keyFunc := kwargs.get("key", nil)
	// selected is the min/max element found so far.
	var selected, selectedKey *Object
	partialFunc := func(o *Object) (raised *BaseException) {
		oKey := o
		if keyFunc != nil {
			oKey, raised = keyFunc.Call(f, Args{o}, nil)
			if raised != nil {
				return raised
			}
		}
		// sel dictates whether o is the new min/max. It defaults to
		// true when selected == nil (we don't yet have a selection).
		sel := true
		if selected != nil {
			result, raised := LT(f, selectedKey, oKey)
			if raised != nil {
				return raised
			}
			lt, raised := IsTrue(f, result)
			if raised != nil {
				return raised
			}
			// Select o when looking for max and selection < o, or
			// when looking for min and o < selection.
			sel = doMax && lt || !doMax && !lt
		}
		if sel {
			selected = o
			selectedKey = oKey
		}
		return nil
	}
	if len(args) == 1 {
		// Take min/max of the single iterable arg passed.
		if raised := seqForEach(f, args[0], partialFunc); raised != nil {
			return nil, raised
		}
		if selected == nil {
			return nil, f.RaiseType(ValueErrorType, fmt.Sprintf("%s() arg is an empty sequence", name))
		}
	} else {
		// Take min/max of the passed args.
		for _, arg := range args {
			if raised := partialFunc(arg); raised != nil {
				return nil, raised
			}
		}
	}
	return selected, nil
}

// numberToBase implements the builtins "bin", "hex", and "oct".
// base must be between 2 and 36, and o must be an instance of
// IntType or LongType.
func numberToBase(prefix string, base int, o *Object) string {
	z := big.Int{}
	switch {
	case o.isInstance(LongType):
		z = toLongUnsafe(o).value
	case o.isInstance(IntType):
		z.SetInt64(int64(toIntUnsafe(o).Value()))
	default:
		panic("numberToBase requires an Int or Long argument")
	}
	s := z.Text(base)
	if s[0] == '-' {
		// Move the negative sign before the prefix.
		return "-" + prefix + s[1:]
	}
	return prefix + s
}

// initIters return list of initiated Iter instances from the list of
// iterables.
func initIters(f *Frame, items []*Object) ([]*Object, *BaseException) {
	l := len(items)
	iters := make([]*Object, l)
	for i, arg := range items {
		iter, raised := Iter(f, arg)
		if raised != nil {
			return nil, raised
		}
		iters[i] = iter
	}
	return iters, nil
}

// zipLongest return the list of aggregates elements from each of the
// iterables. If the iterables are of uneven length, missing values are
// filled-in with None.
func zipLongest(f *Frame, args Args) ([][]*Object, *BaseException) {
	argc := len(args)
	result := make([][]*Object, 0, 2)
	iters, raised := initIters(f, args)
	if raised != nil {
		return nil, raised
	}

	for {
		noItems := true
		elems := make([]*Object, argc)
		for i, iter := range iters {
			if iter == nil {
				continue
			}
			elem, raised := Next(f, iter)
			if raised != nil {
				if raised.isInstance(StopIterationType) {
					iters[i] = nil
					elems[i] = None
					continue
				}
				f.RestoreExc(nil, nil)
				return nil, raised
			}
			noItems = false
			elems[i] = elem
		}
		if noItems {
			break
		}
		result = append(result, elems)
	}
	return result, nil
}

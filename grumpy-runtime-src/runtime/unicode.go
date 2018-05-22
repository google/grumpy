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
	"reflect"
	"unicode"
	"unicode/utf8"
)

var (
	// UnicodeType is the object representing the Python 'unicode' type.
	UnicodeType = newBasisType("unicode", reflect.TypeOf(Unicode{}), toUnicodeUnsafe, BaseStringType)
)

// Unicode represents Python 'unicode' objects. The string value is stored as
// utf-32 data.
type Unicode struct {
	Object
	value []rune
}

// NewUnicode returns a new Unicode holding the given string value. value is
// assumed to be a valid utf-8 string.
func NewUnicode(value string) *Unicode {
	return NewUnicodeFromRunes(bytes.Runes([]byte(value)))
}

// NewUnicodeFromRunes returns a new Unicode holding the given runes.
func NewUnicodeFromRunes(value []rune) *Unicode {
	return &Unicode{Object{typ: UnicodeType}, value}
}

func toUnicodeUnsafe(o *Object) *Unicode {
	return (*Unicode)(o.toPointer())
}

// Encode translates the runes in s into a str with the given encoding.
//
// NOTE: If s contains surrogates (e.g. U+D800), Encode will raise
// UnicodeDecodeError consistent with CPython 3.x but different than 2.x.
func (s *Unicode) Encode(f *Frame, encoding, errors string) (*Str, *BaseException) {
	// TODO: Support custom encodings and error handlers.
	normalized := normalizeEncoding(encoding)
	if normalized != "utf8" {
		return nil, f.RaiseType(LookupErrorType, fmt.Sprintf("unknown encoding: %s", encoding))
	}
	buf := bytes.Buffer{}
	for i, r := range s.Value() {
		switch {
		case utf8.ValidRune(r):
			buf.WriteRune(r)
		case errors == EncodeIgnore:
			// Do nothing
		case errors == EncodeReplace:
			buf.WriteRune(unicode.ReplacementChar)
		case errors == EncodeStrict:
			format := "'%s' codec can't encode character %s in position %d"
			return nil, f.RaiseType(UnicodeEncodeErrorType, fmt.Sprintf(format, encoding, escapeRune(r), i))
		default:
			format := "unknown error handler name '%s'"
			return nil, f.RaiseType(LookupErrorType, fmt.Sprintf(format, errors))
		}
	}
	return NewStr(buf.String()), nil
}

// ToObject upcasts s to an Object.
func (s *Unicode) ToObject() *Object {
	return &s.Object
}

// Value returns the underlying string value held by s.
func (s *Unicode) Value() []rune {
	return s.value
}

func unicodeAdd(f *Frame, v, w *Object) (*Object, *BaseException) {
	unicodeV := toUnicodeUnsafe(v)
	unicodeW, raised := unicodeCoerce(f, w)
	if raised != nil {
		return nil, raised
	}
	lenV := len(unicodeV.Value())
	newLen := lenV + len(unicodeW.Value())
	if newLen < 0 {
		return nil, f.RaiseType(OverflowErrorType, errResultTooLarge)
	}
	value := make([]rune, newLen)
	copy(value, unicodeV.Value())
	copy(value[lenV:], unicodeW.Value())
	return NewUnicodeFromRunes(value).ToObject(), nil
}

func unicodeContains(f *Frame, o *Object, value *Object) (*Object, *BaseException) {
	lhs := toUnicodeUnsafe(o).Value()
	s, raised := unicodeCoerce(f, value)
	if raised != nil {
		return nil, raised
	}
	rhs := s.Value()
	lhsLen, rhsLen := len(lhs), len(rhs)
	maxOffset := lhsLen - rhsLen
	for offset := 0; offset <= maxOffset; offset++ {
		if runeSliceCmp(lhs[offset:offset+rhsLen], rhs) == 0 {
			return True.ToObject(), nil
		}
	}
	return False.ToObject(), nil
}

func unicodeEncode(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	// TODO: Accept unicode for encoding and errors args.
	expectedTypes := []*Type{UnicodeType, StrType, StrType}
	argc := len(args)
	if argc >= 1 && argc < 3 {
		expectedTypes = expectedTypes[:argc]
	}
	if raised := checkMethodArgs(f, "encode", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	encoding := EncodeDefault
	if argc > 1 {
		encoding = toStrUnsafe(args[1]).Value()
	}
	errors := EncodeStrict
	if argc > 2 {
		errors = toStrUnsafe(args[2]).Value()
	}
	ret, raised := toUnicodeUnsafe(args[0]).Encode(f, encoding, errors)
	if raised != nil {
		return nil, raised
	}
	return ret.ToObject(), nil
}

func unicodeEq(f *Frame, v, w *Object) (*Object, *BaseException) {
	return unicodeCompareEq(f, toUnicodeUnsafe(v), w, true)
}

func unicodeGE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return unicodeCompare(f, toUnicodeUnsafe(v), w, False, True, True)
}

// unicodeGetItem returns a slice of string depending on whether index is an
// integer or a slice. If index is neither of those types then a TypeError is
// returned.
func unicodeGetItem(f *Frame, o, key *Object) (*Object, *BaseException) {
	s := toUnicodeUnsafe(o).Value()
	switch {
	case key.typ.slots.Index != nil:
		index, raised := seqCheckedIndex(f, len(s), toIntUnsafe(key).Value())
		if raised != nil {
			return nil, raised
		}
		return NewUnicodeFromRunes([]rune{s[index]}).ToObject(), nil
	case key.isInstance(SliceType):
		slice := toSliceUnsafe(key)
		start, stop, step, sliceLen, raised := slice.calcSlice(f, len(s))
		if raised != nil {
			return nil, raised
		}
		if step == 1 {
			return NewUnicodeFromRunes(s[start:stop]).ToObject(), nil
		}
		result := make([]rune, 0, sliceLen)
		for j := start; j < stop; j += step {
			result = append(result, s[j])
		}
		return NewUnicodeFromRunes([]rune(result)).ToObject(), nil
	}
	return nil, f.RaiseType(TypeErrorType, fmt.Sprintf("unicode indices must be integers or slice, not %s", key.typ.Name()))
}

func unicodeGetNewArgs(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "__getnewargs__", args, UnicodeType); raised != nil {
		return nil, raised
	}
	return NewTuple1(args[0]).ToObject(), nil
}

func unicodeGT(f *Frame, v, w *Object) (*Object, *BaseException) {
	return unicodeCompare(f, toUnicodeUnsafe(v), w, False, False, True)
}

func unicodeHash(f *Frame, o *Object) (*Object, *BaseException) {
	s := toUnicodeUnsafe(o).Value()
	l := len(s)
	if l == 0 {
		return NewInt(0).ToObject(), nil
	}
	h := int(s[0]) << 7
	for _, r := range s {
		h = (1000003 * h) ^ int(r)
	}
	h ^= l
	if h == -1 {
		h = -2
	}
	return NewInt(h).ToObject(), nil
}

func unicodeJoin(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "join", args, UnicodeType, ObjectType); raised != nil {
		return nil, raised
	}
	var result *Object
	raised := seqApply(f, args[1], func(parts []*Object, _ bool) (raised *BaseException) {
		result, raised = unicodeJoinParts(f, toUnicodeUnsafe(args[0]), parts)
		return raised
	})
	if raised != nil {
		return nil, raised
	}
	return result, nil
}

func unicodeLE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return unicodeCompare(f, toUnicodeUnsafe(v), w, True, True, False)
}

func unicodeLen(f *Frame, o *Object) (*Object, *BaseException) {
	return NewInt(len(toUnicodeUnsafe(o).Value())).ToObject(), nil
}

func unicodeLT(f *Frame, v, w *Object) (*Object, *BaseException) {
	return unicodeCompare(f, toUnicodeUnsafe(v), w, True, False, False)
}

func unicodeMul(f *Frame, v, w *Object) (*Object, *BaseException) {
	value := toUnicodeUnsafe(v).Value()
	numChars := len(value)
	n, ok, raised := strRepeatCount(f, numChars, w)
	if raised != nil {
		return nil, raised
	}
	if !ok {
		return NotImplemented, nil
	}
	newLen := numChars * n
	newValue := make([]rune, newLen)
	for i := 0; i < newLen; i += numChars {
		copy(newValue[i:], value)
	}
	return NewUnicodeFromRunes(newValue).ToObject(), nil
}

func unicodeNative(f *Frame, o *Object) (reflect.Value, *BaseException) {
	// Encode to utf-8 when passing data out to Go.
	s, raised := toUnicodeUnsafe(o).Encode(f, EncodeDefault, EncodeStrict)
	if raised != nil {
		return reflect.Value{}, raised
	}
	return reflect.ValueOf(s.Value()), nil
}

func unicodeNE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return unicodeCompareEq(f, toUnicodeUnsafe(v), w, false)
}

func unicodeNew(f *Frame, t *Type, args Args, _ KWArgs) (ret *Object, raised *BaseException) {
	// TODO: Accept keyword arguments: string, encoding, errors.
	if t != UnicodeType {
		// Allocate a plain unicode then copy it's value into an object
		// of the unicode subtype.
		s, raised := unicodeNew(f, UnicodeType, args, nil)
		if raised != nil {
			return nil, raised
		}
		result := toUnicodeUnsafe(newObject(t))
		result.value = toUnicodeUnsafe(s).Value()
		return result.ToObject(), nil
	}
	expectedTypes := []*Type{ObjectType, StrType, StrType}
	argc := len(args)
	if argc < 3 {
		expectedTypes = expectedTypes[:argc]
	}
	if raised := checkMethodArgs(f, "__new__", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	if argc == 0 {
		return NewUnicodeFromRunes(nil).ToObject(), nil
	}
	arg0 := args[0]
	if argc == 1 {
		if unicode := arg0.typ.slots.Unicode; unicode != nil {
			ret, raised = unicode.Fn(f, arg0)
		} else if arg0.typ == UnicodeType {
			ret = toUnicodeUnsafe(arg0).ToObject()
		} else if arg0.isInstance(UnicodeType) {
			// Return a unicode object (not a subtype).
			ret = NewUnicodeFromRunes(toUnicodeUnsafe(arg0).Value()).ToObject()
		} else if str := arg0.typ.slots.Str; str != nil {
			ret, raised = str.Fn(f, arg0)
		} else {
			var s *Str
			if s, raised = Repr(f, arg0); raised == nil {
				ret = s.ToObject()
			}
		}
		if raised != nil {
			return nil, raised
		}
		u, raised := unicodeCoerce(f, ret)
		if raised != nil {
			return nil, raised
		}
		return u.ToObject(), nil
	}
	if !arg0.isInstance(StrType) {
		format := "coercing to Unicode: need str, %s found"
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, arg0.typ.Name()))
	}
	encoding := toStrUnsafe(args[1]).Value()
	errors := "strict"
	if argc > 2 {
		errors = toStrUnsafe(args[2]).Value()
	}
	s, raised := toStrUnsafe(arg0).Decode(f, encoding, errors)
	if raised != nil {
		return nil, raised
	}
	return s.ToObject(), nil
}

func unicodeRepr(_ *Frame, o *Object) (*Object, *BaseException) {
	buf := bytes.Buffer{}
	buf.WriteString("u'")
	for _, r := range toUnicodeUnsafe(o).Value() {
		if escape, ok := escapeMap[r]; ok {
			buf.WriteString(escape)
		} else if r <= unicode.MaxASCII && unicode.IsPrint(r) {
			buf.WriteRune(r)
		} else {
			buf.Write(escapeRune(r))
		}
	}
	buf.WriteRune('\'')
	return NewStr(buf.String()).ToObject(), nil
}

func unicodeStr(f *Frame, o *Object) (*Object, *BaseException) {
	ret, raised := toUnicodeUnsafe(o).Encode(f, EncodeDefault, EncodeStrict)
	if raised != nil {
		return nil, raised
	}
	return ret.ToObject(), nil
}

func unicodeStrip(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	expectedTypes := []*Type{UnicodeType, ObjectType}
	argc := len(args)
	if argc == 1 {
		expectedTypes = expectedTypes[:argc]
	}
	if raised := checkMethodArgs(f, "strip", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	s := toUnicodeUnsafe(args[0])
	charsArg := None
	if argc > 1 {
		charsArg = args[1]
	}
	matchFunc := unicode.IsSpace
	if charsArg != None {
		chars, raised := unicodeCoerce(f, charsArg)
		if raised != nil {
			return nil, raised
		}
		matchFunc = func(r rune) bool {
			for _, c := range chars.Value() {
				if r == c {
					return true
				}
			}
			return false
		}
	}
	runes := s.Value()
	numRunes := len(runes)
	lindex := 0
	for ; lindex < numRunes; lindex++ {
		if !matchFunc(runes[lindex]) {
			break
		}
	}
	rindex := numRunes
	for ; rindex > lindex; rindex-- {
		if !matchFunc(runes[rindex-1]) {
			break
		}
	}
	result := make([]rune, rindex-lindex)
	copy(result, runes[lindex:rindex])
	return NewUnicodeFromRunes(result).ToObject(), nil
}

func initUnicodeType(dict map[string]*Object) {
	dict["__getnewargs__"] = newBuiltinFunction("__getnewargs__", unicodeGetNewArgs).ToObject()
	dict["encode"] = newBuiltinFunction("encode", unicodeEncode).ToObject()
	dict["join"] = newBuiltinFunction("join", unicodeJoin).ToObject()
	dict["strip"] = newBuiltinFunction("strip", unicodeStrip).ToObject()
	UnicodeType.slots.Add = &binaryOpSlot{unicodeAdd}
	UnicodeType.slots.Contains = &binaryOpSlot{unicodeContains}
	UnicodeType.slots.Eq = &binaryOpSlot{unicodeEq}
	UnicodeType.slots.GE = &binaryOpSlot{unicodeGE}
	UnicodeType.slots.GetItem = &binaryOpSlot{unicodeGetItem}
	UnicodeType.slots.GT = &binaryOpSlot{unicodeGT}
	UnicodeType.slots.Hash = &unaryOpSlot{unicodeHash}
	UnicodeType.slots.LE = &binaryOpSlot{unicodeLE}
	UnicodeType.slots.Len = &unaryOpSlot{unicodeLen}
	UnicodeType.slots.LT = &binaryOpSlot{unicodeLT}
	UnicodeType.slots.Mul = &binaryOpSlot{unicodeMul}
	UnicodeType.slots.NE = &binaryOpSlot{unicodeNE}
	UnicodeType.slots.New = &newSlot{unicodeNew}
	UnicodeType.slots.Native = &nativeSlot{unicodeNative}
	UnicodeType.slots.RMul = &binaryOpSlot{unicodeMul}
	UnicodeType.slots.Repr = &unaryOpSlot{unicodeRepr}
	UnicodeType.slots.Str = &unaryOpSlot{unicodeStr}
}

func unicodeCompare(f *Frame, v *Unicode, w *Object, ltResult, eqResult, gtResult *Int) (*Object, *BaseException) {
	rhs := []rune(nil)
	if w.isInstance(UnicodeType) {
		rhs = toUnicodeUnsafe(w).Value()
	} else if w.isInstance(StrType) {
		ret, raised := toStrUnsafe(w).Decode(f, EncodeDefault, EncodeStrict)
		if raised != nil {
			return nil, raised
		}
		rhs = ret.Value()
	} else {
		return NotImplemented, nil
	}
	switch runeSliceCmp(v.Value(), rhs) {
	case -1:
		return ltResult.ToObject(), nil
	case 0:
		return eqResult.ToObject(), nil
	default:
		return gtResult.ToObject(), nil
	}
}

func runeSliceCmp(lhs []rune, rhs []rune) int {
	lhsLen, rhsLen := len(lhs), len(rhs)
	minLen := lhsLen
	if rhsLen < lhsLen {
		minLen = rhsLen
	}
	for i := 0; i < minLen; i++ {
		if lhs[i] < rhs[i] {
			return -1
		}
		if lhs[i] > rhs[i] {
			return 1
		}
	}
	if lhsLen < rhsLen {
		return -1
	}
	if lhsLen > rhsLen {
		return 1
	}
	return 0
}

// unicodeCompareEq returns the result of comparing whether v and w are equal
// (when eq is true) or unequal (when eq is false). It differs from
// unicodeCompare in that it will safely decode w if it has type str and
// therefore will not raise UnicodeDecodeError.
func unicodeCompareEq(f *Frame, v *Unicode, w *Object, eq bool) (*Object, *BaseException) {
	if w.isInstance(UnicodeType) {
		// Do the standard comparison knowing that we won't raise
		// UnicodeDecodeError for w.
		return unicodeCompare(f, v, w, GetBool(!eq), GetBool(eq), GetBool(!eq))
	}
	if !w.isInstance(StrType) {
		return NotImplemented, nil
	}
	lhs := v.Value()
	lhsLen := len(lhs)
	i := 0
	// Decode w as utf-8.
	for _, r := range toStrUnsafe(w).Value() {
		// lhs[i] should never be RuneError so the second part of the
		// condition should catch that case.
		if i >= lhsLen || lhs[i] != r {
			return GetBool(!eq).ToObject(), nil
		}
		i++
	}
	return GetBool((i == lhsLen) == eq).ToObject(), nil
}

func unicodeCoerce(f *Frame, o *Object) (*Unicode, *BaseException) {
	switch {
	case o.isInstance(StrType):
		return toStrUnsafe(o).Decode(f, EncodeDefault, EncodeStrict)
	case o.isInstance(UnicodeType):
		return toUnicodeUnsafe(o), nil
	default:
		format := "coercing to Unicode: need string, %s found"
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, o.typ.Name()))
	}
}

func unicodeJoinParts(f *Frame, s *Unicode, parts []*Object) (*Object, *BaseException) {
	numParts := len(parts)
	if numParts == 0 {
		return NewUnicode("").ToObject(), nil
	}
	sep := s.Value()
	sepLen := len(sep)
	unicodeParts := make([]*Unicode, numParts)
	// Calculate the size of the required buffer.
	numRunes := (numParts - 1) * len(sep)
	for i, part := range parts {
		s, raised := unicodeCoerce(f, part)
		if raised != nil {
			return nil, raised
		}
		unicodeParts[i] = s
		numRunes += len(s.Value())
	}
	// Piece together the result string into buf.
	buf := make([]rune, numRunes)
	offset := 0
	for i, part := range unicodeParts {
		if i > 0 {
			copy(buf[offset:offset+sepLen], sep)
			offset += sepLen
		}
		s := part.Value()
		l := len(s)
		copy(buf[offset:offset+l], s)
		offset += l
	}
	return NewUnicodeFromRunes(buf).ToObject(), nil
}

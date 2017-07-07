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
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

var (
	// StrType is the object representing the Python 'str' type.
	StrType                = newBasisType("str", reflect.TypeOf(Str{}), toStrUnsafe, BaseStringType)
	whitespaceSplitRegexp  = regexp.MustCompile(`\s+`)
	strASCIISpaces         = []byte(" \t\n\v\f\r")
	strInterpolationRegexp = regexp.MustCompile(`^%([#0 +-]?)((\*|[0-9]+)?)((\.(\*|[0-9]+))?)[hlL]?([diouxXeEfFgGcrs%])`)
	internedStrs           = map[string]*Str{}
	caseOffset             = byte('a' - 'A')

	internedName = NewStr("__name__")
)

type stripSide int

const (
	stripSideLeft stripSide = iota
	stripSideRight
	stripSideBoth
)

// InternStr adds s to the interned string map. Subsequent calls to NewStr()
// will return the same underlying Str. InternStr is not thread safe and should
// only be called during module initialization time.
func InternStr(s string) *Str {
	str, _ := internedStrs[s]
	if str == nil {
		str = &Str{Object: Object{typ: StrType}, value: s, hash: NewInt(hashString(s))}
		internedStrs[s] = str
	}
	return str
}

// Str represents Python 'str' objects.
type Str struct {
	Object
	value string
	hash  *Int
}

// NewStr returns a new Str holding the given string value.
func NewStr(value string) *Str {
	if s := internedStrs[value]; s != nil {
		return s
	}
	return &Str{Object: Object{typ: StrType}, value: value}
}

func toStrUnsafe(o *Object) *Str {
	return (*Str)(o.toPointer())
}

// Decode produces a unicode object from the bytes of s assuming they have the
// given encoding. Invalid code points are resolved using a strategy given by
// errors: "ignore" will bypass them, "replace" will substitute the Unicode
// replacement character (U+FFFD) and "strict" will raise UnicodeDecodeError.
//
// NOTE: Decoding UTF-8 data containing surrogates (e.g. U+D800 encoded as
// '\xed\xa0\x80') will raise UnicodeDecodeError consistent with CPython 3.x
// but different than 2.x.
func (s *Str) Decode(f *Frame, encoding, errors string) (*Unicode, *BaseException) {
	// TODO: Support custom encodings and error handlers.
	normalized := normalizeEncoding(encoding)
	if normalized != "utf8" {
		return nil, f.RaiseType(LookupErrorType, fmt.Sprintf("unknown encoding: %s", encoding))
	}
	var runes []rune
	for pos, r := range s.Value() {
		switch {
		case r != utf8.RuneError:
			runes = append(runes, r)
		case errors == EncodeIgnore:
			// Do nothing
		case errors == EncodeReplace:
			runes = append(runes, unicode.ReplacementChar)
		case errors == EncodeStrict:
			format := "'%s' codec can't decode byte 0x%02x in position %d"
			return nil, f.RaiseType(UnicodeDecodeErrorType, fmt.Sprintf(format, encoding, int(s.Value()[pos]), pos))
		default:
			format := "unknown error handler name '%s'"
			return nil, f.RaiseType(LookupErrorType, fmt.Sprintf(format, errors))
		}
	}
	return NewUnicodeFromRunes(runes), nil
}

// ToObject upcasts s to an Object.
func (s *Str) ToObject() *Object {
	return &s.Object
}

// Value returns the underlying string value held by s.
func (s *Str) Value() string {
	return s.value
}

func hashString(s string) int {
	l := len(s)
	if l == 0 {
		return 0
	}
	h := int(s[0]) << 7
	for i := 0; i < l; i++ {
		h = (1000003 * h) ^ int(s[i])
	}
	h ^= l
	if h == -1 {
		h = -2
	}
	return h
}

func strAdd(f *Frame, v, w *Object) (*Object, *BaseException) {
	if w.isInstance(UnicodeType) {
		// CPython explicitly dispatches to unicode here so that's how
		// we do it even though it would seem more natural to override
		// unicode.__radd__.
		ret, raised := toStrUnsafe(v).Decode(f, EncodeDefault, EncodeStrict)
		if raised != nil {
			return nil, raised
		}
		return unicodeAdd(f, ret.ToObject(), w)
	}
	if !w.isInstance(StrType) {
		return NotImplemented, nil
	}
	stringV, stringW := toStrUnsafe(v).Value(), toStrUnsafe(w).Value()
	if len(stringV)+len(stringW) < 0 {
		// This indicates an int overflow.
		return nil, f.RaiseType(OverflowErrorType, errResultTooLarge)
	}
	return NewStr(stringV + stringW).ToObject(), nil
}

func strCapitalize(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "capitalize", args, StrType); raised != nil {
		return nil, raised
	}
	s := toStrUnsafe(args[0]).Value()
	numBytes := len(s)
	if numBytes == 0 {
		return args[0], nil
	}
	b := make([]byte, numBytes)
	b[0] = toUpper(s[0])
	for i := 1; i < numBytes; i++ {
		b[i] = toLower(s[i])
	}
	return NewStr(string(b)).ToObject(), nil
}

func strCenter(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	s, width, fill, raised := strJustDecodeArgs(f, args, "center")
	if raised != nil {
		return nil, raised
	}
	if len(s) >= width {
		return NewStr(s).ToObject(), nil
	}
	marg := width - len(s)
	left := marg/2 + (marg & width & 1)
	return NewStr(pad(s, left, marg-left, fill)).ToObject(), nil
}

func strContains(f *Frame, o *Object, value *Object) (*Object, *BaseException) {
	if value.isInstance(UnicodeType) {
		decoded, raised := toStrUnsafe(o).Decode(f, EncodeDefault, EncodeStrict)
		if raised != nil {
			return nil, raised
		}
		return unicodeContains(f, decoded.ToObject(), value)
	}
	if !value.isInstance(StrType) {
		format := "'in <string>' requires string as left operand, not %s"
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, value.typ.Name()))
	}
	return GetBool(strings.Contains(toStrUnsafe(o).Value(), toStrUnsafe(value).Value())).ToObject(), nil
}

func strCount(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "count", args, StrType, ObjectType); raised != nil {
		return nil, raised
	}
	s := toStrUnsafe(args[0]).Value()
	sep := toStrUnsafe(args[1]).Value()
	cnt := strings.Count(s, sep)
	return NewInt(cnt).ToObject(), nil
}

func strDecode(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	// TODO: Accept unicode for encoding and errors args.
	expectedTypes := []*Type{StrType, StrType, StrType}
	argc := len(args)
	if argc >= 1 && argc < 3 {
		expectedTypes = expectedTypes[:argc]
	}
	if raised := checkMethodArgs(f, "decode", args, expectedTypes...); raised != nil {
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
	s, raised := toStrUnsafe(args[0]).Decode(f, encoding, errors)
	if raised != nil {
		return nil, raised
	}
	return s.ToObject(), nil
}

func strEndsWith(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	return strStartsEndsWith(f, "endswith", args)
}

func strEq(f *Frame, v, w *Object) (*Object, *BaseException) {
	return strCompare(v, w, False, True, False), nil
}

// strFind returns the lowest index in s where the substring sub is found such
// that sub is wholly contained in s[start:end]. Return -1 on failure.
func strFind(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	return strFindOrIndex(f, args, func(s, sub string) (int, *BaseException) {
		return strings.Index(s, sub), nil
	})
}

func strGE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return strCompare(v, w, False, True, True), nil
}

// strGetItem returns a slice of string depending on whether index is an integer
// or a slice. If index is neither of those types then a TypeError is returned.
func strGetItem(f *Frame, o, key *Object) (*Object, *BaseException) {
	s := toStrUnsafe(o).Value()
	switch {
	case key.typ.slots.Index != nil:
		index, raised := IndexInt(f, key)
		if raised != nil {
			return nil, raised
		}
		index, raised = seqCheckedIndex(f, len(s), index)
		if raised != nil {
			return nil, raised
		}
		return NewStr(s[index : index+1]).ToObject(), nil
	case key.isInstance(SliceType):
		slice := toSliceUnsafe(key)
		start, stop, step, sliceLen, raised := slice.calcSlice(f, len(s))
		if raised != nil {
			return nil, raised
		}
		if step == 1 {
			return NewStr(s[start:stop]).ToObject(), nil
		}
		result := make([]byte, 0, sliceLen)
		for j := start; j != stop; j += step {
			result = append(result, s[j])
		}
		return NewStr(string(result)).ToObject(), nil
	}
	return nil, f.RaiseType(TypeErrorType, fmt.Sprintf("string indices must be integers or slice, not %s", key.typ.Name()))
}

func strGetNewArgs(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "__getnewargs__", args, StrType); raised != nil {
		return nil, raised
	}
	return NewTuple1(args[0]).ToObject(), nil
}

func strGT(f *Frame, v, w *Object) (*Object, *BaseException) {
	return strCompare(v, w, False, False, True), nil
}

func strHash(f *Frame, o *Object) (*Object, *BaseException) {
	s := toStrUnsafe(o)
	p := (*unsafe.Pointer)(unsafe.Pointer(&s.hash))
	if v := atomic.LoadPointer(p); v != unsafe.Pointer(nil) {
		return (*Int)(v).ToObject(), nil
	}
	h := NewInt(hashString(toStrUnsafe(o).Value()))
	atomic.StorePointer(p, unsafe.Pointer(h))
	return h.ToObject(), nil
}

func strIndex(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	return strFindOrIndex(f, args, func(s, sub string) (i int, raised *BaseException) {
		i = strings.Index(s, sub)
		if i == -1 {
			raised = f.RaiseType(ValueErrorType, "substring not found")
		}
		return i, raised
	})
}

func strIsAlNum(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "isalnum", args, StrType); raised != nil {
		return nil, raised
	}
	s := toStrUnsafe(args[0]).Value()
	if len(s) == 0 {
		return False.ToObject(), nil
	}
	for i := range s {
		if !isAlNum(s[i]) {
			return False.ToObject(), nil
		}
	}
	return True.ToObject(), nil
}

func strIsAlpha(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "isalpha", args, StrType); raised != nil {
		return nil, raised
	}
	s := toStrUnsafe(args[0]).Value()
	if len(s) == 0 {
		return False.ToObject(), nil
	}
	for i := range s {
		if !isAlpha(s[i]) {
			return False.ToObject(), nil
		}
	}
	return True.ToObject(), nil
}

func strIsDigit(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "isdigit", args, StrType); raised != nil {
		return nil, raised
	}
	s := toStrUnsafe(args[0]).Value()
	if len(s) == 0 {
		return False.ToObject(), nil
	}
	for i := range s {
		if !isDigit(s[i]) {
			return False.ToObject(), nil
		}
	}
	return True.ToObject(), nil
}

func strIsLower(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "islower", args, StrType); raised != nil {
		return nil, raised
	}
	s := toStrUnsafe(args[0]).Value()
	if len(s) == 0 {
		return False.ToObject(), nil
	}
	for i := range s {
		if !isLower(s[i]) {
			return False.ToObject(), nil
		}
	}
	return True.ToObject(), nil
}

func strIsSpace(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "isspace", args, StrType); raised != nil {
		return nil, raised
	}
	s := toStrUnsafe(args[0]).Value()
	if len(s) == 0 {
		return False.ToObject(), nil
	}
	for i := range s {
		if !isSpace(s[i]) {
			return False.ToObject(), nil
		}
	}
	return True.ToObject(), nil
}

func strIsTitle(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "istitle", args, StrType); raised != nil {
		return nil, raised
	}

	s := toStrUnsafe(args[0]).Value()
	if len(s) == 0 {
		return False.ToObject(), nil
	}

	if len(s) == 1 {
		return GetBool(isUpper(s[0])).ToObject(), nil
	}

	cased := false
	previousIsCased := false

	for i := range s {
		if isUpper(s[i]) {
			if previousIsCased {
				return False.ToObject(), nil
			}
			previousIsCased = true
			cased = true
		} else if isLower(s[i]) {
			if !previousIsCased {
				return False.ToObject(), nil
			}
			previousIsCased = true
			cased = true
		} else {
			previousIsCased = false
		}
	}

	return GetBool(cased).ToObject(), nil
}

func strIsUpper(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "isupper", args, StrType); raised != nil {
		return nil, raised
	}
	s := toStrUnsafe(args[0]).Value()
	if len(s) == 0 {
		return False.ToObject(), nil
	}
	for i := range s {
		if !isUpper(s[i]) {
			return False.ToObject(), nil
		}
	}
	return True.ToObject(), nil
}

func strJoin(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "join", args, StrType, ObjectType); raised != nil {
		return nil, raised
	}
	sep := toStrUnsafe(args[0]).Value()
	var result *Object
	raised := seqApply(f, args[1], func(parts []*Object, _ bool) *BaseException {
		numParts := len(parts)
		if numParts == 0 {
			result = NewStr("").ToObject()
			return nil
		}
		// Calculate the size of the required buffer.
		numChars := (numParts - 1) * len(sep)
		for i, part := range parts {
			if part.isInstance(StrType) {
				numChars += len(toStrUnsafe(part).Value())
			} else if part.isInstance(UnicodeType) {
				// Some element was unicode so use the unicode
				// implementation.
				var raised *BaseException
				s, raised := unicodeCoerce(f, args[0])
				if raised != nil {
					return raised
				}
				result, raised = unicodeJoinParts(f, s, parts)
				return raised
			} else {
				format := "sequence item %d: expected string, %s found"
				return f.RaiseType(TypeErrorType, fmt.Sprintf(format, i, part.typ.Name()))
			}
		}
		// Piece together the result string into buf.
		buf := bytes.Buffer{}
		buf.Grow(numChars)
		for i, part := range parts {
			if i > 0 {
				buf.WriteString(sep)
			}
			buf.WriteString(toStrUnsafe(part).Value())
		}
		result = NewStr(buf.String()).ToObject()
		return nil
	})
	if raised != nil {
		return nil, raised
	}
	return result, nil
}

func strLE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return strCompare(v, w, True, True, False), nil
}

func strLen(f *Frame, o *Object) (*Object, *BaseException) {
	return NewInt(len(toStrUnsafe(o).Value())).ToObject(), nil
}

func strLJust(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	s, width, fill, raised := strJustDecodeArgs(f, args, "ljust")
	if raised != nil {
		return nil, raised
	}
	if len(s) >= width {
		return NewStr(s).ToObject(), nil
	}
	return NewStr(pad(s, 0, width-len(s), fill)).ToObject(), nil
}

func strLower(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	expectedTypes := []*Type{StrType}
	if raised := checkMethodArgs(f, "lower", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	s := toStrUnsafe(args[0]).Value()
	numBytes := len(s)
	if numBytes == 0 {
		return args[0], nil
	}
	b := make([]byte, numBytes)
	for i := 0; i < numBytes; i++ {
		b[i] = toLower(s[i])
	}
	return NewStr(string(b)).ToObject(), nil
}

func strLStrip(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	return strStripImpl(f, args, stripSideLeft)
}

func strLT(f *Frame, v, w *Object) (*Object, *BaseException) {
	return strCompare(v, w, True, False, False), nil
}

func strMod(f *Frame, v, w *Object) (*Object, *BaseException) {
	s := toStrUnsafe(v).Value()
	switch {
	case w.isInstance(DictType):
		return nil, f.RaiseType(NotImplementedErrorType, "mappings not yet supported")
	case w.isInstance(TupleType):
		return strInterpolate(f, s, toTupleUnsafe(w))
	default:
		return strInterpolate(f, s, NewTuple1(w))
	}
}

func strMul(f *Frame, v, w *Object) (*Object, *BaseException) {
	s := toStrUnsafe(v).Value()
	n, ok, raised := strRepeatCount(f, len(s), w)
	if raised != nil {
		return nil, raised
	}
	if !ok {
		return NotImplemented, nil
	}
	return NewStr(strings.Repeat(s, n)).ToObject(), nil
}

func strNative(f *Frame, o *Object) (reflect.Value, *BaseException) {
	return reflect.ValueOf(toStrUnsafe(o).Value()), nil
}

func strNE(f *Frame, v, w *Object) (*Object, *BaseException) {
	return strCompare(v, w, True, False, True), nil
}

func strNew(f *Frame, t *Type, args Args, _ KWArgs) (*Object, *BaseException) {
	if t != StrType {
		// Allocate a plain str and then copy it's value into an object
		// of the str subtype.
		s, raised := strNew(f, StrType, args, nil)
		if raised != nil {
			return nil, raised
		}
		result := toStrUnsafe(newObject(t))
		result.value = toStrUnsafe(s).Value()
		return result.ToObject(), nil
	}
	argc := len(args)
	if argc == 0 {
		// Empty string.
		return newObject(t), nil
	}
	if argc != 1 {
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf("str() takes at most 1 argument (%d given)", argc))
	}
	o := args[0]
	if str := o.typ.slots.Str; str != nil {
		result, raised := str.Fn(f, o)
		if raised != nil {
			return nil, raised
		}
		if !result.isInstance(StrType) {
			format := "__str__ returned non-string (type %s)"
			return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, result.typ.Name()))
		}
		return result, nil
	}
	s, raised := Repr(f, o)
	if raised != nil {
		return nil, raised
	}
	return s.ToObject(), nil
}

// strReplace returns a copy of the string s with the first n non-overlapping
// instances of old replaced by sub. If old is empty, it matches at the
// beginning of the string. If n < 0, there is no limit on the number of
// replacements.
func strReplace(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	var raised *BaseException
	// TODO: Support unicode replace.
	expectedTypes := []*Type{StrType, StrType, StrType, ObjectType}
	argc := len(args)
	if argc == 3 {
		expectedTypes = expectedTypes[:argc]
	}
	if raised := checkMethodArgs(f, "replace", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	n := -1
	if argc == 4 {
		n, raised = ToIntValue(f, args[3])
		if raised != nil {
			return nil, raised
		}
	}
	s := toStrUnsafe(args[0]).Value()
	// Returns early if no need to replace.
	if n == 0 {
		return NewStr(s).ToObject(), nil
	}

	old := toStrUnsafe(args[1]).Value()
	sub := toStrUnsafe(args[2]).Value()
	numBytes := len(s)
	// Even if s and old is blank, replace should return sub, except n is negative.
	// This is CPython specific behavior.
	if numBytes == 0 && old == "" && n >= 0 {
		return NewStr("").ToObject(), nil
	}
	// If old is non-blank, pass to strings.Replace.
	if len(old) > 0 {
		return NewStr(strings.Replace(s, old, sub, n)).ToObject(), nil
	}

	// If old is blank, insert sub after every bytes on s and beginning.
	if n < 0 {
		n = numBytes + 1
	}
	// Insert sub at beginning.
	buf := bytes.Buffer{}
	buf.WriteString(sub)
	n--
	// Insert after every byte.
	i := 0
	for n > 0 && i < numBytes {
		buf.WriteByte(s[i])
		buf.WriteString(sub)
		i++
		n--
	}
	// Write the remaining string.
	if i < numBytes {
		buf.WriteString(s[i:])
	}
	return NewStr(buf.String()).ToObject(), nil
}

func strRepr(_ *Frame, o *Object) (*Object, *BaseException) {
	s := toStrUnsafe(o).Value()
	buf := bytes.Buffer{}
	buf.WriteRune('\'')
	numBytes := len(s)
	for i := 0; i < numBytes; i++ {
		r := rune(s[i])
		if escape, ok := escapeMap[r]; ok {
			buf.WriteString(escape)
		} else if r > unicode.MaxASCII || !unicode.IsPrint(r) {
			buf.WriteString(fmt.Sprintf(`\x%02x`, r))
		} else {
			buf.WriteRune(r)
		}
	}
	buf.WriteRune('\'')
	return NewStr(buf.String()).ToObject(), nil
}

func strRFind(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	return strFindOrIndex(f, args, func(s, sub string) (int, *BaseException) {
		return strings.LastIndex(s, sub), nil
	})
}

func strRIndex(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	return strFindOrIndex(f, args, func(s, sub string) (i int, raised *BaseException) {
		i = strings.LastIndex(s, sub)
		if i == -1 {
			raised = f.RaiseType(ValueErrorType, "substring not found")
		}
		return i, raised
	})
}

func strRJust(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	s, width, fill, raised := strJustDecodeArgs(f, args, "rjust")
	if raised != nil {
		return nil, raised
	}
	if len(s) >= width {
		return NewStr(s).ToObject(), nil
	}
	return NewStr(pad(s, width-len(s), 0, fill)).ToObject(), nil
}

func strSplit(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	expectedTypes := []*Type{StrType, ObjectType, IntType}
	argc := len(args)
	if argc == 1 || argc == 2 {
		expectedTypes = expectedTypes[:argc]
	}
	if raised := checkMethodArgs(f, "split", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	sep := ""
	if argc > 1 {
		if arg1 := args[1]; arg1.isInstance(StrType) {
			sep = toStrUnsafe(arg1).Value()
			if sep == "" {
				return nil, f.RaiseType(ValueErrorType, "empty separator")
			}
		} else if arg1 != None {
			return nil, f.RaiseType(TypeErrorType, "expected a str separator")
		}
	}
	maxSplit := -1
	if argc > 2 {
		if i := toIntUnsafe(args[2]).Value(); i >= 0 {
			maxSplit = i + 1
		}
	}
	s := toStrUnsafe(args[0]).Value()
	var parts []string
	if sep == "" {
		s = strings.TrimLeft(s, string(strASCIISpaces))
		parts = whitespaceSplitRegexp.Split(s, maxSplit)
		l := len(parts)
		if l > 0 && strings.Trim(parts[l-1], string(strASCIISpaces)) == "" {
			parts = parts[:l-1]
		}
	} else {
		parts = strings.SplitN(s, sep, maxSplit)
	}
	results := make([]*Object, len(parts))
	for i, part := range parts {
		results[i] = NewStr(part).ToObject()
	}
	return NewList(results...).ToObject(), nil
}

func strSplitLines(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	expectedTypes := []*Type{StrType, ObjectType}
	argc := len(args)
	if argc == 1 {
		expectedTypes = expectedTypes[:1]
	}
	if raised := checkMethodArgs(f, "splitlines", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	keepEnds := false
	if argc == 2 {
		i, raised := ToIntValue(f, args[1])
		if raised != nil {
			return nil, raised
		}
		keepEnds = i != 0
	}
	s := toStrUnsafe(args[0]).Value()
	numChars := len(s)
	start, end := 0, 0
	lines := make([]*Object, 0, 2)
	for start < numChars {
		eol := 0
		for end = start; end < numChars; end++ {
			c := s[end]
			if c == '\n' {
				eol = end + 1
				break
			}
			if c == '\r' {
				eol = end + 1
				if eol < numChars && s[eol] == '\n' {
					eol++
				}
				break
			}
		}
		if end >= numChars {
			eol = end
		}
		line := ""
		if keepEnds {
			line = s[start:eol]
		} else {
			line = s[start:end]
		}
		lines = append(lines, NewStr(line).ToObject())
		start = eol
	}
	return NewList(lines...).ToObject(), nil
}

func strStrip(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	return strStripImpl(f, args, stripSideBoth)
}

func strRStrip(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	return strStripImpl(f, args, stripSideRight)
}

func strStripImpl(f *Frame, args Args, side stripSide) (*Object, *BaseException) {
	expectedTypes := []*Type{StrType, ObjectType}
	argc := len(args)
	if argc == 1 {
		expectedTypes = expectedTypes[:argc]
	}
	if raised := checkMethodArgs(f, "strip", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	s := toStrUnsafe(args[0])
	charsArg := None
	if argc > 1 {
		charsArg = args[1]
	}
	var chars []byte
	switch {
	case charsArg.isInstance(UnicodeType):
		u, raised := s.Decode(f, EncodeDefault, EncodeStrict)
		if raised != nil {
			return nil, raised
		}
		return unicodeStrip(f, Args{u.ToObject(), charsArg}, nil)
	case charsArg.isInstance(StrType):
		chars = []byte(toStrUnsafe(charsArg).Value())
	case charsArg == None:
		chars = strASCIISpaces
	default:
		return nil, f.RaiseType(TypeErrorType, "strip arg must be None, str or unicode")
	}
	byteSlice := []byte(s.Value())
	numBytes := len(byteSlice)
	lindex := 0
	if side == stripSideLeft || side == stripSideBoth {
	LeftStrip:
		for ; lindex < numBytes; lindex++ {
			b := byteSlice[lindex]
			for _, c := range chars {
				if b == c {
					continue LeftStrip
				}
			}
			break
		}
	}
	rindex := numBytes
	if side == stripSideRight || side == stripSideBoth {
	RightStrip:
		for ; rindex > lindex; rindex-- {
			b := byteSlice[rindex-1]
			for _, c := range chars {
				if b == c {
					continue RightStrip
				}
			}
			break
		}
	}
	return NewStr(string(byteSlice[lindex:rindex])).ToObject(), nil
}

func strStartsWith(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	return strStartsEndsWith(f, "startswith", args)
}

func strStr(_ *Frame, o *Object) (*Object, *BaseException) {
	if o.typ == StrType {
		return o, nil
	}
	return NewStr(toStrUnsafe(o).Value()).ToObject(), nil
}

func strSwapCase(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "swapcase", args, StrType); raised != nil {
		return nil, raised
	}
	s := toStrUnsafe(args[0]).Value()
	numBytes := len(s)
	if numBytes == 0 {
		return args[0], nil
	}
	b := make([]byte, numBytes)
	for i := 0; i < numBytes; i++ {
		if isLower(s[i]) {
			b[i] = toUpper(s[i])
		} else if isUpper(s[i]) {
			b[i] = toLower(s[i])
		} else {
			b[i] = s[i]
		}
	}
	return NewStr(string(b)).ToObject(), nil
}

func initStrType(dict map[string]*Object) {
	dict["__getnewargs__"] = newBuiltinFunction("__getnewargs__", strGetNewArgs).ToObject()
	dict["capitalize"] = newBuiltinFunction("capitalize", strCapitalize).ToObject()
	dict["count"] = newBuiltinFunction("count", strCount).ToObject()
	dict["center"] = newBuiltinFunction("center", strCenter).ToObject()
	dict["decode"] = newBuiltinFunction("decode", strDecode).ToObject()
	dict["endswith"] = newBuiltinFunction("endswith", strEndsWith).ToObject()
	dict["find"] = newBuiltinFunction("find", strFind).ToObject()
	dict["index"] = newBuiltinFunction("index", strIndex).ToObject()
	dict["isalnum"] = newBuiltinFunction("isalnum", strIsAlNum).ToObject()
	dict["isalpha"] = newBuiltinFunction("isalpha", strIsAlpha).ToObject()
	dict["isdigit"] = newBuiltinFunction("isdigit", strIsDigit).ToObject()
	dict["islower"] = newBuiltinFunction("islower", strIsLower).ToObject()
	dict["isspace"] = newBuiltinFunction("isspace", strIsSpace).ToObject()
	dict["istitle"] = newBuiltinFunction("istitle", strIsTitle).ToObject()
	dict["isupper"] = newBuiltinFunction("isupper", strIsUpper).ToObject()
	dict["join"] = newBuiltinFunction("join", strJoin).ToObject()
	dict["lower"] = newBuiltinFunction("lower", strLower).ToObject()
	dict["ljust"] = newBuiltinFunction("ljust", strLJust).ToObject()
	dict["lstrip"] = newBuiltinFunction("lstrip", strLStrip).ToObject()
	dict["rfind"] = newBuiltinFunction("rfind", strRFind).ToObject()
	dict["rindex"] = newBuiltinFunction("rindex", strRIndex).ToObject()
	dict["rjust"] = newBuiltinFunction("rjust", strRJust).ToObject()
	dict["split"] = newBuiltinFunction("split", strSplit).ToObject()
	dict["splitlines"] = newBuiltinFunction("splitlines", strSplitLines).ToObject()
	dict["startswith"] = newBuiltinFunction("startswith", strStartsWith).ToObject()
	dict["strip"] = newBuiltinFunction("strip", strStrip).ToObject()
	dict["swapcase"] = newBuiltinFunction("swapcase", strSwapCase).ToObject()
	dict["replace"] = newBuiltinFunction("replace", strReplace).ToObject()
	dict["rstrip"] = newBuiltinFunction("rstrip", strRStrip).ToObject()
	dict["title"] = newBuiltinFunction("title", strTitle).ToObject()
	dict["upper"] = newBuiltinFunction("upper", strUpper).ToObject()
	dict["zfill"] = newBuiltinFunction("zfill", strZFill).ToObject()
	StrType.slots.Add = &binaryOpSlot{strAdd}
	StrType.slots.Contains = &binaryOpSlot{strContains}
	StrType.slots.Eq = &binaryOpSlot{strEq}
	StrType.slots.GE = &binaryOpSlot{strGE}
	StrType.slots.GetItem = &binaryOpSlot{strGetItem}
	StrType.slots.GT = &binaryOpSlot{strGT}
	StrType.slots.Hash = &unaryOpSlot{strHash}
	StrType.slots.LE = &binaryOpSlot{strLE}
	StrType.slots.Len = &unaryOpSlot{strLen}
	StrType.slots.LT = &binaryOpSlot{strLT}
	StrType.slots.Mod = &binaryOpSlot{strMod}
	StrType.slots.Mul = &binaryOpSlot{strMul}
	StrType.slots.NE = &binaryOpSlot{strNE}
	StrType.slots.New = &newSlot{strNew}
	StrType.slots.Native = &nativeSlot{strNative}
	StrType.slots.Repr = &unaryOpSlot{strRepr}
	StrType.slots.RMul = &binaryOpSlot{strMul}
	StrType.slots.Str = &unaryOpSlot{strStr}
}

func strCompare(v, w *Object, ltResult, eqResult, gtResult *Int) *Object {
	if v == w {
		return eqResult.ToObject()
	}
	if !w.isInstance(StrType) {
		return NotImplemented
	}
	s1 := toStrUnsafe(v).Value()
	s2 := toStrUnsafe(w).Value()
	if s1 < s2 {
		return ltResult.ToObject()
	}
	if s1 == s2 {
		return eqResult.ToObject()
	}
	return gtResult.ToObject()
}

func strInterpolate(f *Frame, format string, values *Tuple) (*Object, *BaseException) {
	var buf bytes.Buffer
	valueIndex := 0
	index := strings.Index(format, "%")
	for index != -1 {
		buf.WriteString(format[:index])
		format = format[index:]
		matches := strInterpolationRegexp.FindStringSubmatch(format)
		if matches == nil {
			return nil, f.RaiseType(ValueErrorType, "invalid format spec")
		}
		flags, fieldType := matches[1], matches[7]
		if fieldType != "%" && valueIndex >= len(values.elems) {
			return nil, f.RaiseType(TypeErrorType, "not enough arguments for format string")
		}
		fieldWidth := -1
		if matches[2] == "*" || matches[4] != "" {
			return nil, f.RaiseType(NotImplementedErrorType, "field width not yet supported")
		}
		if matches[2] != "" {
			var err error
			fieldWidth, err = strconv.Atoi(matches[2])
			if err != nil {
				return nil, f.RaiseType(TypeErrorType, fmt.Sprint(err))
			}
		}
		if flags != "" && flags != "0" {
			return nil, f.RaiseType(NotImplementedErrorType, "conversion flags not yet supported")
		}
		var val string
		switch fieldType {
		case "r", "s":
			o := values.elems[valueIndex]
			var s *Str
			var raised *BaseException
			if fieldType == "r" {
				s, raised = Repr(f, o)
			} else {
				s, raised = ToStr(f, o)
			}
			if raised != nil {
				return nil, raised
			}
			val = s.Value()
			if fieldWidth > 0 {
				val = strLeftPad(val, fieldWidth, " ")
			}
			buf.WriteString(val)
			valueIndex++
		case "f":
			o := values.elems[valueIndex]
			if v, ok := floatCoerce(o); ok {
				val := strconv.FormatFloat(v, 'f', 6, 64)
				if fieldWidth > 0 {
					fillchar := " "
					if flags != "" {
						fillchar = flags
					}
					val = strLeftPad(val, fieldWidth, fillchar)
				}
				buf.WriteString(val)
				valueIndex++
			} else {
				return nil, f.RaiseType(TypeErrorType, fmt.Sprintf("float argument required, not %s", o.typ.Name()))
			}
		case "d", "x", "X", "o":
			o := values.elems[valueIndex]
			i, raised := ToInt(f, values.elems[valueIndex])
			if raised != nil {
				return nil, raised
			}
			if fieldType == "d" {
				s, raised := ToStr(f, i)
				if raised != nil {
					return nil, raised
				}
				val = s.Value()
			} else if matches[7] == "o" {
				if o.isInstance(LongType) {
					val = toLongUnsafe(o).Value().Text(8)
				} else {
					val = strconv.FormatInt(int64(toIntUnsafe(i).Value()), 8)
				}
			} else {
				if o.isInstance(LongType) {
					val = toLongUnsafe(o).Value().Text(16)
				} else {
					val = strconv.FormatInt(int64(toIntUnsafe(i).Value()), 16)
				}
				if fieldType == "X" {
					val = strings.ToUpper(val)
				}
			}
			if fieldWidth > 0 {
				fillchar := " "
				if flags != "" {
					fillchar = flags
				}
				val = strLeftPad(val, fieldWidth, fillchar)
			}
			buf.WriteString(val)
			valueIndex++
		case "%":
			val = "%"
			if fieldWidth > 0 {
				val = strLeftPad(val, fieldWidth, " ")
			}
			buf.WriteString(val)
		default:
			format := "conversion type not yet supported: %s"
			return nil, f.RaiseType(NotImplementedErrorType, fmt.Sprintf(format, fieldType))
		}
		format = format[len(matches[0]):]
		index = strings.Index(format, "%")
	}
	if valueIndex < len(values.elems) {
		return nil, f.RaiseType(TypeErrorType, "not all arguments converted during string formatting")
	}
	buf.WriteString(format)
	return NewStr(buf.String()).ToObject(), nil
}

func strRepeatCount(f *Frame, numChars int, mult *Object) (int, bool, *BaseException) {
	var n int
	switch {
	case mult.isInstance(IntType):
		n = toIntUnsafe(mult).Value()
	case mult.isInstance(LongType):
		l := toLongUnsafe(mult).Value()
		if !numInIntRange(l) {
			return 0, false, f.RaiseType(OverflowErrorType, fmt.Sprintf("cannot fit '%s' into an index-sized integer", mult.typ.Name()))
		}
		n = int(l.Int64())
	default:
		return 0, false, nil
	}
	if n <= 0 {
		return 0, true, nil
	}
	if numChars > MaxInt/n {
		return 0, false, f.RaiseType(OverflowErrorType, errResultTooLarge)
	}
	return n, true, nil
}

func adjustIndex(start, end, length int) (int, int) {
	if end > length {
		end = length
	} else if end < 0 {
		end += length
		if end < 0 {
			end = 0
		}
	}
	if start < 0 {
		start += length
		if start < 0 {
			start = 0
		}
	}
	return start, end
}

func strStartsEndsWith(f *Frame, method string, args Args) (*Object, *BaseException) {
	expectedTypes := []*Type{StrType, ObjectType, IntType, IntType}
	argc := len(args)
	if argc == 2 || argc == 3 {
		expectedTypes = expectedTypes[:argc]
	}
	if raised := checkMethodArgs(f, method, args, expectedTypes...); raised != nil {
		return nil, raised
	}
	matchesArg := args[1]
	var matches []string
	switch {
	case matchesArg.isInstance(TupleType):
		elems := toTupleUnsafe(matchesArg).elems
		matches = make([]string, len(elems))
		for i, o := range elems {
			if !o.isInstance(BaseStringType) {
				return nil, f.RaiseType(TypeErrorType, "expected a str")
			}
			s, raised := ToStr(f, o)
			if raised != nil {
				return nil, raised
			}
			matches[i] = s.Value()
		}
	case matchesArg.isInstance(BaseStringType):
		s, raised := ToStr(f, matchesArg)
		if raised != nil {
			return nil, raised
		}
		matches = []string{s.Value()}
	default:
		msg := " first arg must be str, unicode, or tuple, not "
		return nil, f.RaiseType(TypeErrorType, method+msg+matchesArg.typ.Name())
	}
	s := toStrUnsafe(args[0]).Value()
	l := len(s)
	start, end := 0, l
	if argc >= 3 {
		start = toIntUnsafe(args[2]).Value()
	}
	if argc == 4 {
		end = toIntUnsafe(args[3]).Value()
	}
	start, end = adjustIndex(start, end, l)
	if start > end {
		// start == end may still return true when matching ''.
		return False.ToObject(), nil
	}
	s = s[start:end]
	matcher := strings.HasPrefix
	if method == "endswith" {
		matcher = strings.HasSuffix
	}
	for _, match := range matches {
		if matcher(s, match) {
			return True.ToObject(), nil
		}
	}
	return False.ToObject(), nil
}

func strTitle(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	expectedTypes := []*Type{StrType}
	if raised := checkMethodArgs(f, "title", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	s := toStrUnsafe(args[0]).Value()
	numBytes := len(s)
	if numBytes == 0 {
		return args[0], nil
	}
	b := make([]byte, numBytes)
	previousIsCased := false
	for i := 0; i < numBytes; i++ {
		c := s[i]
		switch {
		case isLower(c):
			if !previousIsCased {
				c = toUpper(c)
			}
			previousIsCased = true
		case isUpper(c):
			if previousIsCased {
				c = toLower(c)
			}
			previousIsCased = true
		default:
			previousIsCased = false
		}
		b[i] = c
	}
	return NewStr(string(b)).ToObject(), nil
}

func strUpper(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	expectedTypes := []*Type{StrType}
	if raised := checkMethodArgs(f, "upper", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	s := toStrUnsafe(args[0]).Value()
	numBytes := len(s)
	if numBytes == 0 {
		return args[0], nil
	}
	b := make([]byte, numBytes)
	for i := 0; i < numBytes; i++ {
		b[i] = toUpper(s[i])
	}
	return NewStr(string(b)).ToObject(), nil
}

func strZFill(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "zfill", args, StrType, ObjectType); raised != nil {
		return nil, raised
	}
	s := toStrUnsafe(args[0]).Value()
	width, raised := ToIntValue(f, args[1])
	if raised != nil {
		return nil, raised
	}
	return NewStr(strLeftPad(s, width, "0")).ToObject(), nil
}

func init() {
	InternStr("")
	for i := 0; i < 256; i++ {
		InternStr(string([]byte{byte(i)}))
	}
}

func toLower(b byte) byte {
	if isUpper(b) {
		return b + caseOffset
	}
	return b
}

func toUpper(b byte) byte {
	if isLower(b) {
		return b - caseOffset
	}
	return b
}

func isAlNum(c byte) bool {
	return isAlpha(c) || isDigit(c)
}

func isAlpha(c byte) bool {
	return isUpper(c) || isLower(c)
}

func isDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

func isLower(c byte) bool {
	return 'a' <= c && c <= 'z'
}

func isSpace(c byte) bool {
	switch c {
	case ' ', '\n', '\t', '\v', '\f', '\r':
		return true
	default:
		return false
	}
}

func isUpper(c byte) bool {
	return 'A' <= c && c <= 'Z'
}

func pad(s string, left int, right int, fillchar string) string {
	buf := bytes.Buffer{}

	if left < 0 {
		left = 0
	}

	if right < 0 {
		right = 0
	}

	if left == 0 && right == 0 {
		return s
	}

	buf.Grow(left + len(s) + right)
	buf.WriteString(strings.Repeat(fillchar, left))
	buf.WriteString(s)
	buf.WriteString(strings.Repeat(fillchar, right))

	return buf.String()
}

// strLeftPad returns s padded with fillchar so that its length is at least width.
// Fillchar must be a single character. When fillchar is "0", s starting with a
// sign are handled correctly.
func strLeftPad(s string, width int, fillchar string) string {
	l := len(s)
	if width <= l {
		return s
	}
	buf := bytes.Buffer{}
	buf.Grow(width)
	if l > 0 && fillchar == "0" && (s[0] == '-' || s[0] == '+') {
		buf.WriteByte(s[0])
		s = s[1:]
		l = len(s)
		width--
	}
	// TODO: Support or throw fillchar len more than one.
	buf.WriteString(strings.Repeat(fillchar, width-l))
	buf.WriteString(s)
	return buf.String()
}

type indexFunc func(string, string) (int, *BaseException)

func strFindOrIndex(f *Frame, args Args, fn indexFunc) (*Object, *BaseException) {
	// TODO: Support for unicode substring.
	expectedTypes := []*Type{StrType, StrType, ObjectType, ObjectType}
	argc := len(args)
	if argc == 2 || argc == 3 {
		expectedTypes = expectedTypes[:argc]
	}
	if raised := checkMethodArgs(f, "find/index", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	s := toStrUnsafe(args[0]).Value()
	l := len(s)
	start, end := 0, l
	var raised *BaseException
	if argc >= 3 && args[2] != None {
		start, raised = IndexInt(f, args[2])
		if raised != nil {
			return nil, raised
		}
	}
	if argc == 4 && args[3] != None {
		end, raised = IndexInt(f, args[3])
		if raised != nil {
			return nil, raised
		}
	}
	// Default to an impossible search.
	search, sub := "", "-"
	if start <= l {
		start, end = adjustIndex(start, end, l)
		if start <= end {
			sub = toStrUnsafe(args[1]).Value()
			search = s[start:end]
		}
	}
	index, raised := fn(search, sub)
	if raised != nil {
		return nil, raised
	}
	if index != -1 {
		index += start
	}
	return NewInt(index).ToObject(), nil
}

func strJustDecodeArgs(f *Frame, args Args, name string) (string, int, string, *BaseException) {
	expectedTypes := []*Type{StrType, IntType, StrType}
	if raised := checkMethodArgs(f, name, args, expectedTypes...); raised != nil {
		return "", 0, "", raised
	}
	s := toStrUnsafe(args[0]).Value()
	width := toIntUnsafe(args[1]).Value()
	fill := toStrUnsafe(args[2]).Value()

	if numChars := len(fill); numChars != 1 {
		return s, width, fill, f.RaiseType(TypeErrorType, fmt.Sprintf("%[1]s() argument 2 must be char, not str", name))
	}

	return s, width, fill, nil
}

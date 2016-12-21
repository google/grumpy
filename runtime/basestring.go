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
	"regexp"
	"strings"
)

// EncodeDefault is the system default encoding.
const EncodeDefault = "utf8"

// Error handling modes that dictate the behavior of *Str.Decode and
// *Unicode.Encode when they encounter bad chars.
const (
	// EncodeStrict causes UnicodeError to be raised on bad chars.
	EncodeStrict = "strict"
	// EncodeReplace replaces bad chars with "\ufffd".
	EncodeReplace = "replace"
	// EncodeIgnore discards bad chars.
	EncodeIgnore = "ignore"
)

var (
	// BaseStringType is the object representing the Python 'basestring'
	// type.
	BaseStringType        = newSimpleType("basestring", ObjectType)
	encodingGarbageRegexp = regexp.MustCompile(`[^A-Za-z0-9]+`)
	escapeMap             = map[rune]string{
		'\\': `\\`,
		'\'': `\'`,
		'\n': `\n`,
		'\r': `\r`,
		'\t': `\t`,
	}
)

func initBaseStringType(map[string]*Object) {
	BaseStringType.flags &^= typeFlagInstantiable
}

func normalizeEncoding(encoding string) string {
	return strings.ToLower(encodingGarbageRegexp.ReplaceAllString(encoding, ""))
}

func escapeRune(r rune) []byte {
	const hexTable = "0123456789abcdef"

	if r < 0x100 {
		return []byte{'\\', 'x', hexTable[r>>4], hexTable[r&0x0F]}
	}

	if r < 0x10000 {
		return []byte{'\\', 'u',
			hexTable[r>>12], hexTable[r>>8&0x0F],
			hexTable[r>>4&0x0F], hexTable[r&0x0F]}
	}

	return []byte{'\\', 'U',
		hexTable[r>>28], hexTable[r>>24&0x0F],
		hexTable[r>>20&0x0F], hexTable[r>>16&0x0F],
		hexTable[r>>12&0x0F], hexTable[r>>8&0x0F],
		hexTable[r>>4&0x0F], hexTable[r&0x0F]}
}

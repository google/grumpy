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
	"testing"
)

func TestNormalizeEncoding(t *testing.T) {
	cases := []struct {
		encoding string
		want     string
	}{
		{"utf8", "utf8"},
		{"UTF-16  ", "utf16"},
		{"  __Ascii__", "ascii"},
		{"utf@#(%*#(*%16  ", "utf16"},
		{"", ""},
	}
	for _, cas := range cases {
		if got := normalizeEncoding(cas.encoding); got != cas.want {
			t.Errorf("normalizeEncoding(%q) = %q, want %q", cas.encoding, got, cas.want)
		}
	}
}

func BenchmarkEscapeRune(b *testing.B) {
	b.Run("low values", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			escapeRune(0x10)
		}
	})

	b.Run("mid values", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			escapeRune(0x200)
		}
	})

	b.Run("high values", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			escapeRune(0x20000)
		}
	})
}

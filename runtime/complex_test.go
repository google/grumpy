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

func TestComplexRepr(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(complex(0.0, 0.0)), want: NewStr("0j").ToObject()},
		{args: wrapArgs(complex(0.0, 1.0)), want: NewStr("1j").ToObject()},
		{args: wrapArgs(complex(1.0, 2.0)), want: NewStr("(1+2j)").ToObject()},
		{args: wrapArgs(complex(3.1, -4.2)), want: NewStr("(3.1-4.2j)").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(Repr), &cas); err != "" {
			t.Error(err)
		}
	}
}

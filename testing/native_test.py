# Copyright 2016 Google Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# pylint: disable=g-multiple-import

from '__go__/math' import MaxInt32, Pow10, Signbit
from '__go__/strings' import Count, IndexAny, Repeat
from '__go__/encoding/csv' import NewReader as NewCSVReader
from '__go__/image' import Pt
from '__go__/strings' import NewReader as NewStringReader

assert Count('foo,bar,baz', ',') == 2
assert IndexAny('foobar', 'obr') == 1
assert Repeat('foo', 3) == 'foofoofoo'
assert MaxInt32 == 2147483647
assert Pow10(2.0) == 100.0
assert Signbit(-42.0) == True  # pylint: disable=g-explicit-bool-comparison

# Can access field on unreferenced struct (Pt returns an image.Point struct)
assert Pt(1, 0).X == 1

# Can access field on pointer to struct (NewCSVReader returns a pointer to a
# csv.Reader struct)
assert NewCSVReader(NewStringReader("foo")).LazyQuotes == False

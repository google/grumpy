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

import types


assert isinstance((x for x in ()), types.GeneratorType)
assert list(c for c in 'abc') == ['a', 'b', 'c']
assert [c for c in 'abc'] == ['a', 'b', 'c']
assert [i + j for i in range(2) for j in range(2)] == [0, 1, 1, 2]
assert [c for c in 'foobar' if c in 'aeiou'] == ['o', 'o', 'a']
assert {i: str(i) for i in range(2)} == {0: '0', 1: '1'}

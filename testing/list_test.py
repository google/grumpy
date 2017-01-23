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

a = [0, 1, 2, 3]
b = list(a)
assert a == b
assert a is not b
assert list(()) == []
assert list((0, 1, 2, 3)) == [0, 1, 2, 3]
assert list('') == []
assert list('spam') == ['s', 'p', 'a', 'm']

assert [] is not True
assert [42]

assert [] is not []

assert len([]) == 0
assert len([0]) == 1
assert len([0, 1, 2]) == 3

a = [3, 2, 4, 1]
b = []
c = ["a", "e", "c", "b"]

a.sort()
assert a == [1, 2, 3, 4]
b.sort()
assert b == []
c.sort()
assert c == ["a", "b", "c", "e"]

# Test pop
a = [-1, 0, 1]
assert a.pop() == 1
assert a == [-1, 0]
assert a == [-1, 0]
assert a.pop(0) == -1
assert a == [0]
try:
  a.pop(5)
  assert AssertionError
except IndexError:
  pass
assert a.pop(0) == 0
assert a == []
try:
  a.pop()
  assert AssertionError
except IndexError:
  pass
try:
  a.pop(42, 42)
  assert AssertionError
except TypeError:
  pass
a = [-1, 0, 1]
assert a.pop(1) == 0
assert a == [-1, 1]

# Test extend
a = aa = [3, 2, 4, 1]
b = bb = []
c = cc = ["a", "e", "c", "b"]
a.extend(b)
assert a == [3, 2, 4, 1]
assert a == aa
assert a is aa
b.extend(c)
assert b == ["a", "e", "c", "b"]
assert b is bb
a.extend(tuple())
assert a == [3, 2, 4, 1]
a.extend((6, 7))
assert a == [3, 2, 4, 1, 6, 7]
a.extend(range(3))
assert a == [3, 2, 4, 1, 6, 7, 0, 1, 2]

try:
  a.extend()
  assert AssertionError
except TypeError:
  pass

try:
  a.extend([], [])
  assert AssertionError
except TypeError:
  pass

# Test count
assert [].count(0) == 0
assert [1, 2, 3].count(2) == 1
assert ["a", "b", "a", "a"].count("a") == 3
assert ([2] * 20).count(2) == 20

try:
  [].count()
  assert AssertionError
except TypeError:
  pass

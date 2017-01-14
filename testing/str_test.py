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

# pylint: disable=redefined-outer-name,no-value-for-parameter

import sys


# Test Add
assert "foo" + "bar" == "foobar"
assert "foo" + u"bar" == u"foobar"
assert "baz" + "" == "baz"

# Test Mod
assert "%s" % 42 == "42"
assert "%f" % 3.14 == "3.140000"
assert "abc %d" % 123L == "abc 123"
assert "%d" % 3.14 == "3"
assert "%%" % tuple() == "%"
assert "%r" % "abc" == "'abc'"
assert "%x" % 0x1f == "1f"
assert "%X" % 0xffff == "FFFF"

# Test find
assert "".find("") == 0
assert "".find("", 1) == -1
assert "foobar".find("bar") == 3
assert "foobar".find("bar", 0, -2) == -1
assert "foobar".find("foo", 0, 3) == 0
assert "foobar".find("bar", 3, 5) == -1
assert "foobar".find("bar", 5, 3) == -1
assert "bar".find("foobar") == -1
assert 'abcdefghiabc'.find('abc') == 0
assert 'abcdefghiabc'.find('abc', 1) == 9
assert 'abcdefghiabc'.find('def', 4) == -1
assert 'abc'.find('', 0) == 0
assert 'abc'.find('', 3) == 3
assert 'abc'.find('', 4) == -1
assert 'rrarrrrrrrrra'.find('a') == 2
assert 'rrarrrrrrrrra'.find('a', 4) == 12
assert 'rrarrrrrrrrra'.find('a', 4, 6) == -1
assert ''.find('') == 0
assert ''.find('', 1, 1) == -1
assert ''.find('', sys.maxint, 0) == -1
assert ''.find('xx') == -1
assert ''.find('xx', 1, 1) == -1
assert ''.find('xx', sys.maxint, 0) == -1
# TODO: support long
# assert 'ab'.find('xxx', sys.maxsize + 1, 0) == -1
# TODO: support unicode
# assert "foobar".find(u"bar") == 3
# TODO: support None
# assert 'rrarrrrrrrrra'.find('a', 4, None) == 12
# assert 'rrarrrrrrrrra'.find('a', None, 6) == 2
try:
  "foo".find(123)
  raise AssertionError
except TypeError:
  pass

try:
  'hello'.find()
  raise AssertionError
except TypeError:
  pass

try:
  'hello'.find(42)
  raise AssertionError
except TypeError:
  pass


# Test index
assert "".index("") == 0
assert "foobar".index("bar") == 3
assert "foobar".index("foo", 0, 3) == 0
# TODO: support unicode
# assert "foobar".find(u"bar") == 3
try:
  "".index("", 1)
  raise AssertionError
except ValueError:
  pass
try:
  "foo".index(123)
  raise AssertionError
except TypeError:
  pass


# Test zfill
assert '123'.zfill(2) == '123'
assert '123'.zfill(3) == '123'
assert '123'.zfill(4) == '0123'
assert '+123'.zfill(3) == '+123'
assert '+123'.zfill(4) == '+123'
assert '+123'.zfill(5) == '+0123'
assert '-123'.zfill(3) == '-123'
assert '-123'.zfill(4) == '-123'
assert '-123'.zfill(5) == '-0123'
assert ''.zfill(3) == '000'
assert '34'.zfill(1) == '34'
assert '34'.zfill(4) == '0034'

try:
  '123'.zfill()
  raise AssertionError
except TypeError:
  pass

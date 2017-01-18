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

# pylint: disable=redefined-outer-name

import sys

# Test Add
assert "foo" + "bar" == "foobar"
assert "foo" + u"bar" == u"foobar"
assert "baz" + "" == "baz"

# Test find
assert "".find("") == 0
assert "".find("", 1) == -1
assert "foobar".find("bar") == 3
assert "foobar".find("bar", 0, -2) == -1
assert "foobar".find("foo", 0, 3) == 0
assert "foobar".find("bar", 3, 5) == -1
assert "foobar".find("bar", 5, 3) == -1
assert "bar".find("foobar") == -1
assert "bar".find("a", 0, -1) == 1
assert 'abcdefghiabc'.find('abc') == 0
assert 'abcdefghiabc'.find('abc', 1) == 9
assert 'abcdefghiabc'.find('def', 4) == -1
assert 'abc'.find('', 0) == 0
assert 'abc'.find('', 3) == 3
assert 'abc'.find('c', long(1)) == 2
assert 'abc'.find('c', 0, long(3)) == 2
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
# TODO: Support unicode substring.
# assert "foobar".find(u"bar") == 3
# TODO: Support None.
# assert 'rrarrrrrrrrra'.find('a', 4, None) == 12
# assert 'rrarrrrrrrrra'.find('a', None, 6) == 2


class Foo(object):

  def __index__(self):
    return 3
assert 'abcd'.find('a', Foo()) == -1

# TODO: This raises IndexError under Grumpy but returns -1 for CPython.
# 'ab'.find('xxx', sys.maxsize + 1, 0)

try:
  "foo".find(123)
  raise AssertionError
except TypeError:
  pass

try:
  'foo'.find()  # pylint: disable=no-value-for-parameter
  raise AssertionError
except TypeError:
  pass

try:
  'foo'.find(42)
  raise AssertionError
except TypeError:
  pass

# TODO: Both of these test cases raise TypeError under CPython but raise
# IndexError under Grumpy.
#try:
#  'foobar'.find("bar", "baz")
#  raise AssertionError
#except TypeError:
#  pass

#try:
#  'foobar'.find("bar", 0, "baz")
#  raise AssertionError
#except TypeError:
#  pass

# Test Mod
assert "%s" % 42 == "42"
assert "%f" % 3.14 == "3.140000"
assert "abc %d" % 123L == "abc 123"
assert "%d" % 3.14 == "3"
assert "%%" % tuple() == "%"
assert "%r" % "abc" == "'abc'"
assert "%x" % 0x1f == "1f"
assert "%X" % 0xffff == "FFFF"

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

class A(object):
  def __int__(self):
    return 3

assert '3'.zfill(A()) == '003'

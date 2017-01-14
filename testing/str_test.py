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

# pylint: disable=redefined-outer-name,W0703

import sys


def checkequal(expected, s, attr, *args):
  assert getattr(s, attr)(*args) == expected


def checkraises(expected, s, attr, *args):
  try:
    getattr(s, attr)(*args)
  except Exception as e:
    assert e.__class__ == expected

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
# TODO: support unicode
# assert "foobar".find(u"bar") == 3
assert "foobar".find("bar", 0, -2) == -1
assert "foobar".find("foo", 0, 3) == 0
assert "foobar".find("bar", 3, 5) == -1
assert "foobar".find("bar", 5, 3) == -1
assert "bar".find("foobar") == -1
try:
  "foo".find(123)
except TypeError as e:
  assert str(e) == "'find' requires a 'str' object but received a 'int'" or str(
      e) == 'expected a string or other character buffer object'

checkequal(0, 'abcdefghiabc', 'find', 'abc')
checkequal(9, 'abcdefghiabc', 'find', 'abc', 1)
checkequal(-1, 'abcdefghiabc', 'find', 'def', 4)

checkequal(0, 'abc', 'find', '', 0)
checkequal(3, 'abc', 'find', '', 3)
checkequal(-1, 'abc', 'find', '', 4)

# to check the ability to pass None as defaults
checkequal(2, 'rrarrrrrrrrra', 'find', 'a')
checkequal(12, 'rrarrrrrrrrra', 'find', 'a', 4)
checkequal(-1, 'rrarrrrrrrrra', 'find', 'a', 4, 6)
# TODO: checkMethodArgs to support MainType + NoneType
# checkequal(12, 'rrarrrrrrrrra', 'find', 'a', 4, None)
# checkequal(2, 'rrarrrrrrrrra', 'find', 'a', None, 6)

checkraises(TypeError, 'hello', 'find')
checkraises(TypeError, 'hello', 'find', 42)

checkequal(0, '', 'find', '')
checkequal(-1, '', 'find', '', 1, 1)
checkequal(-1, '', 'find', '', sys.maxint, 0)

checkequal(-1, '', 'find', 'xx')
checkequal(-1, '', 'find', 'xx', 1, 1)
checkequal(-1, '', 'find', 'xx', sys.maxint, 0)

# issue 7458
checkequal(-1, 'ab', 'find', 'xxx', sys.maxsize + 1, 0)


# Test index
assert "".index("") == 0
try:
  "".index("", 1)
except ValueError as e:
  assert str(e) == "substring not found"
assert "foobar".index("bar") == 3
# TODO: support unicode
# assert "foobar".find(u"bar") == 3
assert "foobar".index("foo", 0, 3) == 0
try:
  "foo".index(123)
except TypeError as e:
  assert str(e) == "'find' requires a 'str' object but received a 'int'" or str(
      e) == 'expected a string or other character buffer object'


# Test zfill
checkequal('123', '123', 'zfill', 2)
checkequal('123', '123', 'zfill', 3)
checkequal('0123', '123', 'zfill', 4)
checkequal('+123', '+123', 'zfill', 3)
checkequal('+123', '+123', 'zfill', 4)
checkequal('+0123', '+123', 'zfill', 5)
checkequal('-123', '-123', 'zfill', 3)
checkequal('-123', '-123', 'zfill', 4)
checkequal('-0123', '-123', 'zfill', 5)
checkequal('000', '', 'zfill', 3)
checkequal('34', '34', 'zfill', 1)
checkequal('0034', '34', 'zfill', 4)

checkraises(TypeError, '123', 'zfill')

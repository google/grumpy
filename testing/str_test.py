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

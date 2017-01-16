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

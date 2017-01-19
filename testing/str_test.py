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
assert "".find("", -1) == 0
assert "".find("", None, -1) == 0
assert "foobar".find("bar") == 3
assert "foobar".find("bar", 0, -2) == -1
assert "foobar".find("foo", 0, 3) == 0
assert "foobar".find("bar", 3, 5) == -1
assert "foobar".find("bar", 5, 3) == -1
assert 'foobar'.find("bar", None) == 3
assert 'foobar'.find("bar", 0, None) == 3
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
assert 'rrarrrrrrrrra'.find('a', 4, None) == 12
assert 'rrarrrrrrrrra'.find('a', None, 6) == 2
assert ''.find('') == 0
assert ''.find('', 1, 1) == -1
assert ''.find('', sys.maxint, 0) == -1
assert ''.find('xx') == -1
assert ''.find('xx', 1, 1) == -1
assert ''.find('xx', sys.maxint, 0) == -1
assert 'ab'.find('xxx', sys.maxsize + 1, 0) == -1
# TODO: Support unicode substring.
# assert "foobar".find(u"bar") == 3

class Foo(object):

  def __index__(self):
    return 3
assert 'abcd'.find('a', Foo()) == -1


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

try:
  'foobar'.find("bar", "baz")
  raise AssertionError
except TypeError:
  pass

try:
  'foobar'.find("bar", 0, "baz")
  raise AssertionError
except TypeError:
  pass

# Test Mod
assert "%s" % 42 == "42"
assert "%f" % 3.14 == "3.140000"
assert "abc %d" % 123L == "abc 123"
assert "%d" % 3.14 == "3"
assert "%%" % tuple() == "%"
assert "%r" % "abc" == "'abc'"
assert "%x" % 0x1f == "1f"
assert "%X" % 0xffff == "FFFF"

# Test replace
assert 'one!two!three!'.replace('!', '@', 1) == 'one@two!three!'
assert 'one!two!three!'.replace('!', '') == 'onetwothree'
assert 'one!two!three!'.replace('!', '@', 2) == 'one@two@three!'
assert 'one!two!three!'.replace('!', '@', 3) == 'one@two@three@'
assert 'one!two!three!'.replace('!', '@', 4) == 'one@two@three@'
assert 'one!two!three!'.replace('!', '@', 0) == 'one!two!three!'
assert 'one!two!three!'.replace('!', '@') == 'one@two@three@'
assert 'one!two!three!'.replace('x', '@') == 'one!two!three!'
assert 'one!two!three!'.replace('x', '@', 2) == 'one!two!three!'
assert 'abc'.replace('', '-') == '-a-b-c-'
assert 'abc'.replace('', '-', 3) == '-a-b-c'
assert 'abc'.replace('', '-', 0) == 'abc'
assert ''.replace('', '') == ''
assert ''.replace('', 'a') == 'a'
assert 'abc'.replace('a', '--', 0) == 'abc'
assert 'abc'.replace('xy', '--') == 'abc'
assert '123'.replace('123', '') == ''
assert '123123'.replace('123', '') == ''
assert '123x123'.replace('123', '') == 'x'
assert '\x10\xc4\x80\xe1\x80\x80\xe3\x80\x80\xe8\x80\x80\xef\xbf\xbf'.replace('\xe3\x80\x80', ' ') == '\x10\xc4\x80\xe1\x80\x80 \xe8\x80\x80\xef\xbf\xbf'
assert '\x10\xc4\x80\xe1\x80\x80\xe3\x80\x80\xe8\x80\x80\xef\xbf\xbf'.replace('\x80', ' ') == '\x10\xc4 \xe1  \xe3  \xe8  \xef\xbf\xbf'

import traceback
try:
  class S(str):
    pass

  s = S('abc')
  assert type(s.replace(s, s)) is str
  assert type(s.replace('x', 'y')) is str
  assert type(s.replace('x', 'y', 0)) is str
  # CPython only, pypy supposed to be same as Go
  assert ''.replace('', 'x') == 'x'
  assert ''.replace('', 'x', -1) == 'x'
  assert ''.replace('', 'x', 0) == ''
  assert ''.replace('', 'x', 1) == ''
  assert ''.replace('', 'x', 1000) == ''
  try:
    ''.replace(None, '')
    raise AssertionError
  except TypeError:
    pass
  try:
    ''.replace('', None)
    raise AssertionError
  except TypeError:
    pass
  try:
    ''.replace('', '', None)
    raise AssertionError
  except TypeError:
    pass

  class A(object):
    def __int__(self):
      return 3
  class AL(object):
    def __int__(self):
      return 3L

  class B(object):
    def __index__(self):
      return 3
  class BL(object):
    def __index__(self):
      return 3L

  assert 'aaaaa'.replace('a', 'b', A()) == 'bbbaa'
  # TODO: Support long maxReplace.
  # assert 'aaaaa'.replace('a', 'b', AL()) == 'bbbaa'
  try:
    'aaaaa'.replace('a', 'b', B())
    raise AssertionError
  except TypeError:
    pass
  try:
    'aaaaa'.replace('a', 'b', BL())
    raise AssertionError
  except TypeError:
    pass
except:
  print traceback.print_exc()

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

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

# pylint: disable=redefined-outer-name,pointless-statement

import sys

# Test Add
assert "foo" + "bar" == "foobar"
assert "foo" + u"bar" == u"foobar"
assert "baz" + "" == "baz"

# Test capitalize
assert "".capitalize() == ""
assert "foo".capitalize() == "Foo"
assert "Foo".capitalize() == "Foo"
assert "FOO".capitalize() == "Foo"

# Test count
assert "".count("a") == 0
assert "abcd".count("e") == 0
assert "abccdef".count("c") == 2
assert "abba".count("bb") == 1
assert "abbba".count("bb") == 1
assert "abbbba".count("bb") == 2
assert "five".count("") == 5
assert ("a" * 20).count("a") == 20

try:
  "".count()
  assert AssertionError
except TypeError:
  pass

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

# Test GetItem
class IntIndexType(object):
  def __index__(self):
    return 2

class LongIndexType(object):
  def __index__(self):
    return 2L

class IntIntType(object):
  def __int__(self):
    return 2

class LongIntType(object):
  def __int__(self):
    return 2L

assert "bar"[1] == "a"
assert "bar"[long(1)] == "a"
assert "baz"[-1] == "z"
assert "baz"[IntIndexType()] == "z"
assert "baz"[LongIndexType()] == "z"
assert "bar"[None:2] == "ba"
assert "bar"[1:3] == "ar"
assert "bar"[1:None] == "ar"
assert "foobarbaz"[1:8:2] == "obra"
assert "abc"[None:None:-1] == "cba"
try:
  "baz"[-4]
  raise AssertionError
except IndexError:
  pass
try:
  ""[0]
  raise AssertionError
except IndexError:
  pass
try:
  "foo"[3]
  raise AssertionError
except IndexError:
  pass
try:
  "foo"[3.14] #pylint: disable=invalid-sequence-index
  raise AssertionError
except TypeError:
  pass
try:
  "bar"[1:2:0]
  raise AssertionError
except ValueError:
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

vals = [
    ['-16', '-16', '       -16', '-16', '-000000016'],
    ['-10', '-10', '       -10', '-10', '-000000010'],
    ['-10', '-10', '       -10', '-10', '-000000010'],
    ['-16', '-16', '       -16', '-16', '       -16'],
    ['-16.000000', '-16.000000', '-16.000000', '-16.000000', '-16.000000'],
    ['-16', '-16', '       -16', '-16', '       -16'],
    ['-20', '-20', '       -20', '-20', '-000000020'],
    ['-10', '-10', '       -10', '-10', '-000000010'],
    ['-a', '-a', '        -a', '-a', '-00000000a'],
    ['-A', '-A', '        -A', '-A', '-00000000A'],
    ['-10', '-10', '       -10', '-10', '       -10'],
    ['-10.000000', '-10.000000', '-10.000000', '-10.000000', '-10.000000'],
    ['-10', '-10', '       -10', '-10', '       -10'],
    ['-12', '-12', '       -12', '-12', '-000000012'],
    ['-1', '-1', '        -1', '-1', '-000000001'],
    ['-1', '-1', '        -1', '-1', '-000000001'],
    ['-1', '-1', '        -1', '-1', '-000000001'],
    ['-1', '-1', '        -1', '-1', '        -1'],
    ['-1.000000', '-1.000000', ' -1.000000', '-1.000000', '-01.000000'],
    ['-1', '-1', '        -1', '-1', '        -1'],
    ['-1', '-1', '        -1', '-1', '-000000001'],
    ['0', ' 0', '         0', '00', '0000000000'],
    ['0', ' 0', '         0', '00', '0000000000'],
    ['0', ' 0', '         0', '00', '0000000000'],
    ['0', ' 0', '         0', ' 0', '         0'],
    ['0.000000', '0.000000', '  0.000000', '0.000000', '000.000000'],
    ['0', ' 0', '         0', ' 0', '         0'],
    ['0', ' 0', '         0', '00', '0000000000'],
    ['1', ' 1', '         1', '01', '0000000001'],
    ['1', ' 1', '         1', '01', '0000000001'],
    ['1', ' 1', '         1', '01', '0000000001'],
    ['1', ' 1', '         1', ' 1', '         1'],
    ['1.000000', '1.000000', '  1.000000', '1.000000', '001.000000'],
    ['1', ' 1', '         1', ' 1', '         1'],
    ['1', ' 1', '         1', '01', '0000000001'],
    ['3', ' 3', '         3', '03', '0000000003'],
    ['3', ' 3', '         3', '03', '0000000003'],
    ['3', ' 3', '         3', '03', '0000000003'],
    ['3.14', '3.14', '      3.14', '3.14', '      3.14'],
    ['3.140000', '3.140000', '  3.140000', '3.140000', '003.140000'],
    ['3.14', '3.14', '      3.14', '3.14', '      3.14'],
    ['3', ' 3', '         3', '03', '0000000003'],
    ['10', '10', '        10', '10', '0000000010'],
    ['a', ' a', '         a', '0a', '000000000a'],
    ['A', ' A', '         A', '0A', '000000000A'],
    ['10', '10', '        10', '10', '        10'],
    ['10.000000', '10.000000', ' 10.000000', '10.000000', '010.000000'],
    ['10', '10', '        10', '10', '        10'],
    ['12', '12', '        12', '12', '0000000012'],
    ['16', '16', '        16', '16', '0000000016'],
    ['10', '10', '        10', '10', '0000000010'],
    ['10', '10', '        10', '10', '0000000010'],
    ['16', '16', '        16', '16', '        16'],
    ['16.000000', '16.000000', ' 16.000000', '16.000000', '016.000000'],
    ['16', '16', '        16', '16', '        16'],
    ['20', '20', '        20', '20', '0000000020'],
]

i = 0
for a in [-16, -10, -1, 0, 1, 3.14, 10, 16]:
  for b in "dxXrfso":
    assert [("%" + b) % (a, ), ("%2" + b) % (a, ), ("%10" + b) % (a, ),
            ("%02" + b) % (a, ), ("%010" + b) % (a, )] == vals[i]
    i += 1

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
assert "\xd0\xb2\xd0\xbe\xd0\xbb".replace('', '\0') == "\x00\xd0\x00\xb2\x00\xd0\x00\xbe\x00\xd0\x00\xbb\x00"
assert "\xd0\xb2\xd0\xbe\xd0\xbb".replace('', '\1\2') == '\x01\x02\xd0\x01\x02\xb2\x01\x02\xd0\x01\x02\xbe\x01\x02\xd0\x01\x02\xbb\x01\x02'

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
assert 'aaaaa'.replace('a', 'b', AL()) == 'bbbaa'
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

# Test split
assert "".split() == []
assert " ".split() == []
assert "".split('x') == ['']
assert "a".split() == ['a']
assert " ".split(" ", 1) == ['', '']
assert "aa".split("a", 2) == ['', '', '']
assert " a ".split() == ['a']
assert 'a b c d'.split(None, 1) == ['a', 'b c d']
assert 'a b c d '.split() == ['a', 'b', 'c', 'd']
assert ' a b c d '.split(None, 1) == ['a', 'b c d ']
assert '   a b c d'.split(None, 0) == ['a b c d']

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
assert '3'.zfill(IntIntType()) == '03'
assert '3'.zfill(LongIntType()) == '03'

assert '%o' % 8 == '10'
assert '%o' % -8 == '-10'
assert '%o %o' % (8, -8) == '10 -10'

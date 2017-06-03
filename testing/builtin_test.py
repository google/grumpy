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

# pylint: disable=g-equals-none

# abs(x)

assert abs(1) == 1
assert abs(-1) == 1
assert isinstance(abs(-1), int)

assert abs(long(2)) == 2
assert abs(long(-2)) == 2
assert isinstance(abs(long(-2)), long)

assert abs(3.4) == 3.4
assert abs(-3.4) == 3.4
assert isinstance(abs(-3.4), float)

assert abs(complex(0, 0)) == 0.0
assert abs(complex(3, 4)) == 5.0
assert abs(-complex(3, 4)) == 5.0
assert abs(complex(0.123456e-3, 0)) == 0.000123456
assert abs(complex(0.123456e-3, 3.14151692e+7)) == 31415169.2
assert isinstance(abs(complex(3, 4)), float)
assert repr(abs(complex(-float('inf'), 1.2))) == 'inf'
assert repr(abs(complex(float('nan'), float('inf')))) == 'inf'
assert repr(abs(complex(3.14, float('nan')))) == 'nan'

try:
  abs('a')
except TypeError as e:
  assert str(e) == "bad operand type for abs(): 'str'"
else:
  raise AssertionError('this was supposed to raise an exception')


# all(iterable)

assert all([1, 2, 3])
assert all([])
assert not all([1, 1, 1, 0, 1])

assert all([True, True])
assert not all([False, True, True])

assert all('')
assert all('abc')

try:
  all(13)
except TypeError as e:
  assert str(e) == "'int' object is not iterable"
else:
  raise AssertionError('this was supposed to raise an exception')


# any(iterable)

assert any([1, 2, 3])
assert not any([])
assert any([1, 1, 1, 0, 1])
assert not any([0, 0, 0])

assert any([True, True])
assert any([False, True, True])
assert not any([False, False, False])

assert not any('')
assert any('abc')

try:
  any(13)
except TypeError as e:
  assert str(e) == "'int' object is not iterable"
else:
  raise AssertionError('this was supposed to raise an exception')


# callable(x)

assert not callable(1)
assert not callable(0.1)

assert not callable([1, 2, 3])
assert not callable((1, 2, 3))
assert not callable({'foo': 1, 'bar': 2})

assert callable(lambda x: x + 1)


def foo(x):
  pass

assert callable(foo)


class bar(object):

  def __call__(self, *args, **kwargs):
    pass

assert callable(bar)
assert callable(bar())

# cmp(x)

# Test simple cases.
assert cmp(1, 2) == -1
assert cmp(3, 3) == 0
assert cmp(5, 4) == 1


class Lt(object):

  def __init__(self, x):
    self.lt_called = False
    self.x = x

  def __lt__(self, other):
    self.lt_called = True
    return self.x < other.x


class Eq(object):

  def __init__(self, x):
    self.eq_called = False
    self.x = x

  def __eq__(self, other):
    self.eq_called = True
    return self.x == other.x


class Gt(object):

  def __init__(self, x):
    self.gt_called = False
    self.x = x

  def __gt__(self, other):
    self.gt_called = True
    return self.x > other.x


class RichCmp(Lt, Eq, Gt):

  def __init__(self, x):
    self.x = x


class Cmp(object):

  def __init__(self, x):
    self.cmp_called = False
    self.x = x

  def __cmp__(self, other):
    self.cmp_called = True
    if self.x < other.x:
      return -1
    elif self.x > other.x:
      return 1
    else:
      return 0


class NoCmp(object):

  def __init__(self, x):
    self.x = x

# Test 3-way compare in terms of rich compare.

a, b = RichCmp(1), RichCmp(2)

assert cmp(a, b) == -1
assert a.lt_called

a, b = RichCmp(3), RichCmp(3)

assert cmp(a, b) == 0
assert a.eq_called

a, b = RichCmp(5), RichCmp(4)

assert cmp(a, b) == 1
assert a.gt_called

# Test pure 3-way compare.

a, b = Cmp(1), Cmp(2)

assert cmp(a, b) == -1
assert a.cmp_called

a, b = Cmp(3), Cmp(3)

assert cmp(a, b) == 0
assert a.cmp_called

a, b = Cmp(5), Cmp(4)

assert cmp(a, b) == 1

# Test mixed 3-way and rich compare.

a, b = RichCmp(1), Cmp(2)
assert cmp(a, b) == -1
assert a.lt_called
assert not b.cmp_called

a, b = Cmp(1), RichCmp(2)
assert cmp(a, b) == -1
assert not a.cmp_called
assert b.gt_called

a, b = RichCmp(3), Cmp(3)
assert cmp(a, b) == 0
assert a.eq_called
assert not b.cmp_called

a, b = Cmp(3), RichCmp(3)
assert cmp(a, b) == 0
assert not a.cmp_called
assert b.eq_called

a, b = RichCmp(5), Cmp(4)
assert cmp(a, b) == 1
assert a.gt_called
assert not b.cmp_called

a, b = Cmp(5), RichCmp(4)
assert cmp(a, b) == 1
assert not a.cmp_called
assert b.gt_called

# Test compare on only one object.

a, b = Cmp(1), NoCmp(2)
assert cmp(a, b) == -1
assert a.cmp_called

a, b = NoCmp(1), Cmp(2)
assert cmp(a, b) == -1
assert b.cmp_called

# Test delattr

class Foo(object):
  pass

setattr(Foo, "a", 1)
assert Foo.a == 1  # pylint: disable=no-member

delattr(Foo, "a")
assert getattr(Foo, "a", None) is None

try:
  delattr(Foo, 1, "a")
  assert AssertionError
except TypeError:
  pass

try:
  delattr(Foo)
  assert AssertionError
except TypeError:
  pass

try:
  delattr(Foo, "a", 1)
  assert AssertionError
except TypeError:
  pass

# Test setattr

setattr(Foo, "a", 1)
assert Foo.a == 1  # pylint: disable=no-member

try:
  setattr(Foo, 1, "a")
  assert AssertionError
except TypeError:
  pass

try:
  setattr(Foo)
  assert AssertionError
except TypeError:
  pass

# Test sorted

assert sorted([3, 2, 4, 1]) == [1, 2, 3, 4]
assert sorted([]) == []
assert sorted(["a", "e", "c", "b"]) == ["a", "b", "c", "e"]
assert sorted((3, 1, 5, 2, 4)) == [1, 2, 3, 4, 5]
assert sorted({"foo": 1, "bar": 2}) == ["bar", "foo"]

# Test zip

assert zip('abc', (0, 1, 2)) == [('a', 0), ('b', 1), ('c', 2)]
assert list(zip('abc', range(6))) == zip('abc', range(6))
assert list(zip('abcdef', range(3))) == zip('abcdef', range(3))
assert list(zip('abcdef')) == zip('abcdef')
assert list(zip()) == zip()
assert [tuple(list(pair)) for pair in zip('abc', 'def')] == zip('abc', 'def')
assert [pair for pair in zip('abc', 'def')] == zip('abc', 'def')
assert zip({'b': 1, 'a': 2}) == [('a',), ('b',)]
assert zip(range(5)) == [(0,), (1,), (2,), (3,), (4,)]
assert zip(xrange(5)) == [(0,), (1,), (2,), (3,), (4,)]
assert zip([1, 2, 3], [1], [4, 5, 6]) == [(1, 1, 4)]
assert zip([1], [1, 2, 3], [4, 5, 6]) == [(1, 1, 4)]
assert zip([4, 5, 6], [1], [1, 2, 3]) == [(4, 1, 1)]
assert zip([1], [1, 2, 3], [4]) == [(1, 1, 4)]
assert zip([1, 2], [1, 2, 3], [4]) == [(1, 1, 4)]
assert zip([1, 2, 3, 4], [1, 2, 3], [4]) == [(1, 1, 4)]
assert zip([1], [1, 2], [4, 2, 4]) == [(1, 1, 4)]
assert zip([1, 2, 3], [1, 2], [4]) == [(1, 1, 4)]
assert zip([1, 2, 3], [1, 2], [4], []) == []
assert zip([], [1], [1, 2], [1, 2, 3]) == []
try:
  zip([1, 2, 3], [1, 2], [4], None)
  raise AssertionError
except TypeError:
  pass

# Test map

assert map(str, []) == []
assert map(str, [1, 2, 3]) == ["1", "2", "3"]
assert map(str, (1, 2, 3)) == ["1", "2", "3"]
# assert map(str, (1.0, 2.0, 3.0)) == ["1", "2", "3"]
assert map(str, range(3)) == ["0", "1", "2"]
assert map(str, xrange(3)) == ["0", "1", "2"]
assert map(int, ["1", "2", "3"]) == [1, 2, 3]
assert map(int, "123") == [1, 2, 3]
assert map(int, {"1": "a", "2": "b"}) == [1, 2]
assert map(int, {1: "a", 2: "b"}) == [1, 2]
assert map(lambda a, b: (str(a), float(b or 0) + 0.1),
           [1, 2, 3], [1, 2]) == [('1', 1.1), ('2', 2.1), ('3', 0.1)]
assert map(None, [1, 2, 3]) == [1, 2, 3]
a = [1, 2, 3]
assert map(None, a) == a
assert map(None, a) is not a
assert map(None, (1, 2, 3)) == [1, 2, 3]

# divmod(v, w)

import sys

assert divmod(12, 7) == (1, 5)
assert divmod(-12, 7) == (-2, 2)
assert divmod(12, -7) == (-2, -2)
assert divmod(-12, -7) == (1, -5)
assert divmod(-sys.maxsize - 1, -1) == (sys.maxsize + 1, 0)
assert isinstance(divmod(12, 7), tuple)
assert isinstance(divmod(12, 7)[0], int)
assert isinstance(divmod(12, 7)[1], int)

assert divmod(long(7), long(3)) == (2L, 1L)
assert divmod(long(3), long(-7)) == (-1L, -4L)
assert divmod(long(sys.maxsize), long(-sys.maxsize)) == (-1L, 0L)
assert divmod(long(-sys.maxsize), long(1)) == (-sys.maxsize, 0L)
assert divmod(long(-sys.maxsize), long(-1)) == (sys.maxsize, 0L)
assert isinstance(divmod(long(7), long(3)), tuple)
assert isinstance(divmod(long(7), long(3))[0], long)
assert isinstance(divmod(long(7), long(3))[1], long)

assert divmod(3.25, 1.0) == (3.0, 0.25)
assert divmod(-3.25, 1.0) == (-4.0, 0.75)
assert divmod(3.25, -1.0) == (-4.0, -0.75)
assert divmod(-3.25, -1.0) == (3.0, -0.25)
assert isinstance(divmod(3.25, 1.0), tuple)
assert isinstance(divmod(3.25, 1.0)[0], float)
assert isinstance(divmod(3.25, 1.0)[1], float)

try:
  divmod('a', 'b')
except TypeError as e:
  assert str(e) == "unsupported operand type(s) for divmod(): 'str' and 'str'"
else:
  assert AssertionError

# Check for a bug where zip() and map() were not properly cleaning their
# internal exception state. See:
# https://github.com/google/grumpy/issues/305
sys.exc_clear()
zip((1, 3), (2, 4))
assert not any(sys.exc_info())
map(int, (1, 2, 3))
assert not any(sys.exc_info())

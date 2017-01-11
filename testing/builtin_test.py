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

assert callable(lambda x: x+1)

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

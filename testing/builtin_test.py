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


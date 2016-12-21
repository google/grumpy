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

import types


def gen1():
  yield 1
  yield 2
  yield 3
g = gen1()
assert isinstance(g, types.GeneratorType)
assert list(g) == [1, 2, 3]
assert list(g) == []  # pylint: disable=g-explicit-bool-comparison


def gen2():
  for c in 'foobar':
    yield c
  yield '!'
g = gen2()
assert list(g) == ['f', 'o', 'o', 'b', 'a', 'r', '!']
assert list(g) == []  # pylint: disable=g-explicit-bool-comparison


def gen3():
  raise RuntimeError
  yield 1  # pylint: disable=unreachable
g = gen3()
try:
  g.next()
except RuntimeError:
  pass
assert list(g) == []  # pylint: disable=g-explicit-bool-comparison


def gen4():
  yield g.next()
g = gen4()
try:
  g.next()
except ValueError as e:
  assert 'generator already executing' in str(e), str(e)
else:
  raise AssertionError


def gen5():
  yield
g = gen5()
try:
  g.send('foo')
except TypeError as e:
  assert "can't send non-None value to a just-started generator" in str(e)
else:
  raise AssertionError


def gen6():
  yield 1
  return
  yield 2
g = gen6()
assert list(g) == [1]
assert list(g) == []

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

# pylint: disable=no-value-for-parameter,function-redefined


def foo(a):
  return {'a': a}


assert foo(123) == {'a': 123}
assert foo(a='apple') == {'a': 'apple'}
assert foo(*('bar',)) == {'a': 'bar'}
assert foo(**{'a': 42}) == {'a': 42}
try:
  foo(b='bear')  # pylint: disable=unexpected-keyword-arg
  raise AssertionError
except TypeError as e:
  assert str(e) == "foo() got an unexpected keyword argument 'b'"
try:
  foo()
  raise AssertionError
except TypeError:
  pass
try:
  foo(1, 2, 3)  # pylint: disable=too-many-function-args
  raise AssertionError
except TypeError:
  pass


def foo(a, b):
  return {'a': a, 'b': b}


assert foo(1, 2) == {'a': 1, 'b': 2}
assert foo(1, b='bear') == {'a': 1, 'b': 'bear'}
assert foo(b='bear', a='apple') == {'a': 'apple', 'b': 'bear'}
try:
  foo(1, a='alpha')  # pylint: disable=redundant-keyword-arg
  raise AssertionError
except TypeError as e:
  assert str(e) == "foo() got multiple values for keyword argument 'a'"
try:
  foo(**{123: 'bar'})
  pass
except TypeError:
  pass


def foo(a, b=None):
  return {'a': a, 'b': b}


assert foo(123) == {'a': 123, 'b': None}
assert foo(123, 'bar') == {'a': 123, 'b': 'bar'}
assert foo(a=123, b='bar') == {'a': 123, 'b': 'bar'}
assert foo(*('apple',), **{'b': 'bear'}) == {'a': 'apple', 'b': 'bear'}


def foo(a, *args):
  return {'a': a, 'args': args}


assert foo(1) == {'a': 1, 'args': ()}
assert foo(1, 2, 3) == {'a': 1, 'args': (2, 3)}


def foo(a, **kwargs):
  return {'a': a, 'kwargs': kwargs}


assert foo('bar') == {'a': 'bar', 'kwargs': {}}
assert (foo(**{'a': 'apple', 'b': 'bear'}) ==
        {'a': 'apple', 'kwargs': {'b': 'bear'}})
assert (foo('bar', b='baz', c='qux') ==
        {'a': 'bar', 'kwargs': {'b': 'baz', 'c': 'qux'}})

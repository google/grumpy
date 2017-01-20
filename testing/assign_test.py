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

# pylint: disable=unbalanced-tuple-unpacking


class Foo(object):
  pass


foo = 1
assert foo == 1

foo, bar = 2, 3
assert foo == 2
assert bar == 3

(foo, bar), baz = (4, 5), 6
assert foo == 4
assert bar == 5
assert baz == 6

foo = [7, 8, 9]
bar = foo
assert bar == [7, 8, 9]

try:
  bar, baz = foo
except ValueError as e:
  assert str(e) == 'too many values to unpack'
else:
  raise AssertionError('this was supposed to raise an exception')

try:
  bar, baz, qux, quux = foo
except ValueError as e:
  assert str(e) == 'need more than 3 values to unpack'
else:
  raise AssertionError('this was supposed to raise an exception')

foo = Foo()

foo.bar = 1
assert foo.bar == 1

foo.bar, baz = 2, 3
assert foo.bar == 2
assert baz == 3

foo.bar, (foo.baz, qux) = 4, (5, 6)
assert foo.bar == 4
assert foo.baz == 5
assert qux == 6

foo = bar = baz = 7
assert foo == 7
assert bar == 7
assert baz == 7

foo, bar = baz = 8, 9
assert foo == 8
assert bar == 9
assert baz == (8, 9)

foo = 1
foo += 3
assert foo == 4
foo /= 2
assert foo == 2
foo *= 6
assert foo == 12
foo %= 5
assert foo == 2
foo -= 3
assert foo == -1

foo = []
bar = foo
foo += ["bar", "baz"]
assert foo == ["bar", "baz"]
foo *= 2
assert foo == ["bar", "baz", "bar", "baz"]
assert bar is foo


# Multiple target assignment should only evaluate rhs once.
def foo():  # pylint: disable=function-redefined
  foo_ran[0] += 1
  return 'bar'


foo_ran = [0]
baz = qux = foo()
assert baz == 'bar'
assert qux == 'bar'
assert foo_ran == [1]

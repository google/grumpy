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

x = 'foo'
y = 'wut'
assert x == 'foo'
assert y == 'wut'


def f():
  x = 'bar'
  z = 'baz'
  assert x == 'bar'
  assert y == 'wut'
  assert z == 'baz'
  def g(arg):
    x = 'qux'
    assert x == 'qux'
    assert y == 'wut'
    assert z == 'baz'
    assert arg == 'quux'
    arg = None
  g('quux')


f()


# Delete a local var.
def g():
  foo = 'bar'
  del foo
  try:
    foo
  except UnboundLocalError:
    pass
  else:
    raise AssertionError


g()


# Delete a global.
foo = 'bar'
del foo
try:
  foo
except NameError:
  pass
else:
  raise AssertionError


# Delete a class var.
class Foo(object):
  foo = 'bar'
  del foo
  try:
    foo
  except NameError:
    pass
  else:
    raise AssertionError

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


class Foo(object):

  a = 3
  assert a == 3

  def bar(self):
    assert isinstance(self, Foo)
    return 'bar'

  baz = bar


assert Foo.a == 3

Foo.a = 4
assert Foo.a == 4

foo = Foo()
assert isinstance(foo, Foo)
assert foo.a == 4
foo.a = 5
assert foo.a == 5
assert Foo.a == 4
assert foo.bar() == 'bar'
assert foo.baz() == 'bar'

foo.b = 10
del foo.b
assert not hasattr(foo, 'b')
try:
  del foo.b
except AttributeError:
  pass
else:
  raise AssertionError

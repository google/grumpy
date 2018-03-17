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

# pylint: disable=g-wrong-blank-lines,global-variable-not-assigned,invalid-name,redefined-outer-name,unused-variable

x = 123
def f1():
  global x
  x = 'abc'
f1()
assert x == 'abc'


x = 'foo'
def f2():
  global x
  class x(object):
    pass
f2()
assert isinstance(x, type)
assert x.__name__ == 'x'


x = 3.14
class C1(object):
  global x
  x = 'foo'
assert x == 'foo'


x = 42
def f3():
  global x
  del x
f3()
try:
  print x
  raise AssertionError
except NameError:
  pass


x = 'foo'
def f4():
  x = 'bar'
  def g():
    global x
    def h():
      return x
    return h()
  return g()
assert f4() == 'foo'


x = 3.14
def f5():
  x = 'foo'
  class C(object):
    global x
    y = x
  return C.y
assert f5() == 3.14

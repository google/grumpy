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


class ContextManager(object):

  def __init__(self):
    self.active = False

  def __enter__(self):
    self.active = True

  def __exit__(self, exc_type, value, traceback):
    self.exc_type = exc_type
    self.value = value
    self.traceback = traceback
    self.active = False


a = ContextManager()

# Basic case
assert not a.active
with a:
  assert a.active

assert not a.active

# Exception raised during with clause
try:
  with a:
    assert a.active
    raise RuntimeError('foo')
    assert False  # pylint: disable=unreachable
except RuntimeError:
  assert not a.active

assert not a.active
# Check that all three arguments to __exit__ were passed correctly.
assert a.exc_type == RuntimeError
assert str(a.value) == 'foo'
assert isinstance(a.traceback, types.TracebackType)

# Plays well with finally block.
finally_visited = False
try:
  with a:
    assert a.active
finally:
  finally_visited = True
assert finally_visited

finally_visited = False
except_visited = False
try:
  with a:
    assert a.active
    raise RuntimeError
    assert False  # pylint: disable=unreachable
except RuntimeError:
  except_visited = True
finally:
  finally_visited = True
assert finally_visited
assert except_visited


# 'a' modified during with clause
a_backup = a
with a:
  a = None
  assert a_backup.active

assert not a_backup.active
a = a_backup


# Test with a context manager that returns true (swallowing the exception).
class ExceptionSwallower(object):

  def __init__(self):
    self.active = False

  def __enter__(self):
    self.active = True

  def __exit__(self, exc_type, value, traceback):
    self.exc_type = exc_type
    self.value = value
    self.traceback = traceback
    self.active = False
    return True

b = ExceptionSwallower()
try:
  with b:
    assert b.active
    raise RuntimeError()
    assert False  # pylint: disable=unreachable
except RuntimeError:
  assert False  # Exception should be swallowed by with clause.

assert not b.active
assert b.exc_type is RuntimeError  # Make sure __exit__ got the exception.


# Test missing and broken context managers
class NoExit(object):

  def __enter__(self):
    pass

c = NoExit()

try:
  with c:  # pylint: disable=not-context-manager
    assert False  # Shouldn't get here
except Exception as e:  # pylint: disable=broad-except
  # TODO: Once str.find() is implemented, verify that the attribute error was
  # raised for the correct method (__exit__, not __enter__).
  assert isinstance(e, AttributeError)


class NoEnter(object):

  def __exit__(self, exc_type, value, traceback):
    pass


d = NoEnter()

try:
  with d:  # pylint: disable=not-context-manager
    assert False  # Shouldn't get here
except Exception as e:  # pylint: disable=broad-except
  assert isinstance(e, AttributeError)


f = 'not a context manager'

try:
  with f:  # pylint: disable=not-context-manager
    assert False  # Shouldn't get here
except Exception as e:  # pylint: disable=broad-except
  assert isinstance(e, AttributeError)


class EnterResult(object):

  def __init__(self, value):
    self.value = value

  def __enter__(self):
    return self.value

  def __exit__(self, *args):
    pass


with EnterResult('123') as g:
  pass
assert g == '123'


with EnterResult([1, (2, 3)]) as (h, (i, j)):
  pass
assert h == 1
assert i == 2
assert j == 3


class Foo(object):
  exited = False
  def __enter__(self):
    pass
  def __exit__(self, *args):
    self.exited = True


# This checks for a bug where a with clause inside an except body raises an
# exception because it was checking ExcInfo() to determine whether an exception
# occurred.
try:
  raise AssertionError
except:
  foo = Foo()
  with foo:
    pass
  assert foo.exited


# Return statement should not bypass the with exit handler.
foo = Foo()
def bar():
  with foo:
    return
bar()
assert foo.exited

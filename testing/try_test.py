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

# pylint: disable=bare-except,broad-except,unreachable,redefined-outer-name

# Else should run when no exception raised.
x = 0
try:
  x = 1
except:
  x = 2
else:
  x = 3
assert x == 3

# Bare except handles all.
x = 0
try:
  x = 1
  raise Exception
  x = 2
except:
  x = 3
assert x == 3

# Correct handler triggered.
x = 0
try:
  x = 1
  raise Exception
  x = 2
except TypeError:
  x = 4
except Exception:
  x = 3
assert x == 3

# Else should not run when exception raised.
x = 0
try:
  x = 1
  raise Exception
  x = 2
except Exception:
  x = 3
else:
  x = 4
assert x == 3

# Finally should execute last.
x = 0
try:
  x = 1
finally:
  x = 2
assert x == 2

# Finally should execute when exception raised.
x = 0
try:
  x = 1
  raise Exception
  x = 2
except:
  x = 3
finally:
  x = 4
assert x == 4

# Uncaught exception should propagate to the next handler.
x = 0
try:
  try:
    raise Exception
    x = 1
  except TypeError:
    x = 2
except Exception:
  x = 3
assert x == 3

# Exceptions that pass through a finally, should propagate.
x = 0
try:
  try:
    x = 1
    raise Exception
    x = 2
  finally:
    x = 3
except Exception:
  pass
assert x == 3

# If a function does not handle an exception it should propagate.
x = 0
def f():
  x = 1
  raise Exception
try:
  f()
  x = 2
except Exception:
  x = 3
assert x == 3


def foo():
  # Else should run when no exception raised.
  x = 0
  try:
    x = 1
  except:
    x = 2
  else:
    x = 3
  assert x == 3

  # Bare except handles all.
  x = 0
  try:
    x = 1
    raise Exception
    x = 2
  except:
    x = 3
  assert x == 3

  # Correct handler triggered.
  x = 0
  try:
    x = 1
    raise Exception
    x = 2
  except TypeError:
    x = 4
  except Exception:
    x = 3
  assert x == 3

  # Else should not run when exception raised.
  x = 0
  try:
    x = 1
    raise Exception
    x = 2
  except Exception:
    x = 3
  else:
    x = 4
  assert x == 3

  # Finally should execute last.
  x = 0
  try:
    x = 1
  finally:
    x = 2
  assert x == 2

  # Finally should execute when exception raised.
  x = 0
  try:
    x = 1
    raise Exception
    x = 2
  except:
    x = 3
  finally:
    x = 4
  assert x == 4

  # Uncaught exception should propagate to the next handler.
  x = 0
  try:
    try:
      raise Exception
      x = 1
    except TypeError:
      x = 2
  except Exception:
    x = 3
  assert x == 3

  # Exceptions that pass through a finally, should propagate.
  x = 0
  try:
    try:
      x = 1
      raise Exception
      x = 2
    finally:
      x = 3
  except Exception:
    pass
  assert x == 3

  # If a function does not handle an exception it should propagate.
  x = 0
  def f():
    x = 1
    raise Exception
  try:
    f()
    x = 2
  except Exception:
    x = 3
  assert x == 3


foo()


# Return statement should not bypass the finally.
def foo():
  try:
    return 1
  finally:
    return 2
  return 3


assert foo() == 2


# Break statement should not bypass finally.
x = []
def foo():
  while True:
    try:
      x.append(1)
      break
    finally:
      x.append(2)
  x.append(3)


foo()
assert x == [1, 2, 3]

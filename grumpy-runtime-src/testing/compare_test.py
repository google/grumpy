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

assert 1 < 100
assert -10 <= "foo"
assert "bar" <= "bar"
assert (1, "a", 3) == (1, "a", 3)
assert 15 != 16
assert [] != None  # pylint: disable=g-equals-none,g-explicit-bool-comparison
assert int >= "az"
assert "foo" >= "foo"
assert True > False

# Test rich comparisons.

class RichCmp(object):
  def __init__(self, x):
    self.x = x
    self.lt_called = False
    self.le_called = False
    self.eq_called = False
    self.ge_called = False
    self.gt_called = False

  def __lt__(self, other):
    self.lt_called = True
    return self.x < other.x

  def __le__(self, other):
    self.le_called = True
    return self.x <= other.x

  def __eq__(self, other):
    self.eq_called = True
    return self.x == other.x

  def __ge__(self, other):
    self.ge_called = True
    return self.x >= other.x

  def __gt__(self, other):
    self.gt_called = True
    return self.x > other.x

class Cmp(object):
  def __init__(self, x):
    self.cmp_called = False
    self.x = x

  def __cmp__(self, other):
    self.cmp_called = True
    return cmp(self.x, other.x)

# Test that rich comparison methods are called.

a, b = RichCmp(1), RichCmp(2)
assert a < b
assert a.lt_called

a, b = RichCmp(1), RichCmp(2)
assert a <= b
assert a.le_called

a, b = RichCmp(3), RichCmp(3)
assert a == b
assert a.eq_called

a, b = RichCmp(5), RichCmp(4)
assert a >= b
assert a.ge_called

a, b = RichCmp(5), RichCmp(4)
assert a > b
assert a.gt_called

# Test rich comparison falling back to a 3-way comparison

a, b = Cmp(1), Cmp(2)
assert a < b
assert a.cmp_called

a, b = Cmp(1), Cmp(2)
assert a <= b
assert a.cmp_called

a, b = Cmp(3), Cmp(3)
assert a == b
assert a.cmp_called

a, b = Cmp(5), Cmp(4)
assert a > b
assert a.cmp_called

a, b = Cmp(5), Cmp(4)
assert a >= b
assert a.cmp_called

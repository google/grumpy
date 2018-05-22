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

"""Arithmetic and boolean operator tests."""

import math

import weetest


def TestBoolOps():
  assert ('foo' or 'bar') == 'foo'
  assert ('' or 123) == 123
  assert (0 and 3.14) == 0
  assert (True and False) is False
  assert (0 or 'a' and 'b') == 'b'
  assert (1 and 'a' or 'b') == 'a'


def TestBoolOpsLazyEval():
  def Yes():
    ran.append('Yes')
    return True

  def No():
    ran.append('No')
    return False

  ran = []
  assert Yes() or No()
  assert ran == ['Yes']

  ran = []
  assert not (Yes() and Yes() and No())
  assert ran == ['Yes', 'Yes', 'No']

  ran = []
  assert not (Yes() and No() and Yes())
  assert ran == ['Yes', 'No']

  ran = []
  assert No() or No() or Yes()
  assert ran == ['No', 'No', 'Yes']

  ran = []
  assert Yes() or Yes() or Yes()
  assert ran == ['Yes']


def TestNeg():
  x = 12
  assert -x == -12

  x = 1.1
  assert -x == -1.1

  x = 0.0
  assert -x == -0.0

  x = float('inf')
  assert math.isinf(-x)

  x = -float('inf')
  assert math.isinf(-x)

  x = float('nan')
  assert math.isnan(-x)

  x = long(100)
  assert -x == -100


def TestPos():
  x = 12
  assert +x == 12

  x = 1.1
  assert +x == 1.1

  x = 0.0
  assert +x == 0.0

  x = float('inf')
  assert math.isinf(+x)

  x = +float('inf')
  assert math.isinf(+x)

  x = float('nan')
  assert math.isnan(+x)

  x = long(100)
  assert +x == 100


if __name__ == '__main__':
  weetest.RunTests()

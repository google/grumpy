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

import math

import weetest

# Tests exist for all functions which have logic in the math.py module, instead
# of simply calling the go equivalent.


def TestFactorial():
  assert math.factorial(0) == 1
  assert math.factorial(1) == 1
  assert math.factorial(2) == 2
  assert math.factorial(3) == 6
  assert math.factorial(4) == 24
  assert math.factorial(5) == 120


def TestFactorialError():
  try:
    math.factorial(-1)
  except ValueError:
    pass
  else:
    raise AssertionError
  
  try:
    math.factorial(0.5)
  except ValueError:
    pass
  else:
    raise AssertionError


def TestLdexp():
  assert math.ldexp(1,1) == 2
  assert math.ldexp(1,2) == 4
  assert math.ldexp(1.5,1) == 3
  assert math.ldexp(1.5,2) == 6


def TestLog():
  assert math.log(math.e) == 1
  assert math.log(2,2) == 1
  assert math.log(10,10) == 1
  assert math.log(100,10) == 2


def TestRadians():
  assert math.radians(180) == math.pi
  assert math.radians(360) == 2 * math.pi


def TestDegrees():
  assert math.degrees(math.pi) == 180
  assert math.degrees(2 * math.pi) == 360


if __name__ == '__main__':
  weetest.RunTests()

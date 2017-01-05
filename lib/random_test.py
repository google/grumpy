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

import random

import weetest


def TestSeed():
  random.seed()
  try:
    random.seed("wrongtype")
  except TypeError:
    pass
  else:
    raise AssertionError("TypeError not raised")


def TestRandom():
  a = random.random()
  b = random.random()
  c = random.random()
  assert isinstance(a, float)
  assert 0.0 <= a < 1.0
  assert not a == b == c


def TestRandomInt():
  for _ in range(10):
    a = random.randint(0, 5)
    assert isinstance(a, int)
    assert 0 <= a <= 5

  b = random.randint(1, 1)
  assert b == 1

  try:
    c = random.randint(0.1, 3)
  except ValueError:
    pass
  else:
    raise AssertionError("ValueError not raised")

  try:
    d = random.randint(4, 3)
  except ValueError:
    pass
  else:
    raise AssertionError("ValueError not raised")


def TestRandomChoice():
  seq = [i*2 for i in range(5)]
  for i in range(10):
    item = random.choice(seq)
    item_idx = item/2
    assert seq[item_idx] == item

  try:
    random.choice([])
  except IndexError:
    pass
  else:
    raise AssertionError("IndexError not raised")


if __name__ == '__main__':
  weetest.RunTests()
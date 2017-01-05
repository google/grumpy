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

"""Generate pseudo random numbers. Should not be used for security purposes."""

from __go__.math.rand import Float64, Seed
from __go__.time import Now


def seed(a=None):
  if a is None:
    a = Now().UnixNano()
  Seed(a)


def random():
  return Float64()


def randint(a, b):
  if int(a) != a or int(b) != b:
    raise ValueError("non-integer for randrange()")
  r = (b - a) + 1
  if r < 1:
    raise ValueError("empty range for randrange()")
  return int(a) + int(random() * r)


def choice(seq):
  return seq[int(random() * len(seq))]

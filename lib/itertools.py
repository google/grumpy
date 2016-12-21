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

"""Utilities for iterating over containers."""

import sys


class islice(object):  # pylint: disable=invalid-name
  """Iterator that returns sliced elements from the iterable."""

  def __init__(self, iterable, *args):
    s = slice(*args)
    self._range = iter(xrange(s.start or 0, s.stop or sys.maxint, s.step or 1))
    self._iter = iter(iterable)
    self._i = 0

  def __iter__(self):
    return self

  def next(self):
    nexti = next(self._range)
    while True:
      el = next(self._iter)
      i = self._i
      self._i += 1
      if i == nexti:
        return el

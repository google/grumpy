#!/usr/bin/env python

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

"""Convert a unified diff into a list of modified files and line numbers."""

import sys


class _LineBuffer(object):
  """Iterator over lines in a file supporting one-step rewind."""

  def __init__(self, f):
    self._f = f
    self._prev = None
    self._next = None

  def __iter__(self):
    return self

  def next(self):
    if self._next is not None:
      cur = self._next
    else:
      cur = self._f.readline()
    if not cur:
      raise StopIteration
    self._next = None
    self._prev = cur
    return cur

  def Rewind(self):
    assert self._prev is not None
    self._next = self._prev
    self._prev = None


def _ReadHunks(buf):
  for line in buf:
    if not line.startswith('@@'):
      break
    base = int(line.split()[2].split(',')[0])
    for offset in _ReadHunkBody(buf):
      yield base + offset


def _ReadHunkBody(buf):
  n = 0
  for line in buf:
    prefix = line[0]
    if prefix == ' ':
      n += 1
    elif prefix == '+':
      yield n
      n += 1
    elif prefix != '-' or line.startswith('---'):
      buf.Rewind()
      break


def main():
  buf = _LineBuffer(sys.stdin)
  for line in buf:
    if line.startswith('+++'):
      filename = line.split()[1]
      for n in _ReadHunks(buf):
        print '{}:{}'.format(filename, n)


if __name__ == '__main__':
  main()

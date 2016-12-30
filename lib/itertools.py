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

import _collections
import sys


def chain(*iterables):
  for it in iterables:
    for element in it:
      yield element


def count(start=0, step=1):
  n = start
  while True:
    yield n
    n += step


def imap(function, *iterables):
  iterables = map(iter, iterables)
  while True:
    args = [next(it) for it in iterables]
    if function is None:
      yield tuple(args)
    else:
      yield function(*args)


def islice(iterable, *args):
  s = slice(*args)
  it = iter(xrange(s.start or 0, s.stop or sys.maxint, s.step or 1))
  nexti = next(it)
  for i, element in enumerate(iterable):
    if i == nexti:
      yield element
      nexti = next(it)


def izip(*iterables):
  iterators = map(iter, iterables)
  while iterators:
    yield tuple(map(next, iterators))


def repeat(object, times=None):
  if times is None:
    while True:
      yield object
  else:
    for i in xrange(times):
      yield object


def starmap(function, iterable):
  for args in iterable:
    yield function(*args)


def tee(iterable, n=2):
  it = iter(iterable)
  deques = [_collections.deque() for i in range(n)]
  def gen(mydeque):
    while True:
      if not mydeque:
        newval = next(it)
        for d in deques:
          d.append(newval)
      yield mydeque.popleft()
  return tuple(gen(d) for d in deques)

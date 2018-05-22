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

class chain(object):

  def from_iterable(cls, iterables):
    for it in iterables:
      for element in it:
        yield element

  from_iterable = classmethod(from_iterable)

  def __init__(self, *iterables):
    if not iterables:
      self.iterables = iter([[]])
    else:
      self.iterables = iter(iterables)
    self.curriter = iter(next(self.iterables))

  def __iter__(self):
    return self

  def next(self):
    flag = True
    while flag:
      try:
        ret = next(self.curriter)
        flag = False
      except StopIteration:
        self.curriter = iter(next(self.iterables))
    return ret


def compress(data, selectors):
  return (d for d,s in izip(data, selectors) if s)


def count(start=0, step=1):
  n = start
  while True:
    yield n
    n += step


def cycle(iterable):
  saved = []
  for element in iterable:
    yield element
    saved.append(element)
  while saved:
    for element in saved:
      yield element


def dropwhile(predicate, iterable):
  iterable = iter(iterable)
  for x in iterable:
    if not predicate(x):
      yield x
      break
  for x in iterable:
    yield x


class groupby(object):
  # [k for k, g in groupby('AAAABBBCCDAABBB')] --> A B C D A B
  # [list(g) for k, g in groupby('AAAABBBCCD')] --> AAAA BBB CC D
  def __init__(self, iterable, key=None):
    if key is None:
      key = lambda x: x
    self.keyfunc = key
    self.it = iter(iterable)
    self.tgtkey = self.currkey = self.currvalue = object()

  def __iter__(self):
    return self

  def next(self):
    while self.currkey == self.tgtkey:
      self.currvalue = next(self.it)    # Exit on StopIteration
      self.currkey = self.keyfunc(self.currvalue)
    self.tgtkey = self.currkey
    return (self.currkey, self._grouper(self.tgtkey))
  
  def _grouper(self, tgtkey):
    while self.currkey == tgtkey:
      yield self.currvalue
      self.currvalue = next(self.it)    # Exit on StopIteration
      self.currkey = self.keyfunc(self.currvalue)


def ifilter(predicate, iterable):
  if predicate is None:
    predicate = bool
  for x in iterable:
    if predicate(x):
       yield x


def ifilterfalse(predicate, iterable):
  if predicate is None:
    predicate = bool
  for x in iterable:
    if not predicate(x):
       yield x


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


class ZipExhausted(Exception):
  pass


def izip_longest(*args, **kwds):
  # izip_longest('ABCD', 'xy', fillvalue='-') --> Ax By C- D-
  fillvalue = kwds.get('fillvalue')
  counter = [len(args) - 1]
  def sentinel():
    if not counter[0]:
      raise ZipExhausted
    counter[0] -= 1
    yield fillvalue
  fillers = repeat(fillvalue)
  iterators = [chain(it, sentinel(), fillers) for it in args]
  try:
    while iterators:
      yield tuple(map(next, iterators))
  except ZipExhausted:
    pass


def product(*args, **kwds):
  # product('ABCD', 'xy') --> Ax Ay Bx By Cx Cy Dx Dy
  # product(range(2), repeat=3) --> 000 001 010 011 100 101 110 111
  pools = map(tuple, args) * kwds.get('repeat', 1)
  result = [[]]
  for pool in pools:
    result = [x+[y] for x in result for y in pool]
  for prod in result:
    yield tuple(prod)


def permutations(iterable, r=None):
  pool = tuple(iterable)
  n = len(pool)
  r = n if r is None else r
  for indices in product(range(n), repeat=r):
    if len(set(indices)) == r:
      yield tuple(pool[i] for i in indices)


def combinations(iterable, r):
  pool = tuple(iterable)
  n = len(pool)
  for indices in permutations(range(n), r):
    if sorted(indices) == list(indices):
      yield tuple(pool[i] for i in indices)


def combinations_with_replacement(iterable, r):
  pool = tuple(iterable)
  n = len(pool)
  for indices in product(range(n), repeat=r):
    if sorted(indices) == list(indices):
      yield tuple(pool[i] for i in indices)


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


def takewhile(predicate, iterable):
  for x in iterable:
    if predicate(x):
      yield x
    else:
      break


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


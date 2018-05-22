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

import itertools

import weetest

def TestCycle():
  want = []
  got = []
  for x in itertools.cycle(()):
      got.append(x)
  assert got == want, 'empty cycle yields no elements'

  arg = (0, 1, 2)
  want = (0, 1, 2) * 10
  got = []
  limit = 10 * len(arg)
  counter = 0
  for x in itertools.cycle((0, 1, 2)):
    got.append(x)
    counter += 1
    if counter == limit:
      break
  assert tuple(got) == want, 'tuple(cycle%s) == %s, want %s' % (arg, tuple(got), want)


def TestDropwhile():
  r = range(10)
  cases = [
    ((lambda x: x < 5, r), (5, 6, 7, 8, 9)),
    ((lambda x: True, r), ()),
    ((lambda x: False, r), tuple(r)),
  ]
  for args, want in cases:
    got = tuple(itertools.dropwhile(*args))
    assert got == want, 'tuple(dropwhile%s) == %s, want %s' % (args, got, want)


def TestChain():
  r = range(10)
  cases = [
    ([r], tuple(r)),
    ([r, r], tuple(r) + tuple(r)),
    ([], ())
  ]
  for args, want in cases:
    got = tuple(itertools.chain(*args))
    assert got == want, 'tuple(chain%s) == %s, want %s' % (args, got, want)


def TestFromIterable():
  r = range(10)
  cases = [
    ([r], tuple(r)),
    ([r, r], tuple(r) + tuple(r)),
    ([], ())
  ]
  for args, want in cases:
    got = tuple(itertools.chain.from_iterable(args))
    assert got == want, 'tuple(from_iterable%s) == %s, want %s' % (args, got, want)


def TestIFilter():
  r = range(10)
  cases = [
    ((lambda x: x < 5, r), (0, 1, 2, 3, 4)),
    ((lambda x: False, r), ()),
    ((lambda x: True, r), tuple(r)),
    ((None, r), (1, 2, 3, 4, 5, 6, 7, 8, 9))
  ]
  for args, want in cases:
    got = tuple(itertools.ifilter(*args))
    assert got == want, 'tuple(ifilter%s) == %s, want %s' % (args, got, want)


def TestIFilterFalse():
  r = range(10)
  cases = [
    ((lambda x: x < 5, r), (5, 6, 7, 8, 9)),
    ((lambda x: False, r), tuple(r)),
    ((lambda x: True, r), ()),
    ((None, r), (0,))
  ]
  for args, want in cases:
    got = tuple(itertools.ifilterfalse(*args))
    assert got == want, 'tuple(ifilterfalse%s) == %s, want %s' % (args, got, want)


def TestISlice():
  r = range(10)
  cases = [
      ((r, 5), (0, 1, 2, 3, 4)),
      ((r, 25, 30), ()),
      ((r, 1, None, 3), (1, 4, 7)),
  ]
  for args, want in cases:
    got = tuple(itertools.islice(*args))
    assert got == want, 'tuple(islice%s) == %s, want %s' % (args, got, want)


def TestIZipLongest():
  cases = [
    (('abc', range(6)), (('a', 0), ('b', 1), ('c', 2), (None, 3), (None, 4), (None, 5))),
    ((range(6), 'abc'), ((0, 'a'), (1, 'b'), (2, 'c'), (3, None), (4, None), (5, None))),
    (([1, None, 3], 'ab', range(1)), ((1, 'a', 0), (None, 'b', None), (3, None, None))),
  ]
  for args, want in cases:
    got = tuple(itertools.izip_longest(*args))
    assert got == want, 'tuple(izip_longest%s) == %s, want %s' % (args, got, want)


def TestProduct():
  cases = [
    (([1, 2], ['a', 'b']), ((1, 'a'), (1, 'b'), (2, 'a'), (2, 'b'))),
    (([1], ['a', 'b']), ((1, 'a'), (1, 'b'))),
    (([],), ()),
  ]
  for args, want in cases:
    got = tuple(itertools.product(*args))
    assert got == want, 'tuple(product%s) == %s, want %s' % (args, got, want)


def TestPermutations():
  cases = [
    (('AB',), (('A', 'B'), ('B', 'A'))),
    (('ABC', 2), (('A', 'B'), ('A', 'C'), ('B', 'A'), ('B', 'C'), ('C', 'A'), ('C', 'B'))),
    ((range(3),), ((0, 1, 2), (0, 2, 1), (1, 0, 2), (1, 2, 0), (2, 0, 1), (2, 1, 0))),
    (([],), ((),)),
    (([], 0), ((),)),
    ((range(3), 4), ()),
  ]
  for args, want in cases:
    got = tuple(itertools.permutations(*args))
    assert got == want, 'tuple(permutations%s) == %s, want %s' % (args, got, want)


def TestCombinations():
  cases = [
    ((range(4), 3), ((0, 1, 2), (0, 1, 3), (0, 2, 3), (1, 2, 3))),
  ]
  for args, want in cases:
    got = tuple(itertools.combinations(*args))
    assert got == want, 'tuple(combinations%s) == %s, want %s' % (args, got, want)


def TestCombinationsWithReplacement():
  cases = [
    (([-12], 2), (((-12, -12),))),
    (('AB', 3), (('A', 'A', 'A'), ('A', 'A', 'B'), ('A', 'B', 'B'), ('B', 'B', 'B'))),
    (([], 2), ()),
    (([], 0), ((),))
  ]
  for args, want in cases:
    got = tuple(itertools.combinations_with_replacement(*args))
    assert got == want, 'tuple(combinations_with_replacement%s) == %s, want %s' % (args, got, want)


def TestGroupBy():
  cases = [
    (([1, 2, 2, 3, 3, 3, 4, 4, 4, 4],), [(1, [1]), (2, [2, 2]), (3, [3, 3, 3]), (4, [4, 4, 4, 4])]),
    ((['aa', 'ab', 'abc', 'bcd', 'abcde'], len), [(2, ['aa', 'ab']), (3, ['abc', 'bcd']), (5, ['abcde'])]),
  ]
  for args, want in cases:
    got = [(k, list(v)) for k, v in itertools.groupby(*args)]
    assert got == want, 'groupby %s == %s, want %s' % (args, got, want)


def TestTakewhile():
  r = range(10)
  cases = [
    ((lambda x: x % 2 == 0, r), (0,)),
    ((lambda x: True, r), tuple(r)),
    ((lambda x: False, r), ())
  ]
  for args, want in cases:
    got = tuple(itertools.takewhile(*args))
    assert got == want, 'tuple(takewhile%s) == %s, want %s' % (args, got, want)


if __name__ == '__main__':
  weetest.RunTests()

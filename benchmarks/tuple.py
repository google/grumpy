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

"""Benchmarks for list operations."""

# pylint: disable=pointless-statement

import weetest


def BenchmarkTupleGetItem(b):
  l = (1, 3, 9)
  for _ in xrange(b.N):
    l[2]


def BenchmarkTupleContains3(b):
  t = (1, 3, 9)
  for _ in xrange(b.N):
    9 in t


def BenchmarkTupleContains10(b):
  t = tuple(range(10))
  for _ in xrange(b.N):
    9 in t


if __name__ == '__main__':
  weetest.RunBenchmarks()

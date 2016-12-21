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

"""Benchmarks for comprehensions."""

# pylint: disable=unused-argument

import weetest


def BenchmarkGeneratorExpCreate(b):
  l = []
  for _ in xrange(b.N):
    (x for x in l)  # pylint: disable=pointless-statement


def BenchmarkGeneratorExpIterate(b):
  for _ in (x for x in xrange(b.N)):
    pass


def BenchmarkListCompCreate(b):
  for _ in xrange(b.N):
    [x for x in xrange(1000)]  # pylint: disable=expression-not-assigned


def BenchmarkDictCompCreate(b):
  for _ in xrange(b.N):
    {x: x for x in xrange(1000)}  # pylint: disable=expression-not-assigned


if __name__ == '__main__':
  weetest.RunBenchmarks()

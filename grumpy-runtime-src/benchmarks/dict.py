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

"""Benchmarks for dictionary operations."""

# pylint: disable=pointless-statement

import weetest


def BenchmarkDictCreate(b):  
  for _ in xrange(b.N):
    d = {'one': 1, 'two': 2, 'three': 3}


def BenchmarkDictCreateFunc(b):  
  for _ in xrange(b.N):
    d = dict(one=1, two=2, three=3)

    
def BenchmarkDictGetItem(b):
  d = {42: 123}
  for _ in xrange(b.N):
    d[42]


def BenchmarkDictStringOnlyGetItem(b):
  d = {'foo': 123}
  for _ in xrange(b.N):
    d['foo']


def BenchmarkDictSetItem(b):
  d = {}
  for _ in xrange(b.N):
    d[42] = 123


def BenchmarkDictStringOnlySetItem(b):
  d = {}
  for _ in xrange(b.N):
    d['foo'] = 123


def BenchmarkHashStrCached(b):
  """Hashes the same value repeatedly to exercise any hash caching logic."""
  h = hash  # Prevent builtins lookup each iteration.
  for _ in xrange(b.N):
    h('foobarfoobarfoobar')


if __name__ == '__main__':
  weetest.RunBenchmarks()

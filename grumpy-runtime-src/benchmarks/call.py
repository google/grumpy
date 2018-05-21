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

"""Benchmarks for function calls."""

# pylint: disable=unused-argument

import weetest


def BenchmarkCallNoArgs(b):
  def Foo():
    pass
  for _ in xrange(b.N):
    Foo()


def BenchmarkCallPositionalArgs(b):
  def Foo(a, b, c):
    pass
  for _ in xrange(b.N):
    Foo(1, 2, 3)


def BenchmarkCallKeywords(b):
  def Foo(a, b, c):
    pass
  for _ in xrange(b.N):
    Foo(a=1, b=2, c=3)


def BenchmarkCallDefaults(b):
  def Foo(a=1, b=2, c=3):
    pass
  for _ in xrange(b.N):
    Foo()


def BenchmarkCallVarArgs(b):
  def Foo(*args):
    pass
  for _ in xrange(b.N):
    Foo(1, 2, 3)


def BenchmarkCallKwargs(b):
  def Foo(**kwargs):
    pass
  for _ in xrange(b.N):
    Foo(a=1, b=2, c=3)


if __name__ == '__main__':
  weetest.RunBenchmarks()

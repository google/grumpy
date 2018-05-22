"""Microbenchmark for function call overhead.

This measures simple function calls that are not methods, do not use varargs or
kwargs, and do not use tuple unpacking.

Taken from:
https://github.com/python/performance/blob/9b8d859/performance/benchmarks/bm_call_simple.py
"""

import weetest


def foo(a, b, c, d):
  # 20 calls
  bar(a, b, c)
  bar(a, b, c)
  bar(a, b, c)
  bar(a, b, c)
  bar(a, b, c)
  bar(a, b, c)
  bar(a, b, c)
  bar(a, b, c)
  bar(a, b, c)
  bar(a, b, c)
  bar(a, b, c)
  bar(a, b, c)
  bar(a, b, c)
  bar(a, b, c)
  bar(a, b, c)
  bar(a, b, c)
  bar(a, b, c)
  bar(a, b, c)
  bar(a, b, c)
  bar(a, b, c)


def bar(a, b, c):
  # 20 calls
  baz(a, b)
  baz(a, b)
  baz(a, b)
  baz(a, b)
  baz(a, b)
  baz(a, b)
  baz(a, b)
  baz(a, b)
  baz(a, b)
  baz(a, b)
  baz(a, b)
  baz(a, b)
  baz(a, b)
  baz(a, b)
  baz(a, b)
  baz(a, b)
  baz(a, b)
  baz(a, b)
  baz(a, b)
  baz(a, b)


def baz(a, b):
  # 20 calls
  quux(a)
  quux(a)
  quux(a)
  quux(a)
  quux(a)
  quux(a)
  quux(a)
  quux(a)
  quux(a)
  quux(a)
  quux(a)
  quux(a)
  quux(a)
  quux(a)
  quux(a)
  quux(a)
  quux(a)
  quux(a)
  quux(a)
  quux(a)


def quux(a):
  # 20 calls
  qux()
  qux()
  qux()
  qux()
  qux()
  qux()
  qux()
  qux()
  qux()
  qux()
  qux()
  qux()
  qux()
  qux()
  qux()
  qux()
  qux()
  qux()
  qux()
  qux()


def qux():
  pass


def BenchmarkCallSimple(b):
  for _ in xrange(b.N):
    # 20 calls
    foo(1, 2, 3, 4)
    foo(1, 2, 3, 4)
    foo(1, 2, 3, 4)
    foo(1, 2, 3, 4)
    foo(1, 2, 3, 4)
    foo(1, 2, 3, 4)
    foo(1, 2, 3, 4)
    foo(1, 2, 3, 4)
    foo(1, 2, 3, 4)
    foo(1, 2, 3, 4)
    foo(1, 2, 3, 4)
    foo(1, 2, 3, 4)
    foo(1, 2, 3, 4)
    foo(1, 2, 3, 4)
    foo(1, 2, 3, 4)
    foo(1, 2, 3, 4)
    foo(1, 2, 3, 4)
    foo(1, 2, 3, 4)
    foo(1, 2, 3, 4)
    foo(1, 2, 3, 4)


if __name__ == '__main__':
  weetest.RunBenchmarks()

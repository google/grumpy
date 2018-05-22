"""Microbenchmark for method call overhead.

This measures simple method calls that are predictable, do not use varargs or
kwargs, and do not use tuple unpacking.

Taken from:
https://github.com/python/performance/blob/9b8d859/performance/benchmarks/bm_call_method.py
"""

import weetest


class Foo(object):

  def foo(self, a, b, c, d):
    # 20 calls
    self.bar(a, b, c)
    self.bar(a, b, c)
    self.bar(a, b, c)
    self.bar(a, b, c)
    self.bar(a, b, c)
    self.bar(a, b, c)
    self.bar(a, b, c)
    self.bar(a, b, c)
    self.bar(a, b, c)
    self.bar(a, b, c)
    self.bar(a, b, c)
    self.bar(a, b, c)
    self.bar(a, b, c)
    self.bar(a, b, c)
    self.bar(a, b, c)
    self.bar(a, b, c)
    self.bar(a, b, c)
    self.bar(a, b, c)
    self.bar(a, b, c)
    self.bar(a, b, c)

  def bar(self, a, b, c):
    # 20 calls
    self.baz(a, b)
    self.baz(a, b)
    self.baz(a, b)
    self.baz(a, b)
    self.baz(a, b)
    self.baz(a, b)
    self.baz(a, b)
    self.baz(a, b)
    self.baz(a, b)
    self.baz(a, b)
    self.baz(a, b)
    self.baz(a, b)
    self.baz(a, b)
    self.baz(a, b)
    self.baz(a, b)
    self.baz(a, b)
    self.baz(a, b)
    self.baz(a, b)
    self.baz(a, b)
    self.baz(a, b)

  def baz(self, a, b):
    # 20 calls
    self.quux(a)
    self.quux(a)
    self.quux(a)
    self.quux(a)
    self.quux(a)
    self.quux(a)
    self.quux(a)
    self.quux(a)
    self.quux(a)
    self.quux(a)
    self.quux(a)
    self.quux(a)
    self.quux(a)
    self.quux(a)
    self.quux(a)
    self.quux(a)
    self.quux(a)
    self.quux(a)
    self.quux(a)
    self.quux(a)

  def quux(self, a):
    # 20 calls
    self.qux()
    self.qux()
    self.qux()
    self.qux()
    self.qux()
    self.qux()
    self.qux()
    self.qux()
    self.qux()
    self.qux()
    self.qux()
    self.qux()
    self.qux()
    self.qux()
    self.qux()
    self.qux()
    self.qux()
    self.qux()
    self.qux()
    self.qux()

  def qux(self):
    pass


def BenchmarkCallMethod(b):
  f = Foo()
  for _ in xrange(b.N):
    # 20 calls
    f.foo(1, 2, 3, 4)
    f.foo(1, 2, 3, 4)
    f.foo(1, 2, 3, 4)
    f.foo(1, 2, 3, 4)
    f.foo(1, 2, 3, 4)
    f.foo(1, 2, 3, 4)
    f.foo(1, 2, 3, 4)
    f.foo(1, 2, 3, 4)
    f.foo(1, 2, 3, 4)
    f.foo(1, 2, 3, 4)
    f.foo(1, 2, 3, 4)
    f.foo(1, 2, 3, 4)
    f.foo(1, 2, 3, 4)
    f.foo(1, 2, 3, 4)
    f.foo(1, 2, 3, 4)
    f.foo(1, 2, 3, 4)
    f.foo(1, 2, 3, 4)
    f.foo(1, 2, 3, 4)
    f.foo(1, 2, 3, 4)
    f.foo(1, 2, 3, 4)


if __name__ == '__main__':
  weetest.RunBenchmarks()

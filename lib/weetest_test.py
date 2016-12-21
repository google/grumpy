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

import os
import sys
import tempfile
import time

import weetest


class _Timer(object):

  def __init__(self):
    self.t = 0

  def Reset(self):
    self.t = 0

  def time(self):  # pylint: disable=invalid-name
    return self.t


_timer = _Timer()
# Stub out time for the whole test suite for convenience.
time.time = _timer.time


def TestBenchmark():
  def BenchmarkFoo(b):
    i = 0
    while i < b.N:
      i += 1.0
    _timer.t += b.N

  b = weetest._Benchmark(BenchmarkFoo, 1000)
  _timer.Reset()
  b.Run()
  assert b.duration >= 1000, b.duration
  assert b.N >= 1000


def TestBenchmarkResetTimer():
  def BenchmarkFoo(b):
    _timer.t += 10000  # Do an "expensive" initialization.
    b.ResetTimer()
    i = 0
    while i < b.N:
      i += 1.0
    _timer.t += b.N

  b = weetest._Benchmark(BenchmarkFoo, 1000)
  _timer.Reset()
  b.Run()
  assert b.duration < 2000, b.duration
  assert b.N < 2000, b.duration


def TestRunOneBenchmark():
  def BenchmarkFoo(b):
    i = 0
    while i < b.N:
      i += 1.0
    _timer.t += b.N

  _timer.Reset()
  result = weetest._RunOneBenchmark('BenchmarkFoo', BenchmarkFoo)

  assert result.status == 'passed'
  assert result.properties.get('ops_per_sec') == 1.0
  assert result.duration == _timer.time()


def TestRunOneBenchmarkError():
  def BenchmarkFoo(unused_b):
    raise ValueError
  result = weetest._RunOneBenchmark('BenchmarkFoo', BenchmarkFoo)
  assert result.status == 'error'


def TestRunOneTest():
  def TestFoo():
    # pylint: disable=undefined-loop-variable
    _timer.t += 100
    if case[0]:
      raise case[0]  # pylint: disable=raising-bad-type

  cases = [(None, 'passed'),
           (AssertionError, 'failed'),
           (ValueError, 'error')]
  for case in cases:
    _timer.Reset()
    result = weetest._RunOneTest('TestFoo', TestFoo)
    assert result.status == case[1]
    assert result.duration == 100


def TestWriteXmlFile():
  result_with_properties = weetest._TestResult('foo')
  result_with_properties.properties['bar'] = 'baz'
  cases = [([], ['<testsuite ', 'tests="0"', '</testsuite>']),
           ([weetest._TestResult('foo')],
            ['<testsuite ', 'tests="1"', '<testcase name="foo"']),
           ([weetest._TestResult('foo'), weetest._TestResult('bar')],
            ['tests="2"', '<testcase name="foo"', '<testcase name="bar"']),
           ([result_with_properties],
            ['<testcase name="foo"',
             '<property name="bar" value="baz"></property>'])]
  for case in cases:
    fd, path = tempfile.mkstemp()
    os.close(fd)
    try:
      weetest._WriteXmlFile(path, 100, case[0])
      f = open(path)
      contents = f.read()
      f.close()
      for want in case[1]:
        assert want in contents, contents
    finally:
      os.remove(path)


def TestRunAll():
  class Main(object):

    def __init__(self):
      self.run = {}

    def TestFoo(self):
      self.run['TestFoo'] = True

    def TestBar(self):
      self.run['TestBar'] = True

    def BenchmarkBaz(self, b):
      self.run['BenchmarkBaz'] = True
      _timer.t += b.N

    def BenchmarkQux(self, b):
      self.run['BenchmarkQux'] = True
      _timer.t += b.N

  fd, test_xml = tempfile.mkstemp()
  os.close(fd)
  fd, benchmark_xml = tempfile.mkstemp()
  os.close(fd)
  cases = [('Test', weetest._RunOneTest, None,
            {'TestFoo': True, 'TestBar': True}),
           ('Test', weetest._RunOneTest, test_xml,
            {'TestFoo': True, 'TestBar': True}),
           ('Benchmark', weetest._RunOneBenchmark, None,
            {'BenchmarkBaz': True, 'BenchmarkQux': True}),
           ('Benchmark', weetest._RunOneBenchmark, benchmark_xml,
            {'BenchmarkBaz': True, 'BenchmarkQux': True})]
  for prefix, runner, xml_path, want_run in cases:
    old_main = sys.modules['__main__']
    sys.modules['__main__'] = main = Main()
    if xml_path:
      os.environ['XML_OUTPUT_FILE'] = xml_path
    try:
      weetest._RunAll(prefix, runner)
    finally:
      sys.modules['__main__'] = old_main
      if 'XML_OUTPUT_FILE' in os.environ:
        del os.environ['XML_OUTPUT_FILE']
    assert main.run == want_run, main.run
    if xml_path:
      f = open(xml_path)
      xml = f.read()
      f.close()
      os.remove(xml_path)
      for name in want_run:
        assert '<testcase name="%s"' % name in xml, xml


if __name__ == '__main__':
  # Using keys() avoids "dictionary changed size during iteration" error.
  for test_name in globals().keys():
    if test_name.startswith('Test'):
      globals()[test_name]()

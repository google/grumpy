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

"""Minimal framework for writing tests and benchmarks.

Follows a similar pattern to the Go testing package functionality. To implement
and have test and benchmark functions run automatically, define
Test/BenchmarkXyz() functions in your __main__ module and run them with
weetest.RunTests/Benchmarks() For example:

    import weetest

    def BenchmarkOneThing(b):
      for i in xrange(b.N):
        OneThing()

    def BenchmarkAnotherThing(b):
      i = 0
      while i < b.N:
        AnotherThing()
        i += 1

    if __name__ == '__main__':
      weetest.RunBenchmarks()
"""

import os
import sys
import time
import traceback


# pylint: disable=invalid-name
class _Benchmark(object):
  """Wraps and runs a single user defined benchmark function."""

  def __init__(self, bench_func, target_duration):
    """Set up this benchmark to run bench_func to be run for target_duration."""
    self.bench_func = bench_func
    self.target_duration = target_duration
    self.start_time = 0
    self.N = 1

  def Run(self):
    """Attempt to run this benchmark for the target duration."""
    small_duration = 0.05 * self.target_duration
    self.N = 1
    self._RunOnce()
    while self.duration < self.target_duration:
      if self.duration < small_duration:
        # Grow N very quickly when duration is small.
        N = self.N * 10
      else:
        # Once duration is > 5% of target_duration we should have a good
        # estimate for how many iterations it will take to hit the target. Shoot
        # for 20% beyond that so we don't end up hovering just below the target.
        N = int(self.N * (1.2 * self.target_duration / self.duration))
      if self.N == N:
        self.N += 1
      else:
        self.N = N
      self._RunOnce()

  def _RunOnce(self):
    self.start_time = time.time()
    self.bench_func(self)
    self.duration = time.time() - self.start_time

  def ResetTimer(self):
    """Clears the current elapsed time to discount expensive setup steps."""
    self.start_time = time.time()


class _TestResult(object):
  """The outcome of running a particular benchmark function."""

  def __init__(self, name):
    self.name = name
    self.status = 'not run'
    self.duration = 0
    self.properties = {}


def _RunOneBenchmark(name, test_func):
  """Runs a single benchmark and returns a _TestResult."""
  b = _Benchmark(test_func, 1)
  result = _TestResult(name)
  print name,
  start_time = time.time()
  try:
    b.Run()
  except Exception as e:  # pylint: disable=broad-except
    result.status = 'error'
    print 'ERROR'
    traceback.print_exc()
  else:
    result.status = 'passed'
    ops_per_sec = b.N / b.duration
    result.properties['ops_per_sec'] = ops_per_sec
    print 'PASSED', ops_per_sec
  finally:
    result.duration = time.time() - start_time
  return result


def _RunOneTest(name, test_func):
  """Runs a single test function and returns a _TestResult."""
  result = _TestResult(name)
  start_time = time.time()
  try:
    test_func()
  except AssertionError as e:
    result.status = 'failed'
    print name, 'FAILED'
    traceback.print_exc()
  except Exception as e:  # pylint: disable=broad-except
    result.status = 'error'
    print name, 'ERROR'
    traceback.print_exc()
  else:
    result.status = 'passed'
  finally:
    result.duration = time.time() - start_time
  return result


def _WriteXmlFile(filename, suite_duration, results):
  """Given a list of _BenchmarkResults, writes XML test results to filename."""
  xml_file = open(filename, 'w')
  xml_file.write('<testsuite name="%s" tests="%s" '
                 'time="%f" runner="weetest">\n' %
                 (sys.argv[0], len(results), suite_duration))
  for result in results:
    xml_file.write('  <testcase name="%s" result="completed" '
                   'status="run" time="%f">\n' %
                   (result.name, result.duration))
    if result.properties:
      xml_file.write('    <properties>\n')
      for name in result.properties:
        value = result.properties[name]
        if isinstance(value, float):
          formatted = '%f' % value
        else:
          formatted = str(value)
        xml_file.write('      <property name="%s" value="%s"></property>\n' %
                       (name, formatted))
      xml_file.write('    </properties>\n')
    xml_file.write('  </testcase>\n')
  xml_file.write('</testsuite>')
  xml_file.close()


def _RunAll(test_prefix, runner):
  """Runs all functions in __main__ matching test_prefix using runner."""
  target = os.environ.get('WEETEST_TARGET')
  exit_status = 0
  mod = sys.modules['__main__']
  results = []
  suite_start_time = time.time()
  for name in dir(mod):
    if name.startswith(test_prefix) and (not target or name == target):
      result = runner(name, getattr(mod, name))
      if result.status != 'passed':
        exit_status = 1
      results.append(result)
  suite_duration = time.time() - suite_start_time
  if 'XML_OUTPUT_FILE' in os.environ:
    _WriteXmlFile(os.environ['XML_OUTPUT_FILE'], suite_duration, results)
  return exit_status


def RunBenchmarks():
  """Benchmarks all functions in __main__ with names like BenchmarkXyz()."""
  sys.exit(_RunAll('Benchmark', _RunOneBenchmark))


def RunTests():
  """Runs all functions in __main__ with names like TestXyz()."""
  sys.exit(_RunAll('Test', _RunOneTest))

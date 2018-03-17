#!/usr/bin/env python

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

"""Runs two benchmark programs and compares their results."""

import argparse
import subprocess
import sys


parser = argparse.ArgumentParser()
parser.add_argument('prog1')
parser.add_argument('prog2')
parser.add_argument('--runs', default=1, type=int,
                    help='number of times to run each program')


def main(args):
  results1 = _RunBenchmark(args.prog1)
  benchmarks = set(results1.keys())
  results2 = {}
  for _ in xrange(args.runs - 1):
    _MergeResults(results1, _RunBenchmark(args.prog1), benchmarks)
    _MergeResults(results2, _RunBenchmark(args.prog2), benchmarks)
  _MergeResults(results2, _RunBenchmark(args.prog2), benchmarks)
  for b in sorted(benchmarks):
    print b, '{:+.1%}'.format(results2[b] / results1[b] - 1)


def _MergeResults(merged, results, benchmarks):
  benchmarks = set(benchmarks)
  for k, v in results.iteritems():
    if k not in benchmarks:
      _Die('unmatched benchmark: {}', k)
    merged[k] = max(merged.get(k, 0), v)
    benchmarks.remove(k)
  if benchmarks:
    _Die('missing benchmark(s): {}', ', '.join(benchmarks))


def _RunBenchmark(prog):
  """Executes prog and returns a dict mapping benchmark name -> result."""
  try:
    p = subprocess.Popen([prog], shell=True, stdout=subprocess.PIPE)
  except OSError as e:
    _Die(e)
  out, _ = p.communicate()
  if p.returncode:
    _Die('{} exited with status: {}', prog, p.returncode)
  results = {}
  for line in out.splitlines():
    line = line.strip()
    if not line:
      continue
    parts = line.split()
    if len(parts) != 3:
      _Die('invalid benchmark output: {}', line)
    name, status, result = parts
    if status != 'PASSED':
      _Die('benchmark failed: {}', line)
    try:
      result = float(result)
    except ValueError:
      _Die('invalid benchmark result: {}', line)
    results[name] = result
  return results


def _Die(msg, *args):
  if args:
    msg = msg.format(*args)
  print >> sys.stderr, msg
  sys.exit(1)


if __name__ == '__main__':
  main(parser.parse_args())

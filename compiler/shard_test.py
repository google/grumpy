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

"""Wrapper for unit tests that loads a subset of all test methods."""

from __future__ import unicode_literals

import argparse
import random
import re
import sys
import unittest


class _ShardAction(argparse.Action):

  def __call__(self, parser, args, values, option_string=None):
    match = re.match(r'(\d+)of(\d+)$', values)
    if not match:
      raise argparse.ArgumentError(self, 'bad shard spec: {}'.format(values))
    shard = int(match.group(1))
    count = int(match.group(2))
    if shard < 1 or count < 1 or shard > count:
      raise argparse.ArgumentError(self, 'bad shard spec: {}'.format(values))
    setattr(args, self.dest, (shard, count))


class _ShardTestLoader(unittest.TestLoader):

  def __init__(self, shard, count):
    super(_ShardTestLoader, self).__init__()
    self.shard = shard
    self.count = count

  def getTestCaseNames(self, test_case_cls):
    names = super(_ShardTestLoader, self).getTestCaseNames(test_case_cls)
    state = random.getstate()
    random.seed(self.count)
    random.shuffle(names)
    random.setstate(state)
    n = len(names)
    # self.shard is one-based.
    return names[(self.shard - 1) * n / self.count:self.shard * n / self.count]


class _ShardTestRunner(object):

  def run(self, test):
    result = unittest.TestResult()
    unittest.registerResult(result)
    test(result)
    for kind, errors in [('FAIL', result.failures), ('ERROR', result.errors)]:
      for test, err in errors:
        sys.stderr.write('{} {}\n{}'.format(test, kind, err))
    return result


def main():
  parser = argparse.ArgumentParser()
  parser.add_argument('--shard', default=(1, 1), action=_ShardAction)
  parser.add_argument('unittest_args', nargs='*')
  args = parser.parse_args()
  unittest.main(argv=[sys.argv[0]] + args.unittest_args,
                testLoader=_ShardTestLoader(*args.shard),
                testRunner=_ShardTestRunner)

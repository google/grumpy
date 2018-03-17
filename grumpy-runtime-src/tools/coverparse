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

"""Parse a Go coverage file and prints a message for lines missing coverage."""

import collections
import re
import sys


cover_re = re.compile(r'([^:]+):(\d+)\.\d+,(\d+).\d+ \d+ (\d+)$')


def _ParseCover(f):
  """Return a dict of sets with uncovered line numbers from a Go cover file."""
  uncovered = collections.defaultdict(set)
  for line in f:
    match = cover_re.match(line.rstrip())
    if not match:
      raise RuntimeError('invalid coverage line: {!r}'.format(line))
    filename, line_start, line_end, count = match.groups()
    if not int(count):
      for i in xrange(int(line_start), int(line_end) + 1):
        uncovered[filename].add(i)
  return uncovered


def main():
  with open(sys.argv[1]) as f:
    f.readline()
    uncovered = _ParseCover(f)
  for filename in sorted(uncovered.keys()):
    for lineno in sorted(uncovered[filename]):
      print '{}:{}'.format(filename, lineno)


if __name__ == '__main__':
  main()

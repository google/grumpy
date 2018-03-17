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

d = {'foo': 1, 'bar': 2, 'baz': 3}
try:
  d['qux']
except KeyError:
  pass

assert d['foo'] == 1
assert d['bar'] == 2
assert d['baz'] == 3

d['qux'] = 4
assert d['qux'] == 4

d['foo'] = 5
assert d['foo'] == 5

l = []
for k in d:
  l.append(k)
assert l == ['baz', 'foo', 'bar', 'qux']

try:
  for k in d:
    d['quux'] = 6
except RuntimeError:
  pass
else:
  raise AssertionError

d = {'foo': 1, 'bar': 2, 'baz': 3}
del d['bar']
assert d == {'foo': 1, 'baz': 3}
try:
  del d['bar']
except KeyError:
  pass
else:
  raise AssertionError

# Test clear
d = {1: 1, 2: 2, 3: 3}
d.clear()
assert d == {}

try:
  d.clear()
  assert AssertionError
except TypeError:
  pass

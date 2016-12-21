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

# pylint: disable=unreachable

while False:
  raise AssertionError

while True:
  break
  raise AssertionError

l = []
foo = True
while foo:
  l.append(1)
  foo = False
else:  # pylint: disable=useless-else-on-loop
  l.append(2)
assert l == [1, 2]

l = []
while True:
  break
  l.append(1)
else:
  l.append(2)
assert not l

l = []
foo = True
while foo:
  foo = False
  l.append(1)
  continue
  l.append(2)
assert l == [1]

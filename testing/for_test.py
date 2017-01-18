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

l = []
for i in (1, 2, 3):
  l.append(i)
assert l == [1, 2, 3]

l = []
for i in ():
  l.append(1)
else:  # pylint: disable=useless-else-on-loop
  l.append(2)
assert l == [2]

l = []
for i in (1,):
  l.append(i)
else:  # pylint: disable=useless-else-on-loop
  l.append(2)
assert l == [1, 2]

l = []
for i in (1,):
  l.append(i)
  break
else:
  l.append(2)
assert l == [1]

l = []
for i in (1, 2):
  l.append(i)
  continue
  l.append(3)  # pylint: disable=unreachable
assert l == [1, 2]

l = []
for i, j in [('a', 1), ('b', 2)]:
  l.append(i)
  l.append(j)
assert l == ['a', 1, 'b', 2]

# break and continue statements in an else clause applies to the outer loop.
# See: https://github.com/google/grumpy/issues/123
l = []
for i in range(2):
  l.append(i)
  for j in range(10, 12):
    l.append(j)
  else:
    l.append(12)
    continue
  l.append(-1)
assert l == [0, 10, 11, 12, 1, 10, 11, 12]

l = []
for i in range(10):
  l.append(i)
  for j in range(10, 12):
    l.append(j)
  else:
    break
  l.append(-1)
assert l == [0, 10, 11]

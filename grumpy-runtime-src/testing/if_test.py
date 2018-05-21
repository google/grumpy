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

# pylint: disable=using-constant-test

foo = []
if True:
  foo.append(1)
else:
  foo.append(2)
assert foo == [1]

foo = []
if False:
  foo.append(1)
else:
  foo.append(2)
assert foo == [2]

foo = []
if False:
  foo.append(1)
elif False:
  foo.append(2)
elif True:
  foo.append(3)
assert foo == [3]

foo = []
if False:
  foo.append(1)
elif True:
  foo.append(2)
elif True:
  foo.append(3)
else:
  foo.append(4)
assert foo == [2]

foo = []
if False:
  foo.append(1)
elif False:
  foo.append(2)
elif False:
  foo.append(3)
else:
  foo.append(4)
assert foo == [4]

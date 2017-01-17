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

"""Built-in Python identifiers."""

# pylint: disable=invalid-name

from __go__.grumpy import Builtins


for k, v in Builtins.iteritems():
  globals()[k] = v


# sorted()
def sorted(iterable, **kwargs):
    """sorted(iterable, cmp=None, key=None, reverse=False) --> new sorted list"""
    res = list(iterable)  # make a copy / expand the iterable
    # sort teh copy in place and return
    res.sort(**kwargs)
    return res

globals()["sorted"] = sorted


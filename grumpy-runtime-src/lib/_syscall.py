# Copyright 2017 Google Inc. All Rights Reserved.
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

from '__go__/syscall' import EINTR


def invoke(func, *args):
  while True:
    result = func(*args)
    if isinstance(result, tuple):
      err = result[-1]
      result = result[:-1]
    else:
      err = result
      result = ()
    if err:
      if err == EINTR:
        continue
      raise OSError(err.Error())
    return result

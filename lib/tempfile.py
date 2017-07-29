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

"""Generate temporary files and directories."""

# pylint: disable=g-multiple-import
from '__go__/io/ioutil' import TempDir, TempFile
from '__go__/syscall' import Dup


# pylint: disable=redefined-builtin
def mkdtemp(suffix='', prefix='tmp', dir=None):
  if dir is None:
    dir = ''
  # TODO: Make suffix actually follow the rest of the filename.
  path, err = TempDir(dir, prefix + '-' + suffix)
  if err:
    raise OSError(err.Error())
  return path


def mkstemp(suffix='', prefix='tmp', dir=None, text=False):
  if text:
    raise NotImplementedError
  if dir is None:
    dir = ''
  # TODO: Make suffix actually follow the rest of the filename.
  f, err = TempFile(dir, prefix + '-' + suffix)
  if err:
    raise OSError(err.Error())
  try:
    fd, err = Dup(f.Fd())
    if err:
      raise OSError(err.Error())
    return fd, f.Name()
  finally:
    f.Close()

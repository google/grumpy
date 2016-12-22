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

"""Miscellaneous operating system interfaces."""

# pylint: disable=g-multiple-import
from os import path
import stat as stat_module
import sys
from __go__.os import Chmod, Environ, Remove, Stat
from __go__.path.filepath import Separator
from __go__.grumpy import NewFileFromFD
from __go__.syscall import Close, SYS_FCNTL, Syscall, F_GETFD
from __go__.time import Second

sep = chr(Separator)

environ = {}
for var in Environ():
  k, v = var.split('=', 1)
  environ[k] = v


def chmod(filepath, mode):
  # TODO: Support mode flags other than perms.
  err = Chmod(filepath, stat(filepath).st_mode & ~0o777 | mode & 0o777)
  if err:
    raise OSError(err.Error())


def close(fd):
  err = Close(fd)
  if err:
    raise OSError(err.Error())


def fdopen(fd, mode='r'):  # pylint: disable=unused-argument
  # Ensure this is a valid file descriptor to match CPython behavior.
  _, _, err = Syscall(SYS_FCNTL, fd, F_GETFD, 0)
  if err:
    raise OSError(err.Error())
  return NewFileFromFD(fd)


def remove(filepath):
  if stat_module.S_ISDIR(stat(filepath).st_mode):
    raise OSError('Operation not permitted: ' + filepath)
  err = Remove(filepath)
  if err:
    raise OSError(err.Error())


def rmdir(filepath):
  if not stat_module.S_ISDIR(stat(filepath).st_mode):
    raise OSError('Operation not permitted: ' + filepath)
  err = Remove(filepath)
  if err:
    raise OSError(err.Error())


class StatResult(object):

  def __init__(self, info):
    self._info = info

  def st_mode(self):
    # TODO: This is an incomplete mode flag. It should include S_IFDIR, etc.
    return self._info.Mode()
  # TODO: Make this a decorator once they're implemented.
  st_mode = property(st_mode)

  def st_size(self):
    return self._info.Size()
  # TODO: Make this a decorator once they're implemented.
  st_size = property(st_size)

  def st_mtime(self):
    return float(self._info.ModTime().UnixNano()) / Second
  # TODO: Make this a decorator once they're implemented.
  st_mtime = property(st_mtime)


def stat(filepath):
  info, err = Stat(filepath)
  if err:
    raise OSError(err.Error())
  return StatResult(info)

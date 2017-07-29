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

""""Utilities for manipulating and inspecting OS paths."""

from '__go__/os' import Stat
from '__go__/path/filepath' import Abs, Base, Clean, Dir as dirname, IsAbs as isabs, Join, Split  # pylint: disable=g-multiple-import,unused-import


def abspath(path):
  result, err = Abs(path)
  if err:
    raise OSError(err.Error())
  if isinstance(path, unicode):
    # Grumpy compiler encoded the string into utf-8, so the result can be
    # decoded using utf-8.
    return unicode(result, 'utf-8')
  return result


def basename(path):
  return '' if path.endswith('/') else Base(path)


def exists(path):
  _, err = Stat(path)
  return err is None


def isdir(path):
  info, err = Stat(path)
  if info and err is None:
    return info.Mode().IsDir()
  return False


def isfile(path):
  info, err = Stat(path)
  if info and err is None:
    return info.Mode().IsRegular()
  return False


# NOTE(compatibility): This method uses Go's filepath.Join() method which
# implicitly normalizes the resulting path (pruning extra /, .., etc.) The usual
# CPython behavior is to leave all the cruft. This deviation is reasonable
# because a) result paths will point to the same files and b) one cannot assume
# much about the results of join anyway since it's platform dependent.
def join(*paths):
  if not paths:
    raise TypeError('join() takes at least 1 argument (0 given)')
  parts = []
  for p in paths:
    if isabs(p):
      parts = [p]
    else:
      parts.append(p)
  result = Join(*parts)
  if result and not paths[-1]:
    result += '/'
  return result


def normpath(path):
  result = Clean(path)
  if isinstance(path, unicode):
    return unicode(result, 'utf-8')
  return result


def split(path):
  head, tail = Split(path)
  if len(head) > 1 and head[-1] == '/':
    head = head[:-1]
  return (head, tail)

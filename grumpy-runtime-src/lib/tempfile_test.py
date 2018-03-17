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

import os
import stat
import tempfile

import weetest


def TestMkdTemp():
  path = tempfile.mkdtemp()
  mode = os.stat(path).st_mode
  os.rmdir(path)
  assert stat.S_ISDIR(mode), mode
  assert stat.S_IMODE(mode) == 0o700, mode


def TestMkdTempDir():
  tempdir = tempfile.mkdtemp()
  path = tempfile.mkdtemp(dir=tempdir)
  os.rmdir(path)
  os.rmdir(tempdir)
  assert path.startswith(tempdir)


def TestMkdTempOSError():
  tempdir = tempfile.mkdtemp()
  os.chmod(tempdir, 0o500)
  try:
    tempfile.mkdtemp(dir=tempdir)
  except OSError:
    pass
  else:
    raise AssertionError
  os.rmdir(tempdir)


def TestMkdTempPrefixSuffix():
  path = tempfile.mkdtemp(prefix='foo', suffix='bar')
  os.rmdir(path)
  assert 'foo' in path
  assert 'bar' in path
  # TODO: assert path.endswith('bar')


def TestMksTemp():
  fd, path = tempfile.mkstemp()
  f = os.fdopen(fd, 'w')
  f.write('foobar')
  f.close()
  f = open(path)
  contents = f.read()
  f.close()
  os.remove(path)
  assert contents == 'foobar', contents


def TestMksTempDir():
  tempdir = tempfile.mkdtemp()
  fd, path = tempfile.mkstemp(dir=tempdir)
  os.close(fd)
  os.remove(path)
  os.rmdir(tempdir)
  assert path.startswith(tempdir)


def TestMksTempOSError():
  tempdir = tempfile.mkdtemp()
  os.chmod(tempdir, 0o500)
  try:
    tempfile.mkstemp(dir=tempdir)
  except OSError:
    pass
  else:
    raise AssertionError
  os.rmdir(tempdir)


def TestMksTempPerms():
  fd, path = tempfile.mkstemp()
  os.close(fd)
  mode = os.stat(path).st_mode
  os.remove(path)
  assert stat.S_IMODE(mode) == 0o600, mode


def TestMksTempPrefixSuffix():
  fd, path = tempfile.mkstemp(prefix='foo', suffix='bar')
  os.close(fd)
  os.remove(path)
  assert 'foo' in path
  assert 'bar' in path
  # TODO: assert path.endswith('bar')


if __name__ == '__main__':
  weetest.RunTests()

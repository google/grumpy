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
import time
import tempfile

import weetest


def TestChmod():
  fd, path = tempfile.mkstemp()
  os.close(fd)
  os.chmod(path, 0o644)
  mode = os.stat(path).st_mode & 0o777
  os.remove(path)
  assert mode == 0o644


def TestChmodOSError():
  tempdir = tempfile.mkdtemp()
  try:
    os.chmod(tempdir + '/DoesNotExist', 0o644)
  except OSError:
    pass
  else:
    raise AssertionError


def TestClose():
  fd, _ = tempfile.mkstemp()
  os.close(fd)
  try:
    os.fdopen(fd)
  except OSError:
    pass
  else:
    raise AssertionError


def TestCloseOSError():
  fd, _ = tempfile.mkstemp()
  os.close(fd)
  try:
    os.close(fd)
  except OSError:
    pass
  else:
    raise AssertionError


def TestEnviron():
  assert 'HOME' in os.environ


def TestFDOpen():
  fd, path = tempfile.mkstemp()
  f = os.fdopen(fd, 'w')
  f.write('foobar')
  f.close()
  f = open(path)
  contents = f.read()
  f.close()
  assert contents == 'foobar', contents


def TestFDOpenOSError():
  fd, _ = tempfile.mkstemp()
  os.close(fd)
  try:
    os.fdopen(fd)
  except OSError:
    pass
  else:
    raise AssertionError


def TestRemove():
  fd, path = tempfile.mkstemp()
  os.close(fd)
  os.stat(path)
  os.remove(path)
  try:
    os.stat(path)
  except OSError:
    pass
  else:
    raise AssertionError


def TestRemoveNoExist():
  path = tempfile.mkdtemp()
  try:
    os.remove(path + '/nonexistent')
  except OSError:
    pass
  else:
    raise AssertionError
  finally:
    os.rmdir(path)


def TestRemoveDir():
  path = tempfile.mkdtemp()
  try:
    os.remove(path)
  except OSError:
    pass
  else:
    raise AssertionError
  finally:
    os.rmdir(path)


def TestRmDir():
  path = tempfile.mkdtemp()
  assert stat.S_ISDIR(os.stat(path).st_mode)
  os.rmdir(path)
  try:
    os.stat(path)
  except OSError:
    pass
  else:
    raise AssertionError


def TestRmDirNoExist():
  path = tempfile.mkdtemp()
  try:
    os.rmdir(path + '/nonexistent')
  except OSError:
    pass
  else:
    raise AssertionError
  finally:
    os.rmdir(path)


def TestRmDirFile():
  fd, path = tempfile.mkstemp()
  os.close(fd)
  try:
    os.rmdir(path)
  except OSError:
    pass
  else:
    raise AssertionError
  finally:
    os.remove(path)


def TestStatFile():
  t = time.time()
  fd, path = tempfile.mkstemp()
  os.close(fd)
  st = os.stat(path)
  os.remove(path)
  assert not stat.S_ISDIR(st.st_mode)
  assert stat.S_IMODE(st.st_mode) == 0o600
  assert st.st_size == 0
  # System time and mtime may have different precision so give 10 sec leeway.
  assert st.st_mtime + 10 > t


def TestStatDir():
  path = tempfile.mkdtemp()
  mode = os.stat(path).st_mode
  os.rmdir(path)
  assert stat.S_ISDIR(mode)
  assert stat.S_IMODE(mode) == 0o700


def TestStatNoExist():
  path = tempfile.mkdtemp()
  try:
    os.stat(path + '/nonexistent')
  except OSError:
    pass
  else:
    raise AssertionError
  finally:
    os.rmdir(path)


if __name__ == '__main__':
  weetest.RunTests()

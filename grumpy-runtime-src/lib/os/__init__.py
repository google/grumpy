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
from '__go__/io/ioutil' import ReadDir
from '__go__/os' import (Chdir, Chmod, Environ, Getpid as getpid, Getwd, Pipe,
    ProcAttr, Remove, StartProcess, Stat, Stdout, Stdin,
    Stderr, Mkdir)
from '__go__/path/filepath' import Separator
from '__go__/grumpy' import (NewFileFromFD, StartThread, ToNative)
from '__go__/reflect' import MakeSlice
from '__go__/runtime' import GOOS
from '__go__/syscall' import (Close, SYS_FCNTL, Syscall, F_GETFD, Wait4,
    WaitStatus, WNOHANG)
from '__go__/sync' import WaitGroup
from '__go__/time' import Second
import _syscall
from os import path
import stat as stat_module
import sys


sep = chr(Separator)
error = OSError  # pylint: disable=invalid-name
curdir = '.'
name = 'posix'


environ = {}
for var in Environ():
  k, v = var.split('=', 1)
  environ[k] = v


def mkdir(path, mode=0o777):
  err = Mkdir(path, mode)
  if err:
    raise OSError(err.Error())


def chdir(path):
  err = Chdir(path)
  if err:
    raise OSError(err.Error())


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
  return NewFileFromFD(fd, None)


def listdir(p):
  files, err = ReadDir(p)
  if err:
    raise OSError(err.Error())
  return [x.Name() for x in files]


def getcwd():
  dir, err = Getwd()
  if err:
    raise OSError(err.Error())
  return dir


class _Popen(object):

  def __init__(self, command, mode):
    self.mode = mode
    self.result = None
    self.r, self.w, err = Pipe()
    if err:
      raise OSError(err.Error())
    attr = ProcAttr.new()
    # Create a slice using a reflect.Type returned by ToNative.
    # TODO: There should be a cleaner way to create slices in Python.
    files_type = ToNative(__frame__(), attr.Files).Type()
    files = MakeSlice(files_type, 3, 3).Interface()
    if self.mode == 'r':
      fd = self.r.Fd()
      files[0], files[1], files[2] = Stdin, self.w, Stderr
    elif self.mode == 'w':
      fd = self.w.Fd()
      files[0], files[1], files[2] = self.r, Stdout, Stderr
    else:
      raise ValueError('invalid popen mode: %r', self.mode)
    attr.Files = files
    # TODO: There should be a cleaner way to create slices in Python.
    args_type = ToNative(__frame__(), StartProcess).Type().In(1)
    args = MakeSlice(args_type, 3, 3).Interface()
    shell = environ['SHELL']
    args[0] = shell
    args[1] = '-c'
    args[2] = command
    self.proc, err = StartProcess(shell, args, attr)
    if err:
      raise OSError(err.Error())
    self.wg = WaitGroup.new()
    self.wg.Add(1)
    StartThread(self._thread_func)
    self.file = NewFileFromFD(fd, self.close)

  def _thread_func(self):
    self.result = self.proc.Wait()
    if self.mode == 'r':
      self.w.Close()
    self.wg.Done()

  def close(self, _):
    if self.mode == 'w':
      self.w.Close()
    self.wg.Wait()
    state, err = self.result
    if err:
      raise OSError(err.Error())
    return state.Sys() 


def popen(command, mode='r'):
  return _Popen(command, mode).file


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

  def st_mtime(self):
    return float(self._info.ModTime().UnixNano()) / Second
  # TODO: Make this a decorator once they're implemented.
  st_mtime = property(st_mtime)

  def st_size(self):
    return self._info.Size()
  # TODO: Make this a decorator once they're implemented.
  st_size = property(st_size)


def stat(filepath):
  info, err = Stat(filepath)
  if err:
    raise OSError(err.Error())
  return StatResult(info)


unlink = remove


def waitpid(pid, options):
  status = WaitStatus.new()
  _syscall.invoke(Wait4, pid, status, options, None)
  return pid, _encode_wait_result(status)


def _encode_wait_result(status):
  return status.Signal() | (status.ExitStatus() << 8)

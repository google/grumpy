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

"""System-specific parameters and functions."""

from '__go__/os' import Args
from '__go__/grumpy' import SysModules, MaxInt, Stdin as stdin, Stdout as stdout, Stderr as stderr  # pylint: disable=g-multiple-import
from '__go__/runtime' import (GOOS as platform, Version)
from '__go__/unicode' import MaxRune

argv = []
for arg in Args:
  argv.append(arg)

goversion = Version()
maxint = MaxInt
maxsize = maxint
maxunicode = MaxRune
modules = SysModules
py3kwarning = False
warnoptions = []
# TODO: Support actual byteorder
byteorder = 'little'
version = '2.7.13'

class _Flags(object):
  """Container class for sys.flags."""
  debug = 0
  py3k_warning = 0
  division_warning = 0
  division_new = 0
  inspect = 0
  interactive = 0
  optimize = 0
  dont_write_bytecode = 0
  no_user_site = 0
  no_site = 0
  ignore_environment = 0
  tabcheck = 0
  verbose = 0
  unicode = 0
  bytes_warning = 0
  hash_randomization = 0


flags = _Flags()


def exc_clear():
  __frame__().__exc_clear__()


def exc_info():
  e, tb = __frame__().__exc_info__()  # pylint: disable=undefined-variable
  t = None
  if e:
    t = type(e)
  return t, e, tb


def exit(code=None):  # pylint: disable=redefined-builtin
  raise SystemExit(code)


def _getframe(depth=0):
  f = __frame__()
  while depth > 0 and f is not None:
    f = f.f_back
    depth -= 1
  if f is None:
    raise ValueError('call stack is not deep enough')
  return f

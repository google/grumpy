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

from '__go__/syscall' import (
    FD_SETSIZE as _FD_SETSIZE,
    Select as _Select,
    FdSet as _FdSet,
    Timeval as _Timeval
)
import _syscall
import math


class error(Exception):
  pass


def select(rlist, wlist, xlist, timeout=None):
  rlist_norm = _normalize_fd_list(rlist)
  wlist_norm = _normalize_fd_list(wlist)
  xlist_norm = _normalize_fd_list(xlist)
  all_fds = rlist_norm + wlist_norm + xlist_norm
  if not all_fds:
    nfd = 0
  else:
    nfd = max(all_fds) + 1

  rfds = _make_fdset(rlist_norm)
  wfds = _make_fdset(wlist_norm)
  xfds = _make_fdset(xlist_norm)

  if timeout is None:
    timeval = None
  else:
    timeval = _Timeval.new()
    frac, integer = math.modf(timeout)
    timeval.Sec = int(integer)
    timeval.Usec = int(frac * 1000000.0)
  _syscall.invoke(_Select, nfd, rfds, wfds, xfds, timeval)
  return ([rlist[i] for i, fd in enumerate(rlist_norm) if _fdset_isset(fd, rfds)],
          [wlist[i] for i, fd in enumerate(wlist_norm) if _fdset_isset(fd, wfds)],
          [xlist[i] for i, fd in enumerate(xlist_norm) if _fdset_isset(fd, xfds)])


def _fdset_set(fd, fds):
  idx = fd / (_FD_SETSIZE / len(fds.Bits)) % len(fds.Bits)
  pos = fd % (_FD_SETSIZE / len(fds.Bits))
  fds.Bits[idx] |= 1 << pos


def _fdset_isset(fd, fds):
  idx = fd / (_FD_SETSIZE / len(fds.Bits)) % len(fds.Bits)
  pos = fd % (_FD_SETSIZE / len(fds.Bits))
  return bool(fds.Bits[idx] & (1 << pos))


def _make_fdset(fd_list):
  fds = _FdSet.new()
  for fd in fd_list:
    _fdset_set(fd, fds)
  return fds


def _normalize_fd_list(fds):
  result = []
  # Python permits mutating the select fds list during fileno calls so we can't
  # just use simple iteration over the list. See test_select_mutated in
  # test_select.py
  i = 0
  while i < len(fds):
    fd = fds[i]
    if hasattr(fd, 'fileno'):
      fd = fd.fileno()
    result.append(fd)
    i += 1
  return result

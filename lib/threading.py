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

"""Concurrent programming functionality."""

from __go__.grumpy import StartThread
from __go__.sync import NewCond, type_Mutex as Mutex


class Event(object):
  """Event is a way to signal conditions between threads."""

  def __init__(self):
    self._mutex = Mutex.new()
    self._cond = NewCond(self._mutex)
    self._is_set = False

  def set(self):
    self._mutex.Lock()
    try:
      self._is_set = True
    finally:
      self._mutex.Unlock()
    self._cond.Broadcast()

  # TODO: Support timeout param.
  def wait(self):
    self._mutex.Lock()
    try:
      while not self._is_set:
        self._cond.Wait()
    finally:
      self._mutex.Unlock()
    return True


class Thread(object):
  """Thread is an activity to be executed concurrently."""

  def __init__(self, target=None, args=()):
    self._target = target
    self._args = args
    self._event = Event()

  def run(self):
    self._target(*self._args)

  def start(self):
    StartThread(self._run)

  # TODO: Support timeout param.
  def join(self):
    self._event.wait()

  def _run(self):
    try:
      self.run()
    finally:
      self._event.set()

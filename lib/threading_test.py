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

import threading
import time

import weetest


def TestEvent():
  e = threading.Event()
  target_result = []
  x = 'not ready'
  def Target():
    e.wait()
    target_result.append(x)
  t = threading.Thread(target=Target)
  t.start()
  # Sleeping gives us some confidence that t had the opportunity to wait on e
  # and that if e is broken (e.g. wait() returned immediately) then the test
  # will fail below.
  time.sleep(0.1)
  x = 'ready'
  e.set()
  t.join()
  assert target_result == ['ready']
  target_result[:] = []
  t = threading.Thread(target=Target)
  t.start()
  t.join()
  assert target_result == ['ready']
  target_result[:] = []
  e.clear()
  t = threading.Thread(target=Target)
  t.start()
  time.sleep(0.1)
  assert not target_result
  e.set()
  t.join()
  assert target_result == ['ready']


def TestThread():
  ran = []
  def Target():
    ran.append(True)
  t = threading.Thread(target=Target)
  t.start()
  t.join()
  assert ran


def TestThreadArgs():
  target_args = []
  def Target(*args):
    target_args.append(args)
  t = threading.Thread(target=Target, args=('foo', 42))
  t.start()
  t.join()
  assert target_args == [('foo', 42)]


if __name__ == '__main__':
  weetest.RunTests()

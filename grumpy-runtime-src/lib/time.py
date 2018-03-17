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

"""Time access and conversions."""

from '__go__/time' import Local, Now, Second, Sleep, Unix, Date, UTC # pylint: disable=g-multiple-import


_strftime_directive_map = {
    '%': '%',
    'a': 'Mon',
    'A': 'Monday',
    'b': 'Jan',
    'B': 'January',
    'c': NotImplemented,
    'd': '02',
    'H': '15',
    'I': '03',
    'j': NotImplemented,
    'L': '.000',
    'm': '01',
    'M': '04',
    'p': 'PM',
    'S': '05',
    'U': NotImplemented,
    'W': NotImplemented,
    'w': NotImplemented,
    'X': NotImplemented,
    'x': NotImplemented,
    'y': '06',
    'Y': '2006',
    'Z': 'MST',
    'z': '-0700',
}


class struct_time(tuple):  #pylint: disable=invalid-name,missing-docstring

  def __init__(self, args):
    super(struct_time, self).__init__(tuple, args)
    self.tm_year = self[0]
    self.tm_mon = self[1]
    self.tm_mday = self[2]
    self.tm_hour = self[3]
    self.tm_min = self[4]
    self.tm_sec = self[5]
    self.tm_wday = self[6]
    self.tm_yday = self[7]
    self.tm_isdst = self[8]

  def __repr__(self):
    return ("time.struct_time(tm_year=%s, tm_mon=%s, tm_mday=%s, "
            "tm_hour=%s, tm_min=%s, tm_sec=%s, tm_wday=%s, "
            "tm_yday=%s, tm_isdst=%s)") % self

  def __str__(self):
    return repr(self)


def gmtime(seconds=None):
  t = (Unix(seconds, 0) if seconds else Now()).UTC()
  return struct_time((t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(),
                      t.Second(), (t.Weekday() + 6) % 7, t.YearDay(), 0))


def localtime(seconds=None):
  t = (Unix(seconds, 0) if seconds else Now()).Local()
  return struct_time((t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(),
                      t.Second(), (t.Weekday() + 6) % 7, t.YearDay(), 0))


def mktime(t):
  return float(Date(t[0], t[1], t[2], t[3], t[4], t[5], 0, Local).Unix())


def sleep(secs):
  Sleep(secs * Second)


def time():
  return float(Now().UnixNano()) / Second


def strftime(format, tt=None):  # pylint: disable=missing-docstring,redefined-builtin
  t = Unix(int(mktime(tt)), 0) if tt else Now()
  ret = []
  prev, n = 0, format.find('%', 0, -1)
  while n != -1:
    ret.append(format[prev:n])
    next_ch = format[n + 1]
    c = _strftime_directive_map.get(next_ch)
    if c is NotImplemented:
      raise NotImplementedError('Code: %' + next_ch + ' not yet supported')
    if c:
      ret.append(t.Format(c))
    else:
      ret.append(format[n:n+2])
    n += 2
    prev, n = n, format.find('%', n, -1)
  ret.append(format[prev:])
  return ''.join(ret)


# TODO: Calculate real value for daylight saving.
daylight = 0

# TODO: Use local DST instead of ''.
tzname = (Now().Zone()[0], '')


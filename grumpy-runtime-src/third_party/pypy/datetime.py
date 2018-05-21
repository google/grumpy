"""Concrete date/time and related types -- prototype implemented in Python.

See http://www.zope.org/Members/fdrake/DateTimeWiki/FrontPage

See also http://dir.yahoo.com/Reference/calendars/

For a primer on DST, including many current DST rules, see
http://webexhibits.org/daylightsaving/

For more about DST than you ever wanted to know, see
ftp://elsie.nci.nih.gov/pub/

Sources for time zone and DST data: http://www.twinsun.com/tz/tz-link.htm

This was originally copied from the sandbox of the CPython CVS repository.
Thanks to Tim Peters for suggesting using it.
"""

# from __future__ import division
import time as _time
import math as _math
# import struct as _struct
import _struct

def divmod(x, y):
  x, y = int(x), int(y)
  return x / y, x % y

_SENTINEL = object()

def _cmp(x, y):
    return 0 if x == y else 1 if x > y else -1

def _round(x):
    return int(_math.floor(x + 0.5) if x >= 0.0 else _math.ceil(x - 0.5))

MINYEAR = 1
MAXYEAR = 9999
_MINYEARFMT = 1900

_MAX_DELTA_DAYS = 999999999

# Utility functions, adapted from Python's Demo/classes/Dates.py, which
# also assumes the current Gregorian calendar indefinitely extended in
# both directions.  Difference:  Dates.py calls January 1 of year 0 day
# number 1.  The code here calls January 1 of year 1 day number 1.  This is
# to match the definition of the "proleptic Gregorian" calendar in Dershowitz
# and Reingold's "Calendrical Calculations", where it's the base calendar
# for all computations.  See the book for algorithms for converting between
# proleptic Gregorian ordinals and many other calendar systems.

_DAYS_IN_MONTH = [-1, 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31]

_DAYS_BEFORE_MONTH = [-1]
dbm = 0
for dim in _DAYS_IN_MONTH[1:]:
    _DAYS_BEFORE_MONTH.append(dbm)
    dbm += dim
del dbm, dim

def _is_leap(year):
    "year -> 1 if leap year, else 0."
    return year % 4 == 0 and (year % 100 != 0 or year % 400 == 0)

def _days_before_year(year):
    "year -> number of days before January 1st of year."
    y = year - 1
    return y*365 + y//4 - y//100 + y//400

def _days_in_month(year, month):
    "year, month -> number of days in that month in that year."
    assert 1 <= month <= 12, month
    if month == 2 and _is_leap(year):
        return 29
    return _DAYS_IN_MONTH[month]

def _days_before_month(year, month):
    "year, month -> number of days in year preceding first day of month."
    assert 1 <= month <= 12, 'month must be in 1..12'
    return _DAYS_BEFORE_MONTH[month] + (month > 2 and _is_leap(year))

def _ymd2ord(year, month, day):
    "year, month, day -> ordinal, considering 01-Jan-0001 as day 1."
    assert 1 <= month <= 12, 'month must be in 1..12'
    dim = _days_in_month(year, month)
    assert 1 <= day <= dim, ('day must be in 1..%d' % dim)
    return (_days_before_year(year) +
            _days_before_month(year, month) +
            day)

_DI400Y = _days_before_year(401)    # number of days in 400 years
_DI100Y = _days_before_year(101)    #    "    "   "   " 100   "
_DI4Y   = _days_before_year(5)      #    "    "   "   "   4   "

# A 4-year cycle has an extra leap day over what we'd get from pasting
# together 4 single years.
assert _DI4Y == 4 * 365 + 1

# Similarly, a 400-year cycle has an extra leap day over what we'd get from
# pasting together 4 100-year cycles.
assert _DI400Y == 4 * _DI100Y + 1

# OTOH, a 100-year cycle has one fewer leap day than we'd get from
# pasting together 25 4-year cycles.
assert _DI100Y == 25 * _DI4Y - 1

_US_PER_US = 1
_US_PER_MS = 1000
_US_PER_SECOND = 1000000
_US_PER_MINUTE = 60000000
_SECONDS_PER_DAY = 24 * 3600
_US_PER_HOUR = 3600000000
_US_PER_DAY = 86400000000
_US_PER_WEEK = 604800000000

def _ord2ymd(n):
    "ordinal -> (year, month, day), considering 01-Jan-0001 as day 1."

    # n is a 1-based index, starting at 1-Jan-1.  The pattern of leap years
    # repeats exactly every 400 years.  The basic strategy is to find the
    # closest 400-year boundary at or before n, then work with the offset
    # from that boundary to n.  Life is much clearer if we subtract 1 from
    # n first -- then the values of n at 400-year boundaries are exactly
    # those divisible by _DI400Y:
    #
    #     D  M   Y            n              n-1
    #     -- --- ----        ----------     ----------------
    #     31 Dec -400        -_DI400Y       -_DI400Y -1
    #      1 Jan -399         -_DI400Y +1   -_DI400Y      400-year boundary
    #     ...
    #     30 Dec  000        -1             -2
    #     31 Dec  000         0             -1
    #      1 Jan  001         1              0            400-year boundary
    #      2 Jan  001         2              1
    #      3 Jan  001         3              2
    #     ...
    #     31 Dec  400         _DI400Y        _DI400Y -1
    #      1 Jan  401         _DI400Y +1     _DI400Y      400-year boundary
    n -= 1
    n400, n = divmod(n, _DI400Y)
    year = n400 * 400 + 1   # ..., -399, 1, 401, ...

    # Now n is the (non-negative) offset, in days, from January 1 of year, to
    # the desired date.  Now compute how many 100-year cycles precede n.
    # Note that it's possible for n100 to equal 4!  In that case 4 full
    # 100-year cycles precede the desired day, which implies the desired
    # day is December 31 at the end of a 400-year cycle.
    n100, n = divmod(n, _DI100Y)

    # Now compute how many 4-year cycles precede it.
    n4, n = divmod(n, _DI4Y)

    # And now how many single years.  Again n1 can be 4, and again meaning
    # that the desired day is December 31 at the end of the 4-year cycle.
    n1, n = divmod(n, 365)

    year += n100 * 100 + n4 * 4 + n1
    if n1 == 4 or n100 == 4:
        assert n == 0
        return year-1, 12, 31

    # Now the year is correct, and n is the offset from January 1.  We find
    # the month via an estimate that's either exact or one too large.
    leapyear = n1 == 3 and (n4 != 24 or n100 == 3)
    assert leapyear == _is_leap(year)
    month = (n + 50) >> 5
    preceding = _DAYS_BEFORE_MONTH[month] + (month > 2 and leapyear)
    if preceding > n:  # estimate is too large
        month -= 1
        preceding -= _DAYS_IN_MONTH[month] + (month == 2 and leapyear)
    n -= preceding
    assert 0 <= n < _days_in_month(year, month)

    # Now the year and month are correct, and n is the offset from the
    # start of that month:  we're done!
    return year, month, n+1

# Month and day names.  For localized versions, see the calendar module.
_MONTHNAMES = [None, "Jan", "Feb", "Mar", "Apr", "May", "Jun",
                     "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"]
_DAYNAMES = [None, "Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"]


def _build_struct_time(y, m, d, hh, mm, ss, dstflag):
    wday = (_ymd2ord(y, m, d) + 6) % 7
    dnum = _days_before_month(y, m) + d
    return _time.struct_time((y, m, d, hh, mm, ss, wday, dnum, dstflag))

def _format_time(hh, mm, ss, us):
    # Skip trailing microseconds when us==0.
    result = "%02d:%02d:%02d" % (hh, mm, ss)
    if us:
        result += ".%06d" % us
    return result

# Correctly substitute for %z and %Z escapes in strftime formats.
# def _wrap_strftime(object, format, timetuple):
#     year = timetuple[0]
#     if year < _MINYEARFMT:
#         raise ValueError("year=%d is before %d; the datetime strftime() "
#                          "methods require year >= %d" %
#                          (year, _MINYEARFMT, _MINYEARFMT))
#     # Don't call utcoffset() or tzname() unless actually needed.
#     freplace = None  # the string to use for %f
#     zreplace = None  # the string to use for %z
#     Zreplace = None  # the string to use for %Z

#     # Scan format for %z and %Z escapes, replacing as needed.
#     newformat = []
#     push = newformat.append
#     i, n = 0, len(format)
#     while i < n:
#         ch = format[i]
#         i += 1
#         if ch == '%':
#             if i < n:
#                 ch = format[i]
#                 i += 1
#                 if ch == 'f':
#                     if freplace is None:
#                         freplace = '%06d' % getattr(object,
#                                                     'microsecond', 0)
#                     newformat.append(freplace)
#                 elif ch == 'z':
#                     if zreplace is None:
#                         zreplace = ""
#                         if hasattr(object, "_utcoffset"):
#                             offset = object._utcoffset()
#                             if offset is not None:
#                                 sign = '+'
#                                 if offset < 0:
#                                     offset = -offset
#                                     sign = '-'
#                                 h, m = divmod(offset, 60)
#                                 zreplace = '%c%02d%02d' % (sign, h, m)
#                     assert '%' not in zreplace
#                     newformat.append(zreplace)
#                 elif ch == 'Z':
#                     if Zreplace is None:
#                         Zreplace = ""
#                         if hasattr(object, "tzname"):
#                             s = object.tzname()
#                             if s is not None:
#                                 # strftime is going to have at this: escape %
#                                 Zreplace = s.replace('%', '%%')
#                     newformat.append(Zreplace)
#                 else:
#                     push('%')
#                     push(ch)
#             else:
#                 push('%')
#         else:
#             push(ch)
#     newformat = "".join(newformat)
#     return _time.strftime(newformat, timetuple)

# Just raise TypeError if the arg isn't None or a string.
def _check_tzname(name):
    if name is not None and not isinstance(name, str):
        raise TypeError("tzinfo.tzname() must return None or string, "
                        "not '%s'" % type(name))

# name is the offset-producing method, "utcoffset" or "dst".
# offset is what it returned.
# If offset isn't None or timedelta, raises TypeError.
# If offset is None, returns None.
# Else offset is checked for being in range, and a whole # of minutes.
# If it is, its integer value is returned.  Else ValueError is raised.
def _check_utc_offset(name, offset):
    assert name in ("utcoffset", "dst")
    if offset is None:
        return
    if not isinstance(offset, timedelta):
        raise TypeError("tzinfo.%s() must return None "
                        "or timedelta, not '%s'" % (name, type(offset)))
    days = offset.days
    if days < -1 or days > 0:
        offset = 1440  # trigger out-of-range
    else:
        seconds = days * 86400 + offset.seconds
        minutes, seconds = divmod(seconds, 60)
        if seconds or offset.microseconds:
            raise ValueError("tzinfo.%s() must return a whole number "
                             "of minutes" % name)
        offset = minutes
    if not -1440 < offset < 1440:
        raise ValueError("%s()=%d, must be in -1439..1439" % (name, offset))
    return offset

def _check_int_field(value):
    if isinstance(value, int):
        return int(value)
    if not isinstance(value, float):
        try:
            value = value.__int__()
        except AttributeError:
            pass
        else:
            if isinstance(value, int):
                return int(value)
            elif isinstance(value, long):
                return int(long(value))
            raise TypeError('__int__ method should return an integer')
        raise TypeError('an integer is required')
    raise TypeError('integer argument expected, got float')

def _check_date_fields(year, month, day):
    year = _check_int_field(year)
    month = _check_int_field(month)
    day = _check_int_field(day)
    if not MINYEAR <= year <= MAXYEAR:
        raise ValueError('year must be in %d..%d' % (MINYEAR, MAXYEAR), year)
    if not 1 <= month <= 12:
        raise ValueError('month must be in 1..12', month)
    dim = _days_in_month(year, month)
    if not 1 <= day <= dim:
        raise ValueError('day must be in 1..%d' % dim, day)
    return year, month, day

def _check_time_fields(hour, minute, second, microsecond):
    hour = _check_int_field(hour)
    minute = _check_int_field(minute)
    second = _check_int_field(second)
    microsecond = _check_int_field(microsecond)
    if not 0 <= hour <= 23:
        raise ValueError('hour must be in 0..23', hour)
    if not 0 <= minute <= 59:
        raise ValueError('minute must be in 0..59', minute)
    if not 0 <= second <= 59:
        raise ValueError('second must be in 0..59', second)
    if not 0 <= microsecond <= 999999:
        raise ValueError('microsecond must be in 0..999999', microsecond)
    return hour, minute, second, microsecond

def _check_tzinfo_arg(tz):
    if tz is not None and not isinstance(tz, tzinfo):
        raise TypeError("tzinfo argument must be None or of a tzinfo subclass")


# Notes on comparison:  In general, datetime module comparison operators raise
# TypeError when they don't know how to do a comparison themself.  If they
# returned NotImplemented instead, comparison could (silently) fall back to
# the default compare-objects-by-comparing-their-memory-addresses strategy,
# and that's not helpful.  There are two exceptions:
#
# 1. For date and datetime, if the other object has a "timetuple" attr,
#    NotImplemented is returned.  This is a hook to allow other kinds of
#    datetime-like objects a chance to intercept the comparison.
#
# 2. Else __eq__ and __ne__ return False and True, respectively.  This is
#    so opertaions like
#
#        x == y
#        x != y
#        x in sequence
#        x not in sequence
#        dict[x] = y
#
#    don't raise annoying TypeErrors just because a datetime object
#    is part of a heterogeneous collection.  If there's no known way to
#    compare X to a datetime, saying they're not equal is reasonable.

def _cmperror(x, y):
    raise TypeError("can't compare '%s' to '%s'" % (
                    type(x).__name__, type(y).__name__))

def _normalize_pair(hi, lo, factor):
    if not 0 <= lo <= factor-1:
        inc, lo = divmod(lo, factor)
        hi += inc
    return hi, lo

def _normalize_datetime(y, m, d, hh, mm, ss, us, ignore_overflow=False):
    # Normalize all the inputs, and store the normalized values.
    ss, us = _normalize_pair(ss, us, 1000000)
    mm, ss = _normalize_pair(mm, ss, 60)
    hh, mm = _normalize_pair(hh, mm, 60)
    d, hh = _normalize_pair(d, hh, 24)
    y, m, d = _normalize_date(y, m, d, ignore_overflow)
    return y, m, d, hh, mm, ss, us

def _normalize_date(year, month, day, ignore_overflow=False):
    # That was easy.  Now it gets muddy:  the proper range for day
    # can't be determined without knowing the correct month and year,
    # but if day is, e.g., plus or minus a million, the current month
    # and year values make no sense (and may also be out of bounds
    # themselves).
    # Saying 12 months == 1 year should be non-controversial.
    if not 1 <= month <= 12:
        year, month = _normalize_pair(year, month-1, 12)
        month += 1
        assert 1 <= month <= 12

    # Now only day can be out of bounds (year may also be out of bounds
    # for a datetime object, but we don't care about that here).
    # If day is out of bounds, what to do is arguable, but at least the
    # method here is principled and explainable.
    dim = _days_in_month(year, month)
    if not 1 <= day <= dim:
        # Move day-1 days from the first of the month.  First try to
        # get off cheap if we're only one day out of range (adjustments
        # for timezone alone can't be worse than that).
        if day == 0:    # move back a day
            month -= 1
            if month > 0:
                day = _days_in_month(year, month)
            else:
                year, month, day = year-1, 12, 31
        elif day == dim + 1:    # move forward a day
            month += 1
            day = 1
            if month > 12:
                month = 1
                year += 1
        else:
            ordinal = _ymd2ord(year, month, 1) + (day - 1)
            year, month, day = _ord2ymd(ordinal)

    if not ignore_overflow and not MINYEAR <= year <= MAXYEAR:
        raise OverflowError("date value out of range")
    return year, month, day

def _accum(tag, sofar, num, factor, leftover):
    if isinstance(num, (int, long)):
        prod = num * factor
        rsum = sofar + prod
        return rsum, leftover
    if isinstance(num, float):
        fracpart, intpart = _math.modf(num)
        prod = int(intpart) * factor
        rsum = sofar + prod
        if fracpart == 0.0:
            return rsum, leftover
        assert isinstance(factor, (int, long))
        fracpart, intpart = _math.modf(factor * fracpart)
        rsum += int(intpart)
        return rsum, leftover + fracpart
    raise TypeError("unsupported type for timedelta %s component: %s" %
                    (tag, type(num)))

class timedelta(object):
    """Represent the difference between two datetime objects.

    Supported operators:

    - add, subtract timedelta
    - unary plus, minus, abs
    - compare to timedelta
    - multiply, divide by int/long

    In addition, datetime supports subtraction of two datetime objects
    returning a timedelta, and addition or subtraction of a datetime
    and a timedelta giving a datetime.

    Representation: (days, seconds, microseconds).  Why?  Because I
    felt like it.
    """
    __slots__ = '_days', '_seconds', '_microseconds', '_hashcode'

    def __new__(cls, days=_SENTINEL, seconds=_SENTINEL, microseconds=_SENTINEL,
                milliseconds=_SENTINEL, minutes=_SENTINEL, hours=_SENTINEL, weeks=_SENTINEL):
        x = 0
        leftover = 0.0
        if microseconds is not _SENTINEL:
            x, leftover = _accum("microseconds", x, microseconds, _US_PER_US, leftover)
        if milliseconds is not _SENTINEL:
            x, leftover = _accum("milliseconds", x, milliseconds, _US_PER_MS, leftover)
        if seconds is not _SENTINEL:
            x, leftover = _accum("seconds", x, seconds, _US_PER_SECOND, leftover)
        if minutes is not _SENTINEL:
            x, leftover = _accum("minutes", x, minutes, _US_PER_MINUTE, leftover)
        if hours is not _SENTINEL:
            x, leftover = _accum("hours", x, hours, _US_PER_HOUR, leftover)
        if days is not _SENTINEL:
            x, leftover = _accum("days", x, days, _US_PER_DAY, leftover)
        if weeks is not _SENTINEL:
            x, leftover = _accum("weeks", x, weeks, _US_PER_WEEK, leftover)
        if leftover != 0.0:
            x += _round(leftover)
        return cls._from_microseconds(x)

    @classmethod
    def _from_microseconds(cls, us):
        s, us = divmod(us, _US_PER_SECOND)
        d, s = divmod(s, _SECONDS_PER_DAY)
        return cls._create(d, s, us, False)

    @classmethod
    def _create(cls, d, s, us, normalize):
        if normalize:
            s, us = _normalize_pair(s, us, 1000000)
            d, s = _normalize_pair(d, s, 24*3600)

        if not -_MAX_DELTA_DAYS <= d <= _MAX_DELTA_DAYS:
            raise OverflowError("days=%d; must have magnitude <= %d" % (d, _MAX_DELTA_DAYS))

        self = object.__new__(cls)
        self._days = d
        self._seconds = s
        self._microseconds = us
        self._hashcode = -1
        return self

    def _to_microseconds(self):
        return ((self._days * _SECONDS_PER_DAY + self._seconds) * _US_PER_SECOND +
                self._microseconds)

    def __repr__(self):
        module = "datetime." if self.__class__ is timedelta else ""
        if self._microseconds:
            return "%s(%d, %d, %d)" % (module + self.__class__.__name__,
                                       self._days,
                                       self._seconds,
                                       self._microseconds)
        if self._seconds:
            return "%s(%d, %d)" % (module + self.__class__.__name__,
                                   self._days,
                                   self._seconds)
        return "%s(%d)" % (module + self.__class__.__name__, self._days)

    def __str__(self):
        mm, ss = divmod(self._seconds, 60)
        hh, mm = divmod(mm, 60)
        s = "%d:%02d:%02d" % (hh, mm, ss)
        if self._days:
            def plural(n):
                return n, abs(n) != 1 and "s" or ""
            s = ("%d day%s, " % plural(self._days)) + s
        if self._microseconds:
            s = s + ".%06d" % self._microseconds
        return s

    def total_seconds(self):
        """Total seconds in the duration."""
        # return self._to_microseconds() / 10**6
        return float(self._to_microseconds()) / float(10**6)

    # Read-only field accessors
    @property
    def days(self):
        """days"""
        return self._days

    @property
    def seconds(self):
        """seconds"""
        return self._seconds

    @property
    def microseconds(self):
        """microseconds"""
        return self._microseconds

    def __add__(self, other):
        if isinstance(other, timedelta):
            # for CPython compatibility, we cannot use
            # our __class__ here, but need a real timedelta
            return timedelta._create(self._days + other._days,
                                     self._seconds + other._seconds,
                                     self._microseconds + other._microseconds,
                                     True)
        return NotImplemented

    def __sub__(self, other):
        if isinstance(other, timedelta):
            # for CPython compatibility, we cannot use
            # our __class__ here, but need a real timedelta
            return timedelta._create(self._days - other._days,
                                     self._seconds - other._seconds,
                                     self._microseconds - other._microseconds,
                                     True)
        return NotImplemented

    def __neg__(self):
        # for CPython compatibility, we cannot use
        # our __class__ here, but need a real timedelta
        return timedelta._create(-self._days,
                                 -self._seconds,
                                 -self._microseconds,
                                 True)

    def __pos__(self):
        # for CPython compatibility, we cannot use
        # our __class__ here, but need a real timedelta
        return timedelta._create(self._days,
                                 self._seconds,
                                 self._microseconds,
                                 False)

    def __abs__(self):
        if self._days < 0:
            return -self
        else:
            return self

    def __mul__(self, other):
        if not isinstance(other, (int, long)):
            return NotImplemented
        usec = self._to_microseconds()
        return timedelta._from_microseconds(usec * other)

    __rmul__ = __mul__

    def __div__(self, other):
        if not isinstance(other, (int, long)):
            return NotImplemented
        usec = self._to_microseconds()
        # return timedelta._from_microseconds(usec // other)
        return timedelta._from_microseconds(int(usec) / int(other))

    __floordiv__ = __div__

    # Comparisons of timedelta objects with other.

    def __eq__(self, other):
        if isinstance(other, timedelta):
            return self._cmp(other) == 0
        else:
            return False

    def __ne__(self, other):
        if isinstance(other, timedelta):
            return self._cmp(other) != 0
        else:
            return True

    def __le__(self, other):
        if isinstance(other, timedelta):
            return self._cmp(other) <= 0
        else:
            _cmperror(self, other)

    def __lt__(self, other):
        if isinstance(other, timedelta):
            return self._cmp(other) < 0
        else:
            _cmperror(self, other)

    def __ge__(self, other):
        if isinstance(other, timedelta):
            return self._cmp(other) >= 0
        else:
            _cmperror(self, other)

    def __gt__(self, other):
        if isinstance(other, timedelta):
            return self._cmp(other) > 0
        else:
            _cmperror(self, other)

    def _cmp(self, other):
        assert isinstance(other, timedelta)
        return _cmp(self._getstate(), other._getstate())

    def __hash__(self):
        if self._hashcode == -1:
            self._hashcode = hash(self._getstate())
        return self._hashcode

    def __nonzero__(self):
        return (self._days != 0 or
                self._seconds != 0 or
                self._microseconds != 0)

    # Pickle support.

    def _getstate(self):
        return (self._days, self._seconds, self._microseconds)

    def __reduce__(self):
        return (self.__class__, self._getstate())

timedelta.min = timedelta(-_MAX_DELTA_DAYS)
timedelta.max = timedelta(_MAX_DELTA_DAYS, 24*3600-1, 1000000-1)
timedelta.resolution = timedelta(microseconds=1)

class date(object):
    """Concrete date type.

    Constructors:

    __new__()
    fromtimestamp()
    today()
    fromordinal()

    Operators:

    __repr__, __str__
    __cmp__, __hash__
    __add__, __radd__, __sub__ (add/radd only with timedelta arg)

    Methods:

    timetuple()
    toordinal()
    weekday()
    isoweekday(), isocalendar(), isoformat()
    ctime()
    strftime()

    Properties (readonly):
    year, month, day
    """
    __slots__ = '_year', '_month', '_day', '_hashcode'

    def __new__(cls, year, month=None, day=None):
        """Constructor.

        Arguments:

        year, month, day (required, base 1)
        """
        # if month is None and isinstance(year, bytes) and len(year) == 4 and \
        #         1 <= ord(year[2]) <= 12:
        #     # Pickle support
        #     self = object.__new__(cls)
        #     self.__setstate(year)
        #     self._hashcode = -1
        #     return self
        year, month, day = _check_date_fields(year, month, day)
        self = object.__new__(cls)
        self._year = year
        self._month = month
        self._day = day
        self._hashcode = -1
        return self

    # Additional constructors

    @classmethod
    def fromtimestamp(cls, t):
        "Construct a date from a POSIX timestamp (like time.time())."
        y, m, d, hh, mm, ss, weekday, jday, dst = _time.localtime(t)
        return cls(y, m, d)

    @classmethod
    def today(cls):
        "Construct a date from time.time()."
        t = _time.time()
        return cls.fromtimestamp(t)

    @classmethod
    def fromordinal(cls, n):
        """Contruct a date from a proleptic Gregorian ordinal.

        January 1 of year 1 is day 1.  Only the year, month and day are
        non-zero in the result.
        """
        y, m, d = _ord2ymd(n)
        return cls(y, m, d)

    # Conversions to string

    def __repr__(self):
        """Convert to formal string, for repr().

        >>> dt = datetime(2010, 1, 1)
        >>> repr(dt)
        'datetime.datetime(2010, 1, 1, 0, 0)'

        >>> dt = datetime(2010, 1, 1, tzinfo=timezone.utc)
        >>> repr(dt)
        'datetime.datetime(2010, 1, 1, 0, 0, tzinfo=datetime.timezone.utc)'
        """
        module = "datetime." if self.__class__ is date else ""
        return "%s(%d, %d, %d)" % (module + self.__class__.__name__,
                                   self._year,
                                   self._month,
                                   self._day)

    # XXX These shouldn't depend on time.localtime(), because that
    # clips the usable dates to [1970 .. 2038).  At least ctime() is
    # easily done without using strftime() -- that's better too because
    # strftime("%c", ...) is locale specific.

    def ctime(self):
        "Return ctime() style string."
        weekday = self.toordinal() % 7 or 7
        return "%s %s %2d 00:00:00 %04d" % (
            _DAYNAMES[weekday],
            _MONTHNAMES[self._month],
            self._day, self._year)

    # def strftime(self, format):
    #     "Format using strftime()."
    #     return _wrap_strftime(self, format, self.timetuple())

    def __format__(self, fmt):
        if not isinstance(fmt, (str, unicode)):
            raise ValueError("__format__ expects str or unicode, not %s" %
                             fmt.__class__.__name__)
        if len(fmt) != 0:
            return self.strftime(fmt)
        return str(self)

    def isoformat(self):
        """Return the date formatted according to ISO.

        This is 'YYYY-MM-DD'.

        References:
        - http://www.w3.org/TR/NOTE-datetime
        - http://www.cl.cam.ac.uk/~mgk25/iso-time.html
        """
        # return "%04d-%02d-%02d" % (self._year, self._month, self._day)
        return "%s-%s-%s" % (str(self._year).zfill(4), str(self._month).zfill(2), str(self._day).zfill(2))

    __str__ = isoformat

    # Read-only field accessors
    @property
    def year(self):
        """year (1-9999)"""
        return self._year

    @property
    def month(self):
        """month (1-12)"""
        return self._month

    @property
    def day(self):
        """day (1-31)"""
        return self._day

    # Standard conversions, __cmp__, __hash__ (and helpers)

    def timetuple(self):
        "Return local time tuple compatible with time.localtime()."
        return _build_struct_time(self._year, self._month, self._day,
                                  0, 0, 0, -1)

    def toordinal(self):
        """Return proleptic Gregorian ordinal for the year, month and day.

        January 1 of year 1 is day 1.  Only the year, month and day values
        contribute to the result.
        """
        return _ymd2ord(self._year, self._month, self._day)

    def replace(self, year=None, month=None, day=None):
        """Return a new date with new values for the specified fields."""
        if year is None:
            year = self._year
        if month is None:
            month = self._month
        if day is None:
            day = self._day
        return date.__new__(type(self), year, month, day)

    # Comparisons of date objects with other.

    def __eq__(self, other):
        if isinstance(other, date):
            return self._cmp(other) == 0
        elif hasattr(other, "timetuple"):
            return NotImplemented
        else:
            return False

    def __ne__(self, other):
        if isinstance(other, date):
            return self._cmp(other) != 0
        elif hasattr(other, "timetuple"):
            return NotImplemented
        else:
            return True

    def __le__(self, other):
        if isinstance(other, date):
            return self._cmp(other) <= 0
        elif hasattr(other, "timetuple"):
            return NotImplemented
        else:
            _cmperror(self, other)

    def __lt__(self, other):
        if isinstance(other, date):
            return self._cmp(other) < 0
        elif hasattr(other, "timetuple"):
            return NotImplemented
        else:
            _cmperror(self, other)

    def __ge__(self, other):
        if isinstance(other, date):
            return self._cmp(other) >= 0
        elif hasattr(other, "timetuple"):
            return NotImplemented
        else:
            _cmperror(self, other)

    def __gt__(self, other):
        if isinstance(other, date):
            return self._cmp(other) > 0
        elif hasattr(other, "timetuple"):
            return NotImplemented
        else:
            _cmperror(self, other)

    def _cmp(self, other):
        assert isinstance(other, date)
        y, m, d = self._year, self._month, self._day
        y2, m2, d2 = other._year, other._month, other._day
        return _cmp((y, m, d), (y2, m2, d2))

    def __hash__(self):
        "Hash."
        if self._hashcode == -1:
            self._hashcode = hash(self._getstate())
        return self._hashcode

    # Computations

    def _add_timedelta(self, other, factor):
        y, m, d = _normalize_date(
            self._year,
            self._month,
            self._day + other.days * factor)
        return date(y, m, d)

    def __add__(self, other):
        "Add a date to a timedelta."
        if isinstance(other, timedelta):
            return self._add_timedelta(other, 1)
        return NotImplemented

    __radd__ = __add__

    def __sub__(self, other):
        """Subtract two dates, or a date and a timedelta."""
        if isinstance(other, date):
            days1 = self.toordinal()
            days2 = other.toordinal()
            return timedelta._create(days1 - days2, 0, 0, False)
        if isinstance(other, timedelta):
            return self._add_timedelta(other, -1)
        return NotImplemented

    def weekday(self):
        "Return day of the week, where Monday == 0 ... Sunday == 6."
        return (self.toordinal() + 6) % 7

    # Day-of-the-week and week-of-the-year, according to ISO

    def isoweekday(self):
        "Return day of the week, where Monday == 1 ... Sunday == 7."
        # 1-Jan-0001 is a Monday
        return self.toordinal() % 7 or 7

    def isocalendar(self):
        """Return a 3-tuple containing ISO year, week number, and weekday.

        The first ISO week of the year is the (Mon-Sun) week
        containing the year's first Thursday; everything else derives
        from that.

        The first week is 1; Monday is 1 ... Sunday is 7.

        ISO calendar algorithm taken from
        http://www.phys.uu.nl/~vgent/calendar/isocalendar.htm
        """
        year = self._year
        week1monday = _isoweek1monday(year)
        today = _ymd2ord(self._year, self._month, self._day)
        # Internally, week and day have origin 0
        week, day = divmod(today - week1monday, 7)
        if week < 0:
            year -= 1
            week1monday = _isoweek1monday(year)
            week, day = divmod(today - week1monday, 7)
        elif week >= 52:
            if today >= _isoweek1monday(year+1):
                year += 1
                week = 0
        return year, week+1, day+1

    # Pickle support.

    def _getstate(self):
        yhi, ylo = divmod(self._year, 256)
        return (_struct.pack('4B', yhi, ylo, self._month, self._day),)

    def __setstate(self, string):
        yhi, ylo, self._month, self._day = (ord(string[0]), ord(string[1]),
                                            ord(string[2]), ord(string[3]))
        self._year = yhi * 256 + ylo

    def __reduce__(self):
        return (self.__class__, self._getstate())

_date_class = date  # so functions w/ args named "date" can get at the class

date.min = date(1, 1, 1)
date.max = date(9999, 12, 31)
date.resolution = timedelta(days=1)

class tzinfo(object):
    """Abstract base class for time zone info classes.

    Subclasses must override the name(), utcoffset() and dst() methods.
    """
    __slots__ = ()

    def tzname(self, dt):
        "datetime -> string name of time zone."
        raise NotImplementedError("tzinfo subclass must override tzname()")

    def utcoffset(self, dt):
        "datetime -> minutes east of UTC (negative for west of UTC)"
        raise NotImplementedError("tzinfo subclass must override utcoffset()")

    def dst(self, dt):
        """datetime -> DST offset in minutes east of UTC.

        Return 0 if DST not in effect.  utcoffset() must include the DST
        offset.
        """
        raise NotImplementedError("tzinfo subclass must override dst()")

    def fromutc(self, dt):
        "datetime in UTC -> datetime in local time."

        if not isinstance(dt, datetime):
            raise TypeError("fromutc() requires a datetime argument")
        if dt.tzinfo is not self:
            raise ValueError("dt.tzinfo is not self")

        dtoff = dt.utcoffset()
        if dtoff is None:
            raise ValueError("fromutc() requires a non-None utcoffset() "
                             "result")

        # See the long comment block at the end of this file for an
        # explanation of this algorithm.
        dtdst = dt.dst()
        if dtdst is None:
            raise ValueError("fromutc() requires a non-None dst() result")
        delta = dtoff - dtdst
        if delta:
            dt += delta
            dtdst = dt.dst()
            if dtdst is None:
                raise ValueError("fromutc(): dt.dst gave inconsistent "
                                 "results; cannot convert")
        if dtdst:
            return dt + dtdst
        else:
            return dt

    # Pickle support.

    def __reduce__(self):
        getinitargs = getattr(self, "__getinitargs__", None)
        if getinitargs:
            args = getinitargs()
        else:
            args = ()
        getstate = getattr(self, "__getstate__", None)
        if getstate:
            state = getstate()
        else:
            state = getattr(self, "__dict__", None) or None
        if state is None:
            return (self.__class__, args)
        else:
            return (self.__class__, args, state)

_tzinfo_class = tzinfo

class time(object):
    """Time with time zone.

    Constructors:

    __new__()

    Operators:

    __repr__, __str__
    __cmp__, __hash__

    Methods:

    strftime()
    isoformat()
    utcoffset()
    tzname()
    dst()

    Properties (readonly):
    hour, minute, second, microsecond, tzinfo
    """
    __slots__ = '_hour', '_minute', '_second', '_microsecond', '_tzinfo', '_hashcode'

    def __new__(cls, hour=0, minute=0, second=0, microsecond=0, tzinfo=None):
        """Constructor.

        Arguments:

        hour, minute (required)
        second, microsecond (default to zero)
        tzinfo (default to None)
        """
        # if isinstance(hour, bytes) and len(hour) == 6 and ord(hour[0]) < 24:
        #     # Pickle support
        #     self = object.__new__(cls)
        #     self.__setstate(hour, minute or None)
        #     self._hashcode = -1
        #     return self
        hour, minute, second, microsecond = _check_time_fields(
            hour, minute, second, microsecond)
        _check_tzinfo_arg(tzinfo)
        self = object.__new__(cls)
        self._hour = hour
        self._minute = minute
        self._second = second
        self._microsecond = microsecond
        self._tzinfo = tzinfo
        self._hashcode = -1
        return self

    # Read-only field accessors
    @property
    def hour(self):
        """hour (0-23)"""
        return self._hour

    @property
    def minute(self):
        """minute (0-59)"""
        return self._minute

    @property
    def second(self):
        """second (0-59)"""
        return self._second

    @property
    def microsecond(self):
        """microsecond (0-999999)"""
        return self._microsecond

    @property
    def tzinfo(self):
        """timezone info object"""
        return self._tzinfo

    # Standard conversions, __hash__ (and helpers)

    # Comparisons of time objects with other.

    def __eq__(self, other):
        if isinstance(other, time):
            return self._cmp(other) == 0
        else:
            return False

    def __ne__(self, other):
        if isinstance(other, time):
            return self._cmp(other) != 0
        else:
            return True

    def __le__(self, other):
        if isinstance(other, time):
            return self._cmp(other) <= 0
        else:
            _cmperror(self, other)

    def __lt__(self, other):
        if isinstance(other, time):
            return self._cmp(other) < 0
        else:
            _cmperror(self, other)

    def __ge__(self, other):
        if isinstance(other, time):
            return self._cmp(other) >= 0
        else:
            _cmperror(self, other)

    def __gt__(self, other):
        if isinstance(other, time):
            return self._cmp(other) > 0
        else:
            _cmperror(self, other)

    def _cmp(self, other):
        assert isinstance(other, time)
        mytz = self._tzinfo
        ottz = other._tzinfo
        myoff = otoff = None

        if mytz is ottz:
            base_compare = True
        else:
            myoff = self._utcoffset()
            otoff = other._utcoffset()
            base_compare = myoff == otoff

        if base_compare:
            return _cmp((self._hour, self._minute, self._second,
                         self._microsecond),
                        (other._hour, other._minute, other._second,
                         other._microsecond))
        if myoff is None or otoff is None:
            raise TypeError("can't compare offset-naive and offset-aware times")
        myhhmm = self._hour * 60 + self._minute - myoff
        othhmm = other._hour * 60 + other._minute - otoff
        return _cmp((myhhmm, self._second, self._microsecond),
                    (othhmm, other._second, other._microsecond))

    def __hash__(self):
        """Hash."""
        if self._hashcode == -1:
            tzoff = self._utcoffset()
            if not tzoff:  # zero or None
                self._hashcode = hash(self._getstate()[0])
            else:
                h, m = divmod(self.hour * 60 + self.minute - tzoff, 60)
                if 0 <= h < 24:
                    self._hashcode = hash(time(h, m, self.second, self.microsecond))
                else:
                    self._hashcode = hash((h, m, self.second, self.microsecond))
        return self._hashcode

    # Conversion to string

    def _tzstr(self, sep=":"):
        """Return formatted timezone offset (+xx:xx) or None."""
        off = self._utcoffset()
        if off is not None:
            if off < 0:
                sign = "-"
                off = -off
            else:
                sign = "+"
            hh, mm = divmod(off, 60)
            assert 0 <= hh < 24
            off = "%s%02d%s%02d" % (sign, hh, sep, mm)
        return off

    def __repr__(self):
        """Convert to formal string, for repr()."""
        if self._microsecond != 0:
            s = ", %d, %d" % (self._second, self._microsecond)
        elif self._second != 0:
            s = ", %d" % self._second
        else:
            s = ""
        module = "datetime." if self.__class__ is time else ""
        s= "%s(%d, %d%s)" % (module + self.__class__.__name__,
                             self._hour, self._minute, s)
        if self._tzinfo is not None:
            assert s[-1:] == ")"
            s = s[:-1] + ", tzinfo=%r" % self._tzinfo + ")"
        return s

    def isoformat(self):
        """Return the time formatted according to ISO.

        This is 'HH:MM:SS.mmmmmm+zz:zz', or 'HH:MM:SS+zz:zz' if
        self.microsecond == 0.
        """
        s = _format_time(self._hour, self._minute, self._second,
                         self._microsecond)
        tz = self._tzstr()
        if tz:
            s += tz
        return s

    __str__ = isoformat

    # def strftime(self, format):
    #     """Format using strftime().  The date part of the timestamp passed
    #     to underlying strftime should not be used.
    #     """
    #     # The year must be >= _MINYEARFMT else Python's strftime implementation
    #     # can raise a bogus exception.
    #     timetuple = (1900, 1, 1,
    #                  self._hour, self._minute, self._second,
    #                  0, 1, -1)
    #     return _wrap_strftime(self, format, timetuple)

    def __format__(self, fmt):
        if not isinstance(fmt, (str, unicode)):
            raise ValueError("__format__ expects str or unicode, not %s" %
                             fmt.__class__.__name__)
        if len(fmt) != 0:
            return self.strftime(fmt)
        return str(self)

    # Timezone functions

    def utcoffset(self):
        """Return the timezone offset in minutes east of UTC (negative west of
        UTC)."""
        if self._tzinfo is None:
            return None
        offset = self._tzinfo.utcoffset(None)
        offset = _check_utc_offset("utcoffset", offset)
        if offset is not None:
            offset = timedelta._create(0, offset * 60, 0, True)
        return offset

    # Return an integer (or None) instead of a timedelta (or None).
    def _utcoffset(self):
        if self._tzinfo is None:
            return None
        offset = self._tzinfo.utcoffset(None)
        offset = _check_utc_offset("utcoffset", offset)
        return offset

    def tzname(self):
        """Return the timezone name.

        Note that the name is 100% informational -- there's no requirement that
        it mean anything in particular. For example, "GMT", "UTC", "-500",
        "-5:00", "EDT", "US/Eastern", "America/New York" are all valid replies.
        """
        if self._tzinfo is None:
            return None
        name = self._tzinfo.tzname(None)
        _check_tzname(name)
        return name

    def dst(self):
        """Return 0 if DST is not in effect, or the DST offset (in minutes
        eastward) if DST is in effect.

        This is purely informational; the DST offset has already been added to
        the UTC offset returned by utcoffset() if applicable, so there's no
        need to consult dst() unless you're interested in displaying the DST
        info.
        """
        if self._tzinfo is None:
            return None
        offset = self._tzinfo.dst(None)
        offset = _check_utc_offset("dst", offset)
        if offset is not None:
            offset = timedelta._create(0, offset * 60, 0, True)
        return offset

    # Return an integer (or None) instead of a timedelta (or None).
    def _dst(self):
        if self._tzinfo is None:
            return None
        offset = self._tzinfo.dst(None)
        offset = _check_utc_offset("dst", offset)
        return offset

    def replace(self, hour=None, minute=None, second=None, microsecond=None,
                tzinfo=True):
        """Return a new time with new values for the specified fields."""
        if hour is None:
            hour = self.hour
        if minute is None:
            minute = self.minute
        if second is None:
            second = self.second
        if microsecond is None:
            microsecond = self.microsecond
        if tzinfo is True:
            tzinfo = self.tzinfo
        return time.__new__(type(self),
                            hour, minute, second, microsecond, tzinfo)

    def __nonzero__(self):
        if self.second or self.microsecond:
            return True
        offset = self._utcoffset() or 0
        return self.hour * 60 + self.minute != offset

    # Pickle support.

    def _getstate(self):
        us2, us3 = divmod(self._microsecond, 256)
        us1, us2 = divmod(us2, 256)
        basestate = _struct.pack('6B', self._hour, self._minute, self._second,
                                       us1, us2, us3)
        if self._tzinfo is None:
            return (basestate,)
        else:
            return (basestate, self._tzinfo)

    def __setstate(self, string, tzinfo):
        if tzinfo is not None and not isinstance(tzinfo, _tzinfo_class):
            raise TypeError("bad tzinfo state arg")
        self._hour, self._minute, self._second, us1, us2, us3 = (
            ord(string[0]), ord(string[1]), ord(string[2]),
            ord(string[3]), ord(string[4]), ord(string[5]))
        self._microsecond = (((us1 << 8) | us2) << 8) | us3
        self._tzinfo = tzinfo

    def __reduce__(self):
        return (time, self._getstate())

_time_class = time  # so functions w/ args named "time" can get at the class

time.min = time(0, 0, 0)
time.max = time(23, 59, 59, 999999)
time.resolution = timedelta(microseconds=1)

class datetime(date):
    """datetime(year, month, day[, hour[, minute[, second[, microsecond[,tzinfo]]]]])

    The year, month and day arguments are required. tzinfo may be None, or an
    instance of a tzinfo subclass. The remaining arguments may be ints or longs.
    """
    __slots__ = date.__slots__ + time.__slots__

    def __new__(cls, year, month=None, day=None, hour=0, minute=0, second=0,
                microsecond=0, tzinfo=None):
        # if isinstance(year, bytes) and len(year) == 10 and \
        #         1 <= ord(year[2]) <= 12:
        #     # Pickle support
        #     self = object.__new__(cls)
        #     self.__setstate(year, month)
        #     self._hashcode = -1
        #     return self
        year, month, day = _check_date_fields(year, month, day)
        hour, minute, second, microsecond = _check_time_fields(
            hour, minute, second, microsecond)
        _check_tzinfo_arg(tzinfo)
        self = object.__new__(cls)
        self._year = year
        self._month = month
        self._day = day
        self._hour = hour
        self._minute = minute
        self._second = second
        self._microsecond = microsecond
        self._tzinfo = tzinfo
        self._hashcode = -1
        return self

    # Read-only field accessors
    @property
    def hour(self):
        """hour (0-23)"""
        return self._hour

    @property
    def minute(self):
        """minute (0-59)"""
        return self._minute

    @property
    def second(self):
        """second (0-59)"""
        return self._second

    @property
    def microsecond(self):
        """microsecond (0-999999)"""
        return self._microsecond

    @property
    def tzinfo(self):
        """timezone info object"""
        return self._tzinfo

    @classmethod
    def fromtimestamp(cls, timestamp, tz=None):
        """Construct a datetime from a POSIX timestamp (like time.time()).

        A timezone info object may be passed in as well.
        """
        _check_tzinfo_arg(tz)
        converter = _time.localtime if tz is None else _time.gmtime
        self = cls._from_timestamp(converter, timestamp, tz)
        if tz is not None:
            self = tz.fromutc(self)
        return self

    @classmethod
    def utcfromtimestamp(cls, t):
        "Construct a UTC datetime from a POSIX timestamp (like time.time())."
        return cls._from_timestamp(_time.gmtime, t, None)

    @classmethod
    def _from_timestamp(cls, converter, timestamp, tzinfo):
        t_full = timestamp
        timestamp = int(_math.floor(timestamp))
        frac = t_full - timestamp
        us = _round(frac * 1e6)

        # If timestamp is less than one microsecond smaller than a
        # full second, us can be rounded up to 1000000.  In this case,
        # roll over to seconds, otherwise, ValueError is raised
        # by the constructor.
        if us == 1000000:
            timestamp += 1
            us = 0
        y, m, d, hh, mm, ss, weekday, jday, dst = converter(timestamp)
        ss = min(ss, 59)    # clamp out leap seconds if the platform has them
        return cls(y, m, d, hh, mm, ss, us, tzinfo)

    @classmethod
    def now(cls, tz=None):
        "Construct a datetime from time.time() and optional time zone info."
        t = _time.time()
        return cls.fromtimestamp(t, tz)

    @classmethod
    def utcnow(cls):
        "Construct a UTC datetime from time.time()."
        t = _time.time()
        return cls.utcfromtimestamp(t)

    @classmethod
    def combine(cls, date, time):
        "Construct a datetime from a given date and a given time."
        if not isinstance(date, _date_class):
            raise TypeError("date argument must be a date instance")
        if not isinstance(time, _time_class):
            raise TypeError("time argument must be a time instance")
        return cls(date.year, date.month, date.day,
                   time.hour, time.minute, time.second, time.microsecond,
                   time.tzinfo)

    def timetuple(self):
        "Return local time tuple compatible with time.localtime()."
        dst = self._dst()
        if dst is None:
            dst = -1
        elif dst:
            dst = 1
        return _build_struct_time(self.year, self.month, self.day,
                                  self.hour, self.minute, self.second,
                                  dst)

    def utctimetuple(self):
        "Return UTC time tuple compatible with time.gmtime()."
        y, m, d = self.year, self.month, self.day
        hh, mm, ss = self.hour, self.minute, self.second
        offset = self._utcoffset()
        if offset:  # neither None nor 0
            mm -= offset
            y, m, d, hh, mm, ss, _ = _normalize_datetime(
                y, m, d, hh, mm, ss, 0, ignore_overflow=True)
        return _build_struct_time(y, m, d, hh, mm, ss, 0)

    def date(self):
        "Return the date part."
        return date(self._year, self._month, self._day)

    def time(self):
        "Return the time part, with tzinfo None."
        return time(self.hour, self.minute, self.second, self.microsecond)

    def timetz(self):
        "Return the time part, with same tzinfo."
        return time(self.hour, self.minute, self.second, self.microsecond,
                    self._tzinfo)

    def replace(self, year=None, month=None, day=None, hour=None,
                minute=None, second=None, microsecond=None, tzinfo=True):
        """Return a new datetime with new values for the specified fields."""
        if year is None:
            year = self.year
        if month is None:
            month = self.month
        if day is None:
            day = self.day
        if hour is None:
            hour = self.hour
        if minute is None:
            minute = self.minute
        if second is None:
            second = self.second
        if microsecond is None:
            microsecond = self.microsecond
        if tzinfo is True:
            tzinfo = self.tzinfo
        return datetime.__new__(type(self),
                                year, month, day, hour, minute, second,
                                microsecond, tzinfo)

    def astimezone(self, tz):
        if not isinstance(tz, tzinfo):
            raise TypeError("tz argument must be an instance of tzinfo")

        mytz = self.tzinfo
        if mytz is None:
            raise ValueError("astimezone() requires an aware datetime")

        if tz is mytz:
            return self

        # Convert self to UTC, and attach the new time zone object.
        myoffset = self.utcoffset()
        if myoffset is None:
            raise ValueError("astimezone() requires an aware datetime")
        utc = (self - myoffset).replace(tzinfo=tz)

        # Convert from UTC to tz's local time.
        return tz.fromutc(utc)

    # Ways to produce a string.

    def ctime(self):
        "Return ctime() style string."
        weekday = self.toordinal() % 7 or 7
        return "%s %s %2d %02d:%02d:%02d %04d" % (
            _DAYNAMES[weekday],
            _MONTHNAMES[self._month],
            self._day,
            self._hour, self._minute, self._second,
            self._year)

    def isoformat(self, sep='T'):
        """Return the time formatted according to ISO.

        This is 'YYYY-MM-DD HH:MM:SS.mmmmmm', or 'YYYY-MM-DD HH:MM:SS' if
        self.microsecond == 0.

        If self.tzinfo is not None, the UTC offset is also attached, giving
        'YYYY-MM-DD HH:MM:SS.mmmmmm+HH:MM' or 'YYYY-MM-DD HH:MM:SS+HH:MM'.

        Optional argument sep specifies the separator between date and
        time, default 'T'.
        """
        s = ("%04d-%02d-%02d%c" % (self._year, self._month, self._day, sep) +
             _format_time(self._hour, self._minute, self._second,
                          self._microsecond))
        off = self._utcoffset()
        if off is not None:
            if off < 0:
                sign = "-"
                off = -off
            else:
                sign = "+"
            hh, mm = divmod(off, 60)
            s += "%s%02d:%02d" % (sign, hh, mm)
        return s

    def __repr__(self):
        """Convert to formal string, for repr()."""
        L = [self._year, self._month, self._day,  # These are never zero
             self._hour, self._minute, self._second, self._microsecond]
        if L[-1] == 0:
            del L[-1]
        if L[-1] == 0:
            del L[-1]
        s = ", ".join(map(str, L))
        module = "datetime." if self.__class__ is datetime else ""
        s = "%s(%s)" % (module + self.__class__.__name__, s)
        if self._tzinfo is not None:
            assert s[-1:] == ")"
            s = s[:-1] + ", tzinfo=%r" % self._tzinfo + ")"
        return s

    def __str__(self):
        "Convert to string, for str()."
        return self.isoformat(sep=' ')

    # @classmethod
    # def strptime(cls, date_string, format):
    #     'string, format -> new datetime parsed from a string (like time.strptime()).'
    #     from _strptime import _strptime
    #     # _strptime._strptime returns a two-element tuple.  The first
    #     # element is a time.struct_time object.  The second is the
    #     # microseconds (which are not defined for time.struct_time).
    #     struct, micros = _strptime(date_string, format)
    #     return cls(*(struct[0:6] + (micros,)))

    def utcoffset(self):
        """Return the timezone offset in minutes east of UTC (negative west of
        UTC)."""
        if self._tzinfo is None:
            return None
        offset = self._tzinfo.utcoffset(self)
        offset = _check_utc_offset("utcoffset", offset)
        if offset is not None:
            offset = timedelta._create(0, offset * 60, 0, True)
        return offset

    # Return an integer (or None) instead of a timedelta (or None).
    def _utcoffset(self):
        if self._tzinfo is None:
            return None
        offset = self._tzinfo.utcoffset(self)
        offset = _check_utc_offset("utcoffset", offset)
        return offset

    def tzname(self):
        """Return the timezone name.

        Note that the name is 100% informational -- there's no requirement that
        it mean anything in particular. For example, "GMT", "UTC", "-500",
        "-5:00", "EDT", "US/Eastern", "America/New York" are all valid replies.
        """
        if self._tzinfo is None:
            return None
        name = self._tzinfo.tzname(self)
        _check_tzname(name)
        return name

    def dst(self):
        """Return 0 if DST is not in effect, or the DST offset (in minutes
        eastward) if DST is in effect.

        This is purely informational; the DST offset has already been added to
        the UTC offset returned by utcoffset() if applicable, so there's no
        need to consult dst() unless you're interested in displaying the DST
        info.
        """
        if self._tzinfo is None:
            return None
        offset = self._tzinfo.dst(self)
        offset = _check_utc_offset("dst", offset)
        if offset is not None:
            offset = timedelta._create(0, offset * 60, 0, True)
        return offset

    # Return an integer (or None) instead of a timedelta (or None).
    def _dst(self):
        if self._tzinfo is None:
            return None
        offset = self._tzinfo.dst(self)
        offset = _check_utc_offset("dst", offset)
        return offset

    # Comparisons of datetime objects with other.

    def __eq__(self, other):
        if isinstance(other, datetime):
            return self._cmp(other) == 0
        elif hasattr(other, "timetuple") and not isinstance(other, date):
            return NotImplemented
        else:
            return False

    def __ne__(self, other):
        if isinstance(other, datetime):
            return self._cmp(other) != 0
        elif hasattr(other, "timetuple") and not isinstance(other, date):
            return NotImplemented
        else:
            return True

    def __le__(self, other):
        if isinstance(other, datetime):
            return self._cmp(other) <= 0
        elif hasattr(other, "timetuple") and not isinstance(other, date):
            return NotImplemented
        else:
            _cmperror(self, other)

    def __lt__(self, other):
        if isinstance(other, datetime):
            return self._cmp(other) < 0
        elif hasattr(other, "timetuple") and not isinstance(other, date):
            return NotImplemented
        else:
            _cmperror(self, other)

    def __ge__(self, other):
        if isinstance(other, datetime):
            return self._cmp(other) >= 0
        elif hasattr(other, "timetuple") and not isinstance(other, date):
            return NotImplemented
        else:
            _cmperror(self, other)

    def __gt__(self, other):
        if isinstance(other, datetime):
            return self._cmp(other) > 0
        elif hasattr(other, "timetuple") and not isinstance(other, date):
            return NotImplemented
        else:
            _cmperror(self, other)

    def _cmp(self, other):
        assert isinstance(other, datetime)
        mytz = self._tzinfo
        ottz = other._tzinfo
        myoff = otoff = None

        if mytz is ottz:
            base_compare = True
        else:
            if mytz is not None:
                myoff = self._utcoffset()
            if ottz is not None:
                otoff = other._utcoffset()
            base_compare = myoff == otoff

        if base_compare:
            return _cmp((self._year, self._month, self._day,
                         self._hour, self._minute, self._second,
                         self._microsecond),
                        (other._year, other._month, other._day,
                         other._hour, other._minute, other._second,
                         other._microsecond))
        if myoff is None or otoff is None:
            raise TypeError("can't compare offset-naive and offset-aware datetimes")
        # XXX What follows could be done more efficiently...
        diff = self - other     # this will take offsets into account
        if diff.days < 0:
            return -1
        return diff and 1 or 0

    def _add_timedelta(self, other, factor):
        y, m, d, hh, mm, ss, us = _normalize_datetime(
            self._year,
            self._month,
            self._day + other.days * factor,
            self._hour,
            self._minute,
            self._second + other.seconds * factor,
            self._microsecond + other.microseconds * factor)
        return datetime(y, m, d, hh, mm, ss, us, tzinfo=self._tzinfo)

    def __add__(self, other):
        "Add a datetime and a timedelta."
        if not isinstance(other, timedelta):
            return NotImplemented
        return self._add_timedelta(other, 1)

    __radd__ = __add__

    def __sub__(self, other):
        "Subtract two datetimes, or a datetime and a timedelta."
        if not isinstance(other, datetime):
            if isinstance(other, timedelta):
                return self._add_timedelta(other, -1)
            return NotImplemented

        delta_d = self.toordinal() - other.toordinal()
        delta_s = (self._hour - other._hour) * 3600 + \
                  (self._minute - other._minute) * 60 + \
                  (self._second - other._second)
        delta_us = self._microsecond - other._microsecond
        base = timedelta._create(delta_d, delta_s, delta_us, True)
        if self._tzinfo is other._tzinfo:
            return base
        myoff = self._utcoffset()
        otoff = other._utcoffset()
        if myoff == otoff:
            return base
        if myoff is None or otoff is None:
            raise TypeError("can't subtract offset-naive and offset-aware datetimes")
        return base + timedelta(minutes = otoff-myoff)

    def __hash__(self):
        if self._hashcode == -1:
            tzoff = self._utcoffset()
            if tzoff is None:
                self._hashcode = hash(self._getstate()[0])
            else:
                days = _ymd2ord(self.year, self.month, self.day)
                seconds = self.hour * 3600 + (self.minute - tzoff) * 60 + self.second
                self._hashcode = hash(timedelta(days, seconds, self.microsecond))
        return self._hashcode

    # Pickle support.

    def _getstate(self):
        yhi, ylo = divmod(self._year, 256)
        us2, us3 = divmod(self._microsecond, 256)
        us1, us2 = divmod(us2, 256)
        basestate = _struct.pack('10B', yhi, ylo, self._month, self._day,
                                        self._hour, self._minute, self._second,
                                        us1, us2, us3)
        if self._tzinfo is None:
            return (basestate,)
        else:
            return (basestate, self._tzinfo)

    def __setstate(self, string, tzinfo):
        if tzinfo is not None and not isinstance(tzinfo, _tzinfo_class):
            raise TypeError("bad tzinfo state arg")
        (yhi, ylo, self._month, self._day, self._hour,
            self._minute, self._second, us1, us2, us3) = (ord(string[0]),
                ord(string[1]), ord(string[2]), ord(string[3]),
                ord(string[4]), ord(string[5]), ord(string[6]),
                ord(string[7]), ord(string[8]), ord(string[9]))
        self._year = yhi * 256 + ylo
        self._microsecond = (((us1 << 8) | us2) << 8) | us3
        self._tzinfo = tzinfo

    def __reduce__(self):
        return (self.__class__, self._getstate())


datetime.min = datetime(1, 1, 1)
datetime.max = datetime(9999, 12, 31, 23, 59, 59, 999999)
datetime.resolution = timedelta(microseconds=1)


def _isoweek1monday(year):
    # Helper to calculate the day number of the Monday starting week 1
    # XXX This could be done more efficiently
    THURSDAY = 3
    firstday = _ymd2ord(year, 1, 1)
    firstweekday = (firstday + 6) % 7  # See weekday() above
    week1monday = firstday - firstweekday
    if firstweekday > THURSDAY:
        week1monday += 7
    return week1monday

"""
Some time zone algebra.  For a datetime x, let
    x.n = x stripped of its timezone -- its naive time.
    x.o = x.utcoffset(), and assuming that doesn't raise an exception or
          return None
    x.d = x.dst(), and assuming that doesn't raise an exception or
          return None
    x.s = x's standard offset, x.o - x.d

Now some derived rules, where k is a duration (timedelta).

1. x.o = x.s + x.d
   This follows from the definition of x.s.

2. If x and y have the same tzinfo member, x.s = y.s.
   This is actually a requirement, an assumption we need to make about
   sane tzinfo classes.

3. The naive UTC time corresponding to x is x.n - x.o.
   This is again a requirement for a sane tzinfo class.

4. (x+k).s = x.s
   This follows from #2, and that datimetimetz+timedelta preserves tzinfo.

5. (x+k).n = x.n + k
   Again follows from how arithmetic is defined.

Now we can explain tz.fromutc(x).  Let's assume it's an interesting case
(meaning that the various tzinfo methods exist, and don't blow up or return
None when called).

The function wants to return a datetime y with timezone tz, equivalent to x.
x is already in UTC.

By #3, we want

    y.n - y.o = x.n                             [1]

The algorithm starts by attaching tz to x.n, and calling that y.  So
x.n = y.n at the start.  Then it wants to add a duration k to y, so that [1]
becomes true; in effect, we want to solve [2] for k:

   (y+k).n - (y+k).o = x.n                      [2]

By #1, this is the same as

   (y+k).n - ((y+k).s + (y+k).d) = x.n          [3]

By #5, (y+k).n = y.n + k, which equals x.n + k because x.n=y.n at the start.
Substituting that into [3],

   x.n + k - (y+k).s - (y+k).d = x.n; the x.n terms cancel, leaving
   k - (y+k).s - (y+k).d = 0; rearranging,
   k = (y+k).s - (y+k).d; by #4, (y+k).s == y.s, so
   k = y.s - (y+k).d

On the RHS, (y+k).d can't be computed directly, but y.s can be, and we
approximate k by ignoring the (y+k).d term at first.  Note that k can't be
very large, since all offset-returning methods return a duration of magnitude
less than 24 hours.  For that reason, if y is firmly in std time, (y+k).d must
be 0, so ignoring it has no consequence then.

In any case, the new value is

    z = y + y.s                                 [4]

It's helpful to step back at look at [4] from a higher level:  it's simply
mapping from UTC to tz's standard time.

At this point, if

    z.n - z.o = x.n                             [5]

we have an equivalent time, and are almost done.  The insecurity here is
at the start of daylight time.  Picture US Eastern for concreteness.  The wall
time jumps from 1:59 to 3:00, and wall hours of the form 2:MM don't make good
sense then.  The docs ask that an Eastern tzinfo class consider such a time to
be EDT (because it's "after 2"), which is a redundant spelling of 1:MM EST
on the day DST starts.  We want to return the 1:MM EST spelling because that's
the only spelling that makes sense on the local wall clock.

In fact, if [5] holds at this point, we do have the standard-time spelling,
but that takes a bit of proof.  We first prove a stronger result.  What's the
difference between the LHS and RHS of [5]?  Let

    diff = x.n - (z.n - z.o)                    [6]

Now
    z.n =                       by [4]
    (y + y.s).n =               by #5
    y.n + y.s =                 since y.n = x.n
    x.n + y.s =                 since z and y are have the same tzinfo member,
                                    y.s = z.s by #2
    x.n + z.s

Plugging that back into [6] gives

    diff =
    x.n - ((x.n + z.s) - z.o) =     expanding
    x.n - x.n - z.s + z.o =         cancelling
    - z.s + z.o =                   by #2
    z.d

So diff = z.d.

If [5] is true now, diff = 0, so z.d = 0 too, and we have the standard-time
spelling we wanted in the endcase described above.  We're done.  Contrarily,
if z.d = 0, then we have a UTC equivalent, and are also done.

If [5] is not true now, diff = z.d != 0, and z.d is the offset we need to
add to z (in effect, z is in tz's standard time, and we need to shift the
local clock into tz's daylight time).

Let

    z' = z + z.d = z + diff                     [7]

and we can again ask whether

    z'.n - z'.o = x.n                           [8]

If so, we're done.  If not, the tzinfo class is insane, according to the
assumptions we've made.  This also requires a bit of proof.  As before, let's
compute the difference between the LHS and RHS of [8] (and skipping some of
the justifications for the kinds of substitutions we've done several times
already):

    diff' = x.n - (z'.n - z'.o) =           replacing z'.n via [7]
            x.n  - (z.n + diff - z'.o) =    replacing diff via [6]
            x.n - (z.n + x.n - (z.n - z.o) - z'.o) =
            x.n - z.n - x.n + z.n - z.o + z'.o =    cancel x.n
            - z.n + z.n - z.o + z'.o =              cancel z.n
            - z.o + z'.o =                      #1 twice
            -z.s - z.d + z'.s + z'.d =          z and z' have same tzinfo
            z'.d - z.d

So z' is UTC-equivalent to x iff z'.d = z.d at this point.  If they are equal,
we've found the UTC-equivalent so are done.  In fact, we stop with [7] and
return z', not bothering to compute z'.d.

How could z.d and z'd differ?  z' = z + z.d [7], so merely moving z' by
a dst() offset, and starting *from* a time already in DST (we know z.d != 0),
would have to change the result dst() returns:  we start in DST, and moving
a little further into it takes us out of DST.

There isn't a sane case where this can happen.  The closest it gets is at
the end of DST, where there's an hour in UTC with no spelling in a hybrid
tzinfo class.  In US Eastern, that's 5:MM UTC = 0:MM EST = 1:MM EDT.  During
that hour, on an Eastern clock 1:MM is taken as being in standard time (6:MM
UTC) because the docs insist on that, but 0:MM is taken as being in daylight
time (4:MM UTC).  There is no local time mapping to 5:MM UTC.  The local
clock jumps from 1:59 back to 1:00 again, and repeats the 1:MM hour in
standard time.  Since that's what the local clock *does*, we want to map both
UTC hours 5:MM and 6:MM to 1:MM Eastern.  The result is ambiguous
in local time, but so it goes -- it's the way the local clock works.

When x = 5:MM UTC is the input to this algorithm, x.o=0, y.o=-5 and y.d=0,
so z=0:MM.  z.d=60 (minutes) then, so [5] doesn't hold and we keep going.
z' = z + z.d = 1:MM then, and z'.d=0, and z'.d - z.d = -60 != 0 so [8]
(correctly) concludes that z' is not UTC-equivalent to x.

Because we know z.d said z was in daylight time (else [5] would have held and
we would have stopped then), and we know z.d != z'.d (else [8] would have held
and we have stopped then), and there are only 2 possible values dst() can
return in Eastern, it follows that z'.d must be 0 (which it is in the example,
but the reasoning doesn't depend on the example -- it depends on there being
two possible dst() outcomes, one zero and the other non-zero).  Therefore
z' must be in standard time, and is the spelling we want in this case.

Note again that z' is not UTC-equivalent as far as the hybrid tzinfo class is
concerned (because it takes z' as being in standard time rather than the
daylight time we intend here), but returning it gives the real-life "local
clock repeats an hour" behavior when mapping the "unspellable" UTC hour into
tz.

When the input is 6:MM, z=1:MM and z.d=0, and we stop at once, again with
the 1:MM standard time spelling we want.

So how can this break?  One of the assumptions must be violated.  Two
possibilities:

1) [2] effectively says that y.s is invariant across all y belong to a given
   time zone.  This isn't true if, for political reasons or continental drift,
   a region decides to change its base offset from UTC.

2) There may be versions of "double daylight" time where the tail end of
   the analysis gives up a step too early.  I haven't thought about that
   enough to say.

In any case, it's clear that the default fromutc() is strong enough to handle
"almost all" time zones:  so long as the standard offset is invariant, it
doesn't matter if daylight time transition points change from year to year, or
if daylight time is skipped in some years; it doesn't matter how large or
small dst() may get within its bounds; and it doesn't even matter if some
perverse time zone returns a negative dst()).  So a breaking case must be
pretty bizarre, and a tzinfo subclass can override fromutc() if it is.
"""

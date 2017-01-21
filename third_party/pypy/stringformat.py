# -*- coding: utf-8 -*-
"""Advanced string formatting for Python >= 2.4.

An implementation of the advanced string formatting (PEP 3101).

Author: Florent Xicluna
"""

import re

__all__ = ['FormattableString', 'init']

def isdigit(s):
  for x in s:
    if x < "0" or x > "9":
      return False
  return True

def setdefault(d, k, v):
  if k not in d:
    d[k] = v
  return d[k]

def rjust(s, w, f=" "):
  return (f * (w - len(s)) + s) if w > len(s) else s

def ljust(s, w, f=" "):
  return s + (f * (w - len(s))) if w > len(s) else s

if hasattr(str, 'partition'):
    def partition(s, sep):
        return s.partition(sep)
else:   # Python 2.4
    def partition(s, sep):
        try:
            left, right = s.split(sep, 1)
        except ValueError:
            return s, '', ''
        return left, sep, right

_format_str_re = re.compile(
    r'(%)'                                  # '%'
    r'|((?<!{)(?:{{)+'                      # '{{'
    r'|(?:}})+(?!})'                        # '}}
    r'|{(?:[^{](?:[^{}]+|{[^{}]*})*)?})'    # replacement field
)
_format_sub_re = re.compile(r'({[^{}]*})')  # nested replacement field
_format_spec_re = re.compile(
    r'((?:[^{}]?[<>=^])?)'      # alignment
    r'([-+ ]?)'                 # sign
    r'(#?)' r'(\d*)' r'(,?)'    # base prefix, minimal width, thousands sep
    r'((?:\.\d+)?)'             # precision
    r'(.?)$'                    # type
)
_field_part_re = re.compile(
    r'(?:(\[)|\.|^)'            # start or '.' or '['
    r'((?(1)[^]]*|[^.[]*))'     # part
    r'(?(1)(?:\]|$)([^.[]+)?)'  # ']' and invalid tail
)

if hasattr(re, '__version__'):
    _format_str_sub = _format_str_re.sub
else:
    # Python 2.4 fails to preserve the Unicode type
    def _format_str_sub(repl, s):
        if isinstance(s, unicode):
            return unicode(_format_str_re.sub(repl, s))
        return _format_str_re.sub(repl, s)

if hasattr(int, '__index__'):
    def _is_integer(value):
        return hasattr(value, '__index__')
else:   # Python 2.4
    def _is_integer(value):
        return isinstance(value, (int, long))

try:
    unicode
    def _chr(n):
        return chr(n % 256)
except NameError:   # Python 3
    unicode = str
    _chr = chr


def _strformat(value, format_spec=""):
    """Internal string formatter.

    It implements the Format Specification Mini-Language.
    """
    m = _format_spec_re.match(str(format_spec))
    if not m:
        raise ValueError('Invalid conversion specification')
    align, sign, prefix, width, comma, precision, conversion = m.groups()
    is_numeric = hasattr(value, '__float__')
    is_integer = is_numeric and _is_integer(value)
    if prefix and not is_integer:
        raise ValueError('Alternate form (#) not allowed in %s format '
                         'specifier' % (is_numeric and 'float' or 'string'))
    if is_numeric and conversion == 'n':
        # Default to 'd' for ints and 'g' for floats
        conversion = is_integer and 'd' or 'g'
    elif sign:
        if not is_numeric:
            raise ValueError("Sign not allowed in string format specifier")
        if conversion == 'c':
            raise ValueError("Sign not allowed with integer "
                             "format specifier 'c'")
    if comma:
        # TODO: thousand separator
        pass
    try:
        if ((is_numeric and conversion == 's') or
            (not is_integer and conversion in set('cdoxX'))):
            raise ValueError
        if conversion == 'c':
            conversion = 's'
            value = _chr(value)
        rv = ('%' + prefix + precision + (conversion or 's')) % (value,)
    except ValueError:
        raise ValueError("Unknown format code %r for object of type %r" %
                         (conversion, value.__class__.__name__))
    if sign not in '-' and value >= 0:
        # sign in (' ', '+')
        rv = sign + rv
    if width:
        zero = (width[0] == '0')
        width = int(width)
    else:
        zero = False
        width = 0
    # Fastpath when alignment is not required
    if width <= len(rv):
        if not is_numeric and (align == '=' or (zero and not align)):
            raise ValueError("'=' alignment not allowed in string format "
                             "specifier")
        return rv
    fill, align = align[:-1], align[-1:]
    if not fill:
        fill = zero and '0' or ' '
    if align == '^':
        padding = width - len(rv)
        # tweak the formatting if the padding is odd
        if padding % 2:
            rv += fill
        rv = rv.center(width, fill)
    elif align == '=' or (zero and not align):
        if not is_numeric:
            raise ValueError("'=' alignment not allowed in string format "
                             "specifier")
        if value < 0 or sign not in '-':
            # rv = rv[0] + rv[1:].rjust(width - 1, fill)
            rv = rv[0] + rjust(rv[1:], width - 1, fill)
        else:
            # rv = rv.rjust(width, fill)
            rv = rjust(rv, width, fill)
    elif align in ('>', '=') or (is_numeric and not align):
        # numeric value right aligned by default
        # rv = rv.rjust(width, fill)
        rv = rjust(rv, width, fill)
    else:
        # rv = rv.ljust(width, fill)
        rv = ljust(rv, width, fill)
    return rv


def _format_field(value, parts, conv, spec, want_bytes=False):
    """Format a replacement field."""
    for k, part, _ in parts:
        if k:
            # if part.isdigit():
            if isdigit(part):
                value = value[int(part)]
            else:
                value = value[part]
        else:
            value = getattr(value, part)
    if conv:
        value = ((conv == 'r') and '%r' or '%s') % (value,)
    if hasattr(value, '__format__'):
        value = value.__format__(spec)
    elif hasattr(value, 'strftime') and spec:
        value = value.strftime(str(spec))
    else:
        value = _strformat(value, spec)
    if want_bytes and isinstance(value, unicode):
        return str(value)
    return value


class FormattableString(object):
    """Class which implements method format().

    The method format() behaves like str.format() in python 2.6+.

    >>> FormattableString(u'{a:5}').format(a=42)
    ... # Same as u'{a:5}'.format(a=42)
    u'   42'

    """

    __slots__ = '_index', '_kwords', '_nested', '_string', 'format_string'

    def __init__(self, format_string):
        self._index = 0
        self._kwords = {}
        self._nested = {}

        self.format_string = format_string
        self._string = _format_str_sub(self._prepare, format_string)

    def __eq__(self, other):
        if isinstance(other, FormattableString):
            return self.format_string == other.format_string
        # Compare equal with the original string.
        return self.format_string == other

    def _prepare(self, match):
        # Called for each replacement field.
        part = match.group(0)
        if part == '%':
            return '%%'
        if part[0] == part[-1]:
            # '{{' or '}}'
            assert part == part[0] * len(part)
            return part[:len(part) // 2]
        repl = part[1:-1]
        field, _, format_spec = partition(repl, ':')
        literal, sep, conversion = partition(field, '!')
        if sep and not conversion:
            raise ValueError("end of format while looking for "
                             "conversion specifier")
        if len(conversion) > 1:
            raise ValueError("expected ':' after format specifier")
        if conversion not in 'rsa':
            raise ValueError("Unknown conversion specifier %s" %
                             str(conversion))
        name_parts = _field_part_re.findall(literal)
        if literal[:1] in '.[':
            # Auto-numbering
            if self._index is None:
                raise ValueError("cannot switch from manual field "
                                 "specification to automatic field numbering")
            name = str(self._index)
            self._index += 1
            if not literal:
                # del name_parts[0]
                name_parts.pop(0)
        else:
            name = name_parts.pop(0)[1]
            if isdigit(name) and self._index is not None:
                # Manual specification
                if self._index:
                    raise ValueError("cannot switch from automatic field "
                                     "numbering to manual field specification")
                self._index = None
        empty_attribute = False
        for k, v, tail in name_parts:
            if not v:
                empty_attribute = True
            if tail:
                raise ValueError("Only '.' or '[' may follow ']' "
                                 "in format field specifier")
        if name_parts and k == '[' and not literal[-1] == ']':
            raise ValueError("Missing ']' in format string")
        if empty_attribute:
            raise ValueError("Empty attribute in format string")
        if '{' in format_spec:
            format_spec = _format_sub_re.sub(self._prepare, format_spec)
            rv = (name_parts, conversion, format_spec)
            # self._nested.setdefault(name, []).append(rv)
            setdefault(self._nested, name, []).append(rv)
        else:
            rv = (name_parts, conversion, format_spec)
            # self._kwords.setdefault(name, []).append(rv)
            setdefault(self._kwords, name, []).append(rv)
        return r'%%(%s)s' % id(rv)

    def format(self, *args, **kwargs):
        """Same as str.format() and unicode.format() in Python 2.6+."""
        if args:
            kwargs.update(dict((str(i), value)
                               for (i, value) in enumerate(args)))
        # Encode arguments to ASCII, if format string is bytes
        want_bytes = isinstance(self._string, str)
        params = {}
        for name, items in self._kwords.items():
            value = kwargs[name]
            for item in items:
                parts, conv, spec = item
                params[str(id(item))] = _format_field(value, parts, conv, spec,
                                                      want_bytes)
        for name, items in self._nested.items():
            value = kwargs[name]
            for item in items:
                parts, conv, spec = item
                spec = spec % params
                params[str(id(item))] = _format_field(value, parts, conv, spec,
                                                      want_bytes)
        # return self._string % params
        return re.sub(r'%\((\d+)\)s',
                      lambda x: params.get(x.group(1), x.group(0)), self._string)


# the code below is used to monkey patch builtins
# def _patch_builtin_types():
#     # originally from https://gist.github.com/295200 (Armin R.)
#     import ctypes
#     import sys

#     # figure out size of _Py_ssize_t
#     if hasattr(ctypes.pythonapi, 'Py_InitModule4_64'):
#         _Py_ssize_t = ctypes.c_int64
#     else:
#         _Py_ssize_t = ctypes.c_int

#     class _PyObject(ctypes.Structure):
#         pass
#     _PyObject._fields_ = [
#         ('ob_refcnt', _Py_ssize_t),
#         ('ob_type', ctypes.POINTER(_PyObject)),
#     ]

#     # python with trace
#     if hasattr(sys, 'getobjects'):
#         _PyObject._fields_[0:0] = [
#             ('_ob_next', ctypes.POINTER(_PyObject)),
#             ('_ob_prev', ctypes.POINTER(_PyObject)),
#         ]

#     class _DictProxy(_PyObject):
#         _fields_ = [('dict', ctypes.POINTER(_PyObject))]

#     def get_class_dict(cls):
#         d = getattr(cls, '__dict__', None)
#         if hasattr(d, 'pop'):
#             return d
#         if d is None:
#             raise TypeError('given class does not have a dictionary')
#         setitem = ctypes.pythonapi.PyDict_SetItem
#         ns = {}
#         # Reveal dict behind DictProxy
#         dp = _DictProxy.from_address(id(d))
#         setitem(ctypes.py_object(ns), ctypes.py_object(None), dp.dict)
#         return ns[None]

#     def format(self, *args, **kwargs):
#         """S.format(*args, **kwargs) -> string

#         Return a formatted version of S, using substitutions from args and kwargs.
#         The substitutions are identified by braces ('{' and '}').
#         """
#         return FormattableString(self).format(*args, **kwargs)

#     # This does the actual monkey patch on str and unicode
#     for cls in str, unicode:
#         get_class_dict(cls)['format'] = format


# def init(force=False):
#     if force or not hasattr(str, 'format'):
#         _patch_builtin_types()


# def selftest():
#     import datetime
#     F = FormattableString
#     d = datetime.date(2010, 9, 7)
#     u = unicode

#     # Initialize
#     init(True)

#     assert F(u("{0:{width}.{precision}s}")).format('hello world',
#             width=8, precision=5) == u('hello   ')
#     assert F(u("The year is {0.year}")).format(d) == u("The year is 2010")
#     assert F(u("Tested on {0:%Y-%m-%d}")).format(d) == u("Tested on 2010-09-07")

#     assert u("{0:{width}.{precision}s}").format('hello world',
#             width=8, precision=5) == u('hello   ')
#     assert u("The year is {0.year}").format(d) == u("The year is 2010")
#     assert u("Tested on {0:%Y-%m-%d}").format(d) == u("Tested on 2010-09-07")

#     assert "{0:{width}.{precision}s}".format('hello world',
#             width=8, precision=5) == 'hello   '
#     assert "The year is {0.year}".format(d) == "The year is 2010"
#     assert "Tested on {0:%Y-%m-%d}".format(d) == "Tested on 2010-09-07"

#     print('Test successful')

# if __name__ == '__main__':
# #     selftest()
#       print FormattableString('{a:5} {} {:d} {:<10}').format("abcd", 1234, "a", a=42)

def formatter(self, *args, **kw):
  return FormattableString(self).format(*args, **kw)

str.__dict__["format"] = formatter
# print ['{a:5} {} {:d} {:<10}'.format("abcd", 1234, "a", a=42)]

#
# This module is a pure Python version of pypy.module.struct.
# It is only imported if the vastly faster pypy.module.struct is not
# compiled in.  For now we keep this version for reference and
# because pypy.module.struct is not ootype-backend-friendly yet.
#

"""Functions to convert between Python values and C structs.
Python strings are used to hold the data representing the C struct
and also as format strings to describe the layout of data in the C struct.

The optional first format char indicates byte order, size and alignment:
 @: native order, size & alignment (default)
 =: native order, std. size & alignment
 <: little-endian, std. size & alignment
 >: big-endian, std. size & alignment
 !: same as >

The remaining chars indicate types of args and must match exactly;
these can be preceded by a decimal repeat count:
   x: pad byte (no data);
   c:char;
   b:signed byte;
   B:unsigned byte;
   h:short;
   H:unsigned short;
   i:int;
   I:unsigned int;
   l:long;
   L:unsigned long;
   f:float;
   d:double.
Special cases (preceding decimal count indicates length):
   s:string (array of char); p: pascal string (with count byte).
Special case (only available in native format):
   P:an integer type that is wide enough to hold a pointer.
Special case (not in native mode unless 'long long' in platform C):
   q:long long;
   Q:unsigned long long
Whitespace between formats is ignored.

The variable struct.error is an exception raised on errors."""

import math
import sys

# TODO: XXX Find a way to get information on native sizes and alignments


class StructError(Exception):
  pass
error = StructError

bytes = str


def unpack_int(data, index, size, le):
  _bytes = [b for b in data[index:index + size]]
  if le == 'little':
    _bytes.reverse()
  number = 0
  for b in _bytes:
    number = number << 8 | b
  return int(number)


def unpack_signed_int(data, index, size, le):
  number = unpack_int(data, index, size, le)
  max = (1 << (size * 8))
  if number > (1 << (size * 8 - 1)) - 1:
    number = int(-1 * (max - number))
  return number

INFINITY = 1e200 * 1e200
NAN = INFINITY / INFINITY


def unpack_char(data, index, size, le):
  return data[index:index + size]


def pack_int(number, size, le):
  x = number
  res = []
  for i in range(size):
    res.append(x & 0xff)
    x = x >> 8
  if le == 'big':
    res.reverse()
  return ''.join(chr(x) for x in res)


def pack_signed_int(number, size, le):
  if not isinstance(number, int):
    raise StructError("argument for i,I,l,L,q,Q,h,H must be integer")
  if number > (1 << (8 * size - 1)) - 1 or number < -1 * (1 << (8 * size - 1)):
    raise OverflowError("Number:%i too large to convert" % number)
  return pack_int(number, size, le)


def pack_unsigned_int(number, size, le):
  if not isinstance(number, int):
    raise StructError("argument for i,I,l,L,q,Q,h,H must be integer")
  if number < 0:
    raise TypeError("can't convert negative long to unsigned")
  if number > (1 << (8 * size)) - 1:
    raise OverflowError("Number:%i too large to convert" % number)
  return pack_int(number, size, le)


def pack_char(char, size, le):
  return str(char)


def isinf(x):
  return x != 0.0 and x / 2 == x


def isnan(v):
  return v != v * 1.0 or (v == 1.0 and v == 2.0)


def pack_float(x, size, le):
  unsigned = float_pack(x, size)
  result = []
  for i in range(8):
    result.append((unsigned >> (i * 8)) & 0xFF)
  if le == "big":
    result.reverse()
  return ''.join(chr(x) for x in result)


def unpack_float(data, index, size, le):
  binary = [data[i] for i in range(index, index + 8)]
  if le == "big":
    binary.reverse()
  unsigned = 0
  for i in range(8):
    # unsigned |= binary[i] << (i * 8)
    unsigned |= ord(binary[i]) << (i * 8)
  return float_unpack(unsigned, size, le)


def round_to_nearest(x):
  """Python 3 style round:  round a float x to the nearest int, but
  unlike the builtin Python 2.x round function:

    - return an int, not a float
    - do round-half-to-even, not round-half-away-from-zero.

  We assume that x is finite and nonnegative; except wrong results
  if you use this for negative x.

  """
  int_part = int(x)
  frac_part = x - int_part
  if frac_part > 0.5 or frac_part == 0.5 and int_part & 1 == 1:
    int_part += 1
  return int_part


def float_unpack(Q, size, le):
  """Convert a 32-bit or 64-bit integer created
  by float_pack into a Python float."""

  if size == 8:
    MIN_EXP = -1021  # = sys.float_info.min_exp
    MAX_EXP = 1024   # = sys.float_info.max_exp
    MANT_DIG = 53    # = sys.float_info.mant_dig
    BITS = 64
  elif size == 4:
    MIN_EXP = -125   # C's FLT_MIN_EXP
    MAX_EXP = 128    # FLT_MAX_EXP
    MANT_DIG = 24    # FLT_MANT_DIG
    BITS = 32
  else:
    raise ValueError("invalid size value")

  if Q >> BITS:
    raise ValueError("input out of range")

  # extract pieces
  sign = Q >> BITS - 1
  exp = (Q & ((1 << BITS - 1) - (1 << MANT_DIG - 1))) >> MANT_DIG - 1
  mant = Q & ((1 << MANT_DIG - 1) - 1)

  if exp == MAX_EXP - MIN_EXP + 2:
    # nan or infinity
    result = float('nan') if mant else float('inf')
  elif exp == 0:
    # subnormal or zero
    result = math.ldexp(float(mant), MIN_EXP - MANT_DIG)
  else:
    # normal
    mant += 1 << MANT_DIG - 1
    result = math.ldexp(float(mant), exp + MIN_EXP - MANT_DIG - 1)
  return -result if sign else result


def float_pack(x, size):
  """Convert a Python float x into a 64-bit unsigned integer
  with the same byte representation."""

  if size == 8:
    MIN_EXP = -1021  # = sys.float_info.min_exp
    MAX_EXP = 1024   # = sys.float_info.max_exp
    MANT_DIG = 53    # = sys.float_info.mant_dig
    BITS = 64
  elif size == 4:
    MIN_EXP = -125   # C's FLT_MIN_EXP
    MAX_EXP = 128    # FLT_MAX_EXP
    MANT_DIG = 24    # FLT_MANT_DIG
    BITS = 32
  else:
    raise ValueError("invalid size value")

  sign = math.copysign(1.0, x) < 0.0
  if math.isinf(x):
    mant = 0
    exp = MAX_EXP - MIN_EXP + 2
  elif math.isnan(x):
    mant = 1 << (MANT_DIG - 2)  # other values possible
    exp = MAX_EXP - MIN_EXP + 2
  elif x == 0.0:
    mant = 0
    exp = 0
  else:
    m, e = math.frexp(abs(x))  # abs(x) == m * 2**e
    exp = e - (MIN_EXP - 1)
    if exp > 0:
      # Normal case.
      mant = round_to_nearest(m * (1 << MANT_DIG))
      mant -= 1 << MANT_DIG - 1
    else:
      # Subnormal case.
      if exp + MANT_DIG - 1 >= 0:
        mant = round_to_nearest(m * (1 << exp + MANT_DIG - 1))
      else:
        mant = 0
      exp = 0

    # Special case: rounding produced a MANT_DIG-bit mantissa.
    assert 0 <= mant <= 1 << MANT_DIG - 1
    if mant == 1 << MANT_DIG - 1:
      mant = 0
      exp += 1

    # Raise on overflow (in some circumstances, may want to return
    # infinity instead).
    if exp >= MAX_EXP - MIN_EXP + 2:
      raise OverflowError("float too large to pack in this format")

  # check constraints
  assert 0 <= mant < 1 << MANT_DIG - 1
  assert 0 <= exp <= MAX_EXP - MIN_EXP + 2
  assert 0 <= sign <= 1
  return ((sign << BITS - 1) | (exp << MANT_DIG - 1)) | mant


big_endian_format = {
    'x': {'size': 1, 'alignment': 0, 'pack': None, 'unpack': None},
    'b': {'size': 1, 'alignment': 0, 'pack': pack_signed_int, 'unpack': unpack_signed_int},
    'B': {'size': 1, 'alignment': 0, 'pack': pack_unsigned_int, 'unpack': unpack_int},
    'c': {'size': 1, 'alignment': 0, 'pack': pack_char, 'unpack': unpack_char},
    's': {'size': 1, 'alignment': 0, 'pack': None, 'unpack': None},
    'p': {'size': 1, 'alignment': 0, 'pack': None, 'unpack': None},
    'h': {'size': 2, 'alignment': 0, 'pack': pack_signed_int, 'unpack': unpack_signed_int},
    'H': {'size': 2, 'alignment': 0, 'pack': pack_unsigned_int, 'unpack': unpack_int},
    'i': {'size': 4, 'alignment': 0, 'pack': pack_signed_int, 'unpack': unpack_signed_int},
    'I': {'size': 4, 'alignment': 0, 'pack': pack_unsigned_int, 'unpack': unpack_int},
    'l': {'size': 4, 'alignment': 0, 'pack': pack_signed_int, 'unpack': unpack_signed_int},
    'L': {'size': 4, 'alignment': 0, 'pack': pack_unsigned_int, 'unpack': unpack_int},
    'q': {'size': 8, 'alignment': 0, 'pack': pack_signed_int, 'unpack': unpack_signed_int},
    'Q': {'size': 8, 'alignment': 0, 'pack': pack_unsigned_int, 'unpack': unpack_int},
    'f': {'size': 4, 'alignment': 0, 'pack': pack_float, 'unpack': unpack_float},
    'd': {'size': 8, 'alignment': 0, 'pack': pack_float, 'unpack': unpack_float},
}
default = big_endian_format
formatmode = {'<': (default, 'little'),
              '>': (default, 'big'),
              '!': (default, 'big'),
              '=': (default, sys.byteorder),
              '@': (default, sys.byteorder)
              }


def getmode(fmt):
  try:
    formatdef, endianness = formatmode[fmt[0]]
    index = 1
  except (IndexError, KeyError):
    formatdef, endianness = formatmode['@']
    index = 0
  return formatdef, endianness, index


def getNum(fmt, i):
  num = None
  cur = fmt[i]
  while ('0' <= cur) and (cur <= '9'):
    if num == None:
      num = int(cur)
    else:
      num = 10 * num + int(cur)
    i += 1
    cur = fmt[i]
  return num, i


def calcsize(fmt):
  """calcsize(fmt) -> int
  Return size of C struct described by format string fmt.
  See struct.__doc__ for more on format strings."""

  formatdef, endianness, i = getmode(fmt)
  num = 0
  result = 0
  while i < len(fmt):
    num, i = getNum(fmt, i)
    cur = fmt[i]
    try:
      format = formatdef[cur]
    except KeyError:
      raise StructError("%s is not a valid format" % cur)
    if num != None:
      result += num * format['size']
    else:
      result += format['size']
    num = 0
    i += 1
  return result


def pack(fmt, *args):
  """pack(fmt, v1, v2, ...) -> string
     Return string containing values v1, v2, ... packed according to fmt.
     See struct.__doc__ for more on format strings."""
  formatdef, endianness, i = getmode(fmt)
  args = list(args)
  n_args = len(args)
  result = []
  while i < len(fmt):
    num, i = getNum(fmt, i)
    cur = fmt[i]
    try:
      format = formatdef[cur]
    except KeyError:
      raise StructError("%s is not a valid format" % cur)
    if num == None:
      num_s = 0
      num = 1
    else:
      num_s = num

    if cur == 'x':
      result += [b'\0' * num]
    elif cur == 's':
      if isinstance(args[0], bytes):
        padding = num - len(args[0])
        result += [args[0][:num] + b'\0' * padding]
        args.pop(0)
      else:
        raise StructError("arg for string format not a string")
    elif cur == 'p':
      if isinstance(args[0], bytes):
        padding = num - len(args[0]) - 1

        if padding > 0:
          result += [bytes([len(args[0])]) + args[0]
                     [:num - 1] + b'\0' * padding]
        else:
          if num < 255:
            result += [bytes([num - 1]) + args[0][:num - 1]]
          else:
            result += [bytes([255]) + args[0][:num - 1]]
        args.pop(0)
      else:
        raise StructError("arg for string format not a string")

    else:
      if len(args) < num:
        raise StructError("insufficient arguments to pack")
      for var in args[:num]:
        result += [format['pack'](var, format['size'], endianness)]
      args = args[num:]
    num = None
    i += 1
  if len(args) != 0:
    raise StructError("too many arguments for pack format")
  return b''.join(result)


def unpack(fmt, data):
  """unpack(fmt, string) -> (v1, v2, ...)
     Unpack the string, containing packed C structure data, according
     to fmt.  Requires len(string)==calcsize(fmt).
     See struct.__doc__ for more on format strings."""
  formatdef, endianness, i = getmode(fmt)
  j = 0
  num = 0
  result = []
  length = calcsize(fmt)
  if length != len(data):
    raise StructError("unpack str size does not match format")
  while i < len(fmt):
    num, i = getNum(fmt, i)
    cur = fmt[i]
    i += 1
    try:
      format = formatdef[cur]
    except KeyError:
      raise StructError("%s is not a valid format" % cur)

    if not num:
      num = 1

    if cur == 'x':
      j += num
    elif cur == 's':
      result.append(data[j:j + num])
      j += num
    elif cur == 'p':
      n = data[j]
      if n >= num:
        n = num - 1
      result.append(data[j + 1:j + n + 1])
      j += num
    else:
      for n in range(num):
        result += [format['unpack'](data, j, format['size'], endianness)]
        j += format['size']

  return tuple(result)


def pack_into(fmt, buf, offset, *args):
  data = pack(fmt, *args)
  buffer(buf)[offset:offset + len(data)] = data


def unpack_from(fmt, buf, offset=0):
  size = calcsize(fmt)
  data = buffer(buf)[offset:offset + size]
  if len(data) != size:
    raise error("unpack_from requires a buffer of at least %d bytes"
                % (size,))
  return unpack(fmt, data)


def _clearcache():
  "Clear the internal cache."
  # No cache in this implementation

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

"""Generate pseudo random numbers. Should not be used for security purposes."""

from __go__.math.rand import Uint32, Seed
from __go__.math import Pow as _pow
from __go__.time import Now


BPF = 53  # Number of bits in a float
RECIP_BPF = _pow(2, -BPF)


def _gorandom(nbytes):
  byte_arr = []
  while len(byte_arr) < nbytes:
    i = Uint32()
    byte_arr.append(i & 0xff)
    byte_arr.append(i >> 8 & 0xff)
    byte_arr.append(i >> 16 & 0xff)
    byte_arr.append(i >> 24 & 0xff)
  byte_arr = byte_arr[0:nbytes]
  return byte_arr


def _notimplemented(*args, **kwargs):
  raise NotImplementedError


# This is a slow replacement for int.bit_length.
# We should stop using this if it is implemented.
def _int_bit_length(n):
  bits = 0
  while n:  # 1 bit steps
    n = n / 2
    bits += 1
  return bits


# Replacement for int.from_bytes (big endian)
def _int_from_bytes(bytes):
  i = 0
  n = len(bytes) - 1
  while n >= 0:
    i += bytes[n] << (8 * n)
    n -= 1
  return i


class GrumpyRandom(object):
  """Alternate random number generator using math.rand.read as a replacement
  for urandom. Implemented like SystemRandom from cPythons random.py lib.
  """

  def random(self):
    """Get the next random number in the range [0.0, 1.0)."""
    return (_int_from_bytes(_gorandom(7)) >> 3) * RECIP_BPF

  def getrandbits(self, k):
    """getrandbits(k) -> x.  Generates an int with k random bits."""
    if k <= 0:
      raise ValueError('number of bits must be greater than zero')
    if k != int(k):
      raise TypeError('number of bits should be an integer')
    numbytes = (k + 7) // 8                       # bits / 8 and rounded up
    x = _int_from_bytes(_gorandom(numbytes))
    return x >> (numbytes * 8 - k)                # trim excess bits

  def seed(self, a=None):
    """Seed the golang.math.rand generator."""
    if a is None:
      a = Now().UnixNano()
    Seed(a)

  def _randbelow(self, n):
    """Return a random int in the range [0,n)."""
    # change once int.bit_length is implemented.
    # k = n.bit_length()
    k = _int_bit_length(n)
    r = getrandbits(k)
    while r >= n:
      r = getrandbits(k)
    return r

  def getstate(self):
    raise NotImplementedError('Entropy source does not have state.')

  def setstate(self):
    raise NotImplementedError('Entropy source does not have state.')


# Most of the following code is taken from the cpython std lib.
# Source Lib/random.py
class Random(GrumpyRandom):
  def randrange(self, start, stop=None, step=1, _int=int):
    """Choose a random item from range(start, stop[, step]).
    This fixes the problem with randint() which includes the
    endpoint; in Python this is usually not what you want.
    """
    # This code is a bit messy to make it fast for the
    # common case while still doing adequate error checking.
    istart = _int(start)
    if istart != start:
      raise ValueError("non-integer arg 1 for randrange()")
    if stop is None:
      if istart > 0:
        return self._randbelow(istart)
      raise ValueError("empty range for randrange()")

    # stop argument supplied.
    istop = _int(stop)
    if istop != stop:
      raise ValueError("non-integer stop for randrange()")
    width = istop - istart
    if step == 1 and width > 0:
      return istart + self._randbelow(width)
    if step == 1:
      raise ValueError("empty range for randrange() (%d,%d, %d)" %
                       (istart, istop, width))

    # Non-unit step argument supplied.
    istep = _int(step)
    if istep != step:
      raise ValueError("non-integer step for randrange()")
    if istep > 0:
      n = (width + istep - 1) // istep
    elif istep < 0:
      n = (width + istep + 1) // istep
    else:
      raise ValueError("zero step for randrange()")

    if n <= 0:
      raise ValueError("empty range for randrange()")

    return istart + istep*self._randbelow(n)

  def randint(self, a, b):
    """Return random integer in range [a, b], including both end points.
    """
    return self.randrange(a, b+1)

  def choice(self, seq):
    """Choose a random element from a non-empty sequence."""
    try:
      i = self._randbelow(len(seq))
    except ValueError:
      raise IndexError('Cannot choose from an empty sequence')
    return seq[i]


_inst = Random()
seed = _inst.seed
random = _inst.random
randint = _inst.randint
choice = _inst.choice
randrange = _inst.randrange
getrandbits = _inst.getrandbits
getstate = _inst.getstate
setstate = _inst.setstate


shuffle = _notimplemented
choices = _notimplemented
sample = _notimplemented
uniform = _notimplemented
triangular = _notimplemented
normalvariate = _notimplemented
lognormvariate = _notimplemented
expovariate = _notimplemented
vonmisesvariate = _notimplemented
gammavariate = _notimplemented
gauss = _notimplemented
betavariate = _notimplemented
paretovariate = _notimplemented
weibullvariate = _notimplemented

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

from '__go__/math/rand' import Uint32, Seed
from '__go__/math' import Pow
from '__go__/time' import Now


BPF = 53  # Number of bits in a float
RECIP_BPF = Pow(2, -BPF)


# TODO: The random byte generator currently uses math.rand.Uint32 to generate
# 4 bytes at a time. We should use math.rand.Read to generate the correct
# number of bytes needed. This can be changed once there is a way to
# allocate the needed []byte for Read from python and cast it to a list of
# integers once it is filled.
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
  """Random generator replacement for Grumpy.

  Alternate random number generator using golangs math.rand as a replacement
  for the CPython implementation.
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
    # TODO
    # change once int.bit_length is implemented.
    # k = n.bit_length()
    k = _int_bit_length(n)
    r = self.getrandbits(k)
    while r >= n:
      r = self.getrandbits(k)
    return r

  def getstate(self, *args, **kwargs):
    raise NotImplementedError('Entropy source does not have state.')

  def setstate(self, *args, **kwargs):
    raise NotImplementedError('Entropy source does not have state.')

  def jumpahead(self, *args, **kwargs):
    raise NotImplementedError('Entropy source does not have state.')

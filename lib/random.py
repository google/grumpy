"""Generate pseudo random numbers. Should not be used for security purposes"""

from __go__.math.rand import Float64, Seed
from __go__.time import Now


# default seed with time like in cpython
Seed(Now().UnixNano())


def random():
  return Float64()


def randint(a, b):
  if int(a) != a or int(b) != b:
    raise ValueError("non-integer for randrange()")

  r = (b - a) + 1

  if r < 1:
    raise ValueError("empty range for randrange()")

  return int(a) + int(random() * r)


def choice(seq):
  return seq[int(random() * len(seq))]

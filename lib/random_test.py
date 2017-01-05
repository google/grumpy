import random

import weetest


def TestRandom():
  a = random.random()
  b = random.random()
  c = random.random()
  assert isinstance(a, float)
  assert 0.0 <= a < 1.0
  assert not a == b == c


def TestRandomInt():
  for _ in range(10):
    a = random.randint(0, 5)
    assert isinstance(a, int)
    assert 0 <= a <= 5

  b = random.randint(1, 1)
  assert b == 1

  try:
    c = random.randint(0.1, 3)
  except ValueError:
    pass
  else:
    raise AssertionError("ValueError not raised")

  try:
    d = random.randint(4, 3)
  except ValueError:
    pass
  else:
    raise AssertionError("ValueError not raised")


def TestRandomChoice():
  seq = [i*2 for i in range(5)]
  for i in range(10):
    item = random.choice(seq)
    item_idx = item/2
    assert seq[item_idx] == item

  try:
    random.choice([])
  except IndexError:
    pass
  else:
    raise AssertionError("IndexError not raised")

if __name__ == '__main__':
  weetest.RunTests()
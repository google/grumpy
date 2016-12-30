def eq(a, b):
  return a == b


def le(a, b):
  return a <= b


def lt(a, b):
  return a < b


def ge(a, b):
  return a >= b


def gt(a, b):
  return a > b


def itemgetter(*items):
  if len(items) == 1:
    item = items[0]
    def g(obj):
      return obj[item]
  else:
    def g(obj):
      return tuple(obj[item] for item in items)
  return g


def ne(a, b):
  return a != b

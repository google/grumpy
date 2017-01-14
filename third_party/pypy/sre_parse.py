#
# Secret Labs' Regular Expression Engine
#
# various symbols used by the regular expression engine.
# run this script to update the _sre include files!
#
# Copyright (c) 1998-2001 by Secret Labs AB.  All rights reserved.
#
# See the sre.py file for information on usage and redistribution.
#

"""Internal support module for sre"""

# update when constants are added or removed

MAGIC = 20031017

MAXREPEAT = 2147483648
#from _sre import MAXREPEAT

# SRE standard exception (access as sre.error)
# should this really be here?

class error(Exception):
    pass

# operators

FAILURE = "failure"
SUCCESS = "success"

ANY = "any"
ANY_ALL = "any_all"
ASSERT = "assert"
ASSERT_NOT = "assert_not"
AT = "at"
BIGCHARSET = "bigcharset"
BRANCH = "branch"
CALL = "call"
CATEGORY = "category"
CHARSET = "charset"
GROUPREF = "groupref"
GROUPREF_IGNORE = "groupref_ignore"
GROUPREF_EXISTS = "groupref_exists"
IN = "in"
IN_IGNORE = "in_ignore"
INFO = "info"
JUMP = "jump"
LITERAL = "literal"
LITERAL_IGNORE = "literal_ignore"
MARK = "mark"
MAX_REPEAT = "max_repeat"
MAX_UNTIL = "max_until"
MIN_REPEAT = "min_repeat"
MIN_UNTIL = "min_until"
NEGATE = "negate"
NOT_LITERAL = "not_literal"
NOT_LITERAL_IGNORE = "not_literal_ignore"
RANGE = "range"
REPEAT = "repeat"
REPEAT_ONE = "repeat_one"
SUBPATTERN = "subpattern"
MIN_REPEAT_ONE = "min_repeat_one"

# positions
AT_BEGINNING = "at_beginning"
AT_BEGINNING_LINE = "at_beginning_line"
AT_BEGINNING_STRING = "at_beginning_string"
AT_BOUNDARY = "at_boundary"
AT_NON_BOUNDARY = "at_non_boundary"
AT_END = "at_end"
AT_END_LINE = "at_end_line"
AT_END_STRING = "at_end_string"
AT_LOC_BOUNDARY = "at_loc_boundary"
AT_LOC_NON_BOUNDARY = "at_loc_non_boundary"
AT_UNI_BOUNDARY = "at_uni_boundary"
AT_UNI_NON_BOUNDARY = "at_uni_non_boundary"

# categories
CATEGORY_DIGIT = "category_digit"
CATEGORY_NOT_DIGIT = "category_not_digit"
CATEGORY_SPACE = "category_space"
CATEGORY_NOT_SPACE = "category_not_space"
CATEGORY_WORD = "category_word"
CATEGORY_NOT_WORD = "category_not_word"
CATEGORY_LINEBREAK = "category_linebreak"
CATEGORY_NOT_LINEBREAK = "category_not_linebreak"
CATEGORY_LOC_WORD = "category_loc_word"
CATEGORY_LOC_NOT_WORD = "category_loc_not_word"
CATEGORY_UNI_DIGIT = "category_uni_digit"
CATEGORY_UNI_NOT_DIGIT = "category_uni_not_digit"
CATEGORY_UNI_SPACE = "category_uni_space"
CATEGORY_UNI_NOT_SPACE = "category_uni_not_space"
CATEGORY_UNI_WORD = "category_uni_word"
CATEGORY_UNI_NOT_WORD = "category_uni_not_word"
CATEGORY_UNI_LINEBREAK = "category_uni_linebreak"
CATEGORY_UNI_NOT_LINEBREAK = "category_uni_not_linebreak"

OPCODES = [

    # failure=0 success=1 (just because it looks better that way :-)
    FAILURE, SUCCESS,

    ANY, ANY_ALL,
    ASSERT, ASSERT_NOT,
    AT,
    BRANCH,
    CALL,
    CATEGORY,
    CHARSET, BIGCHARSET,
    GROUPREF, GROUPREF_EXISTS, GROUPREF_IGNORE,
    IN, IN_IGNORE,
    INFO,
    JUMP,
    LITERAL, LITERAL_IGNORE,
    MARK,
    MAX_UNTIL,
    MIN_UNTIL,
    NOT_LITERAL, NOT_LITERAL_IGNORE,
    NEGATE,
    RANGE,
    REPEAT,
    REPEAT_ONE,
    SUBPATTERN,
    MIN_REPEAT_ONE

]

ATCODES = [
    AT_BEGINNING, AT_BEGINNING_LINE, AT_BEGINNING_STRING, AT_BOUNDARY,
    AT_NON_BOUNDARY, AT_END, AT_END_LINE, AT_END_STRING,
    AT_LOC_BOUNDARY, AT_LOC_NON_BOUNDARY, AT_UNI_BOUNDARY,
    AT_UNI_NON_BOUNDARY
]

CHCODES = [
    CATEGORY_DIGIT, CATEGORY_NOT_DIGIT, CATEGORY_SPACE,
    CATEGORY_NOT_SPACE, CATEGORY_WORD, CATEGORY_NOT_WORD,
    CATEGORY_LINEBREAK, CATEGORY_NOT_LINEBREAK, CATEGORY_LOC_WORD,
    CATEGORY_LOC_NOT_WORD, CATEGORY_UNI_DIGIT, CATEGORY_UNI_NOT_DIGIT,
    CATEGORY_UNI_SPACE, CATEGORY_UNI_NOT_SPACE, CATEGORY_UNI_WORD,
    CATEGORY_UNI_NOT_WORD, CATEGORY_UNI_LINEBREAK,
    CATEGORY_UNI_NOT_LINEBREAK
]

# def makedict(list):
#     d = {}
#     i = 0
#     for item in list:
#         d[item] = i
#         i = i + 1
#     return d

# OPCODES = makedict(OPCODES)
# ATCODES = makedict(ATCODES)
# CHCODES = makedict(CHCODES)
ATCODES = {'at_beginning_string': 2, 'at_uni_non_boundary': 11, 'at_uni_boundary': 10, 'at_non_boundary': 4, 'at_loc_non_boundary': 9, 'at_beginning': 0, 'at_end_string': 7, 'at_end': 5, 'at_end_line': 6, 'at_loc_boundary': 8, 'at_boundary': 3, 'at_beginning_line': 1}
CHCODES = {'category_uni_space': 12, 'category_uni_digit': 10, 'category_space': 2, 'category_not_digit': 1, 'category_not_space': 3, 'category_digit': 0, 'category_uni_linebreak': 16, 'category_not_word': 5, 'category_loc_not_word': 9, 'category_word': 4, 'category_uni_word': 14, 'category_uni_not_space': 13, 'category_uni_not_word': 15, 'category_not_linebreak': 7, 'category_loc_word': 8, 'category_linebreak': 6, 'category_uni_not_linebreak': 17, 'category_uni_not_digit': 11}
OPCODES = {'min_until': 23, 'bigcharset': 11, 'min_repeat_one': 31, 'jump': 18, 'at': 6, 'in': 15, 'negate': 26, 'any': 2, 'in_ignore': 16, 'category': 9, 'subpattern': 30, 'charset': 10, 'range': 27, 'max_until': 22, 'mark': 21, 'literal': 19, 'call': 8, 'branch': 7, 'repeat': 28, 'assert': 4, 'failure': 0, 'any_all': 3, 'not_literal_ignore': 25, 'info': 17, 'groupref_ignore': 14, 'groupref_exists': 13, 'success': 1, 'not_literal': 24, 'groupref': 12, 'assert_not': 5, 'repeat_one': 29, 'literal_ignore': 20}

# replacement operations for "ignore case" mode
OP_IGNORE = {
    GROUPREF: GROUPREF_IGNORE,
    IN: IN_IGNORE,
    LITERAL: LITERAL_IGNORE,
    NOT_LITERAL: NOT_LITERAL_IGNORE
}

AT_MULTILINE = {
    AT_BEGINNING: AT_BEGINNING_LINE,
    AT_END: AT_END_LINE
}

AT_LOCALE = {
    AT_BOUNDARY: AT_LOC_BOUNDARY,
    AT_NON_BOUNDARY: AT_LOC_NON_BOUNDARY
}

AT_UNICODE = {
    AT_BOUNDARY: AT_UNI_BOUNDARY,
    AT_NON_BOUNDARY: AT_UNI_NON_BOUNDARY
}

CH_LOCALE = {
    CATEGORY_DIGIT: CATEGORY_DIGIT,
    CATEGORY_NOT_DIGIT: CATEGORY_NOT_DIGIT,
    CATEGORY_SPACE: CATEGORY_SPACE,
    CATEGORY_NOT_SPACE: CATEGORY_NOT_SPACE,
    CATEGORY_WORD: CATEGORY_LOC_WORD,
    CATEGORY_NOT_WORD: CATEGORY_LOC_NOT_WORD,
    CATEGORY_LINEBREAK: CATEGORY_LINEBREAK,
    CATEGORY_NOT_LINEBREAK: CATEGORY_NOT_LINEBREAK
}

CH_UNICODE = {
    CATEGORY_DIGIT: CATEGORY_UNI_DIGIT,
    CATEGORY_NOT_DIGIT: CATEGORY_UNI_NOT_DIGIT,
    CATEGORY_SPACE: CATEGORY_UNI_SPACE,
    CATEGORY_NOT_SPACE: CATEGORY_UNI_NOT_SPACE,
    CATEGORY_WORD: CATEGORY_UNI_WORD,
    CATEGORY_NOT_WORD: CATEGORY_UNI_NOT_WORD,
    CATEGORY_LINEBREAK: CATEGORY_UNI_LINEBREAK,
    CATEGORY_NOT_LINEBREAK: CATEGORY_UNI_NOT_LINEBREAK
}

# flags
SRE_FLAG_TEMPLATE = 1 # template mode (disable backtracking)
SRE_FLAG_IGNORECASE = 2 # case insensitive
SRE_FLAG_LOCALE = 4 # honour system locale
SRE_FLAG_MULTILINE = 8 # treat target as multiline string
SRE_FLAG_DOTALL = 16 # treat target as a single string
SRE_FLAG_UNICODE = 32 # use unicode "locale"
SRE_FLAG_VERBOSE = 64 # ignore whitespace and comments
SRE_FLAG_DEBUG = 128 # debugging
SRE_FLAG_ASCII = 256 # use ascii "locale"

# flags for INFO primitive
SRE_INFO_PREFIX = 1 # has prefix
SRE_INFO_LITERAL = 2 # entire pattern is literal (given by prefix)
SRE_INFO_CHARSET = 4 # pattern starts with character from given set

#
# Secret Labs' Regular Expression Engine
#
# convert re-style regular expression to sre pattern
#
# Copyright (c) 1998-2001 by Secret Labs AB.  All rights reserved.
#
# See the sre.py file for information on usage and redistribution.
#

"""Internal support module for sre"""

# XXX: show string offset and offending character for all errors

import sys

SPECIAL_CHARS = ".\\[{()*+?^$|"
REPEAT_CHARS = "*+?{"

DIGITS = set("0123456789")

OCTDIGITS = set("01234567")
HEXDIGITS = set("0123456789abcdefABCDEF")

WHITESPACE = set(" \t\n\r\v\f")

ESCAPES = {
    r"\a": (LITERAL, ord("\a")),
    r"\b": (LITERAL, ord("\b")),
    r"\f": (LITERAL, ord("\f")),
    r"\n": (LITERAL, ord("\n")),
    r"\r": (LITERAL, ord("\r")),
    r"\t": (LITERAL, ord("\t")),
    r"\v": (LITERAL, ord("\v")),
    r"\\": (LITERAL, ord("\\"))
}

CATEGORIES = {
    r"\A": (AT, AT_BEGINNING_STRING),  # start of string
    r"\b": (AT, AT_BOUNDARY),
    r"\B": (AT, AT_NON_BOUNDARY),
    r"\d": (IN, [(CATEGORY, CATEGORY_DIGIT)]),
    r"\D": (IN, [(CATEGORY, CATEGORY_NOT_DIGIT)]),
    r"\s": (IN, [(CATEGORY, CATEGORY_SPACE)]),
    r"\S": (IN, [(CATEGORY, CATEGORY_NOT_SPACE)]),
    r"\w": (IN, [(CATEGORY, CATEGORY_WORD)]),
    r"\W": (IN, [(CATEGORY, CATEGORY_NOT_WORD)]),
    r"\Z": (AT, AT_END_STRING),  # end of string
}

FLAGS = {
    # standard flags
    "i": SRE_FLAG_IGNORECASE,
    "L": SRE_FLAG_LOCALE,
    "m": SRE_FLAG_MULTILINE,
    "s": SRE_FLAG_DOTALL,
    "x": SRE_FLAG_VERBOSE,
    # extensions
    "a": SRE_FLAG_ASCII,
    "t": SRE_FLAG_TEMPLATE,
    "u": SRE_FLAG_UNICODE,
}


class Pattern(object):
  # master pattern object.  keeps track of global attributes

  def __init__(self):
    self.flags = 0
    self.open = []
    self.groups = 1
    self.groupdict = {}

  def opengroup(self, name=None):
    gid = self.groups
    self.groups = gid + 1
    if name is not None:
      ogid = self.groupdict.get(name, None)
      if ogid is not None:
        raise ("redefinition of group name %s as group %d; "
                    "was group %d" % (repr(name), gid,  ogid))
      self.groupdict[name] = gid
    self.open.append(gid)
    return gid

  def closegroup(self, gid):
    self.open.remove(gid)

  def checkgroup(self, gid):
    return gid < self.groups and gid not in self.open


class SubPattern(object):
  # a subpattern, in intermediate form

  def __init__(self, pattern, data=None):
    self.pattern = pattern
    if data is None:
      data = []
    self.data = data
    self.width = None

  def __iter__(self):
    return iter(self.data)

  def dump(self, level=0):
    nl = 1
    seqtypes = (tuple, list)
    for op, av in self.data:
      print level * "  " + op,
      nl = 0
      if op == "in":
        # member sublanguage
        print()
        nl = 1
        for op, a in av:
          print((level + 1) * "  " + op, a)
      elif op == "branch":
        print()
        nl = 1
        i = 0
        for a in av[1]:
          if i > 0:
            print(level * "  " + "or")
          a.dump(level + 1)
          nl = 1
          i = i + 1
      elif isinstance(av, seqtypes):
        for a in av:
          if isinstance(a, SubPattern):
            if not nl:
              print()
            a.dump(level + 1)
            nl = 1
          else:
            print a,
            nl = 0
      else:
        print av,
        nl = 0
      if not nl:
        print()

  def __repr__(self):
    return repr(self.data)

  def __len__(self):
    return len(self.data)

  def __delitem__(self, index):
    del self.data[index]

  def __getitem__(self, index):
    if isinstance(index, slice):
      return SubPattern(self.pattern, self.data[index])
    return self.data[index]

  def __setitem__(self, index, code):
    self.data[index] = code

  def insert(self, index, code):
    self.data.insert(index, code)

  def append(self, code):
    self.data.append(code)

  def getwidth(self):
    # determine the width (min, max) for this subpattern
    if self.width:
      return self.width
    lo = hi = 0
    UNITCODES = (ANY, RANGE, IN, LITERAL, NOT_LITERAL, CATEGORY)
    REPEATCODES = (MIN_REPEAT, MAX_REPEAT)
    for op, av in self.data:
      if op is BRANCH:
        i = sys.maxsize
        j = 0
        for av in av[1]:
          l, h = av.getwidth()
          i = min(i, l)
          j = max(j, h)
        lo = lo + i
        hi = hi + j
      elif op is CALL:
        i, j = av.getwidth()
        lo = lo + i
        hi = hi + j
      elif op is SUBPATTERN:
        i, j = av[1].getwidth()
        lo = lo + i
        hi = hi + j
      elif op in REPEATCODES:
        i, j = av[2].getwidth()
        lo = lo + int(i) * av[0]
        hi = hi + int(j) * av[1]
      elif op in UNITCODES:
        lo = lo + 1
        hi = hi + 1
      elif op == SUCCESS:
        break
    self.width = int(min(lo, sys.maxsize)), int(min(hi, sys.maxsize))
    return self.width


class Tokenizer(object):

  def __init__(self, string):
    self.istext = isinstance(string, str)
    self.string = string
    self.index = 0
    self.__next()

  def __next(self):
    if self.index >= len(self.string):
      self.next = None
      return
    char = self.string[self.index:self.index + 1]
    # Special case for the str8, since indexing returns a integer
    # XXX This is only needed for test_bug_926075 in test_re.py
    if char and not self.istext:
      char = chr(char[0])
    if char == "\\":
      try:
        c = self.string[self.index + 1]
      except IndexError:
        raise ("bogus escape (end of line)")
      if not self.istext:
        c = chr(c)
      char = char + c
    self.index = self.index + len(char)
    self.next = char

  def match(self, char, skip=1):
    if char == self.next:
      if skip:
        self.__next()
      return 1
    return 0

  def get(self):
    this = self.next
    self.__next()
    return this

  def getwhile(self, n, charset):
    result = ''
    for _ in range(n):
      c = self.next
      if c not in charset:
        break
      result += c
      self.__next()
    return result

  def tell(self):
    return self.index, self.next

  def seek(self, index):
    self.index, self.next = index


def isident(char):
  return "a" <= char <= "z" or "A" <= char <= "Z" or char == "_"


def isdigit(char):
  return "0" <= char <= "9"


def isname(name):
  # check that group name is a valid string
  if not isident(name[0]):
    return False
  for char in name[1:]:
    if not isident(char) and not isdigit(char):
      return False
  return True


def _class_escape(source, escape):
  # handle escape code inside character class
  code = ESCAPES.get(escape)
  if code:
    return code
  code = CATEGORIES.get(escape)
  if code and code[0] == IN:
    return code
  try:
    c = escape[1:2]
    if c == "x":
      # hexadecimal escape (exactly two digits)
      escape += source.getwhile(2, HEXDIGITS)
      if len(escape) != 4:
        raise ValueError
      return LITERAL, int(escape[2:], 16) & 0xff
    elif c == "u" and source.istext:
      # unicode escape (exactly four digits)
      escape += source.getwhile(4, HEXDIGITS)
      if len(escape) != 6:
        raise ValueError
      return LITERAL, int(escape[2:], 16)
    elif c == "U" and source.istext:
      # unicode escape (exactly eight digits)
      escape += source.getwhile(8, HEXDIGITS)
      if len(escape) != 10:
        raise ValueError
      c = int(escape[2:], 16)
      chr(c)  # raise ValueError for invalid code
      return LITERAL, c
    elif c in OCTDIGITS:
      # octal escape (up to three digits)
      escape += source.getwhile(2, OCTDIGITS)
      return LITERAL, int(escape[1:], 8) & 0xff
    elif c in DIGITS:
      raise ValueError
    if len(escape) == 2:
      return LITERAL, ord(escape[1])
  except ValueError:
    pass
  raise ("bogus escape: %s" % repr(escape))


def _escape(source, escape, state):
  # handle escape code in expression
  code = CATEGORIES.get(escape)
  if code:
    return code
  code = ESCAPES.get(escape)
  if code:
    return code
  try:
    c = escape[1:2]
    if c == "x":
      # hexadecimal escape
      escape += source.getwhile(2, HEXDIGITS)
      if len(escape) != 4:
        raise ValueError
      return LITERAL, int(escape[2:], 16) & 0xff
    elif c == "u" and source.istext:
      # unicode escape (exactly four digits)
      escape += source.getwhile(4, HEXDIGITS)
      if len(escape) != 6:
        raise ValueError
      return LITERAL, int(escape[2:], 16)
    elif c == "U" and source.istext:
      # unicode escape (exactly eight digits)
      escape += source.getwhile(8, HEXDIGITS)
      if len(escape) != 10:
        raise ValueError
      c = int(escape[2:], 16)
      chr(c)  # raise ValueError for invalid code
      return LITERAL, c
    elif c == "0":
      # octal escape
      escape += source.getwhile(2, OCTDIGITS)
      return LITERAL, int(escape[1:], 8) & 0xff
    elif c in DIGITS:
      # octal escape *or* decimal group reference (sigh)
      if source.next in DIGITS:
        escape = escape + source.get()
        if (escape[1] in OCTDIGITS and escape[2] in OCTDIGITS and
                source.next in OCTDIGITS):
          # got three octal digits; this is an octal escape
          escape = escape + source.get()
          return LITERAL, int(escape[1:], 8) & 0xff
      # not an octal escape, so this is a group reference
      group = int(escape[1:])
      if group < state.groups:
        if not state.checkgroup(group):
          raise ("cannot refer to open group")
        return GROUPREF, group
      raise ValueError
    if len(escape) == 2:
      return LITERAL, ord(escape[1])
  except ValueError:
    pass
  raise ("bogus escape: %s" % repr(escape))


def _parse_sub(source, state, nested=1):
  # parse an alternation: a|b|c

  items = []
  itemsappend = items.append
  sourcematch = source.match
  while 1:
    itemsappend(_parse(source, state))
    if sourcematch("|"):
      continue
    if not nested:
      break
    if not source.next or sourcematch(")", 0):
      break
    else:
      raise ("pattern not properly closed")

  if len(items) == 1:
    return items[0]

  subpattern = SubPattern(state)
  subpatternappend = subpattern.append

  # check if all items share a common prefix
  while 1:
    prefix = None
    for item in items:
      if not item:
        break
      if prefix is None:
        prefix = item[0]
      elif item[0] != prefix:
        break
    else:
      # all subitems start with a common "prefix".
      # move it out of the branch
      for item in items:
        del item[0]
      subpatternappend(prefix)
      continue  # check next one
    break

  # check if the branch can be replaced by a character set
  for item in items:
    if len(item) != 1 or item[0][0] != LITERAL:
      break
  else:
    # we can store this as a character set instead of a
    # branch (the compiler may optimize this even more)
    set = []
    setappend = set.append
    for item in items:
      setappend(item[0])
    subpatternappend((IN, set))
    return subpattern

  subpattern.append((BRANCH, (None, items)))
  return subpattern


def _parse_sub_cond(source, state, condgroup):
  item_yes = _parse(source, state)
  if source.match("|"):
    item_no = _parse(source, state)
    if source.match("|"):
      raise ("conditional backref with more than two branches")
  else:
    item_no = None
  if source.next and not source.match(")", 0):
    raise ("pattern not properly closed")
  subpattern = SubPattern(state)
  subpattern.append((GROUPREF_EXISTS, (condgroup, item_yes, item_no)))
  return subpattern

_PATTERNENDERS = set("|)")
_ASSERTCHARS = set("=!<")
_LOOKBEHINDASSERTCHARS = set("=!")
_REPEATCODES = set([MIN_REPEAT, MAX_REPEAT])


def _parse(source, state):
  # parse a simple pattern
  subpattern = SubPattern(state)

  # precompute constants into local variables
  subpatternappend = subpattern.append
  sourceget = source.get
  sourcematch = source.match
  _len = len
  PATTERNENDERS = _PATTERNENDERS
  ASSERTCHARS = _ASSERTCHARS
  LOOKBEHINDASSERTCHARS = _LOOKBEHINDASSERTCHARS
  REPEATCODES = _REPEATCODES

  while 1:

    if source.next in PATTERNENDERS:
      break  # end of subpattern
    this = sourceget()
    if this is None:
      break  # end of pattern

    if state.flags & SRE_FLAG_VERBOSE:
      # skip whitespace and comments
      if this in WHITESPACE:
        continue
      if this == "#":
        while 1:
          this = sourceget()
          if this in (None, "\n"):
            break
        continue

    if this and this[0] not in SPECIAL_CHARS:
      subpatternappend((LITERAL, ord(this)))

    elif this == "[":
      # character set
      set = []
      setappend = set.append
# if sourcematch(":"):
# pass # handle character classes
      if sourcematch("^"):
        setappend((NEGATE, None))
      # check remaining characters
      start = set[:]
      while 1:
        this = sourceget()
        if this == "]" and set != start:
          break
        elif this and this[0] == "\\":
          code1 = _class_escape(source, this)
        elif this:
          code1 = LITERAL, ord(this)
        else:
          raise ("unexpected end of regular expression")
        if sourcematch("-"):
          # potential range
          this = sourceget()
          if this == "]":
            if code1[0] is IN:
              code1 = code1[1][0]
            setappend(code1)
            setappend((LITERAL, ord("-")))
            break
          elif this:
            if this[0] == "\\":
              code2 = _class_escape(source, this)
            else:
              code2 = LITERAL, ord(this)
            if code1[0] != LITERAL or code2[0] != LITERAL:
              raise ("bad character range")
            lo = code1[1]
            hi = code2[1]
            if hi < lo:
              raise ("bad character range")
            setappend((RANGE, (lo, hi)))
          else:
            raise ("unexpected end of regular expression")
        else:
          if code1[0] is IN:
            code1 = code1[1][0]
          setappend(code1)

      # XXX: <fl> should move set optimization to compiler!
      if _len(set) == 1 and set[0][0] is LITERAL:
        subpatternappend(set[0])  # optimization
      elif _len(set) == 2 and set[0][0] is NEGATE and set[1][0] is LITERAL:
        subpatternappend((NOT_LITERAL, set[1][1]))  # optimization
      else:
        # XXX: <fl> should add charmap optimization here
        subpatternappend((IN, set))

    elif this and this[0] in REPEAT_CHARS:
      # repeat previous item
      if this == "?":
        min, max = 0, 1
      elif this == "*":
        min, max = 0, MAXREPEAT

      elif this == "+":
        min, max = 1, MAXREPEAT
      elif this == "{":
        if source.next == "}":
          subpatternappend((LITERAL, ord(this)))
          continue
        here = source.tell()
        min, max = 0, MAXREPEAT
        lo = hi = ""
        while source.next in DIGITS:
          lo = lo + source.get()
        if sourcematch(","):
          while source.next in DIGITS:
            hi = hi + sourceget()
        else:
          hi = lo
        if not sourcematch("}"):
          subpatternappend((LITERAL, ord(this)))
          source.seek(here)
          continue
        if lo:
          min = int(lo)
          if min >= MAXREPEAT:
            raise OverflowError("the repetition number is too large")
        if hi:
          max = int(hi)
          if max >= MAXREPEAT:
            raise OverflowError("the repetition number is too large")
          if max < min:
            raise ("bad repeat interval")
      else:
        raise ("not supported")
      # figure out which item to repeat
      if subpattern:
        item = subpattern[-1:]
      else:
        item = None
      if not item or (_len(item) == 1 and item[0][0] == AT):
        raise ("nothing to repeat")
      if item[0][0] in REPEATCODES:
        raise ("multiple repeat")
      if sourcematch("?"):
        subpattern[-1] = (MIN_REPEAT, (min, max, item))
      else:
        subpattern[-1] = (MAX_REPEAT, (min, max, item))

    elif this == ".":
      subpatternappend((ANY, None))

    elif this == "(":
      group = 1
      name = None
      condgroup = None
      if sourcematch("?"):
        group = 0
        # options
        if sourcematch("P"):
          # python extensions
          if sourcematch("<"):
            # named group: skip forward to end of name
            name = ""
            while 1:
              char = sourceget()
              if char is None:
                raise ("unterminated name")
              if char == ">":
                break
              name = name + char
            group = 1
            if not name:
              raise ("missing group name")
            if not isname(name):
              raise ("bad character in group name")
          elif sourcematch("="):
            # named backreference
            name = ""
            while 1:
              char = sourceget()
              if char is None:
                raise ("unterminated name")
              if char == ")":
                break
              name = name + char
            if not name:
              raise ("missing group name")
            if not isname(name):
              raise ("bad character in group name")
            gid = state.groupdict.get(name)
            if gid is None:
              raise ("unknown group name")
            subpatternappend((GROUPREF, gid))
            continue
          else:
            char = sourceget()
            if char is None:
              raise ("unexpected end of pattern")
            raise ("unknown specifier: ?P%s" % char)
        elif sourcematch(":"):
          # non-capturing group
          group = 2
        elif sourcematch("#"):
          # comment
          while 1:
            if source.next is None or source.next == ")":
              break
            sourceget()
          if not sourcematch(")"):
            raise ("unbalanced parenthesis")
          continue
        elif source.next in ASSERTCHARS:
          # lookahead assertions
          char = sourceget()
          dir = 1
          if char == "<":
            if source.next not in LOOKBEHINDASSERTCHARS:
              raise ("syntax error")
            dir = -1  # lookbehind
            char = sourceget()
          p = _parse_sub(source, state)
          if not sourcematch(")"):
            raise ("unbalanced parenthesis")
          if char == "=":
            subpatternappend((ASSERT, (dir, p)))
          else:
            subpatternappend((ASSERT_NOT, (dir, p)))
          continue
        elif sourcematch("("):
          # conditional backreference group
          condname = ""
          while 1:
            char = sourceget()
            if char is None:
              raise ("unterminated name")
            if char == ")":
              break
            condname = condname + char
          group = 2
          if not condname:
            raise ("missing group name")
          if isname(condname):
            condgroup = state.groupdict.get(condname)
            if condgroup is None:
              raise ("unknown group name")
          else:
            try:
              condgroup = int(condname)
            except ValueError:
              raise ("bad character in group name")
        else:
          # flags
          if not source.next in FLAGS:
            raise ("unexpected end of pattern")
          while source.next in FLAGS:
            state.flags = state.flags | FLAGS[sourceget()]
      if group:
        # parse group contents
        if group == 2:
          # anonymous group
          group = None
        else:
          group = state.opengroup(name)
        if condgroup:
          p = _parse_sub_cond(source, state, condgroup)
        else:
          p = _parse_sub(source, state)
        if not sourcematch(")"):
          raise ("unbalanced parenthesis")
        if group is not None:
          state.closegroup(group)
        subpatternappend((SUBPATTERN, (group, p)))
      else:
        while 1:
          char = sourceget()
          if char is None:
            raise ("unexpected end of pattern")
          if char == ")":
            break
          raise ("unknown extension")

    elif this == "^":
      subpatternappend((AT, AT_BEGINNING))

    elif this == "$":
      subpattern.append((AT, AT_END))

    elif this and this[0] == "\\":
      code = _escape(source, this, state)
      subpatternappend(code)

    else:
      raise ("parser error")

  return subpattern


def fix_flags(src, flags):
  # Check and fix flags according to the type of pattern (str or bytes)
  if isinstance(src, str):
    if not flags & SRE_FLAG_ASCII:
      flags |= SRE_FLAG_UNICODE
    elif flags & SRE_FLAG_UNICODE:
      raise ValueError("ASCII and UNICODE flags are incompatible")
  else:
    if flags & SRE_FLAG_UNICODE:
      raise ValueError("can't use UNICODE flag with a bytes pattern")
  return flags


def parse(str, flags=0, pattern=None):
  # parse 're' pattern into list of (opcode, argument) tuples
  source = Tokenizer(str)

  if pattern is None:
    pattern = Pattern()
  pattern.flags = flags
  pattern.str = str
  p = _parse_sub(source, pattern, 0)
  p.pattern.flags = fix_flags(str, p.pattern.flags)

  tail = source.get()
  if tail == ")":
    raise ("unbalanced parenthesis")
  elif tail:
    raise ("bogus characters at end of regular expression")

  if flags & SRE_FLAG_DEBUG:
    p.dump()

  if not (flags & SRE_FLAG_VERBOSE) and p.pattern.flags & SRE_FLAG_VERBOSE:
    # the VERBOSE flag was switched on inside the pattern.  to be
    # on the safe side, we'll parse the whole thing again...
    return parse(str, p.pattern.flags)

  return p


def parse_template(source, pattern):
  # parse 're' replacement string into list of literals and
  # group references
  s = Tokenizer(source)
  sget = s.get
  p = []
  a = p.append

  def literal(literal, p=p, pappend=a):
    if p and p[-1][0] is LITERAL:
      p[-1] = LITERAL, p[-1][1] + literal
    else:
      pappend((LITERAL, literal))
  sep = source[:0]
  if isinstance(sep, str):
    makechar = chr
  else:
    makechar = chr
  while 1:
    this = sget()
    if this is None:
      break  # end of replacement string
    if this and this[0] == "\\":
      # group
      c = this[1:2]
      if c == "g":
        name = ""
        if s.match("<"):
          while 1:
            char = sget()
            if char is None:
              raise ("unterminated group name")
            if char == ">":
              break
            name = name + char
        if not name:
          raise ("missing group name")
        try:
          index = int(name)
          if index < 0:
            raise ("negative group number")
        except ValueError:
          if not isname(name):
            raise ("bad character in group name")
          try:
            index = pattern.groupindex[name]
          except KeyError:
            raise IndexError("unknown group name")
        a((MARK, index))
      elif c == "0":
        if s.next in OCTDIGITS:
          this = this + sget()
          if s.next in OCTDIGITS:
            this = this + sget()
        literal(makechar(int(this[1:], 8) & 0xff))
      elif c in DIGITS:
        isoctal = False
        if s.next in DIGITS:
          this = this + sget()
          if (c in OCTDIGITS and this[2] in OCTDIGITS and
                  s.next in OCTDIGITS):
            this = this + sget()
            isoctal = True
            literal(makechar(int(this[1:], 8) & 0xff))
        if not isoctal:
          a((MARK, int(this[1:])))
      else:
        try:
          this = makechar(ESCAPES[this][1])
        except KeyError:
          pass
        literal(this)
    else:
      literal(this)
  # convert template to groups and literals lists
  i = 0
  groups = []
  groupsappend = groups.append
  literals = [None] * len(p)
  if isinstance(source, str):
    encode = lambda x: x
  else:
    # The tokenizer implicitly decodes bytes objects as latin-1, we must
    # therefore re-encode the final representation.
    encode = lambda x: x.encode('latin-1')
  for c, s in p:
    if c is MARK:
      groupsappend((i, s))
      # literal[i] is already None
    else:
      literals[i] = encode(s)
    i = i + 1
  return groups, literals


def expand_template(template, match):
  g = match.group
  sep = match.string[:0]
  groups, literals = template
  literals = literals[:]
  try:
    for index, group in groups:
      literals[index] = s = g(group)
      if s is None:
        raise ("unmatched group")
  except IndexError:
    raise ("invalid group reference")
  return sep.join(literals)

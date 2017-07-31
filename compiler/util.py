# coding=utf-8

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

"""Utilities for generating Go code."""

from __future__ import unicode_literals

import codecs
import contextlib
import cStringIO
import string
import StringIO
import textwrap


_SIMPLE_CHARS = set(string.digits + string.letters + string.punctuation + " ")
_ESCAPES = {'\t': r'\t', '\r': r'\r', '\n': r'\n', '"': r'\"', '\\': r'\\'}


# This is the max length of a direct allocation tuple supported by the runtime.
# This should match the number of specializations found in tuple.go.
MAX_DIRECT_TUPLE = 6


class CompileError(Exception):

  def __init__(self, node, msg):
    if hasattr(node, 'lineno'):
      msg = 'line {}: {}'.format(node.lineno, msg)
    super(CompileError, self).__init__(msg)


class ParseError(CompileError):
  pass


class ImportError(CompileError):  # pylint: disable=redefined-builtin
  pass


class LateFutureError(ImportError):

  def __init__(self, node):
    msg = 'from __future__ imports must occur at the beginning of the file'
    super(LateFutureError, self).__init__(node, msg)


class Writer(object):
  """Utility class for writing blocks of Go code to a file-like object."""

  def __init__(self, out=None):
    self.out = codecs.getwriter('utf8')(out or cStringIO.StringIO())
    self.indent_level = 0

  def getvalue(self):
    return self.out.getvalue().decode('utf8')

  @contextlib.contextmanager
  def indent_block(self, n=1):
    """A context manager that indents by n on entry and dedents on exit."""
    self.indent(n)
    yield
    self.dedent(n)

  def write(self, output):
    for line in output.split('\n'):
      if line:
        self.out.write(''.join(('\t' * self.indent_level, line, '\n')))

  def write_block(self, block_, body):
    """Outputs the boilerplate necessary for code blocks like functions.

    Args:
      block_: The Block object representing the code block.
      body: String containing Go code making up the body of the code block.
    """
    self.write('for ; πF.State() >= 0; πF.PopCheckpoint() {')
    with self.indent_block():
      self.write('switch πF.State() {')
      self.write('case 0:')
      for checkpoint in block_.checkpoints:
        self.write_tmpl('case $state: goto Label$state', state=checkpoint)
      self.write('default: panic("unexpected function state")')
      self.write('}')
      # Assume that body is aligned with goto labels.
      with self.indent_block(-1):
        self.write(body)
    self.write('}')

  def write_label(self, label):
    with self.indent_block(-1):
      self.write('Label{}:'.format(label))

  def write_py_context(self, lineno, line):
    self.write_tmpl('// line $lineno: $line', lineno=lineno, line=line)

  def write_tmpl(self, tmpl, **kwargs):
    self.write(string.Template(tmpl).substitute(kwargs))

  def write_checked_call2(self, result, call, *args, **kwargs):
    return self.write_tmpl(textwrap.dedent("""\
        if $result, πE = $call; πE != nil {
        \tcontinue
        }"""), result=result.name, call=call.format(*args, **kwargs))

  def write_checked_call1(self, call, *args, **kwargs):
    return self.write_tmpl(textwrap.dedent("""\
        if πE = $call; πE != nil {
        \tcontinue
        }"""), call=call.format(*args, **kwargs))

  def write_temp_decls(self, block_):
    all_temps = block_.free_temps | block_.used_temps
    for temp in sorted(all_temps, key=lambda t: t.name):
      self.write('var {0} {1}\n_ = {0}'.format(temp.name, temp.type_))

  def indent(self, n=1):
    self.indent_level += n

  def dedent(self, n=1):
    self.indent_level -= n


def go_str(value):
  """Returns value as a valid Go string literal."""
  io = StringIO.StringIO()
  io.write('"')
  for c in value:
    if c in _ESCAPES:
      io.write(_ESCAPES[c])
    elif c in _SIMPLE_CHARS:
      io.write(c)
    else:
      io.write(r'\x{:02x}'.format(ord(c)))
  io.write('"')
  return io.getvalue()


def adjust_local_name(name):
  """Returns a Go identifier for the given Python variable name."""
  return 'µ' + name

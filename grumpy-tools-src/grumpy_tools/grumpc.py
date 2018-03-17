#!/usr/bin/env python
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

"""A Python -> Go transcompiler."""

from __future__ import unicode_literals

import argparse
import os
import sys
import textwrap

from .compiler import block
from .compiler import imputil
from .compiler import stmt
from .compiler import util
from .vendor import pythonparser


def main(script=None, modname=None):
  assert script and modname, 'Script "%s" or Modname "%s" is empty' % (script,modname)

  gopath = os.getenv('GOPATH', None)
  if not gopath:
    print >> sys.stderr, 'GOPATH not set'
    return 1

  with open(script) as py_file:
    py_contents = py_file.read()
  try:
    mod = pythonparser.parse(py_contents)
  except SyntaxError as e:
    print >> sys.stderr, '{}: line {}: invalid syntax: {}'.format(
        e.filename, e.lineno, e.text)
    return 2

  # Do a pass for compiler directives from `from __future__ import *` statements
  try:
    future_node, future_features = imputil.parse_future_features(mod)
  except util.CompileError as e:
    print >> sys.stderr, str(e)
    return 2

  importer = imputil.Importer(gopath, modname, script,
                              future_features.absolute_import)
  full_package_name = modname.replace('.', '/')
  mod_block = block.ModuleBlock(importer, full_package_name, script,
                                py_contents, future_features)

  visitor = stmt.StatementVisitor(mod_block, future_node)
  # Indent so that the module body is aligned with the goto labels.
  with visitor.writer.indent_block():
    try:
      visitor.visit(mod)
    except util.ParseError as e:
      print >> sys.stderr, str(e)
      return 2

  writer = util.Writer(sys.stdout)
  tmpl = textwrap.dedent("""\
      package $package
      import πg "grumpy"
      var Code *πg.Code
      func init() {
      \tCode = πg.NewCode("<module>", $script, nil, 0, func(πF *πg.Frame, _ []*πg.Object) (*πg.Object, *πg.BaseException) {
      \t\tvar πR *πg.Object; _ = πR
      \t\tvar πE *πg.BaseException; _ = πE""")
  writer.write_tmpl(tmpl, package=modname.split('.')[-1],
                    script=util.go_str(script))
  with writer.indent_block(2):
    for s in sorted(mod_block.strings):
      writer.write('ß{} := πg.InternStr({})'.format(s, util.go_str(s)))
    writer.write_temp_decls(mod_block)
    writer.write_block(mod_block, visitor.writer.getvalue())
  writer.write_tmpl(textwrap.dedent("""\
    \t\treturn nil, πE
    \t})
    \tπg.RegisterModule($modname, Code)
    }"""), modname=util.go_str(modname))
  return 0

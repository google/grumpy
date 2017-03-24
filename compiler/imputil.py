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

"""Functionality for importing modules in Grumpy."""


from __future__ import unicode_literals

import collections
import os

from pythonparser import algorithm
from pythonparser import ast

from grumpy.compiler import util


_NATIVE_MODULE_PREFIX = '__go__.'


class Path(object):
  """Resolves imported modules based on a search path of directories."""

  def __init__(self, gopath, modname, script):
    self.dirs = []
    if gopath:
      self.dirs.extend(os.path.join(d, 'src', '__python__')
                       for d in gopath.split(os.pathsep))
    dirname, basename = os.path.split(script)
    if basename == '__init__.py':
      self.package_dir = dirname
      self.package_name = modname
    elif (modname.find('.') != -1 and
          os.path.isfile(os.path.join(dirname, '__init__.py'))):
      self.package_dir = dirname
      self.package_name = modname[:modname.rfind('.')]
    else:
      self.package_dir = None
      self.package_name = None

  def resolve_import(self, modname):
    """Find module on the path returning full module name and script path.

    Args:
      modname: Name identified by an import statement, possibly relative.

    Returns:
      A pair (full_name, script), where full_name is the absolute module name
      and script is the filename of the associate .py file.
    """
    if self.package_dir:
      script = self._find_script(self.package_dir, modname)
      if script:
        return '{}.{}'.format(self.package_name, modname), script
    for dirname in self.dirs:
      script = self._find_script(dirname, modname)
      if script:
        return modname, script
    return None, None

  def _find_script(self, dirname, name):
    prefix = os.path.join(dirname, name.replace('.', os.sep))
    script = prefix + '.py'
    if os.path.isfile(script):
      return script
    script = os.path.join(prefix, '__init__.py')
    if os.path.isfile(script):
      return script
    return None


class Import(object):
  """Represents a single module import and all its associated bindings.

  Each import pertains to a single module that is imported. Thus one import
  statement may produce multiple Import objects. E.g. "import foo, bar" makes
  an Import object for module foo and another one for module bar.
  """

  Binding = collections.namedtuple('Binding', ('bind_type', 'alias', 'value'))

  MODULE = "<BindType 'module'>"
  MEMBER = "<BindType 'member'>"

  def __init__(self, name, is_native=False):
    self.name = name
    self.is_native = is_native
    self.bindings = []

  def add_binding(self, bind_type, alias, value):
    self.bindings.append(Import.Binding(bind_type, alias, value))


class ImportVisitor(algorithm.Visitor):
  """Visits import nodes and produces corresponding Import objects."""

  # pylint: disable=invalid-name,missing-docstring,no-init

  def __init__(self, path):
    self.path = path
    self.imports = []

  def visit_Import(self, node):
    for alias in node.names:
      if alias.name.startswith(_NATIVE_MODULE_PREFIX):
        raise util.ImportError(
            node, 'for native imports use "from __go__.xyz import ..." syntax')
      imp = self._resolve_import(node, alias.name)
      if alias.asname:
        imp.add_binding(Import.MODULE, alias.asname, imp.name.count('.'))
      else:
        parts = alias.name.split('.')
        imp.add_binding(Import.MODULE, parts[-1],
                        imp.name.count('.') - len(parts) + 1)
      self.imports.append(imp)

  def visit_ImportFrom(self, node):
    if any(a.name == '*' for a in node.names):
      msg = 'wildcard member import is not implemented: from %s import *' % (
          node.module)
      raise util.ImportError(node, msg)

    if node.module == '__future__':
      return

    if node.module.startswith(_NATIVE_MODULE_PREFIX):
      imp = Import(node.module[len(_NATIVE_MODULE_PREFIX):], is_native=True)
      for alias in node.names:
        asname = alias.asname or alias.name
        imp.add_binding(Import.MEMBER, asname, alias.name)
      self.imports.append(imp)
      return

    member_imp = None
    for alias in node.names:
      asname = alias.asname or alias.name
      full_name, _ = self.path.resolve_import(
          '{}.{}'.format(node.module, alias.name))
      if full_name:
        # Imported name is a submodule within a package, so bind that module.
        imp = Import(full_name)
        imp.add_binding(Import.MODULE, asname, imp.name.count('.'))
        self.imports.append(imp)
      else:
        # A member (not a submodule) is being imported, so bind it.
        if not member_imp:
          member_imp = self._resolve_import(node, node.module)
          self.imports.append(member_imp)
        member_imp.add_binding(Import.MEMBER, asname, alias.name)

  def _resolve_import(self, node, name):
    full_name, _ = self.path.resolve_import(name)
    if not full_name:
      raise util.ImportError(node, 'no such module: {}'.format(name))
    return Import(full_name)


# Parser flags, set on 'from __future__ import *', see parser_flags on
# StatementVisitor below. Note these have the same values as CPython.
FUTURE_DIVISION = 0x2000
FUTURE_ABSOLUTE_IMPORT = 0x4000
FUTURE_PRINT_FUNCTION = 0x10000
FUTURE_UNICODE_LITERALS = 0x20000

# Names for future features in 'from __future__ import *'. Map from name in the
# import statement to a tuple of the flag for parser, and whether we've (grumpy)
# implemented the feature yet.
future_features = {
    "division": (FUTURE_DIVISION, False),
    "absolute_import": (FUTURE_ABSOLUTE_IMPORT, False),
    "print_function": (FUTURE_PRINT_FUNCTION, True),
    "unicode_literals": (FUTURE_UNICODE_LITERALS, False),
}

# These future features are already in the language proper as of 2.6, so
# importing them via __future__ has no effect.
redundant_future_features = ["generators", "with_statement", "nested_scopes"]


def import_from_future(node):
  """Processes a future import statement, returning set of flags it defines."""
  assert isinstance(node, ast.ImportFrom)
  assert node.module == '__future__'
  flags = 0
  for alias in node.names:
    name = alias.name
    if name in future_features:
      flag, implemented = future_features[name]
      if not implemented:
        msg = 'future feature {} not yet implemented by grumpy'.format(name)
        raise util.ParseError(node, msg)
      flags |= flag
    elif name == 'braces':
      raise util.ParseError(node, 'not a chance')
    elif name not in redundant_future_features:
      msg = 'future feature {} is not defined'.format(name)
      raise util.ParseError(node, msg)
  return flags


class FutureFeatures(object):
  def __init__(self):
    self.parser_flags = 0
    self.future_lineno = 0


def visit_future(node):
  """Accumulates a set of compiler flags for the compiler __future__ imports.

  Returns an instance of FutureFeatures which encapsulates the flags and the
  line number of the last valid future import parsed. A downstream parser can
  use the latter to detect invalid future imports that appear too late in the
  file.
  """
  # If this is the module node, do an initial pass through the module body's
  # statements to detect future imports and process their directives (i.e.,
  # set compiler flags), and detect ones that don't appear at the beginning of
  # the file. The only things that can proceed a future statement are other
  # future statements and/or a doc string.
  assert isinstance(node, ast.Module)
  ff = FutureFeatures()
  done = False
  found_docstring = False
  for node in node.body:
    if isinstance(node, ast.ImportFrom):
      modname = node.module
      if modname == '__future__':
        if done:
          raise util.LateFutureError(node)
        ff.parser_flags |= import_from_future(node)
        ff.future_lineno = node.lineno
      else:
        done = True
    elif isinstance(node, ast.Expr) and not found_docstring:
      e = node.value
      if not isinstance(e, ast.Str): # pylint: disable=simplifiable-if-statement
        done = True
      else:
        found_docstring = True
    else:
      done = True
  return ff

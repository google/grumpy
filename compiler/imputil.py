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
import functools
import os

from grumpy.compiler import util
from grumpy.pythonparser import algorithm
from grumpy.pythonparser import ast


_NATIVE_MODULE_PREFIX = '__go__.'


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


class Importer(algorithm.Visitor):
  """Visits import nodes and produces corresponding Import objects."""

  # pylint: disable=invalid-name,missing-docstring,no-init

  def __init__(self, gopath, modname, script, absolute_import):
    self.pathdirs = []
    if gopath:
      self.pathdirs.extend(os.path.join(d, 'src', '__python__')
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
    self.absolute_import = absolute_import

  def generic_visit(self, node):
    raise ValueError('Import cannot visit {} node'.format(type(node).__name__))

  def visit_Import(self, node):
    imports = []
    for alias in node.names:
      if alias.name.startswith(_NATIVE_MODULE_PREFIX):
        raise util.ImportError(
            node, 'for native imports use "from __go__.xyz import ..." syntax')
      imp = self._resolve_import(node, alias.name)
      if alias.asname:
        imp.add_binding(Import.MODULE, alias.asname, imp.name.count('.'))
      else:
        parts = alias.name.split('.')
        imp.add_binding(Import.MODULE, parts[0],
                        imp.name.count('.') - len(parts) + 1)
      imports.append(imp)
    return imports

  def visit_ImportFrom(self, node):
    if any(a.name == '*' for a in node.names):
      msg = 'wildcard member import is not implemented: from %s import *' % (
          node.module)
      raise util.ImportError(node, msg)

    if not node.level and node.module == '__future__':
      return []

    if not node.level and node.module.startswith(_NATIVE_MODULE_PREFIX):
      imp = Import(node.module[len(_NATIVE_MODULE_PREFIX):], is_native=True)
      for alias in node.names:
        asname = alias.asname or alias.name
        imp.add_binding(Import.MEMBER, asname, alias.name)
      return [imp]

    imports = []
    if not node.module:
      # Import of the form 'from .. import foo, bar'. All named imports must be
      # modules, not module members.
      for alias in node.names:
        imp = self._resolve_relative_import(node.level, node, alias.name)
        imp.add_binding(Import.MODULE, alias.asname or alias.name,
                        imp.name.count('.'))
        imports.append(imp)
      return imports

    member_imp = None
    for alias in node.names:
      asname = alias.asname or alias.name
      if node.level:
        resolver = functools.partial(self._resolve_relative_import, node.level)
      else:
        resolver = self._resolve_import
      try:
        imp = resolver(node, '{}.{}'.format(node.module, alias.name))
      except util.ImportError:
        # A member (not a submodule) is being imported, so bind it.
        if not member_imp:
          member_imp = resolver(node, node.module)
          imports.append(member_imp)
        member_imp.add_binding(Import.MEMBER, asname, alias.name)
      else:
        # Imported name is a submodule within a package, so bind that module.
        imp.add_binding(Import.MODULE, asname, imp.name.count('.'))
        imports.append(imp)
    return imports

  def _resolve_import(self, node, modname):
    if not self.absolute_import and self.package_dir:
      if self._script_exists(self.package_dir, modname):
        return Import('{}.{}'.format(self.package_name, modname))
    for dirname in self.pathdirs:
      if self._script_exists(dirname, modname):
        return Import(modname)
    raise util.ImportError(node, 'no such module: {}'.format(modname))

  def _resolve_relative_import(self, level, node, modname):
    if not self.package_dir:
      raise util.ImportError(node, 'attempted relative import in non-package')
    uplevel = level - 1
    if uplevel > self.package_name.count('.'):
      raise util.ImportError(
          node, 'attempted relative import beyond toplevel package')
    dirname = os.path.normpath(os.path.join(
        self.package_dir, *(['..'] * uplevel)))
    if not self._script_exists(dirname, modname):
      raise util.ImportError(node, 'no such module: {}'.format(modname))
    parts = self.package_name.split('.')
    return Import('.'.join(parts[:len(parts)-uplevel]) + '.' + modname)

  def _script_exists(self, dirname, name):
    prefix = os.path.join(dirname, name.replace('.', os.sep))
    return (os.path.isfile(prefix + '.py') or
            os.path.isfile(os.path.join(prefix, '__init__.py')))


_FUTURE_FEATURES = (
    'absolute_import',
    'division',
    'print_function',
    'unicode_literals',
)

_IMPLEMENTED_FUTURE_FEATURES = (
    'absolute_import',
    'print_function',
    'unicode_literals'
)

# These future features are already in the language proper as of 2.6, so
# importing them via __future__ has no effect.
_REDUNDANT_FUTURE_FEATURES = ('generators', 'with_statement', 'nested_scopes')


class FutureFeatures(object):
  """Spec for future feature flags imported by a module."""

  def __init__(self, absolute_import=False, division=False,
               print_function=False, unicode_literals=False):
    self.absolute_import = absolute_import
    self.division = division
    self.print_function = print_function
    self.unicode_literals = unicode_literals


def _make_future_features(node):
  """Processes a future import statement, returning set of flags it defines."""
  assert isinstance(node, ast.ImportFrom)
  assert node.module == '__future__'
  features = FutureFeatures()
  for alias in node.names:
    name = alias.name
    if name in _FUTURE_FEATURES:
      if name not in _IMPLEMENTED_FUTURE_FEATURES:
        msg = 'future feature {} not yet implemented by grumpy'.format(name)
        raise util.ParseError(node, msg)
      setattr(features, name, True)
    elif name == 'braces':
      raise util.ParseError(node, 'not a chance')
    elif name not in _REDUNDANT_FUTURE_FEATURES:
      msg = 'future feature {} is not defined'.format(name)
      raise util.ParseError(node, msg)
  return features


def parse_future_features(mod):
  """Accumulates a set of flags for the compiler __future__ imports."""
  assert isinstance(mod, ast.Module)
  found_docstring = False
  for node in mod.body:
    if isinstance(node, ast.ImportFrom):
      if node.module == '__future__':
        return node, _make_future_features(node)
      break
    elif isinstance(node, ast.Expr) and not found_docstring:
      if not isinstance(node.value, ast.Str):
        break
      found_docstring = True
    else:
      break
  return None, FutureFeatures()

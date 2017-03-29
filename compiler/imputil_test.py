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

"""Tests ImportVisitor and related classes."""

from __future__ import unicode_literals

import os
import shutil
import tempfile
import textwrap
import unittest

import pythonparser

from grumpy.compiler import imputil
from grumpy.compiler import util


class MockPath(object):

  def __init__(self, nonexistent_modules=()):
    self.nonexistent_modules = nonexistent_modules

  def resolve_import(self, modname):
    if modname in self.nonexistent_modules:
      return None, None
    return modname, modname.replace('.', os.sep)


class _MaterializedPathTree(object):
  """Context manager that materializes a tree of files and cleans them up."""

  def __init__(self, spec):
    self.spec = spec
    self.rootdir = None
    self.pydir = None

  def __enter__(self):
    self.rootdir = tempfile.mkdtemp()
    self.pydir = os.path.join(self.rootdir, 'src', '__python__')
    self._materialize(self.rootdir, {'src/': {'__python__/': self.spec}})
    return self

  def __exit__(self, *args):
    shutil.rmtree(self.rootdir)

  def _materialize(self, dirname, spec):
    for name, sub_spec in spec.iteritems():
      if name.endswith('/'):
        subdir = os.path.join(dirname, name[:-1])
        os.mkdir(subdir)
        self._materialize(subdir, sub_spec)
      else:
        with open(os.path.join(dirname, name), 'w'):
          pass


class PathTest(unittest.TestCase):

  def testResolveImportEmptyPath(self):
    path = imputil.Path(None, 'foo', 'foo.py')
    self.assertEqual(path.resolve_import('bar'), (None, None))

  def testResolveImportTopLevelModule(self):
    with _MaterializedPathTree({'bar.py': None}) as tree:
      path = imputil.Path(tree.rootdir, 'foo', 'foo.py')
      want = ('bar', os.path.join(tree.pydir, 'bar.py'))
      self.assertEqual(path.resolve_import('bar'), want)

  def testResolveImportTopLevelPackage(self):
    with _MaterializedPathTree({'bar/': {'__init__.py': None}}) as tree:
      path = imputil.Path(tree.rootdir, 'foo', 'foo.py')
      want = ('bar', os.path.join(tree.pydir, 'bar', '__init__.py'))
      self.assertEqual(path.resolve_import('bar'), want)

  def testResolveImportPackageModuleAbsolute(self):
    spec = {
        'bar/': {
            '__init__.py': None,
            'baz.py': None,
        }
    }
    with _MaterializedPathTree(spec) as tree:
      path = imputil.Path(tree.rootdir, 'foo', 'foo.py')
      want = ('bar.baz', os.path.join(tree.pydir, 'bar', 'baz.py'))
      self.assertEqual(path.resolve_import('bar.baz'), want)

  def testResolveImportPackageModuleRelative(self):
    spec = {
        'bar/': {
            '__init__.py': None,
            'baz.py': None,
        }
    }
    with _MaterializedPathTree(spec) as tree:
      bar_script = os.path.join(tree.pydir, 'bar', '__init__.py')
      path = imputil.Path(tree.rootdir, 'bar', bar_script)
      want = ('bar.baz', os.path.join(tree.pydir, 'bar', 'baz.py'))
      self.assertEqual(path.resolve_import('baz'), want)

  def testResolveImportPackageModuleRelativeFromSubModule(self):
    spec = {
        'bar/': {
            '__init__.py': None,
            'baz.py': None,
            'foo.py': None,
        }
    }
    with _MaterializedPathTree(spec) as tree:
      foo_script = os.path.join(tree.pydir, 'bar', 'foo.py')
      path = imputil.Path(tree.rootdir, 'bar.foo', foo_script)
      want = ('bar.baz', os.path.join(tree.pydir, 'bar', 'baz.py'))
      self.assertEqual(path.resolve_import('baz'), want)


class ImportVisitorTest(unittest.TestCase):

  def testImport(self):
    imp = imputil.Import('foo')
    imp.add_binding(imputil.Import.MODULE, 'foo', 0)
    self._assert_imports_equal(imp, self._visit_import('import foo'))

  def testImportMultiple(self):
    imp1 = imputil.Import('foo')
    imp1.add_binding(imputil.Import.MODULE, 'foo', 0)
    imp2 = imputil.Import('bar')
    imp2.add_binding(imputil.Import.MODULE, 'bar', 0)
    self._assert_imports_equal(
        [imp1, imp2], self._visit_import('import foo, bar'))

  def testImportAs(self):
    imp = imputil.Import('foo')
    imp.add_binding(imputil.Import.MODULE, 'bar', 0)
    self._assert_imports_equal(imp, self._visit_import('import foo as bar'))

  def testImportNativeRaises(self):
    self.assertRaises(util.ImportError, self._visit_import, 'import __go__.fmt')

  def testImportFrom(self):
    imp = imputil.Import('foo.bar')
    imp.add_binding(imputil.Import.MODULE, 'bar', 1)
    self._assert_imports_equal(imp, self._visit_import('from foo import bar'))

  def testImportFromMember(self):
    imp = imputil.Import('foo')
    imp.add_binding(imputil.Import.MEMBER, 'bar', 'bar')
    path = MockPath(nonexistent_modules=('foo.bar',))
    self._assert_imports_equal(
        imp, self._visit_import('from foo import bar', path=path))

  def testImportFromMultiple(self):
    imp1 = imputil.Import('foo.bar')
    imp1.add_binding(imputil.Import.MODULE, 'bar', 1)
    imp2 = imputil.Import('foo.baz')
    imp2.add_binding(imputil.Import.MODULE, 'baz', 1)
    self._assert_imports_equal(
        [imp1, imp2], self._visit_import('from foo import bar, baz'))

  def testImportFromMixedMembers(self):
    imp1 = imputil.Import('foo')
    imp1.add_binding(imputil.Import.MEMBER, 'bar', 'bar')
    imp2 = imputil.Import('foo.baz')
    imp2.add_binding(imputil.Import.MODULE, 'baz', 1)
    path = MockPath(nonexistent_modules=('foo.bar',))
    self._assert_imports_equal(
        [imp1, imp2], self._visit_import('from foo import bar, baz', path=path))

  def testImportFromAs(self):
    imp = imputil.Import('foo.bar')
    imp.add_binding(imputil.Import.MODULE, 'baz', 1)
    self._assert_imports_equal(
        imp, self._visit_import('from foo import bar as baz'))

  def testImportFromAsMembers(self):
    imp = imputil.Import('foo')
    imp.add_binding(imputil.Import.MEMBER, 'baz', 'bar')
    path = MockPath(nonexistent_modules=('foo.bar',))
    self._assert_imports_equal(
        imp, self._visit_import('from foo import bar as baz', path=path))

  def testImportFromWildcardRaises(self):
    self.assertRaises(util.ImportError, self._visit_import, 'from foo import *')

  def testImportFromFuture(self):
    result = self._visit_import('from __future__ import print_function')
    self.assertEqual([], result)

  def testImportFromNative(self):
    imp = imputil.Import('fmt', is_native=True)
    imp.add_binding(imputil.Import.MEMBER, 'Printf', 'Printf')
    self._assert_imports_equal(
        imp, self._visit_import('from __go__.fmt import Printf'))

  def testImportFromNativeMultiple(self):
    imp = imputil.Import('fmt', is_native=True)
    imp.add_binding(imputil.Import.MEMBER, 'Printf', 'Printf')
    imp.add_binding(imputil.Import.MEMBER, 'Println', 'Println')
    self._assert_imports_equal(
        imp, self._visit_import('from __go__.fmt import Printf, Println'))

  def testImportFromNativeAs(self):
    imp = imputil.Import('fmt', is_native=True)
    imp.add_binding(imputil.Import.MEMBER, 'foo', 'Printf')
    self._assert_imports_equal(
        imp, self._visit_import('from __go__.fmt import Printf as foo'))

  def _visit_import(self, source, path=None):
    if not path:
      path = MockPath()
    visitor = imputil.ImportVisitor(path)
    visitor.visit(pythonparser.parse(source).body[0])
    return visitor.imports

  def _assert_imports_equal(self, want, got):
    if isinstance(want, imputil.Import):
      want = [want]
    self.assertEqual([imp.__dict__ for imp in want],
                     [imp.__dict__ for imp in got])


class MakeFutureFeaturesTest(unittest.TestCase):

  def testImportFromFuture(self):
    print_function_features = imputil.FutureFeatures()
    print_function_features.print_function = True
    testcases = [
        ('from __future__ import print_function',
         print_function_features),
        ('from __future__ import generators', imputil.FutureFeatures()),
        ('from __future__ import generators, print_function',
         print_function_features),
    ]

    for tc in testcases:
      source, want = tc
      mod = pythonparser.parse(textwrap.dedent(source))
      node = mod.body[0]
      got = imputil._make_future_features(node)  # pylint: disable=protected-access
      self.assertEqual(want.__dict__, got.__dict__)

  def testImportFromFutureParseError(self):
    testcases = [
        # NOTE: move this group to testImportFromFuture as they are implemented
        # by grumpy
        ('from __future__ import absolute_import',
         r'future feature \w+ not yet implemented'),
        ('from __future__ import division',
         r'future feature \w+ not yet implemented'),
        ('from __future__ import unicode_literals',
         r'future feature \w+ not yet implemented'),
        ('from __future__ import braces', 'not a chance'),
        ('from __future__ import nonexistant_feature',
         r'future feature \w+ is not defined'),
    ]

    for tc in testcases:
      source, want_regexp = tc
      mod = pythonparser.parse(source)
      node = mod.body[0]
      self.assertRaisesRegexp(util.ParseError, want_regexp,
                              imputil._make_future_features, node)  # pylint: disable=protected-access


class ParseFutureFeaturesTest(unittest.TestCase):

  def testVisitFuture(self):
    print_function_features = imputil.FutureFeatures()
    print_function_features.print_function = True
    testcases = [
        ('from __future__ import print_function',
         print_function_features),
        ("""\
        "module docstring"

        from __future__ import print_function
        """, print_function_features),
        ("""\
        "module docstring"

        from __future__ import print_function, with_statement
        from __future__ import nested_scopes
        """, print_function_features),
    ]

    for tc in testcases:
      source, want = tc
      mod = pythonparser.parse(textwrap.dedent(source))
      _, got = imputil.parse_future_features(mod)
      self.assertEqual(want.__dict__, got.__dict__)

  def testVisitFutureLate(self):
    testcases = [
        # future after normal imports
        """\
        import os
        from __future__ import print_function
        """,
        # future after non-docstring expression
        """
        asd = 123
        from __future__ import print_function
        """
    ]

    for source in testcases:
      mod = pythonparser.parse(textwrap.dedent(source))
      self.assertRaises(util.LateFutureError,
                        imputil.parse_future_features, mod)

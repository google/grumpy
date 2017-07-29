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

import copy
import os
import shutil
import tempfile
import textwrap
import unittest

from grumpy.compiler import imputil
from grumpy.compiler import util
from grumpy import pythonparser


class ImportVisitorTest(unittest.TestCase):

  _PATH_SPEC = {
      'foo.py': None,
      'qux.py': None,
      'bar/': {
          'fred/': {
              '__init__.py': None,
              'quux.py': None,
          },
          '__init__.py': None,
          'baz.py': None,
          'foo.py': None,
      },
      'baz.py': None,
  }

  def setUp(self):
    self.rootdir = tempfile.mkdtemp()
    self.pydir = os.path.join(self.rootdir, 'src', '__python__')
    self._materialize_tree(
        self.rootdir, {'src/': {'__python__/': self._PATH_SPEC}})
    foo_script = os.path.join(self.rootdir, 'foo.py')
    self.importer = imputil.Importer(self.rootdir, 'foo', foo_script, False)
    bar_script = os.path.join(self.pydir, 'bar', '__init__.py')
    self.bar_importer = imputil.Importer(
        self.rootdir, 'bar', bar_script, False)
    fred_script = os.path.join(self.pydir, 'bar', 'fred', '__init__.py')
    self.fred_importer = imputil.Importer(
        self.rootdir, 'bar.fred', fred_script, False)

    self.foo_import = imputil.Import(
        'foo', os.path.join(self.pydir, 'foo.py'))
    self.qux_import = imputil.Import(
        'qux', os.path.join(self.pydir, 'qux.py'))
    self.bar_import = imputil.Import(
        'bar', os.path.join(self.pydir, 'bar/__init__.py'))
    self.fred_import = imputil.Import(
        'bar.fred', os.path.join(self.pydir, 'bar/fred/__init__.py'))
    self.quux_import = imputil.Import(
        'bar.fred.quux', os.path.join(self.pydir, 'bar/fred/quux.py'))
    self.baz2_import = imputil.Import(
        'bar.baz', os.path.join(self.pydir, 'bar/baz.py'))
    self.foo2_import = imputil.Import(
        'bar.foo', os.path.join(self.pydir, 'bar/foo.py'))
    self.baz_import = imputil.Import(
        'baz', os.path.join(self.pydir, 'baz.py'))

  def tearDown(self):
    shutil.rmtree(self.rootdir)

  def testImportEmptyPath(self):
    importer = imputil.Importer(None, 'foo', 'foo.py', False)
    self.assertRaises(util.ImportError, importer.visit,
                      pythonparser.parse('import bar').body[0])

  def testImportTopLevelModule(self):
    imp = copy.deepcopy(self.qux_import)
    imp.add_binding(imputil.Import.MODULE, 'qux', 0)
    self._check_imports('import qux', [imp])

  def testImportTopLevelPackage(self):
    imp = copy.deepcopy(self.bar_import)
    imp.add_binding(imputil.Import.MODULE, 'bar', 0)
    self._check_imports('import bar', [imp])

  def testImportPackageModuleAbsolute(self):
    imp = copy.deepcopy(self.baz2_import)
    imp.add_binding(imputil.Import.MODULE, 'bar', 0)
    self._check_imports('import bar.baz', [imp])

  def testImportFromSubModule(self):
    imp = copy.deepcopy(self.baz2_import)
    imp.add_binding(imputil.Import.MODULE, 'baz', 1)
    self._check_imports('from bar import baz', [imp])

  def testImportPackageModuleRelative(self):
    imp = copy.deepcopy(self.baz2_import)
    imp.add_binding(imputil.Import.MODULE, 'baz', 1)
    got = self.bar_importer.visit(pythonparser.parse('import baz').body[0])
    self._assert_imports_equal([imp], got)

  def testImportPackageModuleRelativeFromSubModule(self):
    imp = copy.deepcopy(self.baz2_import)
    imp.add_binding(imputil.Import.MODULE, 'baz', 1)
    foo_script = os.path.join(self.pydir, 'bar', 'foo.py')
    importer = imputil.Importer(self.rootdir, 'bar.foo', foo_script, False)
    got = importer.visit(pythonparser.parse('import baz').body[0])
    self._assert_imports_equal([imp], got)

  def testImportPackageModuleAbsoluteImport(self):
    imp = copy.deepcopy(self.baz_import)
    imp.add_binding(imputil.Import.MODULE, 'baz', 0)
    bar_script = os.path.join(self.pydir, 'bar', '__init__.py')
    importer = imputil.Importer(self.rootdir, 'bar', bar_script, True)
    got = importer.visit(pythonparser.parse('import baz').body[0])
    self._assert_imports_equal([imp], got)

  def testImportMultiple(self):
    imp1 = copy.deepcopy(self.foo_import)
    imp1.add_binding(imputil.Import.MODULE, 'foo', 0)
    imp2 = copy.deepcopy(self.baz2_import)
    imp2.add_binding(imputil.Import.MODULE, 'bar', 0)
    self._check_imports('import foo, bar.baz', [imp1, imp2])

  def testImportAs(self):
    imp = copy.deepcopy(self.foo_import)
    imp.add_binding(imputil.Import.MODULE, 'bar', 0)
    self._check_imports('import foo as bar', [imp])

  def testImportFrom(self):
    imp = copy.deepcopy(self.baz2_import)
    imp.add_binding(imputil.Import.MODULE, 'baz', 1)
    self._check_imports('from bar import baz', [imp])

  def testImportFromMember(self):
    imp = copy.deepcopy(self.foo_import)
    imp.add_binding(imputil.Import.MEMBER, 'bar', 'bar')
    self._check_imports('from foo import bar', [imp])

  def testImportFromMultiple(self):
    imp1 = copy.deepcopy(self.baz2_import)
    imp1.add_binding(imputil.Import.MODULE, 'baz', 1)
    imp2 = copy.deepcopy(self.foo2_import)
    imp2.add_binding(imputil.Import.MODULE, 'foo', 1)
    self._check_imports('from bar import baz, foo', [imp1, imp2])

  def testImportFromMixedMembers(self):
    imp1 = copy.deepcopy(self.bar_import)
    imp1.add_binding(imputil.Import.MEMBER, 'qux', 'qux')
    imp2 = copy.deepcopy(self.baz2_import)
    imp2.add_binding(imputil.Import.MODULE, 'baz', 1)
    self._check_imports('from bar import qux, baz', [imp1, imp2])

  def testImportFromAs(self):
    imp = copy.deepcopy(self.baz2_import)
    imp.add_binding(imputil.Import.MODULE, 'qux', 1)
    self._check_imports('from bar import baz as qux', [imp])

  def testImportFromAsMembers(self):
    imp = copy.deepcopy(self.foo_import)
    imp.add_binding(imputil.Import.MEMBER, 'baz', 'bar')
    self._check_imports('from foo import bar as baz', [imp])

  def testImportFromWildcardRaises(self):
    self.assertRaises(util.ImportError, self.importer.visit,
                      pythonparser.parse('from foo import *').body[0])

  def testImportFromFuture(self):
    self._check_imports('from __future__ import print_function', [])

  def testImportFromNative(self):
    imp = imputil.Import('__go__/fmt', is_native=True)
    imp.add_binding(imputil.Import.MEMBER, 'Printf', 'Printf')
    self._check_imports('from "__go__/fmt" import Printf', [imp])

  def testImportFromNativeMultiple(self):
    imp = imputil.Import('__go__/fmt', is_native=True)
    imp.add_binding(imputil.Import.MEMBER, 'Printf', 'Printf')
    imp.add_binding(imputil.Import.MEMBER, 'Println', 'Println')
    self._check_imports('from "__go__/fmt" import Printf, Println', [imp])

  def testImportFromNativeAs(self):
    imp = imputil.Import('__go__/fmt', is_native=True)
    imp.add_binding(imputil.Import.MEMBER, 'foo', 'Printf')
    self._check_imports('from "__go__/fmt" import Printf as foo', [imp])

  def testRelativeImportNonPackage(self):
    self.assertRaises(util.ImportError, self.importer.visit,
                      pythonparser.parse('from . import bar').body[0])

  def testRelativeImportBeyondTopLevel(self):
    self.assertRaises(util.ImportError, self.bar_importer.visit,
                      pythonparser.parse('from .. import qux').body[0])

  def testRelativeModuleNoExist(self):
    self.assertRaises(util.ImportError, self.bar_importer.visit,
                      pythonparser.parse('from . import qux').body[0])

  def testRelativeModule(self):
    imp = copy.deepcopy(self.foo2_import)
    imp.add_binding(imputil.Import.MODULE, 'foo', 1)
    node = pythonparser.parse('from . import foo').body[0]
    self._assert_imports_equal([imp], self.bar_importer.visit(node))

  def testRelativeModuleFromSubModule(self):
    imp = copy.deepcopy(self.foo2_import)
    imp.add_binding(imputil.Import.MODULE, 'foo', 1)
    baz_script = os.path.join(self.pydir, 'bar', 'baz.py')
    importer = imputil.Importer(self.rootdir, 'bar.baz', baz_script, False)
    node = pythonparser.parse('from . import foo').body[0]
    self._assert_imports_equal([imp], importer.visit(node))

  def testRelativeModuleMember(self):
    imp = copy.deepcopy(self.foo2_import)
    imp.add_binding(imputil.Import.MEMBER, 'qux', 'qux')
    node = pythonparser.parse('from .foo import qux').body[0]
    self._assert_imports_equal([imp], self.bar_importer.visit(node))

  def testRelativeModuleMemberMixed(self):
    imp1 = copy.deepcopy(self.fred_import)
    imp1.add_binding(imputil.Import.MEMBER, 'qux', 'qux')
    imp2 = copy.deepcopy(self.quux_import)
    imp2.add_binding(imputil.Import.MODULE, 'quux', 2)
    node = pythonparser.parse('from .fred import qux, quux').body[0]
    self._assert_imports_equal([imp1, imp2], self.bar_importer.visit(node))

  def testRelativeUpLevel(self):
    imp = copy.deepcopy(self.foo2_import)
    imp.add_binding(imputil.Import.MODULE, 'foo', 1)
    node = pythonparser.parse('from .. import foo').body[0]
    self._assert_imports_equal([imp], self.fred_importer.visit(node))

  def testRelativeUpLevelMember(self):
    imp = copy.deepcopy(self.foo2_import)
    imp.add_binding(imputil.Import.MEMBER, 'qux', 'qux')
    node = pythonparser.parse('from ..foo import qux').body[0]
    self._assert_imports_equal([imp], self.fred_importer.visit(node))

  def _check_imports(self, stmt, want):
    got = self.importer.visit(pythonparser.parse(stmt).body[0])
    self._assert_imports_equal(want, got)

  def _assert_imports_equal(self, want, got):
    self.assertEqual([imp.__dict__ for imp in want],
                     [imp.__dict__ for imp in got])

  def _materialize_tree(self, dirname, spec):
    for name, sub_spec in spec.iteritems():
      if name.endswith('/'):
        subdir = os.path.join(dirname, name[:-1])
        os.mkdir(subdir)
        self._materialize_tree(subdir, sub_spec)
      else:
        with open(os.path.join(dirname, name), 'w'):
          pass


class MakeFutureFeaturesTest(unittest.TestCase):

  def testImportFromFuture(self):
    testcases = [
        ('from __future__ import print_function',
         imputil.FutureFeatures(print_function=True)),
        ('from __future__ import generators', imputil.FutureFeatures()),
        ('from __future__ import generators, print_function',
         imputil.FutureFeatures(print_function=True)),
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
        ('from __future__ import division',
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

  def testFutureFeatures(self):
    testcases = [
        ('from __future__ import print_function',
         imputil.FutureFeatures(print_function=True)),
        ("""\
        "module docstring"

        from __future__ import print_function
        """, imputil.FutureFeatures(print_function=True)),
        ("""\
        "module docstring"

        from __future__ import print_function, with_statement
        from __future__ import nested_scopes
        """, imputil.FutureFeatures(print_function=True)),
        ('from __future__ import absolute_import',
         imputil.FutureFeatures(absolute_import=True)),
        ('from __future__ import absolute_import, print_function',
         imputil.FutureFeatures(absolute_import=True, print_function=True)),
        ('foo = 123\nfrom __future__ import print_function',
         imputil.FutureFeatures()),
        ('import os\nfrom __future__ import print_function',
         imputil.FutureFeatures()),
    ]

    for tc in testcases:
      source, want = tc
      mod = pythonparser.parse(textwrap.dedent(source))
      _, got = imputil.parse_future_features(mod)
      self.assertEqual(want.__dict__, got.__dict__)

  def testUnimplementedFutureRaises(self):
    mod = pythonparser.parse('from __future__ import division')
    msg = 'future feature division not yet implemented by grumpy'
    self.assertRaisesRegexp(util.ParseError, msg,
                            imputil.parse_future_features, mod)

  def testUndefinedFutureRaises(self):
    mod = pythonparser.parse('from __future__ import foo')
    self.assertRaisesRegexp(
        util.ParseError, 'future feature foo is not defined',
        imputil.parse_future_features, mod)


if __name__ == '__main__':
  unittest.main()

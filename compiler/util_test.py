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

"""Tests Writer and other utils."""

from __future__ import unicode_literals

import unittest

import pythonparser

from grumpy.compiler import block
from grumpy.compiler import util
from grumpy.compiler import stmt


class ImportVisitorTest(unittest.TestCase):

  def testImport(self):
    imp = util.Import('foo')
    imp.add_binding(util.Import.MODULE, 'foo', util.Import.ROOT)
    self._assert_imports_equal(imp, self._visit_import('import foo'))

  def testImportMultiple(self):
    imp1 = util.Import('foo')
    imp1.add_binding(util.Import.MODULE, 'foo', util.Import.ROOT)
    imp2 = util.Import('bar')
    imp2.add_binding(util.Import.MODULE, 'bar', util.Import.ROOT)
    self._assert_imports_equal(
        [imp1, imp2], self._visit_import('import foo, bar'))

  def testImportAs(self):
    imp = util.Import('foo')
    imp.add_binding(util.Import.MODULE, 'bar', util.Import.LEAF)
    self._assert_imports_equal(imp, self._visit_import('import foo as bar'))

  def testImportNativeRaises(self):
    self.assertRaises(util.ImportError, self._visit_import, 'import __go__.fmt')

  def testImportFrom(self):
    imp = util.Import('foo.bar')
    imp.add_binding(util.Import.MODULE, 'bar', util.Import.LEAF)
    self._assert_imports_equal(imp, self._visit_import('from foo import bar'))

  def testImportFromMultiple(self):
    imp1 = util.Import('foo.bar')
    imp1.add_binding(util.Import.MODULE, 'bar', util.Import.LEAF)
    imp2 = util.Import('foo.baz')
    imp2.add_binding(util.Import.MODULE, 'baz', util.Import.LEAF)
    self._assert_imports_equal(
        [imp1, imp2], self._visit_import('from foo import bar, baz'))

  def testImportFromAs(self):
    imp = util.Import('foo.bar')
    imp.add_binding(util.Import.MODULE, 'baz', util.Import.LEAF)
    self._assert_imports_equal(
        imp, self._visit_import('from foo import bar as baz'))

  def testImportFromWildcardRaises(self):
    self.assertRaises(util.ImportError, self._visit_import, 'from foo import *')

  def testImportFromFuture(self):
    result = self._visit_import('from __future__ import print_function')
    self.assertEqual([], result)

  def testImportFromNative(self):
    imp = util.Import('fmt', is_native=True)
    imp.add_binding(util.Import.MEMBER, 'Printf', 'Printf')
    self._assert_imports_equal(
        imp, self._visit_import('from __go__.fmt import Printf'))

  def testImportFromNativeMultiple(self):
    imp = util.Import('fmt', is_native=True)
    imp.add_binding(util.Import.MEMBER, 'Printf', 'Printf')
    imp.add_binding(util.Import.MEMBER, 'Println', 'Println')
    self._assert_imports_equal(
        imp, self._visit_import('from __go__.fmt import Printf, Println'))

  def testImportFromNativeAs(self):
    imp = util.Import('fmt', is_native=True)
    imp.add_binding(util.Import.MEMBER, 'foo', 'Printf')
    self._assert_imports_equal(
        imp, self._visit_import('from __go__.fmt import Printf as foo'))

  def _visit_import(self, source):
    return util.ImportVisitor().visit(pythonparser.parse(source).body[0])

  def _assert_imports_equal(self, want, got):
    if isinstance(want, util.Import):
      want = [want]
    self.assertEqual([imp.__dict__ for imp in want],
                     [imp.__dict__ for imp in got])


class WriterTest(unittest.TestCase):

  def testIndentBlock(self):
    writer = util.Writer()
    writer.write('foo')
    with writer.indent_block(n=2):
      writer.write('bar')
    writer.write('baz')
    self.assertEqual(writer.getvalue(), 'foo\n\t\tbar\nbaz\n')

  def testWriteBlock(self):
    writer = util.Writer()
    mod_block = block.ModuleBlock('__main__', '<test>', '',
                                  stmt.FutureFeatures())
    writer.write_block(mod_block, 'BODY')
    output = writer.getvalue()
    dispatch = 'switch πF.State() {\n\tcase 0:\n\tdefault: panic'
    self.assertIn(dispatch, output)
    self.assertIn('return nil, nil\n}', output)

  def testWriteImportBlockEmptyImports(self):
    writer = util.Writer()
    writer.write_import_block({})
    self.assertEqual(writer.getvalue(), '')

  def testWriteImportBlockImportsSorted(self):
    writer = util.Writer()
    imports = {name: block.Package(name) for name in ('a', 'b', 'c')}
    writer.write_import_block(imports)
    self.assertEqual(writer.getvalue(),
                     'import (\n\tπ_a "a"\n\tπ_b "b"\n\tπ_c "c"\n)\n')

  def testWriteMultiline(self):
    writer = util.Writer()
    writer.indent(2)
    writer.write('foo\nbar\nbaz\n')
    self.assertEqual(writer.getvalue(), '\t\tfoo\n\t\tbar\n\t\tbaz\n')

  def testWritePyContext(self):
    writer = util.Writer()
    writer.write_py_context(12, 'print "foo"')
    self.assertEqual(writer.getvalue(), '// line 12: print "foo"\n')

  def testWriteSkipBlankLine(self):
    writer = util.Writer()
    writer.write('foo\n\nbar')
    self.assertEqual(writer.getvalue(), 'foo\nbar\n')

  def testWriteTmpl(self):
    writer = util.Writer()
    writer.write_tmpl('$foo, $bar\n$baz', foo=1, bar=2, baz=3)
    self.assertEqual(writer.getvalue(), '1, 2\n3\n')

  def testIndent(self):
    writer = util.Writer()
    writer.indent(2)
    writer.write('foo')
    self.assertEqual(writer.getvalue(), '\t\tfoo\n')

  def testDedent(self):
    writer = util.Writer()
    writer.indent(4)
    writer.dedent(3)
    writer.write('foo')
    self.assertEqual(writer.getvalue(), '\tfoo\n')


if __name__ == '__main__':
  unittest.main()

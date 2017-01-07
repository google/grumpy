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

import unittest

from grumpy.compiler import block
from grumpy.compiler import util
from grumpy.compiler import stmt


class WriterTest(unittest.TestCase):

  def testIndentBlock(self):
    writer = util.Writer()
    writer.write('foo')
    with writer.indent_block(n=2):
      writer.write('bar')
    writer.write('baz')
    self.assertEqual(writer.out.getvalue(), 'foo\n\t\tbar\nbaz\n')

  def testWriteBlock(self):
    writer = util.Writer()
    mod_block = block.ModuleBlock('__main__', 'grumpy', 'grumpy/lib', '<test>',
                                  [], stmt.FutureFeatures())
    writer.write_block(mod_block, 'BODY')
    output = writer.out.getvalue()
    dispatch = 'switch πF.State() {\n\tcase 0:\n\tdefault: panic'
    self.assertIn(dispatch, output)
    self.assertIn('return nil, nil\n}', output)

  def testWriteImportBlockEmptyImports(self):
    writer = util.Writer()
    writer.write_import_block({})
    self.assertEqual(writer.out.getvalue(), '')

  def testWriteImportBlockImportsSorted(self):
    writer = util.Writer()
    imports = {name: block.Package(name) for name in ('a', 'b', 'c')}
    writer.write_import_block(imports)
    self.assertEqual(writer.out.getvalue(),
                     'import (\n\tπ_a "a"\n\tπ_b "b"\n\tπ_c "c"\n)\n')

  def testWriteMultiline(self):
    writer = util.Writer()
    writer.indent(2)
    writer.write('foo\nbar\nbaz\n')
    self.assertEqual(writer.out.getvalue(), '\t\tfoo\n\t\tbar\n\t\tbaz\n')

  def testWritePyContext(self):
    writer = util.Writer()
    writer.write_py_context(12, 'print "foo"')
    self.assertEqual(writer.out.getvalue(), '// line 12: print "foo"\n')

  def testWriteSkipBlankLine(self):
    writer = util.Writer()
    writer.write('foo\n\nbar')
    self.assertEqual(writer.out.getvalue(), 'foo\nbar\n')

  def testWriteTmpl(self):
    writer = util.Writer()
    writer.write_tmpl('$foo, $bar\n$baz', foo=1, bar=2, baz=3)
    self.assertEqual(writer.out.getvalue(), '1, 2\n3\n')

  def testIndent(self):
    writer = util.Writer()
    writer.indent(2)
    writer.write('foo')
    self.assertEqual(writer.out.getvalue(), '\t\tfoo\n')

  def testDedent(self):
    writer = util.Writer()
    writer.indent(4)
    writer.dedent(3)
    writer.write('foo')
    self.assertEqual(writer.out.getvalue(), '\tfoo\n')


if __name__ == '__main__':
  unittest.main()

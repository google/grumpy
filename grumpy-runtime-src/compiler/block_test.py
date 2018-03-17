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

"""Tests Package, Block, BlockVisitor and related classes."""

from __future__ import unicode_literals

import textwrap
import unittest

from grumpy.compiler import block
from grumpy.compiler import imputil
from grumpy.compiler import util
from grumpy import pythonparser

class PackageTest(unittest.TestCase):

  def testCreate(self):
    package = block.Package('foo/bar/baz')
    self.assertEqual(package.name, 'foo/bar/baz')
    self.assertEqual(package.alias, 'π_fooΓbarΓbaz')

  def testCreateGrump(self):
    package = block.Package('foo/bar/baz', 'myalias')
    self.assertEqual(package.name, 'foo/bar/baz')
    self.assertEqual(package.alias, 'myalias')


class BlockTest(unittest.TestCase):

  def testLoop(self):
    b = _MakeModuleBlock()
    loop = b.push_loop(None)
    self.assertEqual(loop, b.top_loop())
    inner_loop = b.push_loop(None)
    self.assertEqual(inner_loop, b.top_loop())
    b.pop_loop()
    self.assertEqual(loop, b.top_loop())

  def testResolveName(self):
    module_block = _MakeModuleBlock()
    block_vars = {'foo': block.Var('foo', block.Var.TYPE_LOCAL)}
    func1_block = block.FunctionBlock(module_block, 'func1', block_vars, False)
    block_vars = {'bar': block.Var('bar', block.Var.TYPE_LOCAL)}
    func2_block = block.FunctionBlock(func1_block, 'func2', block_vars, False)
    block_vars = {'case': block.Var('case', block.Var.TYPE_LOCAL)}
    keyword_block = block.FunctionBlock(
        module_block, 'keyword_func', block_vars, False)
    class1_block = block.ClassBlock(module_block, 'Class1', set())
    class2_block = block.ClassBlock(func1_block, 'Class2', set())
    self.assertRegexpMatches(self._ResolveName(module_block, 'foo'),
                             r'ResolveGlobal\b.*foo')
    self.assertRegexpMatches(self._ResolveName(module_block, 'bar'),
                             r'ResolveGlobal\b.*bar')
    self.assertRegexpMatches(self._ResolveName(module_block, 'baz'),
                             r'ResolveGlobal\b.*baz')
    self.assertRegexpMatches(self._ResolveName(func1_block, 'foo'),
                             r'CheckLocal\b.*foo')
    self.assertRegexpMatches(self._ResolveName(func1_block, 'bar'),
                             r'ResolveGlobal\b.*bar')
    self.assertRegexpMatches(self._ResolveName(func1_block, 'baz'),
                             r'ResolveGlobal\b.*baz')
    self.assertRegexpMatches(self._ResolveName(func2_block, 'foo'),
                             r'CheckLocal\b.*foo')
    self.assertRegexpMatches(self._ResolveName(func2_block, 'bar'),
                             r'CheckLocal\b.*bar')
    self.assertRegexpMatches(self._ResolveName(func2_block, 'baz'),
                             r'ResolveGlobal\b.*baz')
    self.assertRegexpMatches(self._ResolveName(class1_block, 'foo'),
                             r'ResolveClass\(.*, nil, .*foo')
    self.assertRegexpMatches(self._ResolveName(class2_block, 'foo'),
                             r'ResolveClass\(.*, µfoo, .*foo')
    self.assertRegexpMatches(self._ResolveName(keyword_block, 'case'),
                             r'CheckLocal\b.*µcase, "case"')

  def _ResolveName(self, b, name):
    writer = util.Writer()
    b.resolve_name(writer, name)
    return writer.getvalue()


class BlockVisitorTest(unittest.TestCase):

  def testAssignSingle(self):
    visitor = block.BlockVisitor()
    visitor.visit(_ParseStmt('foo = 3'))
    self.assertEqual(visitor.vars.keys(), ['foo'])
    self.assertRegexpMatches(visitor.vars['foo'].init_expr, r'UnboundLocal')

  def testAssignMultiple(self):
    visitor = block.BlockVisitor()
    visitor.visit(_ParseStmt('foo = bar = 123'))
    self.assertEqual(sorted(visitor.vars.keys()), ['bar', 'foo'])
    self.assertRegexpMatches(visitor.vars['foo'].init_expr, r'UnboundLocal')
    self.assertRegexpMatches(visitor.vars['bar'].init_expr, r'UnboundLocal')

  def testAssignTuple(self):
    visitor = block.BlockVisitor()
    visitor.visit(_ParseStmt('foo, bar = "a", "b"'))
    self.assertEqual(sorted(visitor.vars.keys()), ['bar', 'foo'])
    self.assertRegexpMatches(visitor.vars['foo'].init_expr, r'UnboundLocal')
    self.assertRegexpMatches(visitor.vars['bar'].init_expr, r'UnboundLocal')

  def testAssignNested(self):
    visitor = block.BlockVisitor()
    visitor.visit(_ParseStmt('foo, (bar, baz) = "a", ("b", "c")'))
    self.assertEqual(sorted(visitor.vars.keys()), ['bar', 'baz', 'foo'])
    self.assertRegexpMatches(visitor.vars['foo'].init_expr, r'UnboundLocal')
    self.assertRegexpMatches(visitor.vars['bar'].init_expr, r'UnboundLocal')
    self.assertRegexpMatches(visitor.vars['baz'].init_expr, r'UnboundLocal')

  def testAugAssignSingle(self):
    visitor = block.BlockVisitor()
    visitor.visit(_ParseStmt('foo += 3'))
    self.assertEqual(visitor.vars.keys(), ['foo'])
    self.assertRegexpMatches(visitor.vars['foo'].init_expr, r'UnboundLocal')

  def testVisitClassDef(self):
    visitor = block.BlockVisitor()
    visitor.visit(_ParseStmt('class Foo(object): pass'))
    self.assertEqual(visitor.vars.keys(), ['Foo'])
    self.assertRegexpMatches(visitor.vars['Foo'].init_expr, r'UnboundLocal')

  def testExceptHandler(self):
    visitor = block.BlockVisitor()
    visitor.visit(_ParseStmt(textwrap.dedent("""\
        try:
          pass
        except Exception as foo:
          pass
        except TypeError as bar:
          pass""")))
    self.assertEqual(sorted(visitor.vars.keys()), ['bar', 'foo'])
    self.assertRegexpMatches(visitor.vars['foo'].init_expr, r'UnboundLocal')
    self.assertRegexpMatches(visitor.vars['bar'].init_expr, r'UnboundLocal')

  def testFor(self):
    visitor = block.BlockVisitor()
    visitor.visit(_ParseStmt('for i in foo: pass'))
    self.assertEqual(visitor.vars.keys(), ['i'])
    self.assertRegexpMatches(visitor.vars['i'].init_expr, r'UnboundLocal')

  def testFunctionDef(self):
    visitor = block.BlockVisitor()
    visitor.visit(_ParseStmt('def foo(): pass'))
    self.assertEqual(visitor.vars.keys(), ['foo'])
    self.assertRegexpMatches(visitor.vars['foo'].init_expr, r'UnboundLocal')

  def testImport(self):
    visitor = block.BlockVisitor()
    visitor.visit(_ParseStmt('import foo.bar, baz'))
    self.assertEqual(sorted(visitor.vars.keys()), ['baz', 'foo'])
    self.assertRegexpMatches(visitor.vars['foo'].init_expr, r'UnboundLocal')
    self.assertRegexpMatches(visitor.vars['baz'].init_expr, r'UnboundLocal')

  def testImportFrom(self):
    visitor = block.BlockVisitor()
    visitor.visit(_ParseStmt('from foo.bar import baz, qux'))
    self.assertEqual(sorted(visitor.vars.keys()), ['baz', 'qux'])
    self.assertRegexpMatches(visitor.vars['baz'].init_expr, r'UnboundLocal')
    self.assertRegexpMatches(visitor.vars['qux'].init_expr, r'UnboundLocal')

  def testGlobal(self):
    visitor = block.BlockVisitor()
    visitor.visit(_ParseStmt('global foo, bar'))
    self.assertEqual(sorted(visitor.vars.keys()), ['bar', 'foo'])
    self.assertIsNone(visitor.vars['foo'].init_expr)
    self.assertIsNone(visitor.vars['bar'].init_expr)

  def testGlobalIsParam(self):
    visitor = block.BlockVisitor()
    visitor.vars['foo'] = block.Var('foo', block.Var.TYPE_PARAM, arg_index=0)
    self.assertRaisesRegexp(util.ParseError, 'is parameter and global',
                            visitor.visit, _ParseStmt('global foo'))

  def testGlobalUsedPriorToDeclaration(self):
    node = pythonparser.parse('foo = 42\nglobal foo')
    visitor = block.BlockVisitor()
    self.assertRaisesRegexp(util.ParseError, 'used prior to global declaration',
                            visitor.generic_visit, node)


class FunctionBlockVisitorTest(unittest.TestCase):

  def testArgs(self):
    func = _ParseStmt('def foo(bar, baz, *args, **kwargs): pass')
    visitor = block.FunctionBlockVisitor(func)
    self.assertIn('bar', visitor.vars)
    self.assertIn('baz', visitor.vars)
    self.assertIn('args', visitor.vars)
    self.assertIn('kwargs', visitor.vars)
    self.assertRegexpMatches(visitor.vars['bar'].init_expr, r'Args\[0\]')
    self.assertRegexpMatches(visitor.vars['baz'].init_expr, r'Args\[1\]')
    self.assertRegexpMatches(visitor.vars['args'].init_expr, r'Args\[2\]')
    self.assertRegexpMatches(visitor.vars['kwargs'].init_expr, r'Args\[3\]')

  def testArgsDuplicate(self):
    func = _ParseStmt('def foo(bar, baz, bar=None): pass')
    self.assertRaisesRegexp(util.ParseError, 'duplicate argument',
                            block.FunctionBlockVisitor, func)

  def testYield(self):
    visitor = block.FunctionBlockVisitor(_ParseStmt('def foo(): pass'))
    visitor.visit(_ParseStmt('yield "foo"'))
    self.assertTrue(visitor.is_generator)

  def testYieldExpr(self):
    visitor = block.FunctionBlockVisitor(_ParseStmt('def foo(): pass'))
    visitor.visit(_ParseStmt('foo = (yield)'))
    self.assertTrue(visitor.is_generator)
    self.assertEqual(sorted(visitor.vars.keys()), ['foo'])
    self.assertRegexpMatches(visitor.vars['foo'].init_expr, r'UnboundLocal')


def _MakeModuleBlock():
  importer = imputil.Importer(None, '__main__', '/tmp/foo.py', False)
  return block.ModuleBlock(importer, '__main__', '<test>', '',
                           imputil.FutureFeatures())


def _ParseStmt(stmt_str):
  return pythonparser.parse(stmt_str).body[0]


if __name__ == '__main__':
  unittest.main()

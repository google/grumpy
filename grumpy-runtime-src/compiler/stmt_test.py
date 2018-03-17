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

"""Tests for StatementVisitor."""

from __future__ import unicode_literals

import re
import subprocess
import textwrap
import unittest

from grumpy.compiler import block
from grumpy.compiler import imputil
from grumpy.compiler import shard_test
from grumpy.compiler import stmt
from grumpy.compiler import util
from grumpy import pythonparser
from grumpy.pythonparser import ast


class StatementVisitorTest(unittest.TestCase):

  def testAssertNoMsg(self):
    self.assertEqual((0, 'AssertionError()\n'), _GrumpRun(textwrap.dedent("""\
        try:
          assert False
        except AssertionError as e:
          print repr(e)""")))

  def testAssertMsg(self):
    want = (0, "AssertionError('foo',)\n")
    self.assertEqual(want, _GrumpRun(textwrap.dedent("""\
        try:
          assert False, 'foo'
        except AssertionError as e:
          print repr(e)""")))

  def testBareAssert(self):
    # Assertion errors at the top level of a block should raise:
    # https://github.com/google/grumpy/issues/18
    want = (0, 'ok\n')
    self.assertEqual(want, _GrumpRun(textwrap.dedent("""\
        def foo():
         assert False
        try:
         foo()
        except AssertionError:
         print 'ok'
        else:
         print 'bad'""")))

  def testAssignAttribute(self):
    self.assertEqual((0, '123\n'), _GrumpRun(textwrap.dedent("""\
        e = Exception()
        e.foo = 123
        print e.foo""")))

  def testAssignName(self):
    self.assertEqual((0, 'bar\n'), _GrumpRun(textwrap.dedent("""\
        foo = 'bar'
        print foo""")))

  def testAssignMultiple(self):
    self.assertEqual((0, 'baz baz\n'), _GrumpRun(textwrap.dedent("""\
        foo = bar = 'baz'
        print foo, bar""")))

  def testAssignSubscript(self):
    self.assertEqual((0, "{'bar': None}\n"), _GrumpRun(textwrap.dedent("""\
        foo = {}
        foo['bar'] = None
        print foo""")))

  def testAssignTuple(self):
    self.assertEqual((0, 'a b\n'), _GrumpRun(textwrap.dedent("""\
        baz = ('a', 'b')
        foo, bar = baz
        print foo, bar""")))

  def testAugAssign(self):
    self.assertEqual((0, '42\n'), _GrumpRun(textwrap.dedent("""\
        foo = 41
        foo += 1
        print foo""")))

  def testAugAssignBitAnd(self):
    self.assertEqual((0, '3\n'), _GrumpRun(textwrap.dedent("""\
        foo = 7
        foo &= 3
        print foo""")))

  def testAugAssignPow(self):
    self.assertEqual((0, '64\n'), _GrumpRun(textwrap.dedent("""\
        foo = 8
        foo **= 2
        print foo""")))

  def testClassDef(self):
    self.assertEqual((0, "<type 'type'>\n"), _GrumpRun(textwrap.dedent("""\
        class Foo(object):
          pass
        print type(Foo)""")))

  def testClassDefWithVar(self):
    self.assertEqual((0, 'abc\n'), _GrumpRun(textwrap.dedent("""\
        class Foo(object):
          bar = 'abc'
        print Foo.bar""")))

  def testDeleteAttribute(self):
    self.assertEqual((0, 'False\n'), _GrumpRun(textwrap.dedent("""\
        class Foo(object):
          bar = 42
        del Foo.bar
        print hasattr(Foo, 'bar')""")))

  def testDeleteClassLocal(self):
    self.assertEqual((0, 'False\n'), _GrumpRun(textwrap.dedent("""\
        class Foo(object):
          bar = 'baz'
          del bar
        print hasattr(Foo, 'bar')""")))

  def testDeleteGlobal(self):
    self.assertEqual((0, 'False\n'), _GrumpRun(textwrap.dedent("""\
        foo = 42
        del foo
        print 'foo' in globals()""")))

  def testDeleteLocal(self):
    self.assertEqual((0, 'ok\n'), _GrumpRun(textwrap.dedent("""\
        def foo():
          bar = 123
          del bar
          try:
            print bar
            raise AssertionError
          except UnboundLocalError:
            print 'ok'
        foo()""")))

  def testDeleteNonexistentLocal(self):
    self.assertRaisesRegexp(
        util.ParseError, 'cannot delete nonexistent local',
        _ParseAndVisit, 'def foo():\n  del bar')

  def testDeleteSubscript(self):
    self.assertEqual((0, '{}\n'), _GrumpRun(textwrap.dedent("""\
        foo = {'bar': 'baz'}
        del foo['bar']
        print foo""")))

  def testExprCall(self):
    self.assertEqual((0, 'bar\n'), _GrumpRun(textwrap.dedent("""\
        def foo():
          print 'bar'
        foo()""")))

  def testExprNameGlobal(self):
    self.assertEqual((0, ''), _GrumpRun(textwrap.dedent("""\
        foo = 42
        foo""")))

  def testExprNameLocal(self):
    self.assertEqual((0, ''), _GrumpRun(textwrap.dedent("""\
        foo = 42
        def bar():
          foo
        bar()""")))

  def testFor(self):
    self.assertEqual((0, '1\n2\n3\n'), _GrumpRun(textwrap.dedent("""\
        for i in (1, 2, 3):
          print i""")))

  def testForBreak(self):
    self.assertEqual((0, '1\n'), _GrumpRun(textwrap.dedent("""\
        for i in (1, 2, 3):
          print i
          break""")))

  def testForContinue(self):
    self.assertEqual((0, '1\n2\n3\n'), _GrumpRun(textwrap.dedent("""\
        for i in (1, 2, 3):
          print i
          continue
          raise AssertionError""")))

  def testForElse(self):
    self.assertEqual((0, 'foo\nbar\n'), _GrumpRun(textwrap.dedent("""\
        for i in (1,):
          print 'foo'
        else:
          print 'bar'""")))

  def testForElseBreakNotNested(self):
    self.assertRaisesRegexp(
        util.ParseError, "'continue' not in loop",
        _ParseAndVisit, 'for i in (1,):\n  pass\nelse:\n  continue')

  def testForElseContinueNotNested(self):
    self.assertRaisesRegexp(
        util.ParseError, "'continue' not in loop",
        _ParseAndVisit, 'for i in (1,):\n  pass\nelse:\n  continue')

  def testFunctionDecorator(self):
    self.assertEqual((0, '<b>foo</b>\n'), _GrumpRun(textwrap.dedent("""\
        def bold(fn):
          return lambda: '<b>' + fn() + '</b>'
        @bold
        def foo():
          return 'foo'
        print foo()""")))

  def testFunctionDecoratorWithArg(self):
    self.assertEqual((0, '<b id=red>foo</b>\n'), _GrumpRun(textwrap.dedent("""\
        def tag(name):
          def bold(fn):
            return lambda: '<b id=' + name + '>' + fn() + '</b>'
          return bold
        @tag('red')
        def foo():
          return 'foo'
        print foo()""")))

  def testFunctionDef(self):
    self.assertEqual((0, 'bar baz\n'), _GrumpRun(textwrap.dedent("""\
        def foo(a, b):
          print a, b
        foo('bar', 'baz')""")))

  def testFunctionDefGenerator(self):
    self.assertEqual((0, "['foo', 'bar']\n"), _GrumpRun(textwrap.dedent("""\
        def gen():
          yield 'foo'
          yield 'bar'
        print list(gen())""")))

  def testFunctionDefGeneratorReturnValue(self):
    self.assertRaisesRegexp(
        util.ParseError, 'returning a value in a generator function',
        _ParseAndVisit, 'def foo():\n  yield 1\n  return 2')

  def testFunctionDefLocal(self):
    self.assertEqual((0, 'baz\n'), _GrumpRun(textwrap.dedent("""\
        def foo():
          def bar():
            print 'baz'
          bar()
        foo()""")))

  def testIf(self):
    self.assertEqual((0, 'foo\n'), _GrumpRun(textwrap.dedent("""\
        if 123:
          print 'foo'
        if '':
          print 'bar'""")))

  def testIfElif(self):
    self.assertEqual((0, 'foo\nbar\n'), _GrumpRun(textwrap.dedent("""\
        if True:
          print 'foo'
        elif False:
          print 'bar'
        if False:
          print 'foo'
        elif True:
          print 'bar'""")))

  def testIfElse(self):
    self.assertEqual((0, 'foo\nbar\n'), _GrumpRun(textwrap.dedent("""\
        if True:
          print 'foo'
        else:
          print 'bar'
        if False:
          print 'foo'
        else:
          print 'bar'""")))

  def testImport(self):
    self.assertEqual((0, "<type 'dict'>\n"), _GrumpRun(textwrap.dedent("""\
        import sys
        print type(sys.modules)""")))

  def testImportFutureLateRaises(self):
    regexp = 'from __future__ imports must occur at the beginning of the file'
    self.assertRaisesRegexp(util.ImportError, regexp, _ParseAndVisit,
                            'foo = bar\nfrom __future__ import print_function')

  def testFutureUnicodeLiterals(self):
    want = "u'foo'\n"
    self.assertEqual((0, want), _GrumpRun(textwrap.dedent("""\
        from __future__ import unicode_literals
        print repr('foo')""")))

  def testImportMember(self):
    self.assertEqual((0, "<type 'dict'>\n"), _GrumpRun(textwrap.dedent("""\
        from sys import modules
        print type(modules)""")))

  def testImportConflictingPackage(self):
    self.assertEqual((0, ''), _GrumpRun(textwrap.dedent("""\
        import time
        from "__go__/time" import Now""")))

  def testImportNative(self):
    self.assertEqual((0, '1 1000000000\n'), _GrumpRun(textwrap.dedent("""\
        from "__go__/time" import Nanosecond, Second
        print Nanosecond, Second""")))

  def testImportGrumpy(self):
    self.assertEqual((0, ''), _GrumpRun(textwrap.dedent("""\
        from "__go__/grumpy" import Assert
        Assert(__frame__(), True, 'bad')""")))

  def testImportNativeType(self):
    self.assertEqual((0, "<type 'Duration'>\n"), _GrumpRun(textwrap.dedent("""\
        from "__go__/time" import Duration
        print Duration""")))

  def testImportWildcardMemberRaises(self):
    regexp = 'wildcard member import is not implemented'
    self.assertRaisesRegexp(util.ImportError, regexp, _ParseAndVisit,
                            'from foo import *')
    self.assertRaisesRegexp(util.ImportError, regexp, _ParseAndVisit,
                            'from "__go__/foo" import *')

  def testPrintStatement(self):
    self.assertEqual((0, 'abc 123\nfoo bar\n'), _GrumpRun(textwrap.dedent("""\
        print 'abc',
        print '123'
        print 'foo', 'bar'""")))

  def testPrintFunction(self):
    want = "abc\n123\nabc 123\nabcx123\nabc 123 "
    self.assertEqual((0, want), _GrumpRun(textwrap.dedent("""\
        "module docstring is ok to proceed __future__"
        from __future__ import print_function
        print('abc')
        print(123)
        print('abc', 123)
        print('abc', 123, sep='x')
        print('abc', 123, end=' ')""")))

  def testRaiseExitStatus(self):
    self.assertEqual(1, _GrumpRun('raise Exception')[0])

  def testRaiseInstance(self):
    self.assertEqual((0, 'foo\n'), _GrumpRun(textwrap.dedent("""\
        try:
          raise RuntimeError('foo')
          print 'bad'
        except RuntimeError as e:
          print e""")))

  def testRaiseTypeAndArg(self):
    self.assertEqual((0, 'foo\n'), _GrumpRun(textwrap.dedent("""\
        try:
          raise KeyError('foo')
          print 'bad'
        except KeyError as e:
          print e""")))

  def testRaiseAgain(self):
    self.assertEqual((0, 'foo\n'), _GrumpRun(textwrap.dedent("""\
        try:
          try:
            raise AssertionError('foo')
          except AssertionError:
            raise
        except Exception as e:
          print e""")))

  def testRaiseTraceback(self):
    self.assertEqual((0, ''), _GrumpRun(textwrap.dedent("""\
        import sys
        try:
          try:
            raise Exception
          except:
            e, _, tb = sys.exc_info()
            raise e, None, tb
        except:
          e2, _, tb2 = sys.exc_info()
        assert e is e2
        assert tb is tb2""")))

  def testReturn(self):
    self.assertEqual((0, 'bar\n'), _GrumpRun(textwrap.dedent("""\
        def foo():
          return 'bar'
        print foo()""")))

  def testTryBareExcept(self):
    self.assertEqual((0, ''), _GrumpRun(textwrap.dedent("""\
        try:
          raise AssertionError
        except:
          pass""")))

  def testTryElse(self):
    self.assertEqual((0, 'foo baz\n'), _GrumpRun(textwrap.dedent("""\
        try:
          print 'foo',
        except:
          print 'bar'
        else:
          print 'baz'""")))

  def testTryMultipleExcept(self):
    self.assertEqual((0, 'bar\n'), _GrumpRun(textwrap.dedent("""\
        try:
          raise AssertionError
        except RuntimeError:
          print 'foo'
        except AssertionError:
          print 'bar'
        except:
          print 'baz'""")))

  def testTryFinally(self):
    result = _GrumpRun(textwrap.dedent("""\
        try:
          print 'foo',
        finally:
          print 'bar'
        try:
          print 'foo',
          raise Exception
        finally:
          print 'bar'"""))
    self.assertEqual(1, result[0])
    self.assertIn('foo bar\nfoo bar\n', result[1])
    self.assertIn('Exception\n', result[1])

  def testWhile(self):
    self.assertEqual((0, '2\n1\n'), _GrumpRun(textwrap.dedent("""\
        i = 2
        while i:
          print i
          i -= 1""")))

  def testWhileElse(self):
    self.assertEqual((0, 'bar\n'), _GrumpRun(textwrap.dedent("""\
        while False:
          print 'foo'
        else:
          print 'bar'""")))

  def testWith(self):
    self.assertEqual((0, 'enter\n1\nexit\nenter\n2\nexit\n3\n'),
                     _GrumpRun(textwrap.dedent("""\
        class ContextManager(object):
          def __enter__(self):
            print "enter"

          def __exit__(self, exc_type, value, traceback):
            print "exit"

        a = ContextManager()

        with a:
          print 1

        try:
          with a:
            print 2
            raise RuntimeError
        except RuntimeError:
          print 3
        """)))

  def testWithAs(self):
    self.assertEqual((0, '1 2 3\n'),
                     _GrumpRun(textwrap.dedent("""\
        class ContextManager(object):
          def __enter__(self):
            return (1, (2, 3))
          def __exit__(self, *args):
            pass
        with ContextManager() as [x, (y, z)]:
          print x, y, z
        """)))

  def testWriteExceptDispatcherBareExcept(self):
    visitor = stmt.StatementVisitor(_MakeModuleBlock())
    handlers = [ast.ExceptHandler(type=ast.Name(id='foo')),
                ast.ExceptHandler(type=None)]
    self.assertEqual(visitor._write_except_dispatcher(  # pylint: disable=protected-access
        'exc', 'tb', handlers), [1, 2])
    expected = re.compile(r'ResolveGlobal\(.*foo.*\bIsInstance\(.*'
                          r'goto Label1.*goto Label2', re.DOTALL)
    self.assertRegexpMatches(visitor.writer.getvalue(), expected)

  def testWriteExceptDispatcherBareExceptionNotLast(self):
    visitor = stmt.StatementVisitor(_MakeModuleBlock())
    handlers = [ast.ExceptHandler(type=None),
                ast.ExceptHandler(type=ast.Name(id='foo'))]
    self.assertRaisesRegexp(util.ParseError, r"default 'except:' must be last",
                            visitor._write_except_dispatcher,  # pylint: disable=protected-access
                            'exc', 'tb', handlers)

  def testWriteExceptDispatcherMultipleExcept(self):
    visitor = stmt.StatementVisitor(_MakeModuleBlock())
    handlers = [ast.ExceptHandler(type=ast.Name(id='foo')),
                ast.ExceptHandler(type=ast.Name(id='bar'))]
    self.assertEqual(visitor._write_except_dispatcher(  # pylint: disable=protected-access
        'exc', 'tb', handlers), [1, 2])
    expected = re.compile(
        r'ResolveGlobal\(.*foo.*\bif .*\bIsInstance\(.*\{.*goto Label1.*'
        r'ResolveGlobal\(.*bar.*\bif .*\bIsInstance\(.*\{.*goto Label2.*'
        r'\bRaise\(exc\.ToObject\(\), nil, tb\.ToObject\(\)\)', re.DOTALL)
    self.assertRegexpMatches(visitor.writer.getvalue(), expected)


def _MakeModuleBlock():
  return block.ModuleBlock(None, '__main__', '<test>', '',
                           imputil.FutureFeatures())


def _ParseAndVisit(source):
  mod = pythonparser.parse(source)
  _, future_features = imputil.parse_future_features(mod)
  importer = imputil.Importer(None, 'foo', 'foo.py', False)
  b = block.ModuleBlock(importer, '__main__', '<test>',
                        source, future_features)
  visitor = stmt.StatementVisitor(b)
  visitor.visit(mod)
  return visitor


def _GrumpRun(cmd):
  p = subprocess.Popen(['grumprun'], stdin=subprocess.PIPE,
                       stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
  out, _ = p.communicate(cmd)
  return p.returncode, out


if __name__ == '__main__':
  shard_test.main()

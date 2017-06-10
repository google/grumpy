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

"""Tests for ExprVisitor."""

from __future__ import unicode_literals

import subprocess
import textwrap
import unittest

from grumpy.compiler import block
from grumpy.compiler import imputil
from grumpy.compiler import shard_test
from grumpy.compiler import stmt
from grumpy import pythonparser


def _MakeExprTest(expr):
  def Test(self):
    code = 'assert ({}) == ({!r}), {!r}'.format(expr, eval(expr), expr)  # pylint: disable=eval-used
    self.assertEqual((0, ''), _GrumpRun(code))
  return Test


def _MakeLiteralTest(lit, expected=None):
  if expected is None:
    expected = lit
  def Test(self):
    status, output = _GrumpRun('print repr({}),'.format(lit))
    self.assertEqual(0, status, output)
    self.assertEqual(expected, output.strip())  # pylint: disable=eval-used
  return Test


def _MakeSliceTest(subscript, want):
  """Define a test function that evaluates a slice expression."""
  def Test(self):
    code = textwrap.dedent("""\
        class Slicer(object):
          def __getitem__(self, slice):
            print slice
        Slicer()[{}]""")
    status, output = _GrumpRun(code.format(subscript))
    self.assertEqual(0, status, output)
    self.assertEqual(want, output.strip())
  return Test


class ExprVisitorTest(unittest.TestCase):

  # pylint: disable=invalid-name

  def testAttribute(self):
    code = textwrap.dedent("""\
        class Foo(object):
          bar = 42
        assert Foo.bar == 42""")
    self.assertEqual((0, ''), _GrumpRun(code))

  testBinOpArithmeticAdd = _MakeExprTest('1 + 2')
  testBinOpArithmeticAnd = _MakeExprTest('7 & 12')
  testBinOpArithmeticDiv = _MakeExprTest('8 / 4')
  testBinOpArithmeticFloorDiv = _MakeExprTest('8 // 4')
  testBinOpArithmeticFloorDivRemainder = _MakeExprTest('5 // 2')
  testBinOpArithmeticMod = _MakeExprTest('9 % 5')
  testBinOpArithmeticMul = _MakeExprTest('3 * 2')
  testBinOpArithmeticOr = _MakeExprTest('2 | 6')
  testBinOpArithmeticPow = _MakeExprTest('2 ** 16')
  testBinOpArithmeticSub = _MakeExprTest('10 - 3')
  testBinOpArithmeticXor = _MakeExprTest('3 ^ 5')

  testBoolOpTrueAndFalse = _MakeExprTest('True and False')
  testBoolOpTrueAndTrue = _MakeExprTest('True and True')
  testBoolOpTrueAndExpr = _MakeExprTest('True and 2 == 2')
  testBoolOpTrueOrFalse = _MakeExprTest('True or False')
  testBoolOpFalseOrFalse = _MakeExprTest('False or False')
  testBoolOpFalseOrExpr = _MakeExprTest('False or 2 == 2')

  def testCall(self):
    code = textwrap.dedent("""\
        def foo():
         print 'bar'
        foo()""")
    self.assertEqual((0, 'bar\n'), _GrumpRun(code))

  def testCallKeywords(self):
    code = textwrap.dedent("""\
        def foo(a=1, b=2):
         print a, b
        foo(b=3)""")
    self.assertEqual((0, '1 3\n'), _GrumpRun(code))

  def testCallVarArgs(self):
    code = textwrap.dedent("""\
        def foo(a, b):
          print a, b
        foo(*(123, 'abc'))""")
    self.assertEqual((0, '123 abc\n'), _GrumpRun(code))

  def testCallKwargs(self):
    code = textwrap.dedent("""\
        def foo(a, b=2):
          print a, b
        foo(**{'a': 4})""")
    self.assertEqual((0, '4 2\n'), _GrumpRun(code))

  testCompareLT = _MakeExprTest('1 < 2')
  testCompareLE = _MakeExprTest('7 <= 12')
  testCompareEq = _MakeExprTest('8 == 4')
  testCompareNE = _MakeExprTest('9 != 5')
  testCompareGE = _MakeExprTest('3 >= 2')
  testCompareGT = _MakeExprTest('2 > 6')
  testCompareLTLT = _MakeExprTest('3 < 6 < 9')
  testCompareLTEq = _MakeExprTest('3 < 6 == 9')
  testCompareLTGE = _MakeExprTest('3 < 6 >= -2')
  testCompareGTEq = _MakeExprTest('88 > 12 == 12')
  testCompareInStr = _MakeExprTest('"1" in "abc"')
  testCompareInTuple = _MakeExprTest('1 in (1, 2, 3)')
  testCompareNotInTuple = _MakeExprTest('10 < 12 not in (1, 2, 3)')

  testDictEmpty = _MakeLiteralTest('{}')
  testDictNonEmpty = _MakeLiteralTest("{'foo': 42, 'bar': 43}")

  testSetNonEmpty = _MakeLiteralTest("{'foo', 'bar'}", "set(['foo', 'bar'])")

  testDictCompFor = _MakeExprTest('{x: str(x) for x in range(3)}')
  testDictCompForIf = _MakeExprTest(
      '{x: 3 * x for x in range(10) if x % 3 == 0}')
  testDictCompForFor = _MakeExprTest(
      '{x: y for x in range(3) for y in range(x)}')

  testGeneratorExpFor = _MakeExprTest('tuple(int(x) for x in "123")')
  testGeneratorExpForIf = _MakeExprTest(
      'tuple(x / 3 for x in range(10) if x % 3)')
  testGeneratorExprForFor = _MakeExprTest(
      'tuple(x + y for x in range(3) for y in range(x + 2))')

  testIfExpr = _MakeExprTest('1 if True else 0')
  testIfExprCompound = _MakeExprTest('42 if "ab" == "a" + "b" else 24')
  testIfExprNested = _MakeExprTest(
      '"foo" if "" else "bar" if 0 else "baz"')

  testLambda = _MakeExprTest('(lambda: 123)()')
  testLambda = _MakeExprTest('(lambda a, b: (a, b))("foo", "bar")')
  testLambda = _MakeExprTest('(lambda a, b=3: (a, b))("foo")')
  testLambda = _MakeExprTest('(lambda *args: args)(1, 2, 3)')
  testLambda = _MakeExprTest('(lambda **kwargs: kwargs)(x="foo", y="bar")')

  testListEmpty = _MakeLiteralTest('[]')
  testListNonEmpty = _MakeLiteralTest('[1, 2]')

  testListCompFor = _MakeExprTest('[int(x) for x in "123"]')
  testListCompForIf = _MakeExprTest('[x / 3 for x in range(10) if x % 3]')
  testListCompForFor = _MakeExprTest(
      '[x + y for x in range(3) for y in range(x + 2)]')

  def testNameGlobal(self):
    code = textwrap.dedent("""\
        foo = 123
        assert foo == 123""")
    self.assertEqual((0, ''), _GrumpRun(code))

  def testNameLocal(self):
    code = textwrap.dedent("""\
        def foo():
          bar = 'abc'
          assert bar == 'abc'
        foo()""")
    self.assertEqual((0, ''), _GrumpRun(code))

  testNumInt = _MakeLiteralTest('42')
  testNumLong = _MakeLiteralTest('42L')
  testNumIntLarge = _MakeLiteralTest('12345678901234567890',
                                     '12345678901234567890L')
  testNumFloat = _MakeLiteralTest('102.1')
  testNumFloatOnlyDecimal = _MakeLiteralTest('.5', '0.5')
  testNumFloatNoDecimal = _MakeLiteralTest('5.', '5.0')
  testNumFloatSci = _MakeLiteralTest('1e6', '1000000.0')
  testNumFloatSciCap = _MakeLiteralTest('1E6', '1000000.0')
  testNumFloatSciCapPlus = _MakeLiteralTest('1E+6', '1000000.0')
  testNumFloatSciMinus = _MakeLiteralTest('1e-06')
  testNumComplex = _MakeLiteralTest('3j')

  testSubscriptDictStr = _MakeExprTest('{"foo": 42}["foo"]')
  testSubscriptListInt = _MakeExprTest('[1, 2, 3][2]')
  testSubscriptTupleSliceStart = _MakeExprTest('(1, 2, 3)[2:]')
  testSubscriptTupleSliceStartStop = _MakeExprTest('(1, 2, 3)[10:11]')
  testSubscriptTupleSliceStartStep = _MakeExprTest('(1, 2, 3, 4, 5, 6)[-2::-2]')
  testSubscriptStartStop = _MakeSliceTest('2:3', 'slice(2, 3, None)')
  testSubscriptMultiDim = _MakeSliceTest('1,2,3', '(1, 2, 3)')
  testSubscriptStartStopObjects = _MakeSliceTest(
      'True:False', 'slice(True, False, None)')
  testSubscriptMultiDimSlice = _MakeSliceTest(
      "'foo','bar':'baz':'qux'", "('foo', slice('bar', 'baz', 'qux'))")

  testStrEmpty = _MakeLiteralTest("''")
  testStrAscii = _MakeLiteralTest("'abc'")
  testStrUtf8 = _MakeLiteralTest(r"'\tfoo\n\xcf\x80'")
  testStrQuoted = _MakeLiteralTest('\'"foo"\'', '\'"foo"\'')
  testStrUtf16 = _MakeLiteralTest("u'\\u0432\\u043e\\u043b\\u043d'")

  testTupleEmpty = _MakeLiteralTest('()')
  testTupleNonEmpty = _MakeLiteralTest('(1, 2, 3)')

  testUnaryOpNot = _MakeExprTest('not True')
  testUnaryOpInvert = _MakeExprTest('~4')
  testUnaryOpPos = _MakeExprTest('+4')


def _MakeModuleBlock():
  return block.ModuleBlock(None, '__main__', '<test>', '',
                           imputil.FutureFeatures())


def _ParseExpr(expr):
  return pythonparser.parse(expr).body[0].value


def _ParseAndVisitExpr(expr):
  visitor = stmt.StatementVisitor(_MakeModuleBlock())
  visitor.visit_expr(_ParseExpr(expr))
  return visitor.writer.getvalue()


def _GrumpRun(cmd):
  p = subprocess.Popen(['grumprun'], stdin=subprocess.PIPE,
                       stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
  out, _ = p.communicate(cmd)
  return p.returncode, out


if __name__ == '__main__':
  shard_test.main()

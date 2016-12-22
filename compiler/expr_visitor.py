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

"""Visitor class for traversing Python expressions."""

import ast
import textwrap

from grumpy.compiler import block
from grumpy.compiler import expr
from grumpy.compiler import util


class ExprVisitor(ast.NodeVisitor):
  """Builds and returns a Go expression representing the Python nodes."""

  # pylint: disable=invalid-name,missing-docstring

  def __init__(self, block_, writer):
    self.block = block_
    self.writer = writer

  def generic_visit(self, node):
    msg = 'expression node not yet implemented: ' + type(node).__name__
    raise util.ParseError(node, msg)

  def visit_Attribute(self, node):
    with self.visit(node.value) as obj:
      attr = self.block.alloc_temp()
      self.writer.write_checked_call2(
          attr, 'πg.GetAttr(πF, {}, {}, nil)',
          obj.expr, self.block.intern(node.attr))
    return attr

  def visit_BinOp(self, node):
    result = self.block.alloc_temp()
    with self.visit(node.left) as lhs, self.visit(node.right) as rhs:
      op_type = type(node.op)
      if op_type in ExprVisitor._BIN_OP_TEMPLATES:
        tmpl = ExprVisitor._BIN_OP_TEMPLATES[op_type]
        self.writer.write_checked_call2(
            result, tmpl, lhs=lhs.expr, rhs=rhs.expr)
      else:
        msg = 'binary op not implemented: {}'.format(op_type.__name__)
        raise util.ParseError(node, msg)
    return result

  def visit_BoolOp(self, node):
    result = self.block.alloc_temp()
    with self.block.alloc_temp('bool') as is_true:
      if isinstance(node.op, ast.And):
        cond_expr = '!' + is_true.expr
      else:
        cond_expr = is_true.expr
      end_label = self.block.genlabel()
      num_values = len(node.values)
      for i, n in enumerate(node.values):
        with self.visit(n) as v:
          self.writer.write('{} = {}'.format(result.expr, v.expr))
        if i < num_values - 1:
          self.writer.write_checked_call2(
              is_true, 'πg.IsTrue(πF, {})', result.expr)
          self.writer.write_tmpl(textwrap.dedent("""\
              if $cond_expr {
              \tgoto Label$end_label
              }"""), cond_expr=cond_expr, end_label=end_label)
    self.writer.write_label(end_label)
    return result

  def visit_Call(self, node):
    # Build positional arguments.
    args = expr.nil_expr
    if node.args:
      args = self.block.alloc_temp('[]*πg.Object')
      self.writer.write('{} = πF.MakeArgs({})'.format(args.expr,
                                                      len(node.args)))
      for i, n in enumerate(node.args):
        with self.visit(n) as a:
          self.writer.write('{}[{}] = {}'.format(args.expr, i, a.expr))
    varg = expr.nil_expr
    if node.starargs:
      varg = self.visit(node.starargs)
    # Build keyword arguments
    keywords = expr.nil_expr
    if node.keywords:
      values = []
      for k in node.keywords:
        values.append((util.go_str(k.arg), self.visit(k.value)))
      keywords = self.block.alloc_temp('πg.KWArgs')
      self.writer.write_tmpl('$keywords = πg.KWArgs{', keywords=keywords.name)
      with self.writer.indent_block():
        for k, v in values:
          with v:
            self.writer.write_tmpl('{$name, $value},', name=k, value=v.expr)
      self.writer.write('}')
    kwargs = expr.nil_expr
    if node.kwargs:
      kwargs = self.visit(node.kwargs)
    # Invoke function with all parameters.
    with args, varg, keywords, kwargs, self.visit(node.func) as func:
      result = self.block.alloc_temp()
      if varg is expr.nil_expr and kwargs is expr.nil_expr:
        self.writer.write_checked_call2(result, '{}.Call(πF, {}, {})',
                                        func.expr, args.expr, keywords.expr)
      else:
        self.writer.write_checked_call2(result,
                                        'πg.Invoke(πF, {}, {}, {}, {}, {})',
                                        func.expr, args.expr, varg.expr,
                                        keywords.expr, kwargs.expr)
      if node.args:
        self.writer.write('πF.FreeArgs({})'.format(args.expr))
    return result

  def visit_Compare(self, node):
    result = self.block.alloc_temp()
    lhs = self.visit(node.left)
    n = len(node.ops)
    end_label = self.block.genlabel() if n > 1 else None
    for i, (op, comp) in enumerate(zip(node.ops, node.comparators)):
      rhs = self.visit(comp)
      op_type = type(op)
      if op_type in ExprVisitor._CMP_OP_TEMPLATES:
        tmpl = ExprVisitor._CMP_OP_TEMPLATES[op_type]
        self.writer.write_checked_call2(
            result, tmpl, lhs=lhs.expr, rhs=rhs.expr)
      elif isinstance(op, (ast.In, ast.NotIn)):
        with self.block.alloc_temp('bool') as contains:
          self.writer.write_checked_call2(
              contains, 'πg.Contains(πF, {}, {})', rhs.expr, lhs.expr)
          invert = '' if isinstance(op, ast.In) else '!'
          self.writer.write('{} = πg.GetBool({}{}).ToObject()'.format(
              result.name, invert, contains.expr))
      elif isinstance(op, ast.Is):
        self.writer.write('{} = πg.GetBool({} == {}).ToObject()'.format(
            result.name, lhs.expr, rhs.expr))
      elif isinstance(op, ast.IsNot):
        self.writer.write('{} = πg.GetBool({} != {}).ToObject()'.format(
            result.name, lhs.expr, rhs.expr))
      else:
        raise AssertionError('unrecognized compare op: {}'.format(
            op_type.__name__))
      if i < n - 1:
        with self.block.alloc_temp('bool') as cond:
          self.writer.write_checked_call2(
              cond, 'πg.IsTrue(πF, {})', result.expr)
          self.writer.write_tmpl(textwrap.dedent("""\
              if !$cond {
              \tgoto Label$end_label
              }"""), cond=cond.expr, end_label=end_label)
      lhs.free()
      lhs = rhs
    rhs.free()
    if end_label is not None:
      self.writer.write_label(end_label)
    return result

  def visit_Dict(self, node):
    with self.block.alloc_temp('*πg.Dict') as d:
      self.writer.write('{} = πg.NewDict()'.format(d.name))
      for k, v in zip(node.keys, node.values):
        with self.visit(k) as key, self.visit(v) as value:
          self.writer.write_checked_call1('{}.SetItem(πF, {}, {})',
                                          d.expr, key.expr, value.expr)
      result = self.block.alloc_temp()
      self.writer.write('{} = {}.ToObject()'.format(result.name, d.expr))
    return result

  def visit_DictComp(self, node):
    result = self.block.alloc_temp()
    elt = ast.Tuple(elts=[node.key, node.value], context=ast.Load)
    with self.visit(ast.GeneratorExp(elt, node.generators)) as gen:
      self.writer.write_checked_call2(
          result, 'πg.DictType.Call(πF, πg.Args{{{}}}, nil)', gen.expr)
    return result

  def visit_ExtSlice(self, node):
    result = self.block.alloc_temp()
    with self.block.alloc_temp('[]*πg.Object') as dims:
      self.writer.write('{} = make([]*πg.Object, {})'.format(
          dims.name, len(node.dims)))
      for i, dim in enumerate(node.dims):
        with self.visit(dim) as s:
          self.writer.write('{}[{}] = {}'.format(dims.name, i, s.expr))
      self.writer.write('{} = πg.NewTuple({}...).ToObject()'.format(
          result.name, dims.expr))
    return result

  def visit_GeneratorExp(self, node):
    body = ast.Expr(value=ast.Yield(node.elt), lineno=None)
    for comp_node in reversed(node.generators):
      for if_node in reversed(comp_node.ifs):
        body = ast.If(test=if_node, body=[body], orelse=[], lineno=None)
      body = ast.For(target=comp_node.target, iter=comp_node.iter,
                     body=[body], orelse=[], lineno=None)

    args = ast.arguments(args=[], vararg=None, kwarg=None, defaults=[])
    node = ast.FunctionDef(name='<generator>', args=args, body=[body])
    gen_func = self.visit_function_inline(node)
    result = self.block.alloc_temp()
    self.writer.write_checked_call2(
        result, '{}.Call(πF, nil, nil)', gen_func.expr)
    return result

  def visit_IfExp(self, node):
    else_label, end_label = self.block.genlabel(), self.block.genlabel()
    result = self.block.alloc_temp()
    with self.visit(node.test) as test, self.block.alloc_temp('bool') as cond:
      self.writer.write_checked_call2(
          cond, 'πg.IsTrue(πF, {})', test.expr)
      self.writer.write_tmpl(textwrap.dedent("""\
          if !$cond {
          \tgoto Label$else_label
          }"""), cond=cond.expr, else_label=else_label)
    with self.visit(node.body) as value:
      self.writer.write('{} = {}'.format(result.name, value.expr))
      self.writer.write('goto Label{}'.format(end_label))
    self.writer.write_label(else_label)
    with self.visit(node.orelse) as value:
      self.writer.write('{} = {}'.format(result.name, value.expr))
    self.writer.write_label(end_label)
    return result

  def visit_Index(self, node):
    result = self.block.alloc_temp()
    with self.visit(node.value) as v:
      self.writer.write('{} = {}'.format(result.name, v.expr))
    return result

  def visit_List(self, node):
    with self._visit_seq_elts(node.elts) as elems:
      result = self.block.alloc_temp()
      self.writer.write('{} = πg.NewList({}...).ToObject()'.format(
          result.expr, elems.expr))
    return result

  def visit_ListComp(self, node):
    result = self.block.alloc_temp()
    with self.visit(ast.GeneratorExp(node.elt, node.generators)) as gen:
      self.writer.write_checked_call2(
          result, 'πg.ListType.Call(πF, πg.Args{{{}}}, nil)', gen.expr)
    return result

  def visit_Name(self, node):
    return self.block.resolve_name(self.writer, node.id)

  def visit_Num(self, node):
    if isinstance(node.n, int):
      expr_str = 'NewInt({})'.format(node.n)
    elif isinstance(node.n, long):
      a = abs(node.n)
      gobytes = ''
      while a:
        gobytes = hex(int(a&255)) + ',' + gobytes
        a >>= 8
      expr_str = 'NewLongFromBytes([]byte{{{}}})'.format(gobytes)
      if node.n < 0:
        expr_str = expr_str + '.Neg()'
    elif isinstance(node.n, float):
      expr_str = 'NewFloat({})'.format(node.n)
    else:
      msg = 'number type not yet implemented: ' + type(node.n).__name__
      raise util.ParseError(node, msg)
    return expr.GeneratedLiteral('πg.' + expr_str + '.ToObject()')

  def visit_Slice(self, node):
    result = self.block.alloc_temp()
    lower = upper = step = expr.GeneratedLiteral('πg.None')
    if node.lower:
      lower = self.visit(node.lower)
    if node.upper:
      upper = self.visit(node.upper)
    if node.step:
      step = self.visit(node.step)
    with lower, upper, step:
      self.writer.write_checked_call2(
          result, 'πg.SliceType.Call(πF, πg.Args{{{}, {}, {}}}, nil)',
          lower.expr, upper.expr, step.expr)
    return result

  def visit_Subscript(self, node):
    rhs = self.visit(node.slice)
    result = self.block.alloc_temp()
    with rhs, self.visit(node.value) as lhs:
      self.writer.write_checked_call2(result, 'πg.GetItem(πF, {}, {})',
                                      lhs.expr, rhs.expr)
    return result

  def visit_Str(self, node):
    if isinstance(node.s, unicode):
      expr_str = 'πg.NewUnicode({}).ToObject()'.format(
          util.go_str(node.s.encode('utf-8')))
    else:
      expr_str = '{}.ToObject()'.format(self.block.intern(node.s))
    return expr.GeneratedLiteral(expr_str)

  def visit_Tuple(self, node):
    with self._visit_seq_elts(node.elts) as elems:
      result = self.block.alloc_temp()
      self.writer.write('{} = πg.NewTuple({}...).ToObject()'.format(
          result.expr, elems.expr))
    return result

  def visit_UnaryOp(self, node):
    result = self.block.alloc_temp()
    with self.visit(node.operand) as operand:
      op_type = type(node.op)
      if op_type in ExprVisitor._UNARY_OP_TEMPLATES:
        self.writer.write_checked_call2(
            result, ExprVisitor._UNARY_OP_TEMPLATES[op_type],
            operand=operand.expr)
      elif isinstance(node.op, ast.Not):
        with self.block.alloc_temp('bool') as is_true:
          self.writer.write_checked_call2(
              is_true, 'πg.IsTrue(πF, {})', operand.expr)
          self.writer.write('{} = πg.GetBool(!{}).ToObject()'.format(
              result.name, is_true.expr))
      else:
        msg = 'unary op not implemented: {}'.format(op_type.__name__)
        raise util.ParseError(node, msg)
    return result

  def visit_Yield(self, node):
    if node.value:
      value = self.visit(node.value)
    else:
      value = expr.GeneratedLiteral('πg.None')
    resume_label = self.block.genlabel(is_checkpoint=True)
    self.writer.write('πF.PushCheckpoint({})'.format(resume_label))
    self.writer.write('return {}, nil'.format(value.expr))
    self.writer.write_label(resume_label)
    result = self.block.alloc_temp()
    self.writer.write('{} = πSent'.format(result.name))
    return result

  _BIN_OP_TEMPLATES = {
      ast.BitAnd: 'πg.And(πF, {lhs}, {rhs})',
      ast.BitOr: 'πg.Or(πF, {lhs}, {rhs})',
      ast.BitXor: 'πg.Xor(πF, {lhs}, {rhs})',
      ast.Add: 'πg.Add(πF, {lhs}, {rhs})',
      ast.Div: 'πg.Div(πF, {lhs}, {rhs})',
      # TODO: Support "from __future__ import division".
      ast.FloorDiv: 'πg.Div(πF, {lhs}, {rhs})',
      ast.LShift: 'πg.LShift(πF, {lhs}, {rhs})',
      ast.Mod: 'πg.Mod(πF, {lhs}, {rhs})',
      ast.Mult: 'πg.Mul(πF, {lhs}, {rhs})',
      ast.RShift: 'πg.RShift(πF, {lhs}, {rhs})',
      ast.Sub: 'πg.Sub(πF, {lhs}, {rhs})',
  }

  _CMP_OP_TEMPLATES = {
      ast.Eq: 'πg.Eq(πF, {lhs}, {rhs})',
      ast.Gt: 'πg.GT(πF, {lhs}, {rhs})',
      ast.GtE: 'πg.GE(πF, {lhs}, {rhs})',
      ast.Lt: 'πg.LT(πF, {lhs}, {rhs})',
      ast.LtE: 'πg.LE(πF, {lhs}, {rhs})',
      ast.NotEq: 'πg.NE(πF, {lhs}, {rhs})',
  }

  _UNARY_OP_TEMPLATES = {
      ast.Invert: 'πg.Invert(πF, {operand})',
  }

  def visit_function_inline(self, node):
    """Returns an GeneratedExpr for a function with the given body."""
    # First pass collects the names of locals used in this function. Do this in
    # a separate pass so that we know whether to resolve a name as a local or a
    # global during the second pass.
    func_visitor = block.FunctionBlockVisitor(node)
    for child in node.body:
      func_visitor.visit(child)
    func_block = block.FunctionBlock(self.block, node.name, func_visitor.vars,
                                     func_visitor.is_generator)
    # TODO: Find a better way to reduce coupling between ExprVisitor and
    # StatementVisitor.
    from grumpy.compiler import stmt  # pylint: disable=g-import-not-at-top
    visitor = stmt.StatementVisitor(func_block)
    # Indent so that the function body is aligned with the goto labels.
    with visitor.writer.indent_block():
      visitor._visit_each(node.body)  # pylint: disable=protected-access

    result = self.block.alloc_temp()
    with self.block.alloc_temp('[]πg.FunctionArg') as func_args:
      args = node.args
      argc = len(args.args)
      self.writer.write('{} = make([]πg.FunctionArg, {})'.format(
          func_args.expr, argc))
      # The list of defaults only contains args for which a default value is
      # specified so pad it with None to make it the same length as args.
      defaults = [None] * (argc - len(args.defaults)) + args.defaults
      for i, (a, d) in enumerate(zip(args.args, defaults)):
        with self.visit(d) if d else expr.nil_expr as default:
          tmpl = '$args[$i] = πg.FunctionArg{Name: $name, Def: $default}'
          self.writer.write_tmpl(tmpl, args=func_args.expr, i=i,
                                 name=util.go_str(a.id), default=default.expr)

      vararg = args.vararg if args.vararg else ''
      kwarg = args.kwarg if args.kwarg else ''

      # The function object gets written to a temporary writer because we need
      # it as an expression that we subsequently bind to some variable.
      self.writer.write_tmpl(
          '$result = πg.NewFunction('
          '$name, $args, $vararg, $kwarg, func('
          'πF *πg.Frame, πArgs []*πg.Object) (*πg.Object, *πg.BaseException) {',
          result=result.name, name=util.go_str(node.name), args=func_args.expr,
          vararg=util.go_str(vararg), kwarg=util.go_str(kwarg))
      with self.writer.indent_block():
        # Declare the local variables used in this function.
        for var in func_block.vars.values():
          if var.type != block.Var.TYPE_GLOBAL:
            fmt = 'var {0} *πg.Object = {1}; _ = {0}'
            self.writer.write(fmt.format(
                util.adjust_local_name(var.name), var.init_expr))
        body = visitor.writer.out.getvalue()
        if func_block.checkpoints:
          self.writer.write_block(func_block, body)
          if func_block.is_generator:
            self.writer.write('return πg.NewGenerator('
                              'πBlock, πGlobals).ToObject(), nil')
          else:
            self.writer.write('return πBlock.Exec(πF, πGlobals)')
        else:
          assert not func_block.is_generator
          self.writer.write('var πE *πg.BaseException\n_ = πE')
          self.writer.write_temp_decls(func_block)
          # There's no goto labels so align with the rest of the function.
          with self.writer.indent_block(-1):
            self.writer.write(body)
          self.writer.write('return πg.None, nil')
      self.writer.write('}).ToObject()')
    return result

  def _visit_seq_elts(self, elts):
    result = self.block.alloc_temp('[]*πg.Object')
    self.writer.write('{} = make([]*πg.Object, {})'.format(
        result.expr, len(elts)))
    for i, e in enumerate(elts):
      with self.visit(e) as elt:
        self.writer.write('{}[{}] = {}'.format(result.expr, i, elt.expr))
    return result

  def _node_not_implemented(self, node):
    msg = 'node not yet implemented: ' + type(node).__name__
    raise util.ParseError(node, msg)

  visit_Lambda = _node_not_implemented
  visit_SetComp = _node_not_implemented

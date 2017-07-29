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

"""Visitor class for traversing Python statements."""

from __future__ import unicode_literals

import string
import textwrap

from grumpy.compiler import block
from grumpy.compiler import expr
from grumpy.compiler import expr_visitor
from grumpy.compiler import imputil
from grumpy.compiler import util
from grumpy.pythonparser import algorithm
from grumpy.pythonparser import ast


_NATIVE_TYPE_PREFIX = 'type_'

# Partial list of known vcs for go module import
# Full list can be found at https://golang.org/src/cmd/go/vcs.go
# TODO: Use official vcs.go module instead of partial list
_KNOWN_VCS = [
    'golang.org', 'github.com', 'bitbucket.org', 'git.apache.org',
    'git.openstack.org', 'launchpad.net'
]

_nil_expr = expr.nil_expr


class StatementVisitor(algorithm.Visitor):
  """Outputs Go statements to a Writer for the given Python nodes."""

  # pylint: disable=invalid-name,missing-docstring

  def __init__(self, block_, future_node=None):
    self.block = block_
    self.future_node = future_node
    self.writer = util.Writer()
    self.expr_visitor = expr_visitor.ExprVisitor(self)

  def generic_visit(self, node):
    msg = 'node not yet implemented: {}'.format(type(node).__name__)
    raise util.ParseError(node, msg)

  def visit_expr(self, node):
    return self.expr_visitor.visit(node)

  def visit_Assert(self, node):
    self._write_py_context(node.lineno)
    # TODO: Only evaluate msg if cond is false.
    with self.visit_expr(node.msg) if node.msg else _nil_expr as msg,\
        self.visit_expr(node.test) as cond:
      self.writer.write_checked_call1(
          'πg.Assert(πF, {}, {})', cond.expr, msg.expr)

  def visit_AugAssign(self, node):
    op_type = type(node.op)
    if op_type not in StatementVisitor._AUG_ASSIGN_TEMPLATES:
      fmt = 'augmented assignment op not implemented: {}'
      raise util.ParseError(node, fmt.format(op_type.__name__))
    self._write_py_context(node.lineno)
    with self.visit_expr(node.target) as target,\
        self.visit_expr(node.value) as value,\
        self.block.alloc_temp() as temp:
      self.writer.write_checked_call2(
          temp, StatementVisitor._AUG_ASSIGN_TEMPLATES[op_type],
          lhs=target.expr, rhs=value.expr)
      self._assign_target(node.target, temp.expr)

  def visit_Assign(self, node):
    self._write_py_context(node.lineno)
    with self.visit_expr(node.value) as value:
      for target in node.targets:
        self._tie_target(target, value.expr)

  def visit_Break(self, node):
    if not self.block.loop_stack:
      raise util.ParseError(node, "'break' not in loop")
    self._write_py_context(node.lineno)
    self.writer.write_tmpl(textwrap.dedent("""\
        $breakvar = true
        continue"""), breakvar=self.block.top_loop().breakvar.name)

  def visit_ClassDef(self, node):
    # Since we only care about global vars, we end up throwing away the locals
    # collected by BlockVisitor. But use it anyway since it buys us detection of
    # assignment to vars that are later declared global.
    block_visitor = block.BlockVisitor()
    for child in node.body:
      block_visitor.visit(child)
    global_vars = {v.name for v in block_visitor.vars.values()
                   if v.type == block.Var.TYPE_GLOBAL}
    # Visit all the statements inside body of the class definition.
    body_visitor = StatementVisitor(block.ClassBlock(
        self.block, node.name, global_vars), self.future_node)
    # Indent so that the function body is aligned with the goto labels.
    with body_visitor.writer.indent_block():
      body_visitor._visit_each(node.body)  # pylint: disable=protected-access

    self._write_py_context(node.lineno)
    with self.block.alloc_temp('*πg.Dict') as cls, \
        self.block.alloc_temp() as mod_name, \
        self.block.alloc_temp('[]*πg.Object') as bases, \
        self.block.alloc_temp() as meta:
      self.writer.write('{} = make([]*πg.Object, {})'.format(
          bases.expr, len(node.bases)))
      for i, b in enumerate(node.bases):
        with self.visit_expr(b) as b:
          self.writer.write('{}[{}] = {}'.format(bases.expr, i, b.expr))
      self.writer.write('{} = πg.NewDict()'.format(cls.name))
      self.writer.write_checked_call2(
          mod_name, 'πF.Globals().GetItem(πF, {}.ToObject())',
          self.block.root.intern('__name__'))
      self.writer.write_checked_call1(
          '{}.SetItem(πF, {}.ToObject(), {})',
          cls.expr, self.block.root.intern('__module__'), mod_name.expr)
      tmpl = textwrap.dedent("""
          _, πE = πg.NewCode($name, $filename, nil, 0, func(πF *πg.Frame, _ []*πg.Object) (*πg.Object, *πg.BaseException) {
          \tπClass := $cls
          \t_ = πClass""")
      self.writer.write_tmpl(tmpl, name=util.go_str(node.name),
                             filename=util.go_str(self.block.root.filename),
                             cls=cls.expr)
      with self.writer.indent_block():
        self.writer.write_temp_decls(body_visitor.block)
        self.writer.write_block(body_visitor.block,
                                body_visitor.writer.getvalue())
        self.writer.write('return nil, nil')
      tmpl = textwrap.dedent("""\
          }).Eval(πF, πF.Globals(), nil, nil)
          if πE != nil {
          \tcontinue
          }
          if $meta, πE = $cls.GetItem(πF, $metaclass_str.ToObject()); πE != nil {
          \tcontinue
          }
          if $meta == nil {
          \t$meta = πg.TypeType.ToObject()
          }""")
      self.writer.write_tmpl(
          tmpl, meta=meta.name, cls=cls.expr,
          metaclass_str=self.block.root.intern('__metaclass__'))
      with self.block.alloc_temp() as type_:
        type_expr = ('{}.Call(πF, []*πg.Object{{πg.NewStr({}).ToObject(), '
                     'πg.NewTuple({}...).ToObject(), {}.ToObject()}}, nil)')
        self.writer.write_checked_call2(
            type_, type_expr, meta.expr,
            util.go_str(node.name), bases.expr, cls.expr)
        self.block.bind_var(self.writer, node.name, type_.expr)

  def visit_Continue(self, node):
    if not self.block.loop_stack:
      raise util.ParseError(node, "'continue' not in loop")
    self._write_py_context(node.lineno)
    self.writer.write('continue')

  def visit_Delete(self, node):
    self._write_py_context(node.lineno)
    for target in node.targets:
      if isinstance(target, ast.Attribute):
        with self.visit_expr(target.value) as t:
          self.writer.write_checked_call1(
              'πg.DelAttr(πF, {}, {})', t.expr,
              self.block.root.intern(target.attr))
      elif isinstance(target, ast.Name):
        self.block.del_var(self.writer, target.id)
      elif isinstance(target, ast.Subscript):
        with self.visit_expr(target.value) as t,\
            self.visit_expr(target.slice) as index:
          self.writer.write_checked_call1('πg.DelItem(πF, {}, {})',
                                          t.expr, index.expr)
      else:
        msg = 'del target not implemented: {}'.format(type(target).__name__)
        raise util.ParseError(node, msg)

  def visit_Expr(self, node):
    self._write_py_context(node.lineno)
    self.visit_expr(node.value).free()

  def visit_For(self, node):
    with self.block.alloc_temp() as i:
      with self.visit_expr(node.iter) as iter_expr:
        self.writer.write_checked_call2(i, 'πg.Iter(πF, {})', iter_expr.expr)
      def testfunc(testvar):
        with self.block.alloc_temp() as n:
          self.writer.write_tmpl(textwrap.dedent("""\
              if $n, πE = πg.Next(πF, $i); πE != nil {
              \tisStop, exc := πg.IsInstance(πF, πE.ToObject(), πg.StopIterationType.ToObject())
              \tif exc != nil {
              \t\tπE = exc
              \t} else if isStop {
              \t\tπE = nil
              \t\tπF.RestoreExc(nil, nil)
              \t}
              \t$testvar = !isStop
              } else {
              \t$testvar = true"""), n=n.name, i=i.expr, testvar=testvar.name)
          with self.writer.indent_block():
            self._tie_target(node.target, n.expr)
          self.writer.write('}')
      self._visit_loop(testfunc, node)

  def visit_FunctionDef(self, node):
    self._write_py_context(node.lineno + len(node.decorator_list))
    func = self.visit_function_inline(node)
    self.block.bind_var(self.writer, node.name, func.expr)
    while node.decorator_list:
      decorator = node.decorator_list.pop()
      wrapped = ast.Name(id=node.name)
      decorated = ast.Call(func=decorator, args=[wrapped], keywords=[],
                           starargs=None, kwargs=None)
      target = ast.Assign(targets=[wrapped], value=decorated, loc=node.loc)
      self.visit_Assign(target)

  def visit_Global(self, node):
    self._write_py_context(node.lineno)

  def visit_If(self, node):
    # Collect the nodes for each if/elif/else body and write the dispatching
    # switch statement.
    bodies = []
    # An elif clause is represented as a single If node within the orelse
    # section of the previous If node. Thus this loop terminates once we are
    # done all the elif clauses at which time the orelse var will contain the
    # nodes (if any) for the else clause.
    orelse = [node]
    while len(orelse) == 1 and isinstance(orelse[0], ast.If):
      ifnode = orelse[0]
      with self.visit_expr(ifnode.test) as cond:
        label = self.block.genlabel()
        # We goto the body of the if statement instead of executing it inline
        # because the body itself may be a goto target and Go does not support
        # jumping to targets inside a block.
        with self.block.alloc_temp('bool') as is_true:
          self.writer.write_tmpl(textwrap.dedent("""\
              if $is_true, πE = πg.IsTrue(πF, $cond); πE != nil {
              \tcontinue
              }
              if $is_true {
              \tgoto Label$label
              }"""), is_true=is_true.name, cond=cond.expr, label=label)
      bodies.append((label, ifnode.body, ifnode.lineno))
      orelse = ifnode.orelse
    default_label = end_label = self.block.genlabel()
    if orelse:
      end_label = self.block.genlabel()
      # The else is not represented by ast and thus there is no lineno.
      bodies.append((default_label, orelse, None))
    self.writer.write('goto Label{}'.format(default_label))
    # Write the body of each clause.
    for label, body, lineno in bodies:
      if lineno:
        self._write_py_context(lineno)
      self.writer.write_label(label)
      self._visit_each(body)
      self.writer.write('goto Label{}'.format(end_label))
    self.writer.write_label(end_label)

  def visit_Import(self, node):
    self._write_py_context(node.lineno)
    for imp in self.block.root.importer.visit(node):
      self._import_and_bind(imp)

  def visit_ImportFrom(self, node):
    self._write_py_context(node.lineno)

    if node.module == '__future__' and node != self.future_node:
      raise util.LateFutureError(node)

    for imp in self.block.root.importer.visit(node):
      self._import_and_bind(imp)

  def visit_Module(self, node):
    self._visit_each(node.body)

  def visit_Pass(self, node):
    self._write_py_context(node.lineno)

  def visit_Print(self, node):
    if self.block.root.future_features.print_function:
      raise util.ParseError(node, 'syntax error (print is not a keyword)')
    self._write_py_context(node.lineno)
    with self.block.alloc_temp('[]*πg.Object') as args:
      self.writer.write('{} = make([]*πg.Object, {})'.format(
          args.expr, len(node.values)))
      for i, v in enumerate(node.values):
        with self.visit_expr(v) as arg:
          self.writer.write('{}[{}] = {}'.format(args.expr, i, arg.expr))
      self.writer.write_checked_call1('πg.Print(πF, {}, {})', args.expr,
                                      'true' if node.nl else 'false')

  def visit_Raise(self, node):
    with self.visit_expr(node.exc) if node.exc else _nil_expr as t,\
        self.visit_expr(node.inst) if node.inst else _nil_expr as inst,\
        self.visit_expr(node.tback) if node.tback else _nil_expr as tb:
      if node.inst:
        assert node.exc, 'raise had inst but no type'
      if node.tback:
        assert node.inst, 'raise had tback but no inst'
      self._write_py_context(node.lineno)
      self.writer.write('πE = πF.Raise({}, {}, {})'.format(
          t.expr, inst.expr, tb.expr))
      self.writer.write('continue')

  def visit_Return(self, node):
    assert isinstance(self.block, block.FunctionBlock)
    self._write_py_context(node.lineno)
    if self.block.is_generator and node.value:
      raise util.ParseError(node, 'returning a value in a generator function')
    if node.value:
      with self.visit_expr(node.value) as value:
        self.writer.write('πR = {}'.format(value.expr))
    else:
      self.writer.write('πR = πg.None')
    self.writer.write('continue')

  def visit_Try(self, node):
    # The general structure generated by this method is shown below:
    #
    #       checkpoints.Push(Except)
    #       <try body>
    #       Checkpoints.Pop()
    #       <else body>
    #       goto Finally
    #     Except:
    #       <dispatch table>
    #     Handler1:
    #       <handler 1 body>
    #       Checkpoints.Pop()  // Finally
    #       goto Finally
    #     Handler2:
    #       <handler 2 body>
    #       Checkpoints.Pop()  // Finally
    #       goto Finally
    #     ...
    #     Finally:
    #       <finally body>
    #
    # The dispatch table maps the current exception to the appropriate handler
    # label according to the exception clauses.

    # Write the try body.
    self._write_py_context(node.lineno)
    finally_label = self.block.genlabel(is_checkpoint=bool(node.finalbody))
    if node.finalbody:
      self.writer.write('πF.PushCheckpoint({})'.format(finally_label))
    except_label = None
    if node.handlers:
      except_label = self.block.genlabel(is_checkpoint=True)
      self.writer.write('πF.PushCheckpoint({})'.format(except_label))
    self._visit_each(node.body)
    if except_label:
      self.writer.write('πF.PopCheckpoint()')  # except_label
    if node.orelse:
      self._visit_each(node.orelse)
    if node.finalbody:
      self.writer.write('πF.PopCheckpoint()')  # finally_label
    self.writer.write('goto Label{}'.format(finally_label))

    with self.block.alloc_temp('*πg.BaseException') as exc:
      if except_label:
        with self.block.alloc_temp('*πg.Traceback') as tb:
          self.writer.write_label(except_label)
          self.writer.write_tmpl(textwrap.dedent("""\
              if πE == nil {
                continue
              }
              πE = nil
              $exc, $tb = πF.ExcInfo()"""), exc=exc.expr, tb=tb.expr)
          handler_labels = self._write_except_dispatcher(
              exc.expr, tb.expr, node.handlers)

        # Write the bodies of each of the except handlers.
        for handler_label, except_node in zip(handler_labels, node.handlers):
          self._write_except_block(handler_label, exc.expr, except_node)
          if node.finalbody:
            self.writer.write('πF.PopCheckpoint()')  # finally_label
          self.writer.write('goto Label{}'.format(finally_label))

      # Write the finally body.
      self.writer.write_label(finally_label)
      if node.finalbody:
        with self.block.alloc_temp('*πg.Traceback') as tb:
          self.writer.write('{}, {} = πF.RestoreExc(nil, nil)'.format(
              exc.expr, tb.expr))
          self._visit_each(node.finalbody)
          self.writer.write_tmpl(textwrap.dedent("""\
              if $exc != nil {
              \tπE = πF.Raise($exc.ToObject(), nil, $tb.ToObject())
              \tcontinue
              }
              if πR != nil {
              \tcontinue
              }"""), exc=exc.expr, tb=tb.expr)

  def visit_While(self, node):
    self._write_py_context(node.lineno)
    def testfunc(testvar):
      with self.visit_expr(node.test) as cond:
        self.writer.write_checked_call2(
            testvar, 'πg.IsTrue(πF, {})', cond.expr)
    self._visit_loop(testfunc, node)

  def visit_With(self, node):
    assert len(node.items) == 1, 'multiple items in a with not yet supported'
    item = node.items[0]
    self._write_py_context(node.loc.line())
    # mgr := EXPR
    with self.visit_expr(item.context_expr) as mgr,\
        self.block.alloc_temp() as exit_func,\
        self.block.alloc_temp() as value:
      # The code here has a subtle twist: It gets the exit function attribute
      # from the class, not from the object. This matches the pseudo code from
      # PEP 343 exactly, and is very close to what CPython actually does.  (The
      # CPython implementation actually uses a special lookup which is performed
      # on the object, but skips the instance dictionary: see ceval.c and
      # lookup_maybe in typeobject.c.)

      # exit := type(mgr).__exit__
      self.writer.write_checked_call2(
          exit_func, 'πg.GetAttr(πF, {}.Type().ToObject(), {}, nil)',
          mgr.expr, self.block.root.intern('__exit__'))
      # value := type(mgr).__enter__(mgr)
      self.writer.write_checked_call2(
          value, 'πg.GetAttr(πF, {}.Type().ToObject(), {}, nil)',
          mgr.expr, self.block.root.intern('__enter__'))
      self.writer.write_checked_call2(
          value, '{}.Call(πF, πg.Args{{{}}}, nil)',
          value.expr, mgr.expr)

      finally_label = self.block.genlabel(is_checkpoint=True)
      self.writer.write('πF.PushCheckpoint({})'.format(finally_label))
      if item.optional_vars:
        self._tie_target(item.optional_vars, value.expr)
      self._visit_each(node.body)
      self.writer.write('πF.PopCheckpoint()')
      self.writer.write_label(finally_label)

      with self.block.alloc_temp() as swallow_exc,\
          self.block.alloc_temp('bool') as swallow_exc_bool,\
          self.block.alloc_temp('*πg.BaseException') as exc,\
          self.block.alloc_temp('*πg.Traceback') as tb,\
          self.block.alloc_temp('*πg.Type') as t:
        # temp := exit(mgr, *sys.exec_info())
        tmpl = """\
            $exc, $tb = nil, nil
            if πE != nil {
            \t$exc, $tb = πF.ExcInfo()
            }
            if $exc != nil {
            \t$t = $exc.Type()
            \tif $swallow_exc, πE = $exit_func.Call(πF, πg.Args{$mgr, $t.ToObject(), $exc.ToObject(), $tb.ToObject()}, nil); πE != nil {
            \t\tcontinue
            \t}
            } else {
            \tif $swallow_exc, πE = $exit_func.Call(πF, πg.Args{$mgr, πg.None, πg.None, πg.None}, nil); πE != nil {
            \t\tcontinue
            \t}
            }
        """
        self.writer.write_tmpl(
            textwrap.dedent(tmpl), exc=exc.expr, tb=tb.expr, t=t.name,
            mgr=mgr.expr, exit_func=exit_func.expr,
            swallow_exc=swallow_exc.name)

        # if Exc != nil && swallow_exc != true {
        #   Raise(nil, nil)
        # }
        self.writer.write_checked_call2(
            swallow_exc_bool, 'πg.IsTrue(πF, {})', swallow_exc.expr)
        self.writer.write_tmpl(textwrap.dedent("""\
            if $exc != nil && $swallow_exc != true {
            \tπE = πF.Raise(nil, nil, nil)
            \tcontinue
            }
            if πR != nil {
            \tcontinue
            }"""), exc=exc.expr, swallow_exc=swallow_exc_bool.expr)

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
    visitor = StatementVisitor(func_block, self.future_node)
    # Indent so that the function body is aligned with the goto labels.
    with visitor.writer.indent_block():
      visitor._visit_each(node.body)  # pylint: disable=protected-access

    result = self.block.alloc_temp()
    with self.block.alloc_temp('[]πg.Param') as func_args:
      args = node.args
      argc = len(args.args)
      self.writer.write('{} = make([]πg.Param, {})'.format(
          func_args.expr, argc))
      # The list of defaults only contains args for which a default value is
      # specified so pad it with None to make it the same length as args.
      defaults = [None] * (argc - len(args.defaults)) + args.defaults
      for i, (a, d) in enumerate(zip(args.args, defaults)):
        with self.visit_expr(d) if d else expr.nil_expr as default:
          tmpl = '$args[$i] = πg.Param{Name: $name, Def: $default}'
          self.writer.write_tmpl(tmpl, args=func_args.expr, i=i,
                                 name=util.go_str(a.arg), default=default.expr)
      flags = []
      if args.vararg:
        flags.append('πg.CodeFlagVarArg')
      if args.kwarg:
        flags.append('πg.CodeFlagKWArg')
      # The function object gets written to a temporary writer because we need
      # it as an expression that we subsequently bind to some variable.
      self.writer.write_tmpl(
          '$result = πg.NewFunction(πg.NewCode($name, $filename, $args, '
          '$flags, func(πF *πg.Frame, πArgs []*πg.Object) '
          '(*πg.Object, *πg.BaseException) {',
          result=result.name, name=util.go_str(node.name),
          filename=util.go_str(self.block.root.filename), args=func_args.expr,
          flags=' | '.join(flags) if flags else 0)
      with self.writer.indent_block():
        for var in func_block.vars.values():
          if var.type != block.Var.TYPE_GLOBAL:
            fmt = 'var {0} *πg.Object = {1}; _ = {0}'
            self.writer.write(fmt.format(
                util.adjust_local_name(var.name), var.init_expr))
        self.writer.write_temp_decls(func_block)
        self.writer.write('var πR *πg.Object; _ = πR')
        self.writer.write('var πE *πg.BaseException; _ = πE')
        if func_block.is_generator:
          self.writer.write(
              'return πg.NewGenerator(πF, func(πSent *πg.Object) '
              '(*πg.Object, *πg.BaseException) {')
          with self.writer.indent_block():
            self.writer.write_block(func_block, visitor.writer.getvalue())
            self.writer.write('return nil, πE')
          self.writer.write('}).ToObject(), nil')
        else:
          self.writer.write_block(func_block, visitor.writer.getvalue())
          self.writer.write(textwrap.dedent("""\
              if πE != nil {
              \tπR = nil
              } else if πR == nil {
              \tπR = πg.None
              }
              return πR, πE"""))
      self.writer.write('}), πF.Globals()).ToObject()')
    return result

  _AUG_ASSIGN_TEMPLATES = {
      ast.Add: 'πg.IAdd(πF, {lhs}, {rhs})',
      ast.BitAnd: 'πg.IAnd(πF, {lhs}, {rhs})',
      ast.Div: 'πg.IDiv(πF, {lhs}, {rhs})',
      ast.FloorDiv: 'πg.IFloorDiv(πF, {lhs}, {rhs})',
      ast.LShift: 'πg.ILShift(πF, {lhs}, {rhs})',
      ast.Mod: 'πg.IMod(πF, {lhs}, {rhs})',
      ast.Mult: 'πg.IMul(πF, {lhs}, {rhs})',
      ast.BitOr: 'πg.IOr(πF, {lhs}, {rhs})',
      ast.Pow: 'πg.IPow(πF, {lhs}, {rhs})',
      ast.RShift: 'πg.IRShift(πF, {lhs}, {rhs})',
      ast.Sub: 'πg.ISub(πF, {lhs}, {rhs})',
      ast.BitXor: 'πg.IXor(πF, {lhs}, {rhs})',
  }

  def _assign_target(self, target, value):
    if isinstance(target, ast.Name):
      self.block.bind_var(self.writer, target.id, value)
    elif isinstance(target, ast.Attribute):
      with self.visit_expr(target.value) as obj:
        self.writer.write_checked_call1(
            'πg.SetAttr(πF, {}, {}, {})', obj.expr,
            self.block.root.intern(target.attr), value)
    elif isinstance(target, ast.Subscript):
      with self.visit_expr(target.value) as mapping,\
          self.visit_expr(target.slice) as index:
        self.writer.write_checked_call1('πg.SetItem(πF, {}, {}, {})',
                                        mapping.expr, index.expr, value)
    else:
      msg = 'assignment target not yet implemented: ' + type(target).__name__
      raise util.ParseError(target, msg)

  def _build_assign_target(self, target, assigns):
    if isinstance(target, (ast.Tuple, ast.List)):
      children = []
      for elt in target.elts:
        children.append(self._build_assign_target(elt, assigns))
      tmpl = 'πg.TieTarget{Children: []πg.TieTarget{$children}}'
      return string.Template(tmpl).substitute(children=', '.join(children))
    temp = self.block.alloc_temp()
    assigns.append((target, temp))
    tmpl = 'πg.TieTarget{Target: &$temp}'
    return string.Template(tmpl).substitute(temp=temp.name)

  def _import_and_bind(self, imp):
    """Generates code that imports a module and binds it to a variable.

    Args:
      imp: Import object representing an import of the form "import x.y.z" or
          "from x.y import z". Expects only a single binding.
    """
    # Acquire handles to the Code objects in each Go package and call
    # ImportModule to initialize all modules.
    with self.block.alloc_temp() as mod, \
        self.block.alloc_temp('[]*πg.Object') as mod_slice:
      self.writer.write_checked_call2(
          mod_slice, 'πg.ImportModule(πF, {})', util.go_str(imp.name))

      # Bind the imported modules or members to variables in the current scope.
      for binding in imp.bindings:
        if binding.bind_type == imputil.Import.MODULE:
          self.writer.write('{} = {}[{}]'.format(
              mod.name, mod_slice.expr, binding.value))
          self.block.bind_var(self.writer, binding.alias, mod.expr)
        else:
          self.writer.write('{} = {}[{}]'.format(
              mod.name, mod_slice.expr, imp.name.count('.')))
          # Binding a member of the imported module.
          with self.block.alloc_temp() as member:
            self.writer.write_checked_call2(
                member, 'πg.GetAttr(πF, {}, {}, nil)',
                mod.expr, self.block.root.intern(binding.value))
            self.block.bind_var(self.writer, binding.alias, member.expr)

  def _tie_target(self, target, value):
    if isinstance(target, ast.Name):
      self._assign_target(target, value)
      return

    assigns = []
    self.writer.write_checked_call1(
        'πg.Tie(πF, {}, {})',
        self._build_assign_target(target, assigns), value)
    for t, temp in assigns:
      self._assign_target(t, temp.expr)
      self.block.free_temp(temp)

  def _visit_each(self, nodes):
    for node in nodes:
      self.visit(node)

  def _visit_loop(self, testfunc, node):
    start_label = self.block.genlabel(is_checkpoint=True)
    else_label = self.block.genlabel(is_checkpoint=True)
    end_label = self.block.genlabel()
    with self.block.alloc_temp('bool') as breakvar:
      self.block.push_loop(breakvar)
      self.writer.write('πF.PushCheckpoint({})'.format(else_label))
      self.writer.write('{} = false'.format(breakvar.name))
      self.writer.write_label(start_label)
      self.writer.write_tmpl(textwrap.dedent("""\
          if πE != nil || πR != nil {
          \tcontinue
          }
          if $breakvar {
          \tπF.PopCheckpoint()
          \tgoto Label$end_label
          }"""), breakvar=breakvar.expr, end_label=end_label)
      with self.block.alloc_temp('bool') as testvar:
        testfunc(testvar)
        self.writer.write_tmpl(textwrap.dedent("""\
            if πE != nil || !$testvar {
            \tcontinue
            }
            πF.PushCheckpoint($start_label)\
            """), testvar=testvar.name, start_label=start_label)
      self._visit_each(node.body)
      self.writer.write('continue')
      # End the loop so that break applies to an outer loop if present.
      self.block.pop_loop()
      self.writer.write_label(else_label)
      self.writer.write(textwrap.dedent("""\
          if πE != nil || πR != nil {
          \tcontinue
          }"""))
      if node.orelse:
        self._visit_each(node.orelse)
      self.writer.write_label(end_label)

  def _write_except_block(self, label, exc, except_node):
    self._write_py_context(except_node.lineno)
    self.writer.write_label(label)
    if except_node.name:
      self.block.bind_var(self.writer, except_node.name.id,
                          '{}.ToObject()'.format(exc))
    self._visit_each(except_node.body)
    self.writer.write('πF.RestoreExc(nil, nil)')

  def _write_except_dispatcher(self, exc, tb, handlers):
    """Outputs a Go code that jumps to the appropriate except handler.

    Args:
      exc: Go variable holding the current exception.
      tb: Go variable holding the current exception's traceback.
      handlers: A list of ast.ExceptHandler nodes.

    Returns:
      A list of Go labels indexes corresponding to the exception handlers.

    Raises:
      ParseError: Except handlers are in an invalid order.
    """
    handler_labels = []
    for i, except_node in enumerate(handlers):
      handler_labels.append(self.block.genlabel())
      if except_node.type:
        with self.visit_expr(except_node.type) as type_,\
            self.block.alloc_temp('bool') as is_inst:
          self.writer.write_checked_call2(
              is_inst, 'πg.IsInstance(πF, {}.ToObject(), {})', exc, type_.expr)
          self.writer.write_tmpl(textwrap.dedent("""\
              if $is_inst {
              \tgoto Label$label
              }"""), is_inst=is_inst.expr, label=handler_labels[-1])
      else:
        # This is a bare except. It should be the last handler.
        if i != len(handlers) - 1:
          msg = "default 'except:' must be last"
          raise util.ParseError(except_node, msg)
        self.writer.write('goto Label{}'.format(handler_labels[-1]))
    if handlers[-1].type:
      # There's no bare except, so the fallback is to re-raise.
      self.writer.write(
          'πE = πF.Raise({}.ToObject(), nil, {}.ToObject())'.format(exc, tb))
      self.writer.write('continue')
    return handler_labels

  def _write_py_context(self, lineno):
    if lineno:
      line = self.block.root.buffer.source_line(lineno).strip()
      self.writer.write('// line {}: {}'.format(lineno, line))
      self.writer.write('πF.SetLineno({})'.format(lineno))

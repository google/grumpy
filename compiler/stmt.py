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

import ast
import string
import textwrap

from grumpy.compiler import block
from grumpy.compiler import expr
from grumpy.compiler import expr_visitor
from grumpy.compiler import util


_NATIVE_MODULE_PREFIX = '__go__.'
_NATIVE_TYPE_PREFIX = 'type_'

# Partial list of known vcs for go module import
# Full list can be found at https://golang.org/src/cmd/go/vcs.go
# TODO: Use official vcs.go module instead of partial list
_KNOWN_VCS = [
    'golang.org', 'github.com', 'bitbucket.org', 'git.apache.org',
    'git.openstack.org', 'launchpad.net'
]

_nil_expr = expr.nil_expr


# Parser flags, set on 'from __future__ import *', see parser_flags on
# StatementVisitor below. Note these have the same values as CPython.
FUTURE_DIVISION = 0x2000
FUTURE_ABSOLUTE_IMPORT = 0x4000
FUTURE_PRINT_FUNCTION = 0x10000
FUTURE_UNICODE_LITERALS = 0x20000

# Names for future features in 'from __future__ import *'. Map from name in the
# import statement to a tuple of the flag for parser, and whether we've (grumpy)
# implemented the feature yet.
future_features = {
    "division": (FUTURE_DIVISION, False),
    "absolute_import": (FUTURE_ABSOLUTE_IMPORT, False),
    "print_function": (FUTURE_PRINT_FUNCTION, True),
    "unicode_literals": (FUTURE_UNICODE_LITERALS, False),
}

# These future features are already in the language proper as of 2.6, so
# importing them via __future__ has no effect.
redundant_future_features = ["generators", "with_statement", "nested_scopes"]

late_future = 'from __future__ imports must occur at the beginning of the file'


def import_from_future(node):
  """Processes a future import statement, returning set of flags it defines."""
  assert isinstance(node, ast.ImportFrom)
  assert node.module == '__future__'
  flags = 0
  for alias in node.names:
    name = alias.name
    if name in future_features:
      flag, implemented = future_features[name]
      if not implemented:
        msg = 'future feature {} not yet implemented by grumpy'.format(name)
        raise util.ParseError(node, msg)
      flags |= flag
    elif name == 'braces':
      raise util.ParseError(node, 'not a chance')
    elif name not in redundant_future_features:
      msg = 'future feature {} is not defined'.format(name)
      raise util.ParseError(node, msg)
  return flags


class FutureFeatures(object):
  def __init__(self):
    self.parser_flags = 0
    self.future_lineno = 0


def visit_future(node):
  """Accumulates a set of compiler flags for the compiler __future__ imports.

  Returns an instance of FutureFeatures which encapsulates the flags and the
  line number of the last valid future import parsed. A downstream parser can
  use the latter to detect invalid future imports that appear too late in the
  file.
  """
  # If this is the module node, do an initial pass through the module body's
  # statements to detect future imports and process their directives (i.e.,
  # set compiler flags), and detect ones that don't appear at the beginning of
  # the file. The only things that can proceed a future statement are other
  # future statements and/or a doc string.
  assert isinstance(node, ast.Module)
  ff = FutureFeatures()
  done = False
  found_docstring = False
  for node in node.body:
    if isinstance(node, ast.ImportFrom):
      modname = node.module
      if modname == '__future__':
        if done:
          raise util.ParseError(node, late_future)
        ff.parser_flags |= import_from_future(node)
        ff.future_lineno = node.lineno
      else:
        done = True
    elif isinstance(node, ast.Expr) and not found_docstring:
      e = node.value
      if not isinstance(e, ast.Str): # pylint: disable=simplifiable-if-statement
        done = True
      else:
        found_docstring = True
    else:
      done = True
  return ff


class StatementVisitor(ast.NodeVisitor):
  """Outputs Go statements to a Writer for the given Python nodes."""

  # pylint: disable=invalid-name,missing-docstring

  def __init__(self, block_):
    self.block = block_
    self.future_features = self.block.future_features or FutureFeatures()
    self.writer = util.Writer()
    self.expr_visitor = expr_visitor.ExprVisitor(self.block, self.writer)

  def generic_visit(self, node):
    msg = 'node not yet implemented: {}'.format(type(node).__name__)
    raise util.ParseError(node, msg)

  def visit_Assert(self, node):
    self._write_py_context(node.lineno)
    # TODO: Only evaluate msg if cond is false.
    with self.expr_visitor.visit(node.msg) if node.msg else _nil_expr as msg,\
        self.expr_visitor.visit(node.test) as cond:
      self.writer.write_checked_call1(
          'πg.Assert(πF, {}, {})', cond.expr, msg.expr)

  def visit_AugAssign(self, node):
    op_type = type(node.op)
    if op_type not in StatementVisitor._AUG_ASSIGN_TEMPLATES:
      fmt = 'augmented assignment op not implemented: {}'
      raise util.ParseError(node, fmt.format(op_type.__name__))
    self._write_py_context(node.lineno)
    with self.expr_visitor.visit(node.target) as target,\
        self.expr_visitor.visit(node.value) as value,\
        self.block.alloc_temp() as temp:
      self.writer.write_checked_call2(
          temp, StatementVisitor._AUG_ASSIGN_TEMPLATES[op_type],
          lhs=target.expr, rhs=value.expr)
      self._assign_target(node.target, temp.expr)

  def visit_Assign(self, node):
    self._write_py_context(node.lineno)
    with self.expr_visitor.visit(node.value) as value:
      for target in node.targets:
        self._tie_target(target, value.expr)

  def visit_Break(self, node):
    if not self.block.loop_stack:
      raise util.ParseError(node, "'break' not in loop")
    self._write_py_context(node.lineno)
    self.writer.write('goto Label{}'.format(self.block.top_loop().end_label))

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
        self.block, node.name, global_vars))
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
        with self.expr_visitor.visit(b) as b:
          self.writer.write('{}[{}] = {}'.format(bases.expr, i, b.expr))
      self.writer.write('{} = πg.NewDict()'.format(cls.name))
      self.writer.write_checked_call2(
          mod_name, 'πF.Globals().GetItem(πF, {}.ToObject())',
          self.block.intern('__name__'))
      self.writer.write_checked_call1(
          '{}.SetItem(πF, {}.ToObject(), {})',
          cls.expr, self.block.intern('__module__'), mod_name.expr)
      tmpl = textwrap.dedent("""
          _, πE = πg.NewCode($name, $filename, nil, 0, func(πF *πg.Frame, _ []*πg.Object) (*πg.Object, *πg.BaseException) {
          \tπClass := $cls
          \t_ = πClass""")
      self.writer.write_tmpl(tmpl, name=util.go_str(node.name),
                             filename=util.go_str(self.block.filename),
                             cls=cls.expr)
      with self.writer.indent_block():
        self.writer.write_temp_decls(body_visitor.block)
        self.writer.write_block(body_visitor.block,
                                body_visitor.writer.out.getvalue())
      tmpl = textwrap.dedent("""\
          }).Eval(πF, πF.Globals(), nil, nil)
          if πE != nil {
          \treturn nil, πE
          }
          if $meta, πE = $cls.GetItem(πF, $metaclass_str.ToObject()); πE != nil {
          \treturn nil, πE
          }
          if $meta == nil {
          \t$meta = πg.TypeType.ToObject()
          }""")
      self.writer.write_tmpl(tmpl, meta=meta.name, cls=cls.expr,
                             metaclass_str=self.block.intern('__metaclass__'))
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
    self.writer.write('goto Label{}'.format(self.block.top_loop().start_label))

  def visit_Delete(self, node):
    self._write_py_context(node.lineno)
    for target in node.targets:
      if isinstance(target, ast.Attribute):
        with self.expr_visitor.visit(target.value) as t:
          self.writer.write_checked_call1(
              'πg.DelAttr(πF, {}, {})', t.expr, self.block.intern(target.attr))
      elif isinstance(target, ast.Name):
        self.block.del_var(self.writer, target.id)
      elif isinstance(target, ast.Subscript):
        assert isinstance(target.ctx, ast.Del)
        with self.expr_visitor.visit(target.value) as t,\
            self.expr_visitor.visit(target.slice) as index:
          self.writer.write_checked_call1('πg.DelItem(πF, {}, {})',
                                          t.expr, index.expr)
      else:
        msg = 'del target not implemented: {}'.format(type(target).__name__)
        raise util.ParseError(node, msg)

  def visit_Expr(self, node):
    self._write_py_context(node.lineno)
    self.expr_visitor.visit(node.value).free()

  def visit_For(self, node):
    loop = self.block.push_loop()
    orelse_label = self.block.genlabel() if node.orelse else loop.end_label
    self._write_py_context(node.lineno)
    with self.expr_visitor.visit(node.iter) as iter_expr, \
        self.block.alloc_temp() as i, \
        self.block.alloc_temp() as n:
      self.writer.write_checked_call2(i, 'πg.Iter(πF, {})', iter_expr.expr)
      self.writer.write_label(loop.start_label)
      tmpl = textwrap.dedent("""\
          if $n, πE = πg.Next(πF, $i); πE != nil {
          \tisStop, exc := πg.IsInstance(πF, πE.ToObject(), πg.StopIterationType.ToObject())
          \tif exc != nil {
          \t\tπE = exc
          \t\tcontinue
          \t}
          \tif !isStop {
          \t\tcontinue
          \t}
          \tπE = nil
          \tπF.RestoreExc(nil, nil)
          \tgoto Label$orelse
          }""")
      self.writer.write_tmpl(tmpl, n=n.name, i=i.expr, orelse=orelse_label)
      self._tie_target(node.target, n.expr)
      self._visit_each(node.body)
      self.writer.write('goto Label{}'.format(loop.start_label))

    self.block.pop_loop()
    if node.orelse:
      self.writer.write_label(orelse_label)
      self._visit_each(node.orelse)
    # Avoid label "defined and not used" in case there's no break statements.
    self.writer.write('goto Label{}'.format(loop.end_label))
    self.writer.write_label(loop.end_label)

  def visit_FunctionDef(self, node):
    self._write_py_context(node.lineno)
    func = self.expr_visitor.visit_function_inline(node)
    self.block.bind_var(self.writer, node.name, func.expr)

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
      with self.expr_visitor.visit(ifnode.test) as cond:
        label = self.block.genlabel()
        # We goto the body of the if statement instead of executing it inline
        # because the body itself may be a goto target and Go does not support
        # jumping to targets inside a block.
        with self.block.alloc_temp('bool') as is_true:
          self.writer.write_tmpl(textwrap.dedent("""\
              if $is_true, πE = πg.IsTrue(πF, $cond); πE != nil {
              \treturn nil, πE
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
    for alias in node.names:
      if alias.name.startswith(_NATIVE_MODULE_PREFIX):
        raise util.ParseError(
            node, 'for native imports use "from __go__.xyz import ..." syntax')
      with self._import(alias.name, 0) as mod:
        asname = alias.asname or alias.name.split('.')[0]
        self.block.bind_var(self.writer, asname, mod.expr)

  def visit_ImportFrom(self, node):
    # Wildcard imports are not yet supported.
    for alias in node.names:
      if alias.name == '*':
        msg = 'wildcard member import is not implemented: from %s import %s' % (
            node.module, alias.name)
        raise util.ParseError(node, msg)
    self._write_py_context(node.lineno)
    if node.module.startswith(_NATIVE_MODULE_PREFIX):
      values = [alias.name for alias in node.names]
      with self._import_native(node.module, values) as mod:
        for alias in node.names:
          # Strip the 'type_' prefix when populating the module. This means
          # that, e.g. 'from __go__.foo import type_Bar' will populate foo with
          # a member called Bar, not type_Bar (although the symbol in the
          # importing module will still be type_Bar unless aliased). This bends
          # the semantics of import but makes native module contents more
          # sensible.
          name = alias.name
          if name.startswith(_NATIVE_TYPE_PREFIX):
            name = name[len(_NATIVE_TYPE_PREFIX):]
          with self.block.alloc_temp() as member:
            self.writer.write_checked_call2(
                member, 'πg.GetAttr(πF, {}, {}, nil)',
                mod.expr, self.block.intern(name))
            self.block.bind_var(
                self.writer, alias.asname or alias.name, member.expr)
    elif node.module == '__future__':
      # At this stage all future imports are done in an initial pass (see
      # visit() above), so if they are encountered here after the last valid
      # __future__ then it's a syntax error.
      if node.lineno > self.future_features.future_lineno:
        raise util.ParseError(node, late_future)
    else:
      # NOTE: Assume that the names being imported are all modules within a
      # package. E.g. "from a.b import c" is importing the module c from package
      # a.b, not some member of module b. We cannot distinguish between these
      # two cases at compile time and the Google style guide forbids the latter
      # so we support that use case only.
      for alias in node.names:
        name = '{}.{}'.format(node.module, alias.name)
        with self._import(name, name.count('.')) as mod:
          asname = alias.asname or alias.name
          self.block.bind_var(self.writer, asname, mod.expr)

  def visit_Module(self, node):
    self._visit_each(node.body)

  def visit_Pass(self, node):
    self._write_py_context(node.lineno)

  def visit_Print(self, node):
    if self.future_features.parser_flags & FUTURE_PRINT_FUNCTION:
      raise util.ParseError(node, 'syntax error (print is not a keyword)')
    self._write_py_context(node.lineno)
    with self.block.alloc_temp('[]*πg.Object') as args:
      self.writer.write('{} = make([]*πg.Object, {})'.format(
          args.expr, len(node.values)))
      for i, v in enumerate(node.values):
        with self.expr_visitor.visit(v) as arg:
          self.writer.write('{}[{}] = {}'.format(args.expr, i, arg.expr))
      self.writer.write_checked_call1('πg.Print(πF, {}, {})', args.expr,
                                      'true' if node.nl else 'false')

  def visit_Raise(self, node):
    with self.expr_visitor.visit(node.type) if node.type else _nil_expr as t,\
        self.expr_visitor.visit(node.inst) if node.inst else _nil_expr as inst,\
        self.expr_visitor.visit(node.tback) if node.tback else _nil_expr as tb:
      if node.inst:
        assert node.type, 'raise had inst but no type'
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
      with self.expr_visitor.visit(node.value) as value:
        self.writer.write('return {}, nil'.format(value.expr))
    else:
      self.writer.write('return nil, nil')

  def visit_TryExcept(self, node):  # pylint: disable=g-doc-args
    # The general structure generated by this method is shown below:
    #
    #       checkpoints.Push(Except)
    #       <try body>
    #       Checkpoints.Pop()
    #       <else body>
    #       goto Done
    #     Except:
    #       <dispatch table>
    #     Handler1:
    #       <handler 1 body>
    #       goto Done
    #     Handler2:
    #       <handler 2 body>
    #       goto Done
    #     ...
    #     Done:
    #
    # The dispatch table maps the current exception to the appropriate handler
    # label according to the exception clauses.

    # Write the try body.
    self._write_py_context(node.lineno)
    except_label = self.block.genlabel(is_checkpoint=True)
    done_label = self.block.genlabel()
    self.writer.write('πF.PushCheckpoint({})'.format(except_label))
    self._visit_each(node.body)
    self.writer.write('πF.PopCheckpoint()')
    if node.orelse:
      self._visit_each(node.orelse)
    self.writer.write('goto Label{}'.format(done_label))

    with self.block.alloc_temp('*πg.BaseException') as exc:
      if (len(node.handlers) == 1 and not node.handlers[0].type and
          not node.orelse):
        # When there's just a bare except, no dispatch is required.
        self._write_except_block(except_label, exc.expr, node.handlers[0])
        self.writer.write_label(done_label)
        return

      with self.block.alloc_temp('*πg.Traceback') as tb:
        self.writer.write_label(except_label)
        self.writer.write('{}, {} = πF.ExcInfo()'.format(exc.expr, tb.expr))
        handler_labels = self._write_except_dispatcher(
            exc.expr, tb.expr, node.handlers)

      # Write the bodies of each of the except handlers.
      for handler_label, except_node in zip(handler_labels, node.handlers):
        self._write_except_block(handler_label, exc.expr, except_node)
        self.writer.write('goto Label{}'.format(done_label))

      self.writer.write_label(done_label)

  def visit_TryFinally(self, node):  # pylint: disable=g-doc-args
    # The general structure generated by this method is shown below:
    #
    #       Checkpoints.Push(Finally)
    #       <try body>
    #       Checkpoints.Pop()
    #     Finally:
    #       <finally body>

    # Write the try body.
    self._write_py_context(node.lineno)
    finally_label = self.block.genlabel(is_checkpoint=True)
    self.writer.write('πF.PushCheckpoint({})'.format(finally_label))
    self._visit_each(node.body)
    self.writer.write('πF.PopCheckpoint()')

    # Write the finally body.
    with self.block.alloc_temp('*πg.BaseException') as exc,\
        self.block.alloc_temp('*πg.Traceback') as tb:
      self.writer.write_label(finally_label)
      self.writer.write('πE = nil')
      self.writer.write('{}, {} = πF.RestoreExc(nil, nil)'.format(
          exc.expr, tb.expr))
      self._visit_each(node.finalbody)
      self.writer.write_tmpl(textwrap.dedent("""\
          if $exc != nil {
          \tπE = πF.Raise($exc.ToObject(), nil, $tb.ToObject())
          \tcontinue
          }"""), exc=exc.expr, tb=tb.expr)

  def visit_While(self, node):
    loop = self.block.push_loop()
    self._write_py_context(node.lineno)
    self.writer.write_label(loop.start_label)
    orelse_label = self.block.genlabel() if node.orelse else loop.end_label
    with self.expr_visitor.visit(node.test) as cond,\
        self.block.alloc_temp('bool') as is_true:
      self.writer.write_checked_call2(is_true, 'πg.IsTrue(πF, {})', cond.expr)
      self.writer.write_tmpl(textwrap.dedent("""\
          if !$is_true {
          \tgoto Label$orelse_label
          }"""), is_true=is_true.expr, orelse_label=orelse_label)
      self._visit_each(node.body)
      self.writer.write('goto Label{}'.format(loop.start_label))
    if node.orelse:
      self.writer.write_label(orelse_label)
      self._visit_each(node.orelse)
    # Avoid label "defined and not used" in case there's no break statements.
    self.writer.write('goto Label{}'.format(loop.end_label))
    self.writer.write_label(loop.end_label)
    self.block.pop_loop()

  _AUG_ASSIGN_TEMPLATES = {
      ast.Add: 'πg.IAdd(πF, {lhs}, {rhs})',
      ast.BitAnd: 'πg.IAnd(πF, {lhs}, {rhs})',
      ast.Div: 'πg.IDiv(πF, {lhs}, {rhs})',
      ast.Mod: 'πg.IMod(πF, {lhs}, {rhs})',
      ast.Mult: 'πg.IMul(πF, {lhs}, {rhs})',
      ast.BitOr: 'πg.IOr(πF, {lhs}, {rhs})',
      ast.Sub: 'πg.ISub(πF, {lhs}, {rhs})',
      ast.BitXor: 'πg.IXor(πF, {lhs}, {rhs})',
  }

  def visit_With(self, node):
    self._write_py_context(node.lineno)
    # mgr := EXPR
    with self.expr_visitor.visit(node.context_expr) as mgr,\
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
          mgr.expr, self.block.intern('__exit__'))
      # value := type(mgr).__enter__(mgr)
      self.writer.write_checked_call2(
          value, 'πg.GetAttr(πF, {}.Type().ToObject(), {}, nil)',
          mgr.expr, self.block.intern('__enter__'))
      self.writer.write_checked_call2(
          value, '{}.Call(πF, πg.Args{{{}}}, nil)',
          value.expr, mgr.expr)

      finally_label = self.block.genlabel(is_checkpoint=True)
      self.writer.write('πF.PushCheckpoint({})'.format(finally_label))
      if node.optional_vars:
        self._tie_target(node.optional_vars, value.expr)
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
            $exc, $tb = πF.ExcInfo()
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
            }"""), exc=exc.expr, swallow_exc=swallow_exc_bool.expr)

  def _assign_target(self, target, value):
    if isinstance(target, ast.Name):
      self.block.bind_var(self.writer, target.id, value)
    elif isinstance(target, ast.Attribute):
      assert isinstance(target.ctx, ast.Store)
      with self.expr_visitor.visit(target.value) as obj:
        self.writer.write_checked_call1(
            'πg.SetAttr(πF, {}, {}, {})', obj.expr,
            self.block.intern(target.attr), value)
    elif isinstance(target, ast.Subscript):
      assert isinstance(target.ctx, ast.Store)
      with self.expr_visitor.visit(target.value) as mapping,\
          self.expr_visitor.visit(target.slice) as index:
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

  def _import(self, name, index):
    """Returns an expression for a Module object returned from ImportModule.

    Args:
      name: The fully qualified Python module name, e.g. foo.bar.
      index: The element in the list of modules that this expression should
          select. E.g. for 'foo.bar', 0 corresponds to the package foo and 1
          corresponds to the module bar.
    Returns:
      A Go expression evaluating to an *Object (upcast from a *Module.)
    """
    parts = name.split('.')
    code_objs = []
    for i in xrange(len(parts)):
      package_name = '/'.join(parts[:i + 1])
      if package_name != self.block.full_package_name:
        package = self.block.add_import(package_name)
        code_objs.append('{}.Code'.format(package.alias))
      else:
        code_objs.append('Code')
    mod = self.block.alloc_temp()
    with self.block.alloc_temp('[]*πg.Object') as mod_slice:
      handles_expr = '[]*πg.Code{' + ', '.join(code_objs) + '}'
      self.writer.write_checked_call2(
          mod_slice, 'πg.ImportModule(πF, {}, {})',
          util.go_str(name), handles_expr)
      self.writer.write('{} = {}[{}]'.format(mod.name, mod_slice.expr, index))
    return mod

  def _import_native(self, name, values):
    reflect_package = self.block.add_native_import('reflect')
    import_name = name[len(_NATIVE_MODULE_PREFIX):]
    # Work-around for importing go module from VCS
    # TODO: support bzr|git|hg|svn from any server
    package_name = None
    for x in _KNOWN_VCS:
      if import_name.startswith(x):
        package_name = x + import_name[len(x):].replace('.', '/')
        break
    if not package_name:
      package_name = import_name.replace('.', '/')

    package = self.block.add_native_import(package_name)
    mod = self.block.alloc_temp()
    with self.block.alloc_temp('map[string]*πg.Object') as members:
      self.writer.write_tmpl('$members = map[string]*πg.Object{}',
                             members=members.name)
      for v in values:
        module_attr = v
        with self.block.alloc_temp() as wrapped:
          if v.startswith(_NATIVE_TYPE_PREFIX):
            module_attr = v[len(_NATIVE_TYPE_PREFIX):]
            with self.block.alloc_temp(
                '{}.{}'.format(package.alias, module_attr)) as type_:
              self.writer.write_checked_call2(
                  wrapped, 'πg.WrapNative(πF, {}.ValueOf({}))',
                  reflect_package.alias, type_.expr)
              self.writer.write('{} = {}.Type().ToObject()'.format(
                  wrapped.name, wrapped.expr))
          else:
            self.writer.write_checked_call2(
                wrapped, 'πg.WrapNative(πF, {}.ValueOf({}.{}))',
                reflect_package.alias, package.alias, v)
          self.writer.write('{}[{}] = {}'.format(
              members.name, util.go_str(module_attr), wrapped.expr))
      self.writer.write_checked_call2(mod, 'πg.ImportNativeModule(πF, {}, {})',
                                      util.go_str(name), members.expr)
    return mod

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

  def _write_except_block(self, label, exc, except_node):
    self._write_py_context(except_node.lineno)
    self.writer.write_label(label)
    if except_node.name:
      self.block.bind_var(self.writer, except_node.name.id,
                          '{}.ToObject()'.format(exc))
    self._visit_each(except_node.body)
    self.writer.write('πE = nil')
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
        with self.expr_visitor.visit(except_node.type) as type_,\
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
      line = self.block.lines[lineno - 1].strip()
      self.writer.write('// line {}: {}'.format(lineno, line))
      self.writer.write('πF.SetLineno({})'.format(lineno))

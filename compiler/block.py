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

"""Classes for analyzing and storing the state of Python code blocks."""

from __future__ import unicode_literals

import abc
import collections
import re

from grumpy.compiler import expr
from grumpy.compiler import util
from grumpy.pythonparser import algorithm
from grumpy.pythonparser import ast
from grumpy.pythonparser import source


_non_word_re = re.compile('[^A-Za-z0-9_]')


class Package(object):
  """A Go package import."""

  def __init__(self, name, alias=None):
    self.name = name
      # Use Γ as a separator since it provides readability with a low
      # probability of name collisions.
    self.alias = alias or 'π_' + name.replace('/', 'Γ').replace('.', 'Γ')


class Loop(object):
  """Represents a for or while loop within a particular block."""

  def __init__(self, breakvar):
    self.breakvar = breakvar


class Block(object):
  """Represents a Python block such as a function or class definition."""

  __metaclass__ = abc.ABCMeta

  def __init__(self, parent, name):
    self.root = parent.root if parent else self
    self.parent = parent
    self.name = name
    self.free_temps = set()
    self.used_temps = set()
    self.temp_index = 0
    self.label_count = 0
    self.checkpoints = set()
    self.loop_stack = []
    self.is_generator = False

  @abc.abstractmethod
  def bind_var(self, writer, name, value):
    """Writes Go statements for assigning value to named var in this block.

    This is overridden in the different concrete block types since in Python,
    binding a variable in, e.g. a function is quite different than binding at
    global block.

    Args:
      writer: The Writer object where statements will be written.
      name: The name of the Python variable.
      value: A Go expression to assign to the variable.
    """
    pass

  @abc.abstractmethod
  def del_var(self, writer, name):
    pass

  @abc.abstractmethod
  def resolve_name(self, writer, name):
    """Returns a GeneratedExpr object for accessing the named var in this block.

    This is overridden in the different concrete block types since name
    resolution in Python behaves differently depending on where in what kind of
    block its happening within, e.g. local vars are different than globals.

    Args:
      writer: Writer object where intermediate calculations will be printed.
      name: The name of the Python variable.
    """
    pass

  def genlabel(self, is_checkpoint=False):
    self.label_count += 1
    if is_checkpoint:
      self.checkpoints.add(self.label_count)
    return self.label_count

  def alloc_temp(self, type_='*πg.Object'):
    """Create a new temporary Go variable having type type_ for this block."""
    for v in sorted(self.free_temps, key=lambda k: k.name):
      if v.type_ == type_:
        self.free_temps.remove(v)
        self.used_temps.add(v)
        return v
    self.temp_index += 1
    name = 'πTemp{:03d}'.format(self.temp_index)
    v = expr.GeneratedTempVar(self, name, type_)
    self.used_temps.add(v)
    return v

  def free_temp(self, v):
    """Release the GeneratedTempVar v so it can be reused."""
    self.used_temps.remove(v)
    self.free_temps.add(v)

  def push_loop(self, breakvar):
    loop = Loop(breakvar)
    self.loop_stack.append(loop)
    return loop

  def pop_loop(self):
    self.loop_stack.pop()

  def top_loop(self):
    return self.loop_stack[-1]

  def _resolve_global(self, writer, name):
    result = self.alloc_temp()
    writer.write_checked_call2(
        result, 'πg.ResolveGlobal(πF, {})', self.root.intern(name))
    return result


class ModuleBlock(Block):
  """Python block for a module."""

  def __init__(self, importer, full_package_name,
               filename, src, future_features):
    Block.__init__(self, None, '<module>')
    self.importer = importer
    self.full_package_name = full_package_name
    self.filename = filename
    self.buffer = source.Buffer(src)
    self.strings = set()
    self.future_features = future_features

  def bind_var(self, writer, name, value):
    writer.write_checked_call1(
        'πF.Globals().SetItem(πF, {}.ToObject(), {})',
        self.intern(name), value)

  def del_var(self, writer, name):
    writer.write_checked_call1('πg.DelVar(πF, πF.Globals(), {})',
                               self.intern(name))

  def resolve_name(self, writer, name):
    return self._resolve_global(writer, name)

  def intern(self, s):
    if len(s) > 64 or _non_word_re.search(s):
      return 'πg.NewStr({})'.format(util.go_str(s))
    self.strings.add(s)
    return 'ß' + s


class ClassBlock(Block):
  """Python block for a class definition."""

  def __init__(self, parent, name, global_vars):
    Block.__init__(self, parent, name)
    self.global_vars = global_vars

  def bind_var(self, writer, name, value):
    if name in self.global_vars:
      return self.root.bind_var(writer, name, value)
    writer.write_checked_call1('πClass.SetItem(πF, {}.ToObject(), {})',
                               self.root.intern(name), value)

  def del_var(self, writer, name):
    if name in self.global_vars:
      return self.root.del_var(writer, name)
    writer.write_checked_call1('πg.DelVar(πF, πClass, {})',
                               self.root.intern(name))

  def resolve_name(self, writer, name):
    local = 'nil'
    if name not in self.global_vars:
      # Only look for a local in an outer block when name hasn't been declared
      # global in this block. If it has been declared global then we fallback
      # straight to the global dict.
      block = self.parent
      while not isinstance(block, ModuleBlock):
        if isinstance(block, FunctionBlock) and name in block.vars:
          var = block.vars[name]
          if var.type != Var.TYPE_GLOBAL:
            local = util.adjust_local_name(name)
          # When it is declared global, prefer it to anything in outer blocks.
          break
        block = block.parent
    result = self.alloc_temp()
    writer.write_checked_call2(
        result, 'πg.ResolveClass(πF, πClass, {}, {})',
        local, self.root.intern(name))
    return result


class FunctionBlock(Block):
  """Python block for a function definition."""

  def __init__(self, parent, name, block_vars, is_generator):
    Block.__init__(self, parent, name)
    self.vars = block_vars
    self.parent = parent
    self.is_generator = is_generator

  def bind_var(self, writer, name, value):
    if self.vars[name].type == Var.TYPE_GLOBAL:
      return self.root.bind_var(writer, name, value)
    writer.write('{} = {}'.format(util.adjust_local_name(name), value))

  def del_var(self, writer, name):
    var = self.vars.get(name)
    if not var:
      raise util.ParseError(
          None, 'cannot delete nonexistent local: {}'.format(name))
    if var.type == Var.TYPE_GLOBAL:
      return self.root.del_var(writer, name)
    adjusted_name = util.adjust_local_name(name)
    # Resolve local first to ensure the variable is already bound.
    writer.write_checked_call1('πg.CheckLocal(πF, {}, {})',
                               adjusted_name, util.go_str(name))
    writer.write('{} = πg.UnboundLocal'.format(adjusted_name))

  def resolve_name(self, writer, name):
    block = self
    while not isinstance(block, ModuleBlock):
      if isinstance(block, FunctionBlock):
        var = block.vars.get(name)
        if var:
          if var.type == Var.TYPE_GLOBAL:
            return self._resolve_global(writer, name)
          writer.write_checked_call1('πg.CheckLocal(πF, {}, {})',
                                     util.adjust_local_name(name),
                                     util.go_str(name))
          return expr.GeneratedLocalVar(name)
      block = block.parent
    return self._resolve_global(writer, name)


class Var(object):
  """A Python variable used within a particular block."""

  TYPE_LOCAL = 0
  TYPE_PARAM = 1
  TYPE_GLOBAL = 2

  def __init__(self, name, var_type, arg_index=None):
    self.name = name
    self.type = var_type
    if var_type == Var.TYPE_LOCAL:
      assert arg_index is None
      self.init_expr = 'πg.UnboundLocal'
    elif var_type == Var.TYPE_PARAM:
      assert arg_index is not None
      self.init_expr = 'πArgs[{}]'.format(arg_index)
    else:
      assert arg_index is None
      self.init_expr = None


class BlockVisitor(algorithm.Visitor):
  """Visits nodes in a function or class to determine block variables."""

  # pylint: disable=invalid-name,missing-docstring

  def __init__(self):
    self.vars = collections.OrderedDict()

  def visit_Assign(self, node):
    for target in node.targets:
      self._assign_target(target)
    self.visit(node.value)

  def visit_AugAssign(self, node):
    self._assign_target(node.target)
    self.visit(node.value)

  def visit_ClassDef(self, node):
    self._register_local(node.name)

  def visit_ExceptHandler(self, node):
    if node.name:
      self._register_local(node.name.id)
    self.generic_visit(node)

  def visit_For(self, node):
    self._assign_target(node.target)
    self.generic_visit(node)

  def visit_FunctionDef(self, node):
    # The function being defined is local to this block, i.e. is nested within
    # another function. Note that further nested symbols are not traversed
    # because we don't explicitly visit the function body.
    self._register_local(node.name)

  def visit_Global(self, node):
    for name in node.names:
      self._register_global(node, name)

  def visit_Import(self, node):
    for alias in node.names:
      self._register_local(alias.asname or alias.name.split('.')[0])

  def visit_ImportFrom(self, node):
    for alias in node.names:
      self._register_local(alias.asname or alias.name)

  def visit_With(self, node):
    for item in node.items:
      if item.optional_vars:
        self._assign_target(item.optional_vars)
    self.generic_visit(node)

  def _assign_target(self, target):
    if isinstance(target, ast.Name):
      self._register_local(target.id)
    elif isinstance(target, (ast.Tuple, ast.List)):
      for elt in target.elts:
        self._assign_target(elt)

  def _register_global(self, node, name):
    var = self.vars.get(name)
    if var:
      if var.type == Var.TYPE_PARAM:
        msg = "name '{}' is parameter and global"
        raise util.ParseError(node, msg.format(name))
      if var.type == Var.TYPE_LOCAL:
        msg = "name '{}' is used prior to global declaration"
        raise util.ParseError(node, msg.format(name))
    else:
      self.vars[name] = Var(name, Var.TYPE_GLOBAL)

  def _register_local(self, name):
    if not self.vars.get(name):
      self.vars[name] = Var(name, Var.TYPE_LOCAL)


class FunctionBlockVisitor(BlockVisitor):
  """Visits function nodes to determine variables and generator state."""

  # pylint: disable=invalid-name,missing-docstring

  def __init__(self, node):
    BlockVisitor.__init__(self)
    self.is_generator = False
    node_args = node.args
    args = [a.arg for a in node_args.args]
    if node_args.vararg:
      args.append(node_args.vararg.arg)
    if node_args.kwarg:
      args.append(node_args.kwarg.arg)
    for i, name in enumerate(args):
      if name in self.vars:
        msg = "duplicate argument '{}' in function definition".format(name)
        raise util.ParseError(node, msg)
      self.vars[name] = Var(name, Var.TYPE_PARAM, arg_index=i)

  def visit_Yield(self, unused_node): # pylint: disable=unused-argument
    self.is_generator = True

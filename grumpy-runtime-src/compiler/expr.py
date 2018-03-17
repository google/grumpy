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

"""Classes representing generated expressions."""

from __future__ import unicode_literals

import abc

from grumpy.compiler import util


class GeneratedExpr(object):
  """GeneratedExpr is a generated Go expression in transcompiled output."""

  __metaclass__ = abc.ABCMeta

  def __enter__(self):
    return self

  def __exit__(self, unused_type, unused_value, unused_traceback):
    self.free()

  @abc.abstractproperty
  def expr(self):
    pass

  def free(self):
    pass


class GeneratedTempVar(GeneratedExpr):
  """GeneratedTempVar is an expression result stored in a temporary value."""

  def __init__(self, block_, name, type_):
    self.block = block_
    self.name = name
    self.type_ = type_

  @property
  def expr(self):
    return self.name

  def free(self):
    self.block.free_temp(self)


class GeneratedLocalVar(GeneratedExpr):
  """GeneratedLocalVar is the Go local var corresponding to a Python local."""

  def __init__(self, name):
    self._name = name

  @property
  def expr(self):
    return util.adjust_local_name(self._name)


class GeneratedLiteral(GeneratedExpr):
  """GeneratedLiteral is a generated literal Go expression."""

  def __init__(self, expr):
    self._expr = expr

  @property
  def expr(self):
    return self._expr


nil_expr = GeneratedLiteral('nil')


class BlankVar(GeneratedExpr):
  def __init__(self):
    self.name = '_'

  @property
  def expr(self):
    return '_'


blank_var = BlankVar()

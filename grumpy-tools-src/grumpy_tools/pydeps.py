#!/usr/bin/env python

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

"""Outputs names of modules imported by a script."""
from __future__ import absolute_import

import os
import sys

from .compiler import imputil
from .compiler import util


def main(script=None, modname=None):
  gopath = os.getenv('GOPATH', None)
  if not gopath:
    print >> sys.stderr, 'GOPATH not set'
    return 1

  try:
    imports = imputil.collect_imports(modname, script, gopath)
  except SyntaxError as e:
    print >> sys.stderr, '{}: line {}: invalid syntax: {}'.format(
        e.filename, e.lineno, e.text)
    return 2
  except util.CompileError as e:
    print >> sys.stderr, str(e)
    return 2

  names = set([modname])
  for imp in imports:
    if imp.is_native:
      print imp.name
    else:
      parts = imp.name.split('.')
      # Iterate over all packages and the leaf module.
      for i in xrange(len(parts)):
        name = '.'.join(parts[:i+1])
        if name not in names:
          names.add(name)
          print name


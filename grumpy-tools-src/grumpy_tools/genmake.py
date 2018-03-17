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

"""Generate a Makefile for Python targets in a GOPATH directory."""

import argparse
import os
import subprocess
import sys


parser = argparse.ArgumentParser()
parser.add_argument('dir', help='GOPATH dir to scan for Python modules')
parser.add_argument('-all_target', default='all',
                    help='make target that will build all modules')


def _PrintRule(target, prereqs, rules):
  print '{}: {}'.format(target, ' '.join(prereqs))
  if rules:
    print '\t@mkdir -p $(@D)'
    for rule in rules:
      print '\t@{}'.format(rule)
  print


def main(args):
  try:
    proc = subprocess.Popen('go env GOOS GOARCH', shell=True,
                            stdout=subprocess.PIPE)
  except OSError as e:
    print >> sys.stderr, str(e)
    return 1
  out, _ = proc.communicate()
  if proc.returncode:
    print >> sys.stderr, 'go exited with status: {}'.format(proc.returncode)
    return 1
  goos, goarch = out.split()

  if args.all_target:
    print '{}:\n'.format(args.all_target)

  gopath = os.path.normpath(args.dir)
  pkg_dir = os.path.join(gopath, 'pkg', '{}_{}'.format(goos, goarch))
  pydir = os.path.join(gopath, 'src', '__python__')
  for dirpath, _, filenames in os.walk(pydir):
    for filename in filenames:
      if not filename.endswith('.py'):
        continue
      basename = os.path.relpath(dirpath, pydir)
      if filename != '__init__.py':
        basename = os.path.normpath(
            os.path.join(basename, filename[:-3]))
      modname = basename.replace(os.sep, '.')
      ar_name = os.path.join(pkg_dir, '__python__', basename + '.a')
      go_file = os.path.join(pydir, basename, 'module.go')
      _PrintRule(go_file,
                 [os.path.join(dirpath, filename)],
                 ['grumpc -modname={} $< > $@'.format(modname)])
      recipe = (r"""pydeps -modname=%s $< | awk '{gsub(/\./, "/", $$0); """
                r"""print "%s: %s/__python__/" $$0 ".a"}' > $@""")
      dep_file = os.path.join(pydir, basename, 'module.d')
      _PrintRule(dep_file, [os.path.join(dirpath, filename)],
                 [recipe % (modname, ar_name, pkg_dir)])
      go_package = '__python__/' + basename.replace(os.sep, '/')
      recipe = 'go tool compile -o $@ -p {} -complete -I {} -pack $<'
      _PrintRule(ar_name, [go_file], [recipe.format(go_package, pkg_dir)])
      if args.all_target:
        _PrintRule(args.all_target, [ar_name], [])
      print '-include {}\n'.format(dep_file)


if __name__ == '__main__':
  sys.exit(main(parser.parse_args()))

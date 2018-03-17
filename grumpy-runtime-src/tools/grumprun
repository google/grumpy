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

"""grumprun compiles and runs a snippet of Python using Grumpy.

Usage: $ grumprun -m <module>             # Run the named module.
       $ echo 'print "hola!"' | grumprun  # Execute Python code from stdin.
"""

import argparse
import os
import random
import shutil
import string
import subprocess
import sys
import tempfile

from grumpy.compiler import imputil


parser = argparse.ArgumentParser()
parser.add_argument('-m', '--modname', help='Run the named module')

module_tmpl = string.Template("""\
package main
import (
\t"os"
\t"grumpy"
\tmod "$package"
$imports
)
func main() {
\tgrumpy.ImportModule(grumpy.NewRootFrame(), "traceback")
\tos.Exit(grumpy.RunMain(mod.Code))
}
""")


def main(args):
  gopath = os.getenv('GOPATH', None)
  if not gopath:
    print >> sys.stderr, 'GOPATH not set'
    return 1

  modname = args.modname
  workdir = tempfile.mkdtemp()
  try:
    if modname:
      # Find the script associated with the given module.
      for d in gopath.split(os.pathsep):
        script = imputil.find_script(
            os.path.join(d, 'src', '__python__'), modname)
        if script:
          break
      else:
        print >> sys.stderr, "can't find module", modname
        return 1
    else:
      # Generate a dummy python script on the GOPATH.
      modname = ''.join(random.choice(string.ascii_letters) for _ in range(16))
      py_dir = os.path.join(workdir, 'src', '__python__')
      mod_dir = os.path.join(py_dir, modname)
      os.makedirs(mod_dir)
      script = os.path.join(py_dir, 'module.py')
      with open(script, 'w') as f:
        f.write(sys.stdin.read())
      gopath = gopath + os.pathsep + workdir
      os.putenv('GOPATH', gopath)
      # Compile the dummy script to Go using grumpc.
      fd = os.open(os.path.join(mod_dir, 'module.go'), os.O_WRONLY | os.O_CREAT)
      try:
        p = subprocess.Popen('grumpc ' + script, stdout=fd, shell=True)
        if p.wait():
          return 1
      finally:
        os.close(fd)

    names = imputil.calculate_transitive_deps(modname, script, gopath)
    # Make sure traceback is available in all Python binaries.
    names.add('traceback')
    go_main = os.path.join(workdir, 'main.go')
    package = _package_name(modname)
    imports = ''.join('\t_ "' + _package_name(name) + '"\n' for name in names)
    with open(go_main, 'w') as f:
      f.write(module_tmpl.substitute(package=package, imports=imports))
    return subprocess.Popen('go run ' + go_main, shell=True).wait()
  finally:
    shutil.rmtree(workdir)


def _package_name(modname):
  if modname.startswith('__go__/'):
    return '__python__/' + modname
  return '__python__/' + modname.replace('.', '/')


if __name__ == '__main__':
  sys.exit(main(parser.parse_args()))

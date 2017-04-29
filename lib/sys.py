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

"""System-specific parameters and functions."""

from __go__.os import Args
from __go__.grumpy import SysmoduleDict, SysModules, MaxInt, Stdin, Stdout, Stderr  # pylint: disable=g-multiple-import
from __go__.runtime import Version
from __go__.unicode import MaxRune


__all__ = ('stdin', 'stdout', 'stderr', 'argv', '_goversion',
           'maxint', 'maxsize', 'maxunicode', 'modules', 'py3kwarning',
           'warnoptions', 'byteorder', 'flags', 'exc_info', 'exit')


argv = []
for arg in Args:
    argv.append(arg)

goversion = Version()

stdin = SysmoduleDict['stdin']
stdout = SysmoduleDict['stdout']
stderr = SysmoduleDict['stderr']

maxint = MaxInt
maxsize = maxint
maxunicode = MaxRune
modules = SysModules

py3kwarning = False
warnoptions = []
# TODO: Support actual byteorder
byteorder = 'little'


class _Flags(object):
    """Container class for sys.flags."""
    debug = 0
    py3k_warning = 0
    division_warning = 0
    division_new = 0
    inspect = 0
    interactive = 0
    optimize = 0
    dont_write_bytecode = 0
    no_user_site = 0
    no_site = 0
    ignore_environment = 0
    tabcheck = 0
    verbose = 0
    unicode = 0
    bytes_warning = 0
    hash_randomization = 0


flags = _Flags()


def exc_info():
    e, tb = __frame__().__exc_info__()  # pylint: disable=undefined-variable
    t = None
    if e:
        t = type(e)
    return t, e, tb


def exit(code=None):  # pylint: disable=redefined-builtin
    raise SystemExit(code)


# TODO: Clear this HACK: Should be the last lines of the python part of hybrid stuff
class _SysModule(object):
    def __init__(self):
        for k in ('__name__', '__file__') + __all__:
            SysmoduleDict[k] = globals()[k]
    def __setattr__(self, name, value):
        SysmoduleDict[name] = value
    def __getattribute__(self, name):   # TODO: replace w/ __getattr__ when implemented
        resp = SysmoduleDict.get(name)
        if res is None and name not in SysmoduleDict:
            return super(_SysModule, self).__getattribute__(name)
        return resp

modules = SysModules
modules['sys'] = _SysModule()

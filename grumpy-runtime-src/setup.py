#!/usr/bin/env python
# -*- coding: utf-8 -*-
#
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

"""The setup script."""

import os
import fnmatch
import shutil
import sys
from setuptools import setup, find_packages
from distutils.command.build_py import build_py as BuildPyCommand
from distutils.command.build_ext import build_ext as BuildExtCommand
import subprocess

try:
    with open('README.md') as readme_file:
        readme = readme_file.read()
except:
    readme = ''

setup_requirements = [
    'setuptools>=28.8.0',
]

test_requirements = [
    'pytest',
    # TODO: Put package test requirements here
]

needs_pytest = {'pytest', 'test', 'ptr'}.intersection(sys.argv)
if needs_pytest:
    setup_requirements += ['pytest-runner']

COMMON_OPTIONS = dict(
    version='0.1.8',
    description="Grumpy Runtime & Transpiler",
    long_description=readme,
    author="Dylan Trotter et al.",
    maintainer="Alan Justino et al.",
    maintainer_email="alan.justino@yahoo.com.br",
    url='https://github.com/google/grumpy',
    license="Apache Software License 2.0",
    zip_safe=False,
    keywords='grumpy_runtime',
    python_requires='~=2.7.0',
    classifiers=[
        'Development Status :: 2 - Pre-Alpha',
        'Intended Audience :: Developers',
        'License :: OSI Approved :: Apache Software License',
        'Natural Language :: English',
        "Programming Language :: Python :: 2",
        'Programming Language :: Python :: 2.7',
    ],
    test_suite='tests',
    tests_require=test_requirements,
    setup_requires=setup_requirements,
)


def _run_make(self, *args, **kwargs):
    subprocess.check_call(["""echo "print 'Make Runtime Success'" | make run --debug"""], shell=True)


def _glob_deep(directory, pattern):
    # From: https://stackoverflow.com/a/2186673/798575
    for root, dirs, files in os.walk(directory):
        for basename in files:
            if fnmatch.fnmatch(basename, pattern):
                filename = os.path.join(root, basename)
                yield filename


class BuildMakeCommandInstall(BuildPyCommand):  # Ran on setup.py install
    def run(self, *args, **kwargs):
        # Thanks http://www.digip.org/blog/2011/01/generating-data-files-in-setup.py.html for the tip

        _run_make(self, *args, **kwargs)

        # Makefile creates a "gopath" folder named "build" on root folder. Change it!
        shutil.move('build', 'gopath')

        if not self.dry_run:
            target_dir = os.path.join(self.build_lib, 'grumpy_runtime/data')

            # Remove the symlink used on develop
            shutil.rmtree(target_dir, ignore_errors=True)
            self.mkpath(target_dir)
            shutil.move('gopath', target_dir)

            build_dir = os.path.join(self.build_lib, 'grumpy_runtime')
            built_files = _glob_deep(os.path.join(build_dir, 'gopath'), '*')

            # Strip directory from globbed filenames
            build_dir_len = len(build_dir) + 1 # One more for leading "/"
            built_files = [fn[build_dir_len:] for fn in built_files]

            self.data_files = [
                # (package, src_dir, build_dir, filenames[])
                ('grumpy_runtime', 'grumpy_runtime', build_dir, built_files),
            ]

        super_result = BuildPyCommand.run(self, *args, **kwargs)
        return super_result


class BuildMakeCommandDevelop(BuildExtCommand):  # Ran on setup.py develop
    def run(self, *args, **kwargs):
        # Thanks http://www.digip.org/blog/2011/01/generating-data-files-in-setup.py.html for the tip

        _run_make(self, *args, **kwargs)

        # Makefile creates a "gopath" folder named "build" on root folder. Change it!
        shutil.move('build', 'gopath')

        if not self.dry_run:
            target_dir = os.path.join(self.build_lib, 'grumpy_runtime/data')
            self.mkpath(target_dir)
            shutil.move('gopath', target_dir)

            build_dir = os.path.join(self.build_lib, 'grumpy_runtime')
            built_files = _glob_deep(os.path.join(build_dir, 'gopath'), '*')

            # Strip directory from globbed filenames
            build_dir_len = len(build_dir) + 1 # One more for leading "/"
            built_files = [fn[build_dir_len:] for fn in built_files]

            self.data_files = [
                # (package, src_dir, build_dir, filenames[])
                ('grumpy_runtime', 'grumpy_runtime', build_dir, built_files),
            ]

        super_result = BuildExtCommand.run(self, *args, **kwargs)
        return super_result


GRUMPY_RUNTIME_OPTIONS = dict(
    name='grumpy-runtime',
    requires=['grumpy_tools'],
    install_requires=['grumpy-tools>=0.1.8'],
    packages=find_packages(
        exclude=["*.tests", "*.tests.*", "tests.*", "tests"],
    ),
    include_package_data=True,
    cmdclass={
        'build_py': BuildMakeCommandInstall,   # Ran on setup.py install
        'build_ext': BuildMakeCommandDevelop,  # Ran on setup.py develop
    },
    zip_safe=False,
)
GRUMPY_RUNTIME_OPTIONS.update(COMMON_OPTIONS)

setup(**GRUMPY_RUNTIME_OPTIONS)

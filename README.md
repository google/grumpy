# Grumpy: Go running Python

[![Build Status](https://travis-ci.org/google/grumpy.svg?branch=master)](https://travis-ci.org/google/grumpy)
[![Join the chat at https://gitter.im/grumpy-devel/Lobby](https://badges.gitter.im/grumpy-devel/Lobby.svg)](https://gitter.im/grumpy-devel/Lobby?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

## Overview

Grumpy is a Python to Go source code transcompiler and runtime that is intended
to be a near drop-in replacement for CPython 2.7. The key difference is that it
compiles Python source code to Go source code which is then compiled to native
code, rather than to bytecode. This means that Grumpy has no VM. The compiled Go
source code is a series of calls to the Grumpy runtime, a Go library serving a
similar purpose to the Python C API (although the API is incompatible with
CPython's).

## Limitations

### Things that will probably never be supported by Grumpy

1. `exec`, `eval` and `compile`: These dynamic features of CPython are not
   supported by Grumpy because Grumpy modules consist of statically-compiled Go
   code. Supporting dynamic execution would require bundling Grumpy programs
   with the compilation toolchain, which would be unwieldy and impractically
   slow.

2. C extension modules: Grumpy has a different API and object layout than
   CPython and so supporting C extensions would be difficult. In principle it's
   possible to support them via an API bridge layer like the one that
   [JyNI](http://jyni.org) provides for Jython, but it would be hard to maintain and
   would add significant overhead when calling into and out of extension
   modules.

### Things that Grumpy will support but doesn't yet

There are three basic categories of incomplete functionality:

1. [Language features](https://github.com/google/grumpy/wiki/Missing-features#language-features):
   Most language features are implemented with the notable exception of
   [old-style classes](http://stackoverflow.com/questions/54867/what-is-the-difference-between-old-style-and-new-style-classes-in-python).
   There are also a handful of operators that aren't yet supported.

2. [Builtin functions and types](https://github.com/google/grumpy/wiki/Missing-features#builtins):
   There are a number of missing functions and types in `__builtins__` that have
   not yet been implemented. There are also a lot of methods on builtin types
   that are missing.

3. [Standard library](https://github.com/google/grumpy/wiki/Missing-features#standard-libraries):
   The Python standard library is very large and much of it is pure Python, so
   as the language features and builtins get filled out, many modules will
   just work. But there are also a number of libraries in CPython that are C
   extension modules which will need to be rewritten.

4. C locale support: Go doesn't support locales in the same way that C does. As such,
   some functionality that is locale-dependent may not currently work the same as in
   CPython.

## Running Grumpy Programs

### Method 1: make run:

The simplest way to execute a Grumpy program is to use `make run`, which wraps a
shell script called grumprun that takes Python code on stdin and builds and runs
the code under Grumpy. All of the commands below are assumed to be run from the
root directory of the Grumpy source code distribution:

```
echo "print 'hello, world'" | make run
```

### Method 2: grumpc and grumprun:

For more complicated programs, you'll want to compile your Python source code to
Go using grumpc (the Grumpy compiler) and then build the Go code using `go
build`. Since Grumpy programs are statically linked, all the modules in a
program must be findable by the Grumpy toolchain on the GOPATH. Grumpy looks for
Go packages corresponding to Python modules in the \_\_python\_\_ subdirectory
of the GOPATH. By convention, this subdirectory is also used for staging Python
source code, making it similar to the PYTHONPATH.

The first step is to set up the shell so that the Grumpy toolchain and libraries
can be found. From the root directory of the Grumpy source distribution run:

```
make
export PATH=$PWD/build/bin:$PATH
export GOPATH=$PWD/build
export PYTHONPATH=$PWD/build/lib/python2.7/site-packages
```

You will know things are working if you see the expected output from this
command:

```
echo 'import sys; print sys.version' | grumprun
```

Next, we will write our simple Python module into the \_\_python\_\_ directory:

```
echo 'def hello(): print "hello, world"' > $GOPATH/src/__python__/hello.py
```

To build a Go package from our Python script, run the following:

```
mkdir -p $GOPATH/src/__python__/hello
grumpc -modname=hello $GOPATH/src/__python__/hello.py > \
    $GOPATH/src/__python__/hello/module.go
```

You should now be able to build a Go program that imports the package
"\_\_python\_\_/hello". We can also import this module into Python programs
that are built using grumprun:

```
echo 'from hello import hello; hello()' | grumprun
```

grumprun is doing a few things under the hood here:

1. Compiles the given Python code to a dummy Go package, the same way we
   produced \_\_python\_\_/hello/module.go above
2. Produces a main Go package that imports the Go package from step 1. and
   executes it as our \_\_main\_\_ Python package
3. Executes `go run` on the main package generated in step 2.

## Developing Grumpy

There are three main components and depending on what kind of feature you're
writing, you may need to change one or more of these.

### grumpc

Grumpy converts Python programs into Go programs and `grumpc` is the tool
responsible for parsing Python code and generating Go code from it. `grumpc` is
written in Python and uses the [`pythonparser`](https://github.com/m-labs/pythonparser)
module to accomplish parsing.

The grumpc script itself lives at `tools/grumpc`. It is supported by a number of
Python modules in the `compiler` subdir.

### Grumpy Runtime

The Go code generated by `grumpc` performs operations on data structures that
represent Python objects in running Grumpy programs. These data structures and
operations are defined in the `grumpy` Go library (source is in the runtime
subdir of the source distribution).  This runtime is analogous to the Python C
API and many of the structures and operations defined by `grumpy` have
counterparts in CPython.

### Grumpy Standard Library

Much of the Python standard library is written in Python and thus "just works"
in Grumpy. These parts of the standard library are copied from CPython 2.7
(possibly with light modifications). For licensing reasons, these files are kept
in the `third_party` subdir.

The parts of the standard library that cannot be written in pure Python, e.g.
file and directory operations, are kept in the `lib` subdir. In CPython these
kinds of modules are written as C extensions. In Grumpy they are written in
Python but they use native Go extensions to access facilities not otherwise
available in Python.

### Source Code Overview

- `compiler`: Python package implementating Python -> Go transcompilation logic.
- `lib`: Grumpy-specific Python standard library implementation.
- `runtime`: Go source code for the Grumpy runtime library.
- `third_party/ouroboros`: Pure Python standard libraries copied from the
   [Ouroboros project](https://github.com/pybee/ouroboros).
- `third_party/pypy`: Pure Python standard libraries copied from PyPy.
- `third_party/stdlib`: Pure Python standard libraries copied from CPython.
- `tools`: Transcompilation and utility binaries.

## Contact

Questions? Comments? Drop us a line at [grumpy-users@googlegroups.com](https://groups.google.com/forum/#!forum/grumpy-users)
or join our [Gitter channel](https://gitter.im/grumpy-devel/Lobby)

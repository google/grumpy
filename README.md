# Grumpy: Go running Python

[![Build Status](https://travis-ci.org/google/grumpy.svg?branch=master)](https://travis-ci.org/google/grumpy)

## Overview

Grumpy is a Python to Go source code transcompiler and runtime that is intended
to be a near drop in replacement for CPython 2.7. The key difference is that it
compiles Python source code to Go source code which is then compiled to native
code, rather than to bytecode. This means that Grumpy has no VM. The compiled Go
source code is a series of calls to the Grumpy runtime, a Go library serving a
similar purpose to the Python C API (although the C API is not directly
supported).

## Limitations

### Things that will probably never be supported by Grumpy

1. `exec`, `eval` and `compile`: These dynamic features of CPython are not
   supported by Grumpy because Grumpy modules consist of statically compiled Go
   code. Supporting dynamic execution would require bundling Grumpy programs
   with the compilation toolchain which would be unwieldy and impractically
   slow.

2. C extension modules: Grumpy has a different API and object layout than
   CPython and so supporting C extensions would be difficult. In principle it's
   possible to support them via an API bridge layer like the one that
   [JyNI](http://jyni.org/) provides for Jython but it would be hard to maintain
   and would add significant overhead when calling into and out of extension
   modules.

### Things that Grumpy will support but doesn't yet

There are three basic categories of incomplete functionality:

1. Language features: Most language features are implemented with the notable
   exception of decorators. There are also a handful of operators that aren't
   yet supported.

2. Builtin functions and types: There are a number of missing functions and
   types in `__builtins__` that have not been implemented. There are also a
   lot of methods on builtin types that are missing.

3. Standard library: The Python standard library is very large and much of it
   is pure Python, so as the language features and builtins get filled out, many
   modules will just work. But there are also a number of libraries in CPython
   that are C extension modules that need to be rewritten.
   
 To see the status of a particular feature or standard library module, click
 [here](https://github.com/google/grumpy/wiki/Missing-Features).

## Running Grumpy Programs

### Method 1: grumprun:

The simplest way to execute a Grumpy program is to use `make run`, which wraps a
shell script called grumprun that takes Python code on stdin and builds and runs
the code under Grumpy. All of the commands below are assumed to be run from the
root directory of the Grumpy source code distribution:

```
echo "print 'hello, world'" | make run
```

### Method 2: grumpc:

For more complicated programs you'll want to compile your Python source code to
Go using grumpc (the Grumpy compiler) and then build the Go code using `go
build`.  First, write a simple .py script:

```
echo 'print "hello, world"' > hello.py
```

Next, build the toolchain and export some environment variables that make the
toolchain work:

```
make
export GOPATH=$PWD/build
export PYTHONPATH=$PWD/build/lib/python2.7/site-packages
```

Finally, compile the Python script and build a binary from it:

```
build/bin/grumpc hello.py > hello.go
go build -o hello hello.go
```

Now execute the `./hello` binary to your heart's content.

## Developing Grumpy

There are three main components and depending on what kind of feature you're
writing, you may need to change one or more of these.

### grumpc

Grumpy converts Python programs into Go programs and grumpc is the tool
responsible for parsing Python code and generating Go code from it. grumpc is
written in Python and uses the `ast` module to accomplish parsing.

The grumpc script itself lives at tools/grumpc. It is supported by a number of
Python modules in the compiler subdir.

### Grumpy Runtime

The Go code generated by grumpc performs operations on data structures that
represent Python objects in running Grumpy programs. These data structures and
operations are defined in the `grumpy` Go library (source is in the runtime
subdir of the source distribution).  This runtime is analogous to the Python C
API and many of the structures and operations defined by `grumpy` have
counterparts in CPython.

### Grumpy Standard Library

Much of the Python standard library is written in Python and so it "just works"
in Grumpy. These parts of the standard library are copied from CPython 2.7
(possibly with light modifications). For licensing reasons, these files are kept
in the third_party/stdlib subdir.

The parts of the standard library that cannot be written in pure Python, e.g.
file and directory operations, are kept in the lib subdir. In CPython these
kinds of modules are written as C extensions. In Grumpy they are written in
Python but they use native Go extensions to access facilities not otherwise
available in Python.

### Source Code Overview

- `compiler`: Python package implementating Python -> Go transcompilation logic.
- `lib`: Grumpy-specific Python standard library implementation.
- `runtime`: Go source code for the Grumpy runtime library.
- `third_party/stdlib`: Pure Python standard libraries copied from CPython.
- `tools`: Transcompilation and utility binaries.

## Contact

Questions? Comments? Drop us a line at [grumpy-users@googlegroups.com](https://groups.google.com/forum/#!forum/grumpy-users).

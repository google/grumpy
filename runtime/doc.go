// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
Package grumpy is the Grumpy runtime's Python API, analogous to CPython's C API.

Data model

All Python objects are represented by structs that are binary compatible with
grumpy.Object, so for example the result of the Python expression "object()" is
just an Object pointer. More complex primitive types like str and dict are
represented by structs that augment Object by embedding it as their first field
and holding other data in subsequent fields.  These augmented structs can
themselves be embedded for yet more complex types.

Objects contain a pointer to their Python type, represented by grumpy.Type, and
a pointer to their attribute dict, represented by grumpy.Dict. This dict may be
nil as in the case of str or non-nil as in the case of type objects. Note that
Grumpy objects do not have a refcount since Grumpy relies on Go's garbage
collection to manage object lifetimes.

Every Type object holds references to all its base classes as well as every
class in its MRO list.

Grumpy types also hold a reflect.Type instance known as the type's "basis".  A
type's basis represents the Go struct used to store instances of the type. It
is an important invariant of the Grumpy runtime that an instance of a
particular Python type is stored in the Go struct that is that type's basis.
Violation of this invariant would mean that, for example, a str object could
end up being stored in an unaugmented Object and accessing the str's value
would access invalid memory. This invariant is enforced by Grumpy's API for
primitive types and user defined classes.

Upcasting and downcasting along the basis hierarchy is sometimes necessary, for
example when passing a Str to a function accepting an Object. Upcasts are
accomplished by accessing the embedded base type basis of the subclass, e.g.
accessing the Object member of the Str struct. Downcasting requires
unsafe.Pointer conversions. The safety of these conversions is guaranteed by
the invariant discussed above. E.g. it is valid to cast an *Object with type
StrType to a *Str because it was allocated with storage represented by
StrType's basis, which is struct Str.

Execution model

User defined Python code blocks (modules, classes and functions) are
implemented as Go closures with a state machine that allows the body of the
block to be re-entered for exception handling, yield statements, etc. The
generated code for the body of a code block looks something like this:

	01:	func(f *Frame) (*Object, *BaseException) {
	02:		switch (f.State()) {
	03:		case 0: goto Label0
	04:		case 1: goto Label1
	05:		...
	06:		}
	07:	Label0:
	08:		...
	09:	Label1:
	10:		...
	11:	...
	12:	}

Frame is the basis type for Grumpy's "frame" objects and is very similar to
CPython's type of the same name. The first argument f, therefore, represents a
level in the Python stack. Upon entry into the body, the frame's state variable
is checked and control jumps to the appropriate label. Upon first entry, the
state variable will be 0 and the so execution will start at Label0. Later
invocations may start at other labels. For example, an exception raised in the
try block of a try/finally will cause the function above to return an exception
as its second return value. The caller will then set state to the label
corresponding to the finally clause and call back into the body.

Python exceptions are represented by the BaseException basis struct. Grumpy API
functions and generated code blocks propagate exceptions by returning
*BaseException as their last return value. Exceptions are raised with the
Frame.Raise*() methods which create exception objects to be propagated and set
the exc info indicator for the current frame stack, similar to CPython. Python
except clauses down the stack can then handle the propagated exception.

Each generated body function is owned by a Block struct that is very similar to
CPython's code object. Each Block has a name (e.g. the class' name) and the
filename where the Python code was defined. A block is invoked via the
*Block.Exec method which pushes a new frame on the call stack and then
repeatedly calls the body function. This interplay is depicted below:

	 *Block.Exec
	 --> +-+
	     | | block func
	     |1| --> +-+
	     | |     |2|
	     | | <-- +-+
	     | | --> +-+
	     | |     |2|
	     | | <-- +-+
	 <-- +-+

	1. *Block.Exec repeatedly calls block function until finished or an
	   unhandled exception is encountered

	2. Dispatch switch passes control to appropriate part of block function
	   and executes

When the body returns with a nil exception, the accompanying value is the
returned from the block. If an exception is returned then the "checkpoint
stack" is examined. This data structure stores recovery points within body
that need to be executed when an exception occurs. Expanding on the try/finally
example above, when an exception is raised in the try clause, the finally
checkpoint is popped off the stack and its value is assigned to state. Body
then gets called again and control is passed to the finally label.

To make things concrete, here is a block of code containing a
try/finally:

	01:	try:
	02:		print "foo"
	03:	finally:
	04:		print "bar"

The generated code for this sinippet would look something like this:

	01:	func(f *Frame) (*Object, *BaseException) {
	02:		switch state {
	03:		case 0: goto Label0
	04:		case 1: goto Label1
	05:		}
	06:	Label0:
	07:		// line 1: try:
	08:		f.PushCheckpoint(1)
	09:		// line 2: print foo
	10:		raised = Print(f, []*Object{NewStr("foo").ToObject()})
	11:		if raised != nil {
	12:			return nil, raised
	13:		}
	14:		f.PopCheckpoint()
	15:	Label1:
	16:		exc, tb = πF.RestoreExc(nil, nil)
	17:		// line 4: print bar
	18:		raised = Print(f, []*Object{NewStr("bar").ToObject()})
	19:		if raised != nil {
	20:			return nil, raised
	21:		}
	22:		if exc != nil {
	24:			return nil, f.Raise(exc, nil, tb)
	24:		}
	25:		return None, nil
	26:	}

There are a few relevant things worth noting here:

1. Upon entering the try clause on line 8, a checkpoint pointing to Label1 (the
   finally clause) is pushed onto the stack. If the try clause does not raise,
   the checkpoint is popped on line 14 and control falls through to Label1
   without having to re-enter the body function.

2. Lines 10 and 18 are the two print statements. Exceptions raised during
   execution of these statements are returned immediately. In general,
   Python statements map to one or more Grumpy API function calls which may
   propagate exceptions.

3. Control of the finally clause begins on line 16 where the exception
   indicator is cleared and its original value is stored and re-raised at the
   end of the clause. This matches CPython's behavior where exc info is cleared
   during the finally block.

A stack is used to store checkpoints because checkpoints can be nested.
Continuing the example above, the finally clause itself could be in an except
handler, e.g.:

	01:	try:
	02:		try:
	03:			print "foo"
	04:		finally:
	05:			print "bar"
	06:	except SomeException:
	07:		print "baz"

Once the finally clause completes, it re-raises the exception and control is
passed to the except handler label because it's next in the checkpoint stack.
If the exception is an instance of SomeException then execution continues
within the except clause. If it is some other kind of exception then it will be
returned and control will be passed to the caller to find another checkpoint or
unwind the call stack.

Call model

Python callables are represented by the Function basis struct and the
corresponding Python "function" type. As in CPython, class methods and global
functions are instances of this type. Associated with each instance is a Go
function with the signature:

	func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException)

The args slice and kwargs dict contain the positional and keyword arguments
provided by the caller. Both builtin functions and those in user defined Python
code are called using this convention, however the latter are wrapped in a
layer represented by the FunctionSpec struct that validates arguments and
substitutes absent keyword parameters with their default. Once the spec is
validated, it passes control to the spec function:

	func(f *Frame, args []*Object) (*Object, *BaseException)

Here, the args slice contains an element for each argument present in the
Python function's parameter list, in the same order. Every value is non-nil
since default values have been substituted where necessary by the function
spec. If parameters with the * or ** specifiers are present in the function
signature, they are the last element(s) in args and hold any extra positional
or keyword arguments provided by the caller.

Generated code within the spec function consists of three main parts:

	+----------------------+
	| Spec func            |
	| ---------            |
	| Declare locals       |
	| Declare temporaries  |
	| +------------------+ |
	| | Body func        | |
	| | ----------       | |
	| | Dispatch switch  | |
	| | Labels           | |
	| +------------------+ |
	| Block.Exec(body)     |
	+----------------------+

Locals and temporaries are defined as local variables at the top of the spec
function. Below that, the body function is defined which is stateless except
for what it inherits from its enclosing scope and from the passed frame. This
is important because the body function will be repeatedly reenetered, but all
of the state will have a lifetime longer than any particular invocation because
it belongs to the spec function's scope. Finally, *Block.Exec is called which
drives the state machine, calling into the body function as appropriate.

Generator functions work much the same way except that instead of calling Exec
on the block directly, the block is returned and the generator's next() method
calls Exec until its contents are exhausted.

*/
package grumpy

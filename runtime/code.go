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

package grumpy

import (
	"fmt"
	"reflect"
)

// CodeType is the object representing the Python 'code' type.
var CodeType = newBasisType("code", reflect.TypeOf(Code{}), toCodeUnsafe, ObjectType)

// CodeFlag is a switch controlling the behavior of a Code object.
type CodeFlag int

const (
	// CodeFlagVarArg means a Code object accepts *arg parameters.
	CodeFlagVarArg CodeFlag = 4
	// CodeFlagKWArg means a Code object accepts **kwarg parameters.
	CodeFlagKWArg CodeFlag = 8
)

// Code represents Python 'code' objects.
type Code struct {
	Object
	name     string `attr:"co_name"`
	filename string `attr:"co_filename"`
	// argc is the number of positional arguments.
	argc int `attr:"co_argcount"`
	// minArgc is the number of positional non-keyword arguments (i.e. the
	// minimum number of positional arguments that must be passed).
	minArgc int
	flags   CodeFlag `attr:"co_flags"`
	args    []FunctionArg
	fn      func(*Frame, []*Object) (*Object, *BaseException)
}

// NewCode creates a new Code object that executes the given fn.
func NewCode(name, filename string, args []FunctionArg, flags CodeFlag, fn func(*Frame, []*Object) (*Object, *BaseException)) *Code {
	argc := len(args)
	minArgc := 0
	for ; minArgc < argc; minArgc++ {
		if args[minArgc].Def != nil {
			break
		}
	}
	for _, arg := range args[minArgc:argc] {
		if arg.Def == nil {
			format := "%s() non-keyword arg %s after keyword arg"
			logFatal(fmt.Sprintf(format, name, arg.Name))
		}
	}
	return &Code{Object{typ: CodeType}, name, filename, argc, minArgc, flags, args, fn}
}

func toCodeUnsafe(o *Object) *Code {
	return (*Code)(o.toPointer())
}

// Eval runs the code object c in the context of the given globals.
func (c *Code) Eval(f *Frame, globals *Dict, args Args, kwargs KWArgs) (*Object, *BaseException) {
	// Validate parameters.
	argc := len(args)
	if argc > c.argc && c.flags&CodeFlagVarArg == 0 {
		format := "%s() takes %d arguments (%d given)"
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, c.name, c.argc, argc))
	}
	numBodyArgs := c.argc
	varArgIndex, kwArgIndex := -1, -1
	if c.flags&CodeFlagVarArg != 0 {
		varArgIndex = numBodyArgs
		numBodyArgs++
	}
	if c.flags&CodeFlagKWArg != 0 {
		kwArgIndex = numBodyArgs
		numBodyArgs++
	}
	bodyArgs := f.MakeArgs(numBodyArgs)
	i := 0
	for ; i < argc && i < c.argc; i++ {
		bodyArgs[i] = args[i]
	}
	if varArgIndex != -1 {
		bodyArgs[varArgIndex] = NewTuple(args[i:].makeCopy()...).ToObject()
	}
	var kwargDict *Dict
	if kwArgIndex != -1 {
		kwargDict = NewDict()
		bodyArgs[kwArgIndex] = kwargDict.ToObject()
	}
	for _, kw := range kwargs {
		name := kw.Name
		j := 0
		for ; j < c.argc; j++ {
			if c.args[j].Name == name {
				if bodyArgs[j] != nil {
					format := "%s() got multiple values for keyword argument '%s'"
					return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, c.name, name))
				}
				bodyArgs[j] = kw.Value
				break
			}
		}
		if j == c.argc {
			if kwargDict == nil {
				format := "%s() got an unexpected keyword argument '%s'"
				return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, c.name, name))
			}
			if raised := kwargDict.SetItemString(f, name, kw.Value); raised != nil {
				return nil, raised
			}
		}
	}
	for ; i < c.argc; i++ {
		arg := c.args[i]
		if bodyArgs[i] == nil {
			if arg.Def == nil {
				format := "%s() takes at least %d arguments (%d given)"
				return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, c.name, c.minArgc, argc))
			}
			bodyArgs[i] = arg.Def
		}
	}
	oldExc, oldTraceback := f.ExcInfo()
	next := newChildFrame(f)
	next.code = c
	next.globals = globals
	ret, raised := c.fn(next, bodyArgs)
	next.release()
	f.FreeArgs(bodyArgs)
	if raised == nil {
		// Restore exc_info to what it was when we left the previous
		// frame.
		f.RestoreExc(oldExc, oldTraceback)
		if ret == nil {
			ret = None
		}
	} else {
		_, tb := f.ExcInfo()
		tb = newTraceback(f, tb)
		f.RestoreExc(raised, tb)
	}
	return ret, raised
}

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
	argc      int      `attr:"co_argcount"`
	flags     CodeFlag `attr:"co_flags"`
	paramSpec *ParamSpec
	fn        func(*Frame, []*Object) (*Object, *BaseException)
}

// NewCode creates a new Code object that executes the given fn.
func NewCode(name, filename string, params []Param, flags CodeFlag, fn func(*Frame, []*Object) (*Object, *BaseException)) *Code {
	s := NewParamSpec(name, params, flags&CodeFlagVarArg != 0, flags&CodeFlagKWArg != 0)
	return &Code{Object{typ: CodeType}, name, filename, len(params), flags, s, fn}
}

func toCodeUnsafe(o *Object) *Code {
	return (*Code)(o.toPointer())
}

// Eval runs the code object c in the context of the given globals.
func (c *Code) Eval(f *Frame, globals *Dict, args Args, kwargs KWArgs) (*Object, *BaseException) {
	validated := f.MakeArgs(c.paramSpec.Count)
	if raised := c.paramSpec.Validate(f, validated, args, kwargs); raised != nil {
		return nil, raised
	}
	oldExc, oldTraceback := f.ExcInfo()
	next := newChildFrame(f)
	next.code = c
	next.globals = globals
	ret, raised := c.fn(next, validated)
	next.release()
	f.FreeArgs(validated)
	if raised == nil {
		// Restore exc_info to what it was when we left the previous
		// frame.
		f.RestoreExc(oldExc, oldTraceback)
		if ret == nil {
			ret = None
		}
	} else {
		_, tb := f.ExcInfo()
		if f.code != nil {
			// The root frame has no code object so don't include it
			// in the traceback.
			tb = newTraceback(f, tb)
		}
		f.RestoreExc(raised, tb)
	}
	return ret, raised
}

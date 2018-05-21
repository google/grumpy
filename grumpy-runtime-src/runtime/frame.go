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

// RunState represents the current point of execution within a Python function.
type RunState int

const (
	notBaseExceptionMsg = "exceptions must be derived from BaseException, not %q"
)

// Frame represents Python 'frame' objects.
type Frame struct {
	Object
	*threadState
	back *Frame `attr:"f_back"`
	// checkpoints holds RunState values that should be executed when
	// unwinding the stack due to an exception. Examples of checkpoints
	// include exception handlers and finally blocks.
	checkpoints []RunState
	state       RunState
	globals     *Dict `attr:"f_globals"`
	lineno      int   `attr:"f_lineno"`
	code        *Code `attr:"f_code"`
	taken       bool
}

// NewRootFrame creates a Frame that is the bottom of a new stack.
func NewRootFrame() *Frame {
	f := &Frame{Object: Object{typ: FrameType}}
	f.pushFrame(nil)
	return f
}

// newChildFrame creates a new Frame whose parent frame is back.
func newChildFrame(back *Frame) *Frame {
	f := back.frameCache
	if f == nil {
		f = &Frame{Object: Object{typ: FrameType}}
	} else {
		back.frameCache, f.back = f.back, nil
		// Reset local state late.
		f.checkpoints = f.checkpoints[:0]
		f.state = 0
		f.lineno = 0
	}
	f.pushFrame(back)
	return f
}

func (f *Frame) release() {
	if !f.taken {
		// TODO: Track cache depth and release memory.
		f.frameCache, f.back = f, f.frameCache
		// Clear pointers early.
		f.setDict(nil)
		f.globals = nil
		f.code = nil
	} else if f.back != nil {
		f.back.taken = true
	}
}

// pushFrame adds f to the top of the stack, above back.
func (f *Frame) pushFrame(back *Frame) {
	f.back = back
	if back == nil {
		f.threadState = newThreadState()
	} else {
		f.threadState = back.threadState
	}
}

func toFrameUnsafe(o *Object) *Frame {
	return (*Frame)(o.toPointer())
}

// Globals returns the globals dict for this frame.
func (f *Frame) Globals() *Dict {
	return f.globals
}

// ToObject upcasts f to an Object.
func (f *Frame) ToObject() *Object {
	return &f.Object
}

// SetLineno sets the current line number for the frame.
func (f *Frame) SetLineno(lineno int) {
	f.lineno = lineno
}

// State returns the current run state for f.
func (f *Frame) State() RunState {
	return f.state
}

// PushCheckpoint appends state to the end of f's checkpoint stack.
func (f *Frame) PushCheckpoint(state RunState) {
	f.checkpoints = append(f.checkpoints, state)
}

// PopCheckpoint removes the last element of f's checkpoint stack and returns
// it.
func (f *Frame) PopCheckpoint() {
	numCheckpoints := len(f.checkpoints)
	if numCheckpoints == 0 {
		f.state = -1
	} else {
		f.state = f.checkpoints[numCheckpoints-1]
		f.checkpoints = f.checkpoints[:numCheckpoints-1]
	}
}

// Raise creates an exception and sets the exc info indicator in a way that is
// compatible with the Python raise statement. The semantics are non-trivial
// and are best described here:
// https://docs.python.org/2/reference/simple_stmts.html#the-raise-statement
// If typ, inst and tb are all nil then the currently active exception and
// traceback according to ExcInfo will be used. Raise returns the exception to
// propagate.
func (f *Frame) Raise(typ *Object, inst *Object, tb *Object) *BaseException {
	if typ == nil && inst == nil && tb == nil {
		exc, excTraceback := f.ExcInfo()
		if exc != nil {
			typ = exc.ToObject()
		}
		if excTraceback != nil {
			tb = excTraceback.ToObject()
		}
	}
	if typ == nil {
		typ = None
	}
	if inst == nil {
		inst = None
	}
	if tb == nil {
		tb = None
	}
	// Build the exception if necessary.
	if typ.isInstance(TypeType) {
		t := toTypeUnsafe(typ)
		if !t.isSubclass(BaseExceptionType) {
			return f.RaiseType(TypeErrorType, fmt.Sprintf(notBaseExceptionMsg, t.Name()))
		}
		if !inst.isInstance(t) {
			var args Args
			if inst.isInstance(TupleType) {
				args = toTupleUnsafe(inst).elems
			} else if inst != None {
				args = []*Object{inst}
			}
			var raised *BaseException
			if inst, raised = typ.Call(f, args, nil); raised != nil {
				return raised
			}
		}
	} else if inst == None {
		inst = typ
	} else {
		return f.RaiseType(TypeErrorType, "instance exception may not have a separate value")
	}
	// Validate the exception and traceback object and raise them.
	if !inst.isInstance(BaseExceptionType) {
		return f.RaiseType(TypeErrorType, fmt.Sprintf(notBaseExceptionMsg, inst.typ.Name()))
	}
	e := toBaseExceptionUnsafe(inst)
	var traceback *Traceback
	if tb == None {
		traceback = newTraceback(f, nil)
	} else if tb.isInstance(TracebackType) {
		traceback = toTracebackUnsafe(tb)
	} else {
		return f.RaiseType(TypeErrorType, "raise: arg 3 must be a traceback or None")
	}
	f.RestoreExc(e, traceback)
	return e
}

// RaiseType constructs a new object of type t, passing a single str argument
// built from msg and throws the constructed object.
func (f *Frame) RaiseType(t *Type, msg string) *BaseException {
	return f.Raise(t.ToObject(), NewStr(msg).ToObject(), nil)
}

// ExcInfo returns the exception currently being handled by f's thread and the
// associated traceback.
func (f *Frame) ExcInfo() (*BaseException, *Traceback) {
	return f.threadState.excValue, f.threadState.excTraceback
}

// RestoreExc assigns the exception currently being handled by f's thread and
// the associated traceback. The previously set values are returned.
func (f *Frame) RestoreExc(e *BaseException, tb *Traceback) (*BaseException, *Traceback) {
	f.threadState.excValue, e = e, f.threadState.excValue
	f.threadState.excTraceback, tb = tb, f.threadState.excTraceback
	return e, tb
}

func (f *Frame) reprEnter(o *Object) bool {
	if f.threadState.reprState[o] {
		return true
	}
	if f.threadState.reprState == nil {
		f.threadState.reprState = map[*Object]bool{}
	}
	f.threadState.reprState[o] = true
	return false
}

func (f *Frame) reprLeave(o *Object) {
	delete(f.threadState.reprState, o)
}

// MakeArgs returns an Args slice with the given length. The slice may have
// been previously used, but all elements will be set to nil.
func (f *Frame) MakeArgs(n int) Args {
	if n == 0 {
		return nil
	}
	if n > argsCacheArgc {
		return make(Args, n)
	}
	numEntries := len(f.threadState.argsCache)
	if numEntries == 0 {
		return make(Args, n, argsCacheArgc)
	}
	args := f.threadState.argsCache[numEntries-1]
	f.threadState.argsCache = f.threadState.argsCache[:numEntries-1]
	return args[:n]
}

// FreeArgs clears the elements of args and returns it to the system. It may
// later be returned by calls to MakeArgs and therefore references to slices of
// args should not be held.
func (f *Frame) FreeArgs(args Args) {
	if cap(args) < argsCacheArgc {
		return
	}
	numEntries := len(f.threadState.argsCache)
	if numEntries >= argsCacheSize {
		return
	}
	// Clear args so we don't unnecessarily hold references.
	for i := len(args) - 1; i >= 0; i-- {
		args[i] = nil
	}
	f.threadState.argsCache = f.threadState.argsCache[:numEntries+1]
	f.threadState.argsCache[numEntries] = args
}

// FrameType is the object representing the Python 'frame' type.
var FrameType = newBasisType("frame", reflect.TypeOf(Frame{}), toFrameUnsafe, ObjectType)

func frameExcClear(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "__exc_clear__", args, FrameType); raised != nil {
		return nil, raised
	}
	toFrameUnsafe(args[0]).RestoreExc(nil, nil)
	return None, nil
}

func frameExcInfo(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkMethodVarArgs(f, "__exc_info__", args, FrameType); raised != nil {
		return nil, raised
	}
	excObj, tbObj := None, None
	e, tb := toFrameUnsafe(args[0]).ExcInfo()
	if e != nil {
		excObj = e.ToObject()
	}
	if tb != nil {
		tbObj = tb.ToObject()
	}
	return NewTuple2(excObj, tbObj).ToObject(), nil
}

func initFrameType(dict map[string]*Object) {
	FrameType.flags &= ^(typeFlagInstantiable | typeFlagBasetype)
	dict["__exc_clear__"] = newBuiltinFunction("__exc_clear__", frameExcClear).ToObject()
	dict["__exc_info__"] = newBuiltinFunction("__exc_info__", frameExcInfo).ToObject()
}

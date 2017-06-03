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
	"testing"
)

const (
	testRunStateInvalid RunState = -1
	testRunStateStart            = 0
	testRunStateDone             = 1
)

func TestFrameArgsCache(t *testing.T) {
	f := NewRootFrame()
	args1 := f.MakeArgs(0)
	if args1 != nil {
		t.Errorf("f.MakeArgs(0) = %v, want nil", args1)
	}
	args2 := f.MakeArgs(1)
	if argc := len(args2); argc != 1 {
		t.Errorf("f.MakeArgs(1) had len %d, want len 1", argc)
	}
	if arg0 := args2[0]; arg0 != nil {
		t.Errorf("f.MakeArgs(1)[0] = %v, want nil", arg0)
	}
	args2[0] = None // Make sure this is cleared in MakeArgs result below.
	f.FreeArgs(args2)
	args3 := f.MakeArgs(1)
	if &args2[0] != &args3[0] {
		t.Error("freed arg slice not returned from cache")
	}
	if arg0 := args3[0]; arg0 != nil {
		t.Errorf("f.MakeArgs(1)[0] = %v, want nil", arg0)
	}
	args4 := f.MakeArgs(1000)
	if argc := len(args4); argc != 1000 {
		t.Errorf("f.MakeArgs(1000) had len %d, want len 1", argc)
	}
	// Make sure the cache doesn't overflow when overfed.
	for i := 0; i < 100; i++ {
		f.FreeArgs(make(Args, argsCacheArgc))
	}
	args5 := f.MakeArgs(2)
	if argc := len(args5); argc != 2 {
		t.Errorf("f.MakeArgs(2) had len %d, want len 2", argc)
	}
}

func TestFramePopCheckpoint(t *testing.T) {
	cases := []struct {
		states  []RunState
		want    RunState
		wantTop RunState
	}{
		{nil, testRunStateInvalid, testRunStateInvalid},
		{[]RunState{testRunStateDone}, testRunStateDone, testRunStateInvalid},
		{[]RunState{testRunStateDone, testRunStateStart}, testRunStateStart, testRunStateDone},
	}
	for _, cas := range cases {
		f := NewRootFrame()
		for _, state := range cas.states {
			f.PushCheckpoint(state)
		}
		f.PopCheckpoint()
		if got := f.State(); got != cas.want {
			t.Errorf("%#v.Pop() = %v, want %v", f, got, cas.want)
		} else if numCheckpoints := len(f.checkpoints); numCheckpoints == 0 && cas.wantTop != testRunStateInvalid {
			t.Errorf("%#v.Pop() left checkpoint stack empty, wanted top to be %v", f, cas.wantTop)
		} else if numCheckpoints != 0 && f.checkpoints[numCheckpoints-1] != cas.wantTop {
			t.Errorf("%#v.Pop() left checkpoint stack with top %v, want %v", f, f.State(), cas.wantTop)
		}
	}
}

func TestFramePushCheckpoint(t *testing.T) {
	f := NewRootFrame()
	states := []RunState{testRunStateStart, testRunStateDone}
	for _, state := range states {
		f.PushCheckpoint(state)
		if numCheckpoints := len(f.checkpoints); numCheckpoints == 0 {
			t.Errorf("%#v.Push(%v) left checkpoint stack empty, want non-empty", f, state)
		} else if top := f.checkpoints[numCheckpoints-1]; top != state {
			t.Errorf("%#v.Push(%v) left checkpoint stack top %v, want %v", f, state, top, state)
		}
	}
}

func TestFrameRaise(t *testing.T) {
	f := NewRootFrame()
	raisedFrame := NewRootFrame()
	raisedFrame.RestoreExc(mustCreateException(ValueErrorType, "foo"), newTraceback(raisedFrame, nil))
	tb := newTraceback(f, nil)
	multiArgExc := toBaseExceptionUnsafe(mustNotRaise(ExceptionType.Call(f, []*Object{None, None}, nil)))
	barType := newTestClass("Bar", []*Type{ExceptionType}, newStringDict(map[string]*Object{
		"__new__": newBuiltinFunction("__new__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			return NewStr("Bar").ToObject(), nil
		}).ToObject(),
	}))
	otherTB := newTraceback(NewRootFrame(), nil)
	cases := []struct {
		f       *Frame
		typ     *Object
		inst    *Object
		tb      *Object
		wantExc *BaseException
		wantTB  *Traceback
	}{
		{f, nil, nil, nil, mustCreateException(TypeErrorType, `exceptions must be derived from BaseException, not "NoneType"`), tb},
		{raisedFrame, nil, nil, nil, mustCreateException(ValueErrorType, "foo"), newTraceback(raisedFrame, nil)},
		{f, ExceptionType.ToObject(), nil, nil, mustCreateException(ExceptionType, ""), tb},
		{f, newObject(ExceptionType), nil, nil, toBaseExceptionUnsafe(newObject(ExceptionType)), tb},
		{f, NewInt(42).ToObject(), nil, nil, mustCreateException(TypeErrorType, `exceptions must be derived from BaseException, not "int"`), tb},
		{f, ObjectType.ToObject(), nil, nil, mustCreateException(TypeErrorType, `exceptions must be derived from BaseException, not "object"`), tb},
		{f, AssertionErrorType.ToObject(), NewStr("foo").ToObject(), nil, mustCreateException(AssertionErrorType, "foo"), tb},
		{f, ExceptionType.ToObject(), NewTuple(None, None).ToObject(), nil, multiArgExc, tb},
		{f, ExceptionType.ToObject(), mustCreateException(KeyErrorType, "foo").ToObject(), nil, mustCreateException(KeyErrorType, "foo"), tb},
		{f, barType.ToObject(), nil, nil, mustCreateException(TypeErrorType, `exceptions must be derived from BaseException, not "str"`), tb},
		{f, newObject(StopIterationType), NewInt(123).ToObject(), nil, mustCreateException(TypeErrorType, "instance exception may not have a separate value"), tb},
		{f, RuntimeErrorType.ToObject(), nil, otherTB.ToObject(), mustCreateException(RuntimeErrorType, ""), otherTB},
		{f, RuntimeErrorType.ToObject(), nil, newObject(ObjectType), mustCreateException(TypeErrorType, "raise: arg 3 must be a traceback or None"), tb},
	}
	for _, cas := range cases {
		call := fmt.Sprintf("frame.Raise(%v, %v, %v)", cas.typ, cas.inst, cas.tb)
		// Not using cas.f here because the test may require
		// cas.f is uncleared. If a fresh frame is desired for
		// a particular test, use f.
		f.RestoreExc(nil, nil)
		cas.f.Raise(cas.typ, cas.inst, cas.tb)
		if got := cas.f.Raise(cas.typ, cas.inst, cas.tb); !exceptionsAreEquivalent(got, cas.wantExc) {
			t.Errorf("%s raised %v, want %v", call, got, cas.wantExc)
		} else if e, gotTB := cas.f.ExcInfo(); got != e {
			t.Errorf("%s raised %v but ExcInfo returned %v", call, got, e)
		} else if !reflect.DeepEqual(gotTB, cas.wantTB) {
			t.Errorf("%s produced traceback %v, want %v", call, gotTB, cas.wantTB)
		}
	}
}

func TestFrameRaiseType(t *testing.T) {
	fun := newBuiltinFunction("TestFrameRaiseType", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		if raised := checkFunctionArgs(f, "TestFrameRaiseType", args, TypeType, StrType); raised != nil {
			return nil, raised
		}
		return nil, f.RaiseType(toTypeUnsafe(args[0]), toStrUnsafe(args[1]).Value())
	}).ToObject()
	cases := []invokeTestCase{
		{args: wrapArgs(TypeErrorType, "bar"), wantExc: mustCreateException(TypeErrorType, "bar")},
		{args: wrapArgs(ExceptionType, ""), wantExc: toBaseExceptionUnsafe(mustNotRaise(ExceptionType.Call(NewRootFrame(), wrapArgs(""), nil)))},
		{args: wrapArgs(TupleType, "foo"), wantExc: mustCreateException(TypeErrorType, `exceptions must be derived from BaseException, not "tuple"`)},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestReprEnterLeave(t *testing.T) {
	o := newObject(ObjectType)
	parent := NewRootFrame()
	child := newChildFrame(parent)
	wantParent := NewRootFrame()
	wantParent.reprState = map[*Object]bool{o: true}
	child.reprEnter(o)
	// After child.reprEnter(), expect the parent's reprState to contain o.
	if wantChild := newChildFrame(parent); !reflect.DeepEqual(child, wantChild) {
		t.Errorf("reprEnter: child frame was %#v, want %#v", child, wantChild)
	} else if !reflect.DeepEqual(parent, wantParent) {
		t.Errorf("reprEnter: parent frame was %#v, want %#v", parent, wantParent)
	} else {
		wantParent.reprState = map[*Object]bool{}
		child.reprLeave(o)
		// Expect the parent's reprState to be empty after reprLeave().
		if wantChild := newChildFrame(parent); !reflect.DeepEqual(child, wantChild) {
			t.Errorf("reprLeave: child frame was %#v, want %#v", child, wantChild)
		} else if !reflect.DeepEqual(parent, wantParent) {
			t.Errorf("reprLeave: parent frame was %#v, want %#v", parent, wantParent)
		}
	}
}

func TestFrameRoot(t *testing.T) {
	f1 := NewRootFrame()
	f2 := newChildFrame(f1)
	frames := []*Frame{f1, f2, newChildFrame(f2)}
	for _, f := range frames {
		if f.threadState != f1.threadState {
			t.Errorf("frame threadState was %v, want %v", f.threadState, f1.threadState)
		}
	}
}

func TestFrameExcInfo(t *testing.T) {
	raisedFrame := NewRootFrame()
	raisedExc := mustCreateException(ValueErrorType, "foo")
	raisedTB := newTraceback(raisedFrame, nil)
	raisedFrame.RestoreExc(raisedExc, raisedTB)
	cases := []invokeTestCase{
		{args: wrapArgs(NewRootFrame()), want: NewTuple(None, None).ToObject()},
		{args: wrapArgs(raisedFrame), want: newTestTuple(raisedExc, raisedTB).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(FrameType, "__exc_info__", &cas); err != "" {
			t.Error(err)
		}
	}
}

type checkInvokeResultType int

const (
	checkInvokeResultOk                  checkInvokeResultType = iota
	checkInvokeResultExceptionMismatch                         = iota
	checkInvokeResultReturnValueMismatch                       = iota
)

func checkResult(got, want *Object, gotExc, wantExc *BaseException) checkInvokeResultType {
	if !exceptionsAreEquivalent(gotExc, wantExc) {
		return checkInvokeResultExceptionMismatch
	}
	if got == nil && want == nil {
		return checkInvokeResultOk
	}
	if got != nil && want != nil {
		// Compare exceptions for equivalence but other objects using
		// __eq__.
		if got.isInstance(BaseExceptionType) && want.isInstance(BaseExceptionType) &&
			exceptionsAreEquivalent(toBaseExceptionUnsafe(got), toBaseExceptionUnsafe(want)) {
			return checkInvokeResultOk
		}
		f := NewRootFrame()
		eq, raised := Eq(f, got, want)
		if raised != nil {
			panic(raised)
		}
		b, raised := IsTrue(f, eq)
		if raised != nil {
			panic(raised)
		}
		if b {
			return checkInvokeResultOk
		}
	}
	return checkInvokeResultReturnValueMismatch
}

func checkInvokeResult(callable *Object, args Args, wantRet *Object, wantExc *BaseException) (*Object, checkInvokeResultType) {
	return checkInvokeResultKwargs(callable, args, nil, wantRet, wantExc)
}

func checkInvokeResultKwargs(callable *Object, args Args, kwargs KWArgs, wantRet *Object, wantExc *BaseException) (*Object, checkInvokeResultType) {
	ret, raised := callable.Call(NewRootFrame(), args, kwargs)
	switch checkResult(ret, wantRet, raised, wantExc) {
	case checkInvokeResultExceptionMismatch:
		if raised == nil {
			return nil, checkInvokeResultExceptionMismatch
		}
		return raised.ToObject(), checkInvokeResultExceptionMismatch
	case checkInvokeResultReturnValueMismatch:
		return ret, checkInvokeResultReturnValueMismatch
	default:
		return nil, checkInvokeResultOk
	}
}

type invokeTestCase struct {
	args    Args
	kwargs  KWArgs
	want    *Object
	wantExc *BaseException
}

func runInvokeTestCase(callable *Object, cas *invokeTestCase) string {
	f := NewRootFrame()
	name := mustNotRaise(GetAttr(f, callable, internedName, NewStr("<unknown>").ToObject()))
	if !name.isInstance(StrType) {
		return fmt.Sprintf("%v.__name__ is not a string", callable)
	}
	// Get repr of args before the call in case any of the args are mutated.
	argsRepr, raised := Repr(f, NewTuple(cas.args...).ToObject())
	if raised != nil {
		panic(raised)
	}
	nameStr := toStrUnsafe(name).Value()
	switch got, match := checkInvokeResultKwargs(callable, cas.args, cas.kwargs, cas.want, cas.wantExc); match {
	case checkInvokeResultExceptionMismatch:
		return fmt.Sprintf("%s%s raised %v, want %v", nameStr, argsRepr.Value(), got, cas.wantExc)
	case checkInvokeResultReturnValueMismatch:
		return fmt.Sprintf("%s%s = %v, want %v", nameStr, argsRepr.Value(), got, cas.want)
	default:
		return ""
	}
}

func runInvokeMethodTestCase(t *Type, methodName string, cas *invokeTestCase) string {
	method := mustNotRaise(GetAttr(NewRootFrame(), t.ToObject(), NewStr(methodName), nil))
	return runInvokeTestCase(method, cas)
}

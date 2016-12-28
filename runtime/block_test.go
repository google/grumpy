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
	"testing"
)

func TestBlockExecTryExcept(t *testing.T) {
	type blockArgs struct {
		f *Frame
		e *BaseException
	}
	args := []blockArgs{}
	b := NewBlock("<test>", "foo.py", func(f *Frame, _ *Object) (*Object, *BaseException) {
		e, _ := f.ExcInfo()
		switch f.State() {
		case 0:
			goto Start
		case 1:
			goto Except
		default:
			t.Fatalf("got invalid state %d", f.State())
		}
	Start:
		args = append(args, blockArgs{f, e})
		f.PushCheckpoint(1)
		return nil, f.RaiseType(RuntimeErrorType, "foo")
	Except:
		f.RestoreExc(nil, nil)
		args = append(args, blockArgs{f, e})
		return None, nil
	})
	b.Exec(newFrame(nil), NewDict())
	wantExc := mustCreateException(RuntimeErrorType, "foo")
	if len(args) != 2 {
		t.Errorf("called %d times, want 2", len(args))
	} else if args[0].f == nil || args[0].f != args[1].f {
		t.Errorf("got frames %v %v, want non-nil but identical frames", args[0].f, args[1].f)
	} else if args[0].e != nil {
		t.Errorf("call 0 raised %v, want nil", args[0].e)
	} else if !exceptionsAreEquivalent(args[1].e, wantExc) {
		t.Errorf("call 1 raised %v, want %v", args[1].e, wantExc)
	}
}

func TestBlockExecRaises(t *testing.T) {
	var f1, f2 *Frame
	globals := NewDict()
	b1 := NewBlock("<b1>", "foo.py", func(f *Frame, _ *Object) (*Object, *BaseException) {
		f1 = f
		return nil, f.RaiseType(ValueErrorType, "bar")
	})
	b2 := NewBlock("<b2>", "foo.py", func(f *Frame, _ *Object) (*Object, *BaseException) {
		f2 = f
		return b1.Exec(f, globals)
	})
	b2.Exec(newFrame(nil), NewDict())
	e, tb := f1.ExcInfo()
	wantExc := mustCreateException(ValueErrorType, "bar")
	if !exceptionsAreEquivalent(e, wantExc) {
		t.Errorf("raised %v, want %v", e, wantExc)
	}
	wantTraceback := newTraceback(f1, nil)
	if !reflect.DeepEqual(tb, wantTraceback) {
		t.Errorf("exception traceback was %+v, want %+v", tb, wantTraceback)
	}
}

func TestBlockExecRestoreExc(t *testing.T) {
	e := mustCreateException(RuntimeErrorType, "uh oh")
	ranB1, ranB2 := false, false
	globals := NewDict()
	b1 := NewBlock("<b1>", "foo.py", func(f *Frame, _ *Object) (*Object, *BaseException) {
		if got, _ := f.ExcInfo(); got != e {
			t.Errorf("ExcInfo() = %v, want %v", got, e)
		}
		f.RestoreExc(nil, nil)
		ranB1 = true
		return None, nil
	})
	b2 := NewBlock("<b2>", "foo.py", func(f *Frame, _ *Object) (*Object, *BaseException) {
		f.RestoreExc(e, newTraceback(f, nil))
		b1.Exec(f, globals)
		// The exception was cleared by b1 but when returning to b2, it
		// should have been restored.
		if got, _ := f.ExcInfo(); got != e {
			t.Errorf("ExcInfo() = %v, want <nil>", got)
		}
		f.RestoreExc(nil, nil)
		ranB2 = true
		return None, nil
	})
	b2.Exec(newFrame(nil), globals)
	if !ranB1 {
		t.Error("b1 did not run")
	}
	if !ranB2 {
		t.Error("b2 did not run")
	}
}

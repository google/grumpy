package grumpy

import (
	"testing"
)

func TestXxx(t *testing.T) {
}

func TestNewCodeKeywordsCheck(t *testing.T) {
	oldLogFatal := logFatal
	defer func() { logFatal = oldLogFatal }()
	var got string
	logFatal = func(msg string) {
		got = msg
	}
	NewCode("foo", "foo.py", []FunctionArg{{"bar", None}, {"baz", nil}}, 0, nil)
	if want := "foo() non-keyword arg baz after keyword arg"; got != want {
		t.Errorf("NewCode logged %q, want %q", got, want)
	}
}

func TestNewCode(t *testing.T) {
	testFunc := newBuiltinFunction("TestNewCode", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		if raised := checkFunctionVarArgs(f, "TestNewCode", args, CodeType); raised != nil {
			return nil, raised
		}
		return toCodeUnsafe(args[0]).Eval(f, nil, args[1:], kwargs)
	})
	fn := func(f *Frame, args []*Object) (*Object, *BaseException) {
		return NewTuple(Args(args).makeCopy()...).ToObject(), nil
	}
	cases := []invokeTestCase{
		invokeTestCase{args: wrapArgs(NewCode("f1", "foo.py", nil, 0, fn)), want: NewTuple().ToObject()},
		invokeTestCase{args: wrapArgs(NewCode("f2", "foo.py", []FunctionArg{{"a", nil}}, 0, fn), 123), want: newTestTuple(123).ToObject()},
		invokeTestCase{args: wrapArgs(NewCode("f2", "foo.py", []FunctionArg{{"a", nil}}, 0, fn)), kwargs: wrapKWArgs("a", "apple"), want: newTestTuple("apple").ToObject()},
		invokeTestCase{args: wrapArgs(NewCode("f2", "foo.py", []FunctionArg{{"a", nil}}, 0, fn)), kwargs: wrapKWArgs("b", "bear"), wantExc: mustCreateException(TypeErrorType, "f2() got an unexpected keyword argument 'b'")},
		invokeTestCase{args: wrapArgs(NewCode("f2", "foo.py", []FunctionArg{{"a", nil}}, 0, fn)), wantExc: mustCreateException(TypeErrorType, "f2() takes at least 1 arguments (0 given)")},
		invokeTestCase{args: wrapArgs(NewCode("f2", "foo.py", []FunctionArg{{"a", nil}}, 0, fn), 1, 2, 3), wantExc: mustCreateException(TypeErrorType, "f2() takes 1 arguments (3 given)")},
		invokeTestCase{args: wrapArgs(NewCode("f3", "foo.py", []FunctionArg{{"a", nil}, {"b", nil}}, 0, fn), 1, 2), want: newTestTuple(1, 2).ToObject()},
		invokeTestCase{args: wrapArgs(NewCode("f3", "foo.py", []FunctionArg{{"a", nil}, {"b", nil}}, 0, fn), 1), kwargs: wrapKWArgs("b", "bear"), want: newTestTuple(1, "bear").ToObject()},
		invokeTestCase{args: wrapArgs(NewCode("f3", "foo.py", []FunctionArg{{"a", nil}, {"b", nil}}, 0, fn)), kwargs: wrapKWArgs("b", "bear", "a", "apple"), want: newTestTuple("apple", "bear").ToObject()},
		invokeTestCase{args: wrapArgs(NewCode("f3", "foo.py", []FunctionArg{{"a", nil}, {"b", nil}}, 0, fn), 1), kwargs: wrapKWArgs("a", "alpha"), wantExc: mustCreateException(TypeErrorType, "f3() got multiple values for keyword argument 'a'")},
		invokeTestCase{args: wrapArgs(NewCode("f4", "foo.py", []FunctionArg{{"a", nil}, {"b", None}}, 0, fn), 123), want: newTestTuple(123, None).ToObject()},
		invokeTestCase{args: wrapArgs(NewCode("f4", "foo.py", []FunctionArg{{"a", nil}, {"b", None}}, 0, fn), 123, "bar"), want: newTestTuple(123, "bar").ToObject()},
		invokeTestCase{args: wrapArgs(NewCode("f4", "foo.py", []FunctionArg{{"a", nil}, {"b", None}}, 0, fn)), kwargs: wrapKWArgs("a", 123, "b", "bar"), want: newTestTuple(123, "bar").ToObject()},
		invokeTestCase{args: wrapArgs(NewCode("f5", "foo.py", []FunctionArg{{"a", nil}}, CodeFlagVarArg, fn), 1), want: newTestTuple(1, NewTuple()).ToObject()},
		invokeTestCase{args: wrapArgs(NewCode("f5", "foo.py", []FunctionArg{{"a", nil}}, CodeFlagVarArg, fn), 1, 2, 3), want: newTestTuple(1, newTestTuple(2, 3)).ToObject()},
		invokeTestCase{args: wrapArgs(NewCode("f6", "foo.py", []FunctionArg{{"a", nil}}, CodeFlagKWArg, fn), "bar"), want: newTestTuple("bar", NewDict()).ToObject()},
		invokeTestCase{args: wrapArgs(NewCode("f6", "foo.py", []FunctionArg{{"a", nil}}, CodeFlagKWArg, fn)), kwargs: wrapKWArgs("a", "apple", "b", "bear"), want: newTestTuple("apple", newTestDict("b", "bear")).ToObject()},
		invokeTestCase{args: wrapArgs(NewCode("f6", "foo.py", []FunctionArg{{"a", nil}}, CodeFlagKWArg, fn), "bar"), kwargs: wrapKWArgs("b", "baz", "c", "qux"), want: newTestTuple("bar", newTestDict("b", "baz", "c", "qux")).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(testFunc.ToObject(), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestCodeEvalRestoreExc(t *testing.T) {
	e := mustCreateException(RuntimeErrorType, "uh oh")
	ranC1, ranC2 := false, false
	globals := NewDict()
	c1 := NewCode("<c1>", "foo.py", nil, 0, func(f *Frame, _ []*Object) (*Object, *BaseException) {
		if got, _ := f.ExcInfo(); got != e {
			t.Errorf("ExcInfo() = %v, want %v", got, e)
		}
		f.RestoreExc(nil, nil)
		ranC1 = true
		return None, nil
	})
	c2 := NewCode("<c2>", "foo.py", nil, 0, func(f *Frame, _ []*Object) (*Object, *BaseException) {
		f.RestoreExc(e, newTraceback(f, nil))
		c1.Eval(f, globals, nil, nil)
		// The exception was cleared by c1 but when returning to c2, it
		// should have been restored.
		if got, _ := f.ExcInfo(); got != e {
			t.Errorf("ExcInfo() = %v, want <nil>", got)
		}
		f.RestoreExc(nil, nil)
		ranC2 = true
		return None, nil
	})
	c2.Eval(newFrame(nil), globals, nil, nil)
	if !ranC1 {
		t.Error("c1 did not run")
	}
	if !ranC2 {
		t.Error("c2 did not run")
	}
}

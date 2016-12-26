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
	NewCode("foo", []FunctionArg{{"bar", None}, {"baz", nil}}, 0, nil)
	if want := "foo() non-keyword arg baz after keyword arg"; got != want {
		t.Errorf("NewCode logged %q, want %q", got, want)
	}
}

func TestNewCode(t *testing.T) {
	testFunc := newBuiltinFunction("TestNewCode", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
		if raised := checkFunctionVarArgs(f, "TestNewCode", args, CodeType); raised != nil {
			return nil, raised
		}
		return toCodeUnsafe(args[0]).call(f, args[1:], kwargs)
	})
	fn := func(f *Frame, args []*Object) (*Object, *BaseException) {
		return NewTuple(Args(args).makeCopy()...).ToObject(), nil
	}
	cases := []invokeTestCase{
		invokeTestCase{args: wrapArgs(NewCode("f1", nil, 0, fn)), want: NewTuple().ToObject()},
		invokeTestCase{args: wrapArgs(NewCode("f2", []FunctionArg{{"a", nil}}, 0, fn), 123), want: newTestTuple(123).ToObject()},
		invokeTestCase{args: wrapArgs(NewCode("f2", []FunctionArg{{"a", nil}}, 0, fn)), kwargs: wrapKWArgs("a", "apple"), want: newTestTuple("apple").ToObject()},
		invokeTestCase{args: wrapArgs(NewCode("f2", []FunctionArg{{"a", nil}}, 0, fn)), kwargs: wrapKWArgs("b", "bear"), wantExc: mustCreateException(TypeErrorType, "f2() got an unexpected keyword argument 'b'")},
		invokeTestCase{args: wrapArgs(NewCode("f2", []FunctionArg{{"a", nil}}, 0, fn)), wantExc: mustCreateException(TypeErrorType, "f2() takes at least 1 arguments (0 given)")},
		invokeTestCase{args: wrapArgs(NewCode("f2", []FunctionArg{{"a", nil}}, 0, fn), 1, 2, 3), wantExc: mustCreateException(TypeErrorType, "f2() takes 1 arguments (3 given)")},
		invokeTestCase{args: wrapArgs(NewCode("f3", []FunctionArg{{"a", nil}, {"b", nil}}, 0, fn), 1, 2), want: newTestTuple(1, 2).ToObject()},
		invokeTestCase{args: wrapArgs(NewCode("f3", []FunctionArg{{"a", nil}, {"b", nil}}, 0, fn), 1), kwargs: wrapKWArgs("b", "bear"), want: newTestTuple(1, "bear").ToObject()},
		invokeTestCase{args: wrapArgs(NewCode("f3", []FunctionArg{{"a", nil}, {"b", nil}}, 0, fn)), kwargs: wrapKWArgs("b", "bear", "a", "apple"), want: newTestTuple("apple", "bear").ToObject()},
		invokeTestCase{args: wrapArgs(NewCode("f3", []FunctionArg{{"a", nil}, {"b", nil}}, 0, fn), 1), kwargs: wrapKWArgs("a", "alpha"), wantExc: mustCreateException(TypeErrorType, "f3() got multiple values for keyword argument 'a'")},
		invokeTestCase{args: wrapArgs(NewCode("f4", []FunctionArg{{"a", nil}, {"b", None}}, 0, fn), 123), want: newTestTuple(123, None).ToObject()},
		invokeTestCase{args: wrapArgs(NewCode("f4", []FunctionArg{{"a", nil}, {"b", None}}, 0, fn), 123, "bar"), want: newTestTuple(123, "bar").ToObject()},
		invokeTestCase{args: wrapArgs(NewCode("f4", []FunctionArg{{"a", nil}, {"b", None}}, 0, fn)), kwargs: wrapKWArgs("a", 123, "b", "bar"), want: newTestTuple(123, "bar").ToObject()},
		invokeTestCase{args: wrapArgs(NewCode("f5", []FunctionArg{{"a", nil}}, CodeFlagVarArg, fn), 1), want: newTestTuple(1, NewTuple()).ToObject()},
		invokeTestCase{args: wrapArgs(NewCode("f5", []FunctionArg{{"a", nil}}, CodeFlagVarArg, fn), 1, 2, 3), want: newTestTuple(1, newTestTuple(2, 3)).ToObject()},
		invokeTestCase{args: wrapArgs(NewCode("f6", []FunctionArg{{"a", nil}}, CodeFlagKWArg, fn), "bar"), want: newTestTuple("bar", NewDict()).ToObject()},
		invokeTestCase{args: wrapArgs(NewCode("f6", []FunctionArg{{"a", nil}}, CodeFlagKWArg, fn)), kwargs: wrapKWArgs("a", "apple", "b", "bear"), want: newTestTuple("apple", newTestDict("b", "bear")).ToObject()},
		invokeTestCase{args: wrapArgs(NewCode("f6", []FunctionArg{{"a", nil}}, CodeFlagKWArg, fn), "bar"), kwargs: wrapKWArgs("b", "baz", "c", "qux"), want: newTestTuple("bar", newTestDict("b", "baz", "c", "qux")).ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(testFunc.ToObject(), &cas); err != "" {
			t.Error(err)
		}
	}
}

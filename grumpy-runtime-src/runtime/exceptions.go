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

var (
	// ArithmeticErrorType corresponds to the Python type 'ArithmeticError'.
	ArithmeticErrorType = newSimpleType("ArithmeticError", StandardErrorType)
	// AssertionErrorType corresponds to the Python type 'AssertionError'.
	AssertionErrorType = newSimpleType("AssertionError", StandardErrorType)
	// AttributeErrorType corresponds to the Python type 'AttributeError'.
	AttributeErrorType = newSimpleType("AttributeError", StandardErrorType)
	// BytesWarningType corresponds to the Python type 'BytesWarning'.
	BytesWarningType = newSimpleType("BytesWarning", WarningType)
	// DeprecationWarningType corresponds to the Python type 'DeprecationWarning'.
	DeprecationWarningType = newSimpleType("DeprecationWarning", WarningType)
	// EnvironmentErrorType corresponds to the Python type
	// 'EnvironmentError'.
	EnvironmentErrorType = newSimpleType("EnvironmentError", StandardErrorType)
	// EOFErrorType corresponds to the Python type 'EOFError'.
	EOFErrorType = newSimpleType("EOFError", StandardErrorType)
	// ExceptionType corresponds to the Python type 'Exception'.
	ExceptionType = newSimpleType("Exception", BaseExceptionType)
	// FutureWarningType corresponds to the Python type 'FutureWarning'.
	FutureWarningType = newSimpleType("FutureWarning", WarningType)
	// ImportErrorType corresponds to the Python type 'ImportError'.
	ImportErrorType = newSimpleType("ImportError", StandardErrorType)
	// ImportWarningType corresponds to the Python type 'ImportWarning'.
	ImportWarningType = newSimpleType("ImportWarning", WarningType)
	// IndexErrorType corresponds to the Python type 'IndexError'.
	IndexErrorType = newSimpleType("IndexError", LookupErrorType)
	// IOErrorType corresponds to the Python type 'IOError'.
	IOErrorType = newSimpleType("IOError", EnvironmentErrorType)
	// KeyboardInterruptType corresponds to the Python type 'KeyboardInterrupt'.
	KeyboardInterruptType = newSimpleType("KeyboardInterrupt", BaseExceptionType)
	// KeyErrorType corresponds to the Python type 'KeyError'.
	KeyErrorType = newSimpleType("KeyError", LookupErrorType)
	// LookupErrorType corresponds to the Python type 'LookupError'.
	LookupErrorType = newSimpleType("LookupError", StandardErrorType)
	// MemoryErrorType corresponds to the Python type 'MemoryError'.
	MemoryErrorType = newSimpleType("MemoryError", StandardErrorType)
	// NameErrorType corresponds to the Python type 'NameError'.
	NameErrorType = newSimpleType("NameError", StandardErrorType)
	// NotImplementedErrorType corresponds to the Python type
	// 'NotImplementedError'.
	NotImplementedErrorType = newSimpleType("NotImplementedError", RuntimeErrorType)
	// OSErrorType corresponds to the Python type 'OSError'.
	OSErrorType = newSimpleType("OSError", EnvironmentErrorType)
	// OverflowErrorType corresponds to the Python type 'OverflowError'.
	OverflowErrorType = newSimpleType("OverflowError", ArithmeticErrorType)
	// PendingDeprecationWarningType corresponds to the Python type 'PendingDeprecationWarning'.
	PendingDeprecationWarningType = newSimpleType("PendingDeprecationWarning", WarningType)
	// ReferenceErrorType corresponds to the Python type 'ReferenceError'.
	ReferenceErrorType = newSimpleType("ReferenceError", StandardErrorType)
	// RuntimeErrorType corresponds to the Python type 'RuntimeError'.
	RuntimeErrorType = newSimpleType("RuntimeError", StandardErrorType)
	// RuntimeWarningType corresponds to the Python type 'RuntimeWarning'.
	RuntimeWarningType = newSimpleType("RuntimeWarning", WarningType)
	// StandardErrorType corresponds to the Python type 'StandardError'.
	StandardErrorType = newSimpleType("StandardError", ExceptionType)
	// StopIterationType corresponds to the Python type 'StopIteration'.
	StopIterationType = newSimpleType("StopIteration", ExceptionType)
	// SyntaxErrorType corresponds to the Python type 'SyntaxError'.
	SyntaxErrorType = newSimpleType("SyntaxError", StandardErrorType)
	// SyntaxWarningType corresponds to the Python type 'SyntaxWarning'.
	SyntaxWarningType = newSimpleType("SyntaxWarning", WarningType)
	// SystemErrorType corresponds to the Python type 'SystemError'.
	SystemErrorType = newSimpleType("SystemError", StandardErrorType)
	// SystemExitType corresponds to the Python type 'SystemExit'.
	SystemExitType = newSimpleType("SystemExit", BaseExceptionType)
	// TypeErrorType corresponds to the Python type 'TypeError'.
	TypeErrorType = newSimpleType("TypeError", StandardErrorType)
	// UnboundLocalErrorType corresponds to the Python type
	// 'UnboundLocalError'.
	UnboundLocalErrorType = newSimpleType("UnboundLocalError", NameErrorType)
	// UnicodeDecodeErrorType corresponds to the Python type 'UnicodeDecodeError'.
	UnicodeDecodeErrorType = newSimpleType("UnicodeDecodeError", ValueErrorType)
	// UnicodeEncodeErrorType corresponds to the Python type 'UnicodeEncodeError'.
	UnicodeEncodeErrorType = newSimpleType("UnicodeEncodeError", ValueErrorType)
	// UnicodeErrorType corresponds to the Python type 'UnicodeError'.
	UnicodeErrorType = newSimpleType("UnicodeError", ValueErrorType)
	// UnicodeWarningType corresponds to the Python type 'UnicodeWarning'.
	UnicodeWarningType = newSimpleType("UnicodeWarning", WarningType)
	// UserWarningType corresponds to the Python type 'UserWarning'.
	UserWarningType = newSimpleType("UserWarning", WarningType)
	// ValueErrorType corresponds to the Python type 'ValueError'.
	ValueErrorType = newSimpleType("ValueError", StandardErrorType)
	// WarningType corresponds to the Python type 'Warning'.
	WarningType = newSimpleType("Warning", ExceptionType)
	// ZeroDivisionErrorType corresponds to the Python type
	// 'ZeroDivisionError'.
	ZeroDivisionErrorType = newSimpleType("ZeroDivisionError", ArithmeticErrorType)
)

func systemExitInit(f *Frame, o *Object, args Args, kwargs KWArgs) (*Object, *BaseException) {
	baseExceptionInit(f, o, args, kwargs)
	code := None
	if len(args) > 0 {
		code = args[0]
	}
	if raised := SetAttr(f, o, NewStr("code"), code); raised != nil {
		return nil, raised
	}
	return None, nil
}

func initSystemExitType(map[string]*Object) {
	SystemExitType.slots.Init = &initSlot{systemExitInit}
}

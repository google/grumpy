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
	"io/ioutil"
	"os"
	"testing"
)

func TestImportModule(t *testing.T) {
	f := NewRootFrame()
	invalidModule := newObject(ObjectType)
	foo := newTestModule("foo", "foo/__init__.py")
	bar := newTestModule("foo.bar", "foo/bar/__init__.py")
	baz := newTestModule("foo.bar.baz", "foo/bar/baz/__init__.py")
	qux := newTestModule("foo.qux", "foo/qux/__init__.py")
	fooCode := NewCode("<module>", "foo/__init__.py", nil, 0, func(*Frame, []*Object) (*Object, *BaseException) { return None, nil })
	barCode := NewCode("<module>", "foo/bar/__init__.py", nil, 0, func(*Frame, []*Object) (*Object, *BaseException) { return None, nil })
	bazCode := NewCode("<module>", "foo/bar/baz/__init__.py", nil, 0, func(*Frame, []*Object) (*Object, *BaseException) { return None, nil })
	quxCode := NewCode("<module>", "foo/qux/__init__.py", nil, 0, func(*Frame, []*Object) (*Object, *BaseException) { return None, nil })
	raisesCode := NewCode("<module", "raises.py", nil, 0, func(f *Frame, _ []*Object) (*Object, *BaseException) {
		return nil, f.RaiseType(ValueErrorType, "uh oh")
	})
	circularImported := false
	circularCode := NewCode("<module>", "circular.py", nil, 0, func(f *Frame, _ []*Object) (*Object, *BaseException) {
		if circularImported {
			return nil, f.RaiseType(AssertionErrorType, "circular imported recursively")
		}
		circularImported = true
		if _, raised := ImportModule(f, "circular"); raised != nil {
			return nil, raised
		}
		return None, nil
	})
	circularTestModule := newTestModule("circular", "circular.py").ToObject()
	clearCode := NewCode("<module>", "clear.py", nil, 0, func(f *Frame, _ []*Object) (*Object, *BaseException) {
		if _, raised := SysModules.DelItemString(f, "clear"); raised != nil {
			return nil, raised
		}
		return None, nil
	})
	// NOTE: This test progressively evolves sys.modules, checking after
	// each test case that it's populated appropriately.
	oldSysModules := SysModules
	oldModuleRegistry := moduleRegistry
	defer func() {
		SysModules = oldSysModules
		moduleRegistry = oldModuleRegistry
	}()
	SysModules = newStringDict(map[string]*Object{"invalid": invalidModule})
	moduleRegistry = map[string]*Code{
		"foo":         fooCode,
		"foo.bar":     barCode,
		"foo.bar.baz": bazCode,
		"foo.qux":     quxCode,
		"raises":      raisesCode,
		"circular":    circularCode,
		"clear":       clearCode,
	}
	cases := []struct {
		name           string
		want           *Object
		wantExc        *BaseException
		wantSysModules *Dict
	}{
		{
			"noexist",
			nil,
			mustCreateException(ImportErrorType, "noexist"),
			newStringDict(map[string]*Object{"invalid": invalidModule}),
		},
		{
			"invalid",
			NewTuple(invalidModule).ToObject(),
			nil,
			newStringDict(map[string]*Object{"invalid": invalidModule}),
		},
		{
			"raises",
			nil,
			mustCreateException(ValueErrorType, "uh oh"),
			newStringDict(map[string]*Object{"invalid": invalidModule}),
		},
		{
			"foo",
			NewTuple(foo.ToObject()).ToObject(),
			nil,
			newStringDict(map[string]*Object{
				"foo":     foo.ToObject(),
				"invalid": invalidModule,
			}),
		},
		{
			"foo",
			NewTuple(foo.ToObject()).ToObject(),
			nil,
			newStringDict(map[string]*Object{
				"foo":     foo.ToObject(),
				"invalid": invalidModule,
			}),
		},
		{
			"foo.qux",
			NewTuple(foo.ToObject(), qux.ToObject()).ToObject(),
			nil,
			newStringDict(map[string]*Object{
				"foo":     foo.ToObject(),
				"foo.qux": qux.ToObject(),
				"invalid": invalidModule,
			}),
		},
		{
			"foo.bar.baz",
			NewTuple(foo.ToObject(), bar.ToObject(), baz.ToObject()).ToObject(),
			nil,
			newStringDict(map[string]*Object{
				"foo":         foo.ToObject(),
				"foo.bar":     bar.ToObject(),
				"foo.bar.baz": baz.ToObject(),
				"foo.qux":     qux.ToObject(),
				"invalid":     invalidModule,
			}),
		},
		{
			"circular",
			NewTuple(circularTestModule).ToObject(),
			nil,
			newStringDict(map[string]*Object{
				"circular":    circularTestModule,
				"foo":         foo.ToObject(),
				"foo.bar":     bar.ToObject(),
				"foo.bar.baz": baz.ToObject(),
				"foo.qux":     qux.ToObject(),
				"invalid":     invalidModule,
			}),
		},
		{
			"clear",
			nil,
			mustCreateException(ImportErrorType, "Loaded module clear not found in sys.modules"),
			newStringDict(map[string]*Object{
				"circular":    circularTestModule,
				"foo":         foo.ToObject(),
				"foo.bar":     bar.ToObject(),
				"foo.bar.baz": baz.ToObject(),
				"foo.qux":     qux.ToObject(),
				"invalid":     invalidModule,
			}),
		},
	}
	for _, cas := range cases {
		mods, raised := ImportModule(f, cas.name)
		var got *Object
		if raised == nil {
			got = NewTuple(mods...).ToObject()
		}
		switch checkResult(got, cas.want, raised, cas.wantExc) {
		case checkInvokeResultExceptionMismatch:
			t.Errorf("ImportModule(%q) raised %v, want %v", cas.name, raised, cas.wantExc)
		case checkInvokeResultReturnValueMismatch:
			t.Errorf("ImportModule(%q) = %v, want %v", cas.name, got, cas.want)
		}
		ne := mustNotRaise(NE(f, SysModules.ToObject(), cas.wantSysModules.ToObject()))
		b, raised := IsTrue(f, ne)
		if raised != nil {
			panic(raised)
		}
		if b {
			msg := "ImportModule(%q): sys.modules = %v, want %v"
			t.Errorf(msg, cas.name, SysModules, cas.wantSysModules)
		}
	}
}

func TestModuleGetNameAndFilename(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, m *Module) (*Tuple, *BaseException) {
		name, raised := m.GetName(f)
		if raised != nil {
			return nil, raised
		}
		filename, raised := m.GetFilename(f)
		if raised != nil {
			return nil, raised
		}
		return newTestTuple(name, filename), nil
	})
	cases := []invokeTestCase{
		{args: wrapArgs(newModule("foo", "foo.py")), want: newTestTuple("foo", "foo.py").ToObject()},
		{args: Args{mustNotRaise(ModuleType.Call(NewRootFrame(), wrapArgs("foo"), nil))}, wantExc: mustCreateException(SystemErrorType, "module filename missing")},
		{args: wrapArgs(&Module{Object: Object{typ: ModuleType, dict: NewDict()}}), wantExc: mustCreateException(SystemErrorType, "nameless module")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestModuleInit(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, args ...*Object) (*Tuple, *BaseException) {
		o, raised := ModuleType.Call(f, args, nil)
		if raised != nil {
			return nil, raised
		}
		name, raised := GetAttr(f, o, internedName, None)
		if raised != nil {
			return nil, raised
		}
		doc, raised := GetAttr(f, o, NewStr("__doc__"), None)
		if raised != nil {
			return nil, raised
		}
		return NewTuple(name, doc), nil
	})
	cases := []invokeTestCase{
		{args: wrapArgs("foo"), want: newTestTuple("foo", None).ToObject()},
		{args: wrapArgs("foo", 123), want: newTestTuple("foo", 123).ToObject()},
		{args: wrapArgs(newObject(ObjectType)), wantExc: mustCreateException(TypeErrorType, `'__init__' requires a 'str' object but received a "object"`)},
		{wantExc: mustCreateException(TypeErrorType, "'__init__' requires 2 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestModuleStrRepr(t *testing.T) {
	cases := []invokeTestCase{
		{args: wrapArgs(newModule("foo", "<test>")), want: NewStr("<module 'foo' from '<test>'>").ToObject()},
		{args: wrapArgs(newModule("foo.bar.baz", "<test>")), want: NewStr("<module 'foo.bar.baz' from '<test>'>").ToObject()},
		{args: Args{mustNotRaise(ModuleType.Call(NewRootFrame(), wrapArgs("foo"), nil))}, want: NewStr("<module 'foo' (built-in)>").ToObject()},
		{args: wrapArgs(&Module{Object: Object{typ: ModuleType, dict: newTestDict("__file__", "foo.py")}}), want: NewStr("<module '?' from 'foo.py'>").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(wrapFuncForTest(ToStr), &cas); err != "" {
			t.Error(err)
		}
		if err := runInvokeTestCase(wrapFuncForTest(Repr), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestRunMain(t *testing.T) {
	oldSysModules := SysModules
	defer func() {
		SysModules = oldSysModules
	}()
	cases := []struct {
		code       *Code
		wantCode   int
		wantOutput string
	}{
		{NewCode("<test>", "test.py", nil, 0, func(*Frame, []*Object) (*Object, *BaseException) { return None, nil }), 0, ""},
		{NewCode("<test>", "test.py", nil, 0, func(f *Frame, _ []*Object) (*Object, *BaseException) {
			return nil, f.Raise(SystemExitType.ToObject(), None, nil)
		}), 0, ""},
		{NewCode("<test>", "test.py", nil, 0, func(f *Frame, _ []*Object) (*Object, *BaseException) { return nil, f.RaiseType(TypeErrorType, "foo") }), 1, "TypeError: foo\n"},
		{NewCode("<test>", "test.py", nil, 0, func(f *Frame, _ []*Object) (*Object, *BaseException) { return nil, f.RaiseType(SystemExitType, "foo") }), 1, "foo\n"},
		{NewCode("<test>", "test.py", nil, 0, func(f *Frame, _ []*Object) (*Object, *BaseException) {
			return nil, f.Raise(SystemExitType.ToObject(), NewInt(12).ToObject(), nil)
		}), 12, ""},
	}
	for _, cas := range cases {
		SysModules = NewDict()
		if gotCode, gotOutput, err := runMainAndCaptureStderr(cas.code); err != nil {
			t.Errorf("runMainRedirectStderr() failed: %v", err)
		} else if gotCode != cas.wantCode {
			t.Errorf("RunMain() = %v, want %v", gotCode, cas.wantCode)
		} else if gotOutput != cas.wantOutput {
			t.Errorf("RunMain() output %q, want %q", gotOutput, cas.wantOutput)
		}
	}
}

func runMainAndCaptureStderr(code *Code) (int, string, error) {
	oldStderr := Stderr
	defer func() {
		Stderr = oldStderr
	}()
	r, w, err := os.Pipe()
	if err != nil {
		return 0, "", err
	}
	Stderr = NewFileFromFD(w.Fd(), nil)
	c := make(chan int)
	go func() {
		defer w.Close()
		c <- RunMain(code)
	}()
	result := <-c
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return 0, "", err
	}
	return result, string(data), nil
}

var testModuleType *Type

func init() {
	testModuleType, _ = newClass(NewRootFrame(), TypeType, "testModule", []*Type{ModuleType}, newStringDict(map[string]*Object{
		"__eq__": newBuiltinFunction("__eq__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			if raised := checkMethodArgs(f, "__eq__", args, ModuleType, ObjectType); raised != nil {
				return nil, raised
			}
			if !args[1].isInstance(ModuleType) {
				return NotImplemented, nil
			}
			m1, m2 := toModuleUnsafe(args[0]), toModuleUnsafe(args[1])
			name1, raised := m1.GetName(f)
			if raised != nil {
				return nil, raised
			}
			name2, raised := m2.GetName(f)
			if raised != nil {
				return nil, raised
			}
			if name1.Value() != name2.Value() {
				return False.ToObject(), nil
			}
			file1, raised := m1.GetFilename(f)
			if raised != nil {
				return nil, raised
			}
			file2, raised := m2.GetFilename(f)
			if raised != nil {
				return nil, raised
			}
			return GetBool(file1.Value() == file2.Value()).ToObject(), nil
		}).ToObject(),
		"__ne__": newBuiltinFunction("__ne__", func(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
			if raised := checkMethodArgs(f, "__ne__", args, ModuleType, ObjectType); raised != nil {
				return nil, raised
			}
			eq, raised := Eq(f, args[0], args[1])
			if raised != nil {
				return nil, raised
			}
			isEq, raised := IsTrue(f, eq)
			if raised != nil {
				return nil, raised
			}
			return GetBool(!isEq).ToObject(), nil
		}).ToObject(),
	}))
}

func newTestModule(name, filename string) *Module {
	return &Module{Object: Object{typ: testModuleType, dict: newTestDict("__name__", name, "__file__", filename)}}
}

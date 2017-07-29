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
	"os"
	"reflect"
	"runtime/pprof"
	"strings"
	"sync"
)

type moduleState int

const (
	moduleStateNew moduleState = iota
	moduleStateInitializing
	moduleStateReady
)

var (
	importMutex    sync.Mutex
	moduleRegistry = map[string]*Code{}
	// ModuleType is the object representing the Python 'module' type.
	ModuleType = newBasisType("module", reflect.TypeOf(Module{}), toModuleUnsafe, ObjectType)
	// SysModules is the global dict of imported modules, aka sys.modules.
	SysModules = NewDict()
)

// Module represents Python 'module' objects.
type Module struct {
	Object
	mutex recursiveMutex
	state moduleState
	code  *Code
}

// ModuleInit functions are called when importing Grumpy modules to execute the
// top level code for that module.
type ModuleInit func(f *Frame, m *Module) *BaseException

// RegisterModule adds the named module to the registry so that it can be
// subsequently imported.
func RegisterModule(name string, c *Code) {
	err := ""
	importMutex.Lock()
	if moduleRegistry[name] == nil {
		moduleRegistry[name] = c
	} else {
		err = "module already registered: " + name
	}
	importMutex.Unlock()
	if err != "" {
		logFatal(err)
	}
}

// ImportModule takes a fully qualified module name (e.g. a.b.c) and a slice of
// code objects where the name of the i'th module is the prefix of name
// ending in the i'th dot. The number of dot delimited parts of name must be the
// same as the number of code objects. For each successive prefix, ImportModule
// looks in sys.modules for an existing module with that name and if not
// present creates a new module object, adds it to sys.modules and initializes
// it with the corresponding code object. If the module was already present in
// sys.modules, it is not re-initialized. The returned slice contains each
// package and module initialized in this way in order.
//
// For example, ImportModule(f, "a.b", []*Code{a.Code, b.Code})
// causes the initialization and entry into sys.modules of Grumpy module a and
// then Grumpy module b. The two initialized modules are returned.
//
// If ImportModule is called in two threads concurrently to import the same
// module, both invocations will produce the same module object and the module
// is guaranteed to only be initialized once. The second invocation will not
// return the module until it is fully initialized.
func ImportModule(f *Frame, name string) ([]*Object, *BaseException) {
	if strings.Contains(name, "/") {
		o, raised := importOne(f, name)
		if raised != nil {
			return nil, raised
		}
		return []*Object{o}, nil
	}
	parts := strings.Split(name, ".")
	numParts := len(parts)
	result := make([]*Object, numParts)
	var prev *Object
	for i := 0; i < numParts; i++ {
		name := strings.Join(parts[:i+1], ".")
		o, raised := importOne(f, name)
		if raised != nil {
			return nil, raised
		}
		if prev != nil {
			if raised := SetAttr(f, prev, NewStr(parts[i]), o); raised != nil {
				return nil, raised
			}
		}
		result[i] = o
		prev = o
	}
	return result, nil
}

func importOne(f *Frame, name string) (*Object, *BaseException) {
	var c *Code
	// We do very limited locking here resulting in some
	// sys.modules consistency gotchas.
	importMutex.Lock()
	o, raised := SysModules.GetItemString(f, name)
	if raised == nil && o == nil {
		if c = moduleRegistry[name]; c == nil {
			raised = f.RaiseType(ImportErrorType, name)
		} else {
			o = newModule(name, c.filename).ToObject()
			raised = SysModules.SetItemString(f, name, o)
		}
	}
	importMutex.Unlock()
	if raised != nil {
		return nil, raised
	}
	if o.isInstance(ModuleType) {
		var raised *BaseException
		m := toModuleUnsafe(o)
		m.mutex.Lock(f)
		if m.state == moduleStateNew {
			m.state = moduleStateInitializing
			if _, raised = c.Eval(f, m.Dict(), nil, nil); raised == nil {
				m.state = moduleStateReady
			} else {
				// If the module failed to initialize
				// then before we relinquish the module
				// lock, remove it from sys.modules.
				// Threads waiting on this module will
				// fail when they don't find it in
				// sys.modules below.
				e, tb := f.ExcInfo()
				if _, raised := SysModules.DelItemString(f, name); raised != nil {
					f.RestoreExc(e, tb)
				}
			}
		}
		m.mutex.Unlock(f)
		if raised != nil {
			return nil, raised
		}
		// The result should be what's in sys.modules, not
		// necessarily the originally created module since this
		// is CPython's behavior.
		o, raised = SysModules.GetItemString(f, name)
		if raised != nil {
			return nil, raised
		}
		if o == nil {
			// This can happen in the pathological case
			// where the module clears itself from
			// sys.modules during execution and is handled
			// by CPython in PyImport_ExecCodeModuleEx in
			// import.c.
			format := "Loaded module %s not found in sys.modules"
			return nil, f.RaiseType(ImportErrorType, fmt.Sprintf(format, name))
		}
	}
	return o, nil
}

// newModule creates a new Module object with the given fully qualified name
// (e.g a.b.c) and its corresponding Python filename.
func newModule(name, filename string) *Module {
	d := newStringDict(map[string]*Object{
		"__file__": NewStr(filename).ToObject(),
		"__name__": NewStr(name).ToObject(),
	})
	return &Module{Object: Object{typ: ModuleType, dict: d}}
}

func toModuleUnsafe(o *Object) *Module {
	return (*Module)(o.toPointer())
}

// GetFilename returns the __file__ attribute of m, raising SystemError if it
// does not exist.
func (m *Module) GetFilename(f *Frame) (*Str, *BaseException) {
	fileAttr, raised := GetAttr(f, m.ToObject(), NewStr("__file__"), None)
	if raised != nil {
		return nil, raised
	}
	if !fileAttr.isInstance(StrType) {
		return nil, f.RaiseType(SystemErrorType, "module filename missing")
	}
	return toStrUnsafe(fileAttr), nil
}

// GetName returns the __name__ attribute of m, raising SystemError if it does
// not exist.
func (m *Module) GetName(f *Frame) (*Str, *BaseException) {
	nameAttr, raised := GetAttr(f, m.ToObject(), internedName, None)
	if raised != nil {
		return nil, raised
	}
	if !nameAttr.isInstance(StrType) {
		return nil, f.RaiseType(SystemErrorType, "nameless module")
	}
	return toStrUnsafe(nameAttr), nil
}

// ToObject upcasts m to an Object.
func (m *Module) ToObject() *Object {
	return &m.Object
}

func moduleInit(f *Frame, o *Object, args Args, _ KWArgs) (*Object, *BaseException) {
	expectedTypes := []*Type{StrType, ObjectType}
	argc := len(args)
	if argc == 1 {
		expectedTypes = expectedTypes[:1]
	}
	if raised := checkFunctionArgs(f, "__init__", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	if raised := SetAttr(f, o, internedName, args[0]); raised != nil {
		return nil, raised
	}
	if argc > 1 {
		if raised := SetAttr(f, o, NewStr("__doc__"), args[1]); raised != nil {
			return nil, raised
		}
	}
	return None, nil
}

func moduleRepr(f *Frame, o *Object) (*Object, *BaseException) {
	m := toModuleUnsafe(o)
	name := "?"
	nameAttr, raised := m.GetName(f)
	if raised == nil {
		name = nameAttr.Value()
	} else {
		f.RestoreExc(nil, nil)
	}
	file := "(built-in)"
	fileAttr, raised := m.GetFilename(f)
	if raised == nil {
		file = fmt.Sprintf("from '%s'", fileAttr.Value())
	} else {
		f.RestoreExc(nil, nil)
	}
	return NewStr(fmt.Sprintf("<module '%s' %s>", name, file)).ToObject(), nil
}

func initModuleType(map[string]*Object) {
	ModuleType.slots.Init = &initSlot{moduleInit}
	ModuleType.slots.Repr = &unaryOpSlot{moduleRepr}
}

// RunMain execs the given code object as a module under the name "__main__".
// It handles any exceptions raised during module execution. If no exceptions
// were raised then the return value is zero. If a SystemExit was raised then
// the return value depends on its code attribute: None -> zero, integer values
// are returned as-is. Other code values and exception types produce a return
// value of 1.
func RunMain(code *Code) int {
	if file := os.Getenv("GRUMPY_PROFILE"); file != "" {
		f, err := os.Create(file)
		if err != nil {
			logFatal(err.Error())
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			logFatal(err.Error())
		}
		defer pprof.StopCPUProfile()
	}
	m := newModule("__main__", code.filename)
	m.state = moduleStateInitializing
	f := NewRootFrame()
	f.code = code
	f.globals = m.Dict()
	if raised := SysModules.SetItemString(f, "__main__", m.ToObject()); raised != nil {
		Stderr.writeString(raised.String())
	}
	_, e := code.fn(f, nil)
	if e == nil {
		return 0
	}
	if !e.isInstance(SystemExitType) {
		Stderr.writeString(FormatExc(f))
		return 1
	}
	f.RestoreExc(nil, nil)
	o, raised := GetAttr(f, e.ToObject(), NewStr("code"), nil)
	if raised != nil {
		return 1
	}
	if o.isInstance(IntType) {
		return toIntUnsafe(o).Value()
	}
	if o == None {
		return 0
	}
	if s, raised := ToStr(f, o); raised == nil {
		Stderr.writeString(s.Value() + "\n")
	}
	return 1
}

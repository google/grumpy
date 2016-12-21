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
	"io/ioutil"
	"os"
	"reflect"
	"sync"
)

// File represents Python 'file' objects.
type File struct {
	Object
	// mutex synchronizes the state of the File struct, not access to the
	// underlying os.File. So, for example, when doing file reads and
	// writes we only acquire a read lock.
	mutex sync.RWMutex
	mode  string
	open  bool
	file  *os.File
}

// NewFileFromFD creates a file object from the given file descriptor fd.
func NewFileFromFD(fd uintptr) *File {
	// TODO: Use fcntl or something to get the mode of the descriptor.
	return &File{Object: Object{typ: FileType}, mode: "?", open: true, file: os.NewFile(fd, "<fdopen>")}
}

func toFileUnsafe(o *Object) *File {
	return (*File)(o.toPointer())
}

// ToObject upcasts f to an Object.
func (f *File) ToObject() *Object {
	return &f.Object
}

// FileType is the object representing the Python 'file' type.
var FileType = newBasisType("file", reflect.TypeOf(File{}), toFileUnsafe, ObjectType)

func fileInit(f *Frame, o *Object, args Args, _ KWArgs) (*Object, *BaseException) {
	argc := len(args)
	expectedTypes := []*Type{StrType, StrType}
	if argc == 1 {
		expectedTypes = expectedTypes[:1]
	}
	if raised := checkFunctionArgs(f, "__init__", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	mode := "r"
	if argc > 1 {
		mode = toStrUnsafe(args[1]).Value()
	}
	// TODO: Do something with the binary mode flag.
	var flag int
	switch mode {
	case "a", "ab":
		flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	case "r", "rb":
		flag = os.O_RDONLY
	case "r+", "r+b":
		flag = os.O_RDWR
	case "w", "wb":
		flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	default:
		return nil, f.RaiseType(ValueErrorType, "invalid mode string")
	}
	file := toFileUnsafe(o)
	file.mutex.Lock()
	defer file.mutex.Unlock()
	osFile, err := os.OpenFile(toStrUnsafe(args[0]).Value(), flag, 0644)
	if err != nil {
		return nil, f.RaiseType(IOErrorType, err.Error())
	}
	file.mode = mode
	file.open = true
	file.file = osFile
	return None, nil
}

func fileClose(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "close", args, FileType); raised != nil {
		return nil, raised
	}
	file := toFileUnsafe(args[0])
	file.mutex.Lock()
	defer file.mutex.Unlock()
	if file.open && file.file != nil {
		if err := file.file.Close(); err != nil {
			return nil, f.RaiseType(IOErrorType, err.Error())
		}
	}
	file.open = false
	return None, nil
}

func fileRead(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	expectedTypes := []*Type{FileType, IntType}
	argc := len(args)
	if argc == 1 {
		expectedTypes = expectedTypes[:1]
	}
	if raised := checkMethodArgs(f, "read", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	file := toFileUnsafe(args[0])
	size := -1
	if argc > 1 {
		size = toIntUnsafe(args[1]).Value()
	}
	file.mutex.RLock()
	defer file.mutex.RUnlock()
	if !file.open {
		return nil, f.RaiseType(ValueErrorType, "I/O operation on closed file")
	}
	var data []byte
	var err error
	if size < 0 {
		data, err = ioutil.ReadAll(file.file)
	} else {
		data = make([]byte, size)
		var n int
		n, err = file.file.Read(data)
		data = data[:n]
	}
	if err != nil {
		return nil, f.RaiseType(IOErrorType, err.Error())
	}
	return NewStr(string(data)).ToObject(), nil
}

func fileRepr(f *Frame, o *Object) (*Object, *BaseException) {
	file := toFileUnsafe(o)
	file.mutex.RLock()
	defer file.mutex.RUnlock()
	var openState string
	if file.open {
		openState = "open"
	} else {
		openState = "closed"
	}
	var name string
	if file.file != nil {
		name = file.file.Name()
	} else {
		name = "<uninitialized file>"
	}
	var mode string
	if file.mode != "" {
		mode = file.mode
	} else {
		mode = "<uninitialized file>"
	}
	return NewStr(fmt.Sprintf("<%s file %q, mode %q at %p>", openState, name, mode, file)).ToObject(), nil
}

func fileWrite(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "write", args, FileType, StrType); raised != nil {
		return nil, raised
	}
	file := toFileUnsafe(args[0])
	file.mutex.RLock()
	defer file.mutex.RUnlock()
	if !file.open {
		return nil, f.RaiseType(ValueErrorType, "I/O operation on closed file")
	}
	if _, err := file.file.Write([]byte(toStrUnsafe(args[1]).Value())); err != nil {
		return nil, f.RaiseType(IOErrorType, err.Error())
	}
	return None, nil
}

func initFileType(dict map[string]*Object) {
	dict["close"] = newBuiltinFunction("close", fileClose).ToObject()
	dict["read"] = newBuiltinFunction("read", fileRead).ToObject()
	dict["write"] = newBuiltinFunction("write", fileWrite).ToObject()
	FileType.slots.Init = &initSlot{fileInit}
	FileType.slots.Repr = &unaryOpSlot{fileRepr}
}

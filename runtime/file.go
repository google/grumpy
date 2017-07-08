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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"sync"
)

// File represents Python 'file' objects.
type File struct {
	Object
	// mutex synchronizes the state of the File struct, not access to the
	// underlying os.File. So, for example, when doing file reads and
	// writes we only acquire a read lock.
	mutex       sync.Mutex
	mode        string
	open        bool
	Softspace   int `attr:"softspace" attr_mode:"rw"`
	reader      *bufio.Reader
	file        *os.File
	skipNextLF  bool
	univNewLine bool
	close       *Object
}

// NewFileFromFD creates a file object from the given file descriptor fd.
func NewFileFromFD(fd uintptr, close *Object) *File {
	// TODO: Use fcntl or something to get the mode of the descriptor.
	file := &File{
		Object: Object{typ: FileType},
		mode:   "?",
		open:   true,
		file:   os.NewFile(fd, "<fdopen>"),
	}
	if close != None {
		file.close = close
	}
	file.reader = bufio.NewReader(file.file)
	return file
}

func toFileUnsafe(o *Object) *File {
	return (*File)(o.toPointer())
}

func (f *File) name() string {
	name := "<uninitialized file>"
	if f.file != nil {
		name = f.file.Name()
	}
	return name
}

// ToObject upcasts f to an Object.
func (f *File) ToObject() *Object {
	return &f.Object
}

func (f *File) readLine(maxBytes int) (string, error) {
	var buf bytes.Buffer
	numBytesRead := 0
	for maxBytes < 0 || numBytesRead < maxBytes {
		b, err := f.reader.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if b == '\r' && f.univNewLine {
			f.skipNextLF = true
			buf.WriteByte('\n')
			break
		} else if b == '\n' {
			if f.skipNextLF {
				f.skipNextLF = false
				continue // Do not increment numBytesRead.
			} else {
				buf.WriteByte(b)
				break
			}
		} else {
			buf.WriteByte(b)
		}
		numBytesRead++
	}
	return buf.String(), nil
}

func (f *File) writeString(s string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if !f.open {
		return io.ErrClosedPipe
	}
	if _, err := f.file.Write([]byte(s)); err != nil {
		return err
	}

	return nil
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
	case "r", "rb", "rU", "U":
		flag = os.O_RDONLY
	case "r+", "r+b":
		flag = os.O_RDWR
	// Difference between r+ and a+ is that a+ automatically creates file.
	case "a+":
		flag = os.O_RDWR | os.O_CREATE | os.O_APPEND
	case "w+":
		flag = os.O_RDWR | os.O_CREATE
	case "w", "wb":
		flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	default:
		return nil, f.RaiseType(ValueErrorType, fmt.Sprintf("invalid mode string: %q", mode))
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
	file.reader = bufio.NewReader(osFile)
	file.univNewLine = strings.HasSuffix(mode, "U")
	return None, nil
}

func fileEnter(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "__enter__", args, FileType); raised != nil {
		return nil, raised
	}
	return args[0], nil
}

func fileExit(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodVarArgs(f, "__exit__", args, FileType); raised != nil {
		return nil, raised
	}
	closeFunc, raised := GetAttr(f, args[0], NewStr("close"), nil)
	if raised != nil {
		return nil, raised
	}
	_, raised = closeFunc.Call(f, nil, nil)
	if raised != nil {
		return nil, raised
	}
	return None, nil
}

func fileClose(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "close", args, FileType); raised != nil {
		return nil, raised
	}
	file := toFileUnsafe(args[0])
	file.mutex.Lock()
	defer file.mutex.Unlock()
	ret := None
	if file.open {
		var raised *BaseException
		if file.close != nil {
			ret, raised = file.close.Call(f, args, nil)
		} else if file.file != nil {
			if err := file.file.Close(); err != nil {
				raised = f.RaiseType(IOErrorType, err.Error())
			}
		}
		if raised != nil {
			return nil, raised
		}
	}
	file.open = false
	return ret, nil
}

func fileClosed(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "closed", args, FileType); raised != nil {
		return nil, raised
	}
	file := toFileUnsafe(args[0])
	file.mutex.Lock()
	c := !file.open
	file.mutex.Unlock()
	return GetBool(c).ToObject(), nil
}

func fileFileno(f *Frame, args Args, _ KWArgs) (ret *Object, raised *BaseException) {
	if raised := checkMethodArgs(f, "fileno", args, FileType); raised != nil {
		return nil, raised
	}
	file := toFileUnsafe(args[0])
	file.mutex.Lock()
	if file.open {
		ret = NewInt(int(file.file.Fd())).ToObject()
	} else {
		raised = f.RaiseType(ValueErrorType, "I/O operation on closed file")
	}
	file.mutex.Unlock()
	return ret, raised
}

func fileGetName(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "_get_name", args, FileType); raised != nil {
		return nil, raised
	}
	file := toFileUnsafe(args[0])
	file.mutex.Lock()
	name := file.name()
	file.mutex.Unlock()
	return NewStr(name).ToObject(), nil
}

func fileIter(f *Frame, o *Object) (*Object, *BaseException) {
	return o, nil
}

func fileNext(f *Frame, o *Object) (ret *Object, raised *BaseException) {
	file := toFileUnsafe(o)
	file.mutex.Lock()
	defer file.mutex.Unlock()
	if !file.open {
		return nil, f.RaiseType(ValueErrorType, "I/O operation on closed file")
	}
	line, err := file.readLine(-1)
	if err != nil {
		return nil, f.RaiseType(IOErrorType, err.Error())
	}
	if line == "" {
		return nil, f.Raise(StopIterationType.ToObject(), nil, nil)
	}
	return NewStr(line).ToObject(), nil
}

func fileRead(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	file, size, raised := fileParseReadArgs(f, "read", args)
	if raised != nil {
		return nil, raised
	}
	file.mutex.Lock()
	defer file.mutex.Unlock()
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
		n, err = file.reader.Read(data)
		data = data[:n]
	}
	if err != nil && err != io.EOF {
		return nil, f.RaiseType(IOErrorType, err.Error())
	}
	return NewStr(string(data)).ToObject(), nil
}

func fileReadLine(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	file, size, raised := fileParseReadArgs(f, "readline", args)
	if raised != nil {
		return nil, raised
	}
	file.mutex.Lock()
	defer file.mutex.Unlock()
	if !file.open {
		return nil, f.RaiseType(ValueErrorType, "I/O operation on closed file")
	}
	line, err := file.readLine(size)
	if err != nil {
		return nil, f.RaiseType(IOErrorType, err.Error())
	}
	return NewStr(line).ToObject(), nil
}

func fileReadLines(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	// NOTE: The size hint behavior here is slightly different than
	// CPython. Here we read no more lines than necessary. In CPython a
	// minimum of 8KB or more will be read.
	file, size, raised := fileParseReadArgs(f, "readlines", args)
	if raised != nil {
		return nil, raised
	}
	file.mutex.Lock()
	defer file.mutex.Unlock()
	if !file.open {
		return nil, f.RaiseType(ValueErrorType, "I/O operation on closed file")
	}
	var lines []*Object
	numBytesRead := 0
	for size < 0 || numBytesRead < size {
		line, err := file.readLine(-1)
		if err != nil {
			return nil, f.RaiseType(IOErrorType, err.Error())
		}
		if line != "" {
			lines = append(lines, NewStr(line).ToObject())
		}
		if !strings.HasSuffix(line, "\n") {
			break
		}
		numBytesRead += len(line)
	}
	return NewList(lines...).ToObject(), nil
}

func fileRepr(f *Frame, o *Object) (*Object, *BaseException) {
	file := toFileUnsafe(o)
	file.mutex.Lock()
	defer file.mutex.Unlock()
	var openState string
	if file.open {
		openState = "open"
	} else {
		openState = "closed"
	}
	var mode string
	if file.mode != "" {
		mode = file.mode
	} else {
		mode = "<uninitialized file>"
	}
	return NewStr(fmt.Sprintf("<%s file %q, mode %q at %p>", openState, file.name(), mode, file)).ToObject(), nil
}

func fileWrite(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "write", args, FileType, StrType); raised != nil {
		return nil, raised
	}
	file := toFileUnsafe(args[0])
	file.mutex.Lock()
	defer file.mutex.Unlock()
	if !file.open {
		return nil, f.RaiseType(ValueErrorType, "I/O operation on closed file")
	}
	if _, err := file.file.Write([]byte(toStrUnsafe(args[1]).Value())); err != nil {
		return nil, f.RaiseType(IOErrorType, err.Error())
	}
	return None, nil
}

func initFileType(dict map[string]*Object) {
	// TODO: Make enter/exit into slots.
	dict["__enter__"] = newBuiltinFunction("__enter__", fileEnter).ToObject()
	dict["__exit__"] = newBuiltinFunction("__exit__", fileExit).ToObject()
	dict["close"] = newBuiltinFunction("close", fileClose).ToObject()
	dict["closed"] = newBuiltinFunction("closed", fileClosed).ToObject()
	dict["fileno"] = newBuiltinFunction("fileno", fileFileno).ToObject()
	dict["name"] = newProperty(newBuiltinFunction("_get_name", fileGetName).ToObject(), nil, nil).ToObject()
	dict["read"] = newBuiltinFunction("read", fileRead).ToObject()
	dict["readline"] = newBuiltinFunction("readline", fileReadLine).ToObject()
	dict["readlines"] = newBuiltinFunction("readlines", fileReadLines).ToObject()
	dict["write"] = newBuiltinFunction("write", fileWrite).ToObject()
	FileType.slots.Init = &initSlot{fileInit}
	FileType.slots.Iter = &unaryOpSlot{fileIter}
	FileType.slots.Next = &unaryOpSlot{fileNext}
	FileType.slots.Repr = &unaryOpSlot{fileRepr}
}

func fileParseReadArgs(f *Frame, method string, args Args) (*File, int, *BaseException) {
	expectedTypes := []*Type{FileType, ObjectType}
	argc := len(args)
	if argc == 1 {
		expectedTypes = expectedTypes[:1]
	}
	if raised := checkMethodArgs(f, method, args, expectedTypes...); raised != nil {
		return nil, 0, raised
	}
	size := -1
	if argc > 1 {
		o, raised := IntType.Call(f, args[1:], nil)
		if raised != nil {
			return nil, 0, raised
		}
		size = toIntUnsafe(o).Value()
	}
	return toFileUnsafe(args[0]), size, nil
}

var (
	// Stdin is an alias for sys.stdin.
	Stdin = NewFileFromFD(os.Stdin.Fd(), nil)
	// Stdout is an alias for sys.stdout.
	Stdout = NewFileFromFD(os.Stdout.Fd(), nil)
	// Stderr is an alias for sys.stderr.
	Stderr = NewFileFromFD(os.Stderr.Fd(), nil)
)

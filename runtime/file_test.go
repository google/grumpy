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
	"regexp"
	"testing"
)

var (
	wantFileOperationRead  = 1
	wantFileOperationWrite = 2
)

func TestFileInit(t *testing.T) {
	tempFilename := mustCreateTempFile("TestFileInit", "blah blah")
	defer os.Remove(tempFilename)
	cases := []invokeTestCase{
		{args: wrapArgs(newObject(FileType), tempFilename), want: None},
		{args: wrapArgs(newObject(FileType)), wantExc: mustCreateException(TypeErrorType, "'__init__' requires 2 arguments")},
		{args: wrapArgs(newObject(FileType), tempFilename, "abc"), wantExc: mustCreateException(ValueErrorType, "invalid mode string")},
		{args: wrapArgs(newObject(FileType), tempFilename, "w+"), wantExc: mustCreateException(ValueErrorType, "invalid mode string")},
		{args: wrapArgs(newObject(FileType), "nonexistent-file"), wantExc: mustCreateException(IOErrorType, "open nonexistent-file: no such file or directory")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(FileType, "__init__", &cas); err != "" {
			t.Error(err)
		}
		if len(cas.args) > 0 && cas.args[0].isInstance(FileType) {
			toFileUnsafe(cas.args[0]).file.Close()
		}
	}
}

func TestFileClose(t *testing.T) {
	tempFilename := mustCreateTempFile("TestFileClose", "foo\nbar")
	defer os.Remove(tempFilename)
	closedFile := mustOpenFile(tempFilename)
	// This puts the file into an invalid state since Grumpy thinks
	// it's open even though the underlying file was closed.
	closedFile.file.Close()
	cases := []invokeTestCase{
		{args: wrapArgs(newObject(FileType)), want: None},
		{args: wrapArgs(mustOpenFile(tempFilename)), want: None},
		{args: wrapArgs(closedFile), wantExc: mustCreateException(IOErrorType, "invalid argument")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(FileType, "close", &cas); err != "" {
			t.Error(err)
		}
		if len(cas.args) > 0 && cas.args[0].isInstance(FileType) {
			toFileUnsafe(cas.args[0]).file.Close()
		}
	}
}

func TestFileRead(t *testing.T) {
	tempFilename := mustCreateTempFile("TestFileRead", "foo\nbar")
	defer os.Remove(tempFilename)
	closedFile := mustOpenFile(tempFilename)
	closedFile.file.Close()
	_, closedFileReadError := closedFile.file.Read(make([]byte, 10))
	cases := []invokeTestCase{
		{args: wrapArgs(mustOpenFile(tempFilename)), want: NewStr("foo\nbar").ToObject()},
		{args: wrapArgs(mustOpenFile(tempFilename), 3), want: NewStr("foo").ToObject()},
		{args: wrapArgs(mustOpenFile(tempFilename), 1000), want: NewStr("foo\nbar").ToObject()},
		{args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, "unbound method read() must be called with file instance as first argument (got nothing instead)")},
		{args: wrapArgs(closedFile), wantExc: mustCreateException(IOErrorType, closedFileReadError.Error())},
		{args: wrapArgs(newObject(FileType)), wantExc: mustCreateException(ValueErrorType, "I/O operation on closed file")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(FileType, "read", &cas); err != "" {
			t.Error(err)
		}
		if len(cas.args) > 0 && cas.args[0].isInstance(FileType) {
			toFileUnsafe(cas.args[0]).file.Close()
		}
	}
}

func TestFileStrRepr(t *testing.T) {
	fun := newBuiltinFunction("TestFileStrRepr", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		if raised := checkFunctionArgs(f, "TestFileStrRepr", args, ObjectType, StrType); raised != nil {
			return nil, raised
		}
		o := args[0]
		if o.isInstance(FileType) {
			defer toFileUnsafe(o).file.Close()
		}
		re := regexp.MustCompile(toStrUnsafe(args[1]).Value())
		s, raised := ToStr(f, o)
		if raised != nil {
			return nil, raised
		}
		Assert(f, GetBool(re.MatchString(s.Value())).ToObject(), nil)
		s, raised = Repr(f, o)
		if raised != nil {
			return nil, raised
		}
		Assert(f, GetBool(re.MatchString(s.Value())).ToObject(), nil)
		return None, nil
	}).ToObject()
	tempFilename := mustCreateTempFile("TestFileStrRepr", "foo\nbar")
	defer os.Remove(tempFilename)
	closedFile := mustOpenFile(tempFilename).ToObject()
	mustNotRaise(fileClose(newFrame(nil), []*Object{closedFile}, nil))
	// Open a file for write.
	args := wrapArgs(tempFilename, "wb")
	writeFile := mustNotRaise(FileType.Call(newFrame(nil), args, nil))
	if !writeFile.isInstance(FileType) {
		t.Fatalf("file%v = %v, want file object", args, writeFile)
	}
	cases := []invokeTestCase{
		{args: wrapArgs(mustOpenFile(tempFilename), `^<open file "[^"]+", mode "r" at \w+>$`), want: None},
		{args: wrapArgs(writeFile, `^<open file "[^"]+", mode "wb" at \w+>$`), want: None},
		{args: wrapArgs(newObject(FileType), `^<closed file "<uninitialized file>", mode "<uninitialized file>" at \w+>$`), want: None},
		{args: wrapArgs(closedFile, `^<closed file "[^"]+", mode "r" at \w+>$`), want: None},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestFileWrite(t *testing.T) {
	fun := newBuiltinFunction("TestFileWrite", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		if raised := checkMethodArgs(f, "TestFileWrite", args, StrType, StrType, StrType); raised != nil {
			return nil, raised
		}
		writeFile, raised := FileType.Call(f, args[:2], nil)
		if raised != nil {
			return nil, raised
		}
		write, raised := GetAttr(f, writeFile, NewStr("write"), nil)
		if raised != nil {
			return nil, raised
		}
		if _, raised := write.Call(f, args[2:], nil); raised != nil {
			return nil, raised
		}
		contents, err := ioutil.ReadFile(toStrUnsafe(args[0]).Value())
		if err != nil {
			return nil, f.RaiseType(RuntimeErrorType, fmt.Sprintf("error reading file: %s", err.Error()))
		}
		return NewStr(string(contents)).ToObject(), nil
	}).ToObject()
	// Create a temporary directory and cd to it.
	dir, err := ioutil.TempDir("", "TestFileWrite")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() failed: %s", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir(%q) failed: %s", dir, err)
	}
	defer os.Chdir(oldWd)
	for _, filename := range []string{"truncate.txt", "readonly.txt", "append.txt", "rplus.txt"} {
		if err := ioutil.WriteFile(filename, []byte(filename), 0644); err != nil {
			t.Fatalf("ioutil.WriteFile(%q) failed: %s", filename, err)
		}
	}
	cases := []invokeTestCase{
		{args: wrapArgs("noexist.txt", "w", "foo\nbar"), want: NewStr("foo\nbar").ToObject()},
		{args: wrapArgs("truncate.txt", "w", "new contents"), want: NewStr("new contents").ToObject()},
		{args: wrapArgs("append.txt", "a", "\nbar"), want: NewStr("append.txt\nbar").ToObject()},
		{args: wrapArgs("rplus.txt", "r+", "fooey"), want: NewStr("fooey.txt").ToObject()},
		{args: wrapArgs("readonly.txt", "r", "foo"), wantExc: mustCreateException(IOErrorType, "write readonly.txt: bad file descriptor")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func mustCreateTempFile(prefix, contents string) string {
	osFile, err := ioutil.TempFile("", prefix)
	if err != nil {
		panic(err)
	}
	if _, err := osFile.WriteString(contents); err != nil {
		panic(err)
	}
	if err := osFile.Close(); err != nil {
		panic(err)
	}
	return osFile.Name()
}

func mustOpenFile(filename string) *File {
	args := wrapArgs(filename)
	o := mustNotRaise(FileType.Call(newFrame(nil), args, nil))
	if o == nil || !o.isInstance(FileType) {
		panic(fmt.Sprintf("file%v = %v, want file object", args, o))
	}
	return toFileUnsafe(o)
}

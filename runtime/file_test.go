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
	f := newTestFile("blah blah")
	defer f.cleanup()
	cases := []invokeTestCase{
		{args: wrapArgs(newObject(FileType), f.path), want: None},
		{args: wrapArgs(newObject(FileType)), wantExc: mustCreateException(TypeErrorType, "'__init__' requires 2 arguments")},
		{args: wrapArgs(newObject(FileType), f.path, "abc"), wantExc: mustCreateException(ValueErrorType, `invalid mode string: "abc"`)},
		{args: wrapArgs(newObject(FileType), "nonexistent-file"), wantExc: mustCreateException(IOErrorType, "open nonexistent-file: no such file or directory")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(FileType, "__init__", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestFileClosed(t *testing.T) {
	f := newTestFile("foo\nbar")
	defer f.cleanup()
	closedFile := f.open("r")
	// This puts the file into an invalid state since Grumpy thinks
	// it's open even though the underlying file was closed.
	closedFile.file.Close()
	cases := []invokeTestCase{
		{args: wrapArgs(newObject(FileType)), want: True.ToObject()},
		{args: wrapArgs(f.open("r")), want: False.ToObject()},
		{args: wrapArgs(closedFile), want: False.ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(FileType, "closed", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestFileCloseExit(t *testing.T) {
	f := newTestFile("foo\nbar")
	defer f.cleanup()
	for _, method := range []string{"close", "__exit__"} {
		closedFile := f.open("r")
		// This puts the file into an invalid state since Grumpy thinks
		// it's open even though the underlying file was closed.
		closedFile.file.Close()
		cases := []invokeTestCase{
			{args: wrapArgs(newObject(FileType)), want: None},
			{args: wrapArgs(f.open("r")), want: None},
			{args: wrapArgs(closedFile), wantExc: mustCreateException(IOErrorType, closedFile.file.Close().Error())},
		}
		for _, cas := range cases {
			if err := runInvokeMethodTestCase(FileType, method, &cas); err != "" {
				t.Error(err)
			}
		}
	}
}

func TestFileGetName(t *testing.T) {
	fun := wrapFuncForTest(func(f *Frame, file *File) (*Object, *BaseException) {
		return GetAttr(f, file.ToObject(), NewStr("name"), nil)
	})
	foo := newTestFile("foo")
	defer foo.cleanup()
	cases := []invokeTestCase{
		{args: wrapArgs(foo.open("r")), want: NewStr(foo.path).ToObject()},
		{args: wrapArgs(newObject(FileType)), want: NewStr("<uninitialized file>").ToObject()},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestFileIter(t *testing.T) {
	files := makeTestFiles()
	defer files.cleanup()
	closedFile := files[0].open("r")
	closedFile.file.Close()
	_, closedFileReadError := closedFile.file.Read(make([]byte, 10))
	cases := []invokeTestCase{
		{args: wrapArgs(files[0].open("r")), want: newTestList("foo").ToObject()},
		{args: wrapArgs(files[0].open("rU")), want: newTestList("foo").ToObject()},
		{args: wrapArgs(files[1].open("r")), want: newTestList("foo\n").ToObject()},
		{args: wrapArgs(files[1].open("rU")), want: newTestList("foo\n").ToObject()},
		{args: wrapArgs(files[2].open("r")), want: newTestList("foo\n", "bar").ToObject()},
		{args: wrapArgs(files[2].open("rU")), want: newTestList("foo\n", "bar").ToObject()},
		{args: wrapArgs(files[3].open("r")), want: newTestList("foo\r\n").ToObject()},
		{args: wrapArgs(files[3].open("rU")), want: newTestList("foo\n").ToObject()},
		{args: wrapArgs(files[4].open("r")), want: newTestList("foo\rbar").ToObject()},
		{args: wrapArgs(files[4].open("rU")), want: newTestList("foo\n", "bar").ToObject()},
		{args: wrapArgs(closedFile), wantExc: mustCreateException(IOErrorType, closedFileReadError.Error())},
		{args: wrapArgs(newObject(FileType)), wantExc: mustCreateException(ValueErrorType, "I/O operation on closed file")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(ListType.ToObject(), &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestFileNext(t *testing.T) {
	files := makeTestFiles()
	defer files.cleanup()
	closedFile := files[0].open("r")
	closedFile.file.Close()
	_, closedFileReadError := closedFile.file.Read(make([]byte, 10))
	cases := []invokeTestCase{
		{args: wrapArgs(files[0].open("r")), want: NewStr("foo").ToObject()},
		{args: wrapArgs(files[0].open("rU")), want: NewStr("foo").ToObject()},
		{args: wrapArgs(files[1].open("r")), want: NewStr("foo\n").ToObject()},
		{args: wrapArgs(files[1].open("rU")), want: NewStr("foo\n").ToObject()},
		{args: wrapArgs(files[2].open("r")), want: NewStr("foo\n").ToObject()},
		{args: wrapArgs(files[2].open("rU")), want: NewStr("foo\n").ToObject()},
		{args: wrapArgs(files[3].open("r")), want: NewStr("foo\r\n").ToObject()},
		{args: wrapArgs(files[3].open("rU")), want: NewStr("foo\n").ToObject()},
		{args: wrapArgs(files[4].open("r")), want: NewStr("foo\rbar").ToObject()},
		{args: wrapArgs(files[4].open("rU")), want: NewStr("foo\n").ToObject()},
		{args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, "unbound method next() must be called with file instance as first argument (got nothing instead)")},
		{args: wrapArgs(closedFile), wantExc: mustCreateException(IOErrorType, closedFileReadError.Error())},
		{args: wrapArgs(newObject(FileType)), wantExc: mustCreateException(ValueErrorType, "I/O operation on closed file")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(FileType, "next", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestFileRead(t *testing.T) {
	f := newTestFile("foo\nbar")
	defer f.cleanup()
	closedFile := f.open("r")
	closedFile.file.Close()
	_, closedFileReadError := closedFile.file.Read(make([]byte, 10))
	cases := []invokeTestCase{
		{args: wrapArgs(f.open("r")), want: NewStr("foo\nbar").ToObject()},
		{args: wrapArgs(f.open("r"), 3), want: NewStr("foo").ToObject()},
		{args: wrapArgs(f.open("r"), 1000), want: NewStr("foo\nbar").ToObject()},
		{args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, "unbound method read() must be called with file instance as first argument (got nothing instead)")},
		{args: wrapArgs(closedFile), wantExc: mustCreateException(IOErrorType, closedFileReadError.Error())},
		{args: wrapArgs(newObject(FileType)), wantExc: mustCreateException(ValueErrorType, "I/O operation on closed file")},
		{args: wrapArgs(newObject(FileType), "abc"), wantExc: mustCreateException(ValueErrorType, "invalid literal for int() with base 10: abc")},
		{args: wrapArgs(newObject(FileType), 123, 456), wantExc: mustCreateException(TypeErrorType, "'read' of 'file' requires 2 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(FileType, "read", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestFileReadLine(t *testing.T) {
	files := makeTestFiles()
	defer files.cleanup()
	closedFile := files[0].open("r")
	closedFile.file.Close()
	_, closedFileReadError := closedFile.file.Read(make([]byte, 10))
	partialReadFile := files[5].open("rU")
	partialReadFile.readLine(-1)
	cases := []invokeTestCase{
		{args: wrapArgs(files[0].open("r")), want: NewStr("foo").ToObject()},
		{args: wrapArgs(files[0].open("rU")), want: NewStr("foo").ToObject()},
		{args: wrapArgs(files[1].open("r")), want: NewStr("foo\n").ToObject()},
		{args: wrapArgs(files[1].open("rU")), want: NewStr("foo\n").ToObject()},
		{args: wrapArgs(files[2].open("r")), want: NewStr("foo\n").ToObject()},
		{args: wrapArgs(files[2].open("rU")), want: NewStr("foo\n").ToObject()},
		{args: wrapArgs(files[2].open("r"), 2), want: NewStr("fo").ToObject()},
		{args: wrapArgs(files[2].open("r"), 3), want: NewStr("foo").ToObject()},
		{args: wrapArgs(files[2].open("r"), 4), want: NewStr("foo\n").ToObject()},
		{args: wrapArgs(files[2].open("r"), 5), want: NewStr("foo\n").ToObject()},
		{args: wrapArgs(files[3].open("r")), want: NewStr("foo\r\n").ToObject()},
		{args: wrapArgs(files[3].open("rU")), want: NewStr("foo\n").ToObject()},
		{args: wrapArgs(files[3].open("rU"), 3), want: NewStr("foo").ToObject()},
		{args: wrapArgs(files[3].open("rU"), 4), want: NewStr("foo\n").ToObject()},
		{args: wrapArgs(files[3].open("rU"), 5), want: NewStr("foo\n").ToObject()},
		{args: wrapArgs(files[4].open("r")), want: NewStr("foo\rbar").ToObject()},
		{args: wrapArgs(files[4].open("rU")), want: NewStr("foo\n").ToObject()},
		// Ensure that reading after a \r\n returns the requested
		// number of bytes when possible. Check that the trailing \n
		// does not count toward the bytes read.
		{args: wrapArgs(partialReadFile, 3), want: NewStr("bar").ToObject()},
		{args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, "unbound method readline() must be called with file instance as first argument (got nothing instead)")},
		{args: wrapArgs(closedFile), wantExc: mustCreateException(IOErrorType, closedFileReadError.Error())},
		{args: wrapArgs(newObject(FileType)), wantExc: mustCreateException(ValueErrorType, "I/O operation on closed file")},
		{args: wrapArgs(newObject(FileType), "abc"), wantExc: mustCreateException(ValueErrorType, "invalid literal for int() with base 10: abc")},
		{args: wrapArgs(newObject(FileType), 123, 456), wantExc: mustCreateException(TypeErrorType, "'readline' of 'file' requires 2 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(FileType, "readline", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestFileReadLines(t *testing.T) {
	files := makeTestFiles()
	defer files.cleanup()
	closedFile := files[0].open("r")
	closedFile.file.Close()
	_, closedFileReadError := closedFile.file.Read(make([]byte, 10))
	partialReadFile := files[5].open("rU")
	partialReadFile.readLine(-1)
	cases := []invokeTestCase{
		{args: wrapArgs(files[0].open("r")), want: newTestList("foo").ToObject()},
		{args: wrapArgs(files[0].open("rU")), want: newTestList("foo").ToObject()},
		{args: wrapArgs(files[1].open("r")), want: newTestList("foo\n").ToObject()},
		{args: wrapArgs(files[1].open("rU")), want: newTestList("foo\n").ToObject()},
		{args: wrapArgs(files[2].open("r")), want: newTestList("foo\n", "bar").ToObject()},
		{args: wrapArgs(files[2].open("rU")), want: newTestList("foo\n", "bar").ToObject()},
		{args: wrapArgs(files[2].open("r"), 2), want: newTestList("foo\n").ToObject()},
		{args: wrapArgs(files[2].open("r"), 3), want: newTestList("foo\n").ToObject()},
		{args: wrapArgs(files[2].open("r"), 4), want: newTestList("foo\n").ToObject()},
		{args: wrapArgs(files[2].open("r"), 5), want: newTestList("foo\n", "bar").ToObject()},
		{args: wrapArgs(files[3].open("r")), want: newTestList("foo\r\n").ToObject()},
		{args: wrapArgs(files[3].open("rU")), want: newTestList("foo\n").ToObject()},
		{args: wrapArgs(files[3].open("rU"), 3), want: newTestList("foo\n").ToObject()},
		{args: wrapArgs(files[3].open("rU"), 4), want: newTestList("foo\n").ToObject()},
		{args: wrapArgs(files[3].open("rU"), 5), want: newTestList("foo\n").ToObject()},
		{args: wrapArgs(files[4].open("r")), want: newTestList("foo\rbar").ToObject()},
		{args: wrapArgs(files[4].open("rU")), want: newTestList("foo\n", "bar").ToObject()},
		// Ensure that reading after a \r\n returns the requested
		// number of bytes when possible. Check that the trailing \n
		// does not count toward the bytes read.
		{args: wrapArgs(partialReadFile, 3), want: newTestList("bar\n").ToObject()},
		{args: wrapArgs(), wantExc: mustCreateException(TypeErrorType, "unbound method readlines() must be called with file instance as first argument (got nothing instead)")},
		{args: wrapArgs(closedFile), wantExc: mustCreateException(IOErrorType, closedFileReadError.Error())},
		{args: wrapArgs(newObject(FileType)), wantExc: mustCreateException(ValueErrorType, "I/O operation on closed file")},
		{args: wrapArgs(newObject(FileType), "abc"), wantExc: mustCreateException(ValueErrorType, "invalid literal for int() with base 10: abc")},
		{args: wrapArgs(newObject(FileType), 123, 456), wantExc: mustCreateException(TypeErrorType, "'readlines' of 'file' requires 2 arguments")},
	}
	for _, cas := range cases {
		if err := runInvokeMethodTestCase(FileType, "readlines", &cas); err != "" {
			t.Error(err)
		}
	}
}

func TestFileStrRepr(t *testing.T) {
	fun := newBuiltinFunction("TestFileStrRepr", func(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
		if raised := checkFunctionArgs(f, "TestFileStrRepr", args, ObjectType, StrType); raised != nil {
			return nil, raised
		}
		o := args[0]
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
	f := newTestFile("foo\nbar")
	defer f.cleanup()
	closedFile := f.open("r").ToObject()
	mustNotRaise(fileClose(NewRootFrame(), []*Object{closedFile}, nil))
	// Open a file for write.
	writeFile := f.open("wb")
	cases := []invokeTestCase{
		{args: wrapArgs(f.open("r"), `^<open file "[^"]+", mode "r" at \w+>$`), want: None},
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
	for _, filename := range []string{"truncate.txt", "readonly.txt", "append.txt", "rplus.txt", "aplus.txt", "wplus.txt"} {
		if err := ioutil.WriteFile(filename, []byte(filename), 0644); err != nil {
			t.Fatalf("ioutil.WriteFile(%q) failed: %s", filename, err)
		}
	}
	cases := []invokeTestCase{
		{args: wrapArgs("noexist.txt", "w", "foo\nbar"), want: NewStr("foo\nbar").ToObject()},
		{args: wrapArgs("truncate.txt", "w", "new contents"), want: NewStr("new contents").ToObject()},
		{args: wrapArgs("append.txt", "a", "\nbar"), want: NewStr("append.txt\nbar").ToObject()},

		{args: wrapArgs("rplus.txt", "r+", "fooey"), want: NewStr("fooey.txt").ToObject()},
		{args: wrapArgs("noexistplus1.txt", "r+", "pooey"), wantExc: mustCreateException(IOErrorType, "open noexistplus1.txt: no such file or directory")},

		{args: wrapArgs("aplus.txt", "a+", "\napper"), want: NewStr("aplus.txt\napper").ToObject()},
		{args: wrapArgs("noexistplus3.txt", "a+", "snappbacktoreality"), want: NewStr("snappbacktoreality").ToObject()},

		{args: wrapArgs("wplus.txt", "w+", "destructo"), want: NewStr("destructo").ToObject()},
		{args: wrapArgs("noexistplus2.txt", "w+", "wapper"), want: NewStr("wapper").ToObject()},

		{args: wrapArgs("readonly.txt", "r", "foo"), wantExc: mustCreateException(IOErrorType, "write readonly.txt: bad file descriptor")},
	}
	for _, cas := range cases {
		if err := runInvokeTestCase(fun, &cas); err != "" {
			t.Error(err)
		}
	}
}

type testFile struct {
	path  string
	files []*File
}

func newTestFile(contents string) *testFile {
	osFile, err := ioutil.TempFile("", "")
	if err != nil {
		panic(err)
	}
	if _, err := osFile.WriteString(contents); err != nil {
		panic(err)
	}
	if err := osFile.Close(); err != nil {
		panic(err)
	}
	return &testFile{path: osFile.Name()}
}

func (f *testFile) cleanup() {
	for _, file := range f.files {
		file.file.Close()
	}
	os.Remove(f.path)
}

func (f *testFile) open(mode string) *File {
	args := wrapArgs(f.path, mode)
	o := mustNotRaise(FileType.Call(NewRootFrame(), args, nil))
	if o == nil || !o.isInstance(FileType) {
		panic(fmt.Sprintf("file%v = %v, want file object", args, o))
	}
	file := toFileUnsafe(o)
	f.files = append(f.files, file)
	return file
}

type testFileSlice []*testFile

func makeTestFiles() testFileSlice {
	return []*testFile{
		newTestFile("foo"),
		newTestFile("foo\n"),
		newTestFile("foo\nbar"),
		newTestFile("foo\r\n"),
		newTestFile("foo\rbar"),
		newTestFile("foo\r\nbar\r\nbaz"),
	}
}

func (files testFileSlice) cleanup() {
	for _, f := range files {
		f.cleanup()
	}
}

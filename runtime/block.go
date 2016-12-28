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

// Block is a handle to code that runs in a new scope such as a function, class
// or module.
type Block struct {
	// name is the name of the compiled function or class, or "<module>".
	name string
	// filename is the path of the file where the Python code originated.
	filename string
	// fn is a closure that executes the body of the code block. It may be
	// re-entered multiple times, e.g. for exception handling.
	fn func(*Frame, *Object) (*Object, *BaseException)
}

// NewBlock creates a Block object.
func NewBlock(name, filename string, fn func(*Frame, *Object) (*Object, *BaseException)) *Block {
	return &Block{name, filename, fn}
}

// Exec runs b in the context of a new child frame of back.
func (b *Block) Exec(f *Frame, globals *Dict) (*Object, *BaseException) {
	return b.execInternal(f, nil)
}

func (b *Block) execInternal(f *Frame, sendValue *Object) (*Object, *BaseException) {
	oldExc, oldTraceback := f.ExcInfo()
	// Re-enter function body while we have checkpoint handlers left.
	for {
		ret, raised := b.fn(f, sendValue)
		if raised == nil {
			// Restore exc_info to what it was when we left
			// the previous frame.
			f.RestoreExc(oldExc, oldTraceback)
			return ret, nil
		}
		if len(f.checkpoints) == 0 {
			return nil, raised
		}
		f.state = f.PopCheckpoint()
	}
}

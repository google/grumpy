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
	"testing"
)

func TestRecursiveMutex(t *testing.T) {
	var m recursiveMutex
	f := NewRootFrame()
	m.Lock(f)
	m.Lock(f)
	m.Unlock(f)
	m.Unlock(f)
}

func TestRecursiveMutexUnlockedTooManyTimes(t *testing.T) {
	var m recursiveMutex
	f := NewRootFrame()
	m.Lock(f)
	m.Unlock(f)
	oldLogFatal := logFatal
	logFatal = func(msg string) { panic(msg) }
	defer func() {
		logFatal = oldLogFatal
		if e := recover(); e == nil {
			t.Error("Unlock didn't call logFatal")
		}
	}()
	m.Unlock(f)
}

func TestRecursiveMutexUnlockFrameMismatch(t *testing.T) {
	var m recursiveMutex
	m.Lock(NewRootFrame())
	oldLogFatal := logFatal
	logFatal = func(msg string) { panic(msg) }
	defer func() {
		logFatal = oldLogFatal
		if e := recover(); e == nil {
			t.Error("Unlock didn't call logFatal")
		}
	}()
	m.Unlock(NewRootFrame())
}

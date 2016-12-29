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
	"reflect"
	"sync"
)

var (
	// GeneratorType is the object representing the Python 'generator' type.
	GeneratorType = newBasisType("generator", reflect.TypeOf(Generator{}), toGeneratorUnsafe, ObjectType)
)

type generatorState int

const (
	generatorStateCreated generatorState = iota
	generatorStateReady
	generatorStateRunning
	generatorStateDone
)

// Generator represents Python 'generator' objects.
type Generator struct {
	Object
	mutex sync.Mutex
	state generatorState
	block *Block
	frame *Frame
}

// NewGenerator returns a new Generator object that runs the given Block b.
func NewGenerator(b *Block, globals *Dict) *Generator {
	f := newFrame(nil)
	f.globals = globals
	return &Generator{Object: Object{typ: GeneratorType}, block: b, frame: f}
}

func toGeneratorUnsafe(o *Object) *Generator {
	return (*Generator)(o.toPointer())
}

func (g *Generator) resume(f *Frame, sendValue *Object) (*Object, *BaseException) {
	var raised *BaseException
	g.mutex.Lock()
	oldState := g.state
	switch oldState {
	case generatorStateCreated:
		if sendValue != None {
			raised = f.RaiseType(TypeErrorType, "can't send non-None value to a just-started generator")
		} else {
			g.state = generatorStateRunning
		}
	case generatorStateReady:
		g.state = generatorStateRunning
	case generatorStateRunning:
		raised = f.RaiseType(ValueErrorType, "generator already executing")
	case generatorStateDone:
		raised = f.Raise(StopIterationType.ToObject(), nil, nil)
	}
	g.mutex.Unlock()
	// Concurrent attempts to transition to running state will raise here
	// so it's guaranteed that only one thread will proceed to execute the
	// block below.
	if raised != nil {
		return nil, raised
	}
	g.frame.pushFrame(f)
	// Yield statements push a checkpoint but upon first entry into the
	// generator, the stack will be empty so don't pop in that case.
	if oldState != generatorStateCreated {
		g.frame.state = g.frame.PopCheckpoint()
	}
	result, raised := g.block.execInternal(g.frame, sendValue)
	g.mutex.Lock()
	if result == nil && raised == nil {
		raised = f.Raise(StopIterationType.ToObject(), nil, nil)
	}
	if raised == nil {
		g.state = generatorStateReady
	} else {
		g.state = generatorStateDone
	}
	g.mutex.Unlock()
	return result, raised
}

// ToObject upcasts g to an Object.
func (g *Generator) ToObject() *Object {
	return &g.Object
}

func generatorIter(f *Frame, o *Object) (*Object, *BaseException) {
	return o, nil
}

func generatorNext(f *Frame, o *Object) (*Object, *BaseException) {
	return toGeneratorUnsafe(o).resume(f, None)
}

func generatorSend(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "send", args, GeneratorType, ObjectType); raised != nil {
		return nil, raised
	}
	return toGeneratorUnsafe(args[0]).resume(f, args[1])
}

func initGeneratorType(dict map[string]*Object) {
	dict["send"] = newBuiltinFunction("send", generatorSend).ToObject()
	GeneratorType.flags &= ^(typeFlagBasetype | typeFlagInstantiable)
	GeneratorType.slots.Iter = &unaryOpSlot{generatorIter}
	GeneratorType.slots.Next = &unaryOpSlot{generatorNext}
}

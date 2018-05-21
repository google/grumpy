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
)

// Param describes a parameter to a Python function.
type Param struct {
	// Name is the argument name.
	Name string
	// Def is the default value to use if the argument is not provided. If
	// no default is specified then Def is nil.
	Def *Object
}

// ParamSpec describes a Python function's parameters.
type ParamSpec struct {
	Count       int
	name        string
	minArgs     int
	varArgIndex int
	kwArgIndex  int
	params      []Param
}

// NewParamSpec returns a new ParamSpec that accepts the given positional
// parameters and optional vararg and/or kwarg parameter.
func NewParamSpec(name string, params []Param, varArg bool, kwArg bool) *ParamSpec {
	s := &ParamSpec{name: name, params: params, varArgIndex: -1, kwArgIndex: -1}
	numParams := len(params)
	for ; s.minArgs < numParams; s.minArgs++ {
		if params[s.minArgs].Def != nil {
			break
		}
	}
	for _, p := range params[s.minArgs:numParams] {
		if p.Def == nil {
			format := "%s() non-keyword arg %s after keyword arg"
			logFatal(fmt.Sprintf(format, name, p.Name))
		}
	}
	s.Count = numParams
	if varArg {
		s.varArgIndex = s.Count
		s.Count++
	}
	if kwArg {
		s.kwArgIndex = s.Count
		s.Count++
	}
	return s
}

// Validate ensures that a the args and kwargs passed are valid arguments for
// the param spec s. The validated parameters are output to the validated slice
// which must have len s.Count.
func (s *ParamSpec) Validate(f *Frame, validated []*Object, args Args, kwargs KWArgs) *BaseException {
	if nv := len(validated); nv != s.Count {
		format := "%s(): validated slice was incorrect size: %d, want %d"
		return f.RaiseType(SystemErrorType, fmt.Sprintf(format, s.name, nv, s.Count))
	}
	numParams := len(s.params)
	argc := len(args)
	if argc > numParams && s.varArgIndex == -1 {
		format := "%s() takes %d arguments (%d given)"
		return f.RaiseType(TypeErrorType, fmt.Sprintf(format, s.name, numParams, argc))
	}
	i := 0
	for ; i < argc && i < numParams; i++ {
		validated[i] = args[i]
	}
	if s.varArgIndex != -1 {
		validated[s.varArgIndex] = NewTuple(args[i:].makeCopy()...).ToObject()
	}
	var kwargDict *Dict
	if s.kwArgIndex != -1 {
		kwargDict = NewDict()
		validated[s.kwArgIndex] = kwargDict.ToObject()
	}
	for _, kw := range kwargs {
		name := kw.Name
		j := 0
		for ; j < numParams; j++ {
			if s.params[j].Name == name {
				if validated[j] != nil {
					format := "%s() got multiple values for keyword argument '%s'"
					return f.RaiseType(TypeErrorType, fmt.Sprintf(format, s.name, name))
				}
				validated[j] = kw.Value
				break
			}
		}
		if j == numParams {
			if kwargDict == nil {
				format := "%s() got an unexpected keyword argument '%s'"
				return f.RaiseType(TypeErrorType, fmt.Sprintf(format, s.name, name))
			}
			if raised := kwargDict.SetItemString(f, name, kw.Value); raised != nil {
				return raised
			}
		}
	}
	for ; i < numParams; i++ {
		p := s.params[i]
		if validated[i] == nil {
			if p.Def == nil {
				format := "%s() takes at least %d arguments (%d given)"
				return f.RaiseType(TypeErrorType, fmt.Sprintf(format, s.name, s.minArgs, argc))
			}
			validated[i] = p.Def
		}
	}
	return nil
}

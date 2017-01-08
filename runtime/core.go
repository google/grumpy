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
	"log"
	"os"
	"reflect"
)

var logFatal = func(msg string) { log.Fatal(msg) }

// Abs returns the result of o.__abs__ and is equivalent to the Python
// expression "abs(o)".
func Abs(f *Frame, o *Object) (*Object, *BaseException) {
	abs := o.typ.slots.Abs
	if abs == nil {
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf("bad operand type for abs(): '%s'", o.typ.Name()))
	}
	return abs.Fn(f, o)
}

// Add returns the result of adding v and w together according to the
// __add/radd__ operator.
func Add(f *Frame, v, w *Object) (*Object, *BaseException) {
	return binaryOp(f, v, w, v.typ.slots.Add, v.typ.slots.RAdd, w.typ.slots.RAdd, "+")
}

// And returns the result of the bitwise and operator v & w according to
// __and/rand__.
func And(f *Frame, v, w *Object) (*Object, *BaseException) {
	return binaryOp(f, v, w, v.typ.slots.And, v.typ.slots.RAnd, w.typ.slots.RAnd, "&")
}

// Assert raises an AssertionError if the given cond does not evaluate to true.
// If msg is not nil, it is converted to a string via ToStr() and passed as args
// to the raised exception.
func Assert(f *Frame, cond *Object, msg *Object) *BaseException {
	result, raised := IsTrue(f, cond)
	if raised == nil && !result {
		if msg == nil {
			raised = f.Raise(AssertionErrorType.ToObject(), nil, nil)
		} else {
			var s *Str
			s, raised = ToStr(f, msg)
			if raised == nil {
				raised = f.RaiseType(AssertionErrorType, s.Value())
			}
		}
	}
	return raised
}

// Contains checks whether value is present in seq. It first checks the
// __contains__ method of seq and, if that is not available, attempts to find
// value by iteration over seq. It is equivalent to the Python expression
// "value in seq".
func Contains(f *Frame, seq, value *Object) (bool, *BaseException) {
	if contains := seq.typ.slots.Contains; contains != nil {
		ret, raised := contains.Fn(f, seq, value)
		if raised != nil {
			return false, raised
		}
		return IsTrue(f, ret)
	}
	iter, raised := Iter(f, seq)
	if raised != nil {
		return false, raised
	}
	o, raised := Next(f, iter)
	for ; raised == nil; o, raised = Next(f, iter) {
		eq, raised := Eq(f, o, value)
		if raised != nil {
			return false, raised
		}
		if ret, raised := IsTrue(f, eq); raised != nil {
			return false, raised
		} else if ret {
			return true, nil
		}
	}
	if !raised.isInstance(StopIterationType) {
		return false, raised
	}
	f.RestoreExc(nil, nil)
	return false, nil
}

// DelAttr removes the attribute of o given by name. Equivalent to the Python
// expression delattr(o, name).
func DelAttr(f *Frame, o *Object, name *Str) *BaseException {
	delAttr := o.typ.slots.DelAttr
	if delAttr == nil {
		return f.RaiseType(SystemErrorType, fmt.Sprintf("'%s' object has no __delattr__ method", o.typ.Name()))
	}
	return delAttr.Fn(f, o, name)
}

// DelVar removes the named variable from the given namespace dictionary such
// as a module globals dict.
func DelVar(f *Frame, namespace *Dict, name *Str) *BaseException {
	deleted, raised := namespace.DelItem(f, name.ToObject())
	if raised != nil {
		return raised
	}
	if !deleted {
		return f.RaiseType(NameErrorType, fmt.Sprintf("name '%s' is not defined", name.Value()))
	}
	return nil
}

// DelItem performs the operation del o[key].
func DelItem(f *Frame, o, key *Object) *BaseException {
	delItem := o.typ.slots.DelItem
	if delItem == nil {
		return f.RaiseType(TypeErrorType, fmt.Sprintf("'%s' object does not support item deletion", o.typ.Name()))
	}
	return delItem.Fn(f, o, key)
}

// Div returns the result of dividing v by w according to the __div/rdiv__
// operator.
func Div(f *Frame, v, w *Object) (*Object, *BaseException) {
	return binaryOp(f, v, w, v.typ.slots.Div, v.typ.slots.RDiv, w.typ.slots.RDiv, "/")
}

// Eq returns the equality of v and w according to the __eq__ operator.
func Eq(f *Frame, v, w *Object) (*Object, *BaseException) {
	r, raised := compareRich(f, compareOpEq, v, w)
	if raised != nil {
		return nil, raised
	}
	if r != NotImplemented {
		return r, nil
	}
	return GetBool(compareDefault(f, v, w) == 0).ToObject(), nil
}

// FormatException returns a single-line exception string for the given
// exception object, e.g. "NameError: name 'x' is not defined\n".
func FormatException(f *Frame, e *BaseException) (string, *BaseException) {
	s, raised := ToStr(f, e.ToObject())
	if raised != nil {
		return "", raised
	}
	if len(s.Value()) == 0 {
		return e.typ.Name() + "\n", nil
	}
	return fmt.Sprintf("%s: %s\n", e.typ.Name(), s.Value()), nil
}

// GE returns the result of operation v >= w.
func GE(f *Frame, v, w *Object) (*Object, *BaseException) {
	r, raised := compareRich(f, compareOpGE, v, w)
	if raised != nil {
		return nil, raised
	}
	if r != NotImplemented {
		return r, nil
	}
	return GetBool(compareDefault(f, v, w) >= 0).ToObject(), nil
}

// GetItem returns the result of operation o[key].
func GetItem(f *Frame, o, key *Object) (*Object, *BaseException) {
	getItem := o.typ.slots.GetItem
	if getItem == nil {
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf("'%s' object has no attribute '__getitem__'", o.typ.Name()))
	}
	return getItem.Fn(f, o, key)
}

// GetAttr returns the named attribute of o. Equivalent to the Python expression
// getattr(o, name, def).
func GetAttr(f *Frame, o *Object, name *Str, def *Object) (*Object, *BaseException) {
	// TODO: Fall back to __getattr__.
	getAttribute := o.typ.slots.GetAttribute
	if getAttribute == nil {
		msg := fmt.Sprintf("'%s' has no attribute '%s'", o.typ.Name(), name.Value())
		return nil, f.RaiseType(AttributeErrorType, msg)
	}
	result, raised := getAttribute.Fn(f, o, name)
	if raised != nil && raised.isInstance(AttributeErrorType) && def != nil {
		f.RestoreExc(nil, nil)
		result, raised = def, nil
	}
	return result, raised
}

// GT returns the result of operation v > w.
func GT(f *Frame, v, w *Object) (*Object, *BaseException) {
	r, raised := compareRich(f, compareOpGT, v, w)
	if raised != nil {
		return nil, raised
	}
	if r != NotImplemented {
		return r, nil
	}
	return GetBool(compareDefault(f, v, w) > 0).ToObject(), nil
}

// Hash returns the hash of o according to its __hash__ operator.
func Hash(f *Frame, o *Object) (*Int, *BaseException) {
	hash := o.typ.slots.Hash
	if hash == nil {
		_, raised := hashNotImplemented(f, o)
		return nil, raised
	}
	h, raised := hash.Fn(f, o)
	if raised != nil {
		return nil, raised
	}
	if !h.isInstance(IntType) {
		return nil, f.RaiseType(TypeErrorType, "an integer is required")
	}
	return toIntUnsafe(h), nil
}

// IAdd returns the result of v.__iadd__ if defined, otherwise falls back to
// Add.
func IAdd(f *Frame, v, w *Object) (*Object, *BaseException) {
	return inplaceOp(f, v, w, v.typ.slots.IAdd, Add)
}

// IAnd returns the result of v.__iand__ if defined, otherwise falls back to
// And.
func IAnd(f *Frame, v, w *Object) (*Object, *BaseException) {
	return inplaceOp(f, v, w, v.typ.slots.IAnd, And)
}

// IDiv returns the result of v.__idiv__ if defined, otherwise falls back to
// div.
func IDiv(f *Frame, v, w *Object) (*Object, *BaseException) {
	return inplaceOp(f, v, w, v.typ.slots.IDiv, Div)
}

// IMod returns the result of v.__imod__ if defined, otherwise falls back to
// mod.
func IMod(f *Frame, v, w *Object) (*Object, *BaseException) {
	return inplaceOp(f, v, w, v.typ.slots.IMod, Mod)
}

// IMul returns the result of v.__imul__ if defined, otherwise falls back to
// mul.
func IMul(f *Frame, v, w *Object) (*Object, *BaseException) {
	return inplaceOp(f, v, w, v.typ.slots.IMul, Mul)
}

// Invert returns the result of o.__invert__ and is equivalent to the Python
// expression "~o".
func Invert(f *Frame, o *Object) (*Object, *BaseException) {
	invert := o.typ.slots.Invert
	if invert == nil {
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf("bad operand type for unary ~: '%s'", o.typ.Name()))
	}
	return invert.Fn(f, o)
}

// IOr returns the result of v.__ior__ if defined, otherwise falls back to Or.
func IOr(f *Frame, v, w *Object) (*Object, *BaseException) {
	return inplaceOp(f, v, w, v.typ.slots.IOr, Or)
}

// IsInstance returns true if the type o is an instance of classinfo, or an
// instance of an element in classinfo (if classinfo is a tuple). It returns
// false otherwise. The argument classinfo must be a type or a tuple whose
// elements are types like the isinstance() Python builtin.
func IsInstance(f *Frame, o *Object, classinfo *Object) (bool, *BaseException) {
	return IsSubclass(f, o.typ.ToObject(), classinfo)
}

// IsSubclass returns true if the type o is a subtype of classinfo or a subtype
// of an element in classinfo (if classinfo is a tuple). It returns false
// otherwise. The argument o must be a type and classinfo must be a type or a
// tuple whose elements are types like the issubclass() Python builtin.
func IsSubclass(f *Frame, o *Object, classinfo *Object) (bool, *BaseException) {
	if !o.isInstance(TypeType) {
		return false, f.RaiseType(TypeErrorType, "issubclass() arg 1 must be a class")
	}
	t := toTypeUnsafe(o)
	errorMsg := "classinfo must be a type or tuple of types"
	if classinfo.isInstance(TypeType) {
		return t.isSubclass(toTypeUnsafe(classinfo)), nil
	}
	if !classinfo.isInstance(TupleType) {
		return false, f.RaiseType(TypeErrorType, errorMsg)
	}
	for _, elem := range toTupleUnsafe(classinfo).elems {
		if !elem.isInstance(TypeType) {
			return false, f.RaiseType(TypeErrorType, errorMsg)
		}
		if t.isSubclass(toTypeUnsafe(elem)) {
			return true, nil
		}
	}
	return false, nil
}

// IsTrue returns the truthiness of o according to the __nonzero__ operator.
func IsTrue(f *Frame, o *Object) (bool, *BaseException) {
	switch o {
	case True.ToObject():
		return true, nil
	case False.ToObject(), None:
		return false, nil
	}
	nonzero := o.typ.slots.NonZero
	if nonzero != nil {
		r, raised := nonzero.Fn(f, o)
		if raised != nil {
			return false, raised
		}
		if r.isInstance(IntType) {
			return toIntUnsafe(r).IsTrue(), nil
		}
		msg := fmt.Sprintf("__nonzero__ should return bool, returned %s", r.typ.Name())
		return false, f.RaiseType(TypeErrorType, msg)
	}
	if o.typ.slots.Len != nil {
		l, raised := Len(f, o)
		if raised != nil {
			return false, raised
		}
		return l.IsTrue(), nil
	}
	return true, nil
}

// ISub returns the result of v.__isub__ if defined, otherwise falls back to
// sub.
func ISub(f *Frame, v, w *Object) (*Object, *BaseException) {
	if isub := v.typ.slots.ISub; isub != nil {
		return isub.Fn(f, v, w)
	}
	return Sub(f, v, w)
}

// Iter implements the Python iter() builtin. It returns an iterator for o if
// o is iterable. Otherwise it raises TypeError.
// Note that the iter(f, sentinel) form is not yet supported.
func Iter(f *Frame, o *Object) (*Object, *BaseException) {
	// TODO: Support iter(f, sentinel) usage.
	iter := o.typ.slots.Iter
	if iter != nil {
		return iter.Fn(f, o)
	}
	if o.typ.slots.GetItem != nil {
		return newSeqIterator(o), nil
	}
	return nil, f.RaiseType(TypeErrorType, fmt.Sprintf("'%s' object is not iterable", o.typ.Name()))
}

// IXor returns the result of v.__ixor__ if defined, otherwise falls back to
// Xor.
func IXor(f *Frame, v, w *Object) (*Object, *BaseException) {
	return inplaceOp(f, v, w, v.typ.slots.IXor, Xor)
}

// LE returns the result of operation v <= w.
func LE(f *Frame, v, w *Object) (*Object, *BaseException) {
	r, raised := compareRich(f, compareOpLE, v, w)
	if raised != nil {
		return nil, raised
	}
	if r != NotImplemented {
		return r, nil
	}
	return GetBool(compareDefault(f, v, w) <= 0).ToObject(), nil
}

// Len returns the length of the given sequence object.
func Len(f *Frame, o *Object) (*Int, *BaseException) {
	lenSlot := o.typ.slots.Len
	if lenSlot == nil {
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf("object of type '%s' has no len()", o.typ.Name()))
	}
	r, raised := lenSlot.Fn(f, o)
	if raised != nil {
		return nil, raised
	}
	if !r.isInstance(IntType) {
		return nil, f.RaiseType(TypeErrorType, "an integer is required")
	}
	return toIntUnsafe(r), nil
}

// LShift returns the result of v << w according to the __lshift/rlshift__
// operator.
func LShift(f *Frame, v, w *Object) (*Object, *BaseException) {
	return binaryOp(f, v, w, v.typ.slots.LShift, v.typ.slots.RLShift, w.typ.slots.RLShift, "<<")
}

// LT returns the result of operation v < w.
func LT(f *Frame, v, w *Object) (*Object, *BaseException) {
	r, raised := compareRich(f, compareOpLT, v, w)
	if raised != nil {
		return nil, raised
	}
	if r != NotImplemented {
		return r, nil
	}
	return GetBool(compareDefault(f, v, w) < 0).ToObject(), nil
}

// Mod returns the remainder from the division of v by w according to the
// __mod/rmod__ operator.
func Mod(f *Frame, v, w *Object) (*Object, *BaseException) {
	return binaryOp(f, v, w, v.typ.slots.Mod, v.typ.slots.RMod, w.typ.slots.RMod, "%")
}

// Mul returns the result of multiplying v and w together according to the
// __mul/rmul__ operator.
func Mul(f *Frame, v, w *Object) (*Object, *BaseException) {
	return binaryOp(f, v, w, v.typ.slots.Mul, v.typ.slots.RMul, w.typ.slots.RMul, "*")
}

// Or returns the result of the bitwise or operator v | w according to
// __or/ror__.
func Or(f *Frame, v, w *Object) (*Object, *BaseException) {
	return binaryOp(f, v, w, v.typ.slots.Or, v.typ.slots.ROr, w.typ.slots.ROr, "|")
}

// Index returns the o converted to a Python int or long according to o's
// __index__ slot.
func Index(f *Frame, o *Object) (*Object, *BaseException) {
	if o.isInstance(IntType) || o.isInstance(LongType) {
		return o, nil
	}
	index := o.typ.slots.Index
	if index == nil {
		return nil, nil
	}
	i, raised := index.Fn(f, o)
	if raised != nil {
		return nil, raised
	}
	if !i.isInstance(IntType) && !i.isInstance(LongType) {
		format := "__index__ returned non-(int,long) (type %s)"
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, i.typ.Name()))
	}
	return i, nil
}

// IndexInt returns the value of o converted to a Go int according to o's
// __index__ slot.
// It raises a TypeError if o doesn't have an __index__ method.
// If the returned value doesn't fit into an int, it raises an OverflowError.
func IndexInt(f *Frame, o *Object) (int, *BaseException) {
	i, raised := Index(f, o)
	if raised != nil {
		return 0, raised
	}
	if i.isInstance(IntType) {
		return toIntUnsafe(i).Value(), nil
	}
	if l := toLongUnsafe(i).Value(); numInIntRange(l) {
		return int(l.Int64()), nil
	}
	return 0, f.RaiseType(IndexErrorType, fmt.Sprintf("cannot fit '%s' into an index-sized integer", o.typ))
}

// Invoke calls the given callable with the positional arguments given by args
// and *varargs, and the keyword arguments by keywords and **kwargs. It first
// packs the arguments into slices for the positional and keyword arguments,
// then it passes those to *Object.Call.
func Invoke(f *Frame, callable *Object, args Args, varargs *Object, keywords KWArgs, kwargs *Object) (*Object, *BaseException) {
	if varargs != nil {
		raised := seqApply(f, varargs, func(elems []*Object, _ bool) *BaseException {
			numArgs := len(args)
			packed := make([]*Object, numArgs+len(elems))
			copy(packed, args)
			copy(packed[numArgs:], elems)
			args = packed
			return nil
		})
		if raised != nil {
			return nil, raised
		}
	}
	if kwargs != nil {
		if !kwargs.isInstance(DictType) {
			format := "argument after ** must be a dict, not %s"
			return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(format, kwargs.typ.Name()))
		}
		kwargsDict := toDictUnsafe(kwargs)
		numKeywords := len(keywords)
		numKwargs, raised := Len(f, kwargs)
		if raised != nil {
			return nil, raised
		}
		// Don't bother synchronizing access to len(kwargs) since it's just a
		// hint and it doesn't matter if it's a little off.
		packed := make(KWArgs, numKeywords, numKeywords+numKwargs.Value())
		copy(packed, keywords)
		raised = seqForEach(f, kwargs, func(o *Object) *BaseException {
			if !o.isInstance(StrType) {
				return f.RaiseType(TypeErrorType, "keywords must be strings")
			}
			s := toStrUnsafe(o).Value()
			// Search for dupes linearly assuming small number of keywords.
			for _, kw := range keywords {
				if kw.Name == s {
					format := "got multiple values for keyword argument '%s'"
					return f.RaiseType(TypeErrorType, fmt.Sprintf(format, s))
				}
			}
			item, raised := kwargsDict.GetItem(f, o)
			if raised != nil {
				return raised
			}
			if item == nil {
				return raiseKeyError(f, o)
			}
			packed = append(packed, KWArg{Name: s, Value: item})
			return nil
		})
		if raised != nil {
			return nil, raised
		}
		keywords = packed
	}
	return callable.Call(f, args, keywords)
}

// NE returns the non-equality of v and w according to the __ne__ operator.
func NE(f *Frame, v, w *Object) (*Object, *BaseException) {
	r, raised := compareRich(f, compareOpNE, v, w)
	if raised != nil {
		return nil, raised
	}
	if r != NotImplemented {
		return r, nil
	}
	return GetBool(compareDefault(f, v, w) != 0).ToObject(), nil
}

// Next implements the Python next() builtin. It calls next on the provided
// iterator. It raises TypeError if iter is not an iterator object.
// Note that the next(it, default) form is not yet supported.
func Next(f *Frame, iter *Object) (*Object, *BaseException) {
	// TODO: Support next(it, default) usage.
	next := iter.typ.slots.Next
	if next == nil {
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf("%s object is not an iterator", iter.typ.Name()))
	}
	return next.Fn(f, iter)
}

// Print implements the Python print statement. It calls str() on the given args
// and outputs the results to stdout separated by spaces. Similar to the Python
// print statement.
func Print(f *Frame, args Args, nl bool) *BaseException {
	// TODO: Support outputting to files other than stdout and softspace.
	for i, arg := range args {
		if i > 0 {
			fmt.Print(" ")
		}
		s, raised := ToStr(f, arg)
		if raised != nil {
			return raised
		}
		fmt.Print(s.Value())
	}
	if nl {
		fmt.Println()
	} else if len(args) > 0 {
		fmt.Print(" ")
	}
	return nil
}

// Repr returns a string containing a printable representation of o. This is
// equivalent to the Python expression "repr(o)".
func Repr(f *Frame, o *Object) (*Str, *BaseException) {
	repr := o.typ.slots.Repr
	if repr == nil {
		s, raised := o.typ.FullName(f)
		if raised != nil {
			return nil, raised
		}
		return NewStr(fmt.Sprintf("<%s object at %p>", s, o)), nil
	}
	r, raised := repr.Fn(f, o)
	if raised != nil {
		return nil, raised
	}
	if !r.isInstance(StrType) {
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf("__repr__ returned non-string (type %s)", r.typ.Name()))
	}
	return toStrUnsafe(r), nil
}

// ResolveClass resolves name in the class dict given by class, falling back to
// the provided local if it is non-nil, otherwise falling back to globals.
// This is used by the code generator to resolve names in the context of a class
// definition. If the class definition occurs in a closure in which a local of
// the given name is present then local will be non-nil, otherwise it will be
// nil.
func ResolveClass(f *Frame, class *Dict, local *Object, name *Str) (*Object, *BaseException) {
	if value, raised := class.GetItem(f, name.ToObject()); raised != nil || value != nil {
		return value, raised
	}
	if local != nil {
		if raised := CheckLocal(f, local, name.Value()); raised != nil {
			return nil, raised
		}
		return local, nil
	}
	return ResolveGlobal(f, name)
}

// ResolveGlobal looks up name in the frame's dict of global variables or in
// the Builtins dict if absent. It raises NameError when absent from both.
func ResolveGlobal(f *Frame, name *Str) (*Object, *BaseException) {
	if value, raised := f.Globals().GetItem(f, name.ToObject()); raised != nil || value != nil {
		return value, raised
	}
	value, raised := Builtins.GetItem(f, name.ToObject())
	if raised != nil {
		return nil, raised
	}
	if value == nil {
		return nil, f.RaiseType(NameErrorType, fmt.Sprintf("name '%s' is not defined", name.Value()))
	}
	return value, nil
}

// RShift returns the result of v >> w according to the __rshift/rrshift__
// operator.
func RShift(f *Frame, v, w *Object) (*Object, *BaseException) {
	return binaryOp(f, v, w, v.typ.slots.RShift, v.typ.slots.RRShift, w.typ.slots.RRShift, ">>")
}

// CheckLocal validates that the local variable with the given name and value
// has been bound and raises UnboundLocalError if not.
func CheckLocal(f *Frame, value *Object, name string) *BaseException {
	if value == UnboundLocal {
		format := "local variable '%s' referenced before assignment"
		return f.RaiseType(UnboundLocalErrorType, fmt.Sprintf(format, name))
	}
	return nil
}

// SetAttr sets the attribute of o given by name to value. Equivalent to the
// Python expression setattr(o, name, value).
func SetAttr(f *Frame, o *Object, name *Str, value *Object) *BaseException {
	setAttr := o.typ.slots.SetAttr
	if setAttr == nil {
		return f.RaiseType(SystemErrorType, fmt.Sprintf("'%s' object has no __setattr__ method", o.typ.Name()))
	}
	return setAttr.Fn(f, o, name, value)
}

// SetItem performs the operation o[key] = value.
func SetItem(f *Frame, o, key, value *Object) *BaseException {
	setItem := o.typ.slots.SetItem
	if setItem == nil {
		return f.RaiseType(TypeErrorType, fmt.Sprintf("'%s' object has no attribute '__setitem__'", o.typ.Name()))
	}
	return setItem.Fn(f, o, key, value)
}

// StartThread runs callable in a new goroutine.
func StartThread(callable *Object) {
	go func() {
		f := NewRootFrame()
		_, raised := callable.Call(f, nil, nil)
		if raised != nil {
			s, raised := FormatException(f, raised)
			if raised != nil {
				s = raised.String()
			}
			fmt.Fprintf(os.Stderr, s)
		}
	}()
}

// Sub returns the result of subtracting v from w according to the
// __sub/rsub__ operator.
func Sub(f *Frame, v, w *Object) (*Object, *BaseException) {
	return binaryOp(f, v, w, v.typ.slots.Sub, v.typ.slots.RSub, w.typ.slots.RSub, "-")
}

// TieTarget is a data structure used to facilitate iterator unpacking in
// assignment statements. A TieTarget should have one of Target or Children
// populated but not both.
//
// As an example, the targets in the Python assignment 'foo, bar = ...'
// could be represented as:
//
// 	TieTarget{
// 		Children: []TieTarget{{Target: &foo}, {Target: &bar}},
// 	}
type TieTarget struct {
	// Target is a destination pointer where an unpacked value will be
	// stored.
	Target **Object
	// Children contains a sequence of TieTargets that should be unpacked
	// into.
	Children []TieTarget
}

// Tie takes a (possibly nested) TieTarget and recursively unpacks the
// elements of o by iteration, assigning the results to the Target fields of t.
// If the structure of o is not suitable to be unpacked into t, then an
// exception is raised.
func Tie(f *Frame, t TieTarget, o *Object) *BaseException {
	if t.Target != nil {
		*t.Target = o
		return nil
	}
	iter, raised := Iter(f, o)
	if raised != nil {
		return raised
	}
	for i, child := range t.Children {
		if value, raised := Next(f, iter); raised == nil {
			if raised := Tie(f, child, value); raised != nil {
				return raised
			}
		} else if raised.isInstance(StopIterationType) {
			return f.RaiseType(TypeErrorType, fmt.Sprintf("need more than %d values to unpack", i))
		} else {
			return raised
		}
	}
	_, raised = Next(f, iter)
	if raised == nil {
		return f.RaiseType(ValueErrorType, "too many values to unpack")
	}
	if !raised.isInstance(StopIterationType) {
		return raised
	}
	f.RestoreExc(nil, nil)
	return nil
}

// ToNative converts o to a native Go object according to the __native__
// operator.
func ToNative(f *Frame, o *Object) (reflect.Value, *BaseException) {
	if native := o.typ.slots.Native; native != nil {
		return native.Fn(f, o)
	}
	return reflect.ValueOf(o), nil
}

// ToStr is a convenience function for calling "str(o)".
func ToStr(f *Frame, o *Object) (*Str, *BaseException) {
	result, raised := StrType.Call(f, []*Object{o}, nil)
	if raised != nil {
		return nil, raised
	}
	if !result.isInstance(StrType) {
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf("__str__ returned non-string (type %s)", result.typ.Name()))
	}
	return toStrUnsafe(result), nil
}

// Neg returns the result of o.__neg__ and is equivalent to the Python
// expression "-o".
func Neg(f *Frame, o *Object) (*Object, *BaseException) {
	neg := o.typ.slots.Neg
	if neg == nil {
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf("bad operand type for unary -: '%s'", o.typ.Name()))
	}
	return neg.Fn(f, o)
}

// Xor returns the result of the bitwise xor operator v ^ w according to
// __xor/rxor__.
func Xor(f *Frame, v, w *Object) (*Object, *BaseException) {
	return binaryOp(f, v, w, v.typ.slots.Xor, v.typ.slots.RXor, w.typ.slots.RXor, "^")
}

const (
	errResultTooLarge     = "result too large"
	errUnsupportedOperand = "unsupported operand type(s) for %s: '%s' and '%s'"
)

// binaryOp picks an appropriate operator method (op or rop) from v or w and
// returns its result. It raises TypeError if no appropriate method is found.
// It is similar to CPython's binary_op1 function from abstract.c.
func binaryOp(f *Frame, v, w *Object, op, vrop, wrop *binaryOpSlot, opName string) (*Object, *BaseException) {
	if v.typ != w.typ && w.typ.isSubclass(v.typ) {
		// w is an instance of a subclass of type(v), so prefer w's more
		// specific rop, but only if it is overridden (wrop != vrop).
		if wrop != nil && wrop != vrop {
			r, raised := wrop.Fn(f, w, v)
			if raised != nil {
				return nil, raised
			}
			if r != NotImplemented {
				return r, nil
			}
		}
	}
	if op != nil {
		r, raised := op.Fn(f, v, w)
		if raised != nil {
			return nil, raised
		}
		if r != NotImplemented {
			return r, nil
		}
	}
	if wrop != nil {
		r, raised := wrop.Fn(f, w, v)
		if raised != nil {
			return nil, raised
		}
		if r != NotImplemented {
			return r, nil
		}
	}
	return nil, f.RaiseType(TypeErrorType, fmt.Sprintf(errUnsupportedOperand, opName, v.typ.Name(), w.typ.Name()))
}

func inplaceOp(f *Frame, v, w *Object, slot *binaryOpSlot, fallback binaryOpFunc) (*Object, *BaseException) {
	if slot != nil {
		return slot.Fn(f, v, w)
	}
	return fallback(f, v, w)
}

type compareOp int

const (
	compareOpLT compareOp = iota
	compareOpLE
	compareOpEq
	compareOpNE
	compareOpGE
	compareOpGT
)

var compareOpSwapped = []compareOp{
	compareOpGT,
	compareOpGE,
	compareOpEq,
	compareOpNE,
	compareOpLE,
	compareOpLT,
}

func (op compareOp) swapped() compareOp {
	return compareOpSwapped[op]
}

func (op compareOp) slot(t *Type) *binaryOpSlot {
	switch op {
	case compareOpLT:
		return t.slots.LT
	case compareOpLE:
		return t.slots.LE
	case compareOpEq:
		return t.slots.Eq
	case compareOpNE:
		return t.slots.NE
	case compareOpGE:
		return t.slots.GE
	case compareOpGT:
		return t.slots.GT
	}
	panic(fmt.Sprintf("invalid compareOp value: %d", op))
}

func compareRich(f *Frame, op compareOp, v, w *Object) (*Object, *BaseException) {
	// TODO: Support __cmp__.
	if v.typ != w.typ && w.typ.isSubclass(v.typ) {
		// type(w) is a subclass of type(v) so try to use w's
		// comparison operators since they're more specific.
		slot := op.swapped().slot(w.typ)
		if slot != nil {
			r, raised := slot.Fn(f, w, v)
			if raised != nil {
				return nil, raised
			}
			if r != NotImplemented {
				return r, nil
			}
		}
	}
	slot := op.slot(v.typ)
	if slot != nil {
		r, raised := slot.Fn(f, v, w)
		if raised != nil {
			return nil, raised
		}
		if r != NotImplemented {
			return r, nil
		}
	}
	slot = op.swapped().slot(w.typ)
	if slot != nil {
		return slot.Fn(f, w, v)
	}
	return NotImplemented, nil
}

// compareDefault returns is the fallback logic for object comparison. It
// closely resembles the behavior of CPython's default_3way_compare in object.c.
func compareDefault(f *Frame, v, w *Object) int {
	if v.typ == w.typ {
		pv, pw := uintptr(v.toPointer()), uintptr(w.toPointer())
		if pv < pw {
			return -1
		}
		if pv == pw {
			return 0
		}
		return 1
	}
	if v == None {
		return -1
	}
	if w == None {
		return 1
	}
	// TODO: In default_3way_compare, the number type name is the empty
	// string so it evaluates less than non-number types. Once Grumpy
	// supports the concept of number types, add this behavior.
	if v.typ.Name() < w.typ.Name() {
		return -1
	}
	if v.typ.Name() != w.typ.Name() {
		return 1
	}
	if uintptr(v.typ.toPointer()) < uintptr(w.typ.toPointer()) {
		return -1
	}
	return 1
}

func checkFunctionArgs(f *Frame, function string, args Args, types ...*Type) *BaseException {
	if len(args) != len(types) {
		msg := fmt.Sprintf("'%s' requires %d arguments", function, len(types))
		return f.RaiseType(TypeErrorType, msg)
	}
	for i, t := range types {
		if !args[i].isInstance(t) {
			format := "'%s' requires a '%s' object but received a %q"
			return f.RaiseType(TypeErrorType, fmt.Sprintf(format, function, t.Name(), args[i].typ.Name()))
		}
	}
	return nil
}

func checkFunctionVarArgs(f *Frame, function string, args Args, types ...*Type) *BaseException {
	if len(args) <= len(types) {
		return checkFunctionArgs(f, function, args, types...)
	}
	return checkFunctionArgs(f, function, args[:len(types)], types...)
}

func checkMethodArgs(f *Frame, method string, args Args, types ...*Type) *BaseException {
	if len(args) != len(types) {
		msg := fmt.Sprintf("'%s' of '%s' requires %d arguments", method, types[0].Name(), len(types))
		return f.RaiseType(TypeErrorType, msg)
	}
	for i, t := range types {
		if !args[i].isInstance(t) {
			format := "'%s' requires a '%s' object but received a '%s'"
			return f.RaiseType(TypeErrorType, fmt.Sprintf(format, method, t.Name(), args[i].typ.Name()))
		}
	}
	return nil
}

func checkMethodVarArgs(f *Frame, method string, args Args, types ...*Type) *BaseException {
	if len(args) <= len(types) {
		return checkMethodArgs(f, method, args, types...)
	}
	return checkMethodArgs(f, method, args[:len(types)], types...)
}

func hashNotImplemented(f *Frame, o *Object) (*Object, *BaseException) {
	return nil, f.RaiseType(TypeErrorType, fmt.Sprintf("unhashable type: '%s'", o.typ.Name()))
}

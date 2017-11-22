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
	"bytes"
	"fmt"
	"reflect"
	"sync/atomic"
	"unsafe"
)

var (
	// DictType is the object representing the Python 'dict' type.
	DictType              = newBasisType("dict", reflect.TypeOf(Dict{}), toDictUnsafe, ObjectType)
	dictItemIteratorType  = newBasisType("dictionary-itemiterator", reflect.TypeOf(dictItemIterator{}), toDictItemIteratorUnsafe, ObjectType)
	dictKeyIteratorType   = newBasisType("dictionary-keyiterator", reflect.TypeOf(dictKeyIterator{}), toDictKeyIteratorUnsafe, ObjectType)
	dictValueIteratorType = newBasisType("dictionary-valueiterator", reflect.TypeOf(dictValueIterator{}), toDictValueIteratorUnsafe, ObjectType)
	deletedEntry          = &dictEntry{}
)

const (
	// maxDictSize is the largest number of entries a dictionary can hold.
	// Dict sizes must be a power of two and this is the largest such
	// number representable as int32.
	maxDictSize = 1 << 30
	minDictSize = 8
)

// dictEntry represents a slot in the hash table of a Dict. Entries are
// intended to be immutable so that they can be read atomically.
type dictEntry struct {
	hash  int
	key   *Object
	value *Object
}

// dictTable is the hash table underlying Dict.
type dictTable struct {
	// used is the number of slots in the entries table that contain values.
	used int32
	// fill is the number of slots that are used or once were used but have
	// since been cleared. Thus used <= fill <= len(entries).
	fill int
	// entries is a slice of immutable dict entries. Although elements in
	// the slice will be modified to point to different dictEntry objects
	// as the dictionary is updated, the slice itself (i.e. location in
	// memory and size) will not change for the lifetime of a dictTable.
	// When the table is no longer large enough to hold a dict's contents,
	// a new dictTable will be created.
	entries []*dictEntry
}

// newDictTable allocates a table where at least minCapacity entries can be
// accommodated. minCapacity must be <= maxDictSize.
func newDictTable(minCapacity int) *dictTable {
	// This takes the given capacity and sets all bits less than the highest bit.
	// Adding 1 to that value causes the number to become a multiple of 2 again.
	// The minDictSize is mixed in to make sure the resulting value is at least
	// that big. This implementation makes the function able to be inlined, as
	// well as allows for complete evaluation of constants at compile time.
	numEntries := (minDictSize - 1) | minCapacity
	numEntries |= numEntries >> 1
	numEntries |= numEntries >> 2
	numEntries |= numEntries >> 4
	numEntries |= numEntries >> 8
	numEntries |= numEntries >> 16
	return &dictTable{entries: make([]*dictEntry, numEntries+1)}
}

// loadEntry atomically loads the i'th entry in t and returns it.
func (t *dictTable) loadEntry(i int) *dictEntry {
	p := (*unsafe.Pointer)(unsafe.Pointer(&t.entries[i]))
	return (*dictEntry)(atomic.LoadPointer(p))
}

// storeEntry atomically sets the i'th entry in t to entry.
func (t *dictTable) storeEntry(i int, entry *dictEntry) {
	p := (*unsafe.Pointer)(unsafe.Pointer(&t.entries[i]))
	atomic.StorePointer(p, unsafe.Pointer(entry))
}

func (t *dictTable) loadUsed() int {
	return int(atomic.LoadInt32(&t.used))
}

func (t *dictTable) incUsed(n int) {
	atomic.AddInt32(&t.used, int32(n))
}

// insertAbsentEntry adds the populated entry to t assuming that the key
// specified in entry is absent from t. Since the key is absent, no key
// comparisons are necessary to perform the insert.
func (t *dictTable) insertAbsentEntry(entry *dictEntry) {
	mask := uint(len(t.entries) - 1)
	i := uint(entry.hash) & mask
	perturb := uint(entry.hash)
	index := i
	// The key we're trying to insert is known to be absent from the dict
	// so probe for the first nil entry.
	for ; t.entries[index] != nil; index = i & mask {
		i, perturb = dictNextIndex(i, perturb)
	}
	t.entries[index] = entry
	t.incUsed(1)
	t.fill++
}

// lookupEntry returns the index and entry in t with the given hash and key.
// Elements in the table are updated with immutable entries atomically and
// lookupEntry loads them atomically. So it is not necessary to lock the dict
// to do entry lookups in a consistent way.
func (t *dictTable) lookupEntry(f *Frame, hash int, key *Object) (int, *dictEntry, *BaseException) {
	mask := uint(len(t.entries) - 1)
	i, perturb := uint(hash)&mask, uint(hash)
	// free is the first slot that's available. We don't immediately use it
	// because it has been previously used and therefore an exact match may
	// be found further on.
	free := -1
	var freeEntry *dictEntry
	index := int(i & mask)
	entry := t.loadEntry(index)
	for {
		if entry == nil {
			if free != -1 {
				index = free
				// Store the entry instead of fetching by index
				// later since it may have changed by then.
				entry = freeEntry
			}
			break
		}
		if entry == deletedEntry {
			if free == -1 {
				free = index
			}
		} else if entry.hash == hash {
			o, raised := Eq(f, entry.key, key)
			if raised != nil {
				return -1, nil, raised
			}
			eq, raised := IsTrue(f, o)
			if raised != nil {
				return -1, nil, raised
			}
			if eq {
				break
			}
		}
		i, perturb = dictNextIndex(i, perturb)
		index = int(i & mask)
		entry = t.loadEntry(index)
	}
	return index, entry, nil
}

// writeEntry replaces t's entry at the given index with entry. If writing
// entry would cause t's fill ratio to grow too large then a new table is
// created, the entry is instead inserted there and that table is returned. t
// remains unchanged. When a sufficiently sized table cannot be created, false
// will be returned for the second value, otherwise true will be returned.
func (t *dictTable) writeEntry(f *Frame, index int, entry *dictEntry) (*dictTable, bool) {
	if t.entries[index] == deletedEntry {
		t.storeEntry(index, entry)
		t.incUsed(1)
		return nil, true
	}
	if t.entries[index] != nil {
		t.storeEntry(index, entry)
		return nil, true
	}
	if (t.fill+1)*3 <= len(t.entries)*2 {
		// New entry does not necessitate growing the table.
		t.storeEntry(index, entry)
		t.incUsed(1)
		t.fill++
		return nil, true
	}
	// Grow the table.
	var n int
	if t.used <= 50000 {
		n = int(t.used * 4)
	} else if t.used <= maxDictSize/2 {
		n = int(t.used * 2)
	} else {
		return nil, false
	}
	newTable := newDictTable(n)
	for _, oldEntry := range t.entries {
		if oldEntry != nil && oldEntry != deletedEntry {
			newTable.insertAbsentEntry(oldEntry)
		}
	}
	newTable.insertAbsentEntry(entry)
	return newTable, true
}

// dictEntryIterator is used to iterate over the entries in a dictTable in an
// arbitrary order.
type dictEntryIterator struct {
	index int64
	table *dictTable
}

// newDictEntryIterator creates a dictEntryIterator object for d. It assumes
// that d.mutex is held by the caller.
func newDictEntryIterator(d *Dict) dictEntryIterator {
	return dictEntryIterator{table: d.loadTable()}
}

// next advances this iterator to the next occupied entry and returns it. The
// second return value is true if the dict changed since iteration began, false
// otherwise.
func (iter *dictEntryIterator) next() *dictEntry {
	numEntries := len(iter.table.entries)
	var entry *dictEntry
	for entry == nil {
		// 64bit atomic ops need to be 8 byte aligned. This compile time check
		// verifies alignment by creating a negative constant for an unsigned type.
		// See sync/atomic docs for details.
		const blank = -(unsafe.Offsetof(iter.index) % 8)
		index := int(atomic.AddInt64(&iter.index, 1)) - 1
		if index >= numEntries {
			break
		}
		entry = iter.table.loadEntry(index)
		if entry == deletedEntry {
			entry = nil
		}
	}
	return entry
}

// dictVersionGuard is used to detect when a dict has been modified.
type dictVersionGuard struct {
	dict    *Dict
	version int64
}

func newDictVersionGuard(d *Dict) dictVersionGuard {
	return dictVersionGuard{d, d.loadVersion()}
}

// check returns false if the dict held by g has changed since g was created,
// true otherwise.
func (g *dictVersionGuard) check() bool {
	return g.dict.loadVersion() == g.version
}

// Dict represents Python 'dict' objects. The public methods of *Dict are
// thread safe.
type Dict struct {
	Object
	table *dictTable
	// We use a recursive mutex for synchronization because the hash and
	// key comparison operations may re-enter DelItem/SetItem.
	mutex recursiveMutex
	// version is incremented whenever the Dict is modified. See:
	// https://www.python.org/dev/peps/pep-0509/
	version int64
}

// NewDict returns an empty Dict.
func NewDict() *Dict {
	return &Dict{Object: Object{typ: DictType}, table: newDictTable(0)}
}

func newStringDict(items map[string]*Object) *Dict {
	if len(items) > maxDictSize/2 {
		panic(fmt.Sprintf("dictionary too big: %d", len(items)))
	}
	n := len(items) * 2
	table := newDictTable(n)
	for key, value := range items {
		table.insertAbsentEntry(&dictEntry{hashString(key), NewStr(key).ToObject(), value})
	}
	return &Dict{Object: Object{typ: DictType}, table: table}
}

func toDictUnsafe(o *Object) *Dict {
	return (*Dict)(o.toPointer())
}

// loadTable atomically loads and returns d's underlying dictTable.
func (d *Dict) loadTable() *dictTable {
	p := (*unsafe.Pointer)(unsafe.Pointer(&d.table))
	return (*dictTable)(atomic.LoadPointer(p))
}

// storeTable atomically updates d's underlying dictTable to the one given.
func (d *Dict) storeTable(table *dictTable) {
	p := (*unsafe.Pointer)(unsafe.Pointer(&d.table))
	atomic.StorePointer(p, unsafe.Pointer(table))
}

// loadVersion atomically loads and returns d's version.
func (d *Dict) loadVersion() int64 {
	// 64bit atomic ops need to be 8 byte aligned. This compile time check
	// verifies alignment by creating a negative constant for an unsigned type.
	// See sync/atomic docs for details.
	const blank = -(unsafe.Offsetof(d.version) % 8)
	return atomic.LoadInt64(&d.version)
}

// incVersion atomically increments d's version.
func (d *Dict) incVersion() {
	// 64bit atomic ops need to be 8 byte aligned. This compile time check
	// verifies alignment by creating a negative constant for an unsigned type.
	// See sync/atomic docs for details.
	const blank = -(unsafe.Offsetof(d.version) % 8)
	atomic.AddInt64(&d.version, 1)
}

// DelItem removes the entry associated with key from d. It returns true if an
// item was removed, or false if it did not exist in d.
func (d *Dict) DelItem(f *Frame, key *Object) (bool, *BaseException) {
	originValue, raised := d.putItem(f, key, nil, true)
	if raised != nil {
		return false, raised
	}
	return originValue != nil, nil
}

// DelItemString removes the entry associated with key from d. It returns true
// if an item was removed, or false if it did not exist in d.
func (d *Dict) DelItemString(f *Frame, key string) (bool, *BaseException) {
	return d.DelItem(f, NewStr(key).ToObject())
}

// GetItem looks up key in d, returning the associated value or nil if key is
// not present in d.
func (d *Dict) GetItem(f *Frame, key *Object) (*Object, *BaseException) {
	hash, raised := Hash(f, key)
	if raised != nil {
		return nil, raised
	}
	_, entry, raised := d.loadTable().lookupEntry(f, hash.Value(), key)
	if raised != nil {
		return nil, raised
	}
	if entry != nil && entry != deletedEntry {
		return entry.value, nil
	}
	return nil, nil
}

// GetItemString looks up key in d, returning the associated value or nil if
// key is not present in d.
func (d *Dict) GetItemString(f *Frame, key string) (*Object, *BaseException) {
	return d.GetItem(f, NewStr(key).ToObject())
}

// Pop looks up key in d, returning and removing the associalted value if exist,
// or nil if key is not present in d.
func (d *Dict) Pop(f *Frame, key *Object) (*Object, *BaseException) {
	return d.putItem(f, key, nil, true)
}

// Keys returns a list containing all the keys in d.
func (d *Dict) Keys(f *Frame) *List {
	d.mutex.Lock(f)
	keys := make([]*Object, d.Len())
	i := 0
	for _, entry := range d.table.entries {
		if entry != nil && entry != deletedEntry {
			keys[i] = entry.key
			i++
		}
	}
	d.mutex.Unlock(f)
	return NewList(keys...)
}

// Len returns the number of entries in d.
func (d *Dict) Len() int {
	return d.loadTable().loadUsed()
}

// putItem associates value with key in d, returning the old associated value if
// the key was added, or nil if it was not already present in d.
func (d *Dict) putItem(f *Frame, key, value *Object, overwrite bool) (*Object, *BaseException) {
	hash, raised := Hash(f, key)
	if raised != nil {
		return nil, raised
	}
	d.mutex.Lock(f)
	t := d.table
	v := d.version
	index, entry, raised := t.lookupEntry(f, hash.Value(), key)
	var originValue *Object
	if raised == nil {
		if v != d.version {
			// Dictionary was recursively modified. Blow up instead
			// of trying to recover.
			raised = f.RaiseType(RuntimeErrorType, "dictionary changed during write")
		} else {
			if value == nil {
				// Going to delete the entry.
				if entry != nil && entry != deletedEntry {
					d.table.storeEntry(index, deletedEntry)
					d.table.incUsed(-1)
					d.incVersion()
				}
			} else if overwrite || entry == nil {
				newEntry := &dictEntry{hash.Value(), key, value}
				if newTable, ok := t.writeEntry(f, index, newEntry); ok {
					if newTable != nil {
						d.storeTable(newTable)
					}
					d.incVersion()
				} else {
					raised = f.RaiseType(OverflowErrorType, errResultTooLarge)
				}
			}
			if entry != nil && entry != deletedEntry {
				originValue = entry.value
			}
		}
	}
	d.mutex.Unlock(f)
	return originValue, raised
}

// SetItem associates value with key in d.
func (d *Dict) SetItem(f *Frame, key, value *Object) *BaseException {
	_, raised := d.putItem(f, key, value, true)
	return raised
}

// SetItemString associates value with key in d.
func (d *Dict) SetItemString(f *Frame, key string, value *Object) *BaseException {
	return d.SetItem(f, NewStr(key).ToObject(), value)
}

// ToObject upcasts d to an Object.
func (d *Dict) ToObject() *Object {
	return &d.Object
}

// Update copies the items from the mapping or sequence of 2-tuples o into d.
func (d *Dict) Update(f *Frame, o *Object) (raised *BaseException) {
	var iter *Object
	if o.isInstance(DictType) {
		d2 := toDictUnsafe(o)
		d2.mutex.Lock(f)
		// Concurrent modifications to d2 will cause Update to raise
		// "dictionary changed during iteration".
		iter = newDictItemIterator(d2).ToObject()
		d2.mutex.Unlock(f)
	} else {
		iter, raised = Iter(f, o)
	}
	if raised != nil {
		return raised
	}
	return seqForEach(f, iter, func(item *Object) *BaseException {
		return seqApply(f, item, func(elems []*Object, _ bool) *BaseException {
			if numElems := len(elems); numElems != 2 {
				format := "dictionary update sequence element has length %d; 2 is required"
				return f.RaiseType(ValueErrorType, fmt.Sprintf(format, numElems))
			}
			return d.SetItem(f, elems[0], elems[1])
		})
	})
}

// dictsAreEqual returns true if d1 and d2 have the same keys and values, false
// otherwise. If either d1 or d2 are concurrently modified then RuntimeError is
// raised.
func dictsAreEqual(f *Frame, d1, d2 *Dict) (bool, *BaseException) {
	if d1 == d2 {
		return true, nil
	}
	// Do not hold both locks at the same time to avoid deadlock.
	d1.mutex.Lock(f)
	iter := newDictEntryIterator(d1)
	g1 := newDictVersionGuard(d1)
	len1 := d1.Len()
	d1.mutex.Unlock(f)
	d2.mutex.Lock(f)
	g2 := newDictVersionGuard(d1)
	len2 := d2.Len()
	d2.mutex.Unlock(f)
	if len1 != len2 {
		return false, nil
	}
	result := true
	for entry := iter.next(); entry != nil && result; entry = iter.next() {
		if v, raised := d2.GetItem(f, entry.key); raised != nil {
			return false, raised
		} else if v == nil {
			result = false
		} else {
			eq, raised := Eq(f, entry.value, v)
			if raised != nil {
				return false, raised
			}
			result, raised = IsTrue(f, eq)
			if raised != nil {
				return false, raised
			}
		}
	}
	if !g1.check() || !g2.check() {
		return false, f.RaiseType(RuntimeErrorType, "dictionary changed during iteration")
	}
	return result, nil
}

func dictClear(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "clear", args, DictType); raised != nil {
		return nil, raised
	}
	d := toDictUnsafe(args[0])
	d.mutex.Lock(f)
	d.table = newDictTable(0)
	d.incVersion()
	d.mutex.Unlock(f)
	return None, nil
}

func dictContains(f *Frame, seq, value *Object) (*Object, *BaseException) {
	item, raised := toDictUnsafe(seq).GetItem(f, value)
	if raised != nil {
		return nil, raised
	}
	return GetBool(item != nil).ToObject(), nil
}

func dictCopy(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "copy", args, DictType); raised != nil {
		return nil, raised
	}
	return DictType.Call(f, args, nil)
}

func dictDelItem(f *Frame, o, key *Object) *BaseException {
	deleted, raised := toDictUnsafe(o).DelItem(f, key)
	if raised != nil {
		return raised
	}
	if !deleted {
		return raiseKeyError(f, key)
	}
	return nil
}

func dictEq(f *Frame, v, w *Object) (*Object, *BaseException) {
	if !w.isInstance(DictType) {
		return NotImplemented, nil
	}
	eq, raised := dictsAreEqual(f, toDictUnsafe(v), toDictUnsafe(w))
	if raised != nil {
		return nil, raised
	}
	return GetBool(eq).ToObject(), nil
}

func dictGet(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	expectedTypes := []*Type{DictType, ObjectType, ObjectType}
	argc := len(args)
	if argc == 2 {
		expectedTypes = expectedTypes[:2]
	}
	if raised := checkMethodArgs(f, "get", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	item, raised := toDictUnsafe(args[0]).GetItem(f, args[1])
	if raised == nil && item == nil {
		item = None
		if argc > 2 {
			item = args[2]
		}
	}
	return item, raised
}

func dictHasKey(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "has_key", args, DictType, ObjectType); raised != nil {
		return nil, raised
	}
	return dictContains(f, args[0], args[1])
}

func dictItems(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "items", args, DictType); raised != nil {
		return nil, raised
	}
	d := toDictUnsafe(args[0])
	d.mutex.Lock(f)
	iter := newDictItemIterator(d).ToObject()
	d.mutex.Unlock(f)
	return ListType.Call(f, Args{iter}, nil)
}

func dictIterItems(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "iteritems", args, DictType); raised != nil {
		return nil, raised
	}
	d := toDictUnsafe(args[0])
	d.mutex.Lock(f)
	iter := newDictItemIterator(d).ToObject()
	d.mutex.Unlock(f)
	return iter, nil
}

func dictIterKeys(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "iterkeys", args, DictType); raised != nil {
		return nil, raised
	}
	return dictIter(f, args[0])
}

func dictIterValues(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "itervalues", args, DictType); raised != nil {
		return nil, raised
	}
	d := toDictUnsafe(args[0])
	d.mutex.Lock(f)
	iter := newDictValueIterator(d).ToObject()
	d.mutex.Unlock(f)
	return iter, nil
}

func dictKeys(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "keys", args, DictType); raised != nil {
		return nil, raised
	}
	return toDictUnsafe(args[0]).Keys(f).ToObject(), nil
}

func dictGetItem(f *Frame, o, key *Object) (*Object, *BaseException) {
	item, raised := toDictUnsafe(o).GetItem(f, key)
	if raised != nil {
		return nil, raised
	}
	if item == nil {
		return nil, raiseKeyError(f, key)
	}
	return item, nil
}

func dictInit(f *Frame, o *Object, args Args, kwargs KWArgs) (*Object, *BaseException) {
	var expectedTypes []*Type
	argc := len(args)
	if argc > 0 {
		expectedTypes = []*Type{ObjectType}
	}
	if raised := checkFunctionArgs(f, "__init__", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	d := toDictUnsafe(o)
	if argc > 0 {
		if raised := d.Update(f, args[0]); raised != nil {
			return nil, raised
		}
	}
	for _, kwarg := range kwargs {
		if raised := d.SetItemString(f, kwarg.Name, kwarg.Value); raised != nil {
			return nil, raised
		}
	}
	return None, nil
}

func dictIter(f *Frame, o *Object) (*Object, *BaseException) {
	d := toDictUnsafe(o)
	d.mutex.Lock(f)
	iter := newDictKeyIterator(d).ToObject()
	d.mutex.Unlock(f)
	return iter, nil
}

func dictLen(f *Frame, o *Object) (*Object, *BaseException) {
	d := toDictUnsafe(o)
	ret := NewInt(d.Len()).ToObject()
	return ret, nil
}

func dictNE(f *Frame, v, w *Object) (*Object, *BaseException) {
	if !w.isInstance(DictType) {
		return NotImplemented, nil
	}
	eq, raised := dictsAreEqual(f, toDictUnsafe(v), toDictUnsafe(w))
	if raised != nil {
		return nil, raised
	}
	return GetBool(!eq).ToObject(), nil
}

func dictNew(f *Frame, t *Type, _ Args, _ KWArgs) (*Object, *BaseException) {
	d := toDictUnsafe(newObject(t))
	d.table = &dictTable{entries: make([]*dictEntry, minDictSize, minDictSize)}
	return d.ToObject(), nil
}

func dictPop(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	expectedTypes := []*Type{DictType, ObjectType, ObjectType}
	argc := len(args)
	if argc == 2 {
		expectedTypes = expectedTypes[:2]
	}
	if raised := checkMethodArgs(f, "pop", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	key := args[1]
	d := toDictUnsafe(args[0])
	item, raised := d.Pop(f, key)
	if raised == nil && item == nil {
		if argc > 2 {
			item = args[2]
		} else {
			raised = raiseKeyError(f, key)
		}
	}
	return item, raised
}

func dictPopItem(f *Frame, args Args, _ KWArgs) (item *Object, raised *BaseException) {
	if raised := checkMethodArgs(f, "popitem", args, DictType); raised != nil {
		return nil, raised
	}
	d := toDictUnsafe(args[0])
	d.mutex.Lock(f)
	iter := newDictEntryIterator(d)
	entry := iter.next()
	if entry == nil {
		raised = f.RaiseType(KeyErrorType, "popitem(): dictionary is empty")
	} else {
		item = NewTuple(entry.key, entry.value).ToObject()
		d.table.storeEntry(int(iter.index-1), deletedEntry)
		d.table.incUsed(-1)
		d.incVersion()
	}
	d.mutex.Unlock(f)
	return item, raised
}

func dictRepr(f *Frame, o *Object) (*Object, *BaseException) {
	d := toDictUnsafe(o)
	if f.reprEnter(d.ToObject()) {
		return NewStr("{...}").ToObject(), nil
	}
	defer f.reprLeave(d.ToObject())
	// Lock d so that we get a consistent view of it. Otherwise we may
	// return a state that d was never actually in.
	d.mutex.Lock(f)
	defer d.mutex.Unlock(f)
	var buf bytes.Buffer
	buf.WriteString("{")
	iter := newDictEntryIterator(d)
	i := 0
	for entry := iter.next(); entry != nil; entry = iter.next() {
		if i > 0 {
			buf.WriteString(", ")
		}
		s, raised := Repr(f, entry.key)
		if raised != nil {
			return nil, raised
		}
		buf.WriteString(s.Value())
		buf.WriteString(": ")
		if s, raised = Repr(f, entry.value); raised != nil {
			return nil, raised
		}
		buf.WriteString(s.Value())
		i++
	}
	buf.WriteString("}")
	return NewStr(buf.String()).ToObject(), nil
}

func dictSetDefault(f *Frame, args Args, _ KWArgs) (*Object, *BaseException) {
	argc := len(args)
	if argc == 1 {
		return nil, f.RaiseType(TypeErrorType, "setdefault expected at least 1 arguments, got 0")
	}
	if argc > 3 {
		return nil, f.RaiseType(TypeErrorType, fmt.Sprintf("setdefault expected at most 2 arguments, got %v", argc-1))
	}
	expectedTypes := []*Type{DictType, ObjectType, ObjectType}
	if argc == 2 {
		expectedTypes = expectedTypes[:2]
	}
	if raised := checkMethodArgs(f, "setdefault", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	d := toDictUnsafe(args[0])
	key := args[1]
	var value *Object
	if argc > 2 {
		value = args[2]
	} else {
		value = None
	}
	originValue, raised := d.putItem(f, key, value, false)
	if originValue != nil {
		return originValue, raised
	}
	return value, raised
}

func dictSetItem(f *Frame, o, key, value *Object) *BaseException {
	return toDictUnsafe(o).SetItem(f, key, value)
}

func dictUpdate(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	expectedTypes := []*Type{DictType, ObjectType}
	argc := len(args)
	if argc == 1 {
		expectedTypes = expectedTypes[:1]
	}
	if raised := checkMethodArgs(f, "update", args, expectedTypes...); raised != nil {
		return nil, raised
	}
	d := toDictUnsafe(args[0])
	if argc > 1 {
		if raised := d.Update(f, args[1]); raised != nil {
			return nil, raised
		}
	}
	for _, kwarg := range kwargs {
		if raised := d.SetItemString(f, kwarg.Name, kwarg.Value); raised != nil {
			return nil, raised
		}
	}
	return None, nil
}

func dictValues(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "values", args, DictType); raised != nil {
		return nil, raised
	}
	iter, raised := dictIterValues(f, args, nil)
	if raised != nil {
		return nil, raised
	}
	return ListType.Call(f, Args{iter}, nil)
}

func initDictType(dict map[string]*Object) {
	dict["clear"] = newBuiltinFunction("clear", dictClear).ToObject()
	dict["copy"] = newBuiltinFunction("copy", dictCopy).ToObject()
	dict["get"] = newBuiltinFunction("get", dictGet).ToObject()
	dict["has_key"] = newBuiltinFunction("has_key", dictHasKey).ToObject()
	dict["items"] = newBuiltinFunction("items", dictItems).ToObject()
	dict["iteritems"] = newBuiltinFunction("iteritems", dictIterItems).ToObject()
	dict["iterkeys"] = newBuiltinFunction("iterkeys", dictIterKeys).ToObject()
	dict["itervalues"] = newBuiltinFunction("itervalues", dictIterValues).ToObject()
	dict["keys"] = newBuiltinFunction("keys", dictKeys).ToObject()
	dict["pop"] = newBuiltinFunction("pop", dictPop).ToObject()
	dict["popitem"] = newBuiltinFunction("popitem", dictPopItem).ToObject()
	dict["setdefault"] = newBuiltinFunction("setdefault", dictSetDefault).ToObject()
	dict["update"] = newBuiltinFunction("update", dictUpdate).ToObject()
	dict["values"] = newBuiltinFunction("values", dictValues).ToObject()
	DictType.slots.Contains = &binaryOpSlot{dictContains}
	DictType.slots.DelItem = &delItemSlot{dictDelItem}
	DictType.slots.Eq = &binaryOpSlot{dictEq}
	DictType.slots.GetItem = &binaryOpSlot{dictGetItem}
	DictType.slots.Hash = &unaryOpSlot{hashNotImplemented}
	DictType.slots.Init = &initSlot{dictInit}
	DictType.slots.Iter = &unaryOpSlot{dictIter}
	DictType.slots.Len = &unaryOpSlot{dictLen}
	DictType.slots.NE = &binaryOpSlot{dictNE}
	DictType.slots.New = &newSlot{dictNew}
	DictType.slots.Repr = &unaryOpSlot{dictRepr}
	DictType.slots.SetItem = &setItemSlot{dictSetItem}
}

type dictItemIterator struct {
	Object
	iter  dictEntryIterator
	guard dictVersionGuard
}

// newDictItemIterator creates a dictItemIterator object for d. It assumes that
// d.mutex is held by the caller.
func newDictItemIterator(d *Dict) *dictItemIterator {
	return &dictItemIterator{
		Object: Object{typ: dictItemIteratorType},
		iter:   newDictEntryIterator(d),
		guard:  newDictVersionGuard(d),
	}
}

func toDictItemIteratorUnsafe(o *Object) *dictItemIterator {
	return (*dictItemIterator)(o.toPointer())
}

func (iter *dictItemIterator) ToObject() *Object {
	return &iter.Object
}

func dictItemIteratorIter(f *Frame, o *Object) (*Object, *BaseException) {
	return o, nil
}

func dictItemIteratorNext(f *Frame, o *Object) (ret *Object, raised *BaseException) {
	iter := toDictItemIteratorUnsafe(o)
	entry, raised := dictIteratorNext(f, &iter.iter, &iter.guard)
	if raised != nil {
		return nil, raised
	}
	return NewTuple2(entry.key, entry.value).ToObject(), nil
}

func initDictItemIteratorType(map[string]*Object) {
	dictItemIteratorType.flags &^= typeFlagBasetype | typeFlagInstantiable
	dictItemIteratorType.slots.Iter = &unaryOpSlot{dictItemIteratorIter}
	dictItemIteratorType.slots.Next = &unaryOpSlot{dictItemIteratorNext}
}

type dictKeyIterator struct {
	Object
	iter  dictEntryIterator
	guard dictVersionGuard
}

// newDictKeyIterator creates a dictKeyIterator object for d. It assumes that
// d.mutex is held by the caller.
func newDictKeyIterator(d *Dict) *dictKeyIterator {
	return &dictKeyIterator{
		Object: Object{typ: dictKeyIteratorType},
		iter:   newDictEntryIterator(d),
		guard:  newDictVersionGuard(d),
	}
}

func toDictKeyIteratorUnsafe(o *Object) *dictKeyIterator {
	return (*dictKeyIterator)(o.toPointer())
}

func (iter *dictKeyIterator) ToObject() *Object {
	return &iter.Object
}

func dictKeyIteratorIter(f *Frame, o *Object) (*Object, *BaseException) {
	return o, nil
}

func dictKeyIteratorNext(f *Frame, o *Object) (*Object, *BaseException) {
	iter := toDictKeyIteratorUnsafe(o)
	entry, raised := dictIteratorNext(f, &iter.iter, &iter.guard)
	if raised != nil {
		return nil, raised
	}
	return entry.key, nil
}

func initDictKeyIteratorType(map[string]*Object) {
	dictKeyIteratorType.flags &^= typeFlagBasetype | typeFlagInstantiable
	dictKeyIteratorType.slots.Iter = &unaryOpSlot{dictKeyIteratorIter}
	dictKeyIteratorType.slots.Next = &unaryOpSlot{dictKeyIteratorNext}
}

type dictValueIterator struct {
	Object
	iter  dictEntryIterator
	guard dictVersionGuard
}

// newDictValueIterator creates a dictValueIterator object for d. It assumes
// that d.mutex is held by the caller.
func newDictValueIterator(d *Dict) *dictValueIterator {
	return &dictValueIterator{
		Object: Object{typ: dictValueIteratorType},
		iter:   newDictEntryIterator(d),
		guard:  newDictVersionGuard(d),
	}
}

func toDictValueIteratorUnsafe(o *Object) *dictValueIterator {
	return (*dictValueIterator)(o.toPointer())
}

func (iter *dictValueIterator) ToObject() *Object {
	return &iter.Object
}

func dictValueIteratorIter(f *Frame, o *Object) (*Object, *BaseException) {
	return o, nil
}

func dictValueIteratorNext(f *Frame, o *Object) (*Object, *BaseException) {
	iter := toDictValueIteratorUnsafe(o)
	entry, raised := dictIteratorNext(f, &iter.iter, &iter.guard)
	if raised != nil {
		return nil, raised
	}
	return entry.value, nil
}

func initDictValueIteratorType(map[string]*Object) {
	dictValueIteratorType.flags &^= typeFlagBasetype | typeFlagInstantiable
	dictValueIteratorType.slots.Iter = &unaryOpSlot{dictValueIteratorIter}
	dictValueIteratorType.slots.Next = &unaryOpSlot{dictValueIteratorNext}
}

func raiseKeyError(f *Frame, key *Object) *BaseException {
	s, raised := ToStr(f, key)
	if raised == nil {
		raised = f.RaiseType(KeyErrorType, s.Value())
	}
	return raised
}

func dictNextIndex(i, perturb uint) (uint, uint) {
	return (i << 2) + i + perturb + 1, perturb >> 5
}

func dictIteratorNext(f *Frame, iter *dictEntryIterator, guard *dictVersionGuard) (*dictEntry, *BaseException) {
	// NOTE: The behavior here diverges from CPython where an iterator that
	// is exhausted will always return StopIteration regardless whether the
	// underlying dict is subsequently modified. In Grumpy, an iterator for
	// a dict that has been modified will always raise RuntimeError even if
	// the iterator was exhausted before the modification.
	entry := iter.next()
	if !guard.check() {
		return nil, f.RaiseType(RuntimeErrorType, "dictionary changed during iteration")
	}
	if entry == nil {
		return nil, f.Raise(StopIterationType.ToObject(), nil, nil)
	}
	return entry, nil
}

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

	// Not as a real object = just a memory address. This isn't a pointer
	// since the only usecase for this is to have a unique memory address.
	// By having it a value type, the compiler is able to make all
	// `&deletedEntry` effectively constant (avoids a memory read if this
	// had a pointer type).
	deletedEntry Object
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

func (d dictEntry) isEmpty() bool   { return d.key == nil }
func (d dictEntry) isDeleted() bool { return d.key == &deletedEntry }
func (d dictEntry) isSet() bool     { return !d.isEmpty() && !d.isDeleted() }

// dictTable is the hash table underlying Dict.
type dictTable []dictEntry

// newDictTable allocates a table where at least minCapacity entries can be
// accommodated. minCapacity must be <= maxDictSize.
func newDictTable(minCapacity int) dictTable {
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
	return make(dictTable, numEntries+1)
}

// insertAbsentEntry adds the populated entry to t assuming that the key
// specified in entry is absent from t. Since the key is absent, no key
// comparisons are necessary to perform the insert.
func (t dictTable) insertAbsentEntry(entry dictEntry) {
	mask := uint(len(t) - 1)
	i, perturb := uint(entry.hash)&mask, uint(entry.hash)
	// The key we're trying to insert is known to be absent from the dict
	// so probe for the first empty entry.
top:
	index := i & mask
	if !t[index].isEmpty() {
		i, perturb = dictNextIndex(i, perturb)
		// We avoid a `for` loop so this method can be inlined and save
		// +1ns/call to insertAbsentEntry (which adds up since this is
		// called a lot).
		goto top
	}
	t[index] = entry
}

// lookupEntry returns the index and whether the given hash and key exist in
// the table. Calls to this either should be performed on the read(only) table
// or on the write table while it is locked.
func (t dictTable) lookupEntry(f *Frame, hash int, key *Object) (int, bool, *BaseException) {
	mask := uint(len(t) - 1)
	i, perturb := uint(hash)&mask, uint(hash)
	// free is the first slot that's available. We don't immediately use it
	// because it has been previously used and therefore an exact match may
	// be found further on.
	free := -1
	for {
		index := int(i & mask)
		switch entry := t[index]; entry.key {
		case key:
			return index, true, nil

		case nil:
			if free != -1 {
				index = free
			}
			return index, false, nil

		case &deletedEntry:
			if free == -1 {
				free = index
			}

		default:
			if entry.hash == hash {
				o, raised := Eq(f, entry.key, key)
				if raised != nil {
					return index, false, raised
				}
				if eq, raised := IsTrue(f, o); raised != nil || eq {
					return index, eq, raised
				}
			}
		}
		i, perturb = dictNextIndex(i, perturb)
	}
}

// writeEntry replaces d's entry at the given index with entry. If writing
// entry would cause d's fill ratio to grow too large then a new table is
// created, the entry is instead inserted there and that table is returned. t
// remains unchanged. When a sufficiently sized table cannot be created, false
// will be returned for the second value, otherwise true will be returned.
func (d *Dict) writeEntry(f *Frame, index int, entry dictEntry) (prevEntry dictEntry, ok bool) {
	prevEntry, d.write[index] = d.write[index], entry
	ok = true

	var usedDelta int32
	if entry.isSet() {
		usedDelta++
	}

	if prevEntry.isEmpty() {
		d.fill++
	} else if prevEntry.isSet() {
		usedDelta--
	}

	used := atomic.AddInt32(&d.used, usedDelta)
	if int(d.fill)*3 <= len(d.write)*2 {
		// Write entry does not necessitate growing the table.
		return
	}

	// Grow the table.
	var n int
	if used <= 50000 {
		n = int(used) * 4
	} else if used <= maxDictSize/2 {
		n = int(used) * 2
	} else {
		ok = false
		return
	}

	newTable := newDictTable(n)
	for _, oldEntry := range d.write {
		if oldEntry.isSet() {
			newTable.insertAbsentEntry(oldEntry)
		}
	}
	d.fill = used
	d.write = newTable
	return
}

// dictEntryIterator is used to iterate over the entries in a dictTable in an
// arbitrary order.
type dictEntryIterator struct {
	index int32
	table dictTable
}

// newDictEntryIterator creates a dictEntryIterator object for d.
func newDictEntryIterator(f *Frame, d *Dict) (iter dictEntryIterator) {
	if rtable := d.loadReadTable(); rtable != nil {
		iter.table = *rtable
	} else {
		d.mutex.Lock(f)
		iter.table = d.write
		if iter.table == nil {
			iter.table = *d.loadReadTable()
		} else {
			// Promote to prevent unlocked mutations to the
			// dictTable we are going to iterate over.
			d.promoteWriteToRead()
		}
		d.mutex.Unlock(f)
	}
	return
}

// next advances this iterator to the next occupied entry and returns it.
func (iter *dictEntryIterator) next() (entry dictEntry) {
	numEntries := len(iter.table)
	for !entry.isSet() {
		index := int(atomic.AddInt32(&iter.index, 1)) - 1
		if index >= numEntries {
			// Clear so we don't return a deleted entry and users can just use
			// `isEmpty` for speed.
			entry = dictEntry{}
			break
		}
		entry = iter.table[index]
	}
	return
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
	read *dictTable

	// used is the number of slots in the entries table where
	// slot.value!=nil.
	used int32

	// We use a recursive mutex for synchronization because the hash and
	// key comparison operations may re-enter DelItem/SetItem.
	mutex recursiveMutex
	write dictTable

	// fill is the number of slots where slot.key != nil.
	// Thus used <= fill <= len(entries).
	fill int32

	// The number of reads hitting the write table - helps gauge when the
	// write table should be promoted to the read table.
	misses int32

	// version is incremented whenever the Dict is modified. See:
	// https://www.python.org/dev/peps/pep-0509/
	version int64
}

// NewDict returns an empty Dict.
func NewDict() *Dict {
	return &Dict{
		Object: Object{typ: DictType},
		// We start ready to write so populating is fast(er).
		write: newDictTable(0),
	}
}

func newStringDict(items map[string]*Object) *Dict {
	if len(items) > maxDictSize/2 {
		panic(fmt.Sprintf("dictionary too big: %d", len(items)))
	}
	table := newDictTable(len(items) * 2)
	for key, value := range items {
		table.insertAbsentEntry(dictEntry{hashString(key), NewStr(key).ToObject(), value})
	}
	d := &Dict{
		Object: Object{typ: DictType},
		read:   &table,
		used:   int32(len(items)),
		fill:   int32(len(items)),
	}
	return d
}

func toDictUnsafe(o *Object) *Dict {
	return (*Dict)(o.toPointer())
}

// unsafeReadTablePointer returns `&d.read` as an unsafe pointer.
func (d *Dict) unsafeReadTablePointer() *unsafe.Pointer {
	return (*unsafe.Pointer)(unsafe.Pointer(&d.read))
}

// loadReadTable atomically fetches the read table. If nil, the read table
// isn't available and a fallback to the write table should be tried.
func (d *Dict) loadReadTable() *dictTable {
	return (*dictTable)(atomic.LoadPointer(d.unsafeReadTablePointer()))
}

// promoteWriteToRead promotes the write table to the read table. The mutex
// needs to be held for this operation.
func (d *Dict) promoteWriteToRead() (table dictTable) {
	table, d.write = d.write, nil
	// We must use a pointer to a local variable to prevent setting a
	// pointer to d.write (which would be bad).
	atomic.StorePointer(d.unsafeReadTablePointer(), unsafe.Pointer(&table))
	d.misses = 0
	return
}

// loadVersion atomically loads and returns d's version.
func (d *Dict) loadVersion() int64 {
	// 64bit atomic ops need to be 8 byte aligned. This compile time check
	// verifies alignment by creating a negative constant for an unsigned type.
	// See sync/atomic docs for details.
	const _ = -(unsafe.Offsetof(d.version) % 8)
	return atomic.LoadInt64(&d.version)
}

// incVersion atomically increments d's version.
func (d *Dict) incVersion() {
	// 64bit atomic ops need to be 8 byte aligned. This compile time check
	// verifies alignment by creating a negative constant for an unsigned type.
	// See sync/atomic docs for details.
	const _ = -(unsafe.Offsetof(d.version) % 8)
	atomic.AddInt64(&d.version, 1)
}

// populateWriteTable makes sure that d.write is populated with the dict's
// table, possibly copying it from the read table.
func (d *Dict) populateWriteTable() dictTable {
	if d.write == nil {
		// Copy the read-only table so we can do modifications.
		oldTable := *d.loadReadTable()
		if d.used == d.fill {
			// No deletion markers - use builtin copy for speed.
			d.write = make(dictTable, len(oldTable))
			copy(d.write, oldTable)
		} else {
			// Deletion markers - take the time to clean them out.
			d.write = newDictTable(int(d.used))
			for _, oldEntry := range oldTable {
				if oldEntry.isSet() {
					d.write.insertAbsentEntry(oldEntry)
				}
			}
		}
		// NOTE: d.read remains set until later. This allows reads to
		// happen while d.write is edited. Once we are ready to
		// publish, d.read must be cleared.
		d.fill = d.used
		d.misses = 0
	} else if d.misses > 0 {
		d.misses--
	}
	return d.write
}

// DelItem removes the entry associated with key from d. It returns true if an
// item was removed, or false if it did not exist in d.
func (d *Dict) DelItem(f *Frame, key *Object) (bool, *BaseException) {
	originValue, raised := d.putItem(f, key, nil)
	return originValue != nil, raised
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

	var table dictTable
	if rtable := d.loadReadTable(); rtable != nil {
		table = *rtable
	}

top:
	if table != nil {
		index, exists, raised := table.lookupEntry(f, hash.Value(), key)
		if raised != nil || !exists {
			return nil, raised
		}
		return table[index].value, nil
	}

	d.mutex.Lock(f)
	d.misses++
	table = d.write
	if table == nil {
		table = *d.loadReadTable()
		d.mutex.Unlock(f)
		goto top
	} else if d.misses > d.used {
		table = d.promoteWriteToRead()
		d.mutex.Unlock(f)
		goto top
	}

	index, exists, raised := table.lookupEntry(f, hash.Value(), key)
	// TODO: If the table changes during lookup, do we retry (like in
	// putItem)?
	var value *Object
	if exists && raised == nil {
		value = table[index].value
	}
	d.mutex.Unlock(f)
	return value, raised
}

// GetItemString looks up key in d, returning the associated value or nil if
// key is not present in d.
func (d *Dict) GetItemString(f *Frame, key string) (*Object, *BaseException) {
	return d.GetItem(f, NewStr(key).ToObject())
}

// Pop looks up key in d, returning and removing the associalted value if exist,
// or nil if key is not present in d.
func (d *Dict) Pop(f *Frame, key *Object) (*Object, *BaseException) {
	return d.putItem(f, key, nil)
}

// Keys returns a list containing all the keys in d.
func (d *Dict) Keys(f *Frame) *List {
	var table dictTable
	if rtable := d.loadReadTable(); rtable != nil {
		table = *rtable
	} else {
		d.mutex.Lock(f)
		d.misses++
		table = d.write
		if table == nil {
			table = *d.loadReadTable()
		} else if d.misses > d.used {
			d.promoteWriteToRead()
			d.mutex.Unlock(f)
		} else {
			defer d.mutex.Unlock(f)
		}
	}
	keys := make([]*Object, 0, d.Len())
	for _, entry := range table {
		if entry.isSet() {
			keys = append(keys, entry.key)
		}
	}
	return NewList(keys...)
}

// Len returns the number of entries in d.
func (d *Dict) Len() int {
	return int(atomic.LoadInt32(&d.used))
}

// putItem associates value with key in d, returning the old associated value,
// or nil if it was not already present in d.
func (d *Dict) putItem(f *Frame, key, value *Object) (*Object, *BaseException) {
	hash, raised := Hash(f, key)
	if raised != nil {
		return nil, raised
	}
	hashValue := hash.Value()

	entryKey := key
	if value == nil {
		entryKey = &deletedEntry
	}

	var originValue *Object
	d.mutex.Lock(f)
	v := d.version

top:
	table := d.populateWriteTable()
	index, _, raised := table.lookupEntry(f, hashValue, key)
	if raised == nil {
		if v != d.version {
			// Dictionary was recursively modified. Blow up instead
			// of trying to recover.
			raised = f.RaiseType(RuntimeErrorType, "dictionary changed during write")
		} else if &d.write[0] != &table[0] {
			goto top // Entry lookup caused tables to shift. Try again.
		} else if prevEntry, ok := d.writeEntry(f, index, dictEntry{hashValue, entryKey, value}); ok {
			originValue = prevEntry.value
			if value != nil || originValue != nil {
				d.incVersion()
			}
		} else {
			raised = f.RaiseType(OverflowErrorType, errResultTooLarge)
		}
	}

	if d.read != nil {
		// Time to "publish" the write table. Must use atomic for write
		// since other goroutines might be reading concurrently.
		atomic.StorePointer(d.unsafeReadTablePointer(), nil)
	}

	d.mutex.Unlock(f)
	return originValue, raised
}

// SetItem associates value with key in d.
func (d *Dict) SetItem(f *Frame, key, value *Object) *BaseException {
	_, raised := d.putItem(f, key, value)
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
		// Concurrent modifications to d2 will cause Update to raise
		// "dictionary changed during iteration".
		iter = newDictItemIterator(f, d2).ToObject()
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

	// NOTE: The length, iterator, and version may not be consistent. This
	// is actually OK. If the length is changing concurrently to this call,
	// then the programmer hasn't bothered to implement proper locking in
	// their code and in reality they don't know which statement is
	// happening before and which is happening after (mutator vs. eq).
	// Additionally, it shouldn't matter
	// that the version is potentially one higher (mutation in flight with
	// initial setup) for the same reason - they can't define an ordering.
	//
	// Put another way, if the operation is "atomic" and doesn't bleed back
	// into Python, then this should be too. If it isn't, they should have
	// a lock.
	iter := newDictEntryIterator(f, d1)
	len1 := d1.Len()
	g1 := newDictVersionGuard(d1)

	len2 := d2.Len()
	g2 := newDictVersionGuard(d2)
	if len1 != len2 {
		return false, nil
	}
	result := true
	for entry := iter.next(); !entry.isEmpty() && result; entry = iter.next() {
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
	// Start ready to write...
	d.write = newDictTable(0)
	atomic.StoreInt32(&d.used, 0)
	d.incVersion()
	d.fill = 0

	atomic.StorePointer(d.unsafeReadTablePointer(), nil)
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
	iter := newDictItemIterator(f, d).ToObject()
	return ListType.Call(f, Args{iter}, nil)
}

func dictIterItems(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "iteritems", args, DictType); raised != nil {
		return nil, raised
	}
	d := toDictUnsafe(args[0])
	iter := newDictItemIterator(f, d).ToObject()
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
	iter := newDictValueIterator(f, d).ToObject()
	return iter, nil
}

func dictKeys(f *Frame, args Args, kwargs KWArgs) (*Object, *BaseException) {
	if raised := checkMethodArgs(f, "keys", args, DictType); raised != nil {
		return nil, raised
	}
	return toDictUnsafe(args[0]).Keys(f).ToObject(), nil
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
	iter := newDictKeyIterator(f, d).ToObject()
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
	if t == DictType {
		return NewDict().ToObject(), nil
	}
	d := toDictUnsafe(newObject(t))
	table := newDictTable(0)
	d.read = &table
	return d.ToObject(), nil
}

func dictRepr(f *Frame, o *Object) (*Object, *BaseException) {
	d := toDictUnsafe(o)
	if f.reprEnter(d.ToObject()) {
		return NewStr("{...}").ToObject(), nil
	}
	defer f.reprLeave(d.ToObject())

	// Grab a snapshot of our current state:
	iter := newDictEntryIterator(f, d)

	var buf bytes.Buffer
	buf.WriteString("{")
	i := 0
	for entry := iter.next(); !entry.isEmpty(); entry = iter.next() {
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
	dict["get"] = newBuiltinFunction("get", dictGet).ToObject()
	dict["has_key"] = newBuiltinFunction("has_key", dictHasKey).ToObject()
	dict["items"] = newBuiltinFunction("items", dictItems).ToObject()
	dict["iteritems"] = newBuiltinFunction("iteritems", dictIterItems).ToObject()
	dict["iterkeys"] = newBuiltinFunction("iterkeys", dictIterKeys).ToObject()
	dict["itervalues"] = newBuiltinFunction("itervalues", dictIterValues).ToObject()
	dict["keys"] = newBuiltinFunction("keys", dictKeys).ToObject()
	dict["pop"] = newBuiltinFunction("pop", dictPop).ToObject()
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

// newDictItemIterator creates a dictItemIterator object for d.
func newDictItemIterator(f *Frame, d *Dict) *dictItemIterator {
	return &dictItemIterator{
		Object: Object{typ: dictItemIteratorType},
		iter:   newDictEntryIterator(f, d),
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

// newDictKeyIterator creates a dictKeyIterator object for d.
func newDictKeyIterator(f *Frame, d *Dict) *dictKeyIterator {
	return &dictKeyIterator{
		Object: Object{typ: dictKeyIteratorType},
		iter:   newDictEntryIterator(f, d),
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
	return entry.key, raised
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

// newDictValueIterator creates a dictValueIterator object for d.
func newDictValueIterator(f *Frame, d *Dict) *dictValueIterator {
	return &dictValueIterator{
		Object: Object{typ: dictValueIteratorType},
		iter:   newDictEntryIterator(f, d),
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
	return entry.value, raised
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

func dictIteratorNext(f *Frame, iter *dictEntryIterator, guard *dictVersionGuard) (entry dictEntry, raises *BaseException) {
	// NOTE: The behavior here diverges from CPython where an iterator that
	// is exhausted will always return StopIteration regardless whether the
	// underlying dict is subsequently modified. In Grumpy, an iterator for
	// a dict that has been modified will always raise RuntimeError even if
	// the iterator was exhausted before the modification.
	entry = iter.next()
	if !guard.check() {
		raises = f.RaiseType(RuntimeErrorType, "dictionary changed during iteration")
	} else if entry.isEmpty() {
		raises = f.Raise(StopIterationType.ToObject(), nil, nil)
	}
	return
}

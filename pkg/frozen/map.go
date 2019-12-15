package frozen

import (
	"fmt"

	"github.com/marcelocantos/frozen/pkg/value"
)

// KeyValue represents a key-value pair for insertion into a Map.
type KeyValue struct {
	Key, Value interface{}
}

// KV creates a KeyValue.
func KV(key, val interface{}) KeyValue {
	return KeyValue{Key: key, Value: val}
}

// Hash computes a hash for a KeyValue.
func (kv KeyValue) Hash() uint64 {
	return value.Hash(kv.Key)
}

// Equal returns true iff i is a KeyValue whose key equals this KeyValue's key.
func (kv KeyValue) Equal(i interface{}) bool {
	if kv2, ok := i.(KeyValue); ok {
		return value.Equal(kv.Key, kv2.Key)
	}
	return false
}

// String returns a string representation of a KeyValue.
func (kv KeyValue) String() string {
	return fmt.Sprintf("%#v:%#v", kv.Key, kv.Value)
}

// Map maps keys to values. The zero value is the empty Map.
type Map struct {
	root  *node
	count int
}

var _ value.Key = Set{}

// NewMap create a new Map with kvs as keys and values.
func NewMap(kvs ...KeyValue) Map {
	var b MapBuilder
	for _, kv := range kvs {
		b.Put(kv.Key, kv.Value)
	}
	return b.Finish()
}

// IsEmpty returns true if the Map has no entries.
func (m Map) IsEmpty() bool {
	return m.root == nil
}

// Count returns the number of entries in the Map.
func (m Map) Count() int {
	return m.count
}

// With returns a new Map with key associated with val and all other keys
// retained from m.
func (m Map) With(key, val interface{}) Map {
	c := newUnionComposer(m.Count() + 1)
	root := m.root.apply(c, KV(key, val))
	return Map{root: root, count: c.count()}
}

// Without returns a new Map with all keys retained from m except the elements
// of keys.
func (m Map) Without(keys Set) Map {
	c := newMinusComposer(m.Count())
	root := m.root
	for k := keys.Range(); k.Next(); {
		root = root.apply(c, KV(k.Value(), nil))
	}
	return Map{root: root, count: c.count()}
}

// Has returns true iff the key exists in the map.
func (m Map) Has(key interface{}) bool {
	return m.root.get(KV(key, nil)) != nil
}

// Get returns the value associated with key in m and true iff the key is found.
func (m Map) Get(key interface{}) (interface{}, bool) {
	if kv := m.root.get(KV(key, nil)); kv != nil {
		return kv.(KeyValue).Value, true
	}
	return nil, false
}

// MustGet returns the value associated with key in m or panics if the key is
// not found.
func (m Map) MustGet(key interface{}) interface{} {
	if val, has := m.Get(key); has {
		return val
	}
	panic(fmt.Sprintf("key not found: %v", key))
}

// GetElse returns the value associated with key in m or deflt if the key is not
// found.
func (m Map) GetElse(key, deflt interface{}) interface{} {
	if val, has := m.Get(key); has {
		return val
	}
	return deflt
}

// GetElseFunc returns the value associated with key in m or the result of
// calling deflt if the key is not found.
func (m Map) GetElseFunc(key interface{}, deflt func() interface{}) interface{} {
	if val, has := m.Get(key); has {
		return val
	}
	return deflt()
}

// Keys returns a Set with all the keys in the Map.
func (m Map) Keys() Set {
	return m.Reduce(func(acc, key, _ interface{}) interface{} {
		return acc.(Set).With(key)
	}, Set{}).(Set)
}

// Values returns a Set with all the Values in the Map.
func (m Map) Values() Set {
	return m.Reduce(func(acc, _, val interface{}) interface{} {
		return acc.(Set).With(val)
	}, Set{}).(Set)
}

// Project returns a Map with only keys included from this Map.
func (m Map) Project(keys Set) Map {
	return m.Where(func(key, val interface{}) bool {
		return keys.Has(key)
	})
}

// Where returns a Map with only key-value pairs satisfying pred.
func (m Map) Where(pred func(key, val interface{}) bool) Map {
	return m.Reduce(func(acc, key, val interface{}) interface{} {
		if pred(key, val) {
			return acc.(Map).With(key, val)
		}
		return acc
	}, NewMap()).(Map)
}

// Map returns a Map with keys from this Map, but the values replaced by the
// result of calling f.
func (m Map) Map(f func(key, val interface{}) interface{}) Map {
	return m.Reduce(func(acc, key, val interface{}) interface{} {
		return acc.(Map).With(key, f(key, val))
	}, NewMap()).(Map)
}

// Reduce returns the result of applying f to each key-value pair on the Map.
// The result of each call is used as the acc argument for the next element.
func (m Map) Reduce(f func(acc, key, val interface{}) interface{}, acc interface{}) interface{} {
	for i := m.Range(); i.Next(); {
		acc = f(acc, i.Key(), i.Value())
	}
	return acc
}

// Update returns a Map with key-value pairs from n added or replacing existing
// keys.
func (m Map) Update(n Map) Map {
	return m.merge(n, useRight)
}

// Merge returns a Map with key-value pairs from n merged into m by calling
// compose to get new values.
func (m Map) Merge(n Map, compose func(key, a, b interface{}) interface{}) Map {
	return m.merge(n, func(a, b interface{}) interface{} {
		kva := a.(KeyValue)
		kvb := b.(KeyValue)
		if c := compose(kva.Key, kva.Value, kvb.Value); c != nil {
			return KV(kva.Key, c)
		}
		return nil
	})
}

func (m Map) merge(n Map, compose func(a, b interface{}) interface{}) Map {
	c := newUnionComposer(m.Count() + n.Count())
	c.compose = compose
	root := m.root.merge(n.root, c)
	return Map{root: root, count: c.count()}
}

// Hash computes a hash val for s.
func (m Map) Hash() uint64 {
	var h uint64 = 3167960924819262823
	for i := m.Range(); i.Next(); {
		h ^= 12012876008735959943*value.Hash(i.Key()) + value.Hash(i.Value())
	}
	return h
}

// Equal returns true iff i is a Map with all the same key-value pairs as this
// Map.
func (m Map) Equal(i interface{}) bool {
	if n, ok := i.(Map); ok {
		return m.root.equal(n.root, func(a, b interface{}) bool {
			kva := a.(KeyValue)
			kvb := b.(KeyValue)
			return value.Equal(kva.Key, kvb.Key) && value.Equal(kva.Value, kvb.Value)
		})
	}
	return false
}

// String returns a string representatio of the Map.
func (m Map) String() string {
	return fmt.Sprintf("%v", m)
}

// Format writes a string representation of the Map into state.
func (m Map) Format(state fmt.State, _ rune) {
	state.Write([]byte("("))
	for i, n := m.Range(), 0; i.Next(); n++ {
		if n > 0 {
			state.Write([]byte(", "))
		}
		fmt.Fprintf(state, "%v: %v", i.Key(), i.Value())
	}
	state.Write([]byte(")"))
}

// Range returns a MapIterator over the Map.
func (m Map) Range() *MapIterator {
	return &MapIterator{i: m.root.iterator()}
}

// MapIterator provides for iterating over a Map.
type MapIterator struct {
	i  *nodeIter
	kv KeyValue
}

// Next moves to the next key-value pair or returns false if there are no more.
func (i *MapIterator) Next() bool {
	if i.i.next() {
		var ok bool
		i.kv, ok = i.i.elem.(KeyValue)
		if !ok {
			panic(fmt.Sprintf("Unexpected type: %T", i.i.elem))
		}
		return true
	}
	return false
}

// Key returns the key for the current entry.
func (i *MapIterator) Key() interface{} {
	return i.kv.Key
}

// Value returns the value for the current entry.
func (i *MapIterator) Value() interface{} {
	return i.kv.Value
}

// Item returns the current key-value pair as two return values.
func (i *MapIterator) Item() (key, value interface{}) {
	return i.kv.Key, i.kv.Value
}

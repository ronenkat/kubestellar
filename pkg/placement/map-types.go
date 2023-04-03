/*
Copyright 2023 The KCP Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package placement

// Map is a finite set of (key,value) pairs
// that has at most one value for any given key.
// The collection may or may not be mutable.
// This view of the collection may or may not have a limited scope of validity.
// This view may or may not have concurrency restrictions.
type Map[Key comparable, Val any] interface {
	Emptyable
	Len() int
	LenIsCheap() bool
	Get(Key) (Val, bool)
	Visitable[Pair[Key, Val]]
}

// MutableMap is a Map that can be written to.
type MutableMap[Key comparable, Val any] interface {
	Map[Key, Val]
	MappingReceiver[Key, Val]
}

// MappingReceiver is something that can be given key/value pairs.
// This is the writable aspect of a Map.
// Some DynamicMapProvider implementations require receivers to be comparable.
type MappingReceiver[Key comparable, Val any] interface {
	Put(Key, Val)
	Delete(Key)
}

// MappingReceiverFuncs is a convenient constructor of MappingReceiver from two funcs
type MappingReceiverFuncs[Key comparable, Val any] struct {
	OnPut    func(Key, Val)
	OnDelete func(Key)
}

var _ MappingReceiver[float32, map[string]func()] = MappingReceiverFuncs[float32, map[string]func()]{}
var _ MapChangeReceiver[float32, map[string]func()] = MappingReceiverFuncs[float32, map[string]func()]{}

func (mrf MappingReceiverFuncs[Key, Val]) Put(key Key, val Val) {
	if mrf.OnPut != nil {
		mrf.OnPut(key, val)
	}
}

func (mrf MappingReceiverFuncs[Key, Val]) Delete(key Key) {
	if mrf.OnDelete != nil {
		mrf.OnDelete(key)
	}
}

func (mrf MappingReceiverFuncs[Key, Val]) Create(key Key, val Val) {
	if mrf.OnPut != nil {
		mrf.OnPut(key, val)
	}
}

func (mrf MappingReceiverFuncs[Key, Val]) Update(key Key, oldVal, newVal Val) {
	if mrf.OnPut != nil {
		mrf.OnPut(key, newVal)
	}
}

func (mrf MappingReceiverFuncs[Key, Val]) DeleteWithFinal(key Key, oldVal Val) {
	if mrf.OnDelete != nil {
		mrf.OnDelete(key)
	}
}

type MappingReceiverFork[Key comparable, Val any] []MappingReceiver[Key, Val]

var _ MappingReceiver[int, func()] = MappingReceiverFork[int, func()]{}

func (mrf MappingReceiverFork[Key, Val]) Put(key Key, val Val) {
	for _, mr := range mrf {
		mr.Put(key, val)
	}
}

func (mrf MappingReceiverFork[Key, Val]) Delete(key Key) {
	for _, mr := range mrf {
		mr.Delete(key)
	}
}

// MapChangeReceiver is what a stateful map offers to an observer
type MapChangeReceiver[Key comparable, Val any] interface {
	Create(Key, Val)

	// Update is given key, old value, new value
	Update(Key, Val, Val)

	// DeleteWithFinal is given key and last value
	DeleteWithFinal(Key, Val)
}

// MapChangeReceiverFuncs is a convenient constructor of MapChangeReceiver from three funcs
type MapChangeReceiverFuncs[Key comparable, Val any] struct {
	OnCreate func(Key, Val)
	OnUpdate func(Key, Val, Val)
	OnDelete func(Key, Val)
}

var _ MapChangeReceiver[string, func()] = MapChangeReceiverFuncs[string, func()]{}

func (mrf MapChangeReceiverFuncs[Key, Val]) Create(key Key, val Val) {
	if mrf.OnCreate != nil {
		mrf.OnCreate(key, val)
	}
}

func (mrf MapChangeReceiverFuncs[Key, Val]) Update(key Key, oldVal, newVal Val) {
	if mrf.OnUpdate != nil {
		mrf.OnUpdate(key, oldVal, newVal)
	}
}

func (mrf MapChangeReceiverFuncs[Key, Val]) DeleteWithFinal(key Key, val Val) {
	if mrf.OnDelete != nil {
		mrf.OnDelete(key, val)
	}
}

type MapChangeReceiverFork[Key comparable, Val any] []MapChangeReceiver[Key, Val]

var _ MapChangeReceiver[int, func()] = MapChangeReceiverFork[int, func()]{}

func (mrf MapChangeReceiverFork[Key, Val]) Create(key Key, val Val) {
	for _, mr := range mrf {
		mr.Create(key, val)
	}
}

func (mrf MapChangeReceiverFork[Key, Val]) Update(key Key, oldVal, newVal Val) {
	for _, mr := range mrf {
		mr.Update(key, oldVal, newVal)
	}
}

func (mrf MapChangeReceiverFork[Key, Val]) DeleteWithFinal(key Key, val Val) {
	for _, mr := range mrf {
		mr.DeleteWithFinal(key, val)
	}
}

// MappingReceiverDiscardsPrevious produces a MapChangeReceiver that dumbs down its info to pass along to the given MappingReceiver
func MappingReceiverDiscardsPrevious[Key comparable, Val any](mr MappingReceiver[Key, Val]) MapChangeReceiver[Key, Val] {
	return mappingReceiverDiscardsPrevious[Key, Val]{inner: mr}
}

type mappingReceiverDiscardsPrevious[Key comparable, Val any] struct{ inner MappingReceiver[Key, Val] }

func (mr mappingReceiverDiscardsPrevious[Key, Val]) Create(key Key, val Val) {
	mr.inner.Put(key, val)
}

func (mr mappingReceiverDiscardsPrevious[Key, Val]) Update(key Key, oldVal, newVal Val) {
	mr.inner.Put(key, newVal)
}

func (mr mappingReceiverDiscardsPrevious[Key, Val]) DeleteWithFinal(key Key, val Val) {
	mr.inner.Delete(key)
}

// TransactionalMappingReceiver is one that takes updates in batches
type TransactionalMappingReceiver[Key comparable, Val any] interface {
	Transact(func(MappingReceiver[Key, Val]))
}

// MapReadonly returns a version of the argument that does not support writes
func MapReadonly[Key comparable, Val any](inner Map[Key, Val]) Map[Key, Val] {
	return mapReadonly[Key, Val]{inner}
}

type mapReadonly[Key comparable, Val any] struct {
	Map[Key, Val]
}

func MutableMapWithKeyObserver[Key comparable, Val any](mm MutableMap[Key, Val], observer SetChangeReceiver[Key]) MutableMap[Key, Val] {
	return &mutableMapWithKeyObserver[Key, Val]{mm, observer}
}

type mutableMapWithKeyObserver[Key comparable, Val any] struct {
	MutableMap[Key, Val]
	observer SetChangeReceiver[Key]
}

func (mko *mutableMapWithKeyObserver[Key, Val]) Put(key Key, val Val) {
	mko.MutableMap.Put(key, val)
	mko.observer.Add(key)
}

func (mko *mutableMapWithKeyObserver[Key, Val]) Delete(key Key) {
	mko.MutableMap.Delete(key)
	mko.observer.Remove(key)
}

type TransformMappingReceiver[KeyOriginal, KeyTransformed comparable, ValOriginal, ValTransformed any] struct {
	TransformKey func(KeyOriginal) KeyTransformed
	TransformVal func(ValOriginal) ValTransformed
	Inner        MappingReceiver[KeyTransformed, ValTransformed]
}

var _ MappingReceiver[int, func()] = &TransformMappingReceiver[int, string, func(), []int]{}

func (xr TransformMappingReceiver[KeyOriginal, KeyTransformed, ValOriginal, ValTransformed]) Put(keyOriginal KeyOriginal, valOriginal ValOriginal) {
	keyTransformed := xr.TransformKey(keyOriginal)
	valTransformed := xr.TransformVal(valOriginal)
	xr.Inner.Put(keyTransformed, valTransformed)
}

func (xr TransformMappingReceiver[KeyOriginal, KeyTransformed, ValOriginal, ValTransformed]) Delete(keyOriginal KeyOriginal) {
	keyTransformed := xr.TransformKey(keyOriginal)
	xr.Inner.Delete(keyTransformed)
}

func MappingReceiverAsVisitor[Key comparable, Val any](receiver MappingReceiver[Key, Val]) func(Pair[Key, Val]) error {
	return func(tup Pair[Key, Val]) error {
		receiver.Put(tup.First, tup.Second)
		return nil
	}
}

func MappingReceiverNegativeAsVisitor[Key comparable, Val any](receiver MappingReceiver[Key, Val]) func(Pair[Key, Val]) error {
	return func(tup Pair[Key, Val]) error {
		receiver.Delete(tup.First)
		return nil
	}
}

func MapApply[Key comparable, Val any](theMap Map[Key, Val], receiver MappingReceiver[Key, Val]) {
	theMap.Visit(MappingReceiverAsVisitor(receiver))
}

func MapAddAll[Key comparable, Val any](theMap MutableMap[Key, Val], adds Visitable[Pair[Key, Val]]) {
	adds.Visit(func(add Pair[Key, Val]) error {
		theMap.Put(add.First, add.Second)
		return nil
	})
}

func MapRemoveAll[Key comparable, Val any](theMap MutableMap[Key, Val], goners Visitable[Pair[Key, Val]]) {
	goners.Visit(func(goner Pair[Key, Val]) error {
		theMap.Delete(goner.First)
		return nil
	})
}

// MapGetAdd does a Get and an add if the sought mmapping is missing and desired.
// If the sought mapping is missing and undesired then the result is the zero value of Val.
func MapGetAdd[Key comparable, Val any](theMap MutableMap[Key, Val], key Key, want bool, valGenerator func(Key) Val) Val {
	val, have := theMap.Get(key)
	if have {
		return val
	}
	if want {
		val = valGenerator(key)
		theMap.Put(key, val)
		return val
	}
	var zero Val
	return zero
}

func MapEqual[Key, Val comparable](left, right Map[Key, Val]) bool {
	if left.Len() != right.Len() {
		return false
	}
	if left.Visit(func(tup Pair[Key, Val]) error {
		valRight, has := right.Get(tup.First)
		if !has || tup.Second != valRight {
			return errStop
		}
		return nil
	}) != nil {
		return false
	}
	return true
}

func MapEnumerateDifferences[Key, Val comparable](left, right Map[Key, Val], receiver MapChangeReceiver[Key, Val]) {
	left.Visit(func(tup Pair[Key, Val]) error {
		valRight, has := right.Get(tup.First)
		if !has {
			receiver.DeleteWithFinal(tup.First, tup.Second)
		} else if valRight != tup.Second {
			receiver.Update(tup.First, tup.Second, valRight)
		}
		return nil
	})
	right.Visit(func(tup Pair[Key, Val]) error {
		valLeft, has := left.Get(tup.First)
		if !has {
			receiver.Create(tup.First, tup.Second)
		} else if valLeft != tup.Second {
			receiver.Update(tup.First, valLeft, tup.Second)
		}
		return nil
	})
}

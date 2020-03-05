// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package memory implements the cache.Interface and registers a memory provider.
// All operations are using a sync.RWMutex for synchronization.
package memory

import (
	"errors"
	"fmt"
	cm "github.com/patrickascher/gofw/cache"
	"sync"
	"time"
)

// init register the memory provider.
func init() {
	_ = cm.Register(cm.MEMORY, newMemory)
}

// defaultGCSleepDuration holds the default gc loop waiting time in seconds
var defaultGCSleepDuration = 60

var ErrKeyNotExist = errors.New("cache/memory: key #%v does not exist")

// Memory cache backend
type Memory struct {
	mutex   sync.RWMutex
	options Options
	items   map[string]cm.Valuer
}

// Options for memory provider
type Options struct {
	LoopDuration time.Duration
}

// item implements the Valuer interface
type item struct {
	val     interface{}   //value
	ttl     time.Duration //lifetime
	created time.Time     //time when the value was set
}

// Value returns the value of the item.
func (m *item) Value() interface{} {
	return m.val
}

// expired returns a bool if the value is expired.
func (m *item) expired() bool {
	if m.ttl == cm.INFINITY {
		return false
	}
	return time.Now().Sub(m.created) > m.ttl
}

// newMemory returns a new memory cache by the given options.
func newMemory(opt interface{}) cm.Interface {
	options := Options{LoopDuration: time.Duration(defaultGCSleepDuration) * time.Second}
	if opt != nil {
		options = opt.(Options)
		//TODO create a nicer version
		if options.LoopDuration < 1 {
			options.LoopDuration = time.Duration(defaultGCSleepDuration) * time.Second
		}
	}

	return &Memory{options: options, items: make(map[string]cm.Valuer)}
}

// Get returns the value of the given key.
// Error will return if the key does not exist.
func (m *Memory) Get(key string) (cm.Valuer, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if val, ok := m.items[key]; ok {
		return val, nil
	}
	return nil, fmt.Errorf(ErrKeyNotExist.Error(), key)
}

// GetAll returns all items of the cache as map.
func (m *Memory) GetAll() map[string]cm.Valuer {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.items
}

// Set key/value pair.
// The ttl can be set by duration or forever with cache.INFINITY.
func (m *Memory) Set(key string, value interface{}, ttl time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.items[key] = &item{val: value, created: time.Now(), ttl: ttl}

	return nil
}

// Exist returns true if the key exists.
func (m *Memory) Exist(key string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if _, ok := m.items[key]; ok {
		return true
	}

	return false
}

// Delete removes a given key from the cache.
// Error will return if the key does not exist.
func (m *Memory) Delete(key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.items[key]; !ok {
		return fmt.Errorf(ErrKeyNotExist.Error(), key)
	}

	delete(m.items, key)

	return nil
}

// DeleteAll removes all items from the cache.
func (m *Memory) DeleteAll() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.items = make(map[string]cm.Valuer)

	return nil
}

// GC is an infinity loop. The loop waits for the given time and runs again to delete all expired keys.
func (m *Memory) GC() {
	for {
		<-time.After(m.options.LoopDuration)
		if m.items == nil {
			return
		}
		if keys := m.expiredKeys(); len(keys) != 0 {
			for _, key := range keys {
				_ = m.Delete(key)
			}
		}
	}
}

// expiredKeys returns all expired keys.
func (m *Memory) expiredKeys() (keys []string) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for key, itm := range m.items {
		x := itm.(*item)
		if x.expired() {
			keys = append(keys, key)
		}
	}
	return keys
}

// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package memory implements the cache.Interface and registers an in-memory provider.
// All operations are using a sync.RWMutex for synchronization.
// Benchmark file is available.
package memory

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	cm "github.com/patrickascher/gofw/cache"
)

// init register the in-memory provider.
func init() {
	_ = cm.Register(cm.MEMORY, New)
}

// defaultGCInterval holds the garbage collector waiting time in seconds.
var defaultGCInterval = 60

// Error messages
var (
	ErrKeyNotExist = errors.New("cache/memory: key %v does not exist")
)

// memory cache provider
type memory struct {
	mutex   sync.RWMutex
	options Options
	items   map[string]cm.Valuer
}

// Options for the in-memory provider
type Options struct {
	GCInterval time.Duration
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

// New creates a in-memory cache by the given options.
func New(opt interface{}) cm.Interface {
	options := Options{GCInterval: time.Duration(defaultGCInterval) * time.Second}
	if opt != nil {
		//TODO use a merger like https://github.com/imdario/mergo?
		if opt.(Options).GCInterval > 0 {
			options = opt.(Options)
		}
	}
	return &memory{options: options, items: make(map[string]cm.Valuer)}
}

// Get returns the value of the given key.
// Error will return if the key does not exist.
func (m *memory) Get(key string) (cm.Valuer, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if val, ok := m.items[key]; ok {
		return val, nil
	}
	return nil, fmt.Errorf(ErrKeyNotExist.Error(), key)
}

// GetPrefixed returns all items of the cache as map.
func (m *memory) GetPrefixed(prefix string) map[string]cm.Valuer {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	rv := make(map[string]cm.Valuer, 0)
	for k, item := range m.items {
		if strings.HasPrefix(k, prefix) {
			rv[k] = item
		}
	}

	return rv
}

// GetAll returns all items of the cache as map.
func (m *memory) GetAll() map[string]cm.Valuer {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.items
}

// Set key/value pair.
// The ttl can be set by duration or forever with cache.INFINITY.
func (m *memory) Set(key string, value interface{}, ttl time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.items[key] = &item{val: value, created: time.Now(), ttl: ttl}

	return nil
}

// Exist returns true if the key exists.
func (m *memory) Exist(key string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	_, ok := m.items[key]
	return ok
}

// Delete removes a given key from the cache.
// Error will return if the key does not exist.
func (m *memory) Delete(key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.items[key]; !ok {
		return fmt.Errorf(ErrKeyNotExist.Error(), key)
	}

	delete(m.items, key)

	return nil
}

// DeleteAll removes all items from the cache.
func (m *memory) DeleteAll() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.items = make(map[string]cm.Valuer)

	return nil
}

func (m *memory) DeletePrefixed(prefix string) error {
	items := m.GetPrefixed(prefix)
	for k := range items {
		err := m.Delete(k)
		if err != nil {
			return err
		}
	}
	return nil
}

// GC is an infinity loop. The loop will rerun after an specific interval time which can be set
// in the options (default 60sec).
func (m *memory) GC() {
	for {
		<-time.After(m.options.GCInterval)
		if keys := m.expiredKeys(); len(keys) != 0 {
			for _, key := range keys {
				_ = m.Delete(key)
			}
		}
	}
}

// expiredKeys returns all expired keys.
func (m *memory) expiredKeys() (keys []string) {
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

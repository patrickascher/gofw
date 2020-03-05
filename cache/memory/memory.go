// Package memory implements a cache in memory backend
package memory

import (
	"fmt"
	"github.com/patrickascher/gofw/cache"
	"sync"
	"time"
)

// init register the memory provider.
func init() {
	_ = cache.Register("memory", NewMemoryCache)
}

var memoryCache *Memory

// Memory cache backend
type Memory struct {
	mutex     sync.RWMutex
	dur       time.Duration
	items     map[string]cache.Valuer
	gcSpawned bool
}

type item struct {
	val     interface{}   //value
	ttl     time.Duration //lifetime
	created time.Time     //time when the value was set
}

// Value returns the value of the entry.
func (m *item) Value() interface{} {
	return m.val
}

// Lifetime returns the Lifetime of the entry.
func (m *item) Lifetime() time.Duration {
	return m.ttl
}

// isExpire returns a bool if the value is expired.
func (m *item) isExpire() bool {
	// 0 means forever
	if m.ttl == 0 {
		return false
	}
	return time.Now().Sub(m.created) > m.ttl
}

// NewMemoryCache returns a new MemoryCache.
func NewMemoryCache() cache.Cache {
	if memoryCache == nil {
		memoryCache = &Memory{items: make(map[string]cache.Valuer)}
	} else {
		return memoryCache
	}
	//cache := Memory{items: make(map[string]cache.Item)}
	return memoryCache
}

// Get returns the value of the given key. If no key was found, nil will return.
func (m *Memory) Get(key string) (cache.Valuer, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if val, ok := m.items[key]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("cache-memory: key #%v does not exist", key)
}

// GetAll returns all items of the cache
func (m *Memory) GetAll() map[string]cache.Valuer {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.items
}

// Set sets the value by key. if the ttl is 0 its infinitely stored.
func (m *Memory) Set(key string, value interface{}, ttl time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.items[key] = &item{val: value, created: time.Now(), ttl: ttl}

	return nil
}

// Exist returns a bool if the key exists in the cache
func (m *Memory) Exist(key string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if _, ok := m.items[key]; ok {
		return true
	}

	return false
}

// Delete removes a given key from the cache.
// if the key does not exist or it wasn't removed, a error will return
func (m *Memory) Delete(key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.items[key]; !ok {
		return fmt.Errorf("cache-memory: key #%v does not exist", key)
	}

	delete(m.items, key)

	return nil
}

// DeleteAll removes all items from the cache
func (m *Memory) DeleteAll() error {
	m.mutex.Lock()
	m.items = make(map[string]cache.Valuer)
	m.mutex.Unlock()
	return nil
}

// GC is creating a garbage collector in a new goroutine
func (m *Memory) GC(duration time.Duration) error {
	go m.garbageCollector(duration)
	m.gcSpawned = true

	return nil
}

func (m *Memory) GCSpawned() bool {
	return m.gcSpawned
}

// garbageCollector is running a loop every x time. It removes all expired keys.
func (m *Memory) garbageCollector(duration time.Duration) {
	for {
		<-time.After(duration)
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

// expiredKeys returns all expired values by its key
func (m *Memory) expiredKeys() (keys []string) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for key, itm := range m.items {
		x := itm.(*item)
		if x.isExpire() {
			keys = append(keys, key)
		}
	}
	return keys
}

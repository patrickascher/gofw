// Package cache is a cache manager. Its provides a cache interface and is easy to extend.
// At the moment only a in-memory cache is build
// See https://github.com/patrickascher/go-router for more information and examples.
package cache

import (
	"errors"
	"fmt"
	"time"
)

// Error messages are defined here
var (
	ErrUnknownBackend       = errors.New("cache: unknown cache-backend %q (forgotten import?)")
	ErrNoBackend            = errors.New("cache: empty cache-name or cache is given")
	ErrBackendAlreadyExists = errors.New("cache: cache-backend %#v already exists")
)

// Cache interface for the cache backend
type Cache interface {
	Get(key string) (Item, error)
	GetAll() map[string]Item
	Set(key string, value interface{}, timeout time.Duration) error
	Exist(key string) bool
	Delete(key string) error
	DeleteAll() error
	GC(duration time.Duration) error
}

// Item interface for the stored values
type Item interface {
	Value() interface{}
	Lifetime() time.Duration
}

// Backend is a function of the cache interface.
// Like this the cache backend is getting initialized only when its called
type Backend func() Cache

// cacheStore for all cache backend
var cacheStore = make(map[string]Backend)

// Get creates a new cache backend if its existing otherwise it's returning an error
func Get(cacheName string, duration time.Duration) (Cache, error) {
	instance, ok := cacheStore[cacheName]
	if !ok {
		return nil, fmt.Errorf(ErrUnknownBackend.Error(), cacheName)
	}
	cacheBackend := instance()

	err := cacheBackend.GC(duration)

	if err != nil {
		cacheBackend = nil
	}
	return cacheBackend, nil
}

// Register the cache backend
// If the cache backend name is empty or already exists a error will return
func Register(cacheName string, cache Backend) error {
	if cache == nil || cacheName == "" {
		return ErrNoBackend
	}
	if _, exists := cacheStore[cacheName]; exists {
		return fmt.Errorf(ErrBackendAlreadyExists.Error(), cacheName)
	}
	cacheStore[cacheName] = cache
	return nil
}

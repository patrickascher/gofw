// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package cache is a cache manager. Its provides a cache interface and is easy to extend.
// See https://github.com/patrickascher/go-router for more information and examples.
package cache

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrUnknownProvider       = errors.New("cache: unknown cache-provider %q (forgotten import?)")
	ErrNoProvider            = errors.New("cache: empty cache-name or cache is given")
	ErrProviderAlreadyExists = errors.New("cache: cache-backend %#v already exists")
)

// store for all cache providers
var store = make(map[string]Backend)

// Cache is an interface used by cache providers.
type Cache interface {
	Get(key string) (Valuer, error)
	GetAll() map[string]Valuer

	Set(key string, value interface{}, ttl time.Duration) error

	Exist(key string) bool

	Delete(key string) error
	DeleteAll() error

	GC(duration time.Duration) error
	GCSpawned() bool
}

// Valuer is an interface used to get the value and lifetime of an cache object.
type Valuer interface {
	Value() interface{}
	Lifetime() time.Duration
}

// Backend is a function of the cache interface.
// Like this the cache backend is getting initialized only when its called
type Backend func() Cache

// Instance creates a new cache provider.
// If the provider is not registered an error will return.
func Instance(provider string, duration time.Duration) (Cache, error) {
	instanceFn, ok := store[provider]
	if !ok {
		return nil, fmt.Errorf(ErrUnknownProvider.Error(), provider)
	}
	instance := instanceFn()

	if !instance.GCSpawned() {
		err := instance.GC(duration)
		// TODO return err?
		if err != nil {
			instance = nil
		}
	}

	return instance, nil
}

// Register the cache provider.
// If the cache provider name is empty or already exist, an error will return
func Register(provider string, cache Backend) error {
	if cache == nil || provider == "" {
		return ErrNoProvider
	}
	if _, exists := store[provider]; exists {
		return fmt.Errorf(ErrProviderAlreadyExists.Error(), provider)
	}
	store[provider] = cache
	return nil
}

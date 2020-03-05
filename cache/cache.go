// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package cache provides a cache manager for any type that
// implements the cache.Interface.
package cache

import (
	"errors"
	"fmt"
	"time"
)

const (
	// List of the pre defined cache providers.
	MEMORY = "memory"
	// INFINITY must be used by the cache providers to identify
	// that the value should not get deleted by the garbage collector.
	INFINITY = 0
)

var (
	ErrUnknownProvider       = errors.New("cache: unknown cache-provider %q")
	ErrNoProvider            = errors.New("cache: empty cache-name or cache-provider is nil")
	ErrProviderAlreadyExists = errors.New("cache: cache-provider %#v is already registered")
)

// registry for all cache providers.
var registry = make(map[string]provider)

// Interface is used by cache providers.
type Interface interface {
	// Get returns a Valuer by its key.
	// If it does not exist, an error will return.
	Get(key string) (Valuer, error)
	// GetAll returns all existing Valuer as map.
	GetAll() map[string]Valuer
	// Set a value by its key and lifetime.
	// If a value should not get deleted, cache.INFINITY can be used as time.Duration.
	Set(key string, value interface{}, ttl time.Duration) error
	// Exist checks if a value is set by the given key.
	Exist(key string) bool
	// Delete a value by its key.
	Delete(key string) error
	// DeleteAll values of the cache provider.
	DeleteAll() error
	// GC will spawn the garbage collector in a goroutine.
	// If your cache provider has its own gc (redis, memcached, ...) just return void in this method.
	GC()
}

// Valuer is an interface to get the value of a cache object.
type Valuer interface {
	Value() interface{}
}

// provider is a function which returns the cache interface.
// As argument the provider options can passed.
// Like this the cache provider is getting initialized only when its called.
type provider func(interface{}) Interface

// New returns a specific cache provider by its name and given options.
// The available options are defined in the provider.
// If the provider is not registered an error will return.
func New(provider string, options interface{}) (Interface, error) {
	instanceFn, ok := registry[provider]
	if !ok {
		return nil, fmt.Errorf(ErrUnknownProvider.Error(), provider)
	}
	instance := instanceFn(options)

	// start garbage collector
	go instance.GC()

	return instance, nil
}

// Register the cache provider.
// If the cache provider name is empty or is already registered, an error will return.
func Register(provider string, fn provider) error {
	if fn == nil || provider == "" {
		return ErrNoProvider
	}
	if _, exists := registry[provider]; exists {
		return fmt.Errorf(ErrProviderAlreadyExists.Error(), provider)
	}
	registry[provider] = fn
	return nil
}

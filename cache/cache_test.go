// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cache_test

import (
	"fmt"
	"testing"
	"time"

	cm "github.com/patrickascher/gofw/cache"
	"github.com/stretchr/testify/assert"
)

var mockProvider cm.Interface

func newMock(opt interface{}) cm.Interface {
	mockProvider = &mockCache{options: opt.(options), items: make(map[string]cm.Valuer)}
	return mockProvider
}

type mockCache struct {
	options   options
	items     map[string]cm.Valuer
	gcCounter int
}

type options struct {
	filename string
}

type item struct {
	val     interface{}
	created time.Time
	ttl     time.Duration
}

func (i item) Value() interface{} {
	return "dummy"
}

func (mc *mockCache) Get(k string) (cm.Valuer, error) {
	item, ok := mc.items[k]
	if ok {
		return item, nil
	}
	return nil, fmt.Errorf("cache/mock: %v does not exist", k)
}

func (mc *mockCache) GetAll() map[string]cm.Valuer {
	return mc.items
}

func (mc *mockCache) Set(k string, v interface{}, ttl time.Duration) error {
	mc.items[k] = &item{val: v, created: time.Now(), ttl: ttl}
	return nil
}

func (mc *mockCache) Exist(k string) bool {
	_, ok := mc.items[k]
	return ok
}

func (mc *mockCache) Delete(k string) error {
	if _, ok := mc.items[k]; !ok {
		return fmt.Errorf("cache/mock: %v does not exist", k)
	}

	delete(mc.items, k)
	return nil
}

func (mc *mockCache) DeleteAll() error {
	mc.items = make(map[string]cm.Valuer)
	return nil
}

func (mc *mockCache) GC() {
	mc.gcCounter = mc.gcCounter + 1
}

func TestRegister(t *testing.T) {
	test := assert.New(t)

	// error: no provider-name and provider is given
	err := cm.Register("", nil)
	test.Error(err)
	test.Equal(err.Error(), cm.ErrNoProvider.Error())

	// error: no provider is given
	err = cm.Register("mock", nil)
	test.Error(err)
	test.Equal(err.Error(), cm.ErrNoProvider.Error())

	// error: no provider-name is given
	err = cm.Register("", newMock)
	test.Error(err)
	test.Equal(err.Error(), cm.ErrNoProvider.Error())

	// ok: register successful
	err = cm.Register("mock", newMock)
	test.NoError(err)

	// error: multiple registration
	err = cm.Register("mock", newMock)
	test.Error(err)
	test.Equal(fmt.Sprintf(cm.ErrProviderAlreadyExists.Error(), "mock"), err.Error())
}

func TestNew(t *testing.T) {
	test := assert.New(t)

	// error: no registered dummy cache provider
	cache, err := cm.New("mock2", nil)
	test.Nil(cache)
	test.Error(err)
	test.Equal(fmt.Sprintf(cm.ErrUnknownProvider.Error(), "mock2"), err.Error())

	// ok
	cache, err = cm.New("mock", options{filename: "mock.txt"})
	test.NoError(err)
	test.NotNil(cache)
	test.Equal("mock.txt", mockProvider.(*mockCache).options.filename)
	// TODO checking if GC was called only once
}

// This example demonstrate the basics of the cache interface.
// For more details check the documentation.
func Example() {
	// import the provider package
	// import _ "github.com/patrickascher/gofw/cache/memory"

	// Initialize cache. Each call is creating a new cache instance.
	// The gc will be spawned in the background.
	c, err := cm.New(cm.MEMORY, nil)
	if err != nil {
		return
	}

	// Set a cache item for 5 hours.
	err = c.Set("foo", "bar", 5*time.Hour)

	// Set a cache item infinity
	err = c.Set("John", "Doe", cm.INFINITY)

	// Get a cache by key.
	item, err := c.Get("foo")
	if err != nil && item != nil {
		item.Value() // value of the item "bar"
	}

	// Get all items as map
	items := c.GetAll()
	fmt.Println(items)

	// Check if an key exists
	exists := c.Exist("foo")
	fmt.Println(exists)

	// Delete by key
	err = c.Delete("foo")

	// Delete all items
	err = c.DeleteAll()
}

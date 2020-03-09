// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cache_test

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"fmt"
	cm "github.com/patrickascher/gofw/cache"

	"time"
)

var dummyProvider cm.Interface

func newDummy(opt interface{}) cm.Interface {
	dummyProvider = &dummy{options: opt.(options), items: make(map[string]cm.Valuer)}
	return dummyProvider
}

type dummy struct {
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

func (d *dummy) Get(k string) (cm.Valuer, error) {
	item, ok := d.items[k]
	if ok {
		return item, nil
	}
	return nil, fmt.Errorf("cache/dummy: %v does not exist", k)
}

func (d *dummy) GetAll() map[string]cm.Valuer {
	return d.items
}

func (d *dummy) Set(k string, v interface{}, ttl time.Duration) error {
	d.items[k] = &item{val: v, created: time.Now(), ttl: ttl}
	return nil
}

func (d *dummy) Exist(k string) bool {
	_, ok := d.items[k]
	return ok
}

func (d *dummy) Delete(k string) error {
	if _, ok := d.items[k]; !ok {
		return fmt.Errorf("cache/dummy: %v does not exist", k)
	}

	delete(d.items, k)
	return nil
}

func (d *dummy) DeleteAll() error {
	d.items = make(map[string]cm.Valuer)
	return nil
}

func (d *dummy) GC() {
	d.gcCounter = d.gcCounter + 1
}

func TestRegister(t *testing.T) {
	test := assert.New(t)

	// error: no provider-name and provider is given
	err := cm.Register("", nil)
	test.Error(err)
	test.Equal(err.Error(), cm.ErrNoProvider.Error())

	// error: no provider is given
	err = cm.Register("dummy", nil)
	test.Error(err)
	test.Equal(err.Error(), cm.ErrNoProvider.Error())

	// error: no provider-name is given
	err = cm.Register("", newDummy)
	test.Error(err)
	test.Equal(err.Error(), cm.ErrNoProvider.Error())

	// ok: register successful
	err = cm.Register("dummy", newDummy)
	test.NoError(err)

	// error: multiple registration
	err = cm.Register("dummy", newDummy)
	test.Error(err)
	test.Equal(fmt.Sprintf(cm.ErrProviderAlreadyExists.Error(), "dummy"), err.Error())
}

func TestNew(t *testing.T) {
	test := assert.New(t)

	// error: no registered dummy cache provider
	cache, err := cm.New("dummy2", nil)
	test.Nil(cache)
	test.Error(err)
	test.Equal(fmt.Sprintf(cm.ErrUnknownProvider.Error(), "dummy2"), err.Error())

	// ok
	cache, err = cm.New("dummy", options{filename: "dummy.txt"})
	test.NoError(err)
	test.NotNil(cache)
	test.Equal("dummy.txt", dummyProvider.(*dummy).options.filename)
	// TODO checking if GC was called only once
}

// This example demonstrate the basics of the cache interface.
// For more details check the documentation.
func Example_new() {
	// Initialize cache. Each call is creating a new cache instance.
	c, err := cm.New("memory", nil)
	if err != nil {
		// ...
	}

	// Set a cache item for 5 hours.
	err = c.Set("foo", "bar", 5*time.Hour)
	// Set a cache item infinity
	err = c.Set("John", "Doe", cm.INFINITY)

	// Get a cache by key.
	item, err := c.Get("foo")
	if err != nil {
		_ = item.Value() // value of the item "bar"
	}
	// Get all items as map
	items := c.GetAll()
	fmt.Println(items)

	// Delete if exists
	if c.Exist("foo") {
		err = c.Delete("foo")
	}

	// Delete all items
	err = c.DeleteAll()
}

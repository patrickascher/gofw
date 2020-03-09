// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package memory_test

import (
	"fmt"
	cm "github.com/patrickascher/gofw/cache"
	"testing"
	"time"

	"github.com/patrickascher/gofw/cache/memory"
	"github.com/stretchr/testify/assert"
)

var mem cm.Interface

func init() {
	mem = memory.New(memory.Options{GCInterval: 1})
	go mem.GC()
}

func TestMemory_Set(t *testing.T) {
	// ok
	err := mem.Set("foo", "bar", cm.INFINITY)
	assert.NoError(t, err)

	// ok: redefine
	err = mem.Set("foo", "BAR", cm.INFINITY)
	assert.NoError(t, err)

	// ok
	err = mem.Set("John", "Doe", cm.INFINITY)
	assert.NoError(t, err)
}

func TestMemory_Get(t *testing.T) {
	// ok
	v, err := mem.Get("foo")
	assert.NoError(t, err)
	assert.Equal(t, "BAR", v.Value())

	// error: key does not exist
	v, err = mem.Get("baz")
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf(memory.ErrKeyNotExist.Error(), "baz"), err.Error())
	assert.Nil(t, v)
}

func TestMemory_GetAll(t *testing.T) {
	// ok
	v := mem.GetAll()
	assert.Equal(t, 2, len(v))
	assert.Equal(t, "BAR", v["foo"].Value())
	assert.Equal(t, "Doe", v["John"].Value())
}

func TestMemory_Exist(t *testing.T) {
	// ok
	v := mem.Exist("foo")
	assert.True(t, v)
	v = mem.Exist("baz")
	assert.False(t, v)
}

func TestMemory_GC(t *testing.T) {
	//ok
	err := mem.Set("gc", "val", 500*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(mem.GetAll()))
	time.Sleep(1 * time.Second)
	assert.Equal(t, 2, len(mem.GetAll()))
}

func TestMemory_Delete(t *testing.T) {
	// ok
	err := mem.Delete("John")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(mem.GetAll()))

	// error: key does not exist
	err = mem.Delete("baz")
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf(memory.ErrKeyNotExist.Error(), "baz"), err.Error())
}

func TestMemory_DeleteAll(t *testing.T) {
	// ok
	err := mem.DeleteAll()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(mem.GetAll()))

	// ok - delete with no entries
	err = mem.DeleteAll()
	assert.NoError(t, err)
}

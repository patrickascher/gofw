package memory_test

import (
	"github.com/patrickascher/gofw/cache/memory"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewMemoryCache(t *testing.T) {
	cache := memory.NewMemoryCache
	c := cache()

	// set and get test
	c.Set("test", "ABC", 5*time.Second)
	item, err := c.Get("test")
	assert.NoError(t, err)
	assert.Equal(t, "ABC", item.Value())
	assert.Equal(t, 5*time.Second, item.Lifetime())

	// setting multiple keys and check it
	c.Set("test", "CBA", 5*time.Second)
	c.Set("test2", "ABC", 5*time.Second)
	allItems := c.GetAll()
	assert.Equal(t, 2, len(allItems))
	assert.Equal(t, "CBA", allItems["test"].Value())
	assert.Equal(t, "ABC", allItems["test2"].Value())

	// test non existing keys
	val, err := c.Get("notExisting")
	assert.Equal(t, nil, val)
	assert.Error(t, err)
	assert.Equal(t, false, c.Exist("notExisting"))
	assert.Equal(t, true, c.Exist("test"))

	// delete one key
	assert.Equal(t, nil, c.Delete("test"))
	assert.Equal(t, 1, len(allItems))
	assert.Equal(t, "ABC", allItems["test2"].Value())

	// delete non existing key
	err = c.Delete("notEsssxisting")
	assert.Error(t, err)

	// delete all keys
	err = c.DeleteAll()
	assert.NoError(t, err)
	allItems = c.GetAll()
	assert.Equal(t, 0, len(allItems))
}

// TestMemoryCache_GC testing if the garbage collector is working correctly
func TestMemoryCache_GC(t *testing.T) {
	cache := memory.NewMemoryCache
	c := cache()

	c.Set("inf", "inity", 0)
	c.Set("cached", 1, 1500*time.Millisecond)
	item, err := c.Get("cached")
	assert.NoError(t, err)
	assert.Equal(t, 1, item.Value())
	assert.Equal(t, 1500*time.Millisecond, item.Lifetime())

	c.GC(500 * time.Millisecond)

	//not hitting
	time.Sleep(1 * time.Second)
	assert.Equal(t, 1, item.Value())
	val, err := c.Get("inf")
	assert.NoError(t, err)
	assert.Equal(t, "inity", val.Value())

	//key cached should be deleted
	time.Sleep(1 * time.Second)
	val, err = c.Get("cached")
	assert.Error(t, err)
	assert.Equal(t, nil, val)

	val, err = c.Get("inf")
	assert.NoError(t, err)
	assert.Equal(t, "inity", val.Value())
}

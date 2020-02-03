package cache_test

import (
	_ "github.com/patrickascher/gofw/cache/memory"
	"testing"

	"fmt"
	"github.com/patrickascher/gofw/cache"
	"github.com/patrickascher/gofw/cache/memory"
	"github.com/stretchr/testify/assert"
	"reflect"
	"time"
)

func TestGet(t *testing.T) {
	mem, err := cache.Get("memory", 60*time.Second)
	assert.NoError(t, err)
	assert.Equal(t, "*memory.Memory", reflect.ValueOf(mem).Type().String())
}

func TestRegister(t *testing.T) {
	_, err := cache.Get("redis", 10)
	assert.Equal(t, fmt.Errorf(cache.ErrUnknownBackend.Error(), "redis"), err)

	//empty cache-backend
	err = cache.Register("redis", nil)
	assert.Equal(t, cache.ErrNoBackend, err)

	//empty cache-name
	err = cache.Register("", memory.NewMemoryCache)
	assert.Equal(t, cache.ErrNoBackend, err)

	//already exists - tests also if the register worked
	err = cache.Register("memory", memory.NewMemoryCache)
	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf(cache.ErrBackendAlreadyExists.Error(), "memory"), err)
}

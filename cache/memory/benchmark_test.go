// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package memory_test

import (
	"github.com/patrickascher/gofw/cache"
	"log"
	"strconv"
	"testing"
)

var memCache cache.Interface

func init() {
	mem, err := cache.New("memory", nil)
	if err != nil {
		log.Fatal(err)
	}
	memCache = mem
}

func BenchmarkMemory_Set(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = memCache.Set("key:"+strconv.Itoa(i), i, cache.INFINITY)
	}
}

func BenchmarkMemory_Get(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = memCache.Get("key:" + strconv.Itoa(i))
	}
}

func BenchmarkMemory_Exist(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = memCache.Exist("key:" + strconv.Itoa(i))
	}
}

func BenchmarkMemory_Delete(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = memCache.Delete("key:" + strconv.Itoa(i))
	}
}

func BenchmarkMemory_DeleteAll(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = memCache.DeleteAll()
	}
}

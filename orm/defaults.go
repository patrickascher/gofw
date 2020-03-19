package orm

import (
	"github.com/jinzhu/inflection"
	"github.com/patrickascher/gofw/cache"
	"github.com/patrickascher/gofw/server"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/serenize/snaker"
	"time"
)

// TableName returns a camelCase pluralization of the struct name.
func (m *Model) TableName() string {
	return snaker.CamelToSnake(inflection.Plural(structName(m.caller, false)))
}

// DatabaseName will be empty by default.
func (m *Model) DatabaseName() string {
	return ""
}

// DefaultCache is defining a in-memory cache with a ttl of 6 hours by default.
func (m *Model) DefaultCache() (cache.Cache, time.Duration, error) {
	if server.Cache() != nil {
		return server.Cache(), 0, nil
	}

	c, err := cache.Get("memory", 5*time.Second)
	return c, 6 * time.Hour, err
}

// Builder returns the GlobalBuilder.
// If it's not defined, a error will return.
func (m *Model) Builder() (*sqlquery_.Builder, error) {
	if GlobalBuilder == nil {
		return nil, ErrModelNoBuilder
	}
	return GlobalBuilder, nil
}

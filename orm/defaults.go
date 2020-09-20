package orm

import (
	"errors"
	"github.com/patrickascher/gofw/server"
	"time"

	"github.com/patrickascher/gofw/cache"
	"github.com/patrickascher/gofw/logger"
	"github.com/patrickascher/gofw/sqlquery"
	fwStrings "github.com/patrickascher/gofw/strings"
)

// is used for self referencing m2m relations.
const defaultSelfReferenceAssociationForeignKey = "child_id"

// SoftDelete returns the configured schema of the builder.
// if the activeValue is nil it will perform a "IS NULL" otherwise it will perform a NOT IN
func (m Model) SoftDelete() (field string, deleteValue interface{}, activeValue []string) {
	return "DeletedAt", time.Now(), nil
}

// DefaultSchemaName returns the configured schema of the builder.
func (m Model) DefaultSchemaName() string {
	return m.builder.Config().Schema
}

// DefaultDatabaseName returns the configured database of the builder.
func (m Model) DefaultDatabaseName() string {
	return m.builder.Config().Database
}

// DefaultTableName returns the struct name pluralized in snake_case.
func (m Model) DefaultTableName() string {
	return fwStrings.CamelToSnake(fwStrings.Plural(m.modelName(false)))
}

// DefaultLogger returns the global model logger.
func (m Model) DefaultLogger() *logger.Logger {
	return server.Logger()
}

// DefaultCache returns the server cache.
// Error will return if no default cache is set.
func (m Model) DefaultCache() (cache.Interface, time.Duration, error) {
	if c, err := server.Cache(server.DEFAULT); err == nil {
		return c, cache.INFINITY, nil
	}
	return nil, 0, errors.New("orm: no server cache is defined")
}

// DefaultBuilder returns the GlobalBuilder.
func (m Model) DefaultBuilder() (sqlquery.Builder, error) {
	if b, err := server.Builder(server.DEFAULT); err == nil {
		return b, nil
	}
	return sqlquery.Builder{}, errors.New("orm: no server builder is defined")
}

// DefaultStrategy return the default database strategy.
func (m Model) DefaultStrategy() string {
	return "eager"
}

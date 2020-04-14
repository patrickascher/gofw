package orm2

import (
	"github.com/patrickascher/gofw/cache"
	"github.com/patrickascher/gofw/logger"
	"github.com/patrickascher/gofw/sqlquery"
	fwStrings "github.com/patrickascher/gofw/strings"

	"time"
)

// is used for self referencing m2m relations.
const defaultSelfReferenceAssociationForeignKey = "child_id"

// DefaultSchemaName returns the configured schema of the builder.
func (m Model) DefaultSchemaName() string {
	return m.DefaultBuilder().Config().Schema
}

// DefaultDatabaseName returns the configured database of the builder.
func (m Model) DefaultDatabaseName() string {
	return m.DefaultBuilder().Config().Database
}

// DefaultTableName returns the struct name pluralized in snake_case.
func (m Model) DefaultTableName() string {
	return fwStrings.CamelToSnake(fwStrings.Plural(m.modelName(false)))
}

// DefaultLogger returns the global model logger.
func (m Model) DefaultLogger() *logger.Logger {
	return GlobalLogger
}

// DefaultCache returns the global cache.
// The cache is set to 6 hours by default.
func (m Model) DefaultCache() (cache.Interface, time.Duration, error) {
	return GlobalCache, 6 * time.Hour, nil
}

// DefaultBuilder returns the GlobalBuilder.
func (m Model) DefaultBuilder() sqlquery.Builder {
	return GlobalBuilder
}

// DefaultStrategy return the default database strategy.
func (m Model) DefaultStrategy() string {
	return "eager"
}

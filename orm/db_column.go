package orm

import (
	"github.com/patrickascher/gofw/sqlquery"
)

// Permission for the field.
type Permission struct {
	Read  bool
	Write bool
}

// Column represents a database column and some additional information.
type Column struct {
	StructField string

	SqlSelect  string
	Permission Permission

	Information *sqlquery.Column

	Validator *Validator
}

// Check if field exists
func (c *Column) ExistsInDB() bool {
	return c.Information.Table != ""
}

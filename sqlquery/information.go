// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlquery

import (
	"database/sql"
)

// Information struct.
type Information struct {
	builder  *Builder
	database string // database will be used from the configuration, except the table name has a dot notation.
	table    string
}

// Column represents a database table column.
type Column struct {
	Table         string
	Name          string
	Position      int
	NullAble      bool
	PrimaryKey    bool
	Type          Type
	DefaultValue  sql.NullString
	Length        sql.NullInt64
	Autoincrement bool
}

// ForeignKey represents a table relation.
type ForeignKey struct {
	Name      string
	Primary   Relation
	Secondary Relation
}

// Relation defines the table and column of a relation.
type Relation struct {
	Table  string
	Column string
}

// Describe the requested columns of the database table.
// By default the configure database is used, except the table name has a dot notation.
func (i Information) Describe(columns ...string) ([]*Column, error) {
	return i.builder.driver.Describe(i.builder, i.database, i.table, columns)
}

// ForeignKeys for the requested table.
// By default the configure database is used, except the table name has a dot notation.
func (i Information) ForeignKeys() ([]*ForeignKey, error) {
	return i.builder.driver.ForeignKeys(i.builder, i.database, i.table)
}

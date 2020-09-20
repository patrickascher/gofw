// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package driver contains some out of the box db drivers which implement the sqlquery.Driver interface.
package oracle

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/patrickascher/gofw/sqlquery/types"
	_ "github.com/patrickascher/ora"
)

// Error messages.
var (
	ErrTableDoesNotExist = errors.New("sqlquery: table %s or column does not exist %v")
)

// init register the mysql driver.
func init() {
	sqlquery.Register("oracle", newOracle)
}

// newMysql creates a db connection.
// If the argument *sql.DB is not nil, the existing connection is taken, otherwise a new connection will be created.
func newOracle(cfg sqlquery.Config, db *sql.DB) (sqlquery.DriverI, error) {

	o := &oracle{}
	o.config = cfg
	if db != nil {
		o.connection = db
	} else {
		//oci8
		db, err := sql.Open("ora", fmt.Sprintf("%s/%s@%s:%d/%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Schema))
		if err != nil {
			return nil, err
		}
		o.connection = db
	}

	return o, nil
}

// mysql driver.
type oracle struct {
	connection *sql.DB
	config     sqlquery.Config
}

// Connection returns the existing *sql.DB.
func (o *oracle) Connection() *sql.DB {
	return o.connection
}

// Connection returns the existing *sql.DB.
func (o *oracle) Config() sqlquery.Config {
	return o.config
}

// QuoteCharacterColumn returns the identifier quote character.
func (o *oracle) QuoteCharacterColumn() string {
	return "\""
}

// Describe the database table.
// If the columns argument is set, only the required columns are requested.
func (o *oracle) Describe(b *sqlquery.Builder, db string, table string, columns []string) ([]sqlquery.Column, error) {
	sel := b.Select("USER_TAB_COLUMNS")
	sel.Columns("COLUMN_NAME",
		"COLUMN_ID",
		sqlquery.Raw("case when NULLABLE='Y' THEN 'TRUE' ELSE 'FALSE' END AS \"N\""),
		sqlquery.Raw("'FALSE' AS \"K\""),
		sqlquery.Raw("'FALSE' AS \"U\""),
		"DATA_TYPE",
		sqlquery.Raw("''"), // DATA_DEFAULT - default was deleted because there are some major memory leaks with that. dont need defaults at the moment. fix: switch driver?
		"CHAR_LENGTH",
		sqlquery.Raw("'FALSE' as \"autoincrement\""),
	).Where("table_name = ?", table).Order("COLUMN_ID")

	if len(columns) > 0 {
		sel.Where("COLUMN_NAME IN (?)", columns)
	}

	rows, err := sel.All()

	if err != nil {
		return nil, err
	}

	defer func() {
		rows.Close()
	}()

	var cols []sqlquery.Column
	for rows.Next() {

		var c sqlquery.Column
		c.Table = table // adding Table info

		var t string
		if err := rows.Scan(&c.Name, &c.Position, &c.NullAble, &c.PrimaryKey, &c.Unique, &t, &c.DefaultValue, &c.Length, &c.Autoincrement); err != nil {
			return nil, err
		}

		c.Type = o.TypeMapping(t, c)
		cols = append(cols, c)
	}

	if len(cols) == 0 {
		return nil, fmt.Errorf(ErrTableDoesNotExist.Error(), db+"."+table, columns)
	}

	return cols, nil

}

// ForeignKeys returns the relation of the given table.
// TODO: already set the relation Type (hasOne, hasMany, m2m,...) ? Does this make sense already here instead of the ORM.
func (o *oracle) ForeignKeys(b *sqlquery.Builder, db string, table string) ([]*sqlquery.ForeignKey, error) {
	return nil, errors.New("oracle: foreign keys are not implemented yet!")
}

// Placeholder return a configured placeholder for the db driver.
func (o *oracle) Placeholder() *sqlquery.Placeholder {
	return &sqlquery.Placeholder{Char: ":", Numeric: true}
}

// TypeMapping converts the database type to an unique sqlquery type over different database drives.
func (o *oracle) TypeMapping(raw string, col sqlquery.Column) types.Interface {
	//TODO oracle types
	return types.NewText(raw)
}

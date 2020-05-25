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

	"strings"
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
		"DATA_TYPE",
		"DATA_DEFAULT",
		"CHAR_LENGTH",
		sqlquery.Raw("'FALSE' as \"autoincrement\""),
	).Where("table_name = ?", table)

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
		if err := rows.Scan(&c.Name, &c.Position, &c.NullAble, &c.PrimaryKey, &t, &c.DefaultValue, &c.Length, &c.Autoincrement); err != nil {
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
	//Integer
	if strings.HasPrefix(raw, "bigint") ||
		strings.HasPrefix(raw, "int") ||
		strings.HasPrefix(raw, "mediumint") ||
		strings.HasPrefix(raw, "smallint") ||
		strings.HasPrefix(raw, "tinyint") {

		integer := types.NewInt(raw)
		// Bigint
		if strings.HasPrefix(raw, "bigint") {
			if strings.HasSuffix(raw, "unsigned") {
				integer.Min = 0
				integer.Max = 18446744073709551615 //actually 18446744073709551616 but overflows uint64
			} else {
				integer.Min = -9223372036854775808
				integer.Max = 9223372036854775807
			}
		}

		// Int
		if strings.HasPrefix(raw, "int") {
			if strings.HasSuffix(raw, "unsigned") {
				integer.Min = 0
				integer.Max = 4294967295
			} else {
				integer.Min = -2147483648
				integer.Max = 2147483647
			}
		}

		// MediumInt
		if strings.HasPrefix(raw, "mediumint") {
			if strings.HasSuffix(raw, "unsigned") {
				integer.Min = 0
				integer.Max = 16777215
			} else {
				integer.Min = -8388608
				integer.Max = 8388607
			}
		}

		// SmallInt
		if strings.HasPrefix(raw, "smallint") {
			if strings.HasSuffix(raw, "unsigned") {
				integer.Min = 0
				integer.Max = 65535
			} else {
				integer.Min = -32768
				integer.Max = 32767
			}
		}

		// TinyInt
		if strings.HasPrefix(raw, "tinyint") {
			if strings.HasSuffix(raw, "unsigned") {
				integer.Min = 0
				integer.Max = 255
			} else {
				integer.Min = -128
				integer.Max = 127
			}
		}

		return integer

	}

	// Float
	if strings.HasPrefix(raw, "decimal") ||
		strings.HasPrefix(raw, "float") ||
		strings.HasPrefix(raw, "double") {
		f := types.NewFloat(raw)
		//TODO decimal point
		return f
	}

	// Text
	if strings.HasPrefix(raw, "varchar") ||
		strings.HasPrefix(raw, "char") {
		text := types.NewText(raw)
		if col.Length.Valid {
			text.Size = int(col.Length.Int64)
		}
		return text
	}

	// TextArea
	if strings.HasPrefix(raw, "tinytext") ||
		strings.HasPrefix(raw, "text") ||
		strings.HasPrefix(raw, "mediumtext") ||
		strings.HasPrefix(raw, "longtext") {
		textArea := types.NewTextArea(raw)

		if strings.HasPrefix(raw, "tinytext") {
			textArea.Size = 255
		}

		if strings.HasPrefix(raw, "text") {
			textArea.Size = 65535
		}

		if strings.HasPrefix(raw, "mediumtext") {
			textArea.Size = 16777215
		}

		if strings.HasPrefix(raw, "longtext") {
			textArea.Size = 4294967295
		}

		return textArea
	}

	// Time
	if raw == "time" {
		time := types.NewTime(raw)
		return time
	}

	// Date
	if raw == "date" {
		date := types.NewDate(raw)
		return date
	}

	// DateTime
	if raw == "datetime" || raw == "timestamp" {
		dateTime := types.NewDateTime(raw)
		return dateTime
	}

	return nil
}

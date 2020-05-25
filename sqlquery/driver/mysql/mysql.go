// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package driver contains some out of the box db drivers which implement the sqlquery.Driver interface.
package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/patrickascher/gofw/sqlquery/types"
	"strings"

	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/patrickascher/gofw/sqlquery"
)

// init register the mysql driver.
func init() {
	sqlquery.Register("mysql", newMysql)
}

// Error messages.
var (
	ErrTableDoesNotExist = errors.New("sqlquery: table %s or column does not exist %v")
)

// newMysql creates a db connection.
// If the argument *sql.DB is not nil, the existing connection is taken, otherwise a new connection will be created.
func newMysql(cfg sqlquery.Config, db *sql.DB) (sqlquery.DriverI, error) {

	m := &mysql{}
	m.config = cfg
	if db != nil {
		m.connection = db
	} else {
		db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=true", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database))
		if err != nil {
			return nil, err
		}
		m.connection = db
	}

	m.connection.SetMaxOpenConns(5)

	return m, nil
}

// mysql driver.
type mysql struct {
	connection *sql.DB
	config     sqlquery.Config
}

// Connection returns the existing *sql.DB.
func (m *mysql) Connection() *sql.DB {
	return m.connection
}

// Connection returns the existing *sql.DB.
func (m *mysql) Config() sqlquery.Config {
	return m.config
}

// QuoteCharacterColumn returns the identifier quote character.
func (m *mysql) QuoteCharacterColumn() string {
	return "`"
}

// Describe the database table.
// If the columns argument is set, only the required columns are requested.
func (m *mysql) Describe(b *sqlquery.Builder, db string, table string, columns []string) ([]sqlquery.Column, error) {

	sel := b.Select("information_schema.COLUMNS c")
	sel.Columns("c.COLUMN_NAME",
		"c.ORDINAL_POSITION",
		sqlquery.Raw("IF(c.IS_NULLABLE='YES','TRUE','FALSE') AS N"),
		sqlquery.Raw("IF(COLUMN_KEY='PRI','TRUE','FALSE') AS K"),
		"c.COLUMN_TYPE",
		"c.COLUMN_DEFAULT",
		"c.CHARACTER_MAXIMUM_LENGTH",
		sqlquery.Raw("IF(EXTRA='auto_increment','TRUE','FALSE') AS autoincrement"),
	).
		Where("c.TABLE_SCHEMA = ?", db).
		Where("c.TABLE_NAME = ?", table)

	if len(columns) > 0 {
		sel.Where("c.COLUMN_NAME IN (?)", columns)
	}
	rows, err := sel.All()
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var cols []sqlquery.Column

	for rows.Next() {
		var c sqlquery.Column
		c.Table = table // adding Table info

		var t string
		if err := rows.Scan(&c.Name, &c.Position, &c.NullAble, &c.PrimaryKey, &t, &c.DefaultValue, &c.Length, &c.Autoincrement); err != nil {
			return nil, err
		}
		c.Type = m.TypeMapping(t, c)
		cols = append(cols, c)
	}

	if len(cols) == 0 {
		return nil, fmt.Errorf(ErrTableDoesNotExist.Error(), db+"."+table, columns)
	}

	return cols, nil

}

// ForeignKeys returns the relation of the given table.
// TODO: already set the relation Type (hasOne, hasMany, m2m,...) ? Does this make sense already here instead of the ORM.
func (m *mysql) ForeignKeys(b *sqlquery.Builder, db string, table string) ([]*sqlquery.ForeignKey, error) {

	sel := b.Select("!information_schema.key_column_usage cu, information_schema.table_constraints tc").
		Columns("tc.constraint_name", "tc.table_name", "cu.column_name", "cu.referenced_table_name", "cu.referenced_column_name").
		Where("cu.constraint_name = tc.constraint_name AND cu.table_name = tc.table_name AND tc.constraint_type = 'FOREIGN KEY'").
		Where("cu.table_schema = ?", db).
		Where("tc.table_schema = ?", db).
		Where("tc.table_name = ?", table)

	rows, err := sel.All()
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var fKeys []*sqlquery.ForeignKey

	for rows.Next() {
		f := sqlquery.ForeignKey{Primary: sqlquery.Relation{}, Secondary: sqlquery.Relation{}}
		if err := rows.Scan(&f.Name, &f.Primary.Table, &f.Primary.Column, &f.Secondary.Table, &f.Secondary.Column); err != nil {
			return nil, err
		}
		fKeys = append(fKeys, &f)
	}

	return fKeys, nil
}

// Placeholder return a configured placeholder for the db driver.
func (m *mysql) Placeholder() *sqlquery.Placeholder {
	return &sqlquery.Placeholder{Char: "?", Numeric: false}
}

// TypeMapping converts the database type to an unique sqlquery type over different database drives.
func (m *mysql) TypeMapping(raw string, col sqlquery.Column) types.Interface {

	// Bool
	if strings.HasPrefix(raw, "enum(0,1)") ||
		strings.HasPrefix(raw, "tinyint(1)") {
		f := types.NewBool(raw)
		return f
	}

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

	// ENUM, SET
	if strings.HasPrefix(raw, "enum") {
		enum := types.NewEnum(raw)
		for _, v := range strings.Split(raw[5:len(raw)-1], ",") {
			enum.Values = append(enum.Values, v[1:len(v)-1])
		}
		return enum
	}

	return nil
}

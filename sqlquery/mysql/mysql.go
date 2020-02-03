// Package mysql defines the describe and foreign key syntax.
// It also loads the "github.com/go-sql-driver/mysql" driver
//
// See https://github.com/patrickascher/go-sql for more information and examples.
package mysql

import (
	_ "github.com/go-sql-driver/mysql" //include the mysql driver
	"github.com/patrickascher/gofw/sqlquery"
	"strings"
)

// init registers itself as mysql driver
func init() {
	sqlquery.Register("mysql", &Mysql{})
}

// Mysql driver
type Mysql struct {
}

// Describe select syntax
func (m *Mysql) Describe(database string, table string, builder *sqlquery.Builder, cols []string) *sqlquery.Select {
	sel := builder.Select("information_schema.COLUMNS c")
	sel.Columns("c.COLUMN_NAME",
		"c.ORDINAL_POSITION",
		"!IF(c.IS_NULLABLE='YES','TRUE','FALSE') AS N",
		"!IF(COLUMN_KEY='PRI','TRUE','FALSE') AS K",
		"c.COLUMN_TYPE",
		"c.COLUMN_DEFAULT",
		"c.CHARACTER_MAXIMUM_LENGTH",
		"!IF(EXTRA='auto_increment','TRUE','FALSE') AS autoincrement",
	).
		Where("c.TABLE_SCHEMA = ?", database).
		Where("c.TABLE_NAME = ?", table)

	if len(cols) > 0 {
		sel.Where("c.COLUMN_NAME IN (?)", cols)
	}

	return sel
}

// ForeignKeys select syntax
func (m *Mysql) ForeignKeys(database string, table string, builder *sqlquery.Builder) *sqlquery.Select {
	sel := builder.Select("!information_schema.key_column_usage cu, information_schema.table_constraints tc").
		Columns("tc.constraint_name", "tc.table_name", "cu.column_name", "cu.referenced_table_name", "cu.referenced_column_name").
		Where("cu.constraint_name = tc.constraint_name AND cu.table_name = tc.table_name AND tc.constraint_type = 'FOREIGN KEY'").
		Where("tc.table_schema = ?", database).
		Where("tc.table_name = ?", table)

	return sel
}

// ConvertColumnType is creating a global Type for Database specific types
func (m *Mysql) ConvertColumnType(raw string, col *sqlquery.Column) sqlquery.Type {

	//Integer
	if strings.HasPrefix(raw, "bigint") ||
		strings.HasPrefix(raw, "int") ||
		strings.HasPrefix(raw, "mediumint") ||
		strings.HasPrefix(raw, "smallint") ||
		strings.HasPrefix(raw, "tinyint") {

		integer := sqlquery.NewInt(raw)
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
		f := sqlquery.NewFloat(raw)
		//TODO decimal point
		return f
	}

	// Text
	if strings.HasPrefix(raw, "varchar") ||
		strings.HasPrefix(raw, "char") {
		text := sqlquery.NewText(raw)
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
		textArea := sqlquery.NewTextArea(raw)

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
		time := sqlquery.NewTime(raw)
		return time
	}

	// Date
	if raw == "date" {
		date := sqlquery.NewDate(raw)
		return date
	}

	// DateTime
	if raw == "datetime" || raw == "timestamp" {
		dateTime := sqlquery.NewDateTime(raw)
		return dateTime
	}

	return nil
}

//TODO case "bit,bool - tinyint(1)":
//TODO ENUM
//TODO SET

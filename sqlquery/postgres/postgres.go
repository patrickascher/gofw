// Package postgres defines the describe and foreign key syntax.
// It also loads the "github.com/lib/pq" driver
//
// See https://github.com/patrickascher/go-sql for more information and examples.
package postgres

import (
	_ "github.com/lib/pq" //include the postgres driver
	"github.com/patrickascher/gofw/sqlquery"
	"strings"
)

// init registers itself as mysql driver
func init() {
	sqlquery.Register("postgres", &Postgres{})
}

// Postgres driver
type Postgres struct {
}

// Describe syntax
func (p *Postgres) Describe(database string, table string, builder *sqlquery.Builder, cols []string) *sqlquery.Select {

	sel := builder.Select("information_schema.columns c")
	sel.Columns("c.column_name",
		"c.ordinal_position",
		"!CASE WHEN c.is_nullable = 'YES' THEN 'TRUE' ELSE 'FALSE' END AS nullable",
		"!CASE WHEN (SELECT tcu.constraint_name FROM information_schema.table_constraints tc LEFT JOIN information_schema.constraint_column_usage tcu ON  tcu.table_name = tc.table_name AND tcu.constraint_name = tc.constraint_name AND c.column_name = tcu.column_name WHERE tc.table_name = '"+table+"' AND constraint_type = 'PRIMARY KEY') IS NOT NULL THEN 'TRUE' ELSE 'FALSE' END AS cprimary",
		"c.data_type",
		"c.column_default",
		"c.character_maximum_length",
		"!CASE WHEN c.column_default LIKE 'nextval%' THEN 'TRUE' ELSE 'FALSE' END AS autoincrement").
		Where("c.table_name = ?", table).
		Where("c.table_catalog = ?", database)

	if len(cols) > 0 {
		sel.Where("c.column_name IN (?)", cols)
	}

	return sel
}

// ForeignKeys select syntax
func (p *Postgres) ForeignKeys(database string, table string, builder *sqlquery.Builder) *sqlquery.Select {
	jC := sqlquery.Condition{}
	jC2 := sqlquery.Condition{}

	sel := builder.Select("information_schema.table_constraints tc").
		Columns("tc.constraint_name", "tc.table_name", "kcu.column_name", "ccu.table_name foreign_table_name", "ccu.column_name foreign_column_name").
		Join(sqlquery.LEFT, "information_schema.key_column_usage kcu", jC.On("tc.constraint_name = kcu.constraint_name")).
		Join(sqlquery.LEFT, "information_schema.constraint_column_usage ccu", jC2.On("ccu.constraint_name = tc.constraint_name")).
		Where("constraint_type = ?", "FOREIGN KEY").
		Where("tc.table_name = ?", table).
		Where("tc.constraint_catalog = ?", database)

	return sel
}

// ConvertColumnType is creating a global Type for Database specific types
func (p *Postgres) ConvertColumnType(raw string, col *sqlquery.Column) sqlquery.Type {

	//Integer
	if strings.HasPrefix(raw, "bigint") ||
		strings.HasPrefix(raw, "int8") ||
		strings.HasPrefix(raw, "integer") ||
		strings.HasPrefix(raw, "smallint") {

		integer := sqlquery.NewInt(raw)

		// Bigint
		if strings.HasPrefix(raw, "bigint") || strings.HasPrefix(raw, "int8") {
			if col.Autoincrement {
				integer.Min = 1
				integer.Max = 9223372036854775807 //actually 18446744073709551616 but overflows uint64
			} else {
				integer.Min = -9223372036854775808
				integer.Max = 9223372036854775807
			}
		}

		// Int
		if strings.HasPrefix(raw, "integer") {
			if col.Autoincrement {
				integer.Min = 1
				integer.Max = 2147483647
			} else {
				integer.Min = -2147483648
				integer.Max = 2147483647
			}
		}

		// MediumInt
		if strings.HasPrefix(raw, "smallint") {
			if col.Autoincrement {
				integer.Min = 1
				integer.Max = 32767
			} else {
				integer.Min = -32768
				integer.Max = 32767
			}
		}

		return integer

	}

	// Float
	if strings.HasPrefix(raw, "real") ||
		strings.HasPrefix(raw, "double precision") ||
		strings.HasPrefix(raw, "numeric") {
		f := sqlquery.NewFloat(raw)
		//TODO decimal point
		return f
	}

	// Text
	if strings.HasPrefix(raw, "character") ||
		strings.HasPrefix(raw, "char") ||
		strings.HasPrefix(raw, "varchar") ||
		strings.HasPrefix(raw, "character varying") {
		text := sqlquery.NewText(raw)
		if col.Length.Valid {
			text.Size = int(col.Length.Int64)
		}
		return text
	}

	// TextArea
	if strings.HasPrefix(raw, "text") {
		textArea := sqlquery.NewTextArea(raw)
		//TODO size?
		return textArea
	}

	// Time
	if raw == "time without time zone" || raw == "time with time zone" {
		time := sqlquery.NewTime(raw)
		return time
	}

	// Date
	if raw == "date" {
		date := sqlquery.NewDate(raw)
		return date
	}

	// DateTime
	if raw == "timestamp with time zone" || raw == "timestamp without time zone" {
		dateTime := sqlquery.NewDateTime(raw)
		return dateTime
	}

	return nil
}

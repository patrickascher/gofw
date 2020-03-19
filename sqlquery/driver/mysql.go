// Package mysql defines the describe and foreign key syntax.
// It also loads the "github.com/go-sql-driver/mysql" driver
//
// See https://github.com/patrickascher/go-sql for more information and examples.
package driver

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql" //include the mysql driver
	"github.com/patrickascher/gofw/sqlquery"
	"strings"
)

// init registers itself as mysql driver
func init() {
	sqlquery.Register("mysql", newMysql)
}

// Error messages.
var (
	ErrTableDoesNotExist = errors.New("sqlquery: table %s does not exist")
)

// New creates a in-memory cache by the given options.
func newMysql(cfg sqlquery.Config, db *sql.DB) (sqlquery.Driver, error) {

	m := &mysql{}
	if db != nil {
		m.connection = db
	} else {
		db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=true", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database))
		if err != nil {
			return nil, err
		}
		m.connection = db
	}

	return m, nil
}

// Mysql driver
type mysql struct {
	connection *sql.DB
}

func (m *mysql) Connection() *sql.DB {
	return m.connection
}

func (m *mysql) QuoteCharacterColumn() string {
	return "`"
}

func (m *mysql) Describe(b *sqlquery.Builder, db string, table string, columns []string) ([]*sqlquery.Column, error) {

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

	var cols []*sqlquery.Column

	for rows.Next() {
		var c sqlquery.Column
		c.Table = table // adding Table info

		var t string
		if err := rows.Scan(&c.Name, &c.Position, &c.NullAble, &c.PrimaryKey, &t, &c.DefaultValue, &c.Length, &c.Autoincrement); err != nil {
			return nil, err
		}
		c.Type = m.TypeMapping(t, c)
		cols = append(cols, &c)
	}

	if len(cols) == 0 {
		return nil, fmt.Errorf(ErrTableDoesNotExist.Error(), db+"."+table)
	}

	return cols, nil

}

func (m *mysql) ForeignKeys(b *sqlquery.Builder, db string, table string) ([]*sqlquery.ForeignKey, error) {

	sel := b.Select("!information_schema.key_column_usage cu, information_schema.table_constraints tc").
		Columns("tc.constraint_name", "tc.table_name", "cu.column_name", "cu.referenced_table_name", "cu.referenced_column_name").
		Where("cu.constraint_name = tc.constraint_name AND cu.table_name = tc.table_name AND tc.constraint_type = 'FOREIGN KEY'").
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

func (m *mysql) Placeholder() *sqlquery.Placeholder {
	return &sqlquery.Placeholder{Char: "?", Numeric: false}
}

func (m *mysql) TypeMapping(raw string, col sqlquery.Column) sqlquery.Type {
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

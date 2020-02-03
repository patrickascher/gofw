package sqlquery

import (
	"errors"
	"fmt"
	"strings"
)

// All error messages are defined here
var (
	ErrNoDatabase        = errors.New("sqlquery: no database name is defined")
	ErrTableDoesNotExist = errors.New("sqlquery: table %s does not exist")
)

// Information config
type Information struct {
	builder  *Builder
	table    string
	database string
}

// Column represents a table column
type Column struct {
	Table         string
	Name          string
	Position      int
	NullAble      bool
	PrimaryKey    bool
	Type          Type
	DefaultValue  NullString
	Length        NullInt64
	Autoincrement bool
}

// ForeignKey represents a table relation
type ForeignKey struct {
	Name      string
	Primary   *Relation
	Secondary *Relation
}

// Relation defines the table and column of a relation
type Relation struct {
	Table  string
	Column string
}

// Describe all columns of the given table
func (i *Information) Describe(columns ...string) ([]*Column, error) {

	driver, err := i.getDriver()
	if err != nil {
		return nil, err
	}

	// create select over driver interface
	sel := driver.Describe(i.database, i.table, i.builder, columns)
	rows, err := sel.All()
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var cols []*Column

	for rows.Next() {
		var c Column
		c.Table = i.table // adding Table info

		var t string
		if err := rows.Scan(&c.Name, &c.Position, &c.NullAble, &c.PrimaryKey, &t, &c.DefaultValue, &c.Length, &c.Autoincrement); err != nil {
			return nil, err
		}
		c.Type = driver.ConvertColumnType(t, &c)
		cols = append(cols, &c)
	}

	if len(cols) == 0 {
		return nil, fmt.Errorf(ErrTableDoesNotExist.Error(), i.database+"."+i.table)
	}

	return cols, nil
}

// getDriver returns the driver matched to the builder config adapter
func (i *Information) getDriver() (Driver, error) {
	driver, err := NewDriver(i.builder.conf.Driver())
	if err != nil {
		return nil, err
	}

	// define database and table name
	i.database, i.table = splitDatabaseAndTable(i.table)
	if i.database == "" {
		if i.database = i.builder.conf.DbName(); i.database == "" {
			return nil, ErrNoDatabase
		}
	}
	return driver, nil
}

// ForeignKeys of the given table
func (i *Information) ForeignKeys() ([]*ForeignKey, error) {

	driver, err := i.getDriver()
	if err != nil {
		return nil, err
	}

	// create select over driver interface
	sel := driver.ForeignKeys(i.database, i.table, i.builder)
	rows, err := sel.All()
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var fkeys []*ForeignKey

	for rows.Next() {
		f := ForeignKey{Primary: &Relation{}, Secondary: &Relation{}}
		if err := rows.Scan(&f.Name, &f.Primary.Table, &f.Primary.Column, &f.Secondary.Table, &f.Secondary.Column); err != nil {
			return nil, err
		}
		fkeys = append(fkeys, &f)
	}

	return fkeys, nil
}

// splitDatabaseAndTableName is a helper to split "db.table" into
func splitDatabaseAndTable(s string) (database string, table string) {
	identifier := strings.Split(s, ".")
	if len(identifier) == 1 {
		return "", identifier[0]
	}
	return identifier[0], identifier[1]
}

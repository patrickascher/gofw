// Package sqlquery is a simple SQL generator.
//
// See https://github.com/patrickascher/go-sql for more information and examples.
package sqlquery

import (
	"database/sql"
	"errors"
	"strings"
)

var (
	ErrNoTx = errors.New("sqlquery: no tx exists")
)

// Builder stores the *sql.DB and the Placeholder information.
// This information is needed to translate and execute the query.
type Builder struct {
	driver Driver
	tx     *sql.Tx
	txAuto bool
	conf   Config
}

// New Builder instance with the given configuration.
// If the db argument is nil, a new db connection will be created.
// It is highly recommended to use one open connection to avoid overhead.
func New(cfg Config, db *sql.DB) (Builder, error) {
	d, err := newDriver(cfg, db)
	if err != nil {
		return Builder{}, err
	}
	return Builder{driver: d, conf: cfg}, nil
}

func Raw(c string) string {
	return "!" + c
}

func (b *Builder) Tx() error {
	tx, err := b.driver.Connection().Begin()
	if err != nil {
		return err
	}
	b.tx = tx
	return nil
}

func (b *Builder) Commit() error {
	if b.tx == nil {
		return ErrNoTx
	}
	return b.tx.Commit()
}

func (b *Builder) Rollback() error {
	if b.tx == nil {
		return ErrNoTx
	}
	return b.tx.Rollback()
}

// Information - please see the Information type documentation.
func (b *Builder) Information(table string) *Information {
	database, table := splitDatabaseAndTable(b.conf.Database, table)
	return &Information{builder: b, database: database, table: table}
}

// Select - please see the Select type documentation.
func (b *Builder) Select(from string) *Select {
	return &Select{builder: b, from: from, condition: &Condition{}}
}

// Insert - please see the Select type documentation.
func (b *Builder) Insert(into string) *Insert {
	return &Insert{builder: b, into: into, batch: BATCH}
}

// Update - please see the Update type documentation.
func (b *Builder) Update(table string) *Update {
	return &Update{builder: b, table: table, condition: &Condition{}}
}

// Delete - please see the Delete type documentation.
func (b *Builder) Delete(from string) *Delete {
	return &Delete{builder: b, from: from, condition: &Condition{}}
}

// splitDatabaseAndTableName is a helper to split "db.table" into
func splitDatabaseAndTable(db string, s string) (database string, table string) {
	identifier := strings.Split(s, ".")
	if len(identifier) == 1 {
		return db, identifier[0]
	}
	return identifier[0], identifier[1]
}

// quoteColumns is a helper to quote all identifiers
func (b *Builder) quoteColumns(columns ...string) string {
	colStmt := ""
	for _, c := range columns {
		if colStmt != "" {
			colStmt += ", "
		}

		// dont escape sqlquery.Raw()
		if c[0:1] == "!" {
			colStmt += c[1:]
			continue
		}

		// replace quote characters in the column name.
		// TODO they should get escaped in theory.
		c = strings.Replace(c, b.driver.QuoteCharacterColumn(), "", -1)

		// check if an alias was used
		alias := strings.Split(c, " ")
		var columnSplit []string
		columnSplit = strings.Split(alias[0], ".")
		var rv string
		for _, i := range columnSplit {
			if rv != "" {
				rv += "."
			}
			rv += b.driver.QuoteCharacterColumn() + i + b.driver.QuoteCharacterColumn()
		}
		if len(alias) >= 2 {
			rv += " " + b.quoteColumns(alias[len(alias)-1])
		}

		colStmt += rv
	}

	return colStmt
}

// addColumns adding columns from a key/value pair if they were not added manually before
// needed in update and insert
func (b *Builder) addColumns(columns []string, values map[string]interface{}) []string {
	if len(columns) == 0 {
		for column := range values {
			columns = append(columns, column)
		}
	}
	return columns
}

func (b *Builder) replacePlaceholders(stmt string, p *Placeholder) string {
	n := strings.Count(stmt, PLACEHOLDER)
	for i := 1; i <= n; i++ {
		stmt = strings.Replace(stmt, PLACEHOLDER, p.placeholder(), 1)
	}

	return stmt
}

// first queries one Row by using the *db.QueryRow method.
func (b *Builder) first(stmt string, args []interface{}) *sql.Row {
	// TODO TX
	return b.driver.Connection().QueryRow(stmt, args...)
}

// all queries all rows by using the *db.Query method.
func (b *Builder) all(stmt string, args []interface{}) (*sql.Rows, error) {
	// TODO TX
	return b.driver.Connection().Query(stmt, args...)
}

// exec executes the rendered sql query.
func (b *Builder) exec(stmt string, args [][]interface{}) ([]sql.Result, error) {

	var res []sql.Result
	var err error

	// create a transaction if batching
	if len(args) > 1 && b.tx == nil {
		err = b.Tx()
		if err != nil {
			return nil, err
		}
		b.txAuto = true
	}

	for _, arg := range args {
		// start timer for debug

		var r sql.Result
		var err error

		if b.tx != nil {
			r, err = b.tx.Exec(stmt, arg...)
		} else {
			r, err = b.driver.Connection().Exec(stmt, arg...)
		}

		if err != nil {
			return nil, err
		}
		res = append(res, r)
	}

	// batch auto commit
	if b.txAuto {
		err = b.Commit()
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

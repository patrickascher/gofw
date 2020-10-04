// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// sqlquery is a simple programmatically sql query builder.
// The idea was to create a unique Builder which can be used with any database driver in go.
//
// Features: Unique Placeholder for all database drivers, Batching function for large Inserts, Whitelist, Quote Identifiers, SQL queries and durations log debugging
package sqlquery

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/patrickascher/gofw/logger"
	"strings"
	"time"
)

// Error messages.
var (
	ErrNoTx = errors.New("sqlquery: no tx exists")
)

// Builder type.
type Builder struct {
	driver DriverI
	tx     *sql.Tx
	txAuto bool
	conf   Config

	logger *logger.Logger
}

// New Builder instance with the given configuration.
// If the db argument is nil, a new db connection will be created.
// It is highly recommended to use one open connection to avoid overhead.
// TODO: Idea: only one interface argument, in a type select we can figure out if it was a config or *sql.DB.
func New(cfg Config, db *sql.DB) (Builder, error) {
	d, err := newDriver(cfg, db)
	if err != nil {
		return Builder{}, err
	}
	// Setting config
	d.Connection().SetMaxIdleConns(cfg.MaxIdleConnections)                              // go default 2
	d.Connection().SetMaxOpenConns(cfg.MaxOpenConnections)                              // go default 0
	d.Connection().SetConnMaxLifetime(time.Duration(cfg.MaxConnLifetime) * time.Minute) // go default 0

	// Ping the server, to guarantee a connection.
	err = d.Connection().Ping()
	if err != nil {
		return Builder{}, err
	}

	// pre query
	if cfg.PreQuery != "" {
		_, err := d.Connection().Exec(cfg.PreQuery)
		fmt.Println("running prequery", cfg.PreQuery)
		if err != nil {
			return Builder{}, err
		}
	}

	return Builder{driver: d, conf: cfg}, nil
}

// Raw can be used if the identifier should not be quoted.
func Raw(c string) string {
	return "!" + c
}

func (b *Builder) SetLogger(l *logger.Logger) {
	b.logger = l
}

func (b Builder) Config() Config {
	return b.conf
}

func (b *Builder) log(stmt string, d time.Duration, args ...interface{}) {
	if b.conf.Debug && b.logger != nil {
		b.logger.Debug(fmt.Sprintf("%s with the arguments %v took %s", stmt, args, d))
	}
}

func (b *Builder) Driver() DriverI {
	return b.driver
}

func (b *Builder) HasTx() bool {
	return b.tx != nil
}

// Tx creates a new transaction for the builder.
// If a tx exists, all requests (select, update, insert, delete) will be handled in that transaction.
func (b *Builder) Tx() error {
	tx, err := b.driver.Connection().Begin()
	if err != nil {
		return err
	}
	b.tx = tx
	return nil
}

// Commit the builder transaction.
// Error will return if no transaction was created or there is a commit error.
func (b *Builder) Commit() error {
	if b.tx == nil {
		return ErrNoTx
	}
	fmt.Println("CALLED COMMIT")
	err := b.tx.Commit()
	b.tx = nil
	return err
}

// Rollback the builder transaction.
// Error will return if no transaction was created or there is a rollback error.
func (b *Builder) Rollback() error {
	if b.tx == nil {
		return ErrNoTx
	}
	fmt.Println("CALLED ROLLBACK")

	err := b.tx.Rollback()
	b.tx = nil
	return err
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

// splitDatabaseAndTableName is a helper to split "db.table" into two strings.
func splitDatabaseAndTable(db string, s string) (database string, table string) {
	identifier := strings.Split(s, ".")
	if len(identifier) == 1 {
		return db, identifier[0]
	}
	return identifier[0], identifier[1]
}

// QuoteIdentifier by the driver quote character.
func (b Builder) QuoteIdentifier(col string) string {
	return b.quoteColumns(col)
}

// quoteColumns is a helper to quote all identifiers.
// One or more columns can be added as argument.
// If the column was wrapped in sqlquery.Raw, it will not get escaped.
// "gofw.users AS u" will be converted to `gofw`.`users` AS `u`
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

// addColumns adding columns from a key/value pair. This is used in INSERT and UPDATE if no Columns are set, and the
// columns are added out of the key/value pairs.
func (b *Builder) addColumns(columns []string, values map[string]interface{}) []string {
	if len(columns) == 0 {
		for column := range values {
			columns = append(columns, column)
		}
	}
	return columns
}

// replacePlaceholders switches the sqlquery placeholder to the needed driver placeholder.
// BUG(patrick): condition/render this logic fails if the placeholder is numeric and has the same char as placeholder.
func replacePlaceholders(stmt string, p *Placeholder) string {
	n := strings.Count(stmt, PLACEHOLDER)
	for i := 1; i <= n; i++ {
		stmt = strings.Replace(stmt, PLACEHOLDER, p.placeholder(), 1)
	}

	return stmt
}

// convertArgumentsExtraSlice is needed because Insert uses a slice of an slice if interface because of the batching function.
// That Update and Delete can use the same exec function, the result will also get wrapped.
func convertArgumentsExtraSlice(stmt string, arg []interface{}, err error) (string, [][]interface{}, error) {
	var args [][]interface{}
	if err != nil {
		return "", nil, err
	}
	args = append(args, arg)
	return stmt, args, err
}

// first queries one Row by using the *db.QueryRow method.
func (b *Builder) first(stmt string, args []interface{}) *sql.Row {

	start := time.Now()

	// if tx exists
	if b.tx != nil {
		return b.tx.QueryRow(stmt, args...)
	}
	row := b.driver.Connection().QueryRow(stmt, args...)

	b.log(stmt, time.Since(start), args)

	return row

}

// all, queries all rows by using the *db.Query method.
// If a builder tx is existing, it will be used.
func (b *Builder) all(stmt string, args []interface{}) (*sql.Rows, error) {
	start := time.Now()

	// if tx exists
	if b.tx != nil {
		return b.tx.Query(stmt, args...)
	}
	rows, err := b.driver.Connection().Query(stmt, args...)
	b.log(stmt, time.Since(start), args)

	return rows, err

}

// exec, executes the rendered sql query.
// If a builder tx is existing, it will be used.
// If its a batch function and the builder tx is nil, a tx will be created in background and it will be committed/rolled back automatically.
func (b *Builder) exec(stmt string, args [][]interface{}) ([]sql.Result, error) {

	var res []sql.Result
	var err error

	// create a transaction if batching
	if len(args) > 1 && b.tx == nil {
		fmt.Println("********ADDED AUTOCOMMIT!!!!")
		err = b.Tx()
		if err != nil {
			return nil, err
		}
		b.txAuto = true
	}

	for _, arg := range args {
		var r sql.Result
		var err error

		start := time.Now()

		if b.tx != nil {
			r, err = b.tx.Exec(stmt, arg...)
		} else {
			r, err = b.driver.Connection().Exec(stmt, arg...)
		}

		if err != nil {
			// checking against batch tx
			if b.txAuto {
				b.txAuto = false
				err = b.Rollback()
				if err != nil {
					return nil, err
				}
			}
			return nil, err
		}
		tx := ""
		if b.tx != nil {
			tx = fmt.Sprintf("with TX: %p ", b.tx)
		}
		b.log(tx+stmt, time.Since(start), args)

		res = append(res, r)
	}

	// batch auto commit
	if b.txAuto {
		b.txAuto = false
		err = b.Commit()
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

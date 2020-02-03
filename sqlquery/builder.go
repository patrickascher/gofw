// Package sqlquery is a simple SQL generator.
//
// See https://github.com/patrickascher/go-sql for more information and examples.
package sqlquery

import (
	"database/sql"
	"fmt"
	"time"
)

// Builder stores the *sql.DB and the Placeholder information.
// This information is needed to translate and execute the query.
type Builder struct {
	Adapter     *sql.DB
	conf        Database
	Placeholder *Placeholder
	debugger    Debugger
	quote       string
}

// Debug is printing the executed SQL query to the console.
// As default this is deactivated.
// It can be set over the Config.
func (b *Builder) debugging(s string) {

	if b.debugger != nil {
		b.debugger.Debug(s)
	}
}

// Config returns the Database interface
func (b *Builder) Config() Database {
	return b.conf
}

// NewBuilderFromConfig creates a new instance of the Builder.
// If the adapter configuration is wrong, an error will return
func NewBuilderFromConfig(conf Database) (*Builder, error) {
	//TODO check required config
	db, err := sql.Open(conf.Driver(), conf.DSN())
	if err != nil {
		return nil, err
	}

	return createBuilder(db, conf), nil
}

func createBuilder(db *sql.DB, conf Database) *Builder {
	b := &Builder{Adapter: db, conf: conf, Placeholder: conf.Placeholder()}
	b.quote = conf.QuoteCharacter()
	if conf.Debugger() != nil {
		b.debugger = conf.Debugger()
	}
	return b
}

// NewBuilderFromAdapter creates a new instance of the Builder.
// This means you can use an existing (global) *sql.DB instance.
func NewBuilderFromAdapter(db *sql.DB, conf Database) *Builder {
	//TODO check if required conf is set
	return createBuilder(db, conf)
}

func Raw(sql string) string {
	return "!" + sql
}

// Select creates a new sqlSelect instance.
// SEE Select
func (b *Builder) Select(from string) *Select {
	b.Placeholder.reset() //reset counter
	return &Select{builder: b, from: from, condition: &Condition{}}
}

// Insert creates a new sqlInsert instance.
// SEE Insert
func (b *Builder) Insert(into string) *Insert {
	b.Placeholder.reset() //reset counter
	return &Insert{builder: b, transaction: true, into: into, batch: 50}
}

// Update creates a new sqlUpdate instance.
// SEE Update
func (b *Builder) Update(table string) *Update {
	b.Placeholder.reset() //reset counter
	return &Update{builder: b, table: table, condition: &Condition{}}
}

// Delete creates a new sqlDelete instance.
// SEE Delete
func (b *Builder) Delete(from string) *Delete {
	b.Placeholder.reset() //reset counter
	return &Delete{builder: b, from: from, condition: &Condition{}}
}

// Information returns some table specific data.
// SEE Information
func (b *Builder) Information(table string) *Information {
	return &Information{builder: b, table: table}
}

// NewTx creates a new transaction
func (b *Builder) NewTx() (*sql.Tx, error) {
	tx, err := b.Adapter.Begin()
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CommitTx commits the given transaction
func (b *Builder) CommitTx(tx *sql.Tx) error {
	err := tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// first queries one Row by using the *db.QueryRow method.
func (b *Builder) first(stmt string, args []interface{}) (*sql.Row, error) {
	// start timer for debug
	t := time.Now()

	row := b.Adapter.QueryRow(stmt, args...)

	// debug message if debugging is required
	b.debugging(fmt.Sprintf("FIRST %s %v execution time %s", stmt, args, time.Since(t)))

	return row, nil
}

// first queries one Row by using the *db.QueryRow method with a transaction.
func (b *Builder) firstTx(tx *sql.Tx, stmt string, args []interface{}) (*sql.Row, error) {
	// start timer for debug
	t := time.Now()

	row := tx.QueryRow(stmt, args...)

	// debug message if debugging is required
	b.debugging(fmt.Sprintf("FIRSTTx %s %v execution time %s", stmt, args, time.Since(t)))

	return row, nil
}

// all queries all rows by using the *db.Query method.
func (b *Builder) all(stmt string, args []interface{}) (*sql.Rows, error) {
	// start timer for debug
	t := time.Now()

	rows, err := b.Adapter.Query(stmt, args...)

	// debug message if debugging is required
	b.debugging(fmt.Sprintf("ALL %s execution time %s %#v", stmt, time.Since(t), args))

	return rows, err
}

// all queries all rows by using the *db.Query method with a transaction.
func (b *Builder) allTx(tx *sql.Tx, stmt string, args []interface{}) (*sql.Rows, error) {
	// start timer for debug
	t := time.Now()

	var rows *sql.Rows
	var err error

	rows, err = tx.Query(stmt, args...)
	if err != nil {
		errTx := tx.Rollback()
		if errTx != nil {
			err = errTx
		}
	}

	// debug message if debugging is required
	b.debugging(fmt.Sprintf("ALLTx %s execution time %s %#v", stmt, time.Since(t), args))

	return rows, err
}

// exec executes the rendered sql query.
func (b *Builder) exec(stmt string, args [][]interface{}) ([]sql.Result, error) {

	var res []sql.Result
	var r sql.Result
	var err error

	// create a transaction if batching
	if len(args) > 1 {
		tx, err := b.NewTx()
		if err != nil {
			return nil, err
		}
		res, err = b.execTx(tx, stmt, args)
		if err != nil {
			return nil, err
		}
		err = b.CommitTx(tx)
		if err != nil {
			return nil, err
		}
		return res, nil
	}

	for _, arg := range args {
		// start timer for debug
		t := time.Now()

		r, err = b.Adapter.Exec(stmt, arg...)
		if err != nil {
			return nil, err
		}

		res = append(res, r)
		b.debugging(fmt.Sprintf("Exec %s %v execution time %s", stmt, arg, time.Since(t)))
	}

	return res, nil
}

// exec executes the rendered sql query with a transaction.
func (b *Builder) execTx(tx *sql.Tx, stmt string, args [][]interface{}) ([]sql.Result, error) {

	var res []sql.Result
	var r sql.Result
	var err error

	for _, arg := range args {
		// start timer for debug
		t := time.Now()

		r, err = tx.Exec(stmt, arg...)
		if err != nil {
			errTx := tx.Rollback()
			if errTx != nil {
				err = errTx
			}
		}

		if err != nil {
			return nil, err
		}

		res = append(res, r)
		b.debugging(fmt.Sprintf("ExecTx %s %v execution time %s", stmt, arg, time.Since(t)))
	}

	return res, nil
}

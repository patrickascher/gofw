package sqlquery

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

const BATCH = 50

// Insert type.
type Insert struct {
	builder *Builder

	into  string
	batch int

	lastIDColumn string
	lastIDPtr    interface{}

	columns   []string
	arguments [][]interface{}

	valueSets []map[string]interface{}
}

// Values set the column/value pair
func (s *Insert) Values(values []map[string]interface{}) *Insert {
	s.valueSets = values
	return s
}

// LastInsertedID gets the last id over different drivers
func (s *Insert) LastInsertedID(column string, ID interface{}) *Insert {
	s.lastIDColumn = column
	s.lastIDPtr = ID
	return s
}

// Batch sets the batch counter
func (s *Insert) Batch(b int) *Insert {
	s.batch = b
	return s
}

// Columns define a fixed column order
func (s *Insert) Columns(c ...string) *Insert {
	s.columns = []string{}
	s.columns = append(s.columns, c...)
	return s
}

// String returns the statement and arguments
// An error will return if the arguments and placeholders mismatch or no value was set.
func (s *Insert) String() (stmt string, args [][]interface{}, err error) {
	return s.render()
}

// Exec the sql query through the Builder.
// If its batching, a transaction is called.
// An error will return if the arguments and placeholders mismatch or no value was set.
func (s *Insert) Exec() ([]sql.Result, error) {
	stmt, args, err := s.render()
	if err != nil {
		return nil, err
	}

	if s.lastIDPtr != nil {
		return nil, s.execQueryWithLastId(stmt, args)
	}

	return s.builder.exec(stmt, args)
}

// execQueryWithLastId is modifying the query to get the lastInsertId over different drivers.
// Postgres for example is not returning the result.LastInsertId()
func (s *Insert) execQueryWithLastId(stmt string, args [][]interface{}) error {
	var err error
	if s.builder.conf.Driver == "postgres" {
		stmt = stmt + " RETURNING " + s.builder.quoteColumns(s.lastIDColumn)
		row := s.builder.first(stmt, args[0])
		var i int64
		err = row.Scan(&i)
		if err != nil {
			return err
		}
		reflect.ValueOf(s.lastIDPtr).Elem().SetInt(i)
	} else {
		var res []sql.Result
		res, err = s.builder.exec(stmt, args)
		if err != nil {
			return err
		}
		id, err := res[0].LastInsertId()
		if err != nil {
			return err
		}
		reflect.ValueOf(s.lastIDPtr).Elem().SetInt(id)
	}
	return nil
}

// render generates the sql query.
// An error will return if the arguments and placeholders mismatch or no value was set.
func (s *Insert) render() (stmt string, args [][]interface{}, err error) {

	// no value was set
	if len(s.valueSets) == 0 {
		return "", nil, ErrValueMissing
	}

	//add the columns from the value map
	s.columns = s.builder.addColumns(s.columns, s.valueSets[0])

	//add the arguments from the value map
	err = s.addArguments()
	if err != nil {
		return "", nil, err
	}

	selectStmt := "INSERT INTO " + s.builder.quoteColumns(s.into) + "(" + s.builder.quoteColumns(s.columns...) + ") VALUES "

	//set the value placeholders
	valueStmt := "(" + PLACEHOLDER + strings.Repeat(", "+PLACEHOLDER, len(s.columns)-1) + ")"
	if s.isBatched() {
		selectStmt += valueStmt + strings.Repeat(", "+valueStmt, s.batch-1)
	} else {
		selectStmt += valueStmt + strings.Repeat(", "+valueStmt, len(s.valueSets)-1)
	}

	return s.builder.replacePlaceholders(selectStmt, s.builder.driver.Placeholder()), s.arguments, nil
}

// isBatched checks if a batching is needed
func (s *Insert) isBatched() bool {
	if s.batch == 0 {
		s.batch = BATCH
	}
	return len(s.valueSets) > s.batch
}

// batching the arguments
func (s *Insert) batching() [][]interface{} {

	batchCount := s.batch * len(s.columns)

	var batches [][]interface{}

	for batchCount < len(s.arguments[0]) {
		s.arguments[0], batches = s.arguments[0][batchCount:], append(batches, s.arguments[0][0:batchCount:batchCount])
	}
	return append(batches, s.arguments[0])
}

// addArguments is adding the key/value pairs as argument.
// If the column does not exist in the column slice, an error will return.
// If the actual arguments are bigger as the batch value, the arguments are getting batched.
// Default value for the batch is set through the builder (50)
func (s *Insert) addArguments() error {
	//add all arguments
	var arguments []interface{}
	for _, valueSet := range s.valueSets {
		for _, column := range s.columns {
			if val, ok := valueSet[column]; ok {
				arguments = append(arguments, val)
			} else {
				return fmt.Errorf(ErrColumn.Error(), column)
			}
		}
	}
	s.arguments = append(s.arguments, arguments)

	if s.isBatched() {
		s.arguments = s.batching()
	}

	return nil
}

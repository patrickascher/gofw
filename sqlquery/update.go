package sqlquery

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// Update provides some features for a sql update
type Update struct {
	builder   *Builder
	table     string
	condition *Condition

	columns   []string
	arguments []interface{}

	valueSet map[string]interface{}
}

// Set the column/value pair
func (s *Update) Set(values map[string]interface{}) *Update {
	s.valueSet = values
	return s
}

// Condition adds a ptr to a existing condition.
func (s *Update) Condition(c *Condition) *Update {
	c.Reset(HAVING, LIMIT, ORDER, OFFSET, GROUP, ON)
	s.condition = c
	return s
}

// Columns define a fixed column order
func (s *Update) Columns(c ...string) *Update {
	s.columns = append(s.columns, c...)
	return s
}

// addArguments is adding the key/value pairs as argument.
// If the column does not exist in the column slice, an error will return.
func (s *Update) addArguments() error {
	//add all arguments
	for _, column := range s.columns {
		if val, ok := s.valueSet[strings.Replace(column, s.table+".", "", 1)]; ok {
			s.arguments = append(s.arguments, val)
		} else {
			return fmt.Errorf("sql: Column %v does not exist", column)
		}
	}
	return nil
}

// Where is a wrapper for Condition.Where.
// See: Condition.Where
func (s *Update) Where(stmt string, args ...interface{}) *Update {
	s.condition.Where(stmt, args...)
	return s
}

// render generates the sql query.
// An error will return if the arguments and placeholders mismatch or no value was set.
func (s *Update) render() (stmt string, args []interface{}, err error) {

	//no value is set
	if len(s.valueSet) == 0 {
		return "", []interface{}(nil), errors.New("sql: no Insert value is set")
	}

	//add the columns from the value map
	s.columns = s.builder.addColumns(s.columns, s.valueSet)

	//add the arguments from the value map
	s.addArguments()

	//columns to string
	columns := ""
	for _, c := range s.columns {
		if columns != "" {
			columns += ", "
		}
		columns += s.builder.QuoteIdentifier(c) + " = " + PLACEHOLDER
	}

	selectStmt := "UPDATE " + s.builder.QuoteIdentifier(s.table) + " SET " + columns
	selectStmt = s.builder.replacePlaceholders(selectStmt)

	conditionStmt, err := s.condition.render(s.builder.Placeholder)
	if err != nil {
		return "", []interface{}(nil), err

	}

	return selectStmt + conditionStmt, append(s.arguments, s.condition.arguments()...), nil
}

// String returns the statement and arguments
// An error will return if the arguments and placeholders mismatch or no value was set.
func (s *Update) String() (stmt string, args []interface{}, err error) {
	return s.render()
}

func (s *Update) stmtAndArgs() (string, [][]interface{}, error) {
	var args [][]interface{}
	stmt, arg, err := s.render()
	if err != nil {
		return "", nil, err
	}
	args = append(args, arg)
	return stmt, args, err
}

// Exec the sql query through the Builder.
// An error will return if the arguments and placeholders mismatch, no value was set or the sql query returns one
func (s *Update) Exec() (sql.Result, error) {
	stmt, args, err := s.stmtAndArgs()
	if err != nil {
		return nil, err
	}
	r, err := s.builder.exec(stmt, args)
	if err != nil {
		return nil, err
	}
	return r[0], err
}

// ExecTx is executing the query with a transaction.
// An error will return if the arguments and placeholders mismatch, no value was set or the sql query returns one
func (s *Update) ExecTx(tx *sql.Tx) (sql.Result, error) {
	stmt, args, err := s.stmtAndArgs()
	if err != nil {
		if errTx := tx.Rollback(); errTx != nil {
			return nil, errTx
		}
		return nil, err
	}
	r, err := s.builder.execTx(tx, stmt, args)
	if err != nil {
		return nil, err
	}
	return r[0], err
}

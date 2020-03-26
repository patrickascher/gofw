// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlquery

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// Error messages.
var (
	ErrValueMissing = errors.New("sqlquery: No value is set")
	ErrColumn       = errors.New("sqlquery: Column %v does not exist")
)

// Update type.
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
	c.Reset(ON, GROUP, HAVING, ORDER, LIMIT, OFFSET)
	s.condition = c
	return s
}

// Columns define a fixed column order
func (s *Update) Columns(c ...string) *Update {
	s.columns = []string{}
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
			return fmt.Errorf(ErrColumn.Error(), column)
		}
	}
	return nil
}

// Where - please see the Condition.Where documentation.
func (s *Update) Where(stmt string, args ...interface{}) *Update {
	s.condition.Where(stmt, args...)
	return s
}

// render generates the sql query.
// An error will return if the arguments and placeholders mismatch or no value was set.
func (s *Update) render() (stmt string, args []interface{}, err error) {

	//no value is set
	if len(s.valueSet) == 0 {
		return "", []interface{}(nil), ErrValueMissing
	}

	//add the columns from the value map
	s.columns = s.builder.addColumns(s.columns, s.valueSet)

	//add the arguments from the value map
	err = s.addArguments()
	if err != nil {
		return "", nil, err
	}

	//columns to string
	columns := ""
	for _, c := range s.columns {
		if columns != "" {
			columns += ", "
		}
		columns += s.builder.quoteColumns(c) + " = " + PLACEHOLDER
	}

	selectStmt := "UPDATE " + s.builder.quoteColumns(s.table) + " SET " + columns
	p := s.builder.driver.Placeholder()
	selectStmt = replacePlaceholders(selectStmt, p)
	conditionStmt, err := s.condition.render(p)
	if err != nil {
		return "", []interface{}(nil), err
	}
	if conditionStmt != "" {
		conditionStmt = " " + conditionStmt
	}

	return selectStmt + conditionStmt, append(s.arguments, s.condition.arguments()...), nil
}

// String returns the statement and arguments
// An error will return if the arguments and placeholders mismatch or no value was set.
func (s *Update) String() (stmt string, args []interface{}, err error) {
	return s.render()
}

// Exec the sql query through the Builder.
// An error will return if the arguments and placeholders mismatch, no value was set or the sql query returns one.
func (s *Update) Exec() (sql.Result, error) {
	stmt, args, err := convertArgumentsExtraSlice(s.render())
	if err != nil {
		return nil, err
	}
	r, err := s.builder.exec(stmt, args)
	if err != nil {
		return nil, err
	}
	return r[0], err
}

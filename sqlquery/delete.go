// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlquery

import (
	"database/sql"
)

// Delete type.
type Delete struct {
	builder   *Builder
	from      string
	condition *Condition
}

// Where - please see the Condition.Where documentation.
func (s *Delete) Where(stmt string, args ...interface{}) *Delete {
	s.condition.Where(stmt, args...)
	return s
}

// Condition adds your own condition to the stmt.
// Only WHERE conditions are allowed.
func (s *Delete) Condition(c *Condition) *Delete {
	c.Reset(HAVING, LIMIT, ORDER, OFFSET, GROUP, ON)
	s.condition = c
	return s
}

// render generates the sql query.
// An error will return if the arguments and placeholders mismatch.
func (s *Delete) render() (stmt string, args []interface{}, err error) {
	selectStmt := "DELETE FROM " + s.builder.quoteColumns(s.from)
	conditionStmt, err := s.condition.render(s.builder.driver.Placeholder())
	if err != nil {
		return "", nil, err
	}
	if conditionStmt != "" {
		conditionStmt = " " + conditionStmt
	}

	return selectStmt + conditionStmt, s.condition.arguments(), err
}

// String returns the statement and arguments.
// An error will return if the arguments and placeholders mismatch.
func (s *Delete) String() (stmt string, args []interface{}, err error) {
	return s.render()
}

// Exec the sql query through the Builder.
// An error will return if the arguments and placeholders mismatch or the sql.Exec creates with an error.
func (s *Delete) Exec() (sql.Result, error) {
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

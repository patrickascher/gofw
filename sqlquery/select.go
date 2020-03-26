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

// Allowed join types.
const (
	LEFT = iota + 1
	RIGHT
	INNER
)

// Error messages.
var (
	ErrJoinType = errors.New("sqlquery: wrong join type %#v is used")
)

type join struct {
	builder   *Builder
	joinType  int
	table     string
	condition *Condition
}

// render the join statement
// TODO the condition logic should maybe be changed. At the moment there is one condition for the whole select.
// TODO: It makes more sense if the ON relation has his own condition.
// TODO: In that case, this extra render method could be deleted.
func (j *join) render(c *Condition, p *Placeholder) (string, error) {
	t := ""
	switch j.joinType {
	case LEFT:
		t = "LEFT"
	case RIGHT:
		t = "RIGHT"
	case INNER:
		t = "INNER"
	default:
		return "", fmt.Errorf(ErrJoinType.Error(), j.joinType)
	}

	//render the condition stmt
	space := ""
	if j.condition.on != "" {
		space = " "
	}
	stmt := t + " JOIN " + j.builder.quoteColumns(j.table) + space + j.condition.on

	// testing if the condition itself is correct.
	_, err := j.condition.render(p)
	if err != nil {
		return "", err
	}
	//check if there are some map arguments
	stmt = stmtMapManipulation(c, stmt, j.condition.args[ON], ON)

	for i := 1; i <= strings.Count(j.condition.on, PLACEHOLDER); i++ {
		stmt = strings.Replace(stmt, PLACEHOLDER, p.placeholder(), 1)
	}

	return stmt, nil
}

// Select type.
type Select struct {
	builder *Builder

	columns []string
	join    []join
	from    string

	condition *Condition
}

// Columns set new columns to the select stmt.
// If no columns are added, the * will be used.
func (s *Select) Columns(cols ...string) *Select {
	s.columns = []string{}
	s.columns = append(s.columns, cols...)
	return s
}

// Join - please see the Condition.Join documentation.
func (s *Select) Join(joinType int, table string, condition *Condition) *Select {
	if joinType == 0 || table == "" {
		return s
	}
	s.join = append(s.join, join{builder: s.builder, joinType: joinType, table: table, condition: condition})
	return s
}

// Where - please see the Condition.Where documentation.
func (s *Select) Where(stmt string, args ...interface{}) *Select {
	s.condition.Where(stmt, args...)
	return s
}

// Group - please see the Condition.Group documentation.
func (s *Select) Group(group ...string) *Select {
	s.condition.Group(group...)
	return s
}

// Having - please see the Condition.Having documentation.
func (s *Select) Having(stmt string, args ...interface{}) *Select {
	s.condition.Having(stmt, args...)
	return s
}

// Order - please see the Condition.Order documentation.
func (s *Select) Order(order ...string) *Select {
	s.condition.Order(order...)
	return s
}

// Limit - please see the Condition.Limit documentation.
func (s *Select) Limit(l int) *Select {
	s.condition.Limit(l)
	return s
}

// Offset - please see the Condition.Offset documentation.
func (s *Select) Offset(l int) *Select {
	s.condition.Offset(l)
	return s
}

// Condition adds your own condition to the stmt.
func (s *Select) Condition(c *Condition) *Select {
	c.Reset(ON)
	s.condition = c
	return s
}

// render generates the sql query.
// An error will return if the arguments and placeholders mismatch.
func (s *Select) render() (string, []interface{}, error) {

	columns := s.builder.quoteColumns(s.columns...)
	if columns == "" {
		columns = "*"
	}

	selectStmt := "SELECT " + columns + " FROM " + s.builder.quoteColumns(s.from)

	if len(s.join) > 0 {
		for _, j := range s.join {
			joinStmt, err := j.render(s.condition, s.builder.driver.Placeholder())
			if err != nil {
				return "", nil, err
			}
			if joinStmt != "" {
				joinStmt = " " + joinStmt
			}
			selectStmt = selectStmt + joinStmt
		}
	}
	conditionStmt, err := s.condition.render(s.builder.driver.Placeholder())
	if conditionStmt != "" {
		conditionStmt = " " + conditionStmt
	}

	return selectStmt + conditionStmt, s.condition.arguments(), err
}

// First will return only one row.
// Its a wrapper for DB.QueryRow
func (s Select) First() (*sql.Row, error) {
	stmt, args, err := s.render()
	if err != nil {
		return nil, err
	}
	return s.builder.first(stmt, args), nil
}

// All returns the found rows by using the *db.Query method.
// All returns a *sql.Rows, dont forget to Close it!
// An error will return if the arguments and placeholders mismatch.
func (s Select) All() (*sql.Rows, error) {
	stmt, args, err := s.render()
	if err != nil {
		return nil, err
	}
	return s.builder.all(stmt, args)
}

// String returns the statement and arguments.
// An error will return if the arguments and placeholders mismatch.
func (s Select) String() (string, []interface{}, error) {
	return s.render()
}

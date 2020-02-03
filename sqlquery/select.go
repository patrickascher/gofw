package sqlquery

import (
	"database/sql"
	"fmt"
	"strings"
)

//Exported JOIN types
const (
	LEFT = iota + 1
	RIGHT
	INNER
)

type join struct {
	builder   *Builder
	joinType  int
	table     string
	condition *Condition
}

//render the join statement
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
		return "", fmt.Errorf("sql: wrong join type %#v is used", j.joinType)
	}

	//render the condition stmt
	stmt := t + " JOIN " + j.builder.QuoteIdentifier(j.table) + j.condition.on

	//check if there are some map arguments
	stmt = stmtMapManipulation(c, stmt, j.condition.args[ON], ON)

	for i := 1; i <= strings.Count(j.condition.on, PLACEHOLDER); i++ {
		stmt = strings.Replace(stmt, PLACEHOLDER, p.placeholder(), 1)
	}

	return stmt, nil
}

// Select provides some features for a sql select
type Select struct {
	builder *Builder

	columns []string
	join    []join
	from    string

	condition *Condition
}

// Columns appending
func (s *Select) Columns(cols ...string) *Select {
	s.columns = append(s.columns, cols...)
	return s
}

// Join is a wrapper for Condition.Join.
// See: Condition.Join
func (s *Select) Join(joinType int, table string, condition *Condition) *Select {
	s.join = append(s.join, join{builder: s.builder, joinType: joinType, table: table, condition: condition})
	return s
}

// Condition adds a ptr to a existing condition.
func (s *Select) Condition(c *Condition) *Select {
	c.Reset(ON)
	s.condition = c
	return s
}

// Where is a wrapper for Condition.Where.
// See: Condition.Where
func (s *Select) Where(stmt string, args ...interface{}) *Select {
	s.condition.Where(stmt, args...)
	return s
}

// Group is a wrapper for Condition.Group.
// See: Condition.Group
func (s *Select) Group(group ...string) *Select {
	s.condition.Group(group...)
	return s
}

// Having is a wrapper for Condition.Having.
// See: Condition.Having
func (s *Select) Having(stmt string, args ...interface{}) *Select {
	s.condition.Having(stmt, args...)
	return s
}

// Order is a wrapper for Condition.Order.
// See: Condition.Order
func (s *Select) Order(order ...string) *Select {
	s.condition.Order(order...)
	return s
}

// Limit is a wrapper for Condition.Limit.
// See: Condition.Limit
func (s *Select) Limit(l int) *Select {
	s.condition.Limit(l)
	return s
}

// Offset is a wrapper for Condition.Offset.
// See: Condition.Offset
func (s *Select) Offset(l int) *Select {
	s.condition.Offset(l)
	return s
}

// render generates the sql query.
// An error will return if the arguments and placeholders mismatch.
func (s *Select) render() (string, []interface{}, error) {
	columns := s.builder.escapeColumns(s.columns)
	if columns == "" {
		columns = "*"
	}
	selectStmt := "SELECT " + columns + " FROM " + s.builder.QuoteIdentifier(s.from)

	if len(s.join) > 0 {
		for _, j := range s.join {
			joinStmt, err := j.render(s.condition, s.builder.Placeholder)
			if err != nil {
				return "", nil, err
			}
			selectStmt = selectStmt + " " + joinStmt
		}
	}

	conditionStmt, err := s.condition.render(s.builder.Placeholder)

	return selectStmt + conditionStmt, s.condition.arguments(), err
}

// First will return only one row.
// Its a wrapper for DB.QueryRow
func (s *Select) First() (*sql.Row, error) {
	//s.Limit(1).Offset(0)
	stmt, args, err := s.render()
	if err != nil {
		return nil, err
	}
	return s.builder.first(stmt, args)
}

// FirstTx will return only one row, it is executing the query with a transaction.
// Its a wrapper for DB.QueryRow
func (s *Select) FirstTx(tx *sql.Tx) (*sql.Row, error) {
	//s.Limit(1).Offset(0)
	stmt, args, err := s.render()
	if err != nil {
		if errTx := tx.Rollback(); errTx != nil {
			return nil, errTx
		}
		return nil, err
	}
	return s.builder.firstTx(tx, stmt, args)
}

// String returns the statement and arguments
// An error will return if the arguments and placeholders mismatch.
func (s *Select) String() (string, []interface{}, error) {
	return s.render()
}

// All queries all rows by using the *db.Query method.
// An error will return if the arguments and placeholders mismatch.
func (s *Select) All() (*sql.Rows, error) {
	stmt, args, err := s.render()
	if err != nil {
		return nil, err
	}
	return s.builder.all(stmt, args)
}

// AllTx queries all rows by using the *db.Query method, it is executing the query with a transaction.
// An error will return if the arguments and placeholders mismatch.
func (s *Select) AllTx(tx *sql.Tx) (*sql.Rows, error) {
	stmt, args, err := s.render()
	if err != nil {
		if errTx := tx.Rollback(); errTx != nil {
			return nil, errTx
		}
		return nil, err
	}
	return s.builder.allTx(tx, stmt, args)
}

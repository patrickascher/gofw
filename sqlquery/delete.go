package sqlquery

import (
	"database/sql"
)

// Delete provides some features for a sql delete
type Delete struct {
	builder   *Builder
	from      string
	condition *Condition
}

// Where is a wrapper for Condition.Where.
// See: Condition.Where
func (s *Delete) Where(stmt string, args ...interface{}) *Delete {
	s.condition.Where(stmt, args...)
	return s
}

// Condition adds a ptr to a existing condition.
func (s *Delete) Condition(c *Condition) *Delete {
	c.Reset(HAVING, LIMIT, ORDER, OFFSET, GROUP, ON)
	s.condition = c
	return s
}

// render generates the sql query.
// An error will return if the arguments and placeholders mismatch.
func (s *Delete) render() (stmt string, args []interface{}, err error) {
	selectStmt := "DELETE FROM " + s.builder.QuoteIdentifier(s.from)
	conditionStmt, err := s.condition.render(s.builder.Placeholder)
	return selectStmt + conditionStmt, s.condition.arguments(), err
}

// String returns the statement and arguments
// An error will return if the arguments and placeholders mismatch.
func (s *Delete) String() (stmt string, args []interface{}, err error) {
	return s.render()
}

func (s *Delete) stmtAndArgs() (string, [][]interface{}, error) {
	var args [][]interface{}
	stmt, arg, err := s.render()
	if err != nil {
		return "", nil, err
	}
	args = append(args, arg)
	return stmt, args, err
}

// Exec the sql query through the Builder
// An error will return if the arguments and placeholders mismatch or the sql.Exec creates with an error.
func (s *Delete) Exec() (sql.Result, error) {
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
// An error will return if the arguments and placeholders mismatch or the sql.Exec creates with an error.
func (s *Delete) ExecTx(tx *sql.Tx) (sql.Result, error) {
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

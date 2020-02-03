package sqlquery

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSqlSelect_Columns(t *testing.T) {
	s := Select{}
	s.Columns("id", "name")
	assert.Equal(t, []string{"id", "name"}, s.columns)
}

// TestSqlSelect_Conditions_Select_Join checks the condition, select and join render
func TestSqlSelect_Conditions_Select_Join(t *testing.T) {
	b, err := HelperCreateBuilder()
	if assert.NoError(t, err) {
		joinC := Condition{}
		s := b.Select(TABLE).Columns("id", "name").Join(LEFT, "parts", joinC.On("robots.id = parts.robot_id AND id != ?", 2)).Where("id = ?", 1).Group("id", "name").Having("name = ?", "Wall-E").Order("id", "-name").Limit(5).Offset(1)
		stmt, errStmt := s.condition.render(b.Placeholder)
		if assert.NoError(t, errStmt) {
			if b.Placeholder.Numeric {
				assert.Equal(t, " WHERE id = "+b.Placeholder.Char+"1 GROUP BY id, name HAVING name = "+b.Placeholder.Char+"2 ORDER BY id ASC, name DESC LIMIT 5 OFFSET 1", stmt)
			} else {
				assert.Equal(t, " WHERE id = "+b.Placeholder.Char+" GROUP BY id, name HAVING name = "+b.Placeholder.Char+" ORDER BY id ASC, name DESC LIMIT 5 OFFSET 1", stmt)
			}
			assert.Equal(t, map[int][]interface{}(map[int][]interface{}{WHERE: {1}, HAVING: {"Wall-E"}}), s.condition.args)
		}

		// select render + left join
		b.Placeholder.reset()
		sel, args, errRender := s.render()
		if assert.NoError(t, errRender) {
			if b.Placeholder.Numeric {
				assert.Equal(t, "SELECT "+b.QuoteIdentifier("id")+", "+b.QuoteIdentifier("name")+" FROM "+b.QuoteIdentifier(TABLE)+" LEFT JOIN "+b.QuoteIdentifier("parts")+" ON robots.id = parts.robot_id AND id != "+b.Placeholder.Char+"1 WHERE id = "+b.Placeholder.Char+"2 GROUP BY id, name HAVING name = "+b.Placeholder.Char+"3 ORDER BY id ASC, name DESC LIMIT 5 OFFSET 1", sel)
			} else {
				assert.Equal(t, "SELECT "+b.QuoteIdentifier("id")+", "+b.QuoteIdentifier("name")+" FROM "+b.QuoteIdentifier(TABLE)+" LEFT JOIN "+b.QuoteIdentifier("parts")+" ON robots.id = parts.robot_id AND id != "+b.Placeholder.Char+" WHERE id = "+b.Placeholder.Char+" GROUP BY id, name HAVING name = "+b.Placeholder.Char+" ORDER BY id ASC, name DESC LIMIT 5 OFFSET 1", sel)
			}
			assert.Equal(t, []interface{}([]interface{}{2, 1, "Wall-E"}), args)
		}

		// Right join
		b.Placeholder.reset()
		joinC = Condition{}
		s = b.Select(TABLE).Columns("id", "name").Join(RIGHT, "parts", joinC.On("robots.id = parts.robot_id"))
		stmt, Rargs, errStmt := s.render()
		if assert.NoError(t, errStmt) {
			assert.Equal(t, "SELECT "+b.QuoteIdentifier("id")+", "+b.QuoteIdentifier("name")+" FROM "+b.QuoteIdentifier(TABLE)+" RIGHT JOIN "+b.QuoteIdentifier("parts")+" ON robots.id = parts.robot_id", stmt)
			assert.Equal(t, []interface{}(nil), Rargs)
		}

		// Inner join
		b.Placeholder.reset()
		joinC = Condition{}
		s = b.Select(TABLE).Columns("id", "name").Join(INNER, "parts", joinC.On("robots.id = parts.robot_id"))
		stmt, Iargs, errStmt := s.render()
		if assert.NoError(t, errStmt) {
			assert.Equal(t, "SELECT "+b.QuoteIdentifier("id")+", "+b.QuoteIdentifier("name")+" FROM "+b.QuoteIdentifier(TABLE)+" INNER JOIN "+b.QuoteIdentifier("parts")+" ON robots.id = parts.robot_id", stmt)
			assert.Equal(t, []interface{}(nil), Iargs)
		}

		// Join type does not exist
		s = b.Select(TABLE).Columns("id", "name").Join(100, "parts", joinC.On("robots.id = parts.robot_id"))
		_, _, errStmt = s.render()
		assert.Error(t, errStmt)
	}
}

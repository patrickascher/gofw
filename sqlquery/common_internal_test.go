package sqlquery

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuilder_replacePlaceholders(t *testing.T) {
	b, err := HelperCreateBuilder()
	if assert.NoError(t, err) {

		stmt := "SELECT ? FROM ? WHERE ? = ?"
		expStmt := "SELECT " + b.Placeholder.Char + " FROM " + b.Placeholder.Char + " WHERE " + b.Placeholder.Char + " = " + b.Placeholder.Char
		if b.Placeholder.Numeric {
			expStmt = "SELECT " + b.Placeholder.Char + "1 FROM " + b.Placeholder.Char + "2 WHERE " + b.Placeholder.Char + "3 = " + b.Placeholder.Char + "4"
		}

		assert.Equal(t, expStmt, b.replacePlaceholders(stmt))
	}
}

func TestBuilder_addColumns(t *testing.T) {
	//columns []string, values map[string]interface{}
	b, err := HelperCreateBuilder()
	if assert.NoError(t, err) {
		cols := []string{"id", "name"}
		val := map[string]interface{}{"created": "", "deleted": ""}
		assert.Equal(t, cols, b.addColumns(cols, val))
		//assert.Equal(t,[]string{"created","deleted"},b.addColumns([]string{},val)) //Had to get changed, its a map and this doesn't guarantee the order
		valInternal := b.addColumns([]string{}, val)
		assert.Contains(t, valInternal, "created")
		assert.Contains(t, valInternal, "deleted")
	}
}

func TestBuilder_escapeColumns(t *testing.T) {
	b, err := HelperCreateBuilder()
	if assert.NoError(t, err) {
		cols := []string{"id", "name", "robots.name"}
		assert.Equal(t, b.quote+"id"+b.quote+", "+b.quote+"name"+b.quote+", "+b.quote+"robots"+b.quote+"."+b.quote+"name"+b.quote, b.escapeColumns(cols))
	}
}

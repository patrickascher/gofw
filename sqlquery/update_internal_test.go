package sqlquery

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSqlUpdate_Columns(t *testing.T) {
	b, err := HelperCreateBuilder()
	if assert.NoError(t, err) {
		u := b.Update(TABLE)
		u.Columns("id", "name")
		assert.Equal(t, []string{"id", "name"}, u.columns)
	}
}

func TestSqlUpdate_Set(t *testing.T) {
	values := map[string]interface{}{"name": "Wall-E"}

	b, err := HelperCreateBuilder()
	if assert.NoError(t, err) {
		u := b.Update(TABLE)
		u.Set(values)
		assert.Equal(t, map[string]interface{}{"name": "Wall-E"}, u.valueSet)
	}
}

func TestSqlUpdate_addArguments(t *testing.T) {
	values := map[string]interface{}{"id": 1, "name": "Wall-E"}

	// all given fields form the set
	b, err := HelperCreateBuilder()
	if assert.NoError(t, err) {
		u := b.Update(TABLE)
		u.Columns("id", "name") // this is set special for the test. The arguments is a map, so we can not guarantee the order here.

		u.Set(values)
		u.columns = b.addColumns(u.columns, u.valueSet)

		errArg := u.addArguments()
		assert.NoError(t, errArg)
		assert.Equal(t, []interface{}{1, "Wall-E"}, u.arguments)
	}

	// whitespace given fields form the set
	b, err = HelperCreateBuilder()
	if assert.NoError(t, err) {
		u := b.Update(TABLE)
		u.Columns("id")
		u.Set(values)
		u.columns = b.addColumns(u.columns, u.valueSet)

		errArg := u.addArguments()
		assert.NoError(t, errArg)
		assert.Equal(t, []interface{}{1}, u.arguments)
	}

	// find column also with the table name and replace it with only the column name, otherwise we cannot find it in the value map
	b, err = HelperCreateBuilder()
	if assert.NoError(t, err) {
		u := b.Update(TABLE)
		u.Columns(TABLE + ".id")
		u.Set(values)
		u.columns = b.addColumns(u.columns, u.valueSet)

		errArg := u.addArguments()
		assert.NoError(t, errArg)
		assert.Equal(t, []interface{}{1}, u.arguments)
	}

	// manual added column does not exist
	b, err = HelperCreateBuilder()
	if assert.NoError(t, err) {
		u := b.Update(TABLE)
		u.Columns("id", "doesNotExist")
		u.Set(values)
		u.columns = b.addColumns(u.columns, u.valueSet)

		errArg := u.addArguments()
		assert.Error(t, errArg)
	}
}

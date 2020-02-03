package sqlquery

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSqlInsert_internal_addArguments(t *testing.T) {
	var values []map[string]interface{}
	values = append(values, map[string]interface{}{"id": 10, "name": "Combot"})
	values = append(values, map[string]interface{}{"id": 11, "name": "Uxptron"})

	// add all arguments
	s := Insert{}
	s.valueSets = values
	assert.Equal(t, [][]interface{}([][]interface{}(nil)), s.arguments)
	s.columns = s.builder.addColumns(s.columns, s.valueSets[0])
	err := s.addArguments()
	assert.NoError(t, err)
	//assert.Equal(t, [][]interface{}{{10, "Combot", 11, "Uxptron"}}, s.arguments) - We can not guarantee a fixed order if the columns are not set manually
	assert.Contains(t, s.arguments[0], "Combot")
	assert.Contains(t, s.arguments[0], "Uxptron")
	assert.Contains(t, s.arguments[0], 11)
	assert.Contains(t, s.arguments[0], 10)

	// add only the name arguments
	s = Insert{}
	s.valueSets = values
	assert.Equal(t, [][]interface{}([][]interface{}(nil)), s.arguments)
	s.columns = []string{"name"}
	err = s.addArguments()
	assert.NoError(t, err)
	assert.Equal(t, [][]interface{}([][]interface{}{{"Combot", "Uxptron"}}), s.arguments)

	// manual added column does not exist in value set
	s = Insert{}
	s.valueSets = values
	assert.Equal(t, [][]interface{}([][]interface{}(nil)), s.arguments)
	s.columns = []string{"doesNotExist"}
	err = s.addArguments()
	assert.Error(t, err)
	assert.Equal(t, [][]interface{}(nil), s.arguments)
}

func TestSqlInsert_internal_render(t *testing.T) {
	var values []map[string]interface{}
	values = append(values, map[string]interface{}{"id": 10, "name": "Combot"})
	values = append(values, map[string]interface{}{"id": 11, "name": "Uxptron"})

	// render without value set
	s := Insert{}
	_, _, err := s.render()
	assert.Error(t, err)

	// error because of mismatch manually added columns and value set
	s = Insert{}
	s.columns = []string{"doesNotExist"}
	s.valueSets = values
	_, _, err = s.render()
	assert.Error(t, err)

	// normal insert stmt
	s = Insert{}
	b, err := HelperCreateBuilder()
	if assert.NoError(t, err) {
		s.builder = b
		s.valueSets = values
		s.into = TABLE
		s.Columns("id", "name") // had to use because otherwise i can not guarantee the order in the tests
		stmt, args, errRender := s.render()
		assert.NoError(t, errRender)
		if b.Placeholder.Numeric {
			assert.Equal(t, "INSERT INTO "+b.QuoteIdentifier(TABLE)+"("+b.QuoteIdentifier("id")+", "+b.QuoteIdentifier("name")+") VALUES ("+b.Placeholder.Char+"1, "+b.Placeholder.Char+"2), ("+b.Placeholder.Char+"3, "+b.Placeholder.Char+"4)", stmt)
		} else {
			assert.Equal(t, "INSERT INTO "+b.QuoteIdentifier(TABLE)+"("+b.QuoteIdentifier("id")+", "+b.QuoteIdentifier("name")+") VALUES ("+b.Placeholder.Char+", "+b.Placeholder.Char+"), ("+b.Placeholder.Char+", "+b.Placeholder.Char+")", stmt)
		}
		assert.Equal(t, [][]interface{}([][]interface{}{{10, "Combot", 11, "Uxptron"}}), args)
	}

	// normal insert stmt batched
	s = Insert{}
	b, err = HelperCreateBuilder()
	if assert.NoError(t, err) {
		s.builder = b
		s.valueSets = values
		s.into = TABLE
		s.batch = 1
		s.Columns("id", "name") // had to use because otherwise i can not guarantee the order in the tests
		stmt, args, errRender := s.render()
		assert.NoError(t, errRender)
		if b.Placeholder.Numeric {
			assert.Equal(t, "INSERT INTO "+b.QuoteIdentifier(TABLE)+"("+b.QuoteIdentifier("id")+", "+b.QuoteIdentifier("name")+") VALUES ("+b.Placeholder.Char+"1, "+b.Placeholder.Char+"2)", stmt)
		} else {
			assert.Equal(t, "INSERT INTO "+b.QuoteIdentifier(TABLE)+"("+b.QuoteIdentifier("id")+", "+b.QuoteIdentifier("name")+") VALUES ("+b.Placeholder.Char+", "+b.Placeholder.Char+")", stmt)
		}
		assert.Equal(t, [][]interface{}{{10, "Combot"}, {11, "Uxptron"}}, args)
	}

	// normal insert whitespace column - only name should be added
	s = Insert{}
	b, err = HelperCreateBuilder()
	if assert.NoError(t, err) {
		s.builder = b
		s.columns = []string{"name"}
		s.valueSets = values
		s.into = TABLE
		s.batch = 1
		stmt, args, errRender := s.render()
		assert.NoError(t, errRender)
		if b.Placeholder.Numeric {
			assert.Equal(t, "INSERT INTO "+b.QuoteIdentifier(TABLE)+"("+b.QuoteIdentifier("name")+") VALUES ("+b.Placeholder.Char+"1)", stmt)
		} else {
			assert.Equal(t, "INSERT INTO "+b.QuoteIdentifier(TABLE)+"("+b.QuoteIdentifier("name")+") VALUES ("+b.Placeholder.Char+")", stmt)
		}
		assert.Equal(t, [][]interface{}{{"Combot"}, {"Uxptron"}}, args)
	}
}

func TestSqlInsert_internal_isBatched(t *testing.T) {
	s := Insert{}
	s.valueSets = []map[string]interface{}{{"name": "Cozmo"}, {"name": "Wall-E"}}

	assert.Equal(t, false, s.isBatched())
	s.batch = 1

	assert.Equal(t, true, s.isBatched())
}

func TestSqlInsert_internal_batching(t *testing.T) {
	var values []map[string]interface{}
	values = append(values, map[string]interface{}{"name": "Combot"})
	values = append(values, map[string]interface{}{"name": "Uxptron"})

	//batched
	s := Insert{}
	b, err := HelperCreateBuilder()
	if assert.NoError(t, err) {
		s.builder = b
		s.into = TABLE
		s.columns = []string{"name"}
		s.valueSets = values
		s.batch = 1

		stmt, args, errRender := s.render()
		if assert.NoError(t, errRender) {
			if b.Placeholder.Numeric {
				assert.Equal(t, "INSERT INTO "+b.QuoteIdentifier("robots")+"("+b.QuoteIdentifier("name")+") VALUES ("+b.Placeholder.Char+"1)", stmt)
			} else {
				assert.Equal(t, "INSERT INTO "+b.QuoteIdentifier("robots")+"("+b.QuoteIdentifier("name")+") VALUES ("+b.Placeholder.Char+")", stmt)
			}
			assert.Equal(t, [][]interface{}([][]interface{}{{"Combot"}, {"Uxptron"}}), args)
		}
	}

	//not batched
	s = Insert{}
	b, err = HelperCreateBuilder()
	if assert.NoError(t, err) {
		s.builder = b
		s.into = TABLE
		s.columns = []string{"name"}
		s.valueSets = values
		s.batch = 5

		stmt, args, errRender := s.render()
		if assert.NoError(t, errRender) {
			if b.Placeholder.Numeric {
				assert.Equal(t, "INSERT INTO "+b.QuoteIdentifier("robots")+"("+b.QuoteIdentifier("name")+") VALUES ("+b.Placeholder.Char+"1), ("+b.Placeholder.Char+"2)", stmt)
			} else {
				assert.Equal(t, "INSERT INTO "+b.QuoteIdentifier("robots")+"("+b.QuoteIdentifier("name")+") VALUES ("+b.Placeholder.Char+"), ("+b.Placeholder.Char+")", stmt)
			}
			assert.Equal(t, [][]interface{}{{"Combot", "Uxptron"}}, args)
		}
	}
}

func TestSqlInsert_Batch(t *testing.T) {
	s := Insert{}
	assert.Equal(t, 0, s.batch)
	s.Batch(1)
	assert.Equal(t, 1, s.batch)
}

func TestSqlInsert_Values(t *testing.T) {

	var values []map[string]interface{}
	values = append(values, map[string]interface{}{"name": "Combot"})
	values = append(values, map[string]interface{}{"name": "Uxptron"})

	s := Insert{}
	assert.Equal(t, []map[string]interface{}([]map[string]interface{}(nil)), s.valueSets)
	s.Values(values)
	assert.Equal(t, []map[string]interface{}([]map[string]interface{}{{"name": "Combot"}, {"name": "Uxptron"}}), s.valueSets)
}

func TestSqlInsert_Columns(t *testing.T) {
	s := Insert{}
	assert.Equal(t, []string([]string(nil)), s.columns)
	s.Columns("id", "name")
	assert.Equal(t, []string{"id", "name"}, s.columns)
}

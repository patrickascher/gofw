package orm

import (
	"fmt"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTable_getDatabaseAndTableByString(t *testing.T) {

	db, table := getDatabaseAndTableByString(" customers ")
	assert.Equal(t, "", db)
	assert.Equal(t, "customers", table)

	db, table = getDatabaseAndTableByString("tests.customers")
	assert.Equal(t, "tests", db)
	assert.Equal(t, "customers", table)

}

func TestTable_columnNames_columnByName_columnNamesWithoutAutoincrement(t *testing.T) {
	a := &Column{Permission: Permission{Read: true, Write: true}, Information: &sqlquery_.Column{Name: "a", PrimaryKey: true}}
	b := &Column{Permission: Permission{Read: true, Write: true}, Information: &sqlquery_.Column{Name: "b", Autoincrement: true}}
	c := &Column{Permission: Permission{Read: true, Write: true}, Information: &sqlquery_.Column{Name: "c", PrimaryKey: true}}
	table := Table{Cols: []*Column{
		a,
		b,
		c,
	}}

	// testing columnNames
	assert.Equal(t, []string{"a", "b", "c"}, table.columnNames(READDB, false))

	// testing columnByName
	col, err := table.columnByName("a")
	assert.NoError(t, err)
	assert.Equal(t, a, col)

	col, err = table.columnByName("b")
	assert.NoError(t, err)
	assert.Equal(t, b, col)

	col, err = table.columnByName("c")
	assert.NoError(t, err)
	assert.Equal(t, c, col)

	// Error because column does not exist
	col, err = table.columnByName("d")
	assert.Error(t, err)
	assert.True(t, col == nil)

	// testing columnNamesWithoutAutoincrement
	assert.Equal(t, []string{"a", "c"}, table.columnNamesWithoutAutoincrement())

	// testing primaryKeys
	pkeys := table.PrimaryKeys()
	assert.Equal(t, 2, len(pkeys))
	assert.Equal(t, a, pkeys[0])
	assert.Equal(t, c, pkeys[1])
}

func TestTable_describe(t *testing.T) {

	a := &Column{Permission: Permission{Write: false, Read: false}, StructField: "ID", Information: &sqlquery_.Column{Name: "id", PrimaryKey: true}}
	b := &Column{Permission: Permission{Write: false, Read: false}, Information: &sqlquery_.Column{Name: "first_name", Autoincrement: true}}
	c := &Column{Permission: Permission{Write: false, Read: false}, Information: &sqlquery_.Column{Name: "last_name", PrimaryKey: true}}
	d := &Column{Permission: Permission{Write: false, Read: false}, Information: &sqlquery_.Column{Name: "created_at", PrimaryKey: true}}
	e := &Column{Permission: Permission{Write: false, Read: false}, Information: &sqlquery_.Column{Name: "updated_at", PrimaryKey: true}}
	f := &Column{Permission: Permission{Write: false, Read: false}, Information: &sqlquery_.Column{Name: "deleted_at", PrimaryKey: true}}
	g := &Column{Permission: Permission{Write: false, Read: false}, Information: &sqlquery_.Column{Name: "not_existing", PrimaryKey: true}}

	builder, err := HelperCreateBuilder()
	assert.NoError(t, err)

	// does not exist
	table := Table{Builder: builder, Database: "xxx", Name: "customers", Cols: []*Column{}}
	err = table.describe()
	assert.Error(t, err)

	// ok everything exists
	table = Table{Builder: builder, Database: "tests", Name: "customers", Cols: []*Column{
		a,
		b,
		c,
		d,
		e,
		f,
	}}
	err = table.describe()
	assert.NoError(t, err)
	//TODO TESTS

	// error - column "not_existing" does not exist in the table
	table = Table{Builder: builder, Database: "tests", Name: "customers", Cols: []*Column{
		a,
		b,
		c,
		d,
		e,
		f,
		g,
	}}
	err = table.describe()
	assert.NoError(t, err)
	assert.Equal(t, false, g.Permission.Write)
	assert.Equal(t, false, g.Permission.Read)

	// error - no model field ID is set
	table = Table{Builder: builder, Database: "tests", Name: "customers", Cols: []*Column{
		b,
		c,
		d,
		e,
		f,
	}}
	err = table.describe()
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf(ErrIDPrimary.Error(), "customers"), err.Error())

}

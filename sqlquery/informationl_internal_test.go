package sqlquery

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInformation_splitDatabaseAndTable(t *testing.T) {

	db, table := splitDatabaseAndTable("db.table")
	assert.Equal(t, "db", db)
	assert.Equal(t, "table", table)

	db, table = splitDatabaseAndTable("table")
	assert.Equal(t, "", db)
	assert.Equal(t, "table", table)
}

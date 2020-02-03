package sqlquery_test

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSqlInsert_Exec(t *testing.T) {
	var values []map[string]interface{}
	values = append(values, map[string]interface{}{"id": 10, "name": "Combot"})
	values = append(values, map[string]interface{}{"id": 11, "name": "Uxptron"})

	b, err := HelperCreateBuilder()
	assert.NoError(t, HelperDeleteEntries(b))

	// Normal Insert
	if assert.NoError(t, err) {
		res, errInsert := b.Insert(TABLE).Columns("name").Values(values).Exec()
		if assert.NoError(t, errInsert) {
			num, errRows := res[0].RowsAffected()
			if assert.NoError(t, errRows) {
				assert.Equal(t, int64(2), num)
			}
		}
	}

	// Batched Insert
	if assert.NoError(t, err) {
		res, errInsert := b.Insert(TABLE).Columns("name").Batch(1).Values(values).Exec()
		if assert.NoError(t, errInsert) {
			num, errRows := res[0].RowsAffected()
			if assert.NoError(t, errRows) {
				assert.Equal(t, int64(1), num)
			}
			num, errRows = res[1].RowsAffected()
			if assert.NoError(t, errRows) {
				assert.Equal(t, int64(1), num)
			}
		}
	}

	// Forgot to add a Value
	if assert.NoError(t, err) {
		_, errInsert := b.Insert(TABLE).Columns("name").Batch(1).Exec()
		assert.Error(t, errInsert)
	}

	// Column mismatch
	if assert.NoError(t, err) {
		_, errInsert := b.Insert(TABLE).Columns("id", "name", "doesNotExist").Batch(1).Values(values).Exec()
		assert.Error(t, errInsert)
	}
}

func TestSqlInsert_ExecTx(t *testing.T) {
	var values []map[string]interface{}
	values = append(values, map[string]interface{}{"id": 10, "name": "Combot"})
	values = append(values, map[string]interface{}{"id": 11, "name": "Uxptron"})

	b, err := HelperCreateBuilder()
	assert.NoError(t, err)
	assert.NoError(t, HelperDeleteEntries(b))

	// Normal Insert
	tx, err := b.NewTx()
	if assert.NoError(t, err) {
		if assert.NoError(t, err) {
			res, errInsert := b.Insert(TABLE).Columns("name").Values(values).ExecTx(tx)
			if assert.NoError(t, errInsert) {
				num, errRows := res[0].RowsAffected()
				if assert.NoError(t, errRows) {
					assert.Equal(t, int64(2), num)
				}
			}
		}

		// Batched Insert
		if assert.NoError(t, err) {
			res, errInsert := b.Insert(TABLE).Columns("name").Batch(1).Values(values).ExecTx(tx)
			if assert.NoError(t, errInsert) {
				num, errRows := res[0].RowsAffected()
				if assert.NoError(t, errRows) {
					assert.Equal(t, int64(1), num)
				}
				num, errRows = res[1].RowsAffected()
				if assert.NoError(t, errRows) {
					assert.Equal(t, int64(1), num)
				}
			}
		}

		// Forgot to add a Value - argument mismatch (rollback)
		if assert.NoError(t, err) {
			_, errInsert := b.Insert(TABLE).Columns("name").Batch(1).ExecTx(tx)
			assert.Error(t, errInsert)
		}

		// Column mismatch - argument mismatch (rollback)
		if assert.NoError(t, err) {
			_, errInsert := b.Insert(TABLE).Columns("id", "name", "doesNotExist").Batch(1).Values(values).ExecTx(tx)
			assert.Error(t, errInsert)
		}

		// Error because of rollback already happened
		err = b.CommitTx(tx)
		assert.Error(t, err)
	}
}

func TestSqlInsert_String(t *testing.T) {
	var values []map[string]interface{}
	values = append(values, map[string]interface{}{"id": 10, "name": "Combot"})
	values = append(values, map[string]interface{}{"id": 11, "name": "Uxptron"})

	b, err := HelperCreateBuilder()

	// Normal Insert
	if assert.NoError(t, err) {
		stmt, args, errInsert := b.Insert(TABLE).Columns("name").Values(values).String()
		if assert.NoError(t, errInsert) {
			if b.Placeholder.Numeric {
				assert.Equal(t, "INSERT INTO "+b.QuoteIdentifier(TABLE)+"("+b.QuoteIdentifier("name")+") VALUES ("+b.Placeholder.Char+"1), ("+b.Placeholder.Char+"2)", stmt)
			} else {
				assert.Equal(t, "INSERT INTO "+b.QuoteIdentifier(TABLE)+"("+b.QuoteIdentifier("name")+") VALUES ("+b.Placeholder.Char+"), ("+b.Placeholder.Char+")", stmt)
			}
			assert.Equal(t, [][]interface{}([][]interface{}{{"Combot", "Uxptron"}}), args)
		}
	}
}

package sqlquery_test

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSqlUpdate_String(t *testing.T) {
	values := map[string]interface{}{"id": 1, "name": "Wall-E"}
	b, err := HelperCreateBuilder()
	if assert.NoError(t, err) {
		stmt, args, errString := b.Update(TABLE).Columns("id", "name").Set(values).Where("id = ?", 2).String()
		if assert.NoError(t, errString) {
			if b.Placeholder.Numeric {
				assert.Equal(t, "UPDATE "+b.QuoteIdentifier(TABLE)+" SET "+b.QuoteIdentifier("id")+" = "+b.Placeholder.Char+"1, "+b.QuoteIdentifier("name")+" = "+b.Placeholder.Char+"2 WHERE id = "+b.Placeholder.Char+"3", stmt)
			} else {
				assert.Equal(t, "UPDATE "+b.QuoteIdentifier(TABLE)+" SET "+b.QuoteIdentifier("id")+" = "+b.Placeholder.Char+", "+b.QuoteIdentifier("name")+" = "+b.Placeholder.Char+" WHERE id = "+b.Placeholder.Char, stmt)
			}
			assert.Equal(t, []interface{}{1, "Wall-E", 2}, args)
		}

		//no value set
		_, _, errString = b.Update(TABLE).Columns("id", "name").Where("id = ?", 2).String()
		assert.Error(t, errString)

		//placeholder argument mismatch
		_, _, errString = b.Update(TABLE).Columns("id", "name").Set(values).Where("id = ? ?", 2).String()
		assert.Error(t, errString)
	}

}

func TestSqlUpdate_Exec(t *testing.T) {
	values := map[string]interface{}{"name": "Wall-E2"}
	b, err := HelperCreateBuilder()
	assert.NoError(t, HelperDeleteEntries(b))
	assert.NoError(t, HelperInsertEntries(b))

	if assert.NoError(t, err) {
		res, errExec := b.Update(TABLE).Columns("name").Set(values).Where("id = ?", 2).Exec()
		if assert.NoError(t, errExec) {
			num, errRows := res.RowsAffected()
			assert.NoError(t, errRows)
			assert.Equal(t, int64(1), num)

			s, errFirst := b.Select(TABLE).Where("id = ?", 2).Columns("name").First()
			if assert.NoError(t, errFirst) {
				r := Robot{}
				s.Scan(&r.Name)
				assert.Equal(t, true, r.Name.Valid)
				assert.Equal(t, "Wall-E2", r.Name.String)
			}
		}

		// error mismatch placeholder and arguments
		_, errExec = b.Update(TABLE).Columns("name").Set(values).Where("id = ? ?", 2).Exec()
		assert.Error(t, errExec)
	}
}

func TestSqlUpdate_ExecTx(t *testing.T) {
	values := map[string]interface{}{"name": "Wall-E2"}
	b, err := HelperCreateBuilder()
	assert.NoError(t, HelperDeleteEntries(b))
	assert.NoError(t, HelperInsertEntries(b))

	if assert.NoError(t, err) {
		tx, err := b.NewTx()
		if assert.NoError(t, err) {
			res, errExec := b.Update(TABLE).Columns("name").Set(values).Where("id = ?", 2).ExecTx(tx)
			if assert.NoError(t, errExec) {
				num, errRows := res.RowsAffected()
				assert.NoError(t, errRows)
				assert.Equal(t, int64(1), num)

				s, errFirst := b.Select(TABLE).Where("id = ?", 2).Columns("name").FirstTx(tx)
				if assert.NoError(t, errFirst) {
					r := Robot{}
					s.Scan(&r.Name)
					assert.Equal(t, true, r.Name.Valid)
					assert.Equal(t, "Wall-E2", r.Name.String)
				}
			}

			// error mismatch placeholder and arguments - rollback
			_, errExec = b.Update(TABLE).Columns("name").Set(values).Where("id = ? ?", 2).ExecTx(tx)
			assert.Error(t, errExec)

			// commit error - rollback already happened
			err = b.CommitTx(tx)
			assert.Error(t, err)
		}

	}
}

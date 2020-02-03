package sqlquery_test

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSqlDelete_WhereAndExec(t *testing.T) {
	b, err := HelperCreateBuilder()
	assert.NoError(t, HelperDeleteEntries(b))
	assert.NoError(t, HelperInsertEntries(b))

	if assert.NoError(t, err) {
		res, errExec := b.Delete(TABLE).Where("id = ?", 1).Exec()
		if assert.NoError(t, errExec) {

			num, errNum := res.RowsAffected()
			assert.NoError(t, errNum)
			assert.Equal(t, int64(1), num)

			rows, errSel := b.Select(TABLE).All()
			defer rows.Close()
			if assert.NoError(t, errSel) {

				var robots []Robot
				for rows.Next() {
					r := Robot{}
					errScan := rows.Scan(&r.ID, &r.Name)
					assert.NoError(t, errScan)
					robots = append(robots, r)
				}

				assert.Equal(t, 3, len(robots))
			}
		}
	}

	// delete without any where condition
	res, err := b.Delete(TABLE).Exec() //argument error
	if assert.NoError(t, err) {
		num, errNum := res.RowsAffected()
		assert.NoError(t, errNum)
		assert.Equal(t, int64(3), num)
	}

	// error because of mismatch placeholders and arguments
	_, err = b.Delete(TABLE).Where("id = ??", 1).Exec() //argument error
	assert.Error(t, err)

	// error because of none existing table
	_, err = b.Delete("notExisting").Exec()
	assert.Error(t, err)
}

func TestSqlDelete_WhereAndExecTx(t *testing.T) {
	b, err := HelperCreateBuilder()
	if assert.NoError(t, err) {

		assert.NoError(t, HelperDeleteEntries(b))
		assert.NoError(t, HelperInsertEntries(b))

		tx, err := b.NewTx()
		if assert.NoError(t, err) {
			res, errExec := b.Delete(TABLE).Where("id = ?", 1).ExecTx(tx)
			if assert.NoError(t, errExec) {

				num, errNum := res.RowsAffected()
				assert.NoError(t, errNum)
				assert.Equal(t, int64(1), num)

				rows, errSel := b.Select(TABLE).AllTx(tx)
				defer rows.Close()
				if assert.NoError(t, errSel) {

					var robots []Robot
					for rows.Next() {
						r := Robot{}
						errScan := rows.Scan(&r.ID, &r.Name)
						assert.NoError(t, errScan)
						robots = append(robots, r)
					}

					assert.Equal(t, 3, len(robots))
				}
			}
		}

		// delete without any where condition
		res, err := b.Delete(TABLE).ExecTx(tx) //argument error
		if assert.NoError(t, err) {
			num, errNum := res.RowsAffected()
			assert.NoError(t, errNum)
			assert.Equal(t, int64(3), num)
		}

		// error because of mismatch placeholders and arguments - rollback
		_, err = b.Delete(TABLE).Where("id = ??", 1).ExecTx(tx) //argument error
		assert.Error(t, err)

		// error because of none existing table - rollback
		_, err = b.Delete("notExisting").ExecTx(tx)
		assert.Error(t, err)

		// Rollback already happened
		err = b.CommitTx(tx)
		assert.Error(t, err)
	}

}

func TestSqlDelete_String(t *testing.T) {
	b, err := HelperCreateBuilder()
	if assert.NoError(t, err) {
		stmt, args, errString := b.Delete(TABLE).Where("id = ?", 1).String()
		assert.NoError(t, errString)
		if b.Placeholder.Numeric {
			assert.Equal(t, "DELETE FROM "+b.QuoteIdentifier(TABLE)+" WHERE id = "+b.Placeholder.Char+"1", stmt)
		} else {
			assert.Equal(t, "DELETE FROM "+b.QuoteIdentifier(TABLE)+" WHERE id = ?", stmt)
		}
		assert.Equal(t, []interface{}([]interface{}{1}), args)

		stmt, args, errString = b.Delete(TABLE).String()
		assert.NoError(t, errString)
		assert.Equal(t, "DELETE FROM "+b.QuoteIdentifier(TABLE), stmt)
		assert.Equal(t, []interface{}(nil), args)
	}
}

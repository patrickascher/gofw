package sqlquery_test

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSqlSelect_First(t *testing.T) {

	b, err := HelperCreateBuilder()
	assert.NoError(t, HelperDeleteEntries(b))
	assert.NoError(t, HelperInsertEntries(b))

	if assert.NoError(t, err) {
		//normal select
		row, errSel := b.Select(TABLE).Columns("id", "name").Where("id = ?", 1).First()
		if assert.NoError(t, errSel) {
			r := Robot{}
			errScan := row.Scan(&r.ID, &r.Name)
			if assert.NoError(t, errScan) {
				assert.Equal(t, 1, r.ID)
				assert.Equal(t, true, r.Name.Valid)
				assert.Equal(t, "Cozmo", r.Name.String)
			}
		}

		// mismatch placeholder arguments
		_, errSel = b.Select(TABLE).Columns("id", "name").Where("id = ? ?", 1).First()
		assert.Error(t, errSel)
	}
}

func TestSqlSelect_FirstTx(t *testing.T) {

	b, err := HelperCreateBuilder()
	assert.NoError(t, HelperDeleteEntries(b))
	assert.NoError(t, HelperInsertEntries(b))

	if assert.NoError(t, err) {
		tx, err := b.NewTx()

		if assert.NoError(t, err) {
			//normal select
			row, errSel := b.Select(TABLE).Columns("id", "name").Where("id = ?", 1).FirstTx(tx)
			if assert.NoError(t, errSel) {
				r := Robot{}
				errScan := row.Scan(&r.ID, &r.Name)
				if assert.NoError(t, errScan) {
					assert.Equal(t, 1, r.ID)
					assert.Equal(t, true, r.Name.Valid)
					assert.Equal(t, "Cozmo", r.Name.String)
				}
			}

			// mismatch placeholder arguments - argument mismatch (rollback)
			_, errSel = b.Select(TABLE).Columns("id", "name").Where("id = ? ?", 1).FirstTx(tx)
			assert.Error(t, errSel)

			// commit error because rollback already happened
			err = b.CommitTx(tx)
			assert.Error(t, err)
		}
	}
}

func TestSqlSelect_All(t *testing.T) {

	b, err := HelperCreateBuilder()
	assert.NoError(t, HelperDeleteEntries(b))
	assert.NoError(t, HelperInsertEntries(b))

	if assert.NoError(t, err) {
		//normal select
		rows, errSel := b.Select(TABLE).Columns("id", "name").All()
		defer rows.Close()
		if assert.NoError(t, errSel) {
			var robots []Robot

			for rows.Next() {
				r := Robot{}
				errScan := rows.Scan(&r.ID, &r.Name)
				assert.NoError(t, errScan)
				robots = append(robots, r)
			}

			if assert.Equal(t, 4, len(robots)) {
				assert.Equal(t, 1, robots[0].ID)
				assert.Equal(t, "Cozmo", robots[0].Name.String)
				assert.Equal(t, true, robots[0].Name.Valid)

				assert.Equal(t, 2, robots[1].ID)
				assert.Equal(t, "Wall-E", robots[1].Name.String)
				assert.Equal(t, true, robots[1].Name.Valid)

				assert.Equal(t, 3, robots[2].ID)
				assert.Equal(t, "Spark", robots[2].Name.String)
				assert.Equal(t, true, robots[2].Name.Valid)

				assert.Equal(t, 4, robots[3].ID)
				assert.Equal(t, "Ubimator", robots[3].Name.String)
				assert.Equal(t, true, robots[3].Name.Valid)
			}
		}

		// mismatch placeholder arguments
		_, errSel = b.Select(TABLE).Where("id = ? ?", 1).All()
		assert.Error(t, errSel)
	}
}

func TestSqlSelect_AllTx(t *testing.T) {

	b, err := HelperCreateBuilder()
	assert.NoError(t, HelperDeleteEntries(b))
	assert.NoError(t, HelperInsertEntries(b))

	if assert.NoError(t, err) {
		tx, err := b.NewTx()
		if assert.NoError(t, err) {
			//normal select
			rows, errSel := b.Select(TABLE).Columns("id", "name").AllTx(tx)
			defer rows.Close()
			if assert.NoError(t, errSel) {
				var robots []Robot

				for rows.Next() {
					r := Robot{}
					errScan := rows.Scan(&r.ID, &r.Name)
					assert.NoError(t, errScan)
					robots = append(robots, r)
				}

				if assert.Equal(t, 4, len(robots)) {
					assert.Equal(t, 1, robots[0].ID)
					assert.Equal(t, "Cozmo", robots[0].Name.String)
					assert.Equal(t, true, robots[0].Name.Valid)

					assert.Equal(t, 2, robots[1].ID)
					assert.Equal(t, "Wall-E", robots[1].Name.String)
					assert.Equal(t, true, robots[1].Name.Valid)

					assert.Equal(t, 3, robots[2].ID)
					assert.Equal(t, "Spark", robots[2].Name.String)
					assert.Equal(t, true, robots[2].Name.Valid)

					assert.Equal(t, 4, robots[3].ID)
					assert.Equal(t, "Ubimator", robots[3].Name.String)
					assert.Equal(t, true, robots[3].Name.Valid)
				}
			}

			// mismatch placeholder arguments - rollback
			_, errSel = b.Select(TABLE).Where("id = ? ?", 1).AllTx(tx)
			assert.Error(t, errSel)

			// commit error because rollback already happened
			err = b.CommitTx(tx)
			assert.Error(t, err)
		}

	}
}

func TestSqlSelect_String(t *testing.T) {
	b, err := HelperCreateBuilder()
	if assert.NoError(t, err) {
		stmt, args, errSel := b.Select(TABLE).Where("id = ?", 1).String()
		if assert.NoError(t, errSel) {
			if b.Placeholder.Numeric {
				assert.Equal(t, "SELECT * FROM "+b.QuoteIdentifier(TABLE)+" WHERE id = "+b.Placeholder.Char+"1", stmt)
			} else {
				assert.Equal(t, "SELECT * FROM "+b.QuoteIdentifier(TABLE)+" WHERE id = "+b.Placeholder.Char, stmt)
			}
			assert.Equal(t, []interface{}([]interface{}{1}), args)
		}
	}
}

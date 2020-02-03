package sqlquery

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuilder_first(t *testing.T) {
	b, err := HelperCreateBuilder()
	if assert.NoError(t, err) {
		assert.NoError(t, HelperDeleteEntries(b))
		assert.NoError(t, HelperInsertEntries(b))

		stmt, args, errStmt := b.Select(TABLE).Where(b.QuoteIdentifier("id")+" = ?", 1).String()
		if assert.NoError(t, errStmt) {
			row, errFirst := b.first(stmt, args)
			assert.NoError(t, errFirst)
			assert.IsType(t, &sql.Row{}, row)

			r := Robot{}
			row.Scan(&r.ID, &r.Name)

			assert.Equal(t, 1, r.ID)
			assert.Equal(t, "Cozmo", r.Name.String)
			assert.Equal(t, true, r.Name.Valid)
		}
	}
}

// TestBuilder_first_Tx is a copy of TestBuilder_first with transaction activated
func TestBuilder_first_Tx(t *testing.T) {

	b, err := HelperCreateBuilder()
	assert.NoError(t, HelperDeleteEntries(b))
	assert.NoError(t, HelperInsertEntries(b))

	if assert.NoError(t, err) {
		tx, err := b.NewTx()
		if assert.NoError(t, err) {
			stmt, args, errStmt := b.Select(TABLE).Where(b.QuoteIdentifier("id")+" = ?", 1).String()
			if assert.NoError(t, errStmt) {
				row, errFirst := b.firstTx(tx, stmt, args)
				assert.NoError(t, errFirst)
				assert.IsType(t, &sql.Row{}, row)

				r := Robot{}
				row.Scan(&r.ID, &r.Name)

				assert.Equal(t, 1, r.ID)
				assert.Equal(t, "Cozmo", r.Name.String)
				assert.Equal(t, true, r.Name.Valid)
			}

			assert.NoError(t, b.CommitTx(tx))
		}

	}
}

func TestBuilder_all(t *testing.T) {
	b, err := HelperCreateBuilder()
	assert.NoError(t, HelperDeleteEntries(b))
	assert.NoError(t, HelperInsertEntries(b))

	var robots []Robot

	if assert.NoError(t, err) {
		stmt, args, errStmt := b.Select(TABLE).Columns("id", "name").Where(b.QuoteIdentifier("id")+" IN (?)", []int{1, 2, 3, 4}).String()
		if assert.NoError(t, errStmt) {
			rows, errAll := b.all(stmt, args)
			defer rows.Close()

			assert.NoError(t, errAll)
			assert.IsType(t, &sql.Rows{}, rows)

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
	}
}

// TestBuilder_all_Tx_Err is checking if all is responding correctly with an transaction error
func TestBuilder_all_Tx_Err(t *testing.T) {
	b, err := HelperCreateBuilder()
	assert.NoError(t, err)
	assert.NoError(t, HelperDeleteEntries(b))
	assert.NoError(t, HelperInsertEntries(b))

	tx, err := b.NewTx()
	if assert.NoError(t, err) {
		stmt, args, errStmt := b.Select(TABLE).Columns("notExistingColumn").Where(b.QuoteIdentifier("id")+" IN (?)", []int{1, 2, 3, 4}).String()
		assert.NoError(t, errStmt) //Arguments are ok

		rows, errAll := b.allTx(tx, stmt, args)
		if !assert.Error(t, errAll) {
			defer rows.Close()
		}

		//Commit - Rollback was already called
		assert.Error(t, b.CommitTx(tx))
	}
}

// TestBuilder_all_Tx is a copy of TestBuilder_all with transaction activated
func TestBuilder_all_Tx(t *testing.T) {
	b, err := HelperCreateBuilder()
	assert.NoError(t, err)
	assert.NoError(t, HelperDeleteEntries(b))
	assert.NoError(t, HelperInsertEntries(b))

	var robots []Robot
	tx, err := b.NewTx()
	if assert.NoError(t, err) {
		if assert.NoError(t, err) {
			stmt, args, errStmt := b.Select(TABLE).Columns("id", "name").Where(b.QuoteIdentifier("id")+" IN (?)", []int{1, 2, 3, 4}).String()
			if assert.NoError(t, errStmt) {
				rows, errAll := b.allTx(tx, stmt, args)
				defer rows.Close()

				assert.NoError(t, errAll)
				assert.IsType(t, &sql.Rows{}, rows)

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
		}

		val := map[string]interface{}{"Names": "GoBot"}
		stmt, args, errUpdate := b.Update(TABLE).Set(val).Where(b.QuoteIdentifier("id")+" = ?", 1).String()
		assert.NoError(t, errUpdate)

		// Unknown field Name -> Rollback
		var argsSlice [][]interface{}
		argsSlice = append(argsSlice, args)
		res, errRes := b.execTx(tx, stmt, argsSlice)
		assert.Error(t, errRes)
		assert.Empty(t, res)

		// TX is already rolled back... so no new query allowed
		val = map[string]interface{}{"Name": "GoBot"}
		stmt, args, errUpdate = b.Update(TABLE).Set(val).Where(b.QuoteIdentifier("id")+" = ?", 1).String()
		assert.NoError(t, errUpdate)
		argsSlice = append(argsSlice, args)
		res, errRes = b.execTx(tx, stmt, argsSlice)
		assert.Error(t, errRes)
		assert.Empty(t, res)

		//Commit, but Rollback already happened
		assert.Error(t, b.CommitTx(tx))

		//check robot with the id 1 "Cozmo" must still have his name
		stmt, args, errStmt := b.Select(TABLE).Where(b.QuoteIdentifier("id")+" = ?", 1).String()
		if assert.NoError(t, errStmt) {
			row, errFirst := b.first(stmt, args)
			assert.NoError(t, errFirst)
			assert.IsType(t, &sql.Row{}, row)

			r := Robot{}
			errScan := row.Scan(&r.ID, &r.Name)
			assert.NoError(t, errScan)

			assert.Equal(t, 1, r.ID)
			assert.Equal(t, "Cozmo", r.Name.String)
			assert.Equal(t, true, r.Name.Valid)

		}
	}
}

func TestBuilder_exec(t *testing.T) {
	b, err := HelperCreateBuilder()
	if assert.NoError(t, err) {

		stmt, args, errStmt := b.Delete(TABLE).Where(b.QuoteIdentifier("id")+" = ?", 1).String()
		if assert.NoError(t, errStmt) {
			rows, errExec := b.exec(stmt, [][]interface{}{args})
			assert.NoError(t, errExec)
			assert.IsType(t, []sql.Result{}, rows)
		}
	}
}

func TestBuilder_execTx(t *testing.T) {
	b, err := HelperCreateBuilder()
	if assert.NoError(t, err) {
		tx, err := b.NewTx()
		if assert.NoError(t, err) {
			stmt, args, errStmt := b.Delete(TABLE).Where(b.QuoteIdentifier("id")+" = ?", 1).String()
			if assert.NoError(t, errStmt) {
				rows, errExec := b.execTx(tx, stmt, [][]interface{}{args})
				assert.NoError(t, errExec)
				assert.IsType(t, []sql.Result{}, rows)
			}
			stmt, args, errStmt = b.Delete(TABLE).Where(b.QuoteIdentifier("id")+" = ?", 2).String()
			if assert.NoError(t, errStmt) {
				rows, errExec := b.execTx(tx, stmt, [][]interface{}{args})
				assert.NoError(t, errExec)
				assert.IsType(t, []sql.Result{}, rows)
			}
			err = b.CommitTx(tx)
			assert.NoError(t, err)
		}

	}
}

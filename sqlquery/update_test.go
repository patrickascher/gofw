// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlquery_test

import (
	"fmt"
	"testing"

	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
)

// TestDelete is table driven and checks the render function for different delete stmts.
func TestUpdate(t *testing.T) {
	test := assert.New(t)

	b, err := sqlquery.New(sqlquery.Config{Driver: "test"}, nil)
	test.NoError(err)

	c1 := &sqlquery.Condition{}

	// table driven:
	var tests = []struct {
		expectedArgs []interface{}
		expectedSql  string
		from         string
		where        string
		value        map[string]interface{}
		columns      []string

		condition *sqlquery.Condition

		error    bool
		errorMsg string
	}{
		// err: update without any values
		{expectedSql: "", from: "users", error: true, errorMsg: sqlquery.ErrValueMissing.Error()},
		// ok - testing value and order of it
		{expectedArgs: []interface{}{"John", "Doe"}, expectedSql: "UPDATE 'users' SET 'name' = ?, 'surname' = ?", columns: []string{"name", "surname"}, from: "users", value: map[string]interface{}{"surname": "Doe", "name": "John"}},
		// err: value set name and surname but only name is allowed
		{expectedArgs: []interface{}{"John", "Doe"}, expectedSql: "UPDATE 'users' SET 'surname' = ?, 'name' = ?", from: "users", columns: []string{"nameX"}, value: map[string]interface{}{"name": "John", "surname": "Doe"}, error: true, errorMsg: fmt.Sprintf(sqlquery.ErrColumn.Error(), "nameX")},
		{expectedArgs: []interface{}{"John", "Doe", int64(1), int64(2), int64(3), int64(4), int64(5)}, expectedSql: "UPDATE 'users' SET 'name' = ?, 'surname' = ? WHERE id IN (?, ?, ?, ?, ?)", from: "users", value: map[string]interface{}{"surname": "Doe", "name": "John"}, columns: []string{"name", "surname"}, condition: c1.Where("id IN (?)", []int{1, 2, 3, 4, 5})},
		// argument mismatch
		{expectedSql: "DELETE FROM 'users' WHERE id = 1", from: "users", value: map[string]interface{}{"name": "John", "surname": "Doe"}, where: "id = " + sqlquery.PLACEHOLDER + " OR id = " + sqlquery.PLACEHOLDER, error: true, errorMsg: fmt.Sprintf(sqlquery.ErrPlaceholderMismatch.Error(), "id = ? OR id = ?", 2, 0)},
	}

	for _, tt := range tests {
		t.Run(tt.expectedSql, func(t *testing.T) {

			sel := b.Update(tt.from).Set(tt.value)
			if tt.columns != nil {
				sel.Columns(tt.columns...)
				// multi call to check if they are added correctly
				sel.Columns(tt.columns...)
			}
			if tt.where != "" {
				sel.Where(tt.where)
			}
			if tt.condition != nil {
				sel.Condition(tt.condition)
			}

			sql, args, err := sel.String()
			if tt.error {
				test.Error(err)
				test.Equal("", sql)
				test.Nil(args)
				test.Equal(tt.errorMsg, err.Error())
			} else {
				test.NoError(err, sql)
				test.Equal(tt.expectedSql, sql)
				test.Equal(tt.expectedArgs, args)

			}
		})
	}
}

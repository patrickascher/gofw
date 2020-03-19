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
func TestDelete(t *testing.T) {
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

		condition *sqlquery.Condition

		error    bool
		errorMsg string
	}{
		{expectedSql: "DELETE FROM 'users'", from: "users"},
		{expectedArgs: []interface{}{1}, expectedSql: "DELETE FROM 'users' WHERE id = ?", from: "users", where: "id = " + sqlquery.PLACEHOLDER},
		{expectedArgs: []interface{}{int64(1), int64(2), int64(3), int64(4), int64(5)}, expectedSql: "DELETE FROM 'users' WHERE id IN (?, ?, ?, ?, ?)", from: "users", condition: c1.Where("id IN (?)", []int{1, 2, 3, 4, 5})},
		// argument mismatch
		{expectedSql: "DELETE FROM 'users' WHERE id = 1", from: "users", where: "id = " + sqlquery.PLACEHOLDER + " OR id = " + sqlquery.PLACEHOLDER, error: true, errorMsg: fmt.Sprintf(sqlquery.ErrPlaceholderMismatch.Error(), "id = ? OR id = ?", 2, 1)},
	}

	for _, tt := range tests {
		t.Run(tt.expectedSql, func(t *testing.T) {

			sel := b.Delete(tt.from).Where(tt.where, 1)
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

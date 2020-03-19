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

// TestInsert is table driven and checks the render function for different delete stmts.
func TestInsert(t *testing.T) {
	test := assert.New(t)

	b, err := sqlquery.New(sqlquery.Config{Driver: "test"}, nil)
	test.NoError(err)

	// table driven:
	var tests = []struct {
		expectedArgs [][]interface{}
		expectedSql  string
		into         string
		value        []map[string]interface{}
		columns      []string
		batch        int
		lastID       string

		error    bool
		errorMsg string
	}{
		// err: update without any values
		{expectedSql: "", into: "users", error: true, errorMsg: sqlquery.ErrValueMissing.Error()},
		// ok - testing value and order of it
		{expectedArgs: [][]interface{}{{"John", "Doe"}}, expectedSql: "INSERT INTO 'users'('name', 'surname') VALUES (?, ?)", columns: []string{"name", "surname"}, into: "users", value: []map[string]interface{}{{"surname": "Doe", "name": "John"}}},
		// err: value set name and surname but only name is allowed
		{expectedArgs: [][]interface{}{{"John", "Doe"}}, expectedSql: "UPDATE 'users' SET 'surname' = ?, 'name' = ?", into: "users", columns: []string{"nameX"}, value: []map[string]interface{}{{"name": "John", "surname": "Doe"}}, error: true, errorMsg: fmt.Sprintf(sqlquery.ErrColumn.Error(), "nameX")},
		// ok: multiple value
		{expectedArgs: [][]interface{}{{"John", "Doe", "Bar", "Foo"}}, expectedSql: "INSERT INTO 'users'('name', 'surname') VALUES (?, ?), (?, ?)", columns: []string{"name", "surname"}, into: "users", value: []map[string]interface{}{{"surname": "Doe", "name": "John"}, {"surname": "Foo", "name": "Bar"}}},
		// ok: batch value
		{batch: 1, expectedArgs: [][]interface{}{{"John", "Doe"}, {"Bar", "Foo"}}, expectedSql: "INSERT INTO 'users'('name', 'surname') VALUES (?, ?)", columns: []string{"name", "surname"}, into: "users", value: []map[string]interface{}{{"surname": "Doe", "name": "John"}, {"surname": "Foo", "name": "Bar"}}},
		// ok: batch default size 50 is used
		{lastID: "ID", batch: 0, expectedArgs: [][]interface{}{{"John", "Doe", "Bar", "Foo"}}, expectedSql: "INSERT INTO 'batchcheck'('name', 'surname') VALUES (?, ?), (?, ?)", columns: []string{"name", "surname"}, into: "batchcheck", value: []map[string]interface{}{{"surname": "Doe", "name": "John"}, {"surname": "Foo", "name": "Bar"}}},
	}

	for _, tt := range tests {
		t.Run(tt.expectedSql, func(t *testing.T) {

			sel := b.Insert(tt.into).Values(tt.value)
			if tt.columns != nil {
				sel.Columns(tt.columns...)
				// multiple call - to test if they are added correctly
				sel.Columns(tt.columns...)
			}
			if tt.batch != 0 || tt.into == "batchcheck" {
				sel.Batch(tt.batch)
			}
			if tt.lastID != "" {
				id := 0
				sel.LastInsertedID(tt.lastID, &id)
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

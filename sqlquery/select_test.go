// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlquery_test

import (
	"fmt"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestSelect is table driven and checks the render function for different selects.
func TestSelect(t *testing.T) {
	test := assert.New(t)

	b, err := sqlquery.New(sqlquery.Config{Driver: "test"}, nil)
	test.NoError(err)

	jc1 := &sqlquery.Condition{}
	jc2 := &sqlquery.Condition{}
	jc3 := &sqlquery.Condition{}
	jc4 := &sqlquery.Condition{}
	c := &sqlquery.Condition{}

	// table driven:
	var tests = []struct {
		expectedSql  string
		expectedArgs []interface{}
		from         string
		columns      []string
		group        []string
		order        []string
		limit        int
		offset       int

		having string

		where string

		joinType      int
		JoinTable     string
		JoinCondition *sqlquery.Condition

		condition *sqlquery.Condition

		error    bool
		errorMsg string
	}{
		{expectedSql: "SELECT * FROM 'users'", from: "users"},
		{expectedSql: "SELECT 'name', 'surname' FROM 'users'", from: "users", columns: []string{"name", "surname"}},
		{expectedSql: "SELECT * FROM 'users' ORDER BY name ASC, surname ASC", from: "users", order: []string{"name", "surname"}},
		{expectedSql: "SELECT * FROM 'users' LIMIT 1", from: "users", limit: 1},
		{expectedSql: "SELECT * FROM 'users' OFFSET 5", from: "users", offset: 5},
		{expectedArgs: []interface{}{1}, expectedSql: "SELECT * FROM 'users' WHERE id = ?", from: "users", where: "id = " + sqlquery.PLACEHOLDER},
		{expectedArgs: []interface{}{2}, expectedSql: "SELECT * FROM 'users' HAVING id = ?", from: "users", having: "id = " + sqlquery.PLACEHOLDER},
		{expectedArgs: []interface{}{1}, expectedSql: "SELECT * FROM 'users' LEFT JOIN 'departments' ON users.dep = departments.id AND users.id = ?", from: "users", joinType: sqlquery.LEFT, JoinTable: "departments", JoinCondition: jc1.On("users.dep = departments.id AND users.id = "+sqlquery.PLACEHOLDER, 1)},
		{expectedArgs: []interface{}{1}, expectedSql: "SELECT * FROM 'users' RIGHT JOIN 'departments' ON users.dep = departments.id AND users.id = ?", from: "users", joinType: sqlquery.RIGHT, JoinTable: "departments", JoinCondition: jc2.On("users.dep = departments.id AND users.id = "+sqlquery.PLACEHOLDER, 1)},
		{expectedArgs: []interface{}{1}, expectedSql: "SELECT * FROM 'users' INNER JOIN 'departments' ON users.dep = departments.id AND users.id = ?", from: "users", joinType: sqlquery.INNER, JoinTable: "departments", JoinCondition: jc3.On("users.dep = departments.id AND users.id = "+sqlquery.PLACEHOLDER, 1)},
		{expectedArgs: []interface{}{1, 1, 2}, expectedSql: "SELECT 'name', 'surname' FROM 'users' LEFT JOIN 'departments' ON users.dep = departments.id AND users.id = ? WHERE id = ? HAVING id = ? ORDER BY id,created ASC LIMIT 1 OFFSET 5", from: "users", where: "id = " + sqlquery.PLACEHOLDER, having: "id = " + sqlquery.PLACEHOLDER, columns: []string{"name", "surname"}, order: []string{"id,created"}, limit: 1, offset: 5, joinType: sqlquery.LEFT, JoinTable: "departments", JoinCondition: jc1.On("users.dep = departments.id AND users.id = "+sqlquery.PLACEHOLDER, 1)},
		// err: wrong join type
		{expectedSql: "SELECT * FROM 'users' INNER JOIN 'departments' ON users.dep = departments.id AND users.id = ?", from: "users", joinType: 5, JoinTable: "departments", JoinCondition: jc1.On("users.dep = departments.id AND users.id = "+sqlquery.PLACEHOLDER, 1), error: true, errorMsg: fmt.Sprintf(sqlquery.ErrJoinType.Error(), 5)},
		// err: missing argument ON
		{expectedSql: "SELECT * FROM 'users' LEFT JOIN 'departments' ON users.dep = departments.id AND users.id = ?", from: "users", joinType: sqlquery.LEFT, JoinTable: "departments", JoinCondition: jc4.On("users.dep = departments.id AND users.id = " + sqlquery.PLACEHOLDER), error: true, errorMsg: fmt.Sprintf(sqlquery.ErrPlaceholderMismatch.Error(), "users.dep = departments.id AND users.id = ?", 1, 0)},
		{expectedSql: "SELECT 'id', 'name' FROM 'users' GROUP BY company ORDER BY id ASC, name ASC LIMIT 10", from: "users", columns: []string{"id", "name"}, condition: c.Order("id", "name").Limit(10).Group("company")},
	}

	for _, tt := range tests {
		t.Run(tt.expectedSql, func(t *testing.T) {

			sel := b.Select(tt.from).Columns(tt.columns...).Columns(tt.columns...).Group(tt.group...).Order(tt.order...).Having(tt.having, 2).Where(tt.where, 1).Join(tt.joinType, tt.JoinTable, tt.JoinCondition)
			if tt.offset != 0 {
				sel.Offset(tt.offset)
			}
			if tt.limit != 0 {
				sel.Limit(tt.limit)
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

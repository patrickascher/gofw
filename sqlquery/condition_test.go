// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlquery_test

import (
	"strings"
	"testing"

	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
)

func TestCondition_ConfigAndReset(t *testing.T) {
	test := assert.New(t)

	c := sqlquery.Condition{}
	c.Limit(1)
	c.Offset(10)
	c.Order("created")
	c.Where("id = "+sqlquery.PLACEHOLDER, 1)
	c.Having("id = "+sqlquery.PLACEHOLDER, 2)
	c.Group("id", "name")
	c.On("user.id = company.id AND user.id != "+sqlquery.PLACEHOLDER, 5)

	test.Equal("WHERE id = 1 GROUP BY id, name HAVING id = 2 ORDER BY created ASC LIMIT 1 OFFSET 10", c.Config(true, sqlquery.WHERE, sqlquery.GROUP, sqlquery.HAVING, sqlquery.ORDER, sqlquery.LIMIT, sqlquery.OFFSET))
	test.Equal("WHERE id = "+sqlquery.PLACEHOLDER+" GROUP BY id, name HAVING id = "+sqlquery.PLACEHOLDER+" ORDER BY created ASC LIMIT 1 OFFSET 10", c.Config(false, sqlquery.WHERE, sqlquery.GROUP, sqlquery.HAVING, sqlquery.ORDER, sqlquery.LIMIT, sqlquery.OFFSET))

	c.Reset(sqlquery.WHERE)
	test.Equal("GROUP BY id, name HAVING id = 2 ORDER BY created ASC LIMIT 1 OFFSET 10", c.Config(true, sqlquery.WHERE, sqlquery.GROUP, sqlquery.HAVING, sqlquery.ORDER, sqlquery.LIMIT, sqlquery.OFFSET))
	test.Equal("GROUP BY id, name HAVING id = "+sqlquery.PLACEHOLDER+" ORDER BY created ASC LIMIT 1 OFFSET 10", c.Config(false, sqlquery.WHERE, sqlquery.GROUP, sqlquery.HAVING, sqlquery.ORDER, sqlquery.LIMIT, sqlquery.OFFSET))

	c.Reset(sqlquery.GROUP)
	test.Equal("HAVING id = 2 ORDER BY created ASC LIMIT 1 OFFSET 10", c.Config(true, sqlquery.WHERE, sqlquery.GROUP, sqlquery.HAVING, sqlquery.ORDER, sqlquery.LIMIT, sqlquery.OFFSET))
	test.Equal("HAVING id = "+sqlquery.PLACEHOLDER+" ORDER BY created ASC LIMIT 1 OFFSET 10", c.Config(false, sqlquery.WHERE, sqlquery.GROUP, sqlquery.HAVING, sqlquery.ORDER, sqlquery.LIMIT, sqlquery.OFFSET))

	c.Reset(sqlquery.HAVING)
	test.Equal("ORDER BY created ASC LIMIT 1 OFFSET 10", c.Config(true, sqlquery.WHERE, sqlquery.GROUP, sqlquery.HAVING, sqlquery.ORDER, sqlquery.LIMIT, sqlquery.OFFSET))
	test.Equal("ORDER BY created ASC LIMIT 1 OFFSET 10", c.Config(false, sqlquery.WHERE, sqlquery.GROUP, sqlquery.HAVING, sqlquery.ORDER, sqlquery.LIMIT, sqlquery.OFFSET))

	c.Reset(sqlquery.ORDER)
	test.Equal("LIMIT 1 OFFSET 10", c.Config(true, sqlquery.WHERE, sqlquery.GROUP, sqlquery.HAVING, sqlquery.ORDER, sqlquery.LIMIT, sqlquery.OFFSET))
	test.Equal("LIMIT 1 OFFSET 10", c.Config(false, sqlquery.WHERE, sqlquery.GROUP, sqlquery.HAVING, sqlquery.ORDER, sqlquery.LIMIT, sqlquery.OFFSET))

	c.Reset(sqlquery.LIMIT)
	test.Equal("OFFSET 10", c.Config(true, sqlquery.WHERE, sqlquery.GROUP, sqlquery.HAVING, sqlquery.ORDER, sqlquery.LIMIT, sqlquery.OFFSET))
	test.Equal("OFFSET 10", c.Config(false, sqlquery.WHERE, sqlquery.GROUP, sqlquery.HAVING, sqlquery.ORDER, sqlquery.LIMIT, sqlquery.OFFSET))

	c.Reset(sqlquery.OFFSET)
	test.Equal("", c.Config(true, sqlquery.WHERE, sqlquery.GROUP, sqlquery.HAVING, sqlquery.ORDER, sqlquery.LIMIT, sqlquery.OFFSET))
	test.Equal("", c.Config(false, sqlquery.WHERE, sqlquery.GROUP, sqlquery.HAVING, sqlquery.ORDER, sqlquery.LIMIT, sqlquery.OFFSET))

	test.Equal("ON user.id = company.id AND user.id != 5", c.Config(true, sqlquery.ON))
	test.Equal("ON user.id = company.id AND user.id != "+sqlquery.PLACEHOLDER, c.Config(false, sqlquery.ON))
	c.Reset(sqlquery.ON)
	test.Equal("", c.Config(true, sqlquery.ON))
	test.Equal("", c.Config(false, sqlquery.ON))

	//reset all
	c = sqlquery.Condition{}
	c.Limit(1)
	c.Offset(10)
	c.Order("created")
	c.Where("id = "+sqlquery.PLACEHOLDER, 1)
	c.Having("id = "+sqlquery.PLACEHOLDER, 2)
	c.Group("id", "name")
	c.On("user.id = company.id AND user.id != "+sqlquery.PLACEHOLDER, 5)
	c.Reset()
	test.Equal("", c.Config(true, sqlquery.WHERE, sqlquery.GROUP, sqlquery.HAVING, sqlquery.ORDER, sqlquery.LIMIT, sqlquery.OFFSET))

}

func TestCondition_Where(t *testing.T) {
	test := assert.New(t)
	c := sqlquery.Condition{}
	c.Where("id = "+sqlquery.PLACEHOLDER+" AND name="+sqlquery.PLACEHOLDER, 1, "John Doe")
	test.Equal("WHERE id = 1 AND name=John Doe", c.Config(true, sqlquery.WHERE))

	// ok: calling WHERE again - conditions are getting chained.
	c.Where("surname="+sqlquery.PLACEHOLDER, "Bar")
	test.Equal("WHERE id = 1 AND name=John Doe AND surname=Bar", c.Config(true, sqlquery.WHERE))

	// error: because of argument mismatch. internally c.error is set.
	// string is not getting added.
	c.Where("additional="+sqlquery.PLACEHOLDER, "Foo", 1)
	test.Equal("WHERE id = 1 AND name=John Doe AND surname=Bar", c.Config(true, sqlquery.WHERE))
}

func TestCondition_Group(t *testing.T) {
	test := assert.New(t)
	c := sqlquery.Condition{}

	c.Group()
	test.Equal("", c.Config(true, sqlquery.GROUP))

	c.Group("")
	test.Equal("", c.Config(true, sqlquery.GROUP))

	c.Group("branch")
	test.Equal("GROUP BY branch", c.Config(true, sqlquery.GROUP))
}

func TestCondition_Having(t *testing.T) {
	test := assert.New(t)
	c := sqlquery.Condition{}
	c.Having("id = "+sqlquery.PLACEHOLDER+" AND name="+sqlquery.PLACEHOLDER, 1, "John Doe")
	test.Equal("HAVING id = 1 AND name=John Doe", c.Config(true, sqlquery.HAVING))

	// ok: calling HAVING again - conditions are getting chained.
	c.Having("surname="+sqlquery.PLACEHOLDER, "Bar")
	test.Equal("HAVING id = 1 AND name=John Doe AND surname=Bar", c.Config(true, sqlquery.HAVING))

	// error: because of argument mismatch. internally c.error is set.
	// string is not getting added.
	c.Having("additional="+sqlquery.PLACEHOLDER, "Foo", 1)
	test.Equal("HAVING id = 1 AND name=John Doe AND surname=Bar", c.Config(true, sqlquery.HAVING))
}

func TestCondition_Order(t *testing.T) {
	test := assert.New(t)
	c := sqlquery.Condition{}

	c.Order()
	test.Equal("", c.Config(true, sqlquery.ORDER))

	c.Order("")
	test.Equal("", c.Config(true, sqlquery.ORDER))

	c.Order("branch")
	test.Equal("ORDER BY branch ASC", c.Config(true, sqlquery.ORDER))

	c.Order("company", "branch")
	test.Equal("ORDER BY company ASC, branch ASC", c.Config(true, sqlquery.ORDER))

	c.Order("-company", "-branch")
	test.Equal("ORDER BY company DESC, branch DESC", c.Config(true, sqlquery.ORDER))

	c.Order("-company", "branch")
	test.Equal("ORDER BY company DESC, branch ASC", c.Config(true, sqlquery.ORDER))
}

func TestCondition_Limit(t *testing.T) {
	test := assert.New(t)
	c := sqlquery.Condition{}

	// TODO check if DB allow 0 values
	c.Limit(0)
	test.Equal("LIMIT 0", c.Config(true, sqlquery.LIMIT))

	c.Limit(1)
	test.Equal("LIMIT 1", c.Config(true, sqlquery.LIMIT))
}

func TestCondition_Offset(t *testing.T) {
	test := assert.New(t)
	c := sqlquery.Condition{}

	// TODO check if DB allow 0 values
	c.Offset(0)
	test.Equal("OFFSET 0", c.Config(true, sqlquery.OFFSET))

	c.Offset(10)
	test.Equal("OFFSET 10", c.Config(true, sqlquery.OFFSET))
}

func TestCondition_On(t *testing.T) {
	test := assert.New(t)
	c := sqlquery.Condition{}
	c.On("user.company = company.id AND id = "+sqlquery.PLACEHOLDER+" AND name="+sqlquery.PLACEHOLDER, 1, "John Doe")
	test.Equal("ON user.company = company.id AND id = 1 AND name=John Doe", c.Config(true, sqlquery.ON))

	// ok: calling ON again - conditions are getting replaced.
	c.On("user.company = company.id AND id = "+sqlquery.PLACEHOLDER+" AND name="+sqlquery.PLACEHOLDER, 1, "John Doe")
	test.Equal("ON user.company = company.id AND id = 1 AND name=John Doe", c.Config(true, sqlquery.ON))

	// ok: ON is used but contains an error, because its getting reset, the value is empty.
	c.On("user.company = company.id AND id = "+sqlquery.PLACEHOLDER+" AND name="+sqlquery.PLACEHOLDER, 1, 2, "John Doe")
	test.Equal("", c.Config(true, sqlquery.ON))

}

func TestCondition_WhereWithIN(t *testing.T) {
	test := assert.New(t)
	c := sqlquery.Condition{}
	c.Where("id IN ("+sqlquery.PLACEHOLDER+")", []int{1, 2, 3, 4, 5})
	test.Equal("WHERE id IN (1, 2, 3, 4, 5)", c.Config(true, sqlquery.WHERE))
	test.Equal("WHERE id IN ("+sqlquery.PLACEHOLDER+strings.Repeat(", "+sqlquery.PLACEHOLDER, 4)+")", c.Config(false, sqlquery.WHERE))

	c.Reset(sqlquery.WHERE)
	c.Where("name IN ("+sqlquery.PLACEHOLDER+")", []string{"John", "Doe", "Foo", "Bar"})
	test.Equal("WHERE name IN (John, Doe, Foo, Bar)", c.Config(true, sqlquery.WHERE))
	test.Equal("WHERE name IN ("+sqlquery.PLACEHOLDER+strings.Repeat(", "+sqlquery.PLACEHOLDER, 3)+")", c.Config(false, sqlquery.WHERE))
}

func TestCondition_WhereEmpty(t *testing.T) {
	test := assert.New(t)
	c := sqlquery.Condition{}
	c.Where("")
	test.Equal("", c.Config(true, sqlquery.WHERE))
	test.Equal("", c.Config(false, sqlquery.WHERE))
}

func TestCondition_WhereExample(t *testing.T) {
	test := assert.New(t)
	c := sqlquery.Condition{}

	var arguments []interface{}
	var where []string

	where = append(where, "(STANDORT = ? AND CAR IN (?))")
	where = append(where, "(STANDORT = ?)")
	where = append(where, "(STANDORT = ? AND CAR IN (?))")

	arguments = append(arguments, "INNSBRUCK")
	arguments = append(arguments, []string{"BMW", "BMW2"})
	arguments = append(arguments, "WERNDORF_2")
	arguments = append(arguments, "SENEC2")
	arguments = append(arguments, []string{"KTM1", "KTM2", "KTM3"})

	c.Where(strings.Join(where, " OR "), arguments...)

	test.Equal("WHERE (STANDORT = INNSBRUCK AND CAR IN (BMW, BMW2)) OR (STANDORT = WERNDORF_2) OR (STANDORT = SENEC2 AND CAR IN (KTM1, KTM2, KTM3))", c.Config(true, sqlquery.WHERE))
	test.Equal("WHERE (STANDORT = ? AND CAR IN (?, ?)) OR (STANDORT = ?) OR (STANDORT = ? AND CAR IN (?, ?, ?))", c.Config(false, sqlquery.WHERE))
}

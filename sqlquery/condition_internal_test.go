// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlquery

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCondition_render(t *testing.T) {
	test := assert.New(t)
	c := Condition{}

	// TESTING arguments order
	// Testing wrong type on multiple insert
	// testing argument mismatch
	p := &Placeholder{}
	p.Char = "$"
	p.Numeric = true

	// ok: testing render and arguments order
	c.Limit(1).Offset(10).Order("created").Where("id = "+PLACEHOLDER, 1).Having("id = "+PLACEHOLDER, 2).Where("name = "+PLACEHOLDER, "Foo").Having("surname = "+PLACEHOLDER, "Bar").Group("id", "name")
	sql, err := c.render(p)
	test.NoError(err)
	test.Equal([]interface{}{1, "Foo", 2, "Bar"}, c.arguments())

	test.Equal("WHERE id = $1 AND name = $2 GROUP BY id, name HAVING id = $3 AND surname = $4 ORDER BY created ASC LIMIT 1 OFFSET 10", sql)

	// err: test argument mismatch
	c.Where("id = "+PLACEHOLDER, 1, 2)
	sql, err = c.render(p)
	test.Error(err)
	test.Equal(fmt.Sprintf(ErrPlaceholderMismatch.Error(), "id = ?", 1, 2), err.Error())
	test.Equal("", sql)

	// err: test argument unsupported type
	c.Where("id IN ("+PLACEHOLDER+")", []*Placeholder{p})
	sql, err = c.render(p)
	test.Error(err)
	test.Equal(fmt.Sprintf(ErrArgumentType.Error(), "ptr"), err.Error())
	test.Equal("", sql)
}

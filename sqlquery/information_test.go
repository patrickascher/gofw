// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlquery_test

import (
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInformation_Describe(t *testing.T) {
	test := assert.New(t)

	b, err := sqlquery.New(sqlquery.Config{Driver: "test", Database: "company"}, nil)
	test.NoError(err)

	// ok - columns (name,surname) will be described in company.user
	_, err = b.Information("user").Describe("name", "surname")
	test.NoError(err)
	test.Equal("company", mockProvider.describeDb)
	test.Equal("user", mockProvider.describeTable)
	test.Equal([]string{"name", "surname"}, mockProvider.describeCols)

	// ok - columns (name,surname) will be described in c.user
	_, err = b.Information("c.user").Describe("name", "surname")
	test.NoError(err)
	test.Equal("c", mockProvider.describeDb)
	test.Equal("user", mockProvider.describeTable)
	test.Equal([]string{"name", "surname"}, mockProvider.describeCols)
}

func TestInformation_ForeignKeys(t *testing.T) {
	test := assert.New(t)

	b, err := sqlquery.New(sqlquery.Config{Driver: "test", Database: "company"}, nil)
	test.NoError(err)

	// ok - fk of the table company.user
	_, err = b.Information("user").ForeignKeys()
	test.NoError(err)
	test.Equal("company", mockProvider.fkDb)
	test.Equal("user", mockProvider.fkTable)

	// ok - fk of the table c.user
	_, err = b.Information("c.user").ForeignKeys()
	test.NoError(err)
	test.Equal("c", mockProvider.fkDb)
	test.Equal("user", mockProvider.fkTable)
}

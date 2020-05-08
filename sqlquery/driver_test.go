// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlquery_test

import (
	"database/sql"
	"github.com/patrickascher/gofw/sqlquery/types"

	"github.com/patrickascher/gofw/sqlquery"
)

var mockProvider *mockDriver

// New creates a in-memory cache by the given options.
func mockMock(cfg sqlquery.Config, db *sql.DB) (sqlquery.DriverI, error) {
	mockProvider = &mockDriver{}
	mockProvider.cfg = cfg
	return mockProvider, nil
}

type mockDriver struct {
	describeDb    string
	describeTable string
	describeCols  []string

	fkTable string
	fkDb    string

	cfg sqlquery.Config
}

func (m *mockDriver) Config() sqlquery.Config {
	return m.cfg
}

func (m *mockDriver) Connection() *sql.DB {
	//return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=true", c.Username, c.Password, c.Host, c.Port, c.Database)
	return nil
}

func (m *mockDriver) QuoteCharacterColumn() string {
	return "'"
}

func (m *mockDriver) Describe(b *sqlquery.Builder, db string, table string, cols []string) ([]sqlquery.Column, error) {
	m.describeDb = db
	m.describeTable = table
	m.describeCols = cols
	return nil, nil
}

func (m *mockDriver) ForeignKeys(b *sqlquery.Builder, db string, table string) ([]*sqlquery.ForeignKey, error) {
	m.fkDb = db
	m.fkTable = table
	return nil, nil
}

func (m *mockDriver) Placeholder() *sqlquery.Placeholder {
	return &sqlquery.Placeholder{Char: "?", Numeric: false}
}

func (m *mockDriver) TypeMapping(s string, column sqlquery.Column) types.Interface {
	return nil
}

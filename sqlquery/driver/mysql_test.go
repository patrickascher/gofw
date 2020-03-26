// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package driver

import (
	"testing"

	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
)

func TestMysql_TypeMapping(t *testing.T) {

	test := assert.New(t)
	m := mysql{}

	var tests = []struct {
		raw  string
		kind string
	}{
		{raw: "bigint", kind: "Integer"},
		{raw: "int", kind: "Integer"},
		{raw: "mediumint", kind: "Integer"},
		{raw: "smallint", kind: "Integer"},
		{raw: "tinyint", kind: "Integer"},
		{raw: "bigint unsigned", kind: "Integer"},
		{raw: "int unsigned", kind: "Integer"},
		{raw: "mediumint unsigned", kind: "Integer"},
		{raw: "smallint unsigned", kind: "Integer"},
		{raw: "tinyint unsigned", kind: "Integer"},

		{raw: "decimal", kind: "Float"},
		{raw: "float", kind: "Float"},
		{raw: "double", kind: "Float"},

		{raw: "varchar", kind: "Text"},
		{raw: "char", kind: "Text"},

		{raw: "tinytext", kind: "TextArea"},
		{raw: "text", kind: "TextArea"},
		{raw: "mediumtext", kind: "TextArea"},
		{raw: "longtext", kind: "TextArea"},

		{raw: "time", kind: "Time"},
		{raw: "date", kind: "Date"},
		{raw: "datetime", kind: "DateTime"},
		{raw: "timestamp", kind: "DateTime"},

		{raw: "xxx", kind: "nil"},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			c := sqlquery.Column{}
			ct := m.TypeMapping(tt.raw, c)

			if tt.kind == "nil" {
				test.Nil(ct)
			} else {
				test.Equal(tt.kind, ct.Kind())
				test.Equal(tt.raw, ct.Raw())
			}
		})
	}
}

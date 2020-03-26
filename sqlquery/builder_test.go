// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlquery_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
)

func init() {
	// ok: register successful
	_ = sqlquery.Register("test", mockMock)
}

func TestRegister(t *testing.T) {
	test := assert.New(t)

	// error: no provider-name and provider is given
	err := sqlquery.Register("", nil)
	test.Error(err)
	test.Equal(err.Error(), sqlquery.ErrNoProvider.Error())

	// error: no provider is given
	err = sqlquery.Register("mock", nil)
	test.Error(err)
	test.Equal(err.Error(), sqlquery.ErrNoProvider.Error())

	// error: no provider-name is given
	err = sqlquery.Register("", mockMock)
	test.Error(err)
	test.Equal(err.Error(), sqlquery.ErrNoProvider.Error())

	// ok: register successful
	err = sqlquery.Register("mock", mockMock)
	test.NoError(err)

	// error: multiple registration
	err = sqlquery.Register("mock", mockMock)
	test.Error(err)
	test.Equal(fmt.Sprintf(sqlquery.ErrProviderAlreadyExists.Error(), "mock"), err.Error())
}

func TestNew(t *testing.T) {
	test := assert.New(t)

	// error: no config driver -> unregistered empty driver
	b, err := sqlquery.New(sqlquery.Config{Driver: "mock2"}, nil)
	test.Equal(sqlquery.Builder{}, b)
	test.Error(err)
	test.Equal(fmt.Sprintf(sqlquery.ErrUnknownProvider.Error(), "mock2"), err.Error())

	// error: no registered dummy cache provider
	b, err = sqlquery.New(sqlquery.Config{Driver: "mock2"}, nil)
	test.Equal(sqlquery.Builder{}, b)
	test.Error(err)
	test.Equal(fmt.Sprintf(sqlquery.ErrUnknownProvider.Error(), "mock2"), err.Error())

	// ok
	b, err = sqlquery.New(sqlquery.Config{Driver: "mock"}, nil)
	test.NoError(err)
	test.NotNil(b)
}

func TestBuilder_Information(t *testing.T) {
	test := assert.New(t)

	b, err := sqlquery.New(sqlquery.Config{Driver: "mock", Database: "company"}, nil)
	test.NoError(err)

	_, err = b.Information("user").Describe("name", "surname")
	test.NoError(err)
	test.Equal("company", mockProvider.describeDb)
	test.Equal("user", mockProvider.describeTable)

	// ok - columns (name,surname) will be described in c.user
	_, err = b.Information("c.user").Describe("name", "surname")
	test.NoError(err)
	test.Equal("c", mockProvider.describeDb)
	test.Equal("user", mockProvider.describeTable)
}

func TestBuilder_QuoteIdentifier(t *testing.T) {
	test := assert.New(t)

	b, err := sqlquery.New(sqlquery.Config{Driver: "mock", Database: "company"}, nil)
	test.NoError(err)

	test.Equal("'gofw'.'user' 'usr'", b.QuoteIdentifier("gofw.user AS usr"))
}

func TestBuilder_Driver(t *testing.T) {
	test := assert.New(t)

	b, err := sqlquery.New(sqlquery.Config{Driver: "mock", Database: "company"}, nil)
	test.NoError(err)

	test.IsType(&mockDriver{}, b.Driver())
	test.Equal(sqlquery.Config{Driver: "mock", Database: "company"}, b.Driver().Config())
}

func TestBuilder_Select(t *testing.T) {
	test := assert.New(t)

	b, err := sqlquery.New(sqlquery.Config{Driver: "mock", Database: "company"}, nil)
	test.NoError(err)

	sel := b.Select("user")

	test.NoError(err)
	test.IsType(&sqlquery.Select{}, sel)
}

// Example of the basic usage of the builder. For more details please see the Examples in the methods.
func Example() {
	cfg := sqlquery.Config{
		Driver:   "mysql",
		Host:     "127.0.0.1",
		Port:     3319,
		Username: "root",
		Password: "root",
		Database: "gofw",
	}

	// Builder from config
	builder, err := sqlquery.New(cfg, nil)
	if err != nil {
		//...
	}

	// Select stmt
	row, err := builder.Select("users").First()
	if err != nil {
		//...
	}
	fmt.Println(row)

	// Insert stmt
	rows, err := builder.Insert("users").Values([]map[string]interface{}{{"id:": 1, "name": "John", "surname": "Doe"}}).Exec()
	if err != nil {
		//...
	}
	fmt.Println(rows)

	// Update stmt
	res, err := builder.Update("users").Set(map[string]interface{}{"name": "John", "surname": "Doe"}).Where("id = ?", 1).Exec()
	if err != nil {
		//...
	}
	fmt.Println(res)

	// Delete stmt
	res, err = builder.Delete("users").Where("id = ?", 1).Exec()
	if err != nil {
		//...
	}
	fmt.Println(res)

	// Describe stmt
	cols, err := builder.Information("users").Describe()
	if err != nil {
		//...
	}
	fmt.Println(cols)

	// ForeignKey stmt
	fks, err := builder.Information("users").ForeignKeys()
	if err != nil {
		//...
	}
	fmt.Println(fks)
}

func ExampleNew() {
	cfg := sqlquery.Config{
		Driver:   "mysql",
		Host:     "127.0.0.1",
		Port:     3319,
		Username: "root",
		Password: "root",
		Database: "gofw",
	}

	// Builder from config.
	builder, err := sqlquery.New(cfg, nil)
	if err != nil {
		//...
	}
	fmt.Println(builder)

	// Builder from adapter - like this a global open pool could be provided.
	db, err := sql.Open("mysql", "dns")
	if err != nil {
		//...
	}
	builder, err = sqlquery.New(cfg, db)
	if err != nil {
		//...
	}
	fmt.Println(builder)
}

func ExampleBuilder_Tx() {
	// Builder from config.
	b, err := sqlquery.New(sqlquery.Config{}, nil)
	if err != nil {
		//...
	}

	// start tx
	err = b.Tx()
	if err != nil {
		//...
	}

	// some bundled stmts
	_, err = b.Update("users").Set(map[string]interface{}{"id": 1, "name": "John"}).Where("id = ?", 1).Exec()
	if err != nil {
		err = b.Rollback()
		//...
	}
	_, err = b.Delete("users").Where("id = ?", 10).Exec()
	if err != nil {
		err = b.Rollback()
		//...
	}

	// commit tx
	err = b.Commit()
	if err != nil {
		//...
	}

}

func ExampleBuilder_Select() {

	// Builder from config.
	b, err := sqlquery.New(sqlquery.Config{}, nil)
	if err != nil {
		//...
	}

	// select a single row
	row, err := b.Select("users").
		Columns("id", "name").
		Where("id = ?", 1).
		First()
	if err != nil {
		//...
	}
	fmt.Println(row)

	// select all rows
	jc := sqlquery.Condition{}
	rows, err := b.Select("users").
		Columns("id", "name").
		Where("id > ?", 10).
		Limit(10).
		Offset(5).
		Order("surname").
		Group("age").
		Join(sqlquery.LEFT, "company", jc.On("user.company = company.id AND company.branches > ?", 2)).
		All()
	if err != nil {
		//...
	}
	_ = rows.Close()
	fmt.Println(rows)
}

func ExampleBuilder_Insert() {

	// Builder from config.
	b, err := sqlquery.New(sqlquery.Config{}, nil)
	if err != nil {
		//...
	}

	lastID := 0
	res, err := b.Insert("users").
		Batch(1).                                                    //if more than 1 values are set, a batching would be executed.
		Columns("id", "name").                                       // set specific columns order
		Values([]map[string]interface{}{{"id": 1, "name": "John"}}). // values
		LastInsertedID("id", &lastID).                               // set last inserted id
		Exec()                                                       //execute

	if err != nil {
		//...
	}
	fmt.Println(res)
}

func ExampleBuilder_Update() {

	// Builder from config.
	b, err := sqlquery.New(sqlquery.Config{}, nil)
	if err != nil {
		//...
	}

	res, err := b.Update("users").
		Columns("id", "name").
		Set(map[string]interface{}{"name": "Bar"}).
		Where("id = ?", 2).
		Exec() //execute

	if err != nil {
		//...
	}
	fmt.Println(res)
}

func ExampleBuilder_Delete() {

	// Builder from config.
	b, err := sqlquery.New(sqlquery.Config{}, nil)
	if err != nil {
		//...
	}

	res, err := b.Delete("users").
		Where("id = ?", 2).
		Exec() //execute

	if err != nil {
		//...
	}
	fmt.Println(res)
}

func ExampleBuilder_Information() {

	// Builder from config.
	b, err := sqlquery.New(sqlquery.Config{}, nil)
	if err != nil {
		//...
	}

	// Describe table columns.
	cols, err := b.Information("users").Describe("id", "name")
	if err != nil {
		//...
	}
	fmt.Println(cols)

	// FKs of the table.
	fks, err := b.Information("users").ForeignKeys()
	if err != nil {
		//...
	}
	fmt.Println(fks)
}

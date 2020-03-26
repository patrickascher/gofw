// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlquery_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/patrickascher/gofw/sqlquery"
	"github.com/patrickascher/gofw/sqlquery/driver"
	"github.com/stretchr/testify/assert"
)

func getBuilder() (sqlquery.Builder, error) {
	cfg := sqlquery.Config{
		Driver:   "mysql",
		Host:     "127.0.0.1",
		Port:     3319,
		Username: "root",
		Password: "root",
		Database: "gofw",
	}

	return sqlquery.New(cfg, nil)
}

func truncateTables() error {
	b, err := getBuilder()
	if err != nil {
		return err
	}

	_, err = b.Delete("users").Exec()
	if err != nil {
		return err
	}
	_, err = b.Delete("companies").Exec()
	if err != nil {
		return err
	}

	return nil
}

func insertDummyData() error {
	b, err := getBuilder()
	if err != nil {
		return err
	}

	// ok: insert entry
	_, err = b.Insert("users").Values([]map[string]interface{}{{"id": 1, "name": "John", "surname": "Doe"}, {"id": 2, "name": "Bar", "surname": "Foo"}}).Exec()
	if err != nil {
		return err
	}
	return nil
}

func countRows(b *sqlquery.Builder) int {
	var err error
	if b == nil {
		b1, err := getBuilder()
		if err != nil {
			return 0
		}
		b = &b1
	}

	rows, err := b.Select("users").All()
	if err != nil {
		return 0
	}
	i := 0
	for rows.Next() {
		i++
	}
	err = rows.Close()
	if err != nil {
		return 0
	}
	return i
}
func TestDb_FirstAll(t *testing.T) {
	test := assert.New(t)

	err := truncateTables()
	test.NoError(err)
	b, err := getBuilder()
	test.NoError(err)

	test.Equal("mysql", b.Driver().Config().Driver)

	// ok: first with no entries - empty row
	r, err := b.Select("users").First()
	test.NoError(err)

	user := struct {
		id      int
		name    string
		surname string
		company sql.NullInt64
	}{}
	err = r.Scan(&user.id, &user.name, &user.surname, &user.company)
	test.Error(err)
	test.Equal(sql.ErrNoRows, err)

	// insert entry
	err = insertDummyData()
	test.NoError(err)

	// ok: first found row
	r, err = b.Select("users").First()
	test.NoError(err)

	user = struct {
		id      int
		name    string
		surname string
		company sql.NullInt64
	}{}
	err = r.Scan(&user.id, &user.name, &user.surname, &user.company)
	test.NoError(err)
	test.Equal(1, user.id)
	test.Equal("John", user.name)
	test.Equal("Doe", user.surname)

	// err: first with wrong syntax
	r, err = b.Select("users").Where("id = "+sqlquery.PLACEHOLDER+" AND ?", 1).First()
	test.Error(err)
	test.Nil(r)

	// ok: All
	rows, err := b.Select("users").All()
	test.NoError(err)
	i := 0
	result := ""
	for rows.Next() {
		rows.Scan(&user.id, &user.name, &user.surname, &user.company)
		result += user.name
		result += user.surname
		i++
	}
	err = rows.Close()
	test.NoError(err)
	test.True(i == 2)
	test.Equal("JohnDoeBarFoo", result)

	// err: all with wrong syntax
	rows, err = b.Select("users").Where("id = "+sqlquery.PLACEHOLDER+" AND ?", 1).All()
	test.Error(err)
	test.Nil(rows)
}

func TestDb_InsertAndBatch(t *testing.T) {
	test := assert.New(t)

	err := truncateTables()
	test.NoError(err)
	b, err := getBuilder()
	test.NoError(err)

	// ok: insert entry
	_, err = b.Insert("users").Values([]map[string]interface{}{{"id": 1, "name": "John", "surname": "Doe"}, {"id": 2, "name": "Bar", "surname": "Foo"}}).Exec()
	test.NoError(err)
	test.Equal(2, countRows(nil))

	// err: wrong syntax (values are missing)
	_, err = b.Insert("users").Exec()
	test.Error(err)

	// ok: last id
	err = truncateTables()
	test.NoError(err)
	lID := 0
	_, err = b.Insert("users").LastInsertedID("id", &lID).Values([]map[string]interface{}{{"id": 1, "name": "John", "surname": "Doe"}, {"id": 2, "name": "Bar", "surname": "Foo"}}).Exec()
	test.NoError(err)
	test.Equal(2, lID)

	// error: idx field does not exist
	err = truncateTables()
	test.NoError(err)
	lID = 0
	_, err = b.Insert("users").LastInsertedID("idx", &lID).Values([]map[string]interface{}{{"idx": 1, "name": "John", "surname": "Doe"}, {"idx": 2, "name": "Bar", "surname": "Foo"}}).Exec()
	test.Error(err)

	// ok: batch insert
	err = truncateTables()
	test.NoError(err)
	_, err = b.Insert("users").Batch(1).Values([]map[string]interface{}{{"id": 1, "name": "John", "surname": "Doe"}, {"id": 2, "name": "Bar", "surname": "Foo"}}).Exec()
	test.NoError(err)
	// check if two entries were made
	test.Equal(2, countRows(nil))
}

func TestDb_InsertTx(t *testing.T) {
	test := assert.New(t)

	err := truncateTables()
	test.NoError(err)
	b, err := getBuilder()
	test.NoError(err)

	err = b.Commit()
	test.Error(err)
	test.Equal(sqlquery.ErrNoTx.Error(), err.Error())

	err = b.Rollback()
	test.Error(err)
	test.Equal(sqlquery.ErrNoTx.Error(), err.Error())

	// new TX
	err = b.Tx()
	test.NoError(err)

	// ok: insert entry not committed, no entries.
	_, err = b.Insert("users").Values([]map[string]interface{}{{"id": 1, "name": "John", "surname": "Doe"}, {"id": 2, "name": "Bar", "surname": "Foo"}}).Exec()
	test.NoError(err)
	// check outside the tx
	test.Equal(0, countRows(nil))
	// check all inside the tx
	test.Equal(2, countRows(&b))
	// check first inside the tx
	row, err := b.Select("users").First()
	test.NoError(err)
	user := struct {
		id      int
		name    string
		surname string
		company sql.NullInt64
	}{}
	err = row.Scan(&user.id, &user.name, &user.surname, &user.company)
	test.NoError(err)
	test.Equal(1, user.id)
	test.Equal("John", user.name)
	test.Equal("Doe", user.surname)

	err = b.Rollback()
	test.NoError(err)

	test.Equal(0, countRows(nil))

	// -----------------------------------------

	// err: tx was already committed or rolled back
	_, err = b.Insert("users").Values([]map[string]interface{}{{"id": 1, "name": "John", "surname": "Doe"}, {"id": 2, "name": "Bar", "surname": "Foo"}}).Exec()
	test.Error(err)

	// ok: recreated tx
	err = b.Tx()
	test.NoError(err)

	_, err = b.Insert("users").Values([]map[string]interface{}{{"id": 1, "name": "John", "surname": "Doe"}, {"id": 2, "name": "Bar", "surname": "Foo"}}).Exec()
	test.NoError(err)
	test.Equal(0, countRows(nil))

	test.Equal(0, countRows(nil))

	err = b.Commit()
	test.NoError(err)

	test.Equal(2, countRows(nil))

}

func TestDb_Delete(t *testing.T) {
	test := assert.New(t)

	err := truncateTables()
	test.NoError(err)
	b, err := getBuilder()
	test.NoError(err)

	// ok: check delete *
	err = insertDummyData()
	test.NoError(err)
	_, err = b.Delete("users").Exec()
	test.NoError(err)
	test.Equal(0, countRows(nil))

	// ok: check delete where
	err = truncateTables()
	test.NoError(err)
	err = insertDummyData()
	test.NoError(err)
	_, err = b.Delete("users").Where("id = "+sqlquery.PLACEHOLDER, 1).Exec()
	test.NoError(err)
	test.Equal(1, countRows(nil))

	// error: table does not exist
	err = truncateTables()
	test.NoError(err)
	err = insertDummyData()
	test.NoError(err)
	_, err = b.Delete("userx").Exec()
	test.Error(err)
	test.Equal(2, countRows(nil))

	// error: where syntax error
	err = truncateTables()
	test.NoError(err)
	err = insertDummyData()
	test.NoError(err)
	_, err = b.Delete("users").Where("id = "+sqlquery.PLACEHOLDER+" AND "+sqlquery.PLACEHOLDER, 1).Exec()
	test.Error(err)
	test.Equal(2, countRows(nil))

}

func TestDb_Update(t *testing.T) {
	test := assert.New(t)

	err := truncateTables()
	test.NoError(err)
	b, err := getBuilder()
	test.NoError(err)

	// ok: update id 1
	err = insertDummyData()
	test.NoError(err)
	_, err = b.Update("users").Where("id = "+sqlquery.PLACEHOLDER, 1).Set(map[string]interface{}{"name": "John_", "surname": "Doe_"}).Exec()
	test.NoError(err)
	test.Equal(2, countRows(nil))
	r, err := b.Select("users").Where("id = "+sqlquery.PLACEHOLDER, 1).First()
	test.NoError(err)
	user := struct {
		id      int
		name    string
		surname string
		company sql.NullInt64
	}{}
	err = r.Scan(&user.id, &user.name, &user.surname, &user.company)
	test.NoError(err)
	test.Equal("John_", user.name)
	test.Equal("Doe_", user.surname)

	// err: table does not exist
	err = truncateTables()
	test.NoError(err)
	err = insertDummyData()
	test.NoError(err)
	_, err = b.Update("usersx").Where("id = "+sqlquery.PLACEHOLDER, 1).Set(map[string]interface{}{"name": "John_", "surname": "Doe_"}).Exec()
	test.Error(err)

	// err: value is not set
	err = truncateTables()
	test.NoError(err)
	err = insertDummyData()
	test.NoError(err)
	_, err = b.Update("usersx").Where("id = "+sqlquery.PLACEHOLDER, 1).Exec()
	test.Error(err)
	test.Equal(sqlquery.ErrValueMissing.Error(), err.Error())
}

func TestDb_NewWithAdapter(t *testing.T) {
	test := assert.New(t)

	cfg := sqlquery.Config{
		Driver:   "mysql",
		Host:     "127.0.0.1",
		Port:     3319,
		Username: "root",
		Password: "root",
		Database: "gofw",
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=true", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database))
	test.NoError(err)

	// setting own adapter.
	_, err = sqlquery.New(cfg, db)
	test.NoError(err)
}

func TestDb_Describe(t *testing.T) {
	test := assert.New(t)

	err := truncateTables()
	test.NoError(err)
	b, err := getBuilder()
	test.NoError(err)

	// ok
	cols, err := b.Information("users").Describe()
	test.NoError(err)
	test.True(len(cols) > 0)

	// err: table does not exist
	cols, err = b.Information("usersx").Describe()
	test.Error(err)
	test.Equal(fmt.Sprintf(driver.ErrTableDoesNotExist.Error(), "gofw.usersx"), err.Error()) //TODO database should be dynamic

	// ok: specific columns
	cols, err = b.Information("users").Describe("id", "name")
	test.NoError(err)
	test.True(len(cols) == 2)
	test.Equal("id", cols[0].Name)
	test.Equal(true, cols[0].Autoincrement)
	test.Equal(sql.NullString{String: "", Valid: false}, cols[0].DefaultValue)
	test.Equal(sql.NullInt64{Int64: 0, Valid: false}, cols[0].Length)
	test.Equal(false, cols[0].NullAble)
	test.Equal(1, cols[0].Position)
	test.Equal(true, cols[0].PrimaryKey)
	test.Equal("users", cols[0].Table)
	//	test.Equal(nil, cols[0].Type)

	test.Equal("name", cols[1].Name)
	test.Equal(false, cols[1].Autoincrement)
	test.Equal(sql.NullString{String: "", Valid: false}, cols[1].DefaultValue)
	test.Equal(sql.NullInt64{Int64: 100, Valid: true}, cols[1].Length)
	test.Equal(true, cols[1].NullAble)
	test.Equal(2, cols[1].Position)
	test.Equal(false, cols[1].PrimaryKey)
	test.Equal("users", cols[1].Table)
	//	test.Equal(nil, cols[1].Type)

}

func TestDb_Foreignkey(t *testing.T) {
	test := assert.New(t)

	err := truncateTables()
	test.NoError(err)
	b, err := getBuilder()
	test.NoError(err)

	// ok
	fk, err := b.Information("users").ForeignKeys()
	test.NoError(err)
	test.True(len(fk) > 0)
	test.True(len(fk[0].Name) > 0)
	test.Equal("users", fk[0].Primary.Table)
	test.Equal("company", fk[0].Primary.Column)
	test.Equal("companies", fk[0].Secondary.Table)
	test.Equal("id", fk[0].Secondary.Column)

	fmt.Println(fk)

}

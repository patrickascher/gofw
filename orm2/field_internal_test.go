package orm2

import (
	"fmt"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

type carNoPrimary struct {
	Model
	Name string
}

type carTableDoesNotExist struct {
	Model
	ID    int
	Names string
}

type carFieldDoesNotExist struct {
	Model
	ID     int
	Engine string
}

func (c carFieldDoesNotExist) DefaultTableName() string {
	return "cars"
}

type carAdditionalPrimary struct {
	Model
	ID    int
	Brand string `orm:"primary"`
}

func (c carAdditionalPrimary) DefaultTableName() string {
	return "cars"
}

type carCustomPrimary struct {
	Model
	ID    int `orm:"primary"`
	Brand string
}

func (c carCustomPrimary) DefaultTableName() string {
	return "cars"
}

// test - no primary key is defined
func TestModel_createFieldsNoPrimary(t *testing.T) {
	test := assert.New(t)
	c := &carNoPrimary{}
	err := c.Init(c)
	test.Error(err)
	test.Equal(fmt.Sprintf(errPrimaryKey.Error(), c.modelName(true)), err.Error())
}

// test - car db table does not exist
func TestModel_carTabledDoesNotExist(t *testing.T) {
	test := assert.New(t)
	c := &carTableDoesNotExist{}
	err := c.Init(c)
	test.Error(err)
	test.True(strings.Contains(err.Error(), "sqlquery: table orm_test.car_table_does_not_exists or column does not exist [id names created_at updated_at deleted_at]"))
}

// test - car db table does not exist
func TestModel_carFieldDoesNotExist(t *testing.T) {
	test := assert.New(t)
	c := &carFieldDoesNotExist{}
	err := c.Init(c)
	test.Error(err)
	test.Equal(fmt.Sprintf(errDBColumn.Error(), "engine", "cars"), err.Error())
}

// test - custom primary (not declared as primary in db table)
func TestModel_carAdditionalPrimary(t *testing.T) {
	test := assert.New(t)
	c := &carAdditionalPrimary{}
	err := c.Init(c)
	test.Error(err)
	test.Equal(errDbSync.Error(), err.Error())
}

// test - custom primary (declared on db table)
func TestModel_carCustomPrimary(t *testing.T) {
	test := assert.New(t)
	c := &carCustomPrimary{}
	err := c.Init(c)
	test.NoError(err)
}

// test create Fields
func TestModel_createFields(t *testing.T) {
	test := assert.New(t)

	c := &car{}
	err := c.Init(c)
	assert.NoError(t, err)

	// checking if all fields exist (added createdAt, ignored Owner relation)
	fmt.Println(c.fields)
	if assert.Equal(t, 8, len(c.fields)) {
		// table driven tests
		var tests = []struct {
			Name        string
			Information sqlquery.Column
			Permission  Permission
			custom      bool
			sql         string
		}{
			{Name: "ID", Information: sqlquery.Column{Table: "cars", Name: "id", PrimaryKey: true}, Permission: Permission{Read: true, Write: true}, custom: false, sql: ""},
			{Name: "OwnerID", Information: sqlquery.Column{Table: "cars", Name: "owner_id", PrimaryKey: false}, Permission: Permission{Read: true, Write: true}, custom: false, sql: ""},
			{Name: "Brand", Information: sqlquery.Column{Table: "cars", Name: "brand", PrimaryKey: false}, Permission: Permission{Read: true, Write: true}, custom: false, sql: ""},
			{Name: "Type", Information: sqlquery.Column{Table: "cars", Name: "type", PrimaryKey: false}, Permission: Permission{Read: false, Write: false}, custom: false, sql: ""},
			{Name: "Custom", Information: sqlquery.Column{Table: "", Name: "custom", PrimaryKey: false}, Permission: Permission{Read: true, Write: true}, custom: true, sql: ""},
			{Name: "YearCheck", Information: sqlquery.Column{Table: "cars", Name: "year", PrimaryKey: false}, Permission: Permission{Read: true, Write: true}, custom: false, sql: "Concat(id,'.',brand,year)"},
			{Name: "CustomOne", Information: sqlquery.Column{Table: "", Name: "custom_one", PrimaryKey: false}, Permission: Permission{Read: true, Write: true}, custom: true, sql: ""},
			{Name: "CreatedAt", Information: sqlquery.Column{Table: "cars", Name: "created_at", PrimaryKey: false}, Permission: Permission{Read: true, Write: true}, custom: false, sql: ""},
		}

		for k, tt := range tests {
			t.Run(tt.Name, func(t *testing.T) {
				test.Equal(tt.Name, c.fields[k].Name)
				test.Equal(tt.Information.Table, c.fields[k].Information.Table)
				test.Equal(tt.Information.Name, c.fields[k].Information.Name)
				test.Equal(tt.Information.PrimaryKey, c.fields[k].Information.PrimaryKey)
				test.Equal(tt.Permission, c.fields[k].Permission)
				test.Equal(tt.custom, c.fields[k].Custom)
				test.Equal(tt.sql, c.fields[k].SqlSelect)
			})
		}
	}

}

func TestModel_configFieldByTag(t *testing.T) {
	test := assert.New(t)

	f := &Field{}

	// Permission
	test.Equal(Permission{Read: false, Write: false}, f.Permission)
	configFieldByTag(f, "permission:r")
	test.Equal(Permission{Read: true, Write: false}, f.Permission)
	configFieldByTag(f, "permission:w")
	test.Equal(Permission{Read: false, Write: true}, f.Permission)
	configFieldByTag(f, "permission")
	test.Equal(Permission{Read: false, Write: false}, f.Permission)

	// Custom
	test.False(f.Custom)
	configFieldByTag(f, "custom")
	test.True(f.Custom)

	// Primary
	test.False(f.Information.PrimaryKey)
	configFieldByTag(f, "primary")
	test.True(f.Information.PrimaryKey)

	// Column name
	test.Equal("", f.Information.Name)
	configFieldByTag(f, "column:test")
	test.Equal("test", f.Information.Name)

	// custom select
	test.Equal("", f.SqlSelect)
	configFieldByTag(f, "select:Concat(test)")
	test.Equal("Concat(test)", f.SqlSelect)

}

func TestModel_parseTag(t *testing.T) {
	test := assert.New(t)

	// table driven tests
	var tests = []struct {
		Tag string
		Map map[string]string
	}{
		{Tag: "column:abc;fk:rel", Map: map[string]string{"column": "abc", "fk": "rel"}},
		{Tag: "column", Map: map[string]string{"column": ""}},
		{Tag: ":;", Map: map[string]string{}},
	}

	for _, tt := range tests {
		t.Run(tt.Tag, func(t *testing.T) {
			test.Equal(tt.Map, parseTags(tt.Tag))
		})
	}
}

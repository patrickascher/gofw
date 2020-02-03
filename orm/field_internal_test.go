package orm

import (
	"fmt"
	"github.com/serenize/snaker"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestField_newValueInstanceFromType(t *testing.T) {
	cust := Customerfk{}

	custVal := newValueInstanceFromType(reflect.ValueOf(cust).FieldByName("Info").Type())
	assert.Equal(t, "Contactfk", custVal.Type().Name())
	assert.Equal(t, reflect.Struct, custVal.Type().Kind())
	custVal = newValueInstanceFromType(reflect.ValueOf(cust).FieldByName("Orders").Type())
	assert.Equal(t, "Orderfk", custVal.Type().Name())
	assert.Equal(t, reflect.Struct, custVal.Type().Kind())
	custVal = newValueInstanceFromType(reflect.ValueOf(cust).FieldByName("Service").Type())
	assert.Equal(t, "Servicefk", custVal.Type().Name())
	assert.Equal(t, reflect.Struct, custVal.Type().Kind())

	custprt := Customerptr{}
	custVal = newValueInstanceFromType(reflect.ValueOf(custprt).FieldByName("Info").Type())
	assert.Equal(t, "Contactfk", custVal.Type().Name())
	assert.Equal(t, reflect.Struct, custVal.Type().Kind())
	custVal = newValueInstanceFromType(reflect.ValueOf(custprt).FieldByName("Orders").Type())
	assert.Equal(t, "Orderfk", custVal.Type().Name())
	assert.Equal(t, reflect.Struct, custVal.Type().Kind())
	custVal = newValueInstanceFromType(reflect.ValueOf(custprt).FieldByName("Service").Type())
	assert.Equal(t, "Servicefk", custVal.Type().Name())
	assert.Equal(t, reflect.Struct, custVal.Type().Kind())

}

func TestField_implementsInterface(t *testing.T) {

	cust := Customerfk{}
	custVal, exists := reflect.TypeOf(cust).FieldByName("Info")
	if assert.True(t, exists) {
		assert.True(t, implementsInterface(custVal))
	}
	custVal, exists = reflect.TypeOf(cust).FieldByName("Orders")
	if assert.True(t, exists) {
		assert.True(t, implementsInterface(custVal))
	}
	custVal, exists = reflect.TypeOf(cust).FieldByName("Service")
	if assert.True(t, exists) {
		assert.True(t, implementsInterface(custVal))
	}

	custprt := Customerptr{}
	custVal, exists = reflect.TypeOf(custprt).FieldByName("Info")
	if assert.True(t, exists) {
		assert.True(t, implementsInterface(custVal))
	}
	custVal, exists = reflect.TypeOf(cust).FieldByName("Orders")
	if assert.True(t, exists) {
		assert.True(t, implementsInterface(custVal))
	}
	custVal, exists = reflect.TypeOf(cust).FieldByName("Service")
	if assert.True(t, exists) {
		assert.True(t, implementsInterface(custVal))
	}

}

func TestField_isStructLoopAndaddLoadedStruct(t *testing.T) {
	cust := Customerfk{}

	// addmin a loaded model to the parent.
	cust.loadedRel = append(cust.loadedRel, "testModel1")

	// check if the loaded models are getting passed to the child
	custVal := newValueInstanceFromType(reflect.ValueOf(cust).FieldByName("Info").Type())
	cust.addLoadedStruct(custVal.FieldByName("loadedRel"))
	//TODO test? unexported?

	// check if loop gets recognized
	assert.True(t, cust.isStructLoop("testModel1"))
	assert.False(t, cust.isStructLoop("testModel2"))

}

func TestField_initializeModelByValue(t *testing.T) {
	cust := Customerfk{}
	relation := newValueInstanceFromType(reflect.ValueOf(cust).FieldByName("Info").Type())

	// relation not initialized
	assert.False(t, relation.FieldByName("isInitialized").Bool())

	err := cust.initializeModelByValue(relation)
	assert.NoError(t, err)

	// relation must be initialized now
	assert.True(t, relation.FieldByName("isInitialized").Bool())

	// error because orm.Model forgot to get embedded
	type Usr struct {
		ID int
	}
	usr := Usr{}
	relation = newValueInstanceFromType(reflect.ValueOf(usr).Type())
	err = cust.initializeModelByValue(relation)
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf(ErrModelInit.Error(), "orm.Usr"), err.Error())

}

func TestField_getFieldsOrRelations_Fields(t *testing.T) {
	cust := Customerfk{}

	fields := cust.getFieldsOrRelations(&cust, false)
	assert.Equal(t, "ID", fields[0].Name)
	assert.Equal(t, "int", fields[0].Type.Name())
	assert.Equal(t, reflect.StructTag(""), fields[0].Tag)
	assert.Equal(t, []int([]int{2}), fields[0].Index)
	assert.Equal(t, false, fields[0].Anonymous)
	assert.Equal(t, "", fields[0].PkgPath)

	assert.Equal(t, "FirstName", fields[1].Name)
	assert.Equal(t, "String", fields[1].Type.Name())
	assert.Equal(t, reflect.StructTag(""), fields[1].Tag)
	assert.Equal(t, []int([]int{3}), fields[1].Index)
	assert.Equal(t, false, fields[1].Anonymous)
	assert.Equal(t, "", fields[1].PkgPath)

	assert.Equal(t, "LastName", fields[2].Name)
	assert.Equal(t, "String", fields[2].Type.Name())
	assert.Equal(t, reflect.StructTag(""), fields[2].Tag)
	assert.Equal(t, []int([]int{4}), fields[2].Index)
	assert.Equal(t, false, fields[2].Anonymous)
	assert.Equal(t, "", fields[2].PkgPath)

	assert.Equal(t, "CreatedAt", fields[3].Name)
	assert.Equal(t, "Time", fields[3].Type.Name())
	assert.Equal(t, reflect.StructTag(""), fields[3].Tag)
	assert.Equal(t, []int([]int{0}), fields[3].Index) // embedded counter starts again at 0
	assert.Equal(t, false, fields[3].Anonymous)
	assert.Equal(t, "", fields[3].PkgPath)

	assert.Equal(t, "UpdatedAt", fields[4].Name)
	assert.Equal(t, "Time", fields[4].Type.Name())
	assert.Equal(t, reflect.StructTag(""), fields[4].Tag)
	assert.Equal(t, []int([]int{1}), fields[4].Index)
	assert.Equal(t, false, fields[4].Anonymous)
	assert.Equal(t, "", fields[4].PkgPath)

	assert.Equal(t, "DeletedAt", fields[5].Name)
	assert.Equal(t, "Time", fields[5].Type.Name())
	assert.Equal(t, reflect.StructTag(""), fields[5].Tag)
	assert.Equal(t, []int([]int{2}), fields[5].Index)
	assert.Equal(t, false, fields[5].Anonymous)
	assert.Equal(t, "", fields[5].PkgPath)

}

func TestField_getFieldsOrRelations_Relations(t *testing.T) {
	cust := Customerfk{}

	relations := cust.getFieldsOrRelations(&cust, true)

	assert.Equal(t, "Info", relations[0].Name)
	assert.Equal(t, "Contactfk", relations[0].Type.Name())
	assert.Equal(t, reflect.Struct, relations[0].Type.Kind())
	assert.Equal(t, reflect.StructTag(""), relations[0].Tag)
	assert.Equal(t, []int([]int{6}), relations[0].Index)
	assert.Equal(t, false, relations[0].Anonymous)
	assert.Equal(t, "", relations[0].PkgPath)

	assert.Equal(t, "Orders", relations[1].Name)
	assert.Equal(t, reflect.Slice, relations[1].Type.Kind())
	assert.Equal(t, "Orderfk", relations[1].Type.Elem().Name())
	assert.Equal(t, reflect.StructTag(""), relations[1].Tag)
	assert.Equal(t, []int([]int{7}), relations[1].Index)
	assert.Equal(t, false, relations[1].Anonymous)
	assert.Equal(t, "", relations[1].PkgPath)

	assert.Equal(t, "Service", relations[2].Name)
	assert.Equal(t, reflect.Slice, relations[2].Type.Kind())
	assert.Equal(t, "Servicefk", relations[2].Type.Elem().Name())
	assert.Equal(t, reflect.StructTag(""), relations[2].Tag)
	assert.Equal(t, []int([]int{8}), relations[2].Index)
	assert.Equal(t, false, relations[2].Anonymous)
	assert.Equal(t, "", relations[2].PkgPath)
}

func TestField_addStructFieldsToTableColumn(t *testing.T) {

	cust := Customerfk{}
	cust.caller = &cust
	cust.table = &Table{}

	cust.addStructFieldsToTableColumn(cust.caller)

	// Check fields
	fields := []string{"ID", "FirstName", "LastName", "CreatedAt", "UpdatedAt", "DeletedAt"}
	for i, field := range fields {
		assert.Equal(t, field, cust.table.Cols[i].StructField)
		assert.Equal(t, &Permission{Read: true, Write: true}, cust.table.Cols[i].Permission)
		assert.Equal(t, snaker.CamelToSnake(field), cust.table.Cols[i].Information.Name)
	}
}

func TestField_parseTags(t *testing.T) {

	// testing empty tag
	m, err := parseTags("")
	assert.NoError(t, err)
	assert.True(t, m == nil)

	// testing: remove trailing separator, whitespaces
	m, err = parseTags(" permission: rw ; role: admin ; ")
	assert.NoError(t, err)
	assert.Equal(t, map[string]string(map[string]string{"permission": "rw", "role": "admin"}), m)

	// testing bad syntax
	m, err = parseTags(" permission; rw ; role: admin ; ")
	assert.Error(t, err)
	assert.True(t, m == nil)
}

func TestField_configColumnByTag(t *testing.T) {
	cust := Customerfk{}
	cust.caller = &cust
	cust.table = &Table{}
	cust.addStructFieldsToTableColumn(cust.caller)

	// empty tag
	err := configColumnByTag(cust.table.Cols[0], "")
	assert.NoError(t, err)

	// column name
	err = configColumnByTag(cust.table.Cols[0], "column:customer_id")
	assert.NoError(t, err)
	assert.Equal(t, "customer_id", cust.table.Cols[0].Information.Name)

	// permission W
	err = configColumnByTag(cust.table.Cols[0], "permission:w")
	assert.NoError(t, err)
	assert.Equal(t, &Permission{Read: false, Write: true}, cust.table.Cols[0].Permission)

	// permission R
	err = configColumnByTag(cust.table.Cols[0], "permission:r")
	assert.NoError(t, err)
	assert.Equal(t, &Permission{Read: true, Write: false}, cust.table.Cols[0].Permission)

	// permission RW
	err = configColumnByTag(cust.table.Cols[0], "permission:rw")
	assert.NoError(t, err)
	assert.Equal(t, &Permission{Read: true, Write: true}, cust.table.Cols[0].Permission)

	// select
	err = configColumnByTag(cust.table.Cols[0], "select:Count(*)")
	assert.NoError(t, err)
	assert.Equal(t, "Count(*)", cust.table.Cols[0].SqlSelect)

	// error
	err = configColumnByTag(cust.table.Cols[0], "key::value")
	assert.Error(t, err)

}

func TestField_isUnexportedField(t *testing.T) {

	cust := Customerfk{}
	field, exists := reflect.TypeOf(cust).FieldByName("ID")
	assert.True(t, exists)
	assert.False(t, isUnexportedField(field))

	field, exists = reflect.TypeOf(cust).FieldByName("Info")
	assert.True(t, exists)
	assert.False(t, isUnexportedField(field))

	// unexported struct field
	field, exists = reflect.TypeOf(cust).FieldByName("unexp")
	assert.True(t, exists)
	assert.True(t, isUnexportedField(field))

	// orm.Model struct
	field, exists = reflect.TypeOf(cust).FieldByName("Model")
	assert.True(t, exists)
	assert.True(t, isUnexportedField(field))

}

func TestField_reflectField(t *testing.T) {
	cust := Customerfk{}
	// field
	v := reflectField(&cust, "ID")
	assert.Equal(t, "int", v.Type().Name())
	// struct
	v = reflectField(&cust, "Info")
	assert.Equal(t, "Contactfk", v.Type().Name())
	// slice
	v = reflectField(&cust, "Orders")
	assert.Equal(t, "Orderfk", v.Type().Elem().Name())
}

func TestField_columnExists(t *testing.T) {
	cust := Customerfk{}
	cust.caller = &cust
	cust.table = &Table{}
	cust.addStructFieldsToTableColumn(cust.caller)

	assert.True(t, columnExists(&cust, "id"))
	assert.True(t, columnExists(&cust, "first_name"))
	assert.True(t, columnExists(&cust, "last_name"))

	assert.True(t, columnExists(&cust, "created_at"))
	assert.True(t, columnExists(&cust, "updated_at"))
	assert.True(t, columnExists(&cust, "deleted_at"))

	assert.False(t, columnExists(&cust, "id_"))

}
func TestField_fieldExists(t *testing.T) {
	cust := Customerfk{}
	cust.caller = &cust
	cust.table = &Table{}
	cust.addStructFieldsToTableColumn(cust.caller)

	assert.True(t, fieldExists(&cust, "ID"))
	assert.True(t, fieldExists(&cust, "FirstName"))
	assert.True(t, fieldExists(&cust, "LastName"))

	assert.True(t, fieldExists(&cust, "CreatedAt"))
	assert.True(t, fieldExists(&cust, "UpdatedAt"))
	assert.True(t, fieldExists(&cust, "DeletedAt"))

	assert.False(t, fieldExists(&cust, "unexp"))
}

func TestField_getColumnNameFromField(t *testing.T) {
	cust := Customerfk{}
	cust.caller = &cust
	cust.table = &Table{}
	cust.addStructFieldsToTableColumn(cust.caller)

	col, err := getColumnNameFromField(&cust, "ID")
	assert.NoError(t, err)
	assert.Equal(t, "id", col)

	col, err = getColumnNameFromField(&cust, "FirstName")
	assert.NoError(t, err)
	assert.Equal(t, "first_name", col)

	col, err = getColumnNameFromField(&cust, "LastName")
	assert.NoError(t, err)
	assert.Equal(t, "last_name", col)

	col, err = getColumnNameFromField(&cust, "CreatedAt")
	assert.NoError(t, err)
	assert.Equal(t, "created_at", col)

	col, err = getColumnNameFromField(&cust, "UpdatedAt")
	assert.NoError(t, err)
	assert.Equal(t, "updated_at", col)

	col, err = getColumnNameFromField(&cust, "DeletedAt")
	assert.NoError(t, err)
	assert.Equal(t, "deleted_at", col)

	col, err = getColumnNameFromField(&cust, "unexp")
	assert.Error(t, err)
	assert.Equal(t, "", col)
}

func TestField_checkPrimaryFieldsEmpty(t *testing.T) {
	cust := Customerfk{}
	err := cust.Initialize(&cust)
	assert.NoError(t, err)

	assert.True(t, checkPrimaryFieldsEmpty(&cust))

	cust.ID = 1
	assert.False(t, checkPrimaryFieldsEmpty(&cust))
}

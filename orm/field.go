package orm

import (
	"fmt"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/serenize/snaker"
	"reflect"
	"strings"
	"unsafe"
)

// newValueInstanceOfField creates a new value of the type.
// It ensures that the return value is a Value and no Pointer.
// If a Slice is given, it will take the struct type defined in the slice.
func newValueInstanceFromType(field reflect.Type) reflect.Value {

	// convert slice to single element
	var v reflect.Value
	if field.Kind() == reflect.Slice {
		//handel ptr
		if field.Elem().Kind() == reflect.Ptr {
			v = reflect.New(field.Elem().Elem())
		} else {
			v = reflect.New(field.Elem())
		}
	} else {
		if field.Kind() == reflect.Ptr {
			v = reflect.New(field.Elem())
		} else {
			v = reflect.New(field)
		}
	}

	// convert from ptr to value
	return reflect.Indirect(v)
}

// TODO: NewInstanceFromType is experimental, needed in grid for the select values...
func NewInstanceFromType(field reflect.Type) reflect.Value {
	return newValueInstanceFromType(field)
}

// implementsInterface checks if the given Field type implements the model.Interface.
func implementsInterface(field reflect.StructField) bool {
	i := reflect.TypeOf((*Interface)(nil)).Elem()
	v := newValueInstanceFromType(field.Type)

	return v.Addr().Type().Implements(i)
}

// isStructLoop checks if there is a loop defined.
// It checks if the struct was already loaded, if so, the relation will not get initialized again.
func (m Model) isStructLoop(rel string) bool {
	for _, n := range m.loadedRel {
		if rel == n {
			return true
		}
	}
	return false
}

// addLoadedStruct adds the loaded struct name to the slice of loaded relations.
func (m Model) addLoadedStruct(v reflect.Value) {
	v = reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
	v.Set(reflect.ValueOf(m.loadedRel))
}

// initializeModelByValue checks if the struct has the method Initialize and calls it.
// It also adds the parent loadedRelations, to avoid struct loops.
// Error will return if the method Initialize is not available or the method itself returns an error.
func (m Model) initializeModelByValue(value reflect.Value) error {
	c, d, err := m.caller.Cache()
	if err != nil {
		return err
	}
	if value.Addr().MethodByName("SetCache").IsValid() {
		rv := value.Addr().MethodByName("SetCache").Call([]reflect.Value{reflect.ValueOf(c), reflect.ValueOf(d)})
		// error handling
		if rv[0].Interface() != nil {
			return rv[0].Interface().(error)
		}
	}
	init := value.Addr().MethodByName("Initialize")
	if init.IsValid() {
		// add loadedRelations
		m.addLoadedStruct(value.FieldByName("loadedRel")) // FIX it: TODO it should be the other way around!!! get the m.loadedRel and set it to value.loadedRel
		// call init
		rv := init.Call([]reflect.Value{value.Addr()})
		// error handling
		if rv[0].Interface() != nil {
			return rv[0].Interface().(error)
		}

		return nil
	}
	return fmt.Errorf(ErrModelInit.Error(), value.Type().String())
}

// getFieldsOrRelations returns all exported fields or relations.
// It will skip the field if its unexported or skipped by tag '-'.
// Embedded structs are supported and their fields getting added as well.
// Info: caller is a interface{} instead of model.Interface because the embedded structs can be a normal struct (calls itself on iterating through fields).
func (m *Model) getFieldsOrRelations(caller interface{}, relation bool) []reflect.StructField {

	var fields []reflect.StructField
	var timeFields []reflect.StructField

	//reflect the caller
	v := reflect.ValueOf(caller)
	in := reflect.Indirect(v)

	if in.IsValid() {
		callerType := in.Type()
		//go through all caller fields
		for i := 0; i < callerType.NumField(); i++ {
			field := callerType.Field(i)

			// skip unexported, embedded fields and fields with the "-" tag
			if isUnexportedField(field) || field.Tag.Get(TagName) == TagSkip ||
				(implementsInterface(field) && relation == false) ||
				(!implementsInterface(field) && relation == true) {

				// adding CreatedAt, UpdatedAt, DeletedAt as embedded fields
				if field.Type.String() == "orm.Model" && relation == false {
					af := m.getFieldsOrRelations(in.FieldByName(field.Name).Interface(), false)
					timeFields = append(fields, af...)
				}

				continue
			}

			// adding embedded struct fields
			if field.Anonymous && relation == false {
				af := m.getFieldsOrRelations(in.FieldByName(field.Name).Interface(), false)
				fields = append(fields, af...)
				continue
			}

			fields = append(fields, field)
		}

		// adding time fields (createdAt, updatedAt, deletedAt) at the end
		if timeFields != nil {
			fields = append(fields, timeFields...)
		}
	}

	return fields
}

type Type interface {
	Kind() string
	Raw() string
}

type customType struct {
}

func (c *customType) Raw() string {
	return CustomImpl
}
func (c *customType) Kind() string {
	return CustomImpl
}

type CustomType struct {
}

func (ct *CustomType) Raw() string {
	return CustomImpl
}
func (ct *CustomType) Kind() string {
	return CustomImpl
}

// addStructFieldsToTableColumn adds the table columns.
// By default the column name is the snake_style of the struct field name.
// Permission Tag is getting set
// It parses the field tag and checks for the tag column,permission and select.
func (m *Model) addStructFieldsToTableColumn(caller interface{}) {
	for _, field := range m.getFieldsOrRelations(caller, false) {
		// create field and db column
		col := &Column{}
		col.StructField = field.Name
		col.Information = &sqlquery.Column{Name: snaker.CamelToSnake(col.StructField)}
		col.Permission = Permission{Read: true, Write: true}

		if m.strategy == CustomImpl && (col.StructField != CREATE && col.StructField != UPDATE && col.StructField != DELETE) {
			col.Information.Type = &CustomType{}
		}

		// parse tag and config the column
		_ = configColumnByTag(col, field.Tag.Get(TagName))

		col.Validator = NewValidator(field.Tag.Get(TagValidate))

		// add to model fields
		m.table.Cols = append(m.table.Cols, col)
	}
}

// parseTags returns key/value pairs of the given tag.
// If there is only the key set, the value will be an empty string.
func parseTags(tag string) (map[string]string, error) {

	if tag == "" {
		return nil, nil
	}

	// remove trailing separator
	tag = strings.TrimSpace(tag)
	if tag[len(tag)-1:] == TagSeparator {
		tag = tag[0 : len(tag)-1]
	}

	// configure model
	values := map[string]string{}
	for _, t := range strings.Split(tag, TagSeparator) {
		tag := strings.Split(t, TagKeyValue)
		if len(tag) != 2 {
			tag = append(tag, "")
			//return nil, ErrTagSyntax
		}

		// remove spaces
		tag[0] = strings.TrimSpace(tag[0])
		tag[1] = strings.TrimSpace(tag[1])

		values[tag[0]] = tag[1]

	}

	return values, nil
}

// configColumnByTag is parsing the field tags column,permission and select.
func configColumnByTag(col *Column, tag string) error {
	// skip if there is no defined tag
	if tag == "" {
		return nil
	}

	tags, err := parseTags(tag)
	if err != nil {
		return err
	}

	for k, v := range tags {
		switch k {
		case CustomImpl:
			col.Information.Type = &customType{}
		case "column":
			col.Information.Name = v
		case "permission":
			col.Permission.Read = false
			col.Permission.Write = false
			if strings.Contains(v, "r") {
				col.Permission.Read = true
			}
			if strings.Contains(v, "w") {
				col.Permission.Write = true
			}
		case "select":
			col.SqlSelect = v
		}
	}

	return nil
}

// isUnexportedField returns true if its the model struct or an unexported field.
func isUnexportedField(field reflect.StructField) bool {
	if field.Type.String() == "orm.Model" {
		return true //TODO check if that's also right when a user has an alias for import
	}
	if field.PkgPath != "" {
		return true
	}
	return false
}

// reflectField returns a struct field value by the given Interface and field name (string).
func reflectField(m Interface, field string) reflect.Value {
	return reflect.ValueOf(m).Elem().FieldByName(field)
}

// columnExists checks if the given column name exists in the given model (Interface).
func columnExists(m Interface, column string) bool {
	t := m.Table()
	for _, col := range t.Cols {
		if col.Information.Name == column {
			return true
		}
	}
	return false
}

// checkPrimaryFieldsEmpty is a helper to check if all primary fields have a value.
func checkPrimaryFieldsEmpty(m Interface) bool {

	pkey := m.Table().PrimaryKeys()
	if len(pkey) == 0 {
		return false
	}

	for _, col := range pkey {
		field := reflectField(m, col.StructField)
		if !isZeroOfUnderlyingType(field.Interface()) {
			return false
		}
	}
	return true
}

// fieldExists checks if the given struct field exists in the given model (Interface).
func fieldExists(m Interface, field string) bool {
	_, err := getColumnNameFromField(m, field)
	if err == nil {
		return true
	}
	return false
}

// getColumnNameFromField returns the column name of the given struct field.
// Will return an error if the field does not exist in the given model (Interface).
func getColumnNameFromField(m Interface, field string) (string, error) {
	if m.Table() != nil {
		for _, col := range m.Table().Cols {
			if col.StructField == strings.Title(field) { // ucfirst because of struct could be unexported but Fields has to be exported.
				return col.Information.Name, nil
			}
		}
	}

	return "", fmt.Errorf(ErrModelFieldNotFound.Error(), field, structName(m, true))
}

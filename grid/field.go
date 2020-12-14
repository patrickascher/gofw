package grid

import (
	"encoding/json"
	"fmt"
	"sort"
)

var errFieldNotFound = "grid: field %s was not found"

type Field struct {
	grid *Grid
	mode int
	// Struct field name or json name if defined.
	id string
	// referenceID is the database column name (used for conditions: orm - db column name)
	referenceId string
	// is it a primary key
	primary bool
	// field type (Integer, Text, hasOne,...)
	fieldType string
	// title - value can be different in the grid modes.
	_title map[int]string
	// description - value can be different in the grid modes.
	_description map[int]string
	// position - value can be different in the grid modes.
	_position map[int]int
	// remove - value can be different in the grid modes.
	_remove map[int]bool
	// hidden - value can be different in the grid modes.
	_hidden map[int]bool
	// view - value can be different in the grid modes.
	_view map[int]string
	// is read only
	readOnly bool
	// sorting allowed
	sortable bool
	// filter allowed
	filterable bool
	// customize filter
	where string
	// grouping allowed
	groupable bool
	// additional options
	options map[string]interface{}
	// callback of the field, used after the source results are fetched.
	callback          interface{}
	callbackArguments []interface{}
	// has additional fields (relations)
	fields   []Field
	relation bool
	// has an error
	error error
}

func (g Grid) sortFields() []Field {
	sort.Slice(g.fields, func(i, j int) bool {
		return g.fields[i].Position() < g.fields[j].Position()
	})
	return g.fields
}

// MarshalJson is used to create the header information of the field.
func (f Field) MarshalJSON() ([]byte, error) {
	rv := map[string]interface{}{}

	rv["id"] = f.id
	rv["type"] = f.fieldType
	if f.primary {
		rv["primary"] = f.primary
	}

	rv["title"] = f.Title()

	if v := f.Description(); v != "" {
		rv["description"] = v
	}
	rv["position"] = f.Position()

	if v := f.IsRemoved(); v {
		rv["remove"] = v
	}
	if v := f.IsHidden(); v {
		rv["hidden"] = v
	}
	if v := f.View(); v != "" {
		rv["view"] = v
	}
	if f.readOnly {
		rv["readOnly"] = f.readOnly
	}
	if f.sortable {
		rv["sortable"] = f.sortable
	}
	if f.filterable {
		rv["filterable"] = f.filterable
	}
	if f.groupable {
		rv["groupable"] = f.groupable
	}
	if f.readOnly {
		rv["readOnly"] = f.readOnly
	}
	if len(f.options) > 0 {
		rv["options"] = f.options
	}
	if len(f.fields) > 0 {
		rv["fields"] = f.fields
	}

	return json.Marshal(rv)
}

func (f *Field) SetDatabaseId(id string) {
	f.referenceId = id
}

func (f *Field) DatabaseId() string {
	return f.referenceId
}

func (f *Field) SetPrimary(primary bool) {
	f.primary = primary
}

func (f *Field) IsPrimary() bool {
	return f.primary
}

func (f *Field) SetWhere(where string) *Field {
	f.where = where
	return f
}

func (f *Field) SetFields(fields []Field) {
	f.fields = fields
}

func (f *Field) SetRelation(r bool) *Field {
	f.relation = r
	return f
}

func (f *Field) IsRelation() bool {
	return f.relation
}

func FieldsToString(fields []Field) []string {
	var rv []string
	for _, f := range fields {
		if !f.IsRemoved() {
			rv = append(rv, f.Id())
		}
	}
	return rv
}

// Field will return the field by the given name.
// If it was not found, an error will be set.
func (f *Field) Field(name string) *Field {

	for _, fn := range f.fields {
		if fn.id == name {
			return &fn
		}
	}

	// not found
	f.setError(fmt.Errorf(errFieldNotFound, name))
	return f
}

func (f *Field) Fields() []Field {
	return f.fields
}

func (f *Field) SetMode(mode int) {
	f.mode = mode
}

func (f *Field) SetId(id string) *Field {
	f.id = id
	return f
}

func (f Field) Id() string {
	return f.id
}

func (f *Field) SetReadOnly(readOnly bool) *Field {
	f.readOnly = readOnly
	return f
}

func (f Field) IsReadOnly() bool {
	return f.readOnly
}

func (f *Field) SetFieldType(fieldType string) *Field {
	f.fieldType = fieldType
	return f
}

func (f Field) FieldType() string {
	return f.fieldType
}

func (f *Field) SetTitle(title interface{}) *Field {
	f._title = setValueString(title)
	return f
}

func (f Field) Title() string {
	return f.grid.controller.T(f._title[f.mode])
}

func (f *Field) SetDescription(description interface{}) *Field {
	f._description = setValueString(description)
	return f
}

func (f Field) Description() string {
	return f._description[f.mode]
}

func (f *Field) SetPosition(position interface{}) *Field {
	f._position = setValueInt(position)
	return f
}

func (f Field) Position() int {
	return f._position[f.mode]
}

func (f *Field) SetRemove(remove interface{}) *Field {
	f._remove = setValueBool(remove)
	return f
}

func (f Field) IsRemoved() bool {
	// if the mode is Callback
	if f.mode != VTable && f.mode != VCreate && f.mode != VDetails && f.mode != VUpdate && f.mode != Export {
		return true
	}
	return f._remove[f.mode]
}

func (f *Field) SetHidden(hidden interface{}) *Field {
	f._hidden = setValueBool(hidden)
	return f
}

func (f Field) IsHidden() bool {
	return f._hidden[f.mode]
}

func (f *Field) SetSortable(sortable bool) *Field {
	f.sortable = sortable
	return f
}

func (f Field) IsSortable() bool {
	return f.sortable
}

func (f *Field) SetGroupable(groupable bool) *Field {
	f.groupable = groupable
	return f
}

func (f Field) IsGroupable() bool {
	return f.groupable
}

func (f *Field) SetFilterable(filterable bool) *Field {
	f.filterable = filterable
	return f
}

func (f Field) IsFilterable() bool {
	return f.filterable
}

func (f *Field) SetOption(key string, value interface{}) *Field {
	if f.options == nil {
		f.options = map[string]interface{}{}
	}
	if key == FeSelect && len(value.(Select).Items) > 0 {
		f.SetFieldType("Select")
		sel := value.(Select)
		sel.ValueField = "value"
		sel.TextField = "text"
		value = sel
		f.SetOption(FeReturnObject, false)
	}
	f.options[key] = value
	return f
}

func (f *Field) Options() map[string]interface{} {
	return f.options
}

func (f *Field) Option(key string) interface{} {
	if v, ok := f.options[key]; ok {
		return v
	}
	return nil
}

func (f *Field) SetView(view interface{}) *Field {
	f._view = setValueString(view)
	return f
}

func (f *Field) View() string {
	return f._view[f.mode]
}

func (f *Field) SetCallback(callback interface{}, args ...interface{}) *Field {
	f.callback = callback
	f.callbackArguments = args
	return f
}

func (f *Field) Callback() (callback interface{}, args interface{}) {
	return f.callback, f.callbackArguments
}

func (f *Field) setError(err error) {
	f.error = err
}

func (f Field) Error() error {
	return f.error
}

func setFieldModeRecursively(g *Grid, fields []Field) {

	// backend create,update and view should have the same settings
	mode := g.Mode()
	switch g.Mode() {
	case CREATE:
		mode = VCreate
	case UPDATE:
		mode = VUpdate
	case FILTERCONFIG:
		mode = VTable
	}

	// recursively add mode
	for k, f := range fields {
		fields[k].grid = g
		fields[k].mode = mode
		if g.config.Policy == 1 {
			fields[k].SetRemove(true)
		}
		if len(f.fields) > 0 {
			setFieldModeRecursively(g, fields[k].fields)
		}
	}
}

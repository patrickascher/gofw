package grid2

import (
	"encoding/json"
	"fmt"
	"sort"
)

var errFieldNotFound = "grid: field %s was not found"

type Field struct {
	// Struct field name or other identifier
	id string
	// referenceID (used for orm - db column name)
	referenceId string
	// is it a primary key
	primary bool
	// field type (Integer, Text, hasOne,...)
	fieldType string
	// title - value can be different in the grid modes.
	title value
	// description - value can be different in the grid modes.
	description value
	// position - value can be different in the grid modes.
	position value
	// remove - value can be different in the grid modes.
	remove value
	// hidden - value can be different in the grid modes.
	hidden value
	// view - value can be different in the grid modes.
	view value
	// is read only
	readOnly bool
	// sorting allowed
	sortable bool
	// filter allowed
	filterable bool
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

func (f *Field) SetReferenceId(id string) {
	f.referenceId = id
}

func (f *Field) SetPrimary(primary bool) {
	f.primary = primary
}

func (f *Field) SetFields(fields []Field) {
	f.fields = fields
}

func (f *Field) SetRelation(name bool) *Field {
	f.relation = true
	return f
}

func (f *Field) IsRelation() bool {
	return f.relation
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
	setValueHelper(&f.title, title)
	return f
}

func (f Field) Title() string {
	return f.title.getString()
}

func (f *Field) SetDescription(description interface{}) *Field {
	setValueHelper(&f.description, description)
	return f
}

func (f Field) Description() string {
	return f.description.getString()
}

func (f *Field) SetPosition(position interface{}) *Field {
	setValueHelper(&f.position, position)
	return f
}

func (f Field) Position() int {
	return f.position.getInt()
}

func (f *Field) SetRemove(remove interface{}) *Field {
	setValueHelper(&f.remove, remove)
	return f
}

func (f Field) IsRemoved() bool {
	return f.remove.getBool()
}

func (f *Field) SetHidden(hidden interface{}) *Field {
	setValueHelper(&f.hidden, hidden)
	return f
}

func (f Field) IsHidden() bool {
	return f.hidden.getBool()
}

func (f *Field) SetSortable(sortable bool) *Field {
	f.sortable = sortable
	return f
}

func (f Field) IsSortable() bool {
	return f.sortable
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

func (f *Field) Option(key string) interface{} {
	if v, ok := f.options[key]; ok {
		return v
	}
	return nil
}

func (f *Field) SetView(view interface{}) *Field {
	setValueHelper(&f.view, view)
	return f
}

func (f *Field) View() string {
	return f.view.getString()
}

func (f *Field) SetCallback(callback interface{}, args ...interface{}) *Field {
	f.callback = callback
	f.callbackArguments = args
	return f
}

func (f *Field) setError(err error) {
	f.error = err
}

func (f Field) Error() error {
	return f.error
}

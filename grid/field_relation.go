package grid

import (
	"fmt"
	"github.com/patrickascher/gofw/orm"
	"reflect"
)

// Interface of the fields.
type Interface interface {
	getTitle() string
	getDescription() string
	getPosition() int
	setPosition(interface{})
	setHide(interface{})
	getHide() bool
	getRemove() bool
	setRemove(interface{})
	getSort() bool
	getFilter() bool

	setReadOnly(bool)
	getReadOnly() bool

	//getFieldValidator() map[string]validator.Validation
	//for fields
	//getFieldPrimary() interface{}
	//getFieldDefault() interface{}
	//for relations
	getFieldType() FieldType
	setFieldType(FieldType)

	getFieldName() string
	setJsonName(string)
	getJsonName() string

	getColumn() *orm.Field
	getFields() map[string]Interface

	getCallback() reflect.Value
	getCallbackArgs() []interface{}
}

// setValueHelper is a helper for configure the fields
// It is taking a ptr to a value struct or a single go type.
// If only a go type is set, all modes will have the same value.
// With the value struct, the value for each mode can be set individual.
func setValueHelper(field *value, v interface{}) {
	tpe := reflect.TypeOf(v).String()
	switch tpe {
	case "*grid.value":
		field.setByValue(v.(*value))
		return
	default:
		field.set(v)
		return
	}
}

// --------------------------------------------------------------------------
// common declares function which are used in fields and relations
type common struct {
	title       *value
	description *value
	position    *value
	remove      *value
	hide        *value
	readOnly    bool
	//permission  *value TODO

	filter bool
	sort   bool

	fieldType FieldType

	column      *orm.Field
	jsonName    string
	association *orm.Relation

	callback     reflect.Value
	callbackArgs []interface{}
}

func (c *common) setJsonName(v string) {
	c.jsonName = v
}

func (c *common) getJsonName() string {
	return c.jsonName
}

func (c *common) getTitle() string {
	return c.title.getString()
}

func (c *common) setReadOnly(v bool) {
	c.readOnly = v
}

func (c *common) getReadOnly() bool {
	return c.readOnly
}

func (c *common) setTitle(v interface{}) {
	setValueHelper(c.title, v)
}

func (c *common) getDescription() string {
	return c.description.getString()
}

func (c *common) setDescription(v interface{}) {
	setValueHelper(c.description, v)
}

func (c *common) getPosition() int {
	return c.position.getInt()
}

func (c *common) setPosition(v interface{}) {
	setValueHelper(c.position, v)
}

func (c *common) getHide() bool {
	return c.hide.getBool()
}

func (c *common) setHide(v interface{}) {
	setValueHelper(c.hide, v)
}

func (c *common) getRemove() bool {
	return c.remove.getBool()
}

func (c *common) setRemove(v interface{}) {
	setValueHelper(c.remove, v)
}

func (c *common) getSort() bool {
	return c.sort
}

func (c *common) setSort(b bool) {
	c.sort = b
}

func (c *common) getFilter() bool {
	return c.filter
}

func (c *common) setFilter(b bool) {
	c.filter = b
}

func (c *common) getFieldType() FieldType {
	return c.fieldType
}

func (c *common) setFieldType(f FieldType) {
	c.fieldType = f
}

func (c *common) getColumn() *orm.Field {
	return c.column
}

func (c *common) getCallback() reflect.Value {
	return c.callback
}

func (c *common) getCallbackArgs() []interface{} {
	return c.callbackArgs
}

// defaultCommon is a helper for all the default values.
func defaultCommon(g *Grid) common {
	c := common{}
	c.title = valueWithGrid("", g)
	c.description = valueWithGrid("", g)
	c.remove = valueWithGrid(false, g)
	c.hide = valueWithGrid(false, g)
	c.position = valueWithGrid(0, g)
	c.filter = true
	c.sort = true

	c.fieldType = DefaultFieldType(g)

	return c
}

// --------------------------------------------------------------------------

type field struct {
	common
	//validator *validator.Validator //default from db column
}

// DefaultField is creting a new field struct with default values [see defaultCommon].
func defaultField(g *Grid) *field {
	f := field{common: defaultCommon(g)}
	return &f
}

func (f *field) FieldType() FieldType {
	return f.fieldType
}

func (f *field) SetSelect(sql string) *field {
	f.column.SqlSelect = sql
	return f
}

func (f *field) SetReadOnly(b bool) *field {
	f.readOnly = b
	return f
}

// SetTitle to the field.
func (f *field) SetTitle(title interface{}) *field {
	f.setTitle(title)
	return f
}

// SetDescription to the field.
func (f *field) SetDescription(desc interface{}) *field {
	f.setDescription(desc)
	return f
}

// SetPosition to the field.
func (f *field) SetPosition(pos interface{}) *field {
	f.setPosition(pos)
	return f
}

// SetHide to the field.
func (f *field) SetHide(hide interface{}) *field {
	f.setHide(hide)
	return f
}

// SetRemove to the field.
func (f *field) SetRemove(remove interface{}) *field {
	f.setRemove(remove)
	return f
}

// SetSort to the field.
func (f *field) SetSort(sort bool) *field {
	f.setSort(sort)
	return f
}

// SetFilter to the field.
func (f *field) SetFilter(filter bool) *field {
	f.setFilter(filter)
	return f
}

// getFields returns all the time nil on "field".
func (f *field) getFields() map[string]Interface {
	return nil
}

// getFieldName returns the struct field name
func (f *field) getFieldName() string {
	return f.getColumn().Name
}

func (f *field) SetCallback(fn interface{}, args ...interface{}) {
	v := reflect.ValueOf(fn)
	f.callback = v
	f.callbackArgs = args
}

// --------------------------------------------------------------------------

type relation struct {
	common
	//dbTable      string //needed?
	//relationType string //yes but over columns????
	name   string
	fields map[string]Interface

	decorator           string
	decoratorEscapeHTML bool

	skipRelation string
}

// DefaultRelation is creting a new field struct with default values [see defaultCommon].
func defaultRelation(g *Grid) *relation {
	r := relation{common: defaultCommon(g)}
	r.fields = make(map[string]Interface, 0)
	return &r
}

func (r *relation) FieldType() FieldType {
	return r.fieldType
}

func (r *relation) SetReadOnly(b bool) *relation {
	r.readOnly = b
	return r
}

func (r *relation) SetCallback(fn interface{}, args ...interface{}) {
	v := reflect.ValueOf(fn)
	r.callback = v
	r.callbackArgs = args
}

// SetTitle to the relation.
func (r *relation) SetTitle(title interface{}) *relation {
	r.setTitle(title)
	return r
}

// SetDescription to the relation.
func (r *relation) SetDescription(desc interface{}) *relation {
	r.setDescription(desc)
	return r
}

// SetPosition to the relation.
func (r *relation) SetPosition(pos interface{}) *relation {
	r.setPosition(pos)
	return r
}

// SetHide to the relation.
func (r *relation) SetHide(hide interface{}) *relation {
	r.setHide(hide)
	return r
}

// SetRemove to the relation.
func (r *relation) SetRemove(remove interface{}) *relation {
	r.setRemove(remove)
	return r
}

// SetSort to the relation.
func (r *relation) SetSort(sort bool) *relation {
	r.setSort(sort)
	return r
}

// SetFilter to the relation.
func (r *relation) SetFilter(filter bool) *relation {
	r.setFilter(filter)
	return r
}

// getFields returns all the fields of a relation.
func (r *relation) getFields() map[string]Interface {
	return r.fields
}

// getFieldName returns the struct field name
func (r *relation) getFieldName() string {
	return r.name
}

// setFieldName sets the field name
func (r *relation) setFieldName(name string) {
	r.name = name
}

// TODO at the moment no relation is sortable - create logic for it.
func (r *relation) getSort() bool {
	return false
}

// Relation is returning a relation by the given name. If the relation does not exist, a error will return.
func (r *relation) Relation(n string) (*relation, error) {
	if f, ok := r.fields[n]; ok {
		return f.(*relation), nil
	}
	return nil, fmt.Errorf(ErrFieldOrRelation.Error(), n)
}

// Field is returning a field by the given name. If the field does not exist, a error will return.
func (r *relation) Field(f string) (*field, error) {
	if f, ok := r.fields[f]; ok {
		return f.(*field), nil
	}
	return nil, fmt.Errorf(ErrFieldOrRelation.Error(), f)
}

// TODO better solution, delete this after.
// Select contains general information to build a HTML select
type Select struct {
	Options interface{} `json:"items,omitempty"`
	ValueK  string      `json:"valueKey,omitempty"`
	TextK   string      `json:"textKey,omitempty"`
}

func (s *Select) Items() interface{} {
	return s.Options
}

func (s *Select) SetItems(items interface{}) SelectI {
	s.Options = items
	return s
}

func (s *Select) ValueKey() string {
	return s.ValueK
}

func (s *Select) SetValueKey(k string) SelectI {
	s.ValueK = k
	return s
}

func (s *Select) TextKey() string {
	return s.TextK
}

func (s *Select) SetTextKey(k string) SelectI {
	s.TextK = k
	return s
}

type SelectI interface {
	Items() interface{}
	SetItems(interface{}) SelectI
	ValueKey() string
	SetValueKey(string) SelectI
	TextKey() string
	SetTextKey(string) SelectI
}

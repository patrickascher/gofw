package orm2

import (
	"fmt"
	"github.com/patrickascher/gofw/sqlquery"
	"reflect"
	"strings"
)

// NewScope returns a new scope from a model.
// This function should be used in custom strategies to access the model data.
// If a custom white/blacklist is defined, the permission of model fields and relations getting set.
// Error will return if there were some permission error.s
func NewScope(m *Model, action string) (Scope, error) {
	s := Scope{model: m}

	// whitelist/blacklist
	err := setFieldPermission(s, action)
	if err != nil {
		return Scope{}, err
	}

	return s, nil
}

// NewScopeFromInterface returns a scope from an orm.interface.
// This can be useful to access internal model data.
func NewScopeFromInterface(m Interface) Scope {
	return Scope{model: m.model()}
}

// Scope is a helper function for the loading strategies or callbacks.
// Its used to access the internal model fields or relations.
// Some basic helper functions are defined.
type Scope struct {
	model *Model
}

// Builder returns the caller Builder{}.
func (scope Scope) Builder() sqlquery.Builder {
	return scope.model.DefaultBuilder()
}

// Model returns the scope model.
func (scope Scope) Model() *Model {
	return scope.model
}

// Caller returns the model caller.
func (scope Scope) Caller() Interface {
	return scope.model.caller
}

// TableName returns the full qualified table name.
// TODO schemas
func (scope Scope) TableName() string {
	return scope.Builder().Config().Database + "." + scope.Caller().DefaultTableName()
}

// Columns return all columns as string by the given permission.
// The permission can be read:true or write:true for selects or exec query statements.
// SqlSelect indicates if the real column name should be returned or the modified sql select.
func (scope Scope) Columns(p Permission, SqlSelect bool) []string {
	var rv []string
	for _, f := range scope.Fields(p) {

		// custom sql selects
		if SqlSelect && f.SqlSelect != "" {
			rv = append(rv, sqlquery.Raw("("+f.SqlSelect+") AS "+scope.Builder().QuoteIdentifier(f.Information.Name)))
			continue
		}

		rv = append(rv, f.Information.Name)
	}
	return rv
}

// PrimaryKeysFieldName return all primary keys as string
func (scope Scope) PrimaryKeysFieldName() []string {
	var rv []string
	for _, f := range scope.model.fields {
		if f.Information.PrimaryKey {
			rv = append(rv, f.Name)
		}
	}
	return rv
}

// PrimaryKeysFieldName return all primary keys as Field{} struct.
func (scope Scope) PrimaryKeys() []Field {
	var rv []Field
	for _, f := range scope.model.fields {
		if f.Information.PrimaryKey {
			rv = append(rv, f)
		}
	}
	return rv
}

// CallerField return a reflect.Value of the caller models given field.
func (scope Scope) CallerField(field string) reflect.Value {
	return reflect.ValueOf(scope.model.caller).Elem().FieldByName(field)
}

// Field returns the field by the given name.
// Error will return if it does not exist.
func (scope Scope) Field(name string) (Field, error) {

	for _, f := range scope.model.fields {
		if f.Name == name {
			return f, nil
		}
	}
	return Field{}, fmt.Errorf(errStructField.Error(), name, scope.model.name)
}

// Fields return all Fields by the given permission.
// Custom fields are skipped. //TODO allow?
// The permission can be read:true or write:true for selects or exec query statements.
func (scope Scope) Fields(p Permission) []Field {
	var rv []Field
	for _, f := range scope.model.fields {

		// skipping custom types
		// TODO custom Relations have only callbacks?
		if f.Custom {
			continue
		}

		// skip if permission is not permitted
		if (p.Read && !f.Permission.Read) || (p.Write && !f.Permission.Write) {
			continue
		}

		rv = append(rv, f)
	}
	return rv
}

// Relation returns a relation by name.
// Permission is ignored at the moment.
// Custom Relations are also skipped? TODO: allow?
func (scope Scope) Relation(relation string, p Permission) Relation {
	for _, r := range scope.model.relations {
		if r.Field == relation {
			if r.Custom {
				continue
			}
			return r
		}
	}
	return Relation{}
}

// Relations return all defined relation in this scope.
// The permission can be read:true or write:true for selects or exec query statements.
func (scope Scope) Relations(p Permission) []Relation {
	var rv []Relation
	for _, r := range scope.model.relations {
		if r.Custom {
			continue
		}
		// skip if permission is not permitted
		if (p.Read && !r.Permission.Read) || (p.Write && !r.Permission.Write) {
			continue
		}

		rv = append(rv, r)
	}
	return rv
}

// IsPolymorphic returns true if a polymorphic is defined on that relation
func (scope Scope) IsPolymorphic(relation Relation) bool {
	if relation.Polymorphic.Value != "" {
		return true
	}
	return false
}

// CachedModel returns a scope to the cached model.
// Error will return if the cache model was not found.
func (scope Scope) CachedModel(model string) (Scope, error) {

	c, _, err := scope.Caller().Cache()
	if err != nil {
		return Scope{}, err
	}

	v, err := c.Get(model)
	if err != nil {
		return Scope{}, err
	}
	m := v.Value().(Model)

	return Scope{&m}, err
}

// InitCallerRelation returns an orm.Interface of the given relation Field.
// If the caller field is an Ptr or struct the reference will be taken and initialized.
// If the caller field is a slice the struct will be returned and initialized.
// The new relation model will be initialized with the parent cache and the white/blacklist for that relation.
// Error will return on setCache or Init.
func (scope Scope) InitCallerRelation(relField string) (Interface, error) {

	f := scope.CallerField(relField)
	r := scope.Relation(relField, Permission{})
	var relation Interface
	var err error

	switch f.Kind() {
	case reflect.Ptr:
		if reflect.ValueOf(f.Interface().(Interface)).IsNil() {
			f.Set(newValueInstanceFromType(scope.CallerField(relField).Type()).Addr())
		}
		relation = f.Interface().(Interface)
	case reflect.Struct:
		relation = f.Addr().Interface().(Interface)
	case reflect.Slice:
		//field, _ := reflect.TypeOf(reflect.ValueOf(caller).Elem().Interface()).FieldByName(field)
		// relation = newValueInstanceFromType(field.Type).Addr().Interface().(Interface)
		relation = newValueInstanceFromType(r.Type).Addr().Interface().(Interface)
	}

	if relation != nil {
		err = scope.addParentCache(relation)
		if err != nil {
			return nil, err
		}
		err = relation.Init(relation)
		if err != nil {
			return nil, err
		}
		scope.addParentWbList(relation, relField)
	}

	return relation, err
}

// addParentCache passes the cache from the parent model to the child model.
func (scope Scope) addParentCache(relation Interface) error {
	c, d, err := scope.Caller().Cache()
	if err != nil {
		return err
	}
	err = relation.SetCache(c, d)
	if err != nil {
		return err
	}

	return nil
}

// addParentWbList passes the wb list from the parent to the child model if a dot notation was uses.
func (scope Scope) addParentWbList(relation Interface, field string) {

	if scope.Model().wbList == nil {
		return
	}

	var fields []string
	for _, a := range scope.Model().wbList.fields {
		if strings.HasPrefix(a, field+".") {
			fields = append(fields, strings.Replace(a, field+".", "", 1))
		}
	}

	if len(fields) > 0 {
		relation.SetWBList(scope.Model().wbList.policy, fields...)
	}
}

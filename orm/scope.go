package orm

import (
	"database/sql"
	driverI "database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/patrickascher/gofw/sqlquery"
)

var (
	errParentModel  = "orm: parent model %s was not found"
	errInstanceType = "orm: the scope %s and the given type %s must be of the same type"
	errInfinityLoop = errors.New("orm: 🎉 congratulation you created an infinity loop")
	errModelNil     = "orm: relation orm %s is nil"
)

// Scope is a helper function for the loading strategies or callbacks.
// Its used to access the internal model fields or relations.
// Some basic helper functions are defined.
type Scope struct {
	model *Model
}

// Builder returns a ptr to the model builder.
func (scope Scope) Builder() *sqlquery.Builder {
	return &scope.model.builder
}

// Model returns a ptr to the orm model.
func (scope Scope) Model() *Model {
	return scope.model
}

// Parent returns the parent model by name or the root parent if the name is empty.
// The name must be the orm struct name incl. namespace.
// Error will return if no parent exists or the given name does not exist.
func (scope Scope) Parent(name string) (*Model, error) {
	p := scope.model.parentModel
	for p != nil {
		// return root parent
		if name == "" && p.parentModel == nil {
			return p, nil
		}
		// return named parent
		if p.name == name {
			return p, nil
		}
		p = p.parentModel
	}

	return nil, fmt.Errorf(errParentModel, name)
}

// SetBackReference sets a back reference if the model was already loaded.
// This will avoid loops.
// At the moment only on hasOne and belongsTo relations possible.
//
// TODO: create back referencing for m2m.
func (scope Scope) SetBackReference(rel Relation) error {
	c, err := scope.Parent(rel.Type.String())
	if err != nil {
		return err
	}

	f := scope.CallerField(rel.Field)
	return SetReflectValue(f, reflect.ValueOf(c.caller))
}

// SetReflectValue is a helper to set a fields value without worrying about the field type.
// The field type and the value type must be the same with the exception of int.
// Int32,int64 and nullInt will be mapped.
// TODO create a better solution for this, what with int8,int16,...
func SetReflectValue(field reflect.Value, value reflect.Value) error {
	switch field.Kind() {
	case reflect.Ptr:
		if value.CanAddr() {
			field.Set(value.Addr())
		} else {
			field.Set(value)
		}
	case reflect.Struct:
		scannerI := reflect.TypeOf((*sql.Scanner)(nil)).Elem()
		if field.Addr().Type().Implements(scannerI) && field.Type() != value.Type() {
			f := field.Addr().Interface().(sql.Scanner)
			_ = f.Scan(value)
		} else {
			field.Set(value)
		}
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.Ptr {
			field.Set(reflect.Append(field, reflect.Indirect(value).Addr()))
		} else {
			field.Set(reflect.Append(field, value))
		}
	default:
		if field.Type() != value.Type() {
			if v, ok := value.Interface().(driverI.Valuer); ok {
				vv, err := v.Value()
				if err != nil {
					return err
				}
				value = reflect.ValueOf(vv)
			}
		}
		// int mapping, create a better solution
		if field.Kind() == reflect.Int && value.Kind() == reflect.Int64 {
			value = reflect.ValueOf(int(value.Interface().(int64)))
		}
		if field.Kind() == reflect.Int64 && value.Kind() == reflect.Int {
			value = reflect.ValueOf(value.Int())
		}

		field.Set(value)
	}
	return nil
}

// Caller returns the orm caller.
func (scope Scope) Caller() Interface {
	return scope.model.caller
}

// Name returns the orm model name.
// Namespace will be included if 'ns' is true.
// If no namespace is required the orm model name will be titled.
func (scope Scope) Name(ns bool) string {
	return scope.model.modelName(ns)
}

// TableName returns the full qualified table name.
// TODO schema is missing
func (scope Scope) TableName() string {
	return scope.Builder().Config().Database + "." + scope.Caller().DefaultTableName()
}

// Columns return all columns as string by the given permission.
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

// PrimaryKeys return all primary keys as []Field.
func (scope Scope) PrimaryKeys() []Field {
	var rv []Field
	for _, f := range scope.model.fields {
		if f.Information.PrimaryKey {
			rv = append(rv, f)
		}
	}
	return rv
}

// PrimaryKeysFieldName return all primary keys by the field name as string.
func (scope Scope) PrimaryKeysFieldName() []string {
	var rv []string
	for _, f := range scope.PrimaryKeys() {
		if f.Information.PrimaryKey {
			rv = append(rv, f.Name)
		}
	}
	return rv
}

// PrimariesSet checks if all primaries have a non zero value.
func (scope Scope) PrimariesSet() bool {
	for _, f := range scope.model.fields {
		if f.Information.PrimaryKey {
			if scope.CallerField(f.Name).IsZero() {
				return false
			}
		}
	}
	return true
}

// IsEmpty checks if the orm model fields and relations are empty by the given permission.
func (scope Scope) IsEmpty(perm Permission) bool {
	for _, f := range scope.Fields(perm) {
		if !scope.CallerField(f.Name).IsZero() {
			if f.Name == CreatedAt || f.Name == UpdatedAt || f.Name == DeletedAt {
				continue
			}
			return false
		}
	}

	for _, r := range scope.Relations(perm) {
		// TODO simplify this whole code.

		// for slice and ptr
		if scope.CallerField(r.Field).IsZero() ||
			(scope.CallerField(r.Field).Type().Kind() == reflect.Slice && scope.CallerField(r.Field).Len() == 0) {
			continue
		}
		// must be initialized
		if scope.CallerField(r.Field).Type().Kind() == reflect.Struct {
			_ = scope.InitRelation(scope.CallerField(r.Field).Addr().Interface().(Interface), r.Field)
			if !scope.CallerField(r.Field).Addr().Interface().(Interface).Scope().IsEmpty(perm) {
				return false
			}
		}
		if scope.CallerField(r.Field).Type().Kind() == reflect.Ptr {
			_ = scope.InitRelation(scope.CallerField(r.Field).Interface().(Interface), r.Field)
			if !scope.CallerField(r.Field).Interface().(Interface).Scope().IsEmpty(perm) {
				return false
			}
		}
	}
	return true
}

// CallerField return a reflect.Value of the orm caller struct field.
func (scope Scope) CallerField(field string) reflect.Value {
	return reflect.ValueOf(scope.model.caller).Elem().FieldByName(field)
}

// Field by the given name.
// No specific permission is checked. No need for it yet.
// Error will return if it does not exist.
func (scope Scope) Field(name string) (Field, error) {

	for _, f := range scope.model.fields {
		if f.Name == name {
			return f, nil
		}
	}
	return Field{}, fmt.Errorf(errStructField, name, scope.model.name)
}

// Fields return all Fields by the given permission.
// Custom fields are skipped.
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

// Relation by name.
// Error will return if the relation does not exist.
func (scope Scope) Relation(relation string, p Permission) (Relation, error) {
	for _, r := range scope.Relations(p) {
		if r.Field == relation {
			return r, nil
		}
	}
	return Relation{}, fmt.Errorf(errStructField, relation, scope.model.name)
}

// Relations return all relations of the orm by the given permission.
// Custom relations are skipped.
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

// IsPolymorphic checks if the given relation is polymorphic.
func (scope Scope) IsPolymorphic(relation Relation) bool {
	if relation.Polymorphic.Value != "" {
		return true
	}
	return false
}

// CachedModel returns a scope to the cached model.
// Error will return if the cache model was not found.
func (scope Scope) CachedModel(model string) (Interface, error) {
	if scope.model.cache == nil {
		return nil, errNoCache
	}

	v, err := scope.model.cache.Get(model)
	if err != nil {
		return nil, err
	}
	m := v.Value().(Model)

	return &m, err
}

// ScanValues, scans a row set into the orm fields.
func (scope Scope) ScanValues(p Permission) []interface{} {
	var values []interface{}
	for _, col := range scope.Fields(p) {
		values = append(values, scope.CallerField(col.Name).Addr().Interface())
	}
	return values
}

// NewScopeFromType is used to create a new orm instance of the given type.
// The given type must be the same type as the scope where its getting called because
// the fields/relations are getting copied.
//
// The cache will be passed from the scope orm. The fields and relations are
// copied from the active scope orm instance, because there could have been some permission changes.
// Only used in eager.all.
func (scope Scope) NewScopeFromType(p reflect.Type) (*Scope, error) {
	v := newValueInstanceFromType(p)
	model := v.Addr().Interface().(Interface)

	// check if the scope and given type are the same.
	if scope.Name(true) != strings.Replace(p.String(), "*", "", -1) {
		//return nil, fmt.Errorf(errInstanceType, scope.Name(true), p.String())
	}

	// copy the scope cache to the new orm instance.
	model.model().cache, model.model().cacheTTL = scope.model.cache, scope.model.cacheTTL
	// init the orm instance.
	err := model.Init(model)
	if err != nil {
		return nil, err
	}

	// copy fields/relation permission from parent.
	if scope.Name(true) == strings.Replace(p.String(), "*", "", -1) {
		copy(model.model().fields, scope.model.fields)
		copy(model.model().relations, scope.model.relations)
	}
	model.model().loopDetection = scope.model.loopDetection

	return model.Scope(), nil
}

// EqualWith checks if the given orm model is equal with the scope orm model.
// A []ChangedValue will return with all the changes recursively (fields and relations).
// On relations and slices the operation info (create, update or delete) is given.
// All time fields are excluded of this check.
// On hasMany or m2m relations on DELETE operation the index will be the Field "ID".
// TODO: check for the real primary field(s) and set the correct index.
// TODO: id on DELETE no Primary is given, 0 will be set as index. Throw error?
func (scope Scope) EqualWith(snapshot Interface) ([]ChangedValue, error) {

	var cv []ChangedValue
	writePerm := Permission{Write: true}

	// normal fields
	for _, f := range scope.Fields(writePerm) {
		// skip the automatic time fields
		if f.Name == CreatedAt || f.Name == UpdatedAt || f.Name == DeletedAt {
			continue
		}

		oldV := snapshot.Scope().CallerField(f.Name).Interface()
		newV := scope.CallerField(f.Name).Interface()
		if oldV != newV {
			cv = append(cv, ChangedValue{Operation: UPDATE, Field: f.Name, OldV: oldV, NewV: newV})
		}
	}
	// if there were any changes on the normal fields, the UpdatedAt field gets set as changed field.
	if len(cv) > 0 {
		cv = append(cv, ChangedValue{Operation: UPDATE, Field: UpdatedAt})
	}

	// relations fields
	for _, rel := range scope.Relations(writePerm) {

		switch rel.Kind {
		case HasOne, BelongsTo:
			// relation interface
			relationI, err := scope.InitCallerRelation(rel.Field, false)
			if err != nil {
				return nil, err
			}
			// relation snapshot interface
			relationSnapshotI, err := snapshot.Scope().InitCallerRelation(rel.Field, false)
			if err != nil {
				return nil, err
			}

			// check if the relation is equal with the relation snapshot
			changes, err := relationI.Scope().EqualWith(relationSnapshotI)
			if err != nil {
				return nil, err
			}

			// if there were any changes
			if len(changes) > 0 {
				op := UPDATE
				if relationI.Scope() != relationSnapshotI.Scope() {

					// if the relation model is empty, delete all existing entries.
					if relationI.Scope().IsEmpty(Permission{}) {
						op = DELETE
					} else if relationSnapshotI.Scope().IsEmpty(Permission{}) {
						// if the relation snapshot was empty, create all entries.
						op = CREATE
					} else if !relationI.Scope().PrimariesSet() || Int(relationI.Scope().CallerField("ID").Interface()) != Int(relationSnapshotI.Scope().CallerField("ID").Interface()) {
						// if there were entries before but the new added relation has no primary key set or has an new ID.
						// this can happens if the user adds manually a new slice.
						// the old relation snapshot IDs will be deleted at the end.
						// TODO BUG(patrick): just checking against the ID field is unsafe because the user could have defined his own primary key, a function to compare the primary keys must be created.
						op = CREATE
					}
				}
				cv = append(cv, ChangedValue{Operation: op, Field: rel.Field, ChangedValue: changes})
			}
		case HasMany, ManyToMany:

			newLength := scope.CallerField(rel.Field).Len()
			oldLength := snapshot.Scope().CallerField(rel.Field).Len()

			// no entries exist
			if newLength == 0 && oldLength == 0 {
				continue
			}

			op := UPDATE
			// if there are no entries in the relation snapshot.
			if oldLength == 0 {
				cv = append(cv, ChangedValue{Operation: CREATE, Field: rel.Field})
				continue
			}
			// if there are no entries in the relation.
			if newLength == 0 {
				cv = append(cv, ChangedValue{Operation: DELETE, Field: rel.Field})
				continue
			}

			var changes []ChangedValue
		newSliceLoop:
			// iterating over the new entries
			for i := 0; i < newLength; i++ {
				// slice interface
				sliceI := reflect.Indirect(scope.CallerField(rel.Field).Index(i)).Addr().Interface().(Interface)
				err := scope.InitRelation(sliceI, rel.Field)
				if err != nil {
					return nil, err
				}

				// new entry - if primary keys are not set
				if !sliceI.Scope().PrimariesSet() {
					changes = append(changes, ChangedValue{Operation: CREATE, Index: i, Field: rel.Field})
				} else {

					// iterating over the relation snapshot
					for n := 0; n < snapshot.Scope().CallerField(rel.Field).Len(); n++ {
						// slice snapshot interface
						sliceSnapshotI := reflect.Indirect(snapshot.Scope().CallerField(rel.Field).Index(n)).Addr().Interface().(Interface)
						err := scope.InitRelation(sliceSnapshotI, rel.Field)
						if err != nil {
							return nil, err
						}

						// TODO BUG(patrick): if the primary is a string (uuid)
						// TODO befere there was a orm.Int() function, now its comparing directly. will throw an error on int & int64.
						if sliceSnapshotI.Scope().CallerField("ID").Interface() == sliceI.Scope().CallerField("ID").Interface() {

							changesSlice, err := sliceI.Scope().EqualWith(sliceSnapshotI)
							if err != nil {
								return nil, err
							}
							if len(changesSlice) > 0 {
								changes = append(changes, ChangedValue{Operation: UPDATE, Index: i, Field: rel.Field, ChangedValue: changesSlice})
							}

							// if there were no changes, delete from snapshot slice. because all existing snapshot slices will get delete at the end.
							result := reflect.AppendSlice(snapshot.Scope().CallerField(rel.Field).Slice(0, n), snapshot.Scope().CallerField(rel.Field).Slice(n+1, snapshot.Scope().CallerField(rel.Field).Len()))
							snapshot.Scope().CallerField(rel.Field).Set(result)
							continue newSliceLoop
						}
					}
					// if the slice was not found in the snapshot slice, create it.
					changes = append(changes, ChangedValue{Operation: CREATE, Index: i, Field: rel.Field})
				}
			}

			// all still existing snapshot slices, will get deleted. because they are represented in the new relation slice.
			if snapshot.Scope().CallerField(rel.Field).Len() > 0 {
				for n := 0; n < snapshot.Scope().CallerField(rel.Field).Len(); n++ {
					changes = append(changes, ChangedValue{Operation: DELETE, Index: Int(reflect.Indirect(snapshot.Scope().CallerField(rel.Field).Index(n)).FieldByName("ID").Interface()), Field: rel.Field})
				}
			}

			// if there ware any changes, add it.
			if len(changes) > 0 {
				cv = append(cv, ChangedValue{Operation: op, Field: rel.Field, ChangedValue: changes})
			}
		}
	}

	return cv, nil
}

// TimeFields return the existing time fields of the database by the given permission.
func (scope Scope) TimeFields(p Permission) []string {
	var rv []string
	for _, f := range scope.Fields(p) {
		if f.Information.Name == "created_at" || f.Information.Name == "updated_at" || f.Information.Name == "deleted_at" {
			rv = append(rv, f.Name)
		}
	}
	return rv
}

// InitCallerRelation returns an orm.Interface of the given relation Field.
// If the caller field is an Ptr or struct the reference will be taken and initialized.
// If the caller field is a slice the struct will be returned and initialized.
// The new relation model will be initialized with the parent cache, white/blacklist, loopDetection and builder.
func (scope Scope) InitCallerRelation(relField string, noParent bool) (Interface, error) {

	f := scope.CallerField(relField)
	r, err := scope.Relation(relField, Permission{})
	if err != nil {
		return nil, err
	}

	// get the relation field
	var relationI Interface
	switch f.Kind() {
	case reflect.Ptr:
		if reflect.ValueOf(f.Interface().(Interface)).IsNil() {
			f.Set(newValueInstanceFromType(scope.CallerField(relField).Type()).Addr())
		}
		relationI = f.Interface().(Interface)
	case reflect.Struct:
		relationI = f.Addr().Interface().(Interface)
	case reflect.Slice:
		relationI = newValueInstanceFromType(r.Type).Addr().Interface().(Interface)
	}

	// initialize the relation field
	err = scope.InitRelation(relationI, relField)
	if err != nil {
		return nil, err
	}

	if noParent {
		relationI.model().parentModel = nil
	}

	return relationI, err
}

// InitRelation initializes the given relation.
// * If the relation is already initialized, the caller gets set. Otherwise the cache will be set and the orm model gets initialized.
// * If the argument relationField is not empty, its a child orm model and the parent will be set.
// * parent wb list, loopDetection and the builder will be passed to the child of the scope orm model.
func (scope Scope) InitRelation(relationI Interface, relationField string) error {

	// relationI is nil - ptr reference
	if reflect.ValueOf(relationI).IsNil() {
		return fmt.Errorf(errModelNil, reflect.TypeOf(relationI).String())
	}

	// initialize model or only replace the caller if its already initialized.
	if !relationI.model().isInitialized {
		relationI.model().cache, relationI.model().cacheTTL = scope.model.cache, scope.model.cacheTTL
		err := relationI.Init(relationI)
		if err != nil {
			return err
		}
	} else {
		relationI.model().caller = relationI                       //needed
		relationI.model().scope = &Scope{model: relationI.model()} //needed
	}

	// if no string is given, its no relations - its the root element which should have no parent.
	if relationField != "" {
		relationI.model().parentModel = scope.model.caller.model() // set the correct parent
	}

	// passing parent fields
	scope.addParentWbList(relationI, relationField)

	// copy the loopDetection, map is referenced by. so all the changes would also be in the parent models.
	relationI.model().loopDetection = make(map[string][]string, len(scope.model.loopDetection))
	for index := range scope.model.loopDetection {
		relationI.model().loopDetection[index] = make([]string, len(scope.model.loopDetection[index]))
		copy(relationI.model().loopDetection[index], scope.model.loopDetection[index])
	}

	relationI.model().builder = scope.model.builder

	return nil
}

// checkLoopMap is checking if the relation model was already asked before with the same where condition.
// Error will return if the same condition was already asked before.
func (scope Scope) checkLoopMap(args string) error {
	rel := scope.Name(true)
	counter := 0
	for _, b := range scope.model.loopDetection[rel] {
		if b == args {
			counter++
		}
		if counter >= 1 {
			return errInfinityLoop
		}
	}
	if scope.model.loopDetection == nil {
		scope.model.loopDetection = map[string][]string{}
	}

	scope.model.loopDetection[rel] = append(scope.model.loopDetection[rel], args)
	return nil
}

// addParentWbList passes the wb list from the parent to the child orm model, if a dot notation exists.
// If the field is empty the root wb list is copied. This is for example needed by new slice orm models.
// For example Car.ID: the field ID will be added to the child car orm model.
func (scope Scope) addParentWbList(relation Interface, field string) {

	if scope.model.wbList == nil {
		return
	}

	// on self reference, add the same wb list of parent
	// also used for snapshot
	if scope.model.name == relation.model().modelName(true) {
		relation.SetWBList(scope.model.wbList.policy, scope.model.wbList.fields...)
		return
	}

	// otherwise add only if the relation is defined
	var fields []string
	for _, a := range scope.model.wbList.fields {
		if strings.HasPrefix(a, field+".") {
			fields = append(fields, strings.Replace(a, field+".", "", 1))
		}
	}

	if len(fields) > 0 {
		relation.SetWBList(scope.model.wbList.policy, fields...)
	}
}

// ChangedValue keeps recursively information of changed values.
type ChangedValue struct {
	Field        string
	OldV         interface{}
	NewV         interface{}
	Operation    string
	Index        int
	ChangedValue []ChangedValue
}

// AppendChangedValue adds the changedValue if it does not exist yet by the given field name.
func (scope Scope) AppendChangedValue(cV ChangedValue) {
	if scope.ChangedValueByFieldName(cV.Field) == nil {
		scope.model.changedValues = append(scope.model.changedValues, cV)
	}
}

// SetChangedValues sets the changedValues field of the scope.
// This is used to pass the values to a child orm model.
func (scope Scope) SetChangedValues(cV []ChangedValue) {
	scope.model.changedValues = cV
}

// ChangedValueByName returns a *changedValue by the field name.
// Nil will return if it does not exist.
func (scope Scope) ChangedValueByFieldName(field string) *ChangedValue {
	for _, c := range scope.model.changedValues {
		if c.Field == field {
			return &c
		}
	}
	return nil
}

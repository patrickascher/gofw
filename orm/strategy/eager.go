package strategy

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"

	"github.com/patrickascher/gofw/orm"
	"github.com/patrickascher/gofw/slices"
	"github.com/patrickascher/gofw/sqlquery"
)

func init() {
	_ = orm.Register("eager", &EagerLoading{})
}

type EagerLoading struct {
}

// First requests one row by the given condition.
// Relation hasOne and belongsTo will call orm First().
// Relation hasMany and manyToMany will call orm All().
// TODO: Custom relation conditions have to be improved, atm only depth 1.
func (e EagerLoading) First(scope *orm.Scope, c *sqlquery.Condition, perm orm.Permission) error {

	b := scope.Builder()

	// build select
	row, err := b.Select(scope.TableName()).Columns(scope.Columns(perm, true)...).Condition(c).First()
	if err != nil {
		return err
	}

	// scan all variables to fill it with values
	err = row.Scan(scope.ScanValues(perm)...)
	if err != nil {
		return err
	}

	// relations
	for _, relation := range scope.Relations(perm) {

		// back-referencing if the struct was already loaded, to avoid loops
		if err := scope.SetBackReference(relation); err == nil {
			return nil
		}

		// init relation
		c := &sqlquery.Condition{}
		rel, err := scope.InitCallerRelation(relation.Field, false)
		if err != nil {
			return err
		}

		// handling relation
		switch relation.Kind {
		case orm.HasOne, orm.BelongsTo:
			addWhere(scope, relation, c, scope.CallerField(relation.ForeignKey.Name).Interface())

			// CUSTOM Condition on relations, TODO improvements
			if v := scope.Model().RelationCondition(relation.Field); v != nil {
				c = v
			}

			err = rel.First(c)
			if err != nil {
				// reset initialized model to zero value.
				if err == sql.ErrNoRows {
					scope.CallerField(relation.Field).Set(reflect.New(scope.CallerField(relation.Field).Type()).Elem())
					continue
				}
				return err
			}
		case orm.HasMany, orm.ManyToMany:
			if relation.Kind == orm.HasMany {
				addWhere(scope, relation, c, scope.CallerField(relation.ForeignKey.Name).Interface())
			} else {
				// TODO INNER JOIN faster?
				c.Where(b.QuoteIdentifier(relation.AssociationForeignKey.Information.Name)+" IN (SELECT "+b.QuoteIdentifier(relation.JoinTable.AssociationForeignKey)+" FROM "+b.QuoteIdentifier(relation.JoinTable.Name)+" WHERE "+b.QuoteIdentifier(relation.JoinTable.ForeignKey)+" = ?)", scope.CallerField(relation.ForeignKey.Name).Interface())
			}
			c.Order(relation.AssociationForeignKey.Information.Name) // TODO make it possible to create a order by tag or something

			// CUSTOM Condition on relations, TODO improvements
			if v := scope.Model().RelationCondition(relation.Field); v != nil {
				c = v
			}

			// reset the slice.
			// needed if there is something like append slice -> update -> first (would double the slices)
			if scope.CallerField(relation.Field).Type().Elem().Kind() == reflect.Ptr {
				scope.CallerField(relation.Field).Set(reflect.New(reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(reflect.New(relation.Type).Interface())), 0, 0).Type()).Elem())
			} else {
				scope.CallerField(relation.Field).Set(reflect.New(reflect.MakeSlice(reflect.SliceOf(relation.Type), 0, 0).Type()).Elem())
			}

			err = rel.All(scope.CallerField(relation.Field).Addr().Interface(), c)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// All request all rows by the given condition.
// All foreign keys are collected after the main select, all relations are handled by one request (m2m has 2 selects one for the join table and one for the main data).
// The data is mapped after that.
// TODO back-references end in an infinity loop error at the moment
// TODO better solution of defer in a for loop, func?
// TODO defer in loop?
func (e EagerLoading) All(res interface{}, scope *orm.Scope, c *sqlquery.Condition) error {

	readPerm := orm.Permission{Read: true}
	b := scope.Builder()

	// build select
	rows, err := b.Select(scope.TableName()).Columns(scope.Columns(readPerm, true)...).Condition(c).All()
	if err != nil {
		return err
	}
	defer rows.Close()

	// result slice of the user interface
	resultSlice := reflect.ValueOf(res).Elem()

	// scan db results into orm model
	for rows.Next() {
		// new instance of slice element
		cScope, err := scope.NewScopeFromType(resultSlice.Type().Elem())
		if err != nil {
			return err
		}
		//add the values
		err = rows.Scan(cScope.ScanValues(readPerm)...)
		if err != nil {
			return err
		}
		// adding ptr or value depending on users struct definition
		err = orm.SetReflectValue(resultSlice, reflect.ValueOf(cScope.Caller()).Elem())
		if err != nil {
			return err
		}
	}

	// no rows were found
	if resultSlice.Len() == 0 {
		return nil
	}

	// in - collects all foreign keys of the main model. This is needed to create a select for all needed data.
	// Later on, the data will be mapped to the correct result again.
	in := map[string][]interface{}{}
	for _, relation := range scope.Relations(readPerm) {
		f := relation.ForeignKey.Name
		if _, ok := in[f]; !ok {
			for n := 0; n < resultSlice.Len(); n++ {
				i, err := orm.SanitizeValue(reflect.Indirect(resultSlice.Index(n)).FieldByName(f).Interface())
				if err != nil {
					return err
				}
				if _, exist := slices.ExistInterface(in[f], i); !exist {
					in[f] = append(in[f], i)
				}
			}
		}
		// Special case, if its a many to many relation.
		// To avoid multiple sql selects, all many2many results of this relation and the main model ids will be loaded with two queries.
		// The first query checks the join table for all needed IDs. The second query requests the model by the ids.
		m2mMap := map[string][]interface{}{}
		var m2mAll []interface{}
		if relation.Kind == orm.ManyToMany {

			var fkInt []interface{}
			for row := 0; row < resultSlice.Len(); row++ {
				v, err := orm.SanitizeValue(reflect.Indirect(resultSlice.Index(row)).FieldByName(relation.ForeignKey.Name).Interface())
				if err != nil {
					return err
				}
				fkInt = append(fkInt, v)
			}
			rows, err := b.Select(relation.JoinTable.Name).Columns(relation.JoinTable.ForeignKey, relation.JoinTable.AssociationForeignKey).Where(b.QuoteIdentifier(relation.JoinTable.ForeignKey)+" IN (?)", fkInt).All()
			if err != nil {
				return err
			}
			defer rows.Close() // TODO better solution for defer, func?

			// map fk,afk
			for rows.Next() {
				var fk string
				var afk string

				err = rows.Scan(&fk, &afk)
				if err != nil {
					return err
				}
				m2mMap[fk] = append(m2mMap[fk], afk)
				if _, exists := slices.ExistInterface(m2mAll, afk); !exists {
					m2mAll = append(m2mAll, afk)
				}
			}
		}

		// loading the requested model data
		if (relation.Kind != orm.ManyToMany && len(in[relation.ForeignKey.Name]) > 0) ||
			(relation.Kind == orm.ManyToMany && len(m2mAll) > 0) {

			// create condition
			c := &sqlquery.Condition{}
			if relation.Kind != orm.ManyToMany {
				addWhere(scope, relation, c, in[f])
			} else {
				c.Where(b.QuoteIdentifier(relation.ForeignKey.Information.Name)+" IN (?)", m2mAll)
			}

			// create an empty slice of the relation type
			rRes := reflect.New(reflect.MakeSlice(reflect.SliceOf(relation.Type), 0, 0).Type()).Interface()
			// create relation model
			rModel, err := scope.InitCallerRelation(relation.Field, false)
			if err != nil {
				return err
			}

			// CUSTOM Condition on relations, TODO improvements
			if v := scope.Model().RelationCondition(relation.Field); v != nil {
				c = v
			}

			// request all relation data
			err = rModel.All(rRes, c)
			if err != nil {
				return err
			}

			// map data to parent model
			rResElem := reflect.ValueOf(rRes).Elem()
			for row := 0; row < resultSlice.Len(); row++ { // parent model result
				for y := 0; y < rResElem.Len(); y++ { // relation result set
					// PARENT ID
					modelField := reflect.Indirect(resultSlice.Index(row)).FieldByName(relation.ForeignKey.Name)
					parentID, err := orm.SanitizeToString(modelField.Interface())
					if err != nil {
						return err
					}

					// set data to the main model.
					if relation.Kind == orm.ManyToMany {

						v, ok := m2mMap[reflect.ValueOf(parentID).String()]
						v2, err := orm.SanitizeToString(reflect.Indirect(rResElem.Index(y)).FieldByName(relation.AssociationForeignKey.Name).Interface())
						if err != nil {
							return err
						}

						_, exists := slices.ExistInterface(v, v2)
						if ok && exists {
							err = orm.SetReflectValue(reflect.Indirect(resultSlice.Index(row)).FieldByName(relation.Field), rResElem.Index(y))
							if err != nil {
								return err
							}
						}
					} else {
						if (!scope.IsPolymorphic(relation) && compareValues(parentID, reflect.Indirect(rResElem.Index(y)).FieldByName(relation.AssociationForeignKey.Name).Interface())) ||
							scope.IsPolymorphic(relation) && compareValues(parentID, reflect.Indirect(rResElem.Index(y)).FieldByName(relation.Polymorphic.Field.Name).Interface()) {
							err = orm.SetReflectValue(reflect.Indirect(resultSlice.Index(row)).FieldByName(relation.Field), rResElem.Index(y))
							if err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}

	return nil
}

// Create an entry.
// BelongsTo relation is handled before the main entry. If the belongsTo primary already exists in the database it will be updated instead of created.
// Autoincrement field will be skipped if the value is zero. Last inserted ID will be set to the struct later on.
// Error will return if no columns of the root struct have any value.
// Relations are skipped if the value is zero.
// HasOne - entries are created only. There is no check if the ID already exists in the db.
// HasMany - if no other relations exists on that orm, a batch insert can be made because the inserted ID is not required. Otherwise a create will be called for every entry. No update or db check is been made. (TODO Callbacks must be handeldt both ways)
// ManyToMany - if the ID exists in the db, its been updated otherwise its been created. the join table is batched.
func (e EagerLoading) Create(scope *orm.Scope) error {
	writePerm := orm.Permission{Write: true}
	b := scope.Builder()

	// handling belongsTo relations first
	for _, relation := range scope.Relations(writePerm) {
		if relation.Kind == orm.BelongsTo {

			// skip if the belongsTo relation object is completely empty
			if scope.CallerField(relation.Field).IsZero() {
				continue
			}

			// init the relation model
			// parent is not loaded to avoid to update the relations of the belongsTo orm model.
			rel, err := scope.InitCallerRelation(relation.Field, true)
			if err != nil {
				return err
			}

			// create or update the entry
			err = createOrUpdate(rel, false)
			if err != nil {
				return err
			}

			// set related id to the parent model
			err = orm.SetReflectValue(scope.CallerField(relation.ForeignKey.Name), rel.Scope().CallerField(relation.AssociationForeignKey.Name))
			if err != nil {
				return err
			}
		}
	}

	// get the struct variable for the scan
	insertValue := map[string]interface{}{}

	var insertColumns []string
	var autoincrement orm.Field
	for _, f := range scope.Fields(writePerm) {
		// skipping autoincrement fields if no value is set
		if f.Information.Autoincrement && scope.CallerField(f.Name).IsZero() {
			autoincrement = f
			continue
		}

		// skip empty values
		if scope.CallerField(f.Name).IsZero() {
			continue
		}

		insertValue[f.Information.Name] = scope.CallerField(f.Name).Interface()
		insertColumns = append(insertColumns, f.Information.Name)
	}

	if len(insertColumns) == 0 {
		return errors.New("orm: no value is given")
	}

	insert := b.Insert(scope.TableName()).Columns(insertColumns...).Values([]map[string]interface{}{insertValue})
	if autoincrement.Name != "" {
		insert.LastInsertedID(autoincrement.Information.Name, scope.CallerField(autoincrement.Name).Addr().Interface())
	}
	_, err := insert.Exec()
	if err != nil {
		return err
	}

	// handle the other relations
	for _, relation := range scope.Relations(writePerm) {
		// skip if no value is given
		if scope.CallerField(relation.Field).IsZero() {
			continue
		}

		switch relation.Kind {
		case orm.HasOne:
			rel, err := scope.InitCallerRelation(relation.Field, false)
			if err != nil {
				return err
			}

			// set parent ID to relation model
			err = setValue(scope, relation, reflect.Indirect(reflect.ValueOf(rel.Scope().Caller())))
			if err != nil {
				return err
			}

			err = rel.Create()
			if err != nil {
				return err
			}
		case orm.HasMany:
			rel, err := scope.InitCallerRelation(relation.Field, false)
			if err != nil {
				return err
			}
			relScope := rel.Scope()

			// if no relations on that orm exist, batch insert can be made because there are only normal fields, otherwise for every element a insert query
			// has to be made because we need the lastID for further relations
			if len(relScope.Relations(writePerm)) == 0 {
				slice := scope.CallerField(relation.Field)

				var values []map[string]interface{}
				var cols []string
				for i := 0; i < slice.Len(); i++ {
					// skip if the added value is an empty struct
					if reflect.Indirect(slice.Index(i)).IsZero() {
						continue
					}

					// set parent ID to relation model
					err = setValue(scope, relation, reflect.Indirect(slice.Index(i)))
					if err != nil {
						return err
					}

					// get the struct variable for the scan
					value := map[string]interface{}{}
					for _, f := range relScope.Fields(writePerm) {
						// skipping autoincrement fields if no value is set
						if f.Information.Autoincrement && reflect.Indirect(slice.Index(i)).FieldByName(f.Name).IsZero() {
							continue
						}
						value[f.Information.Name] = reflect.Indirect(slice.Index(i)).FieldByName(f.Name).Interface()
						if i == 0 {
							cols = append(cols, f.Information.Name)
						}
					}
					values = append(values, value)
				}

				if slice.Len() > 0 {
					_, err = b.Insert(relScope.TableName()).Columns(cols...).Values(values).Exec()
					if err != nil {
						return err
					}
				}

			} else {
				slice := scope.CallerField(relation.Field)
				for i := 0; i < slice.Len(); i++ {
					// get related struct
					var r orm.Interface
					if slice.Index(i).Kind() == reflect.Ptr {
						r = slice.Index(i).Interface().(orm.Interface)
					} else {
						r = slice.Index(i).Addr().Interface().(orm.Interface)
					}
					err = scope.InitRelation(r, relation.Field)
					if err != nil {
						return err
					}

					// set parent ID to relation model
					err = setValue(scope, relation, reflect.Indirect(reflect.ValueOf(r.Scope().Caller())))
					if err != nil {
						return err
					}

					// create the entries
					err = r.Create()
					if err != nil {
						return err
					}
				}
			}
		case orm.ManyToMany:

			// TODO m2m performance batch insert? joinTable is already batched.
			// insert in relation table only if PRIMARY is empty
			var ids []interface{}
			slice := scope.CallerField(relation.Field)

			for i := 0; i < slice.Len(); i++ {

				var r orm.Interface
				if slice.Index(i).Kind() == reflect.Ptr {
					r = slice.Index(i).Interface().(orm.Interface)
				} else {
					r = slice.Index(i).Addr().Interface().(orm.Interface)
				}

				// check if its a self reference and a ptr to itself (already initialized)
				var selfReferencedObject bool
				if relation.SelfReference == true && r.Scope() != nil {
					selfReferencedObject = true
				}

				err = scope.InitRelation(r, relation.Field)
				if err != nil {
					return err
				}

				// create or update the entry
				err = createOrUpdate(r, selfReferencedObject)
				if err != nil {
					return err
				}

				// add last inserted id for junction table
				v, err := orm.SanitizeValue(r.Scope().CallerField(relation.AssociationForeignKey.Name).Interface())
				if err != nil {
					return err
				}
				ids = append(ids, v)
			}

			if len(ids) > 0 {
				// add values
				var val []map[string]interface{}
				for _, associationId := range ids {
					val = append(val, map[string]interface{}{relation.JoinTable.ForeignKey: scope.CallerField(relation.ForeignKey.Name).Interface(), relation.JoinTable.AssociationForeignKey: associationId})
				}

				// insert into join table
				_, err = b.Insert(relation.JoinTable.Name).
					Columns(relation.JoinTable.ForeignKey, relation.JoinTable.AssociationForeignKey).
					Values(val).Exec()
				if err != nil {
					return fmt.Errorf("orm: eager m2m create: %w", err)
				}
			}
		}
	}

	return nil
}

func (e EagerLoading) Update(scope *orm.Scope, c *sqlquery.Condition) error {

	writePerm := orm.Permission{Write: true}
	b := scope.Builder()

	// handling belongsTo relations first
	for _, relation := range scope.Relations(writePerm) {
		if relation.Kind == orm.BelongsTo {
			if cV := scope.ChangedValueByFieldName(relation.Field); cV != nil {
				if cV.Field == relation.Field { //TODO
					rel, err := scope.InitCallerRelation(relation.Field, false)
					if err != nil {
						return err
					}

					switch cV.Operation {
					case orm.CREATE:
						// create or update the entry
						err = createOrUpdate(rel, false)
						if err != nil {
							return err
						}
						err = orm.SetReflectValue(scope.CallerField(relation.ForeignKey.Name), rel.Scope().CallerField(relation.AssociationForeignKey.Name))
						if err != nil {
							return err
						}
						scope.AppendChangedValue(orm.ChangedValue{Field: relation.ForeignKey.Name})
					case orm.UPDATE:
						err = orm.SetReflectValue(scope.CallerField(relation.ForeignKey.Name), rel.Scope().CallerField(relation.AssociationForeignKey.Name))
						if err != nil {
							return err
						}
						rel.Scope().SetChangedValues(cV.ChangedValue)
						err = rel.Update()
						if err != nil {
							return err
						}
					case orm.DELETE:
						err = orm.SetReflectValue(scope.CallerField(relation.ForeignKey.Name), reflect.Zero(scope.CallerField(relation.ForeignKey.Name).Type()))
						if err != nil {
							return err
						}
						scope.AppendChangedValue(orm.ChangedValue{Field: relation.ForeignKey.Name})
						// No real delete of belongsTo because there could be references? needed to really delete?
					}
				}
			}
		}
	}

	// set value
	value := map[string]interface{}{}
	var column []string
	for _, f := range scope.Fields(writePerm) {
		if scope.ChangedValueByFieldName(f.Name) != nil {
			column = append(column, f.Information.Name)
			value[f.Information.Name] = scope.CallerField(f.Name).Interface()
		}
	}

	// only update if columns are writeable
	if len(value) > 0 {
		// exec
		_, err := b.Update(scope.TableName()).Condition(c).Columns(column...).Set(value).Exec()
		if err != nil {
			return err
		}
	}

	for _, relation := range scope.Relations(writePerm) {
		switch relation.Kind {
		case orm.HasOne:
			if cV := scope.ChangedValueByFieldName(relation.Field); cV != nil {
				if cV.Field == relation.Field {

					rel, err := scope.InitCallerRelation(relation.Field, false)
					if err != nil {
						return err
					}
					relScope := rel.Scope()

					switch cV.Operation {
					case orm.CREATE:

						// set parent ID to relation model
						err = setValue(scope, relation, reflect.Indirect(reflect.ValueOf(relScope.Caller())))
						if err != nil {
							return err
						}

						deleteModel := b.Delete(relScope.TableName())
						c := sqlquery.NewCondition()
						addWhere(scope, relation, c, scope.CallerField(relation.ForeignKey.Name).Interface())
						// Delete all old references, this can happen if a user adds a new model and an old exists already.
						// TODO this should be a model.Delete instead of builder - callback wise.
						_, err = deleteModel.Condition(c).Exec()
						if err != nil {
							return err
						}

						err = rel.Create()
					case orm.UPDATE:

						// set parent ID to relation model
						err = setValue(scope, relation, reflect.Indirect(reflect.ValueOf(rel.Scope().Caller())))
						if err != nil {
							return err
						}

						relScope.SetChangedValues(cV.ChangedValue)
						err = rel.Update()
					case orm.DELETE:
						deleteModel := b.Delete(relScope.TableName())
						c := sqlquery.NewCondition()
						addWhere(scope, relation, c, scope.CallerField(relation.ForeignKey.Name).Interface())
						_, err = deleteModel.Condition(c).Exec()
						if err != nil {
							return err
						}

					}
					if err != nil {
						return err
					}
				}
			}
		case orm.HasMany:
			if cV := scope.ChangedValueByFieldName(relation.Field); cV != nil {
				rel, err := scope.InitCallerRelation(relation.Field, false)
				if err != nil {
					return err
				}
				relScope := rel.Scope()

				switch cV.Operation {
				case orm.CREATE:
					for i := 0; i < scope.CallerField(relation.Field).Len(); i++ {

						// set parent ID to relation model
						err = setValue(scope, relation, reflect.Indirect(scope.CallerField(relation.Field).Index(i)))
						if err != nil {
							return err
						}

						err = scope.InitRelation(reflect.Indirect(scope.CallerField(relation.Field).Index(i)).Addr().Interface().(orm.Interface), relation.Field)

						if err != nil {
							return err
						}
						err = reflect.Indirect(scope.CallerField(relation.Field).Index(i)).Addr().Interface().(orm.Interface).Create()
						if err != nil {
							return err
						}
					}
				case orm.UPDATE:

					var deleteID []interface{}
					for _, changes := range cV.ChangedValue {
						switch changes.Operation {
						case orm.CREATE:
							// set parent ID to relation model
							err = setValue(scope, relation, reflect.Indirect(scope.CallerField(relation.Field).Index(changes.Index.(int))))
							if err != nil {
								return err
							}

							err = scope.InitRelation(reflect.Indirect(scope.CallerField(relation.Field).Index(changes.Index.(int))).Addr().Interface().(orm.Interface), relation.Field)
							if err != nil {
								return err
							}

							// TODO what if the id already exists.
							err = reflect.Indirect(scope.CallerField(relation.Field).Index(changes.Index.(int))).Addr().Interface().(orm.Interface).Create()
							if err != nil {
								return err
							}
						case orm.UPDATE:
							tmpUpdate := reflect.Indirect(scope.CallerField(relation.Field).Index(changes.Index.(int))).Addr().Interface().(orm.Interface)
							err = scope.InitRelation(tmpUpdate, relation.Field)
							if err != nil {
								return err
							}

							// set parent ID to relation model
							err = setValue(scope, relation, reflect.Indirect(scope.CallerField(relation.Field).Index(changes.Index.(int))))
							if err != nil {
								return err
							}

							tmpUpdate.Scope().SetChangedValues(changes.ChangedValue)
							err = tmpUpdate.Update()
							if err != nil {
								return err
							}
						case orm.DELETE:
							deleteID = append(deleteID, changes.Index)
						}
					}
					if len(deleteID) > 0 {
						deleteModel := b.Delete(relScope.TableName())
						pKeys := relScope.PrimaryKeys() // TODO allow multiple primary keys
						deleteModel.Where(b.QuoteIdentifier(pKeys[0].Information.Name)+" IN (?)", deleteID)
						_, err = deleteModel.Exec()
						if err != nil {
							return err
						}
					}
				case orm.DELETE:
					deleteModel := b.Delete(relScope.TableName())
					c := sqlquery.NewCondition()
					addWhere(scope, relation, c, scope.CallerField(relation.ForeignKey.Name).Interface())
					_, err = deleteModel.Condition(c).Exec()
					if err != nil {
						return err
					}
				}
			}
		case orm.ManyToMany:

			if cV := scope.ChangedValueByFieldName(relation.Field); cV != nil {
				switch cV.Operation {
				case orm.CREATE:
					var joinTable []map[string]interface{}
					for i := 0; i < scope.CallerField(relation.Field).Len(); i++ {

						tmpCreate := reflect.Indirect(scope.CallerField(relation.Field).Index(i)).Addr().Interface().(orm.Interface)
						err := scope.InitRelation(tmpCreate, relation.Field)
						if err != nil {
							return err
						}

						// create or update the entry
						err = createOrUpdate(tmpCreate, relation.SelfReference)
						if err != nil {
							return err
						}

						joinTable = append(joinTable, map[string]interface{}{relation.JoinTable.ForeignKey: scope.CallerField(relation.ForeignKey.Name).Interface(), relation.JoinTable.AssociationForeignKey: reflect.Indirect(scope.CallerField(relation.Field).Index(i)).FieldByName(relation.AssociationForeignKey.Name).Interface()})
					}
					// batch insert
					if len(joinTable) > 0 {
						_, err := b.Insert(relation.JoinTable.Name).Columns(relation.JoinTable.ForeignKey, relation.JoinTable.AssociationForeignKey).Values(joinTable).Exec()
						if err != nil {
							return err
						}
					}
				case orm.UPDATE:

					var deleteID []interface{}
					var createID []map[string]interface{}
					for _, changes := range cV.ChangedValue {

						switch changes.Operation {
						case orm.CREATE:

							tmpCreate := reflect.Indirect(scope.CallerField(relation.Field).Index(changes.Index.(int))).Addr().Interface().(orm.Interface)
							err := scope.InitRelation(tmpCreate, relation.Field)
							if err != nil {
								return err
							}

							// create or update the entry
							err = createOrUpdate(tmpCreate, false)
							if err != nil {
								return err
							}

							createID = append(createID, map[string]interface{}{relation.JoinTable.ForeignKey: scope.CallerField(relation.ForeignKey.Name).Interface(), relation.JoinTable.AssociationForeignKey: reflect.Indirect(scope.CallerField(relation.Field).Index(changes.Index.(int))).FieldByName(relation.AssociationForeignKey.Name).Interface()})
						case orm.UPDATE:
							tmpUpdate := reflect.Indirect(scope.CallerField(relation.Field).Index(changes.Index.(int))).Addr().Interface().(orm.Interface)

							err := scope.InitRelation(tmpUpdate, relation.Field)
							if err != nil {
								return err
							}

							tmpScope := tmpUpdate.Scope()
							tmpScope.SetChangedValues(changes.ChangedValue)
							err = tmpUpdate.Update()
							if err != nil {
								return err
							}
						case orm.DELETE:
							deleteID = append(deleteID, changes.Index)
						}
					}

					if len(deleteID) > 0 {
						_, err := b.Delete(relation.JoinTable.Name).
							Where(b.QuoteIdentifier(relation.JoinTable.ForeignKey)+" = ?", scope.CallerField(relation.ForeignKey.Name).Interface()).
							Where(b.QuoteIdentifier(relation.JoinTable.AssociationForeignKey)+" IN (?)", deleteID).
							Exec()
						if err != nil {
							return err
						}
					}

					if len(createID) > 0 {
						_, err := b.Insert(relation.JoinTable.Name).Columns(relation.JoinTable.ForeignKey, relation.JoinTable.AssociationForeignKey).Values(createID).Exec()
						if err != nil {
							return err
						}
					}

				case orm.DELETE:
					_, err := b.Delete(relation.JoinTable.Name).Where(b.QuoteIdentifier(relation.JoinTable.ForeignKey)+" = ?", scope.CallerField(relation.ForeignKey.Name).Interface()).Exec()
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}
func (e EagerLoading) Delete(scope *orm.Scope, c *sqlquery.Condition) error {

	b := scope.Builder()
	writePerm := orm.Permission{Write: true}
	// handling belongsTo relations first
	for _, relation := range scope.Relations(writePerm) {
		switch relation.Kind {
		case orm.BelongsTo:
			// ignore - belongsTo - stays untouched
		case orm.HasOne, orm.HasMany:
			// hasOne - deleteSql - ignore softDelete if the main struct has none.
			var deleteSql *sqlquery.Delete
			if scope.IsPolymorphic(relation) {
				deleteSql = b.Delete(relation.Polymorphic.Field.Information.Table)
				deleteSql.Where(b.QuoteIdentifier(relation.Polymorphic.Field.Information.Name)+" = ?", scope.CallerField(relation.ForeignKey.Name).Interface())
				deleteSql.Where(b.QuoteIdentifier(relation.Polymorphic.Type.Information.Name)+" = ?", relation.Polymorphic.Value)
			} else {
				deleteSql = b.Delete(relation.AssociationForeignKey.Information.Table)
				deleteSql.Where(b.QuoteIdentifier(relation.AssociationForeignKey.Information.Name)+" = ?", scope.CallerField(relation.ForeignKey.Name).Interface())
			}
			_, err := deleteSql.Exec()
			if err != nil {
				return err
			}
		case orm.ManyToMany:
			// hasManyToMany - only junction table entries are getting deleted - for the association table use SQL CASCADE or a callbacks
			_, err := b.Delete(relation.JoinTable.Name).Where(b.QuoteIdentifier(relation.JoinTable.ForeignKey)+" = ?", scope.CallerField(relation.ForeignKey.Name).Interface()).Exec()
			if err != nil {
				return err
			}
		}
	}

	// exec
	_, err := b.Delete(scope.TableName()).Condition(c).Exec()
	if err != nil {
		return err
	}

	return nil
}

// addWhere helper. it handles polymorphic and slices.
func addWhere(scope *orm.Scope, relation orm.Relation, c *sqlquery.Condition, value interface{}) {

	op := " = ?"
	if reflect.TypeOf(value).Kind() == reflect.Slice {
		op = " IN (?)"
	}

	if scope.IsPolymorphic(relation) {
		c.Where(scope.Builder().QuoteIdentifier(relation.Polymorphic.Field.Information.Name)+op, value)
		c.Where(scope.Builder().QuoteIdentifier(relation.Polymorphic.Type.Information.Name)+" = ?", relation.Polymorphic.Value)
	} else {
		c.Where(scope.Builder().QuoteIdentifier(relation.AssociationForeignKey.Information.Name)+op, value)
	}
}

// compareValues is a helper function to sanitize the value to a string and compare it.
func compareValues(v1 interface{}, v2 interface{}) bool {
	s1, err := orm.SanitizeToString(v1)
	if err != nil {
		return false
	}
	s2, err := orm.SanitizeToString(v2)
	if err != nil {
		return false
	}

	return s1 == s2
}

// createOrUpdate is a helper to create an entry if the primary keys are missing.
// It updates an entry if primary keys exist and its existing in the database, otherwise it will create the entry.
func createOrUpdate(rel orm.Interface, selfReference bool) error {

	if !rel.Scope().PrimariesSet() {
		return rel.Create()
	} else {

		// if only the belongsTo foreign key and the manyToMany join table should be updated.
		if parent, err := rel.Scope().Parent(""); err == nil && parent.Scope().ReferencesOnly() {
			return nil
		}

		// on self reference there is a problem with a loop, so the changed value is not checked again.
		if !selfReference {
			rel.Scope().TakeSnapshot()
		}
		err := rel.Update()
		// if the ID does not exist yet, error will be thrown. then create it.
		if err == sql.ErrNoRows {
			err = rel.Create()
		}

		if selfReference {
			rel.Scope().UnsetParent()
		}
		return err
	}
}

// setValue is a helper to set the parent foreign key to the relation field.
// Its taking care of polymorphic.
func setValue(scope *orm.Scope, relation orm.Relation, field reflect.Value) error {
	if !scope.IsPolymorphic(relation) {
		err := orm.SetReflectValue(field.FieldByName(relation.AssociationForeignKey.Name), scope.CallerField(relation.ForeignKey.Name))
		if err != nil {
			return err
		}
	} else {
		err := orm.SetReflectValue(field.FieldByName(relation.Polymorphic.Field.Name), scope.CallerField(relation.ForeignKey.Name))
		if err != nil {
			return err
		}
		err = orm.SetReflectValue(field.FieldByName(relation.Polymorphic.Type.Name), reflect.ValueOf(relation.Polymorphic.Value))
		if err != nil {
			return err
		}
	}
	return nil
}

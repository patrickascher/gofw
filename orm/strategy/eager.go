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

// addWhere helper for the First method.
func addWhere(scope *orm.Scope, relation orm.Relation, c *sqlquery.Condition) {
	if scope.IsPolymorphic(relation) {
		c.Where(scope.Builder().QuoteIdentifier(relation.Polymorphic.Field.Information.Name)+" = ?", scope.CallerField(relation.ForeignKey.Name).Interface())
		c.Where(scope.Builder().QuoteIdentifier(relation.Polymorphic.Type.Information.Name)+" = ?", relation.Polymorphic.Value)
	} else {
		c.Where(scope.Builder().QuoteIdentifier(relation.AssociationForeignKey.Information.Name)+" = ?", scope.CallerField(relation.ForeignKey.Name).Interface())
	}
}
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

		// init Rel
		c := &sqlquery.Condition{}
		rel, err := scope.InitCallerRelation(relation.Field, false)
		if err != nil {
			return err
		}

		// handling relation
		switch relation.Kind {
		case orm.HasOne, orm.BelongsTo:
			addWhere(scope, relation, c)
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
				addWhere(scope, relation, c)
			} else {
				// TODO INNER JOIN faster?
				c.Where(b.QuoteIdentifier(relation.AssociationForeignKey.Information.Name)+" IN (SELECT "+b.QuoteIdentifier(relation.JoinTable.AssociationForeignKey)+" FROM "+b.QuoteIdentifier(relation.JoinTable.Name)+" WHERE "+b.QuoteIdentifier(relation.JoinTable.ForeignKey)+" = ?)", scope.CallerField(relation.ForeignKey.Name).Interface())
			}
			c.Order(relation.AssociationForeignKey.Information.Name) // TODO make it possible to create a order by tag or something

			// CUSTOM Condition on relations, TODO improvements
			if v := scope.Model().RelationCondition(relation.Field); v != nil {
				c = v
			}

			err = rel.All(scope.CallerField(relation.Field).Addr().Interface(), c)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// TODO back-references end in an infinity loop error at the moment
// TODO at the moment only INTs are working (whats with UUIDs,strings)...
// TODO better solution of defer in a for loop, func?
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
	// Later on the data will be mapped to the correct result again.
	// TODO at the moment only INTs are working (whats with UUIDs,strings)...
	in := map[string][]int{}
	for _, relation := range scope.Relations(readPerm) {
		f := relation.ForeignKey.Name
		if _, ok := in[f]; !ok {
			for n := 0; n < resultSlice.Len(); n++ {
				// TODO type switch if string (uuid)
				i := orm.Int(reflect.Indirect(resultSlice.Index(n)).FieldByName(f).Interface())
				if _, exist := slices.ExistInt(in[f], i); !exist {
					in[f] = append(in[f], i)
				}
			}
		}

		// Special case, if its a many to many relation.
		// To avoid multiple sql selects, all many2many results of this relation and the main model ids will be loaded with two queries.
		// The first query checks the join table for all needed IDs. The second query requests the model by the ids.
		m2mMap := map[int][]int{}
		var m2mAll []int
		if relation.Kind == orm.ManyToMany {

			var fkInt []int
			for row := 0; row < resultSlice.Len(); row++ {
				fkInt = append(fkInt, orm.Int(reflect.Indirect(resultSlice.Index(row)).FieldByName(relation.ForeignKey.Name).Interface()))
			}

			rows, err := b.Select(relation.JoinTable.Name).Columns(relation.JoinTable.ForeignKey, relation.JoinTable.AssociationForeignKey).Where(b.QuoteIdentifier(relation.JoinTable.ForeignKey)+" IN (?)", fkInt).All()
			if err != nil {
				return err
			}
			defer rows.Close() // TODO better solution for defer, func?

			// map fk,afk
			for rows.Next() {
				var fk int
				var afk int
				err = rows.Scan(&fk, &afk)
				if err != nil {
					return err
				}
				m2mMap[fk] = append(m2mMap[fk], afk)
				if _, exists := slices.ExistInt(m2mAll, afk); !exists {
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
				if scope.IsPolymorphic(relation) {
					c.Where(b.QuoteIdentifier(relation.Polymorphic.Field.Information.Name)+" IN (?)", in[f])
					c.Where(b.QuoteIdentifier(relation.Polymorphic.Type.Information.Name)+" = ?", relation.Polymorphic.Value)
				} else {
					c.Where(b.QuoteIdentifier(relation.AssociationForeignKey.Information.Name)+" IN (?)", in[f])
				}
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
					parentID := orm.Int(modelField.Interface())

					// set data to the main model.
					if relation.Kind == orm.ManyToMany {

						v, ok := m2mMap[parentID]
						_, exists := slices.ExistInt(v, orm.Int(reflect.Indirect(rResElem.Index(y)).FieldByName(relation.AssociationForeignKey.Name).Interface()))
						if ok && exists {
							err = orm.SetReflectValue(reflect.Indirect(resultSlice.Index(row)).FieldByName(relation.Field), rResElem.Index(y))
							if err != nil {
								return err
							}
						}
					} else {
						if (!scope.IsPolymorphic(relation) && parentID == orm.Int(reflect.Indirect(rResElem.Index(y)).FieldByName(relation.AssociationForeignKey.Name).Interface())) ||
							scope.IsPolymorphic(relation) && parentID == orm.Int(reflect.Indirect(rResElem.Index(y)).FieldByName(relation.Polymorphic.Field.Name).Interface()) {
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

// TODO: on hasMany orm which have an additional relation a insert is done for every single item because we need the last inserted ID for further relations. Solution tx options?
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
			// parent is not loaded because the belongsTo orm could be updated.
			rel, err := scope.InitCallerRelation(relation.Field, true)
			if err != nil {
				return err
			}

			// create new entry if primary fields are empty, otherwise update entry
			if !rel.Scope().PrimariesSet() {
				err = rel.Create()
			} else {
				err = rel.Update()
			}
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
			if !scope.IsPolymorphic(relation) {
				err = orm.SetReflectValue(rel.Scope().CallerField(relation.AssociationForeignKey.Name), scope.CallerField(relation.ForeignKey.Name))
				if err != nil {
					return err
				}
			} else {
				err = orm.SetReflectValue(rel.Scope().CallerField(relation.Polymorphic.Field.Name), scope.CallerField(relation.ForeignKey.Name))
				if err != nil {
					return err
				}
				err = orm.SetReflectValue(rel.Scope().CallerField(relation.Polymorphic.Type.Name), reflect.ValueOf(relation.Polymorphic.Value))
				if err != nil {
					return err
				}
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
					if !scope.IsPolymorphic(relation) {
						err = orm.SetReflectValue(reflect.Indirect(slice.Index(i)).FieldByName(relation.AssociationForeignKey.Name), scope.CallerField(relation.ForeignKey.Name))
						if err != nil {
							return err
						}
					} else {
						err = orm.SetReflectValue(reflect.Indirect(slice.Index(i)).FieldByName(relation.Polymorphic.Field.Name), scope.CallerField(relation.ForeignKey.Name))
						if err != nil {
							return err
						}
						err = orm.SetReflectValue(reflect.Indirect(slice.Index(i)).FieldByName(relation.Polymorphic.Type.Name), reflect.ValueOf(relation.Polymorphic.Value))
						if err != nil {
							return err
						}
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
					if !scope.IsPolymorphic(relation) {
						err = orm.SetReflectValue(r.Scope().CallerField(relation.AssociationForeignKey.Name), scope.CallerField(relation.ForeignKey.Name))
						if err != nil {
							return err
						}
					} else {
						err = orm.SetReflectValue(r.Scope().CallerField(relation.Polymorphic.Field.Name), scope.CallerField(relation.ForeignKey.Name))
						if err != nil {
							return err
						}
						err = orm.SetReflectValue(r.Scope().CallerField(relation.Polymorphic.Type.Name), reflect.ValueOf(relation.Polymorphic.Value))
						if err != nil {
							return err
						}
					}

					// create the entries
					err = r.Create()
					if err != nil {
						return err
					}
				}
			}
		case orm.ManyToMany:
			// TODO m2m performance batch insert?
			// insert in relation table only if PRIMARY is empty
			var ids []int // TODO fix whats with UUIDs or none sequence?
			slice := scope.CallerField(relation.Field)
			for i := 0; i < slice.Len(); i++ {

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

				rScope := r.Scope()

				// if primary fields are empty, create entry otherwise ignore
				// TODO check also Association FK?
				if !rScope.PrimariesSet() {
					err = r.Create()
					if err != nil {
						return err
					}
				}

				// add last inserted id for junction table
				ids = append(ids, orm.Int(rScope.CallerField(relation.AssociationForeignKey.Name).Interface()))
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
					relScope := rel.Scope()

					switch cV.Operation {
					case orm.CREATE:

						// TODO same code as in hasMany update/create
						// if primary fields are set, dont create it on the association table.
						// the changed values map of the orm.model is just checking against the snapshot, so it does not know if the value is in the database or not.
						// to avoid duplicated entries, there is a check. // maybe make a real db check?
						if !relScope.PrimariesSet() {
							err = rel.Create()
							if err != nil {
								return err
							}
						}

						err = orm.SetReflectValue(scope.CallerField(relation.ForeignKey.Name), relScope.CallerField(relation.AssociationForeignKey.Name))
						if err != nil {
							return err
						}
						scope.AppendChangedValue(orm.ChangedValue{Field: relation.ForeignKey.Name})
					case orm.UPDATE:
						err = orm.SetReflectValue(scope.CallerField(relation.ForeignKey.Name), relScope.CallerField(relation.AssociationForeignKey.Name))
						if err != nil {
							return err
						}
						relScope.SetChangedValues(cV.ChangedValue)
					case orm.DELETE:
						err = orm.SetReflectValue(scope.CallerField(relation.ForeignKey.Name), reflect.Zero(scope.CallerField(relation.ForeignKey.Name).Type()))
						if err != nil {
							return err
						}
						scope.AppendChangedValue(orm.ChangedValue{Field: relation.ForeignKey.Name})
					}
					if err != nil {
						return err
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

						deleteModel := b.Delete(relScope.TableName())
						if scope.IsPolymorphic(relation) {
							err = orm.SetReflectValue(relScope.CallerField(relation.Polymorphic.Field.Name), scope.CallerField(relation.ForeignKey.Name))
							if err != nil {
								return err
							}
							err = orm.SetReflectValue(relScope.CallerField(relation.Polymorphic.Type.Name), reflect.ValueOf(relation.Polymorphic.Value))
							if err != nil {
								return err
							}

							deleteModel.Where(b.QuoteIdentifier(relation.Polymorphic.Field.Information.Name)+" = ?", scope.CallerField(relation.ForeignKey.Name).Interface())
							deleteModel.Where(b.QuoteIdentifier(relation.Polymorphic.Type.Information.Name)+" = ?", relation.Polymorphic.Value)
						} else {
							err = orm.SetReflectValue(relScope.CallerField(relation.AssociationForeignKey.Name), scope.CallerField(relation.ForeignKey.Name))
							if err != nil {
								return err
							}

							deleteModel.Where(b.QuoteIdentifier(relation.AssociationForeignKey.Information.Name)+" = ?", scope.CallerField(relation.ForeignKey.Name).Interface())
						}

						// Delete all old references, this can happen if a user adds a new model and an old exists already.
						// TODO this should be a model.Delete instead of builder - callback wise.
						_, err = deleteModel.Exec()
						if err != nil {
							return err
						}

						err = rel.Create()
					case orm.UPDATE:
						// if poly, set the relation fields. It could be that the user only adds the primary key for an update.
						if scope.IsPolymorphic(relation) {
							err = orm.SetReflectValue(rel.Scope().CallerField(relation.Polymorphic.Field.Name), scope.CallerField(relation.ForeignKey.Name))
							if err != nil {
								return err
							}
							err = orm.SetReflectValue(rel.Scope().CallerField(relation.Polymorphic.Type.Name), reflect.ValueOf(relation.Polymorphic.Value))
							if err != nil {
								return err
							}
						}
						relScope.SetChangedValues(cV.ChangedValue)
						err = rel.Update()
					case orm.DELETE:
						deleteModel := b.Delete(relScope.TableName())
						if scope.IsPolymorphic(relation) {
							deleteModel.Where(b.QuoteIdentifier(relation.Polymorphic.Field.Information.Name)+" = ?", scope.CallerField(relation.ForeignKey.Name).Interface())
							deleteModel.Where(b.QuoteIdentifier(relation.Polymorphic.Type.Information.Name)+" = ?", relation.Polymorphic.Value)
						} else {
							deleteModel.Where(b.QuoteIdentifier(relation.AssociationForeignKey.Information.Name)+" = ?", scope.CallerField(relation.ForeignKey.Name).Interface())
						}
						_, err = deleteModel.Exec()
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
						if !scope.IsPolymorphic(relation) {
							err = orm.SetReflectValue(reflect.Indirect(scope.CallerField(relation.Field).Index(i)).FieldByName(relation.AssociationForeignKey.Name), scope.CallerField(relation.ForeignKey.Name))
							if err != nil {
								return err
							}
						} else {
							err = orm.SetReflectValue(reflect.Indirect(scope.CallerField(relation.Field).Index(i)).FieldByName(relation.Polymorphic.Field.Name), scope.CallerField(relation.ForeignKey.Name))
							if err != nil {
								return err
							}
							err = orm.SetReflectValue(reflect.Indirect(scope.CallerField(relation.Field).Index(i)).FieldByName(relation.Polymorphic.Type.Name), reflect.ValueOf(relation.Polymorphic.Value))
							if err != nil {
								return err
							}
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

					var deleteID []int
					for _, changes := range cV.ChangedValue {
						switch changes.Operation {
						case orm.CREATE:
							//TODO same as create above - helper
							if !scope.IsPolymorphic(relation) {
								err = orm.SetReflectValue(reflect.Indirect(scope.CallerField(relation.Field).Index(changes.Index)).FieldByName(relation.AssociationForeignKey.Name), scope.CallerField(relation.ForeignKey.Name))
								if err != nil {
									return err
								}
							} else {
								err = orm.SetReflectValue(reflect.Indirect(scope.CallerField(relation.Field).Index(changes.Index)).FieldByName(relation.Polymorphic.Field.Name), scope.CallerField(relation.ForeignKey.Name))
								if err != nil {
									return err
								}
								err = orm.SetReflectValue(reflect.Indirect(scope.CallerField(relation.Field).Index(changes.Index)).FieldByName(relation.Polymorphic.Type.Name), reflect.ValueOf(relation.Polymorphic.Value))
								if err != nil {
									return err
								}
							}
							err = scope.InitRelation(reflect.Indirect(scope.CallerField(relation.Field).Index(changes.Index)).Addr().Interface().(orm.Interface), relation.Field)
							if err != nil {
								return err
							}

							// TODO what if the id already exists.
							err = reflect.Indirect(scope.CallerField(relation.Field).Index(changes.Index)).Addr().Interface().(orm.Interface).Create()
							if err != nil {
								return err
							}
						case orm.UPDATE:
							tmpUpdate := reflect.Indirect(scope.CallerField(relation.Field).Index(changes.Index)).Addr().Interface().(orm.Interface)
							err = scope.InitRelation(tmpUpdate, relation.Field)
							if err != nil {
								return err
							}

							tmpScope := tmpUpdate.Scope()
							// TODO create helper, same as before on create
							if !scope.IsPolymorphic(relation) {
								err = orm.SetReflectValue(reflect.Indirect(scope.CallerField(relation.Field).Index(changes.Index)).FieldByName(relation.AssociationForeignKey.Name), scope.CallerField(relation.ForeignKey.Name))
								if err != nil {
									return err
								}
							} else {
								err = orm.SetReflectValue(reflect.Indirect(scope.CallerField(relation.Field).Index(changes.Index)).FieldByName(relation.Polymorphic.Field.Name), scope.CallerField(relation.ForeignKey.Name))
								if err != nil {
									return err
								}
								err = orm.SetReflectValue(reflect.Indirect(scope.CallerField(relation.Field).Index(changes.Index)).FieldByName(relation.Polymorphic.Type.Name), reflect.ValueOf(relation.Polymorphic.Value))
								if err != nil {
									return err
								}
							}

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
					if scope.IsPolymorphic(relation) {
						deleteModel.Where(b.QuoteIdentifier(relation.Polymorphic.Field.Information.Name)+" = ?", scope.CallerField(relation.ForeignKey.Name).Interface())
						deleteModel.Where(b.QuoteIdentifier(relation.Polymorphic.Type.Information.Name)+" = ?", relation.Polymorphic.Value)
					} else {
						deleteModel.Where(b.QuoteIdentifier(relation.AssociationForeignKey.Information.Name)+" = ?", scope.CallerField(relation.ForeignKey.Name).Interface())
					}
					_, err = deleteModel.Exec()
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

						// TODO same code as in update
						// if primary fields are set, dont create it on the association table.
						// the changed values map of the orm.model is just checking against the snapshot, so it does not know if the value is in the database or not.
						// to avoid duplicated entries, there is a check. // maybe make a real db check?
						if !tmpCreate.Scope().PrimariesSet() {
							err = tmpCreate.Create()
							if err != nil {
								return err
							}
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

					var deleteID []int
					var createID []map[string]interface{}
					for _, changes := range cV.ChangedValue {

						switch changes.Operation {
						case orm.CREATE:

							tmpCreate := reflect.Indirect(scope.CallerField(relation.Field).Index(changes.Index)).Addr().Interface().(orm.Interface)
							err := scope.InitRelation(tmpCreate, relation.Field)
							if err != nil {
								return err
							}

							// TODO same code as in create
							// if primary fields are set, dont create it on the association table.
							// the changed values map of the orm.model is just checking against the snapshot, so it does not know if the value is in the database or not.
							// to avoid duplicated entries, there is a check. // maybe make a real db check?
							if !tmpCreate.Scope().PrimariesSet() {
								err = tmpCreate.Create()
								if err != nil {
									return err
								}
							}

							createID = append(createID, map[string]interface{}{relation.JoinTable.ForeignKey: scope.CallerField(relation.ForeignKey.Name).Interface(), relation.JoinTable.AssociationForeignKey: reflect.Indirect(scope.CallerField(relation.Field).Index(changes.Index)).FieldByName(relation.AssociationForeignKey.Name).Interface()})
						case orm.UPDATE:
							tmpUpdate := reflect.Indirect(scope.CallerField(relation.Field).Index(changes.Index)).Addr().Interface().(orm.Interface)

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

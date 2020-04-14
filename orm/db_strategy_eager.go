package orm

import (
	"fmt"
	"github.com/guregu/null"
	"github.com/patrickascher/gofw/sqlquery"
	"reflect"
	"strings"
	"unsafe"
)

func init() {
	_ = Register("eager", &EagerLoading{})
}

// EagerLoading strategy
type EagerLoading struct {
}

// First will request one row by the given condition and adds it to the given caller struct.
// Relations:
// HasOne, BelongsTo: Handled in one request each
// HasMany, ManyToMany: Handled in one request each
func (e EagerLoading) First(m Interface, c *sqlquery.Condition) error {
	b := m.Table().Builder

	// get the struct variable for the scan
	var values []interface{}
	for _, col := range m.Table().Columns(READDB) {
		values = append(values, reflectField(m, col.StructField).Addr().Interface())
	}

	// build select for the main struct
	schema := ""
	//if b.Config().Driver() == "postgres" {
	//	schema = ".public"
	//}

	row, err := b.Select(m.Table().Database + schema + "." + m.Table().Name).Columns(m.Table().columnNames(READDB, !m.disableCustomSql())...).Condition(c).First()
	if err != nil {
		return err
	}

	// scan all variables to fill it with values
	err = row.Scan(values...)
	if err != nil {
		if err.Error() != "sql: no rows in result set" {
			return err
		}
		return nil
	}

	// handle relations
	for field, relation := range m.Table().Relations(m.whiteBlacklist(), READALL) {

		c := &sqlquery.Condition{}
		if val, ok := m.RelationCondition()[field]; ok {
			c = &val
		}

		rel, err := initRelation(m, field)
		if err != nil {
			return err
		}

		// set white - blacklist from parent
		rel.setWhiteBlacklist(RelationWhiteBlackList(m.whiteBlacklist(), field))
		rel.setParent(m)
		relField, _ := reflect.TypeOf(m).Elem().FieldByName(field)
		rel.callback().setRelField(relField)
		rel.setLoopMap(m.getLoopMap())

		switch relation.Type {
		case CustomSlice, CustomStruct:
			var err error
			if relation.Type == CustomSlice {
				rel.callback().setMode("First")
				err = rel.All(reflectField(m, field).Addr().Interface(), nil)
			} else {
				err = rel.First(nil)
			}
			if err != nil {
				return err
			}
		case HasOne, BelongsTo:
			c.Where(b.QuoteIdentifier(relation.AssociationTable.Information.Name)+" = ?", reflectField(m, relation.StructTable.StructField).Interface())

			err = rel.First(c)
			if err != nil {
				return err
			}
		case HasMany, ManyToMany, ManyToManySR:
			if relation.Type == HasMany {
				c.Where(b.QuoteIdentifier(relation.AssociationTable.Information.Name)+" = ?", reflectField(m, relation.StructTable.StructField).Interface())
			} else {
				c.Where(b.QuoteIdentifier(relation.AssociationTable.Information.Name)+" IN (SELECT "+b.QuoteIdentifier(relation.JunctionTable.AssociationColumn)+" FROM "+b.QuoteIdentifier(relation.JunctionTable.Table)+" WHERE "+b.QuoteIdentifier(relation.JunctionTable.StructColumn)+" = ?)", reflectField(m, relation.StructTable.StructField).Interface())
			}

			c.Order("id") // TODO make it possible to create a order by tag or something
			err = rel.All(reflectField(m, field).Addr().Interface(), c)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// All returns all rows by its condition.
// It fetches all results, loops over it and fetches the pk. With that primary keys the relation requests will get called.
// After that, in a loop the values are added to the correct result entry in the slice.
// Relation:
// HasOne, BelongsTo and HasMany are handled in one sql statement.
// ManyToMany: for each result a own sql statement is made.
// TODO improvements, manyToMany
func (e EagerLoading) All(res interface{}, m Interface, c *sqlquery.Condition) error {

	// checking if the res is a ptr
	if reflect.TypeOf(res).Kind() != reflect.Ptr {
		return ErrResultPtr
	}

	// build select for the main struct
	b := m.Table().Builder
	// build select for the main struct
	schema := ""
	//if b.Config().Driver() == "postgres" {
	//	schema = ".public"
	//}
	rows, err := b.Select(m.Table().Database + schema + "." + m.Table().Name).Columns(m.Table().columnNames(READDB, !m.disableCustomSql())...).Condition(c).All()
	if err != nil {
		return err
	}
	defer rows.Close()

	// slice value
	resultSlice := reflect.ValueOf(res).Elem()

	//convert db results to struct result
	for rows.Next() {
		//new instance of slice element
		sliceElement := newValueInstanceFromType(resultSlice.Type().Elem())
		sliceModel := sliceElement.Addr().Interface().(Interface)

		// here the caller gets set. This is used at the moment for the MarshalJSON method.
		// if we need later on more data, m.Initialize could be called, Initialize is 4 times slower as just set the caller at the moment.
		field := reflect.ValueOf(sliceModel).Elem().FieldByName("caller")
		field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
		field.Set(reflect.ValueOf(sliceModel))

		// get the struct variable for the scan
		var values []interface{}
		for _, col := range m.Table().Columns(READDB) {
			values = append(values, reflectField(sliceModel, col.StructField).Addr().Interface())
		}

		//add the values
		err = rows.Scan(values...)
		if err != nil {
			return err
		}

		// adding ptr or value depending on users struct definition
		setValue(resultSlice, reflect.ValueOf(sliceModel).Elem())
	}

	// checking if there were any rows
	reflectRes := reflect.ValueOf(res).Elem()
	if reflectRes.Len() == 0 {
		return nil
	}

	// handle relations
	in := map[string][]int{} // TODO fix this, at the moment only int is possible - whats with uuid´s?
	for field, relation := range m.Table().Relations(m.whiteBlacklist(), READALL) {

		// TODO atm for each result row there is an extra request. This should be combined to one sql request. for that we need to have an reference of the parent struct. write the sql with the builder.
		if relation.Type == ManyToMany || relation.Type == ManyToManySR {
			for row := 0; row < reflectRes.Len(); row++ { //Parent result

				// create condition
				c := sqlquery.Condition{}
				if val, ok := m.RelationCondition()[field]; ok {
					c = val
				}
				c.Where(b.QuoteIdentifier(relation.AssociationTable.Information.Table+"."+relation.AssociationTable.Information.Name)+" IN (SELECT "+b.QuoteIdentifier(relation.JunctionTable.AssociationColumn)+" FROM "+b.QuoteIdentifier(relation.JunctionTable.Table)+" WHERE "+b.QuoteIdentifier(relation.JunctionTable.StructColumn)+" = ?)", reflectRes.Index(row).FieldByName(relation.StructTable.StructField).Interface())

				// create a new result set for the query
				resultSet := reflect.New(reflect.MakeSlice(reflect.SliceOf(newValueInstanceFromType(reflectField(m, field).Type()).Type()), 0, 0).Type()).Interface()
				if reflectRes.Index(row).FieldByName(field).Type().Elem().Kind() == reflect.Ptr {
					resultSet = reflect.New(reflect.MakeSlice(reflect.SliceOf(newValueInstanceFromType(reflectField(m, field).Type()).Addr().Type()), 0, 0).Type()).Interface()
				}

				// create model
				relationModel, err := initRelation(m, field)
				if err != nil {
					return err
				}
				// set white- blacklist from parent
				relationModel.setWhiteBlacklist(RelationWhiteBlackList(m.whiteBlacklist(), field))
				relationModel.setLoopMap(m.getLoopMap())
				relationModel.setParent(m)

				err = relationModel.All(resultSet, &c)
				if err != nil {
					return err
				}

				// set data
				reflect.Indirect(reflectRes.Index(row)).FieldByName(field).Set(reflect.ValueOf(resultSet).Elem())
			}
			continue
		}

		// if fk field does not exist yet - add to map
		f := relation.StructTable.StructField
		if _, ok := in[f]; !ok {
			for n := 0; n < reflectRes.Len(); n++ {
				reflectRes.Index(n).FieldByName(f).Interface()

				switch reflectRes.Index(n).FieldByName(f).Interface().(type) {
				case int:
					if !inSlice(reflectRes.Index(n).FieldByName(f).Interface().(int), in[f]) {
						in[f] = append(in[f], reflectRes.Index(n).FieldByName(f).Interface().(int))
					}
				case null.Int:
					if reflectRes.Index(n).FieldByName(f).Interface().(null.Int).Valid == true && !inSlice(int(reflectRes.Index(n).FieldByName(f).Interface().(null.Int).Int64), in[f]) {
						in[f] = append(in[f], int(reflectRes.Index(n).FieldByName(f).Interface().(null.Int).Int64))
					}
				}

			}
		}

		if len(in[f]) > 0 {

			// create condition
			c := &sqlquery.Condition{}
			if val, ok := m.RelationCondition()[field]; ok {
				c = &val
			}
			c.Where(b.QuoteIdentifier(relation.AssociationTable.Information.Name)+" IN (?)", in[f])

			// Create an empty slice for resultSet
			resultSet := reflect.New(reflect.MakeSlice(reflect.SliceOf(newValueInstanceFromType(reflectField(m, field).Type()).Type()), 0, 0).Type()).Interface()

			// create model
			relationModel, err := initRelation(m, field)
			if err != nil {
				return err
			}
			// set white- blacklist from parent
			relationModel.setWhiteBlacklist(RelationWhiteBlackList(m.whiteBlacklist(), field))
			relationModel.setLoopMap(m.getLoopMap())
			relationModel.setParent(m)
			relField, _ := reflect.TypeOf(m).Elem().FieldByName(field)
			relationModel.callback().setRelField(relField)

			err = relationModel.All(resultSet, c)
			if err != nil {
				return err
			}

			// loop over all parent results to add the correct data
			res := reflect.ValueOf(resultSet).Elem()
			for row := 0; row < reflectRes.Len(); row++ { //Parent result
				for y := 0; y < res.Len(); y++ { //result set

					int64 := int64(0)
					switch reflectRes.Index(row).FieldByName(relation.StructTable.StructField).Interface().(type) {
					case int:
						int64 = reflectRes.Index(row).FieldByName(relation.StructTable.StructField).Int()
					case null.Int:
						nullInt := reflectRes.Index(row).FieldByName(relation.StructTable.StructField).Interface().(null.Int)
						if nullInt.Valid == true {
							int64 = nullInt.Int64
						}
					}

					if int64 == res.Index(y).FieldByName(relation.AssociationTable.StructField).Int() {
						switch relation.Type {
						case HasOne, BelongsTo, HasMany, CustomStruct, CustomSlice:
							setValue(reflectRes.Index(row).FieldByName(field), res.Index(y))
						}
					}

				}
			}
		}
	}

	return nil
}

// Create an entry by the given struct value.
// Its skipping the autoincrement field if there is no value set, otherwise it wil get inserted with the given value.
//  check if fields has privilege "w"
// TODO improvements, HasMany and ManyToMany are creating a sql statement for each entry - batch?
func (e EagerLoading) Create(m Interface) error {
	var err error

	// handling belongsTo relations before the main entry
	for field, relation := range m.Table().Relations(m.whiteBlacklist(), WRITEDB) {

		if relation.Type == BelongsTo {

			// initialize related model and create the entry
			rel, errTmp := initRelation(m, field)
			if errTmp != nil {
				err = errTmp
				return err
			}

			// set white- blacklist from parent
			rel.setWhiteBlacklist(RelationWhiteBlackList(m.whiteBlacklist(), field))
			rel.setLoopMap(m.getLoopMap())
			rel.setParent(m)

			// create new entry if primary fields are empty, otherwise update entry
			if checkPrimaryFieldsEmpty(rel) {
				err = rel.Create()
			} else {
				err = rel.Update()
			}

			if err != nil {
				return err
			}

			// set related id to the parent model
			reflectField(m, relation.StructTable.StructField).Set(reflectField(rel, relation.AssociationTable.StructField))
		}

		continue
	}

	// get the struct variable for the scan
	var values []map[string]interface{}
	value := map[string]interface{}{}
	var cols []string
	var autoincrement *Column

	for _, col := range m.Table().Columns(WRITEDB) {

		// skipping autoincrement fields if no value is set
		if col.Information.Autoincrement && isZeroOfUnderlyingType(reflect.ValueOf(m).Elem().FieldByName(col.StructField).Interface()) {
			autoincrement = col
			continue
		}

		value[col.Information.Name] = reflectField(m, col.StructField).Interface()
		cols = append(cols, col.Information.Name)
	}

	values = append(values, value)

	// build insert
	b := m.Table().Builder
	schema := ""
	//if b.Config().Driver() == "postgres" {
	//	schema = ".public"
	//}

	insert := b.Insert(m.Table().Database + schema + "." + m.Table().Name).Columns(cols...).Values(values)
	if autoincrement != nil {
		insert.LastInsertedID(autoincrement.Information.Name, reflectField(m, autoincrement.StructField).Addr().Interface())
	}
	_, err = insert.Exec()
	if err != nil {
		return err
	}

	// handle all other relations
	for field, relation := range m.Table().Relations(m.whiteBlacklist(), WRITEDB) {

		switch relation.Type {
		case HasOne, CustomStruct:
			rel, errTmp := initRelation(m, field)
			if errTmp != nil {
				err = errTmp
				return err
			}
			// set white- blacklist from parent
			rel.setWhiteBlacklist(RelationWhiteBlackList(m.whiteBlacklist(), field))
			rel.setLoopMap(m.getLoopMap())
			rel.setParent(m)

			// set parent ID
			reflect.Indirect(reflectField(m, field)).FieldByName(relation.AssociationTable.StructField).Set(reflectField(m, relation.StructTable.StructField))
			// create the entry
			err = rel.Create()
			if err != nil {
				return err
			}
		case HasMany, CustomSlice:
			// TODO atm it creates a insert for one entry - performance wise combine it into multiple insert - problem i need a good solution for the last insert id
			slice := reflectField(m, field)
			for i := 0; i < slice.Len(); i++ {
				// get related struct
				var r Interface
				if slice.Index(i).Kind() == reflect.Ptr {
					r = slice.Index(i).Interface().(Interface)
				} else {
					r = slice.Index(i).Addr().Interface().(Interface)
				}

				// set parent ID
				reflectField(r, relation.AssociationTable.StructField).Set(reflectField(m, relation.StructTable.StructField))

				// add cache to related model
				cache, ttl, errTmp := m.Cache()
				if errTmp != nil {
					err = errTmp
					return err
				}

				if !r.HasCache() {
					_ = r.SetCache(cache, ttl)
				}

				// init related model
				err = r.Initialize(r)
				if err != nil {
					return err
				}
				// set white- blacklist from parent
				r.setWhiteBlacklist(RelationWhiteBlackList(m.whiteBlacklist(), field))
				r.setLoopMap(m.getLoopMap())
				r.setParent(m)

				// create the entries
				err = r.Create()
				if err != nil {
					return err
				}
			}
		case ManyToMany, ManyToManySR:
			// TODO atm it creates a insert for one entry - performance wise combine it into multiple insert - problem i need a good solution for the last insert id
			// insert in relation table only if PRIMARY is empty
			var ids []int64 // TODO fix whats with uuids or none sequence?
			slice := reflectField(m, field)
			for i := 0; i < reflectField(m, field).Len(); i++ {

				var r Interface
				if slice.Index(i).Kind() == reflect.Ptr {
					r = slice.Index(i).Interface().(Interface)
				} else {
					r = slice.Index(i).Addr().Interface().(Interface)
				}
				cache, ttl, errTmp := m.Cache()
				if errTmp != nil {
					err = errTmp
					return err
				}

				if !r.HasCache() {
					_ = r.SetCache(cache, ttl)
				}

				err = r.Initialize(r)
				if err != nil {
					return err
				}
				// set white- blacklist from parent
				r.setWhiteBlacklist(RelationWhiteBlackList(m.whiteBlacklist(), field))
				r.setLoopMap(m.getLoopMap())
				r.setParent(m)

				// if primary fields are empty, create entry otherwise ignore
				if checkPrimaryFieldsEmpty(r) {
					err = r.Create()
					if err != nil {
						return err
					}
				}

				// add last inserted id for junction table
				ids = append(ids, reflectField(r, relation.AssociationTable.StructField).Int())
			}

			if len(ids) > 0 {
				// add values
				var val []map[string]interface{}
				for _, associationId := range ids {
					val = append(val, map[string]interface{}{relation.JunctionTable.StructColumn: reflectField(m, relation.StructTable.StructField).Interface(), relation.JunctionTable.AssociationColumn: associationId})
				}

				// insert into junction table
				_, err = b.Insert(relation.JunctionTable.Table).
					Columns(relation.JunctionTable.StructColumn, relation.JunctionTable.AssociationColumn).
					Values(val).Exec()
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Update an entry by the given struct value.
// Relations:
// HasOne,BelongsTo - Creates the entry if pkeys are empty or updates the entry if pkeys are ok
// HasMany - Deletes everything and adds creates for each entry again.
// ManyToMany - Create entries in the association table if the primary keys are empty otherwise dont do anything. Delete everything from junction table and enter the ids again.
// TODO HasMany creates a new statement for each entry - batch
// TODO ManyToMayn logic is the same as in create, DRY
// TODO validate struct before update
// TODO update pkey not possible atm!
func (e EagerLoading) Update(m Interface, c *sqlquery.Condition) error {
	var err error

	// handling belongsTo relations before the main entry
	for field, relation := range m.Table().Relations(m.whiteBlacklist(), WRITEDB) {

		if relation.Type == BelongsTo {

			// initialize related model and create the entry
			rel, errTmp := initRelation(m, field)
			if errTmp != nil {
				err = errTmp
				return errTmp
			}
			// set white- blacklist from parent
			rel.setWhiteBlacklist(RelationWhiteBlackList(m.whiteBlacklist(), field))
			rel.setLoopMap(m.getLoopMap())
			rel.setParent(m)

			// create new entry if primary fields are empty, otherwise update entry
			if checkPrimaryFieldsEmpty(rel) {
				err = rel.Create()
			} else {
				err = rel.Update()
			}
			if err != nil {
				return err
			}

			// set related id to the parent model
			reflectField(m, relation.StructTable.StructField).Set(reflectField(rel, relation.AssociationTable.StructField))
		}
		continue
	}

	// then the main entry
	b := m.Table().Builder

	// set value
	value := map[string]interface{}{}

	for _, col := range m.Table().Columns(WRITEDB) {
		value[col.Information.Name] = reflectField(m, col.StructField).Interface()
	}

	// update only if there are some changes in the main table otherwise ignore
	if len(value) > 0 {
		// exec
		update := b.Update(m.Table().Name).Condition(c).Columns(m.Table().columnNames(WRITEDB, false)...).Set(value)
		res, errTmp := update.Exec()
		if errTmp != nil {
			err = errTmp
			return err
		}

		// checking if rows were effected
		i, errTmp := res.RowsAffected()
		if errTmp != nil {
			err = errTmp
			return errTmp
		}
		if i == 0 {
			//stmt, args, _ := update.String()
			//err = fmt.Errorf(ErrUpdateZero.Error()+" - "+strings.Replace(stmt, b.Placeholder.Char, "%v", 1), args...)
			//return err
		}
	}

	// handling relations
	for field, relation := range m.Table().Relations(m.whiteBlacklist(), WRITEDB) {
		switch relation.Type {
		case HasOne, CustomStruct:
			rel, errTmp := initRelation(m, field)
			if errTmp != nil {
				err = errTmp
				return errTmp
			}
			// set white- blacklist from parent
			rel.setWhiteBlacklist(RelationWhiteBlackList(m.whiteBlacklist(), field))
			rel.setLoopMap(m.getLoopMap())
			rel.setParent(m)

			// add reference
			reflect.Indirect(reflectField(m, field)).FieldByName(relation.AssociationTable.StructField).Set(reflectField(m, relation.StructTable.StructField))

			if relation.Type == CustomStruct {
				//err = rel.Update()
				//fmt.Println(relation.Type, field, reflect.TypeOf(rel.Table().strategy), err)

				return rel.Update()
			}

			// this is needed to indicate if the entry has to get created or updated.
			// logic, if all primary keys and association keys have a none zero value, its a update otherwise it will get created.
			newEntry := checkPrimaryFieldsEmpty(rel) || isZeroOfUnderlyingType(reflect.Indirect(reflectField(m, field)).FieldByName(relation.AssociationTable.StructField).Interface())

			// create new entry if primary fields are empty, otherwise update entry
			if newEntry {
				//TODO Delete old entries if exist... just to be sure: If someone just creates a new struct with an ID and adds some data without doing a First() or All() before
				err = rel.Create()
			} else {
				err = rel.Update()
			}
			if err != nil {
				return err
			}
		case HasMany, CustomSlice: // TODO at the moment for each entry there is a sql statement.

			if relation.Type != CustomSlice {
				// delete all entries
				// TODO not working with depth > 1 relations ....
				_, err = b.Delete(relation.AssociationTable.Information.Table).Where(relation.AssociationTable.Information.Name+" = ?", reflectField(m, relation.StructTable.StructField).Interface()).Exec()
				if err != nil {
					return err
				}
			}

			/* Would work but i need a snapshot first... to identify the old entries
			// loop over all entries
			rel := reflectField(m, field)
			for i := 0; i < rel.Len(); i++ {
				var entry Interface
				if rel.Index(i).Kind() == reflect.Ptr {
					entry = rel.Index(i).Interface().(Interface)
				} else {
					entry = rel.Index(i).Addr().Interface().(Interface)
				}
				// initialize model
				err = entry.Initialize(entry)
				if err != nil {
					return err
				}
				entry.SetTx(m.Tx())
				if !checkPrimaryFieldsEmpty(entry){
					err = entry.Delete()
					if err != nil{
						return err
					}
				}
			}*/

			// loop over all entries
			rel := reflectField(m, field)
			for i := 0; i < rel.Len(); i++ {

				var entry Interface
				if rel.Index(i).Kind() == reflect.Ptr {
					entry = rel.Index(i).Interface().(Interface)
				} else {
					entry = rel.Index(i).Addr().Interface().(Interface)
				}

				// add cache to model
				cache, ttl, errTmp := m.Cache()
				if errTmp != nil {
					err = errTmp
					return err
				}

				if !entry.HasCache() {
					_ = entry.SetCache(cache, ttl)
				}

				// initialize model
				err = entry.Initialize(entry)
				if err != nil {
					return err
				}
				// set white- blacklist from parent
				entry.setWhiteBlacklist(RelationWhiteBlackList(m.whiteBlacklist(), field))
				entry.setLoopMap(m.getLoopMap())
				entry.setParent(m)
				entry.callback().setMode("Update")

				// adding parent id to relation field
				reflectField(entry, relation.AssociationTable.StructField).Set(reflectField(m, relation.StructTable.StructField))

				err = entry.Create()
				if err != nil {
					return err
				}
			}

		case ManyToMany, ManyToManySR:
			// MANY-TO-MANY: Delete * from junction, add secondary (create or update), add to junction again

			// delete all entries from junction table
			_, err = b.Delete(relation.JunctionTable.Table).Where(relation.JunctionTable.StructColumn+" = ?", reflectField(m, relation.StructTable.StructField).Interface()).Exec()
			if err != nil {
				return err
			}

			// TODO all after here same logic as create - combine?????

			// loop over all entries
			rel := reflectField(m, field)
			var ids []int64 // TODO fix whats with uuids or none sequence?
			for i := 0; i < rel.Len(); i++ {

				var entry Interface
				if rel.Index(i).Kind() == reflect.Ptr {
					entry = rel.Index(i).Interface().(Interface)
				} else {
					entry = rel.Index(i).Addr().Interface().(Interface)
				}

				// add cache to model
				cache, ttl, errTmp := m.Cache()
				if errTmp != nil {
					err = errTmp
					return errTmp
				}

				if !entry.HasCache() {
					_ = entry.SetCache(cache, ttl)
				}

				// initialize model
				err = entry.Initialize(entry)
				if err != nil {
					return err
				}
				// set white- blacklist from parent
				entry.setWhiteBlacklist(RelationWhiteBlackList(m.whiteBlacklist(), field))
				entry.setLoopMap(m.getLoopMap())
				entry.setParent(m)

				// if primary fields are empty, create entry in association table otherwise ignore
				if checkPrimaryFieldsEmpty(entry) {
					err = entry.Create()
					if err != nil {
						return err
					}
				}

				// get id of association entry
				ids = append(ids, reflectField(entry, relation.AssociationTable.StructField).Int())
			}

			// prepare data for junction table - add values
			var val []map[string]interface{}
			for _, associationId := range ids {
				val = append(val, map[string]interface{}{relation.JunctionTable.StructColumn: reflectField(m, relation.StructTable.StructField).Interface(), relation.JunctionTable.AssociationColumn: associationId})
			}

			// insert into junction table
			if len(val) > 0 {
				_, err = b.Insert(relation.JunctionTable.Table).Columns(relation.JunctionTable.StructColumn, relation.JunctionTable.AssociationColumn).Values(val).Exec()
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Delete deletes the entry by its primary key
// SoftDelete is happening on model base.
// Relation:
// BelongsTo - will be ignored
// HasOne, HasMany are deleted by there reference
// ManyToMany - deletes only the junction table. If you have to delete the associated table, use cascade in the db.
func (e EagerLoading) Delete(m Interface, c *sqlquery.Condition) error {
	var err error

	// then the main entry
	b := m.Table().Builder

	// handle relations first
	for field, relation := range m.Table().Associations {

		switch relation.Type {
		case CustomStruct:
			relModel, err := initRelation(m, field)
			if err != nil {
				return err
			}
			err = relModel.Delete()
			if err != nil {
				return err
			}
		case CustomSlice:

			// loop over all entries
			rel := reflectField(m, field)
			for i := 0; i < rel.Len(); i++ {

				var entry Interface
				if rel.Index(i).Kind() == reflect.Ptr {
					entry = rel.Index(i).Interface().(Interface)
				} else {
					entry = rel.Index(i).Addr().Interface().(Interface)
				}

				// add cache to model
				cache, ttl, errTmp := m.Cache()
				if errTmp != nil {
					err = errTmp
					return err
				}

				if !entry.HasCache() {
					_ = entry.SetCache(cache, ttl)
				}

				// initialize model
				err = entry.Initialize(entry)
				if err != nil {
					return err
				}
				entry.setLoopMap(m.getLoopMap())
				entry.setParent(m)

				err = entry.Delete()
				if err != nil {
					return err
				}

			}
		case BelongsTo:
			// ignore - belongsTo - stays untouched
		case HasOne:
			// hasOne - delete - ignore softDelete if the main struct has none.
			_, err = b.Delete(relation.AssociationTable.Information.Table).Where(relation.AssociationTable.Information.Name+" = ?", reflectField(m, relation.StructTable.StructField).Interface()).Exec()
			if err != nil {
				return err
			}
		case HasMany:
			// hasMany - delete - ignore softDelete if the main struct has none.
			_, err = b.Delete(relation.AssociationTable.Information.Table).Where(b.QuoteIdentifier(relation.AssociationTable.Information.Name)+" = ?", reflectField(m, relation.StructTable.StructField).Interface()).Exec()
			if err != nil {
				return err
			}
		case ManyToMany, ManyToManySR:
			// hasManyToMany - only junction table entries are getting deleted - for the association table use SQL CASCADE or a callbacks
			_, err = b.Delete(relation.JunctionTable.Table).Where(relation.JunctionTable.StructColumn+" = ?", reflectField(m, relation.StructTable.StructField).Interface()).Exec()
			if err != nil {
				return err
			}
		}
	}

	// exec
	deleteSql := b.Delete(m.Table().Name).Condition(c)
	res, errTmp := deleteSql.Exec()
	if errTmp != nil {
		err = errTmp
		return err
	}

	// checking if rows were effected
	i, errTmp := res.RowsAffected()
	if errTmp != nil {
		err = errTmp
		return err
	}
	if i != 1 {
		stmt, args, _ := deleteSql.String()
		err = fmt.Errorf(ErrDeleteNotFound.Error()+" - "+strings.Replace(stmt, b.Driver().Placeholder().Char, "%v", 1), args...)
		return err
	}

	return nil
}

// setValue is a helper to set the value correct to the underlying type.
func setValue(field reflect.Value, value reflect.Value) {

	switch field.Kind() {
	case reflect.Ptr:
		field.Set(value.Addr())
	case reflect.Struct:
		field.Set(value)
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.Ptr {
			field.Set(reflect.Append(field, value.Addr()))
		} else {
			field.Set(reflect.Append(field, value))
		}
	}
}

// inSlice is a helper to check if the value is already in the slice.
func inSlice(search int, list []int) bool {
	for _, v := range list {
		if v == search {
			return true
		}
	}
	return false
}

// isZeroOfUnderlyingType checks if the given value is a zero value.
func isZeroOfUnderlyingType(x interface{}) bool {
	return x == reflect.Zero(reflect.TypeOf(x)).Interface()
}

// initRelation is initializing a struct field which implements the (model) Interface.
// TODO check if already init?
func initRelation(m Interface, field string) (Interface, error) {
	f := reflectField(m, field)
	var relation Interface
	var err error

	switch f.Kind() {
	case reflect.Ptr:

		if reflect.ValueOf(f.Interface().(Interface)).IsNil() {
			f.Set(newValueInstanceFromType(reflectField(m, field).Type()).Addr())
		}

		relation = f.Interface().(Interface)
	case reflect.Struct:
		relation = f.Addr().Interface().(Interface)
	case reflect.Slice:
		field, _ := reflect.TypeOf(reflect.ValueOf(m).Elem().Interface()).FieldByName(field)
		relation = newValueInstanceFromType(field.Type).Addr().Interface().(Interface)
	}

	if relation != nil {
		c, d, errTmp := m.Cache()
		if errTmp != nil {
			err = errTmp
			return nil, err
		}
		if !relation.HasCache() {
			_ = relation.SetCache(c, d)
		}
		err = relation.Initialize(relation)
	}
	return relation, err
}

package orm

import (
	"fmt"
	"github.com/jinzhu/inflection"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/serenize/snaker"
	"reflect"
	"strings"
	"unsafe"
)

// getRelationByTag returns the given relation tag type.
// If no relation tag is set, it will return an empty string.
// If the type is different than hasOne,belongsTo,hasMany or manyToMany, an error will return.
func getRelationByTag(tag string) (string, error) {
	// User defined relation type
	switch tag {
	case HasOne, BelongsTo, HasMany, ManyToMany:
		return tag, nil
	case "":
		return "", nil
	default:
		return "", fmt.Errorf(ErrRelationType.Error(), tag)
	}
}

// getRelationByType is checking the relation type by struct type.
// Ptr or struct are defined as hasOne or belongsTo - it checks the foreign key to detect which relations it is.
// A slice is defined as hasMany or manyToMany - it checks if a junction table exists to detect which relation it is.
func getRelationByType(mainModel Interface, relationModel Interface, rel reflect.StructField) (string, error) {

	if rel.Type.Kind() == reflect.Ptr || rel.Type.Kind() == reflect.Struct {

		// checking DB hasOne
		fk, err := getForeignKey(relationModel, mainModel, rel)
		if err != nil {
			// return "", err //TODO only allow field not found errors here
		}

		if fk != nil {
			// checking if manually a fk was set which detects a belongsTo
			if fk.Name == BelongsTo {
				return BelongsTo, nil
			}
			return HasOne, nil
		}

		// checking DB belongsTo
		fk, err = getForeignKey(mainModel, relationModel, rel)
		if err != nil {
			return "", err
		}
		if fk != nil {
			return BelongsTo, nil
		}

		return HasOne, nil // TODO here should be an error - not detected????
	}

	if rel.Type.Kind() == reflect.Slice {
		if hasManyToMany(mainModel, relationModel) {
			return ManyToMany, nil
		}
		return HasMany, nil
	}

	return "", nil // TODO here should be an error - not detected????
}

// hasManyToMany checks if a junction table exists for the two given models (Interface).
// The struct names of the models are getting combined to a plural snakestyle ex: UserPost = checking table name -> user_posts
// Returns true if the table exists.
func hasManyToMany(mainModel Interface, relationModel Interface) bool {
	// TODO create a tag to define junction table name + fk
	junctionTable := snaker.CamelToSnake(inflection.Plural(structName(mainModel, false) + structName(relationModel, false)))
	cols, _ := mainModel.Table().Builder.Information(junctionTable).Describe()
	if len(cols) > 0 {
		return true
	}
	return false
}

// getManyToMany returns the foreign keys of the relations.
// At the moment no junctionTable name or FK can be set by Tag.
// Returns an error if field does not exist in struct or table does not exist (builder).
func getManyToMany(mainModel Interface, relationModel Interface) ([]*sqlquery.ForeignKey, error) {
	junctionTable := snaker.CamelToSnake(inflection.Plural(structName(mainModel, false) + structName(relationModel, false)))
	_, err := mainModel.Table().Builder.Information(junctionTable).Describe()
	if err != nil {
		return nil, err
	}
	var fks []*sqlquery.ForeignKey

	// check if its a self-reference
	if structName(mainModel, true) == structName(relationModel, true) {

		fKeys, err := mainModel.Table().Builder.Information(junctionTable).ForeignKeys()
		if err != nil {
			return nil, err
		}

		for _, fk := range fKeys {
			if fk.Secondary.Table == mainModel.Table().Name {
				fks = append(fks, fk)
			}
		}

		return fks, nil
	}

	// junction table - relation model
	fk2, err := getForeignKeyByDb(relationModel, mainModel, junctionTable)

	if err != nil {
		return nil, err
	}
	fks = append(fks, fk2)

	// junction table - main model
	fk, err := getForeignKeyByDb(mainModel, relationModel, junctionTable)
	if err != nil {
		return nil, err
	}
	fks = append(fks, fk)

	return fks, nil
}

// getRelation combines getRelationByTag and getRelationByType.
// Tag has higher priority than the automatic logic.
// Returns the errors of getRelationBy or getRelationByType.
func getRelation(mainModel Interface, relationModel Interface, relation reflect.StructField) (string, error) {

	// checking if a relation is defined by tag
	rel, err := getRelationByTag(relation.Tag.Get(TagRelation))
	if err != nil {
		return "", err
	}
	if rel != "" {
		return rel, nil
	}

	// automatic logic
	return getRelationByType(mainModel, relationModel, relation)
}

// getForeignKeyByTag return the given foreignKey tag.
// If no fk tag is set, a nil pointer will return.
// If the defined field is not found in the underlying struct, an error will return.
// If the FK is set by TAG the name of the foreign key will be "tag" as identifier for a user added fks.
func getForeignKeyByTag(mainModel Interface, relationModel Interface, relation reflect.StructField) (*sqlquery.ForeignKey, error) {

	// if tag is empty
	fk := relation.Tag.Get(TagFK)
	fk = strings.TrimSpace(fk)
	if fk == "" {
		return nil, nil
	}

	// declarations
	foreignKey := &sqlquery.ForeignKey{}

	// if only a single field is defined in the tag
	if !strings.Contains(fk, TagSeparator) {

		rel, err := getRelationByTag(relation.Tag.Get(TagRelation))
		if err != nil {
			return nil, err
		}

		// checking field has a belongsTo relation or by the field definition
		if rel == BelongsTo || (fieldExists(mainModel, structName(relationModel, false)+fk) && fieldExists(relationModel, fk)) { //TODO create tests for belongsTo TAG
			foreignKey.Name = BelongsTo
			col, err := getColumnNameFromField(mainModel, structName(relationModel, false)+fk)
			if err != nil {
				return nil, err
			}
			relCol, err := getColumnNameFromField(relationModel, fk)
			if err != nil {
				return nil, err
			}
			foreignKey.Primary = sqlquery.Relation{Table: mainModel.Table().Name, Column: col}
			foreignKey.Secondary = sqlquery.Relation{Table: relationModel.Table().Name, Column: relCol}
			return foreignKey, nil
		}

		// definition for all other relation types (hasOne, hasMany, manyToMany)
		foreignKey.Name = "tag"
		col, err := getColumnNameFromField(mainModel, fk)
		if err != nil {
			return nil, err
		}

		relCol, err := getColumnNameFromField(relationModel, structName(mainModel, false)+fk)
		if err != nil {
			// if its a custom type, check the StructField instead of Table.Cols.
			relCol = structName(mainModel, false) + fk
			if !reflect.Indirect(reflect.ValueOf(relationModel)).FieldByName(relCol).IsValid() {
				return nil, err
			}
		}
		foreignKey.Primary = sqlquery.Relation{Table: mainModel.Table().Name, Column: col}
		tn := "" // needed for custom types - set an empty Table name otherwise stategy will break
		if relationModel.Table() != nil {
			tn = relationModel.Table().Name
		}
		foreignKey.Secondary = sqlquery.Relation{Table: tn, Column: relCol}
		return foreignKey, nil

	}

	parsedFK, err := parseTags(fk)
	if err != nil {
		return nil, err
	}

	col, err := getColumnNameFromField(mainModel, parsedFK["field"])
	if err != nil {
		return nil, err
	}

	relCol, err := getColumnNameFromField(relationModel, parsedFK["associationField"])
	if err != nil {
		// if its a custom type, check the StructField instead of Table.Cols.
		relCol = parsedFK["associationField"]
		if !reflect.Indirect(reflect.ValueOf(relationModel)).FieldByName(relCol).IsValid() {
			return nil, err
		}
	}

	foreignKey.Name = "tag"
	if strings.HasPrefix(parsedFK["field"], structName(relationModel, false)) {
		foreignKey.Name = BelongsTo //TODO create tests for belongsTo TAG
	}
	foreignKey.Primary = sqlquery.Relation{Table: mainModel.Table().Name, Column: col}
	tn := "" // needed for custom types - set an empty Table name otherwise stategy will break
	if relationModel.Table() != nil {
		tn = relationModel.Table().Name
	}
	foreignKey.Secondary = sqlquery.Relation{Table: tn, Column: relCol}
	return foreignKey, nil

}

// getForeignKeyByDb checks if there is a foreign key for the given models (Interface) defined.
// A third parameter can be used for other table names like to check a junction table for example.
// It will return an error if the struct field does not exist. (checked in both models (Interfaces)).
// Will return no error if the foreign key was not found - is done in addRelation for a better error message.
func getForeignKeyByDb(mainModel Interface, relationModel Interface, specialTable string) (*sqlquery.ForeignKey, error) {

	db, mainTable := mainModel.Table().Database, mainModel.Table().Name
	relationTable := relationModel.Table().Name

	tableName := db + "." + mainTable
	var st string
	if specialTable != "" {
		tableName = specialTable
		_, st = getDatabaseAndTableByString(specialTable) //split db and table if a . is given
	}

	fKeys, err := mainModel.Table().Builder.Information(tableName).ForeignKeys()
	if err != nil {
		return nil, err
	}

	for _, fKey := range fKeys {
		if (fKey.Primary.Table == mainTable && fKey.Secondary.Table == relationTable) || (fKey.Primary.Table == st && fKey.Secondary.Table == relationTable) {

			if relationTable == fKey.Secondary.Table {

				if specialTable == "" && !columnExists(mainModel, fKey.Primary.Column) { // on junctionTable this is not necessary because its called twice and only fk.Secondary is important to check!
					return nil, fmt.Errorf(ErrModelFieldNotFound.Error(), fKey.Primary.Column, structName(mainModel, true))
				}

				if !columnExists(relationModel, fKey.Secondary.Column) {
					return nil, fmt.Errorf(ErrModelFieldNotFound.Error(), fKey.Secondary.Column, structName(relationModel, true))
				}
			}

			return fKey, nil
		}
	}

	return nil, nil
}

// getForeignKey combines getForeignKeyByTag and getForeignKeyByDb.
// Will return errors of getForeignKeyByTag / getForeignKeyByDb.
func getForeignKey(mainModel Interface, relationModel Interface, relation reflect.StructField) (*sqlquery.ForeignKey, error) {
	// checking fk tag
	fk, err := getForeignKeyByTag(mainModel, relationModel, relation)
	if err != nil {
		return nil, err
	}
	if fk != nil {
		return fk, nil
	}

	// returning defined logic
	return getForeignKeyByDb(mainModel, relationModel, "")
}

// addRelation adds a relation to the table.Associations map.
// The key is the struct field name.
// Will return an errors of methods getForeignKey / getManyToMany.
func (m *Model) addRelation(model Interface) error {
	//TODO foreignKey was maybe already called before in getRelationByType (m2m) - performance improvements multiple describe.
RelationLoop:
	for _, relation := range m.getFieldsOrRelations(model, true) {
		// check if its a struct relation loop
		// TODO fix: example Customer -> Order -> Customer. Customer hasOne Order and Order belongsTo Customer. If its getting initialized like this. the next time you initialize Order, the belongsTo relation is missing.
		if m.isStructLoop(relation.Type.String()) {
			continue
		}

		// create new instance and initialize it
		relationModel := newValueInstanceFromType(relation.Type)

		// checking if its a self-reference
		if structName(relationModel.Interface(), true) == structName(model, true) {
			f := relationModel.FieldByName("skipRel")
			f = reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
			f.Set(reflect.ValueOf(true))
		}

		// checking if its a custom implementation
		relationModelExec := relationModel.Addr().Interface().(Interface)
		customImp := relationModelExec.Custom()
		tags, err := parseTags(relation.Tag.Get(TagName))
		if err != nil {
			return err
		}
		for k := range tags {
			switch k {
			case "custom":
				customImp = true
			}
		}
		if customImp {
			custType := CustomStruct
			if relation.Type.Kind() == reflect.Slice {
				custType = CustomSlice
			}

			fk, err := getForeignKeyByTag(model, relationModel.Addr().Interface().(Interface), relation)
			if err != nil {
				return err
			}
			var structCol *Column
			var associationCol *Column

			if fk == nil {
				//Default logic fields must exist = Main Struct "ID", Child Struct "NameOfMainStructID"

				// structCol always exists, because the ID field is mandatory in the orm.
				structCol = &Column{}
				structCol.StructField = "ID"
				structCol.Information = &sqlquery.Column{Name: "ID"}

				// checking if the Field "NameOfMainStructID" exists.
				associationCol = &Column{}
				associationCol.StructField = structName(model, false) + "ID"
				associationCol.Information = &sqlquery.Column{Name: associationCol.StructField}
				f := relationModel.FieldByName(associationCol.StructField)
				if !f.IsValid() {
					return fmt.Errorf(ErrForeignKeyNotFound.Error(), custType, relation.Name, structName(model, true))
				}
			} else {
				structCol, err = m.table.columnByName(fk.Primary.Column)
				if err != nil {
					return err
				}

				associationCol = &Column{}
				associationCol.StructField = fk.Secondary.Column
				associationCol.Information = &sqlquery.Column{Name: fk.Secondary.Column}
			}

			m.table.Associations[relation.Name] = &Association{Type: custType, StructTable: structCol, AssociationTable: associationCol}

			err = relationModel.Addr().Interface().(Interface).SetStrategy("custom")
			if err != nil {
				return err
			}

			err = m.initializeModelByValue(relationModel)
			if err != nil {
				return err
			}

			continue RelationLoop
		}

		// initialize relation model
		err = m.initializeModelByValue(relationModel)
		if err != nil {
			return err
		}

		rel, err := getRelation(model, relationModel.Addr().Interface().(Interface), relation)
		if err != nil {
			return err
		}

		switch rel {
		case HasOne, HasMany:

			fk, err := getForeignKey(relationModel.Addr().Interface().(Interface), model, relation)
			if err != nil {
				return err
			}
			if fk == nil {
				return fmt.Errorf(ErrForeignKeyNotFound.Error(), rel, relation.Name, structName(model, true))
			}

			structCol, err := m.table.columnByName(fk.Secondary.Column)
			if err != nil {
				return err
			}

			associationCol, err := relationModel.Addr().Interface().(Interface).Table().columnByName(fk.Primary.Column)
			if err != nil {
				return err
			}

			m.table.Associations[relation.Name] = &Association{Type: rel, StructTable: structCol, AssociationTable: associationCol}

		case BelongsTo:
			fk, err := getForeignKey(model, relationModel.Addr().Interface().(Interface), relation)
			if err != nil {
				return err
			}
			if fk == nil {
				return fmt.Errorf(ErrForeignKeyNotFound.Error(), BelongsTo, relation.Name, structName(model, true))
			}

			structCol, err := m.table.columnByName(fk.Primary.Column)
			if err != nil {
				return err
			}

			associationCol, err := relationModel.Addr().Interface().(Interface).Table().columnByName(fk.Secondary.Column)
			if err != nil {
				return err
			}

			m.table.Associations[relation.Name] = &Association{Type: BelongsTo, StructTable: structCol, AssociationTable: associationCol}

		case ManyToMany:
			fks, err := getManyToMany(model, relationModel.Addr().Interface().(Interface))
			if err != nil {
				return err
			}

			if len(fks) != 2 || fks[0] == nil || fks[1] == nil {
				return fmt.Errorf(ErrForeignKeyNotFound.Error(), ManyToMany, relation.Name, structName(model, true))
			}
			jT := &JunctionTable{Table: fks[0].Primary.Table, StructColumn: fks[0].Primary.Column, AssociationColumn: fks[1].Primary.Column}

			structCol, err := m.table.columnByName(fks[0].Secondary.Column)
			if err != nil {
				return err
			}

			associationCol, err := relationModel.Addr().Interface().(Interface).Table().columnByName(fks[1].Secondary.Column)
			if err != nil {
				return err
			}
			tpe := ManyToMany
			if structCol.Information.Table == associationCol.Information.Table {
				tpe = ManyToManySR
			}
			m.table.Associations[relation.Name] = &Association{Type: tpe, StructTable: structCol, AssociationTable: associationCol, JunctionTable: jT}
		}
	}
	return nil
}

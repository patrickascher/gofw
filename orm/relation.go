package orm

import (
	"errors"
	"fmt"
	"reflect"
	strings2 "strings"

	"github.com/patrickascher/gofw/strings"
)

// available relations.
const (
	HasOne     = "hasOne"
	BelongsTo  = "belongsTo"
	HasMany    = "hasMany"
	ManyToMany = "m2m"
)

// struct tag definition.
const (
	tagRelation                  = "relation"
	tagPolymorphic               = "polymorphic"
	tagPolymorphicValue          = "polymorphic_value"
	tagForeignKey                = "fk"
	tagAssociationForeignKey     = "afk"
	tagJoinTable                 = "join_table"
	tagJoinForeignKey            = "join_fk"
	tagJoinAssociationForeignKey = "join_afk"
)

// error messages
var (
	errStructField           = "orm: field %s does not exist in %s"
	errSelfReference         = errors.New("orm: self reference is only allowed on many to many")
	errRelationKind          = "orm: %s.%s relation kind %s is not allowed on %s"
	errJoinFK                = "orm: fk field %s or association fk field %s does not exit in table %s"
	errPolymorphicNotAllowed = "orm: %s - polymorphic is not allowed on relation type %s"
	errBackReference         = "orm: back-reference fields %s must be a %s in %s"
)

// JoinTable holds the information for the m2m junction table.
// The name, foreignKey and associationForeignKey are the direct column names of the db table.
type JoinTable struct {
	Name                  string
	ForeignKey            string
	AssociationForeignKey string
}

// Polymorphic is available for hasOne and hasMany relationships.
type Polymorphic struct {
	Field Field
	Type  Field
	Value string
}

// Relation keeps some information about the relation type and connection fields.
type Relation struct {
	// Type for easier reflection
	Type reflect.Type
	// Kind of the relation (hasOne, hasMany,...)
	Kind string
	// The struct field name
	Field string
	// FK of the main struct
	ForeignKey Field
	// AFK of the child struct
	AssociationForeignKey Field
	// Self referencing relationship
	SelfReference bool
	// Custom struct without orm.Model embedded
	Custom bool

	Permission  Permission
	Validator   *validator
	Polymorphic Polymorphic
	JoinTable   JoinTable
}

// foreignKey of the model by tag.
// If the tag is empty, the first primary field will be taken.
// Error will return if the field does not exist.
// Primary field always exists (already checked in createFields).
// TODO change if more than one pk are allowed
func (m *Model) foreignKey(tag string) (Field, error) {
	scope := Scope{model: m}

	if tag != "" {
		f, err := scope.Field(tag)
		if err != nil {
			return Field{}, err
		}

		return *f, nil
	}

	return scope.PrimaryKeys()[0], nil
}

// polymorphicRestriction - if there are polymorphic tags set on belongsTo,M2M an error will return.
func polymorphicRestriction(tags map[string]string, field string, rel string) error {
	if _, ok := tags[tagPolymorphic]; ok {
		return fmt.Errorf(errPolymorphicNotAllowed, field, rel)
	}
	if _, ok := tags[tagPolymorphicValue]; ok {
		return fmt.Errorf(errPolymorphicNotAllowed, field, rel)

	}
	return nil
}

// polymorphic of the relation by tag.
// If a polymorphic is defined, it will check if the needed fields exist.
// Error will return if {name}ID {name}Type does not exist.
// As default value the struct name is taken.
func (m Model) polymorphic(tags map[string]string, rel Interface) (Polymorphic, error) {
	if v, ok := tags[tagPolymorphic]; ok {
		// {name}ID
		f, err := rel.Scope().Field(v + "ID")
		if err != nil {
			return Polymorphic{}, err
		}
		// {name}Type
		t, err := rel.Scope().Field(v + "Type")
		if err != nil {
			return Polymorphic{}, err
		}
		// value
		val := ""
		if v, ok := tags[tagPolymorphicValue]; ok {
			val = v
		}
		if val == "" {
			val = m.modelName(false)
		}

		return Polymorphic{Field: *f, Type: *t, Value: val}, nil
	}

	return Polymorphic{}, nil
}

// associationForeignKey will return the defined a.fk by tag.
// If no tag is defined, by default the a.fk will be the parent struct {name}ID.
// Error will return if the field does not exist.
func (m Model) associationForeignKey(fieldName string, rel Interface) (Field, error) {
	if fieldName == "" {
		fieldName = m.modelName(false) + "ID"
	}

	f, err := rel.Scope().Field(fieldName)
	// because Field is now returning a ptr
	if f == nil {
		f = &Field{}
	}
	return *f, err
}

// joinTable builds a JoinTable{} by tag or default values.
// Tags "join_table, join_fk, join_afk" can be used to configure the join Table.
// If there are no tags defined, the following defaults will be used.
// 		join_table: Parent struct name + child struct name in plural (customer_addresses).
// 		join_fk: parent struct name + parent struct pkey (customer_id)
// 		join_afk: child struct name + child struct pkey (address_id)
// A check will be called if the table and fields exist in the database, otherwise an error will return.
func joinTable(m Interface, rel Interface, tags map[string]string, selfRef bool) (JoinTable, error) {
	joinTable := tags[tagJoinTable]
	joinFK := tags[tagJoinForeignKey]
	joinAFK := tags[tagJoinAssociationForeignKey]
	if joinTable == "" {
		joinTable = strings.CamelToSnake(strings.Plural(m.model().modelName(false) + rel.model().modelName(false)))
	}
	if joinFK == "" {
		fk, _ := m.model().foreignKey(tags[tagForeignKey])
		joinFK = strings.CamelToSnake(m.model().modelName(false) + fk.Name)
	}
	if joinAFK == "" {
		if !selfRef {
			afk, _ := rel.model().foreignKey(tags[tagAssociationForeignKey])
			joinAFK = strings.CamelToSnake(rel.model().modelName(false) + afk.Name)
		} else {
			joinAFK = defaultSelfReferenceAssociationForeignKey
		}
	}

	// check if the database table exists
	b := m.Scope().Builder()
	cols, err := b.Information(m.DefaultDatabaseName()+"."+joinTable).Describe(joinFK, joinAFK)
	if err != nil {
		return JoinTable{}, err
	}
	// check if the required fields exist.
	if len(cols) != 2 {
		return JoinTable{}, fmt.Errorf(errJoinFK, joinFK, joinAFK, joinTable)
	}

	return JoinTable{Name: joinTable, ForeignKey: joinFK, AssociationForeignKey: joinAFK}, nil
}

// backReferencePointer is checking if the backreference is a * or a []*.
// TODO rename the function, because this is only checking if the given field is a * or a []*.
func (m Model) backReferencePointer(relation reflect.StructField) error {
	if (relation.Type.Kind() == reflect.Struct) ||
		(relation.Type.Kind() == reflect.Slice && relation.Type.Elem().Kind() != reflect.Ptr) {
		v := "ptr"
		if relation.Type.Kind() == reflect.Slice {
			v = "[]ptr"
		}
		return fmt.Errorf(errBackReference, relation.Name, v, m.modelName(true))
	}

	return nil
}

// createRelations adds a relation to the relations map.
// * Relations will be checked against the type.
// * Relation loops will be avoided, self referencing is allowed.
// * If the fk, a.fk, join table or polymorphic field does not exist, an error will return.
// * The configured relation is added to the orm model.
// * validation has the only function to store the user added tags. Needed for grid frontend later on. The isValid function validates the whole struct anyway.
func (m *Model) createRelations() error {

	for _, relation := range m.structFields(m.caller, true) {
		// parse tags
		tags := parseTags(relation.Tag.Get(tagName))

		validator := &validator{Config: relation.Tag.Get(tagValidate)}

		// get relation kind by tag or default definition.
		rel, err := m.relationKind(tags, relation)
		if err != nil {
			return err
		}

		// custom relationship
		if _, ok := tags[tagCustom]; ok {
			m.relations = append(m.relations, Relation{Validator: validator, Kind: rel, SelfReference: false, Field: relation.Name, Custom: true})
			continue
		}

		// get an instance of the given type.
		v := newValueInstanceFromType(relation.Type)
		// check if its a self reference
		var rModel Interface
		selfRef := m.modelName(true) == v.Type().String()
		if selfRef && rel != ManyToMany {
			return errSelfReference
		}

		// if the orm model type was already loaded in a parent instance and if the fields are pointer or []*.
		if loadedRel, err := m.scope.Parent(v.Type().String()); err == nil {
			// set model instead of initialize it.
			rModel = loadedRel.caller

			// checking if both fields are * or []*
			// this block has to get rewritten when the back-referencing will be needed.
			err = m.backReferencePointer(relation)
			if err != nil {
				return err
			}
			// needed because the parent relations are not initialized yet.
			for _, r := range m.structFields(rModel, true) {
				if strings2.Replace(strings2.Replace(r.Type.String(), "[]", "", -1), "*", "", -1) == m.name {
					err = rModel.model().backReferencePointer(r)
					if err != nil {
						return err
					}
				}
			}
		} else {
			// Initialize the relation model
			// loops are avoided.
			rModel, err = m.initializeModelByValue(v)
			if err != nil {
				return err
			}
		}

		defaultPermission := Permission{Write: true, Read: true}
		if v, ok := tags[tagPermission]; ok {
			defaultPermission.Read = false
			defaultPermission.Write = false
			if strings2.Contains(v, "r") {
				defaultPermission.Read = true
			}
			if strings2.Contains(v, "w") {
				defaultPermission.Write = true
			}
		}

		// relation type switch
		switch rel {
		case HasOne, HasMany:
			// HasOne or HasMany relation.
			// FK tags are checked first, if none was found, the first primary key is used.
			// AFK tags are checked first, if none was found the field name is parent model name + ID
			// If a polymorphic is defined, the AFK will be ignored
			fk, err := m.foreignKey(tags[tagForeignKey])
			if err != nil {
				return err
			}

			poly, err := m.polymorphic(tags, rModel)
			if err != nil {
				return err
			}

			// a.fk is not needed if there is a polymorphic.
			var afk Field
			if poly.Value == "" {
				afk, err = m.associationForeignKey(tags[tagAssociationForeignKey], rModel)
				if err != nil {
					return err
				}
			}
			m.relations = append(m.relations, Relation{Validator: validator, Type: v.Type(), Permission: defaultPermission, Kind: rel, SelfReference: selfRef, Field: relation.Name, ForeignKey: fk, Polymorphic: poly, AssociationForeignKey: afk})
		case BelongsTo:
			// Polymorphic is not allowed on m2m
			err = polymorphicRestriction(tags, relation.Name, rel)
			if err != nil {
				return err
			}

			// BelongsTo relation.
			// FK tags are checked first, if none was found, the field name is child model name + ID
			// AFK tags are checked first, if none was found the first primary key is taken.
			fk, err := rModel.model().associationForeignKey(tags[tagForeignKey], m)
			if err != nil {
				return err
			}
			afk, err := rModel.model().foreignKey(tags[tagAssociationForeignKey])
			if err != nil {
				return err
			}

			m.relations = append(m.relations, Relation{Validator: validator, Type: v.Type(), Permission: defaultPermission, Kind: rel, SelfReference: selfRef, Field: relation.Name, ForeignKey: fk, AssociationForeignKey: afk})
		case ManyToMany:
			// Polymorphic is not allowed on m2m
			err = polymorphicRestriction(tags, relation.Name, rel)
			if err != nil {
				return err
			}

			// ManyToMany relation.
			//
			// Customer ID
			// FK (ID), AFK = (ID)
			// join_table = customer_addresses customer_id, address_id
			// Address ID
			fk, err := m.foreignKey(tags[tagForeignKey])
			if err != nil {
				return err
			}
			afk, err := rModel.model().foreignKey(tags[tagAssociationForeignKey])
			if err != nil {
				return err
			}

			jTable, err := joinTable(m, rModel, tags, selfRef)
			if err != nil {
				return err
			}

			m.relations = append(m.relations, Relation{Validator: validator, Type: v.Type(), Permission: defaultPermission, Kind: rel, SelfReference: selfRef, Field: relation.Name, ForeignKey: fk, AssociationForeignKey: afk, JoinTable: jTable})
		}
	}

	return nil
}

// isTagRelationAllowed checks if the given tag or default is allowed with the used struct type.
func isTagRelationAllowed(field reflect.StructField, r string) bool {

	// struct, ptr
	if field.Type.Kind() == reflect.Struct || (field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct) {
		if r == HasOne || r == BelongsTo {
			return true
		}
		return false
	}

	// []
	if field.Type.Kind() == reflect.Slice {
		if r == HasMany || r == ManyToMany {
			return true
		}
	}

	return false
}

// relationKind return the relation as string.
// An error will return if its a misconfiguration.
// hasOne,belongsTo must be a struct or a ptr to a struct. (default: struct,ptr to struct = hasOne).
// hasMany,m2m must be a slice (default: slice = hasMany).
// self referencing (default value is manyToMany).
func (m Model) relationKind(tags map[string]string, field reflect.StructField) (string, error) {

	// check if tag relation is set and valid
	if tag, ok := tags[tagRelation]; ok {
		if !isTagRelationAllowed(field, tag) {
			return "", fmt.Errorf(errRelationKind, m.name, field.Name, tag, field.Type.Kind())
		}
		return tag, nil
	}

	// default values
	if field.Type.Kind() == reflect.Struct || (field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct) {
		return HasOne, nil
	}
	if field.Type.Kind() == reflect.Slice {
		// self reference is m2m by default
		if strings2.Replace(field.Type.Elem().String(), "*", "", -1) == m.modelName(true) {
			return ManyToMany, nil
		}
		return HasMany, nil
	}

	return "", fmt.Errorf(errRelationKind, m.name, field.Name, field.Name, field.Type.Kind())
}

// implementsInterface checks if the given field type implements the orm interface.
func implementsInterface(field reflect.StructField) bool {
	i := reflect.TypeOf((*Interface)(nil)).Elem()
	v := newValueInstanceFromType(field.Type)

	return v.Addr().Type().Implements(i)
}

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

// initializeModelByValue init a reflect.Value.
// * Its checked against the already initialized relations to avoid loops.
// * The parent cache will be passed to the child model.
// * The parent initRelations will be passed to the child model.
// * The model gets initialized and an ptr to the model will be returned.
// getStructFields already checks if the relation implements the orm interface.
func (m Model) initializeModelByValue(r reflect.Value) (Interface, error) {
	rel := r.Addr().Interface().(Interface)
	rel.model().cache, rel.model().cacheTTL = m.cache, m.cacheTTL
	rel.model().parentModel = &m

	// init
	err := rel.Init(rel)
	if err != nil {
		return nil, err
	}

	return r.Addr().Interface().(Interface), nil
}

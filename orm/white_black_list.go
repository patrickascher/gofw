package orm

import (
	"reflect"
	"strings"
)

// Black and Whitelist constants
const (
	BLACKLIST = 0
	WHITELIST = 1
)

// WhiteBlackList struct
type WhiteBlackList struct {
	policy int
	list   []string
}

// NewWhiteBlackList creates a new WhiteBlackList and sets the policy and list.
// The list is a slice of string which should be the name of the struct field or relation.
func NewWhiteBlackList(policy int, list []string) *WhiteBlackList {
	wb := WhiteBlackList{}
	wb.policy = policy
	wb.list = list
	return &wb
}

// setDefault sets all read and write permission of all columns to the given boolean.
// used if a model is read from cache.
// only sets the Permission on real database columns.
func SetDefaultPermission(m Interface, d bool) {
	for _, col := range m.Table().Cols {
		if col.ExistsInDB() || reflect.TypeOf(m.Table().strategy).Elem().Name() == "StrategyCustom" { // indicator that it is a real db column field and not
			col.Permission.Read = d
			col.Permission.Write = d

			tag, ok := reflect.TypeOf(m).Elem().FieldByName(col.StructField)
			if ok {
				configColumnByTag(col, tag.Tag.Get(TagName))
			}
		} else {
			col.Permission.Read = false
			col.Permission.Write = false
		}
	}
}

// isRelationDisabled checks if a relation of a struct is disabled through the black/whitelist.
func (wb *WhiteBlackList) isRelationDisabled(rel string) bool {
	exists := false
	for _, a := range wb.list {
		if a == rel || (wb.policy == WHITELIST && strings.HasPrefix(a, rel+".")) {
			exists = true
		}
	}

	if wb.policy == BLACKLIST {
		return exists
	}
	return !exists
}

// RelationWhiteBlackList creates a new WhiteBlackList struct for the relation .
// It's taking the parent white/blacklist and checks if there is any definition for the relation model.
func RelationWhiteBlackList(wbParent *WhiteBlackList, rel string) *WhiteBlackList {

	if wbParent == nil {
		return nil
	}

	var fields []string
	for _, a := range wbParent.list {
		if strings.HasPrefix(a, rel+".") {
			fields = append(fields, strings.Replace(a, rel+".", "", 1))
		}
	}

	if len(fields) > 0 {
		return NewWhiteBlackList(wbParent.policy, fields)
	}

	return nil
}

// setFieldPermission sets the permission read/write for all columns for the given black/whitelist.
func (wb *WhiteBlackList) setFieldPermission(m *Model, calledFrom string) error {
	// resetting all Permissions
	defVal := true
	if wb.policy == WHITELIST {
		defVal = false
	}
	SetDefaultPermission(m, defVal)
	// normal fields
	i := 1
loop:
	for _, listField := range wb.list {

		// needed that no constraint is breaking...
		// only needed for create and TODO also for update?
		for _, conf := range m.Table().Associations {
			if conf.Type == BelongsTo && calledFrom == "create" {
				conf.StructTable.Permission.Write = true
			}
		}

		for _, col := range m.Table().Cols {
			// always allow add primary keys
			if col.Information.PrimaryKey {
				col.Permission.Read = true

				// the id does not have to get updated all the time, first/all/update does not need it.
				// at the moment the update function can not update an id... so this will work for now.
				// TODO for later, check if the field is set by the user, if so, it also has to get updated.
				if calledFrom == "create" {
					col.Permission.Write = true
				} else {
					col.Permission.Write = false
				}
			}

			if col.StructField == listField {
				if !col.Information.PrimaryKey { // needed that no pkey can be blacklisted
					col.Permission.Read = !defVal
					col.Permission.Write = !defVal
					i++
				}

				// checking if its a reference key - not allowed to blacklist
				if wb.policy == BLACKLIST {
					for _, rel := range m.Table().Associations {
						if rel.StructTable == col || rel.AssociationTable == col {
							col.Permission.Read = true
							col.Permission.Write = true
							i++
						}
					}
				}

				continue loop
			}
		}
	}

	return nil

}

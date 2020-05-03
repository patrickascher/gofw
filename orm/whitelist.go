package orm

import (
	"errors"
	"fmt"
	"strings"

	"github.com/patrickascher/gofw/slices"
)

// white/blacklist policy
const (
	BLACKLIST = 0
	WHITELIST = 1
)

var errRelation = errors.New("orm: relation is type nil")

// whiteBlackList struct
type whiteBlackList struct {
	policy int
	fields []string
}

// newWBList creates a new white/blacklist with the given policy and fields.
func newWBList(policy int, fields []string) *whiteBlackList {
	wb := &whiteBlackList{}
	wb.policy = policy
	wb.fields = fields
	return wb
}

// setFieldPermission sets the permission read/write for all columns for the given black/whitelist.
// It is called on first, all, create, update and delete.
// The fields wb fields are not getting decreased, because they are added to the child object on a self referencing model.
func (scope *Scope) setFieldPermission(action string) error {

	// adding/removing mandatory fields of the wb list.
	// foreign keys, primary keys and the time fields are always allowed.
	err := addMandatoryFields(scope, action)
	if err != nil {
		return err
	}

	// if no wb list is defined, return.
	if scope.model.wbList == nil {
		return nil
	}

	fmt.Println("BLACKLIST AFTER", scope.model.wbList.fields)

	whitelisted := false
	if scope.model.wbList.policy == WHITELIST {
		whitelisted = true
	}

	// loop over all field
fields:
	for i := range scope.model.fields {
		// set all fields to the opposite value.
		scope.model.fields[i].Permission.Read = !whitelisted
		scope.model.fields[i].Permission.Write = !whitelisted

		// loop over the given white/blacklist fields and set the given policy
		for _, wbField := range scope.model.wbList.fields {
			if scope.model.fields[i].Name == wbField {
				scope.model.fields[i].Permission.Read = whitelisted
				scope.model.fields[i].Permission.Write = whitelisted
				continue fields
			}
		}
	}

	// loop over relations
relations:
	for i := range scope.model.relations {
		// set all fields to the opposite value.
		scope.model.relations[i].Permission.Read = !whitelisted
		scope.model.relations[i].Permission.Write = !whitelisted

		// loop over the white/blacklist fields and set the given policy
		// if there is a dot notation (example User.Name) the User relation is set to required on a whitelist.
		for _, wbField := range scope.model.wbList.fields {
			if scope.model.relations[i].Field == wbField || (scope.model.wbList.policy == WHITELIST && strings.HasPrefix(wbField, scope.model.relations[i].Field+".")) {
				// if the relation exist, by the name add or remove it
				// if whitelist and its a relation dot notation, add the relation because its mandatory.
				scope.model.relations[i].Permission.Read = whitelisted
				scope.model.relations[i].Permission.Write = whitelisted

				if scope.model.relations[i].Field == wbField {
					// if a relation is added completely and there is also a dot notation on that relation, remove it because the whole relation is added anyway.
					if deleteFields := slices.PrefixedWith(scope.model.wbList.fields, scope.model.relations[i].Field+"."); len(deleteFields) > 0 {
						for _, deleteField := range deleteFields {
							if i, exists := slices.Exists(scope.model.wbList.fields, deleteField); exists {
								scope.model.wbList.fields = append(scope.model.wbList.fields[:i], scope.model.wbList.fields[i+1:]...)
							}
						}
					}
				}

				continue relations
			}
		}
	}

	return nil
}

// addMandatoryFields will add all foreign keys, primary keys and time fields.
// If a full relation is added, it will be ignored on Blacklist, because all fields are added or removed anyway. On Whitelist the foreign key is added.
// If something like Relation.Child1.Child2.Name exists, it will recursively add all mandatory keys.
// If the  wb list is empty, this will skip.
func addMandatoryFields(scope *Scope, action string) error {

	// skip if no  wb list is defined
	if scope.model.wbList == nil {
		return nil
	}

	// return value
	var rv []string

	// always allow primary keys
	// needed for select + relations, create,update,delete
	pKeys := scope.PrimaryKeysFieldName()
	rv = append(rv, pKeys...)

	// relation permission
	// TODO: check if fits - specific permission is deleted, because otherwise the user can disable mandatory fields by tag.
	perm := Permission{}
	//var perm Permission
	//switch action {
	//case FIRST, ALL:
	//	perm = Permission{Read: false}
	//case CREATE, UPDATE, DELETE:
	//	perm = Permission{Write: false}
	//}

	// time fields are always added if they exist in the database.
	tf := scope.TimeFields(perm)
	rv = append(rv, tf...)

	for _, relation := range scope.Relations(perm) {

		// Whole relations on white or blacklist can be ignored because they are added completely.
		// Only on a whitelist the relation fk is added because the data is needed for referencing.
		// The fk is only added if it does not exist yet.
		if _, exists := slices.Exists(scope.model.wbList.fields, relation.Field); exists {
			if scope.model.wbList.policy == WHITELIST {
				if _, exists := slices.Exists(rv, relation.ForeignKey.Name); !exists {
					rv = append(rv, relation.ForeignKey.Name)
				}
			}
			continue
		}

		// Relation with dot notations adds all mandatory fields recursively.
		// The fields are only added if they dont exist yet.
		if relChild := slices.PrefixedWith(scope.model.wbList.fields, relation.Field+"."); relChild != nil {
			for _, rc := range relChild {
				relFields, err := mandatoryKeys(scope, perm, relation, strings.Split(rc, ".")[1:])
				if err != nil {
					return err
				}
				for _, relField := range relFields {
					if _, exists := slices.Exists(rv, relField); !exists {
						rv = append(rv, relField)
					}
				}
			}
		} else {
			// If relation is not blacklisted, the foreign key field is required and not allowed to be blacklisted.
			if scope.model.wbList.policy == BLACKLIST {
				if _, exists := slices.Exists(rv, relation.ForeignKey.Name); !exists {
					rv = append(rv, relation.ForeignKey.Name)
				}
			}
		}
	}

	// If its a whitelist, the needed keys will be merged with the existing list.
	// On a blacklist, the required fields will be excluded of the wb list, because they are mandatory for the relation chain.
	if scope.model.wbList.policy == WHITELIST {
		scope.model.wbList.fields = slices.MergeUnique(scope.model.wbList.fields, rv)
	} else {
		for _, r := range rv {
			if p, exists := slices.Exists(scope.model.wbList.fields, r); exists {
				scope.model.wbList.fields = append(scope.model.wbList.fields[:p], scope.model.wbList.fields[p+1:]...)
			}
		}
	}

	// if the whole wb list is dissolve because the the blacklisted fields are required, set the wb list to nil.
	if len(scope.model.wbList.fields) == 0 {
		scope.model.wbList = nil
	}

	return nil
}

// mandatoryKeys recursively adds all required keys (primary, fk, afk, poly)
// If the policy is Whitelist, all fields are added additional to the custom wb list.
// On Blacklist, the fields are removed from the wb list, because they are mandatory to guarantee the relation chain.
func mandatoryKeys(scope *Scope, p Permission, relation Relation, fields []string) ([]string, error) {
	var rv []string

	// if a relation id of the type nil, return error
	if relation.Type == nil {
		return rv, errRelation
	}

	// relation model from cache
	relModel, err := scope.CachedModel(relation.Type.String())
	if err != nil {
		return nil, err
	}

	// add all primary keys of the relation model
	for _, pkey := range relModel.Scope().PrimaryKeysFieldName() {
		if _, exists := slices.Exists(rv, relation.Field+"."+pkey); !exists {
			rv = append(rv, relation.Field+"."+pkey)
		}
	}

	// add the foreign key
	if _, exists := slices.Exists(rv, relation.ForeignKey.Name); !exists {
		rv = append(rv, relation.ForeignKey.Name)
	}

	// add polymorphic fields or association foreign key
	if scope.IsPolymorphic(relation) {
		if _, exists := slices.Exists(rv, relation.Field+"."+relation.Polymorphic.Field.Name); !exists {
			rv = append(rv, relation.Field+"."+relation.Polymorphic.Field.Name)
		}
		if _, exists := slices.Exists(rv, relation.Field+"."+relation.Polymorphic.Type.Name); !exists {
			rv = append(rv, relation.Field+"."+relation.Polymorphic.Type.Name)
		}
	} else {
		if _, exists := slices.Exists(rv, relation.Field+"."+relation.AssociationForeignKey.Name); !exists {
			rv = append(rv, relation.Field+"."+relation.AssociationForeignKey.Name)
		}
	}

	// if the depth of the added fields are bigger than 1, check the relation and run it recursively.
	if len(fields) > 1 {
		// if there is still a relation, get the cached model
		childRel, err := relModel.Scope().Relation(fields[0], p)
		if err != nil {
			return nil, err
		}
		childModel, err := scope.CachedModel(childRel.Type.String())
		if err != nil {
			return nil, err
		}

		// recursively add all mandatory fields
		relFields, err := mandatoryKeys(childModel.Scope(), p, childRel, fields[1:])
		if err != nil {
			return nil, err
		}
		for _, relField := range relFields {
			if _, exists := slices.Exists(rv, relation.Field+"."+relField); !exists {
				rv = append(rv, relation.Field+"."+relField)
			}
		}
	}

	return rv, nil
}

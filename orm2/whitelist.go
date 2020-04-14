package orm2

import (
	"fmt"
	"github.com/patrickascher/gofw/slices"
	"strings"
)

const (
	BLACKLIST = 0
	WHITELIST = 1
)

// WhiteBlackList struct
type whiteBlackList struct {
	policy int
	fields []string
}

// NewWBList creates a new white/blacklist with the given policy and fields.
func NewWBList(policy int, fields []string) *whiteBlackList {
	wb := &whiteBlackList{}
	wb.policy = policy
	wb.fields = fields
	return wb
}

// Fields return all wb list fields.
// If the wb list is nil, nil will return.
func (w *whiteBlackList) Fields() []string {
	if w == nil {
		return nil
	}
	return w.fields
}

// setDefaultPermission resets the field and relation permission by the cached model.
// TODO check if needed, otherwise delete
func (m *Model) setDefaultPermission() error {
	if !m.isInitialized {
		return fmt.Errorf(errInit.Error(), m.modelName(true))
	}

	// accessing cache // TODO if needed switch to scope
	v, err := m.cache.Get(m.caller.modelName(true))
	if err != nil || v == nil {
		return err
	}

	// reset field the permission
	for i, cachedField := range v.Value().(Model).fields {
		m.fields[i].Permission = cachedField.Permission
	}

	// reset relation permission
	for i, cachedRelation := range v.Value().(Model).relations {
		m.relations[i].Permission = cachedRelation.Permission
	}

	return nil
}

// setFieldPermission sets the permission read/write for all columns for the given black/whitelist.
// It is called by the scope on every action (First,All,...)
func setFieldPermission(scope Scope, action string) error {

	m := scope.Model()

	// adding/removing mandatory fields of the wb list.
	err := addMandatoryFields(scope, action)
	if err != nil {
		return err
	}

	// if no wb list is defined
	if scope.Model().wbList == nil {
		return nil
	}

	whitelisted := false
	if m.wbList.policy == WHITELIST {
		whitelisted = true
	}

	// loop over all field
fields:
	for i := range m.fields {

		// set all other fields
		m.fields[i].Permission.Read = !whitelisted
		m.fields[i].Permission.Write = !whitelisted

		// loop over the white/blacklist fields
		for n, wbField := range m.wbList.fields {
			if m.fields[i].Name == wbField {
				m.fields[i].Permission.Read = whitelisted
				m.fields[i].Permission.Write = whitelisted

				// decrease
				m.wbList.fields = append(m.wbList.fields[:n], m.wbList.fields[n+1:]...)
				continue fields
			}
		}
	}

	// loop over relations
relations:
	for i := range m.relations {
		m.relations[i].Permission.Read = !whitelisted
		m.relations[i].Permission.Write = !whitelisted

		// loop over the white/blacklist fields
		for n, wbField := range m.wbList.fields {
			if m.relations[i].Field == wbField || (m.wbList.policy == WHITELIST && strings.HasPrefix(wbField, m.relations[i].Field+".")) {
				// if the relation exist, by the name add or remove it
				// if whitelist and its a relation dot notation, add the relation because its mandatory.
				m.relations[i].Permission.Read = whitelisted
				m.relations[i].Permission.Write = whitelisted

				// decrease
				if m.relations[i].Field == wbField {
					m.wbList.fields = append(m.wbList.fields[:n], m.wbList.fields[n+1:]...)
				}
				continue relations
			}
		}
	}

	return nil
}

// addMandatoryFields will add all relation relevant fields and primary keys.
// If a full relation is added, it will be ignored on Blacklist,Whitelist because all fields are added or removed anyway.
// If something like Relation.Child1.Child2.Name exists, it will recursively add all relevant keys(pkey, fk, afk, poly).
// If the custom wb list is empty, nil will return.
func addMandatoryFields(scope Scope, action string) error {
	m := scope.Model()

	// skip if no custom wb list is defined
	if m.wbList == nil {
		return nil
	}

	// return value
	var rv []string

	// always add primary keys
	// needed for select + relations, create,update,delete
	pKeys := scope.PrimaryKeysFieldName()
	rv = append(rv, pKeys...)

	// relation
	var perm Permission
	switch action {
	case FIRST, ALL:
		perm = Permission{Read: true}
	case CREATE, UPDATE, DELETE:
		perm = Permission{Write: true}
	}

	for _, relation := range scope.Relations(perm) {

		// Relations on white or blacklist can be ignored because they are added completly.
		if _, exists := slices.Exists(m.wbList.fields, relation.Field); exists {
			continue
		}

		// Relation with dot notations (recursively added)
		if relChild := slices.PrefixedWith(m.wbList.fields, relation.Field+"."); relChild != nil {
			for _, rc := range relChild {

				tmp, err := mandatoryKeys(scope, perm, relation, strings.Split(rc, ".")[1:])
				if err != nil {
					return err
				}
				rv = append(rv, tmp...)
			}
		} else {
			if scope.model.wbList.policy == BLACKLIST {
				if _, exists := slices.Exists(rv, relation.ForeignKey.Name); !exists {
					rv = append(rv, relation.ForeignKey.Name)
				}
			}
		}
	}

	if m.wbList.policy == WHITELIST {
		m.wbList.fields = slices.MergeUnique(m.wbList.fields, rv)
	} else {
		for _, r := range rv {
			// skip relations because they are passed to the child later on.
			if strings.Contains(r, ".") {
				continue
			}
			if p, exists := slices.Exists(m.wbList.fields, r); exists {
				m.wbList.fields = append(m.wbList.fields[:p], m.wbList.fields[p+1:]...)
			}
		}
	}

	// delete if not needed anymore
	if len(m.wbList.fields) == 0 {
		m.wbList = nil
	}

	return nil
}

// mandatoryKeys recursively adds all relevant keys (primary, fk, afk, poly)
// If the policy is Whitelist, all fields are added additional to the custom wb list.
// On Blacklist, the fields are removed from the custom wb list, because they are mandatory to guarantee the relation chain.
func mandatoryKeys(scope Scope, p Permission, relation Relation, fields []string) ([]string, error) {
	var rv []string

	// get relation model from cache
	model := relation.Type.String()
	relScope, err := scope.CachedModel(model)
	if err != nil {
		return nil, err
	}

	// add all primary keys of the relation model
	for _, pkey := range relScope.PrimaryKeysFieldName() {
		if _, exists := slices.Exists(rv, relation.Field+"."+pkey); !exists {
			rv = append(rv, relation.Field+"."+pkey)
		}
	}

	// fk
	if _, exists := slices.Exists(rv, relation.ForeignKey.Name); !exists {
		rv = append(rv, relation.ForeignKey.Name)
	}

	// polymorphic
	if scope.IsPolymorphic(relation) {
		if _, exists := slices.Exists(rv, relation.Field+"."+relation.Polymorphic.Field.Name); !exists {
			rv = append(rv, relation.Field+"."+relation.Polymorphic.Field.Name)
		}
		if _, exists := slices.Exists(rv, relation.Field+"."+relation.Polymorphic.Type.Name); !exists {
			rv = append(rv, relation.Field+"."+relation.Polymorphic.Type.Name)
		}
	} else {
		// afk
		if _, exists := slices.Exists(rv, relation.Field+"."+relation.AssociationForeignKey.Name); !exists {
			rv = append(rv, relation.Field+"."+relation.AssociationForeignKey.Name)
		}
	}

	if len(fields) > 1 {
		// if there are still relations, add FK
		childRel := relScope.Relation(fields[0], p) // error
		relScope, err := scope.CachedModel(childRel.Type.String())
		if err != nil {
			return nil, err
		}

		rec, err := mandatoryKeys(relScope, p, childRel, fields[1:])
		if err != nil {
			return nil, err
		}
		for _, rec_ := range rec {
			if _, exists := slices.Exists(rv, relation.Field+"."+rec_); !exists {
				rv = append(rv, relation.Field+"."+rec_)
			}
		}
	}

	return rv, nil
}

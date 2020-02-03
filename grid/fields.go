package grid

import (
	"fmt"
	"strings"
)

func getFieldByDotNotation(search string, list map[string]Interface) (Interface, error) {
	relField := strings.Split(search, ".")
	lastField := list
	for i, f := range relField {
		if value, exist := lastField[f]; exist {
			if i == len(relField)-1 {
				return value, nil
			} else {
				lastField = value.getFields()
			}
		}
	}
	return nil, fmt.Errorf(ErrFieldOrRelation.Error(), search)
}

// getFieldByName returns the field by a given string.
// examples: Name, Parts.ID
func (g *Grid) getFieldByName(field string) (Interface, error) {

	// dot notation
	relField := strings.Split(field, ".")
	if len(relField) > 1 {
		return getFieldByDotNotation(field, g.fields)
	}

	// normal field
	f, err := g.Field(field)
	if err == nil {
		return f, nil
	}

	// relation
	r, err := g.Relation(field)
	if err == nil {
		return r, nil
	}

	return nil, fmt.Errorf(ErrFieldOrRelation.Error(), field)
}

// SetFieldsReadOnly allows to set multiple fields to read only.
// The value can be bool only because readOnly is just relevant for edit.
// Example: g.SetFieldsReadOnly(true,"ID","Name")
func (g *Grid) SetFieldsReadOnly(v bool, fields ...string) error {
	for _, field := range fields {
		f, err := g.getFieldByName(field)
		if err != nil {
			return err
		}
		f.setReadOnly(v)
	}

	return nil
}

// SetFieldsRemove allows to set multiple fields to remove.
// example: g.SetFieldsRemove(true,"ID","Name")
func (g *Grid) SetFieldsRemove(v interface{}, fields ...string) error {
	for _, field := range fields {
		f, err := g.getFieldByName(field)
		if err != nil {
			return err
		}
		f.setRemove(v)
	}

	return nil
}

// SetFieldsHidden allows to set multiple fields to remove.
// example: g.SetFieldsHidden(true,"ID","Name")
func (g *Grid) SetFieldsHidden(v interface{}, fields ...string) error {
	for _, field := range fields {
		f, err := g.getFieldByName(field)
		if err != nil {
			return err
		}
		f.setHide(v)
	}

	return nil
}

package grid

import (
	"fmt"
	"github.com/patrickascher/gofw/orm"
	"sort"
)

var relationCounter = 500

// head contains all information for one field or relation.
type head struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`

	FieldType map[string]interface{} `json:"type,omitempty"` // new----

	FieldDefault string `json:"default,omitempty"`
	FieldPrimary bool   `json:"primary,omitempty"`
	FieldName    string `json:"name"`

	Filter   bool `json:"filter,omitempty"`
	Sort     bool `json:"sortable,omitempty"`
	Remove   bool `json:"remove,omitempty"`
	Hide     bool `json:"hide,omitempty"`
	ReadOnly bool `json:"readOnly,omitempty"`

	Fields []*head `json:"fields,omitempty"`
}

// position is needed to reorganize the fields.
type position struct {
	pos   int
	field string
}

// sortHeaderInfo is sorting the fields by its position.
// the position is default the same as the database position, all relations are coming after it - alphabetical sorted.
// Be aware that an empty relation position is always last field position +1.
func sortHeaderInfo(fields map[string]Interface) []position {

	// sort relations alphabet if pos == 500
	var relations []string
	for _, f := range fields {
		if f.getFields() != nil && f.getPosition() >= relationCounter {
			relations = append(relations, f.getFieldName())
		}
	}
	if len(relations) > 0 {
		sort.Strings(relations)
		i := relationCounter
		for k, rel := range relations {
			fields[rel].setPosition(i + k)
		}
	}

	// sorting all fields
	var pos []position
	for k, f := range fields {
		if f.getRemove() {
			continue
		}
		pos = append(pos, position{pos: f.getPosition(), field: k})
	}

	sort.Slice(pos, func(i, j int) bool {
		return pos[i].pos < pos[j].pos
	})

	return pos
}

// headerFieldsLoop is going recursive over all fields and relations to fetch all relations and fields.
func headerFieldsLoop(fields map[string]Interface, jsonName bool) []*head {

	if fields == nil {
		return nil
	}

	sortedFields := sortHeaderInfo(fields)

	headInfo := make([]*head, len(sortedFields))
	//for k, f := range fields {
	for k, sortedField := range sortedFields {

		// getRemove fields are already removed in the sortHeaderInfo method.
		// position not needed because its already sorted in the backend.

		f := fields[sortedField.field]
		f.setPosition(k + 1) // correct the positions because the relations are starting with 500

		// checking if a json name is set
		fName := f.getFieldName()
		if jsonName && f.getJsonName() != "" {
			fName = f.getJsonName()
		}

		headInfo[k] = &head{
			Title:       f.getTitle(),
			Description: f.getDescription(),
			FieldName:   fName, // needed for edit/delete link, sorting,

			Filter:   f.getFilter(),
			Sort:     f.getSort(),
			Hide:     f.getHide(),
			ReadOnly: f.getReadOnly(),
			Fields:   headerFieldsLoop(f.getFields(), jsonName),
			//Select:   f.getSelect(),
		}
		fmt.Println("****", sortedField.field, f.getFieldName())

		if len(headInfo[k].Fields) == 0 { // = no relation
			headInfo[k].FieldPrimary = f.getColumn().Information.PrimaryKey
			f.getFieldType().SetValidator(f.getColumn().Validator.Config)
			headInfo[k].FieldDefault = f.getColumn().Information.DefaultValue.String
		}

		// a BelongsTo relation is always required
		if f.getFieldType().Name() == orm.BelongsTo {
			f.getFieldType().SetValidator("required")
		}

		// set the fieldtype FieldType
		headInfo[k].FieldType = fieldTypeToMap(f.getFieldType())

	}
	return headInfo
}

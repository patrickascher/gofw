package grid

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Field(t *testing.T) {
	test := assert.New(t)

	// set fields by normal values
	field := Field{}
	field.SetMode(VTable)

	field.SetId("id")
	test.Equal("id", field.id)
	test.Equal("id", field.Id())

	field.SetDatabaseId("rid")
	test.Equal("rid", field.referenceId)

	field.SetPrimary(true)
	test.Equal(true, field.primary)

	field.SetReadOnly(true)
	test.True(field.IsReadOnly())
	test.Equal(true, field.readOnly)

	field.SetFieldType("fieldType")
	test.Equal("fieldType", field.fieldType)
	test.Equal("fieldType", field.FieldType())

	field.SetTitle(NewValue("title"))
	test.Equal(5, len(field._title))
	test.Equal("title", field.Title())

	field.SetDescription(NewValue("description"))
	test.Equal(5, len(field._description))
	test.Equal("description", field.Description())

	field.SetPosition(NewValue(1))
	test.Equal(5, len(field._position))
	test.Equal(1, field.Position())

	field.SetRemove(NewValue(true))
	test.Equal(5, len(field._remove))
	test.Equal(true, field.IsRemoved())

	field.SetHidden(NewValue(true))
	test.Equal(5, len(field._hidden))
	test.Equal(true, field.IsHidden())

	field.SetView(NewValue("view"))
	test.Equal(5, len(field._view))
	test.Equal("view", field.View())

	field.SetFilterable(true)
	test.Equal(true, field.filterable)
	test.Equal(true, field.IsFilterable())

	field.SetSortable(true)
	test.Equal(true, field.sortable)
	test.Equal(true, field.IsSortable())

	field.SetOption("select", Select{})
	test.Equal(1, len(field.options))
	test.Equal(Select{}, field.options["select"])

	field.SetCallback("callback", 1, 2)
	test.Equal("callback", field.callback)
	test.Equal(2, len(field.callbackArguments))

	test.NoError(field.error)
	field.setError(errors.New(""))
	test.Error(field.error)
	test.Error(field.Error())

	field.SetFields([]Field{{id: "child"}})
	test.Equal(1, len(field.fields))
	test.Equal("child", field.Field("child").Id())
	test.Equal("id", field.Field("xy").Id()) // the old field is returned and an error is set
	test.Error(field.Field("xy").Error())

	// set fields by struct values
	field = Field{}
	field.SetMode(VTable)

	field.SetId("id")
	field.SetFieldType("fieldType")

	field.SetTitle("title")
	test.Equal(5, len(field._title))
	test.Equal("title", field.Title())

	field.SetDescription("description")
	test.Equal(5, len(field._description))
	test.Equal("description", field.Description())

	field.SetPosition(1)
	test.Equal(5, len(field._position))
	test.Equal(1, field.Position())

	field.SetRemove(true)
	test.Equal(5, len(field._remove))
	test.Equal(true, field.IsRemoved())

	field.SetHidden(true)
	test.Equal(5, len(field._hidden))
	test.Equal(true, field.IsHidden())

	field.SetView("view")
	test.Equal(5, len(field._view))
	test.Equal("view", field.View())

	field.SetFilterable(true)
	field.SetSortable(true)
	field.SetOption("select", Select{})
	field.SetCallback("callback", 1, 2)
	field.setError(errors.New(""))

	test.Equal("id", field.id)
	test.Equal("fieldType", field.fieldType)
	test.Equal(true, field.filterable)
	test.Equal(true, field.sortable)
	test.Equal(Select{}, field.options["select"])
	test.Equal("callback", field.callback)
	test.Equal(2, len(field.callbackArguments))
	test.Error(field.error)

}

func TestField_MarshalJSON(t *testing.T) {
	test := assert.New(t)

	field := Field{}
	field.id = "id"
	field.fieldType = "type"
	field.primary = true
	field.SetTitle("title")
	field.SetDescription("desc")
	field.SetPosition(1)
	field.SetRemove(true)
	field.SetHidden(true)
	field.SetView("view")
	field.readOnly = true
	field.sortable = true
	field.filterable = true
	field.options = map[string]interface{}{"option": "value"}
	field.fields = append(field.fields, Field{id: "child"})

	bytes, err := field.MarshalJSON()
	test.NoError(err)

	fNew := map[string]interface{}{}
	err = json.Unmarshal(bytes, &fNew)
	test.NoError(err)
	test.Equal(11, len(fNew))

	// test empty/zero values
	field = Field{}
	field.id = "id"
	field.fieldType = "type"
	field.primary = false

	field.SetTitle(NewValue("title"))
	field.SetDescription(NewValue("desc"))
	field.SetPosition(NewValue(1))
	field.SetRemove(NewValue(true))
	field.SetHidden(NewValue(true))
	field.SetView(NewValue("view"))

	field.readOnly = false
	field.sortable = false
	field.filterable = false

	bytes, err = field.MarshalJSON()
	test.NoError(err)

	fNew = map[string]interface{}{}
	err = json.Unmarshal(bytes, &fNew)
	test.NoError(err)
	test.Equal(5, len(fNew))
}

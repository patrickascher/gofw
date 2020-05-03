package grid

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"strings"
	"testing"
)

func valueHelper(i interface{}) value {
	return value{grid: nil, table: i, details: i, update: i, create: i}
}
func Test_Field(t *testing.T) {
	test := assert.New(t)

	grid := New(newController(httptest.NewRequest("GET", "https://localhost/users", strings.NewReader(""))))

	// set fields by normal values
	field := Field{}

	field.SetId("id")
	test.Equal("id", field.id)
	test.Equal("id", field.Id())

	field.SetReferenceId("rid")
	test.Equal("rid", field.referenceId)

	field.SetPrimary(true)
	test.Equal(true, field.primary)

	field.SetReadOnly(true)
	test.True(field.IsReadOnly())
	test.Equal(true, field.readOnly)

	field.SetFieldType("fieldType")
	test.Equal("fieldType", field.fieldType)
	test.Equal("fieldType", field.FieldType())

	field.title = *grid.NewValue("title")
	field.SetTitle(grid.NewValue("title"))
	test.Equal(grid.NewValue("title"), &field.title)
	test.Equal("title", field.Title())

	field.description = *grid.NewValue("description")
	field.SetDescription(grid.NewValue("description"))
	test.Equal(grid.NewValue("description"), &field.description)
	test.Equal("description", field.Description())

	field.position = *grid.NewValue(1)
	field.SetPosition(grid.NewValue(1))
	test.Equal(grid.NewValue(1), &field.position)
	test.Equal(1, field.Position())

	field.remove = *grid.NewValue(true)
	field.SetRemove(grid.NewValue(true))
	test.Equal(grid.NewValue(true), &field.remove)
	test.Equal(true, field.IsRemoved())

	field.hidden = *grid.NewValue(true)
	field.SetHidden(grid.NewValue(true))
	test.Equal(grid.NewValue(true), &field.hidden)
	test.Equal(true, field.IsHidden())

	field.view = *grid.NewValue("view")
	field.SetView(grid.NewValue("view"))
	test.Equal(grid.NewValue("view"), &field.view)
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

	field.SetId("id")
	field.SetFieldType("fieldType")

	field.title = *grid.NewValue("")
	field.SetTitle(grid.NewValue("title").SetDetails("title_details"))

	field.description = *grid.NewValue("")
	field.SetDescription(grid.NewValue("description").SetDetails("description_details"))

	field.position = *grid.NewValue(0)
	field.SetPosition(grid.NewValue(1).SetDetails(2))

	field.remove = *grid.NewValue(false)
	field.SetRemove(grid.NewValue(true).SetDetails(false))

	field.hidden = *grid.NewValue(false)
	field.SetHidden(grid.NewValue(true).SetDetails(false))

	field.view = *grid.NewValue("")
	field.SetView(grid.NewValue("view").SetDetails("view_details"))

	field.SetFilterable(true)
	field.SetSortable(true)
	field.SetOption("select", Select{})
	field.SetCallback("callback", 1, 2)
	field.setError(errors.New(""))

	test.Equal("id", field.id)
	test.Equal("fieldType", field.fieldType)
	test.Equal(&value{grid: grid, table: "title", details: "title_details", create: "title", update: "title"}, &field.title)
	test.Equal(&value{grid: grid, table: "description", details: "description_details", create: "description", update: "description"}, &field.description)
	test.Equal(&value{grid: grid, table: 1, details: 2, create: 1, update: 1}, &field.position)
	test.Equal(&value{grid: grid, table: true, details: false, create: true, update: true}, &field.remove)
	test.Equal(&value{grid: grid, table: true, details: false, create: true, update: true}, &field.hidden)
	test.Equal(&value{grid: grid, table: "view", details: "view_details", create: "view", update: "view"}, &field.view)
	test.Equal(true, field.filterable)
	test.Equal(true, field.sortable)
	test.Equal(Select{}, field.options["select"])
	test.Equal("callback", field.callback)
	test.Equal(2, len(field.callbackArguments))
	test.Error(field.error)

}

func TestField_MarshalJSON(t *testing.T) {
	test := assert.New(t)
	grid := New(newController(httptest.NewRequest("GET", "https://localhost/users", strings.NewReader(""))))

	field := Field{}
	field.id = "id"
	field.fieldType = "type"
	field.primary = true
	field.title = *grid.NewValue("title")
	field.description = *grid.NewValue("desc")
	field.position = *grid.NewValue(1)
	field.remove = *grid.NewValue(true)
	field.hidden = *grid.NewValue(true)
	field.view = *grid.NewValue("view")
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
	test.Equal(14, len(fNew))

	// test empty/zero values
	field = Field{}
	field.id = "id"
	field.fieldType = "type"
	field.primary = false
	field.title = *grid.NewValue("title")
	field.description = *grid.NewValue("")
	field.position = *grid.NewValue(1)
	field.remove = *grid.NewValue(false)
	field.hidden = *grid.NewValue(false)
	field.view = *grid.NewValue("")
	field.readOnly = false
	field.sortable = false
	field.filterable = false

	bytes, err = field.MarshalJSON()
	test.NoError(err)

	fNew = map[string]interface{}{}
	err = json.Unmarshal(bytes, &fNew)
	test.NoError(err)
	test.Equal(4, len(fNew))
}

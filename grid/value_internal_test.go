package grid

import (
	"github.com/patrickascher/gofw/controller"
	"github.com/patrickascher/gofw/controller/context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// helper to change the different grid modes
func newController(r *http.Request) controller.Interface {
	c := controller.Controller{}
	rw := httptest.NewRecorder()
	ctx := context.New(r, rw)
	c.SetContext(ctx)
	return &c
}

// Testing if the value is set for all items if a normal type is used and if the value is set by mode.
func TestNewValue(t *testing.T) {
	test := assert.New(t)

	grid := New(newController(httptest.NewRequest("GET", "https://localhost/users", strings.NewReader(""))))
	v := grid.NewValue("title")
	test.Equal(grid, v.grid)

	// test if every value is set to the same
	test.Equal("title", v.create)
	test.Equal("title", v.details)
	test.Equal("title", v.update)
	test.Equal("title", v.table)
	test.Equal("title", v.getString())

	// test empty value
	v.table = nil
	test.Equal("", v.getString())

	v = grid.NewValue("title").SetTable("title_table").SetCreate("title_create").SetDetails("title_details").SetUpdate("title_edit")
	test.Equal(grid, v.grid)
	// test if every value is set to the same
	test.Equal("title_create", v.create)
	test.Equal("title_details", v.details)
	test.Equal("title_edit", v.update)
	test.Equal("title_table", v.table)
	test.Equal("title_table", v.getString())
	// test empty value
	v.table = nil
	test.Equal("", v.getString())
}

// Test if the return value is correct for by mode.
func TestValue_Mode(t *testing.T) {

	test := assert.New(t)

	grid := New(newController(httptest.NewRequest("GET", "https://localhost/users", strings.NewReader(""))))
	v := grid.NewValue("title").SetDetails("title_details").SetCreate("title_create").SetUpdate("title_edit")
	test.Equal(grid, v.grid)

	// Table View
	test.Equal("title", v.getString())
	// Details View
	grid.controller = newController(httptest.NewRequest("GET", "https://localhost/users?mode=details", strings.NewReader("")))
	test.Equal("title_details", v.getString())
	// Create View
	grid.controller = newController(httptest.NewRequest("GET", "https://localhost/users?mode=create", strings.NewReader("")))
	test.Equal("title_create", v.getString())
	// Edit View
	grid.controller = newController(httptest.NewRequest("GET", "https://localhost/users?mode=update", strings.NewReader("")))
	test.Equal("title_edit", v.getString())

	// mode does not exist
	grid.controller = newController(httptest.NewRequest("POST", "https://localhost/users?mode=update", strings.NewReader("")))
	test.Equal("title_create", v.getString())
}

// Test if the returned type is correct.
func TestValue_Types(t *testing.T) {

	test := assert.New(t)

	grid := New(newController(httptest.NewRequest("GET", "https://localhost/users", strings.NewReader(""))))

	vString := grid.NewValue("title")
	test.Equal("title", vString.getString())
	vString = grid.NewValue(nil)
	test.Equal("", vString.getString())

	vBool := grid.NewValue(true)
	test.Equal(true, vBool.getBool())
	vBool = grid.NewValue(nil)
	test.Equal(false, vBool.getBool())

	vInt := grid.NewValue(1)
	test.Equal(1, vInt.getInt())
	vInt = grid.NewValue(nil)
	test.Equal(0, vInt.getInt())

	vInterface := grid.NewValue([]string{"title"})
	test.Equal([]string{"title"}, vInterface.getInterface())
	vInterface = grid.NewValue(nil)
	test.Equal(nil, vInterface.getInterface())
}

// Test the setValueHelper which offers to enter normal go types or pass a new *value struct.
func TestValue_SetValue(t *testing.T) {
	test := assert.New(t)

	grid := New(newController(httptest.NewRequest("GET", "https://localhost/users", strings.NewReader(""))))
	v := grid.NewValue("title")

	setValueHelper(v, "title_2")
	test.Equal("title_2", v.table)
	test.Equal("title_2", v.details)
	test.Equal("title_2", v.update)
	test.Equal("title_2", v.create)

	setValueHelper(v, grid.NewValue("title").SetTable("title_table"))
	test.Equal("title_table", v.table)
	test.Equal("title", v.details)
	test.Equal("title", v.update)
	test.Equal("title", v.create)
}

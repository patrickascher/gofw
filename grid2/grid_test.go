package grid2_test

import (
	"errors"
	"github.com/patrickascher/gofw/controller"
	"github.com/patrickascher/gofw/controller/context"
	"github.com/patrickascher/gofw/grid2"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

// helper to change the different grid modes
func newController(r *http.Request) (controller.Interface, *httptest.ResponseRecorder) {
	c := controller.Controller{}
	rw := httptest.NewRecorder()
	ctx := context.New(r, rw)
	c.SetContext(ctx)
	return &c, rw
}

type mockSource struct {
	throwError       error
	throwErrorAll    error
	throwErrorFields error

	initCalled     bool
	uFieldCalled   bool
	callbackCalled bool

	oneCalled    bool
	allCalled    bool
	createCalled bool
	updateCalled bool
	deleteCalled bool
}

func (m *mockSource) Init(grid *grid2.Grid) error {
	m.initCalled = true
	return m.throwError
}

// Fields of the grid.
func (m *mockSource) Fields(grid *grid2.Grid) ([]grid2.Field, error) {
	if m.throwErrorFields != nil {
		return nil, m.throwErrorFields
	}
	var rv []grid2.Field
	field := grid2.Field{}
	field.SetId("id").SetTitle("title").SetSortable(true).SetPrimary(true)
	rv = append(rv, field)
	return rv, m.throwError
}

// UpdatedFields is called before render. The grid fields have the user updated configurations.
func (m *mockSource) UpdatedFields(grid *grid2.Grid) error {
	m.uFieldCalled = true
	return m.throwError
}

// Callback is called on a callback request of the grid.
func (m *mockSource) Callback(callback string, grid *grid2.Grid) (interface{}, error) {
	m.callbackCalled = true
	return nil, m.throwError
}

// One request a single row by the given condition.
func (m *mockSource) One(c *sqlquery.Condition, grid *grid2.Grid) (interface{}, error) {
	m.oneCalled = true
	return "some data", m.throwError
}

// All data by the given condition.
func (m *mockSource) All(c *sqlquery.Condition, grid *grid2.Grid) (interface{}, error) {
	m.allCalled = true
	return "some data", m.throwErrorAll
}

// Create the object
func (m *mockSource) Create(grid *grid2.Grid) (interface{}, error) {
	m.createCalled = true
	return nil, m.throwError
}

// Update the object
func (m *mockSource) Update(grid *grid2.Grid) error {
	m.updateCalled = true
	return m.throwError
}

// Delete the object by the given condition.
func (m *mockSource) Delete(c *sqlquery.Condition, grid *grid2.Grid) error {
	m.deleteCalled = true
	return m.throwError
}

// Count all the existing object by the given condition.
func (m *mockSource) Count(c *sqlquery.Condition) (int, error) {
	return 10, m.throwError
}

func TestNew(t *testing.T) {
	test := assert.New(t)
	grid := grid2.New(nil)
	test.IsType(&grid2.Grid{}, grid)
}

func TestGrid_Controller(t *testing.T) {
	test := assert.New(t)
	c := controller.Controller{}
	grid := grid2.New(&c)
	test.Equal(&c, grid.Controller())
}

func TestGrid_SetSource_Fields_Field(t *testing.T) {
	test := assert.New(t)

	r := httptest.NewRequest("GET", "https://localhost/users", strings.NewReader(""))
	// new controller
	c := controller.Controller{}
	rw := httptest.NewRecorder()
	ctx := context.New(r, rw)
	c.SetContext(ctx)

	grid := grid2.New(&c)
	mock := &mockSource{}
	err := grid.SetSource(mock)
	test.NoError(err)

	// check if the init method was called
	test.Equal(true, mock.initCalled)
	// check if the fields were added by the source
	test.Equal(1, len(grid.Fields()))
	//test field
	test.Equal("id", grid.Field("id").Id())
	//test non existing field
	test.Equal("", grid.Field("xy").Id())
	test.NotNil(grid.Field("xy").Error())

	// SetSource errors
	controller, rw := newController(httptest.NewRequest("GET", "https://localhost/users?sort=id&noheader=1", strings.NewReader("")))
	g := grid2.New(controller)
	mock = &mockSource{}
	mock.throwError = errors.New("")
	err = g.SetSource(mock) // error on source.Init()
	test.Error(err)
	mock.throwError = nil
	mock.throwErrorFields = errors.New("")
	err = g.SetSource(mock) // error on source.Fields()
	test.Error(err)
}

func TestGrid_Mode(t *testing.T) {
	test := assert.New(t)

	// VTable - GET no mode param
	c, _ := newController(httptest.NewRequest("GET", "https://localhost/users", strings.NewReader("")))
	grid := grid2.New(c)
	test.Equal(grid2.VTable, grid.Mode())
	// VCreate - GET with mode param "create"
	c, _ = newController(httptest.NewRequest("GET", "https://localhost/users?mode=create", strings.NewReader("")))
	grid = grid2.New(c)
	test.Equal(grid2.VCreate, grid.Mode())
	// VUpdate - GET with mode param "update"
	c, _ = newController(httptest.NewRequest("GET", "https://localhost/users?mode=update", strings.NewReader("")))
	grid = grid2.New(c)
	test.Equal(grid2.VUpdate, grid.Mode())
	// VDetails - GET with mode param "details"
	c, _ = newController(httptest.NewRequest("GET", "https://localhost/users?mode=details", strings.NewReader("")))
	grid = grid2.New(c)
	test.Equal(grid2.VDetails, grid.Mode())
	// VCallback - GET with mode param "callback"
	c, _ = newController(httptest.NewRequest("GET", "https://localhost/users?mode=callback", strings.NewReader("")))
	grid = grid2.New(c)
	test.Equal(grid2.CALLBACK, grid.Mode())

	// CREATE - POST
	c, _ = newController(httptest.NewRequest("POST", "https://localhost/users", strings.NewReader("")))
	grid = grid2.New(c)
	test.Equal(grid2.CREATE, grid.Mode())
	// UPDATE - PUT
	c, _ = newController(httptest.NewRequest("PUT", "https://localhost/users", strings.NewReader("")))
	grid = grid2.New(c)
	test.Equal(grid2.UPDATE, grid.Mode())
	// DELETE - DELETE
	c, _ = newController(httptest.NewRequest("DELETE", "https://localhost/users", strings.NewReader("")))
	grid = grid2.New(c)
	test.Equal(grid2.DELETE, grid.Mode())

	// not defined - PATCH
	c, _ = newController(httptest.NewRequest("PATCH", "https://localhost/users", strings.NewReader("")))
	grid = grid2.New(c)
	test.Equal(0, grid.Mode())
}

func TestGrid_Render(t *testing.T) {
	test := assert.New(t)

	// error - no source was added
	c, rw := newController(httptest.NewRequest("GET", "https://localhost/users?mode=details", strings.NewReader("")))
	g := grid2.New(c)
	g.Render()
	test.Equal(500, rw.Code)

	// CREATE
	c, rw = newController(httptest.NewRequest("POST", "https://localhost/users", strings.NewReader("")))
	g = grid2.New(c)
	mock := &mockSource{}
	err := g.SetSource(mock)
	test.NoError(err)
	g.Render()
	test.True(mock.createCalled)
	test.Equal(200, rw.Code)
	mock.throwError = errors.New("")
	g.Render()
	test.Equal(500, rw.Code)

	// UPDATE
	c, rw = newController(httptest.NewRequest("PUT", "https://localhost/users", strings.NewReader("")))
	g = grid2.New(c)
	mock = &mockSource{}
	err = g.SetSource(mock)
	test.NoError(err)
	g.Render()
	test.True(mock.updateCalled)
	test.Equal(200, rw.Code)
	mock.throwError = errors.New("")
	g.Render()
	test.Equal(500, rw.Code)

	// DELETE - primary is missing
	c, rw = newController(httptest.NewRequest("DELETE", "https://localhost/users", strings.NewReader("")))
	g = grid2.New(c)
	mock = &mockSource{}
	err = g.SetSource(mock)
	test.NoError(err)
	g.Render()
	test.False(mock.deleteCalled)
	test.Equal(500, rw.Code)
	mock.throwError = errors.New("")
	g.Render()
	test.Equal(500, rw.Code)

	// DELETE - primary is missing
	c, rw = newController(httptest.NewRequest("DELETE", "https://localhost/users?id=1", strings.NewReader("")))
	g = grid2.New(c)
	mock = &mockSource{}
	err = g.SetSource(mock)
	test.NoError(err)
	g.Render()
	test.True(mock.deleteCalled)
	test.Equal(200, rw.Code)
	mock.throwError = errors.New("")
	g.Render()
	test.Equal(500, rw.Code)

	// TODO CALLBACK - when it is created

	// VCREATE - check if the header is added correctly
	c, rw = newController(httptest.NewRequest("GET", "https://localhost/users?mode=create", strings.NewReader("")))
	g = grid2.New(c)
	mock = &mockSource{}
	err = g.SetSource(mock)
	test.NoError(err)
	g.Render()
	test.Equal(200, rw.Code)
	test.Equal(1, len(c.Context().Response.Data("head").([]grid2.Field)))

	// VUpdate, VDetails - no primary added
	c, rw = newController(httptest.NewRequest("GET", "https://localhost/users?mode=update", strings.NewReader("")))
	g = grid2.New(c)
	mock = &mockSource{}
	err = g.SetSource(mock)
	test.NoError(err)
	g.Render()
	test.Equal(500, rw.Code)

	// VUpdate, VDetails - with primary
	c, rw = newController(httptest.NewRequest("GET", "https://localhost/users?mode=update&id=1", strings.NewReader("")))
	g = grid2.New(c)
	mock = &mockSource{}
	err = g.SetSource(mock)
	test.NoError(err)
	g.Render()
	test.Equal(200, rw.Code)
	test.Equal(1, len(c.Context().Response.Data("head").([]grid2.Field)))
	test.Equal("some data", c.Context().Response.Data("data"))
	test.True(mock.oneCalled)
	mock.throwError = errors.New("")
	g.Render()
	test.Equal(500, rw.Code)

	// VTable - error sort field does not exist
	c, rw = newController(httptest.NewRequest("GET", "https://localhost/users?sort=ABC", strings.NewReader("")))
	g = grid2.New(c)
	mock = &mockSource{}
	err = g.SetSource(mock)
	test.NoError(err)
	g.Render()
	test.Equal(500, rw.Code)

	// VTable
	c, rw = newController(httptest.NewRequest("GET", "https://localhost/users?sort=id", strings.NewReader("")))
	g = grid2.New(c)
	mock = &mockSource{}
	err = g.SetSource(mock)
	test.NoError(err)
	g.Render()
	test.Equal(200, rw.Code)
	test.Equal(1, len(c.Context().Response.Data("head").([]grid2.Field)))
	test.Equal("some data", c.Context().Response.Data("data"))
	test.Equal("*grid2.pagination", reflect.TypeOf(c.Context().Response.Data("pagination")).String())
	test.True(mock.allCalled)
	mock.throwError = errors.New("")
	g.Render() // error on pagination because of the source.Count()
	test.Equal(500, rw.Code)
	mock.throwError = nil
	mock.throwErrorAll = errors.New("")
	g.Render() // error on source.All()
	test.Equal(500, rw.Code)

	// VTable without header
	c, rw = newController(httptest.NewRequest("GET", "https://localhost/users?sort=id&noheader=1", strings.NewReader("")))
	g = grid2.New(c)
	mock = &mockSource{}
	err = g.SetSource(mock)
	test.NoError(err)
	g.Render()
	test.Equal(200, rw.Code)
	test.Nil(c.Context().Response.Data("head"))
	test.Equal("some data", c.Context().Response.Data("data"))
	test.Equal("*grid2.pagination", reflect.TypeOf(c.Context().Response.Data("pagination")).String())
}

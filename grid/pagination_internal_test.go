package grid

import (
	"errors"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockSource struct {
}

func (m *mockSource) Init(grid *Grid) error {
	return nil
}

// Fields of the grid.
func (m *mockSource) Fields(grid *Grid) ([]Field, error) {
	return nil, nil
}

// UpdatedFields is called before render. The grid fields have the user updated configurations.
func (m *mockSource) UpdatedFields(grid *Grid) error {
	return nil
}

// Callback is called on a callback request of the grid.
func (m *mockSource) Callback(callback string, grid *Grid) (interface{}, error) {
	return nil, nil
}

// One request a single row by the given condition.
func (m *mockSource) One(c *sqlquery.Condition, grid *Grid) (interface{}, error) {
	return nil, nil
}

// All data by the given condition.
func (m *mockSource) All(c *sqlquery.Condition, grid *Grid) (interface{}, error) {
	return nil, nil
}

// Create the object
func (m *mockSource) Create(grid *Grid) (interface{}, error) {
	return nil, nil
}

// Update the object
func (m *mockSource) Update(grid *Grid) error {
	return nil
}

// Delete the object by the given condition.
func (m *mockSource) Delete(c *sqlquery.Condition, grid *Grid) error {
	return nil
}

// Count all the existing object by the given condition.
func (m *mockSource) Count(c *sqlquery.Condition) (int, error) {
	if c.Config(false, sqlquery.WHERE) == "" {
		return 10, nil
	}
	return 10, errors.New("test error")
}

func TestPagination_next(t *testing.T) {
	p := pagination{}

	p.CurrentPage = 1
	p.TotalPages = 10
	assert.Equal(t, 2, p.next())

	p.CurrentPage = 9
	p.TotalPages = 10
	assert.Equal(t, 10, p.next())

	p.CurrentPage = 10
	p.TotalPages = 10
	assert.Equal(t, 0, p.next())

	p.CurrentPage = 11
	p.TotalPages = 10
	assert.Equal(t, 0, p.next())
}

func TestPagination_prev(t *testing.T) {
	p := pagination{}
	p.CurrentPage = -10
	p.TotalPages = 10
	assert.Equal(t, 0, p.prev())

	p.CurrentPage = 0
	p.TotalPages = 10
	assert.Equal(t, 0, p.prev())

	p.CurrentPage = 1
	p.TotalPages = 10
	assert.Equal(t, 0, p.prev())

	p.CurrentPage = 10
	p.TotalPages = 10
	assert.Equal(t, 9, p.prev())

	p.CurrentPage = 100
	p.TotalPages = 10
	assert.Equal(t, 10, p.prev())
}

func TestPagination_offset(t *testing.T) {
	p := pagination{}
	p.Limit = 15

	p.CurrentPage = 0
	assert.Equal(t, 0, p.offset())

	p.CurrentPage = 1
	assert.Equal(t, 0, p.offset())

	p.CurrentPage = 2
	assert.Equal(t, 15, p.offset())

	p.CurrentPage = 3
	assert.Equal(t, 30, p.offset())
}

func TestPagination_totalPages(t *testing.T) {
	p := pagination{}

	p.Total = 10
	p.Limit = 5
	assert.Equal(t, 2, p.totalPages())
	p.Total = 11
	p.Limit = 5
	assert.Equal(t, 3, p.totalPages())
	p.Total = 15
	p.Limit = 5
	assert.Equal(t, 3, p.totalPages())
	p.Total = 16
	p.Limit = 5
	assert.Equal(t, 4, p.totalPages())

	// no rows exist
	p.Total = 0
	assert.Equal(t, 1, p.totalPages())

	// infinity limit
	p.Total = 15
	p.Limit = -1
	assert.Equal(t, 1, p.totalPages())
}

func TestPagination_generate(t *testing.T) {

	body := strings.NewReader("")
	r := httptest.NewRequest("GET", "https://localhost/users", body)
	g := New(newController(r))

	err := g.SetSource(&mockSource{})
	assert.NoError(t, err)

	p, err := g.newPagination(nil)
	g.controller.Set("pagination", p)
	assert.NoError(t, err)
	assert.Equal(t, &pagination{Limit: defaultLimit, Prev: 0, Next: 0, CurrentPage: 1, Total: 10, TotalPages: 1}, g.controller.Context().Response.Data("pagination"))

	// mock is throwing an error if a condition exist.
	p, err = g.newPagination(sqlquery.NewCondition().Where("id > 1"))
	g.controller.Set("pagination", p)
	assert.Error(t, err)
	assert.Nil(t, g.controller.Context().Response.Data("pagination"))

}

func TestPagination_paginationParam(t *testing.T) {
	body := strings.NewReader("")
	r := httptest.NewRequest("GET", "https://localhost/users?limit=5&page=2", body)
	g := New(newController(r))

	p := pagination{}
	assert.Equal(t, 5, p.paginationParam(g, "limit"))
	assert.Equal(t, 2, p.paginationParam(g, "page"))

	body = strings.NewReader("")
	r = httptest.NewRequest("GET", "https://localhost/users", body)
	g = New(newController(r))
	p = pagination{}
	assert.Equal(t, defaultLimit, p.paginationParam(g, "limit"))
	assert.Equal(t, 1, p.paginationParam(g, "page"))
}

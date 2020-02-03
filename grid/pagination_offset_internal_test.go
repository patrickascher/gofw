package grid

import (
	"github.com/patrickascher/gofw/orm"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPagination_next(t *testing.T) {
	p := paginationOffset{}
	p.CurrentPage = 1
	p.TotalPages = 10

	assert.Equal(t, 2, p.next())

	p.CurrentPage = 10
	p.TotalPages = 10
	assert.Equal(t, 0, p.next())
}

func TestPagination_prev(t *testing.T) {
	p := paginationOffset{}
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
	p := paginationOffset{}
	p.CurrentPage = 1
	p.Limit = 15

	assert.Equal(t, 0, p.offset())

	p.CurrentPage = 2
	assert.Equal(t, 15, p.offset())

	p.CurrentPage = 3
	assert.Equal(t, 30, p.offset())
}

func TestPagination_totalPages(t *testing.T) {
	p := paginationOffset{}
	p.Total = 15
	p.Limit = 5
	assert.Equal(t, 3, p.totalPages())

	p.Total = 0
	assert.Equal(t, 1, p.totalPages())

	p.Total = 15
	p.Limit = -1
	assert.Equal(t, 1, p.totalPages())
}

func TestPagination_generate(t *testing.T) {

	err := deleteAll()
	if assert.NoError(t, err) {
		err = insertAll()
		if assert.NoError(t, err) {
			body := strings.NewReader("")
			r := httptest.NewRequest("GET", "https://localhost/users", body)
			g := defaultGrid(r)

			orm.GlobalBuilder, _ = HelperCreateBuilder()

			cust := Customerfk{}
			err := g.Source(&cust)
			assert.NoError(t, err)

			p := paginationOffset{}
			p.generate(g, nil)
			assert.NoError(t, err)

			assert.Equal(t, &paginationOffset{Limit: PaginationDefaultLimit, Prev: 0, Next: 0, CurrentPage: 1, Total: 5, TotalPages: 1}, g.controller.Context().Response.Data("pagination"))
		}
	}
}

func TestPagination_paginationParam(t *testing.T) {
	body := strings.NewReader("")
	r := httptest.NewRequest("GET", "https://localhost/users?limit=5&page=2", body)
	g := defaultGrid(r)

	p := paginationOffset{}
	assert.Equal(t, 5, p.paginationParam(g, "limit"))
	assert.Equal(t, 2, p.paginationParam(g, "page"))

	body = strings.NewReader("")
	r = httptest.NewRequest("GET", "https://localhost/users", body)
	g = defaultGrid(r)
	p = paginationOffset{}
	assert.Equal(t, PaginationDefaultLimit, p.paginationParam(g, "limit"))
	assert.Equal(t, 1, p.paginationParam(g, "page"))
}

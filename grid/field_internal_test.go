package grid

import (
	"github.com/patrickascher/gofw/orm"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestField_setValueHelper(t *testing.T) {

	v := &value{}
	setValueHelper(v, "test")
	assert.Equal(t, "test", v.grid)
	assert.Equal(t, "test", v.create)
	assert.Equal(t, "test", v.update)
	assert.Equal(t, "test", v.details)

	v2 := &value{}
	v2.details = "details"
	v2.grid = "grid"
	v2.create = "create"
	v2.update = "update"

	setValueHelper(v, v2)
	assert.Equal(t, "grid", v.grid)
	assert.Equal(t, "create", v.create)
	assert.Equal(t, "update", v.update)
	assert.Equal(t, "details", v.details)
}

func TestField_title(t *testing.T) {
	r := httptest.NewRequest("GET", "https://localhost/users", strings.NewReader(""))
	g := defaultGrid(r)
	c := defaultCommon(g)
	c.setTitle("title")
	assert.Equal(t, "title", c.getTitle())
}
func TestField_description(t *testing.T) {
	r := httptest.NewRequest("GET", "https://localhost/users", strings.NewReader(""))
	g := defaultGrid(r)
	c := defaultCommon(g)
	c.setDescription("desc")
	assert.Equal(t, "desc", c.getDescription())
}
func TestField_position(t *testing.T) {
	r := httptest.NewRequest("GET", "https://localhost/users", strings.NewReader(""))
	g := defaultGrid(r)
	c := defaultCommon(g)
	c.setPosition(1)
	assert.Equal(t, 1, c.getPosition())
}
func TestField_hide(t *testing.T) {
	r := httptest.NewRequest("GET", "https://localhost/users", strings.NewReader(""))
	g := defaultGrid(r)
	c := defaultCommon(g)
	c.setHide(true)
	assert.Equal(t, true, c.getHide())
}
func TestField_remove(t *testing.T) {
	r := httptest.NewRequest("GET", "https://localhost/users", strings.NewReader(""))
	g := defaultGrid(r)
	c := defaultCommon(g)
	c.setRemove(true)
	assert.Equal(t, true, c.getRemove())
}
func TestField_sort(t *testing.T) {
	r := httptest.NewRequest("GET", "https://localhost/users", strings.NewReader(""))
	g := defaultGrid(r)
	c := defaultCommon(g)
	c.setSort(true)
	assert.Equal(t, true, c.getSort())
}
func TestField_filter(t *testing.T) {
	r := httptest.NewRequest("GET", "https://localhost/users", strings.NewReader(""))
	g := defaultGrid(r)
	c := defaultCommon(g)
	c.setFilter(true)
	assert.Equal(t, true, c.getFilter())
}

func TestField_column(t *testing.T) {
	r := httptest.NewRequest("GET", "https://localhost/users", strings.NewReader(""))
	g := defaultGrid(r)
	c := defaultCommon(g)
	col := &orm.Field{}
	c.column = col
	assert.Equal(t, col, c.getColumn())
}

func TestField_defaultCommon(t *testing.T) {
	r := httptest.NewRequest("GET", "https://localhost/users", strings.NewReader(""))
	g := defaultGrid(r)
	c := defaultCommon(g)

	assert.Equal(t, "", c.getTitle())
	assert.Equal(t, "", c.getDescription())
	assert.Equal(t, false, c.getRemove())
	assert.Equal(t, false, c.getHide())
	assert.Equal(t, 0, c.getPosition())
	assert.Equal(t, true, c.filter)
	assert.Equal(t, true, c.sort)
}

func TestField_defaultField(t *testing.T) {
	r := httptest.NewRequest("GET", "https://localhost/users", strings.NewReader(""))
	g := defaultGrid(r)
	f := defaultField(g)

	assert.Equal(t, "", f.getTitle())
	assert.Equal(t, "", f.getDescription())
	assert.Equal(t, false, f.getRemove())
	assert.Equal(t, false, f.getHide())
	assert.Equal(t, 0, f.getPosition())
	assert.Equal(t, true, f.filter)
	assert.Equal(t, true, f.sort)
}

func TestField_defaultRelation(t *testing.T) {
	r := httptest.NewRequest("GET", "https://localhost/users", strings.NewReader(""))
	g := defaultGrid(r)
	rel := defaultRelation(g)

	assert.Equal(t, "", rel.getTitle())
	assert.Equal(t, "", rel.getDescription())
	assert.Equal(t, false, rel.getRemove())
	assert.Equal(t, false, rel.getHide())
	assert.Equal(t, 0, rel.getPosition())
	assert.Equal(t, true, rel.filter)
	assert.Equal(t, true, rel.sort)
	assert.Equal(t, 0, len(rel.fields))
}

func TestField_getFields_Field(t *testing.T) {
	r := httptest.NewRequest("GET", "https://localhost/users", strings.NewReader(""))
	g := defaultGrid(r)
	f := defaultField(g)
	assert.Equal(t, 0, len(f.getFields()))
}

func TestField_getFields_Relation(t *testing.T) {
	r := httptest.NewRequest("GET", "https://localhost/users", strings.NewReader(""))
	g := defaultGrid(r)
	rel := defaultRelation(g)
	assert.Equal(t, 0, len(rel.getFields()))
}

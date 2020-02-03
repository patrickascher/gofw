package grid

import (
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestField_Field_Set(t *testing.T) {
	r := httptest.NewRequest("GET", "https://localhost/users", strings.NewReader(""))
	g := defaultGrid(r)
	f := defaultField(g)

	f.SetTitle("title")
	assert.Equal(t, "title", f.getTitle())

	f.SetDescription("desc")
	assert.Equal(t, "desc", f.getDescription())

	f.SetPosition(1)
	assert.Equal(t, 1, f.getPosition())

	f.SetHide(true)
	assert.Equal(t, true, f.getHide())

	f.SetRemove(true)
	assert.Equal(t, true, f.getRemove())

	f.SetSort(false)
	assert.Equal(t, false, f.getSort())

	f.SetFilter(false)
	assert.Equal(t, false, f.getFilter())
}

func TestField_Relation_Set(t *testing.T) {
	r := httptest.NewRequest("GET", "https://localhost/users", strings.NewReader(""))
	g := defaultGrid(r)
	rel := defaultRelation(g)

	rel.SetTitle("title")
	assert.Equal(t, "title", rel.getTitle())

	rel.SetDescription("desc")
	assert.Equal(t, "desc", rel.getDescription())

	rel.SetPosition(1)
	assert.Equal(t, 1, rel.getPosition())

	rel.SetHide(true)
	assert.Equal(t, true, rel.getHide())

	rel.SetRemove(true)
	assert.Equal(t, true, rel.getRemove())

	rel.SetSort(false)
	assert.Equal(t, false, rel.getSort())

	rel.SetFilter(false)
	assert.Equal(t, false, rel.getFilter())

	_, err := rel.Relation("xxx")
	assert.Error(t, err)
	_, err = rel.Field("xxx")
	assert.Error(t, err)

	rel.fields["id"] = defaultField(g)
	rel.fields["orders"] = defaultRelation(g)

	fieldX, err := rel.Field("id")
	assert.NoError(t, err)
	assert.True(t, fieldX != nil)
	relationX, err := rel.Relation("orders")
	assert.NoError(t, err)
	assert.True(t, relationX != nil)

}

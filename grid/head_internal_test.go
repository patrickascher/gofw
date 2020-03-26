package grid

import (
	"github.com/patrickascher/gofw/orm"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHead_sortHeaderInfo(t *testing.T) {
	body := strings.NewReader("")
	r := httptest.NewRequest("GET", "https://localhost/users", body)
	g := defaultGrid(r)

	fields := map[string]Interface{}
	fields["id"] = defaultField(g)
	fields["firstName"] = defaultField(g)
	fields["lastName"] = defaultField(g)

	fields["id"].setPosition(1)
	fields["firstName"].setPosition(3)
	fields["lastName"].setPosition(2)

	pos := sortHeaderInfo(fields)
	assert.Equal(t, position{pos: 1, field: "id"}, pos[0])
	assert.Equal(t, position{pos: 2, field: "lastName"}, pos[1])
	assert.Equal(t, position{pos: 3, field: "firstName"}, pos[2])
}

func TestHead_headerFieldsLoop(t *testing.T) {

	body := strings.NewReader("")
	r := httptest.NewRequest("GET", "https://localhost/users", body)
	g := defaultGrid(r)

	b, err := HelperCreateBuilder()
	assert.NoError(t, err)

	orm.GlobalBuilder = &b
	custom := Customerfk{}
	err = g.SetSource(&custom, nil)
	assert.NoError(t, err)

	//remove LastName
	l, err := g.Field("LastName")
	assert.NoError(t, err)
	l.SetRemove(true)

	// relation
	rel, err := g.Relation("Info")
	assert.NoError(t, err)
	rel.SetPosition(8)
	rel, err = g.Relation("Orders")
	assert.NoError(t, err)
	rel.SetPosition(9)
	rel, err = g.Relation("Service")
	assert.NoError(t, err)
	rel.SetPosition(10)
	rel, err = g.Relation("Account")
	assert.NoError(t, err)
	rel.SetPosition(11)

	head := headerFieldsLoop(g.fields, false)

	assert.Equal(t, "ID", head[0].Title)
	assert.Equal(t, "", head[0].Description)
	//assert.Equal(t, 1, head[0].Position)
	//assert.Equal(t, "Int", head[0].FieldType)
	assert.Equal(t, true, head[0].Filter)
	assert.Equal(t, true, head[0].Sort)
	assert.Equal(t, false, head[0].Remove)
	assert.Equal(t, true, head[0].Hide)
	assert.Equal(t, 0, len(head[0].Fields))
	assert.Equal(t, true, head[0].FieldPrimary)
	assert.Equal(t, "", head[0].FieldDefault)

	assert.Equal(t, "FirstName", head[1].Title)
	assert.Equal(t, "", head[1].Description)
	//assert.Equal(t, 2, head[1].Position)
	//	assert.Equal(t, "Text", head[1].FieldType)
	assert.Equal(t, true, head[1].Filter)
	assert.Equal(t, true, head[1].Sort)
	assert.Equal(t, false, head[1].Remove)
	assert.Equal(t, false, head[1].Hide)
	assert.Equal(t, 0, len(head[1].Fields))
	assert.Equal(t, false, head[1].FieldPrimary)
	assert.Equal(t, "", head[1].FieldDefault)

	// check if LastName was removed
	assert.True(t, head[2].Title != "LastName")

	assert.Equal(t, "Info", head[2].Title)
	assert.Equal(t, 3, len(head[2].Fields))

	//assert.Equal(t, 3, head[2].Position)

	assert.Equal(t, "Orders", head[3].Title)
	assert.Equal(t, 3, len(head[3].Fields))

	//assert.Equal(t, 4, head[3].Position)

	assert.Equal(t, "Service", head[4].Title)
	assert.Equal(t, 2, len(head[4].Fields))

	//assert.Equal(t, 5, head[4].Position)

	assert.Equal(t, "Account", head[5].Title)
	assert.Equal(t, 2, len(head[5].Fields))

	//assert.Equal(t, 6, head[5].Position)
}

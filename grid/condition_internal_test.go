package grid

import (
	"fmt"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGrid_conditionOne(t *testing.T) {
	test := assert.New(t)

	grid := New(newController(httptest.NewRequest("GET", "https://localhost/users", strings.NewReader(""))))

	// no params were set
	c, err := grid.conditionOne()
	test.Nil(c)
	test.Error(err)
	test.Equal(errPrimaryMissing.Error(), err.Error())

	// no primary is set
	grid = New(newController(httptest.NewRequest("GET", "https://localhost/users?xy=1", strings.NewReader(""))))
	grid.fields = append(grid.fields, Field{id: "ID", referenceId: "id"})
	c, err = grid.conditionOne()
	test.Nil(c)
	test.Error(err)
	test.Equal(errPrimaryMissing.Error(), err.Error())

	// primary not found in param
	grid = New(newController(httptest.NewRequest("GET", "https://localhost/users?xy=1", strings.NewReader(""))))
	grid.fields = append(grid.fields, Field{id: "ID", primary: true, referenceId: "id"})
	c, err = grid.conditionOne()
	test.Nil(c)
	test.Error(err)
	test.Equal(errPrimaryMissing.Error(), err.Error())

	// primary  found in param
	grid = New(newController(httptest.NewRequest("GET", "https://localhost/users?ID=1", strings.NewReader(""))))
	grid.fields = append(grid.fields, Field{id: "ID", primary: true, referenceId: "id"})
	c, err = grid.conditionOne()
	test.NotNil(c)
	test.NoError(err)
	test.Equal("WHERE id = 1", c.Config(true, sqlquery.WHERE))

	// pre defined sql where
	grid = New(newController(httptest.NewRequest("GET", "https://localhost/users?ID=1", strings.NewReader(""))))
	grid.fields = append(grid.fields, Field{id: "ID", primary: true, referenceId: "id"})
	grid.SetCondition(sqlquery.NewCondition().Where("a=b"))
	c, err = grid.conditionOne()
	test.NotNil(c)
	test.NoError(err)
	test.Equal("WHERE a=b AND id = 1", c.Config(true, sqlquery.WHERE))

	// skip if its a relation (error because no primary was defined)
	grid = New(newController(httptest.NewRequest("GET", "https://localhost/users?ID=1", strings.NewReader(""))))
	grid.fields = append(grid.fields, Field{id: "ID", primary: true, referenceId: "id", fields: []Field{{id: "child"}}})
	c, err = grid.conditionOne()
	test.Nil(c)
	test.Error(err)
}

func TestGrid_conditionAll(t *testing.T) {
	test := assert.New(t)

	grid := New(newController(httptest.NewRequest("GET", "https://localhost/users", strings.NewReader(""))))

	// no params were set
	c, err := grid.conditionAll()
	test.Equal("", c.Config(true, sqlquery.WHERE))
	test.NoError(err)

	// no params were set - with default condition
	grid.SetCondition(sqlquery.NewCondition().Where("a=b"))
	c, err = grid.conditionAll()
	test.Equal("WHERE a=b", c.Config(true, sqlquery.WHERE))
	test.NoError(err)

	// field does not exist
	grid = New(newController(httptest.NewRequest("GET", "https://localhost/users?sort=abc", strings.NewReader(""))))
	grid.fields = append(grid.fields, Field{id: "ID", primary: true, referenceId: "id"})
	c, err = grid.conditionAll()
	test.Error(err)
	test.Equal(fmt.Sprintf(errSortPermission, "abc"), err.Error())

	// field has no sort permission
	grid = New(newController(httptest.NewRequest("GET", "https://localhost/users?sort=ID", strings.NewReader(""))))
	grid.fields = append(grid.fields, Field{id: "ID", primary: true, referenceId: "id"})
	c, err = grid.conditionAll()
	test.Error(err)
	test.Equal(fmt.Sprintf(errSortPermission, "ID"), err.Error())

	// field order permission ok
	grid = New(newController(httptest.NewRequest("GET", "https://localhost/users?sort=ID", strings.NewReader(""))))
	grid.fields = append(grid.fields, Field{id: "ID", sortable: true, primary: true, referenceId: "id"})
	c, err = grid.conditionAll()
	test.NoError(err)
	test.Equal("ORDER BY id ASC", c.Config(true, sqlquery.ORDER))

	// testing if pre definde order is getting reset
	grid = New(newController(httptest.NewRequest("GET", "https://localhost/users?sort=ID", strings.NewReader(""))))
	grid.SetCondition(sqlquery.NewCondition().Order("test"))
	grid.fields = append(grid.fields, Field{id: "ID", sortable: true, primary: true, referenceId: "id"})
	c, err = grid.conditionAll()
	test.NoError(err)
	test.Equal("ORDER BY id ASC", c.Config(true, sqlquery.ORDER))

	// testing desc ordering
	grid = New(newController(httptest.NewRequest("GET", "https://localhost/users?sort=-ID", strings.NewReader(""))))
	grid.SetCondition(sqlquery.NewCondition().Order("test"))
	grid.fields = append(grid.fields, Field{id: "ID", sortable: true, primary: true, referenceId: "id"})
	c, err = grid.conditionAll()
	test.NoError(err)
	test.Equal("ORDER BY id DESC", c.Config(true, sqlquery.ORDER))

	// testing empty sort param
	grid = New(newController(httptest.NewRequest("GET", "https://localhost/users?sort", strings.NewReader(""))))
	grid.SetCondition(sqlquery.NewCondition().Order("test"))
	grid.fields = append(grid.fields, Field{id: "ID", sortable: true, primary: true, referenceId: "id"})
	c, err = grid.conditionAll()
	test.NoError(err)
	test.Equal("", c.Config(true, sqlquery.ORDER))

	// testing filter - field does not exist
	grid = New(newController(httptest.NewRequest("GET", "https://localhost/users?filter_xy=1", strings.NewReader(""))))
	grid.SetCondition(sqlquery.NewCondition().Order("test"))
	grid.fields = append(grid.fields, Field{id: "ID", primary: true, referenceId: "id"})
	c, err = grid.conditionAll()
	test.Error(err)
	test.Equal(fmt.Sprintf(errFilterPermission, "xy"), err.Error())
	test.Nil(c)

	// no permission
	grid = New(newController(httptest.NewRequest("GET", "https://localhost/users?filter_ID=1", strings.NewReader(""))))
	grid.SetCondition(sqlquery.NewCondition().Order("test"))
	grid.fields = append(grid.fields, Field{id: "ID", primary: true, referenceId: "id"})
	c, err = grid.conditionAll()
	test.Error(err)
	test.Equal(fmt.Sprintf(errFilterPermission, "ID"), err.Error())
	test.Nil(c)

	// filter
	grid = New(newController(httptest.NewRequest("GET", "https://localhost/users?filter_ID=1", strings.NewReader(""))))
	grid.SetCondition(sqlquery.NewCondition().Order("test"))
	grid.fields = append(grid.fields, Field{id: "ID", filterable: true, primary: true, referenceId: "id"})
	c, err = grid.conditionAll()
	test.NoError(err)
	test.Equal("WHERE ID = 1", c.Config(true, sqlquery.WHERE))

	// filter with per defined condition
	grid = New(newController(httptest.NewRequest("GET", "https://localhost/users?filter_ID=1", strings.NewReader(""))))
	grid.SetCondition(sqlquery.NewCondition().Where("a=b"))
	grid.fields = append(grid.fields, Field{id: "ID", filterable: true, primary: true, referenceId: "id"})
	c, err = grid.conditionAll()
	test.NoError(err)
	test.Equal("WHERE a=b AND ID = 1", c.Config(true, sqlquery.WHERE))

	// filter with multiple values
	grid = New(newController(httptest.NewRequest("GET", "https://localhost/users?filter_ID=1,2,3", strings.NewReader(""))))
	grid.fields = append(grid.fields, Field{id: "ID", filterable: true, primary: true, referenceId: "id"})
	c, err = grid.conditionAll()
	test.NoError(err)
	test.Equal("WHERE ID IN(1, 2, 3)", c.Config(true, sqlquery.WHERE))
}

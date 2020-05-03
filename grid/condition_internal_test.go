package grid

import (
	"github.com/patrickascher/gofw/orm"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCondition_conditionAll(t *testing.T) {
	body := strings.NewReader("")
	r := httptest.NewRequest("GET", "https://localhost/users?sort=ID,-FirstName&filter_ID=1", body)
	g := defaultGrid(r)
	b, err := HelperCreateBuilder()
	orm.GlobalBuilder = b
	cust := Customerfk{}
	err = g.SetSource(&cust, nil)
	assert.NoError(t, err)
	c, err := conditionAll(g)
	assert.NoError(t, err)
	cExp := sqlquery.Condition{}
	cExp.Order("id", "-first_name")
	cExp.Where("id = ?", "1") // TODO fix params types
	assert.Equal(t, &cExp, c)

	// no params
	r = httptest.NewRequest("GET", "https://localhost/users", body)
	g = defaultGrid(r)
	cust2 := Customerfk{}
	err = g.SetSource(&cust2, nil)
	assert.NoError(t, err)
	c, err = conditionAll(g)
	assert.NoError(t, err)
	cExp = sqlquery.Condition{}
	assert.Equal(t, &cExp, c)

	// sort field does not exist
	r = httptest.NewRequest("GET", "https://localhost/users?sort=xxx", body)
	g = defaultGrid(r)
	cust3 := Customerfk{}
	err = g.SetSource(&cust3, nil)
	assert.NoError(t, err)
	c, err = conditionAll(g)
	assert.Error(t, err)
	assert.True(t, c == nil)

	// filter field does not exist
	r = httptest.NewRequest("GET", "https://localhost/users?filter_xxx=1", body)
	g = defaultGrid(r)
	cust4 := Customerfk{}
	err = g.SetSource(&cust4, nil)
	assert.NoError(t, err)
	c, err = conditionAll(g)
	assert.Error(t, err)
	assert.True(t, c == nil)
}

func TestCondition_addSortCondition(t *testing.T) {
	body := strings.NewReader("")
	r := httptest.NewRequest("GET", "https://localhost/users?sort=ID,-FirstName", body)
	g := defaultGrid(r)
	b, err := HelperCreateBuilder()
	orm.GlobalBuilder = b
	cust := Customerfk{}
	err = g.SetSource(&cust, nil)
	assert.NoError(t, err)

	// single param
	c := &sqlquery.Condition{}
	cExp := &sqlquery.Condition{}
	cExp.Order("id", "-first_name") // TODO: why is here no fqdn?
	err = addSortCondition(g, "ID,-FirstName", c)
	assert.NoError(t, err)
	assert.Equal(t, cExp, c)

	// field does not exist
	c = &sqlquery.Condition{}
	err = addSortCondition(g, "ID,-Name", c)
	assert.Error(t, err)

	// no params
	c = &sqlquery.Condition{}
	err = addSortCondition(g, "", c)
	assert.NoError(t, err)
}

func TestCondition_addFilterCondition(t *testing.T) {
	body := strings.NewReader("")
	r := httptest.NewRequest("GET", "https://localhost/users?filter_id=1", body)
	g := defaultGrid(r)
	b, err := HelperCreateBuilder()
	orm.GlobalBuilder = b
	cust := Customerfk{}
	err = g.SetSource(&cust, nil)
	assert.NoError(t, err)

	// single param
	c := &sqlquery.Condition{}
	cExp := &sqlquery.Condition{}
	cExp.Where("id = ?", "1") // TODO: why is here no fqdn?
	err = addFilterCondition(g, "ID", []string{"1"}, c)
	assert.NoError(t, err)
	assert.Equal(t, cExp, c)

	body = strings.NewReader("")
	r = httptest.NewRequest("GET", "https://localhost/users?filter_id=1,2,3", body)
	g = defaultGrid(r)
	cust2 := Customerfk{} //TODO FIX: why is a redeclaring not working?
	err = g.SetSource(&cust2, nil)
	assert.NoError(t, err)
	// multiple params
	c = &sqlquery.Condition{}
	cExp = &sqlquery.Condition{}
	cExp.Where("id IN(?)", []string{"1", "2", "3"}) // TODO: why is here no fqdn?
	err = addFilterCondition(g, "ID", []string{"1,2,3"}, c)
	assert.NoError(t, err)
	assert.Equal(t, cExp, c)
}

func TestCondition_isSortAllowed(t *testing.T) {
	body := strings.NewReader("")
	r := httptest.NewRequest("GET", "https://localhost/users", body)
	g := defaultGrid(r)
	b, err := HelperCreateBuilder()
	orm.GlobalBuilder = b
	cust := Customerfk{}
	err = g.SetSource(&cust, nil)
	assert.NoError(t, err)

	// allowed
	allowed, err := isSortAllowed(g, "ID")
	assert.Equal(t, "id", allowed)

	assert.NoError(t, err)

	// not allowed
	f, err := g.Field("ID")
	assert.NoError(t, err)
	f.SetSort(false)
	allowed, err = isSortAllowed(g, "ID")
	assert.Error(t, err)
}

func TestCondition_isFilterAllowed(t *testing.T) {
	body := strings.NewReader("")
	r := httptest.NewRequest("GET", "https://localhost/users", body)
	g := defaultGrid(r)
	b, err := HelperCreateBuilder()
	orm.GlobalBuilder = b
	cust := Customerfk{}
	err = g.SetSource(&cust, nil)
	assert.NoError(t, err)

	// allowed
	allowed, err := isFilterAllowed(g, "ID")
	assert.NoError(t, err)
	assert.Equal(t, "id", allowed)

	// not allowed
	f, err := g.Field("ID")
	assert.NoError(t, err)
	f.SetFilter(false)
	allowed, err = isFilterAllowed(g, "ID")
	assert.Error(t, err)
}

func TestCondition_getFieldByDbName(t *testing.T) {
	body := strings.NewReader("")
	r := httptest.NewRequest("GET", "https://localhost/users", body)
	g := defaultGrid(r)
	b, err := HelperCreateBuilder()
	orm.GlobalBuilder = b
	cust := Customerfk{}
	err = g.SetSource(&cust, nil)
	assert.NoError(t, err)

	// exists
	f, err := g.getFieldByName("ID")
	assert.NoError(t, err)
	assert.Equal(t, "ID", f.getTitle())

	// not existing
	f, err = g.getFieldByName("xxx")
	assert.Error(t, err)
	assert.True(t, f == nil)
}

func TestCondition_checkPrimaryParams(t *testing.T) {
	body := strings.NewReader("")
	r := httptest.NewRequest("GET", "https://localhost/users", body)
	g := defaultGrid(r)
	b, err := HelperCreateBuilder()
	orm.GlobalBuilder = b
	cust := Customerfk{}
	err = g.SetSource(&cust, nil)
	assert.NoError(t, err)
	// err no params are set
	c, err := checkPrimaryParams(g)
	assert.Error(t, err)
	assert.True(t, c == nil)

	body = strings.NewReader("")
	r = httptest.NewRequest("GET", "https://localhost/users?test=1", body)
	g = defaultGrid(r)
	cust2 := Customerfk{}
	err = g.SetSource(&cust2, nil)
	assert.NoError(t, err)
	// err pkey is not set
	c, err = checkPrimaryParams(g)
	assert.Error(t, err)
	assert.True(t, c == nil)

	body = strings.NewReader("")
	r = httptest.NewRequest("GET", "https://localhost/users?ID=1", body)
	g = defaultGrid(r)
	cust3 := Customerfk{}
	err = g.SetSource(&cust3, nil)
	assert.NoError(t, err)
	// everything ok
	c, err = checkPrimaryParams(g)
	assert.NoError(t, err)
	con := sqlquery.Condition{}
	con.Where("customerfks.id = ?", "1") // TODO fix?
	assert.Equal(t, &con, c)
}

package grid

import (
	"bytes"
	"github.com/patrickascher/gofw/cache"
	"github.com/patrickascher/gofw/cache/memory"
	"github.com/patrickascher/gofw/controller"
	"github.com/patrickascher/gofw/controller/context"
	"github.com/patrickascher/gofw/orm"
	"github.com/patrickascher/gofw/sqlquery"
	_ "github.com/patrickascher/gofw/sqlquery/driver"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestGrid_New(t *testing.T) {
	c := controller.Controller{}
	grid := New(&c)
	assert.Equal(t, "*grid.Grid", reflect.TypeOf(grid).String())
}

func TestGrid_Source(t *testing.T) {

	// error because no cache is set
	c := controller.Controller{}
	grid := New(&c)
	customer := Customerfk{}
	err := grid.SetSource(&customer, nil)
	assert.Error(t, err)

	// everything ok - createFields is tested in an extra test case
	c = controller.Controller{}
	cache, _ := cache.New("memory", memory.Options{GCInterval: 5 * time.Minute})
	builder, err := HelperCreateBuilder()
	assert.NoError(t, err)
	orm.GlobalBuilder = &builder
	c.SetCache(cache)

	r := httptest.NewRequest("GET", "https://localhost/users?sort=ID,-FirstName&filter_ID=1", nil)
	rw := httptest.NewRecorder()
	ctx := context.New(r, rw)
	c.SetContext(ctx)

	grid = New(&c)
	customer2 := Customerfk{}
	err = grid.SetSource(&customer2, nil)
	assert.NoError(t, err)
	assert.True(t, grid.sourceAdded)
}

func TestGrid_Mode(t *testing.T) {

	r := httptest.NewRequest("GET", "https://localhost/users", strings.NewReader(""))
	g := defaultGrid(r)
	assert.Equal(t, ViewGrid, g.Mode())

	r = httptest.NewRequest("GET", "https://localhost/users?mode=new", strings.NewReader(""))
	g = defaultGrid(r)
	assert.Equal(t, ViewCreate, g.Mode())

	r = httptest.NewRequest("GET", "https://localhost/users?mode=details&id=1", strings.NewReader(""))
	g = defaultGrid(r)
	assert.Equal(t, ViewDetails, g.Mode())

	r = httptest.NewRequest("GET", "https://localhost/users?mode=edit&id=1", strings.NewReader(""))
	g = defaultGrid(r)
	assert.Equal(t, ViewEdit, g.Mode())

	r = httptest.NewRequest("POST", "https://localhost/users", strings.NewReader(""))
	g = defaultGrid(r)
	assert.Equal(t, CREATE, g.Mode())

	r = httptest.NewRequest("PUT", "https://localhost/users", strings.NewReader(""))
	g = defaultGrid(r)
	assert.Equal(t, UPDATE, g.Mode())

	r = httptest.NewRequest("DELETE", "https://localhost/users", strings.NewReader(""))
	g = defaultGrid(r)
	assert.Equal(t, DELETE, g.Mode())

	r = httptest.NewRequest("OPTION", "https://localhost/users", strings.NewReader(""))
	g = defaultGrid(r)
	assert.Equal(t, 0, g.Mode())
}

func TestGrid_Disable(t *testing.T) {
	r := httptest.NewRequest("GET", "https://localhost/users", strings.NewReader(""))
	g := defaultGrid(r)

	assert.False(t, g.disablePagination)
	assert.False(t, g.disableHeader)
	err := g.Disable(PAGINATION)
	assert.NoError(t, err)
	assert.True(t, g.disablePagination)
	assert.False(t, g.disableHeader)
	err = g.Disable(HEADER)
	assert.NoError(t, err)
	assert.True(t, g.disablePagination)
	assert.True(t, g.disableHeader)

	err = g.Disable("xxx")
	assert.Error(t, err)
}

func TestGrid_Field_Relation(t *testing.T) {
	body := strings.NewReader("")
	r := httptest.NewRequest("GET", "https://localhost/users", body)
	g := defaultGrid(r)

	b, err := HelperCreateBuilder()
	assert.NoError(t, err)
	orm.GlobalBuilder = &b

	cust := Customerfk{}
	err = g.SetSource(&cust, nil)
	assert.NoError(t, err)

	id, err := g.Field("ID")
	assert.NoError(t, err)
	assert.Equal(t, "ID", id.getTitle())

	notExisting, err := g.Field("xxx")
	assert.Error(t, err)
	assert.True(t, notExisting == nil)

	relInfo, err := g.Relation("Info")
	assert.NoError(t, err)
	assert.Equal(t, "Info", relInfo.getTitle())
	assert.Equal(t, 3, len(relInfo.getFields()))

	relInfo, err = g.Relation("InfXX")
	assert.Error(t, err)
	assert.True(t, relInfo == nil)
}

func TestGrid_getRelationName(t *testing.T) {

	type Adr struct {
		ID int
	}

	type User struct {
		AdrPtr  *Adr
		Adr     Adr
		Adrs    []Adr
		AdrsPtr []*Adr
	}

	user := User{}
	assert.Equal(t, "grid.Adr", getRelationName(reflect.ValueOf(user.Adr)))
	assert.Equal(t, "grid.Adr", getRelationName(reflect.ValueOf(user.AdrPtr)))
	assert.Equal(t, "grid.Adr", getRelationName(reflect.ValueOf(user.Adrs)))
	assert.Equal(t, "grid.Adr", getRelationName(reflect.ValueOf(user.AdrsPtr)))
}

func TestGrid_httpMethod(t *testing.T) {
	body := strings.NewReader("")
	r := httptest.NewRequest("GET", "https://localhost/users", body)
	g := defaultGrid(r)

	assert.Equal(t, "GET", g.httpMethod())
}

func TestGrid_createFields(t *testing.T) {
	body := strings.NewReader("")
	r := httptest.NewRequest("GET", "https://localhost/users", body)
	g := defaultGrid(r)

	b, err := HelperCreateBuilder()
	assert.NoError(t, err)
	orm.GlobalBuilder = &b

	cust := Customerfk{}
	err = g.SetSource(&cust, nil)
	assert.NoError(t, err)

	assert.Equal(t, 8, len(g.fields))

	assert.Equal(t, "ID", g.fields["ID"].getTitle())
	assert.Equal(t, "", g.fields["ID"].getDescription())
	assert.Equal(t, 1, g.fields["ID"].getPosition())
	assert.Equal(t, true, g.fields["ID"].getHide())
	assert.Equal(t, false, g.fields["ID"].getRemove())
	assert.Equal(t, true, g.fields["ID"].getSort())
	assert.Equal(t, true, g.fields["ID"].getFilter())
	assert.Equal(t, "Integer", g.fields["ID"].getFieldType().Name())
	assert.True(t, g.fields["ID"].getFields() == nil)
	assert.True(t, g.fields["ID"].getColumn() != nil)

	assert.Equal(t, "FirstName", g.fields["FirstName"].getTitle())
	assert.Equal(t, "", g.fields["FirstName"].getDescription())
	assert.Equal(t, 2, g.fields["FirstName"].getPosition())
	assert.Equal(t, false, g.fields["FirstName"].getHide())
	assert.Equal(t, false, g.fields["FirstName"].getRemove())
	assert.Equal(t, true, g.fields["FirstName"].getSort())
	assert.Equal(t, true, g.fields["FirstName"].getFilter())
	assert.Equal(t, "Text", g.fields["FirstName"].getFieldType().Name())
	assert.True(t, g.fields["FirstName"].getFields() == nil)
	assert.True(t, g.fields["FirstName"].getColumn() != nil)

	assert.Equal(t, "LastName", g.fields["LastName"].getTitle())
	assert.Equal(t, "", g.fields["LastName"].getDescription())
	assert.Equal(t, 3, g.fields["LastName"].getPosition())
	assert.Equal(t, false, g.fields["LastName"].getHide())
	assert.Equal(t, false, g.fields["LastName"].getRemove())
	assert.Equal(t, true, g.fields["LastName"].getSort())
	assert.Equal(t, true, g.fields["LastName"].getFilter())
	assert.Equal(t, "Text", g.fields["LastName"].getFieldType().Name())
	assert.True(t, g.fields["LastName"].getFields() == nil)
	assert.True(t, g.fields["LastName"].getColumn() != nil)

	/*assert.Equal(t, "CreatedAt", g.fields["CreatedAt"].getTitle())
	assert.Equal(t, "", g.fields["CreatedAt"].getDescription())
	assert.Equal(t, 4, g.fields["CreatedAt"].getPosition())
	assert.Equal(t, false, g.fields["CreatedAt"].getHide())
	assert.Equal(t, false, g.fields["CreatedAt"].getRemove())
	assert.Equal(t, true, g.fields["CreatedAt"].getSort())
	assert.Equal(t, true, g.fields["CreatedAt"].getFilter())
	assert.Equal(t, "DateTime", g.fields["CreatedAt"].getFieldType())
	assert.True(t, g.fields["CreatedAt"].getFields() == nil)
	assert.True(t, g.fields["CreatedAt"].getColumn() != nil)

	assert.Equal(t, "UpdatedAt", g.fields["UpdatedAt"].getTitle())
	assert.Equal(t, "", g.fields["UpdatedAt"].getDescription())
	assert.Equal(t, 5, g.fields["UpdatedAt"].getPosition())
	assert.Equal(t, false, g.fields["UpdatedAt"].getHide())
	assert.Equal(t, false, g.fields["UpdatedAt"].getRemove())
	assert.Equal(t, true, g.fields["UpdatedAt"].getSort())
	assert.Equal(t, true, g.fields["UpdatedAt"].getFilter())
	assert.Equal(t, "DateTime", g.fields["UpdatedAt"].getFieldType())
	assert.True(t, g.fields["UpdatedAt"].getFields() == nil)
	assert.True(t, g.fields["UpdatedAt"].getColumn() != nil)

	assert.Equal(t, "DeletedAt", g.fields["DeletedAt"].getTitle())
	assert.Equal(t, "", g.fields["DeletedAt"].getDescription())
	assert.Equal(t, 6, g.fields["DeletedAt"].getPosition())
	assert.Equal(t, false, g.fields["DeletedAt"].getHide())
	assert.Equal(t, false, g.fields["DeletedAt"].getRemove())
	assert.Equal(t, true, g.fields["DeletedAt"].getSort())
	assert.Equal(t, true, g.fields["DeletedAt"].getFilter())
	assert.Equal(t, "DateTime", g.fields["DeletedAt"].getFieldType())
	assert.True(t, g.fields["DeletedAt"].getFields() == nil)
	assert.True(t, g.fields["DeletedAt"].getColumn() != nil)*/

	assert.Equal(t, "AccountId", g.fields["AccountId"].getTitle())
	assert.Equal(t, "", g.fields["AccountId"].getDescription())
	assert.Equal(t, 7, g.fields["AccountId"].getPosition())
	assert.Equal(t, false, g.fields["AccountId"].getHide())
	assert.Equal(t, true, g.fields["AccountId"].getRemove())
	assert.Equal(t, true, g.fields["AccountId"].getSort())
	assert.Equal(t, true, g.fields["AccountId"].getFilter())
	assert.Equal(t, "Integer", g.fields["AccountId"].getFieldType().Name())
	assert.True(t, g.fields["AccountId"].getFields() == nil)
	assert.True(t, g.fields["AccountId"].getColumn() != nil)

	assert.Equal(t, "Info", g.fields["Info"].getTitle())
	assert.Equal(t, "", g.fields["Info"].getDescription())
	//assert.Equal(t,8,g.fields["Info"].getPosition()) //Position is not checked here because its not fixed yet association is a map.
	assert.Equal(t, false, g.fields["Info"].getHide())
	assert.Equal(t, false, g.fields["Info"].getRemove())
	assert.Equal(t, false, g.fields["Info"].getSort())
	assert.Equal(t, true, g.fields["Info"].getFilter())
	assert.Equal(t, "hasOne", g.fields["Info"].getFieldType().Name())
	assert.Equal(t, 3, len(g.fields["Info"].getFields()))
	assert.Equal(t, "ID", g.fields["Info"].getFields()["ID"].getTitle())
	assert.Equal(t, "CustomerID", g.fields["Info"].getFields()["CustomerID"].getTitle())
	assert.Equal(t, "Phone", g.fields["Info"].getFields()["Phone"].getTitle())

	assert.Equal(t, "Orders", g.fields["Orders"].getTitle())
	assert.Equal(t, "", g.fields["Orders"].getDescription())
	//assert.Equal(t,8,g.fields["Orders"].getPosition()) //Position is not checked here because its not fixed yet association is a map.
	assert.Equal(t, false, g.fields["Orders"].getHide())
	assert.Equal(t, false, g.fields["Orders"].getRemove())
	assert.Equal(t, false, g.fields["Orders"].getSort())
	assert.Equal(t, true, g.fields["Orders"].getFilter())
	assert.Equal(t, "hasMany", g.fields["Orders"].getFieldType().Name())
	assert.Equal(t, 3, len(g.fields["Orders"].getFields()))
	assert.Equal(t, "ID", g.fields["Orders"].getFields()["ID"].getTitle())
	assert.Equal(t, "CustomerID", g.fields["Orders"].getFields()["CustomerID"].getTitle())
	assert.Equal(t, "CreatedAt", g.fields["Orders"].getFields()["CreatedAt"].getTitle())

	assert.Equal(t, "Service", g.fields["Service"].getTitle())
	assert.Equal(t, "", g.fields["Service"].getDescription())
	//assert.Equal(t,8,g.fields["Service"].getPosition()) //Position is not checked here because its not fixed yet association is a map.
	assert.Equal(t, false, g.fields["Service"].getHide())
	assert.Equal(t, false, g.fields["Service"].getRemove())
	assert.Equal(t, false, g.fields["Service"].getSort())
	assert.Equal(t, true, g.fields["Service"].getFilter())
	assert.Equal(t, "manyToMany", g.fields["Service"].getFieldType().Name())
	assert.Equal(t, 2, len(g.fields["Service"].getFields()))
	assert.Equal(t, "ID", g.fields["Service"].getFields()["ID"].getTitle())
	assert.Equal(t, "Name", g.fields["Service"].getFields()["Name"].getTitle())

	assert.Equal(t, "Account", g.fields["Account"].getTitle())
	assert.Equal(t, "", g.fields["Account"].getDescription())
	//assert.Equal(t,8,g.fields["Service"].getPosition()) //Position is not checked here because its not fixed yet association is a map.
	assert.Equal(t, false, g.fields["Account"].getHide())
	assert.Equal(t, false, g.fields["Account"].getRemove())
	assert.Equal(t, false, g.fields["Account"].getSort())
	assert.Equal(t, true, g.fields["Account"].getFilter())
	assert.Equal(t, "belongsTo", g.fields["Account"].getFieldType().Name())
	assert.Equal(t, 2, len(g.fields["Account"].getFields()))
	assert.Equal(t, "ID", g.fields["Account"].getFields()["ID"].getTitle())
	assert.Equal(t, "Name", g.fields["Account"].getFields()["Name"].getTitle())
}

func TestGrid_headerInfo(t *testing.T) {
	body := strings.NewReader("")
	r := httptest.NewRequest("GET", "https://localhost/users", body)
	g := defaultGrid(r)

	b, err := HelperCreateBuilder()
	assert.NoError(t, err)
	orm.GlobalBuilder = &b

	cust := Customerfk{}
	err = g.SetSource(&cust, nil)
	assert.NoError(t, err)

	g.headerInfo()

	assert.True(t, g.controller.Context().Response.Data("head") != nil)
}

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func TestGrid_marshalModel(t *testing.T) {
	body := strings.NewReader("")
	r := httptest.NewRequest("GET", "https://localhost/users", body)
	g := defaultGrid(r)

	b, err := HelperCreateBuilder()
	assert.NoError(t, err)
	orm.GlobalBuilder = &b

	cust := Customerfk{}
	err = g.SetSource(&cust, nil)
	assert.NoError(t, err)

	// invalid json  "Read will..."
	g.controller.Context().Request.Raw().Body = nopCloser{bytes.NewBufferString("Read will...")}
	err = g.unmarshalModel()
	assert.Equal(t, ErrJsonInvalid, err)

	// empty "{}" TODO check empty model to turn this into an error!
	g.controller.Context().Request.Raw().Body = nopCloser{bytes.NewBufferString("{}")}
	err = g.unmarshalModel()
	assert.NoError(t, err)

	// Wrong Field set
	g.controller.Context().Request.Raw().Body = nopCloser{bytes.NewBufferString("{\"NOT\":\"Existing\"}")}
	err = g.unmarshalModel()
	assert.Error(t, err)

	// Everything right
	g.controller.Context().Request.Raw().Body = nopCloser{bytes.NewBufferString("{\"Id\":1}")}
	err = g.unmarshalModel()
	assert.NoError(t, err)

	// Error no model is set
	g.controller.Context().Request.Raw().Body = nopCloser{bytes.NewBufferString("{\"Id\":1}")}
	g.src = nil
	err = g.unmarshalModel()
	assert.Error(t, err)

	// Body is not a Reader
	g.controller.Context().Request.Raw().Body = nil
	g.src = nil
	err = g.unmarshalModel()
	assert.Equal(t, ErrRequestBody, err)
}

func TestGrid_readOne(t *testing.T) {

	err := deleteAll()
	if assert.NoError(t, err) {
		err = insertAll()
		if assert.NoError(t, err) {

			body := strings.NewReader("")
			r := httptest.NewRequest("GET", "https://localhost/users", body)
			g := defaultGrid(r)

			b, err := HelperCreateBuilder()
			assert.NoError(t, err)
			orm.GlobalBuilder = &b

			cust := Customerfk{}
			err = g.SetSource(&cust, nil)
			assert.NoError(t, err)

			c := sqlquery.Condition{}
			c.Where("id = ?", 1)
			g.readOne(&c)

			assert.Equal(t, 1, cust.ID)
			assert.True(t, g.controller.Context().Response.Data("data") != nil)
		}
	}

}

func TestGrid_readAll(t *testing.T) {

	err := deleteAll()
	if assert.NoError(t, err) {
		err = insertAll()
		if assert.NoError(t, err) {

			body := strings.NewReader("")
			r := httptest.NewRequest("GET", "https://localhost/users", body)
			g := defaultGrid(r)

			b, err := HelperCreateBuilder()
			assert.NoError(t, err)
			orm.GlobalBuilder = &b

			cust := Customerfk{}
			err = g.SetSource(&cust, nil)
			assert.NoError(t, err)

			g.readAll()

			counter := reflect.ValueOf(g.controller.Context().Response.Data("data").(*[]Customerfk)).Elem().Interface().([]Customerfk)
			assert.Equal(t, 5, len(counter))
			assert.True(t, g.controller.Context().Response.Data("head") != "")
			assert.True(t, g.controller.Context().Response.Data("pagination") != "")

		}
	}
}

func TestGrid_delete(t *testing.T) {

	err := deleteAll()
	if assert.NoError(t, err) {
		err = insertAll()
		if assert.NoError(t, err) {

			body := strings.NewReader("")
			r := httptest.NewRequest("GET", "https://localhost/users", body)
			g := defaultGrid(r)

			b, err := HelperCreateBuilder()
			assert.NoError(t, err)
			orm.GlobalBuilder = &b

			cust := Customerfk{}
			err = g.SetSource(&cust, nil)
			assert.NoError(t, err)

			c := sqlquery.Condition{}
			c.Where("id = ?", 1)
			err = g.delete(&c)
			assert.NoError(t, err)

			// still the same amount of entries because of softDelete
			c = sqlquery.Condition{}
			count, err := g.src.Count(&c)
			assert.NoError(t, err)
			assert.Equal(t, 5, count)

			c = sqlquery.Condition{}
			c.Where("id = ?", 1)
			err = g.readOne(&c)
			assert.NoError(t, err)
			assert.True(t, cust.DeletedAt.Valid)

		}
	}
}

func TestGrid_create(t *testing.T) {

	err := deleteAll()
	if assert.NoError(t, err) {

		body := strings.NewReader(`{"FirstName":"Test123","LastName":"Stoate","CreatedAt":"2019-02-23T00:00:00Z","UpdatedAt":"2020-03-02T00:00:00Z","DeletedAt":"2020-10-02T00:00:00Z","Info":{"Phone":"000-000-123"},"Orders":[{"CreatedAt":"2010-07-21T00:00:00Z"},{"CreatedAt":"2010-07-22T00:00:00Z"},{"CreatedAt":"2010-07-23T00:00:00Z"}],"Service":[{"Name":"paypal"},{"Name":"banking"},{"Name":"appstore"},{"Name":"playstore"}]}`)
		r := httptest.NewRequest("GET", "https://localhost/users", body)
		g := defaultGrid(r)

		b, err := HelperCreateBuilder()
		assert.NoError(t, err)
		orm.GlobalBuilder = &b

		cust := Customerfk{}
		err = g.SetSource(&cust, nil)
		assert.NoError(t, err)

		err = g.create()
		assert.NoError(t, err)
		assert.Equal(t, "Test123", cust.FirstName.String)
	}
}

func TestGrid_update(t *testing.T) {

	err := deleteAll()
	if assert.NoError(t, err) {
		err = insertAll()
		if assert.NoError(t, err) {
			body := strings.NewReader(`{"ID":1,"FirstName":"TreschaUpd","LastName":"Stoate","CreatedAt":"2019-02-23T00:00:00Z","UpdatedAt":"2020-03-02T00:00:00Z","DeletedAt":"2020-10-02T00:00:00Z","Info":{"Phone":"000-000-123"},"Orders":[{"CreatedAt":"2010-07-21T00:00:00Z"},{"CreatedAt":"2010-07-22T00:00:00Z"},{"CreatedAt":"2010-07-23T00:00:00Z"}],"Service":[{"Name":"paypal"},{"Name":"banking"},{"Name":"appstore"},{"Name":"playstore"}]}`)
			r := httptest.NewRequest("GET", "https://localhost/users", body)
			g := defaultGrid(r)

			b, err := HelperCreateBuilder()
			assert.NoError(t, err)
			orm.GlobalBuilder = &b

			cust := Customerfk{}
			err = g.SetSource(&cust, nil)
			assert.NoError(t, err)

			err = g.update()
			assert.NoError(t, err)
			assert.Equal(t, "TreschaUpd", cust.FirstName.String)
		}
	}
}

func TestGrid_Render(t *testing.T) {

	// ViewGrid
	body := strings.NewReader("")
	r := httptest.NewRequest("GET", "https://localhost/users", body)
	g := defaultGrid(r)

	b, err := HelperCreateBuilder()
	assert.NoError(t, err)
	orm.GlobalBuilder = &b

	cust := Customerfk{}
	err = g.SetSource(&cust, nil)
	assert.NoError(t, err)
	g.Render()
	assert.True(t, g.controller.Context().Response.Data("data") != nil)

	// ViewCreate
	body = strings.NewReader("")
	r = httptest.NewRequest("GET", "https://localhost/users?mode=new", body)
	g = defaultGrid(r)

	b, err = HelperCreateBuilder()
	assert.NoError(t, err)
	orm.GlobalBuilder = &b

	cust2 := Customerfk{}
	err = g.SetSource(&cust2, nil)
	assert.NoError(t, err)
	g.Render()
	assert.True(t, g.controller.Context().Response.Data("data") == nil)
	assert.True(t, g.controller.Context().Response.Data("head") != nil)

	// ViewEdit / ViewDetails
	body = strings.NewReader("")
	r = httptest.NewRequest("GET", "https://localhost/users?mode=details&ID=1", body)
	g = defaultGrid(r)

	b, err = HelperCreateBuilder()
	assert.NoError(t, err)
	orm.GlobalBuilder = &b

	cust3 := Customerfk{}
	err = g.SetSource(&cust3, nil)
	assert.NoError(t, err)
	g.Render()
	assert.True(t, g.controller.Context().Response.Data("data") != nil)
	assert.True(t, g.controller.Context().Response.Data("head") != nil)

	// Create
	body = strings.NewReader(`{"FirstName":"Test123","LastName":"Stoate","CreatedAt":"2019-02-23T00:00:00Z","UpdatedAt":"2020-03-02T00:00:00Z","DeletedAt":"2020-10-02T00:00:00Z","Info":{"Phone":"000-000-123"},"Orders":[{"CreatedAt":"2010-07-21T00:00:00Z"},{"CreatedAt":"2010-07-22T00:00:00Z"},{"CreatedAt":"2010-07-23T00:00:00Z"}],"Service":[{"Name":"paypal"},{"Name":"banking"},{"Name":"appstore"},{"Name":"playstore"}]}`)
	r = httptest.NewRequest("POST", "https://localhost/users", body)
	g = defaultGrid(r)

	b, err = HelperCreateBuilder()
	assert.NoError(t, err)
	orm.GlobalBuilder = &b

	cust4 := Customerfk{}
	err = g.SetSource(&cust4, nil)
	assert.NoError(t, err)
	g.Render()
	assert.Equal(t, "Test123", cust4.FirstName.String)

	// Update
	body = strings.NewReader(`{"ID":1,"FirstName":"TreschaUpd","LastName":"Stoate","CreatedAt":"2019-02-23T00:00:00Z","UpdatedAt":"2020-03-02T00:00:00Z","DeletedAt":"2020-10-02T00:00:00Z","Info":{"Phone":"000-000-123"},"Orders":[{"CreatedAt":"2010-07-21T00:00:00Z"},{"CreatedAt":"2010-07-22T00:00:00Z"},{"CreatedAt":"2010-07-23T00:00:00Z"}],"Service":[{"Name":"paypal"},{"Name":"banking"},{"Name":"appstore"},{"Name":"playstore"}]}`)
	r = httptest.NewRequest("PUT", "https://localhost/users", body)
	g = defaultGrid(r)

	b, err = HelperCreateBuilder()
	assert.NoError(t, err)
	orm.GlobalBuilder = &b

	cust5 := Customerfk{}
	err = g.SetSource(&cust5, nil)
	assert.NoError(t, err)
	g.Render()
	assert.Equal(t, "TreschaUpd", cust5.FirstName.String)

	// Delete
	body = strings.NewReader("")
	r = httptest.NewRequest("DELETE", "https://localhost/users?ID=1", body)
	g = defaultGrid(r)

	b, err = HelperCreateBuilder()
	assert.NoError(t, err)
	orm.GlobalBuilder = &b

	cust6 := Customerfk{}
	err = g.SetSource(&cust6, nil)
	assert.NoError(t, err)
	g.Render()
	assert.True(t, cust6.DeletedAt.Valid)
}

func defaultGrid(r *http.Request) *Grid {
	// new controller
	c := controller.Controller{}

	//cache
	cache, _ := cache.New("memory", memory.Options{GCInterval: 5 * time.Minute})
	c.SetCache(cache)

	rw := httptest.NewRecorder()
	ctx := context.New(r, rw)
	c.SetContext(ctx)

	//new grid
	grid := New(&c)
	return grid
}

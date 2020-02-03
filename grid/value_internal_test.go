package grid

import (
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"strings"
	"testing"
)

// defaultValue helper for the tests
func defaultValue() (*value, *Grid) {
	body := strings.NewReader("")
	r := httptest.NewRequest("GET", "https://localhost/users", body)
	g := defaultGrid(r)
	return valueWithGrid("test", g), g
}

// TestValue_Value testing if the value is set to all modes
func TestValue_Value(t *testing.T) {
	v := Value("read")

	assert.Equal(t, "read", v.grid)
	assert.Equal(t, "read", v.details)
	assert.Equal(t, "read", v.create)
	assert.Equal(t, "read", v.update)
	assert.True(t, v.g == nil)
}

// TestValue_valueWithGrid testing if the value gets set with the grid
func TestValue_valueWithGrid(t *testing.T) {
	v, g := defaultValue()
	assert.Equal(t, "test", v.get())
	assert.Equal(t, g, v.g)
}

// TestValue_set testing if the value gets set and get
func TestValue_set(t *testing.T) {
	v, _ := defaultValue()

	//set a new value to all modes
	v.set("read")
	assert.Equal(t, "read", v.get())
}
func TestValue_get(t *testing.T) {

	body := strings.NewReader("")
	r := httptest.NewRequest("GET", "https://localhost/users", body)
	g := defaultGrid(r)
	v := valueWithGrid("test", g)
	v.grid = "grid"
	v.create = "create"
	v.details = "details"
	v.update = "update"
	//ViewGrid
	assert.Equal(t, "grid", v.get())

	body = strings.NewReader("")
	r = httptest.NewRequest("GET", "https://localhost/users?mode=new", body)
	g = defaultGrid(r)
	v = valueWithGrid("test", g)
	v.grid = "grid"
	v.create = "create"
	v.details = "details"
	v.update = "update"
	//ViewCreate
	assert.Equal(t, "create", v.get())

	body = strings.NewReader("")
	r = httptest.NewRequest("GET", "https://localhost/users?mode=details&id=1", body)
	g = defaultGrid(r)
	v = valueWithGrid("test", g)
	v.grid = "grid"
	v.create = "create"
	v.details = "details"
	v.update = "update"
	//ViewDetails
	assert.Equal(t, "details", v.get())

	body = strings.NewReader("")
	r = httptest.NewRequest("GET", "https://localhost/users?mode=edit&id=1", body)
	g = defaultGrid(r)
	v = valueWithGrid("test", g)
	v.grid = "grid"
	v.create = "create"
	v.details = "details"
	v.update = "update"
	//ViewEdit
	assert.Equal(t, "update", v.get())

	body = strings.NewReader("")
	r = httptest.NewRequest("DELETE", "https://localhost/users", body)
	g = defaultGrid(r)
	v = valueWithGrid("test", g)
	v.grid = "grid"
	v.create = "create"
	v.details = "details"
	v.update = "update"
	//No Value set for other this HTTP Method
	assert.Equal(t, nil, v.get())

}
func TestValue_setByValue(t *testing.T) {
	v, _ := defaultValue()

	assert.Equal(t, "test", v.grid)
	assert.Equal(t, "test", v.details)
	assert.Equal(t, "test", v.create)
	assert.Equal(t, "test", v.update)

	val := Value("grid").Details("details").Edit("edit").Create("create")
	v.setByValue(val)

	assert.Equal(t, "grid", v.grid)
	assert.Equal(t, "details", v.details)
	assert.Equal(t, "create", v.create)
	assert.Equal(t, "edit", v.update)

	val = Value("grid").Grid("g2")
	v.setByValue(val)
	assert.Equal(t, "g2", v.grid)
}

func TestValue_getBool(t *testing.T) {
	v, _ := defaultValue()

	v.grid = true
	assert.Equal(t, true, v.getBool())

	//default
	v.grid = nil
	assert.Equal(t, false, v.getBool())
}
func TestValue_getInt(t *testing.T) {
	v, _ := defaultValue()

	v.grid = 333
	assert.Equal(t, 333, v.getInt())

	//default
	v.grid = nil
	assert.Equal(t, 0, v.getInt())
}
func TestValue_getString(t *testing.T) {
	v, _ := defaultValue()

	v.grid = "grid"
	assert.Equal(t, "grid", v.getString())

	//default
	v.grid = nil
	assert.Equal(t, "", v.getString())
}

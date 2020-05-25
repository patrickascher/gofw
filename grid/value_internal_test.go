package grid

import (
	"github.com/patrickascher/gofw/cache"
	_ "github.com/patrickascher/gofw/cache/memory"
	"github.com/patrickascher/gofw/controller"
	"github.com/patrickascher/gofw/controller/context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

// helper to change the different grid modes
func newController(r *http.Request) controller.Interface {
	c := controller.Controller{}

	ca, _ := cache.New("memory", nil)
	c.SetCache(ca)

	rw := httptest.NewRecorder()
	ctx := context.New(r, rw)
	c.SetContext(ctx)
	return &c
}

// Testing if the value is set for all items if a normal type is used and if the value is set by mode.
func TestNewValue(t *testing.T) {
	test := assert.New(t)

	v := NewValue("title")
	// test if every value is set to the same
	test.Equal("title", v.create)
	test.Equal("title", v.details)
	test.Equal("title", v.update)
	test.Equal("title", v.table)
	test.Equal("title", v.export)

	v = NewValue("title").SetExport("title_export").SetTable("title_table").SetCreate("title_create").SetDetails("title_details").SetUpdate("title_edit")
	// test if every value is set to the same
	test.Equal("title_create", v.create)
	test.Equal("title_details", v.details)
	test.Equal("title_edit", v.update)
	test.Equal("title_table", v.table)
	test.Equal("title_export", v.export)
}

// Test if the return value is correct for by mode.
func TestValue_Mode(t *testing.T) {
	test := assert.New(t)

	field := Field{}
	field.SetTitle(NewValue("title").SetDetails("title_details").SetCreate("title_create").SetUpdate("title_edit"))

	field.SetMode(VTable)
	test.Equal("title", field.Title())

	field.SetMode(VDetails)
	test.Equal("title_details", field.Title())

	field.SetMode(VCreate)
	test.Equal("title_create", field.Title())

	field.SetMode(VUpdate)
	test.Equal("title_edit", field.Title())
}

package orm_test

import (
	"fmt"
	"github.com/patrickascher/gofw/cache"
	"github.com/patrickascher/gofw/controller"
	"github.com/patrickascher/gofw/controller/context"
	grid2 "github.com/patrickascher/gofw/grid"
	"github.com/patrickascher/gofw/orm"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestGrid(t *testing.T) {
	test := assert.New(t)
	car := car{}
	src := orm.Grid(&car)
	test.Equal("*orm.gridSource", reflect.TypeOf(src).String())
}

func TestGridSource_Init(t *testing.T) {
	test := assert.New(t)
	car := car{}
	src := orm.Grid(&car)

	err := src.Init(nil)
	test.NoError(err)

	test.Equal("orm_test.car", car.Scope().Name(true))
	test.True(car.Scope().ReferencesOnly())
}

// helper to change the different grid modes
func newController(r *http.Request) (controller.Interface, *httptest.ResponseRecorder) {
	c := controller.Controller{}
	rw := httptest.NewRecorder()
	ctx := context.New(r, rw)

	ca, _ := cache.New("memory", nil)
	c.SetCache(ca)

	c.SetContext(ctx)
	return &c, rw
}

func TestGridSource_Fields_Policy(t *testing.T) {

	controller, _ := newController(httptest.NewRequest("GET", "https://localhost/", strings.NewReader("")))

	for n := 0; n < 2; n++ {

		tmpWblist := "blacklist"
		tmpRemoved := false
		if n == 1 {
			tmpWblist = "whitelist"
			tmpRemoved = true
		}

		grid := grid2.New(controller, &grid2.Config{Policy: n})

		test := assert.New(t)
		car := car{}
		src := orm.Grid(&car)

		err := src.Init(nil)
		test.NoError(err)

		fields, err := src.Fields(grid)
		test.NoError(err)

		test.Equal(6, len(fields))

		test.Equal("ID", fields[0].Id())
		test.True(fields[0].IsRemoved())

		// table driven:
		var tests = []struct {
			mode              int
			id                string
			referenceId       string
			primary           bool
			fieldType         string
			title             string
			description       string
			position          int
			remove            bool
			hidden            bool
			view              string
			readOnly          bool
			sortable          bool
			filterable        bool
			options           map[string]interface{}
			callback          interface{}
			callbackArguments []interface{}
			fields            int
			relation          bool
			error             error
		}{
			{id: "ID", referenceId: "cars.id", primary: true, fieldType: "Integer", title: "ID", description: "", position: 0, remove: true, hidden: false, view: "", readOnly: false, sortable: true, filterable: true, relation: false, options: map[string]interface{}{"validate": "omitempty,numeric,min=0,max=4294967295"}, callback: nil, callbackArguments: nil, fields: 0, error: nil},
			{id: "OwnerID", referenceId: "cars.owner_id", primary: false, fieldType: "Integer", title: "OwnerID", description: "", position: 1, remove: true, hidden: false, view: "", readOnly: false, sortable: true, filterable: true, relation: false, options: map[string]interface{}{"validate": "omitempty,numeric,min=0,max=4294967295"}, callback: nil, callbackArguments: nil, fields: 0, error: nil},
			{id: "Owner", referenceId: "", primary: false, fieldType: orm.BelongsTo, title: "Owner", description: "", position: 2, remove: tmpRemoved, hidden: false, view: "", readOnly: false, sortable: false, filterable: false, relation: true, options: map[string]interface{}{"select": grid2.Select{TextField: "Name", ValueField: "ID", Items: []grid2.SelectItem(nil), Api: "", Condition: "", OrmField: ""}}, callback: nil, callbackArguments: nil, fields: 2, error: nil},
			{id: "Driver", referenceId: "", primary: false, fieldType: orm.ManyToMany, title: "Driver", description: "", position: 3, remove: tmpRemoved, hidden: false, view: "", readOnly: false, sortable: false, filterable: false, relation: true, options: map[string]interface{}{"select": grid2.Select{TextField: "Name", ValueField: "ID", Items: []grid2.SelectItem(nil), Api: "", Condition: "", OrmField: ""}}, callback: nil, callbackArguments: nil, fields: 2, error: nil},
			{id: "Radio", referenceId: "", primary: false, fieldType: orm.HasOne, title: "Radio", description: "", position: 4, remove: tmpRemoved, hidden: false, view: "", readOnly: false, sortable: false, filterable: false, relation: true, options: map[string]interface{}(nil), callback: nil, callbackArguments: nil, fields: 5, error: nil},
			{id: "Liquid", referenceId: "", primary: false, fieldType: orm.HasMany, title: "Liquid", description: "", position: 5, remove: tmpRemoved, hidden: false, view: "", readOnly: false, sortable: false, filterable: false, relation: true, options: map[string]interface{}(nil), callback: nil, callbackArguments: nil, fields: 5, error: nil},
			// Wheel and Brand only have Write permission
		}

		for i, tt := range tests {
			t.Run(tt.id+" "+tmpWblist, func(t *testing.T) {
				fields[i].SetMode(grid2.VTable)

				test.Equal(tt.id, fields[i].Id())
				test.Equal(tt.referenceId, fields[i].DatabaseId())
				test.Equal(tt.primary, fields[i].IsPrimary())
				test.Equal(tt.fieldType, fields[i].FieldType())
				test.Equal(tt.title, fields[i].Title())
				test.Equal(tt.description, fields[i].Description())
				test.Equal(tt.position, fields[i].Position())
				test.Equal(tt.remove, fields[i].IsRemoved())
				test.Equal(tt.hidden, fields[i].IsHidden())
				test.Equal(tt.view, fields[i].View())
				test.Equal(tt.readOnly, fields[i].IsReadOnly())
				test.Equal(tt.sortable, fields[i].IsSortable())
				test.Equal(tt.filterable, fields[i].IsFilterable())
				test.Equal(tt.relation, fields[i].IsRelation())
				test.Equal(tt.options, fields[i].Options())
				cbk, args := fields[i].Callback()
				test.Equal(tt.callback, cbk)
				test.Equal(tt.callbackArguments, args)
				test.Equal(tt.fields, len(fields[i].Fields()))
				if len(fields[i].Fields()) > 0 {
					for y := 0; y < len(fields[i].Fields()); y++ {
						fields[i].Fields()[y].SetMode(grid2.VTable)
						if fields[i].Fields()[y].IsPrimary() || fields[i].Fields()[y].Id() == "CarID" || fields[i].Fields()[y].Id() == "CarType" {
							test.Equal(true, fields[i].Fields()[y].IsRemoved())
						} else {
							fmt.Println(fields[i].Id(), fields[i].Fields()[y].Id(), fields[i].Fields()[y].IsRemoved(), fields[i].Fields()[y].IsRelation())
							test.Equal(tmpRemoved, fields[i].Fields()[y].IsRemoved())
						}
					}
				}

			})
		}
	}
}

func TestGridSource_Fields_Json(t *testing.T) {
	controller, _ := newController(httptest.NewRequest("GET", "https://localhost/", strings.NewReader("")))
	grid := grid2.New(controller, nil)

	test := assert.New(t)
	car := carJsonNameAndSkip{}
	src := orm.Grid(&car)

	err := src.Init(nil)
	test.NoError(err)

	fields, err := src.Fields(grid)
	test.NoError(err)
	fields[0].SetMode(grid2.VTable)

	test.Equal(1, len(fields)) // OwnerID is skipped
	test.Equal("pid", fields[0].Id())
	test.Equal("ID", fields[0].Title())
	test.Equal("cars.id", fields[0].DatabaseId())
}

func TestGridSource_Fields_Enum_BelongsTo(t *testing.T) {
	for i := 0; i < 2; i++ {
		controller, _ := newController(httptest.NewRequest("GET", "https://localhost/?mode=update", strings.NewReader("")))
		grid := grid2.New(controller, &grid2.Config{Policy: i})

		test := assert.New(t)
		car := carEnum{}
		src := orm.Grid(&car)

		err := src.Init(nil)
		test.NoError(err)

		fields, err := src.Fields(grid)
		test.NoError(err)
		fields[0].SetMode(grid2.VUpdate)
		fields[1].SetMode(grid2.VUpdate)
		fields[2].SetMode(grid2.VUpdate)
		fields[3].SetMode(grid2.VUpdate)

		test.Equal(4, len(fields))
		test.Equal("ID", fields[0].Id())

		test.Equal("Enum", fields[1].Id())
		test.Equal(map[string]interface{}{"select": grid2.Select{TextField: "text", ValueField: "value", Items: []grid2.SelectItem{grid2.SelectItem{Text: "TEST", Value: "TEST"}, grid2.SelectItem{Text: "TEST2", Value: "TEST2"}}, Api: "", Condition: "", OrmField: ""}, "validate": "omitempty,oneof=TEST TEST2", "vueReturnObject": false}, fields[1].Options())

		test.Equal("OwnerID", fields[2].Id())
		if i == 0 { // Blacklist
			test.Equal(false, fields[2].IsRemoved())
		} else { //Whitelist
			test.Equal(true, fields[2].IsRemoved())
		}
		test.Equal("Owner", fields[3].Id())
		test.Equal(true, fields[3].IsRemoved()) // always removed because on edit only ownerID is loaded
	}
}

func TestGridSource_UpdateFields(t *testing.T) {
	test := assert.New(t)

	for i := 0; i < 2; i++ {
		controller, _ := newController(httptest.NewRequest("GET", "https://localhost/", strings.NewReader("")))
		grid := grid2.New(controller, &grid2.Config{Policy: i})

		car := car{}
		src := orm.Grid(&car)

		err := grid.SetSource(src)
		test.NoError(err)

		err = src.UpdatedFields(grid)

		policy, list := car.WBList()
		test.Equal(orm.WHITELIST, policy)
		if i == 0 {
			// BLACKLIST
			test.NoError(err)
			test.Equal([]string{"Owner.Name", "Driver.Name", "Radio.Brand", "Radio.Note", "Liquid.Brand", "Liquid.Note"}, list)
		} else {
			// WHITELIST - no fields are added
			test.Error(err)
			test.Equal([]string(nil), list)

			// allow some Fields
			grid.Field("ID").SetRemove(false)
			grid.Field("Owner").SetRemove(false)
			grid.Field("OwnerID").SetReadOnly(true).SetRemove(false)
			grid.Field("Radio.Brand").SetRemove(false)

			err = src.UpdatedFields(grid)
			policy, list := car.WBList()
			test.Equal(orm.WHITELIST, policy)

			// check read only
			f, err := car.Scope().Field("OwnerID")
			test.NoError(err)
			test.Equal(orm.Permission{Read: true, Write: false}, f.Permission)

			test.Equal([]string{"ID", "OwnerID", "Owner", "Radio.Brand"}, list)
		}
	}
}

func TestGridSource_Callback(t *testing.T) {
	test := assert.New(t)

	controller, _ := newController(httptest.NewRequest("GET", "https://localhost/?f=Owner", strings.NewReader("")))
	grid := grid2.New(controller, nil)

	car := car{}
	src := orm.Grid(&car)

	err := grid.SetSource(src)
	test.NoError(err)

	// with link field
	sel, err := src.Callback(grid2.FeSelect, grid)
	test.NoError(err)
	test.Equal(2, len(sel.([]owner)))

	// with defined OrmField.
	grid.Field("Owner").SetOption(grid2.FeSelect, grid2.Select{OrmField: "Owner"})
	sel, err = src.Callback(grid2.FeSelect, grid)
	test.NoError(err)
	test.Equal(2, len(sel.([]owner)))
}

func TestGridSource_First(t *testing.T) {
	test := assert.New(t)

	err := createEntries(&car{})
	test.NoError(err)

	c := car{}
	src := orm.Grid(&c)
	err = src.Init(nil)
	test.NoError(err)

	res, err := src.First(nil, nil)
	test.NoError(err)
	test.Equal(int64(1), res.(*car).ID.Int64)

}

func TestGridSource_All(t *testing.T) {
	test := assert.New(t)

	err := createEntries(&car{})
	test.NoError(err)

	c := car{}
	src := orm.Grid(&c)
	err = src.Init(nil)
	test.NoError(err)

	res, err := src.All(nil, nil)
	test.NoError(err)
	test.Equal(2, len(res.([]car)))
}

func TestGridSource_Count(t *testing.T) {
	test := assert.New(t)

	err := createEntries(&car{})
	test.NoError(err)

	c := car{}
	src := orm.Grid(&c)
	err = src.Init(nil)
	test.NoError(err)

	count, err := src.Count(nil, nil)
	test.NoError(err)
	test.Equal(2, count)
}

func TestGridSource_Create(t *testing.T) {
	test := assert.New(t)

	controller, _ := newController(httptest.NewRequest("POST", "https://localhost/", strings.NewReader("{\"Brand\":\"NewCreation\"}")))
	grid := grid2.New(controller, nil)

	c := car{}
	src := orm.Grid(&c)
	err := src.Init(nil)
	test.NoError(err)

	count, err := src.Count(nil, grid)
	test.NoError(err)

	res, err := src.Create(grid)
	test.NoError(err)
	test.True(res.(map[string]interface{})["ID"].(orm.NullInt).Valid)

	count2, err := src.Count(nil, grid)
	test.NoError(err)
	test.Equal(count+1, count2)
}

func TestGridSource_Update(t *testing.T) {
	test := assert.New(t)

	controller, _ := newController(httptest.NewRequest("PUT", "https://localhost/", strings.NewReader("{\"ID\":1,\"Brand\":\"Updated\"}")))
	grid := grid2.New(controller, nil)

	err := createEntries(&car{})
	test.NoError(err)

	c := car{}
	src := orm.Grid(&c)
	err = src.Init(nil)
	test.NoError(err)

	err = src.Update(grid)
	test.NoError(err)

	res, err := src.First(sqlquery.NewCondition().Where("id=1"), grid)
	test.NoError(err)
	test.Equal("Updated", res.(*car).Brand)
}

func TestGridSource_Delete(t *testing.T) {
	test := assert.New(t)

	controller, _ := newController(httptest.NewRequest("DELETE", "https://localhost/", strings.NewReader("")))
	grid := grid2.New(controller, nil)

	err := createEntries(&car{})
	test.NoError(err)

	c := car{}
	src := orm.Grid(&c)
	err = src.Init(nil)
	test.NoError(err)

	count, err := src.Count(nil, grid)
	test.NoError(err)

	err = src.Delete(sqlquery.NewCondition().Where("id=1"), grid)
	test.NoError(err)

	count2, err := src.Count(nil, grid)
	test.NoError(err)
	test.Equal(count-1, count2)
}

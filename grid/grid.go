// Package grid is a module which transforms a orm struct into a CRUD backend.
package grid

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/patrickascher/gofw/controller"
	"github.com/patrickascher/gofw/grid/config"
	"github.com/patrickascher/gofw/orm"
	"github.com/patrickascher/gofw/sqlquery"
	"io/ioutil"
	"reflect"
	"strconv"
)

// Identifiers for grid mode
const (
	ViewGrid = iota + 1
	ViewDetails
	ViewCreate
	ViewEdit
	SelectValues
	CREATE
	UPDATE
	DELETE
	CBK
)

// Configuration keys
const (
	HEADER                 = "header"
	PAGINATION             = "pagination"
	CALLBACK               = "callback"
	PaginationDefaultLimit = 15
)

// all errors are defined here
var (
	ErrJsonInvalid     = errors.New("grid: json is invalid")
	ErrRequestBody     = errors.New("grid: request body is empty")
	ErrCache           = errors.New("grid: controller must have a cache")
	ErrFieldOrRelation = errors.New("grid: field or relation %#v does not exist")
	ErrDisable         = errors.New("grid: function not found to disable %s")
	ErrAction          = errors.New("grid: this action is not allowed")
)

// Grid is holding all information.
type Grid struct {
	src          orm.Interface
	srcCondition *sqlquery_.Condition

	fields     map[string]Interface
	controller controller.Interface
	config     *config.Config

	disableHeader     bool
	disablePagination bool
	disableCallback   bool

	sourceAdded bool
}

func (g *Grid) blacklistedFields(fields map[string]Interface, parent string) []string {
	var rv []string

	if fields == nil {
		fields = g.fields
	}

	for _, field := range fields {
		if field.getRemove() {
			name := parent
			if name != "" {
				name += "."
			}
			rv = append(rv, name+field.getFieldName())
			continue
		}

		if len(field.getFields()) > 0 {
			name := parent
			if name != "" {
				name += "."
			}
			prv := g.blacklistedFields(field.getFields(), name+field.getFieldName())
			if len(prv) > 0 {
				rv = append(rv, prv...)
			}
		}
	}

	return rv
}

// Config returns the grid configuration.
func (g *Grid) Config() *config.Config {
	return g.config
}

// New creates a new grid instance.
// The controller interface is needed because of the context and cache.
func New(c controller.Interface) *Grid {
	g := &Grid{}
	g.fields = make(map[string]Interface, 0)
	g.config = &config.Config{}

	g.controller = c

	return g
}

func (g *Grid) Controller() controller.Interface {
	return g.controller
}

func (g *Grid) Source() orm.Interface {
	return g.src
}

// Source connects a model to the grid.
// The controller must have a cache otherwise an error will return.
func (g *Grid) SetSource(m orm.Interface, condition *sqlquery_.Condition) error {
	// init model
	if g.controller.HasCache() {
		err := m.SetCache(g.controller.Cache(), 0) //infinity cache
		if err != nil {
			return err
		}
	} else {
		// At the moment a cache is required, to avoid that somebody forgot to add it!
		// if this gets deleted one day, the createFields - Relation sector has to be rewritten.
		return ErrCache
	}

	// adding custom condition
	g.srcCondition = condition

	// init model
	err := m.Initialize(m)
	if err != nil {
		return err
	}

	// setting the grid source
	g.src = m

	// getting the fields from the struct and its relations
	err = g.createFields(m, nil)
	if err != nil {
		return err
	}

	// identifier if a source was set
	g.sourceAdded = true

	return nil
}

// Mode is returning the current Grid mode by param and HTTP method.
func (g *Grid) Mode() int {

	// checking callback
	m, err := g.controller.Context().Request.Param("mode")
	if err == nil && m[0] == CALLBACK {
		return CBK
	}

	switch g.httpMethod() {
	case "GET":
		m, err := g.controller.Context().Request.Param("select") // TODO this can be handled with the new callback mode.
		if err == nil {
			return SelectValues
		}
		m, err = g.controller.Context().Request.Param("mode")
		if err != nil {
			return ViewGrid
		}
		switch m[0] {
		case "new":
			return ViewCreate
		case "edit":
			return ViewEdit
		case "details":
			return ViewDetails
		}
	case "POST":
		return CREATE
	case "PUT":
		return UPDATE
	case "DELETE":
		return DELETE
	}
	return 0
}

// Disable is a helper to disable defaults of the grid
// Please use the constants PAGINATION and HEADER for it.
// An error will return if the option is unknown.
func (g *Grid) Disable(d string) error {

	switch d {
	case PAGINATION:
		g.disablePagination = true
	case HEADER:
		g.disableHeader = true
	case CALLBACK:
		g.disableCallback = true
	default:
		return fmt.Errorf(ErrDisable.Error(), d)
	}

	return nil
}

// Field is returning a grid field by its struct name.
// Error will return when no field was found.
func (g *Grid) Field(f string) (*field, error) {
	if f, ok := g.fields[f]; ok && len(f.getFields()) == 0 {
		return f.(*field), nil
	}
	return nil, fmt.Errorf(ErrFieldOrRelation.Error(), f)
}

// Relation is returning a grid relation by its struct name.
// Error will return when no relation was found.
func (g *Grid) Relation(rel string) (*relation, error) {

	if r, ok := g.fields[rel]; ok && len(r.getFields()) > 0 {
		return r.(*relation), nil
	}
	return nil, fmt.Errorf(ErrFieldOrRelation.Error(), rel)
}

func whereIDLoop(id int, table string, col1 string, col2 string) ([]int, error) {
	fmt.Println("++++start WHERE loop++++")
	ids := []int{}
	b := orm.GlobalBuilder
	rows, err := b.Select(table).Columns(col1).Where(col2+" = ?", id).All()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
		id2, err := whereIDLoop(id, table, col1, col2)
		if err != nil {
			return nil, err
		}
		if id2 != nil {
			ids = append(ids, id2...)
		}
	}

	err = rows.Close()
	if err != nil {
		return nil, err
	}

	fmt.Println("++++end WHERE loop++++")

	return ids, nil
}

// Render is rendering the grid by its given mode
func (g *Grid) Render() {

	if !g.sourceAdded {
		g.controller.Error(500, "grid: source must be added before render")
		return
	}

	mode := g.Mode()
	switch mode {
	case CBK:
		// get param field
		p, err := g.controller.Context().Request.Param("field")
		if err != nil {
			g.controller.Error(500, err.Error())
			return
		}

		// get param function name
		fn, err := g.controller.Context().Request.Param("fn")
		if err != nil {
			g.controller.Error(500, err.Error())
			return
		}

		field := reflect.Indirect(reflect.ValueOf(g.src)).FieldByName(p[0])
		if field.IsValid() {
			fieldType := reflect.TypeOf(field.Interface())
			modelType := orm.NewInstanceFromType(fieldType)
			method := modelType.Addr().MethodByName(fn[0])
			// CHECK if method exists
			if method.IsValid() {
				// checking the IN arguments
				numIn := method.Type().NumIn()
				if numIn != 1 {
					g.controller.Error(500, fmt.Sprintf("grid: callback %v must have a ptr to grid arguments", fn[0]))
					return
				}
				for i := 0; i < numIn; i++ {
					argType := method.Type().In(i)
					if argType != reflect.TypeOf(&Grid{}) {
						g.controller.Error(500, fmt.Sprintf("grid: callback %v must have a ptr to grid arguments", fn[0]))
					}
				}

				// checking the OUT arguments
				numOut := method.Type().NumOut()
				if numOut != 1 {
					g.controller.Error(500, fmt.Sprintf("grid: callback %v must return an error value", fn[0]))
				}
				for i := 0; i < numOut; i++ {
					rvType := method.Type().Out(i)
					if rvType.String() != "error" {
						g.controller.Error(500, fmt.Sprintf("grid: callback %v must return an error value", fn[0]))
					}
				}

				// finally call the method with the arguments
				errM := modelType.Addr().MethodByName(fn[0]).Call([]reflect.Value{reflect.ValueOf(g)})
				if !errM[0].IsNil() {
					err = errM[0].Interface().(error)
					g.controller.Error(500, err.Error())
					return
				}
			}
		}

		return
	case SelectValues: // get select values for belongsTo and m2m
		val, _ := g.controller.Context().Request.Param("select") //error already handled in getMode
		valueKey, err := g.controller.Context().Request.Param("value")
		if err != nil {
			g.controller.Error(500, err.Error())
			return
		}
		textKey, err := g.controller.Context().Request.Param("text")
		if err != nil {
			g.controller.Error(500, err.Error())
			return
		}

		selectField := reflect.Indirect(reflect.ValueOf(g.src)).FieldByName(val[0])
		selectFieldType := reflect.TypeOf(selectField.Interface())

		modelType := orm.NewInstanceFromType(selectFieldType)
		model := modelType.Addr().Interface().(orm.Interface)

		err = model.SetCache(g.controller.Cache(), 0)
		if err != nil {
			g.controller.Error(500, err.Error())
			return
		}
		err = model.Initialize(model)
		if err != nil {
			g.controller.Error(500, err.Error())
			return
		}

		resultSet := reflect.New(reflect.MakeSlice(reflect.SliceOf(modelType.Type()), 0, 0).Type()).Interface()
		model.SetWhitelist(valueKey[0], textKey[0])

		c := sqlquery_.Condition{}
		pkey := g.src.Table().PrimaryKeys() // TODO: what if more pkeys? not possible in M2M?
		id, err := g.controller.Context().Request.Param(pkey[0].StructField)
		if err == nil {
			for name, rel := range g.src.Table().Associations {
				if name == val[0] && rel.Type == orm.ManyToManySR {
					i, err := strconv.Atoi(id[0]) // TODO: what if a id is not an int?
					if err != nil {
						g.controller.Error(500, err.Error())
						return
					}
					ids, err := whereIDLoop(i, rel.JunctionTable.Table, rel.JunctionTable.StructColumn, rel.JunctionTable.AssociationColumn)
					ids = append(ids, i)
					if err != nil {
						g.controller.Error(500, err.Error())
						return
					}
					c.Where(pkey[0].Information.Name+" NOT IN (?)", ids)
				}
			}
		}

		err = model.All(resultSet, &c)
		if err != nil {
			g.controller.Error(500, err.Error())
			return
		}

		g.controller.Set("values", resultSet)
		return

	case ViewGrid: // readAll
		// callback before
		if !g.disableCallback {
			err := orm.CallMethodIfExist(g.src, []string{"GridBeforeViewAll", "GridBeforeView"}, g)
			if err != nil {
				g.controller.Error(500, err.Error())
				return
			}
		}

		err := g.readAll()
		if err != nil {
			g.controller.Error(500, err.Error())
			return
		}

		// callback after
		if !g.disableCallback {
			err = orm.CallMethodIfExist(g.src, []string{"GridAfterViewAll", "GridAfterView"}, g)
			if err != nil {
				g.controller.Error(500, err.Error())
				return
			}
		}
	case ViewCreate: // readOne
		// callback before
		if !g.disableCallback {
			err := orm.CallMethodIfExist(g.src, []string{"GridBeforeViewCreate", "GridBeforeView"}, g)
			if err != nil {
				g.controller.Error(500, err.Error())
				return
			}
		}
		if g.config.Action.New.Disable {
			g.controller.Error(500, ErrAction.Error())
			return
		}

		g.headerInfo()

		// callback after
		if !g.disableCallback {
			err := orm.CallMethodIfExist(g.src, []string{"GridAfterViewCreate", "GridAfterView"}, g)
			if err != nil {
				g.controller.Error(500, err.Error())
				return
			}
		}
	case ViewEdit, ViewDetails: // readOne

		callbacksBefore := []string{"GridBeforeViewDetails", "GridBeforeView"}
		callbacksAfter := []string{"GridAfterViewDetails", "GridAfterView"}
		if !g.disableCallback {
			if mode == ViewEdit {
				callbacksBefore = []string{"GridBeforeViewEdit", "GridBeforeView"}
				callbacksAfter = []string{"GridAfterViewEdit", "GridAfterView"}
			}

			// callback before
			err := orm.CallMethodIfExist(g.src, callbacksBefore, g)
			if err != nil {
				g.controller.Error(500, err.Error())
				return
			}
		}

		if g.config.Action.Edit.Disable {
			g.controller.Error(500, ErrAction.Error())
			return
		}

		// create condition
		c, err := checkPrimaryParams(g)
		if err != nil {
			g.controller.Error(500, err.Error())
			return
		}

		// request data
		g.src.DisableCustomSql(true) // only has to get disabled here, in the other orm modes its not getting used.
		err = g.readOne(c)
		if err != nil {
			g.controller.Error(500, err.Error())
			return
		}

		// callback after
		if !g.disableCallback {
			err = orm.CallMethodIfExist(g.src, callbacksAfter, g)
			if err != nil {
				g.controller.Error(500, err.Error())
				return
			}
		}
	case CREATE: // create entry
		err := g.create()
		if err != nil {
			g.controller.Error(500, err.Error())
		}
		return
	case UPDATE: // update entry
		//		c, err := checkPrimaryParams(g)
		err := g.update()
		if err != nil {
			g.controller.Error(500, err.Error())
		}
		return
	case DELETE: // delete entry

		if g.config.Action.Delete.Disable {
			g.controller.Error(500, ErrAction.Error())
			return
		}

		c, err := checkPrimaryParams(g)
		if err != nil {
			g.controller.Error(500, err.Error())
			return
		}
		err = g.delete(c)
		if err != nil {
			g.controller.Error(500, err.Error())
		}
		return
	default:
		g.controller.Error(500, fmt.Sprintf("grid: mode %#v is unknown", mode))
		return
	}
}

// getRelationName returns the name of the relation.
// Its getting checked against a slice, slice ptr, ptr and slice.
func getRelationName(v reflect.Value) string {
	if v.Type().Kind() == reflect.Slice {
		//handel ptr
		if v.Type().Elem().Kind() == reflect.Ptr {
			return v.Type().Elem().Elem().String()
		}
		return v.Type().Elem().String()
	}
	if v.Kind() == reflect.Ptr {
		return v.Type().Elem().String()
	}
	return v.Type().String()
}

// createFields is adding the model fields and relations to grid
// needed for the later configuration of the grid
func (g *Grid) createFields(m orm.Interface, parent *relation) error {

	//Columns
	orm.SetDefaultPermission(m, true)

	for _, col := range m.Table().Columns(orm.READVIEW) {

		field := defaultField(g)
		field.setTitle(col.StructField)
		// Hide autoincrement field in create and update. Also hide fk fields.
		if col.Information.Autoincrement == true || (parent != nil && parent.association.AssociationTable.StructField == col.StructField) {
			field.hide.set(true)
			//field.hide.Create(true).Edit(true)
		}
		field.setPosition(col.Information.Position)
		field.getFieldType().SetName(col.Information.Type.Kind())

		if reflect.ValueOf(m).MethodByName("GridCallback" + col.StructField).IsValid() {
			field.SetCallback("GridCallback" + col.StructField)
		}

		// set default fieldName (json name tag)
		f, ok := reflect.TypeOf(m).Elem().FieldByName(col.StructField)
		if ok {
			jsonTagName := jsonTagName(f.Tag.Get("json"))
			if jsonTagName != "" {
				field.setJsonName(jsonTagName)
			}
		}

		// disable sort und filter on custom types
		if col.Information.Type.Kind() == orm.CustomImpl {
			field.SetFilter(false).SetSort(false)
		}

		// setReadOnly
		if col.Permission.Read == true && col.Permission.Write == false {
			field.SetReadOnly(true)
		}

		field.column = col // adding column, maybe we need some reference later on. TODO check if needed

		// add field to grid fields
		if parent != nil {
			parent.fields[col.StructField] = field
		} else {
			g.fields[col.StructField] = field
		}
	}

	// Relations
	// relations are getting a high position number to not interfere with the normal fields.
	// The position will get normalized later again.
	positionRelation := relationCounter
	for fieldName, rel := range m.Table().Associations {

		if parent != nil && parent.skipRelation == fieldName {
			continue // self reference infinity loop TODO better solution...
		}

		r := defaultRelation(g)
		r.setTitle(fieldName)
		r.setPosition(positionRelation)

		nv := orm.NewInstanceFromType(reflect.ValueOf(m).Elem().FieldByName(fieldName).Type())
		customFieldType := nv.Addr().MethodByName("GridFieldType")
		if customFieldType.IsValid() {
			in := make([]reflect.Value, 1)
			in[0] = reflect.ValueOf(g)
			r.setFieldType(customFieldType.Call(in)[0].Interface().(FieldType))
		} else {
			r.getFieldType().SetName(rel.Type)
		}

		r.setFieldName(fieldName)

		r.association = rel // needed to automatically hide some relation fields
		positionRelation++

		// checking if relation field exists
		rv := reflect.Indirect(reflect.ValueOf(m)).FieldByName("caller").Elem().Elem().FieldByName(fieldName)
		if !rv.IsValid() {
			continue
		}

		// check if field Callback exists
		// schema: GridCallback{FieldName}
		if reflect.ValueOf(m).MethodByName("GridCallback" + fieldName).IsValid() {
			r.SetCallback("GridCallback" + fieldName)
		}

		reflectRel := orm.NewInstanceFromType(reflect.ValueOf(m).Elem().FieldByName(fieldName).Type())
		if reflectRel.Addr().MethodByName("GridCallback").IsValid() {
			fmt.Println("---->", reflectRel.Addr().MethodByName("GridCallback").Interface())
			r.SetCallback(reflectRel.Addr().MethodByName("GridCallback").Interface())
		}

		// getting model from cache, this logic has to change if a cache is not required anymore
		relationModel, err := g.controller.Cache().Get(getRelationName(rv))
		if err != nil {
			return err
		}

		// add field to grid fields
		if parent == nil {
			g.fields[fieldName] = r
		} else {
			parent.fields[fieldName] = r
		}

		// check relations
		r.skipRelation = fieldName

		err = g.createFields(relationModel.Value().(orm.Interface).Caller(), r)
		if err != nil {
			return err
		}

		// add select for belongsTo relations, and remove the relationField ex UserID in main struct.
		if !customFieldType.IsValid() && (r.getFieldType().Name() == orm.BelongsTo || r.getFieldType().Name() == orm.ManyToMany || r.getFieldType().Name() == orm.ManyToManySR) {
			sel := &Select{}
			for _, col := range r.getFields() {
				if col.getPosition() == 1 {
					sel.SetValueKey(col.getFieldName())
				}
				if col.getPosition() == 2 {
					sel.SetTextKey(col.getFieldName())
				}
			}

			r.getFieldType().SetOption("select", sel)

			if val, ok := g.fields[rel.StructTable.StructField]; r.getFieldType().Name() == orm.BelongsTo && ok {
				val.setRemove(true)
			}
		}
	}
	return nil
}

// headerInfo is returning the field information for the frontend
func (g *Grid) headerInfo() {
	if !g.disableHeader {
		g.controller.Set("head", headerFieldsLoop(g.fields, true))
	}
}

// httpMethod is a helper to get the actual request method
func (g *Grid) httpMethod() string {
	return g.controller.Context().Request.Method()
}

// marshalModel is needed for create and update.
// It checks if the request json is correct.
// Only struct fields are allowed.
// Empty struct is not allowed.
// Errors will return if one of the rules are not satisfied.
func (g *Grid) unmarshalModel() error {
	// reading the body request
	body := g.controller.Context().Request.Raw().Body
	if body == nil {
		return ErrRequestBody
	}
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	// check if the json is valid
	if !json.Valid(b) {
		return ErrJsonInvalid
	}

	// unmarshal the request to the model struct
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	for dec.More() {
		err := dec.Decode(g.src)
		if err != nil {
			return err
		}
	}

	// check if the model has any data
	//if g.src.IsEmpty() {
	//	return ErrModelIsEmpty
	//}

	return nil
}

// readOne fetches only one db entry by its primary key
func (g *Grid) readOne(c *sqlquery_.Condition) error {

	// request data
	err := g.src.First(c)
	if err != nil {
		return err
	}

	// set data to controller response
	g.headerInfo()
	g.controller.Set("data", g.src)

	return nil
}

// readAll fetches all entries by the given filter and sorting options.
func (g *Grid) readAll() error {

	// creating result slice
	model := reflect.New(reflect.Indirect(reflect.ValueOf(g.src)).Type()).Elem() //new type
	modelSlice := reflect.MakeSlice(reflect.SliceOf(model.Type()), 0, 0)
	modelSliceNew := reflect.New(modelSlice.Type()) //
	resultSlice := modelSliceNew.Interface()

	// check if header information is disabled
	if h, err := g.controller.Context().Request.Param("head"); err != nil || h[0] == "0" {
		err = g.Disable(HEADER)
		if err != nil {
			return err
		}
	}

	// add condition with the filter and sorting condition by params
	c, err := conditionAll(g)
	if err != nil {
		return err
	}

	// generate pagination - its adding the limit and offset to the condition
	if !g.disablePagination {
		pagination := paginationOffset{}
		err = pagination.generate(g, c)
		if err != nil {
			return err
		}
	}

	// request all data
	fmt.Println("WILL GET BLACKLISTED:::", g.blacklistedFields(nil, ""))
	g.src.SetBlacklist(g.blacklistedFields(nil, "")...)
	err = g.src.All(resultSlice, c)
	if err != nil {
		return err
	}

	// adding config
	g.controller.Set("config", g.config)

	// set data to controller response
	g.headerInfo()

	// callbacks
	resCallback, err := g.callback(resultSlice)
	if err != nil {
		return err
	}
	if resCallback != nil {
		// setting the result data
		g.controller.Set("data", resCallback)
		return nil
	}

	// setting the result data
	g.controller.Set("data", resultSlice)

	return nil
}

// create is marshaling the json request and try to create an entry.
// If there is an error with marshal or with create, a error will return.
func (g *Grid) create() error {
	err := g.unmarshalModel()
	if err != nil {
		return err
	}

	err = g.src.Create()
	if err != nil {
		return err

	}

	//response with the pkey values
	pkeys := make(map[string]interface{}, 0)
	for _, col := range g.src.Table().Cols {
		if col.Information.PrimaryKey {
			pkeys[col.StructField] = reflect.Indirect(reflect.ValueOf(g.src)).FieldByName(col.StructField).Interface()
		}
	}
	g.controller.Set("pkeys", pkeys)

	// response with the pkey(s) value
	return nil
}

// update is marshaling the json request and try to update an entry.
// If there is an error with marshal or with update, a error will return.
func (g *Grid) update() error {
	err := g.unmarshalModel()
	if err != nil {
		return err
	}

	err = g.src.Update()
	if err != nil {
		return err

	}
	return nil
}

// delete an db entry by its primary
func (g *Grid) delete(c *sqlquery_.Condition) error {
	err := g.src.First(c)
	if err != nil {
		return err
	}
	err = g.src.Delete()
	if err != nil {
		return err
	}
	return nil
}

package grid

import (
	"errors"
	"fmt"
	"github.com/patrickascher/gofw/cache"
	"github.com/patrickascher/gofw/controller"
	"github.com/patrickascher/gofw/server"
	"github.com/patrickascher/gofw/sqlquery"
	"net/http"
	"strings"
	"time"
)

// Grid modes
const (
	CREATE = iota + 1
	UPDATE
	DELETE
	CALLBACK
	HEAD
	VTable
	VDetails
	VUpdate
	VCreate
	Export
)

// Callbacks
const (
	BeforeFirst = iota + 1
	AfterFirst
	BeforeAll
	AfterAll
	BeforeCreate
	AfterCreate
	BeforeUpdate
	AfterUpdate
	BeforeDelete
	AfterDelete
)

// Frontend constants
const (
	FeSelect       = "select"
	FeDecorator    = "decorator"
	FeNoEscaping   = "noEscaping"
	FeReturnObject = "vueReturnObject"
)

const cachePrefix = "grid_"

// Error messages
var (
	errCache    = errors.New("grid: cache is required")
	errSource   = "grid: no source is added in %s action %s"
	errWrapper  = "grid: %w"
	errSecurity = "grid: the mode %s is not allowed"
)

type SourceI interface {
	// Init is called right after the source was added.
	// This function can be used if the source has to get initialized after it was added.
	Init(grid *Grid) error
	// Fields of the grid.
	Fields(grid *Grid) ([]Field, error)
	// UpdatedFields is called before render. The grid fields have the user updated configurations.
	UpdatedFields(grid *Grid) error
	// Callback is called on a callback request of the grid.
	Callback(callback string, grid *Grid) (interface{}, error)
	// First request a single row by the given condition.
	First(c *sqlquery.Condition, grid *Grid) (interface{}, error)
	// All data by the given condition.
	All(c *sqlquery.Condition, grid *Grid) (interface{}, error)
	// Create the object
	Create(grid *Grid) (interface{}, error)
	// Update the object
	Update(grid *Grid) error
	// Delete the object by the given condition.
	Delete(c *sqlquery.Condition, grid *Grid) error
	// Count all the existing object by the given condition.
	Count(c *sqlquery.Condition, grid *Grid) (int, error)
}

type Select struct {
	TextField  string       `json:",omitempty"`
	ValueField string       `json:",omitempty"`
	Items      []SelectItem `json:",omitempty"`

	Api       string `json:",omitempty"`
	Condition string `json:",omitempty"`
	OrmField  string `json:"-"`
}

type SelectItem struct {
	Text  interface{} `json:"text"`
	Value interface{} `json:"value"`
}

type Grid struct {
	// the given source.
	src SourceI
	// for additional conditions on the source object.
	srcCondition *sqlquery.Condition
	// identifier if the source was added.
	sourceAdded bool
	// the given controller.
	controller controller.Interface
	// grid fields
	fields []Field
	// the given controller.
	config Config
	//callbacks
	callbacks map[int]func(*Grid) error
}

// New creates a grid instance with the given controller.
// the controller is used to fetch all the request data and add the response.
func New(c controller.Interface, config *Config) *Grid {
	grid := &Grid{controller: c}
	// TODO config correctly, at the moment only for testing.
	// TODO also check config in the render mode if allowed.
	if config != nil {
		grid.config = *config
	} else {
		grid.config.Policy = 1 // whitelist
	}

	return grid
}

func (g *Grid) IsCallback() bool {
	if g.Mode() == CALLBACK {
		g.Render()
		return true
	}
	return false
}

// AddCallback to the grid.
// (Before/After)First,All,Create,Update,Delete exists.
func (g *Grid) AddCallback(name int, fn func(*Grid) error) {
	g.callbacks[name] = fn
}

// callback internal calls the callback function if exists.
func (g *Grid) callback(name int) error {
	if fn, ok := g.callbacks[name]; ok {
		return fn(g)
	}
	return nil
}

// Config of the grid.
func (g *Grid) Config() *Config {
	return &g.config
}

// Controller returns the grid controller.
// This data could be useful in the implemented source.
func (g *Grid) Controller() controller.Interface {
	return g.controller
}

// SetCondition adds a condition on the primary source.
func (g *Grid) SetCondition(c *sqlquery.Condition) *Grid {
	g.srcCondition = c
	return g
}

func (g *Grid) gridID() string {

	if g.config.ID == "" {
		g.config.ID = g.controller.Name() + ":" + g.controller.Action()
	}

	return g.config.ID
}

// SetSource to the grid.
// Fields are getting fetched from the source.
func (g *Grid) SetSource(src SourceI) error {

	serverCache, err := server.Cache(server.DEFAULT)
	if err != nil {
		return err
	}

	// call the source init function
	err = src.Init(g)
	if err != nil {
		return err
	}

	// add source
	g.src = src
	g.sourceAdded = true

	// get the source fields
	var fields []Field
	if v, err := serverCache.Get(cachePrefix + g.gridID()); err == nil {
		t := time.Now()
		fields = v.Value().([]Field)

		fmt.Println("CACHED FIELDS::", time.Since(t))
	} else {
		t := time.Now()
		fields, err = g.src.Fields(g)
		if err != nil {
			return err
		}
		err = serverCache.Set(cachePrefix+g.gridID(), fields, cache.INFINITY)
		if err != nil {
			return err
		}
		fmt.Println("SET FIELDS::", time.Since(t))
	}

	// make a deep copy to avoid that the cached slice will be changed
	g.fields = copySlice(fields)

	// set grid mode to the fields
	setFieldModeRecursively(g, g.fields)

	return nil
}

func copySlice(fields []Field) []Field {
	rv := make([]Field, len(fields))
	copy(rv, fields)
	for k := range rv {
		if len(rv[k].fields) > 0 {
			rv[k].fields = copySlice(rv[k].fields)
		}
	}
	return rv
}

// Fields return all defined grid fields.
func (g *Grid) Fields() []Field {
	return g.fields
}

// Field by name.
// If the field was not found a new Field with an error is created.
// This helps the user to avoid annoying error if statements. If there was an error,
// the grid will automatically response with an error message. Or you can call field.Error() != nil to check if an error happend.
func (g *Grid) Field(name string) *Field {

	loop := strings.Split(name, ".")

	fields := g.fields
	for i := 0; i < len(loop); i++ {
		for k, f := range fields {
			if f.id == loop[i] && i < len(loop)-1 {
				fields = fields[k].fields
			}
			if f.id == loop[i] && i == len(loop)-1 {
				return &fields[k]
			}
		}
	}

	return &Field{error: fmt.Errorf("Field %s does not exist", name)}
}

// Mode by the given url / http method.
// POST = grid create
// PUT = grid update
// DELETE = grid delete
// GET without any mode param = grid view table
// GET with mode param "create" = grid view create
// GET with mode param "update" = grid view update
// GET with mode param "details" = grid view details
// GET with mode param "callback" = grid view callback
// everything else will return 0
func (g *Grid) Mode() int {
	// Requested HTTP method of the controller, always uppercase.
	switch g.controller.Context().Request.Method() {
	case http.MethodGet:
		// if the param mode does not exist, its the grid view.
		m, notExisting := g.controller.Context().Request.Param("mode")
		if notExisting != nil {
			return VTable
		}
		switch m[0] {
		case "head":
			return HEAD
		case "callback":
			return CALLBACK
		case "create":
			return VCreate
		case "update":
			return VUpdate
		case "details":
			return VDetails
		case "export":
			return Export
		}
	case http.MethodPost:
		return CREATE
	case http.MethodPut:
		return UPDATE
	case http.MethodDelete:
		return DELETE
	}
	return 0
}

// security security checks if the request mode is allowed by the configuration.
func (g *Grid) security() error {

	switch g.Mode() {
	case CREATE:
		if g.config.Action.DisableCreate && g.config.Action.DisableFilter {
			return fmt.Errorf(errSecurity, "create")
		}
	case VCreate:
		if g.config.Action.DisableCreate {
			return fmt.Errorf(errSecurity, "create")
		}
	case UPDATE, VUpdate:
		if g.config.Action.DisableUpdate {
			return fmt.Errorf(errSecurity, "update")
		}
	case DELETE:
		if g.config.Action.DisableDelete {
			return fmt.Errorf(errSecurity, "delete")
		}
	case VDetails:
		if g.config.Action.DisableDetail {
			return fmt.Errorf(errSecurity, "details")
		}
	}

	return nil
}

// Render the grid by the defined grid mode.
func (g *Grid) Render() {
	// source is mandatory
	if !g.sourceAdded {
		g.controller.Error(500, fmt.Errorf(errSource, g.controller.Name(), g.controller.Context().Request.FullURL()))
		return
	}

	// security check
	if err := g.security(); err != nil {
		g.controller.Error(500, err)
		return
	}

	// update the user config in the source
	err := g.src.UpdatedFields(g)
	if err != nil {
		g.controller.Error(500, fmt.Errorf(errWrapper, err))
		return
	}

	if g.config.Title != "" {
		g.controller.Set("title", g.config.Title)
	}

	// add filter to grid config
	if f, ok := getFilterList(g); ok == nil {
		g.config.Filter.List = f
	}

	mode := g.Mode()
	switch mode {
	case HEAD:
		g.controller.Set("head", g.sortFields())
		return
	case CREATE:
		// SAVE filter
		if m, err := g.controller.Context().Request.Param("mode"); err == nil && m[0] == "filter" {
			err = g.SetSource(Orm(&UserGrid{}))
			if err != nil {
				g.controller.Error(500, fmt.Errorf(errWrapper, err))
				return
			}
			pk, err := g.src.Create(g)
			if err != nil {
				g.controller.Error(500, fmt.Errorf(errWrapper, err))
				return
			}
			g.controller.Set("pkeys", pk)
			return
		}

		err = g.callback(BeforeCreate)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err))
			return
		}

		pk, err := g.src.Create(g)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err))
			return
		}
		g.controller.Set("pkeys", pk)

		err = g.callback(AfterCreate)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err))
			return
		}
		return
	case UPDATE:
		err = g.callback(BeforeUpdate)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err))
			return
		}

		err := g.src.Update(g)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err))
			return
		}

		err = g.callback(AfterUpdate)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err))
			return
		}
		return
	case DELETE:
		err = g.callback(BeforeDelete)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err))
			return
		}

		c, err := g.conditionFirst()
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err))
			return
		}
		err = g.src.Delete(c, g)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err))
			return
		}

		err = g.callback(AfterDelete)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err))
			return
		}
		return
	case Export:
		c, err := g.conditionAll()
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err))
			return
		}

		t, err := g.Controller().Context().Request.Param("type")
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err))
			return
		}
		g.controller.Set("head", FieldsToString(g.sortFields()))

		values, err := g.src.All(c, g)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err))
			return
		}

		g.controller.SetRenderType(t[0])
		g.controller.Set("data", values)
		values = nil
	case VTable:
		err = g.callback(BeforeAll)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err))
			return
		}

		c, err := g.conditionAll()
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err))
			return
		}

		// add header as long as the param noheader is not given.
		pagination, err := g.newPagination(c)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err))
			return
		}

		if _, err := g.controller.Context().Request.Param("noheader"); err != nil {
			g.controller.Set("head", g.sortFields())
			g.controller.Set("pagination", pagination)
		}

		values, err := g.src.All(c, g)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err))
			return
		}

		err = g.callback(AfterAll)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err))
			return
		}

		// adding data
		g.controller.Set("config", g.config)
		g.controller.Set("data", values)
		return
	case VUpdate, VDetails:
		err = g.callback(BeforeFirst)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err))
			return
		}

		g.controller.Set("head", g.sortFields())

		c, err := g.conditionFirst()
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err))
			return
		}
		values, err := g.src.First(c, g)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err))
			return
		}

		err = g.callback(AfterFirst)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err))
			return
		}
		g.controller.Set("data", values)
		return
	case VCreate:
		g.controller.Set("head", g.sortFields())
		return
	case CALLBACK:
		callback, err := g.controller.Context().Request.Param("callback")
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err))
			return
		}
		values, err := g.src.Callback(callback[0], g)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err))
			return
		}
		g.controller.Set("data", values)
	}
}

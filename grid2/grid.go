package grid2

import (
	"fmt"
	"github.com/patrickascher/gofw/controller"
	"github.com/patrickascher/gofw/sqlquery"
	"net/http"
	"strings"
)

const (
	CREATE = iota + 1
	UPDATE
	DELETE
	CALLBACK
	VTable
	VDetails
	VUpdate
	VCreate
)

// Frontend constants
const (
	FeSelect       = "select"
	FeDecorator    = "decorator"
	FeNoEscaping   = "noEscaping"
	FeReturnObject = "vueReturnObject"
)

var (
	errSource  = "grid: no source is added in %s action %s"
	errWrapper = "grid: %w"
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
	// One request a single row by the given condition.
	One(c *sqlquery.Condition, grid *Grid) (interface{}, error)
	// All data by the given condition.
	All(c *sqlquery.Condition, grid *Grid) (interface{}, error)
	// Create the object
	Create(grid *Grid) (interface{}, error)
	// Update the object
	Update(grid *Grid) error
	// Delete the object by the given condition.
	Delete(c *sqlquery.Condition, grid *Grid) error
	// Count all the existing object by the given condition.
	Count(c *sqlquery.Condition) (int, error)
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
	Key   interface{} `json:"value"`
	Value interface{} `json:"text"`
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
}

// New creates a grid instance with the given controller.
// the controller is used to fetch all the request data and add the response.
func New(c controller.Interface) *Grid {
	grid := &Grid{controller: c}
	// TODO config correctly, at the moment only for testing.
	// TODO also check config in the render mode if allowed.
	grid.config = Config{Action: Action{DisableCreate: false, DisableUpdate: false, DisableDetails: false}}
	return grid
}

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

// SetSource to the grid.
// Fields are getting fetched from the source.
func (g *Grid) SetSource(src SourceI) error {

	// call the source init function
	err := src.Init(g)
	if err != nil {
		return err
	}

	// add source
	g.src = src
	g.sourceAdded = true

	// get the source fields
	g.fields, err = g.src.Fields(g)
	if err != nil {
		return err
	}

	return nil
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
		case "callback":
			return CALLBACK
		case "create":
			return VCreate
		case "update":
			return VUpdate
		case "details":
			return VDetails
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

// Render the grid by the defined grid mode.
func (g *Grid) Render() {
	// source is mandatory
	if !g.sourceAdded {
		g.controller.Error(500, fmt.Sprintf(errSource, g.controller.Name(), g.controller.Context().Request.FullURL()))
		return
	}

	// update the user config in the source
	err := g.src.UpdatedFields(g)
	if err != nil {
		g.controller.Error(500, fmt.Errorf(errWrapper, err).Error())
		return
	}

	mode := g.Mode()
	switch mode {
	case CREATE:
		pk, err := g.src.Create(g)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err).Error())
			return
		}
		g.controller.Set("pkeys", pk)

		return
	case UPDATE:
		err := g.src.Update(g)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err).Error())
			return
		}
		return
	case DELETE:
		c, err := g.conditionOne()
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err).Error())
			return
		}
		err = g.src.Delete(c, g)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err).Error())
			return
		}
		return
	case VTable:
		// add header as long as the param noheader is not given.
		if _, err := g.controller.Context().Request.Param("noheader"); err != nil {
			g.controller.Set("head", g.sortFields())
		}

		c, err := g.conditionAll()
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err).Error())
			return
		}

		pagination, err := g.newPagination(c)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err).Error())
			return
		}
		g.controller.Set("pagination", pagination)

		// adding config
		g.controller.Set("config", g.config)

		values, err := g.src.All(c, g)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err).Error())
			return
		}

		//TODO callbacks on data

		g.controller.Set("data", values)
		return
	case VUpdate, VDetails:
		g.controller.Set("head", g.sortFields())

		c, err := g.conditionOne()
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err).Error())
			return
		}
		values, err := g.src.One(c, g)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err).Error())
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
			g.controller.Error(500, fmt.Errorf(errWrapper, err).Error())
			return
		}
		values, err := g.src.Callback(callback[0], g)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrapper, err).Error())
			return
		}
		g.controller.Set("data", values)
	}

}

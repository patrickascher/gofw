package orm

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/patrickascher/gofw/grid"
	"github.com/patrickascher/gofw/slices"
	"github.com/patrickascher/gofw/sqlquery"
	"io/ioutil"
	"reflect"
	"strings"
)

var ErrRequestBody = errors.New("grid: request body is empty")
var ErrJsonInvalid = errors.New("grid: json is invalid")

type gridSource struct {
	orm Interface
}

func Grid(orm Interface) *gridSource {
	g := &gridSource{}
	g.orm = orm
	return g
}

func (g *gridSource) Init(grid *grid.Grid) error {
	return g.orm.Init(g.orm)
}

func gridFields(scope *Scope, g *grid.Grid, parent string) ([]grid.Field, error) {
	// normal fields
	var rv []grid.Field
	i := 0

	for _, f := range scope.Fields(Permission{Read: true}) {
		field := grid.Field{}
		field.SetId(f.Name)
		if jsonTagName(scope, &field) {
			continue
		}
		field.SetPrimary(f.Information.PrimaryKey)
		field.SetReferenceId(f.Information.Table + "." + f.Information.Name)
		field.SetFieldType(f.Information.Type.Kind())
		field.SetTitle(g.NewValue(f.Name))
		field.SetDescription(g.NewValue(""))
		field.SetPosition(g.NewValue(i))
		field.SetRemove(g.NewValue(false))
		field.SetHidden(g.NewValue(false))
		field.SetView(g.NewValue(""))
		if f.Validator.Config != "" {
			field.SetOption(tagValidate, f.Validator.Config)
		}
		// remove,hide, view is empty
		field.SetSortable(true)
		field.SetFilterable(true)

		// field manipulations
		// Primary keys are not shown on frontend header.
		// In the backend they are loaded anyway because they are mandatory for relations.
		if f.Information.PrimaryKey {
			field.SetRemove(true)
		}

		for _, relation := range scope.Relations(Permission{Read: true}) {
			if relation.Kind == BelongsTo {
				if field.Id() == relation.ForeignKey.Name {
					field.SetRemove(true)
				}
			}
		}

		// options, callback, fields empty
		rv = append(rv, field)
		i++
	}

	for _, r := range scope.Relations(Permission{Read: true}) {

		if _, err := scope.Parent(r.Type.String()); err == nil {
			continue
		}

		if r.Kind == BelongsTo && g.Mode() != grid.VTable {
			for k := range rv {
				if rv[k].Id() == r.ForeignKey.Name {
					if parent != "" {
						parent += "."
					}
					rv[k].SetFieldType(r.Kind)
					rv[k].SetRemove(false)
					rv[k].SetOption(grid.FeSelect, grid.Select{OrmField: parent + r.Field, TextField: "Name", ValueField: r.AssociationForeignKey.Name})
					rv[k].SetOption("vueReturnObject", false)
				}
			}

			// relation has to get blacklisted. otherwise the single refid will not be updated.
			field := grid.Field{}
			field.SetId(r.Field)
			field.SetRemove(g.NewValue(true))
			field.SetRelation(true)
			rv = append(rv, field)

			continue
		}

		field := grid.Field{}
		field.SetId(r.Field)
		field.SetRelation(true)
		if jsonTagName(scope, &field) {
			continue
		}
		field.SetFieldType(r.Kind)
		field.SetTitle(g.NewValue(r.Field))
		field.SetDescription(g.NewValue(""))
		field.SetPosition(g.NewValue(i))
		field.SetRemove(g.NewValue(false))
		field.SetHidden(g.NewValue(false))
		field.SetView(g.NewValue(""))
		if r.Validator.Config != "" {
			field.SetOption(tagValidate, r.Validator.Config)
		}
		field.SetSortable(false)
		field.SetFilterable(false)
		// options, callback (depending on kind)

		// add options for BelongsTo and ManyToMany relations.
		if r.Kind == BelongsTo || r.Kind == ManyToMany {
			field.SetOption(grid.FeSelect, grid.Select{TextField: "Name", ValueField: r.AssociationForeignKey.Name})
		}

		// recursively add fields
		rScope, err := scope.NewScopeFromType(r.Type)
		if err != nil {
			return nil, err
		}
		rScope.model.parentModel = scope.model // adding parent to avoid loops

		rField, err := gridFields(rScope, g, r.Field)
		if err != nil {
			return nil, err
		}

		// field manipulations
		// FK,AFK,Poly are removed.
		// TODO better logic?
		for k, relField := range rField {
			if relField.Id() == r.ForeignKey.Name {
				rField[k].SetRemove(true)
			}
			if scope.IsPolymorphic(r) {
				if relField.Id() == r.Polymorphic.Field.Name || relField.Id() == r.Polymorphic.Type.Name {
					rField[k].SetRemove(true)
				}
			} else {
				if relField.Id() == r.AssociationForeignKey.Name {
					rField[k].SetRemove(true)
				}
			}
		}

		if len(rField) > 0 {
			field.SetFields(rField)
		}

		rv = append(rv, field)
		i++

	}

	return rv, nil
}

func (g *gridSource) Fields(grid *grid.Grid) ([]grid.Field, error) {
	return gridFields(g.orm.Scope(), grid, "")
}

func blacklistedFields(g *gridSource, fields []grid.Field, parent string) ([]string, error) {
	var blacklist []string

	if parent != "" {
		parent += "."
	}

	for _, f := range fields {

		if !f.IsRelation() {

			if f.IsRemoved() {
				if _, exists := slices.Exists(blacklist, parent+f.Id()); !exists {
					blacklist = append(blacklist, parent+f.Id())
				}
			} else {
				if parent == "" { // because relation fields can not be fetched here
					ormField, err := g.orm.Scope().Field(f.Id())
					if err != nil {
						return nil, err
					}
					ormField.Permission.Read = true
				}
			}

			if f.IsReadOnly() {
				if parent == "" { // because relation fields can not be fetched here
					ormField, err := g.orm.Scope().Field(f.Id())
					if err != nil {
						return nil, err
					}
					ormField.Permission.Write = false
				}
			}
		} else {
			if f.IsRemoved() {
				if _, exists := slices.Exists(blacklist, parent+f.Id()); !exists {
					blacklist = append(blacklist, parent+f.Id())
				}
			} else {
				bList, err := blacklistedFields(g, f.Fields(), f.Id())
				if err != nil {
					return nil, err
				}
				if len(bList) > 0 {
					for _, tmp := range bList {
						if _, exists := slices.Exists(blacklist, tmp); !exists {
							blacklist = append(blacklist, tmp)
						}
					}
				}
			}
		}
	}
	return blacklist, nil
}

func (g *gridSource) UpdatedFields(grid *grid.Grid) error {

	blacklist, err := blacklistedFields(g, grid.Fields(), "")
	if err != nil {
		return err
	}

	if len(blacklist) > 0 {
		g.orm.SetWBList(BLACKLIST, blacklist...)
	}

	return nil
}

func (g *gridSource) Callback(callback string, gr *grid.Grid) (interface{}, error) {

	if callback == grid.FeSelect {
		selectField, err := gr.Controller().Context().Request.Param("f")
		if err != nil {
			return nil, err
		}
		// get the defined grid object

		selField := gr.Field(selectField[0])
		if selField.Error() != nil {
			return nil, selField.Error()
		}
		sel := selField.Option(grid.FeSelect).(grid.Select)

		var relScope *Scope
		fields := strings.Split(selectField[0], ".")
		if sel.OrmField != "" {
			fields = strings.Split(sel.OrmField, ".")
		}

		// get relation field of the orm
		relation, err := g.orm.Scope().Relation(fields[0], Permission{Read: true})
		if err != nil {
			return nil, err
		}

		if len(fields) > 1 {
			// create a new orm object
			relScope, err := g.orm.Scope().NewScopeFromType(relation.Type)
			if err != nil {
				return nil, err
			}
			// get relation field of the orm
			relation, err = relScope.Relation(fields[1], Permission{Read: true})
			if err != nil {
				return nil, err
			}
		}

		// create a new orm object
		relScope, err = g.orm.Scope().NewScopeFromType(relation.Type)
		if err != nil {
			return nil, err
		}

		// create the result slice
		rRes := reflect.New(reflect.MakeSlice(reflect.SliceOf(relation.Type), 0, 0).Type()).Interface()
		// set whitelist fields
		reqFields := []string{sel.ValueField}
		textFields := strings.Split(sel.TextField, ",")
		for k, tf := range textFields {
			textFields[k] = strings.Trim(tf, " ")
			reqFields = append(reqFields, strings.Trim(tf, " "))
		}
		relScope.model.SetWBList(WHITELIST, reqFields...)
		// request the data
		err = relScope.model.All(rRes, sqlquery.NewCondition().Where(sel.Condition))
		if err != nil {
			return nil, err
		}

		return rRes, nil
	}

	return nil, nil
}

func (g *gridSource) One(c *sqlquery.Condition, grid *grid.Grid) (interface{}, error) {

	err := g.orm.First(c)
	if err != nil {
		return nil, err
	}

	return g.orm, nil
}

func (g *gridSource) All(c *sqlquery.Condition, grid *grid.Grid) (interface{}, error) {
	// creating result slice
	model := reflect.New(reflect.ValueOf(g.orm).Elem().Type()).Elem() //new type
	resultSlice := reflect.New(reflect.MakeSlice(reflect.SliceOf(model.Type()), 0, 0).Type()).Interface()
	err := g.orm.All(resultSlice, c)
	if err != nil {
		return nil, err
	}

	return resultSlice, nil
}

func (g *gridSource) Create(grid *grid.Grid) (interface{}, error) {
	err := g.unmarshalModel(grid)
	if err != nil {
		return nil, err
	}

	err = g.orm.Create()
	if err != nil {
		return 0, err
	}

	//response with the pkey values
	pkeys := make(map[string]interface{}, 0)
	for _, f := range g.orm.Scope().Fields(Permission{Read: true}) {
		if f.Information.PrimaryKey {
			pkeys[f.Name] = reflect.Indirect(reflect.ValueOf(g.orm)).FieldByName(f.Name).Interface()
		}
	}

	// return with the pkey(s) value
	return pkeys, nil
}

func (g *gridSource) Update(grid *grid.Grid) error {
	err := g.unmarshalModel(grid)
	if err != nil {
		return err
	}

	err = g.orm.Update()
	if err != nil {
		return err

	}
	return nil
}
func (g *gridSource) Delete(c *sqlquery.Condition, grid *grid.Grid) error {
	err := g.orm.First(c)
	if err != nil {
		return err
	}
	err = g.orm.Delete()
	if err != nil {
		return err
	}
	return nil
}

func (g *gridSource) Count(c *sqlquery.Condition) (int, error) {
	return g.orm.Count(c)
}

// marshalModel is needed for create and update.
// It checks if the request json is correct.
// Only struct fields are allowed.
// Empty struct is not allowed.
// Errors will return if one of the rules are not satisfied.
func (g *gridSource) unmarshalModel(grid *grid.Grid) error {
	// reading the body request
	body := grid.Controller().Context().Request.Raw().Body
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
		err := dec.Decode(g.orm)
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

func jsonTagName(scope *Scope, f *grid.Field) bool {

	rField, ok := reflect.TypeOf(scope.Caller()).Elem().FieldByName(f.Id())
	if ok {
		jsonTag := rField.Tag.Get("json")
		if jsonTag == "-" {
			return true
		}

		if jsonTag != "" && jsonTag != "-" {
			if commaIdx := strings.Index(jsonTag, ","); commaIdx != 0 {
				if commaIdx > 0 {
					f.SetId(jsonTag[:commaIdx])
				} else {
					f.SetId(jsonTag)
				}
			}
		}
	}

	return false
}

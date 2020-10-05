package grid

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/patrickascher/gofw/orm"
	"github.com/patrickascher/gofw/slices"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/patrickascher/gofw/sqlquery/types"
	"reflect"
	"strings"
)

var ErrRequestBody = errors.New("grid: request body is empty")
var ErrJsonInvalid = errors.New("grid: json is invalid")

type gridSource struct {
	orm orm.Interface
}

// Grid converts an orm model to an grid source.
func Orm(orm orm.Interface) *gridSource {
	g := &gridSource{}
	g.orm = orm
	return g
}

// Init is called when the source is added to the grid.
func (g *gridSource) Init(grid *Grid) error {
	err := g.orm.Init(g.orm)
	if err != nil {
		return err
	}
	g.orm.Scope().SetReferencesOnly(true)
	return nil
}

// Fields return all defined fields for the frontend.
func (g *gridSource) Fields(grid *Grid) ([]Field, error) {
	return gridFields(g.orm.Scope(), grid, "")
}

func (g *gridSource) Interface() interface{} {
	return g.orm
}

// UpdatedFields is called before grid.Render to update the entered user config to the grid.
func (g *gridSource) UpdatedFields(gr *Grid) error {
	if gr.Mode() != CALLBACK && gr.Mode() != DELETE {

		// get user defined field config
		whitelist, err := whitelistFields(g, gr.Fields(), "")
		if err != nil {
			return err
		}

		if len(whitelist) > 0 {
			g.orm.SetWBList(orm.WHITELIST, whitelist...)
			g.orm.Scope().SetWhitelistExplict(true) // must be called after the WB list is set.
		} else {
			return errors.New("no fields are configured")
		}
	}
	return nil
}

// Callbacks of the frontend.
// At the moment FeSelect is implemented for all select fields.
// select OrmField is used because it can differ from the called select field. TODO simplify, use one version link or Field.
func (g *gridSource) Callback(callback string, gr *Grid) (interface{}, error) {
	if callback == FeUnique {
		selectField, err := gr.Controller().Context().Request.Param("f")
		if err != nil {
			return nil, err
		}
		value, err := gr.Controller().Context().Request.Param("v")
		if err != nil {
			return nil, err
		}
		// set all primary keys
		for _, pk := range g.orm.Scope().PrimaryKeysFieldName() {
			value, err := gr.Controller().Context().Request.Param(pk)
			if err != nil {
				continue
				//return nil, err  // no primary if its a new entry
			}
			err = orm.SetReflectValue(g.orm.Scope().CallerField(pk), reflect.ValueOf(value[0]))
			if err != nil {
				return nil, err
			}
		}
		g.orm.Scope().CallerField(selectField[0]).Set(reflect.ValueOf(value[0]))
		fl := orm.OrmToFieldLevel(selectField[0], g.orm.Scope().Caller())
		return orm.ValidateUnique(fl), nil
	}
	if callback == FeSelect {
		selectField, err := gr.Controller().Context().Request.Param("f")
		if err != nil {
			return nil, err
		}

		// get the defined grid object
		selField := gr.Field(selectField[0])
		if selField.Error() != nil {
			return nil, selField.Error()
		}
		sel := selField.Option(FeSelect).(Select)

		var relScope *orm.Scope
		fields := strings.Split(selectField[0], ".")
		if sel.OrmField != "" {
			fields = strings.Split(sel.OrmField, ".")
		}

		// get relation field of the orm
		relation, err := g.orm.Scope().Relation(fields[0], orm.Permission{Read: true})
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
			relation, err = relScope.Relation(fields[1], orm.Permission{Read: true})
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
		rRes := reflect.New(reflect.MakeSlice(reflect.SliceOf(relation.Type), 0, 0).Type())
		// set whitelist fields
		reqFields := []string{sel.ValueField}
		textFields := strings.Split(sel.TextField, ",")
		for k, tf := range textFields {
			textFields[k] = strings.Trim(tf, " ")
			reqFields = append(reqFields, strings.Trim(tf, " "))
		}
		relScope.Model().SetWBList(orm.WHITELIST, reqFields...)
		// request the data
		err = relScope.Model().All(rRes.Interface(), sqlquery.NewCondition().Where(sel.Condition))
		if err != nil {
			return nil, err
		}

		return reflect.Indirect(rRes).Interface(), nil
	}

	return nil, nil
}

// First returns one row. Used for the grid details and edit view.
func (g *gridSource) First(c *sqlquery.Condition, grid *Grid) (interface{}, error) {
	// fetch data
	err := g.orm.First(c)
	if err != nil {
		return nil, err
	}

	return g.orm, nil
}

// All returns all rows by the given condition.
func (g *gridSource) All(c *sqlquery.Condition, grid *Grid) (interface{}, error) {
	// creating result slice
	model := reflect.New(reflect.ValueOf(g.orm).Elem().Type()).Elem() //new type
	resultSlice := reflect.New(reflect.MakeSlice(reflect.SliceOf(model.Type()), 0, 0).Type())
	// fetch data
	err := g.orm.All(resultSlice.Interface(), c)
	if err != nil {
		return nil, err
	}

	return reflect.Indirect(resultSlice).Interface(), nil
}

// Create a new entry.
func (g *gridSource) Create(grid *Grid) (interface{}, error) {

	err := g.unmarshalModel(grid)
	if err != nil {
		return nil, err
	}

	// TODO CREATE CALLBACKS on CREATE
	if reflect.TypeOf(g.orm).String() == "*auth.User" {
		reflect.ValueOf(g.orm.Scope().Caller()).MethodByName("SetPassword").Call([]reflect.Value{reflect.ValueOf(g.orm.Scope().CallerField("Password").Interface().(orm.NullString).String)})
	}

	err = g.orm.Create()
	if err != nil {
		return 0, err
	}

	//response with the pkey values, because the frontend is reloading the view by id.
	pkeys := make(map[string]interface{}, 0)
	for _, f := range g.orm.Scope().Fields(orm.Permission{Read: true}) {
		if f.Information.PrimaryKey {
			pkeys[f.Name] = reflect.Indirect(reflect.ValueOf(g.orm)).FieldByName(f.Name).Interface()
		}
	}

	// return with the pkey(s) value
	return pkeys, nil
}

// Update the entry
func (g *gridSource) Update(grid *Grid) error {
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

// Delete an entry.
// First the whole row will be fetched and deleted afterwards. To guarantee that relations are deleted as well.
func (g *gridSource) Delete(c *sqlquery.Condition, grid *Grid) error {
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

// Count returns a number of existing rows.
func (g *gridSource) Count(c *sqlquery.Condition, grid *Grid) (int, error) {
	return g.orm.Count(c)
}

// unmarshalModel is needed for create and update.
// It checks if the request json is correct.
// Only struct fields are allowed.
// Empty struct is not allowed.
// Errors will return if one of the rules are not satisfied.
func (g *gridSource) unmarshalModel(gr *Grid) error {
	body := gr.Controller().Context().Request.Body()
	if body == nil {
		return ErrRequestBody
	}

	// check if the json is valid
	if !json.Valid(body) {
		return ErrJsonInvalid
	}

	// unmarshal the request to the model struct
	dec := json.NewDecoder(bytes.NewReader(body))
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

// skipJsonByTag will return true if the json skip tag exists.
// otherwise it checks if a json name is set, and sets the field ID.
func skipJsonByTagOrSetJsonName(scope *orm.Scope, f *Field) bool {

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

func whitelistFields(g *gridSource, fields []Field, parent string) ([]string, error) {
	var whitelist []string

	if parent != "" {
		parent += "."
	}

	for k, f := range fields {
		// check if its a normal field
		if !f.IsRelation() {
			// add if its not removed
			if !f.IsRemoved() {
				if _, exists := slices.Exists(whitelist, parent+f.Id()); !exists {
					whitelist = append(whitelist, parent+f.Id())
				}

				// set write permission to false if its read only.
				// only working on root level.
				if f.IsReadOnly() && parent == "" { // because relation fields can not be fetched
					ormField, err := g.orm.Scope().Field(f.Id())
					if err != nil {
						return nil, err
					}
					ormField.Permission.Write = false
				}
			}
		} else {

			// If sub notations are allowed of the relation, allow the relation but dont whitelist the whole.
			bList, err := whitelistFields(g, f.Fields(), parent+f.Id())
			if err != nil {
				return nil, err
			}
			// if there is a dot notation, add it.
			if len(bList) > 0 {
				for _, tmp := range bList {
					if _, exists := slices.Exists(whitelist, tmp); !exists {
						whitelist = append(whitelist, tmp)
					}
				}
				fields[k].SetRemove(false)
			} else {
				// If the whole relation is allowed.
				if !f.IsRemoved() {
					// add the complete relation if there were no fields set.
					if _, exists := slices.Exists(whitelist, parent+f.Id()); !exists {
						whitelist = append(whitelist, parent+f.Id())
					}
				}
			}

		}
	}

	return whitelist, nil
}

// gridFields is recursively adding the orm fields to the grid.
// If the grid policy is Whitelist, all fields are removed by default for performance reasons and the user has to add it manually on the Grid.
// If the grid policy is Blacklist, all fields are added by default.
// On Type bool or select, the items are added to the frontend.
// Primary, fk, afk keys are removed by default - whitelist & blacklist.
// All the validator tags are added for the frontend.
// self referenced relations are skipped.
// BelongsTo has a special case, on views != table view, only the root fk is loaded instead of the relation, because its a dropdown in the frontend and only the ID is needed.
// TODO logic for m2m, same?
// Select Items are added automatically by callback. The Value field is the afk and the text field is the third Field? TODO why third and not 2nd?
func gridFields(scope *orm.Scope, g *Grid, parent string) ([]Field, error) {

	var rv []Field
	i := 0

	// normal fields
	for _, f := range scope.Fields(orm.Permission{Read: true}) {
		field := Field{}
		field.SetId(f.Name)
		if skipJsonByTagOrSetJsonName(scope, &field) {
			continue
		}
		field.SetPrimary(f.Information.PrimaryKey)
		field.SetDatabaseId(f.Information.Table + "." + f.Information.Name)
		field.SetFieldType(f.Information.Type.Kind())
		field.SetTitle("ORM§§" + scope.Name(true) + "§§" + f.Name)
		//field.SetDescription(grid.NewValue(""))
		field.SetPosition(i)
		field.SetRemove(false)
		field.SetHidden(NewValue(false))
		//field.SetView(g.NewValue(""))
		field.SetSortable(true)
		field.SetFilterable(true)
		field.SetGroupable(true)
		// set validation tag
		if f.Validator.Config != "" {
			field.SetOption(orm.TagValidate, f.Validator.Config)
		}

		if f.Information.Type.Kind() == "Select" {
			var items []SelectItem
			sel := f.Information.Type.(types.Select)
			for _, i := range sel.Items() {
				items = append(items, SelectItem{Text: i, Value: i})
			}
			field.SetOption(FeSelect, Select{Items: items})
		}

		// field manipulations
		// Primary keys are not shown on frontend header.
		// In the backend they are loaded anyway because they are mandatory for relations.
		if f.Information.PrimaryKey {
			field.SetRemove(true)
		}

		// the association key is getting removed.
		for _, relation := range scope.Relations(orm.Permission{Read: true}) {
			if relation.Kind == orm.BelongsTo {
				if field.Id() == relation.ForeignKey.Name {
					field.SetRemove(true)
				}
			}
		}

		rv = append(rv, field)
		i++
	}

	// relation fields
	for _, r := range scope.Relations(orm.Permission{Read: true}) {

		// skip on self referencing orm models
		if _, err := scope.Parent(r.Type.String()); err == nil {
			continue
		}

		// if its a belongsTo relation and the view is no grid table, only the association field is loaded because this is
		// a dropdown on the frontend and only the ID is needed.
		// TODO if this changes to a combobox in the future, this code has to get changed.
		if r.Kind == orm.BelongsTo && g.Mode() != VTable {
			for k := range rv {
				if rv[k].Id() == r.ForeignKey.Name {
					if parent != "" {
						parent += "."
					}
					rv[k].SetFieldType(r.Kind)
					rv[k].SetRemove(false)
					rv[k].SetOption(FeSelect, Select{OrmField: parent + r.Field, TextField: r.Type.Field(2).Name, ValueField: r.AssociationForeignKey.Name})
					rv[k].SetOption("vueReturnObject", false)
				}
			}

			// the whole relation has to get removed. otherwise the single foreign key will not with the existing logic.
			// why: the relation object is not changing in the frontend, only the afk on the root model. But the afk is manipulated automatically of the orm on create.
			// another reason why this is handled like this, the whole belongsTo object has to get loaded for every select item - performance risk.
			field := Field{}
			field.SetId(r.Field)
			field.SetRemove(true)
			field.SetRelation(true)
			rv = append(rv, field)

			continue
		}

		// relation field
		field := Field{}
		field.SetId(r.Field)
		field.SetRelation(true)
		if skipJsonByTagOrSetJsonName(scope, &field) {
			continue
		}
		field.SetFieldType(r.Kind)
		field.SetTitle("ORM§§" + scope.Name(true) + "§§" + r.Field)

		//field.SetDescription(g.NewValue(""))
		field.SetPosition(i)
		field.SetRemove(false)
		//field.SetHidden(false)
		//field.SetView(g.NewValue(""))
		if r.Validator.Config != "" {
			field.SetOption(orm.TagValidate, r.Validator.Config)
		}
		field.SetSortable(false)
		field.SetFilterable(false)
		field.SetGroupable(false)

		// add options for BelongsTo and ManyToMany relations.
		if r.Kind == orm.BelongsTo || r.Kind == orm.ManyToMany {
			// experimental (model,id, title...) by default always the third field is taken.
			field.SetOption(FeSelect, Select{TextField: r.Type.Field(2).Name, ValueField: r.AssociationForeignKey.Name})
		}

		// recursively add fields
		rScope, err := scope.NewScopeFromType(r.Type)
		if err != nil {
			return nil, err
		}
		rScope.SetParent(scope.Model()) // adding parent to avoid loops

		rField, err := gridFields(rScope, g, r.Field)
		if err != nil {
			return nil, err
		}

		// field manipulations - FK,AFK,Poly are removed.
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

		// adding relation fields
		if len(rField) > 0 {
			field.SetFields(rField)
		}

		rv = append(rv, field)
		i++

	}

	return rv, nil
}

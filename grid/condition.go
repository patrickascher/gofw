package grid

import (
	"errors"
	"fmt"
	"github.com/patrickascher/gofw/sqlquery"
	"strings"
)

// Condition configuration
const (
	ConditionSeparator    = ","
	ConditionFilterPrefix = "filter_"
)

// All errors are defined here
var (
	ErrSort        = errors.New("grid: the field %#v is not sortable")
	ErrField       = errors.New("grid: the field %#v was not found")
	ErrFilter      = errors.New("grid: the field %#v is not filterable")
	ErrPrimaryKeys = errors.New("grid: primary key(s) are not set")
)

// conditionAll creates a condition for the VIEW_GRID.
// Used in readAll.
func conditionAll(g *Grid) (*sqlquery.Condition, error) {
	c := &sqlquery.Condition{}

	if g.srcCondition != nil {
		*c = *g.srcCondition
	}

	params, err := g.controller.Context().Request.Params()
	if err != nil {
		return nil, err
	}

	for key, param := range params {
		if key == "sort" {
			c.Reset(sqlquery.ORDER)
			err := addSortCondition(g, param[0], c)
			if err != nil {
				return nil, err
			}
		}
		if strings.HasPrefix(key, ConditionFilterPrefix) {
			err := addFilterCondition(g, key[len(ConditionFilterPrefix):], param, c)
			if err != nil {
				return nil, err
			}
		}
	}

	return c, nil
}

// addSortCondition adds an order by condition with the given params.
// If the field is not allowed to sort or the field does not exist, an error will return.
func addSortCondition(g *Grid, params string, c *sqlquery.Condition) error {
	sortFields := strings.Split(params, ConditionSeparator)
	var orderFields []string

	// skip if no arguments
	if len(sortFields) == 1 && sortFields[0] == "" {
		return nil
	}

	// checking if the field is allowed for sorting
	for _, f := range sortFields {

		columnField, err := isSortAllowed(g, f)
		if err != nil {
			return err
		}

		orderFields = append(orderFields, columnField)
	}

	// adding order
	c.Order(orderFields...)

	return nil
}

// addFilterCondition adds a where condition with the given params.
// If the field is not allowed to filter or the field does not exist, an error will return.
// If there are more than one argument, its converting it in a WHERE IN condition.
func addFilterCondition(g *Grid, field string, params []string, c *sqlquery.Condition) error {
	field, err := isFilterAllowed(g, field)
	if err != nil {
		return err
	}

	args := strings.Split(params[0], ConditionSeparator)
	if len(args) > 1 {
		c.Where(field+" IN(?)", args)
	}

	if len(args) == 1 {
		c.Where(field+" = ?", args[0])
	}

	return nil
}

// isSortAllowed checks if the field exists and if the field is sortable and returns the db column name.
// Otherwise an error will return
func isSortAllowed(g *Grid, name string) (string, error) {
	fieldName := name
	prefix := ""
	if strings.HasPrefix(name, "-") {
		fieldName = name[1:]
		prefix = "-"
	}

	// check if field exists
	field, err := g.getFieldByStructName(fieldName)
	if err != nil {
		return "", err
	}

	// check if sort is allowed
	if !field.getSort() {
		return "", fmt.Errorf(ErrSort.Error(), fieldName)
	}

	return prefix + field.getColumn().Information.Name, nil
}

// isFilterAllowed returns an error if the field does not exist or the field is not allowed as filter.
// If everything is ok, the db column name will return.
func isFilterAllowed(g *Grid, name string) (string, error) {
	field, err := g.getFieldByStructName(name)
	if err != nil {
		return "", err
	}
	// filter is not allowed
	if !field.getFilter() {
		return "", fmt.Errorf(ErrFilter.Error(), name)
	}

	// filter allowed
	return field.getColumn().Information.Name, nil
}

// getFieldByName returns the field by structname
// If it does not exist, an error will return
func (g *Grid) getFieldByStructName(name string) (Interface, error) {

	if _, ok := g.fields[name]; !ok {
		return nil, fmt.Errorf(ErrField.Error(), name)

	}
	return g.fields[name], nil
}

// checkPrimaryParams checks all params if the primary key(s) exists
// if one is empty, a error will return.
// if everything is correct, a condition will return.
// Used in read one, update, delete.
func checkPrimaryParams(g *Grid) (*sqlquery.Condition, error) {

	// if no params exist
	params, err := g.controller.Context().Request.Params()
	if err != nil || len(params) == 0 {
		return nil, ErrPrimaryKeys
	}

	c := &sqlquery.Condition{}

	// looping all fields
	for _, field := range g.fields {
		// checking all primary keys
		if len(field.getFields()) == 0 && field.getColumn().Information.PrimaryKey {
			v, err := g.controller.Context().Request.Param(field.getColumn().Name)
			if err != nil {
				return nil, ErrPrimaryKeys
			}
			c.Where(field.getColumn().Information.Table+"."+field.getColumn().Information.Name+" = ?", v[0])
		}
	}

	return c, nil
}

package grid2

import (
	"errors"
	"fmt"
	"github.com/patrickascher/gofw/sqlquery"
	"strings"
)

const (
	ConditionSeparator    = ","
	ConditionFilterPrefix = "filter_"
)

var (
	errPrimaryMissing   = errors.New("grid: primary is not set")
	errSortPermission   = "grid: field %s id not allowed to sort or does not exist"
	errFilterPermission = "grid: field %s id not allowed to filter or does not exist"
)

// conditionOne returns a condition for one row. This is used for the grid views update, details and create.
// if a grid src.condition exists, its getting passed through.
// error will return if not all primary keys are filled or non is existing.
func (g *Grid) conditionOne() (*sqlquery.Condition, error) {

	// if no params exist, exit
	params, err := g.controller.Context().Request.Params()
	if err != nil || len(params) == 0 {
		return nil, errPrimaryMissing
	}

	// create a new condition.
	// if a user condition exist, the value will be copied. This should provide additional security.
	c := sqlquery.NewCondition()
	if g.srcCondition != nil {
		*c = *g.srcCondition
	}

	// looping all fields
	var primary bool
	for _, f := range g.fields {
		// skipping relations
		if len(f.fields) > 0 {
			continue
		}

		// checking all primary keys
		if f.primary {
			v, err := g.controller.Context().Request.Param(f.id)

			if err != nil {
				return nil, errPrimaryMissing
			}
			primary = true
			c.Where(f.referenceId+" = ?", v[0])
		}
	}

	// no primary was defined in the grid fields
	if !primary {
		return nil, errPrimaryMissing
	}
	return c, nil
}

// conditionAll return a condtion for the grid table view.
// if a grid src.condition exists, its getting passed through.
// sort and filter params are checked. (sort=ID,-Name) (filter_ID=1&filter_Name=Patrick,Tom)
// Error will return if the sort/filter key does not exist or has no permission.
func (g *Grid) conditionAll() (*sqlquery.Condition, error) {

	// create a new condition.
	// if a user condition exist, the value will be copied.
	c := sqlquery.NewCondition()
	if g.srcCondition != nil {
		*c = *g.srcCondition
	}

	// get all request params of the controller.
	params, err := g.controller.Context().Request.Params()
	if err != nil {
		return nil, err
	}

	// iterate over the params.
	// check if the key sort exists or the key is prefixed with filter_
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
		prefix := ""
		if strings.HasPrefix(f, "-") {
			f = f[1:]
			prefix = "-"
		}
		if gridField := g.Field(f); gridField.error == nil && gridField.IsSortable() {
			orderFields = append(orderFields, prefix+gridField.referenceId)
		} else {
			return fmt.Errorf(errSortPermission, f)
		}
	}

	// adding order
	c.Order(orderFields...)

	return nil
}

// addFilterCondition adds a where condition with the given params.
// If the field is not allowed to filter or the field does not exist, an error will return.
// If there are more than one argument, a WHERE IN (?) will be added.
func addFilterCondition(g *Grid, field string, params []string, c *sqlquery.Condition) error {

	if gridField := g.Field(field); gridField.error == nil && gridField.IsFilterable() {
		args := strings.Split(params[0], ConditionSeparator)
		if len(args) > 1 {
			c.Where(field+" IN(?)", args)
		}

		if len(args) == 1 {
			c.Where(field+" = ?", args[0])
		}
		return nil
	}

	return fmt.Errorf(errFilterPermission, field)
}

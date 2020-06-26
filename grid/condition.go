package grid

import (
	"errors"
	"fmt"
	"github.com/patrickascher/gofw/sqlquery"
	"strconv"
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
func (g *Grid) conditionFirst() (*sqlquery.Condition, error) {

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

	// TODO first check filter,
	filter, err := g.controller.Context().Request.Param("filter")
	if err == nil {

		id, err := strconv.Atoi(filter[0])
		if err != nil {
			return nil, err
		}

		uFilter, err := getFilterByID(id, g)
		if err != nil {
			return nil, err
		}

		// set active Filter
		g.config.Filter.Active.ID = uFilter.ID
		if uFilter.RowsPerPage.Valid {
			g.config.Filter.Active.RowsPerPage = int(uFilter.RowsPerPage.Int64)
		}

		// Position/Disable fields
		for _, f := range uFilter.Fields {
			if gridField := g.Field(f.Key); gridField.error == nil && !gridField.IsRemoved() {
				if f.Pos.Valid {
					gridField.SetPosition(int(f.Pos.Int64))
				}
				if f.Show == false {
					gridField.SetRemove(true)
				}
			}
		}

		// Add filters
		for _, f := range uFilter.Filters {
			if gridField := g.Field(f.Key); gridField.error == nil && gridField.IsFilterable() {
				switch f.Op {
				case "=", ">=", "<=":
					c.Where(gridField.referenceId+" "+f.Op+" ?", f.Value)
				case "IN", "NOT IN":
					c.Where(gridField.referenceId+" "+f.Op+" (?)", strings.Split(f.Value, ","))
				case "Like":
					c.Where(gridField.referenceId+" LIKE %?%", f.Value)
				case "RLike":
					c.Where(gridField.referenceId+" LIKE %?", f.Value)
				case "LLike":
					c.Where(gridField.referenceId+" LIKE ?%", f.Value)
				default:
					return nil, fmt.Errorf(errFilterPermission, f.Key)
				}
			} else {
				return nil, fmt.Errorf(errFilterPermission, f.Key)
			}
		}

		// Add sorts
		var sort string
		// add grouping as fist param
		if uFilter.GroupBy.Valid {
			sort += uFilter.GroupBy.String
			g.config.Filter.Active.Group = uFilter.GroupBy.String
			//TODO ASC or DESC
		}
		// add order by
		for _, s := range uFilter.Sorting {
			if sort != "" {
				sort += ", "
			}
			op := "ASC"
			if s.Desc {
				op = "DESC"
			}

			if gridField := g.Field(s.Key); gridField.error == nil && gridField.IsSortable() {
				sort += gridField.referenceId + " " + op
				g.config.Filter.Active.Sort = append(g.config.Filter.Active.Sort, s.Key+" "+op)
			} else {
				return nil, fmt.Errorf(errSortPermission, s.Key)
			}

		}
		c.Order(sort)
	}
	// Then add additional sorting? + resert sort first

	// iterate over the params.
	// check if the key sort exists or the key is prefixed with filter_
	for key, param := range params {
		if key == "sort" {
			c.Reset(sqlquery.ORDER)
			g.config.Filter.Active.Sort = nil // reset config order
			err := addSortCondition(g, param[0], c)
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

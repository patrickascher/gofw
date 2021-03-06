package grid

import (
	"errors"
	"fmt"
	"github.com/patrickascher/gofw/sqlquery"
	"strconv"
	"strings"
)

const (
	ConditionFilterSeparator = ";"
	ConditionSeparator       = ","
	ConditionFilterPrefix    = "filter_"
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

// escape is needed to escape the % and _ for sql.
// at the moment it will only be prefixed with a backslash.
// TODO for the future every driver should have his own escaping function.
func escape(v string) string {
	return strings.ReplaceAll(strings.ReplaceAll(v, "%", "\\%"), "_", "\\_")
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
		if len(uFilter.Fields) > 0 {
			for i, f := range g.Fields() {
				remove := true
				for _, uf := range uFilter.Fields {
					if uf.Key == f.id {
						g.fields[i].SetPosition(uf.Pos)
						remove = false
						break
					}
				}
				g.fields[i].SetRemove(remove)
			}
		}

		// Add filters
		//TODO Mysql,Oracle have different ways to add/sub dates. create a driver based date function.
		for _, f := range uFilter.Filters {
			if gridField := g.Field(f.Key); gridField.error == nil && gridField.IsFilterable() {

				switch f.Op {
				case "TODAY":
					c.Where(gridField.referenceId + " >= DATENOW")
					c.Where(gridField.referenceId + " <= DATENOW")
				case "YESTERDAY":
					c.Where(gridField.referenceId + " >= DATENOW-1")
					c.Where(gridField.referenceId + " < DATENOW")
				case "WEEK":
					c.Where(gridField.referenceId + " = WEEK")
				case "LWEEK":
					c.Where(gridField.referenceId + " = WEEK-1")
				case "MONTH":
					c.Where(gridField.referenceId + " = MONTH")
				case "LMONTH":
					c.Where(gridField.referenceId + " = MONTH-1")
				case "YEAR":
					c.Where(gridField.referenceId + " = YEAR")
				case "LYEAR":
					c.Where(gridField.referenceId + " = YEAR-1")
				case "!=", "=", ">=", "<=":
					c.Where(gridField.referenceId+" "+f.Op+" ?", escape(f.Value.String))
				case "IN":
					c.Where(gridField.referenceId+" "+f.Op+" (?)", strings.Split(escape(f.Value.String), ConditionFilterSeparator))
				case "NOTIN":
					c.Where(gridField.referenceId+" NOT IN (?)", strings.Split(escape(f.Value.String), ConditionFilterSeparator))
				case "NULL":
					c.Where(gridField.referenceId + " IS NULL")
				case "NOTNULL":
					c.Where(gridField.referenceId + " IS NOT NULL")
				case "Like":
					c.Where(gridField.referenceId+" LIKE ?", "%%"+escape(f.Value.String)+"%%")
				case "RLike":
					c.Where(gridField.referenceId+" LIKE ?", escape(f.Value.String)+"%%")
				case "LLike":
					c.Where(gridField.referenceId+" LIKE ?", "%%"+escape(f.Value.String))
				default:
					return nil, fmt.Errorf(errFilterPermission, escape(f.Key))
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
		if strings.HasPrefix(key, ConditionFilterPrefix) {
			err := addFilterCondition(g, key[len(ConditionFilterPrefix):], param, c)
			if err != nil {
				return nil, err
			}
		}
	}

	return c, nil
}

// addFilterCondition adds a where condition with the given params.
// If the field is not allowed to filter or the field does not exist, an error will return.
// If there are more than one argument, a WHERE IN (?) will be added.
func addFilterCondition(g *Grid, field string, params []string, c *sqlquery.Condition) error {

	if gridField := g.Field(field); gridField.error == nil && gridField.IsFilterable() {

		args := strings.Split(escape(params[0]), ConditionFilterSeparator)

		if gridField.where != "" {
			c.Where(gridField.where, "%%"+args[0]+"%%")
			return nil
		}

		if len(args) > 1 {
			c.Where(field+" IN(?)", args)
		}

		if len(args) == 1 {
			c.Where(field+" LIKE ?", "%%"+args[0]+"%%")
		}

		return nil
	}

	return fmt.Errorf(errFilterPermission, field)
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

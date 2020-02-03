package sqlquery

import (
	"strconv"
	"strings"

	"errors"
	"fmt"
	"reflect"
)

//Exported CONDITION types
const (
	WHERE = iota + 1
	HAVING
	LIMIT
	ORDER
	OFFSET
	GROUP
	ON
)

// Condition is holding all condition statements and arguments.
//TODO: rewrite the conditions that the actual rendering t string is happening in the render function
type Condition struct {
	where  string
	having string
	limit  string
	order  string
	offset string
	group  string
	on     string

	args  map[int][]interface{}
	error error
}

// This was only created for the loop detection in the orm module.
// TODO create a more usefull function here or figure something out in the orm module.
func (c *Condition) Config(condition ...int) string {
	for _, fn := range condition {
		switch fn {
		case WHERE:
			return fmt.Sprintf(strings.Replace(c.where, PLACEHOLDER, "%v", -1), c.args[WHERE])
		case HAVING:
			return fmt.Sprintf(strings.Replace(c.having, PLACEHOLDER, "%v", -1), c.args[HAVING])
		case LIMIT:
			return c.limit
		case ORDER:
			return c.order
		case OFFSET:
			return c.offset
		case GROUP:
			return c.group
		case ON:
			return fmt.Sprintf(strings.Replace(c.on, PLACEHOLDER, "%v", -1), c.args[ON])
		}
	}
	return ""
}

// Reset the condition by a specific clause.
func (c *Condition) Reset(reset ...int) {

	for _, fn := range reset {
		switch fn {
		case WHERE:
			c.where = ""
			if c.args[WHERE] != nil {
				c.args[WHERE] = []interface{}{}
			}
		case HAVING:
			c.having = ""
			if c.args[HAVING] != nil {
				c.args[HAVING] = []interface{}{}
			}
		case LIMIT:
			c.limit = ""
		case ORDER:
			c.order = ""
		case OFFSET:
			c.offset = ""
		case GROUP:
			c.group = ""
		case ON:
			c.on = ""
			if c.args[ON] != nil {
				c.args[ON] = []interface{}{}
			}
		}
	}
}

// Where can be called multiple times.
// The condition gets connected by an AND.
// This means, if you need an OR Condition, be aware to set the right brackets or write the whole Condition in one Where call.
func (c *Condition) Where(stmt string, args ...interface{}) *Condition {
	c.conditionHelper(WHERE, stmt, args)
	return c
}

// Group by condition.
// Usage c.Group("id","id2")
func (c *Condition) Group(group ...string) *Condition {

	for _, stmt := range group {

		if c.group != "" {
			c.group += ", "
		}

		c.group = c.group + stmt
	}

	c.group = " GROUP BY " + c.group
	return c
}

// Having can be called multiple times. The condition gets connected by an AND.
// This means, if you need an OR Condition, be aware to set the right brackets or write the whole Condition in one Having call.
func (c *Condition) Having(stmt string, args ...interface{}) *Condition {
	c.conditionHelper(HAVING, stmt, args)
	return c
}

// Order by condition.
// Usage con.Order("id DESC")
func (c *Condition) Order(order ...string) *Condition {
	for _, stmt := range order {

		if c.order != "" {
			c.order += ", "
		}

		// uppercase user asc,desc inserts
		stmt = strings.Replace(stmt, " asc", " ASC", 1)
		stmt = strings.Replace(stmt, " desc", " DESC", 1)

		if strings.HasPrefix(stmt, "-") {
			stmt = stmt[1:] + " DESC"
		} else if !strings.HasSuffix(strings.ToUpper(stmt), "ASC") && !strings.HasSuffix(strings.ToUpper(stmt), "DESC") {
			stmt = stmt + " ASC"
		}

		c.order = c.order + stmt
	}

	c.order = " ORDER BY " + c.order
	return c
}

// Limit condition.
// Usage c.Limit(5)
func (c *Condition) Limit(l int) *Condition {
	c.limit = " LIMIT " + strconv.Itoa(l)
	return c
}

// Offset by condition.
// Usage c.Offset(1)
func (c *Condition) Offset(o int) *Condition {
	c.offset = " OFFSET " + strconv.Itoa(o)
	return c
}

// On condition for sql joins.
// Usage: c.On("company.id = employee.id",nil)
func (c *Condition) On(stmt string, args ...interface{}) *Condition {
	c.conditionHelper(ON, stmt, args)
	return c
}

// stmtMapManipulation is a helper for adding arguments which are the type array or slice.
// It manipulates the statement ex.: `Where("id IN (?)",[1,2,3]` into `id IN (?,?,?)` and it  appends all given arguments.
func stmtMapManipulation(c *Condition, stmt string, args []interface{}, conditionType int) string {

	//initialize arguments
	if len(c.args) == 0 {
		c.args = make(map[int][]interface{})
	}

	//manipulate statement
	if len(args) >= 1 {
		for i := 0; i < len(args); i++ {
			//handel array arguments
			if reflect.ValueOf(args[i]).Kind() == reflect.Array || reflect.ValueOf(args[i]).Kind() == reflect.Slice {
				//split after placeholder and only replace the map placeholder
				spStmt := strings.SplitAfter(stmt, PLACEHOLDER)
				spStmt[i] = strings.Replace(spStmt[i], PLACEHOLDER, PLACEHOLDER+strings.Repeat(", "+PLACEHOLDER, reflect.ValueOf(args[i]).Len()-1), -1)
				stmt = strings.Join(spStmt, "")
			}
			// add single or (map,slice) arguments
			c.addArgument(conditionType, args[i])
		}
	}

	return stmt
}

// addArgument appends all given arguments to Condition.args.
// in a slice or array all int's are casted to an int64
func (c *Condition) addArgument(conditionType int, args interface{}) {

	//Array/Slice arguments
	if reflect.ValueOf(args).Kind() == reflect.Array || reflect.ValueOf(args).Kind() == reflect.Slice {
		for n := 0; n < reflect.ValueOf(args).Len(); n++ {
			switch reflect.ValueOf(args).Index(n).Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				val := reflect.ValueOf(args).Index(n).Int()
				c.args[conditionType] = append(c.args[conditionType], val)
			case reflect.String:
				val := reflect.ValueOf(args).Index(n).String()
				c.args[conditionType] = append(c.args[conditionType], val)
			default:
				c.error = errors.New("sql: this argument type is not allowed")
			}
		}
		return
	}

	//single argument
	c.args[conditionType] = append(c.args[conditionType], args)
}

// arguments merges the condition arguments in the right order.
func (c *Condition) arguments() []interface{} {
	var arguments []interface{}
	arguments = append(arguments, c.args[ON]...)
	arguments = append(arguments, c.args[WHERE]...)
	arguments = append(arguments, c.args[HAVING]...)

	return arguments
}

// conditionHelper adding the given condition as string to the struct.
// The
func (c *Condition) conditionHelper(conditionType int, stmt string, args []interface{}) {

	//no statement given
	if stmt == "" {
		return
	}

	sqlStmt := ""

	//compare placeholders and arguments length, return error if there is a mismatch
	if strings.Count(stmt, PLACEHOLDER) != len(args) {
		c.error = fmt.Errorf("%v placeholder(%v) and arguments(%v) does not fit", sqlStmt, strings.Count(stmt, PLACEHOLDER), len(args))
		return
	}

	stmt = stmtMapManipulation(c, stmt, args, conditionType)

	//building statement
	switch conditionType {
	case WHERE:
		if c.where == "" {
			sqlStmt += " WHERE"
		}
		if sqlStmt != " WHERE" {
			sqlStmt += c.where + " AND"
		}
		c.where = sqlStmt + " " + stmt
	case HAVING:
		if c.having == "" {
			sqlStmt += " HAVING"
		}
		if sqlStmt != " HAVING" {
			sqlStmt += c.having + " AND"
		}
		c.having = sqlStmt + " " + stmt
	case ON:
		c.on = " ON " + stmt
	}

}

// render the condition stmt in the right order.
// Error will return if the arguments does not fit to the placeholder.
func (c *Condition) render(p *Placeholder) (string, error) {

	if c.error != nil {
		return "", c.error
	}

	//replace the package placeholder with the driver placeholder
	condition := c.where + c.group + c.having + c.order + c.limit + c.offset
	for i := 1; i <= strings.Count(c.where+c.group+c.having+c.order+c.limit+c.offset, PLACEHOLDER); i++ {
		condition = strings.Replace(condition, PLACEHOLDER, p.placeholder(), 1)
	}

	return condition, nil
}

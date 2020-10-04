// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlquery

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Allowed conditions
const (
	WHERE = iota + 1
	HAVING
	LIMIT
	ORDER
	OFFSET
	GROUP
	ON
)

// Error messages.
var (
	ErrArgumentType        = errors.New("sqlquery: this argument type %v is not allowed")
	ErrPlaceholderMismatch = errors.New("sqlquery: %v placeholder(%v) and arguments(%v) does not fit")
)

// Condition type.
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

func Copy(c *Condition) *Condition {
	newc := NewCondition()

	newc.where = c.where
	newc.having = c.having
	newc.limit = c.limit
	newc.order = c.order
	newc.offset = c.offset
	newc.group = c.group
	newc.on = c.on
	newc.error = c.error
	newc.args = make(map[int][]interface{})

	for i := range c.args {
		newc.args[i] = make([]interface{}, len(c.args[i]))
		copy(newc.args[i], c.args[i])
	}

	return newc
}

func NewCondition() *Condition {
	return &Condition{}
}

// Config returns the requested condition.
// The first argument defines if the result should include the values or not.
// Caution, the values are not escaped and should only be used for test or debugging. On database level these are placeholders and getting escaped by the driver.
// 		c.Config(false,WHERE,ORDER) // WHERE and ORDER are being rendered.
func (c *Condition) Config(values bool, condition ...int) string {
	var stmt []string

	for _, fn := range condition {
		switch fn {
		case WHERE:
			s := c.where
			if s != "" {
				if values {
					stmt = append(stmt, fmt.Sprintf(strings.Replace(c.where, PLACEHOLDER, "%v", -1), c.args[WHERE]...))
				} else {
					stmt = append(stmt, s)
				}
			}
		case HAVING:
			s := c.having
			if s != "" {
				if values {
					stmt = append(stmt, fmt.Sprintf(strings.Replace(c.having, PLACEHOLDER, "%v", -1), c.args[HAVING]...))
				} else {
					stmt = append(stmt, s)
				}
			}
		case LIMIT:
			if c.limit != "" {
				stmt = append(stmt, c.limit)
			}
		case ORDER:
			if c.order != "" {
				stmt = append(stmt, c.order)
			}
		case OFFSET:
			if c.offset != "" {
				stmt = append(stmt, c.offset)
			}
		case GROUP:
			if c.group != "" {
				stmt = append(stmt, c.group)
			}
		case ON:
			s := c.on
			if s != "" {
				if values {
					stmt = append(stmt, fmt.Sprintf(strings.Replace(c.on, PLACEHOLDER, "%v", -1), c.args[ON]...))
				} else {
					stmt = append(stmt, s)
				}
			}
		}
	}

	return strings.Join(stmt, " ")
}

// Reset the condition by one or more conditions.
// If the argument is empty, all conditions are reset.
//		c.Reset() // all will be reset
// 		c.Reset(WHERE,HAVING) // only WHERE and HAVING are reset.
func (c *Condition) Reset(reset ...int) {

	if len(reset) == 0 {
		reset = []int{ON, WHERE, GROUP, HAVING, ORDER, LIMIT, OFFSET}
	}

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

// DEPRECATED - this is just a quick fix. the wholde sqlquery.condition has to get rewritten because of the where string.
// not enough manipulation chances.
func (c *Condition) SetWhere(w string) {
	c.where = w
}

// Where condition.
// Where can be called multiple times on a sql statement and gets chained by AND.
// If you need an OR Condition, be aware to set the right brackets or write the whole condition in one WHERE call.
// Arrays and slices can be passed as argument.
//		c.Where("id = ?",1)
//		c.Where("id IN (?)",[]int{10,11,12})
func (c *Condition) Where(stmt string, args ...interface{}) *Condition {
	c.conditionHelper(WHERE, stmt, args)
	return c
}

// Group condition.
// Group should only be called once. If its called more often, the last values count.
// Column names are not quoted TODO?.
// 		c.Group("id","name") // GROUP BY id, name
func (c *Condition) Group(group ...string) *Condition {
	// skipping empty call or string
	if len(group) == 0 || (len(group) == 1 && group[0] == "") {
		return c
	}

	// reset if group would be used twice
	c.Reset(GROUP)
	for _, stmt := range group {
		if c.group != "" {
			c.group += ", "
		}
		c.group = c.group + stmt
	}

	c.group = "GROUP BY " + c.group
	return c
}

// Having condition.
// Having can be called multiple times on a sql statement and gets chained by AND.
// If you need an OR Condition, be aware to set the right brackets or write the whole condition in one HAVING call.
// Arrays and slices can be passed as argument.
//		c.Having("amount > ?",100)
//		c.Having("id IN (?)",[]int{10,11,12})
func (c *Condition) Having(stmt string, args ...interface{}) *Condition {
	c.conditionHelper(HAVING, stmt, args)
	return c
}

// Order condition.
// If a column has a `-` prefix, DESC order will get set.
// Order should only be called once. If its called more often, the last values count.
// Column names are not quoted TODO?.
// 		c.Order("id","-name") // ORDER BY id ASC, name DESC
func (c *Condition) Order(order ...string) *Condition {
	// skipping empty call or string
	if len(order) == 0 || (len(order) == 1 && order[0] == "") {
		return c
	}

	c.Reset(ORDER)
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

	c.order = "ORDER BY " + c.order
	return c
}

// Limit condition.
// Limit should be called once. If its called more often, the last values count.
//		c.Limit(10)
func (c *Condition) Limit(l int) *Condition {
	c.limit = "LIMIT " + strconv.Itoa(l)
	return c
}

// Offset condition.
// Offset should be called once. If its called more often, the last values count.
//		c.Offset(5)
func (c *Condition) Offset(o int) *Condition {
	c.offset = "OFFSET " + strconv.Itoa(o)
	return c
}

// On condition for sql joins.
// On should only be called once. If its called more often, the last values count.
// Arrays and slices can be passed as argument.
//		c.On("user.company = company.id AND user.id > ?",100)
func (c *Condition) On(stmt string, args ...interface{}) *Condition {
	c.Reset(ON)
	c.conditionHelper(ON, stmt, args)
	return c
}

// stmtMapManipulation is a helper for adding arguments which are the type array or slice.
// It manipulates the statement ex.: `Where("id IN (?)",[1,2,3]` into `id IN (?,?,?)` and it appends all given arguments.
func stmtMapManipulation(c *Condition, stmt string, args []interface{}, conditionType int) string {

	// initialize arguments
	if len(c.args) == 0 {
		c.args = make(map[int][]interface{})
	}

	// manipulate statement
	if len(args) >= 1 {
		for i := 0; i < len(args); i++ {
			// handel array arguments
			if reflect.ValueOf(args[i]).Kind() == reflect.Array || reflect.ValueOf(args[i]).Kind() == reflect.Slice {
				//split after placeholder and only replace the map placeholder
				spStmt := strings.SplitAfter(stmt, PLACEHOLDER)
				// because of this logic, the append placeholders need a different name without ?. TODO create a more prof. solution.
				spStmt[i] = strings.Replace(spStmt[i], PLACEHOLDER, PLACEHOLDER+strings.Repeat(", "+PLACEHOLDER_APPEND, reflect.ValueOf(args[i]).Len()-1), -1)
				stmt = strings.Join(spStmt, "")
			}
			// add single or (map,slice) arguments
			c.addArgument(conditionType, args[i])
		}
	}
	stmt = strings.Replace(stmt, PLACEHOLDER_APPEND, PLACEHOLDER, -1)

	return stmt
}

// addArgument appends all given arguments to Condition.args.
// in a slice or array all int's are casted to an int64
// Only int and string types are allowed.
func (c *Condition) addArgument(conditionType int, args interface{}) {

	// Array/Slice arguments
	if reflect.ValueOf(args).Kind() == reflect.Array || reflect.ValueOf(args).Kind() == reflect.Slice {
		for n := 0; n < reflect.ValueOf(args).Len(); n++ {
			switch t := reflect.TypeOf(reflect.ValueOf(args).Index(n).Interface()).Kind(); t {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				var val int64
				if t == reflect.Int || t == reflect.Int32 {
					val = int64(reflect.ValueOf(args).Index(n).Interface().(int))
				}
				if t == reflect.Int8 {
					val = int64(reflect.ValueOf(args).Index(n).Interface().(int8))
				}
				if t == reflect.Int16 {
					val = int64(reflect.ValueOf(args).Index(n).Interface().(int16))
				}
				if t == reflect.Int64 {
					val = reflect.ValueOf(args).Index(n).Interface().(int64)
				}

				c.args[conditionType] = append(c.args[conditionType], val)
			case reflect.String:
				val := reflect.ValueOf(args).Index(n).Interface().(string)
				c.args[conditionType] = append(c.args[conditionType], val)
			default:
				c.error = fmt.Errorf(ErrArgumentType.Error(), reflect.ValueOf(args).Index(n).Kind())
			}
		}
		return
	}

	// single argument
	c.args[conditionType] = append(c.args[conditionType], args)
}

// arguments merges the condition arguments in the right order (ON, WHERE, HAVING)
func (c *Condition) arguments() []interface{} {
	var arguments []interface{}
	arguments = append(arguments, c.args[ON]...)
	arguments = append(arguments, c.args[WHERE]...)
	arguments = append(arguments, c.args[HAVING]...)
	return arguments
}

// conditionHelper for ON, WHERE and HAVING.
// The stmt will be set to the correct condition variable.
// Error will return if there is an argument/placeholder mismatch.
func (c *Condition) conditionHelper(conditionType int, stmt string, args []interface{}) {

	//no statement given
	if stmt == "" {
		return
	}

	sqlStmt := ""

	// compare placeholders and arguments length, return error if there is a mismatch
	if strings.Count(stmt, PLACEHOLDER) != len(args) {
		c.error = fmt.Errorf(ErrPlaceholderMismatch.Error(), stmt, strings.Count(stmt, PLACEHOLDER), len(args))
		return
	}

	stmt = stmtMapManipulation(c, stmt, args, conditionType)
	if c.error != nil {
		return
	}

	// building statement
	switch conditionType {
	case WHERE:
		if c.where == "" {
			sqlStmt += "WHERE"
		}
		if sqlStmt != "WHERE" {
			sqlStmt += c.where + " AND"
		}
		c.where = sqlStmt + " " + stmt
	case HAVING:
		if c.having == "" {
			sqlStmt += "HAVING"
		}
		if sqlStmt != "HAVING" {
			sqlStmt += c.having + " AND"
		}
		c.having = sqlStmt + " " + stmt
	case ON:
		c.on = "ON " + stmt
	}

}

// render the condition stmt in the right order.
// Error will return if the arguments does not fit to the placeholder.
func (c *Condition) render(p *Placeholder) (string, error) {

	if c.error != nil {
		return "", c.error
	}

	return replacePlaceholders(c.Config(false, WHERE, GROUP, HAVING, ORDER, LIMIT, OFFSET), p), nil
}

package sqlquery

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCondition_internal_Reset(t *testing.T) {
	c := Condition{}

	c.Where("id = ?", 1)
	assert.Equal(t, " WHERE id = ?", c.where)
	c.Reset(WHERE)
	assert.Equal(t, "", c.where)
	assert.Equal(t, []interface{}([]interface{}{}), c.args[WHERE])

	c.Group("id", "name")
	assert.Equal(t, " GROUP BY id, name", c.group)
	c.Reset(GROUP)
	assert.Equal(t, "", c.group)

	c.Having("id = ?", 1)
	assert.Equal(t, " HAVING id = ?", c.having)
	c.Reset(HAVING)
	assert.Equal(t, "", c.having)
	assert.Equal(t, []interface{}([]interface{}{}), c.args[HAVING])

	c.Order("id", "-name")
	assert.Equal(t, " ORDER BY id ASC, name DESC", c.order)
	c.Reset(ORDER)
	assert.Equal(t, "", c.order)

	c.Limit(5)
	assert.Equal(t, " LIMIT 5", c.limit)
	c.Reset(LIMIT)
	assert.Equal(t, "", c.limit)

	c.Offset(10)
	assert.Equal(t, " OFFSET 10", c.offset)
	c.Reset(OFFSET)
	assert.Equal(t, "", c.offset)

	c.On("robots.id = parts.robot_id AND robot_id != ?", 1)
	assert.Equal(t, " ON robots.id = parts.robot_id AND robot_id != ?", c.on)
	c.Reset(ON)
	assert.Equal(t, "", c.on)
	assert.Equal(t, []interface{}([]interface{}{}), c.args[ON])

	//ALL
	c = Condition{}
	c.Where("id = ?", 1).Group("name").Having("name = ?", "Cozmo").Order("id", "-name").Limit(10).Offset(5)
	c.Reset(WHERE, GROUP, HAVING, ORDER, LIMIT, OFFSET)
	assert.Equal(t, "", c.where)
	assert.Equal(t, "", c.group)
	assert.Equal(t, "", c.having)
	assert.Equal(t, "", c.order)
	assert.Equal(t, "", c.limit)
	assert.Equal(t, "", c.offset)
}

func TestCondition_internal_Where(t *testing.T) {
	//no stmt
	c := Condition{}
	c.Where("", 1)
	assert.Equal(t, "", c.where)
	assert.Equal(t, []interface{}([]interface{}(nil)), c.args[WHERE])

	//string, int
	c.Reset(WHERE)
	c.Where("id = ?", 2).Where("name = ?", "Wall-E")
	assert.Equal(t, " WHERE id = ? AND name = ?", c.where)
	assert.Equal(t, []interface{}([]interface{}{2, "Wall-E"}), c.args[WHERE])

	//map int
	c.Reset(WHERE)
	c.Where("id IN (?)", []int{1, 2, 3, 4}).Where("name = ?", "Wall-E")
	assert.Equal(t, " WHERE id IN (?, ?, ?, ?) AND name = ?", c.where)
	assert.Equal(t, []interface{}([]interface{}{int64(1), int64(2), int64(3), int64(4), "Wall-E"}), c.args[WHERE])

	//map string
	c.Reset(WHERE)
	c.Where("name IN (?)", []string{"Cozmo", "Wall-E"})
	assert.Equal(t, " WHERE name IN (?, ?)", c.where)
	assert.Equal(t, []interface{}([]interface{}{"Cozmo", "Wall-E"}), c.args[WHERE])

	//map not supported type
	c.Reset(WHERE)
	c.Where("price IN (?)", []float64{3.3, 2.0})
	assert.Equal(t, " WHERE price IN (?, ?)", c.where)
	assert.Equal(t, []interface{}([]interface{}{}), c.args[WHERE])
}

func TestCondition_internal_Group(t *testing.T) {
	c := Condition{}
	c.Group("id", "name")
	assert.Equal(t, " GROUP BY id, name", c.group)
}

func TestCondition_internal_Having(t *testing.T) {
	c := Condition{}
	c.Having("id = ?", 2).Having("name = ?", "Wall-E")
	assert.Equal(t, " HAVING id = ? AND name = ?", c.having)
	assert.Equal(t, []interface{}([]interface{}{2, "Wall-E"}), c.args[HAVING])
}

func TestCondition_internal_On(t *testing.T) {
	c := Condition{}
	c.On("robots.id = parts.robot_id AND robot.id != ?", 1)
	assert.Equal(t, " ON robots.id = parts.robot_id AND robot.id != ?", c.on)
	assert.Equal(t, []interface{}([]interface{}{1}), c.args[ON])
}

func TestCondition_internal_Order(t *testing.T) {
	c := Condition{}
	c.Order("id")
	assert.Equal(t, " ORDER BY id ASC", c.order)

	c.Reset(ORDER)
	c.Order("id asc")
	assert.Equal(t, " ORDER BY id ASC", c.order)

	c.Reset(ORDER)
	c.Order("-id")
	assert.Equal(t, " ORDER BY id DESC", c.order)

	c.Reset(ORDER)
	c.Order("id desc")
	assert.Equal(t, " ORDER BY id DESC", c.order)

	c.Reset(ORDER)
	c.Order("id", "-name")
	assert.Equal(t, " ORDER BY id ASC, name DESC", c.order)

	c.Reset(ORDER)
	c.Order("id asc", "name desc", "-brand", "owner")
	assert.Equal(t, " ORDER BY id ASC, name DESC, brand DESC, owner ASC", c.order)
}

func TestCondition_internal_Limit(t *testing.T) {
	c := Condition{}
	c.Limit(5)
	assert.Equal(t, " LIMIT 5", c.limit)
}

func TestCondition_internal_Offset(t *testing.T) {
	c := Condition{}
	c.Offset(10)
	assert.Equal(t, " OFFSET 10", c.offset)
}

// stmtMapManipulation
func TestCondition_internal_stmtMapManipulation(t *testing.T) {
	// single
	c := &Condition{}
	stmt := stmtMapManipulation(c, "WHERE id = ?", []interface{}{1}, WHERE)
	assert.Equal(t, "WHERE id = ?", stmt)
	assert.Equal(t, []interface{}{1}, c.args[WHERE])

	// map
	c = &Condition{}
	stmt = stmtMapManipulation(c, "WHERE id IN (?)", []interface{}{[]int{1, 2, 3, 4}}, WHERE)
	assert.Equal(t, "WHERE id IN (?, ?, ?, ?)", stmt)
	assert.Equal(t, []interface{}{int64(1), int64(2), int64(3), int64(4)}, c.args[WHERE])
}

// addArgument
func TestCondition_internal_addArgumentAndArguments(t *testing.T) {
	// single
	c := Condition{}
	c.args = make(map[int][]interface{}) // this is normally happening in stmtMapManipulation
	c.addArgument(WHERE, int(1))
	c.addArgument(HAVING, int(2))
	c.addArgument(ON, int(3))

	// multi
	c.addArgument(WHERE, []int{1, 2, 3, 4})
	c.addArgument(WHERE, []string{"Cozmo"})

	c.addArgument(HAVING, []int{5, 6, 7, 8})
	c.addArgument(HAVING, []string{"Wall-E"})

	c.addArgument(ON, []int{9, 10, 11, 12})
	c.addArgument(ON, []string{"Ubimator"})

	assert.Equal(t, []interface{}([]interface{}{int(1), int64(1), int64(2), int64(3), int64(4), "Cozmo"}), c.args[WHERE])
	assert.Equal(t, []interface{}([]interface{}{int(2), int64(5), int64(6), int64(7), int64(8), "Wall-E"}), c.args[HAVING])
	assert.Equal(t, []interface{}([]interface{}{int(3), int64(9), int64(10), int64(11), int64(12), "Ubimator"}), c.args[ON])

	// slice type not implemented
	c.addArgument(ON, []float32{3.3})
	assert.Error(t, c.error)

	// arguments()
	assert.Equal(t, []interface{}{int(3), int64(9), int64(10), int64(11), int64(12), "Ubimator", int(1), int64(1), int64(2), int64(3), int64(4), "Cozmo", int(2), int64(5), int64(6), int64(7), int64(8), "Wall-E"}, c.arguments())

	// calling render with an c.error
	b, err := HelperCreateBuilder()
	if assert.NoError(t, err) {
		stmt, errRender := c.render(b.Placeholder)
		assert.Error(t, errRender)
		assert.Equal(t, "", stmt)
	}
}

// render
func TestCondition_internal_render(t *testing.T) {
	c := Condition{}
	c.Where("id = ?", 1).Group("name").Having("name = ?", "Cozmo").Order("id", "-name").Limit(10).Offset(5)

	b, err := HelperCreateBuilder()
	if assert.NoError(t, err) {
		stmt, errRender := c.render(b.Placeholder)
		assert.NoError(t, errRender)

		if b.Placeholder.Numeric {
			assert.Equal(t, " WHERE id = "+b.Placeholder.Char+"1 GROUP BY name HAVING name = "+b.Placeholder.Char+"2 ORDER BY id ASC, name DESC LIMIT 10 OFFSET 5", stmt)
		} else {
			assert.Equal(t, " WHERE id = "+b.Placeholder.Char+" GROUP BY name HAVING name = "+b.Placeholder.Char+" ORDER BY id ASC, name DESC LIMIT 10 OFFSET 5", stmt)
		}
	}
}

// conditionHelper is already tested in WHERE,GROUP and ON
func TestCondition_internal_conditionHelper(t *testing.T) {
	c := Condition{}
	assert.Equal(t, nil, c.error)
	c.conditionHelper(WHERE, "id = ? AND name = ?", []interface{}{1})
	assert.Error(t, c.error)
}

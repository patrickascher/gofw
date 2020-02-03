package sqlquery_test

import (
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCondition_Where(t *testing.T) {
	c := sqlquery.Condition{}
	assert.IsType(t, &sqlquery.Condition{}, c.Where("id = ?", 1))
}

func TestCondition_Group(t *testing.T) {
	c := sqlquery.Condition{}
	assert.IsType(t, &sqlquery.Condition{}, c.Group("id", "name"))
}

func TestCondition_Having(t *testing.T) {
	c := sqlquery.Condition{}
	assert.IsType(t, &sqlquery.Condition{}, c.Having("id = ?", 1))
}

func TestCondition_Order(t *testing.T) {
	c := sqlquery.Condition{}
	assert.IsType(t, &sqlquery.Condition{}, c.Order("id", "name"))
}

func TestCondition_Limit(t *testing.T) {
	c := sqlquery.Condition{}
	assert.IsType(t, &sqlquery.Condition{}, c.Limit(5))
}

func TestCondition_Offset(t *testing.T) {
	c := sqlquery.Condition{}
	assert.IsType(t, &sqlquery.Condition{}, c.Offset(10))
}

func TestCondition_On(t *testing.T) {
	c := sqlquery.Condition{}
	assert.IsType(t, &sqlquery.Condition{}, c.On("table.one = table.two"))
}

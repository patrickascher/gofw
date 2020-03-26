package orm

import (
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestModel_structName(t *testing.T) {
	prt := &Customer{}
	assert.Equal(t, "Customer", structName(prt, false))
	assert.Equal(t, "orm.Customer", structName(prt, true))

	cust := Customer{}
	assert.Equal(t, "Customer", structName(cust, false))
	assert.Equal(t, "orm.Customer", structName(cust, true))
}

func TestModel_initBuilder(t *testing.T) {
	cust := Customer{}
	cust.caller = &cust

	// no builder defined or nil
	GlobalBuilder = nil
	_, err := cust.initBuilder()
	if assert.Error(t, err) {
		assert.Equal(t, ErrModelNoBuilder.Error(), err.Error())
	}

	// no builder not defined or nil
	GlobalBuilder = &sqlquery.Builder{}
	b, err := cust.initBuilder()
	if assert.NoError(t, err) {
		assert.Equal(t, GlobalBuilder, b)
	}

	// User defined Builder but nil
	cnb := CustomerNilBuilder{}
	cnb.caller = &cnb
	_, err = cnb.initBuilder()
	if assert.Error(t, err) {
		assert.Equal(t, ErrModelNoBuilder.Error(), err.Error())
	}

	// User defined Builder ok
	cb := CustomerBuilder{}
	cb.caller = &cb
	b, err = cb.initBuilder()
	if assert.NoError(t, err) {
		assert.Equal(t, &sqlquery.Builder{}, b)
	}
}

func TestModel_initTable(t *testing.T) {
	GlobalBuilder, _ = HelperCreateBuilder()
	cb := Customer{}
	cb.caller = &cb
	cb.table = &Table{Builder: GlobalBuilder}
	err := cb.initTable()

	if assert.NoError(t, err) {
		assert.Equal(t, []string{"orm.Customer"}, cb.loadedRel)
		assert.Equal(t, GlobalBuilder, cb.table.Builder)
		assert.Equal(t, "customers", cb.table.Name)
		assert.Equal(t, "tests", cb.table.Database)
		assert.Equal(t, &EagerLoading{}, cb.table.strategy)
		assert.Equal(t, Associations{}, cb.table.Associations)
	}

	// testing if no builder exists
	cnb := CustomerNilBuilder{}
	cnb.caller = &cnb
	err = cnb.initTable()
	assert.Error(t, err)

	// addStructFieldsToTableColumn are tested in field.go
	// table.describe are tested in db_table.go
}

func TestModel_isInit(t *testing.T) {
	cb := Customer{}

	// Model is init
	cb.isInitialized = true
	assert.True(t, cb.isInit())

	// model is not init
	cb.isInitialized = false
	assert.False(t, cb.isInit())
}

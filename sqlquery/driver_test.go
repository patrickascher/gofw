package sqlquery_test

import (
	"fmt"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/patrickascher/gofw/sqlquery/mysql"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestNewDriver(t *testing.T) {
	mem, err := sqlquery.NewDriver("mysql")
	assert.NoError(t, err)
	assert.Equal(t, "*mysql.Mysql", reflect.ValueOf(mem).Type().String())
}

func TestRegister(t *testing.T) {
	_, err := sqlquery.NewDriver("xy")
	assert.Equal(t, fmt.Errorf(sqlquery.ErrUnknownDriver.Error(), "xy"), err)

	//empty cache-backend
	err = sqlquery.Register("redis", nil)
	assert.Equal(t, sqlquery.ErrNoDriver, err)

	//empty cache-name
	err = sqlquery.Register("", &mysql.Mysql{})
	assert.Equal(t, sqlquery.ErrNoDriver, err)

	//already exists - tests also if the register worked
	err = sqlquery.Register("redis", &mysql.Mysql{})
	assert.NoError(t, err)
	err = sqlquery.Register("redis", &mysql.Mysql{})
	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf(sqlquery.ErrDriverAlreadyExists.Error(), "redis"), err)
}

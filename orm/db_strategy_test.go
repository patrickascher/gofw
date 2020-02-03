package orm_test

import (
	"testing"

	"fmt"
	"github.com/patrickascher/gofw/orm"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {

	err := orm.Register("mock2", &StrategyMock{})
	assert.NoError(t, err)

	err = orm.Register("", &StrategyMock{})
	assert.Error(t, err)
	assert.Equal(t, orm.ErrStrategyNotGiven.Error(), err.Error())

	err = orm.Register("mock3", nil)
	assert.Error(t, err)
	assert.Equal(t, orm.ErrStrategyNotGiven.Error(), err.Error())

	err = orm.Register("mock", &StrategyMock{})
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf(orm.ErrStrategyAlreadyExists.Error(), "mock"), err.Error())
}

func TestNewStrategy(t *testing.T) {
	s, err := orm.NewStrategy("mock")
	assert.NoError(t, err)
	assert.True(t, s != nil)

	s, err = orm.NewStrategy("mocks")
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf(orm.ErrStrategyUnknown.Error(), "mocks"), err.Error())

	assert.True(t, s == nil)

}

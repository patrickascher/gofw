package orm_test

import (
	"testing"

	"github.com/patrickascher/gofw/orm"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
)

// mock object that implements the strategy interface.
type strategyMock struct {
}

func (s *strategyMock) First(m *orm.Scope, c *sqlquery.Condition, p orm.Permission) error {
	return nil
}

func (s *strategyMock) All(res interface{}, m *orm.Scope, c *sqlquery.Condition) error {
	return nil
}

func (s *strategyMock) Create(m *orm.Scope) error {
	return nil
}

func (s *strategyMock) Update(m *orm.Scope, c *sqlquery.Condition) error {
	return nil
}

func (s *strategyMock) Delete(m *orm.Scope, c *sqlquery.Condition) error {
	return nil
}

func TestRegister(t *testing.T) {
	test := assert.New(t)

	// error: no provider-name and provider is given
	err := orm.Register("", nil)
	test.Error(err)
	test.Equal("orm: empty strategy-name or strategy-provider is nil", err.Error())

	// error: no provider is given
	err = orm.Register("mock", nil)
	test.Error(err)
	test.Equal("orm: empty strategy-name or strategy-provider is nil", err.Error())

	// error: no provider-name is given
	err = orm.Register("", &strategyMock{})
	test.Error(err)
	test.Equal("orm: empty strategy-name or strategy-provider is nil", err.Error())

	// ok: register successful
	err = orm.Register("mock", &strategyMock{})
	test.NoError(err)

	// error: multiple registration
	err = orm.Register("mock", &strategyMock{})
	test.Error(err)
	test.Equal("orm: strategy-provider \"mock\" is already registered", err.Error())
}

func TestNew(t *testing.T) {
	test := assert.New(t)

	// error: no registered dummy cache provider
	s, err := orm.NewStrategy("mock2")
	test.Nil(s)
	test.Error(err)
	test.Equal("orm: unknown strategy-provider \"mock2\"", err.Error())

	// ok
	s, err = orm.NewStrategy("mock")
	test.NotNil(s)
	test.NoError(err)
}

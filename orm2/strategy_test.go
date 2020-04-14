package orm2_test

import (
	"github.com/patrickascher/gofw/orm2"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
	"testing"
)

type StrategyMock struct {
}

func (s *StrategyMock) First(m orm2.Scope, c *sqlquery.Condition) error {
	return nil
}

func (s *StrategyMock) All(res interface{}, m orm2.Scope, c *sqlquery.Condition) error {
	return nil
}

func (s *StrategyMock) Create(m orm2.Scope) error {
	return nil
}

func (s *StrategyMock) Update(m orm2.Scope, c *sqlquery.Condition) error {
	return nil
}

func (s *StrategyMock) Delete(m orm2.Scope, c *sqlquery.Condition) error {
	return nil
}

func TestRegister(t *testing.T) {
	test := assert.New(t)

	// error: no provider-name and provider is given
	err := orm2.Register("", nil)
	test.Error(err)
	test.Equal("orm: empty strategy-name or strategy-provider is nil", err.Error())

	// error: no provider is given
	err = orm2.Register("mock", nil)
	test.Error(err)
	test.Equal("orm: empty strategy-name or strategy-provider is nil", err.Error())

	// error: no provider-name is given
	err = orm2.Register("", &StrategyMock{})
	test.Error(err)
	test.Equal("orm: empty strategy-name or strategy-provider is nil", err.Error())

	// ok: register successful
	err = orm2.Register("mock", &StrategyMock{})
	test.NoError(err)

	// error: multiple registration
	err = orm2.Register("mock", &StrategyMock{})
	test.Error(err)
	test.Equal("orm: strategy-provider \"mock\" is already registered", err.Error())
}

func TestNew(t *testing.T) {
	test := assert.New(t)

	// error: no registered dummy cache provider
	s, err := orm2.NewStrategy("mock2")
	test.Nil(s)
	test.Error(err)
	test.Equal("orm: unknown strategy-provider \"mock2\"", err.Error())

	// ok
	s, err = orm2.NewStrategy("mock")
	test.NotNil(s)
	test.NoError(err)
}

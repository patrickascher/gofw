package orm

import (
	"github.com/patrickascher/gofw/sqlquery"
)

func init() {
	_ = Register("custom", &StrategyCustom{})
}

// EagerLoading strategy
type StrategyCustom struct {
}

func (s StrategyCustom) First(m Interface, c *sqlquery.Condition) error {
	return nil
}

func (s StrategyCustom) All(res interface{}, m Interface, c *sqlquery.Condition) error {
	return nil
}

func (s StrategyCustom) Create(m Interface) error {
	return nil
}

func (s StrategyCustom) Update(m Interface, c *sqlquery.Condition) error {
	return nil
}

func (s StrategyCustom) Delete(m Interface, c *sqlquery.Condition) error {
	return nil
}

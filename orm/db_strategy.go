package orm

import (
	"fmt"
	"github.com/patrickascher/gofw/sqlquery"
)

var strategyStore = make(map[string]Strategy)

// Strategy interface
type Strategy interface {
	First(m Interface, c *sqlquery.Condition) error
	All(res interface{}, m Interface, c *sqlquery.Condition) error
	Create(m Interface) error
	Update(m Interface, c *sqlquery.Condition) error
	Delete(m Interface, c *sqlquery.Condition) error
}

// Register the strategy
// If the strategy name is empty or already exists a error will return
func Register(name string, strategy Strategy) error {
	if strategy == nil || name == "" {
		return ErrStrategyNotGiven
	}
	if _, exists := strategyStore[name]; exists {
		return fmt.Errorf(ErrStrategyAlreadyExists.Error(), name)
	}
	strategyStore[name] = strategy
	return nil
}

// NewStrategy returns the strategy
func NewStrategy(name string) (Strategy, error) {
	instance, ok := strategyStore[name]
	if !ok {
		return nil, fmt.Errorf(ErrStrategyUnknown.Error(), name)
	}
	return instance, nil
}

package orm2

import (
	"errors"
	"fmt"
	"github.com/patrickascher/gofw/sqlquery"
)

var registry = make(map[string]Strategy)

// Error messages.
var (
	errNoProvider            = errors.New("orm: empty strategy-name or strategy-provider is nil")
	errUnknownProvider       = errors.New("orm: unknown strategy-provider %q")
	errProviderAlreadyExists = errors.New("orm: strategy-provider %#v is already registered")
)

// Strategy interface
type Strategy interface {
	First(Scope, *sqlquery.Condition) error
	All(interface{}, Scope, *sqlquery.Condition) error
	Create(Scope) error
	Update(Scope, *sqlquery.Condition) error
	Delete(Scope, *sqlquery.Condition) error
}

// Register the strategy
// If the strategy name is empty or already exists a error will return
func Register(name string, strategy Strategy) error {
	if strategy == nil || name == "" {
		return errNoProvider
	}
	if _, exists := registry[name]; exists {
		return fmt.Errorf(errProviderAlreadyExists.Error(), name)
	}
	registry[name] = strategy
	return nil
}

// NewStrategy returns the strategy
func NewStrategy(name string) (Strategy, error) {
	instance, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf(errUnknownProvider.Error(), name)
	}
	return instance, nil
}

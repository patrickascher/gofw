package sqlquery

import (
	"errors"
	"fmt"
)

var driverStore = make(map[string]Driver)

// Error messages are defined here
var (
	ErrUnknownDriver       = errors.New("sqlquery: unknown driver %q (forgotten import?)")
	ErrNoDriver            = errors.New("sqlquery: empty driver-name or driver is given")
	ErrDriverAlreadyExists = errors.New("sqlquery: driver %#v already exists")
)

// Driver interface
type Driver interface {
	Describe(db string, table string, builder *Builder, cols []string) *Select
	ForeignKeys(db string, table string, builder *Builder) *Select
	ConvertColumnType(t string, column *Column) Type
}

// Register the driver
// If the driver name is empty or already exists a error will return
func Register(name string, driver Driver) error {
	if driver == nil || name == "" {
		return ErrNoDriver
	}
	if _, exists := driverStore[name]; exists {
		return fmt.Errorf(ErrDriverAlreadyExists.Error(), name)
	}
	driverStore[name] = driver
	return nil
}

// NewDriver returns the sqlquery driver
func NewDriver(name string) (Driver, error) {
	instance, ok := driverStore[name]
	if !ok {
		return nil, fmt.Errorf(ErrUnknownDriver.Error(), name)
	}
	return instance, nil
}

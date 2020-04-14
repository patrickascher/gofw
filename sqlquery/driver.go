// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlquery

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/patrickascher/gofw/sqlquery/types"
)

// Error messages.
var (
	ErrUnknownProvider       = errors.New("sqlquery/driver: unknown driver %q")
	ErrNoProvider            = errors.New("sqlquery/driver: empty driver-name or driver is nil")
	ErrProviderAlreadyExists = errors.New("sqlquery/driver: driver %#v is already registered")
)

// registry for all cache providers.
var registry = make(map[string]driver)

// driver is a function which returns the driver interface.
// The first argument is the config struct, followed by an open connection.
// If the opened connection is nil, a connection will get created by the driver.
// Like this the driver is getting initialized only when its called.
type driver func(config Config, db *sql.DB) (DriverI, error)

// Driver interface.
type DriverI interface {
	// Connection should return an open connection. The implementation depends on the driver.
	// As proposal, a connection should be unique to avoid overhead. Store it in on package level if its created.
	Connection() *sql.DB

	// The character which will be used for quoting columns.
	QuoteCharacterColumn() string

	// Describe the columns of the given table.
	Describe(b *Builder, db string, table string, cols []string) ([]Column, error)

	// ForeignKeys of the given table.
	ForeignKeys(b *Builder, db string, table string) ([]*ForeignKey, error)

	// Placeholder for the go driver.
	Placeholder() *Placeholder

	// Config returns the given configuration of the driver.
	Config() Config

	// TypeMapping should unify the different column types of different database types.
	TypeMapping(string, Column) types.Interface
}

// newDriver returns a sqlquery driver.
// Error will return if the diver is not registered.
func newDriver(cfg Config, connection *sql.DB) (DriverI, error) {
	instanceFn, ok := registry[cfg.Driver]
	if !ok {
		return nil, fmt.Errorf(ErrUnknownProvider.Error(), cfg.Driver)
	}

	return instanceFn(cfg, connection)
}

// Register the sqlquery drive. This should be called in the init() of the providers.
// If the sqlquery name/driver is empty or is already registered, an error will return.
func Register(name string, driver driver) error {
	if driver == nil || name == "" {
		return ErrNoProvider
	}
	if _, exists := registry[name]; exists {
		return fmt.Errorf(ErrProviderAlreadyExists.Error(), name)
	}
	registry[name] = driver
	return nil
}

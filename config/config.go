// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package config provides a config manager for any type that
// implements the config.Interface.
package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
)

const (
	// JSON pre-defined config provider.
	JSON = "json"
	// ENV is the default name to check the system environment variable os.GetEnv().
	ENV = "ENV"
)

// registry for all config providers.
var registry = make(map[string]provider)

// env for the provider.
var env string

// Error messages.
var (
	ErrNoProvider            = errors.New("config: empty config-name or config-provider is nil")
	ErrUnknownProvider       = errors.New("config: unknown config-provider %q")
	ErrProviderAlreadyExists = errors.New("config: config-provider %#v is already registered")
	ErrConfigPtr             = errors.New("config: struct must be a ptr")
)

// Interface is used by config providers.
type Interface interface {
	// Parse the given ptr struct.
	Parse(config interface{}, env string, options interface{}) error
}

// provider is a function which returns the config interface.
// Like this the config provider is getting initialized only when its called.
type provider func() Interface

// Register the config provider. This should be called in the init() of the providers.
// If the config provider/name is empty or is already registered, an error will return.
func Register(provider string, fn provider) error {
	if fn == nil || provider == "" {
		return ErrNoProvider
	}
	if _, exists := registry[provider]; exists {
		return fmt.Errorf(ErrProviderAlreadyExists.Error(), provider)
	}
	registry[provider] = fn
	return nil
}

// New will call the parse function on the given provider.
// The config must be a ptr to the given struct.
// Options and environment are passed through to the provider. For more information check the provider documentation.
// If no specific environment was set by SetEnv() before, the os.Env("ENV") will be used.
// If the provider is not registered, the parsing fails or the cfg kind is not ptr, an error will return.
func New(provider string, config interface{}, options interface{}) error {
	instanceFn, ok := registry[provider]
	if !ok {
		return fmt.Errorf(ErrUnknownProvider.Error(), provider)
	}

	if reflect.ValueOf(config).Kind() != reflect.Ptr {
		return ErrConfigPtr
	}

	instance := instanceFn()
	return instance.Parse(config, env, options)
}

// SetEnv allows a custom environment variable. This must be set before New() is called.
func SetEnv(e string) {
	env = e
}

// Env returns the actual environment variable.
func Env() string {
	if env == "" {
		return os.Getenv(ENV)
	}
	return env
}

/*
// IsSet checks if a specific field exists and has (no) zero value (recursively).
// By default its checking if the field has an none zero value, if a zero value should be allowed, prefix the field with a 0 (example: "0User.Role.Name")
// example: User.Role.Name searches if the User struct has a field Role and the struct Role has a field Name which has no zero value.
// At the moment only struct is implemented.
//
// Deprecated: This should not be used anymore. It will get rewritten in a struct module.
func IsSet(s string, cfg interface{}) bool {
	allowZero := ""
	if strings.HasPrefix(s, "0") {
		s = s[1:]
		allowZero = "0"
	}
	search := strings.Split(s, ".")

	t := reflect.TypeOf(cfg).Kind()
	switch t {
	case reflect.Struct:
		for i, v := range search {
			rv := reflect.Indirect(reflect.ValueOf(cfg))
			if !rv.FieldByName(v).IsValid() {
				return false
			}

			if len(search)-1 > i {
				return IsSet(allowZero+strings.Join(search[i+1:], "."), rv.FieldByName(v).Interface())
			}

			if allowZero == "0" {
				return true
			}
			t := reflect.TypeOf(rv.FieldByName(v).Interface())
			return rv.FieldByName(v).Interface() != reflect.Zero(t).Interface()
		}

	}

	return false
}
*/

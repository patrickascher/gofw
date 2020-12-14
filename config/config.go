// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package config provides a config manager for any type that
// implements the config.Interface.
package config

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator"
	"github.com/patrickascher/gofw/callback"
	"os"
	"reflect"
)

const (
	// JSON pre-defined config provider.
	JSON = "json"
	// ENV is the default name to check the system environment variable os.GetEnv().
	ENV = "ENV"
)

const (
	CallbackBeforeParse = "BeforeParse"
	CallbackAfterParse  = "AfterParse"
	CallbackBeforeValid = "BeforeValid"
	CallbackAfterValid  = "AfterValid"
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
// By default validate can be used on the struct to ensure all mandatory data is set.
// Callbacks BeforeParse, BeforeValid, AfterValid, AfterParse can be used.

func New(provider string, config interface{}, options interface{}) error {
	instanceFn, ok := registry[provider]
	if !ok {
		return fmt.Errorf(ErrUnknownProvider.Error(), provider)
	}

	if reflect.ValueOf(config).Kind() != reflect.Ptr {
		return ErrConfigPtr
	}

	err := callback.StructMethod(CallbackBeforeParse, config)
	if err != nil {
		return err
	}

	instance := instanceFn()
	err = instance.Parse(config, env, options)
	if err != nil {
		return err
	}

	err = callback.StructMethod(CallbackBeforeValid, config)
	if err != nil {
		return err
	}

	validate := validator.New()
	err = validate.Struct(config)
	if err != nil {
		return err
	}

	err = callback.StructMethod(CallbackAfterValid, config)
	if err != nil {
		return err
	}

	err = callback.StructMethod(CallbackAfterParse, config)
	if err != nil {
		return err
	}

	return nil
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

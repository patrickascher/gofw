// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package config_test

import (
	"fmt"
	"github.com/patrickascher/gofw/config/json"
	"os"
	"testing"

	"github.com/patrickascher/gofw/config"
	"github.com/stretchr/testify/assert"
)

var mockProvider = &mockConfig{}

type cfg struct {
	Called bool
}

type mockConfig struct {
	env     string
	options mockOptions
}

type mockOptions struct {
	Separator string
}

func (c *mockConfig) Parse(conf interface{}, env string, opt interface{}) error {
	conf.(*cfg).Called = true
	c.env = env
	c.options = opt.(mockOptions)
	return nil
}

func newMock() config.Interface {
	mockProvider = &mockConfig{}
	return mockProvider
}

func TestRegister(t *testing.T) {
	test := assert.New(t)

	// error: no provider-name and provider is given
	err := config.Register("", nil)
	test.Error(err)
	test.Equal(err.Error(), config.ErrNoProvider.Error())

	// error: no provider is given
	err = config.Register("mock", nil)
	test.Error(err)
	test.Equal(err.Error(), config.ErrNoProvider.Error())

	// error: no provider-name is given
	err = config.Register("", newMock)
	test.Error(err)
	test.Equal(err.Error(), config.ErrNoProvider.Error())

	// ok: register successful
	err = config.Register("mock", newMock)
	test.NoError(err)

	// error: multiple registration
	err = config.Register("mock", newMock)
	test.Error(err)
	test.Equal(fmt.Sprintf(config.ErrProviderAlreadyExists.Error(), "mock"), err.Error())
}

func TestNew(t *testing.T) {
	test := assert.New(t)

	// error: no registered dummy cache provider
	err := config.New("mock2", "", nil)
	test.Error(err)
	test.Equal(fmt.Sprintf(config.ErrUnknownProvider.Error(), "mock2"), err.Error())

	// error: config is no ptr
	err = config.New("mock", "", mockOptions{Separator: ";"})
	test.Error(err)
	test.Equal(config.ErrConfigPtr.Error(), err.Error())

	// ok
	c := &cfg{}
	err = config.New("mock", c, mockOptions{Separator: ";"})
	test.NoError(err)
	test.Equal(";", mockProvider.options.Separator)
	test.Equal(true, c.Called)
}

func TestSetEnv(t *testing.T) {
	config.SetEnv("development")
	assert.Equal(t, "development", config.Env())
}

func TestEnv(t *testing.T) {
	test := assert.New(t)

	// ok: used defined env should have priority
	config.SetEnv("development")
	err := os.Setenv(config.ENV, "system")
	test.NoError(err)
	test.Equal("development", config.Env())

	// ok: system is used
	config.SetEnv("")
	err = os.Setenv(config.ENV, "system")
	test.NoError(err)
	test.Equal("system", config.Env())

	// ok: none is set
	config.SetEnv("")
	err = os.Setenv(config.ENV, "")
	test.NoError(err)
	test.Equal("", config.Env())
}

// This example demonstrate the basics of the config interface.
// For more details check the documentation.
func Example() {
	// import the provider package
	// import _ "github.com/patrickascher/gofw/config/json"

	// Cfg is the configuration struct which should get marshaled.
	type Cfg struct {
		Database string
		Port     int
	}

	// define the config variable
	cfg := Cfg{}

	// SetEnv can be used to set a environment variable. By default the os.Env("ENV") will be used.
	config.SetEnv("staging")

	// New calls the given config provider and passes the struct, env and options.
	// cfg must be a pointer, otherwise the value can not be set.
	err := config.New(config.JSON, &cfg, json.Options{Filepath: "config/app.json"})
	if err != nil {
		return
	}

	// The value of the cfg variable will be set.
	// Cfg{Database:"127.0.0.1",Port:3306}
}

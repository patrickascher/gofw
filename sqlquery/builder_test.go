package sqlquery_test

import (
	"fmt"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
	"testing"
)

func init() {
	// ok: register successful
	_ = sqlquery.Register("test", mockMock)
}

func TestRegister(t *testing.T) {
	test := assert.New(t)

	// error: no provider-name and provider is given
	err := sqlquery.Register("", nil)
	test.Error(err)
	test.Equal(err.Error(), sqlquery.ErrNoProvider.Error())

	// error: no provider is given
	err = sqlquery.Register("mock", nil)
	test.Error(err)
	test.Equal(err.Error(), sqlquery.ErrNoProvider.Error())

	// error: no provider-name is given
	err = sqlquery.Register("", mockMock)
	test.Error(err)
	test.Equal(err.Error(), sqlquery.ErrNoProvider.Error())

	// ok: register successful
	err = sqlquery.Register("mock", mockMock)
	test.NoError(err)

	// error: multiple registration
	err = sqlquery.Register("mock", mockMock)
	test.Error(err)
	test.Equal(fmt.Sprintf(sqlquery.ErrProviderAlreadyExists.Error(), "mock"), err.Error())
}

func TestNew(t *testing.T) {
	test := assert.New(t)

	// error: no config driver -> unregistered empty driver
	b, err := sqlquery.New(sqlquery.Config{Driver: "mock2"}, nil)
	test.Equal(sqlquery.Builder{}, b)
	test.Error(err)
	test.Equal(fmt.Sprintf(sqlquery.ErrUnknownProvider.Error(), "mock2"), err.Error())

	// error: no registered dummy cache provider
	b, err = sqlquery.New(sqlquery.Config{Driver: "mock2"}, nil)
	test.Equal(sqlquery.Builder{}, b)
	test.Error(err)
	test.Equal(fmt.Sprintf(sqlquery.ErrUnknownProvider.Error(), "mock2"), err.Error())

	// ok
	b, err = sqlquery.New(sqlquery.Config{Driver: "mock"}, nil)
	test.NoError(err)
	test.NotNil(b)
}

func TestBuilder_Information(t *testing.T) {
	test := assert.New(t)

	b, err := sqlquery.New(sqlquery.Config{Driver: "mock", Database: "company"}, nil)
	test.NoError(err)

	_, err = b.Information("user").Describe("name", "surname")
	test.NoError(err)
	test.Equal("company", mockProvider.describeDb)
	test.Equal("user", mockProvider.describeTable)

	// ok - columns (name,surname) will be described in c.user
	_, err = b.Information("c.user").Describe("name", "surname")
	test.NoError(err)
	test.Equal("c", mockProvider.describeDb)
	test.Equal("user", mockProvider.describeTable)
}

func TestBuilder_Select(t *testing.T) {
	test := assert.New(t)

	b, err := sqlquery.New(sqlquery.Config{Driver: "mock", Database: "company"}, nil)
	test.NoError(err)

	sel := b.Select("user")

	test.NoError(err)
	test.IsType(&sqlquery.Select{}, sel)
}

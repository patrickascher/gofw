package config_test

import (
	"errors"
	"fmt"
	"github.com/patrickascher/gofw/config"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

type TestReader struct {
	env string
}

type TestReaderOptions struct {
	TriggerErr bool
}

func (p *TestReader) Parse(conf interface{}, options config.Options) error {
	o := options.(*TestReaderOptions)
	if o.TriggerErr == true {
		return errors.New("test error")
	}
	return nil
}

func (p *TestReader) Env(env string) {
	p.env = env
}

func (p *TestReader) reset() {
	p.env = ""
}

var reader = &TestReader{}

func TestRegister(t *testing.T) {

	// Register test reader
	err := config.Register("test", reader)
	assert.NoError(t, err)

	// Error empty reader
	err = config.Register("test", nil)
	assert.Error(t, err)
	assert.Equal(t, config.ErrNoReader, err)

	// Error empty reader name
	err = config.Register("", reader)
	assert.Error(t, err)
	assert.Equal(t, config.ErrNoReader, err)

	// Error registering it twice
	err = config.Register("test", reader)
	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf(config.ErrReaderAlreadyExists.Error(), "test"), err)
}

func TestExist(t *testing.T) {

	type Role struct {
		Name string
	}
	type User struct {
		Role Role
	}
	type Cfg struct {
		User User
	}

	cfg := Cfg{}

	reader.reset()
	err := os.Setenv(config.ENV, "devsys")
	assert.NoError(t, err)
	err = config.Parse("test", &cfg, &TestReaderOptions{})
	assert.NoError(t, err)

	// field exist but no value
	assert.False(t, config.IsSet("User", cfg))
	assert.False(t, config.IsSet("User.Role", cfg))
	assert.False(t, config.IsSet("User.Role.Name", cfg))
	assert.True(t, config.IsSet("0User", cfg))
	assert.True(t, config.IsSet("0User.Role", cfg))
	assert.True(t, config.IsSet("0User.Role.Name", cfg))
	// field does not exist
	assert.False(t, config.IsSet("Users", cfg))
	assert.False(t, config.IsSet("User.Roles", cfg))
	assert.False(t, config.IsSet("User.Roles.Name", cfg))
	assert.False(t, config.IsSet("User.Role.Names", cfg))

	// no struct type
	test := "test"
	assert.False(t, config.IsSet("t", test))

	cfg.User.Role.Name = "Wall-E"
	// field exist but no value
	assert.True(t, config.IsSet("User", cfg))
	assert.True(t, config.IsSet("User.Role", cfg))
	assert.True(t, config.IsSet("User.Role.Name", cfg))
}

func TestParse(t *testing.T) {
	type Cfg struct{}

	// Reader does not exist
	err := config.Parse("notExisting", &Cfg{}, &TestReaderOptions{})
	assert.Error(t, err)

	// Reader exist - sys environment
	reader.reset()
	err = os.Setenv(config.ENV, "devsys")
	assert.NoError(t, err)
	err = config.Parse("test", &Cfg{}, &TestReaderOptions{})
	assert.NoError(t, err)
	assert.Equal(t, "devsys", reader.env)

	// Reader exist - manual environment
	reader.reset()
	config.Env("dev")
	err = config.Parse("test", &Cfg{}, &TestReaderOptions{})
	assert.NoError(t, err)
	assert.Equal(t, "dev", reader.env)

	// Reader triggered error
	reader.reset()
	err = config.Parse("test", &Cfg{}, &TestReaderOptions{TriggerErr: true})
	assert.Error(t, err)
}

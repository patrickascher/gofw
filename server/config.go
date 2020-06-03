package server

import (
	"errors"
	"github.com/patrickascher/gofw/sqlquery"
	"reflect"
)

var cfg *Config

const DEFAULT = "default"

var (
	ErrConfigNotLoaded = errors.New("server: config is not loaded")
)

type Config struct {
	Databases    []*sqlquery.Config `json:"databases" validate:"min=1"`
	Server       Server             `json:"server" validate:"required"`
	Router       RouterProvider     `json:"router" validate:"required"`
	CacheManager []CacheProvider    `json:"caches" validate:"min=1"`
}

type Server struct {
	HTTPPort int    `json:"httpPort" validate:"required"`
	AppPath  string `json:"appPath" validate:"required"`
}

type RouterProvider struct {
	Provider    string      `json:"provider" validate:"required"`
	Favicon     string      `json:"favicon"`
	Directories []UrlSource `json:"directories"`
	Files       []UrlSource `json:"files"`
}

type UrlSource struct {
	Url    string `json:"url"`
	Source string `json:"source"`
}

type CacheProvider struct {
	Provider string `json:"provider" validate:"required"`
	GCCycle  int64  `json:"cycle" validate:"required"` // int * Minutes
}

// config returns the loaded configuration.
// If it was not loaded yet, a error will return.
func config() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}
	return nil, ErrConfigNotLoaded
}

func loadConfig(userConfig interface{}) *Config {
	rv := reflect.Indirect(reflect.ValueOf(userConfig))
	if rv.IsValid() {
		for i := 0; i < rv.NumField(); i++ {
			if rv.Field(i).Type().String() == "server.Config" {
				return rv.Field(i).Addr().Interface().(*Config)
			}
		}
	}
	return nil
}

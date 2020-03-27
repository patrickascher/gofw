package server

import (
	"database/sql"
	"errors"
	"github.com/patrickascher/gofw/sqlquery"
	"reflect"
)

var (
	cfg = &Cfg{}
	db  *sql.DB
)

var (
	ErrConfigNotLoaded = errors.New("server: config is not loaded")
)

type Cfg struct {
	Database     *sqlquery.Config `json:"database"`
	Server       Server           `json:"server"`
	Router       RouterProvider   `json:"router"`
	CacheManager CacheProvider    `json:"cache"`
}

type Server struct {
	HTTPPort int `json:"httpPort"`
}

type RouterProvider struct {
	Provider    string      `json:"provider"`
	Favicon     string      `json:"favicon"`
	Directories []UrlSource `json:"directories"`
	Files       []UrlSource `json:"files"`
}

type UrlSource struct {
	Url    string `json:"url"`
	Source string `json:"source"`
}

type CacheProvider struct {
	Provider string `json:"provider"`
	GCCycle  int64  `json:"cycle"`
}

// config returns the loaded configuration.
// If it was not loaded yet, a error will return.
func config() (*Cfg, error) {
	if cfg != nil {
		return cfg, nil
	}
	return &Cfg{}, ErrConfigNotLoaded
}

func loadConfig(userConfig interface{}) *Cfg {
	rv := reflect.Indirect(reflect.ValueOf(userConfig))
	if rv.IsValid() {
		for i := 0; i < rv.NumField(); i++ {
			if rv.Field(i).Type().Name() == "Cfg" {
				return rv.Field(i).Addr().Interface().(*Cfg)
			}
		}
	}
	return nil
}

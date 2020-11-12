package server

import (
	"errors"
	"github.com/patrickascher/gofw/cache/memory"
	"github.com/patrickascher/gofw/logger/console"
	"github.com/patrickascher/gofw/router/httprouter"
	"time"

	"github.com/patrickascher/gofw/cache"
	"github.com/patrickascher/gofw/logger"
	"github.com/patrickascher/gofw/router"
	"github.com/patrickascher/gofw/sqlquery"
)

var (
	cfgLogger  *logger.Logger
	cfgCache   []cache.Interface
	cfgBuilder []sqlquery.Builder
	cfgRouter  *router.Manager
)

// Logger returns the default log.
// Logger is always defined.
func Logger() *logger.Logger {
	return cfgLogger
}

// initLogger is setting a console log.
func initLogger() error {
	var err error

	c, err := console.New(console.Options{Color: true})
	if err != nil {
		return err
	}

	err = logger.Register("console", logger.Config{Writer: c})
	if err != nil {
		return err
	}

	cfgLogger, err = logger.Get("console")
	if err != nil {
		return err
	}

	return nil
}

// Builder returns the configured database.
// If no database is defined, the builder will be nil..
func Builder(name string) (sqlquery.Builder, error) {
	if name == DEFAULT {
		return cfgBuilder[0], nil
	}
	for _, b := range cfgBuilder {
		if (b.Config().Name != "" && b.Config().Name == name) ||
			(b.Config().Name == "" && b.Config().Driver == name) {
			return b, nil
		}
	}
	return sqlquery.Builder{}, errors.New("server: builder does not exist " + name)
}

// initBuilder initialize a builder of the defined database config.
func initBuilder() error {
	c, err := config()
	if err != nil {
		return err
	}

	for _, db := range c.Databases {
		if db.Host != "" {
			b, err := sqlquery.New(*db, nil)
			if err != nil {
				return err
			}
			if db.Debug {
				b.SetLogger(Logger())
			}
			cfgBuilder = append(cfgBuilder, b)
		}
	}

	return nil
}

// Cache returns the configured cache.
// If no cache is defined, this will be nil.
func Cache(name string) (cache.Interface, error) {
	if name == DEFAULT {
		return cfgCache[0], nil
	}

	// TODO set a name/config to the cache. otherwise i can not iterate over it.
	return nil, nil
}

// initCache initialize the cache provider if set in the config.
func initCache() error {
	c, err := config()
	if err != nil {
		return err
	}

	for _, ca := range c.CacheManager {
		if ca.Provider != "" {
			c, err := cache.New(ca.Provider, memory.Options{GCInterval: time.Duration(ca.GCCycle) * time.Minute})
			if err != nil {
				return err
			}
			cfgCache = append(cfgCache, c)
		}
	}

	return nil
}

// Cache returns the configured cache.
// If no cache is defined, this will be nil.
func Router() *router.Manager {
	return cfgRouter
}

// initRouter init the router manager if the config is set
func initRouter() error {
	c, err := config()
	if err != nil {
		return err
	}
	if c.Router.Provider != "" {
		rm, err := router.New(c.Router.Provider, httprouter.Options{CatchAllKeyValuePair: true})
		if err != nil {
			return err
		}

		err = rm.SetFavicon(c.Router.Favicon)
		if err != nil {
			return err
		}

		for _, dir := range c.Router.Directories {
			err = rm.AddPublicDir(dir.Url, dir.Source)
			if err != nil {
				return err
			}
		}

		for _, file := range c.Router.Files {
			err = rm.AddPublicFile(file.Url, file.Source)
			if err != nil {
				return err
			}
		}
		cfgRouter = rm
	}

	return nil
}

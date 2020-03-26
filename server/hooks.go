package server

import (
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
	cfgCache   cache.Interface
	cfgBuilder sqlquery.Builder
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
// If no database is defined, the builder will be nil.
func Builder() *sqlquery.Builder {
	return &cfgBuilder
}

// initBuilder initialize a builder of the defined database config.
func initBuilder() error {
	c, err := config()
	if err != nil {
		return err
	}

	if c.Database.Host != "" {
		cfgBuilder, err = sqlquery.New(*c.Database, nil)
		if err != nil {
			return err
		}
		if c.Database.Debug {
			cfgBuilder.SetLogger(Logger())
		}
	}
	return nil
}

// Cache returns the configured cache.
// If no cache is defined, this will be nil.
func Cache() cache.Interface {
	return cfgCache
}

// initCache initialize the cache provider if set in the config.
func initCache() error {
	c, err := config()
	if err != nil {
		return err
	}

	if c.CacheManager.Provider != "" {

		c, err := cache.New(c.CacheManager.Provider, memory.Options{GCInterval: time.Duration(c.CacheManager.GCCycle) * time.Minute})
		if err != nil {
			return err
		}

		cfgCache = c
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
		cfgRouter = rm
	}

	return nil
}

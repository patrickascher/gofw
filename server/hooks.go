package server

import (
	"time"

	"github.com/patrickascher/gofw/cache"
	"github.com/patrickascher/gofw/logger"
	"github.com/patrickascher/gofw/router"
	"github.com/patrickascher/gofw/sqlquery"
)

var (
	cfgLogger  *logger.Logger
	cfgCache   cache.Cache
	cfgBuilder *sqlquery_.Builder
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
	cfgLogger, err = logger.Get(logger.CONSOLE)
	if err != nil {
		return err
	}
	return nil
}

// Builder returns the configured database.
// If no database is defined, the builder will be nil.
func Builder() *sqlquery_.Builder {
	return cfgBuilder
}

// initBuilder initialize a builder of the defined database config.
func initBuilder() error {
	c, err := config()
	if err != nil {
		return err
	}

	if c.Database.Host != "" {
		cfgBuilder, err = sqlquery_.NewBuilderFromConfig(c.Database)
		if err != nil {
			return err
		}
	}
	return nil
}

// Cache returns the configured cache.
// If no cache is defined, this will be nil.
func Cache() cache.Cache {
	return cfgCache
}

// initCache initialize the cache provider if set in the config.
func initCache() error {
	c, err := config()
	if err != nil {
		return err
	}

	if c.CacheManager.Provider != "" {
		c, err := cache.Get(c.CacheManager.Provider, time.Duration(int64(time.Second)*c.CacheManager.GCCycle))
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

		rm, err := router.Get(c.Router.Provider)
		if err != nil {
			return err
		}

		err = rm.Favicon(c.Router.Favicon)
		if err != nil {
			return err
		}

		for _, dir := range c.Router.Directories {
			rm.PublicDir(dir.Url, dir.Source)
		}
		cfgRouter = rm
	}

	return nil
}

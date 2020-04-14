// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package orm transfers a struct into an ORM by simple embedding the orm.Model.
//
// A model requires one or more primary keys which can be set by tag. If tag is set, the field ID will be defined as primary key.
// An error will return if no key is set.
//
// Configuration
// If the model has the function Default(TableName, DatabaseName, SchemaName, Builder, Logger, Cache) some default values can be set.
// By default, the TableName is the plural of the struct name. Database and SchemaName are taken from the builder configuration.
//
// Tags
// custom: if set, the field is declared as custom field. This means the field is not required in the database table.
// primary: set a field as primary field. if none is set, it checks if the field ID exists and sets this as default primary.
// column: set a table column name. by default the column name is snake style of the field name.
// permission: rw can be set for read and write. if none is required just type permission. The read and write privileges will be set to false.
// select: if a custom sql statement is required.
// relation: hasOne, belongsTo, hasMany, m2m
// join_table, join_fk, join_akf
// polymorphic:
// polymorphic_value:
//
// restrictions:
// self reference is only allowed on m2m
// polymorphic is only allowed on hasOne, hasMany
// embedded fields must be Exported structs and no orm2.Model in it is allowed. (TODO fix this, maybe useful?)
package orm2

import (
	"errors"
	"fmt"
	"github.com/guregu/null"
	"github.com/patrickascher/gofw/cache"
	"github.com/patrickascher/gofw/cache/memory"
	"github.com/patrickascher/gofw/logger"
	"github.com/patrickascher/gofw/logger/console"
	"github.com/patrickascher/gofw/sqlquery"
	_ "github.com/patrickascher/gofw/sqlquery/driver"

	"log"
	"reflect"
	"strings"
	"time"
)

const (
	CREATE = "create"
	UPDATE = "update"
	DELETE = "delete"
	FIRST  = "first"
	ALL    = "all"
)

var (
	GlobalBuilder sqlquery.Builder
	GlobalCache   cache.Interface
	GlobalLogger  *logger.Logger
)

var (
	errNoCache    = errors.New("orm: no cache or cache duration is defined")
	errInitPtr    = errors.New("orm: model must be a ptr")
	errBuilder    = errors.New("orm: no builder is defined")
	errDb         = errors.New("orm: db or table name is not defined")
	errBeforeInit = errors.New("orm: %s must be called before the Init method")
	errInit       = errors.New("orm: %s must be initialized before")
)

func init() {
	c := sqlquery.Config{}
	c.Driver = "mysql"
	c.Database = "orm_test"
	c.Schema = "public"
	c.Username = "root"
	c.Password = "root"
	c.Host = "127.0.0.1"
	c.Port = 3319
	c.Debug = true

	var err error
	GlobalBuilder, err = sqlquery.New(c, nil)
	if err != nil {
		log.Fatal(err)
	}

	GlobalCache, err = cache.New("memory", memory.Options{GCInterval: 1 * time.Minute})
	if err != nil {
		log.Fatal(err)
	}

	cLogger, err := console.New(console.Options{Color: true})
	err = logger.Register("model", logger.Config{Writer: cLogger})
	if err != nil {
		log.Fatal(err)
	}
	GlobalLogger, err = logger.Get("model")
	if err != nil {
		log.Fatal(err)
	}

	GlobalBuilder.SetLogger(GlobalLogger)
}

type Interface interface {
	Init(c Interface) error

	Cache() (cache.Interface, time.Duration, error)
	SetCache(p cache.Interface, duration time.Duration) error

	WBList() (policy int, fields []string)
	SetWBList(policy int, fields ...string)

	// Default values
	DefaultLogger() *logger.Logger
	DefaultCache() (manager cache.Interface, ttl time.Duration, error error)
	DefaultBuilder() sqlquery.Builder
	DefaultTableName() string
	DefaultDatabaseName() string
	DefaultSchemaName() string
	DefaultStrategy() string

	// needed for relations
	associationForeignKey(tag string, rel Interface) (Field, error)
	foreignKey(tag string) (Field, error)
	modelName(ns bool) string
	model() *Model
	setCaller(p Interface)

	// initRelation returns a model interface if the relation was already initialized.
	// The name must be the full model name. If none was found, nil will return.
	initRelation(string) Interface
	// addInitRelation adds a relation as initialized. The string must be the full model name incl. namespace.
	addInitRelation(string, Interface)
	// setInitRelations sets the initRel variable. This is used to pass all already initialized models to a child object.
	setInitRelations(map[string]Interface)

	First(c *sqlquery.Condition) error
}

func (m *Model) model() *Model {
	return m
}

func (m *Model) SetWBList(p int, fields ...string) {
	m.wbList = NewWBList(p, fields)
}

func (m *Model) WBList() (p int, fields []string) {
	if m.wbList == nil {
		return WHITELIST, nil
	}
	return m.wbList.policy, m.wbList.fields
}

type Model struct {

	// caller
	caller    Interface
	name      string
	fields    []Field
	relations []Relation

	initRel map[string]Interface

	wbList *whiteBlackList
	// Builder
	//builder    sqlquery.Builder
	//dbName     string
	//schemaName string
	//tableName  string

	// cache
	cache    cache.Interface
	cacheTTL time.Duration

	// strategy
	strategyVal   string
	isInitialized bool

	// Embedded time fields
	CreatedAt null.Time
	UpdatedAt null.Time
	DeletedAt null.Time
}

func (m *Model) setCaller(c Interface) {
	m.caller = c
}

// Init the orm model.
// This method must be called before the orm functions can be used.
// All mandatory configs will be checked (Cache, Builder, DB, Table name).
// The struct gets initialized and all relations and fields gets created.
// The database is checked against the fields and relations.
// After the init the orm model gets cached for performance reasons.
func (m *Model) Init(c Interface) error {

	// checks if the given argument is a ptr and not nil
	val := reflect.ValueOf(c)
	if !val.IsValid() || val.Kind() != reflect.Ptr {
		return errInitPtr
	}

	// set th caller
	m.caller = c

	// if no cache was set, call Cache.
	if m.cache == nil {
		_, _, err := c.Cache()
		if err != nil {
			return err
		}
	}

	// set the model name incl namespace
	m.name = val.Type().Elem().String()

	// check if cache exists
	if m.cache.Exist(m.name) {
		m.DefaultLogger().Debug("Init cached", m.modelName(false)) //TODO can be removed
		v, err := m.cache.Get(m.name)
		if err != nil {
			return err
		}
		*m = v.Value().(Model)
		m.caller = c

		cFields := make([]Field, len(m.fields))
		copy(cFields, m.fields)
		m.fields = cFields

		cRelations := make([]Relation, len(m.relations))
		copy(cRelations, m.relations)
		m.relations = cRelations
		return nil
	}

	m.DefaultLogger().Trace("Init", m.modelName(false)) //TODO can be removed

	// TODO set builder like cache? user defined? own tx?
	b := c.DefaultBuilder()
	if b.Driver() == nil {
		return errBuilder
	}
	if c.DefaultDatabaseName() == "" || c.DefaultTableName() == "" {
		return errDb
	}

	// build all exported struct fields
	err := m.createFields()
	if err != nil {
		return err
	}

	// mark this orm model as initialized, to avoid relation loops.
	m.addInitRelation(m.name, m.caller)

	// build all exported relations.
	err = m.createRelations()
	if err != nil {
		return err
	}

	// set strategy

	// todo validation
	// todo callbacks

	// set model as value
	m.isInitialized = true
	err = m.cache.Set(m.name, *m, m.cacheTTL)
	if err != nil {
		return err
	}

	cFields := make([]Field, len(m.fields))
	copy(cFields, m.fields)
	m.fields = cFields

	cRelations := make([]Relation, len(m.relations))
	copy(cRelations, m.relations)
	m.relations = cRelations
	return nil
}

// Cache returns the given cache. If none was defined yet the model function
// DefaultCache is called. If no cache provider was defined, an error will return.
// TODO panic if caller is nil, fix it! (err not init?)
func (m *Model) Cache() (cache.Interface, time.Duration, error) {
	var err error

	// If no cache was defined, call the DefaultCache.
	if m.cache == nil {
		m.cache, m.cacheTTL, err = m.caller.DefaultCache()
		if err != nil {
			return nil, 0, err
		}
	}

	// If no cache is set, an error will return.
	// A cache is mandatory for the orm, because of performance.
	if m.cache == nil || m.cacheTTL == 0 {
		return nil, 0, errNoCache
	}

	return m.cache, m.cacheTTL, nil
}

// SetCache sets a custom cache to the orm model.
// The method must be called before the orm model is initialized.
// Error will return if the cache provider is nil, no time duration is set or the model was already initialized.
func (m *Model) SetCache(c cache.Interface, ttl time.Duration) error {
	if m.isInitialized {
		return fmt.Errorf(errBeforeInit.Error(), "SetCache")
	}
	if c == nil || ttl == 0 {
		return errNoCache
	}
	m.cache = c
	m.cacheTTL = ttl
	return nil
}

// addInitRelation adds a new initialized relation.
// The string name must be the orm model name incl. namespace.
func (m *Model) addInitRelation(s string, r Interface) {
	if m.initRel == nil {
		m.initRel = map[string]Interface{}
	}
	m.initRel[s] = r
}

// setInitRelations sets the relations of the model.
// This is used to pass the initialized relations of a model.
func (m *Model) setInitRelations(s map[string]Interface) {
	m.initRel = s
}

// initRelation returns a orm Interface by the orm model name.
// Nil will return if the relation was not added yet.
func (m *Model) initRelation(s string) Interface {
	if val, ok := m.initRel[s]; ok {
		return val
	}
	return nil
}

// modelName returns the struct name with or without the namespace.
// model name will always be titled (first letter uppercase) also if the struct is unexported.
func (m Model) modelName(ns bool) string {
	name := m.name

	if idx := strings.Index(name, "."); !ns && idx != -1 {
		return strings.Title(name[idx+1:])
	}
	return name
}

func (m Model) strategy() (Strategy, error) {

	if m.strategyVal == "" {
		m.strategyVal = m.caller.DefaultStrategy()
	}

	s, err := NewStrategy(m.strategyVal)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// First will check the first founded row by its condition and adds it values to the struct fields.
// Everything handled in the loading strategy.
// It will return an error if the model is not initialized or the strategy returns an error.
func (m *Model) First(c *sqlquery.Condition) error {
	if !m.isInitialized {
		return errInit
	}

	// TODO Callbacks before

	// create sql condition
	if c == nil {
		c = &sqlquery.Condition{}
	}

	// TODO loop maps (role x -> role y -> role x)

	s, err := m.strategy()
	if err != nil {
		return err
	}

	scope, err := NewScope(m, FIRST)
	if err != nil {
		return err
	}

	err = s.First(scope, c)
	if err != nil {
		return err
	}

	// TODO Callbacks after

	return err
}

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
// fk,afk
// join_table, join_fk, join_akf
// polymorphic:
// polymorphic_value:
//
// restrictions:
// self reference is only allowed on m2m
// polymorphic is only allowed on hasOne, hasMany
// embedded fields must be Exported structs and no orm2.Model in it is allowed. (TODO fix this, maybe useful?)
package orm

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	valid "github.com/go-playground/validator"
	"github.com/patrickascher/gofw/cache"
	"github.com/patrickascher/gofw/logger"
	"github.com/patrickascher/gofw/sqlquery"
	_ "github.com/patrickascher/gofw/sqlquery/driver"
)

const (
	CREATE = "create"
	UPDATE = "update"
	DELETE = "delete"
	FIRST  = "first"
	ALL    = "all"
)

const (
	CreatedAt = "CreatedAt"
	UpdatedAt = "UpdatedAt"
	DeletedAt = "DeletedAt"
)

var (
	validate *valid.Validate
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
	errInit       = errors.New("orm: must be initialized before")
	errResultPtr  = errors.New("orm: result variable must be a ptr in %s.All()")

	ErrUpdateNoChanges = errors.New("orm: the model %s was not updated because there were no changes")
)

func init() {

	// global validator
	validate = valid.New()
	validate.SetTagName(tagValidate)
	validate.RegisterCustomTypeFunc(ValidateValuer, NullInt{}, NullFloat{}, NullString{}, NullTime{})
	/*
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
	*/

}

type Interface interface {
	Init(c Interface) error

	Cache() (cache.Interface, time.Duration, error)
	SetCache(p cache.Interface, duration time.Duration) error

	WBList() (policy int, fields []string)
	SetWBList(policy int, fields ...string)

	// Default values
	// TODO move to scope?
	DefaultLogger() *logger.Logger
	DefaultCache() (manager cache.Interface, ttl time.Duration, error error)
	DefaultBuilder() sqlquery.Builder
	DefaultTableName() string
	DefaultDatabaseName() string
	DefaultSchemaName() string
	DefaultStrategy() string

	Scope() *Scope
	model() *Model

	First(c *sqlquery.Condition) error
	All(result interface{}, c *sqlquery.Condition) error
	Create() error
	Update() error
	Delete() error
	Count(c *sqlquery.Condition) (int, error)

	// experimental
	SetRelationCondition(string, sqlquery.Condition)
}

type Model struct {

	// name of the model incl. namespace.
	name string
	// identifier if the model was already initialized.
	isInitialized bool
	// the caller orm.
	caller Interface
	// the orm struct fields.
	fields []Field
	// the orm relation fields.
	relations []Relation
	// identifier if a sql transaction was added by the system.
	autoTx bool
	// parent orm
	parentModel *Model
	// changedValues, needed for update, that only changed values are updated
	changedValues []ChangedValue
	// white/black list
	wbList *whiteBlackList
	// identifier for a loop
	loopDetection map[string][]string
	// orm builder.
	builder sqlquery.Builder
	// orm scope.
	scope *Scope
	// cache
	cache cache.Interface
	// cache ttl
	cacheTTL time.Duration
	// strategy
	strategyVal string

	//experimental
	relationCondition map[string]sqlquery.Condition

	// Embedded time fields
	CreatedAt *NullTime `orm:"permission:w" json:",omitempty"`
	UpdatedAt *NullTime `orm:"permission:w" json:",omitempty"`
	DeletedAt *NullTime `orm:"permission:w" json:",omitempty"`
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
		//m.DefaultLogger().Debug("Init cached", m.modelName(false)) //TODO can be removed
		v, err := m.cache.Get(m.name)
		if err != nil {
			return err
		}
		*m = v.Value().(Model)

		m.caller = c
		m.parentModel = nil
		m.scope = &Scope{m}

		m.copyFieldRelationSlices()

		return nil
	}

	//m.DefaultLogger().Trace("Init", m.modelName(false)) //TODO can be removed

	// set scope
	m.scope = &Scope{m}

	// set builder
	b := c.DefaultBuilder()
	if b.Driver() == nil {
		return errBuilder
	}
	m.builder = b

	// check if database name and table name is defined.
	if c.DefaultDatabaseName() == "" || c.DefaultTableName() == "" {
		return errDb
	}

	// build all exported struct fields
	err := m.createFields()
	if err != nil {
		return err
	}

	// build all exported relations.
	err = m.createRelations()
	if err != nil {
		return err
	}

	// must be called here because the relations are required
	err = m.addDBValidation()
	if err != nil {
		return err
	}

	// todo set strategy
	// todo callbacks

	// set model as value
	m.isInitialized = true
	err = m.cache.Set(m.name, *m, m.cacheTTL)
	if err != nil {
		return err
	}

	m.copyFieldRelationSlices()

	return nil
}

func (m *Model) RelationCondition(relation string) *sqlquery.Condition {
	if v, ok := m.relationCondition[relation]; ok {
		tmp := v
		return &tmp
	}
	return nil
}

func (m *Model) SetRelationCondition(relation string, condition sqlquery.Condition) {
	if m.relationCondition == nil {
		m.relationCondition = make(map[string]sqlquery.Condition, 1)
	}
	m.relationCondition[relation] = condition
}

// copyFieldRelationSlices is needed that the cached fields and relations of the orm model are not getting changed.
func (m *Model) copyFieldRelationSlices() {
	cFields := make([]Field, len(m.fields))
	copy(cFields, m.fields)
	m.fields = cFields

	cRelations := make([]Relation, len(m.relations))
	copy(cRelations, m.relations)
	m.relations = cRelations
}

// Scope of the model.
func (m *Model) Scope() *Scope {
	return m.scope
}

// SetWBList a white/blacklist to the orm.
func (m *Model) SetWBList(p int, fields ...string) {
	m.wbList = newWBList(p, fields)
}

// WBList of the orm.
func (m *Model) WBList() (p int, fields []string) {
	if m.wbList == nil {
		return WHITELIST, nil
	}
	return m.wbList.policy, m.wbList.fields
}

// Cache returns the given cache. If none was defined yet the model function
// DefaultCache is called. If no cache provider was defined, an error will return.
func (m *Model) Cache() (cache.Interface, time.Duration, error) {
	var err error

	// If no cache was defined, call the DefaultCache.
	if m.cache == nil {
		if m.caller == nil {
			return nil, 0, errInit
		}
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
	//if c == nil || ttl == 0 { // ttl Zero is infinity.
	if c == nil {
		return errNoCache
	}
	m.cache = c
	m.cacheTTL = ttl
	return nil
}

// First will check the first founded row by its condition and adds it values to the struct fields.
// Everything handled in the loading strategy.
// It will return an error if the model is not initialized or the strategy returns an error.
func (m *Model) First(c *sqlquery.Condition) error {
	if !m.isInitialized {
		return fmt.Errorf(errInit.Error(), reflect.TypeOf(m.caller))
	}

	// reset loop detection TODO in every mode (ALL,CREATE,UPDATE,DELETE)
	if m.parentModel == nil {
		m.loopDetection = nil
	}

	// TODO Callbacks before

	// create sql condition
	if c == nil {
		c = &sqlquery.Condition{}
	}

	s, err := m.strategy()
	if err != nil {
		return err
	}

	err = m.scope.setFieldPermission(FIRST)
	if err != nil {
		return err
	}

	err = m.scope.checkLoopMap(c.Config(true, sqlquery.WHERE))
	if err != nil {
		return err
	}

	err = s.First(m.scope, c, Permission{Read: true})
	if err != nil {
		return err
	}
	if m.parentModel == nil {
		m.loopDetection = nil
	}

	// TODO Callbacks after

	return nil
}

// All will return all rows by its condition and puts it in the given result.
// Everything handled in the loading strategy.
// It will return an error if the model is not initialized or the strategy returns an error.
func (m *Model) All(result interface{}, c *sqlquery.Condition) error {
	if !m.isInitialized {
		return fmt.Errorf(errInit.Error(), "all", reflect.TypeOf(m.caller))
	}

	// checking if the res is a ptr
	if result == nil || reflect.TypeOf(result).Kind() != reflect.Ptr {
		return fmt.Errorf(errResultPtr.Error(), m.name)
	}

	// TODO Callbacks before
	//m.resSet = result <- needed for callbacks

	if c == nil {
		c = &sqlquery.Condition{}
	}

	s, err := m.strategy()
	if err != nil {
		return err
	}

	err = m.scope.setFieldPermission(ALL)
	if err != nil {
		return err
	}

	err = m.scope.checkLoopMap(c.Config(true, sqlquery.WHERE))
	if err != nil {
		return err
	}

	now := time.Now()
	err = s.All(result, m.scope, c)
	if err != nil {
		return err
	}
	fmt.Println(time.Since(now))
	// TODO Callbacks after
	if m.parentModel == nil {
		m.loopDetection = nil
	}
	return nil
}

// Create an entry with the actual struct value.
// Everything handled in the loading strategy.
// If there is no custom transaction added, it will add one by default and also commits it automatically if everything is ok. Otherwise a Rollback will be called.
// It will return an error if the model is not initialized, tx  error, the strategy returns an error or a commit error happens.
func (m *Model) Create() (err error) {
	defer func() { modelDefer(m, err) }()

	if !m.isInitialized {
		err = fmt.Errorf(errInit.Error(), "Create", reflect.TypeOf(m.caller))
		return
	}

	// TODO callback before

	// set the CreatedAt info if exists
	// it only gets saved if the field exists in the db (permission is set)
	m.CreatedAt = &NullTime{NullTime: sql.NullTime{time.Now(), true}}

	// if the model is empty no need for creating.
	if m.scope.IsEmpty(Permission{Write: true}) {
		return nil
	}

	err = m.scope.setFieldPermission(CREATE)
	if err != nil {
		return
	}

	err = m.isValid()
	if err != nil {
		return
	}

	s, err := m.strategy()
	if err != nil {
		return
	}

	// call delete on strategy
	err = m.addAutoTX()
	if err != nil {
		return
	}

	now := time.Now()
	err = s.Create(m.scope)
	if err != nil {
		return
	}
	fmt.Println(time.Since(now))

	// call delete on strategy
	err = m.commitAutoTX()
	if err != nil {
		return
	}

	// TODO callback after
	return
}

// modelDefer function for create, update and delete.
// it checks if a a tx was added and rolls it back.
func modelDefer(m *Model, err error) {
	if err != nil && m.isInitialized && m.autoTx && m.Scope().Builder().HasTx() {
		rErr := m.Scope().Builder().Rollback()
		if rErr != nil {
			panic(rErr)
		}
	}
}

// Update an entry with the actual struct value.
// Everything handled in the loading strategy.
// If there is no custom transaction added, it will add one by default and also commits it automatically if everything is ok. Otherwise a Rollback will be called.
// It will return an error if the model is not initialized, tx  error, the strategy returns an error or a commit error happens.
func (m *Model) Update() (err error) {
	defer func() { modelDefer(m, err) }()

	if !m.isInitialized {
		err = fmt.Errorf(errInit.Error(), "Update", reflect.TypeOf(m.caller))
		return
	}

	if !m.scope.PrimariesSet() {
		err = fmt.Errorf(errPrimaryKey, m.name)
		return
	}

	// must be called before isValid
	err = m.scope.setFieldPermission(UPDATE)
	if err != nil {
		return
	}

	err = m.isValid()
	if err != nil {
		return
	}

	// create where condition
	c := &sqlquery.Condition{}
	for _, col := range m.scope.PrimaryKeys() {
		c.Where(m.scope.Builder().QuoteIdentifier(col.Information.Name)+" = ?", m.scope.CallerField(col.Name).Interface())
	}

	s, err := m.strategy()
	if err != nil {
		return
	}

	// TODO callback before

	// call delete on strategy
	err = m.addAutoTX()
	if err != nil {
		return
	}

	// snapshot
	// TODO option to disable snapshot
	if m.parentModel == nil {

		// reset condition loop
		m.loopDetection = nil

		// init snapshot
		snapshot := newValueInstanceFromType(reflect.TypeOf(m.caller)).Addr().Interface().(Interface)
		err = m.scope.InitRelation(snapshot, "")
		if err != nil {
			return
		}

		err = s.First(snapshot.Scope(), c, Permission{Write: true})
		if err != nil {
			return
		}

		m.changedValues, err = m.scope.EqualWith(snapshot)
		if err != nil {
			return
		}
	}

	// if no data was changed
	if m.changedValues == nil {
		err = fmt.Errorf(ErrUpdateNoChanges.Error(), m.name)
		return
	}

	fmt.Println("changed ------>", m.changedValues)
	// set the UpdatedAt info if exists
	// it only gets saved if the field exists in the db (permission is set)
	m.UpdatedAt = &NullTime{NullTime: sql.NullTime{time.Now(), true}}

	// TODO validate the struct
	now := time.Now()
	err = s.Update(m.scope, c)
	if err != nil {
		return
	}
	fmt.Println(time.Since(now))

	err = m.commitAutoTX()
	if err != nil {
		return
	}
	// TODO callback after

	return nil
}

// Delete the orm model by its primary keys.
// Error will return if no primaries are set.
func (m *Model) Delete() (err error) {
	defer func() { modelDefer(m, err) }()

	if !m.isInitialized {
		err = fmt.Errorf(errInit.Error(), "Delete", reflect.TypeOf(m.caller))
		return
	}

	err = m.scope.setFieldPermission(DELETE)
	if err != nil {
		return
	}

	if !m.scope.PrimariesSet() {
		err = fmt.Errorf(errPrimaryKey, m.name)
		return
	}

	// create where condition
	c := &sqlquery.Condition{}
	for _, col := range m.scope.PrimaryKeys() {
		c.Where(m.scope.Builder().QuoteIdentifier(col.Information.Name)+" = ?", m.scope.CallerField(col.Name).Interface())
	}

	s, err := m.strategy()
	if err != nil {
		return
	}

	// TODO callback before

	// call delete on strategy
	err = m.addAutoTX()
	if err != nil {
		return
	}

	err = s.Delete(m.scope, c)
	if err != nil {
		return
	}

	err = m.commitAutoTX()
	if err != nil {
		return
	}

	// TODO callback after

	return nil
}

// Count the existing rows by the given condition.
func (m *Model) Count(c *sqlquery.Condition) (int, error) {
	if !m.isInitialized {
		return 0, fmt.Errorf(errInit.Error(), "Delete", reflect.TypeOf(m.caller))
	}

	if c == nil {
		c = &sqlquery.Condition{}
	}

	row, err := m.builder.Select(m.Scope().TableName()).Condition(c).Columns(sqlquery.Raw("COUNT(*)")).First()
	if err != nil {
		return 0, err
	}

	var count int
	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}

	// logic here not in strategy
	return count, nil
}

// model of the orm.
func (m *Model) model() *Model {
	return m
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

// strategy sets the orm strategy if not added manually.
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

// addAutoTX adds a transaction on Create, Update, Delete if there was non added by the user.
func (m *Model) addAutoTX() error {
	fmt.Println(m.scope.Builder().HasTx(), m.caller.Scope().Builder().HasTx())
	if !m.scope.Builder().HasTx() && m.parentModel == nil && len(m.relations) > 0 {
		m.autoTx = true
		return m.Scope().Builder().Tx()
	}
	return nil
}

// commitAutoTX if exists and added by the system.
func (m *Model) commitAutoTX() error {
	if m.autoTx {
		m.autoTx = false
		fmt.Println("***** called commit !!!!!!!")
		return m.Scope().Builder().Commit()
	}
	return nil
}

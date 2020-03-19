// Package orm converts any struct to a full orm.
//
// See https://github.com/patrickascher/go-orm for more information and examples
package orm

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/patrickascher/gofw/cache"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/serenize/snaker"
	"gopkg.in/go-playground/validator.v9"
	"reflect"
	"strings"
	"time"
	"unsafe"
)

// GlobalBuilder for a global db connection
var GlobalBuilder *sqlquery_.Builder
var validate *validator.Validate

// Error for Model
var (
	ErrModelNotInitialized = errors.New("model: model is not initialized")
	ErrSetCache            = errors.New("model: cache must be set before the model gets initialized")
	ErrModelNoBuilder      = errors.New("model: no builder defined")
	ErrModelInit           = errors.New("model: method Initialize was not found in %s (forgot to embed Model struct?)")
	ErrModelFieldNotFound  = errors.New("model: field %s not found in columns of %s")
	ErrModelColumnNotFound = errors.New("model: column %s not found in columns of %s")
)

// Error for Strategy
var (
	ErrStrategyUnknown       = errors.New("model: unknown strategy %q (forgotten import?)")
	ErrStrategyNotGiven      = errors.New("model: empty strategy-name or driver is given")
	ErrStrategyAlreadyExists = errors.New("model: strategy %#v already exists")
)

// Errors for Table
var (
	ErrTableColumnNotFound = errors.New("model: column %s does not exist in table %s")
	ErrIDPrimary           = errors.New("model: no ID field exists or is not a primary key in table %s")
)

// other errors
var (
	ErrTagSyntax          = errors.New("model: bad tag syntax")
	ErrForeignKeyNotFound = errors.New("model: foreign key was not found for %s %s %s")
	ErrRelationType       = errors.New("model: relation type %s is not supported")
	ErrResultPtr          = errors.New("model: the result must be a pointer")
	ErrUpdateZero         = errors.New("model: update affected zero rows (update)")
	ErrDeletePk           = errors.New("model: one or more primary keys are empty (delete)")
	ErrDeleteNotFound     = errors.New("model: entry could not be found (delete)")
)

// Tag key names
const (
	TagRelation = "relation"
	TagFK       = "fk"
	TagName     = "orm"
	TagValidate = "validate"

	TagSkip      = "-"
	TagSeparator = ";"
	TagKeyValue  = ":"
)

// Loading strategies
const (
	Eager = "eager"
	//Lazy  = "lazy"
)

// Relation types
const (
	HasOne       = "hasOne"
	BelongsTo    = "belongsTo"
	HasMany      = "hasMany"
	ManyToMany   = "manyToMany"
	ManyToManySR = "manyToManySR"
	CustomStruct = "customStruct"
	CustomSlice  = "customSlice"
	CustomImpl   = "custom"
)

const (
	CallbackBefore = "OrmBefore"
	CallbackAfter  = "OrmAfter"
)

// Mode identifier
const (
	CREATE = "CreatedAt"
	UPDATE = "UpdatedAt"
	DELETE = "DeletedAt"
)

// Interface of orm models
type Interface interface {
	Initialize(caller Interface) error

	// defaults
	TableName() string
	DatabaseName() string
	Builder() (*sqlquery_.Builder, error)
	DefaultCache() (cache.Cache, time.Duration, error)
	Custom() bool

	// orm
	First(c *sqlquery_.Condition) error
	All(result interface{}, c *sqlquery_.Condition) error
	Create() error
	Update() error
	Delete() error
	Count(c *sqlquery_.Condition) (int, error)

	DisableSnapshot(bool)
	DisableCallback(bool)
	disableCallback() bool

	DisableCustomSql(bool)
	disableCustomSql() bool

	// cache
	Cache() (cache.Cache, time.Duration, error)
	SetCache(cache.Cache, time.Duration) error
	HasCache() bool

	// Transaction
	Tx() *sql.Tx
	SetTx(tx *sql.Tx)

	// helper for strategy
	Table() *Table
	SetStrategy(string) error

	// condition for relations
	SetRelationCondition(string, *sqlquery_.Condition) error
	RelationCondition() map[string]*sqlquery_.Condition

	// White and Blacklist
	SetWhitelist(...string) *Model
	SetBlacklist(...string) *Model
	whiteBlacklist() *WhiteBlackList
	WhiteBlacklist() *WhiteBlackList

	setWhiteBlacklist(*WhiteBlackList)

	setLoopMap(loopMap map[string][]string)
	getLoopMap() map[string][]string

	setParent(p Interface)
	hasParent() bool

	equalWith(y Interface, parent string) []ChangedValues
	Caller() Interface

	callback() *Callback
	parent() Interface
	resultSet() interface{}
}

// Model struct contains all fields and relations of the database table.
type Model struct {
	CreatedAt *sqlquery_.NullTime `orm:"permission:w" json:",omitempty"` //no nice solution but must be a ptr otherwise json omitempty will not work
	UpdatedAt *sqlquery_.NullTime `orm:"permission:w" json:",omitempty"` //no nice solution but must be a ptr otherwise json omitempty will not work
	DeletedAt *sqlquery_.NullTime `orm:"permission:w" json:",omitempty"` //no nice solution but must be a ptr otherwise json omitempty will not work

	caller Interface // for result
	table  *Table

	isInitialized bool
	skipRel       bool // needed for self-references

	//calledFromParent bool // TODO:delete - indicator if the model method is called directly or triggered by a through a relation.
	parentModel Interface
	resSet      interface{}
	cbk         *Callback

	whiteOrBlackList *WhiteBlackList

	relationCondition map[string]*sqlquery_.Condition

	cache    cache.Cache
	cacheTTL time.Duration
	tx       *sql.Tx
	customTx bool
	strategy string

	disableCb       bool //TODO create a better disable solution
	disableSnapshot bool //TODO create a better disable solution
	disableCustSql  bool //TODO create a better disable solution

	loopDetection bool
	loadedRel     []string            // needed to detect struct loops
	loopMap       map[string][]string // as map key the relation name is set and as slice string the already called arguments are set.
}

func (m *Model) Caller() Interface {
	return m.caller
}

func (m *Model) resultSet() interface{} {
	return m.resSet
}
func (m *Model) parent() Interface {
	return m.parentModel
}
func (m *Model) callback() *Callback {
	return m.cbk
}

func (m *Model) DisableSnapshot(b bool) {
	m.disableSnapshot = b
}

func (m *Model) Custom() bool {
	return false
}

func (m *Model) DisableCallback(b bool) {
	m.disableCb = b
}
func (m *Model) disableCallback() bool {
	return m.disableCb
}

func (m *Model) DisableCustomSql(b bool) {
	m.disableCustSql = b
}
func (m *Model) disableCustomSql() bool {
	return m.disableCustSql
}

func (m *Model) hasParent() bool {
	return m.parentModel != nil
}

func (m *Model) setParent(p Interface) {
	m.parentModel = p
}

// getLoopMap returns the loopMap.
func (m *Model) getLoopMap() map[string][]string {
	return m.loopMap
}

// setLoopMap sets the loopMap.
func (m *Model) setLoopMap(loopMap map[string][]string) {
	m.loopMap = loopMap
}

// checkLoopMap is checking if the relation was already asked before with the same where condition.
// this is a dirty way to regorgnice a infinity loop. after 10 loops its stopping.
// TODO create a different solution for this. maybe with a validation? works for now but its ugly^10!
func (m *Model) checkLoopMap(args string) error {
	if m.loopDetection {
		rel := structName(m.caller, true)
		counter := 0
		for _, b := range m.loopMap[rel] {
			if b == args {
				counter = counter + 1
			}
			if counter == 10 {
				return errors.New("congratulation you created a infinity loop")
			}
		}
		m.loopMap[rel] = append(m.loopMap[rel], args)
	}
	return nil
}

// SetRelationCondition adds a special condition for a relation.
func (m *Model) SetRelationCondition(name string, c *sqlquery_.Condition) error {
	if !m.isInit() {
		return ErrModelNotInitialized
	}

	for relName := range m.Table().Associations {
		if relName == name {
			m.relationCondition[relName] = c
			return nil
		}
	}

	return ErrModelFieldNotFound
}

func (m *Model) RelationCondition() map[string]*sqlquery_.Condition {
	return m.relationCondition
}

// SetWhitelist sets explicit fields/relations to the CRUD.
func (m *Model) SetWhitelist(list ...string) *Model {
	if len(list) == 0 {
		SetDefaultPermission(m, true)
		m.whiteOrBlackList = nil
	} else {
		m.whiteOrBlackList = NewWhiteBlackList(WHITELIST, list)
	}
	return m
}

// SetBlacklist removes fields/relations from the CRUD.
func (m *Model) SetBlacklist(list ...string) *Model {
	if len(list) == 0 {
		m.whiteOrBlackList = nil
		SetDefaultPermission(m, true)
	} else {
		m.whiteOrBlackList = NewWhiteBlackList(BLACKLIST, list)
	}
	return m
}

// whiteBlacklist returns the list
func (m *Model) whiteBlacklist() *WhiteBlackList {
	return m.whiteOrBlackList
}

func (m *Model) WhiteBlacklist() *WhiteBlackList {
	return m.whiteOrBlackList
}

// setWhiteBlacklist sets the list
func (m *Model) setWhiteBlacklist(wb *WhiteBlackList) {
	m.whiteOrBlackList = wb
	if wb == nil {
		SetDefaultPermission(m, true)
	}
}

// Tx returns the transaction of the model.
// returns nil if none is existing.
func (m *Model) Tx() *sql.Tx {
	return m.tx
}

// SetTx to set your own transaction.
// Keep in mind you also have to commit it on your own.
func (m *Model) SetTx(tx *sql.Tx) {
	m.tx = tx
	m.customTx = true
}

// Cache returns the current cache.
// If no cache is set, the default cache gets called.
func (m *Model) Cache() (cache.Cache, time.Duration, error) {
	if m.cache == nil {
		c, ttl, err := m.caller.DefaultCache()
		return c, ttl, err
	}
	return m.cache, m.cacheTTL, nil
}

func (m *Model) HasCache() bool {
	return m.cache != nil
}

// SetCache to add some custom cache for the model
func (m *Model) SetCache(c cache.Cache, d time.Duration) error {
	if m.isInitialized {
		return ErrSetCache
	}
	m.cache = c
	m.cacheTTL = d
	return nil
}

func CloneValue(source interface{}, destin interface{}) {
	x := reflect.ValueOf(source)
	if x.Kind() == reflect.Ptr {
		starX := x.Elem()
		y := reflect.New(starX.Type())
		starY := y.Elem()
		starY.Set(starX)
		reflect.ValueOf(destin).Elem().Set(y.Elem())
	} else {
		destin = x.Interface()
	}
}

// Initialize the model
//
// Workflow:
// - Initialize is setting the caller
// - Checking if the struct is already cached (if so, return from cache)
// - Validator is added
// - Initialize the DB Table
//    - Initialize default builder
//        - Call the struct `Builder`
//        - Check if builder is defined
//    - Call the struct `DatabaseName` and check if a specific database name is defined, otherwise use the one from the config
//    - Call the struct `TableName`
//    - Create and set the `orm.Table` struct to the model
//    - Set the loading strategy `Eager` (atm hardcoded, already prepaired to make it config able)
//    - Set variable `loadedRel` - to avoid struct loops
//    - Add all exported struct fields as `orm.Column`
//    - Describe the database table with the needed columns.
//        - Add db table column information to struct (name, position, nullable, primarykey, type, defaultvalue, length, autoincrement)
//        - RETURN ERROR IF STRUCT FIELD ID IS MISSING OR IS NO PRIMARYKEY.
// - Initialize Relations
// - Add information that the struct is initialized
// - Add to cache
func (m *Model) Initialize(caller Interface) error {
	// set caller
	m.caller = caller
	m.relationCondition = make(map[string]*sqlquery_.Condition)
	m.loopDetection = true
	m.loopMap = make(map[string][]string) // TODO define size of relations

	// initialize cache
	c, ttl, err := m.Cache()
	if err != nil {
		return err
	}

	// return cached model if exists
	modelName := structName(m.caller, true)
	if c.Exist(modelName) {
		cModel, err := c.Get(modelName)
		if err != nil {
			return err
		}
		mVal := cModel.Value().(*Model)

		//TODO find a better solution to copy the cached model to the new caller.
		field := reflect.ValueOf(caller).Elem().FieldByName("table")
		field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
		field.Set(reflect.ValueOf(mVal.table))
		SetDefaultPermission(m.caller, true) // TODO this should be handled different. maybe not *whitelist?

		//TODO find a better solution to set isInitialized
		field = reflect.ValueOf(caller).Elem().FieldByName("isInitialized")
		field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
		field.Set(reflect.ValueOf(true))

		//TODO find a better solution to set callbacks
		field = reflect.ValueOf(caller).Elem().FieldByName("cbk")
		field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
		newCallback := &Callback{}
		CloneValue(mVal.cbk, newCallback)
		field.Set(reflect.ValueOf(newCallback))
		newCallback.setCaller(caller)
		newCallback.setMode("")

		return nil
	}
	// -----------------------------------------------------------------------------------------------------------------

	// initialize the models table and columns
	err = m.initTable()
	if err != nil {
		return err
	}

	// initialize Relation
	if !m.skipRel {
		err = m.addRelation(m.caller)
		if err != nil {
			return err
		}
	}

	// add validation to the model
	validate = validator.New()       // TODO global?
	validate.SetTagName(TagValidate) // TODO global?
	validate.RegisterCustomTypeFunc(ValidateValuer, sqlquery_.NullInt64{}, sqlquery_.NullFloat64{}, sqlquery_.NullBool{}, sqlquery_.NullString{}, sqlquery_.NullTime{})

	err = m.addDBValidation()
	if err != nil {
		return err
	}

	// add callback struct
	m.cbk, err = NewCallback(m.caller)
	if err != nil {
		return err
	}

	// set flag that the struct is initialized
	m.isInitialized = true
	err = c.Set(modelName, m, ttl)
	if err != nil {
		return err
	}

	return nil
}

// First will check the first founded row by its condition and adds it values to the struct fields.
// Everything handled in the loading strategy.
// It will return an error if the model is not initialized or the strategy returns an error.
func (m *Model) First(c *sqlquery_.Condition) error {
	if !m.isInit() {
		return ErrModelNotInitialized
	}

	// reset resultSet
	m.resSet = nil

	// callback before
	// no tx.rollback is needed because there are only selects.
	err := m.cbk.callIfExists("First", true)
	if err != nil {
		return err
	}

	// create sql condition
	if c == nil {
		c = &sqlquery_.Condition{}
	}

	// configure white or blacklist, if set
	if m.whiteOrBlackList != nil {
		err := m.whiteOrBlackList.setFieldPermission(m, "first")
		if err != nil {
			return err
		}
	}

	err = m.checkLoopMap(c.Config(sqlquery_.WHERE))
	if err != nil {
		return err
	}

	err = m.table.strategy.First(m.caller, c)

	// callback after
	// no tx.rollback is needed because there are only selects.
	err = m.cbk.callIfExists("First", false)
	if err != nil {
		return err
	}

	return err
}

// CallMethodIfExist checks if the given interface has one of the given callback methods.
// The first method which exists, gets called and all others are getting ignored then.
// Arguments can be added as third argument.
// if there is one return argument, it gets treated as error return value.
func CallMethodIfExist(model Interface, callbacks []string, args ...interface{}) error {
	// check if callbacks are disabled
	if model.Caller().disableCallback() {
		return nil
	}

	for i, callback := range callbacks {

		cb := reflect.ValueOf(model.Caller()).MethodByName(callback)
		if cb.IsValid() {

			// check if i >0 = globalCallback (Before/After), add mode on first place.
			if i > 0 && (callback == CallbackBefore || callback == CallbackAfter) {
				mode := strings.Replace(callbacks[0], callback, "", 1)
				args = append([]interface{}{mode}, args...)

				// adding a empty result set
				if mode != "All" {
					var res []interface{}
					args = append(args, &res)
				}
			}

			in := make([]reflect.Value, len(args))
			for k, v := range args {
				in[k] = reflect.ValueOf(v)
			}

			if cb.Type().NumIn() != len(in) {
				in = make([]reflect.Value, 0) //TODO: should a error be thrown? or just call it without any args?
			}

			out := cb.Call(in)
			if len(out) == 1 && !out[0].IsNil() {
				return out[0].Interface().(error)
			}

			return nil
		}
	}

	return nil
}

// All will return all rows by its condition and puts it in the given result.
// Everything handled in the loading strategy.
// It will return an error if the model is not initialized or the strategy returns an error.
func (m *Model) All(result interface{}, c *sqlquery_.Condition) error {
	if !m.isInit() {
		return ErrModelNotInitialized
	}

	// set resultSet to model
	m.resSet = result

	// callback before
	// no tx.rollback is needed because there are only selects.
	err := m.cbk.callIfExists("All", true)
	if err != nil {
		return err
	}

	if c == nil {
		c = &sqlquery_.Condition{}
	}

	// configure white or blacklist, if set
	if m.whiteOrBlackList != nil {
		err := m.whiteOrBlackList.setFieldPermission(m, "All")
		if err != nil {
			return err
		}
	}

	err = m.checkLoopMap(c.Config(sqlquery_.WHERE))
	if err != nil {
		return err
	}

	err = m.table.strategy.All(result, m.caller, c)
	if err != nil {
		return err
	}

	// callback after
	// no tx.rollback is needed because there are only selects.
	err = m.cbk.callIfExists("All", false)
	if err != nil {
		return err
	}

	return nil
}

// timestampFieldExists checks if the given timestamp field exists
// used for checks on createdAt, updatedAt or deletedAt
func (m *Model) timestampFieldExists(field string) bool {
	col, err := m.Table().columnByName(snaker.CamelToSnake(field))
	if err != nil {
		return false
	}

	if col.ExistsInDB() { // don't check Information.Name because that is set always...
		return true
	}
	return false
}

// setTimestampOn sets the actual timestamp to the given field.
// it returns an error if the field does not exist.
// If its createdAt (the updatedAt and deletedAt write permission is getting removed)
// If its updatedAt (the createdAt and deletedAt write permission is getting removed)
// If its DeletedAt (the createdAt and updatedAt write permission is getting removed)
// is is necessary because of a bug in null.Time which is not allowed to be nil if its a ptr.
func (m *Model) setTimestampOn(field string) error {

	timestamp := reflectField(m, field)

	col, err := m.Table().columnByName(snaker.CamelToSnake(field))
	if err != nil {
		return err
	}
	col.Permission.Write = true

	t := sqlquery_.NullTime{Time: time.Now(), Valid: true}
	switch timestamp.Kind() {
	case reflect.Ptr:
		timestamp.Set(reflect.ValueOf(&t))
	case reflect.Struct:
		timestamp.Set(reflect.ValueOf(t))
	}

	// disable the other two timestamp fields if exist.
	// TODO this is needed at the moment because the *null.Time struct can not handle nil values????
	// solution: implement our own nullTime struct.
	switch field {
	case CREATE:
		if m.timestampFieldExists(UPDATE) {
			col, err := m.Table().columnByName(snaker.CamelToSnake(UPDATE))
			if err != nil {
				return err
			}
			col.Permission.Write = false
		}
		if m.timestampFieldExists(DELETE) {
			col, err := m.Table().columnByName(snaker.CamelToSnake(DELETE))
			if err != nil {
				return err
			}
			col.Permission.Write = false
		}
	case UPDATE:
		if m.timestampFieldExists(CREATE) {
			col, err := m.Table().columnByName(snaker.CamelToSnake(CREATE))
			if err != nil {
				return err
			}
			col.Permission.Write = false
		}
		if m.timestampFieldExists(DELETE) {
			col, err := m.Table().columnByName(snaker.CamelToSnake(DELETE))
			if err != nil {
				return err
			}
			col.Permission.Write = false
		}
	case DELETE:
		if m.timestampFieldExists(CREATE) {
			col, err := m.Table().columnByName(snaker.CamelToSnake(CREATE))
			if err != nil {
				return err
			}
			col.Permission.Write = false
		}
		if m.timestampFieldExists(UPDATE) {
			col, err := m.Table().columnByName(snaker.CamelToSnake(UPDATE))
			if err != nil {
				return err
			}
			col.Permission.Write = false
		}
	}

	return nil
}

// Create an entry with the actual struct value.
// Everything handled in the loading strategy.
// If there is no custom transaction added, it will add one by default and also commits it automatically if everything is ok. Otherwise a Rollback will be called.
// It will return an error if the model is not initialized, tx  error, the strategy returns an error or a commit error happens.
func (m *Model) Create() error {
	var err error
	callRollbackOnErr := true

	defer func() {
		if p := recover(); p != nil {
			_ = m.Tx().Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil && callRollbackOnErr {
			_ = m.Tx().Rollback() // err is non-nil; don't change it
		}
		return
	}()

	if !m.isInit() {
		if m.tx == nil {
			callRollbackOnErr = false
		}
		err = ErrModelNotInitialized
		return err
	}

	// reset resultSet
	m.resSet = nil

	// transaction
	err = m.addTx()
	if err != nil {
		callRollbackOnErr = false
		return err
	}

	// callback before
	err = m.cbk.callIfExists("Create", true)
	if err != nil {
		return err
	}

	// configure white or blacklist, if set
	if m.whiteOrBlackList != nil {
		err = m.whiteOrBlackList.setFieldPermission(m, "create")
		if err != nil {
			return err
		}
	}

	// set the CreatedAt info if exists
	if m.timestampFieldExists(CREATE) {
		err = m.setTimestampOn(CREATE)
		if err != nil {
			return err
		}
	}

	// validate the struct
	err = m.isValid()
	if err != nil {
		return err
	}

	// call create on strategy
	err = m.table.strategy.Create(m.caller)
	if err != nil {
		callRollbackOnErr = false // rollback is already done in strategy
		return err
	}

	// callback after
	// its before the TX, that the transaction can still fail and rollback
	err = m.cbk.callIfExists("Create", false)
	if err != nil {
		return err
	}

	// commit if tx was not added manually
	err = m.commit()
	if err != nil {
		callRollbackOnErr = false // no rollback needed
		return err
	}

	return nil
}

// Update an entry with the actual struct value.
// Everything handled in the loading strategy.
// If there is no custom transaction added, it will add one by default and also commits it automatically if everything is ok. Otherwise a Rollback will be called.
// It will return an error if the model is not initialized, tx  error, the strategy returns an error or a commit error happens.
func (m *Model) Update() error {
	var err error
	callRollbackOnErr := true

	defer func() {
		if p := recover(); p != nil {
			_ = m.Tx().Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil && callRollbackOnErr {
			_ = m.Tx().Rollback() // err is non-nil; don't change it
		}
		return
	}()

	if !m.isInit() {
		if m.tx == nil {
			callRollbackOnErr = false
		}
		return ErrModelNotInitialized
	}

	// reset resultSet
	m.resSet = nil

	// check if primary fields exist
	if checkPrimaryFieldsEmpty(m.caller) {
		if m.tx == nil {
			callRollbackOnErr = false
		}
		err = ErrDeletePk
		return err
	}

	// create where condition
	c := &sqlquery_.Condition{}
	for _, col := range m.Table().PrimaryKeys() {
		c.Where(m.Table().Builder.QuoteIdentifier(col.Information.Name)+" = ?", reflectField(m.caller, col.StructField).Interface())
	}

	// transaction
	err = m.addTx()
	if err != nil {
		return err
	}

	// callback before
	// its after the TX that the default/custom transactions can be used in callbacks as well.
	err = m.cbk.callIfExists("Update", true)
	if err != nil {
		return err
	}

	// configure white or blacklist, if set
	if m.whiteOrBlackList != nil {
		err := m.whiteOrBlackList.setFieldPermission(m, "update")
		if err != nil {
			return err
		}
	}

	// snapshot
	if !m.hasParent() && !m.disableSnapshot {
		fmt.Println("############# SNAP SNAP START ###############")
		snapshot := newValueInstanceFromType(reflect.TypeOf(m.caller)).Addr().Interface().(Interface)

		cache2, ttl, errTmp := m.caller.Cache()
		if errTmp != nil {
			err = errTmp
			return err
		}

		err = snapshot.SetCache(cache2, ttl)
		if err != nil {
			return err
		}

		err = snapshot.Initialize(snapshot)
		if err != nil {
			return err
		}

		err = snapshot.First(c)
		if err != nil {
			return err
		}

		changes := m.equalWith(snapshot, "")
		changesKeys := m.getFieldsFromChanges(changes)

		// if a user added a whitelist on his own, only add changes in the user list
		//TODO this logic should be moved into the white_black_list.go create a compare method
		if m.whiteOrBlackList != nil && m.whiteOrBlackList.list != nil {

			if m.whiteOrBlackList.policy == WHITELIST {
				for i, userWhitelist := range m.whiteOrBlackList.list {
					exists := false
					for _, chaKeys := range changesKeys {
						if userWhitelist == chaKeys || strings.HasPrefix(chaKeys+".", userWhitelist) {
							exists = true
						}
					}
					if !exists {
						if len(changesKeys) >= i+2 {
							m.whiteOrBlackList.list = append(m.whiteOrBlackList.list[:i], m.whiteOrBlackList.list[i+1:]...)
						} else {
							m.whiteOrBlackList.list = m.whiteOrBlackList.list[:i]
						}
					}
				}
				changesKeys = m.whiteOrBlackList.list
			} else {
				for i, chaKeys := range changesKeys {
					exists := false
					for _, userWhitelist := range m.whiteOrBlackList.list {
						if userWhitelist == chaKeys { // take care of dot notation... this is a little bit tricky if there is only one field blacklisted in a relation. then we have to get all the other relation fields.
							exists = true
						}
					}
					if exists {
						if len(changesKeys) >= i+2 {
							changesKeys = append(changesKeys[:i], changesKeys[i+1:]...)
						} else {
							changesKeys = changesKeys[:i]
						}
					}
				}
			}
		}

		if changesKeys == nil || len(changesKeys) == 0 {
			//TODO - error no data was changed???
			fmt.Println("#### SNAPSHOT #### no data was changed", m.equalWith(snapshot, ""))
			return nil
		}

		m.SetWhitelist(changesKeys...)
		err = m.whiteOrBlackList.setFieldPermission(m, "update") // has to be called again to configure it again.
		if err != nil {
			return err
		}

		fmt.Println(m.getFieldsFromChanges(changes), changesKeys, "############# SNAP SNAP END###############")
	}

	// set the UpdatedAt info if exists
	if m.timestampFieldExists(UPDATE) {
		err = m.setTimestampOn(UPDATE)
		if err != nil {
			return err
		}
	}

	// validate the struct
	err = m.isValid()
	if err != nil {
		return err
	}

	// call update on strategy
	err = m.table.strategy.Update(m.caller, c)
	if err != nil {
		callRollbackOnErr = false // rollback is already done in strategy
		return err
	}

	// callback after
	// its before the TX that the default/custom transactions can be used in callbacks as well.
	err = m.cbk.callIfExists("Update", false)
	if err != nil {
		return err
	}

	// commit if tx was not added manually
	err = m.commit()
	if err != nil {
		callRollbackOnErr = false // no rollback needed
		return err
	}

	return nil
}

type ChangedValues struct {
	field    string
	relation string //at the moment only belongsTo is set. nothing else needed so far. TODO for later when more time
	index    int
	old      interface{}
	new      interface{}
}

func (m *Model) getFieldsFromChanges(ne []ChangedValues) []string {
	var keys []string
	for _, f := range ne {
		keys = append(keys, f.field)
	}
	return keys
}

func (m *Model) equalWith(snapshot Interface, parent string) []ChangedValues {

	var ne []ChangedValues

	// normal fields
	if parent != "" {
		parent = parent + "."
	}

	for _, col := range m.caller.Table().Columns(WRITEDB) {
		// skip the automatic time fields
		if col.StructField == DELETE || col.StructField == UPDATE || col.StructField == CREATE {
			continue
		}

		newValue := reflectField(m.caller, col.StructField).Interface()
		old := reflectField(snapshot, col.StructField).Interface()
		if old != newValue {
			ne = append(ne, ChangedValues{field: parent + col.StructField, old: old, new: newValue})
		}
	}

	// relations
	for field, rel := range m.caller.Table().Relations(m.whiteBlacklist(), WRITEDB) {

		switch rel.Type {
		case HasOne, CustomStruct:
			// in a hasOne relation all given fields/relations are checked with the snapshot.
			newValue := reflectField(m.caller, field).Addr().Interface().(Interface)
			cache, ttl, _ := m.caller.Cache()
			_ = newValue.SetCache(cache, ttl)
			_ = newValue.Initialize(newValue)
			oldValue := reflectField(snapshot, field).Addr().Interface().(Interface)

			res := newValue.equalWith(oldValue, parent+field)
			if res != nil {
				ne = append(ne, res...)
			}
		case BelongsTo:
			// a belongsTo relations is checked like:
			// User{BelongsID} == Belongs{Id}. There is no need to check the whole BelongsTo entry because
			// its just connected with the foreign key.
			actualModelData := reflectField(m.caller, field).Addr().Interface().(Interface)

			oldValue := reflectField(snapshot, rel.StructTable.StructField)
			newValue := reflectField(actualModelData, rel.AssociationTable.StructField)

			if !oldValue.IsValid() || !newValue.IsValid() || oldValue.Interface() != newValue.Interface() {
				ne = append(ne, ChangedValues{relation: BelongsTo, field: rel.StructTable.StructField, old: oldValue.Interface(), new: newValue.Interface()})
			}
		case HasMany, CustomSlice:
			// has many checks if the length of the new and old slice is different.
			// if so, a change gets returned. At the moment no specifics over the old and new value are returned (just the field name for the whitelist)
			// if the length is the same, every field is getting checked if its different. If a field is not equal a changedValue will return without any specifics at the moment.
			// if this gets implemented, an index is needed. already added it to the ChangedValues struct.
			newValue := reflectField(m.caller, field)
			oldValue := reflectField(snapshot, field)

			if newValue.Len() != oldValue.Len() {
				//ne = append(ne, ChangedValues{field: field, old: oldValue.Len(), new: newValue.Len()})
				ne = append(ne, ChangedValues{field: field, old: oldValue.Len(), new: newValue.Len()})
			} else {
				// check values
				var _ne []ChangedValues
				for i := 0; i < newValue.Len(); i++ {

					newValueI := reflect.Indirect(newValue.Index(i)).Addr().Interface().(Interface)
					cache, ttl, _ := m.caller.Cache()
					_ = newValueI.SetCache(cache, ttl)
					_ = newValueI.Initialize(newValueI)
					oldValueI := reflect.Indirect(oldValue.Index(i)).Addr().Interface().(Interface)

					res := newValueI.equalWith(oldValueI, parent+field)
					if res != nil {
						_ne = append(_ne, res...)
					}
				}

				if _ne != nil {
					ne = append(ne, ChangedValues{field: field, old: "not implemented yet", new: "not implemented yet"})
				}
			}

		case ManyToMany, ManyToManySR:
			// check only the linked ids
			newValue := reflectField(m.caller, field)
			oldValue := reflectField(snapshot, field)

			var oldInt []int
			var newInt []int

			for i := 0; i < newValue.Len(); i++ {
				v := reflectField(newValue.Index(i).Addr().Interface().(Interface), rel.AssociationTable.StructField)
				newInt = append(newInt, int(v.Int()))
			}
			for i := 0; i < oldValue.Len(); i++ {
				v := reflectField(oldValue.Index(i).Addr().Interface().(Interface), rel.AssociationTable.StructField)
				oldInt = append(oldInt, int(v.Int()))
			}

			// check if its the same value
			same := true
			for _, i := range oldInt {
				if !inSlice(i, newInt) {
					same = false
				}
			}
			for _, i := range newInt {
				if !inSlice(i, oldInt) {
					same = false
				}
			}

			if newValue.Len() != oldValue.Len() || !same {
				ne = append(ne, ChangedValues{field: field, old: oldInt, new: newInt})
			}
		}

	}

	return ne
}

// Delete an entry by the given primary key(s).
// If a softDelete field is existing, it will update that field with the current timestamp instead of deleting the entry.
// Everything handled in the loading strategy.
// If there is no custom transaction added, it will add one by default and also commits it automatically if everything is ok. Otherwise a Rollback will be called.
// It will return an error if the model is not initialized, tx  error, the strategy returns an error or a commit error happens.
func (m *Model) Delete() error {

	var err error
	callRollbackOnErr := true

	defer func() {
		if p := recover(); p != nil {
			_ = m.Tx().Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil && callRollbackOnErr {
			_ = m.Tx().Rollback() // err is non-nil; don't change it
		}
		return
	}()

	if !m.isInit() {
		if m.tx == nil {
			callRollbackOnErr = false
		}
		err = ErrModelNotInitialized
		return err
	}

	// reset resultSet
	m.resSet = nil

	// check if primary fields exist to avoid a delete *
	if checkPrimaryFieldsEmpty(m.caller) {
		return ErrDeletePk
	}

	// create where condition
	c := &sqlquery_.Condition{}
	for _, col := range m.Table().PrimaryKeys() {
		c.Where(m.Table().Builder.QuoteIdentifier(col.Information.Name)+" = ?", reflectField(m.caller, col.StructField).Interface())
	}

	// set the DeletedAt info if exists
	if m.timestampFieldExists(DELETE) {
		if m.tx == nil {
			callRollbackOnErr = false
		}
		err = m.setTimestampOn(DELETE)
		if err != nil {
			return err
		}
		err = m.SetWhitelist(DELETE).Update()
		if err != nil {
			return err
		}
	} else {

		// transaction
		err = m.addTx()
		if err != nil {
			return err
		}

		// callback before
		// its after the TX that the default/custom transactions can be used in callbacks as well.
		err = m.cbk.callIfExists("Delete", true)
		if err != nil {
			return err
		}

		// call delete on strategy
		err = m.table.strategy.Delete(m.caller, c)
		if err != nil {
			callRollbackOnErr = false // rollback is already done in strategy
			return err
		}

		// callback after
		// its before the TX that the default/custom transactions can be used in callbacks as well.
		err = m.cbk.callIfExists("Delete", false)
		if err != nil {
			return err
		}

		// commit if tx was not added manually
		err = m.commit()
		if err != nil {
			callRollbackOnErr = false // no rollback needed anymore
			return err
		}
	}

	return nil
}

// commit the transaction
func (m *Model) commit() error {
	if !m.customTx {
		if err := m.tx.Commit(); err != nil {
			if errR := m.tx.Rollback(); errR != nil {
				return errR
			}
			return err
		}
		m.tx = nil
	}

	return nil
}

// Count the existing rows by the given condition.
func (m *Model) Count(c *sqlquery_.Condition) (int, error) {
	if !m.isInit() {
		return 0, ErrModelNotInitialized
	}

	b := m.Table().Builder
	if c == nil {
		c = &sqlquery_.Condition{}
	}
	row, err := b.Select(m.Table().Name).Condition(c).Columns("!COUNT(*)").First()
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

// Table information
func (m *Model) Table() *Table {
	return m.table
}

// isInit checks if the model got already initialized.
func (m Model) isInit() bool {
	return m.isInitialized
}

// addTx adds a new transaction to the model.
// It sets customTX to false as identifier.
func (m *Model) addTx() error {
	// add tx only if it was not added manually already
	if m.tx == nil {
		tx, err := m.Table().Builder.NewTx()
		if err != nil {
			return err
		}

		m.tx = tx
		m.customTx = false
	}
	return nil
}

// SetStrategy to the model
func (m *Model) SetStrategy(s string) error {
	if m.isInit() {
		err := m.setStrategy(s)
		if err != nil {
			return err
		}
	} else {
		m.strategy = s
	}
	return nil
}

// SetStrategy to the model
func (m *Model) setStrategy(s string) error {

	st, err := NewStrategy(s)
	if err != nil {
		return err
	}
	m.table.strategy = st

	return nil
}

// initTable is getting the database, table name and all table columns.
// It will return an error if a struct field does not exist in the table or no Builder is defined.
// It also ensures that the Field ID is given and is a primary key in the table.
func (m *Model) initTable() error {
	// initialize builder for the table
	b, err := m.initBuilder()
	if err != nil {
		return err
	}

	// check if user defined his own database, otherwise take database from config
	db := m.caller.DatabaseName()
	if db == "" {
		db = b.Config().DbName()
	}

	// get struct table name
	t := m.caller.TableName()

	// create new table struct with builder and the database and table name
	m.table = &Table{Builder: b, Name: t, Database: db, Associations: make(Associations)}

	// add strategy
	loadingStrategy := Eager
	if m.strategy != "" {
		loadingStrategy = m.strategy
	}
	err = m.setStrategy(loadingStrategy)
	if err != nil {
		return err
	}

	// adding all exported struct fields as table columns
	m.loadedRel = append(m.loadedRel, structName(m.caller, true))
	m.addStructFieldsToTableColumn(m.caller)

	if m.strategy == CustomImpl {
		return nil
	}

	// describe table columns and merge information
	return m.table.describe()
}

// initBuilder checks if a builder is given, otherwise an error will return.
func (m *Model) initBuilder() (*sqlquery_.Builder, error) {
	//checking default config
	b, err := m.caller.Builder()
	if err != nil {
		return nil, err
	}

	// check if the custom added Builder has a value
	if b == nil {
		return nil, ErrModelNoBuilder
	}

	return b, nil
}

// structName is a helper to get the name of the struct with or without the namespace.
func structName(s interface{}, withNamespace bool) string {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		if withNamespace == true {
			return t.Elem().String()
		}
		return t.Elem().Name()
	}
	if withNamespace == true {
		return t.String()
	}
	return t.Name()
}

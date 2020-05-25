package orm_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/patrickascher/gofw/cache"
	"github.com/patrickascher/gofw/orm"
	_ "github.com/patrickascher/gofw/orm/strategy"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
)

type Role struct {
	orm.Model
	ID   int
	Name orm.NullString

	Roles []*Role
}

func truncate(ormI orm.Interface) error {
	b := ormI.DefaultBuilder()

	// owner belongsTo
	_, err := b.Delete("owners").Exec()
	if err != nil {
		return err
	}

	// wheels hasMany
	_, err = b.Delete("wheels").Exec()
	if err != nil {
		return err
	}

	// wheels hasMany
	_, err = b.Delete("wheels").Exec()
	if err != nil {
		return err
	}

	// components hasOne,hasMany poly
	_, err = b.Delete("components").Exec()
	if err != nil {
		return err
	}

	// join table m2m drivers
	_, err = b.Delete("car_drivers").Exec()
	if err != nil {
		return err
	}

	// drivers m2m
	_, err = b.Delete("drivers").Exec()
	if err != nil {
		return err
	}

	// cars main
	_, err = b.Delete("cars").Exec()
	if err != nil {
		return err
	}

	// roles join table
	_, err = b.Delete("role_roles").Exec()
	if err != nil {
		return err
	}

	// roles
	_, err = b.Delete("roles").Exec()
	if err != nil {
		return err
	}

	return nil
}

func createEntries(ormI orm.Interface) error {
	b := ormI.DefaultBuilder()

	err := truncate(ormI)
	if err != nil {
		return err
	}

	// owners
	_, err = b.Insert("owners").Columns("id", "name").Values([]map[string]interface{}{{"id": 1, "name": "John Doe"}, {"id": 2, "name": "Foo Bar"}}).Exec()
	if err != nil {
		return fmt.Errorf("owners: %w", err)
	}

	// cars
	cars := []map[string]interface{}{
		{"id": 1, "brand": "BMW", "owner_id": 1, "type": "M3", "year": 2000},
		{"id": 2, "brand": "Mercedes", "owner_id": 2, "type": "SLK", "year": 2001},
	}
	_, err = b.Insert("cars").Columns("id", "brand", "owner_id", "type", "year").Values(cars).Exec()
	if err != nil {
		return fmt.Errorf("cars: %w", err)
	}

	// drivers
	drivers := []map[string]interface{}{
		{"id": 1, "name": "Pat"},
		{"id": 2, "name": "Tom"},
		{"id": 3, "name": "Marc"},
	}
	_, err = b.Insert("drivers").Columns("id", "name").Values(drivers).Exec()
	if err != nil {
		return fmt.Errorf("drivers: %w", err)
	}

	// owner hasMany
	carDriversMap := []map[string]interface{}{
		{"car_id": 1, "driver_id": 1},
		{"car_id": 1, "driver_id": 2},
		{"car_id": 2, "driver_id": 3},
	}
	_, err = b.Insert("car_drivers").Columns("car_id", "driver_id").Values(carDriversMap).Exec()
	if err != nil {
		return fmt.Errorf("car_drivers: %w", err)
	}

	// components
	components := []map[string]interface{}{
		{"id": 1, "car_id": 1, "car_type": "radio", "brand": "AEG", "note": "Bass"},
		{"id": 2, "car_id": 1, "car_type": "liquid", "brand": "Molly", "note": "cheap"},
		{"id": 3, "car_id": 2, "car_type": "liquid", "brand": "Molly", "note": "cheap"},
	}
	_, err = b.Insert("components").Columns("id", "car_id", "car_type", "brand", "note").Values(components).Exec()
	if err != nil {
		return fmt.Errorf("components: %w", err)
	}

	// roles
	roles := []map[string]interface{}{
		{"id": 1, "name": "Admin"},
		{"id": 2, "name": "Writer"},
		{"id": 3, "name": "User"},
		{"id": 4, "name": "Guest"},
	}
	_, err = b.Insert("roles").Columns("id", "name").Values(roles).Exec()
	if err != nil {
		return fmt.Errorf("roles: %w", err)
	}

	// roles
	rolesMap := []map[string]interface{}{
		{"role_id": 1, "child_id": 2},
		{"role_id": 1, "child_id": 3},
		{"role_id": 2, "child_id": 3},
	}
	_, err = b.Insert("role_roles").Columns("role_id", "child_id").Values(rolesMap).Exec()
	if err != nil {
		return fmt.Errorf("role_roles: %w", err)
	}

	return nil
}

type car struct {
	orm.Model

	ID      orm.NullInt `orm:"select:id"`
	OwnerID orm.NullInt

	Brand string `orm:"permission:w"`

	// relation tests (belongsTo, m2m, hasMany)
	Owner  *owner    `orm:"relation:belongsTo"`
	Driver []*driver `orm:"relation:m2m"`
	Wheels []wheel   `orm:"permission:w"`

	// polymorphic tests (hasOne, hasMany)
	Radio  radio    `orm:"polymorphic:Car;polymorphic_value:radio"`
	Liquid []liquid `orm:"polymorphic:Car;polymorphic_value:liquid"`
}
type carWithoutCache struct {
	orm.Model

	ID      orm.NullInt `orm:"select:id"`
	OwnerID orm.NullInt
}

type carEnum struct {
	orm.Model

	ID   orm.NullInt
	Enum orm.NullString

	OwnerID orm.NullInt
	Owner   *owner `orm:"relation:belongsTo"`
}

type carJsonNameAndSkip struct {
	orm.Model

	ID      orm.NullInt `json:"pid"`
	OwnerID orm.NullInt `json:"-"`
}

func (c carEnum) DefaultTableName() string {
	return "cars"
}
func (c carJsonNameAndSkip) DefaultTableName() string {
	return "cars"
}
func (c carWithoutCache) DefaultTableName() string {
	return "cars"
}
func (c carWithoutCache) DefaultCache() (cache.Interface, time.Duration, error) {
	return nil, 0, nil
}

type carBackRef struct {
	orm.Model

	ID      orm.NullInt `orm:"select:id"`
	OwnerID orm.NullInt
	Owner   *ownerBackRef `orm:"relation:belongsTo;fk:OwnerID"`
}

func (c carBackRef) DefaultTableName() string {
	return "cars"
}

type ownerBackRef struct {
	orm.Model

	ID   int
	Name orm.NullString

	// relation test hasOne
	Car *carBackRef `orm:"afk:OwnerID"`
}

func (o ownerBackRef) DefaultTableName() string {
	return "owners"
}

type owner struct {
	orm.Model

	ID   int
	Name orm.NullString

	// relation test hasOne
	//Car *car
}

type driver struct {
	orm.Model

	ID   int
	Name orm.NullString
	//Car  []*car `orm:"relation:m2m;join_table:car_drivers;"`
}

type wheel struct {
	orm.Model

	ID    int
	Brand orm.NullString
	Note  orm.NullString

	CarID int
}
type radio struct {
	orm.Model
	Component
}

func (r radio) DefaultTableName() string {
	return "components"
}

type liquid struct {
	orm.Model
	Component
}

func (l liquid) DefaultTableName() string {
	return "components"
}

type Component struct {
	ID    int
	Brand orm.NullString
	Note  orm.NullString

	CarID   int
	CarType string
}

type carNoBuilder struct {
	orm.Model

	ID      orm.NullInt `orm:"select:id"`
	OwnerID orm.NullInt
}

func (c carNoBuilder) DefaultBuilder() sqlquery.Builder {
	return sqlquery.Builder{}
}

type carNoTableName struct {
	orm.Model

	ID      orm.NullInt `orm:"select:id"`
	OwnerID orm.NullInt
}

func (c carNoTableName) DefaultTableName() string {
	return ""
}
func (c carNoTableName) DefaultDatabaseName() string {
	return ""
}

type carErrCache struct {
	orm.Model

	ID      orm.NullInt `orm:"select:id"`
	OwnerID orm.NullInt
}

func (c carErrCache) DefaultCache() (cache.Interface, time.Duration, error) {
	return orm.GlobalCache, 6 * time.Hour, errors.New("cache error")
}

func TestModel_Init(t *testing.T) {
	test := assert.New(t)

	// ok
	c := &car{}
	err := c.Init(c)
	c.Liquid = append(c.Liquid, liquid{})
	test.NoError(err)
	test.Equal("orm_test.car", c.Scope().Name(true))
	test.Equal(c, c.Scope().Caller())
	test.True(len(c.Scope().Fields(orm.Permission{})) > 0)
	test.True(len(c.Scope().Relations(orm.Permission{})) > 0)
	id, err := c.Scope().Field("ID")
	test.NoError(err)
	id.Permission = orm.Permission{Read: true, Write: false}
	test.Equal(orm.Permission{Read: true, Write: false}, id.Permission)

	// ok from cache
	cCache := &car{}
	err = cCache.Init(cCache)
	test.NoError(err)
	test.Equal("orm_test.car", c.Scope().Name(true))
	test.Equal(cCache, cCache.Scope().Caller())
	test.True(len(cCache.Scope().Fields(orm.Permission{})) > 0)
	test.True(len(cCache.Scope().Relations(orm.Permission{})) > 0)
	test.False(len(cCache.Liquid) == 1)
	id, err = cCache.Scope().Field("ID")
	test.NoError(err)
	test.Equal(orm.Permission{Read: true, Write: true}, id.Permission)

	// error no cache is defined
	c2 := &carWithoutCache{}
	err = c2.Init(c2)
	test.Error(err)

	// no pointer value is set
	c = &car{}
	err = c.Init(nil)
	test.Error(err)

	// no builder is given
	cBuilder := &carNoBuilder{}
	err = cBuilder.Init(cBuilder)
	test.Error(err)

	// no tablename or database name is given
	cTable := &carNoTableName{}
	err = cTable.Init(cTable)
	test.Error(err)
}

func TestModel_Scope(t *testing.T) {
	c := car{}
	err := c.Init(&c)
	assert.NoError(t, err)
	assert.IsType(t, &orm.Scope{}, c.Scope())
}

func TestModel_SetWBList(t *testing.T) {
	c := car{}
	err := c.Init(&c)
	assert.NoError(t, err)

	// no list is defined
	p, f := c.WBList()
	assert.Equal(t, orm.WHITELIST, p)
	assert.Equal(t, []string(nil), f)

	// defined list
	c.SetWBList(orm.BLACKLIST, "ID")
	p, f = c.WBList()
	assert.Equal(t, orm.BLACKLIST, p)
	assert.Equal(t, []string{"ID"}, f)
}

func TestModel_Cache_SetCache(t *testing.T) {
	c := car{}

	// err orm is not init
	cache_, ttl, err := c.Cache()
	assert.Equal(t, time.Duration(0), ttl)
	assert.Equal(t, nil, cache_)
	assert.Error(t, err)

	err = c.Init(&c)
	assert.NoError(t, err)

	// ok
	cache_, ttl, err = c.Cache()
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(cache.INFINITY), ttl)
	// set cache error model already init
	err = c.SetCache(cache_, ttl)
	assert.Error(t, err)

	// no cache
	cNoCache := carWithoutCache{}
	err = cNoCache.Init(&cNoCache)
	assert.Error(t, err)
	// set cache - error cache is nil
	err = cNoCache.SetCache(nil, ttl)
	assert.Error(t, err)
	// set cache - ttl is 0
	// 0 is infinity, so its allowed
	err = cNoCache.SetCache(cache_, 0)
	assert.NoError(t, err)
	// set cache - object was not init before
	err = cNoCache.SetCache(cache_, ttl)
	assert.NoError(t, err)
	// init now ok
	err = cNoCache.Init(&cNoCache)
	assert.NoError(t, err)

	// err on default cache
	cErrCache := carErrCache{}
	err = cErrCache.Init(&cErrCache)
	assert.Error(t, err)

}

func TestModel_First(t *testing.T) {
	// create db entries
	err := createEntries(&Role{})
	assert.NoError(t, err)

	c := car{}
	// err - not initialized
	err = c.First(nil)
	assert.Error(t, err)

	// init
	err = c.Init(&c)
	assert.NoError(t, err)

	// fetch first
	err = c.First(nil)
	assert.NoError(t, err)
	assert.True(t, c.ID.Int64 != 0)

	err = truncate(&Role{})
	assert.NoError(t, err)
}

func TestModel_All(t *testing.T) {
	// create db entries
	err := createEntries(&Role{})
	assert.NoError(t, err)

	c := car{}
	var cRes []car
	// err - not initialized
	err = c.All(&cRes, nil)
	assert.Error(t, err)

	// init
	err = c.Init(&c)
	assert.NoError(t, err)

	// err - result is nil
	err = c.All(nil, nil)
	assert.Error(t, err)

	// err - result is no ptr
	err = c.All(cRes, nil)
	assert.Error(t, err)

	// fetch all
	err = c.All(&cRes, nil)
	assert.NoError(t, err)
	assert.True(t, len(cRes) > 0)
}

func TestModel_Create(t *testing.T) {
	// create db entries
	err := truncate(&Role{})
	assert.NoError(t, err)

	c := car{}

	// err - not initialized
	err = c.Create()
	assert.Error(t, err)

	// init
	err = c.Init(&c)
	assert.NoError(t, err)

	// ok - no save action because its empty
	err = c.Create()
	assert.NoError(t, err)
	assert.False(t, c.Scope().Builder().HasTx())

	// err - brand is mandatory
	c.ID = orm.NewNullInt(1)
	err = c.Create()
	assert.Error(t, err)
	assert.False(t, c.Scope().Builder().HasTx())

	// ok - brand is mandatory
	c.Brand = "XY"
	err = c.Create()
	assert.NoError(t, err)
	assert.False(t, c.Scope().Builder().HasTx())
}

func TestModel_Update(t *testing.T) {
	// create db entries
	err := createEntries(&Role{})
	assert.NoError(t, err)

	c := car{}

	// err - not initialized
	err = c.Update()
	assert.Error(t, err)

	// init
	err = c.Init(&c)
	assert.NoError(t, err)

	// err - primary is not set
	err = c.Update()
	assert.Error(t, err)
	assert.False(t, c.Scope().Builder().HasTx())

	// err - brand is mandatory
	c.ID = orm.NewNullInt(1)
	err = c.Update()
	assert.Error(t, err)
	assert.False(t, c.Scope().Builder().HasTx())

	// ok - brand is mandatory
	c.Brand = "XY"
	err = c.Update()
	assert.NoError(t, err)
	assert.False(t, c.Scope().Builder().HasTx())

	// no err - no changes - = nil return
	assert.False(t, c.Scope().Builder().HasTx())

	err = c.Update()
	assert.NoError(t, err)
	assert.False(t, c.Scope().Builder().HasTx())
}

func TestModel_Delete(t *testing.T) {
	// create db entries
	err := createEntries(&Role{})
	assert.NoError(t, err)

	c := car{}

	// err - not initialized
	err = c.Delete()
	assert.Error(t, err)

	// init
	err = c.Init(&c)
	assert.NoError(t, err)

	// err - not primaries set
	err = c.Delete()
	assert.Error(t, err)

	//ok
	c.ID = orm.NewNullInt(1)
	err = c.Delete()
	assert.NoError(t, err)

}

func TestModel_Count(t *testing.T) {
	// create db entries
	err := createEntries(&Role{})
	assert.NoError(t, err)

	r := Role{}

	//error because not init
	count, err := r.Count(nil)
	assert.Error(t, err)
	assert.Equal(t, 0, count)

	// init
	err = r.Init(&r)
	assert.NoError(t, err)

	count, err = r.Count(nil)
	assert.NoError(t, err)
	assert.Equal(t, 4, count)
}

func TestModel_SetRelationCondition(t *testing.T) {
	// create db entries
	err := createEntries(&Role{})
	assert.NoError(t, err)

	r := Role{}

	// init
	err = r.Init(&r)
	assert.NoError(t, err)

	// normal call
	err = r.First(sqlquery.NewCondition().Where("id = 1"))
	assert.NoError(t, err)
	assert.Equal(t, 2, len(r.Roles))

	// setting a relation Condition
	r.SetRelationCondition("Roles", *sqlquery.NewCondition().Where("id = roles.id  AND roles.Name = ? ", "User"))
	err = r.First(sqlquery.NewCondition().Where("id = 1"))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(r.Roles))
	assert.Equal(t, "WHERE id = roles.id  AND roles.Name = User ", r.RelationCondition("Roles").Config(true, sqlquery.WHERE))
}

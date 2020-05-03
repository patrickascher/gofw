package orm

import (
	"reflect"
	"testing"
	"time"

	"github.com/patrickascher/gofw/cache"
	"github.com/stretchr/testify/assert"
)

type car struct {
	Model

	ID        int
	OwnerID   NullInt
	Brand     string     `validate:"oneof=BMW BMW2"`
	Type      NullString `orm:"permission"` // reset permission
	Custom    string     `orm:"custom"`
	YearCheck NullInt    `orm:"column:year;select:Concat(id,'.',brand,year)"`
	Embedded

	// relation tests (belongsTo, m2m, hasMany)
	Owner  *owner   `orm:"relation:belongsTo;permission:r"`
	Driver []driver `orm:"relation:m2m"`

	Wheels []wheel

	// polymorphic tests (hasOne, hasMany)
	Radio  radio    `orm:"polymorphic:Car;polymorphic_value:radio"`
	Liquid []liquid `orm:"polymorphic:Car;polymorphic_value:liquid"`

	CustomDriver []*driver `orm:"custom"`
}

type Embedded struct {
	CustomOne string `orm:"custom"`
}

type owner struct {
	Model

	ID   int
	Name NullString

	// relation test hasOne
	Car *car
}

type driver struct {
	Model

	ID   int
	Name NullString
}

type wheel struct {
	Model

	ID    int
	Brand NullString
	Note  NullString

	CarID int
}
type radio struct {
	Model
	Component
}

func (r radio) DefaultTableName() string {
	return "components"
}

type liquid struct {
	Model
	Component
}

func (l liquid) DefaultTableName() string {
	return "components"
}

type Component struct {
	ID    int
	Brand NullString
	Note  NullString

	CarID   int
	CarType string
}

type ComponentNoID struct {
	Model

	ID    int
	Brand NullString
	Note  NullString

	CarType string
}

func (c ComponentNoID) DefaultTableName() string {
	return "components"
}

type ComponentNoType struct {
	Model

	ID    int
	Brand NullString
	Note  NullString

	CarID int
}

func (c ComponentNoType) DefaultTableName() string {
	return "components"
}

type Role struct {
	Model
	ID   int
	Name NullString

	Roles []*Role
}

type RoleWrongRelation struct {
	Model
	ID   int
	Name NullString

	Roles []RoleWrongRelation `orm:"relation:hasMany"`
}

func (r RoleWrongRelation) DefaultTableName() string {
	return "roles"
}

type carErrPolyM2m struct {
	Model

	ID int
	// polymorphic error
	Liquid []liquid `orm:"relation:m2m;polymorphic:Car;polymorphic_value:liquid"`
}

func (c carErrPolyM2m) DefaultTableName() string {
	return "cars"
}

type carErrPolyB2 struct {
	Model

	ID int
	// polymorphic error
	Liquid liquid `orm:"relation:belongsTo;polymorphic:Car;polymorphic_value:liquid"`
}

func (c carErrPolyB2) DefaultTableName() string {
	return "cars"
}

type carErrRelation struct {
	Model
	ID     int
	Liquid liquid `orm:"relation:m2m"`
}

func (c carErrRelation) DefaultTableName() string {
	return "cars"
}

type carBelongsToFKErr struct {
	Model
	ID      int
	OwnerID NullInt
	Owner   *owner `orm:"relation:belongsTo;fk:foo"`
}

func (c carBelongsToFKErr) DefaultTableName() string {
	return "cars"
}

type carBelongsToAFKErr struct {
	Model
	ID      int
	OwnerID NullInt
	Owner   *owner `orm:"relation:belongsTo;afk:bar"`
}

func (c carBelongsToAFKErr) DefaultTableName() string {
	return "cars"
}

type carM2MFkErr struct {
	Model
	ID     int
	Driver []driver `orm:"relation:m2m;fk:foo"`
}

func (c carM2MFkErr) DefaultTableName() string {
	return "cars"
}

type carM2MAfkErr struct {
	Model
	ID     int
	Driver []driver `orm:"relation:m2m;afk:bar"`
}

func (c carM2MAfkErr) DefaultTableName() string {
	return "cars"
}

type carM2MJoinErr struct {
	Model
	ID     int
	Driver []driver `orm:"relation:m2m;join_table:foo"`
}

func (c carM2MJoinErr) DefaultTableName() string {
	return "cars"
}

type carHasManyFkErr struct {
	Model

	ID     int
	Wheels []wheel `orm:"fk:foo"`
}

func (c carHasManyFkErr) DefaultTableName() string {
	return "cars"
}

type carHasManyAfkErr struct {
	Model

	ID     int
	Wheels []wheel `orm:"afk:bar"`
}

func (c carHasManyAfkErr) DefaultTableName() string {
	return "cars"
}

type carHasManyPolyErr struct {
	Model

	ID     int
	Wheels []wheel `orm:"polymorphic:poly"`
}

func (c carHasManyPolyErr) DefaultTableName() string {
	return "cars"
}

type carCacheErr struct {
	Model
	ID     int
	Driver []driver `orm:"relation:m2m;join_table:car_drivers;join_fk:car_id"`
}

func (c carCacheErr) DefaultTableName() string {
	return "cars"
}
func (c carCacheErr) DefaultCache() (cache.Interface, time.Duration, error) {
	return nil, 0, nil
}

func TestModel_model(t *testing.T) {
	c := car{}
	err := c.Init(&c)
	assert.NoError(t, err)
	assert.IsType(t, &Model{}, c.model())
}

func TestModel_modelName(t *testing.T) {
	c := car{}
	err := c.Init(&c)
	assert.NoError(t, err)
	assert.IsType(t, reflect.TypeOf(c).String(), c.modelName(true))
	assert.IsType(t, "Car", c.modelName(false))
}

func TestModel_strategy(t *testing.T) {
	c := car{}
	err := c.Init(&c)
	assert.NoError(t, err)
	s, err := c.strategy()
	assert.NoError(t, err)
	assert.Equal(t, "*strategy.EagerLoading", reflect.TypeOf(s).String())

}

func TestModel_addAutoTX_commitAutoTX(t *testing.T) {
	c := car{}
	err := c.Init(&c)
	assert.NoError(t, err)

	// manually added tx
	err = c.builder.Tx()
	assert.NoError(t, err)
	assert.True(t, c.builder.HasTx())
	assert.False(t, c.autoTx)
	err = c.commitAutoTX() // nothing should happen because it was no auto tx
	assert.NoError(t, err)
	assert.True(t, c.builder.HasTx())

	// auto tx should be ignored
	err = c.addAutoTX()
	assert.NoError(t, err)
	assert.True(t, c.builder.HasTx())
	assert.False(t, c.autoTx)

	// reset TX
	err = c.builder.Commit()
	assert.NoError(t, err)

	err = c.addAutoTX()
	assert.NoError(t, err)
	assert.True(t, c.builder.HasTx())
	assert.True(t, c.autoTx)
	err = c.commitAutoTX() // tx commit
	assert.False(t, c.builder.HasTx())
	assert.False(t, c.autoTx)

}

package orm2

import (
	"github.com/patrickascher/gofw/cache"
	"time"
)

type car struct {
	Model

	ID        int
	OwnerID   int
	Brand     string
	Type      string `orm:"permission"`
	Custom    string `orm:"custom"`
	YearCheck string `orm:"column:year;select:Concat(id,'.',brand,year)"`
	Embedded

	// relation tests (belongsTo, m2m, hasMany)
	Owner  owner    `orm:"relation:belongsTo"`
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
	Name string

	// relation test hasOne
	Car *car
}

type driver struct {
	Model

	ID   int
	Name string
}

type wheel struct {
	Model

	ID    int
	Brand string
	Note  string

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
	Brand string
	Note  string

	CarID   int
	CarType string
}

type ComponentNoID struct {
	Model

	ID    int
	Brand string
	Note  string

	CarType string
}

func (c ComponentNoID) DefaultTableName() string {
	return "components"
}

type ComponentNoType struct {
	Model

	ID    int
	Brand string
	Note  string

	CarID int
}

func (c ComponentNoType) DefaultTableName() string {
	return "components"
}

type Role struct {
	Model
	ID   int
	Name string

	Roles []Role
}

type RoleWrongRelation struct {
	Model
	ID   int
	Name string

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
	OwnerID int
	Owner   *owner `orm:"relation:belongsTo;fk:foo"`
}

func (c carBelongsToFKErr) DefaultTableName() string {
	return "cars"
}

type carBelongsToAFKErr struct {
	Model
	ID      int
	OwnerID int
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

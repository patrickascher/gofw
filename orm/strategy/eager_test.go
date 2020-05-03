package strategy_test

import (
	"fmt"
	"testing"

	"github.com/patrickascher/gofw/orm"
	"github.com/patrickascher/gofw/orm/strategy"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
)

// this is copied of model_test
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

	// drivers
	wheels := []map[string]interface{}{
		{"id": 1, "brand": "Goodyear", "note": "X", "car_id": 1},
		{"id": 2, "brand": "Pirelli", "note": "X", "car_id": 1},
	}
	_, err = b.Insert("wheels").Columns("id", "brand", "note", "car_id").Values(wheels).Exec()
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

type Role struct {
	orm.Model
	ID   int
	Name orm.NullString

	Roles []*Role
}
type car struct {
	orm.Model

	ID        int
	OwnerID   orm.NullInt
	Brand     string         `validate:"omitempty,oneof=BMW BMW2"`
	Type      orm.NullString `orm:"permission"` // reset permission
	Custom    string         `orm:"custom"`
	YearCheck orm.NullInt    `orm:"column:year"`
	Embedded

	// relation tests (belongsTo, m2m, hasMany)
	Owner  *owner   `orm:"relation:belongsTo"`
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
	orm.Model

	ID   int
	Name orm.NullString

	// relation test hasOne
	//Car *car // disabled because its not working on All yet.
}

type driver struct {
	orm.Model

	ID   int
	Name orm.NullString
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

func TestEagerLoading_First(t *testing.T) {
	test := assert.New(t)

	err := createEntries(&car{})
	test.NoError(err)

	c := &car{}
	err = c.Init(c)
	test.NoError(err)

	s := strategy.EagerLoading{}

	err = s.First(c.Scope(), sqlquery.NewCondition().Where("id =?", 1), orm.Permission{Read: true})
	test.NoError(err)

	// normal fields
	test.Equal(1, c.ID)
	test.Equal("BMW", c.Brand)
	test.Equal(orm.NewNullInt(1), c.OwnerID)
	test.Equal(orm.NullString{}, c.Type)
	test.Equal("", c.Custom)
	test.Equal(orm.NewNullInt(2000), c.YearCheck)

	// belongsTo
	test.Equal(1, c.Owner.ID)
	test.Equal(orm.NewNullString("John Doe"), c.Owner.Name)
	//test.Equal(c, c.Owner.Car) // back reference - disabled because its not working on All.

	// hasOne polymorphic
	test.Equal(1, c.Radio.ID)
	test.Equal(orm.NewNullString("AEG"), c.Radio.Brand)
	test.Equal("radio", c.Radio.CarType)
	test.Equal(1, c.Radio.ID)

	// hasMany
	test.Equal(2, len(c.Wheels))
	test.Equal(1, c.Wheels[0].ID)
	test.Equal(orm.NewNullString("Goodyear"), c.Wheels[0].Brand)
	test.Equal(orm.NewNullString("X"), c.Wheels[0].Note)
	test.Equal(1, c.Wheels[0].CarID)
	test.Equal(2, c.Wheels[1].ID)
	test.Equal(orm.NewNullString("Pirelli"), c.Wheels[1].Brand)
	test.Equal(orm.NewNullString("X"), c.Wheels[1].Note)
	test.Equal(1, c.Wheels[1].CarID)

	// hasMany polymorphic
	test.Equal(1, len(c.Liquid))
	test.Equal(2, c.Liquid[0].ID)
	test.Equal(orm.NewNullString("Molly"), c.Liquid[0].Brand)
	test.Equal(1, c.Liquid[0].CarID)
	test.Equal("liquid", c.Liquid[0].CarType)

	// m2m
	test.Equal(2, len(c.Driver))
	test.Equal(1, c.Driver[0].ID)
	test.Equal(orm.NewNullString("Pat"), c.Driver[0].Name)
	test.Equal(2, c.Driver[1].ID)
	test.Equal(orm.NewNullString("Tom"), c.Driver[1].Name)

	// custom field
	test.Equal(0, len(c.CustomDriver))

	// Whitelist ID & Owner (OwnerID is added automatically)
	// Need to call First on the model, because the field wb definition happens there.
	c = &car{}
	err = c.Init(c)
	test.NoError(err)
	c.SetWBList(orm.WHITELIST, "Owner")
	err = c.First(sqlquery.NewCondition().Where("id =?", 1))
	test.NoError(err)
	// normal fields
	test.Equal(1, c.ID)
	test.Equal("", c.Brand)
	test.Equal(orm.NewNullInt(1), c.OwnerID)
	test.Equal(orm.NullString{}, c.Type)
	test.Equal("", c.Custom)
	test.Equal(orm.NullInt{}, c.YearCheck)
	// belongsTo
	test.Equal(1, c.Owner.ID)
	test.Equal(orm.NewNullString("John Doe"), c.Owner.Name)
	//test.Equal(c, c.Owner.Car) // back reference
	test.Equal(0, c.Radio.ID)
	test.Equal(0, len(c.Wheels))
	test.Equal(0, len(c.Liquid))
	test.Equal(0, len(c.Driver))
	test.Equal(0, len(c.CustomDriver))

	// Blacklist all Relations ID & Owner + OwnerID (fk field of the relation)
	// Need to call First on the model, because the field wb definition happens there.
	c = &car{}
	err = c.Init(c)
	test.NoError(err)
	c.SetWBList(orm.BLACKLIST, "OwnerID", "Owner", "Driver", "Wheels", "Liquid", "Radio")
	err = c.First(sqlquery.NewCondition().Where("id =?", 1))
	test.NoError(err)
	// normal fields
	test.Equal(1, c.ID)
	test.Equal("BMW", c.Brand)
	test.Equal(orm.NullInt{}, c.OwnerID)
	test.Equal(orm.NewNullString("M3"), c.Type)
	test.Equal("", c.Custom)
	test.Equal(orm.NewNullInt(2000), c.YearCheck)
	// belongsTo
	test.Nil(c.Owner)
	test.Equal(0, c.Radio.ID)
	test.Equal(0, len(c.Wheels))
	test.Equal(0, len(c.Liquid))
	test.Equal(0, len(c.Driver))
	test.Equal(0, len(c.CustomDriver))

	// Test if there is will return an error if one (radio) relation has no entries
	c = &car{}
	err = c.Init(c)
	test.NoError(err)
	err = c.First(sqlquery.NewCondition().Where("id =?", 2))
	test.NoError(err)
	test.Equal(0, c.Radio.ID)

	err = truncate(&car{})
	test.NoError(err)
}

func TestEagerLoading_First_SelfReference(t *testing.T) {
	test := assert.New(t)

	err := createEntries(&Role{})
	test.NoError(err)

	r := &Role{}
	err = r.Init(r)
	test.NoError(err)

	s := strategy.EagerLoading{}

	err = s.First(r.Scope(), sqlquery.NewCondition().Where("id =?", 1), orm.Permission{Read: true})
	test.NoError(err)

	test.Equal(1, r.ID)
	test.Equal(orm.NewNullString("Admin"), r.Name)
	test.Equal(2, len(r.Roles))

	test.Equal(2, r.Roles[0].ID)
	test.Equal(orm.NewNullString("Writer"), r.Roles[0].Name)
	test.Equal(1, len(r.Roles[0].Roles))
	test.Equal(3, r.Roles[0].Roles[0].ID)
	test.Equal(orm.NewNullString("User"), r.Roles[0].Roles[0].Name)
	test.Equal(0, len(r.Roles[0].Roles[0].Roles))

	test.Equal(3, r.Roles[1].ID)
	test.Equal(orm.NewNullString("User"), r.Roles[1].Name)
	test.Equal(0, len(r.Roles[1].Roles))

	err = truncate(&car{})
	test.NoError(err)
}

func TestEagerLoading_All(t *testing.T) {
	test := assert.New(t)

	err := createEntries(&car{})
	test.NoError(err)

	c := &car{}
	err = c.Init(c)
	test.NoError(err)

	s := strategy.EagerLoading{}

	var res []car
	err = s.All(&res, c.Scope(), sqlquery.NewCondition())
	test.NoError(err)

	// normal fields - result set 1
	test.Equal(1, res[0].ID)
	test.Equal("BMW", res[0].Brand)
	test.Equal(orm.NewNullInt(1), res[0].OwnerID)
	test.Equal(orm.NullString{}, res[0].Type)
	test.Equal("", res[0].Custom)
	test.Equal(orm.NewNullInt(2000), res[0].YearCheck)
	// normal fields - result set 2
	test.Equal(2, res[1].ID)
	test.Equal("Mercedes", res[1].Brand)
	test.Equal(orm.NewNullInt(2), res[1].OwnerID)
	test.Equal(orm.NullString{}, res[1].Type)
	test.Equal("", res[1].Custom)
	test.Equal(orm.NewNullInt(2001), res[1].YearCheck)

	// belongsTo - result set 1
	test.Equal(1, res[0].Owner.ID)
	test.Equal(orm.NewNullString("John Doe"), res[0].Owner.Name)
	//test.Equal(c, res[0].Owner.Car) // back reference
	// belongsTo - result set 2
	test.Equal(2, res[1].Owner.ID)
	test.Equal(orm.NewNullString("Foo Bar"), res[1].Owner.Name)
	//test.Equal(c, res[0].Owner.Car) // back reference

	// hasOne polymorphic - result set 1
	test.Equal(1, res[0].Radio.ID)
	test.Equal(orm.NewNullString("AEG"), res[0].Radio.Brand)
	test.Equal("radio", res[0].Radio.CarType)
	test.Equal(1, res[0].Radio.ID)
	// hasOne polymorphic - result set 2
	test.Equal(0, res[1].Radio.ID)
	test.Equal(orm.NullString{}, res[1].Radio.Brand)
	test.Equal("", res[1].Radio.CarType)
	test.Equal(0, res[1].Radio.ID)

	// hasMany - result set 1
	test.Equal(2, len(res[0].Wheels))
	test.Equal(1, res[0].Wheels[0].ID)
	test.Equal(orm.NewNullString("Goodyear"), res[0].Wheels[0].Brand)
	test.Equal(orm.NewNullString("X"), res[0].Wheels[0].Note)
	test.Equal(1, res[0].Wheels[0].CarID)
	test.Equal(2, res[0].Wheels[1].ID)
	test.Equal(orm.NewNullString("Pirelli"), res[0].Wheels[1].Brand)
	test.Equal(orm.NewNullString("X"), res[0].Wheels[1].Note)
	test.Equal(1, res[0].Wheels[1].CarID)
	// hasMany - result set 2
	test.Equal(0, len(res[1].Wheels))

	// hasMany polymorphic - result set 1
	test.Equal(1, len(res[0].Liquid))
	test.Equal(2, res[0].Liquid[0].ID)
	test.Equal(orm.NewNullString("Molly"), res[0].Liquid[0].Brand)
	test.Equal(1, res[0].Liquid[0].CarID)
	test.Equal("liquid", res[0].Liquid[0].CarType)
	// hasMany polymorphic - result set 2
	test.Equal(1, len(res[1].Liquid))
	test.Equal(3, res[1].Liquid[0].ID)
	test.Equal(orm.NewNullString("Molly"), res[1].Liquid[0].Brand)
	test.Equal(2, res[1].Liquid[0].CarID)
	test.Equal("liquid", res[1].Liquid[0].CarType)

	// m2m - result set 1
	test.Equal(2, len(res[0].Driver))
	test.Equal(1, res[0].Driver[0].ID)
	test.Equal(orm.NewNullString("Pat"), res[0].Driver[0].Name)
	test.Equal(2, res[0].Driver[1].ID)
	test.Equal(orm.NewNullString("Tom"), res[0].Driver[1].Name)
	// m2m - result set 2
	test.Equal(1, len(res[1].Driver))
	test.Equal(3, res[1].Driver[0].ID)
	test.Equal(orm.NewNullString("Marc"), res[1].Driver[0].Name)

	// custom field
	test.Equal(0, len(res[0].CustomDriver))
	test.Equal(0, len(res[1].CustomDriver))

	// Whitelist ID & Owner (OwnerID is added automatically)
	// Need to call First on the model, because the field wb definition happens there.
	c = &car{}
	err = c.Init(c)
	test.NoError(err)
	c.SetWBList(orm.WHITELIST, "Owner")
	var res2 []car
	err = c.All(&res2, sqlquery.NewCondition())
	test.NoError(err)
	// normal fields - result set 1
	test.Equal(1, res2[0].ID)
	test.Equal("", res2[0].Brand)
	test.Equal(orm.NewNullInt(1), res2[0].OwnerID)
	test.Equal(orm.NullString{}, res2[0].Type)
	test.Equal("", res2[0].Custom)
	test.Equal(orm.NullInt{}, res2[0].YearCheck)
	// normal fields - result set 2
	test.Equal(2, res2[1].ID)
	test.Equal("", res2[1].Brand)
	test.Equal(orm.NewNullInt(2), res2[1].OwnerID)
	test.Equal(orm.NullString{}, res2[1].Type)
	test.Equal("", res2[1].Custom)
	test.Equal(orm.NullInt{}, res2[1].YearCheck)

	// belongsTo - result set 1
	test.Equal(1, res2[0].Owner.ID)
	test.Equal(orm.NewNullString("John Doe"), res2[0].Owner.Name)
	//test.Equal(c, res[0].Owner.Car) // back reference
	test.Equal(0, res2[0].Radio.ID)
	test.Equal(0, len(res2[0].Wheels))
	test.Equal(0, len(res2[0].Liquid))
	test.Equal(0, len(res2[0].Driver))
	test.Equal(0, len(res2[0].CustomDriver))
	// belongsTo - result set 2
	test.Equal(2, res2[1].Owner.ID)
	test.Equal(orm.NewNullString("Foo Bar"), res2[1].Owner.Name)
	//test.Equal(c, res[0].Owner.Car) // back reference
	test.Equal(0, res2[1].Radio.ID)
	test.Equal(0, len(res2[1].Wheels))
	test.Equal(0, len(res2[1].Liquid))
	test.Equal(0, len(res2[1].Driver))
	test.Equal(0, len(res2[1].CustomDriver))

	// Blacklist all Relations ID & Owner + OwnerID (fk field of the relation)
	// Need to call First on the model, because the field wb definition happens there.
	c = &car{}
	err = c.Init(c)
	test.NoError(err)
	c.SetWBList(orm.BLACKLIST, "OwnerID", "Owner", "Driver", "Wheels", "Liquid", "Radio")
	var res3 []car
	err = c.All(&res3, sqlquery.NewCondition())
	test.NoError(err)
	// normal fields - result set 1
	test.Equal(1, res3[0].ID)
	test.Equal("BMW", res3[0].Brand)
	test.Equal(orm.NullInt{}, res3[0].OwnerID)
	test.Equal(orm.NewNullString("M3"), res3[0].Type)
	test.Equal("", res3[0].Custom)
	test.Equal(orm.NewNullInt(2000), res3[0].YearCheck)
	// normal fields - result set 2
	test.Equal(2, res3[1].ID)
	test.Equal("Mercedes", res3[1].Brand)
	test.Equal(orm.NullInt{}, res3[1].OwnerID)
	test.Equal(orm.NewNullString("SLK"), res3[1].Type)
	test.Equal("", res3[1].Custom)
	test.Equal(orm.NewNullInt(2001), res3[1].YearCheck)
	// belongsTo - result set 1
	test.Nil(res3[0].Owner)
	test.Equal(0, res3[0].Radio.ID)
	test.Equal(0, len(res3[0].Wheels))
	test.Equal(0, len(res3[0].Liquid))
	test.Equal(0, len(res3[0].Driver))
	test.Equal(0, len(res3[0].CustomDriver))
	// belongsTo - result set 2
	test.Nil(res3[1].Owner)
	test.Equal(0, res3[1].Radio.ID)
	test.Equal(0, len(res3[1].Wheels))
	test.Equal(0, len(res3[1].Liquid))
	test.Equal(0, len(res3[1].Driver))
	test.Equal(0, len(res3[1].CustomDriver))

	// test empty result
	c = &car{}
	err = c.Init(c)
	test.NoError(err)
	var res4 []car
	err = c.All(&res4, sqlquery.NewCondition().Where("id < 0"))
	test.NoError(err)
	test.Equal(0, len(res4))
}

func TestEagerLoading_All_SelfReference(t *testing.T) {
	test := assert.New(t)

	err := createEntries(&Role{})
	test.NoError(err)

	r := &Role{}
	err = r.Init(r)
	test.NoError(err)

	s := strategy.EagerLoading{}

	var res []Role
	err = s.All(&res, r.Scope(), sqlquery.NewCondition())
	test.NoError(err)

	// normal fields - result set 1
	test.Equal(1, res[0].ID)
	test.Equal(orm.NewNullString("Admin"), res[0].Name)
	test.Equal(2, len(res[0].Roles))
	// Roles of roles
	test.Equal(2, res[0].Roles[0].ID)
	test.Equal(orm.NewNullString("Writer"), res[0].Roles[0].Name)
	test.Equal(1, len(res[0].Roles[0].Roles))
	test.Equal(3, res[0].Roles[0].Roles[0].ID)
	test.Equal(orm.NewNullString("User"), res[0].Roles[0].Roles[0].Name)
	test.Equal(0, len(res[0].Roles[0].Roles[0].Roles))
	test.Equal(3, res[0].Roles[1].ID)
	test.Equal(orm.NewNullString("User"), res[0].Roles[1].Name)
	test.Equal(0, len(res[0].Roles[1].Roles))

	// result set 2
	test.Equal(2, res[1].ID)
	test.Equal(orm.NewNullString("Writer"), res[1].Name)
	test.Equal(1, len(res[1].Roles))
	test.Equal(3, res[1].Roles[0].ID)
	test.Equal(orm.NewNullString("User"), res[1].Roles[0].Name)
	test.Equal(0, len(res[1].Roles[0].Roles))

	// result set 3
	test.Equal(3, res[2].ID)
	test.Equal(orm.NewNullString("User"), res[2].Name)
	test.Equal(0, len(res[2].Roles))

	// result set 4
	test.Equal(4, res[3].ID)
	test.Equal(orm.NewNullString("Guest"), res[3].Name)
	test.Equal(0, len(res[3].Roles))
}

func TestEagerLoading_Create(t *testing.T) {
	test := assert.New(t)

	err := truncate(&car{})
	test.NoError(err)

	c := &car{}
	err = c.Init(c)
	test.NoError(err)

	s := strategy.EagerLoading{}

	// no value is given
	err = s.Create(c.Scope())
	test.Error(err)

	// normal fields
	c.ID = 1
	c.Brand = "BMW"
	c.Type = orm.NewNullString("M3")
	c.Custom = ""
	c.YearCheck = orm.NewNullInt(2000)
	c.CustomOne = ""
	// relations
	// belongsTo
	c.Owner = &owner{Name: orm.NewNullString("Pat")}
	// M2M
	c.Driver = append(c.Driver, driver{Name: orm.NewNullString("Marc")})
	// hasMany
	c.Wheels = append(c.Wheels, wheel{Brand: orm.NewNullString("Goodyear")}, wheel{Brand: orm.NewNullString("Pirelli")}, wheel{})
	// hasOne polymorphic
	c.Radio.Brand = orm.NewNullString("AEG")
	c.Radio.Note = orm.NewNullString("Base")
	// hasMany polymorphic
	c.Liquid = append(c.Liquid, liquid{Component: Component{Brand: orm.NewNullString("Molly")}}, liquid{Component: Component{Brand: orm.NewNullString("Dot4")}})

	// Create entry
	err = s.Create(c.Scope())
	test.NoError(err)
	tmpOwnerID := c.Owner.ID

	// check entry
	c = &car{}
	err = c.Init(c)
	test.NoError(err)
	err = s.First(c.Scope(), sqlquery.NewCondition().Where("id = ?", 1), orm.Permission{Read: true})
	test.NoError(err)
	// normal fields
	test.Equal(1, c.ID)
	test.Equal("BMW", c.Brand)
	test.Equal(orm.NullString{}, c.Type) // permission was set to false in write and read.
	test.Equal("", c.Custom)
	test.Equal(orm.NewNullInt(2000), c.YearCheck)
	test.Equal("", c.CustomOne)
	test.Equal(tmpOwnerID, int(c.OwnerID.Int64))
	// relations
	// belongsTo
	test.Equal(tmpOwnerID, c.Owner.ID)
	test.Equal(orm.NewNullString("Pat"), c.Owner.Name)
	// M2M
	test.Equal(1, len(c.Driver))
	test.Equal(orm.NewNullString("Marc"), c.Driver[0].Name)
	// hasMany
	test.Equal(2, len(c.Wheels))
	test.Equal(orm.NewNullString("Goodyear"), c.Wheels[0].Brand)
	test.Equal(orm.NewNullString("Pirelli"), c.Wheels[1].Brand)
	// hasOne polymorphic
	test.Equal(orm.NewNullString("AEG"), c.Radio.Brand)
	test.Equal(orm.NewNullString("Base"), c.Radio.Note)
	// hasMany polymorphic
	test.Equal(2, len(c.Liquid))
	test.Equal(orm.NewNullString("Molly"), c.Liquid[0].Brand)
	test.Equal(orm.NewNullString("Dot4"), c.Liquid[1].Brand)

	// Create empty with existing Owner ID
	c = &car{}
	err = c.Init(c)
	test.NoError(err)
	c.Brand = "BMW2"
	c.Owner = &owner{ID: tmpOwnerID, Name: orm.NewNullString("Tom")}
	err = s.Create(c.Scope())
	test.NoError(err)

	o := &owner{}
	err = o.Init(o)
	test.NoError(err)
	err = o.First(nil)
	err = o.Init(o)
	test.Equal(c.Owner.ID, o.ID)

	err = truncate(&car{})
	test.NoError(err)
}

func TestEagerLoading_Create_SelfReference(t *testing.T) {
	test := assert.New(t)

	err := truncate(&Role{})
	test.NoError(err)

	r := &Role{}
	err = r.Init(r)
	test.NoError(err)

	s := strategy.EagerLoading{}

	// no value is given
	err = s.Create(r.Scope())
	test.Error(err)

	r.ID = 1
	r.Name = orm.NewNullString("Admin")
	r.Roles = append(r.Roles, &Role{Name: orm.NewNullString("Writer"), Roles: []*Role{{Name: orm.NewNullString("Guest")}}})
	err = s.Create(r.Scope())
	test.NoError(err)

	var res []Role
	err = s.All(&res, r.Scope(), sqlquery.NewCondition())
	test.NoError(err)
	test.Equal(3, len(res))

	err = truncate(&car{})
	test.NoError(err)
}

func TestEagerLoading_Update(t *testing.T) {
	test := assert.New(t)

	c := newCar()

	// error - no primary is set
	err := c.Update()
	test.Error(err)

	// belongsTo Delete - Change Brand and deleted all relations
	c = newCar()
	c.ID = 1
	c.Brand = "BMW2"
	err = c.Update()
	test.NoError(err)
	cCheck := getId(1)
	test.Equal("BMW2", cCheck.Brand)
	test.Nil(cCheck.Owner)            // deleted because it was not set anymore
	test.Equal(0, len(cCheck.Wheels)) // deleted because it was not set anymore
	test.Equal(0, len(cCheck.Driver)) // deleted because it was not set anymore
	test.Equal(0, cCheck.Radio.ID)    // deleted because it was not set anymore
	test.Equal(0, len(cCheck.Liquid)) // deleted because it was not set anymore

	// belongsTo Create - Change Brand and deleted all relations, aside owner was created
	c = newCar()
	c.ID = 1
	c.Brand = "BMW2"
	c.Owner = &owner{Name: orm.NewNullString("Pat2")}
	// update data
	err = c.Update()
	test.NoError(err)
	// error update twice - no changes
	err = c.Update()
	test.Error(err)
	// check data
	cCheck = getId(1)
	test.Equal("BMW2", cCheck.Brand)
	test.NotNil(cCheck.Owner)
	test.Equal(orm.Int(c.Owner.ID), orm.Int(cCheck.OwnerID))
	test.Equal(c.OwnerID, cCheck.OwnerID)
	test.Equal("Pat2", cCheck.Owner.Name.String)
	test.Equal(0, len(cCheck.Wheels)) // deleted because it was not set anymore
	test.Equal(0, len(cCheck.Driver)) // deleted because it was not set anymore
	test.Equal(0, cCheck.Radio.ID)    // deleted because it was not set anymore
	test.Equal(0, len(cCheck.Liquid)) // deleted because it was not set anymore

	// belongsTo Update - Change Brand and deleted all relations, aside owner was updated
	c = newCar()
	c.ID = 1
	c.Brand = "BMW2"
	c.Owner = &owner{ID: 1, Name: orm.NewNullString("Pat2")}
	// update data
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getId(1)
	test.Equal("BMW2", cCheck.Brand)
	test.NotNil(cCheck.Owner)
	test.Equal(orm.Int(c.Owner.ID), orm.Int(cCheck.OwnerID))
	test.Equal(c.OwnerID, cCheck.OwnerID)
	test.Equal(0, len(cCheck.Wheels)) // deleted because it was not set anymore
	test.Equal(0, len(cCheck.Driver)) // deleted because it was not set anymore
	test.Equal(0, cCheck.Radio.ID)    // deleted because it was not set anymore
	test.Equal(0, len(cCheck.Liquid)) // deleted because it was not set anymore

	// hasOne - deleted because its empty in the new struct
	c = newCar()
	c.ID = 1
	c.Brand = "BMW2"
	// update data
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getId(1)
	test.Equal("BMW2", cCheck.Brand)
	test.Nil(cCheck.Owner)
	test.Equal(0, len(cCheck.Wheels)) // deleted because it was not set anymore
	test.Equal(0, len(cCheck.Driver)) // deleted because it was not set anymore
	test.Equal(0, cCheck.Radio.ID)    // deleted because it was not set anymore
	test.Equal(0, len(cCheck.Liquid)) // deleted because it was not set anymore

	// hasOne - added
	c = newCar()
	c.ID = 1
	c.Brand = "BMW2"
	c.Radio.Brand = orm.NewNullString("AEG")
	// update data
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getId(1)
	test.Equal("BMW2", cCheck.Brand)
	test.Nil(cCheck.Owner)
	test.Equal(0, len(cCheck.Wheels)) // deleted because it was not set anymore
	test.Equal(0, len(cCheck.Driver)) // deleted because it was not set anymore
	test.Equal(c.Radio.ID, cCheck.Radio.ID)
	test.Equal(orm.NewNullString("AEG"), cCheck.Radio.Brand)
	test.Equal(0, len(cCheck.Liquid)) // deleted because it was not set anymore
	// check that the old radio got deleted and only the added one is in the database.
	r := &radio{}
	err = r.Init(r)
	test.NoError(err)
	var rRes []radio
	err = r.All(&rRes, sqlquery.NewCondition().Where("car_id = ? AND car_type = ?", c.ID, "radio"))
	test.NoError(err)
	test.Equal(1, len(rRes))

	// hasOne - update
	c = newCar()
	c.ID = 1
	c.Brand = "BMW2"
	c.Radio.ID = 1
	c.Radio.Brand = orm.NewNullString("AEG2")
	// update data
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getId(1)
	test.Equal("BMW2", cCheck.Brand)
	test.Nil(cCheck.Owner)
	test.Equal(0, len(cCheck.Wheels)) // deleted because it was not set anymore
	test.Equal(0, len(cCheck.Driver)) // deleted because it was not set anymore
	test.Equal(c.Radio.ID, cCheck.Radio.ID)
	test.Equal(orm.NewNullString("AEG2"), cCheck.Radio.Brand)
	test.Equal(0, len(cCheck.Liquid)) // deleted because it was not set anymore

	// hasMany - deleted because its empty in the new struct
	c = newCar()
	c.ID = 1
	c.Brand = "BMW2"
	// update data
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getId(1)
	test.Equal(0, len(cCheck.Wheels)) // deleted because it was not set anymore

	// hasMany - update (create one, delete the others)
	c = newCar()
	c.ID = 1
	c.Wheels = append(c.Wheels, wheel{Brand: orm.NewNullString("Goodyear")})
	// update data
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getId(1)
	test.Equal(1, len(cCheck.Wheels))                                 // deleted because it was not set anymore
	test.Equal(orm.NewNullString("Goodyear"), cCheck.Wheels[0].Brand) // deleted because it was not set anymore
	test.Equal(c.Wheels[0].ID, cCheck.Wheels[0].ID)                   // deleted because it was not set anymore

	// hasMany - update (update one, delete the others)
	c = newCar()
	c.ID = 1
	c.Wheels = append(c.Wheels, wheel{ID: 1, Brand: orm.NewNullString("Goodyear1")})
	// update data
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getId(1)
	test.Equal(1, len(cCheck.Wheels))                                  // deleted because it was not set anymore
	test.Equal(orm.NewNullString("Goodyear1"), cCheck.Wheels[0].Brand) // deleted because it was not set anymore
	test.Equal(c.Wheels[0].ID, cCheck.Wheels[0].ID)                    // deleted because it was not set anymore

	// hasMany - create
	c = newCar()
	c.ID = 1
	// delete wheels
	err = c.Update()
	test.NoError(err)
	c.Wheels = append(c.Wheels, wheel{Brand: orm.NewNullString("Goodyear")})
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getId(1)
	test.Equal(1, len(cCheck.Wheels))                                 // deleted because it was not set anymore
	test.Equal(orm.NewNullString("Goodyear"), cCheck.Wheels[0].Brand) // deleted because it was not set anymore
	test.Equal(c.Wheels[0].ID, cCheck.Wheels[0].ID)                   // deleted because it was not set anymore

	// hasMany poly - deleted because its empty in the new struct
	c = newCar()
	c.ID = 1
	c.Brand = "BMW2"
	// update data
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getId(1)
	test.Equal(0, len(cCheck.Liquid)) // deleted because it was not set anymore

	// hasMany poly - update (create one, delete the others)
	c = newCar()
	c.ID = 1
	c.Liquid = append(c.Liquid, liquid{Component: Component{Brand: orm.NewNullString("Molly")}})
	// update data
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getId(1)
	test.Equal(1, len(cCheck.Liquid))
	test.Equal(orm.NewNullString("Molly"), cCheck.Liquid[0].Brand)
	test.Equal(c.Liquid[0].ID, cCheck.Liquid[0].ID)

	// hasMany - update (update one, delete the others)
	c = newCar()
	c.ID = 1
	c.Liquid = append(c.Liquid, liquid{Component: Component{ID: 2, Brand: orm.NewNullString("Molly1")}})
	// update data
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getId(1)
	test.Equal(1, len(cCheck.Liquid))
	test.Equal(orm.NewNullString("Molly1"), cCheck.Liquid[0].Brand)
	test.Equal(c.Liquid[0].ID, cCheck.Liquid[0].ID)

	// hasMany - poly create
	c = newCar()
	c.ID = 1
	// delete wheels
	err = c.Update()
	test.NoError(err)
	c.Liquid = append(c.Liquid, liquid{Component: Component{ID: 2, Brand: orm.NewNullString("Molly1")}})
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getId(1)
	test.Equal(1, len(cCheck.Liquid))                               // deleted because it was not set anymore
	test.Equal(orm.NewNullString("Molly1"), cCheck.Liquid[0].Brand) // deleted because it was not set anymore
	test.Equal(c.Liquid[0].ID, cCheck.Liquid[0].ID)                 // deleted because it was not set anymore

	// m2m- deleted because its empty in the new struct
	c = newCar()
	c.ID = 1
	c.Brand = "BMW2"
	// update data
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getId(1)
	test.Equal(0, len(cCheck.Driver)) // deleted because it was not set anymore

	// m2m - update (create one, delete the others)
	c = newCar()
	c.ID = 1
	c.Driver = append(c.Driver, driver{Name: orm.NewNullString("Pat")})
	// update data
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getId(1)
	test.Equal(1, len(cCheck.Driver))
	test.Equal(orm.NewNullString("Pat"), cCheck.Driver[0].Name)
	test.Equal(c.Driver[0].ID, cCheck.Driver[0].ID)

	// m2m - update (update one, delete the others)
	c = newCar()
	c.ID = 1
	c.Driver = append(c.Driver, driver{ID: 1, Name: orm.NewNullString("Pat2")})
	// update data
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getId(1)
	test.Equal(1, len(cCheck.Driver))
	test.Equal(orm.NewNullString("Pat2"), cCheck.Driver[0].Name)
	test.Equal(c.Driver[0].ID, cCheck.Driver[0].ID)

	// m2m - create
	c = newCar()
	c.ID = 1
	// delete wheels
	err = c.Update()
	test.NoError(err)
	c.Driver = append(c.Driver, driver{Name: orm.NewNullString("Pat")})
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getId(1)
	test.Equal(1, len(cCheck.Driver))                           // deleted because it was not set anymore
	test.Equal(orm.NewNullString("Pat"), cCheck.Driver[0].Name) // deleted because it was not set anymore
	test.Equal(c.Driver[0].ID, cCheck.Driver[0].ID)             // deleted because it was not set anymore

	err = truncate(&car{})
	test.NoError(err)
}

func TestEagerLoading_Delete(t *testing.T) {
	test := assert.New(t)

	err := createEntries(&car{})
	test.NoError(err)

	c := &car{}
	err = c.Init(c)
	test.NoError(err)

	s := strategy.EagerLoading{}

	err = s.Delete(c.Scope(), sqlquery.NewCondition().Where("id =?", 1))
	test.NoError(err)

	err = c.First(sqlquery.NewCondition().Where("id=?", 1))
	test.Error(err) // no rows found

	err = truncate(&car{})
	test.NoError(err)
}

func newCar() *car {
	_ = createEntries(&car{})
	c := &car{}
	_ = c.Init(c)
	return c
}
func getId(id int) *car {
	c := &car{}
	_ = c.Init(c)
	_ = c.First(sqlquery.NewCondition().Where("id = ?", id))
	return c
}

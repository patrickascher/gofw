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

	// pkstrings
	_, err = b.Delete("pkstrings").Exec()
	if err != nil {
		return err
	}
	// pkstring_owners
	_, err = b.Delete("pkstring_owners").Exec()
	if err != nil {
		return err
	}
	// pkstring_ones
	_, err = b.Delete("pkstring_ones").Exec()
	if err != nil {
		return err
	}
	// pkstring_one_polies
	_, err = b.Delete("pkstring_one_polies").Exec()
	if err != nil {
		return err
	}
	// pkstring_manies
	_, err = b.Delete("pkstring_manies").Exec()
	if err != nil {
		return err
	}
	// pkstring_many_polies
	_, err = b.Delete("pkstring_many_polies").Exec()
	if err != nil {
		return err
	}
	// pkstring_m2_m_s
	_, err = b.Delete("pkstring_m2_m_s").Exec()
	if err != nil {
		return err
	}
	// pkstring_pkstring_m2_m_s
	_, err = b.Delete("pkstring_pkstring_m2_m_s").Exec()
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

	// string tables
	tmp := []map[string]interface{}{
		{"id": "ID1", "name": "Stringer", "pkstring_owner_id": "OWNER1"},
		{"id": "ID2", "name": "Stringer2", "pkstring_owner_id": "OWNER2"},
	}
	_, err = b.Insert("pkstrings").Columns("id", "name", "pkstring_owner_id").Values(tmp).Exec()
	if err != nil {
		return fmt.Errorf("pkstrings: %w", err)
	}
	tmp = []map[string]interface{}{
		{"id": "OWNER1", "name": "Owner"},
		{"id": "OWNER2", "name": "Owner2"},
	}
	_, err = b.Insert("pkstring_owners").Columns("id", "name").Values(tmp).Exec()
	if err != nil {
		return fmt.Errorf("pkstring_owners: %w", err)
	}
	tmp = []map[string]interface{}{
		{"id": "ONE1", "pkstring_id": "ID1", "name": "One"},
		{"id": "ONE2", "pkstring_id": "ID2", "name": "One2"},
	}
	_, err = b.Insert("pkstring_ones").Columns("id", "pkstring_id", "name").Values(tmp).Exec()
	if err != nil {
		return fmt.Errorf("pkstring_ones: %w", err)
	}

	tmp = []map[string]interface{}{
		{"id": "ONEPOLY1", "name": "One Poly", "poly_id": "ID1", "poly_type": "Pkstring"},
		{"id": "ONEPOLY2", "name": "One Poly2", "poly_id": "ID2", "poly_type": "Pkstring"},
	}
	_, err = b.Insert("pkstring_one_polies").Columns("id", "name", "poly_id", "poly_type").Values(tmp).Exec()
	if err != nil {
		return fmt.Errorf("pkstring_one_polies: %w", err)
	}

	tmp = []map[string]interface{}{
		{"id": "MANY1", "pkstring_id": "ID1", "name": "Many"},
		{"id": "MANY2", "pkstring_id": "ID2", "name": "Many2"},
	}
	_, err = b.Insert("pkstring_manies").Columns("id", "pkstring_id", "name").Values(tmp).Exec()
	if err != nil {
		return fmt.Errorf("pkstring_manies: %w", err)
	}

	tmp = []map[string]interface{}{
		{"id": "MANYPOLY1", "name": "Many Poly", "poly_id": "ID1", "poly_type": "Pkstring"},
		{"id": "MANYPOLY2", "name": "Many Poly2", "poly_id": "ID2", "poly_type": "Pkstring"},
	}
	_, err = b.Insert("pkstring_many_polies").Columns("id", "name", "poly_id", "poly_type").Values(tmp).Exec()
	if err != nil {
		return fmt.Errorf("pkstring_many_polies: %w", err)
	}

	tmp = []map[string]interface{}{
		{"id": "M2M1", "name": "M2m"},
		{"id": "M2M2", "name": "M2m2"},
	}
	_, err = b.Insert("pkstring_m2_m_s").Columns("id", "name").Values(tmp).Exec()
	if err != nil {
		return fmt.Errorf("pkstring_m2_m_s: %w", err)
	}

	tmp = []map[string]interface{}{
		{"pkstring_id": "ID1", "pkstring_m2_m_id": "M2M1"},
		{"pkstring_id": "ID2", "pkstring_m2_m_id": "M2M2"},
	}
	_, err = b.Insert("pkstring_pkstring_m2_m_s").Columns("pkstring_id", "pkstring_m2_m_id").Values(tmp).Exec()
	if err != nil {
		return fmt.Errorf("pkstring_pkstring_m2_m_s: %w", err)
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

// Test a new entry but the BelongsTo/M2M relation has an id which exists/not exist in the database.
// in both cases the element should be created/updated.
func TestEagerLoading_Create_BelongsTo_M2M_Update(t *testing.T) {
	test := assert.New(t)

	err := truncate(&car{})
	test.NoError(err)

	c := &car{}
	err = c.Init(c)
	test.NoError(err)

	s := strategy.EagerLoading{}

	// TEST ID does not exist in the db yet , belongsTo, M2M
	// normal fields
	c.ID = 1
	c.Brand = "BMW"
	// belongsTo
	c.Owner = &owner{ID: 1, Name: orm.NewNullString("Pat")} // ID does not exist in the DB, so it should be created
	// M2M
	c.Driver = append(c.Driver, driver{ID: 1, Name: orm.NewNullString("Marc")}) // ID does not exist in the DB, so it should be created
	err = s.Create(c.Scope())
	test.NoError(err)

	err = c.First(sqlquery.NewCondition().Where("id = ?", 1))
	test.NoError(err)

	test.Equal(1, c.ID)
	test.Equal("BMW", c.Brand)
	test.Equal(1, c.Owner.ID)
	test.Equal("Pat", c.Owner.Name.String)
	if test.Equal(1, len(c.Driver)) {
		test.Equal(1, c.Driver[0].ID)
		test.Equal("Marc", c.Driver[0].Name.String)
	}

	// TEST UPDATE of BelongsTo and M2M
	// normal fields
	c = &car{}
	err = c.Init(c)
	test.NoError(err)

	c.ID = 2
	c.Brand = "BMW"
	// relations
	// belongsTo
	c.Owner = &owner{ID: 1, Name: orm.NewNullString("Pat2")} // ID does not exist in the DB, so it should be created
	// M2M
	c.Driver = append(c.Driver, driver{ID: 1, Name: orm.NewNullString("Marc2")}) // ID does not exist in the DB, so it should be created
	err = s.Create(c.Scope())
	test.NoError(err)

	err = c.First(sqlquery.NewCondition().Where("id = ?", 2))
	test.NoError(err)

	test.Equal(2, c.ID)
	test.Equal("BMW", c.Brand)
	test.Equal(1, c.Owner.ID)
	test.Equal("Pat2", c.Owner.Name.String)
	if test.Equal(1, len(c.Driver)) {
		test.Equal(1, c.Driver[0].ID)
		test.Equal("Marc2", c.Driver[0].Name.String)
	}

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
	// no error update twice - will return nil
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getId(1)
	test.Equal("BMW2", cCheck.Brand)
	test.NotNil(cCheck.Owner)
	test.Equal(int64(c.Owner.ID), cCheck.OwnerID.Int64)
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
	test.Equal(int64(c.Owner.ID), cCheck.OwnerID.Int64)
	test.Equal(c.OwnerID, cCheck.OwnerID)
	test.Equal(0, len(cCheck.Wheels)) // deleted because it was not set anymore
	test.Equal(0, len(cCheck.Driver)) // deleted because it was not set anymore
	test.Equal(0, cCheck.Radio.ID)    // deleted because it was not set anymore
	test.Equal(0, len(cCheck.Liquid)) // deleted because it was not set anymore

	// Testing New ID which does not exist in the DB YET - BelongsTO and M2M
	c = newCar()
	c.ID = 1
	c.Brand = "BMW2"
	// belongsTo - new ID which does not exist in the DB yet
	c.Owner = &owner{ID: 10, Name: orm.NewNullString("New")}
	// m2m - ID DOES NOT EXIST in the db yet.
	c.Driver = append(c.Driver, driver{ID: 10, Name: orm.NewNullString("Driver")})
	err = c.Update()
	test.NoError(err)
	cCheck = getId(1)
	test.Equal("BMW2", cCheck.Brand)
	test.NotNil(cCheck.Owner)
	test.Equal(int64(10), cCheck.OwnerID.Int64)
	test.Equal(10, cCheck.Owner.ID)
	test.Equal("New", cCheck.Owner.Name.String)
	test.Equal(1, len(cCheck.Driver))
	test.Equal(10, cCheck.Driver[0].ID)
	test.Equal("Driver", cCheck.Driver[0].Name.String)

	// test update ID which does already exist in the DB, BelongsTO and M2M
	c = newCar()
	c.ID = 1
	c.Brand = "BMW"
	// belongsTo - update ID which does exist in the DB yet
	c.Owner = &owner{ID: 1, Name: orm.NewNullString("New22")}
	// m2m - ID DOES EXIST in the db yet.
	c.Driver = append(c.Driver, driver{ID: 1, Name: orm.NewNullString("Driver2")})
	err = c.Update()
	test.NoError(err)

	cCheck = getId(1)
	test.Equal("BMW", cCheck.Brand)
	test.NotNil(cCheck.Owner)
	test.Equal(int64(1), cCheck.OwnerID.Int64)
	test.Equal(1, cCheck.Owner.ID)
	test.Equal("New22", cCheck.Owner.Name.String)
	test.Equal(1, len(cCheck.Driver))
	test.Equal(1, cCheck.Driver[0].ID)
	test.Equal("Driver2", cCheck.Driver[0].Name.String)

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

func newStringPrimary() *Pkstring {
	_ = createEntries(&Pkstring{})
	c := &Pkstring{}
	_ = c.Init(c)
	return c
}
func getStringPrimaryId(id string) *Pkstring {
	c := &Pkstring{}
	_ = c.Init(c)
	_ = c.First(sqlquery.NewCondition().Where("id = ?", id))
	return c
}

type Pkstring struct {
	orm.Model
	ID              string
	Name            string
	PkstringOwnerID orm.NullString

	// belongsTO
	Owner PkstringOwner `orm:"relation:belongsTo"`
	// hasOne
	One PkstringOne
	// hasOne poly
	OnePoly PkstringOnePoly `orm:"polymorphic:Poly;"`
	// hasMany
	Many []PkstringMany
	// hasMany poly
	ManyPoly []PkstringManyPoly `orm:"polymorphic:Poly;"`
	// m2m
	M2M []PkstringM2M `orm:"relation:m2m"`
}
type PkstringOwner struct {
	orm.Model
	ID   string
	Name string
}
type PkstringOne struct {
	orm.Model
	ID         string
	PkstringID string
	Name       string
}
type PkstringOnePoly struct {
	orm.Model
	ID   string
	Name string

	PolyID   string
	PolyType string
}
type PkstringMany struct {
	orm.Model
	ID         string
	PkstringID string
	Name       string
}

type PkstringManyPoly struct {
	orm.Model
	ID   string
	Name string

	PolyID   string
	PolyType string
}

type PkstringM2M struct {
	orm.Model
	ID   string
	Name string
}

// test with primary key string
func TestEagerLoading_First_PrimaryString(t *testing.T) {
	test := assert.New(t)

	err := createEntries(&Pkstring{})
	test.NoError(err)

	m := Pkstring{}
	err = m.Init(&m)
	test.NoError(err)

	err = m.First(nil)

	test.Equal("ID1", m.ID)
	test.Equal("Stringer", m.Name)
	// belongsTo
	test.Equal("OWNER1", m.PkstringOwnerID.String)
	test.Equal("OWNER1", m.Owner.ID)
	test.Equal("Owner", m.Owner.Name)
	// has One
	test.Equal("ONE1", m.One.ID)
	test.Equal("ID1", m.One.PkstringID)
	test.Equal("One", m.One.Name)
	// has One poly
	test.Equal("ONEPOLY1", m.OnePoly.ID)
	test.Equal("One Poly", m.OnePoly.Name)
	test.Equal("ID1", m.OnePoly.PolyID)
	test.Equal("Pkstring", m.OnePoly.PolyType)
	// has many
	test.Equal(1, len(m.Many))
	test.Equal("MANY1", m.Many[0].ID)
	test.Equal("ID1", m.Many[0].PkstringID)
	test.Equal("Many", m.Many[0].Name)
	// has many poly
	test.Equal(1, len(m.ManyPoly))
	test.Equal("MANYPOLY1", m.ManyPoly[0].ID)
	test.Equal("Many Poly", m.ManyPoly[0].Name)
	test.Equal("ID1", m.ManyPoly[0].PolyID)
	test.Equal("Pkstring", m.ManyPoly[0].PolyType)
	// m2m
	test.Equal(1, len(m.M2M))
	test.Equal("M2M1", m.M2M[0].ID)
	test.Equal("M2m", m.M2M[0].Name)

	err = truncate(&Pkstring{})
	test.NoError(err)
}

// test with primary key string
func TestEagerLoading_All_PrimaryString(t *testing.T) {
	test := assert.New(t)

	err := createEntries(&Pkstring{})
	test.NoError(err)

	m := Pkstring{}
	err = m.Init(&m)
	test.NoError(err)

	var res []Pkstring
	err = m.All(&res, nil)
	test.NoError(err)

	test.Equal("ID1", res[0].ID)
	test.Equal("Stringer", res[0].Name)
	// belongsTo
	test.Equal("OWNER1", res[0].PkstringOwnerID.String)
	test.Equal("OWNER1", res[0].Owner.ID)
	test.Equal("Owner", res[0].Owner.Name)
	// has One
	test.Equal("ONE1", res[0].One.ID)
	test.Equal("ID1", res[0].One.PkstringID)
	test.Equal("One", res[0].One.Name)
	// has One poly
	test.Equal("ONEPOLY1", res[0].OnePoly.ID)
	test.Equal("One Poly", res[0].OnePoly.Name)
	test.Equal("ID1", res[0].OnePoly.PolyID)
	test.Equal("Pkstring", res[0].OnePoly.PolyType)
	// has many
	test.Equal(1, len(res[0].Many))
	test.Equal("MANY1", res[0].Many[0].ID)
	test.Equal("ID1", res[0].Many[0].PkstringID)
	test.Equal("Many", res[0].Many[0].Name)
	// has many poly
	test.Equal(1, len(res[0].ManyPoly))
	test.Equal("MANYPOLY1", res[0].ManyPoly[0].ID)
	test.Equal("Many Poly", res[0].ManyPoly[0].Name)
	test.Equal("ID1", res[0].ManyPoly[0].PolyID)
	test.Equal("Pkstring", res[0].ManyPoly[0].PolyType)
	// m2m
	test.Equal(1, len(res[0].M2M))
	test.Equal("M2M1", res[0].M2M[0].ID)
	test.Equal("M2m", res[0].M2M[0].Name)

	// RESULTSET 2
	test.Equal("ID2", res[1].ID)
	test.Equal("Stringer2", res[1].Name)
	// belongsTo
	test.Equal("OWNER2", res[1].PkstringOwnerID.String)
	test.Equal("OWNER2", res[1].Owner.ID)
	test.Equal("Owner2", res[1].Owner.Name)
	// has One
	test.Equal("ONE2", res[1].One.ID)
	test.Equal("ID2", res[1].One.PkstringID)
	test.Equal("One2", res[1].One.Name)
	// has One poly
	test.Equal("ONEPOLY2", res[1].OnePoly.ID)
	test.Equal("One Poly2", res[1].OnePoly.Name)
	test.Equal("ID2", res[1].OnePoly.PolyID)
	test.Equal("Pkstring", res[1].OnePoly.PolyType)
	// has many
	test.Equal(1, len(res[1].Many))
	test.Equal("MANY2", res[1].Many[0].ID)
	test.Equal("ID2", res[1].Many[0].PkstringID)
	test.Equal("Many2", res[1].Many[0].Name)
	// has many poly
	test.Equal(1, len(res[1].ManyPoly))
	test.Equal("MANYPOLY2", res[1].ManyPoly[0].ID)
	test.Equal("Many Poly2", res[1].ManyPoly[0].Name)
	test.Equal("ID2", res[1].ManyPoly[0].PolyID)
	test.Equal("Pkstring", res[1].ManyPoly[0].PolyType)
	// m2m
	test.Equal(1, len(res[1].M2M))
	test.Equal("M2M2", res[1].M2M[0].ID)
	test.Equal("M2m2", res[1].M2M[0].Name)

	err = truncate(&Pkstring{})
	test.NoError(err)
}

// test with primary key string
func TestEagerLoading_Create_PrimaryString(t *testing.T) {
	test := assert.New(t)

	err := truncate(&Pkstring{})
	test.NoError(err)

	m := Pkstring{}
	err = m.Init(&m)
	test.NoError(err)

	m.Name = "CreateString"
	m.ID = "CUSTID"
	//belongsTo
	m.Owner = PkstringOwner{ID: "CUSTOWNERID", Name: "CustOwner"}
	//hasOne
	m.One.ID = "CUSTONE"
	m.One.Name = "CustOne"
	//hasOne poly
	m.OnePoly.ID = "CUSTONEPOLY"
	m.OnePoly.Name = "CustOnePoly"
	//hasMany
	m.Many = append(m.Many, PkstringMany{ID: "CUSTMANY", Name: "CustMany"})
	//hasManyPoly
	m.ManyPoly = append(m.ManyPoly, PkstringManyPoly{ID: "CUSTMANYPOLY", Name: "CustManyPoly"})
	//m2m
	m.M2M = append(m.M2M, PkstringM2M{ID: "CUSTM2M", Name: "CustM2m"})

	err = m.Create()
	test.NoError(err)

	m = Pkstring{}
	err = m.Init(&m)
	test.NoError(err)
	err = m.First(sqlquery.NewCondition().Where("id = ?", "CUSTID"))
	test.NoError(err)

	test.Equal("CUSTID", m.ID)
	test.Equal("CreateString", m.Name)
	// belongsTo
	test.Equal("CUSTOWNERID", m.PkstringOwnerID.String)
	test.Equal("CUSTOWNERID", m.Owner.ID)
	test.Equal("CustOwner", m.Owner.Name)
	// has One
	test.Equal("CUSTONE", m.One.ID)
	test.Equal("CUSTID", m.One.PkstringID)
	test.Equal("CustOne", m.One.Name)
	// has One poly
	test.Equal("CUSTONEPOLY", m.OnePoly.ID)
	test.Equal("CustOnePoly", m.OnePoly.Name)
	test.Equal("CUSTID", m.OnePoly.PolyID)
	test.Equal("Pkstring", m.OnePoly.PolyType)
	// has many
	test.Equal(1, len(m.Many))
	test.Equal("CUSTMANY", m.Many[0].ID)
	test.Equal("CUSTID", m.Many[0].PkstringID)
	test.Equal("CustMany", m.Many[0].Name)
	// has many poly
	test.Equal(1, len(m.ManyPoly))
	test.Equal("CUSTMANYPOLY", m.ManyPoly[0].ID)
	test.Equal("CustManyPoly", m.ManyPoly[0].Name)
	test.Equal("CUSTID", m.ManyPoly[0].PolyID)
	test.Equal("Pkstring", m.ManyPoly[0].PolyType)
	// m2m
	test.Equal(1, len(m.M2M))
	test.Equal("CUSTM2M", m.M2M[0].ID)
	test.Equal("CustM2m", m.M2M[0].Name)

	err = truncate(&Pkstring{})
	test.NoError(err)
}

// test with primary key string
func TestEagerLoading_Update_PrimaryString(t *testing.T) {
	test := assert.New(t)

	c := newStringPrimary()

	// error - no primary is set
	err := c.Update()
	test.Error(err)

	// belongsTo Delete - Change Brand and deleted all relations
	c = newStringPrimary()
	c.ID = "ID1"
	c.Name = "CustId2"
	err = c.Update()
	test.NoError(err)
	cCheck := getStringPrimaryId("ID1")
	test.Equal("CustId2", cCheck.Name)
	test.Equal("", cCheck.Owner.ID)
	test.Equal(0, len(cCheck.Many))     // deleted because it was not set anymore
	test.Equal(0, len(cCheck.ManyPoly)) // deleted because it was not set anymore
	test.Equal("", cCheck.One.ID)       // deleted because it was not set anymore
	test.Equal("", cCheck.OnePoly.ID)   // deleted because it was not set anymore
	test.Equal(0, len(cCheck.M2M))      // deleted because it was not set anymore

	// belongsTo Create - Change Brand and deleted all relations, aside owner was created
	c = newStringPrimary()
	c.ID = "ID1"
	c.Name = "CustId2"
	c.Owner = PkstringOwner{ID: "CUSTOWNERID", Name: "CustOwner"}
	// update data
	err = c.Update()
	test.NoError(err)
	// no error update twice - will return nil
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getStringPrimaryId("ID1")
	test.Equal("CustId2", cCheck.Name)
	test.NotNil(cCheck.Owner)
	test.Equal(orm.NewNullString("CUSTOWNERID"), cCheck.PkstringOwnerID)
	test.Equal(c.PkstringOwnerID, cCheck.PkstringOwnerID)
	test.Equal("CustOwner", cCheck.Owner.Name)
	test.Equal(0, len(cCheck.Many))     // deleted because it was not set anymore
	test.Equal(0, len(cCheck.ManyPoly)) // deleted because it was not set anymore
	test.Equal("", cCheck.One.ID)       // deleted because it was not set anymore
	test.Equal("", cCheck.OnePoly.ID)   // deleted because it was not set anymore
	test.Equal(0, len(cCheck.M2M))      // deleted because it was not set anymore

	// belongsTo Update - Change Brand and deleted all relations, aside owner was updated
	c = newStringPrimary()
	c.ID = "ID1"
	c.Name = "CustId2"
	c.Owner = PkstringOwner{ID: "CUSTOWNERID", Name: "CustOwner2"}
	// update data
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getStringPrimaryId("ID1")
	test.Equal("CustId2", cCheck.Name)
	test.NotNil(cCheck.Owner)
	test.Equal(orm.NewNullString("CUSTOWNERID"), cCheck.PkstringOwnerID)
	test.Equal(c.PkstringOwnerID, cCheck.PkstringOwnerID)
	test.Equal("CustOwner2", cCheck.Owner.Name)
	test.Equal(0, len(cCheck.Many))     // deleted because it was not set anymore
	test.Equal(0, len(cCheck.ManyPoly)) // deleted because it was not set anymore
	test.Equal("", cCheck.One.ID)       // deleted because it was not set anymore
	test.Equal("", cCheck.OnePoly.ID)   // deleted because it was not set anymore
	test.Equal(0, len(cCheck.M2M))      // deleted because it was not set anymore

	// Testing New ID which does not exist in the DB YET - BelongsTO and M2M
	c = newStringPrimary()
	c.ID = "ID1"
	c.Name = "CustId2"
	// belongsTo - new ID which does not exist in the DB yet
	c.Owner = PkstringOwner{ID: "CUSTOWNERIDNEW", Name: "CustOwner2"}
	// m2m - ID DOES NOT EXIST in the db yet.
	c.M2M = append(c.M2M, PkstringM2M{ID: "M2MNEW", Name: "M2m"})
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getStringPrimaryId("ID1")
	test.Equal("CustId2", cCheck.Name)
	test.NotNil(cCheck.Owner)
	test.Equal(orm.NewNullString("CUSTOWNERIDNEW"), cCheck.PkstringOwnerID)
	test.Equal(c.PkstringOwnerID, cCheck.PkstringOwnerID)
	test.Equal("CustOwner2", cCheck.Owner.Name)
	test.Equal(0, len(cCheck.Many))     // deleted because it was not set anymore
	test.Equal(0, len(cCheck.ManyPoly)) // deleted because it was not set anymore
	test.Equal("", cCheck.One.ID)       // deleted because it was not set anymore
	test.Equal("", cCheck.OnePoly.ID)   // deleted because it was not set anymore
	test.Equal(1, len(cCheck.M2M))
	test.Equal("M2MNEW", cCheck.M2M[0].ID)
	test.Equal("M2m", cCheck.M2M[0].Name)

	// test update ID which does already exist in the DB, BelongsTO and M2M
	c = newStringPrimary()
	c.ID = "ID1"
	c.Name = "CustId2"
	// belongsTo - new ID which does not exist in the DB yet
	c.Owner = PkstringOwner{ID: "OWNER1", Name: "CustOwner2"}
	// m2m - ID DOES NOT EXIST in the db yet.
	c.M2M = append(c.M2M, PkstringM2M{ID: "M2M1", Name: "M2m2"})
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getStringPrimaryId("ID1")
	test.Equal("CustId2", cCheck.Name)
	test.NotNil(cCheck.Owner)
	test.Equal(orm.NewNullString("OWNER1"), cCheck.PkstringOwnerID)
	test.Equal(c.PkstringOwnerID, cCheck.PkstringOwnerID)
	test.Equal("CustOwner2", cCheck.Owner.Name)
	test.Equal(0, len(cCheck.Many))     // deleted because it was not set anymore
	test.Equal(0, len(cCheck.ManyPoly)) // deleted because it was not set anymore
	test.Equal("", cCheck.One.ID)       // deleted because it was not set anymore
	test.Equal("", cCheck.OnePoly.ID)   // deleted because it was not set anymore
	test.Equal(1, len(cCheck.M2M))
	test.Equal("M2M1", cCheck.M2M[0].ID)
	test.Equal("M2m2", cCheck.M2M[0].Name)

	// hasOne - added
	c = newStringPrimary()
	c.ID = "ID1"
	c.Name = "CustId2"
	c.One.ID = "ONE1"
	c.One.Name = "one1"
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getStringPrimaryId("ID1")
	test.Equal("CustId2", cCheck.Name)
	test.Equal("", cCheck.Owner.ID)     // deleted because it was not set anymore
	test.Equal(0, len(cCheck.Many))     // deleted because it was not set anymore
	test.Equal(0, len(cCheck.ManyPoly)) // deleted because it was not set anymore
	test.Equal("ONE1", cCheck.One.ID)
	test.Equal("one1", cCheck.One.Name)
	test.Equal("", cCheck.OnePoly.ID) // deleted because it was not set anymore
	test.Equal(0, len(cCheck.M2M))    // deleted because it was not set anymore
	// check that the old radio got deleted and only the added one is in the database.
	r := &PkstringOne{}
	err = r.Init(r)
	test.NoError(err)
	var rRes []PkstringOne
	err = r.All(&rRes, sqlquery.NewCondition().Where("pkstring_id = ?", c.ID))
	test.NoError(err)
	test.Equal(1, len(rRes))

	// hasOne - update
	c = newStringPrimary()
	c.ID = "ID1"
	c.Name = "CustId2"
	c.One.ID = "ONE1"
	c.One.Name = "OneUpdate"
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getStringPrimaryId("ID1")
	test.Equal("CustId2", cCheck.Name)
	test.Equal("", cCheck.Owner.ID)     // deleted because it was not set anymore
	test.Equal(0, len(cCheck.Many))     // deleted because it was not set anymore
	test.Equal(0, len(cCheck.ManyPoly)) // deleted because it was not set anymore
	test.Equal("ONE1", cCheck.One.ID)
	test.Equal("OneUpdate", cCheck.One.Name)
	test.Equal("", cCheck.OnePoly.ID) // deleted because it was not set anymore
	test.Equal(0, len(cCheck.M2M))    // deleted because it was not set anymore

	// hasMany - update (create one, delete the others)
	c = newStringPrimary()
	c.ID = "ID1"
	c.Name = "CustId2"
	c.Many = append(c.Many, PkstringMany{ID: "MANY3", Name: "many3"})
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getStringPrimaryId("ID1")
	test.Equal("CustId2", cCheck.Name)
	test.Equal("", cCheck.Owner.ID) // deleted because it was not set anymore
	test.Equal(1, len(cCheck.Many))
	test.Equal("MANY3", cCheck.Many[0].ID)
	test.Equal("many3", cCheck.Many[0].Name)
	test.Equal(0, len(cCheck.ManyPoly)) // deleted because it was not set anymore
	test.Equal("", cCheck.One.ID)       // deleted because it was not set anymore
	test.Equal("", cCheck.OnePoly.ID)   // deleted because it was not set anymore
	test.Equal(0, len(cCheck.M2M))      // deleted because it was not set anymore

	// hasMany - update (update one, delete the others)
	c = newStringPrimary()
	c.ID = "ID1"
	c.Name = "CustId2"
	c.Many = append(c.Many, PkstringMany{ID: "MANY1", Name: "many3"})
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getStringPrimaryId("ID1")
	test.Equal("CustId2", cCheck.Name)
	test.Equal("", cCheck.Owner.ID) // deleted because it was not set anymore
	test.Equal(1, len(cCheck.Many))
	test.Equal("MANY1", cCheck.Many[0].ID)
	test.Equal("many3", cCheck.Many[0].Name)
	test.Equal(0, len(cCheck.ManyPoly)) // deleted because it was not set anymore
	test.Equal("", cCheck.One.ID)       // deleted because it was not set anymore
	test.Equal("", cCheck.OnePoly.ID)   // deleted because it was not set anymore
	test.Equal(0, len(cCheck.M2M))      // deleted because it was not set anymore

	// hasMany - create
	c = newStringPrimary()
	c.ID = "ID1"
	c.Name = "CustId2"
	err = c.Update()
	test.NoError(err)
	c.Many = append(c.Many, PkstringMany{ID: "MANY1", Name: "many3"})
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getStringPrimaryId("ID1")
	test.Equal("CustId2", cCheck.Name)
	test.Equal("", cCheck.Owner.ID) // deleted because it was not set anymore
	test.Equal(1, len(cCheck.Many))
	test.Equal("MANY1", cCheck.Many[0].ID)
	test.Equal("many3", cCheck.Many[0].Name)
	test.Equal(0, len(cCheck.ManyPoly)) // deleted because it was not set anymore
	test.Equal("", cCheck.One.ID)       // deleted because it was not set anymore
	test.Equal("", cCheck.OnePoly.ID)   // deleted because it was not set anymore
	test.Equal(0, len(cCheck.M2M))      // deleted because it was not set anymore

	// hasMany poly - update (create one, delete the others)
	c = newStringPrimary()
	c.ID = "ID1"
	c.Name = "CustId2"
	c.ManyPoly = append(c.ManyPoly, PkstringManyPoly{ID: "MANYPoly3", Name: "manypoly3"})
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getStringPrimaryId("ID1")
	test.Equal("CustId2", cCheck.Name)
	test.Equal("", cCheck.Owner.ID) // deleted because it was not set anymore
	test.Equal(0, len(cCheck.Many)) // deleted because it was not set anymore
	test.Equal(1, len(cCheck.ManyPoly))
	test.Equal("MANYPoly3", cCheck.ManyPoly[0].ID)
	test.Equal("manypoly3", cCheck.ManyPoly[0].Name)
	test.Equal("", cCheck.One.ID)     // deleted because it was not set anymore
	test.Equal("", cCheck.OnePoly.ID) // deleted because it was not set anymore
	test.Equal(0, len(cCheck.M2M))    // deleted because it was not set anymore

	// hasMany - update (update one, delete the others)
	c = newStringPrimary()
	c.ID = "ID1"
	c.Name = "CustId2"
	c.ManyPoly = append(c.ManyPoly, PkstringManyPoly{ID: "MANYPOLY1", Name: "manypoly3"})
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getStringPrimaryId("ID1")
	test.Equal("CustId2", cCheck.Name)
	test.Equal("", cCheck.Owner.ID)
	test.Equal(0, len(cCheck.Many)) // deleted because it was not set anymore
	test.Equal(1, len(cCheck.ManyPoly))
	test.Equal("MANYPOLY1", cCheck.ManyPoly[0].ID)
	test.Equal("manypoly3", cCheck.ManyPoly[0].Name)
	test.Equal("", cCheck.One.ID)     // deleted because it was not set anymore
	test.Equal("", cCheck.OnePoly.ID) // deleted because it was not set anymore
	test.Equal(0, len(cCheck.M2M))    // deleted because it was not set anymore

	// hasMany - poly create
	c = newStringPrimary()
	c.ID = "ID1"
	c.Name = "CustId2"
	err = c.Update()
	test.NoError(err)
	c.ManyPoly = append(c.ManyPoly, PkstringManyPoly{ID: "MANYPOLY3", Name: "manypoly3"})
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getStringPrimaryId("ID1")
	test.Equal("CustId2", cCheck.Name)
	test.Equal("", cCheck.Owner.ID) // deleted because it was not set anymore
	test.Equal(0, len(cCheck.Many)) // deleted because it was not set anymore
	test.Equal(1, len(cCheck.ManyPoly))
	test.Equal("MANYPOLY3", cCheck.ManyPoly[0].ID)
	test.Equal("manypoly3", cCheck.ManyPoly[0].Name)
	test.Equal("", cCheck.One.ID)     // deleted because it was not set anymore
	test.Equal("", cCheck.OnePoly.ID) // deleted because it was not set anymore
	test.Equal(0, len(cCheck.M2M))    // deleted because it was not set anymore

	// m2m - update (create one, delete the others)
	c = newStringPrimary()
	c.ID = "ID1"
	c.Name = "CustId2"
	c.M2M = append(c.M2M, PkstringM2M{ID: "M2M3", Name: "m2m3"})
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getStringPrimaryId("ID1")
	test.Equal("CustId2", cCheck.Name)
	test.Equal("", cCheck.Owner.ID) // deleted because it was not set anymore
	test.Equal(0, len(cCheck.Many)) // deleted because it was not set anymore
	test.Equal(0, len(cCheck.ManyPoly))
	test.Equal("", cCheck.One.ID)     // deleted because it was not set anymore
	test.Equal("", cCheck.OnePoly.ID) // deleted because it was not set anymore
	test.Equal(1, len(cCheck.M2M))    // deleted because it was not set anymore
	test.Equal("M2M3", cCheck.M2M[0].ID)
	test.Equal("m2m3", cCheck.M2M[0].Name)

	// m2m - update (update one, delete the others)
	c = newStringPrimary()
	c.ID = "ID1"
	c.Name = "CustId2"
	c.M2M = append(c.M2M, PkstringM2M{ID: "M2M1", Name: "m2m3"})
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getStringPrimaryId("ID1")
	test.Equal("CustId2", cCheck.Name)
	test.Equal("", cCheck.Owner.ID) // deleted because it was not set anymore
	test.Equal(0, len(cCheck.Many)) // deleted because it was not set anymore
	test.Equal(0, len(cCheck.ManyPoly))
	test.Equal("", cCheck.One.ID)     // deleted because it was not set anymore
	test.Equal("", cCheck.OnePoly.ID) // deleted because it was not set anymore
	test.Equal(1, len(cCheck.M2M))    // deleted because it was not set anymore
	test.Equal("M2M1", cCheck.M2M[0].ID)
	test.Equal("m2m3", cCheck.M2M[0].Name)

	// m2m - create
	c = newStringPrimary()
	c.ID = "ID1"
	c.Name = "CustId2"
	err = c.Update()
	test.NoError(err)
	c.M2M = append(c.M2M, PkstringM2M{ID: "M2M3", Name: "m2m3"})
	err = c.Update()
	test.NoError(err)
	// check data
	cCheck = getStringPrimaryId("ID1")
	test.Equal("CustId2", cCheck.Name)
	test.Equal("", cCheck.Owner.ID) // deleted because it was not set anymore
	test.Equal(0, len(cCheck.Many)) // deleted because it was not set anymore
	test.Equal(0, len(cCheck.ManyPoly))
	test.Equal("", cCheck.One.ID)     // deleted because it was not set anymore
	test.Equal("", cCheck.OnePoly.ID) // deleted because it was not set anymore
	test.Equal(1, len(cCheck.M2M))
	test.Equal("M2M3", cCheck.M2M[0].ID)
	test.Equal("m2m3", cCheck.M2M[0].Name)

	err = truncate(&car{})
	test.NoError(err)
}

func TestEagerLoading_Delete_PrimaryString(t *testing.T) {
	test := assert.New(t)

	err := createEntries(&Pkstring{})
	test.NoError(err)

	c := &Pkstring{}
	err = c.Init(c)
	test.NoError(err)

	s := strategy.EagerLoading{}

	err = s.Delete(c.Scope(), sqlquery.NewCondition().Where("id =?", "ID1"))
	test.NoError(err)

	err = c.First(sqlquery.NewCondition().Where("id=?", "ID1"))
	test.Error(err) // no rows found

	err = truncate(&Pkstring{})
	test.NoError(err)
}

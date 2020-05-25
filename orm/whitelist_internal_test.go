package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewWBList testing if the given policy and fields are added correctly
func TestNewWBList(t *testing.T) {
	test := assert.New(t)

	// whitelist factory
	w := newWBList(WHITELIST, []string{"ID", "Name"})
	test.Equal(WHITELIST, w.policy)
	test.Equal([]string{"ID", "Name"}, w.fields)

	// blacklist factory
	b := newWBList(BLACKLIST, []string{"ID", "Name"})
	test.Equal(BLACKLIST, b.policy)
	test.Equal([]string{"ID", "Name"}, b.fields)
}

func Test_setFieldPermission(t *testing.T) {
	test := assert.New(t)

	c := &car{}
	err := c.Init(c)
	if test.NoError(err) {

		// testing if wb list is still nil
		err = c.scope.setFieldPermission()
		if test.NoError(err) {
			test.Nil(c.wbList)

			// user added whitelist
			c.SetWBList(WHITELIST, "Brand", "Owner.Name", "Liquid", "Liquid.Brand", "Liquid.ID", "does not exist")
			err = c.scope.setFieldPermission()
			if test.NoError(err) {
				// Liquid.Brand and Liquid.ID got removed because the whole relation was added.
				test.Equal(&whiteBlackList{policy: 1, fields: []string{"Brand", "Owner.Name", "Liquid", "does not exist", "ID", "CreatedAt", "UpdatedAt", "Owner.ID", "OwnerID"}}, c.wbList)
				fields := c.scope.Fields(Permission{Read: true})
				test.Equal(5, len(fields)) // ID,Brand,OwnerID,CreatedAt,UpdatedAt
				relations := c.scope.Relations(Permission{Read: true})
				test.Equal(2, len(relations)) // Owner, Liquid
			}

			// user added blacklist
			c.SetWBList(BLACKLIST, "Brand", "Owner.Name", "Liquid", "Liquid.Brand", "Liquid.ID")
			err = c.scope.setFieldPermission()
			if test.NoError(err) {
				// Other fields were removed because they are mandatory
				test.Equal(&whiteBlackList{policy: 0, fields: []string{"Brand", "Owner.Name", "Liquid"}}, c.wbList)
				fields := c.scope.Fields(Permission{Read: true})
				test.Equal(6, len(fields)) // ID,OwnerID,Type,YearCheck,CreatedAt,UpdatedAt
				relations := c.scope.Relations(Permission{Read: true})
				test.Equal(4, len(relations)) // all relations aside Liquid and the custom one.
			}
		}
	}
}

// testing if double entered keys will be unique in the wb list.
func TestModel_SetWBList_Unique(t *testing.T) {
	test := assert.New(t)

	c := &car{}
	err := c.Init(c)
	test.NoError(err)

	c.SetWBList(WHITELIST, "ID", "ID", "Owner.Name", "Owner.Name")
	test.NoError(err)

	test.Equal(WHITELIST, c.wbList.policy)
	test.Equal([]string{"ID", "Owner.Name"}, c.wbList.fields)

	err = addMandatoryFields(c.scope)
	test.NoError(err)
	test.Equal([]string{"ID", "Owner.Name", "CreatedAt", "UpdatedAt", "Owner.ID", "OwnerID"}, c.wbList.fields)
}

// Test_addMandatoryFields testing if all mandatory fields were added.
func Test_addMandatoryFields(t *testing.T) {
	test := assert.New(t)

	c := &car{}
	err := c.Init(c)
	err = addMandatoryFields(c.scope)
	test.NoError(err)
	test.Nil(c.wbList)

	// all fk,afk,poly and primary keys must be loaded that the relation relevant data is given.
	c.SetWBList(WHITELIST, "Owner.Name")
	err = addMandatoryFields(c.scope)
	test.NoError(err)
	test.Equal([]string{"Owner.Name", "ID", "CreatedAt", "UpdatedAt", "Owner.ID", "OwnerID"}, c.wbList.fields)

	// full relation name
	c.SetWBList(WHITELIST, "Owner")
	err = addMandatoryFields(c.scope)
	test.NoError(err)
	test.Equal([]string{"Owner", "ID", "CreatedAt", "UpdatedAt", "OwnerID"}, c.wbList.fields)

	// child relation
	c.SetWBList(WHITELIST, "Owner.Car.ID")
	err = addMandatoryFields(c.scope)
	test.NoError(err)
	test.Equal([]string{"Owner.Car.ID", "ID", "CreatedAt", "UpdatedAt", "Owner.ID", "OwnerID", "Owner.Car.OwnerID"}, c.wbList.fields)

	// polymorphic relation
	c.SetWBList(WHITELIST, "Liquid.Brand")
	err = addMandatoryFields(c.scope)
	test.NoError(err)
	test.Equal([]string{"Liquid.Brand", "ID", "CreatedAt", "UpdatedAt", "Liquid.ID", "Liquid.CarID", "Liquid.CarType"}, c.wbList.fields)

	// On blacklist the owner can name is disabled but the mandatory fields are deleted of the blacklist.
	c.SetWBList(BLACKLIST, "Owner.Name", "OwnerID", "ID", "Owner.ID")
	err = addMandatoryFields(c.scope)
	test.NoError(err)
	test.Equal([]string{"Owner.Name"}, c.wbList.fields)

	// full relation name - no additional keys needed to remove
	c.SetWBList(BLACKLIST, "Owner")
	err = addMandatoryFields(c.scope)
	test.NoError(err)
	test.Equal([]string{"Owner"}, c.wbList.fields)

	// child relation - theses are all mandatory keys and can not be removed.
	c.SetWBList(BLACKLIST, "Owner.Car.ID", "ID", "Owner.ID", "OwnerID", "Owner.Car.OwnerID")
	err = addMandatoryFields(c.scope)
	test.NoError(err)
	test.Nil(c.wbList)

	// child relation - theses are all mandatory keys and can not be removed.
	c.SetWBList(BLACKLIST, "Brand", "Owner.Car.Name", "Owner.Car.ID", "ID", "Owner.ID", "OwnerID", "Owner.Car.OwnerID")
	err = addMandatoryFields(c.scope)
	test.NoError(err)
	test.Equal([]string{"Brand", "Owner.Car.Name"}, c.wbList.fields)

	// polymorphic relation
	c.SetWBList(BLACKLIST, "Liquid.Brand", "ID", "Liquid.ID", "Liquid.CarID", "Liquid.CarType")
	err = addMandatoryFields(c.scope)
	test.NoError(err)
	test.Equal([]string{"Liquid.Brand"}, c.wbList.fields)

	o := &owner{}
	err = o.Init(o)
	test.NoError(err)
	o.SetWBList(BLACKLIST, "Name", "ID", "Car.ID", "Car.OwnerID")
	err = addMandatoryFields(o.scope)
	test.NoError(err)
	test.Equal([]string{"Name"}, o.wbList.fields)
}

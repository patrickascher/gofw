package orm2

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewWBList(t *testing.T) {
	test := assert.New(t)

	// whitelist factory
	w := NewWBList(WHITELIST, []string{"ID", "Name"})
	test.Equal(WHITELIST, w.policy)
	test.Equal([]string{"ID", "Name"}, w.fields)
	// blacklist factory
	b := NewWBList(BLACKLIST, []string{"ID", "Name"})
	test.Equal(BLACKLIST, b.policy)
	test.Equal([]string{"ID", "Name"}, b.fields)
}

func Test_setDefaultPermission(t *testing.T) {
	test := assert.New(t)

	m := &car{}
	err := m.setDefaultPermission()
	// err: not initialized, needed because of the cache.
	test.Error(err)

	// init model
	err = m.Init(m)
	test.NoError(err)

	// check default permission
	test.Equal(Permission{Read: true, Write: true}, m.fields[0].Permission)
	test.Equal(Permission{Read: true, Write: true}, m.fields[2].Permission)

	// manually manipulate permission
	m.fields[0].Permission = Permission{Read: false, Write: false}
	m.fields[2].Permission = Permission{Read: false, Write: false}
	test.Equal(Permission{Read: false, Write: false}, m.fields[0].Permission)
	test.Equal(Permission{Read: false, Write: false}, m.fields[2].Permission)

	// resetting the permission
	err = m.setDefaultPermission()
	test.NoError(err)
	test.Equal(Permission{Read: true, Write: true}, m.fields[0].Permission)
	test.Equal(Permission{Read: true, Write: true}, m.fields[2].Permission)

	// fake caller to test not existing in cache
	m.caller = &driver{}
	err = m.setDefaultPermission()
	test.Error(err)
	test.Equal("cache/memory: key  does not exist", err.Error())

}
func Test_addMandatoryFields(t *testing.T) {
	test := assert.New(t)

	driver := &driver{}
	err := driver.Init(driver)
	test.NoError(err)
	driver.First(nil)

	fmt.Println(driver.wbList)

	wheel := &wheel{}
	err = wheel.Init(wheel)
	test.NoError(err)
	wheel.First(nil)

	fmt.Println(wheel.wbList)
}
